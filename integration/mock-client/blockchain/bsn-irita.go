package blockchain

import (
	"fmt"
)

func handleBSNIritaRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "query_status":
		return GetCannedResponses("birita", msg)

	case "query_blockResults":
		return GetCannedResponses("birita", msg)

	case "query_serviceRequest":
		return GetCannedResponses("birita", msg)

	default:
		return nil, fmt.Errorf("unexpected method: %v", msg.Method)
	}
}
