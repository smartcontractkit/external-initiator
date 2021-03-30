package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
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

func (sm *mockBlockchainManager) CreateJobRun(t string, params interface{}) (map[string]interface{}, error) {
	switch t {
	case blockchain.FMJobRun:
		return map[string]interface{}{}, nil
	}

	return nil, errors.New("job run type not implemented")
}
func TestNewFluxMonitor(t *testing.T) {
	cryptoapis, _ := url.Parse("http://localhost:8081")
	coingecko, _ := url.Parse("http://localhost:8082")
	amberdata, _ := url.Parse("http://localhost:8083")
	triggerJobRun := make(chan subscriber.Event)
	var fmConfig FluxMonitorConfig

	fmConfig.Adapters = []url.URL{*cryptoapis, *coingecko, *amberdata}
	fmConfig.From = "BTC"
	fmConfig.To = "USD"
	fmConfig.Multiply = 18
	fmConfig.Threshold = decimal.NewFromFloat(0.01)
	fmConfig.AbsoluteThreshold = decimal.NewFromInt(0)
	fmConfig.Heartbeat = 15 * time.Second
	fmConfig.PollInterval = 1 * time.Second

	fm, err := NewFluxMonitor(fmConfig, triggerJobRun, &mockBlockchainManager{})
	require.NoError(t, err)
	go func() {
		for range time.Tick(time.Second * 2) {
			fmt.Println("New round event, initiated")
			FAEvents <- blockchain.FMEventNewRound{
				RoundID:         fm.state.RoundID + 1,
				OracleInitiated: true,
			}
		}
	}()
	go func() {
		for range time.Tick(time.Second * 7) {
			fmt.Println("New round event, not initiated")
			FAEvents <- blockchain.FMEventNewRound{
				RoundID:         fm.state.RoundID + 1,
				OracleInitiated: false,
			}
		}
	}()
	go func() {
		for range time.Tick(time.Second * 9) {
			newAnswer := &big.Int{}
			newAnswer = newAnswer.Add(&fm.state.LatestAnswer, big.NewInt(1))
			fmt.Println("Answer updated: ", newAnswer)
			FAEvents <- blockchain.FMEventAnswerUpdated{
				LatestAnswer: fm.state.LatestAnswer,
			}
		}
	}()
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
	for {
		job := <-triggerJobRun
		go func() {
			fmt.Println("Job triggered", job)
			fmt.Println("Current state", prettyPrint(fm.state))
		}()
	}
}
