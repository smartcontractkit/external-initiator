package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func getHmyLogResponse(address string, topics []string) hmyLogResponse {
	return hmyLogResponse{
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

func hmyInterfaceToJson(in interface{}) json.RawMessage {
	bz, _ := json.Marshal(in)
	return bz
}

var hmyAddress = common.HexToAddress("0x0")
var hmyAddress2 = common.HexToAddress("0x1")
var hmyHash = common.HexToHash("0x123")

func Test_hmyLogRequestToResponse(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    hmyLogResponse
		wantErr bool
	}{
		{
			"correct hmy_log request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			getHmyLogResponse(address.String(), []string{hash.String()}),
			false,
		},
		{
			"hmy_log request with empty topics",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, address.String())),
				},
			},
			getHmyLogResponse(address.String(), nil),
			false,
		},
		{
			"misformed payload",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[{"topics":[null],"address":"0x0"}]`),
				},
			},
			hmyLogResponse{},
			true,
		},
		{
			"empty request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[]`),
				},
			},
			hmyLogResponse{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hmyLogRequestToResponse(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("hmyLogRequestToResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hmyLogRequestToResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHmyAddressesFromMap(t *testing.T) {
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
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, hmyAddress.String())),
				},
			},
			[]common.Address{address},
			false,
		},
		{
			"multiple addresses",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s", "%s"]`, hmyAddress.String(), hmyAddress.String(), hmyAddress.String())),
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
					"something_else": json.RawMessage(fmt.Sprintf(`["%s"]`, hmyAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHmyAddressesFromMap(tt.args.req)
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

func Test_getHmyTopicsFromMap(t *testing.T) {
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
					"topics": json.RawMessage(fmt.Sprintf(`[["%s"]]`, hmyHash.String())),
				},
			},
			[][]common.Hash{{hmyHash}},
			false,
		},
		{
			"multiple topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[["%s","%s"],["%s"]]`, hmyHash.String(), hmyHash.String(), hmyHash.String())),
				},
			},
			[][]common.Hash{{hmyHash, hmyHash}, {hmyHash}},
			false,
		},
		{
			"nil in topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[null,["%s"]]`, hmyHash.String())),
				},
			},
			[][]common.Hash{{hmyHash}},
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
			got, err := getHmyTopicsFromMap(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHmyTopicsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHmyTopicsFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleHmyBlockNumber(t *testing.T) {
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
				JsonrpcMessage{ID: []byte(`123`), Method: "hmy_blockNumber"},
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
			got, ok := GetCannedResponse("hmy", tt.args.msg)
			if ok != tt.wantOk {
				t.Errorf("handleHmyBlockNumber() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleHmyBlockNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleHmyGetLogs(t *testing.T) {
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
					Result:  hmyInterfaceToJson([]hmyLogResponse{getHmyLogResponse(address.String(), []string{hash.String()})}),
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
			got, err := handleHmyGetLogs(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleHmyGetLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleHmyGetLogs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleHmyRequest(t *testing.T) {
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
					Method: "hmy_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, hmyAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "hmy_subscribe",
				},
				{
					Version: "2.0",
					Method:  "hmy_subscribe",
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, hmyInterfaceToJson(getHmyLogResponse(hmyAddress.String(), nil)))),
				},
			},
			false,
		},
		{
			"fails hmy_subscribe on RPC",
			args{
				"rpc",
				JsonrpcMessage{
					Method: "hmy_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, hmyAddress.String())),
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
					Method: "hmy_getLogs",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hmyHash.String(), hmyAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Result:  hmyInterfaceToJson([]hmyLogResponse{getHmyLogResponse(hmyAddress.String(), []string{hmyHash.String()})}),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleHmyRequest(tt.args.conn, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleHmyRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleHmyRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleHmySubscribe(t *testing.T) {
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
					Method: "hmy_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, hmyAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "hmy_subscribe",
				},
				{
					Version: "2.0",
					Method:  "hmy_subscribe",
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, hmyInterfaceToJson(getHmyLogResponse(hmyAddress.String(), nil)))),
				},
			},
			false,
		},
		{
			"incorrect params array",
			args{
				JsonrpcMessage{
					Method: "hmy_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, hmyAddress.String())),
				},
			},
			nil,
			true,
		},
		{
			"incorrect map string interface",
			args{
				JsonrpcMessage{
					Method: "hmy_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"address":["%s"]}]`, hmyAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleHmySubscribe(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleHmySubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleHmySubscribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleHmyMapStringInterface(t *testing.T) {
	type args struct {
		in map[string]json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    hmyLogResponse
		wantErr bool
	}{
		{
			"correct hmy_log request",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, hmyHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, hmyAddress.String())),
				},
			},
			getHmyLogResponse(hmyAddress.String(), []string{hmyHash.String()}),
			false,
		},
		{
			"hmy_log request with empty topics",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(`[]`),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, hmyAddress.String())),
				},
			},
			getHmyLogResponse(hmyAddress.String(), nil),
			false,
		},
		{
			"hmy_log request with no topics",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, hmyAddress.String())),
				},
			},
			hmyLogResponse{},
			true,
		},
		{
			"uses first address",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, hmyHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s"]`, hmyAddress.String(), hmyAddress2.String())),
				},
			},
			getHmyLogResponse(hmyAddress.String(), []string{hmyHash.String()}),
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
			hmyLogResponse{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleHmyMapStringInterface(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleHmyMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleHmyMapStringInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}
