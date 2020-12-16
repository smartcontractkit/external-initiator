package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func handleKeeperRequest(_ string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "eth_call":
		return handleEthCall(msg)
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
}

type ethCallMessage struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to"`
	Gas      string `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Value    string `json:"value,omitempty"`
	Data     string `json:"data,omitempty"`
}

func msgToEthCall(msg JsonrpcMessage) (*ethCallMessage, error) {
	var params []json.RawMessage
	err := json.Unmarshal(msg.Params, &params)
	if err != nil {
		return nil, err
	}

	if len(params) != 2 {
		return nil, errors.New("unexpected amount of params")
	}

	var ethCall ethCallMessage
	err = json.Unmarshal(params[0], &ethCall)
	if err != nil {
		return nil, err
	}
	return &ethCall, nil
}

func handleEthCall(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	data, err := msgToEthCall(msg)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(data.Data, "0xb7d06888") {
		return nil, errors.New("unknown function selector")
	}

	return []JsonrpcMessage{
		{
			Version: "2.0",
			ID:      msg.ID,
			Result:  []byte(`"0x000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"`),
		},
	}, nil
}
