package blockchain

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

var testAbi abi.ABI
var encoderAbi abi.ABI

const encodingAbi = `[
{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "performData",
				"type": "bytes"
			},
			{
				"internalType": "uint256",
				"name": "maxLinkPayment",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "gasLimit",
				"type": "uint256"
			},
			{
				"internalType": "int256",
				"name": "gasWei",
				"type": "int256"
			},
			{
				"internalType": "int256",
				"name": "linkEth",
				"type": "int256"
			}
		],
		"name": "checkUpkeepInverse",
		"outputs": [],
		"stateMutability": "view",
		"type": "function"
	}
]`

func TestMain(m *testing.M) {
	var err error
	testAbi, err = abi.JSON(bytes.NewReader([]byte(UpkeepRegistryInterface)))
	if err != nil {
		panic(err)
	}

	encoderAbi, err = abi.JSON(bytes.NewReader([]byte(encodingAbi)))
	if err != nil {
		panic(err)
	}

	m.Run()
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
	from := common.HexToAddress("0xfff5a916Ea1049f591B07B1B62d800d16630002e")
	data, err := testAbi.Pack(checkMethod, big.NewInt(123), from)
	require.NoError(t, err)
	dataHex := bytesToHex(data)
	t.Run("ABI packs correct data string", func(t *testing.T) {
		expected := "0xc41b813a000000000000000000000000000000000000000000000000000000000000007b000000000000000000000000fff5a916ea1049f591b07b1b62d800d16630002e"
		assert.Equal(t, expected, dataHex)
	})

	type fields struct {
		address common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			"Empty address",
			fields{},
			[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x0000000000000000000000000000000000000000","data":"%s"},"latest"]}`, dataHex)),
			false,
		},
		{
			"With address",
			fields{
				address: common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
			},
			[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"%s","data":"%s"},"latest"]}`, address.String(), dataHex)),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := keeperSubscription{
				address:  tt.fields.address,
				from:     from,
				abi:      testAbi,
				upkeepId: big.NewInt(123),
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
	encodedData, err := encoderAbi.Pack("checkUpkeepInverse", []byte("sample data"), big.NewInt(123), big.NewInt(123), big.NewInt(123), big.NewInt(123))
	require.NoError(t, err)
	require.True(t, len(encodedData) > 4)
	encodedData = encodedData[4:] // Remove function selector to just get the data
	errMessage := interface{}("error message")

	tests := []struct {
		name     string
		response JsonrpcMessage
		want     []subscriber.Event
		wantErr  bool
	}{
		{
			"Error",
			JsonrpcMessage{
				Result: []byte{},
				Error:  &errMessage,
			},
			nil,
			false,
		},
		{
			"Empty response",
			JsonrpcMessage{
				Result: []byte(`"0x"`),
			},
			nil,
			true,
		},
		{
			"Invalid response",
			JsonrpcMessage{
				Result: []byte(`"0x12341234"`),
			},
			nil,
			true,
		},
		{
			"UpkeepRegistryInterface unpack with bytes",
			JsonrpcMessage{
				Result: []byte(fmt.Sprintf(`"%s"`, bytesToHex(encodedData))),
			},
			[]subscriber.Event{
				[]byte(`{"address":"0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE","format":"preformatted","fromAddresses":["0x0000000000000000000000000000000000000123"],"functionSelector":"0x7bbaf1ea","result":"0x000000000000000000000000000000000000000000000000000000000000007b0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000b73616d706c652064617461000000000000000000000000000000000000000000"}`),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ethCall := keeperSubscription{
				address:  common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
				abi:      testAbi,
				from:     common.HexToAddress("0x0000000000000000000000000000000000000123"),
				upkeepId: big.NewInt(123),
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

func Test_keeperSubscription_isCooldownDone(t *testing.T) {
	type fields struct {
		cooldown         *big.Int
		lastInitiatedRun *big.Int
		blockHeight      *big.Int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"continues if cooldown has passed",
			fields{
				cooldown:         big.NewInt(1),
				lastInitiatedRun: big.NewInt(1),
				blockHeight:      big.NewInt(2),
			},
			true,
		},
		{
			"continues if cooldown has passed by a large amount",
			fields{
				cooldown:         big.NewInt(1),
				lastInitiatedRun: big.NewInt(1),
				blockHeight:      big.NewInt(1000),
			},
			true,
		},
		{
			"continues if there is no cooldown",
			fields{
				cooldown:         big.NewInt(0),
				lastInitiatedRun: big.NewInt(1),
				blockHeight:      big.NewInt(1),
			},
			true,
		},
		{
			"waits if cooldown has not completed",
			fields{
				cooldown:         big.NewInt(2),
				lastInitiatedRun: big.NewInt(1),
				blockHeight:      big.NewInt(2),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := keeperSubscription{
				cooldown:         tt.fields.cooldown,
				blockHeight:      tt.fields.blockHeight,
				lastInitiatedRun: tt.fields.lastInitiatedRun,
			}
			if got := keeper.isCooldownDone(); got != tt.want {
				t.Errorf("isCooldownDone() = %v, want %v", got, tt.want)
			}
		})
	}
}
