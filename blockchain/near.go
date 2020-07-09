package blockchain

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const (
	// NEAR platform name
	NEAR = "near"
	// maxRequests max number of requests contract returns
	maxRequests = 10
)

// NEARQueryCall is a JSON-RPC Params strutc for NEAR JSON-RPC query Method
type NEARQueryCall struct {
	RequestType string `json:"request_type"`
	Finality    string `json:"finality"`
	AccountID   string `json:"account_id"`
	MethodName  string `json:"method_name"`
	ArgsBase64  string `json:"args_base64"`
}

// NEARQueryResult is a result struct for NEAR JSON-RPC NEARQueryCall response
type NEARQueryResult struct {
	Result      []uint8 `json:"result"`
	Logs        []uint8 `json:"logs"`
	BlockHeight uint64  `json:"block_height"`
	BlockHash   string  `json:"block_hash"`
}

// NEARVersion type contains NEAR version info
type NEARVersion struct {
	Build   string `json:"build"`
	Version string `json:"version"`
}

// NEARValidator type contains NEAR validator info
type NEARValidator struct {
	AccountID string `json:"account_id"`
	IsSlashed bool   `json:"is_slashed"`
}

// NEARSyncInfo type contains NEAR sync info
type NEARSyncInfo struct {
	LatestBlockHash   string `json:"latest_block_hash"`
	LatestBlockHeight uint32 `json:"latest_block_height"`
	LatestBlockTime   string `json:"latest_block_time"`
	LatestStateRoot   string `json:"latest_state_root"`
	Syncing           bool   `json:"syncing"`
}

// NEARStatus is a result type for NEAR JSON-RPC status response, contains NEAR network status info
type NEARStatus struct {
	ChainID               string          `json:"chain_id"`
	LatestProtocolVersion uint16          `json:"latest_protocol_version"`
	ProtocolVersion       uint16          `json:"protocol_version"`
	RPCAddr               string          `json:"rpc_addr"`
	SyncInfo              NEARSyncInfo    `json:"sync_info"`
	Validators            []NEARValidator `json:"validators"`
}

type NEARManager struct {
	connectionType subscriber.Type
	status         *NEARStatus
}

// createNEARManager creates a new instance of NEARManager with the provided
// connection type and store.Subscription config.
func createNEARManager(connectionType subscriber.Type, config store.Subscription) (*NEARManager, error) {
	if connectionType != subscriber.RPC {
		return nil, errors.New("only RPC connections are allowed for NEAR")
	}

	return &NEARManager{
		connectionType: connectionType,
	}, nil
}

// GetTriggerJson generates a JSON payload to the NEAR node
// using the config in NEARManager.
//
// If NEARManager is using RPC: Returns a "query" request.
func (m NEARManager) GetTriggerJson() []byte {
	// TODO: hardcoded client account
	clientAccount := "client.oracle.testnet"
	args := fmt.Sprintf(`{"account": "%v", "max_requests": "%v"}`, clientAccount, maxRequests)

	queryCall := NEARQueryCall{
		RequestType: "call_function",
		Finality:    "final",
		AccountID:   "oracle.oracle.testnet", // TODO: hardcoded oracle account
		MethodName:  "get_requests",
		ArgsBase64:  b64.StdEncoding.EncodeToString([]byte(args)),
	}

	queryCallBytes, err := json.Marshal(queryCall)
	if err != nil {
		return nil
	}

	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
	}

	switch m.connectionType {
	case subscriber.RPC:
		msg.Method = "query"
		msg.Params = json.RawMessage(string(queryCallBytes))
	default:
		return nil
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil
	}

	return bytes
}

func (m NEARManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	logger.Debugw("Parsing response", "ExpectsMock", ExpectsMock)

	var msg JsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("Failed parsing JSON-RPC message: ", err)
		return nil, false
	}

	// TODO: build []subscriber.Event
	return nil, false
}

// GetTestJson generates a JSON payload to test
// the connection to the NEAR node.
//
// If NEARManager is using WebSocket:
// Returns nil.
//
// If NEARManager is using RPC:
// Returns a request to get the network status.
func (m NEARManager) GetTestJson() []byte {
	if m.connectionType == subscriber.RPC {
		msg := JsonrpcMessage{
			Version: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "status",
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
// NEAR node after sending GetTestJson(), and returns
// the error from parsing, if any.
//
// If NEARManager is using WebSocket:
// Returns nil.
//
// If NEARManager is using RPC:
// Attempts to parse the status in the response.
// If successful, stores the status in NEARManager.
func (m NEARManager) ParseTestResponse(data []byte) error {
	if m.connectionType == subscriber.RPC {
		var msg JsonrpcMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}

		var res NEARStatus
		if err := json.Unmarshal(msg.Result, &res); err != nil {
			return err
		}
		m.status = &res
	}

	return nil
}
