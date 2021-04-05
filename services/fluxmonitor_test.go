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
			LatestAnswer: *big.NewInt(40000),
			RoundID:      1,
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

func createMockAdapter(result string) *url.URL {
	var payload []byte
	var statusCode int
	if result != "" {
		statusCode = 200
		payload, _ = json.Marshal(map[string]interface{}{"jobRunID": "1", "result": result, "statusCode": statusCode})
	} else {
		statusCode = 400
		payload, _ = json.Marshal(map[string]interface{}{"jobRunID": "1", "status": "errored", "statusCode": statusCode})
	}

	mockAdapter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(fmt.Sprint(bytes.NewBuffer(payload))))
	}))
	adapter, _ := url.Parse(mockAdapter.URL)
	return adapter
}
func TestNewFluxMonitor(t *testing.T) {
	tests := []struct {
		name              string
		adapterResults    []string
		threshold         float64
		absoluteThreshold float64
		heartbeat         time.Duration
		pollInterval      time.Duration
		want              string
	}{
		{
			"1 adapter",
			[]string{"50000"},
			0.01,
			600,
			3 * time.Second,
			1 * time.Second,
			"50000",
		},
		{
			"2 adapters",
			[]string{"50000", "51000"},
			0.01,
			600,
			15 * time.Millisecond,
			3 * time.Millisecond,
			"50500",
		},
		{
			"3 adapters",
			[]string{"50000", "51000", "52000"},
			0.01,
			600,
			15 * time.Millisecond,
			3 * time.Millisecond,
			"51000",
		},
		{
			"3 adapters, 1 errored",
			[]string{"50000", "51000", ""},
			0.01,
			600,
			15 * time.Millisecond,
			3 * time.Millisecond,
			"50500",
		},
		{
			"3 adapters, 2 errored",
			[]string{"50000", "", ""},
			0.01,
			600,
			15 * time.Millisecond,
			3 * time.Millisecond,
			"no_job",
		},
		{
			"3 adapters, 1 errored, no heartbeat",
			[]string{"50000", "51000", ""},
			0.01,
			600,
			0,
			3 * time.Millisecond,
			"50500",
		},
		{
			"3 adapters, 1 errored, no polling",
			[]string{"50000", "51000", ""},
			0.01,
			600,
			15 * time.Millisecond,
			0,
			"50500",
		},
		{
			"3 adapters, 1 errored, no heartbeat, no polling",
			[]string{"50000", "51000", ""},
			0.01,
			600,
			0,
			0,
			"50500",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("Starting test: ", tt.name)

			var mockAdapters []url.URL
			for _, v := range tt.adapterResults {
				mockAdapters = append(mockAdapters, *createMockAdapter(v))
			}

			triggerJobRun := make(chan subscriber.Event, 10)
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
			fmt.Println(prettyPrint(fm.state))
			// time.Sleep(1 * time.Second)
			fmt.Println("New round event, initiated: ", fm.state.RoundID+1)
			FAEvents <- blockchain.FMEventNewRound{
				RoundID:         fm.state.RoundID + 1,
				OracleInitiated: false,
			}
			// FAEvents <- blockchain.FMEventAnswerUpdated{
			// 	LatestAnswer: fm.state.LatestAnswer,
			// }
			if tt.want == "no_job" {
				// timeout in case no job run is going to be send, in order not to continue running
				go func() {
					// might set this as variable
					time.Sleep(3 * time.Second)
					fmt.Println("Timeout of 3 seconds. Sending a mock job to end the test")
					triggerJobRun <- map[string]interface{}{"result": "no_job"}
				}()
			}

			job := <-triggerJobRun
			fmt.Println("Job triggered", job["result"])

			if got := job["result"]; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
			}

			fm.Stop()

		})
	}

}

// func TestNewFluxMonitor_Run(t *testing.T) {
// 	cryptoapis, _ := url.Parse("http://localhost:8081")
// 	coingecko, _ := url.Parse("http://localhost:8082")
// 	amberdata, _ := url.Parse("http://localhost:8083")
// 	triggerJobRun := make(chan subscriber.Event)
// 	var fmConfig FluxMonitorConfig

// 	fmConfig.Adapters = []url.URL{*cryptoapis, *coingecko, *amberdata}
// 	fmConfig.From = "BTC"
// 	fmConfig.To = "USD"
// 	fmConfig.Multiply = 18
// 	fmConfig.Threshold = decimal.NewFromFloat(0.01)
// 	fmConfig.AbsoluteThreshold = decimal.NewFromInt(0)
// 	fmConfig.Heartbeat = 15 * time.Second
// 	fmConfig.PollInterval = 1 * time.Second

// 	fm, err := NewFluxMonitor(fmConfig, triggerJobRun, &mockBlockchainManager{})
// 	require.NoError(t, err)
// 	go func() {
// 		for range time.Tick(time.Second * 2) {
// 			fmt.Println("New round event, initiated")
// 			FAEvents <- blockchain.FMEventNewRound{
// 				RoundID:         fm.state.RoundID + 1,
// 				OracleInitiated: true,
// 			}
// 		}
// 	}()
// 	go func() {
// 		for range time.Tick(time.Second * 7) {
// 			fmt.Println("New round event, not initiated")
// 			FAEvents <- blockchain.FMEventNewRound{
// 				RoundID:         fm.state.RoundID + 1,
// 				OracleInitiated: false,
// 			}
// 		}
// 	}()
// 	go func() {
// 		for range time.Tick(time.Second * 9) {
// 			newAnswer := &big.Int{}
// 			newAnswer = newAnswer.Add(&fm.state.LatestAnswer, big.NewInt(1))
// 			fmt.Println("Answer updated: ", newAnswer)
// 			FAEvents <- blockchain.FMEventAnswerUpdated{
// 				LatestAnswer: fm.state.LatestAnswer,
// 			}
// 		}
// 	}()
// 	// go func() {
// 	// 	for range time.Tick(time.Second * 17) {
// 	// 		fmt.Println("Permissions false")
// 	// 		FAEvents <- blockchain.FMEventPermissionsUpdated{
// 	// 			CanSubmit: false,
// 	// 		}
// 	// 	}
// 	// }()
// 	// go func() {
// 	// 	for range time.Tick(time.Second * 6) {
// 	// 		fmt.Println("Permissions true")
// 	// 		FAEvents <- blockchain.FMEventPermissionsUpdated{
// 	// 			CanSubmit: true,
// 	// 		}
// 	// 	}
// 	// }()
// 	for {
// 		job := <-triggerJobRun
// 		go func() {
// 			fmt.Println("Job triggered", job)
// 			fmt.Println("Current state", prettyPrint(fm.state))
// 		}()
// 	}
// }
