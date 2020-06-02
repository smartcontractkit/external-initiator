package blockchain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type smartContactEvent struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []*notifyEventInfo
}

func TestHandleGetSmartCodeEvent(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "getsmartcodeevent",
	}

	rsp, err := handleOntRequest(req)
	assert.NoError(t, err)
	events := make([]*smartContactEvent, 0)
	err = json.Unmarshal(rsp[0].Result, &events)
	assert.NoError(t, err)
}

func TestHandleGetBlockCount(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "getblockcount",
	}

	rsp, ok := GetCannedResponse("ont", req)
	assert.True(t, ok)
	var count uint32
	err := json.Unmarshal(rsp[0].Result, &count)
	assert.NoError(t, err)
	assert.Equal(t, count, uint32(1))
}
