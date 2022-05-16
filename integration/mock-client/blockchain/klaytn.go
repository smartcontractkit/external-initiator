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
			return handleKlaytnGetLogs(msg)
		}
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
}

func handleKlaytnMapStringInterface(in map[string]json.RawMessage) (klaytnLogResponse, error) {
	topics, err := getTopicsFromMap(in)
	if err != nil {
		return klaytnLogResponse{}, err
	}

	var topicsStr []string
	if len(topics) > 0 {
		for _, t := range topics[0] {
			topicsStr = append(topicsStr, t.String())
		}
	}

	addresses, err := getAddressesFromMap(in)
	if err != nil {
		return klaytnLogResponse{}, err
	}

	return klaytnLogResponse{
		LogIndex:         "0x0",
		BlockNumber:      "0x2",
		BlockHash:        "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionHash:  "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionIndex: "0x0",
		Address:          addresses[0].String(),
		Data:             "0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
		Topics:           topicsStr,
	}, nil
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

	log, err := handleKlaytnMapStringInterface(filter)
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

type klaytnLogResponse struct {
	LogIndex         string   `json:"logIndex"`
	BlockNumber      string   `json:"blockNumber"`
	BlockHash        string   `json:"blockHash"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

func klaytnLogRequestToResponse(msg JsonrpcMessage) (klaytnLogResponse, error) {
	var reqs []map[string]json.RawMessage
	err := json.Unmarshal(msg.Params, &reqs)
	if err != nil {
		return klaytnLogResponse{}, err
	}

	if len(reqs) != 1 {
		return klaytnLogResponse{}, fmt.Errorf("expected exactly 1 filter in request, got %d", len(reqs))
	}

	r, err := handleKlaytnMapStringInterface(reqs[0])
	if err != nil {
		return klaytnLogResponse{}, err
	}
	return klaytnLogResponse(r), nil
}

func handleKlaytnGetLogs(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	log, err := klaytnLogRequestToResponse(msg)
	if err != nil {
		return nil, err
	}

	logs := []klaytnLogResponse{log}
	data, err := json.Marshal(logs)
	if err != nil {
		return nil, err
	}

	return []JsonrpcMessage{
		{
			Version: "2.0",
			ID:      msg.ID,
			Result:  data,
		},
	}, nil
}
