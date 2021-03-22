package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink/core/logger"
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
}

func ParseFMSpec(jsonSpec json.RawMessage) FluxMonitorConfig {
	var fmConfig FluxMonitorConfig
	res := gjson.GetBytes(jsonSpec, "feeds.#.url")
	var adapters []url.URL
	for _, adapter := range res.Array() {
		url, _ := url.Parse(adapter.String())
		adapters = append(adapters, *url)
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

type FluxAggregatorState struct {
	currentAnswer decimal.Decimal
}

type AdapterResponse struct {
	Price decimal.Decimal `json:"result"`
	// might need error response as well
}
type FluxMonitor struct {
	state  FluxAggregatorState
	config FluxMonitorConfig

	// subscriber subscriber.ISubscriber
	latestResult decimal.Decimal

	// quitOnce     sync.Once

	// chBlockchainEvents chan subscriber.Event
	chDeviation chan decimal.Decimal
	// chNewround  chan FluxAggregatorState
	// chClose chan struct{}
}

func NewFluxMonitor(config FluxMonitorConfig, triggerJobRun chan decimal.Decimal) *FluxMonitor {
	logger.Info(fmt.Sprintf("New FluxMonitor with config: %+v", config))
	srv := FluxMonitor{
		config: config,
	}
	go srv.startPoller()
	srv.hitTrigger(triggerJobRun)
	return &srv
}

func (fm *FluxMonitor) hitTrigger(triggerJobRun chan decimal.Decimal) {
	fm.chDeviation = make(chan decimal.Decimal)
	ticker := time.NewTicker(fm.config.Heartbeat)
	defer ticker.Stop()
	for {

		select {
		// case <-fm.chNewround:
		// 	fmt.Println("New round started")
		// 	fm.state.currentAnswer = fm.latestResult
		// 	fmt.Println("New answer: ", fm.state.currentAnswer)
		case <-fm.chDeviation:
			logger.Info("Deviation threshold hit")
			fm.state.currentAnswer = <-fm.chDeviation
			triggerJobRun <- fm.state.currentAnswer
			logger.Info("New answer: ", fm.state.currentAnswer)
		case <-ticker.C:
			logger.Info("Heartbeat")
			// if adapters not working this is going to resubmit an old value
			fm.state.currentAnswer = fm.latestResult
			triggerJobRun <- fm.state.currentAnswer

			logger.Info("New answer: ", fm.state.currentAnswer)
			// 	// case <-fm.chClose:
			// 	// fmt.Println("shut down")
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
			// case <-fm.chClose:
			// 	fmt.Println("shut down")
		}
	}
}

func getAdapterResponse(endpoint url.URL, from string, to string) (*decimal.Decimal, error) {
	logger.Info("Requesting data from adapter: ", endpoint.String())
	data := map[string]string{"from": from, "to": to}
	values := map[string]interface{}{"id": "0", "data": data}
	json_data, err := json.Marshal(values)

	if err != nil {
		logger.Error("Marshal error: ", err)
		return nil, err
	}
	//TODO: initialize this outside of method and use env var to set the timeout
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Post(endpoint.String(), "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		logger.Error(err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		return nil, fmt.Errorf("%s returned 400", endpoint.String())
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %v from endpoint %s", resp.StatusCode, endpoint.String())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ReadAll error: ", err)
		return nil, err
	}

	var response AdapterResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		fmt.Println("Unmarshal error: ", err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("Response from %s: %s ", endpoint.String(), response.Price))
	return &response.Price, nil

}

func (fm *FluxMonitor) poll() {
	numSources := len(fm.config.Adapters)
	ch := make(chan *decimal.Decimal)
	for _, adapter := range fm.config.Adapters {
		go func(adapter url.URL) {
			var price, _ = getAdapterResponse(adapter, fm.config.From, fm.config.To)
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
		logger.Info("Unable to get values from more than 50% of data sources")
		return
	}

	median := calculateMedian(values)
	fm.latestResult = median
	logger.Info("Latest result: ", median)
	if outOfDeviation(fm.state.currentAnswer, fm.latestResult, fm.config.Threshold, fm.config.AbsoluteThreshold) {
		fm.chDeviation <- fm.latestResult
	}

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

// func (fm *FluxMonitor) stop() {
// 	fm.quitOnce.Do(func() {
// 		close(fm.chClose)
// 	})
// }

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
