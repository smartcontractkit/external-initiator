package blockchain

import (
	"reflect"
	"testing"
)

func Test_handleEthCall(t *testing.T) {
	tests := []struct {
		name    string
		msg     JsonrpcMessage
		want    []JsonrpcMessage
		wantErr bool
	}{
		{
			name: "standard eth_call",
			msg: JsonrpcMessage{
				Version: "2.0",
				ID:      []byte(`1`),
				Method:  "eth_call",
				Params:  []byte(`[{"data":"0xb7d06888"},"latest"]`),
			},
			want: []JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`1`),
					Result:  []byte(`"0x000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"`),
				},
			},
			wantErr: false,
		},
		{
			name: "unknown function selector",
			msg: JsonrpcMessage{
				Version: "2.0",
				ID:      []byte(`1`),
				Method:  "eth_call",
				Params:  []byte(`[{"data":"0x1fbe2fb6"},"latest"]`),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleEthCall(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleEthCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEthCall() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_msgToEthCall(t *testing.T) {
	tests := []struct {
		name    string
		msg     JsonrpcMessage
		want    *ethCallMessage
		wantErr bool
	}{
		{
			name: "valid message",
			msg: JsonrpcMessage{
				Version: "2.0",
				ID:      []byte(`1`),
				Method:  "eth_call",
				Params:  []byte(`[{"data":"0xb7d06888"},"latest"]`),
			},
			want:    &ethCallMessage{Data: "0xb7d06888"},
			wantErr: false,
		},
		{
			name: "invalid params order",
			msg: JsonrpcMessage{
				Version: "2.0",
				ID:      []byte(`1`),
				Method:  "eth_call",
				Params:  []byte(`["latest",{"data":"0xb7d06888"}]`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid number of params",
			msg: JsonrpcMessage{
				Version: "2.0",
				ID:      []byte(`1`),
				Method:  "eth_call",
				Params:  []byte(`[{"data":"0xb7d06888"}]`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no params",
			msg: JsonrpcMessage{
				Version: "2.0",
				ID:      []byte(`1`),
				Method:  "eth_call",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := msgToEthCall(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("msgToEthCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("msgToEthCall() got = %v, want %v", got, tt.want)
			}
		})
	}
}
