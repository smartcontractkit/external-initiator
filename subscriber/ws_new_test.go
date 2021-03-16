package subscriber

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/external-initiator/store"
)

var upgrader = websocket.Upgrader{} // use default options

func TestWebsocketSubscriber_SubscribeToEvents(t *testing.T) {
	t.Run("subscribes and ignores confirmation message", func(t *testing.T) {
		wss := WebsocketSubscriber{Endpoint: wsMockUrl.String(), Manager: TestsMockManager{true}}
		events := make(chan Event)

		sub, err := wss.SubscribeToEvents(events, store.RuntimeConfig{})
		if err != nil {
			t.Errorf("SubscribeToEvents() error = %v", err)
			return
		}
		defer sub.Unsubscribe()

		event := <-events
		mockevent := string(event)

		if mockevent == "confirmation" {
			t.Error("SubscribeToEvents() got unexpected confirmation")
			return
		}

		if mockevent != "event" {
			t.Errorf("SubscribeToEvents() got unexpected message = %v", mockevent)
			return
		}
	})

	t.Run("fails subscribe to invalid URL", func(t *testing.T) {
		wss := WebsocketSubscriber{Endpoint: "", Manager: TestsMockManager{false}}
		events := make(chan Event)

		sub, err := wss.SubscribeToEvents(events, store.RuntimeConfig{})
		if err == nil {
			sub.Unsubscribe()
			t.Error("SubscribeToEvents() expected error, but got nil")
			return
		}
	})

	t.Run("subscribes and attempts reconnect", func(t *testing.T) {
		wss := WebsocketSubscriber{Endpoint: wsMockUrl.String(), Manager: &TestsReconnectManager{}}
		events := make(chan Event)

		sub, err := wss.SubscribeToEvents(events, store.RuntimeConfig{})
		if err != nil {
			t.Errorf("SubscribeToEvents() error = %v", err)
			return
		}
		defer sub.Unsubscribe()

		event := <-events
		mockevent := string(event)

		if mockevent != "event" {
			t.Errorf("SubscribeToEvents() got unexpected message = %v", mockevent)
			return
		}
	})
}

type TestsReconnectManager struct {
	connections int
}

func (m TestsReconnectManager) ParseResponse(data []byte) ([]Event, bool) {
	return []Event{data}, true
}

func (m *TestsReconnectManager) GetTriggerJson() []byte {
	count := m.connections
	m.connections++
	switch count {
	case 0:
		return []byte(`close`)
	default:
		return []byte(`true`)
	}
}

func (m TestsReconnectManager) GetTestJson() []byte {
	return nil
}

func (m TestsReconnectManager) ParseTestResponse(data []byte) error {
	return nil
}

func TestWebsocketSubscriber_Test(t *testing.T) {
	type fields struct {
		Endpoint string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"succeeds connecting to valid endpoint",
			fields{Endpoint: wsMockUrl.String()},
			false,
		},
		{
			"fails connecting to invalid endpoint",
			fields{Endpoint: "ws://localhost:9999/invalid"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wss := WebsocketSubscriber{
				Endpoint: tt.fields.Endpoint,
				Manager:  TestsMockManager{},
			}
			if err := wss.Test(); (err != nil) != tt.wantErr {
				t.Errorf("Test() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
