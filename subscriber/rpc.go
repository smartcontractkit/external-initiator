package subscriber

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
)

// RpcSubscriber holds the configuration for
// a not-yet-active RPC subscription.
type RpcSubscriber struct {
	Endpoint string
	Interval time.Duration
	Manager  JsonManager
}

// Test sends a POST request using GetTestJson()
// as payload, and returns the error from
// calling ParseTestResponse() on the response.
func (rpc RpcSubscriber) Test() error {
	resp, err := sendPostRequest(rpc.Endpoint, rpc.Manager.GetTestJson())
	if err != nil {
		return err
	}

	return rpc.Manager.ParseTestResponse(resp)
}

// rpcSubscription holds an active RPC subscription.
type rpcSubscription struct {
	endpoint string
	done     chan struct{}
	events   chan<- Event
	manager  JsonManager
}

func (rpc rpcSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from RPC endpoint", rpc.endpoint)
	close(rpc.done)
}

func (rpc rpcSubscription) poll() {
	logger.Debugf("Polling %s\n", rpc.endpoint)

	resp, err := sendPostRequest(rpc.endpoint, rpc.manager.GetTriggerJson())
	if err != nil {
		logger.Errorf("Failed polling %s: %v\n", rpc.endpoint, err)
		return
	}

	events, ok := rpc.manager.ParseResponse(resp)
	if !ok {
		return
	}

	for _, event := range events {
		rpc.events <- event
	}
}

func (rpc rpcSubscription) readMessages(interval time.Duration) {
	timer := time.NewTicker(interval)
	defer timer.Stop()

	// Poll before waiting for ticker
	rpc.poll()

	for {
		select {
		case <-rpc.done:
			return
		case <-timer.C:
			rpc.poll()
		}
	}
}

func sendPostRequest(url string, body []byte) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode < 200 || r.StatusCode >= 400 {
		return nil, errors.New("got unexpected status code")
	}

	return ioutil.ReadAll(r.Body)
}

func (rpc RpcSubscriber) SubscribeToEvents(channel chan<- Event, _ store.RuntimeConfig) (ISubscription, error) {
	logger.Infof("Using RPC endpoint: %s\n", rpc.Endpoint)

	subscription := rpcSubscription{
		endpoint: rpc.Endpoint,
		done:     make(chan struct{}),
		events:   channel,
		manager:  rpc.Manager,
	}

	interval := rpc.Interval
	if interval <= time.Duration(0) {
		interval = 5 * time.Second
	}

	go subscription.readMessages(interval)

	return subscription, nil
}
