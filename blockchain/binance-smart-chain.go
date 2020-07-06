package blockchain

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/eth"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"math/big"
)

const BSC = "binance-smart-chain"

// The bscManager implements the subscriber.JsonManager interface and allows
// for interacting with ETH nodes over RPC or WS.
type bscManager struct {
	fq *bscFilterQuery
	p  subscriber.Type
}

// createBscManager creates a new instance of bscManager with the provided
// connection type and store.Subscription config.
func createBscManager(p subscriber.Type, config store.Subscription) bscManager {
	var addresses []common.Address
	for _, a := range config.BinanceSmartChain.Addresses {
		addresses = append(addresses, common.HexToAddress(a))
	}

	topics := [][]common.Hash{{
		models.RunLogTopic20190207withoutIndexes,
		common.HexToHash(StringToBytes32(config.Job)),
	}}

	return bscManager{
		fq: &bscFilterQuery{
			Addresses: addresses,
			Topics:    topics,
		},
		p: p,
	}
}

// GetTriggerJson generates a JSON payload to the ETH node
// using the config in bscManager.
//
// If bscManager is using WebSocket:
// Creates a new "eth_subscribe" subscription.
//
// If bscManager is using RPC:
// Sends a "eth_getLogs" request.
func (e bscManager) GetTriggerJson() []byte {
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
// If bscManager is using WebSocket:
// Returns nil.
//
// If bscManager is using RPC:
// Sends a request to get the latest block number.
func (e bscManager) GetTestJson() []byte {
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
// If bscManager is using WebSocket:
// Returns nil.
//
// If bscManager is using RPC:
// Attempts to parse the block number in the response.
// If successful, stores the block number in bscManager.
func (e bscManager) ParseTestResponse(data []byte) error {
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

type bscSubscribeResponse struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

// ParseResponse parses the response from the
// ETH node, and returns a slice of subscriber.Events
// and if the parsing was successful.
//
// If bscManager is using RPC:
// If there are new events, update bscManager with
// the latest block number it sees.
func (e bscManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	logger.Debugw("Parsing response", "ExpectsMock", ExpectsMock)

	var msg jsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("failed parsing JSON-RPC message:", msg)
		return nil, false
	}

	var events []subscriber.Event

	switch e.p {
	case subscriber.WS:
		var res bscSubscribeResponse
		if err := json.Unmarshal(msg.Params, &res); err != nil {
			logger.Error("unmarshal:", err)
			return nil, false
		}

		var evt eth.Log
		if err := json.Unmarshal(res.Result, &evt); err != nil {
			logger.Error("unmarshal:", err)
			return nil, false
		}

		if evt.Removed {
			return nil, false
		}

		request, err := logEventToOracleRequest(evt)
		if err != nil {
			logger.Error("failed to get oracle request:", err)
			return nil, false
		}

		event, err := json.Marshal(request)
		if err != nil {
			logger.Error("marshal:", err)
			return nil, false
		}

		events = append(events, event)

	case subscriber.RPC:
		var rawEvents []eth.Log
		if err := json.Unmarshal(msg.Result, &rawEvents); err != nil {
			logger.Error(err)
			return nil, false
		}

		for _, evt := range rawEvents {
			if evt.Removed {
				continue
			}

			request, err := logEventToOracleRequest(evt)
			if err != nil {
				logger.Error("failed to get oracle request:", err)
				return nil, false
			}

			event, err := json.Marshal(request)
			if err != nil {
				logger.Error("failed marshaling request:", err)
				continue
			}
			events = append(events, event)

			// Check if we can update the "fromBlock" in the query,
			// so we only get new events from blocks we haven't queried yet
			// Increment the block number by 1, since we want events from *after* this block number
			curBlkn := &big.Int{}
			curBlkn = curBlkn.Add(big.NewInt(int64(evt.BlockNumber)), big.NewInt(1))

			fromBlkn, err := hexutil.DecodeBig(e.fq.FromBlock)
			if err != nil && !(e.fq.FromBlock == "latest" || e.fq.FromBlock == "") {
				logger.Error(err)
				continue
			}

			// If our query "fromBlock" is "latest", or our current "fromBlock" is in the past compared to
			// the last event we received, we want to update the query
			if e.fq.FromBlock == "latest" || e.fq.FromBlock == "" || curBlkn.Cmp(fromBlkn) > 0 {
				e.fq.FromBlock = hexutil.EncodeBig(curBlkn)
			}
		}
	}

	return events, true
}

type bscFilterQuery struct {
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

func (q bscFilterQuery) toMapInterface() (interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != "" || q.ToBlock != "" {
			return nil, errors.New("cannot specify both BlockHash and FromBlock/ToBlock")
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
