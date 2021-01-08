package blockchain

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
)

type MockIoTeXServer struct {
	iotexapi.UnimplementedAPIServiceServer
}

func (*MockIoTeXServer) GetChainMeta(context.Context, *iotexapi.GetChainMetaRequest) (*iotexapi.GetChainMetaResponse, error) {
	return &iotexapi.GetChainMetaResponse{
		ChainMeta: &iotextypes.ChainMeta{
			Height: 1000,
		},
	}, nil
}

func (*MockIoTeXServer) GetLogs(_ context.Context, req *iotexapi.GetLogsRequest) (*iotexapi.GetLogsResponse, error) {
	addresses := req.GetFilter().GetAddress()
	var contract string
	if len(addresses) != 0 {
		contract = addresses[0]
	}

	var topic [][]byte
	topics := req.GetFilter().Topics
	if len(topics) != 0 {
		topic = topics[0].GetTopic()
	}

	data, err := hexToBytes("0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864")
	if err != nil {
		return nil, err
	}
	return &iotexapi.GetLogsResponse{
		Logs: []*iotextypes.Log{
			{
				ContractAddress: contract,
				Topics:          topic,
				Data:            data,
				BlkHeight:       req.GetByRange().GetFromBlock(),
				Index:           uint32(0),
			},
		},
	}, nil
}

func hexToBytes(x string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(x, "0x"))
}
