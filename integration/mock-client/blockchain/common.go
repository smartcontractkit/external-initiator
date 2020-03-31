package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
)

type JsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *interface{}    `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func HandleRequest(conn, platform string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch platform {
	case "eth":
		return HandleEthRequest(conn, msg)
	default:
		return nil, errors.New(fmt.Sprint("unexpected platform: ", platform))
	}
}
