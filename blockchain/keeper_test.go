package blockchain

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

var testAbi abi.ABI
var abiJson = []byte(`[
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
		"name": "bytes32",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "_value",
				"type": "bytes32"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs":[],
		"name": "int256",
		"outputs": [
			{
				"internalType": "int256",
				"name": "_value",
				"type": "int256"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs":[],
		"name": "address",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
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

func TestMain(m *testing.M) {
	var err error
	testAbi, err = abi.JSON(bytes.NewReader(abiJson))
	if err != nil {
		panic(err)
	}

	m.Run()
}

func newAbiFunctionHelper(t *testing.T, method string) solFunctionHelper {
	helper, err := NewSolFunctionHelper(abiJson, method, models.FunctionSelector{}, "")
	require.NoError(t, err)
	return *helper
}

func newFunctionHelper(t *testing.T, functionSelector models.FunctionSelector, returnType string) solFunctionHelper {
	helper, err := NewSolFunctionHelper(nil, "", functionSelector, returnType)
	require.NoError(t, err)
	return *helper
}

func Test_ethCallSubscriber_GetTestJson(t *testing.T) {
	ethCall := keeperSubscriber{}
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
			ethCall := keeperSubscriber{}
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
		helper  solFunctionHelper
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
			"Address and function selector from ABI and method name",
			fields{
				address: common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
				helper:  newAbiFunctionHelper(t, "empty"),
			},
			[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"%s","data":"%s"},"latest"]}`, address.String(), hexutil.Encode(emptyFuncSelect[:]))),
			false,
		},
		{
			"Address and function selector without ABI",
			fields{
				address: common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
				helper:  newFunctionHelper(t, models.HexToFunctionSelector(hexutil.Encode(emptyFuncSelect[:])), "bool"),
			},
			[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"%s","data":"%s"},"latest"]}`, address.String(), hexutil.Encode(emptyFuncSelect[:]))),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := keeperSubscription{
				address: tt.fields.address,
				helper:  tt.fields.helper,
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
			ethCall := keeperSubscription{}
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
	bytes32 := StringToBytes32("9adf89a689154509a053d6e3383304b5")

	type fields struct {
		helper solFunctionHelper
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
				helper: newAbiFunctionHelper(t, "empty"),
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
				helper: newAbiFunctionHelper(t, "boolean"),
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
				helper: newAbiFunctionHelper(t, "boolean"),
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
				helper: newAbiFunctionHelper(t, "addresses"),
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
				helper: newAbiFunctionHelper(t, "addresses"),
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
				helper: newAbiFunctionHelper(t, "addresses"),
			},
			JsonrpcMessage{
				Result: []byte(`"0x"`),
			},
			nil,
			true,
		},
		{
			"An address value",
			fields{
				helper: newAbiFunctionHelper(t, "address"),
			},
			JsonrpcMessage{
				Result: []byte(`"0x0000000000000000000000000000000000000000000000000000000000000000"`),
			},
			[]subscriber.Event{
				[]byte(fmt.Sprintf(`{"result":"%s"}`, common.HexToAddress("0").String())),
			},
			false,
		},
		{
			"A bytes32 value",
			fields{
				helper: newAbiFunctionHelper(t, "bytes32"),
			},
			JsonrpcMessage{
				Result: []byte(fmt.Sprintf(`"%s"`, bytes32.Hex())),
			},
			[]subscriber.Event{
				[]byte(fmt.Sprintf(`{"result":"%s"}`, bytes32.Hex())),
			},
			false,
		},
		{
			"An address value without ABI",
			fields{
				helper: newFunctionHelper(t, models.FunctionSelector{}, "address"),
			},
			JsonrpcMessage{
				Result: []byte(`"0x0000000000000000000000000000000000000000000000000000000000000000"`),
			},
			[]subscriber.Event{
				[]byte(fmt.Sprintf(`{"result":"%s"}`, common.HexToAddress("0").String())),
			},
			false,
		},
		{
			"A bytes32 value without ABI",
			fields{
				helper: newFunctionHelper(t, models.FunctionSelector{}, "bytes32"),
			},
			JsonrpcMessage{
				Result: []byte(fmt.Sprintf(`"%s"`, bytes32.Hex())),
			},
			[]subscriber.Event{
				[]byte(fmt.Sprintf(`{"result":"%s"}`, bytes32.Hex())),
			},
			false,
		},
		{
			"Filled address set without ABI",
			fields{
				helper: newFunctionHelper(t, models.FunctionSelector{}, "address[]"),
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
			"An int256 value without ABI",
			fields{
				helper: newFunctionHelper(t, models.FunctionSelector{}, "int256"),
			},
			JsonrpcMessage{
				Result: []byte(`"0x0000000000000000000000000000000000000000000000000000000a9d93b118"`),
			},
			[]subscriber.Event{
				[]byte(`{"result":45593375000}`),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := keeperSubscription{
				helper: tt.fields.helper,
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

func Test_keeperSubscription_cooldownDone(t *testing.T) {
	type fields struct {
		cooldown         *big.Int
		lastInitiatedRun *big.Int
	}
	type args struct {
		blockHeight *big.Int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"continues if cooldown has passed",
			fields{
				cooldown:         big.NewInt(1),
				lastInitiatedRun: big.NewInt(1),
			},
			args{
				blockHeight: big.NewInt(2),
			},
			true,
		},
		{
			"continues if cooldown has passed by a large amount",
			fields{
				cooldown:         big.NewInt(1),
				lastInitiatedRun: big.NewInt(1),
			},
			args{
				blockHeight: big.NewInt(1000),
			},
			true,
		},
		{
			"continues if there is no cooldown",
			fields{
				cooldown:         big.NewInt(0),
				lastInitiatedRun: big.NewInt(1),
			},
			args{
				blockHeight: big.NewInt(1),
			},
			true,
		},
		{
			"waits if cooldown has not completed",
			fields{
				cooldown:         big.NewInt(2),
				lastInitiatedRun: big.NewInt(1),
			},
			args{
				blockHeight: big.NewInt(2),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := keeperSubscription{
				cooldown:         tt.fields.cooldown,
				lastInitiatedRun: tt.fields.lastInitiatedRun,
			}
			if got := keeper.cooldownDone(tt.args.blockHeight); got != tt.want {
				t.Errorf("cooldownDone() = %v, want %v", got, tt.want)
			}
		})
	}
}
