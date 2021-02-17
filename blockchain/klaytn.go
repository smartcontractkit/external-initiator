package blockchain

import (
	"encoding/json"
	"math/big"

	"github.com/klaytn/klaytn/common/hexutil"

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

type klaytnLogResponse struct {
	LogIndex         string   `json:"logIndex"`
	BlockNumber      string   `json:"blockNumber"`
	BlockHash        string   `json:"blockHash"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

// ParseResponse parses the response from the
// Klaytn node, and returns a slice of subscriber.Events
// and if the parsing was successful.
//
// If klaytnManager is using RPC:
// If there are new events, update klaytnManager with
// the latest block number it sees.
func (k klaytnManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	logger.Debugw("Parsing response", "ExpectsMock", ExpectsMock)

	var msg JsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("failed parsing msg: ", msg)
		return nil, false
	}

	var events []subscriber.Event

	switch k.p {
	case subscriber.WS:
		var res ethSubscribeResponse
		if err := json.Unmarshal(msg.Params, &res); err != nil {
			logger.Error("unmarshal:", err)
			return nil, false
		}

		var evt klaytnLogResponse
		if err := json.Unmarshal(res.Result, &evt); err != nil {
			logger.Error("unmarshal:", err)
			return nil, false
		}

		event, err := json.Marshal(evt)
		if err != nil {
			logger.Error("marshal:", err)
			return nil, false
		}
		logger.Warnw("receive message from subscribe", "evt", evt, "message", event)
		events = append(events, event)

	case subscriber.RPC:
		var rawEvents []klaytnLogResponse
		if err := json.Unmarshal(msg.Result, &rawEvents); err != nil {
			return nil, false
		}

		for _, evt := range rawEvents {
			event, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			events = append(events, event)

			// Check if we can update the "fromBlock" in the query,
			// so we only get new events from blocks we haven't queried yet
			curBlkn, err := hexutil.DecodeBig(evt.BlockNumber)
			if err != nil {
				continue
			}
			// Increment the block number by 1, since we want events from *after* this block number
			curBlkn.Add(curBlkn, big.NewInt(1))

			fromBlkn, err := hexutil.DecodeBig(k.fq.FromBlock)
			if err != nil && !(k.fq.FromBlock == "latest" || k.fq.FromBlock == "") {
				continue
			}

			// If our query "fromBlock" is "latest", or our current "fromBlock" is in the past compared to
			// the last event we received, we want to update the query
			if k.fq.FromBlock == "latest" || k.fq.FromBlock == "" || curBlkn.Cmp(fromBlkn) > 0 {
				k.fq.FromBlock = hexutil.EncodeBig(curBlkn)
			}
		}

	default:
		logger.Errorw(ErrSubscriberType.Error(), "type", k.p)
		return nil, false
	}

	return events, true
}
