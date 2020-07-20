package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func handleOntRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "getsmartcodeevent":
		return handleGetSmartCodeEvent(msg)
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
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
		States: []interface{}{hex.EncodeToString([]byte("oracleRequest")), "mock", "01", "02", "03", "04",
			"05", "06", "07", "", "08"},
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
