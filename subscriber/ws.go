package subscriber

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type WebsocketSubscriber struct {
	Endpoint string
	Manager  Manager
}

func (wss WebsocketSubscriber) Test() error {
	c, _, err := websocket.DefaultDialer.Dial(wss.Endpoint, nil)
	if err != nil {
		return err
	}
	c.Close()
	return nil
}

type wsConn struct {
	connection *websocket.Conn
	closing    bool
}

type WebsocketSubscription struct {
	conn      *wsConn
	done      chan bool
	events    chan<- Event
	confirmed bool
	manager   Manager
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
