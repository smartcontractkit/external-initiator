package blockchain

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/eth"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/smartcontractkit/chainlink/core/store/models"
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

	// Hard-set the topics to match the OracleRequest()
	// event emitted by the oracle contract provided.
	topics := [][]common.Hash{{
		models.RunLogTopic20190207withoutIndexes,
	}, {
		common.HexToHash(StringToBytes32(config.Job)),
	}}

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
	LogIndex         string   			`json:"logIndex"`
	EpochNumber      string   			`json:"epochNumber"`
	BlockHash        common.Hash 		`json:"blockHash"`
	TransactionHash  common.Hash 		`json:"transactionHash"`
	TransactionIndex string   			`json:"transactionIndex"`
	Address          common.Address `json:"address"`
	Data             string   			`json:"data"`
	Topics           []common.Hash 	`json:"topics"`
}

//convert cfxLogResponse type to eth.Log type
func Cfx2EthResponse(cfx cfxLogResponse) (eth.Log, error) {
	blockNumber, err_block := hexutil.DecodeUint64(cfx.EpochNumber)
	txIndex, err_tx := hexutil.DecodeUint64(cfx.TransactionIndex)
	index, err_ind := hexutil.DecodeUint64(cfx.LogIndex)
	data, err_dat := hexutil.Decode(cfx.Data)

	//conversion error handling
	if (err_block != nil || err_tx != nil || err_ind != nil || err_dat != nil) {
			var err_str = ""
			if err_block != nil {
				err_str += " blockNumber"
			}

			if err_tx != nil {
				err_str += " txIndex"
			}

			if err_ind != nil {
				err_str += " index"
			}

			if err_dat != nil {
				err_str += " data"
			}

			return eth.Log{}, errors.New(err_str)
	}

	return eth.Log{
		Address: cfx.Address,
		Topics: cfx.Topics,
		Data: data,
		BlockNumber: blockNumber,
		TxHash: cfx.TransactionHash,
		TxIndex: uint(txIndex),
		BlockHash: cfx.BlockHash,
		Index: uint(index),
	}, nil
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

			//filtering + error handling
			evt_eth, err := Cfx2EthResponse(evt)
			if err != nil {
				logger.Error("failed to convert to ETH log type", nil)
				return nil, false
			}

			request, err := logEventToOracleRequest(evt_eth)
			if err != nil {
				logger.Error("failed to get oracle request:", err)
				return nil, false
			}

			event, err := json.Marshal(request)
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
			arg["toEpoch"] = "latest_state"
		} else {
			arg["toEpoch"] = q.toEpoch
		}
	}
	return arg, nil
}
