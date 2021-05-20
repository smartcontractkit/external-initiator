package subscriber

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/external-initiator/eitest"
	"github.com/smartcontractkit/external-initiator/store"
)

var rpcMockUrl *url.URL
var wsMockUrl *url.URL
var upgrader = websocket.Upgrader{} // use default options

func getEndpoint(u url.URL) store.Endpoint {
	return store.Endpoint{Url: u.String()}
}

func wrapJsonRPC(s interface{}, id json.RawMessage, wrap bool) []byte {
	// handle RPC + WS cases
	if wrap {
		sBytes, _ := json.Marshal(s)
		msg, _ := json.Marshal(JsonrpcMessage{Result: sBytes, ID: id})
		return msg
	}

	// handle WS Core case
	if str, ok := s.(string); ok {
		return []byte(str)
	}

	log.Fatal("wrapJsonRPC: invalid type")
	return []byte{}
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
		_, _ = w.Write(wrapJsonRPC(responses[r.URL.Path], nil, true))
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

			// check if json message (will wrap response if necessary)
			var jsonMsg JsonrpcMessage
			var jsonCheck bool
			var id json.RawMessage
			err = json.Unmarshal(message, &jsonMsg)
			if err == nil {
				message = []byte(jsonMsg.Method)
				jsonCheck = true
				id = jsonMsg.ID
			}

			switch string(message) {
			case "true":
				// Send confirmation message
				err = c.WriteMessage(mt, wrapJsonRPC("confirmation", id, jsonCheck))
				if err != nil {
					log.Println("write:", err)
					return
				}
			case "close":
				// Close connection prematurely
				return
			}

			// Send event message
			err = c.WriteMessage(mt, wrapJsonRPC("event", id, jsonCheck))
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

func TestNewSubscriber(t *testing.T) {
	type fields struct {
		Url string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"succeeds with valid RPC endpoint",
			fields{Url: rpcMockUrl.String()},
			false,
		},
		{
			"succeeds with valid WS endpoint",
			fields{Url: wsMockUrl.String()},
			false,
		},
		{
			"fails with invalid endpoint",
			fields{Url: "not_real"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := store.Endpoint{Url: tt.fields.Url, RefreshInt: 10}
			if _, err := NewSubscriber(endpoint); (err != nil) != tt.wantErr {
				t.Errorf("NewSubscriber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewJsonrpcMessage(t *testing.T) {
	type fields struct {
		nonce  uint64
		method string
		params json.RawMessage
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"succeeds with valid parameters",
			fields{nonce: 1, method: "test", params: []byte("{}")},
			false,
		},
		{
			"fails with invalid params",
			fields{nonce: 1, method: "test", params: []byte("fail")},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewJsonrpcMessage(tt.fields.nonce, tt.fields.method, tt.fields.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJsonrpcMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
