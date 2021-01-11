package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmrpc "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/smartcontractkit/external-initiator/blockchain"
)

func TestHandleQueryStatus(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "status",
	}

	rsp, ok := GetCannedResponse("birita", req)
	assert.True(t, ok)

	var status tmrpc.ResultStatus
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

	var blockResult tmrpc.ResultBlockResults
	err := tmjson.Unmarshal(rsp[0].Result, &blockResult)
	assert.NoError(t, err)
	assert.Equal(t, int64(7753), blockResult.Height)
}

func TestHandleQueryServiceRequest(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "abci_query",
		Params:  []byte(`{"path":"/custom/service/request","data:":"01","height":"0","prove":false}`),
	}

	rsp, err := handleQueryABCI(req)
	assert.NoError(t, err)

	var abciResponse tmrpc.ResultABCIQuery
	err = tmjson.Unmarshal(rsp[0].Result, &abciResponse)
	assert.NoError(t, err)

	var request blockchain.BIritaServiceRequest
	err = tmjson.Unmarshal(abciResponse.Response.Value, &request)
	assert.NoError(t, err)
	assert.Equal(t, "oracle", request.ServiceName)
	assert.Equal(t, "iaa1l4vp69jt8ghxtyrh6jm8jp022km50sg35eqcae", request.Provider)
}
