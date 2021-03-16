package services

import (
	"fmt"
	"math/big"
	"net/url"
	"sync"
	"time"

	"github.com/smartcontractkit/external-initiator/subscriber"
)

type fluxAggregatorState struct {
	currentAnswer *big.Int
}

type fluxMonitorConfig struct {
	multiply  *big.Int
	threshold *big.Int
	heartbeat time.Duration
}

type FluxMonitor struct {
	state  fluxAggregatorState
	config fluxMonitorConfig

	subscriber subscriber.ISubscriberNew

	adapters []url.URL

	quitOnce sync.Once

	chBlockchainEvents chan subscriber.Event
	chDeviation        chan *big.Int
	chNewround         chan fluxAggregatorState
	chClose            chan struct{}
}

func NewFluxMonitor(subscriber subscriber.ISubscriberNew) (*FluxMonitor, error) {
	return nil, nil
}

func (fm *FluxMonitor) hitTrigger() {
	timer := time.NewTimer(fm.config.heartbeat)
	defer timer.Stop()

	select {
	case <-fm.chNewround:
		fmt.Println("new round started")
	case <-fm.chDeviation:
		fmt.Println("hit deviation threshold")
	case <-timer.C:
		fmt.Println("heartbeat")
	case <-fm.chClose:
		fmt.Println("shut down")
	}
}

func (fm *FluxMonitor) startPoller(pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("do poll")
		case <-fm.chClose:
			fmt.Println("shut down")
		}
	}
}

func (fm *FluxMonitor) poll() {
	numSources := len(fm.adapters)
	ch := make(chan *big.Int)
	for _, adapter := range fm.adapters {
		go func(adapter url.URL) {
			// Poll the adapter
			fmt.Println(adapter.String())
			ch <- big.NewInt(123)
		}(adapter)
	}

	var values []*big.Int
	for i := 0; i < numSources; i++ {
		val := <-ch
		if val == nil {
			continue
		}
		values = append(values, val)
	}

	if len(values) < numSources/2 {
		fmt.Println("Unable to get values from more than 50% of data sources")
		return
	}

	median := getMedian(values)
	percDifference := getDifference(new(big.Int).Div(fm.state.currentAnswer, fm.config.multiply), median)
	if percDifference.Cmp(fm.config.threshold) <= 0 {
		return
	}

	fm.chDeviation <- median
}

func (fm *FluxMonitor) stop() {
	fm.quitOnce.Do(func() {
		close(fm.chClose)
	})
}

func quicksortBigint(valsRef *[]*big.Int) {
	vals := *valsRef
	if len(vals) < 1 {
		return
	}

	pivotIndex := len(vals) / 2
	var smallerItems []*big.Int
	var largerItems []*big.Int

	for i, val := range vals {
		if i == pivotIndex {
			continue
		}
		if val.Cmp(vals[pivotIndex]) < 0 {
			smallerItems = append(smallerItems, val)
		} else {
			largerItems = append(largerItems, val)
		}
	}

	quicksortBigint(&smallerItems)
	quicksortBigint(&largerItems)

	merged := append(append(smallerItems, vals[pivotIndex]), largerItems...)
	for i := 0; i < len(vals); i++ {
		vals[i] = merged[i]
	}

	valsRef = &vals
}

func getMedian(vals []*big.Int) *big.Int {
	if len(vals) < 1 {
		return nil
	}

	quicksortBigint(&vals)

	middleIndex := len(vals) / 2

	if len(vals)%2 != 0 {
		return vals[middleIndex]
	}

	return new(big.Int).Div(new(big.Int).Add(vals[middleIndex-1], vals[middleIndex]), big.NewInt(2))
}

func getDifference(v1 *big.Int, v2 *big.Int) *big.Int {
	absDiff := new(big.Int).Abs(new(big.Int).Sub(v1, v2))
	total := new(big.Int).Add(v1, v2)

	percDiff := new(big.Int).Div(absDiff, new(big.Int).Div(total, big.NewInt(2)))
	return new(big.Int).Mul(percDiff, big.NewInt(100))
}
