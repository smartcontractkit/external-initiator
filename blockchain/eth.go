package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"math/big"
)

const ETH = "ethereum"

// The EthManager implements the subscriber.Manager interface and allows
// for interacting with ETH nodes over RPC or WS.
type EthManager struct {
	fq *filterQuery
	p  subscriber.Type
}

// CreateEthManager creates a new instance of EthManager with the provided
// connection type and store.EthSubscription config.
func CreateEthManager(p subscriber.Type, config store.EthSubscription) EthManager {
	var addresses []common.Address
	for _, a := range config.Addresses {
		addresses = append(addresses, common.HexToAddress(a))
	}

	var topics [][]common.Hash
	var t []common.Hash
	for _, value := range config.Topics {
		if len(value) < 1 {
			continue
		}
		t = append(t, common.HexToHash(value))
	}
	topics = append(topics, t)

	return EthManager{
		fq: &filterQuery{
			Addresses: addresses,
			Topics:    topics,
		},
		p: p,
	}
}

// GetTriggerJson generates a JSON payload to the ETH node
// using the config in EthManager.
//
// If EthManager is using WebSocket:
// Creates a new "eth_subscribe" subscription.
//
// If EthManager is using RPC:
// Sends a "eth_getLogs" request.
func (e EthManager) GetTriggerJson() []byte {
	if e.p == subscriber.RPC && e.fq.FromBlock == "" {
		e.fq.FromBlock = "latest"
	}

	filter, err := e.fq.toMapInterface()
	if err != nil {
		return nil
	}

	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil
	}

	msg := jsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
	}

	switch e.p {
	case subscriber.WS:
		msg.Method = "eth_subscribe"
		msg.Params = json.RawMessage(`["logs",` + string(filterBytes) + `]`)
	case subscriber.RPC:
		msg.Method = "eth_getLogs"
		msg.Params = json.RawMessage(`[` + string(filterBytes) + `]`)
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil
	}

	return bytes
}

// GetTestJson generates a JSON payload to test
// the connection to the ETH node.
//
// If EthManager is using WebSocket:
// Returns nil.
//
// If EthManager is using RPC:
// Sends a request to get the latest block number.
func (e EthManager) GetTestJson() []byte {
	if e.p == subscriber.RPC {
		msg := jsonrpcMessage{
			Version: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "eth_blockNumber",
		}

		bytes, err := json.Marshal(msg)
		if err != nil {
			return nil
		}

		return bytes
	}

	return nil
}

// ParseTestResponse parses the response from the
// ETH node after sending GetTestJson(), and returns
// the error from parsing, if any.
//
// If EthManager is using WebSocket:
// Returns nil.
//
// If EthManager is using RPC:
// Attempts to parse the block number in the response.
// If successful, stores the block number in EthManager.
func (e EthManager) ParseTestResponse(data []byte) error {
	if e.p == subscriber.RPC {
		var msg jsonrpcMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}
		var res string
		if err := json.Unmarshal(msg.Result, &res); err != nil {
			return err
		}
		e.fq.FromBlock = res
	}

	return nil
}

type ethSubscribeResponse struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

type ethLogResponse struct {
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
// ETH node, and returns a slice of subscriber.Events
// and if the parsing was successful.
//
// If EthManager is using RPC:
// If there are new events, update EthManager with
// the latest block number it sees.
func (e EthManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	var msg jsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, false
	}

	var events []subscriber.Event

	switch e.p {
	case subscriber.WS:
		var res ethSubscribeResponse
		if err := json.Unmarshal(msg.Params, &res); err != nil {
			return nil, false
		}

		var evt ethLogResponse
		if err := json.Unmarshal(res.Result, &evt); err != nil {
			return nil, false
		}

		event, err := json.Marshal(evt)
		if err != nil {
			return nil, false
		}

		events = append(events, event)

	case subscriber.RPC:
		var rawEvents []ethLogResponse
		if err := json.Unmarshal(msg.Result, &rawEvents); err != nil {
			return nil, false
		}

		for _, evt := range rawEvents {
			// Check if we can update the "fromBlock" in the query,
			// so we only get new events from blocks we haven't queried yet
			curBlkn, err := hexutil.DecodeBig(evt.BlockNumber)
			if err != nil {
				continue
			}
			// Increment the block number by 1, since we want events from *after* this block number
			curBlkn.Add(curBlkn, big.NewInt(1))

			fromBlkn, err := hexutil.DecodeBig(e.fq.FromBlock)
			if err != nil {
				continue
			}

			// If our query "fromBlock" is "latest", or our current "fromBlock" is in the past compared to
			// the last event we received, we want to update the query
			if e.fq.FromBlock == "latest" || e.fq.FromBlock == "" || curBlkn.Cmp(fromBlkn) > 0 {
				e.fq.FromBlock = hexutil.EncodeBig(curBlkn)
			}

			event, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			events = append(events, event)
		}
	}

	return events, true
}

type filterQuery struct {
	BlockHash *common.Hash     // used by eth_getLogs, return logs only from block with this hash
	FromBlock string           // beginning of the queried range, nil means genesis block
	ToBlock   string           // end of the range, nil means latest block
	Addresses []common.Address // restricts matches to events created by specific contracts

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	Topics [][]common.Hash
}

func (q filterQuery) toMapInterface() (interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != "" || q.ToBlock != "" {
			return nil, fmt.Errorf("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == "" {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = q.FromBlock
		}
		if q.ToBlock == "" {
			arg["toBlock"] = "latest"
		} else {
			arg["toBlock"] = q.ToBlock
		}
	}
	return arg, nil
}

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *interface{}    `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}
