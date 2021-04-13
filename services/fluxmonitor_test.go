package services

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

func prettyPrint(i interface{}) string {
	// does not print big.int pointers
	// s, _ := json.MarshalIndent(i, "", "\t")
	// return string(s)
	return fmt.Sprintf("%+v", i)
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
		payload = []byte(fmt.Sprintf(`{"jobRunID": "1", "result": "%s", "statusCode": %d}`, result, statusCode))
	} else {
		statusCode = 400
		payload = []byte(fmt.Sprintf(`{"jobRunID": "1", "status": "errored", "statusCode": %d}`, statusCode))
	}

	mockAdapter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write(payload)
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
		// multiply          int32
		want string
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
		{
			"3 adapters, no threshold",
			[]string{"50500", "51000", "50000"},
			0,
			0,
			15 * time.Millisecond,
			3 * time.Millisecond,
			"50500",
		},
		{
			"3 adapters, 1 non-expected response",
			[]string{"50000", "51000", "wrong"},
			0.01,
			600,
			15 * time.Millisecond,
			3 * time.Millisecond,
			"50500",
		},
	}
	for _, tt := range tests {
		var mockAdapters []url.URL
		for _, v := range tt.adapterResults {
			mockAdapters = append(mockAdapters, *createMockAdapter(v))
		}

		triggerJobRun := make(chan subscriber.Event, 10)
		var fmConfig FluxMonitorConfig

		fmConfig.Adapters = mockAdapters
		fmConfig.RequestData = `{"from":"BTC","to":"USD"}`
		// fmConfig.Multiply = tt.multiply
		fmConfig.Threshold = tt.threshold
		fmConfig.AbsoluteThreshold = tt.absoluteThreshold
		fmConfig.Heartbeat = tt.heartbeat
		fmConfig.PollInterval = tt.heartbeat
		fmConfig.RuntimeConfig = store.RuntimeConfig{FMAdapterTimeout: 1 * time.Second, FMAdapterRetryAttempts: 1, FMAdapterRetryDelay: 1 * time.Second}

		t.Run("1 new round event tests: "+tt.name, func(t *testing.T) {
			fmt.Printf("Testing %s", t.Name())
			fm, err := NewFluxMonitor("test", fmConfig, triggerJobRun, &mockBlockchainManager{})
			require.NoError(t, err)
			defer fm.Stop()
			wg := sync.WaitGroup{}
			wg.Add(1)
			fmt.Println(prettyPrint(fm.state))
			fmt.Println("New round event, not initiated: ", fm.state.RoundID+1)
			FAEvents <- blockchain.FMEventNewRound{
				RoundID:         fm.state.RoundID + 1,
				OracleInitiated: false,
			}
			go func() {
				defer wg.Done()
				waitForTrigger(t, triggerJobRun, tt.want, 2*time.Second)
			}()
			wg.Wait()
		})

		t.Run("2 rounds tests: "+tt.name, func(t *testing.T) {
			fmt.Printf("Testing %s", t.Name())
			fm, err := NewFluxMonitor("test", fmConfig, triggerJobRun, &mockBlockchainManager{})
			require.NoError(t, err)
			defer fm.Stop()
			wg := sync.WaitGroup{}
			wg.Add(1)
			fmt.Println("FluxMonitor state: ", prettyPrint(fm.state))
			fmt.Println("1st round event, non initiated: ", fm.state.RoundID+1)
			FAEvents <- blockchain.FMEventNewRound{
				RoundID:         fm.state.RoundID + 1,
				OracleInitiated: false,
			}
			go func() {
				defer wg.Done()
				waitForTrigger(t, triggerJobRun, tt.want, 2*time.Second)
			}()
			wg.Wait()
			wg.Add(1)
			fmt.Println("FluxMonitor state: ", prettyPrint(fm.state))
			fmt.Println("2nd round event, non initiated: ", fm.state.RoundID+1)
			FAEvents <- blockchain.FMEventNewRound{
				RoundID:         fm.state.RoundID + 1,
				OracleInitiated: false,
			}
			go func() {
				defer wg.Done()
				waitForTrigger(t, triggerJobRun, tt.want, 2*time.Second)
			}()
			wg.Wait()
		})

		// TODO: could be handled better?
		// we want to check if job is triggered after only certain event, therefore makes sense to test only cases that do not have ticker triggers
		if strings.Contains(tt.name, "no heartbeat, no polling") {
			t.Run("Initiated round: "+tt.name, func(t *testing.T) {
				fmt.Printf("Testing %s", t.Name())
				fm, err := NewFluxMonitor("test", fmConfig, triggerJobRun, &mockBlockchainManager{})
				require.NoError(t, err)
				defer fm.Stop()
				wg := sync.WaitGroup{}
				wg.Add(1)
				fmt.Println("FluxMonitor state: ", prettyPrint(fm.state))
				fmt.Println("Initiated round event: ", fm.state.RoundID+1)
				FAEvents <- blockchain.FMEventNewRound{
					RoundID:         fm.state.RoundID + 1,
					OracleInitiated: true,
				}
				go func() {
					defer wg.Done()
					waitForTrigger(t, triggerJobRun, "no_job", 2*time.Second)
				}()
				wg.Wait()
			})

			t.Run("Answer updated: "+tt.name, func(t *testing.T) {
				fmt.Printf("Testing %s", t.Name())
				fm, err := NewFluxMonitor("test", fmConfig, triggerJobRun, &mockBlockchainManager{})
				require.NoError(t, err)
				defer fm.Stop()
				wg := sync.WaitGroup{}
				wg.Add(1)
				fmt.Println("FluxMonitor state: ", prettyPrint(fm.state))
				fmt.Println("Answer updated: ", fm.state.RoundID+1)
				FAEvents <- blockchain.FMEventAnswerUpdated{
					LatestAnswer: *fm.state.LatestAnswer.Add(&fm.state.LatestAnswer, big.NewInt(int64(fm.config.AbsoluteThreshold+1))),
				}
				go func() {
					defer wg.Done()
					waitForTrigger(t, triggerJobRun, tt.want, 2*time.Second)
				}()
				wg.Wait()
			})

			t.Run("Answer updated, but inside deviation threshold: "+tt.name, func(t *testing.T) {
				fmt.Printf("Testing %s", t.Name())
				fm, err := NewFluxMonitor("test", fmConfig, triggerJobRun, &mockBlockchainManager{})
				require.NoError(t, err)
				defer fm.Stop()
				wg := sync.WaitGroup{}
				wg.Add(1)
				fmt.Println("FluxMonitor state: ", prettyPrint(fm.state))
				fmt.Println("1st round event, non initiated: ", fm.state.RoundID+1)
				FAEvents <- blockchain.FMEventNewRound{
					RoundID:         fm.state.RoundID + 1,
					OracleInitiated: false,
				}
				go func() {
					defer wg.Done()
					waitForTrigger(t, triggerJobRun, tt.want, 2*time.Second)
				}()
				wg.Wait()
				wg.Add(1)

				fmt.Println("FluxMonitor state: ", prettyPrint(fm.state))
				fmt.Println("Answer updated first time: ", fm.state.RoundID+1)
				FAEvents <- blockchain.FMEventAnswerUpdated{
					LatestAnswer: *big.NewInt(fm.latestResult.IntPart()),
				}
				go func() {
					defer wg.Done()
					waitForTrigger(t, triggerJobRun, "no_job", 2*time.Second)
				}()
				wg.Wait()
				wg.Add(1)

				fmt.Println("FluxMonitor state: ", prettyPrint(fm.state))
				fmt.Println("Answer updated without deviation: ", fm.state.RoundID+1)
				FAEvents <- blockchain.FMEventAnswerUpdated{
					LatestAnswer: fm.state.LatestAnswer,
				}
				go func() {
					defer wg.Done()
					waitForTrigger(t, triggerJobRun, "no_job", 2*time.Second)
				}()
				wg.Wait()
			})

		}
	}
}

func waitForTrigger(t *testing.T, triggerJobRun chan subscriber.Event, want string, timeoutInterval time.Duration) {
	timeout := time.NewTimer(timeoutInterval)
	defer timeout.Stop()

	select {
	case job := <-triggerJobRun:
		fmt.Println("Job triggered", job["result"])
		if want == "no_job" {
			t.Errorf("Job received. Want %s", want)
		}
		if got := job["result"]; !reflect.DeepEqual(fmt.Sprintf("%s", got), fmt.Sprintf("%s", want)) {
			t.Errorf("GetTriggerJson() = %s, want %s", got, want)
		}

	case <-timeout.C:
		fmt.Println("Job timeout")
		if want != "no_job" {
			t.Errorf("No Job received, timeout. Want %s", want)
		}
	}
}
