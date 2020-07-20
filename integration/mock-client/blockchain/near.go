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
				return nil, fmt.Errorf("failed to load canned responses for: %v", "near")
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
		// TODO: https://www.pivotaltracker.com/story/show/173896260
		return nil, fmt.Errorf("unexpected method: %v", msg.Method)
	}

	return nil, fmt.Errorf("unexpected connection: %v", conn)
}

// buildResponseID builds a response ID for supplied JSON-RPC message,
// that can be used to find disk stored canned respones.
func buildResponseID(msg JsonrpcMessage) (string, error) {
	if msg.Method == "" {
		return "", fmt.Errorf("failed to build response ID (Method not available): %v", msg)
	}

	var params blockchain.NEARQueryCallFunction
	err := json.Unmarshal(msg.Params, &params)
	if err != nil {
		return "", err
	}

	if params.MethodName == "" {
		return msg.Method, nil
	}

	return fmt.Sprintf("%v_%v", msg.Method, params.MethodName), nil
}
