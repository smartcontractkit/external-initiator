package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/tidwall/gjson"

	"go.uber.org/zap"
)

type FluxMonitorConfig struct {
	Adapters          []url.URL
	RequestData       string
	Multiply          int32
	Threshold         float64
	AbsoluteThreshold float64
	Heartbeat         time.Duration
	PollInterval      time.Duration

	RuntimeConfig store.RuntimeConfig
}

func ParseFMSpec(jsonSpec json.RawMessage, runtimeConfig store.RuntimeConfig) (FluxMonitorConfig, error) {
	var fmConfig FluxMonitorConfig

	res := gjson.GetBytes(jsonSpec, "feeds.#.url")
	var adapters []url.URL
	for _, adapter := range res.Array() {
		u, _ := url.Parse(adapter.String())
		adapters = append(adapters, *u)
	}

	fmConfig.Adapters = adapters
	fmConfig.RequestData = gjson.GetBytes(jsonSpec, "requestData").Raw
	fmConfig.Multiply = int32(gjson.GetBytes(jsonSpec, "precision").Int())
	fmConfig.Threshold = gjson.GetBytes(jsonSpec, "threshold").Float()
	fmConfig.AbsoluteThreshold = gjson.GetBytes(jsonSpec, "absoluteThreshold").Float()
	fmConfig.RuntimeConfig = runtimeConfig

	var err error
	if !gjson.GetBytes(jsonSpec, "idleTimer.disabled").Bool() {
		fmConfig.Heartbeat, err = time.ParseDuration(gjson.GetBytes(jsonSpec, "idleTimer.duration").String())
		if err != nil {
			return FluxMonitorConfig{}, errors.Wrap(err, "unable to parse idleTimer duration")
		}
		if fmConfig.Heartbeat < 1*time.Second {
			return FluxMonitorConfig{}, errors.New("idleTimer duration is less than 1s")
		}
	}
	if !gjson.GetBytes(jsonSpec, "pollTimer.disabled").Bool() {
		fmConfig.PollInterval, err = time.ParseDuration(gjson.GetBytes(jsonSpec, "pollTimer.period").String())
		if err != nil {
			return FluxMonitorConfig{}, errors.Wrap(err, "unable to parse pollTimer period")
		}
		if fmConfig.PollInterval < 1*time.Second {
			return FluxMonitorConfig{}, errors.New("pollTimer period is less than 1s")
		}
	}

	return fmConfig, nil
}

type AdapterResponse struct {
	Price *decimal.Decimal `json:"result"`
	// might need error response as well
}

type FluxMonitor struct {
	state  blockchain.FluxAggregatorState
	config FluxMonitorConfig

	blockchain blockchain.Manager

	latestResult           decimal.Decimal
	latestResultTimestamp  time.Time
	latestSubmittedRoundID uint32
	latestInitiatedRoundID uint32

	pollTicker     utils.PausableTicker
	idleTimer      utils.ResettableTimer
	idleTimerReset chan struct{}

	quitOnce   sync.Once
	tStart     sync.Once
	tStop      sync.Once
	checkMutex sync.Mutex
	pollMutex  sync.Mutex

	httpClient http.Client

	chJobTrigger  chan subscriber.Event
	chClose       chan struct{}
	chTickerClose chan struct{}

	logger *zap.SugaredLogger
}

func NewFluxMonitor(job string, config FluxMonitorConfig, triggerJobRun chan subscriber.Event, blockchainManager blockchain.Manager) (*FluxMonitor, error) {
	fm := FluxMonitor{
		config:  config,
		chClose: make(chan struct{}),
		httpClient: http.Client{
			Timeout: config.RuntimeConfig.FMAdapterTimeout,
		},
		blockchain:     blockchainManager,
		chJobTrigger:   triggerJobRun,
		pollTicker:     utils.NewPausableTicker(config.PollInterval),
		idleTimer:      utils.NewResettableTimer(),
		idleTimerReset: make(chan struct{}, 1),
		logger:         logger.Default.With("job", job),
	}
	fm.logger.Infof("New FluxMonitor with config: %+v", config)

	FAEvents := make(chan interface{})

	state, err := fm.blockchain.Request(blockchain.FMRequestState)
	if err != nil {
		return nil, err
	}

	faState, ok := state.(*blockchain.FluxAggregatorState)
	if !ok {
		return nil, errors.New("didn't receive valid FluxAggregatorState")
	}
	if faState == nil {
		return nil, errors.New("received nil FluxAggregatorState")
	}

	fm.state = *faState

	err = fm.blockchain.Subscribe(blockchain.FMSubscribeEvents, FAEvents)
	if err != nil {
		return nil, err
	}

	fm.canSubmitUpdated()
	go fm.eventListener(FAEvents)

	return &fm, nil
}

func (fm *FluxMonitor) Stop() {
	fm.quitOnce.Do(func() {
		close(fm.chClose)
	})
}

func (fm *FluxMonitor) canSubmitUpdated() {
	if fm.state.CanSubmit {
		logger.Info("Starting tickers because node is eligible to submit")
		fm.startTickers()
	} else {
		logger.Info("Stoping tickers because node is not eligible to submit")
		fm.stopTickers()
	}
}

func (fm *FluxMonitor) startTickers() {
	fm.tStart.Do(func() {
		fm.chTickerClose = make(chan struct{})
		go fm.pollingTicker()
		go fm.heartbeatTimer()
		fm.tStop = sync.Once{}
	})
}

func (fm *FluxMonitor) stopTickers() {
	if fm.chTickerClose != nil {
		fm.tStop.Do(func() {
			close(fm.chTickerClose)
			fm.tStart = sync.Once{}
		})
	}
}

func (fm *FluxMonitor) eventListener(ch <-chan interface{}) {
	defer fm.Stop()
	logger.Info("FM listening for events")
	for {
		select {
		case rawEvent := <-ch:
			switch event := rawEvent.(type) {
			case blockchain.FMEventNewRound:
				fm.logger.Debug("Got new round event: ", event)
				fm.resetHeartbeatTimer()
				fm.state.RoundID = event.RoundID
				if event.OracleInitiated {
					fm.latestInitiatedRoundID = event.RoundID
					continue
				}
				err := fm.checkAndSendJob(false)
				if err != nil {
					fm.logger.Error(err)
				}
			case blockchain.FMEventAnswerUpdated:
				fm.logger.Debug("Got answer updated event: ", event)
				fm.state.LatestAnswer = event.LatestAnswer
				fm.checkDeviation()
			case blockchain.FMEventPermissionsUpdated:
				fm.logger.Debug("Got permissions updated event: ", event)
				fm.state.CanSubmit = event.CanSubmit
				fm.canSubmitUpdated()
			}
		case <-fm.chClose:
			return
		}
	}
}

func (fm *FluxMonitor) canSubmitToRound(initiate bool) bool {
	if !fm.state.CanSubmit {
		fm.logger.Info("Oracle can't submit to this feed")
		return false
	}

	if initiate {
		if int32(fm.state.RoundID+1-fm.latestInitiatedRoundID) <= fm.state.RestartDelay {
			fm.logger.Info("Oracle needs to wait until restart delay passes until it can initiate a new round")
			return false
		}

		if fm.latestSubmittedRoundID >= fm.state.RoundID+1 {
			fm.logger.Info("Oracle already initiated this round")
			return false
		}
	} else {
		if fm.latestSubmittedRoundID != 0 && fm.latestSubmittedRoundID >= fm.state.RoundID {
			fm.logger.Info("Oracle already submitted to this round")
			return false
		}
	}

	return true
}

func (fm *FluxMonitor) checkAndSendJob(initiate bool) error {
	// Add a lock for checks so we prevent multiple rounds being started at the same time
	fm.checkMutex.Lock()
	defer fm.checkMutex.Unlock()

	if !fm.canSubmitToRound(initiate) {
		return errors.New("oracle can't submit to this round")
	}

	roundId := fm.state.RoundID
	if initiate {
		roundId++
	}

	jobRequest, err := fm.blockchain.CreateJobRun(blockchain.FMJobRun, roundId)
	if err != nil {
		return errors.Wrap(err, "failed to create job request from blockchain manager")
	}

	if !fm.ValidLatestResult() {
		err := fm.pollWithRetry()
		if err != nil {
			return err
		}
	}

	// Add common keys that should always be included
	jobRequest["result"] = fm.latestResult
	jobRequest["payment"] = fm.state.Payment.String()
	fm.logger.Info("Triggering Job Run: ", jobRequest)
	fm.chJobTrigger <- jobRequest

	fm.latestSubmittedRoundID = roundId
	return nil
}

func (fm *FluxMonitor) resetHeartbeatTimer() {
	if fm.config.Heartbeat != 0 {
		fm.idleTimerReset <- struct{}{}
	}
}

func (fm *FluxMonitor) heartbeatTimer() {
	if fm.config.Heartbeat == 0 {
		fm.logger.Info("Not starting heartbeat timer, because config.Heartbeat is not set")
		return
	}
	fm.logger.Info("Started heartbeat timer")

	fm.idleTimer.Reset(fm.config.Heartbeat)
	defer fm.idleTimer.Stop()

	for {
		select {
		case <-fm.idleTimerReset:
			fm.logger.Info("Resetting the heartbeat timer")
			fm.idleTimer.Reset(fm.config.Heartbeat)
			continue
		case <-fm.idleTimer.Ticks():
			fm.logger.Info("Heartbeat trigger")
			err := fm.checkAndSendJob(true)
			if err != nil {
				fm.logger.Error("Failed to initiate new round at heartbeat: ", err)
			}
		case <-fm.chClose:
			fm.logger.Info("FluxMonitor stopped")
			return
		case <-fm.chTickerClose:
			fm.logger.Info("Stopping heartbeat timer")
			return
		}
	}
}

func (fm *FluxMonitor) pollingTicker() {
	if fm.config.PollInterval == 0 {
		fm.logger.Info("Not starting polling adapters, because config.PollInterval is not set")
		return
	}

	fm.pollTicker.Resume()
	defer fm.pollTicker.Destroy()

	for {
		select {
		case <-fm.pollTicker.Ticks():
			fm.pollTicker.Pause()
			fm.logger.Info("polling adapters")
			err := fm.pollWithRetry()
			if err != nil {
				fm.logger.Error(err)
			} else {
				// We don't want to check deviation against an outdated value
				fm.checkDeviation()
			}
			fm.pollTicker.Resume()
		case <-fm.chClose:
			fm.logger.Info("stopping polling adapter")
			return
		case <-fm.chTickerClose:
			fm.logger.Info("stopping poller")
			return
		}
	}
}

// could use the fm service values but also can use arguments for standalone functionality and potential reuseability
// httpClient passing and initial setup could be handled in standalone service if we don't want to interfere with fm service state
func (fm *FluxMonitor) getAdapterResponse(endpoint url.URL, requestData string) (*decimal.Decimal, error) {
	fm.logger.Info("Requesting data from adapter: ", endpoint.String())
	resp, err := fm.httpClient.Post(endpoint.String(), "application/json", bytes.NewBuffer([]byte(requestData)))
	if err != nil {
		return nil, err
	}
	defer logger.ErrorIfCalling(resp.Body.Close)

	if resp.StatusCode == 400 {
		return nil, fmt.Errorf("%s returned 400", endpoint.String())
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %v from endpoint %s", resp.StatusCode, endpoint.String())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fm.logger.Infof("%s returned: %s ", endpoint.String(), body)

	var response AdapterResponse
	return response.Price, json.Unmarshal(body, &response)
}

func (fm *FluxMonitor) ValidLatestResult() bool {
	fm.pollMutex.Lock()
	defer fm.pollMutex.Unlock()

	if time.Since(fm.latestResultTimestamp) <= fm.config.PollInterval+fm.config.RuntimeConfig.FMAdapterTimeout {
		// The result we have is from within our polling interval, so we can use it
		fmt.Println("poll result is valid for use")
		return true
	}
	fmt.Println("poll result is outdated(or empty) and is not valid for use")
	return false
}

func (fm *FluxMonitor) pollWithRetry() error {
	fm.pollMutex.Lock()
	defer fm.pollMutex.Unlock()
	for i := 0; i < int(fm.config.RuntimeConfig.FMAdapterRetryAttempts); i++ {
		err := fm.poll()
		if err != nil {
			fm.logger.Error("Failed polling adapters: ", err)
			if i < int(fm.config.RuntimeConfig.FMAdapterRetryAttempts) {
				fm.logger.Debugf("Waiting %s before trying again...", fm.config.RuntimeConfig.FMAdapterRetryDelay.String())
				time.Sleep(fm.config.RuntimeConfig.FMAdapterRetryDelay)
			}
			continue
		}

		return nil
	}

	return fmt.Errorf("unable to get a poll result after %d attempts", fm.config.RuntimeConfig.FMAdapterRetryAttempts)
}

// poll() should only be called by pollWithRetry(), as it holds the mutex lock
func (fm *FluxMonitor) poll() error {
	numSources := len(fm.config.Adapters)
	ch := make(chan *decimal.Decimal)
	for _, adapter := range fm.config.Adapters {
		go func(adapter url.URL) {
			price, err := fm.getAdapterResponse(adapter, fm.config.RequestData)
			if err != nil {
				fm.logger.Error(fmt.Sprintf("Adapter response error. URL: %s requestData: %s error: ", adapter.Host, fm.config.RequestData), err)
				price = nil
			}
			ch <- price
		}(adapter)
	}

	var values []decimal.Decimal
	for i := 0; i < numSources; i++ {
		val := <-ch
		if val == nil {
			continue
		}
		values = append(values, *val)
	}

	if len(values) <= numSources/2 {
		return errors.New("unable to get values from more than 50% of data sources")
	}

	median := calculateMedian(values)
	fm.latestResult = median
	fm.latestResultTimestamp = time.Now()
	fm.logger.Info("Latest result from adapter polling: ", median)
	return nil
}

func (fm *FluxMonitor) checkDeviation() {
	if !outOfDeviation(decimal.NewFromBigInt(&fm.state.LatestAnswer, -fm.config.Multiply), fm.latestResult, fm.config.Threshold, fm.config.AbsoluteThreshold) {
		return
	}

	logger.ErrorIf(fm.checkAndSendJob(true))
}

func calculateMedian(prices []decimal.Decimal) decimal.Decimal {
	sort.Slice(prices, func(i, j int) bool {
		return (prices[i]).LessThan(prices[j])
	})
	mNumber := len(prices) / 2

	if len(prices)%2 == 1 {
		return prices[mNumber]
	}

	return (prices[mNumber-1].Add(prices[mNumber])).Div(decimal.NewFromInt(2))
}

func outOfDeviation(currentAnswer, nextAnswer decimal.Decimal, percentageThreshold, absoluteThreshold float64) bool {
	loggerFields := []interface{}{
		"threshold", percentageThreshold,
		"absoluteThreshold", absoluteThreshold,
		"currentAnswer", currentAnswer,
		"nextAnswer", nextAnswer,
	}
	logger.Infow(
		"Deviation checker ", loggerFields...)
	if absoluteThreshold == 0 && percentageThreshold == 0 {
		logger.Debugw(
			"Deviation thresholds both zero; short-circuiting deviation checker to "+
				"true, regardless of feed values", loggerFields...)
		return true
	}

	diff := currentAnswer.Sub(nextAnswer).Abs()
	loggerFields = append(loggerFields, "absoluteDeviation", diff)

	if !diff.GreaterThan(decimal.NewFromFloat(absoluteThreshold)) {
		logger.Debugw("Absolute deviation threshold not met", loggerFields...)
		return false
	}

	if currentAnswer.IsZero() {
		if nextAnswer.IsZero() {
			logger.Debugw("Relative deviation is undefined; can't satisfy threshold", loggerFields...)
			return false
		}
		logger.Infow("Threshold met: relative deviation is ∞", loggerFields...)
		return true
	}

	// 100*|new-old|/|old|: Deviation (relative to curAnswer) as a percentage
	percentage := diff.Div(currentAnswer.Abs()).Mul(decimal.NewFromInt(100))
	loggerFields = append(loggerFields, "percentage", percentage)

	if percentage.LessThan(decimal.NewFromFloat(percentageThreshold)) {
		logger.Debugw("Relative deviation threshold not met", loggerFields...)
		return false
	}

	logger.Infow("Relative and absolute deviation thresholds both met", loggerFields...)
	return true
}
