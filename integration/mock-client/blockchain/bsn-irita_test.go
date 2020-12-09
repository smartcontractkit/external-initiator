package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irisnet/service-sdk-go/service"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
)

func TestHandleQueryStatus(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "status",
	}

	rsp, ok := GetCannedResponse("birita", req)
	assert.True(t, ok)

	var status tmtypes.ResultStatus
	err := tmjson.Unmarshal(rsp[0].Result, &status)
	assert.NoError(t, err)
	assert.Equal(t, int64(7753), status.SyncInfo.LatestBlockHeight)
}

func TestHandleQueryBlockResults(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "block_results",
	}

	rsp, ok := GetCannedResponse("birita", req)
	assert.True(t, ok)

	var blockResult tmtypes.ResultBlockResults
	err := tmjson.Unmarshal(rsp[0].Result, &blockResult)
	assert.NoError(t, err)
	assert.Equal(t, int64(7753), blockResult.Height)
}

func TestHandleQueryServiceRequest(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "abci_query",
		Params:  []byte(`["/custom/service/request",{"request_id":"1"},"0",false]`),
	}

	rsp, err := handleQueryABCI(req)
	assert.NoError(t, err)

	var request service.Request
	err = tmjson.Unmarshal(rsp[0].Result, &request)
	assert.NoError(t, err)
	assert.Equal(t, "oracle", request.ServiceName)
	assert.Equal(t, "iaa1l4vp69jt8ghxtyrh6jm8jp022km50sg35eqcae", request.Provider.String())
}
