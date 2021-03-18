package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink/core/logger"
)

type FluxAggregatorState struct {
	currentAnswer decimal.Decimal
}

type FluxMonitorConfig struct {
	multiply          *big.Int
	threshold         decimal.Decimal
	absoluteThreshold decimal.Decimal
	heartbeat         time.Duration
	pollInterval      time.Duration
}

type AdapterResponse struct {
	Price decimal.Decimal `json:"result"`
	// might need error response as well
}
type FluxMonitor struct {
	state  FluxAggregatorState
	config FluxMonitorConfig

	// subscriber subscriber.ISubscriber

	adapters []url.URL
	from     string
	to       string
	quitOnce sync.Once

	// chBlockchainEvents chan subscriber.Event
	// chDeviation chan *decimal.Decimal
	// chNewround  chan FluxAggregatorState
	chClose chan struct{}
}

func NewFluxMonitor(adapters []url.URL, from string, to string, multiply *big.Int, threshold decimal.Decimal, absoluteThreshold decimal.Decimal, heartbeat time.Duration, pollInterval time.Duration) *FluxMonitor {
	srv := FluxMonitor{
		adapters: adapters,
		from:     from,
		to:       to,
		config: FluxMonitorConfig{
			multiply:          multiply,
			threshold:         threshold,
			absoluteThreshold: absoluteThreshold,
			heartbeat:         heartbeat,
			pollInterval:      pollInterval,
		},
	}
	srv.startPoller()
	return &srv
}

// func (fm *FluxMonitor) hitTrigger() {
// 	timer := time.NewTimer(fm.config.heartbeat)
// 	defer timer.Stop()

// 	select {
// 	case <-fm.chNewround:
// 		fmt.Println("new round started")
// 	case <-fm.chDeviation:
// 		fmt.Println("hit deviation threshold")
// 	case <-timer.C:
// 		fmt.Println("heartbeat")
// 	case <-fm.chClose:
// 		fmt.Println("shut down")
// 	}
// }

func (fm *FluxMonitor) startPoller() {
	ticker := time.NewTicker(fm.config.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("polling adapters")
			fm.poll()
		case <-fm.chClose:
			fmt.Println("shut down")
		}
	}
}

func getAdapterResponse(endpoint url.URL, from string, to string) (*decimal.Decimal, error) {
	data := map[string]string{"from": from, "to": to}
	values := map[string]interface{}{"id": "0", "data": data}
	json_data, err := json.Marshal(values)

	if err != nil {
		fmt.Println("Marshal error: ", err)
		return nil, err
	}

	resp, err := http.Post(endpoint.String(), "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		fmt.Println(err)
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
		fmt.Println("ReadAll error: ", err)
		return nil, err
	}

	var response AdapterResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		fmt.Println("Unmarshal error: ", err)
		return nil, err
	}

	fmt.Println(response.Price)

	return &response.Price, nil

}

func (fm *FluxMonitor) poll() {
	numSources := len(fm.adapters)
	ch := make(chan *decimal.Decimal)
	for _, adapter := range fm.adapters {
		go func(adapter url.URL) {
			fmt.Println(adapter.String())
			var price, _ = getAdapterResponse(adapter, fm.from, fm.to)
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
		fmt.Println("Unable to get values from more than 50% of data sources")
		return
	}

	median := calculateMedian(values)
	fmt.Println("Median: ")
	fmt.Println(median)
	if outOfDeviation(fm.state.currentAnswer, median, fm.config.threshold, fm.config.absoluteThreshold) {
		fm.state.currentAnswer = median
	}

	// fm.chDeviation <- median
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

func (fm *FluxMonitor) stop() {
	fm.quitOnce.Do(func() {
		close(fm.chClose)
	})
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
