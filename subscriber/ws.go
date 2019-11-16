package subscriber

import (
	"fmt"
	"github.com/gorilla/websocket"
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

type WebsocketSubscription struct {
	connection *websocket.Conn
	done       chan bool
	events     chan<- Event
	confirmed  bool
	manager    Manager
}

func (wss WebsocketSubscription) Unsubscribe() {
	_ = wss.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	_ = wss.connection.Close()
}

func (wss WebsocketSubscription) readMessages() {
	for {
		_, message, err := wss.connection.ReadMessage()
		if err != nil {
			_ = wss.connection.Close()
			// TODO: Attempt reconnect
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

func (wss WebsocketSubscriber) SubscribeToEvents(channel chan<- Event, confirmation ...interface{}) (ISubscription, error) {
	fmt.Printf("Connecting to WS endpoint: %s\n", wss.Endpoint)

	c, _, err := websocket.DefaultDialer.Dial(wss.Endpoint, nil)
	if err != nil {
		return nil, err
	}

	subscription := WebsocketSubscription{
		connection: c,
		events:     channel,
		confirmed:  len(confirmation) != 0, // If passed as a param, do not expect confirmation message
		manager:    wss.Manager,
	}

	go subscription.readMessages()

	err = subscription.connection.WriteMessage(websocket.TextMessage, wss.Manager.GetTriggerJson())
	if err != nil {
		subscription.Unsubscribe()
		return nil, err
	}

	fmt.Printf("Connected to %s\n", wss.Endpoint)

	return subscription, nil
}
