package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func handleCfxRequest(conn string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "cfx_getLogs":
		return handleCfxGetLogs(msg)
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
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
		EpochNumber:      "0x1",
		BlockHash:        "0x0",
		TransactionHash:  "0x0",
		TransactionIndex: "0x0",
		Address:          addresses[0].String(),
		Data:             "0x0",
		Topics:           topicsStr,
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

func getCfxAddressesFromMap(req map[string]json.RawMessage) ([]common.Address, error) {
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

func cfxLogRequestToResponse(msg JsonrpcMessage) (cfxLogResponse, error) {
	var reqs []map[string]json.RawMessage
	err := json.Unmarshal(msg.Params, &reqs)
	if err != nil {
		return cfxLogResponse{}, err
	}

	if len(reqs) != 1 {
		return cfxLogResponse{}, fmt.Errorf("Expected exactly 1 filter in request, got %d", len(reqs))
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
