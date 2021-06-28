package blockchain

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const Xinfin = "xinfin"

// The xinfinManager implements the subscriber.JsonManager interface and allows
// for interacting with Xinfin nodes over RPC or WS.
// It is different from eth that it uses 'eth' instead of 'eth' in method.
// If you are subscribing something other than Log, it could have different
// struct from eth's.
type xinfinManager struct {
	ethManager
}

// createXinfinManager creates a new instance of xinfinManager with the provided
// connection type and store.EthSubscription config.
func createXinfinManager(p subscriber.Type, config store.Subscription) xinfinManager {

	var addresses []common.Address
	for _, a := range config.Xinfin.Addresses {
		b := strings.Replace(a,"xdc","0x",1)
		addresses = append(addresses, common.HexToAddress(b))

	}

	var topics [][]common.Hash
	var t []common.Hash
	for _, value := range config.Xinfin.Topics {
		if len(value) < 1 {
			continue
		}
		t = append(t, common.HexToHash(value))
	}
	topics = append(topics, t)

	return xinfinManager{
		ethManager{
			fq: &filterQuery{
				Addresses: addresses,
				Topics:    topics,
			},
			p:            p,
			endpointName: config.EndpointName,
			jobid:        config.Job,
		},
	}
}

// GetTriggerJson generates a JSON payload to the Xinfin node
// using the config in xinfinManager.
//
// If xinfinManager is using WebSocket:
// Creates a new "eth_subscribe" subscription.
//
// If xinfinManager is using RPC:
// Sends a "eth_getLogs" request.
func (k xinfinManager) GetTriggerJson() []byte {
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
		msg.Method = "eth_subscribe"
		msg.Params = json.RawMessage(`["logs",` + string(filterBytes) + `]`)
	case subscriber.RPC:
		msg.Method = "eth_getLogs"
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
// the connection to the Xinfin node.
//
// If xinfinManager is using WebSocket:
// Returns nil.
//
// If xinfinManager is using RPC:
// Sends a request to get the latest block number.
func (k xinfinManager) GetTestJson() []byte {
	if k.p == subscriber.RPC {
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
// Xinfin node after sending GetTestJson(), and returns
// the error from parsing, if any.
//
// If xinfinManager is using WebSocket:
// Returns nil.
//
// If xinfinManager is using RPC:
// Attempts to parse the block number in the response.
// If successful, stores the block number in xinfinManager.
func (k xinfinManager) ParseTestResponse(data []byte) error {
	return k.ethManager.ParseTestResponse(data)
}

// ParseResponse parses the response from the
// Xinfin node, and returns a slice of subscriber.Events
// and if the parsing was successful.
//
// If xinfinManager is using RPC:
// If there are new events, update xinfinManager with
// the latest block number it sees.
func (k xinfinManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	promLastSourcePing.With(prometheus.Labels{"endpoint": k.endpointName, "jobid": k.jobid}).SetToCurrentTime()
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

		var evt models.Log
		if err := json.Unmarshal(res.Result, &evt); err != nil {
			logger.Error("unmarshal:", err)
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
		logger.Warnw("receive message from subscribe", "evt", evt, "message", event)
		events = append(events, event)

	case subscriber.RPC:
		var rawEvents []models.Log

		actualValue := string(msg.Result[:])
		modifiedValue := strings.Replace(actualValue,"xdc","0x",1)
		toSendData := []byte(modifiedValue)
		
		if err := json.Unmarshal(toSendData, &rawEvents); err != nil {
			logger.Error("unmarshal:", err)
			return nil, false
		}

		for _, evt := range rawEvents {
			request, err := logEventToOracleRequest(evt)
			if err != nil {
				logger.Error("failed to get oracle request:", err, evt.Data, evt.Address)
				return nil, false
			}

			event, err := json.Marshal(request)
			if err != nil {
				continue
			}
			events = append(events, event)

			// Check if we can update the "fromBlock" in the query,
			// so we only get new events from blocks we haven't queried yet
			// Increment the block number by 1, since we want events from *after* this block number
			curBlkn := &big.Int{}
			curBlkn = curBlkn.Add(big.NewInt(int64(evt.BlockNumber)), big.NewInt(1))

			fromBlkn, err := hexutil.DecodeBig(k.fq.FromBlock)
			if err != nil && !(k.fq.FromBlock == "latest" || k.fq.FromBlock == "") {
				logger.Error("Failed to get block number from event:", err)
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
