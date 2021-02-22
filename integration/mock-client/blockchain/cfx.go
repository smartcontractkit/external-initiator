package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func handleCfxRequest(conn string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	if conn == "ws" {
		switch msg.Method {
		case "cfx_subscribe":
			return handleCfxSubscribe(msg)
		}
	} else {
		switch msg.Method {
		case "cfx_getLogs":
			return handleCfxGetLogs(msg)
		}
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
}

type cfxSubscribeResponse struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

func handleCfxMapStringInterface(in map[string]json.RawMessage) (cfxLogResponse, error) {
	topics, err := getCfxTopicsFromMap(in)
	if err != nil {
		return cfxLogResponse{}, err
	}

	var topicsStr []string
	if len(topics) > 0 {
		for _, t := range topics[0] {
			topicsStr = append(topicsStr, t.String())
		}
	}

	addresses, err := getCfxAddressesFromMap(in)
	if err != nil {
		return cfxLogResponse{}, err
	}

	return cfxLogResponse{
		LogIndex:         "0x0",
		EpochNumber:      "0x2",
		BlockHash:        "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionHash:  "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionIndex: "0x0",
		Address:          addresses[0],
		Data:             "0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
		Topics:           topicsStr,
	}, nil
}

func handleCfxSubscribe(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
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

	log, err := handleCfxMapStringInterface(filter)
	if err != nil {
		return nil, err
	}

	logBz, err := json.Marshal(log)
	if err != nil {
		return nil, err
	}

	subResp := cfxSubscribeResponse{
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
			Method:  "cfx_subscribe",
		},
		{
			Version: "2.0",
			ID:      msg.ID,
			Method:  "cfx_subscribe",
			Params:  subBz,
		},
	}, nil
}

type cfxLogResponse struct {
	LogIndex         string   `json:"logIndex"`
	EpochNumber      string   `json:"epochNumber"`
	BlockHash        string   `json:"blockHash"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

func getCfxTopicsFromMap(req map[string]json.RawMessage) ([][]common.Hash, error) {
	topicsInterface, ok := req["topics"]
	if !ok {
		return nil, errors.New("no topics included")
	}

	var topicsArr []*[]string
	err := json.Unmarshal(topicsInterface, &topicsArr)
	if err != nil {
		return nil, err
	}

	var finalTopics [][]common.Hash
	for _, t := range topicsArr {
		if t == nil {
			continue
		}

		topics := make([]common.Hash, len(*t))
		for i, s := range *t {
			topics[i] = common.HexToHash(s)
		}

		finalTopics = append(finalTopics, topics)
	}

	return finalTopics, nil
}

func getCfxAddressesFromMap(req map[string]json.RawMessage) ([]string, error) {
	addressesInterface, ok := req["address"]
	if !ok {
		return nil, errors.New("no addresses included")
	}

	var addresses []string
	err := json.Unmarshal(addressesInterface, &addresses)
	if err != nil {
		return nil, err
	}

	if len(addresses) < 1 {
		return nil, errors.New("no addresses provided")
	}

	return addresses, nil
}

func cfxLogRequestToResponse(msg JsonrpcMessage) (cfxLogResponse, error) {
	var reqs []map[string]json.RawMessage
	err := json.Unmarshal(msg.Params, &reqs)
	if err != nil {
		return cfxLogResponse{}, err
	}

	if len(reqs) != 1 {
		return cfxLogResponse{}, fmt.Errorf("expected exactly 1 filter in request, got %d", len(reqs))
	}

	return handleCfxMapStringInterface(reqs[0])
}

func handleCfxGetLogs(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	log, err := cfxLogRequestToResponse(msg)
	if err != nil {
		return nil, err
	}

	logs := []cfxLogResponse{log}
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
