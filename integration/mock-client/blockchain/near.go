package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/smartcontractkit/external-initiator/blockchain"
)

func handleNEARRequest(conn string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	if conn == "rpc" {
		switch msg.Method {
		case "query":
			responses, ok := GetCannedResponses("near")
			if !ok {
				return nil, fmt.Errorf("Failed to load canned responses for: %v", "near")
			}

			respID, err := buildResponseID(msg)
			if err != nil {
				return nil, err
			}
			responseList, ok := responses[respID]
			if !ok {
				errResp := responses["error_MethodNotFound"]
				return errResp, nil
			}

			return setJsonRpcId(msg.ID, responseList), nil
		}
		// TODO: why don't we return a JSON-RPC error here?
		// NEAR API Example: {"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found","data":"nonexistent_method_123"},"id":"chainlink"}
		return nil, fmt.Errorf("Unexpected method: %v", msg.Method)
	}

	return nil, fmt.Errorf("Unexpected connection: %v", conn)
}

// buildResponseID builds a response ID for supplied JSON-RPC message,
// that can be used to find disk stored canned respones.
func buildResponseID(msg JsonrpcMessage) (string, error) {
	if msg.Method == "" {
		return "", fmt.Errorf("Failed to build response ID (Method not available): %v", msg)
	}

	var params blockchain.NEARQueryCall
	err := json.Unmarshal(msg.Params, &params)
	if err != nil {
		return "", err
	}

	if params.MethodName == "" {
		return msg.Method, nil
	}

	return fmt.Sprintf("%v_%v", msg.Method, params.MethodName), nil
}
