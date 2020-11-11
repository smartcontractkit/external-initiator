package blockchain

import (
	"fmt"
)

func handleBSNIritaRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "query_status":
		return handleQueryStatus(msg)

	case "query_blockResults":
		return handleQueryBlockResults(msg)

	case "query_serviceRequest":
		return handleQueryServiceRequest(msg)

	default:
		return nil, fmt.Errorf("unexpected method: %v", msg.Method)
	}
}

func handleQueryStatus(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	return GetCannedResponses("birita", msg)
}

func handleQueryBlockResults(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	return GetCannedResponses("birita", msg)
}

func handleQueryServiceRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	return GetCannedResponses("birita", msg)
}
