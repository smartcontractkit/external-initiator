// Package subscriber holds logic to communicate between the
// external initiator service and the external endpoints it
// subscribes to.
package subscriber

import "github.com/smartcontractkit/external-initiator/store"

// Type holds the connection type for the subscription
type Type int

const (
	// WS are connections made over WebSocket
	WS Type = iota
	// RPC are connections made by POSTing a JSON payload
	// to the external endpoint.
	RPC
	// Client are connections encapsulated in its
	// entirety by the blockchain implementation.
	Client
	// Unknown is just a placeholder for when
	// it cannot be determined how connections
	// should be made. When this is returned,
	// it should be considered an error.
	Unknown
)

// SubConfig holds the configuration required to connect
// to the external endpoint.
type SubConfig struct {
	Endpoint string
}

// Event is the individual event that occurs during
// the subscription.
type Event []byte

// JsonManager holds the interface for generating blockchain
// specific payloads and parsing the response for the
// appropriate blockchain.
type JsonManager interface {
	// Get JSON payload to send when opening a new subscription
	GetTriggerJson() []byte
	// Parse the response returned after sending GetTriggerJson()
	ParseResponse(data []byte) ([]Event, bool)
	// Get JSON payload to send when testing a connection
	GetTestJson() []byte
	// Parse the response returned after sending GetTestJson()
	ParseTestResponse(data []byte) error
}

// ISubscription holds the interface for interacting
// with an active subscription.
type ISubscription interface {
	// Unsubscribe closes the connection to the external endpoint
	// and stops any processes related to this subscription.
	Unsubscribe()
}

// ISubscriber holds the interface for interacting
// with a not-yet-active subscription.
type ISubscriber interface {
	// SubscribeToEvents subscribes to events using the endpoint and configuration
	// as set in ISubscriber. All events will be sent in the channel. If anything is
	// passed as a param after channel, it will not expect to receive any confirmation
	// message after opening the initial subscription.
	SubscribeToEvents(channel chan<- Event, runtimeConfig store.RuntimeConfig) (ISubscription, error)
	// Test attempts to open a connection using the endpoint and configuration
	// as set in ISubscriber. If connection is succesful, it sends GetTestJson() as a payload
	// and attempts to parse response with ParseTestResponse().
	Test() error
}

// IParser holds the interface for parsing data
// from the external endpoint into an array of Events
// based on the blockchain's parser.
type IParser interface {
	ParseResponse(data []byte) ([]Event, bool)
}
