package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/external-initiator/blockchain"

	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

// does not print big.int pointers
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

type mockBlockchainManager struct{}

var FAEvents = make(chan<- interface{})

func (sm *mockBlockchainManager) Request(t string) (interface{}, error) {
	switch t {
	case blockchain.FMRequestState:
		return &blockchain.FluxAggregatorState{
			CanSubmit:    true,
			LatestAnswer: *big.NewInt(50000),
			// RestartDelay: 2,
		}, nil
		// return &FluxAggregatorState{}, nil
		// maybe initialize with reasonable defaults
	}
	return nil, errors.New("request type is not implemented")
}

func (sm *mockBlockchainManager) Subscribe(t string, ch chan<- interface{}) error {
	switch t {
	case blockchain.FMSubscribeEvents:
		FAEvents = ch
		return nil
	}
	return errors.New("subscribe type is not implemented")
}

func (sm *mockBlockchainManager) CreateJobRun(t string, roundId uint32) (map[string]interface{}, error) {
	switch t {
	case blockchain.FMJobRun:
		return map[string]interface{}{}, nil
	}

	return nil, errors.New("job run type not implemented")
}

func createMockAdapter(result int64) *url.URL {
	payload, _ := json.Marshal(map[string]interface{}{"jobRunID": "1", "result": result, "statusCode": 200})
	mockAdapter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, bytes.NewBuffer(payload))
	}))
	adapter, _ := url.Parse(mockAdapter.URL)
	return adapter

	// return mockAdapter
}
func TestNewFluxMonitor(t *testing.T) {
	tests := []struct {
		name              string
		adapterResults    []int64
		threshold         decimal.Decimal
		absoluteThreshold decimal.Decimal
		heartbeat         time.Duration
		pollInterval      time.Duration
		want              string
	}{
		{
			"1 adapter",
			[]int64{50000},
			decimal.NewFromFloat(0.01),
			decimal.NewFromInt(600),
			15 * time.Millisecond,
			3 * time.Millisecond,
			"50000",
		},
		{
			"2 adapters",
			[]int64{50000, 51000},
			decimal.NewFromFloat(0.01),
			decimal.NewFromInt(600),
			15 * time.Millisecond,
			3 * time.Millisecond,
			"50500",
		},
		{
			"3 adapters",
			[]int64{50000, 51000, 52000},
			decimal.NewFromFloat(0.01),
			decimal.NewFromInt(600),
			15 * time.Millisecond,
			3 * time.Millisecond,
			"51000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var mockAdapters []url.URL
			for _, v := range tt.adapterResults {
				mockAdapters = append(mockAdapters, *createMockAdapter(v))
			}

			triggerJobRun := make(chan subscriber.Event)
			var fmConfig FluxMonitorConfig

			fmConfig.Adapters = mockAdapters
			fmConfig.From = "BTC"
			fmConfig.To = "USD"
			fmConfig.Multiply = 18
			fmConfig.Threshold = tt.threshold
			fmConfig.AbsoluteThreshold = tt.absoluteThreshold
			fmConfig.Heartbeat = tt.heartbeat
			fmConfig.PollInterval = tt.heartbeat

			fm, err := NewFluxMonitor(fmConfig, triggerJobRun, &mockBlockchainManager{})
			require.NoError(t, err)
			fmt.Println("New round event, initiated")
			FAEvents <- blockchain.FMEventNewRound{
				RoundID:         fm.state.RoundID + 1,
				OracleInitiated: true,
			}
			job := <-triggerJobRun
			fmt.Println("Job triggered", job["result"])
			fm.Stop()

			if err == nil {
				if got := job["result"]; !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
				}
			}
		})
	}
	// go func() {
	// 	for range time.Tick(time.Second * 2) {
	// 		fmt.Println("New round event, initiated")
	// 		FAEvents <- blockchain.FMEventNewRound{
	// 			RoundID:         fm.state.RoundID + 1,
	// 			OracleInitiated: true,
	// 		}
	// 	}
	// }()
	// go func() {
	// 	for range time.Tick(time.Second * 7) {
	// 		fmt.Println("New round event, not initiated")
	// 		FAEvents <- blockchain.FMEventNewRound{
	// 			RoundID:         fm.state.RoundID + 1,
	// 			OracleInitiated: false,
	// 		}
	// 	}
	// }()
	// go func() {
	// 	for range time.Tick(time.Second * 9) {
	// 		newAnswer := &big.Int{}
	// 		newAnswer = newAnswer.Add(&fm.state.LatestAnswer, big.NewInt(1))
	// 		fmt.Println("Answer updated: ", newAnswer)
	// 		fm.state.LatestAnswer = *newAnswer
	// 		FAEvents <- blockchain.FMEventAnswerUpdated{
	// 			LatestAnswer: fm.state.LatestAnswer,
	// 		}
	// 	}
	// }()
	// go func() {
	// 	for range time.Tick(time.Second * 17) {
	// 		fmt.Println("Permissions false")
	// 		FAEvents <- blockchain.FMEventPermissionsUpdated{
	// 			CanSubmit: false,
	// 		}
	// 	}
	// }()
	// go func() {
	// 	for range time.Tick(time.Second * 6) {
	// 		fmt.Println("Permissions true")
	// 		FAEvents <- blockchain.FMEventPermissionsUpdated{
	// 			CanSubmit: true,
	// 		}
	// 	}
	// }()
	// 	for {
	// 		job := <-triggerJobRun
	// 		go func() {
	// 			fmt.Println("Job triggered", job)
	// 			fmt.Println("Current state", prettyPrint(fm.state))
	// 		}()
	// 	}
}
