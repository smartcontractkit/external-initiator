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
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/tidwall/gjson"
)

type FluxMonitorConfig struct {
	Adapters          []url.URL
	From              string
	To                string
	Multiply          int32
	Threshold         float64
	AbsoluteThreshold float64
	Heartbeat         time.Duration
	PollInterval      time.Duration
	AdapterTimeout    time.Duration
}

func ParseFMSpec(jsonSpec json.RawMessage) (FluxMonitorConfig, error) {
	var fmConfig FluxMonitorConfig

	res := gjson.GetBytes(jsonSpec, "feeds.#.url")
	var adapters []url.URL
	for _, adapter := range res.Array() {
		u, _ := url.Parse(adapter.String())
		adapters = append(adapters, *u)
	}

	fmConfig.Adapters = adapters
	fmConfig.From = gjson.GetBytes(jsonSpec, "requestData.data.from").String()
	fmConfig.To = gjson.GetBytes(jsonSpec, "requestData.data.to").String()
	fmConfig.Multiply = int32(gjson.GetBytes(jsonSpec, "precision").Int())
	fmConfig.Threshold = gjson.GetBytes(jsonSpec, "threshold").Float()
	fmConfig.AbsoluteThreshold = gjson.GetBytes(jsonSpec, "absoluteThreshold").Float()

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

	pollTicker utils.PausableTicker
	idleTimer  utils.ResettableTimer

	quitOnce   sync.Once
	tStart     sync.Once
	tStop      sync.Once
	checkMutex sync.Mutex
	pollMutex  sync.Mutex

	httpClient http.Client

	chJobTrigger  chan subscriber.Event
	chClose       chan struct{}
	chTickerClose chan struct{}
}

func NewFluxMonitor(config FluxMonitorConfig, triggerJobRun chan subscriber.Event, blockchainManager blockchain.Manager) (*FluxMonitor, error) {
	logger.Infof("New FluxMonitor with config: %+v", config)

	fm := FluxMonitor{
		config:  config,
		chClose: make(chan struct{}),
		httpClient: http.Client{
			Timeout: config.AdapterTimeout,
		},
		blockchain:   blockchainManager,
		chJobTrigger: triggerJobRun,
		pollTicker:   utils.NewPausableTicker(config.PollInterval),
		idleTimer:    utils.NewResettableTimer(),
	}
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
	// make an initial submission on startup
	logger.ErrorIf(fm.checkAndSendJob(false))
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
		fm.startTickers()
	} else {
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
	fm.tStop.Do(func() {
		close(fm.chTickerClose)
		fm.tStart = sync.Once{}
	})
}

func (fm *FluxMonitor) eventListener(ch <-chan interface{}) {
	defer fm.Stop()

	for {
		select {
		case rawEvent := <-ch:
			switch event := rawEvent.(type) {
			case blockchain.FMEventNewRound:
				logger.Debug("Got new round event: ", event)
				fm.state.RoundID = event.RoundID
				if event.OracleInitiated {
					fm.latestInitiatedRoundID = event.RoundID
					continue
				}
				logger.ErrorIf(fm.checkAndSendJob(false))
				fm.idleTimer.Reset(fm.config.Heartbeat)
			case blockchain.FMEventAnswerUpdated:
				logger.Debug("Got answer updated event: ", event)
				fm.state.LatestAnswer = event.LatestAnswer
				fm.checkDeviation()
			case blockchain.FMEventPermissionsUpdated:
				logger.Debug("Got permissions updated event: ", event)
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
		logger.Info("Oracle can't submit to this feed")

		return false
	}

	if initiate && int32(fm.state.RoundID+1-fm.latestInitiatedRoundID) <= fm.state.RestartDelay {
		logger.Info("Oracle needs to wait until restart delay passes until it can initiate a new round")

		return false
	}

	if fm.latestSubmittedRoundID != 0 && fm.latestSubmittedRoundID >= fm.state.RoundID {
		logger.Info("Oracle already submitted to this round")

		return false
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
		logger.Error(err)
		return errors.New("failed to create job request from blockchain manager")
	}

	if time.Since(fm.latestResultTimestamp) > fm.config.PollInterval+fm.config.AdapterTimeout {
		logger.Warn("Polling again because result is old or have not been set yet")
		err := fm.poll()
		if err != nil {
			return errors.Wrap(err, "unable to get result from polling")
		}
	}

	// Add common keys that should always be included
	jobRequest["result"] = fm.latestResult.String()
	jobRequest["payment"] = fm.state.Payment.String()
	logger.Info("Triggering Job Run: ", jobRequest)
	fm.chJobTrigger <- jobRequest

	fm.latestSubmittedRoundID = fm.state.RoundID
	return nil
}

func (fm *FluxMonitor) heartbeatTimer() {
	if fm.config.Heartbeat == 0 {
		logger.Info("Not starting heartbeat ticker, because config.Heartbeat is not set")
		return
	}
	fm.idleTimer.Reset(fm.config.Heartbeat)
	defer fm.idleTimer.Stop()

	for {
		select {
		case <-fm.idleTimer.Ticks():
			logger.Info("Heartbeat")
			err := fm.poll()
			if err != nil {
				logger.Error("Failed to poll at heartbeat: ", err)
			}
			err = fm.checkAndSendJob(true)
			if err != nil {
				logger.Error("Failed to initiate new round at heartbeat: ", err)
			}
		case <-fm.chClose:
			logger.Info("FluxMonitor stopped")
			return
		case <-fm.chTickerClose:
			logger.Info("Stopping heartbeat timer")
			return
		}
	}
}

func (fm *FluxMonitor) pollingTicker() {
	if fm.config.PollInterval == 0 {
		logger.Info("Not starting polling adapters, because config.PollInterval is not set")
		return
	}
	logger.ErrorIf(fm.poll())
	fm.pollTicker.Resume()
	defer fm.pollTicker.Destroy()

	for {
		select {
		case <-fm.pollTicker.Ticks():
			fm.pollTicker.Pause()
			logger.Info("polling adapters")
			logger.ErrorIf(fm.poll())
			fm.checkDeviation()
			fm.pollTicker.Resume()
		case <-fm.chClose:
			fmt.Println("stopping polling adapter")
			return
		case <-fm.chTickerClose:
			logger.Info("stopping poller")
			return
		}
	}
}

// could use the fm service values but also can use arguments for standalone functionality and potential reuseability
// httpClient passing and initial setup could be handled in standalone service if we don't want to interfere with fm service state
func (fm *FluxMonitor) getAdapterResponse(endpoint url.URL, from string, to string) (*decimal.Decimal, error) {
	logger.Info("Requesting data from adapter: ", endpoint.String())
	data := map[string]string{"from": from, "to": to}
	values := map[string]interface{}{"id": "0", "data": data}
	payload, err := json.Marshal(values)

	if err != nil {
		return nil, err
	}

	resp, err := fm.httpClient.Post(endpoint.String(), "application/json",
		bytes.NewBuffer(payload))

	if err != nil {
		return nil, err
	}

	defer logger.ErrorIfCalling(resp.Body.Close)
	// potentially log the actual error messages
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

	logger.Infof("%s returned: %s ", endpoint.String(), body)

	var response AdapterResponse
	return response.Price, json.Unmarshal(body, &response)
}

func (fm *FluxMonitor) poll() error {
	numSources := len(fm.config.Adapters)
	ch := make(chan *decimal.Decimal)
	for _, adapter := range fm.config.Adapters {
		go func(adapter url.URL) {
			var price, _ = fm.getAdapterResponse(adapter, fm.config.From, fm.config.To)
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
	fm.pollMutex.Lock()
	fm.latestResult = median
	fm.latestResultTimestamp = time.Now()
	fm.pollMutex.Unlock()
	logger.Info("Latest result from adapter polling: ", median)
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
