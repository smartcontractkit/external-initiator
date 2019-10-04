package subscriber

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type RpcSubscriber struct {
	Endpoint url.URL
	Interval time.Duration
	Parser   IParser
}

type RpcSubscription struct {
	endpoint  url.URL
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
			fmt.Printf("Polling %s\n", rpc.endpoint.String())
			resp, err := sendGetRequest(rpc.endpoint.String())
			if err != nil {
				fmt.Printf("Failed polling %s: %v\n", rpc.endpoint.String(), err)
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

	if r.StatusCode < 200 || r.StatusCode > 299 {
		return nil, errors.New("got unexpected status code")
	}

	return ioutil.ReadAll(r.Body)
}

func (rpc RpcSubscriber) SubscribeToEvents(channel chan<- Event, filter Filter, confirmation ...interface{}) (ISubscription, error) {
	fmt.Printf("Using RPC endpoint: %s", rpc.Endpoint.String())

	subscription := RpcSubscription{
		endpoint:  rpc.Endpoint,
		done:      make(chan struct{}),
		events:    channel,
		confirmed: len(confirmation) != 0, // If passed as a param, do not expect confirmation message
		parser:    rpc.Parser,
	}

	go subscription.readMessages(rpc.Interval, filter)

	return subscription, nil
}
