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
		EpochNumber:      "0x1",
		BlockHash:        "0x0",
		TransactionHash:  "0x0",
		TransactionIndex: "0x0",
		Address:          address,
		Data:             "0x0",
		Topics:           topics,
	}
}

// func interfaceToJson(in interface{}) json.RawMessage {
// 	bz, _ := json.Marshal(in)
// 	return bz
// }

// var address = common.HexToAddress("0x0")
// var address2 = common.HexToAddress("0x1")
// var hash = common.HexToHash("0x123")

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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			getCfxLogResponse(address.String(), []string{hash.String()}),
			false,
		},
		{
			"cfx_log request with empty topics",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, address.String())),
				},
			},
			getCfxLogResponse(address.String(), nil),
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
		want    []common.Address
		wantErr bool
	}{
		{
			"correct payload",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
				},
			},
			[]common.Address{address},
			false,
		},
		{
			"multiple addresses",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s", "%s"]`, address.String(), address.String(), address.String())),
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
					"something_else": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
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
					"topics": json.RawMessage(fmt.Sprintf(`[["%s"]]`, hash.String())),
				},
			},
			[][]common.Hash{{hash}},
			false,
		},
		{
			"multiple topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[["%s","%s"],["%s"]]`, hash.String(), hash.String(), hash.String())),
				},
			},
			[][]common.Hash{{hash, hash}, {hash}},
			false,
		},
		{
			"nil in topics",
			args{
				map[string]json.RawMessage{
					"topics": json.RawMessage(fmt.Sprintf(`[null,["%s"]]`, hash.String())),
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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`123`),
					Result:  interfaceToJson([]cfxLogResponse{getCfxLogResponse(address.String(), []string{hash.String()})}),
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
			"fails cfx_subscribe on RPC",
			args{
				"rpc",
				JsonrpcMessage{
					Method: "cfx_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, address.String())),
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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Result:  interfaceToJson([]cfxLogResponse{getCfxLogResponse(address.String(), []string{hash.String()})}),
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
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, hash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
				},
			},
			getCfxLogResponse(address.String(), []string{hash.String()}),
			false,
		},
		{
			"cfx_log request with empty topics",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(`[]`),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
				},
			},
			getCfxLogResponse(address.String(), nil),
			false,
		},
		{
			"cfx_log request with no topics",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
				},
			},
			cfxLogResponse{},
			true,
		},
		{
			"uses first address",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, hash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s", "%s"]`, address.String(), address2.String())),
				},
			},
			getCfxLogResponse(address.String(), []string{hash.String()}),
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
