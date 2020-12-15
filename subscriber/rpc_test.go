package subscriber

import (
	"testing"
	"time"

	"github.com/smartcontractkit/external-initiator/store"
)

func TestRpcSubscriber_SubscribeToEvents(t *testing.T) {
	t.Run("subscribes to rpc endpoint", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "/test/1"
		rpc := RpcSubscriber{Endpoint: u.String(), Manager: TestsMockManager{true}, Interval: 1 * time.Second}

		events := make(chan Event)

		sub, err := rpc.SubscribeToEvents(events, store.RuntimeConfig{})
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

func TestSendPostRequest(t *testing.T) {
	t.Run("succeeds on normal response", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "/test/2"

		_, err := sendPostRequest(u.String(), TestsMockManager{}.GetTriggerJson())
		if err != nil {
			t.Errorf("sendGetRequest() got unexpected error = %v", err)
			return
		}
	})

	t.Run("fails on bad status", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "/fails"

		_, err := sendPostRequest(u.String(), TestsMockManager{}.GetTriggerJson())
		if err == nil {
			t.Error("sendGetRequest() expected error, but got nil")
			return
		}
	})
}

func TestRpcSubscriber_Test(t *testing.T) {
	type fields struct {
		Endpoint string
		Manager  JsonManager
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"succeeds connecting to valid endpoint",
			fields{Endpoint: rpcMockUrl.String(), Manager: TestsMockManager{}},
			false,
		},
		{
			"fails connecting to invalid endpoint",
			fields{Endpoint: "http://localhost:9999/invalid", Manager: TestsMockManager{}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpc := RpcSubscriber{
				Endpoint: tt.fields.Endpoint,
				Manager:  tt.fields.Manager,
			}
			if err := rpc.Test(); (err != nil) != tt.wantErr {
				t.Errorf("Test() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
