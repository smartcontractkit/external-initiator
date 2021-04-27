package blockchain

import (
	"encoding/json"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const Klaytn = "klaytn"

// The klaytnManager implements the subscriber.JsonManager interface and allows
// for interacting with Klaytn nodes over RPC or WS.
// It is different from eth that it uses 'klay' instead of 'eth' in method.
// If you are subscribing something other than Log, it could have different
// struct from eth's.
type klaytnManager struct {
	ethManager
}

// createKlaytnManager creates a new instance of klaytnManager with the provided
// connection type and store.EthSubscription config.
func createKlaytnManager(p subscriber.Type, config store.Subscription) klaytnManager {
	return klaytnManager{
		createEthManager(p, config),
	}
}

// GetTriggerJson generates a JSON payload to the Klaytn node
// using the config in klaytnManager.
//
// If klaytnManager is using WebSocket:
// Creates a new "klay_subscribe" subscription.
//
// If klaytnManager is using RPC:
// Sends a "klay_getLogs" request.
func (k klaytnManager) GetTriggerJson() []byte {
	if k.p == subscriber.RPC && k.fq.FromBlock == "" {
		k.fq.FromBlock = "latest"
	}

	filter, err := k.fq.toMapInterface()
	if err != nil {
		return nil
	}

	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil
	}

	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
	}

	switch k.p {
	case subscriber.WS:
		msg.Method = "klay_subscribe"
		msg.Params = json.RawMessage(`["logs",` + string(filterBytes) + `]`)
	case subscriber.RPC:
		msg.Method = "klay_getLogs"
		msg.Params = json.RawMessage(`[` + string(filterBytes) + `]`)
	default:
		logger.Errorw(ErrSubscriberType.Error(), "type", k.p)
		return nil
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil
	}

	return bytes
}

// GetTestJson generates a JSON payload to test
// the connection to the Klaytn node.
//
// If klaytnManager is using WebSocket:
// Returns nil.
//
// If klaytnManager is using RPC:
// Sends a request to get the latest block number.
func (k klaytnManager) GetTestJson() []byte {
	if k.p == subscriber.RPC {
		msg := JsonrpcMessage{
			Version: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "klay_blockNumber",
		}

		bytes, err := json.Marshal(msg)
		if err != nil {
			return nil
		}

		return bytes
	}

	return nil
}
