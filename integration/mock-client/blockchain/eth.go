package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func handleEthRequest(conn string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	if conn == "ws" {
		switch msg.Method {
		case "eth_subscribe":
			return handleEthSubscribe(msg)
		}
	} else {
		switch msg.Method {
		case "eth_getLogs":
			return handleEthGetLogs(msg)
		}
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
}

type ethSubscribeResponse struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

func handleMapStringInterface(in map[string]json.RawMessage) (ethLogResponse, error) {
	topics, err := getTopicsFromMap(in)
	if err != nil {
		return ethLogResponse{}, err
	}

	var topicsStr []string
	if len(topics) > 0 {
		for _, t := range topics[0] {
			topicsStr = append(topicsStr, t.String())
		}
	}

	addresses, err := getAddressesFromMap(in)
	if err != nil {
		return ethLogResponse{}, err
	}

	return ethLogResponse{
		LogIndex:         "0x0",
		BlockNumber:      "0x1",
		BlockHash:        "0x0",
		TransactionHash:  "0x0",
		TransactionIndex: "0x0",
		Address:          addresses[0].String(),
		Data:             "0x0",
		Topics:           topicsStr,
	}, nil
}

func handleEthSubscribe(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
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
			Method:  "eth_subscribe",
		},
		{
			Version: "2.0",
			ID:      msg.ID,
			Method:  "eth_subscribe",
			Params:  subBz,
		},
	}, nil
}

type ethLogResponse struct {
	LogIndex         string   `json:"logIndex"`
	BlockNumber      string   `json:"blockNumber"`
	BlockHash        string   `json:"blockHash"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

func getTopicsFromMap(req map[string]json.RawMessage) ([][]common.Hash, error) {
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

func getAddressesFromMap(req map[string]json.RawMessage) ([]common.Address, error) {
	addressesInterface, ok := req["address"]
	if !ok {
		return nil, errors.New("no addresses included")
	}

	var addresses []common.Address
	err := json.Unmarshal(addressesInterface, &addresses)
	if err != nil {
		return nil, err
	}

	if len(addresses) < 1 {
		return nil, errors.New("no addresses provided")
	}

	return addresses, nil
}

func ethLogRequestToResponse(msg JsonrpcMessage) (ethLogResponse, error) {
	var reqs []map[string]json.RawMessage
	err := json.Unmarshal(msg.Params, &reqs)
	if err != nil {
		return ethLogResponse{}, err
	}

	if len(reqs) != 1 {
		return ethLogResponse{}, fmt.Errorf("expected exactly 1 filter in request, got %d", len(reqs))
	}

	return handleMapStringInterface(reqs[0])
}

func handleEthGetLogs(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	log, err := ethLogRequestToResponse(msg)
	if err != nil {
		return nil, err
	}

	logs := []ethLogResponse{log}
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
