package evm

import (
	"math/big"
	"reflect"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var emptyAddressSet []eth.Address
var emptyTopicSet [][]eth.Hash

func Test_toFilterArg(t *testing.T) {
	type args struct {
		q FilterQuery
	}

	blockHash := eth.HexToHash("abc")

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
				Addresses: []eth.Address{},
				Topics:    [][]eth.Hash{},
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

func TestBytesToHex(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Adds 0x prefix with leading 0",
			args{[]byte{0}},
			"0x00",
		},
		{
			"Bigger numbers have even length",
			args{[]byte{255, 255}},
			"0xffff",
		},
		{
			"Converts string data",
			args{[]byte(`"test"`)},
			"0x227465737422",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesToHex(tt.args.data); got != tt.want {
				t.Errorf("BytesToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateEvmFilterQuery(t *testing.T) {
	type args struct {
		jobid        string
		strAddresses []string
	}
	tests := []struct {
		name string
		args args
		want *FilterQuery
	}{
		{
			"creates a basic filter query with jobid and addresses",
			args{"abc123", []string{"0x123", "0xabc"}},
			&FilterQuery{
				Addresses: []eth.Address{
					eth.HexToAddress("0x123"),
					eth.HexToAddress("0xabc"),
				},
				Topics: [][]eth.Hash{
					{eth.HexToHash("0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65")},
					{eth.HexToHash("0x6162633132330000000000000000000000000000000000000000000000000000")},
				},
			},
		},
		{
			"supports no jobid or addresses",
			args{"", []string{}},
			&FilterQuery{
				Addresses: emptyAddressSet,
				Topics: [][]eth.Hash{
					{eth.HexToHash("0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65")},
					{eth.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateEvmFilterQuery(tt.args.jobid, tt.args.strAddresses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateEvmFilterQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterQuery_ToMapInterface(t *testing.T) {
	blockHash := eth.HexToHash("abc")

	type fields struct {
		BlockHash *eth.Hash
		FromBlock string
		ToBlock   string
		Addresses []eth.Address
		Topics    [][]eth.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		{
			"cannot specify both Blockhash and FromBlock",
			fields{
				BlockHash: &blockHash,
				FromBlock: hexutil.EncodeBig(big.NewInt(3234512922)),
			},
			nil,
			true,
		},
		{
			"cannot specify both Blockhash and ToBlock",
			fields{
				BlockHash: &blockHash,
				ToBlock:   hexutil.EncodeBig(big.NewInt(3234512922)),
			},
			nil,
			true,
		},
		{
			"empty query passes",
			fields{},
			map[string]interface{}{
				"address":   emptyAddressSet,
				"fromBlock": "0x0",
				"toBlock":   "latest",
				"topics":    emptyTopicSet,
			},
			false,
		},
		{
			"can specify block hash",
			fields{
				BlockHash: &blockHash,
			},
			map[string]interface{}{
				"address":   emptyAddressSet,
				"blockHash": blockHash,
				"topics":    emptyTopicSet,
			},
			false,
		},
		{
			"can specify from and to block",
			fields{
				FromBlock: "0x123",
				ToBlock:   "0x1234",
			},
			map[string]interface{}{
				"address":   emptyAddressSet,
				"fromBlock": "0x123",
				"toBlock":   "0x1234",
				"topics":    emptyTopicSet,
			},
			false,
		},
		{
			"can specify address and topics",
			fields{
				Addresses: []eth.Address{
					eth.HexToAddress("0x123"),
				},
				Topics: [][]eth.Hash{
					{eth.HexToHash("0xabc")},
					{eth.HexToHash("0x123")},
				},
			},
			map[string]interface{}{
				"address": []eth.Address{
					eth.HexToAddress("0x123"),
				},
				"fromBlock": "0x0",
				"toBlock":   "latest",
				"topics": [][]eth.Hash{
					{eth.HexToHash("0xabc")},
					{eth.HexToHash("0x123")},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := FilterQuery{
				BlockHash: tt.fields.BlockHash,
				FromBlock: tt.fields.FromBlock,
				ToBlock:   tt.fields.ToBlock,
				Addresses: tt.fields.Addresses,
				Topics:    tt.fields.Topics,
			}
			got, err := q.ToMapInterface()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMapInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToMapInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToBytes32(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want eth.Hash
	}{
		{
			"right-pads a simple string",
			args{"abc123"},
			eth.HexToHash("0x6162633132330000000000000000000000000000000000000000000000000000"),
		},
		{
			"empty string gives empty hash",
			args{""},
			eth.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToBytes32(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToBytes32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ParseBlockNumberResult(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			"parses a correct response",
			args{[]byte(`"0x01"`)},
			1,
			false,
		},
		{
			"parses even if there's no leading 0",
			args{[]byte(`"0x1"`)},
			1,
			false,
		},
		{
			"fails on invalid payload",
			args{[]byte(`"abc"`)},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBlockNumberResult(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBlockNumberResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseBlockNumberResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}
