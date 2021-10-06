package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/utils"
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

func ParseFMSpec(jsonSpec json.RawMessage, runtimeConfig store.RuntimeConfig) (fmConfig FluxMonitorConfig, err error) {
	var fmSpecErrors []string

	fmSpec := gjson.GetBytes(jsonSpec, "fluxmonitor")

	res := fmSpec.Get("feeds.#.url")
	var adapters []url.URL
	for _, adapter := range res.Array() {
		u, _ := url.Parse(adapter.String())
		adapters = append(adapters, *u)
	}

	fmConfig.Adapters = adapters
	fmConfig.RequestData = fmSpec.Get("requestData").Raw
	fmConfig.Multiply = int32(fmSpec.Get("precision").Int())
	fmConfig.Threshold = fmSpec.Get("threshold").Float()
	fmConfig.AbsoluteThreshold = fmSpec.Get("absoluteThreshold").Float()
	fmConfig.RuntimeConfig = runtimeConfig

	if len(fmConfig.Adapters) == 0 {
		fmSpecErrors = append(fmSpecErrors, "at least one feed url should be specified. Example format: "+
			"["+
			"{ \"url\": \"http://localhost:8080\" },"+
			"{ \"url\": \"http://localhost:8081\" }"+
			"]")
	}

	//This is better not be strictly validated so it is flexible for different kind of requests
	if fmConfig.RequestData == "" {
		fmSpecErrors = append(fmSpecErrors, "no requestData")
	}
	if fmConfig.Multiply == 0 {
		fmSpecErrors = append(fmSpecErrors, "no precision")
	}
	if fmConfig.Threshold <= 0 {
		fmSpecErrors = append(fmSpecErrors, "'threshold' must be positive")
	}
	if fmConfig.AbsoluteThreshold < 0 {
		fmSpecErrors = append(fmSpecErrors, "'absoluteThreshold' must be non-negative")
	}

	if fmSpec.Get("pollTimer.disabled").Bool() && fmSpec.Get("idleTimer.disabled").Bool() {
		fmSpecErrors = append(fmSpecErrors, "must enable pollTimer, idleTimer, or both")
	}

	if !fmSpec.Get("idleTimer.disabled").Bool() {
		fmConfig.Heartbeat, err = time.ParseDuration(fmSpec.Get("idleTimer.duration").String())
		if err != nil {
			fmSpecErrors = append(fmSpecErrors, "unable to parse idleTimer duration")
		}
		if fmConfig.Heartbeat < 1*time.Second {
			fmSpecErrors = append(fmSpecErrors, "idleTimer duration is less than 1s")
		}
	}
	if !fmSpec.Get("pollTimer.disabled").Bool() {
		fmConfig.PollInterval, err = time.ParseDuration(fmSpec.Get("pollTimer.period").String())
		if err != nil {
			fmSpecErrors = append(fmSpecErrors, "unable to parse pollTimer period")
		}
		if fmConfig.PollInterval < 1*time.Second {
			fmSpecErrors = append(fmSpecErrors, "pollTimer period is less than 1s")
		}
	}

	if len(fmSpecErrors) > 0 {
		return fmConfig, errors.New(strings.Join(fmSpecErrors, ", "))
	}

	return
}

type AdapterResponse struct {
	Price *decimal.Decimal `json:"result"`
	// might need error response as well
}

type FluxMonitor struct {
	state  common.FluxAggregatorState
	config FluxMonitorConfig

	blockchain common.FluxMonitorManager

	latestResult           decimal.Decimal
	latestResultTimestamp  time.Time
	latestSubmittedRoundID uint32
	latestRoundTimestamp   int64
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

func NewFluxMonitor(job string, config FluxMonitorConfig, triggerJobRun chan subscriber.Event, blockchainManager common.FluxMonitorManager) (*FluxMonitor, error) {
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

	err := fm.GetState()
	if err != nil {
		return nil, err
	}

	err = fm.blockchain.SubscribeEvents(context.TODO(), FAEvents)
	if err != nil {
		return nil, err
	}

	fm.canSubmitUpdated()
	go fm.eventListener(FAEvents)

	return &fm, nil
}

func (fm *FluxMonitor) GetState() error {
	state, err := fm.blockchain.GetState(context.TODO())
	if err != nil {
		return err
	} else if state == nil {
		return errors.New("received nil FluxAggregatorState")
	}

	fm.logger.Infof("[GetState]: (on-chain)LastReportedRoundId(%d), (on-chain)LastStartedRoundId (%d),(local)latestSubmittedRoundID (%d), (local)latestInitiatedRoundID (%d), syncing local state", state.LastReportedRoundId, state.LastStartedRoundId, fm.latestSubmittedRoundID, fm.latestInitiatedRoundID)
	fm.state = *state
	fm.latestSubmittedRoundID = state.LastReportedRoundId
	fm.latestInitiatedRoundID = state.LastStartedRoundId
	return nil
}

func (fm *FluxMonitor) Stop() {
	fm.quitOnce.Do(func() {
		close(fm.chClose)
		fm.blockchain.Stop()
	})
}

func (fm *FluxMonitor) canSubmitUpdated() {
	if fm.state.CanSubmit {
		fm.logger.Info("Oracle is eligible to submit")
		fm.startTickers()
	} else {
		fm.logger.Info("Oracle is not eligible to submit")
		fm.stopTickers()
	}
}

func (fm *FluxMonitor) startTickers() {
	fm.tStart.Do(func() {
		fm.logger.Info("Starting tickers")
		fm.chTickerClose = make(chan struct{})
		go fm.pollingTicker()
		go fm.heartbeatTimer()
		fm.tStop = sync.Once{}
	})
}

func (fm *FluxMonitor) stopTickers() {
	if fm.chTickerClose != nil {
		fm.logger.Info("Stoping tickers")
		fm.tStop.Do(func() {
			close(fm.chTickerClose)
			fm.tStart = sync.Once{}
		})
	}
}

func (fm *FluxMonitor) eventListener(ch <-chan interface{}) {
	defer fm.Stop()
	fm.logger.Info("FM listening for events")
	for {
		select {
		case rawEvent := <-ch:
			switch event := rawEvent.(type) {
			// case common.FMSubmissionReceived:
			// 	fm.logger.Debugf("Got submission received event: %+v", event)
			// 	if fm.latestSubmittedRoundID == event.RoundID {
			// 		fm.state.LatestAnswer = event.LatestAnswer // tracks it's latest round submission between AnswerUpdated events (otherwise will keep trying to submit if rounds unfulfilled)
			// 	}
			case common.FMEventNewRound:
				fm.logger.Debugf("Got new round event: %+v", event)
				fm.resetHeartbeatTimer()
				fm.state.RoundID = event.RoundID
				fm.latestRoundTimestamp = time.Now().Unix() // track when new round is started to keep track of timeout for starting new round
				if event.OracleInitiated || fm.latestSubmittedRoundID >= event.RoundID {
					continue
				}
				err := fm.checkAndSendJob(false)
				if err != nil {
					fm.logger.Error(err)
				}
			case common.FMEventAnswerUpdated:
				fm.logger.Debugf("Got answer updated event: %+v", event)
				fm.state.LatestAnswer = event.LatestAnswer
				fm.state.RoundID = event.RoundID
				fm.latestRoundTimestamp = 0 // set to 0 when round complete (timeout not needed to be taken into account, rounds can be initiated whenever)
				fm.checkDeviation()
			case common.FMEventPermissionsUpdated:
				fm.logger.Debugf("Got permissions updated event: %+v", event)
				fm.state.CanSubmit = event.CanSubmit
				fm.canSubmitUpdated()
			case common.FMEventRoundDetailsUpdated:
				fm.logger.Debugf("Got round details updated event: %+v", event)
				fm.state.Payment = event.Payment
				fm.state.Timeout = event.Timeout
				fm.state.RestartDelay = event.RestartDelay
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
		// temporarily disabling some checks below

		// if int32(fm.state.RoundID+1-fm.latestInitiatedRoundID) <= fm.state.RestartDelay {
		// 	fm.logger.Info("Oracle needs to wait until restart delay passes until it can initiate a new round")
		// 	return false
		// }

		// check if timeout parameter is met (timestamp of 0 means new round can be initiated without waiting for timeout)
		current := time.Now().Unix()
		if fm.latestRoundTimestamp > 0 && uint32(current-fm.latestRoundTimestamp) < fm.state.Timeout {
			fm.logger.Info("Oracle needs to wait until timeout delay passes until it can ignore the current round and initiate a new round")
			fm.logger.Debugf("[fluxmonitor/canSubmitToRound] latestRound (%d), current (%d), %d < %d", fm.latestRoundTimestamp, current, uint32(current-fm.latestRoundTimestamp), fm.state.Timeout)
			return false
		}
		fm.logger.Debugf("[fluxmonitor/canSubmitToRound] timeout check passed, %d > %d", uint32(current-fm.latestRoundTimestamp), fm.state.Timeout)

		// if fm.latestInitiatedRoundID >= fm.state.RoundID+1 {
		// 	fm.logger.Info("Oracle already initiated this round")
		// 	return false
		// }
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
	if err := fm.GetState(); err != nil {
		fm.logger.Error(err)
	}
	roundId := fm.state.RoundID

	if initiate {
		if roundId < fm.latestSubmittedRoundID {
			roundId = fm.latestSubmittedRoundID
		}
		roundId++

	}

	if !fm.canSubmitToRound(initiate) {
		return errors.New("oracle can't submit to this round")
	}

	if !fm.ValidLatestResult() {
		err := fm.pollWithRetry()
		if err != nil {
			return err
		}
	}

	// check if latest result to submit is within bounds
	vString := strings.ReplaceAll(fm.latestResult.StringFixed(int32(fm.state.Decimals)), ".", "") // convert to scaled value
	var value big.Int
	value.SetString(vString, 10)
	if value.Cmp(&fm.state.MinSubmission) < 0 {
		return fmt.Errorf("[fluxmonitor/checkAndSendJob]: latest result (%s) below min submission (%s)", value.String(), fm.state.MinSubmission.String())
	}
	if value.Cmp(&fm.state.MaxSubmission) > 0 {
		return fmt.Errorf("[fluxmonitor/checkAndSendJob]: latest result (%s) above max submission (%s)", value.String(), fm.state.MaxSubmission.String())
	}

	jobRequest := fm.blockchain.CreateJobRun(roundId)
	// Add common keys that should always be included
	jobRequest["result"] = fm.latestResult.String()
	jobRequest["payment"] = fm.state.Payment.String()
	fm.logger.Info("Triggering Job Run: ", jobRequest)
	fm.chJobTrigger <- jobRequest

	fm.latestSubmittedRoundID = roundId
	if initiate {
		fm.latestInitiatedRoundID = roundId
	}
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
			fm.idleTimer.Reset(fm.config.Heartbeat)
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
		fm.logger.Info("poll result is valid for use")
		return true
	}
	fm.logger.Info("poll result is outdated (or empty) and is not valid for use")
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

	if len(values) == 0 {
		return errors.New("unable to get values from any of the data sources")
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
	// Temporarily disabling it this way, to have the routines and process as usual, for better actual testing.
	// Will need to disable the whole service potentially for Terra
	// logger.ErrorIf(fm.checkAndSendJob(true))
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
