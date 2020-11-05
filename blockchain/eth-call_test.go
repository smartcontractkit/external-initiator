package blockchain

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

var testAbi abi.ABI

func TestMain(m *testing.M) {
	var err error
	abiJson := []byte(`[
	{
		"inputs":[],
		"name": "empty",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs":[],
		"name": "boolean",
		"outputs": [
			{
				"internalType": "bool",
				"name": "_value",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs":[],
		"name": "addresses",
		"outputs": [
			{
				"internalType": "address[]",
				"name": "_value",
				"type": "address[]"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs":[
			{
				"internalType": "address[]",
				"name": "_value",
				"type": "address[]"
			}
		],
		"name": "addressesTwo",
		"outputs": [
			{
				"internalType": "address[]",
				"name": "_value",
				"type": "address[]"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`)
	testAbi, err = abi.JSON(bytes.NewReader(abiJson))
	if err != nil {
		panic(err)
	}

	m.Run()
}

func Test_ethCallSubscriber_GetTestJson(t *testing.T) {
	ethCall := ethCallSubscriber{}
	got := ethCall.GetTestJson()
	want := []byte(`{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}`)
	if !bytes.Equal(got, want) {
		t.Errorf("GetTestJson() got = %s, want %s", got, want)
	}
}

func Test_ethCallSubscriber_ParseTestResponse(t *testing.T) {
	tests := []struct {
		name    string
		resp    []byte
		wantErr bool
	}{
		{
			"valid response",
			[]byte(`{"jsonrpc":"2.0","online":true}`),
			false,
		},
		{
			"empty response",
			[]byte{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := ethCallSubscriber{}
			if err := ethCall.ParseTestResponse(tt.resp); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ethCallSubscription_getCallPayload(t *testing.T) {
	address := common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
	emptyFuncSelect, err := testAbi.Pack("empty")
	require.NoError(t, err)

	type fields struct {
		address common.Address
		abi     abi.ABI
		method  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			"Empty payload",
			fields{},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x0000000000000000000000000000000000000000","data":"0x"},"latest"]}`),
			false,
		},
		{
			"With address",
			fields{
				address: common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
			},
			[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"%s","data":"0x"},"latest"]}`, address.String())),
			false,
		},
		{
			"With address and function selector",
			fields{
				address: common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
				abi:     testAbi,
				method:  "empty",
			},
			[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"%s","data":"%s"},"latest"]}`, address.String(), hexutil.Encode(emptyFuncSelect[:]))),
			false,
		},
		{
			"Fails on unknown method",
			fields{
				address: address,
				abi:     testAbi,
				method:  "doesNotExist",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := ethCallSubscription{
				address: tt.fields.address,
				abi:     tt.fields.abi,
				method:  tt.fields.method,
			}
			got, err := ethCall.getCallPayload()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCallPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCallPayload() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_ethCallSubscription_getSubscribePayload(t *testing.T) {
	tests := []struct {
		name    string
		want    []byte
		wantErr bool
	}{
		{
			"Subscribes to newHeads",
			[]byte(`{"jsonrpc":"2.0","id":2,"method":"eth_subscribe","params":["newHeads"]}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := ethCallSubscription{}
			got, err := ethCall.getSubscribePayload()
			if (err != nil) != tt.wantErr {
				t.Errorf("getSubscribePayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSubscribePayload() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ethCallSubscription_parseResponse(t *testing.T) {
	type fields struct {
		abi    abi.ABI
		method string
	}
	tests := []struct {
		name     string
		fields   fields
		response JsonrpcMessage
		want     []subscriber.Event
		wantErr  bool
	}{
		{
			"No events",
			fields{
				abi:    testAbi,
				method: "empty",
			},
			JsonrpcMessage{
				Result: []byte(`"0x"`),
			},
			nil,
			false,
		},
		{
			"Bool false",
			fields{
				abi:    testAbi,
				method: "boolean",
			},
			JsonrpcMessage{
				Result: []byte(fmt.Sprintf(`"%s"`, common.HexToHash("0").String())),
			},
			nil,
			false,
		},
		{
			"Bool true",
			fields{
				abi:    testAbi,
				method: "boolean",
			},
			JsonrpcMessage{
				Result: []byte(fmt.Sprintf(`"%s"`, common.HexToHash("1").String())),
			},
			[]subscriber.Event{{}},
			false,
		},
		{
			"Empty address set",
			fields{
				abi:    testAbi,
				method: "addresses",
			},
			JsonrpcMessage{
				Result: []byte(fmt.Sprintf(`"%s"`, common.HexToHash("0").String())),
			},
			nil,
			false,
		},
		{
			"Filled address set",
			fields{
				abi:    testAbi,
				method: "addresses",
			},
			JsonrpcMessage{
				Result: []byte(`"0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"`),
			},
			[]subscriber.Event{
				[]byte(fmt.Sprintf(`{"result":"%s"}`, common.HexToAddress("0").String())),
				[]byte(fmt.Sprintf(`{"result":"%s"}`, common.HexToAddress("1").String())),
			},
			false,
		},
		{
			"Invalid ABI unpack",
			fields{
				abi:    testAbi,
				method: "addresses",
			},
			JsonrpcMessage{
				Result: []byte(`"0x"`),
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := ethCallSubscription{
				abi:    tt.fields.abi,
				method: tt.fields.method,
				key:    defaultResponseKey,
			}
			got, err := ethCall.parseResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResponse() got = %s, want %s", got, tt.want)
			}
		})
	}
}
