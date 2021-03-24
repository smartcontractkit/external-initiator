package services

import (
	"encoding/json"
	"fmt"
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
	fm := NewFluxMonitor(fmConfig, triggerJobRun)
	fm.state.LatestRoundID = 1
	fm.state.CurrentRoundID = 2
	fm.state.LatestAnswer = decimal.NewFromInt(50000)
	fm.state.CanSubmit = true

	go func() {
		for range time.Tick(time.Second * 25) {
			fm.state.LatestRoundID += 1
			fm.state.CurrentRoundID += 1
			fm.state.LatestAnswer = decimal.NewFromInt32(50000 + fm.state.LatestRoundID)
			fmt.Println("New state", prettyPrint(fm.state))
		}
	}()
	go func() {
		for range time.Tick(time.Second * 35) {
			fm.state.CanSubmit = !fm.state.CanSubmit
			fmt.Println("New state", prettyPrint(fm.state))
		}
	}()
	for {
		// we should send here event from FluxMonitor
		job := <-triggerJobRun
		go func() {
			fmt.Println("Job triggered", string(job))
			fmt.Println("Current state", prettyPrint(fm.state))
		}()
	}
}
