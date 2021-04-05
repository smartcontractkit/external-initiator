package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/smartcontractkit/external-initiator/blockchain"

	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

var mockAdapterUrl string

func TestMain(m *testing.M) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.URL.Path != "/success" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":1234}`))
		return
	}))
	defer ts.Close()

	mockAdapterUrl = ts.URL

	code := m.Run()
	os.Exit(code)
}

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
		}, nil
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

func TestNewFluxMonitor(t *testing.T) {
	mockAdapter, err := url.Parse(mockAdapterUrl)
	require.NoError(t, err)
	successAdapter := *mockAdapter
	successAdapter.Path = "/success"

	triggerJobRun := make(chan subscriber.Event)
	var fmConfig FluxMonitorConfig

	fmConfig.Adapters = []url.URL{successAdapter}
	fmConfig.From = "BTC"
	fmConfig.To = "USD"
	fmConfig.Multiply = 18
	fmConfig.Threshold = 0.01
	fmConfig.AbsoluteThreshold = 0
	fmConfig.Heartbeat = 15 * time.Second
	fmConfig.PollInterval = 1 * time.Second

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		timeout := time.NewTimer(2 * time.Second)
		defer timeout.Stop()

		select {
		case <-triggerJobRun:
			fmt.Println("Got job run trigger")
		case <-timeout.C:
			t.Error("Did not get job run")
			t.Fail()
		}
	}()

	_, err = NewFluxMonitor(fmConfig, triggerJobRun, &mockBlockchainManager{})
	require.NoError(t, err)

	wg.Wait()
}
