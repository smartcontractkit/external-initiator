package subscriber

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getWS(u url.URL) (*jsonRpcWebsocketConnection, error) {
	endpoint := getEndpoint(u)
	return NewWebsocketConnection(endpoint)
}

func TestNewWebsocketConnection(t *testing.T) {
	t.Run("creates new WS connection", func(t *testing.T) {
		u := *wsMockUrl
		ws, err := getWS(u)
		if err != nil {
			t.Errorf("NewWebsocketConnection() error = %v", err)
			return
		}

		// checking return parameters
		assert.Equal(t, u.String(), ws.wsCore.endpoint)
		assert.Equal(t, WS, ws.Type())

		ws.Stop() // check for parameter change after Stop() call
		assert.Equal(t, true, ws.wsCore.stopped)
	})

	t.Run("fails creating connection", func(t *testing.T) {
		u, _ := url.Parse("ws://localhost:8080")

		_, err := getWS(*u)
		if err == nil {
			t.Errorf("NewWebsocketConnection() expected error, but go nil")
			return
		}
	})
}

func TestWSSubscribe(t *testing.T) {
	t.Run("successfully subscribes", func(t *testing.T) {
		u := *wsMockUrl
		ws, err := getWS(u)
		if err != nil {
			t.Errorf("TestSubscribe() error = %v", err)
			return
		}

		listener := make(chan json.RawMessage)
		err = ws.Subscribe(context.TODO(), "sub", testMethod+"_unsubscribe", emptyBytes, listener)
		if err != nil {
			t.Errorf("TestSubscribe() error = %v", err)
			return
		}

		data := <-listener
		t.Log(data)

	})
}
