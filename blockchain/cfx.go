package blockchain

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const CFX = "conflux"

// The cfxManager implements the subscriber.JsonManager interface and allows
// for interacting with CFX nodes over RPC.
type cfxManager struct {
	fq *cfxFilterQuery
	p  subscriber.Type
}

// createCfxManager creates a new instance of cfxManager with the provided
// connection type and store.cfxSubscription config.
func createCfxManager(p subscriber.Type, config store.Subscription) cfxManager {
	var addresses []common.Address
	for _, a := range config.Conflux.Addresses {
		addresses = append(addresses, common.HexToAddress(a))
	}

	var topics [][]common.Hash
	var t []common.Hash
	for _, value := range config.Conflux.Topics {
		if len(value) < 1 {
			continue
		}
		t = append(t, common.HexToHash(value))
	}
	topics = append(topics, t)

	return cfxManager{
		fq: &cfxFilterQuery{
			Addresses: addresses,
			Topics:    topics,
		},
		p: p,
	}
}

// GetTriggerJson generates a JSON payload to the CFX node
// using the config in cfxManager.

// cfxManager is using RPC:
// Sends a "cfx_getLogs" request.
func (e cfxManager) GetTriggerJson() []byte {
	if e.p == subscriber.RPC && e.fq.fromEpoch == "" {
		e.fq.fromEpoch = "latest_state"
	}

	filter, err := e.fq.toMapInterface()
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

	msg.Method = "cfx_getLogs"
	msg.Params = json.RawMessage(`[` + string(filterBytes) + `]`)

	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil
	}

	return bytes
}

// GetTestJson generates a JSON payload to test
// the connection to the CFX node.
//
// cfxManager is using RPC:
// Sends a request to get the latest epoch number.
func (e cfxManager) GetTestJson() []byte {
	if e.p == subscriber.RPC {
		msg := JsonrpcMessage{
			Version: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "cfx_epochNumber",
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
// CFX node after sending GetTestJson(), and returns
// the error from parsing, if any.
//
// cfxManager is using RPC:
// Attempts to parse the block number in the response.
// If successful, stores the block number in cfxManager.
func (e cfxManager) ParseTestResponse(data []byte) error {
	if e.p == subscriber.RPC {
		var msg JsonrpcMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}
		var res string
		if err := json.Unmarshal(msg.Result, &res); err != nil {
			return err
		}
		e.fq.fromEpoch = res
	}

	return nil
}

type cfxLogResponse struct {
	LogIndex         string   `json:"logIndex"`
	EpochNumber      string   `json:"epochNumber"`
	BlockHash        string   `json:"blockHash"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

// ParseResponse parses the response from the
// CFX node, and returns a slice of subscriber.Events
// and if the parsing was successful.
//
// If cfxManager is using RPC:
// If there are new events, update cfxManager with
// the latest block number it sees.
func (e cfxManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	logger.Debugw("Parsing response", "ExpectsMock", ExpectsMock)

	var msg JsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("failed parsing msg: ", msg)
		return nil, false
	}

	var events []subscriber.Event

	switch e.p {
	case subscriber.RPC:
		var rawEvents []cfxLogResponse
		if err := json.Unmarshal(msg.Result, &rawEvents); err != nil {
			return nil, false
		}

		for _, evt := range rawEvents {
			event, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			events = append(events, event)

			// Check if we can update the "fromEpoch" in the query,
			// so we only get new events from blocks we haven't queried yet
			curBlkn, err := hexutil.DecodeBig(evt.EpochNumber)
			if err != nil {
				continue
			}
			// Increment the block number by 1, since we want events from *after* this block number
			curBlkn.Add(curBlkn, big.NewInt(1))

			fromBlkn, err := hexutil.DecodeBig(e.fq.fromEpoch)
			if err != nil && !(e.fq.fromEpoch == "latest_mined" || e.fq.fromEpoch == "") {
				continue
			}

			// If our query "fromEpoch" is "latest_mined", or our current "fromEpoch" is in the past compared to
			// the last event we received, we want to update the query
			if e.fq.fromEpoch == "latest_mined" || e.fq.fromEpoch == "" || curBlkn.Cmp(fromBlkn) > 0 {
				e.fq.fromEpoch = hexutil.EncodeBig(curBlkn)
			}
		}
	}

	return events, true
}

type cfxFilterQuery struct {
	BlockHash *common.Hash     // used by cfx_getLogs, return logs only from block with this hash
	fromEpoch string           // beginning of the queried range, nil means genesis block
	toEpoch   string           // end of the range, nil means latest block
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

func (q cfxFilterQuery) toMapInterface() (interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.fromEpoch != "" || q.toEpoch != "" {
			return nil, errors.New("cannot specify both BlockHash and fromEpoch/toEpoch")
		}
	} else {
		if q.fromEpoch == "" {
			arg["fromEpoch"] = "0x0"
		} else {
			arg["fromEpoch"] = q.fromEpoch
		}
		if q.toEpoch == "" {
			arg["toEpoch"] = "latest_mined"
		} else {
			arg["toEpoch"] = q.toEpoch
		}
	}
	return arg, nil
}
