package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func Test_handleKlaytnBlockNumber(t *testing.T) {
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
				JsonrpcMessage{ID: []byte(`123`), Method: "klay_blockNumber"},
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
			got, ok := GetCannedResponse("klay", tt.args.msg)
			if ok != tt.wantOk {
				t.Errorf("handleKlaytnBlockNumber() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleKlaytnBlockNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleKlaytnRequest(t *testing.T) {
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
					Method: "klay_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, address.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "klay_subscribe",
				},
				{
					Version: "2.0",
					Method:  "klay_subscribe",
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
					Method: "klay_subscribe",
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
					Method: "klay_getLogs",
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
			got, err := handleKlaytnRequest(tt.args.conn, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleKlaytnRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleKlaytnRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleKlaytnSubscribe(t *testing.T) {
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
					Method: "klay_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"topics":[null],"address":["%s"]}]`, address.String())),
				},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					Method:  "klay_subscribe",
				},
				{
					Version: "2.0",
					Method:  "klay_subscribe",
					Params:  json.RawMessage(fmt.Sprintf(`{"subscription":"test","result":%s}`, interfaceToJson(getEthLogResponse(address.String(), nil)))),
				},
			},
			false,
		},
		{
			"incorrect params array",
			args{
				JsonrpcMessage{
					Method: "klay_subscribe",
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
					Method: "klay_subscribe",
					Params: json.RawMessage(fmt.Sprintf(`["logs",{"address":["%s"]}]`, address.String())),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleKlaytnSubscribe(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleKlaytnSubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleKlaytnSubscribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}
