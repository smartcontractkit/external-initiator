package conflux

import (
	"reflect"
	"testing"

	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleAddress = "cfxtest:acdjv47k166p1pt4e8yph9rbcumrpbn2u69wyemxv0"

func Test_cfxFilterQuery_toMapInterface(t *testing.T) {
	type fields struct {
		BlockHash *eth.Hash
		FromEpoch string
		ToEpoch   string
		Addresses []cfxaddress.Address
		Topics    [][]eth.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		{
			"empty toEpoch becomes latest_state",
			fields{},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]eth.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "latest_state",
			},
			false,
		},
		{
			"uses non-empty toEpoch",
			fields{ToEpoch: "0x1"},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]eth.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "0x1",
			},
			false,
		},
		{
			"empty fromEpoch becomes 0x0",
			fields{},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]eth.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "latest_state",
			},
			false,
		},
		{
			"uses non-empty fromEpoch",
			fields{FromEpoch: "0x1"},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]eth.Hash{},
				"fromEpoch": "0x1",
				"toEpoch":   "latest_state",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := filterQuery{
				BlockHash: tt.fields.BlockHash,
				FromEpoch: tt.fields.FromEpoch,
				ToEpoch:   tt.fields.ToEpoch,
				Addresses: tt.fields.Addresses,
				Topics:    tt.fields.Topics,
			}
			got, err := q.toMapInterface()
			if (err != nil) != tt.wantErr {
				t.Errorf("toMapInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			keys := []string{"fromEpoch", "toEpoch"}
			for _, key := range keys {
				assert.Equal(t, got[key], tt.want[key])
			}
		})
	}
}

func Test_cfx2EthResponse(t *testing.T) {
	addr, err := cfxaddress.NewFromBase32(exampleAddress)
	require.NoError(t, err)

	type args struct {
		cfx cfxLogResponse
	}
	tests := []struct {
		name    string
		args    args
		want    models.Log
		wantErr bool
	}{
		{
			"parses a regular CFX log response",
			args{cfxLogResponse{
				LogIndex:         "0x0",
				EpochNumber:      "0x2",
				BlockHash:        eth.HexToHash("0xabc"),
				TransactionHash:  eth.HexToHash("0x123"),
				TransactionIndex: "0x0",
				Address:          addr,
				Data:             "0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
				Topics: []eth.Hash{
					eth.HexToHash("0xabc123"),
				},
			}},
			models.Log{
				Address: addr.MustGetCommonAddress(),
				Topics: []eth.Hash{
					eth.HexToHash("0xabc123"),
				},
				Data:        eth.Hex2Bytes("0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864"),
				BlockNumber: 2,
				TxHash:      eth.HexToHash("0x123"),
				TxIndex:     0,
				BlockHash:   eth.HexToHash("0xabc"),
				Index:       0,
				Removed:     false,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfx2EthResponse(tt.args.cfx)
			if (err != nil) != tt.wantErr {
				t.Errorf("cfx2EthResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cfx2EthResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createFilterQuery(t *testing.T) {
	addr, err := cfxaddress.NewFromBase32(exampleAddress)
	require.NoError(t, err)

	type args struct {
		jobid        string
		strAddresses []string
	}
	tests := []struct {
		name    string
		args    args
		want    filterQuery
		wantErr bool
	}{
		{
			"creates a standard filter query",
			args{"test123", []string{exampleAddress}},
			filterQuery{
				Addresses: []cfxaddress.Address{addr},
				Topics: [][]eth.Hash{
					{eth.HexToHash("0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65")},
					{eth.HexToHash("0x7465737431323300000000000000000000000000000000000000000000000000")},
				},
			},
			false,
		},
		{
			"supports empty jobid",
			args{"", []string{exampleAddress}},
			filterQuery{
				Addresses: []cfxaddress.Address{addr},
				Topics: [][]eth.Hash{
					{eth.HexToHash("0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65")},
					{eth.HexToHash("0x00")},
				},
			},
			false,
		},
		{
			"fails on invalid address",
			args{"test123", []string{"not a proper address"}},
			filterQuery{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createFilterQuery(tt.args.jobid, tt.args.strAddresses)
			if (err != nil) != tt.wantErr {
				t.Errorf("createFilterQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createFilterQuery() got = %v, want %v", got, tt.want)
			}
		})
	}
}
