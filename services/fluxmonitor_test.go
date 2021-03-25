package services

import (
	"encoding/json"
	"fmt"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
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
	fmConfig.PollInterval = 5 * time.Second

	sub := store.Subscription{}

	fm, err := NewFluxMonitor(fmConfig, triggerJobRun, sub)
	require.NoError(t, err)
	canSubmit := true
	fm.state.CanSubmit = &canSubmit
	go func() {
		for range time.Tick(time.Second * 2) {
			fmt.Println("New round event")
			fm.chNewRound <- *fm.state.CurrentRoundID + 1
		}
	}()
	go func() {
		for range time.Tick(time.Second * 7) {
			fmt.Println("New answer event")
			fm.chAnswerUpdated <- fm.state.LatestAnswer.Add(decimal.NewFromInt32(10))
		}
	}()
	go func() {
		for range time.Tick(time.Second * 25) {
			fmt.Println("Oracle permissions changed")
			fm.chCanSubmit <- !*fm.state.CanSubmit
		}
	}()
	for {
		job := <-triggerJobRun
		go func() {
			fmt.Println("Job triggered", string(job))
			fmt.Println("Current state", prettyPrint(fm.state))
		}()
	}
}
