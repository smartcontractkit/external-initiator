package subscriber

import (
	"github.com/gorilla/websocket"
	"net/url"
	"testing"
)

var upgrader = websocket.Upgrader{} // use default options

type MockFilter struct {
	confirmation bool
}

func (mf MockFilter) Json() []byte {
	if mf.confirmation {
		return []byte(`true`)
	}
	return []byte(`false`)
}

func TestWebsocketSubscriber_SubscribeToEvents(t *testing.T) {
	wss := WebsocketSubscriber{Endpoint: *wsMockUrl}

	t.Run("subscribes and ignores confirmation message", func(t *testing.T) {
		events := make(chan Event)
		filter := MockFilter{true}

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
		filter := MockFilter{false}

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
		filter := MockFilter{false}

		nonExistantWss := WebsocketSubscriber{Endpoint: url.URL{}}

		sub, err := nonExistantWss.SubscribeToEvents(events, filter)
		if err == nil {
			sub.Unsubscribe()
			t.Error("SubscribeToEvents() expected error, but got nil")
			return
		}
	})
}
