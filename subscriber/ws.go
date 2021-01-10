package subscriber

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
)

// WebsocketSubscriber holds the configuration for
// a not-yet-active WS subscription.
type WebsocketSubscriber struct {
	Endpoint string
	Manager  JsonManager
}

// Test sends a opens a WS connection to the endpoint.
func (wss WebsocketSubscriber) Test() error {
	c, _, err := websocket.DefaultDialer.Dial(wss.Endpoint, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	testPayload := wss.Manager.GetTestJson()
	if testPayload == nil {
		return nil
	}

	resp := make(chan []byte)

	go func() {
		var body []byte
		_, body, err = c.ReadMessage()
		if err != nil {
			close(resp)
		}
		resp <- body
	}()

	err = c.WriteMessage(websocket.BinaryMessage, testPayload)
	if err != nil {
		return err
	}

	// Set timeout for response to 5 seconds
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()

	select {
	case <-t.C:
		return errors.New("timeout from test payload")
	case body, ok := <-resp:
		if !ok {
			return errors.New("failed reading test response from WS endpoint")
		}
		return wss.Manager.ParseTestResponse(body)
	}
}

type wsConn struct {
	connection *websocket.Conn
	closing    bool
}

type websocketSubscription struct {
	conn      *wsConn
	events    chan<- Event
	confirmed bool
	manager   JsonManager
	endpoint  string
}

func (wss websocketSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from WS endpoint", wss.endpoint)
	wss.conn.closing = true
	_ = wss.conn.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	_ = wss.conn.connection.Close()
}

func (wss websocketSubscription) forceClose() {
	wss.conn.closing = false
	_ = wss.conn.connection.Close()
}

func (wss websocketSubscription) readMessages() {
	for {
		_, message, err := wss.conn.connection.ReadMessage()
		if err != nil {
			_ = wss.conn.connection.Close()
			if !wss.conn.closing {
				wss.reconnect()
				return
			}
			return
		}

		// First message is a confirmation with the subscription id
		// Ignore this
		if !wss.confirmed {
			wss.confirmed = true
			continue
		}

		events, ok := wss.manager.ParseResponse(message)
		if !ok {
			continue
		}

		for _, event := range events {
			wss.events <- event
		}
	}
}

func (wss websocketSubscription) init() {
	go wss.readMessages()

	err := wss.conn.connection.WriteMessage(websocket.TextMessage, wss.manager.GetTriggerJson())
	if err != nil {
		wss.forceClose()
		return
	}

	logger.Infof("Connected to %s\n", wss.endpoint)
}

func (wss websocketSubscription) reconnect() {
	logger.Warnf("Lost WS connection to %s\nRetrying in %vs", wss.endpoint, 3)
	time.Sleep(3 * time.Second)

	c, _, err := websocket.DefaultDialer.Dial(wss.endpoint, nil)
	if err != nil {
		logger.Error("Reconnect failed:", err)
		wss.reconnect()
		return
	}

	wss.conn.connection = c
	wss.init()
}

func (wss WebsocketSubscriber) SubscribeToEvents(channel chan<- Event, _ store.RuntimeConfig) (ISubscription, error) {
	logger.Infof("Connecting to WS endpoint: %s\n", wss.Endpoint)

	c, _, err := websocket.DefaultDialer.Dial(wss.Endpoint, nil)
	if err != nil {
		return nil, err
	}

	subscription := websocketSubscription{
		conn:      &wsConn{connection: c},
		events:    channel,
		confirmed: false,
		manager:   wss.Manager,
		endpoint:  wss.Endpoint,
	}
	subscription.init()

	return subscription, nil
}
