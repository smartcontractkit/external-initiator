package evm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func Test_toFilterArg(t *testing.T) {
	type args struct {
		q FilterQuery
	}

	blockHash := common.HexToHash("abc")

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cannot specify both Blockhash and FromBlock",
			args{FilterQuery{
				BlockHash: &blockHash,
				FromBlock: hexutil.EncodeBig(big.NewInt(3234512922)),
			}},
			true,
		},
		{
			"cannot specify both Blockhash and ToBlock",
			args{FilterQuery{
				BlockHash: &blockHash,
				ToBlock:   hexutil.EncodeBig(big.NewInt(3234512922)),
			}},
			true,
		},
		{
			"regular query passes",
			args{FilterQuery{
				Addresses: []common.Address{},
				Topics:    [][]common.Hash{},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.args.q.ToMapInterface()
			if (err != nil) != tt.wantErr {
				t.Errorf("toFilterArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
