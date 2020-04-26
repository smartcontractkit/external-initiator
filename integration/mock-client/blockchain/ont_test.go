package blockchain

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SmartContactEvent struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []*NotifyEventInfo
}

func TestHandleGetSmartCodeEvent(t *testing.T) {
	req := JsonRpcRequest{
		Version: "2.0",
		Id: "1",
		Method: "getsmartcodeevent",
	}

	rsp, err := HandleOntRequest(req)
	assert.NoError(t, err)
	events := make([]*SmartContactEvent, 0)
	err = json.Unmarshal(rsp.Result, &events)
	assert.NoError(t, err)
}

func TestHandleGetBlockCount(t *testing.T) {
	req := JsonRpcRequest{
		Version: "2.0",
		Id: "1",
		Method: "getblockcount",
	}

	rsp, err := HandleOntRequest(req)
	assert.NoError(t, err)
	var count uint32
	err = json.Unmarshal(rsp.Result, &count)
	assert.NoError(t, err)
	assert.Equal(t, count, uint32(1))
}