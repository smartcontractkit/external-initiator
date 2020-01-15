package subscriber

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
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
		_, body, err := c.ReadMessage()
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

type WebsocketSubscription struct {
	conn      *wsConn
	events    chan<- Event
	confirmed bool
	manager   JsonManager
	endpoint  string
}

func (wss WebsocketSubscription) Unsubscribe() {
	fmt.Println("Unsubscribing from WS endpoint", wss.endpoint)
	wss.conn.closing = true
	_ = wss.conn.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	_ = wss.conn.connection.Close()
}

func (wss WebsocketSubscription) forceClose() {
	wss.conn.closing = false
	_ = wss.conn.connection.Close()
}

func (wss WebsocketSubscription) readMessages() {
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

func (wss WebsocketSubscription) init() {
	go wss.readMessages()

	err := wss.conn.connection.WriteMessage(websocket.TextMessage, wss.manager.GetTriggerJson())
	if err != nil {
		wss.forceClose()
		return
	}

	fmt.Printf("Connected to %s\n", wss.endpoint)
}

func (wss WebsocketSubscription) reconnect() {
	fmt.Printf("Lost WS connection to %s\nRetrying in %vs\n", wss.endpoint, 3)
	time.Sleep(3 * time.Second)

	c, _, err := websocket.DefaultDialer.Dial(wss.endpoint, nil)
	if err != nil {
		fmt.Println("Reconnect failed:", err)
		wss.reconnect()
		return
	}

	wss.conn.connection = c
	wss.init()
}

func (wss WebsocketSubscriber) SubscribeToEvents(channel chan<- Event, confirmation ...interface{}) (ISubscription, error) {
	fmt.Printf("Connecting to WS endpoint: %s\n", wss.Endpoint)

	c, _, err := websocket.DefaultDialer.Dial(wss.Endpoint, nil)
	if err != nil {
		return nil, err
	}

	subscription := WebsocketSubscription{
		conn:      &wsConn{connection: c},
		events:    channel,
		confirmed: len(confirmation) != 0, // If passed as a param, do not expect confirmation message
		manager:   wss.Manager,
		endpoint:  wss.Endpoint,
	}
	subscription.init()

	return subscription, nil
}
