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

// NEARQuery is a JSON-RPC Params struct for NEAR JSON-RPC query Method
type NEARQuery struct {
	RequestType string `json:"request_type"`
	Finality    string `json:"finality"`
	AccountID   string `json:"account_id"`
}

// NEARQueryCall is a JSON-RPC Params struct for NEAR JSON-RPC query Method
// where "request_type": "call_function".
//
// NEAR "call_function" request type, calls method_name in contract account_id
// as view function with data as parameters.
type NEARQueryCall struct {
	// TODO: how to embed NEARQuery here?
	RequestType string `json:"request_type"`
	Finality    string `json:"finality"`
	AccountID   string `json:"account_id"`
	MethodName  string `json:"method_name"`
	ArgsBase64  string `json:"args_base64"` // base64-encoded
}

// NEARQueryResult is a result struct for NEAR JSON-RPC NEARQueryCall response
type NEARQueryResult struct {
	Result      json.RawMessage `json:"result"`
	Logs        []byte          `json:"logs"`
	BlockHeight uint64          `json:"block_height"`
	BlockHash   string          `json:"block_hash"`
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

// NEAROracleRequestArgs contains the oracle request arguments
type NEAROracleRequestArgs struct {
	CallerAccount   string `json:"caller_account"`
	RequestSpec     string `json:"request_spec"` // base64-encoded
	CallbackAddress string `json:"callback_address"`
	CallbackMethod  string `json:"callback_method"`
	Data            string `json:"data"`       // base64-encoded
	Payment         uint64 `json:"payment"`    // in LINK tokens
	Expiration      uint64 `json:"expiration"` // in nanoseconds
}

// NEAROracleRequest is the request returned by the oracle get_requests function
type NEAROracleRequest struct {
	Nonce   uint32                `json:"nonce"`
	Request NEAROracleRequestArgs `json:"request"`
}

type nearFilter struct {
	JobID      string
	AccountIDs []string
}

type nearManager struct {
	filter         nearFilter
	connectionType subscriber.Type
	status         *NEARStatus
}

// createNearManager creates a new instance of nearManager with the provided
// connection type and store.Subscription config.
func createNearManager(connectionType subscriber.Type, config store.Subscription) (*nearManager, error) {
	if connectionType != subscriber.RPC {
		return nil, errors.New("only RPC connections are allowed for NEAR")
	}

	var accountIDs []string
	for _, id := range config.NEAR.AccountIds {
		accountIDs = append(accountIDs, id)
	}

	return &nearManager{
		filter: nearFilter{
			JobID:      config.Job,
			AccountIDs: accountIDs,
		},
		connectionType: connectionType,
	}, nil
}

// GetTriggerJson generates a JSON payload to the NEAR node
// using the config in nearManager.
//
// If nearManager is using RPC: Returns a "query" request.
func (m nearManager) GetTriggerJson() []byte {
	// TODO: hardcoded client account
	// We are not interested to query requests for a specific client,
	// but all requests made through a specific contract.
	clientAccount := "client.oracle.testnet"
	args := fmt.Sprintf(`{"account": "%v", "max_requests": "%v"}`, clientAccount, maxRequests)

	queryCall := NEARQueryCall{
		RequestType: "call_function",
		Finality:    "final",
		// TODO: hardcoded first oracle account
		// How do we support query for multiple oracle accounts/contracts?
		AccountID:  m.filter.AccountIDs[0],
		MethodName: "get_requests",
		ArgsBase64: b64.StdEncoding.EncodeToString([]byte(args)),
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

func (m nearManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	logger.Debugw("Parsing NEAR response", "ExpectsMock", ExpectsMock)

	var msg JsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("Failed parsing JSON-RPC message: ", err)
		return nil, false
	}

	var queryResult NEARQueryResult
	if err := json.Unmarshal(msg.Result, &queryResult); err != nil {
		logger.Error("Failed parsing NEARQueryResult:", err)
		return nil, false
	}

	var oracleRequests []NEAROracleRequest
	if err := json.Unmarshal(queryResult.Result, &oracleRequests); err != nil {
		logger.Error("Failed parsing NEAROracleRequests:", err)
		return nil, false
	}

	var events []subscriber.Event

	for _, r := range oracleRequests {
		request := r.Request

		jobID, err := b64.StdEncoding.DecodeString(request.RequestSpec)
		if err != nil {
			logger.Error("Failed parsing NEAROracleRequestArgs.RequestSpec:", err)
			return nil, false
		}

		// Check if our jobID matches
		if !matchesJobID(m.filter.JobID, string(jobID)) {
			continue
		}

		event, err := json.Marshal(request)
		if err != nil {
			logger.Error("failed marshaling request:", err)
			continue
		}
		events = append(events, event)
	}

	return events, false
}

// GetTestJson generates a JSON payload to test
// the connection to the NEAR node.
//
// If nearManager is using WebSocket:
// Returns nil.
//
// If nearManager is using RPC:
// Returns a request to get the network status.
func (m nearManager) GetTestJson() []byte {
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
// If nearManager is using WebSocket:
// Returns nil.
//
// If nearManager is using RPC:
// Attempts to parse the status in the response.
// If successful, stores the status in nearManager.
func (m nearManager) ParseTestResponse(data []byte) error {
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
