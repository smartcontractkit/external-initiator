package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/tidwall/gjson"
)

type FluxMonitorConfig struct {
	Adapters          []url.URL
	From              string
	To                string
	Multiply          int64
	Threshold         decimal.Decimal
	AbsoluteThreshold decimal.Decimal
	Heartbeat         time.Duration
	PollInterval      time.Duration
	AdapterTimeout    time.Duration
}

func ParseFMSpec(jsonSpec json.RawMessage) FluxMonitorConfig {
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
	fmConfig.Multiply = gjson.GetBytes(jsonSpec, "precision").Int()
	fmConfig.Threshold, _ = decimal.NewFromString(gjson.GetBytes(jsonSpec, "threshold").String())
	fmConfig.AbsoluteThreshold, _ = decimal.NewFromString(gjson.GetBytes(jsonSpec, "absoluteThreshold").String())
	fmConfig.Heartbeat, _ = time.ParseDuration(gjson.GetBytes(jsonSpec, "idleTimer.duration").String())
	fmConfig.PollInterval, _ = time.ParseDuration(gjson.GetBytes(jsonSpec, "pollTimer.period").String())
	return fmConfig
}

type AdapterResponse struct {
	Price decimal.Decimal `json:"result"`
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

	quitOnce   sync.Once
	tStart     sync.Once
	tStop      sync.Once
	checkMutex sync.Mutex

	httpClient http.Client

	chJobTrigger  chan subscriber.Event
	chClose       chan struct{}
	chTickerClose chan struct{}
}

func NewFluxMonitor(config FluxMonitorConfig, triggerJobRun chan subscriber.Event, blockchainManager blockchain.Manager) (*FluxMonitor, error) {
	logger.Infof("New FluxMonitor with config: %+v", config)

	fm := FluxMonitor{
		config:        config,
		chClose:       make(chan struct{}),
		chTickerClose: make(chan struct{}),
		httpClient: http.Client{
			Timeout: config.AdapterTimeout,
		},
		blockchain:   blockchainManager,
		chJobTrigger: triggerJobRun,
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
	// make an initial sumbission on startup
	fm.checkAndSendJob(false)
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
		go fm.startPoller()
		go fm.hitTrigger()
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
				fm.state.RoundID = event.RoundID
				if event.OracleInitiated {
					fm.latestInitiatedRoundID = event.RoundID
					continue
				}
				fm.checkAndSendJob(false)
			case blockchain.FMEventAnswerUpdated:
				fmt.Println("State change")
				fm.state.LatestAnswer = event.LatestAnswer
				fm.checkDeviation()
			case blockchain.FMEventPermissionsUpdated:
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

	if fm.latestSubmittedRoundID >= fm.state.RoundID {
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
		return err
	}

	// If latestResult is an old value or have not been set yet, try to fetch new
	if time.Since(fm.latestResultTimestamp) > fm.config.PollInterval+fm.config.AdapterTimeout {
		logger.Warn("Polling again because result is old")
		err := fm.poll()
		if err != nil {
			logger.Error("cannot retrieve result from polling")
			return err
		}
	}

	// Add common keys that should always be included
	jobRequest["result"] = fm.latestResult.String()
	jobRequest["payment"] = fm.state.Payment.String()
	logger.Info("Triggering Job Run with latest result: ", fm.latestResult)
	fm.chJobTrigger <- jobRequest

	fm.latestSubmittedRoundID = fm.state.RoundID
	return nil
}

func (fm *FluxMonitor) hitTrigger() {
	ticker := time.NewTicker(fm.config.Heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("Heartbeat")
			fm.poll()
			fm.checkAndSendJob(true)
		case <-fm.chClose:
			logger.Info("FluxMonitor stopped")
			return
		case <-fm.chTickerClose:
			logger.Info("Stopping heartbeat timer")
			return
		}
	}
}

func (fm *FluxMonitor) startPoller() {
	fm.poll()
	ticker := time.NewTicker(fm.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("polling adapters")
			fm.poll()
			fm.checkDeviation()
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

	var response AdapterResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}
	logger.Infof("Response from %s: %s ", endpoint.String(), response.Price)
	return &response.Price, nil

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

	var values []*decimal.Decimal
	for i := 0; i < numSources; i++ {
		val := <-ch
		if val == nil {
			continue
		}
		values = append(values, val)
	}

	if len(values) <= numSources/2 {
		return errors.New("unable to get values from more than 50% of data sources")
	}

	median := calculateMedian(values)
	fm.latestResult = median
	fm.latestResultTimestamp = time.Now()
	logger.Info("Latest result from adapter polling: ", median)
	return nil
}

func (fm *FluxMonitor) checkDeviation() {
	if !outOfDeviation(decimal.NewFromBigInt(&fm.state.LatestAnswer, 0), fm.latestResult, fm.config.Threshold, fm.config.AbsoluteThreshold) {
		return
	}

	fm.checkAndSendJob(true)
}

func calculateMedian(prices []*decimal.Decimal) decimal.Decimal {
	sort.Slice(prices, func(i, j int) bool {
		return (*prices[i]).LessThan(*prices[j])
	})
	mNumber := len(prices) / 2

	if len(prices)%2 == 1 {
		return *prices[mNumber]
	}

	return (prices[mNumber-1].Add(*prices[mNumber])).Div(decimal.NewFromInt(2))
}

func outOfDeviation(currentAnswer, nextAnswer, percentageThreshold, absoluteThreshold decimal.Decimal) bool {
	loggerFields := []interface{}{
		"threshold", percentageThreshold,
		"absoluteThreshold", absoluteThreshold,
		"currentAnswer", currentAnswer,
		"nextAnswer", nextAnswer,
	}
	logger.Infow(
		"Deviation checker ", loggerFields...)
	if absoluteThreshold == decimal.NewFromInt(0) && percentageThreshold == decimal.NewFromInt(0) {
		logger.Debugw(
			"Deviation thresholds both zero; short-circuiting deviation checker to "+
				"true, regardless of feed values", loggerFields...)
		return true
	}

	diff := currentAnswer.Sub(nextAnswer).Abs()

	if !diff.GreaterThan(absoluteThreshold) {
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

	if percentage.LessThan(percentageThreshold) {
		logger.Debugw("Relative deviation threshold not met", loggerFields...)
		return false
	}
	logger.Infow("Relative and absolute deviation thresholds both met", loggerFields...)
	return true
}
