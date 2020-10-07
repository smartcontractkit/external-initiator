package blockchain

import (
	"context"
	"testing"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIoTeXMockServer(t *testing.T) {
	serv := &MockIoTeXServer{}
	t.Run("iotexGetChainMeta", func(t *testing.T) {
		ctx := context.Background()
		resp, err := serv.GetChainMeta(ctx, &iotexapi.GetChainMetaRequest{})
		require.NoError(t, err)
		assert.Equal(t, uint64(1000), resp.GetChainMeta().GetHeight())
	})

	t.Run("iotexGetLogs", func(t *testing.T) {
		ctx := context.Background()
		contract := "io12345678"
		height := uint64(1000)
		req := &iotexapi.GetLogsRequest{
			Filter: &iotexapi.LogsFilter{
				Address: []string{contract},
			},
			Lookup: &iotexapi.GetLogsRequest_ByRange{
				ByRange: &iotexapi.GetLogsByRange{
					FromBlock: height,
					Count:     1,
				},
			},
		}
		resp, err := serv.GetLogs(ctx, req)
		require.NoError(t, err)
		assert.NotZero(t, len(resp.GetLogs()))
		log := resp.GetLogs()[0]
		assert.Equal(t, contract, log.GetContractAddress())
		assert.Equal(t, height, log.GetBlkHeight())
		assert.NotNil(t, log.GetData())
	})
}
