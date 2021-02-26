package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func getCfxLogResponse(address string, topics []string) cfxLogResponse {
	return cfxLogResponse{
		LogIndex:         "0x0",
		EpochNumber:      "0x2",
		BlockHash:        "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionHash:  "0xabc0000000000000000000000000000000000000000000000000000000000000",
		TransactionIndex: "0x0",
		Address:          address,
		Data:             "0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
		Topics:           topics,
	}
}

func cfxInterfaceToJson(in interface{}) json.RawMessage {
	bz, _ := json.Marshal(in)
	return bz
}

var cfxAddress = common.HexToAddress("0x0")
var cfxAddress2 = common.HexToAddress("0x1")
var cfxHash = common.HexToHash("0x123")

func Test_cfxLogRequestToResponse(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    cfxLogResponse
		wantErr bool
	}{
		{
			"correct cfx_log request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, cfxHash.String(), cfxAddress.String())),
				},
			},
			getCfxLogResponse(cfxAddress.String(), []string{cfxHash.String()}),
			false,
		},
		{
			"cfx_log request with empty topics",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, cfxAddress.String())),
				},
			},
			getCfxLogResponse(cfxAddress.String(), nil),
			false,
		},
		{
			"misformed payload",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[{"topics":[null],"address":"0x0"}]`),
				},
			},
			cfxLogResponse{},
			true,
		},
		{
			"empty request",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(`[]`),
				},
			},
			cfxLogResponse{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfxLogRequestToResponse(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("cfxLogRequestToResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cfxLogRequestToResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCfxAddressesFromMap(t *testing.T) {
	type args struct {
		req map[string]json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"correct payload",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, cfxAddress.String())),
				},
			},
			[]string{cfxAddress.String()},
			false,
		},
		{
			"multiple addresses",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s", "%s"]`, cfxAddress.String(), cfxAddress.String(), cfxAddress.String())),
				},
			},
			[]string{address.String(), address.String(), address.String()},
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
					"something_else": json.RawMessage(fmt.Sprintf(`["%s"]`, cfxAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCfxAddressesFromMap(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCfxAddressesFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCfxAddressesFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCfxTopicsFromMap(t *testing.T) {
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
					"topics": json.RawMessage(fmt.Sprintf(`[["%s"]]`, cfxHash.String())),
				},
			},
			[][]common.Hash{{hash}},
			false,
		},
		{
			"multiple topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[["%s","%s"],["%s"]]`, cfxHash.String(), cfxHash.String(), cfxHash.String())),
				},
			},
			[][]common.Hash{{hash, hash}, {hash}},
			false,
		},
		{
			"nil in topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[null,["%s"]]`, cfxHash.String())),
				},
			},
			[][]common.Hash{{hash}},
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
			got, err := getCfxTopicsFromMap(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCfxTopicsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCfxTopicsFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleCfxEpochNumber(t *testing.T) {
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
			"returns a epoch number with the correct ID",
			args{
				JsonrpcMessage{ID: []byte(`123`), Method: "cfx_epochNumber"},
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
			got, ok := GetCannedResponse("cfx", tt.args.msg)
			if ok != tt.wantOk {
				t.Errorf("handleCfxEpochNumber() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleCfxEpochNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleCfxGetLogs(t *testing.T) {
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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, cfxHash.String(), cfxAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`123`),
					Result:  cfxInterfaceToJson([]cfxLogResponse{getCfxLogResponse(cfxAddress.String(), []string{cfxHash.String()})}),
				},
			},
			false,
		},
		{
			"fails on missing address",
			args{
				JsonrpcMessage{
					ID:     []byte(`123`),
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":[]}]`, cfxHash.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleCfxGetLogs(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCfxGetLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleCfxGetLogs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleCfxRequest(t *testing.T) {
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
					Method: "cfx_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, cfxAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "cfx_subscribe",
				},
				{
					Version: "2.0",
					Method:  "cfx_subscribe",
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, cfxInterfaceToJson(getCfxLogResponse(cfxAddress.String(), nil)))),
				},
			},
			false,
		},
		{
			"fails cfx_subscribe on RPC",
			args{
				"rpc",
				JsonrpcMessage{
					Method: "cfx_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, cfxAddress.String())),
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
					Method: "cfx_getLogs",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, cfxHash.String(), cfxAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Result:  cfxInterfaceToJson([]cfxLogResponse{getCfxLogResponse(cfxAddress.String(), []string{cfxHash.String()})}),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleCfxRequest(tt.args.conn, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCfxRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleCfxRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleCfxSubscribe(t *testing.T) {
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
					Method: "cfx_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, cfxAddress.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "cfx_subscribe",
				},
				{
					Version: "2.0",
					Method:  "cfx_subscribe",
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, cfxInterfaceToJson(getCfxLogResponse(cfxAddress.String(), nil)))),
				},
			},
			false,
		},
		{
			"incorrect params array",
			args{
				JsonrpcMessage{
					Method: "cfx_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, cfxAddress.String())),
				},
			},
			nil,
			true,
		},
		{
			"incorrect map string interface",
			args{
				JsonrpcMessage{
					Method: "cfx_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"address":["%s"]}]`, cfxAddress.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleCfxSubscribe(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCfxSubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleCfxSubscribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleCfxMapStringInterface(t *testing.T) {
	type args struct {
		in map[string]json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    cfxLogResponse
		wantErr bool
	}{
		{
			"correct cfx_log request",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, cfxHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, cfxAddress.String())),
				},
			},
			getCfxLogResponse(cfxAddress.String(), []string{cfxHash.String()}),
			false,
		},
		{
			"cfx_log request with empty topics",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(`[]`),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, cfxAddress.String())),
				},
			},
			getCfxLogResponse(cfxAddress.String(), nil),
			false,
		},
		{
			"cfx_log request with no topics",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, cfxAddress.String())),
				},
			},
			cfxLogResponse{},
			true,
		},
		{
			"uses first address",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, cfxHash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s"]`, cfxAddress.String(), cfxAddress2.String())),
				},
			},
			getCfxLogResponse(cfxAddress.String(), []string{cfxHash.String()}),
			false,
		},
		{
			"fails on no addresses",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, cfxHash.String())),
					"address": json.RawMessage(`[]`),
				},
			},
			cfxLogResponse{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleCfxMapStringInterface(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCfxMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleCfxMapStringInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}
