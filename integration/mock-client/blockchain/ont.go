package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

func handleOntRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "getblockcount":
		return handleGetBlockCount(msg)
	case "getsmartcodeevent":
		return handleGetSmartCodeEvent(msg)
	}

	return nil, errors.New(fmt.Sprint("unexpected method: ", msg.Method))
}

func handleGetBlockCount(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	r, _ := json.Marshal(uint32(1))
	return []JsonrpcMessage{
		{
			ID:     msg.ID,
			Result: r,
		},
	}, nil
}

type executeNotify struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []notifyEventInfo
}

type notifyEventInfo struct {
	ContractAddress string
	States          interface{}
}

func handleGetSmartCodeEvent(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	eInfos := make([]*executeNotify, 0)
	nEI := notifyEventInfo{
		ContractAddress: "0x2aD9B7b9386c2f45223dDFc4A4d81C2957bAE19A",
		States: []interface{}{hex.EncodeToString([]byte("oracleRequest")), "mock"},
	}
	eInfo := &executeNotify{
		Notify: []notifyEventInfo{nEI},
	}
	eInfos = append(eInfos, eInfo)

	data, err := json.Marshal(eInfos)
	if err != nil {
		return nil, err
	}

	return []JsonrpcMessage{
		{
			ID:     msg.ID,
			Result: data,
		},
	}, nil
}
