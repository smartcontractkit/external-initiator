package subscriber

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var wsMockUrl *url.URL

var upgrader = websocket.Upgrader{} // use default options

func TestMain(m *testing.M) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)

			if string(message) == "true" {
				// Send confirmation message
				err = c.WriteMessage(mt, []byte("confirmation"))
				if err != nil {
					log.Println("write:", err)
					break
				}
			}

			// Send event message
			err = c.WriteMessage(mt, []byte("event"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	wsMockUrl = u
	if err != nil {
		log.Fatal(err)
	}
	u.Scheme = "ws"

	code := m.Run()
	os.Exit(code)
}

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
