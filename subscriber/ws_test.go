package subscriber

import (
	"github.com/gorilla/websocket"
	"testing"
)

var upgrader = websocket.Upgrader{} // use default options

type TestsMockFilter struct {
	confirmation bool
}

func (mf TestsMockFilter) Json() []byte {
	if mf.confirmation {
		return []byte(`true`)
	}
	return []byte(`false`)
}

func TestWebsocketSubscriber_SubscribeToEvents(t *testing.T) {
	wss := WebsocketSubscriber{Endpoint: wsMockUrl.String(), Parser: MockParser{}}

	t.Run("subscribes and ignores confirmation message", func(t *testing.T) {
		events := make(chan Event)
		filter := TestsMockFilter{true}

		sub, err := wss.SubscribeToEvents(events, filter)
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

	t.Run("subscribes and does not expect confirmation message", func(t *testing.T) {
		events := make(chan Event)
		filter := TestsMockFilter{false}

		sub, err := wss.SubscribeToEvents(events, filter, false)
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

	t.Run("fails subscribe to invalid URL", func(t *testing.T) {
		events := make(chan Event)
		filter := TestsMockFilter{false}

		nonExistantWss := WebsocketSubscriber{Endpoint: ""}

		sub, err := nonExistantWss.SubscribeToEvents(events, filter)
		if err == nil {
			sub.Unsubscribe()
			t.Error("SubscribeToEvents() expected error, but got nil")
			return
		}
	})
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
			}
			if err := wss.Test(); (err != nil) != tt.wantErr {
				t.Errorf("Test() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
