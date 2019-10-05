package subscriber

import (
	"testing"
	"time"
)

type Parser struct{}

func (parser Parser) ParseResponse(data []byte) ([]Event, bool) {
	return []Event{data}, true
}

func TestRpcSubscriber_SubscribeToEvents(t *testing.T) {
	t.Run("subscribes to rpc endpoint", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "/test/1"
		rpc := RpcSubscriber{Endpoint: u, Parser: Parser{}, Interval: 1 * time.Second}

		events := make(chan Event)
		filter := MockFilter{true}

		sub, err := rpc.SubscribeToEvents(events, filter)
		if err != nil {
			t.Errorf("SubscribeToEvents() error = %v", err)
			return
		}
		defer sub.Unsubscribe()

		event := <-events
		mockevent := string(event)
		if mockevent != "1" {
			t.Errorf("SubscribeToEvents() got unexpected first message = %v", mockevent)
			return
		}
		event = <-events
		mockevent = string(event)
		if mockevent != "2" {
			t.Errorf("SubscribeToEvents() got unexpected second message = %v", mockevent)
			return
		}
		return
	})
}

func TestSendGetRequest(t *testing.T) {
	t.Run("succeeds on normal response", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "/test/2"

		_, err := sendGetRequest(u.String())
		if err != nil {
			t.Errorf("sendGetRequest() got unexpected error = %v", err)
			return
		}
	})

	t.Run("fails on bad status", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "/fails"

		_, err := sendGetRequest(u.String())
		if err == nil {
			t.Error("sendGetRequest() expected error, but got nil")
			return
		}
	})
}
