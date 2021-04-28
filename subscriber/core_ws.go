package subscriber

import (
	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 15 * 1024 * 1024
)

type WebsocketConnection struct {
	endpoint string

	conn *websocket.Conn

	quitOnce   sync.Once
	writeMutex sync.Mutex

	chClose chan struct{}
	stopped bool
}

func NewCoreWebsocketConnection(endpoint store.Endpoint) (*WebsocketConnection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(endpoint.Url, nil)
	if err != nil {
		return nil, err
	}

	wsc := &WebsocketConnection{
		endpoint: endpoint.Url,
		conn:     conn,
		chClose:  make(chan struct{}),
	}

	return wsc, nil
}

func (wsc *WebsocketConnection) Type() Type {
	return CoreWS
}

func (wsc *WebsocketConnection) Stop() {
	wsc.quitOnce.Do(func() {
		wsc.stopped = true
		close(wsc.chClose)
	})
}

func (wsc *WebsocketConnection) resetConnection() {
	if wsc.stopped {
		return
	}

	attempts := 0
	for {
		if wsc.stopped {
			return
		}

		attempts++

		conn, _, err := websocket.DefaultDialer.Dial(wsc.endpoint, nil)
		if err != nil {
			logger.Error(err)
			var fac time.Duration
			if attempts < 5 {
				fac = time.Duration(attempts * 2)
			} else {
				fac = 10
			}
			time.Sleep(fac * time.Second)
			continue
		}

		wsc.conn = conn
		break
	}
}

func (wsc *WebsocketConnection) Read(ch chan<- []byte) {
	defer wsc.resetConnection()

	wsc.conn.SetReadLimit(maxMessageSize)
	for {
		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			// TODO: Reconnect
			return
		}

		ch <- message
	}
}

func (wsc *WebsocketConnection) SendMessage(payload []byte) error {
	wsc.writeMutex.Lock()
	defer wsc.writeMutex.Unlock()

	err := wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		return err
	}
	return wsc.conn.WriteMessage(websocket.TextMessage, payload)
}
