package subscriber

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type RpcSubscriber struct {
	Endpoint string
	Interval time.Duration
	Manager  Manager
}

func (rpc RpcSubscriber) Test() error {
	resp, err := sendPostRequest(rpc.Endpoint, rpc.Manager.GetTestJson())
	if err != nil {
		return err
	}

	return rpc.Manager.ParseTestResponse(resp)
}

type RpcSubscription struct {
	endpoint string
	done     chan struct{}
	events   chan<- Event
	manager  Manager
}

func (rpc RpcSubscription) Unsubscribe() {
	close(rpc.done)
}

func (rpc RpcSubscription) readMessages(interval time.Duration) {
	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case <-rpc.done:
			return
		case <-timer.C:
			fmt.Printf("Polling %s\n", rpc.endpoint)
			resp, err := sendPostRequest(rpc.endpoint, rpc.manager.GetTriggerJson())
			if err != nil {
				fmt.Printf("Failed polling %s: %v\n", rpc.endpoint, err)
				continue
			}

			events, ok := rpc.manager.ParseResponse(resp)
			if !ok {
				continue
			}

			for _, event := range events {
				rpc.events <- event
			}
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

func (rpc RpcSubscriber) SubscribeToEvents(channel chan<- Event, confirmation ...interface{}) (ISubscription, error) {
	fmt.Printf("Using RPC endpoint: %s", rpc.Endpoint)

	subscription := RpcSubscription{
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
