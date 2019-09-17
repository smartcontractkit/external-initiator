package subscriber

import (
	"fmt"
	"log"
	"net/url"
	"time"
)

type RpcSubscriber struct {
	Endpoint url.URL
	Interval time.Duration
}

type RpcSubscription struct {
	endpoint  url.URL
	done      chan struct{}
	events    chan<- Event
	confirmed bool
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
			fmt.Printf("Requesting data from %s\n", rpc.endpoint.String())
			// TODO: Send request
			// TODO: Parse response
		}
	}
}

func (rpc RpcSubscriber) SubscribeToEvents(channel chan<- Event, filter Filter, confirmation ...interface{}) (ISubscription, error) {
	log.Printf("Using RPC endpoint: %s", rpc.Endpoint.String())

	subscription := RpcSubscription{
		endpoint:  rpc.Endpoint,
		done:      make(chan struct{}),
		events:    channel,
		confirmed: len(confirmation) != 0, // If passed as a param, do not expect confirmation message
	}

	go subscription.readMessages(rpc.Interval, filter)

	return subscription, nil
}
