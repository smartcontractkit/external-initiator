package blockchain

import (
	"encoding/json"
	"fmt"
)

func handleBSNIritaRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "status", "block_results":
		rsp, ok := GetCannedResponse("birita", msg)
		if !ok {
			return nil, fmt.Errorf("failed to handle BSN-IRITA request for method %s", msg.Method)
		}

		return rsp, nil

	case "abci_query":
		return handleQueryABCI(msg)

	default:
		return nil, fmt.Errorf("unexpected method: %v", msg.Method)
	}
}

func handleQueryABCI(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	var params []json.RawMessage
	err := json.Unmarshal(msg.Params, &params)
	if err != nil {
		return nil, err
	}

	if len(params) != 4 {
		return nil, fmt.Errorf("incorrect length of params array: %v", len(params))
	}

	var path string
	err = json.Unmarshal(params[0], &path)
	if err != nil {
		return nil, err
	}

	if path == "/custom/service/request" {
		return handleQueryServiceRequest(msg)
	}

	return []JsonrpcMessage{
		{
			ID:     msg.ID,
			Result: []byte{},
		},
	}, nil
}

func handleQueryServiceRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	msg.Method = "abci_query_service_request"

	rsp, ok := GetCannedResponse("birita", msg)
	if !ok {
		return nil, fmt.Errorf("failed to handle BSN-IRITA request for service request query")
	}

	return rsp, nil
}
