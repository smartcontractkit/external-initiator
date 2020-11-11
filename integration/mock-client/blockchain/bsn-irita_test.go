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

	rsp, err := handleBSNIritaRequest(req)
	assert.NoError(t, err)

	var status tmtypes.SyncInfo
	err = json.Unmarshal(rsp[0].Result, &status)
	assert.NoError(t, err)
	assert.Equal(t, status.LatestBlockHeight, 1)
}

func TestHandleQueryBlockResults(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "query_blockResults",
	}

	rsp, err := handleBSNIritaRequest(req)
	assert.NoError(t, err)

	var blockResult tmtypes.ResultBlockResults
	err = json.Unmarshal(rsp[0].Result, &blockResult)
	assert.NoError(t, err)
	assert.Equal(t, blockResult.Height, 1)
}

func TestHandleQueryServiceRequest(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "query_serviceRequest",
	}

	rsp, err := handleBSNIritaRequest(req)
	assert.NoError(t, err)

	var request service.Request
	err = json.Unmarshal(rsp[0].Result, &request)
	assert.NoError(t, err)
	assert.Equal(t, request.ServiceName, "oracle")
}
