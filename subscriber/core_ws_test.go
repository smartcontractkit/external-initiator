package subscriber

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getWSCore(u url.URL) (*WebsocketConnection, error) {
	endpoint := getEndpoint(u)
	endpoint.RefreshInt = interval
	return NewCoreWebsocketConnection(endpoint)
}

func TestNewCoreWebsocketConnection(t *testing.T) {
	t.Run("creates new CoreWS connection", func(t *testing.T) {
		u := *wsMockUrl
		ws, err := getWSCore(u)
		if err != nil {
			t.Errorf("NewCoreWebsocketConnection() error = %v", err)
			return
		}

		// checking return parameters
		assert.Equal(t, u.String(), ws.endpoint)
		assert.Equal(t, CoreWS, ws.Type())

		ws.Stop() // check for parameter change after Stop() call
		assert.Equal(t, true, ws.stopped)
	})

	t.Run("fails creating connection", func(t *testing.T) {
		u, _ := url.Parse("ws://localhost:8080")

		_, err := getWSCore(*u)
		if err == nil {
			t.Errorf("NewCoreWebsocketConnection() expected error, but go nil")
			return
		}
	})
}

func TestSendMessage(t *testing.T) {
	t.Run("sends to websocket connection", func(t *testing.T) {
		u := *wsMockUrl
		ws, err := getWSCore(u)
		if err != nil {
			t.Errorf("getWSCore() error = %v", err)
			return
		}
		defer ws.Stop()

		err = ws.SendMessage([]byte("true"))
		if err != nil {
			t.Errorf("SendMessage() error = %v", err)
			return
		}
	})
}

func TestRead(t *testing.T) {
	t.Run("reads from websocket connection", func(t *testing.T) {
		u := *wsMockUrl
		ws, err := getWSCore(u)
		if err != nil {
			t.Errorf("getWSCore() error = %v", err)
			return
		}
		defer ws.Stop()

		listener := make(chan []byte)
		go func() {
			ws.Read(listener) //blocking call with for loop
		}()

		ws.SendMessage([]byte("true")) // trigger message

		res := <-listener //first response
		assert.Equal(t, "confirmation", string(res))

		res = <-listener //second response
		assert.Equal(t, "event", string(res))
	})
}
