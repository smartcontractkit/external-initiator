package blockchain

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const ETH = "ethereum"

// The ethManager implements the subscriber.JsonManager interface and allows
// for interacting with ETH nodes over RPC or WS.
type ethManager struct {
	fq           *filterQuery
	p            subscriber.Type
	endpointName string
	jobid        string
}

// createEthManager creates a new instance of ethManager with the provided
// connection type and store.EthSubscription config.
func createEthManager(p subscriber.Type, config store.Subscription) ethManager {
	var addresses []common.Address
	for _, a := range config.Ethereum.Addresses {
		addresses = append(addresses, common.HexToAddress(a))
	}

	var topics [][]common.Hash
	var t []common.Hash
	for _, value := range config.Ethereum.Topics {
		if len(value) < 1 {
			continue
		}
		t = append(t, common.HexToHash(value))
	}
	topics = append(topics, t)

	return ethManager{
		fq: &filterQuery{
			Addresses: addresses,
			Topics:    topics,
		},
		p:            p,
		endpointName: config.EndpointName,
		jobid:        config.Job,
	}
}

// GetTriggerJson generates a JSON payload to the ETH node
// using the config in ethManager.
//
// If ethManager is using WebSocket:
// Creates a new "eth_subscribe" subscription.
//
// If ethManager is using RPC:
// Sends a "eth_getLogs" request.
func (e ethManager) GetTriggerJson() []byte {
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

	msg := JsonrpcMessage{
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
	default:
		logger.Errorw(ErrSubscriberType.Error(), "type", e.p)
		return nil
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
// If ethManager is using WebSocket:
// Returns nil.
//
// If ethManager is using RPC:
// Sends a request to get the latest block number.
func (e ethManager) GetTestJson() []byte {
	if e.p == subscriber.RPC {
		msg := JsonrpcMessage{
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
// If ethManager is using WebSocket:
// Returns nil.
//
// If ethManager is using RPC:
// Attempts to parse the block number in the response.
// If successful, stores the block number in ethManager.
func (e ethManager) ParseTestResponse(data []byte) error {
	if e.p == subscriber.RPC {
		var msg JsonrpcMessage
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
// If ethManager is using RPC:
// If there are new events, update ethManager with
// the latest block number it sees.
func (e ethManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	promLastSourcePing.With(prometheus.Labels{"endpoint": e.endpointName, "jobid": e.jobid}).SetToCurrentTime()
	logger.Debugw("Parsing response", "ExpectsMock", ExpectsMock)

	var msg JsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("failed parsing msg: ", msg)
		return nil, false
	}

	var events []subscriber.Event

	switch e.p {
	case subscriber.WS:
		var res ethSubscribeResponse
		if err := json.Unmarshal(msg.Params, &res); err != nil {
			logger.Error("unmarshal:", err)
			return nil, false
		}

		var evt ethLogResponse
		if err := json.Unmarshal(res.Result, &evt); err != nil {
			logger.Error("unmarshal:", err)
			return nil, false
		}

		event, err := json.Marshal(evt)
		if err != nil {
			logger.Error("marshal:", err)
			return nil, false
		}
		events = append(events, event)

	case subscriber.RPC:
		var rawEvents []ethLogResponse
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

			fromBlkn, err := hexutil.DecodeBig(e.fq.FromBlock)
			if err != nil && !(e.fq.FromBlock == "latest" || e.fq.FromBlock == "") {
				continue
			}

			// If our query "fromBlock" is "latest", or our current "fromBlock" is in the past compared to
			// the last event we received, we want to update the query
			if e.fq.FromBlock == "latest" || e.fq.FromBlock == "" || curBlkn.Cmp(fromBlkn) > 0 {
				e.fq.FromBlock = hexutil.EncodeBig(curBlkn)
			}
		}

	default:
		logger.Errorw(ErrSubscriberType.Error(), "type", e.p)
		return nil, false
	}

	return events, true
}
