package subscriber

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/chainlink/core/logger"
	"go.uber.org/atomic"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	errorRequestTimeout = errors.New("request timed out")
)

type subscribeRequest struct {
	method string
	params json.RawMessage
	ch     chan<- json.RawMessage
}

type websocketConnection struct {
	requests              []subscribeRequest
	subscriptionListeners map[string]chan<- json.RawMessage
	nonceListeners        map[uint64]chan<- json.RawMessage

	conn *websocket.Conn

	quitOnce sync.Once

	writeMutex sync.Mutex
	nonce      atomic.Uint64

	chSubscriptionIds chan string
	chClose           chan struct{}
}

func newWebsocketConnection(endpoint string) (*websocketConnection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, err
	}

	wsc := &websocketConnection{
		conn:              conn,
		chSubscriptionIds: make(chan string),
		chClose:           make(chan struct{}),
	}

	go wsc.read()

	return wsc, nil
}

func (wsc *websocketConnection) read() {
	wsc.conn.SetReadLimit(maxMessageSize)
	for {
		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			// TODO: Reconnect
			return
		}

		go wsc.processIncomingMessage(message)
	}
}

func (wsc *websocketConnection) processIncomingMessage(payload json.RawMessage) {
	var msg JsonrpcMessage
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		logger.Errorf("Unable to unmarshal payload: %s", payload)
		return
	}

	var nonce uint64
	err = json.Unmarshal(msg.ID, &nonce)
	if err == nil && nonce > 0 {
		ch, ok := wsc.nonceListeners[nonce]
		if !ok {
			logger.Errorf("Could not find listener for nonce: %v", nonce)
			return
		}
		ch <- msg.Result
	}

	var params struct {
		Subscription string `json:"subscription"`
	}
	err = json.Unmarshal(msg.Params, &params)
	if err != nil {
		logger.Errorf("Unable to find subscription ID in message: %s", payload)
		return
	}

	ch, ok := wsc.subscriptionListeners[params.Subscription]
	if !ok {
		logger.Errorf("Could not find listener for subscription: %s", params.Subscription)
		return
	}

	ch <- msg.Params
}

func (wsc *websocketConnection) subscribe(method string, params json.RawMessage, ch chan<- json.RawMessage) error {
	wsc.requests = append(wsc.requests, subscribeRequest{method, params, ch})

	nonce := wsc.nonce.Inc()
	payload, err := NewJsonrpcMessage(nonce, method, params)
	if err != nil {
		return err
	}

	listener := make(chan json.RawMessage)
	wsc.nonceListeners[nonce] = listener

	err = wsc.sendMessage(payload)
	if err != nil {
		return err
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case result := <-listener:
		var subscriptionId string
		err = json.Unmarshal(result, &subscriptionId)
		if err != nil {
			return err
		}

		wsc.subscriptionListeners[subscriptionId] = ch
		return nil
	case <-timer.C:
		return errorRequestTimeout
	}
}

func (wsc *websocketConnection) sendMessage(payload json.RawMessage) error {
	wsc.writeMutex.Lock()
	defer wsc.writeMutex.Unlock()

	err := wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		return err
	}
	return wsc.conn.WriteMessage(websocket.TextMessage, payload)
}

func (wsc *websocketConnection) close() {
	wsc.quitOnce.Do(func() {
		close(wsc.chClose)
	})
}

// WebsocketSubscriber holds the configuration for
// a not-yet-active WS subscription.
type WebsocketSubscriberNew struct {
	endpoint string

	wsc *websocketConnection

	onClose sync.Once
}

func NewWebsocketSubscriber(endpoint string) (*WebsocketSubscriberNew, error) {
	wsc, err := newWebsocketConnection(endpoint)
	if err != nil {
		return nil, err
	}

	return &WebsocketSubscriberNew{
		endpoint: endpoint,
		wsc:      wsc,
	}, nil
}

func (wss *WebsocketSubscriberNew) Subscribe(method string, params json.RawMessage, ch chan<- []byte) (func(), error) {
	msgs := make(chan json.RawMessage)
	err := wss.wsc.subscribe(method, params, msgs)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case msg := <-msgs:
				ch <- msg
			case <-wss.wsc.chClose:
				return
			}
		}
	}()

	unsubscribe := func() {
		// TODO: Implement Unsubscribe
	}

	return unsubscribe, nil
}

func (wss *WebsocketSubscriberNew) Request(method string, params json.RawMessage) (result []byte, err error) {
	// TODO: Implement
	return nil, err
}

func (wss *WebsocketSubscriberNew) Stop() {
	wss.onClose.Do(func() {
		wss.wsc.close()
	})
}
