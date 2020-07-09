package blockchain

import (
	"encoding/json"
	"time"

	"github.com/smartcontractkit/external-initiator/subscriber"
)

const (
	// NEAR platform name
	NEAR             = "near"
	NEARScanInterval = 5 * time.Second
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

type NEARManager struct {
	p      subscriber.Type
	status *NEARStatus
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

// GetTestJson generates a JSON payload to test
// the connection to the NEAR node.
//
// If NEARManager is using WebSocket:
// Returns nil.
//
// If NEARManager is using RPC:
// Returns a request to get the network status.
func (m NEARManager) GetTestJson() []byte {
	if m.p == subscriber.RPC {
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
	if m.p == subscriber.RPC {
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
