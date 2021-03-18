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
	"time"

	"github.com/shopspring/decimal"
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
}
type FluxMonitor struct {
	state  FluxAggregatorState
	config FluxMonitorConfig

	// subscriber subscriber.ISubscriber

	adapters []url.URL
	from     string
	to       string
	// quitOnce sync.Once

	// chBlockchainEvents chan subscriber.Event
	// chDeviation chan *decimal.Decimal
	// chNewround  chan FluxAggregatorState
	// chClose     chan struct{}
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
			// case <-fm.chClose:
			// 	fmt.Println("shut down")
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
	fm.state.currentAnswer = median
	// percDifference := getDifference(new(big.Int).Div(fm.state.currentAnswer, fm.config.multiply), median)
	// if percDifference.Cmp(fm.config.threshold) <= 0 {
	// 	return
	// }

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

// func (fm *FluxMonitor) stop() {
// 	fm.quitOnce.Do(func() {
// 		close(fm.chClose)
// 	})
// }

// func getDifference(v1 *decimal.Decimal, v2 *decimal.Decimal) *decimal.Decimal {
// 	absDiff := new(decimal.Decimal).Abs(new(decimal.Decimal).Sub(*v1, *v2))
// 	total := new(decimal.Decimal).Add(*v1, *v2)

// 	percDiff := new(decimal.Decimal).Div(absDiff, new(decimal.Decimal).Div(total, decimal.NewFromInt(2)))
// 	return new(decimal.Decimal).Mul(percDiff, decimal.NewFromInt(100))
// }
