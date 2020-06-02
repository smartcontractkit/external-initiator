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
		BlockNumber:      "0x1",
		BlockHash:        "0x0",
		TransactionHash:  "0x0",
		TransactionIndex: "0x0",
		Address:          address,
		Data:             "0x0",
		Topics:           topics,
	}
}

func interfaceToJson(in interface{}) json.RawMessage {
	bz, _ := json.Marshal(in)
	return bz
}

var address = common.HexToAddress("0x0")
var address2 = common.HexToAddress("0x1")
var hash = common.HexToHash("0x123")

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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			getEthLogResponse(address.String(), []string{hash.String()}),
			false,
		},
		{
			"eth_log request with empty topics",
			args{
				JsonrpcMessage{
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, address.String())),
				},
			},
			getEthLogResponse(address.String(), nil),
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

func Test_getAddressesFromMap(t *testing.T) {
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
			got, err := getAddressesFromMap(tt.args.req)
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

func Test_getTopicsFromMap(t *testing.T) {
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
			got, err := getTopicsFromMap(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTopicsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTopicsFromMap() got = %v, want %v", got, tt.want)
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
			got, ok := GetCannedResponse("eth", tt.args.msg)
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
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`123`),
					Result:  interfaceToJson([]ethLogResponse{getEthLogResponse(address.String(), []string{hash.String()})}),
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
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, address.String())),
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
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, interfaceToJson(getEthLogResponse(address.String(), nil)))),
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
					Method: "eth_getLogs",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[["%s"]],"address":["%s"]}]`, hash.String(), address.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Result:  interfaceToJson([]ethLogResponse{getEthLogResponse(address.String(), []string{hash.String()})}),
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

func Test_handleEthSubscribe(t *testing.T) {
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
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, address.String())),
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
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, interfaceToJson(getEthLogResponse(address.String(), nil)))),
				},
			},
			false,
		},
		{
			"incorrect params array",
			args{
				JsonrpcMessage{
					Method: "eth_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`[{"topics":[null],"address":["%s"]}]`, address.String())),
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
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"address":["%s"]}]`, address.String())),
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
				t.Errorf("handleEthSubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEthSubscribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleMapStringInterface(t *testing.T) {
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
					"topics":  json.RawMessage(fmt.Sprintf(`[["%s"]]`, hash.String())),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
				},
			},
			getEthLogResponse(address.String(), []string{hash.String()}),
			false,
		},
		{
			"eth_log request with empty topics",
			args{
				map[string]json.RawMessage{
					"topics":  json.RawMessage(`[]`),
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
				},
			},
			getEthLogResponse(address.String(), nil),
			false,
		},
		{
			"eth_log request with no topics",
			args{
				map[string]json.RawMessage{
					"address": json.RawMessage(fmt.Sprintf(`["%s"]`, address.String())),
				},
			},
			ethLogResponse{},
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
			getEthLogResponse(address.String(), []string{hash.String()}),
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
			ethLogResponse{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleMapStringInterface(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleMapStringInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}
