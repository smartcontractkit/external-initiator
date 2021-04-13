// Package subscriber holds logic to communicate between the
// external initiator service and the external endpoints it
// subscribes to.
package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/smartcontractkit/external-initiator/store"
)

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
type Event map[string]interface{}

// IParser holds the interface for parsing data
// from the external endpoint into an array of Events
// based on the blockchain's parser.
type IParser interface {
	ParseResponse(data []byte) ([]Event, bool)
}

// JsonrpcMessage declares JSON-RPC message type
type JsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *interface{}    `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func NewJsonrpcMessage(nonce uint64, method string, params json.RawMessage) ([]byte, error) {
	id, err := json.Marshal(nonce)
	if err != nil {
		return nil, err
	}

	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	return json.Marshal(msg)
}

// ISubscriber holds the interface for interacting with a blockchain node
type ISubscriber interface {
	// Subscribe to events of type t. Events are pushed to ch.
	Subscribe(ctx context.Context, method, unsubscribeMethod string, params json.RawMessage, ch chan<- json.RawMessage) error
	// Request data of type t.
	Request(ctx context.Context, method string, params json.RawMessage) (result json.RawMessage, err error)
	// Stop the subscriber and Stop all connections
	Stop()
}

func NewSubscriber(endpoint store.Endpoint) (ISubscriber, error) {
	u, err := url.Parse(endpoint.Url)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "ws", "wss":
		return NewWebsocketConnection(endpoint.Url)
	case "http", "https":
		// TODO: Implement
		return nil, nil
	}

	return nil, errors.New("unexpected URL scheme")
}
