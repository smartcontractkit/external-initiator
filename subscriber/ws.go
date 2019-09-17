package subscriber

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type WebsocketSubscriber struct {
	Endpoint url.URL
}

type WebsocketSubscription struct {
	connection *websocket.Conn
	done       chan bool
	events     chan<- Event
	confirmed  bool
}

func (wss WebsocketSubscription) Unsubscribe() {
	_ = wss.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	_ = wss.connection.Close()
}

func (wss WebsocketSubscription) readMessages() {
	for {
		_, message, err := wss.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			_ = wss.connection.Close()
			close(wss.events)
			return
		}

		var event Event
		err = json.Unmarshal(message, &event)
		if err != nil {
			log.Println("Unable to unmarshal message: ", err)
			continue
		}

		// First message is a confirmation with the subscription id
		// Ignore this
		if !wss.confirmed {
			wss.confirmed = true
			continue
		}

		wss.events <- event
	}
}

func (wss WebsocketSubscriber) SubscribeToEvents(channel chan<- Event, filter Filter, confirmation ...interface{}) (ISubscription, error) {
	fmt.Printf("Connecting to WS endpoint: %s\n", wss.Endpoint.String())

	c, _, err := websocket.DefaultDialer.Dial(wss.Endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	subscription := WebsocketSubscription{
		connection: c,
		events:     channel,
		confirmed:  len(confirmation) != 0, // If passed as a param, do not expect confirmation message
	}

	go subscription.readMessages()

	err = subscription.connection.WriteMessage(websocket.TextMessage, filter.Json())
	if err != nil {
		subscription.Unsubscribe()
		return nil, err
	}

	fmt.Printf("Connected to %s\n", wss.Endpoint.String())

	return subscription, nil
}
