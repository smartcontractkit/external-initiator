package blockchain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irisnet/service-sdk-go/service"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
)

func TestHandleQueryStatus(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "query_status",
	}

	rsp, ok := GetCannedResponse("birita", req)
	assert.True(t, ok)

	var status tmtypes.SyncInfo
	err = json.Unmarshal(rsp[0].Result, &status)
	assert.NoError(t, err)
	assert.Equal(t, status.LatestBlockHeight, 7753)
}

func TestHandleQueryBlockResults(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "query_blockResults",
	}

	rsp, ok := GetCannedResponse("birita", req)
	assert.True(t, ok)

	var blockResult tmtypes.ResultBlockResults
	err = json.Unmarshal(rsp[0].Result, &blockResult)
	assert.NoError(t, err)
	assert.Equal(t, blockResult.Height, 7753)
}

func TestHandleQueryServiceRequest(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "query_serviceRequest",
	}

	rsp, ok := GetCannedResponse("birita", req)
	assert.True(t, ok)

	var request service.Request
	err = json.Unmarshal(rsp[0].Result, &request)
	assert.NoError(t, err)
	assert.Equal(t, "oracle", request.ServiceName)
	assert.Equal(t, "iaa1l4vp69jt8ghxtyrh6jm8jp022km50sg35eqcae", request.Provider)
}
