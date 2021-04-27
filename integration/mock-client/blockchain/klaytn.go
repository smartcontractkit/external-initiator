package blockchain

import (
	"encoding/json"
	"fmt"
)

// handleKlaytnRequest handles Klaytn request.
// It is different from eth that it uses 'klay' instead of 'eth' in method.
func handleKlaytnRequest(conn string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	if conn == "ws" {
		switch msg.Method {
		case "klay_subscribe":
			return handleKlaytnSubscribe(msg)
		}
	} else {
		switch msg.Method {
		case "klay_getLogs":
			return handleEthGetLogs(msg)
		}
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
}

func handleKlaytnSubscribe(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	var contents []json.RawMessage
	err := json.Unmarshal(msg.Params, &contents)
	if err != nil {
		return nil, err
	}

	if len(contents) != 2 {
		return nil, fmt.Errorf("possibly incorrect length of params array: %v", len(contents))
	}

	var filter map[string]json.RawMessage
	err = json.Unmarshal(contents[1], &filter)
	if err != nil {
		return nil, err
	}

	log, err := handleMapStringInterface(filter)
	if err != nil {
		return nil, err
	}

	logBz, err := json.Marshal(log)
	if err != nil {
		return nil, err
	}

	subResp := ethSubscribeResponse{
		Subscription: "test",
		Result:       logBz,
	}

	subBz, err := json.Marshal(subResp)
	if err != nil {
		return nil, err
	}

	return []JsonrpcMessage{
		// Send a confirmation message first
		// This is currently ignored, so don't fill
		{
			Version: "2.0",
			ID:      msg.ID,
			Method:  "klay_subscribe",
		},
		{
			Version: "2.0",
			ID:      msg.ID,
			Method:  "klay_subscribe",
			Params:  subBz,
		},
	}, nil
}
