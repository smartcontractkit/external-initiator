package blockchain

import (
	"encoding/base64"
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
	// maxNumAccounts max number of accounts "get_all_requests" contract fn returns
	maxNumAccounts = 1000
	// maxRequests max number of requests "get_all_requests" contract fn returns
	maxRequests = 1000
)

// NEARQueryCallFunction is a JSON-RPC Params struct for NEAR JSON-RPC query Method
// where "request_type": "call_function".
//
// NEAR "call_function" request type, calls method_name in contract account_id
// as view function with data as parameters.
type NEARQueryCallFunction struct {
	RequestType string `json:"request_type"`
	Finality    string `json:"finality"`
	AccountID   string `json:"account_id"`
	MethodName  string `json:"method_name"`
	ArgsBase64  string `json:"args_base64"` // base64-encoded
}

// NEARQueryResult is a result struct for NEAR JSON-RPC NEARQueryCallFunction response
type NEARQueryResult struct {
	Result      []byte `json:"result"`
	Logs        []byte `json:"logs"`
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
}

// NEARVersion type contains NEAR build & version info
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

// NEAROracleFnGetAllRequestsArgs represents function arguments for NEAR oracle 'get_all_requests' function
type NEAROracleFnGetAllRequestsArgs struct {
	MaxNumAccounts uint16 `json:"max_num_accounts"`
	MaxRequests    uint16 `json:"max_requests"`
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

// NEAROracleRequest is the request returned by the oracle 'get_requests' or 'get_all_requests' function
type NEAROracleRequest struct {
	Nonce   string                `json:"nonce"`
	Request NEAROracleRequestArgs `json:"request"`
}

// nearFilter holds the context data used to filter oracle requests for this subscription
type nearFilter struct {
	JobID      string
	AccountIDs []string
}

// nearManager implements NEAR subscription management
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

	return &nearManager{
		filter: nearFilter{
			JobID:      config.Job,
			AccountIDs: config.NEAR.AccountIds,
		},
		connectionType: connectionType,
	}, nil
}

// GetTriggerJson generates a JSON payload to the NEAR node
// using the config in nearManager.
//
// If nearManager is using RPC: Returns a "query" request.
func (m nearManager) GetTriggerJson() []byte {
	// We get all requests made through a contract, with some limits.
	args := NEAROracleFnGetAllRequestsArgs{
		MaxNumAccounts: maxNumAccounts,
		MaxRequests:    maxRequests,
	}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		logger.Error("Failed to marshal NEAROracleFnGetAllRequestsArgs:", err)
		return nil
	}

	queryCall := NEARQueryCallFunction{
		RequestType: "call_function",
		Finality:    "final",
		// TODO: hardcoded first oracle account
		// How do we support query for multiple oracle accounts/contracts?
		AccountID:  m.filter.AccountIDs[0],
		MethodName: "get_all_requests",
		ArgsBase64: base64.StdEncoding.EncodeToString(argsBytes),
	}

	queryCallBytes, err := json.Marshal(queryCall)
	if err != nil {
		logger.Error("Failed to marshal NEARQueryCallFunction:", err)
		return nil
	}

	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
	}

	switch m.connectionType {
	case subscriber.RPC:
		msg.Method = "query"
		msg.Params = queryCallBytes
	default:
		return nil
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal JsonrpcMessage:", err)
		return nil
	}

	return bytes
}

// ParseResponse generates []subscriber.Event from JSON-RPC response, requested using the GetTriggerJson message
func (m nearManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	logger.Debugw("Parsing NEAR response", "ExpectsMock", ExpectsMock)

	var msg JsonrpcMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("Failed parsing JSON-RPC message: ", err)
		return nil, false
	}

	oracleRequestsMap, err := ParseNEAROracleRequestsMap(msg)
	if err != nil {
		logger.Error("Failed parsing NEAROracleRequests map:", err)
		return nil, false
	}

	var events []subscriber.Event

	for _, oracleRequests := range oracleRequestsMap {
		for _, r := range oracleRequests {
			request := r.Request

			requestSpecBytes, err := base64.StdEncoding.DecodeString(request.RequestSpec)
			if err != nil {
				logger.Error("Failed parsing NEAROracleRequestArgs.RequestSpec:", err)
				return nil, false
			}
			requestSpec := fmt.Sprintf("%s", requestSpecBytes)

			// Check if our jobID matches
			if !matchesJobID(m.filter.JobID, requestSpec) {
				continue
			}

			event, err := json.Marshal(request)
			if err != nil {
				logger.Error("Failed marshaling request:", err)
				continue
			}
			events = append(events, event)
		}
	}

	return events, true
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

// ParseNEARQueryResult will unmarshal JsonrpcMessage as a NEAR standard NEARQueryResult
func ParseNEARQueryResult(msg JsonrpcMessage) (*NEARQueryResult, error) {
	var queryResult NEARQueryResult
	if err := json.Unmarshal(msg.Result, &queryResult); err != nil {
		return nil, err
	}
	return &queryResult, nil
}

// ParseNEAROracleRequests will unmarshal JsonrpcMessage result.result as []NEAROracleRequest
func ParseNEAROracleRequests(msg JsonrpcMessage) ([]NEAROracleRequest, error) {
	queryResult, err := ParseNEARQueryResult(msg)
	if err != nil {
		return nil, err
	}

	var oracleRequests []NEAROracleRequest
	if err := json.Unmarshal(queryResult.Result, &oracleRequests); err != nil {
		return nil, err
	}

	return oracleRequests, nil
}

// ParseNEAROracleRequestsMap will unmarshal JsonrpcMessage result.result as map[string][]NEAROracleRequest
func ParseNEAROracleRequestsMap(msg JsonrpcMessage) (map[string][]NEAROracleRequest, error) {
	queryResult, err := ParseNEARQueryResult(msg)
	if err != nil {
		return nil, err
	}

	var oracleRequestsMap map[string][]NEAROracleRequest
	if err := json.Unmarshal(queryResult.Result, &oracleRequestsMap); err != nil {
		return nil, err
	}

	return oracleRequestsMap, nil
}
