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
		case "eth_blockNumber":
			return handleEthBlockNumber(msg)
		case "eth_getLogs":
			return handleEthGetLogs(msg)
		}
	}

	return nil, errors.New(fmt.Sprint("unexpected method: ", msg.Method))
}

type ethSubscribeResponse struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

func handleMapStringInterface(in map[string]interface{}) (ethLogResponse, error) {
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
	var contents []interface{}
	err := json.Unmarshal(msg.Params, &contents)
	if err != nil {
		return nil, err
	}

	if len(contents) != 2 {
		return nil, errors.New(fmt.Sprint("possibly incorrect length of params array:", len(contents)))
	}

	req, ok := contents[1].(map[string]interface{})
	if !ok {
		return nil, errors.New("type cast to map[string]interface{} failed")
	}

	log, err := handleMapStringInterface(req)
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
			Version: msg.Version,
			ID:      msg.ID,
			Method:  "eth_subscribe",
		},
		{
			Version: msg.Version,
			ID:      msg.ID,
			Method:  "eth_subscribe",
			Params:  subBz,
		},
	}, nil
}

func handleEthBlockNumber(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	return []JsonrpcMessage{
		{
			Version: "2.0",
			ID:      msg.ID,
			Result:  []byte(`"0x0"`),
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

func getTopicsFromMap(req map[string]interface{}) ([][]common.Hash, error) {
	topicsInterface, ok := req["topics"]
	if !ok {
		return nil, errors.New("no topics included")
	}

	topicsArr, ok := topicsInterface.([]interface{})
	if !ok {
		return nil, errors.New("could not cast topics to []interface{}")
	}

	var finalTopics [][]common.Hash
	for _, t := range topicsArr {
		if t == nil {
			continue
		}

		topic, ok := t.(*[]string)
		if !ok || topic == nil {
			continue
		}

		topics := make([]common.Hash, len(*topic))
		for i, s := range *topic {
			topics[i] = common.HexToHash(s)
		}

		finalTopics = append(finalTopics, topics)
	}

	return finalTopics, nil
}

func getAddressesFromMap(req map[string]interface{}) ([]common.Address, error) {
	addressesInterface, ok := req["address"]
	if !ok {
		return nil, errors.New("no addresses included")
	}

	addressesIntf, ok := addressesInterface.([]interface{})
	if !ok {
		return nil, errors.New("could not cast addresses to []interface{}")
	}

	var addresses []common.Address
	for _, intf := range addressesIntf {
		str, ok := intf.(string)
		if !ok {
			return nil, errors.New("unable to cast into string")
		}

		addresses = append(addresses, common.HexToAddress(str))
	}

	if len(addresses) < 1 {
		return nil, errors.New("no addresses provided")
	}

	return addresses, nil
}

func ethLogRequestToResponse(msg JsonrpcMessage) (ethLogResponse, error) {
	var reqs []map[string]interface{}
	err := json.Unmarshal(msg.Params, &reqs)
	if err != nil {
		return ethLogResponse{}, err
	}

	if len(reqs) != 1 {
		return ethLogResponse{}, errors.New(fmt.Sprintf("Expected exactly 1 filter in request, got %d", len(reqs)))
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
