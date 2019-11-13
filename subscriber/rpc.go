package subscriber

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type RpcSubscriber struct {
	Endpoint string
	Interval time.Duration
	Parser   IParser
}

func (rpc RpcSubscriber) Test() error {
	_, err := sendGetRequest(rpc.Endpoint)
	return err
}

type RpcSubscription struct {
	endpoint  string
	done      chan struct{}
	events    chan<- Event
	confirmed bool
	parser    IParser
}

func (rpc RpcSubscription) Unsubscribe() {
	close(rpc.done)
}

func (rpc RpcSubscription) readMessages(interval time.Duration, filter Filter) {
	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case <-rpc.done:
			return
		case <-timer.C:
			fmt.Printf("Polling %s\n", rpc.endpoint)
			resp, err := sendGetRequest(rpc.endpoint)
			if err != nil {
				fmt.Printf("Failed polling %s: %v\n", rpc.endpoint, err)
				continue
			}

			events, ok := rpc.parser.ParseResponse(resp)
			if !ok {
				continue
			}

			for _, event := range events {
				rpc.events <- event
			}
		}
	}
}

func sendGetRequest(url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
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

func (rpc RpcSubscriber) SubscribeToEvents(channel chan<- Event, filter Filter, confirmation ...interface{}) (ISubscription, error) {
	fmt.Printf("Using RPC endpoint: %s", rpc.Endpoint)

	subscription := RpcSubscription{
		endpoint:  rpc.Endpoint,
		done:      make(chan struct{}),
		events:    channel,
		confirmed: len(confirmation) != 0, // If passed as a param, do not expect confirmation message
		parser:    rpc.Parser,
	}

	interval := rpc.Interval
	if interval <= time.Duration(0) {
		interval = 5 * time.Second
	}

	go subscription.readMessages(interval, filter)

	return subscription, nil
}
