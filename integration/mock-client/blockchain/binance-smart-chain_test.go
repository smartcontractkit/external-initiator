package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func getBscLogResponse(address string, topics []string) bscLogResponse {
	return bscLogResponse{
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

func bscInterfaceToJson(in interface{}) json.RawMessage {
	bz, _ := json.Marshal(in)
	return bz
}

var bscAddress = common.HexToAddress("0x0")
var bscAddress2 = common.HexToAddress("0x1")
var bscHash = common.HexToHash("0x123")

func Test_bscLogRequestToResponse(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    bscLogResponse
		wantErr bool
	}{
		{
			"correct eth_log request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			getBscLogResponse(address.String(), []string{hash.String()}),
			false,
		},
		{
			"eth_log request with empty topics",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, address.String())),
				},
			},
			getBscLogResponse(address.String(), nil),
			false,
		},
		{
			"misformed payload",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[{"topics":[null],"address":"0x0"}]`),
				},
			},
			bscLogResponse{},
			true,
		},
		{
			"empty request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[]`),
				},
			},
			bscLogResponse{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bscLogRequestToResponse(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("bscLogRequestToResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bscLogRequestToResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getBscAddressesFromMap(t *testing.T) {
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
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, bscAddress.String())),
				},
			},
			[]common.Address{address},
			false,
		},
		{
			"multiple addresses",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s", "%s"]`, bscAddress.String(), bscAddress.String(), bscAddress.String())),
				},
			},
			[]common.Address{address, address, address},
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
					"something_else": json.RawMessage(fmt.Sprintf(`["%s"]`, bscAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBscAddressesFromMap(tt.args.req)
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

func Test_getBscTopicsFromMap(t *testing.T) {
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
					"topics": json.RawMessage(fmt.Sprintf(`[["%s"]]`, bscHash.String())),
				},
			},
			[][]common.Hash{{bscHash}},
			false,
		},
		{
			"multiple topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[["%s","%s"],["%s"]]`, bscHash.String(), bscHash.String(), bscHash.String())),
				},
			},
			[][]common.Hash{{bscHash, bscHash}, {bscHash}},
			false,
		},
		{
			"nil in topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[null,["%s"]]`, bscHash.String())),
				},
			},
			[][]common.Hash{{bscHash}},
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
			got, err := getBscTopicsFromMap(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBscTopicsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBscTopicsFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleBscBlockNumber(t *testing.T) {
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
				t.Errorf("handleBscBlockNumber() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleBscBlockNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleBscGetLogs(t *testing.T) {
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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`123`),
					Result:  bscInterfaceToJson([]bscLogResponse{getBscLogResponse(address.String(), []string{hash.String()})}),
				},
			},
			false,
		},
		{
			"fails on missing address",
			args{
				JsonrpcMessage{
					ID:     []byte(`123`),
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":[]}]`, hash.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleBscGetLogs(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleBscGetLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleBscGetLogs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleBscRequest(t *testing.T) {
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
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, bscAddress.String())),
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
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, bscInterfaceToJson(getBscLogResponse(bscAddress.String(), nil)))),
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
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, bscAddress.String())),
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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, bscHash.String(), bscAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Result:  bscInterfaceToJson([]bscLogResponse{getBscLogResponse(bscAddress.String(), []string{bscHash.String()})}),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleBscRequest(tt.args.conn, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleBscRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleBscRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleBscSubscribe(t *testing.T) {
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
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, bscAddress.String())),
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
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, bscInterfaceToJson(getBscLogResponse(bscAddress.String(), nil)))),
				},
			},
			false,
		},
		{
			"incorrect params array",
			args{
				JsonrpcMessage{
					Method: "eth_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, bscAddress.String())),
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
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"address":["%s"]}]`, bscAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleBscSubscribe(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleBscSubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleBscSubscribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleBscMapStringInterface(t *testing.T) {
	type args struct {
		in map[string]json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    bscLogResponse
		wantErr bool
	}{
		{
			"correct eth_log request",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, bscHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, bscAddress.String())),
				},
			},
			getBscLogResponse(bscAddress.String(), []string{bscHash.String()}),
			false,
		},
		{
			"eth_log request with empty topics",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(`[]`),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, bscAddress.String())),
				},
			},
			getBscLogResponse(bscAddress.String(), nil),
			false,
		},
		{
			"eth_log request with no topics",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, bscAddress.String())),
				},
			},
			bscLogResponse{},
			true,
		},
		{
			"uses first address",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, bscHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s"]`, bscAddress.String(), bscAddress2.String())),
				},
			},
			getBscLogResponse(bscAddress.String(), []string{bscHash.String()}),
			false,
		},
		{
			"fails on no addresses",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, hash.String())),
					"address": json.RawMessage(`[]`),
				},
			},
			bscLogResponse{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleBscMapStringInterface(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleBscMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleBscMapStringInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}
