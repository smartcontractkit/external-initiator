package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func getEthLogResponse(address string, topics []string) ethLogResponse {
	return ethLogResponse{
		LogIndex:         "0x0",
		BlockNumber:      "0x2",
		BlockHash:        "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionHash:  "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionIndex: "0x0",
		Address:          address,
		Data:             "0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
		Topics:           topics,
	}
}

func ethInterfaceToJson(in interface{}) json.RawMessage {
	bz, _ := json.Marshal(in)
	return bz
}

var ethAddress = common.HexToAddress("0x0")
var ethAddress2 = common.HexToAddress("0x1")
var ethHash = common.HexToHash("0x123")

func Test_ethLogRequestToResponse(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    ethLogResponse
		wantErr bool
	}{
		{
			"correct eth_log request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, ethHash.String(), ethAddress.String())),
				},
			},
			getEthLogResponse(ethAddress.String(), []string{ethHash.String()}),
			false,
		},
		{
			"eth_log request with empty topics",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, ethAddress.String())),
				},
			},
			getEthLogResponse(ethAddress.String(), nil),
			false,
		},
		{
			"misformed payload",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[{"topics":[null],"address":"0x0"}]`),
				},
			},
			ethLogResponse{},
			true,
		},
		{
			"empty request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[]`),
				},
			},
			ethLogResponse{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ethLogRequestToResponse(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ethLogRequestToResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ethLogRequestToResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEthAddressesFromMap(t *testing.T) {
	type args struct {
		req map[string]json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    []common.Address
		wantErr bool
	}{
		{
			"correct payload",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, ethAddress.String())),
				},
			},
			[]common.Address{ethAddress},
			false,
		},
		{
			"multiple addresses",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s", "%s"]`, ethAddress.String(), ethAddress.String(), ethAddress.String())),
				},
			},
			[]common.Address{ethAddress, ethAddress, ethAddress},
			false,
		},
		{
			"no addresses",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(`[]`),
				},
			},
			nil,
			true,
		},
		{
			"missing address key",
			args{
				map[string]json.RawMessage{
					"something_else": json.RawMessage(fmt.Sprintf(`["%s"]`, ethAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEthAddressesFromMap(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAddressesFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAddressesFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEthTopicsFromMap(t *testing.T) {
	type args struct {
		req map[string]json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    [][]common.Hash
		wantErr bool
	}{
		{
			"correct payload",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[["%s"]]`, ethHash.String())),
				},
			},
			[][]common.Hash{{ethHash}},
			false,
		},
		{
			"multiple topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[["%s","%s"],["%s"]]`, ethHash.String(), ethHash.String(), ethHash.String())),
				},
			},
			[][]common.Hash{{ethHash, ethHash}, {ethHash}},
			false,
		},
		{
			"nil in topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[null,["%s"]]`, ethHash.String())),
				},
			},
			[][]common.Hash{{ethHash}},
			false,
		},
		{
			"nil topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(`[null]`),
				},
			},
			nil,
			false,
		},
		{
			"missing topics key",
			args{
				map[string]json.RawMessage{
					"something_else": json.RawMessage(`[null]`),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEthTopicsFromMap(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getEthTopicsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEthTopicsFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleEthBlockNumber(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name   string
		args   args
		want   []JsonrpcMessage
		wantOk bool
	}{
		{
			"returns a block number with the correct ID",
			args{
				JsonrpcMessage{ID: []byte(`123`), Method: "eth_blockNumber"},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`123`),
					Result:  []byte(`"0x0"`),
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetCannedResponse("binance-smart-chain", tt.args.msg)
			if ok != tt.wantOk {
				t.Errorf("handleEthBlockNumber() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEthBlockNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleEthGetLogs(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    []JsonrpcMessage
		wantErr bool
	}{
		{
			"returns a response with the correct ID",
			args{
				JsonrpcMessage{
					ID:     []byte(`123`),
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, ethHash.String(), ethAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`123`),
					Result:  ethInterfaceToJson([]ethLogResponse{getEthLogResponse(ethAddress.String(), []string{ethHash.String()})}),
				},
			},
			false,
		},
		{
			"fails on missing address",
			args{
				JsonrpcMessage{
					ID:     []byte(`123`),
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":[]}]`, ethHash.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleEthGetLogs(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleEthGetLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEthGetLogs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleEthRequest(t *testing.T) {
	type args struct {
		conn string
		msg  JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    []JsonrpcMessage
		wantErr bool
	}{
		{
			"handles WS subscribe",
			args{
				"ws",
				JsonrpcMessage{
					Method: "eth_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, ethAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "eth_subscribe",
				},
				{
					Version: "2.0",
					Method:  "eth_subscribe",
					Params:  json.RawMessage(fmt.Sprintf(`{"suethription":"test","result":%s}`, ethInterfaceToJson(getEthLogResponse(ethAddress.String(), nil)))),
				},
			},
			false,
		},
		{
			"fails eth_subscribe on RPC",
			args{
				"rpc",
				JsonrpcMessage{
					Method: "eth_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, ethAddress.String())),
				},
			},
			nil,
			true,
		},
		{
			"gets logs",
			args{
				"rpc",
				JsonrpcMessage{
					Method: "eth_getLogs",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, ethHash.String(), ethAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Result:  ethInterfaceToJson([]ethLogResponse{getEthLogResponse(ethAddress.String(), []string{ethHash.String()})}),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleEthRequest(tt.args.conn, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleEthRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEthRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleEthSuethribe(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    []JsonrpcMessage
		wantErr bool
	}{
		{
			"handles correct subscribe",
			args{
				JsonrpcMessage{
					Method: "eth_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, ethAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "eth_subscribe",
				},
				{
					Version: "2.0",
					Method:  "eth_subscribe",
					Params:  json.RawMessage(fmt.Sprintf(`{"suethription":"test","result":%s}`, ethInterfaceToJson(getEthLogResponse(ethAddress.String(), nil)))),
				},
			},
			false,
		},
		{
			"incorrect params array",
			args{
				JsonrpcMessage{
					Method: "eth_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, ethAddress.String())),
				},
			},
			nil,
			true,
		},
		{
			"incorrect map string interface",
			args{
				JsonrpcMessage{
					Method: "eth_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"address":["%s"]}]`, ethAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleEthSubscribe(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleEthSuethribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEthSuethribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleEthMapStringInterface(t *testing.T) {
	type args struct {
		in map[string]json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    ethLogResponse
		wantErr bool
	}{
		{
			"correct eth_log request",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, ethHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, ethAddress.String())),
				},
			},
			getEthLogResponse(ethAddress.String(), []string{ethHash.String()}),
			false,
		},
		{
			"eth_log request with empty topics",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(`[]`),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, ethAddress.String())),
				},
			},
			getEthLogResponse(ethAddress.String(), nil),
			false,
		},
		{
			"eth_log request with no topics",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, ethAddress.String())),
				},
			},
			ethLogResponse{},
			true,
		},
		{
			"uses first address",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, ethHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s"]`, ethAddress.String(), ethAddress2.String())),
				},
			},
			getEthLogResponse(ethAddress.String(), []string{ethHash.String()}),
			false,
		},
		{
			"fails on no addresses",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, ethHash.String())),
					"address": json.RawMessage(`[]`),
				},
			},
			ethLogResponse{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleEthMapStringInterface(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleEthMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEthMapStringInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}
