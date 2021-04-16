package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	maxMessageSize = 15 * 1024 * 1024
)

var (
	errorRequestTimeout = errors.New("request timed out")
)

type websocketConnection struct {
	endpoint string

	requests              []*subscribeRequest
	subscriptionListeners map[string]chan<- json.RawMessage
	nonceListeners        map[uint64]chan<- json.RawMessage

	conn *websocket.Conn

	quitOnce sync.Once

	writeMutex sync.Mutex
	nonce      atomic.Uint64

	chSubscriptionIds chan string
	chClose           chan struct{}
	stopped           bool
}

func NewWebsocketConnection(endpoint string) (*websocketConnection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, err
	}

	wsc := &websocketConnection{
		endpoint:              endpoint,
		conn:                  conn,
		subscriptionListeners: make(map[string]chan<- json.RawMessage),
		nonceListeners:        make(map[uint64]chan<- json.RawMessage),
		chSubscriptionIds:     make(chan string),
		chClose:               make(chan struct{}),
	}

	go wsc.read()

	return wsc, nil
}

func (wsc *websocketConnection) Type() Type {
	return WS
}

func (wsc *websocketConnection) Stop() {
	wsc.quitOnce.Do(func() {
		wsc.stopped = true
		close(wsc.chClose)
	})
}

func (wsc *websocketConnection) Subscribe(ctx context.Context, method, unsubscribeMethod string, params json.RawMessage, ch chan<- json.RawMessage) error {
	req := wsc.newSubscribeRequest(ctx, method, unsubscribeMethod, params, ch)
	err := wsc.subscribe(req)
	if err != nil {
		return err
	}

	return nil
}

func (wsc *websocketConnection) Request(ctx context.Context, method string, params json.RawMessage) (result json.RawMessage, err error) {
	listener := make(chan json.RawMessage, 1)
	nonce := wsc.nonce.Inc()
	wsc.nonceListeners[nonce] = listener
	defer func() {
		delete(wsc.nonceListeners, nonce)
		close(listener)
	}()

	payload, err := NewJsonrpcMessage(nonce, method, params)
	if err != nil {
		return nil, err
	}

	err = wsc.sendMessage(payload)
	if err != nil {
		return nil, err
	}

	select {
	case msg := <-listener:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (wsc *websocketConnection) resetConnection() {
	if wsc.stopped {
		return
	}

	wsc.subscriptionListeners = make(map[string]chan<- json.RawMessage)
	wsc.nonceListeners = make(map[uint64]chan<- json.RawMessage)
	wsc.nonce.Store(0)

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

	for _, request := range wsc.requests {
		if request == nil || request.stopped {
			continue
		}
		logger.ErrorIf(wsc.subscribe(request))
	}
}

func (wsc *websocketConnection) read() {
	defer wsc.resetConnection()

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
			return
		}
		ch <- msg.Result
		return
	}

	var params struct {
		Subscription string          `json:"subscription"`
		Result       json.RawMessage `json:"result"`
	}
	err = json.Unmarshal(msg.Params, &params)
	if err != nil {
		return
	}

	ch, ok := wsc.subscriptionListeners[params.Subscription]
	if !ok {
		// TODO: Should be improved in a way
		time.Sleep(1 * time.Second)
		ch, ok = wsc.subscriptionListeners[params.Subscription]
		if !ok {
			return
		}
	}

	ch <- params.Result
}

func (wsc *websocketConnection) subscribe(req *subscribeRequest) error {
	subscriptionId, err := wsc.getSubscriptionId(req)
	if err != nil {
		return err
	}

	listener := make(chan json.RawMessage)
	wsc.subscriptionListeners[subscriptionId] = listener

	go func() {
		defer func() {
			delete(wsc.subscriptionListeners, subscriptionId)
			close(listener)
		}()

		for {
			select {
			case msg := <-listener:
				req.ch <- msg
			case <-req.ctx.Done():
				req.stopped = true
				payload, err := NewJsonrpcMessage(wsc.nonce.Inc(), req.unsubscribeMethod, []byte(fmt.Sprintf(`["%s"]`, subscriptionId)))
				if err != nil {
					logger.Error(err)
					return
				}
				logger.ErrorIf(wsc.sendMessage(payload))
				return
			case <-wsc.chClose:
				return
			}
		}
	}()

	return nil
}

func (wsc *websocketConnection) getSubscriptionId(req *subscribeRequest) (string, error) {
	nonce := wsc.nonce.Inc()
	payload, err := NewJsonrpcMessage(nonce, req.method, req.params)
	if err != nil {
		return "", err
	}

	listener := make(chan json.RawMessage)
	wsc.nonceListeners[nonce] = listener
	defer func() {
		delete(wsc.nonceListeners, nonce)
		close(listener)
	}()

	err = wsc.sendMessage(payload)
	if err != nil {
		return "", err
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case result := <-listener:
		var subscriptionId string
		err = json.Unmarshal(result, &subscriptionId)
		return subscriptionId, err
	case <-timer.C:
		return "", errorRequestTimeout
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

type subscribeRequest struct {
	ctx    context.Context
	method string
	params json.RawMessage
	ch     chan<- json.RawMessage

	unsubscribeMethod string
	stopped           bool
}

func (wsc *websocketConnection) newSubscribeRequest(ctx context.Context, method, unsubscribeMethod string, params json.RawMessage, ch chan<- json.RawMessage) *subscribeRequest {
	req := &subscribeRequest{
		ctx:               ctx,
		method:            method,
		params:            params,
		ch:                ch,
		unsubscribeMethod: unsubscribeMethod,
	}
	wsc.requests = append(wsc.requests, req)
	return req
}
