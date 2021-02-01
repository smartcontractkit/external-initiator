package subscriber

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/external-initiator/eitest"
)

var rpcMockUrl *url.URL
var wsMockUrl *url.URL

type TestsMockManager struct {
	confirmation bool
}

func (m TestsMockManager) ParseResponse(data []byte) ([]Event, bool) {
	return []Event{data}, true
}

func (m TestsMockManager) GetTriggerJson() []byte {
	if m.confirmation {
		return []byte(`true`)
	}
	return []byte(`false`)
}

func (m TestsMockManager) GetTestJson() []byte {
	return nil
}

func (m TestsMockManager) ParseTestResponse(data []byte) error {
	return nil
}

func TestMain(m *testing.M) {
	responses := make(map[string]int)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fails" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		responses[r.URL.Path] = responses[r.URL.Path] + 1
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprint(responses[r.URL.Path])))
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	rpcMockUrl = u
	if err != nil {
		log.Fatal(err)
	}

	ws := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var c *websocket.Conn
		c, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer eitest.MustClose(c)
		for {
			var mt int
			var message []byte
			mt, message, err = c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)

			switch string(message) {
			case "true":
				// Send confirmation message
				err = c.WriteMessage(mt, []byte("confirmation"))
				if err != nil {
					log.Println("write:", err)
					return
				}
			case "close":
				// Close connection prematurely
				return
			}

			// Send event message
			err = c.WriteMessage(mt, []byte("event"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}))
	defer ws.Close()

	wsMockUrl, err = url.Parse(ws.URL)
	if err != nil {
		log.Fatal(err)
	}
	wsMockUrl.Scheme = "ws"

	code := m.Run()
	os.Exit(code)
}
