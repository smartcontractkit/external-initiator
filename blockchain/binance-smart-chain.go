package blockchain

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const BSC = "binance-smart-chain"

// The bscManager implements the subscriber.JsonManager interface and allows
// for interacting with ETH nodes over RPC or WS.
type bscManager struct {
	ethManager
}

// createBscManager creates a new instance of bscManager with the provided
// connection type and store.Subscription config.
func createBscManager(p subscriber.Type, config store.Subscription) bscManager {
	return bscManager{
		ethManager{
			fq: createEvmFilterQuery(config.Job, config.BinanceSmartChain.Addresses),
			p:  p,
		},
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
	return e.ethManager.GetTriggerJson()
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
	return e.ethManager.GetTestJson()
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
	return e.ethManager.ParseTestResponse(data)
}

// ParseResponse parses the response from the
// ETH node, and returns a slice of subscriber.Events
// and if the parsing was successful.
//
// If bscManager is using RPC:
// If there are new events, update bscManager with
// the latest block number it sees.
func (e bscManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	logger.Debugw("Parsing Binance Smart Chain response", "ExpectsMock", ExpectsMock)

	var msg JsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("failed parsing JSON-RPC message:", msg)
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

		var evt models.Log
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
		var rawEvents []models.Log
		if err := json.Unmarshal(msg.Result, &rawEvents); err != nil {
			logger.Error("unmarshal:", err)
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
				logger.Error("Failed to get block number from event:", err)
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
