package blockchain

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_nearManager_GetTestJson(t *testing.T) {
	type args struct {
		filter         nearFilter
		connectionType subscriber.Type
	}
	filter := nearFilter{
		JobID:      "job#1",
		AccountIDs: []string{"oracle.chainlink.testnet"},
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"returns JSON when using RPC",
			args{filter: filter, connectionType: subscriber.RPC},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"status"}`),
		},
		{
			"returns empty when using WS",
			args{filter: filter, connectionType: subscriber.WS},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: filter, connectionType: tt.args.connectionType}
			if got := m.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nearManager_ParseTestResponse(t *testing.T) {
	type fields struct {
		filter         nearFilter
		connectionType subscriber.Type
	}
	filter := nearFilter{
		JobID:      "job#1",
		AccountIDs: []string{"oracle.chainlink.testnet"},
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"does nothing for WS",
			fields{filter: filter, connectionType: subscriber.WS},
			args{},
			false,
		},
		{
			"fails unmarshal payload",
			fields{filter: filter, connectionType: subscriber.RPC},
			args{[]byte(`error`)},
			true,
		},
		{
			"fails unmarshal result",
			fields{filter: filter, connectionType: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: filter, connectionType: tt.fields.connectionType}
			if err := m.ParseTestResponse(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_nearManager_GetTriggerJson(t *testing.T) {
	type args struct {
		filter         nearFilter
		connectionType subscriber.Type
	}
	filter := nearFilter{
		JobID:      "job#1",
		AccountIDs: []string{"oracle.chainlink.testnet"},
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"returns JSON-RPC message when using RPC",
			args{filter: filter, connectionType: subscriber.RPC},
			true,
		},
		{
			"returns empty when using WS",
			args{filter: filter, connectionType: subscriber.WS},
			false,
		},
		{
			"returns empty when using Client",
			args{filter: filter, connectionType: subscriber.Client},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: tt.args.filter, connectionType: tt.args.connectionType}
			if got := m.GetTriggerJson(); (got != nil) != tt.want {
				t.Errorf("GetTriggerJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NEAROracleRequest_Unmarshal(t *testing.T) {
	result := []byte{123, 34, 99, 108, 105, 101, 110, 116, 46, 111, 46, 109, 105, 107,
		101, 46, 116, 101, 115, 116, 110, 101, 116, 34, 58, 91, 123, 34, 110, 111, 110,
		99, 101, 34, 58, 34, 49, 34, 44, 34, 114, 101, 113, 117, 101, 115, 116, 34, 58,
		123, 34, 99, 97, 108, 108, 101, 114, 95, 97, 99, 99, 111, 117, 110, 116, 34, 58,
		34, 99, 108, 105, 101, 110, 116, 46, 111, 46, 109, 105, 107, 101, 46, 116, 101,
		115, 116, 110, 101, 116, 34, 44, 34, 114, 101, 113, 117, 101, 115, 116, 95, 115,
		112, 101, 99, 34, 58, 34, 100, 87, 53, 112, 99, 88, 86, 108, 73, 72, 78, 119, 90,
		87, 77, 103, 97, 87, 81, 61, 34, 44, 34, 99, 97, 108, 108, 98, 97, 99, 107, 95,
		97, 100, 100, 114, 101, 115, 115, 34, 58, 34, 99, 108, 105, 101, 110, 116, 46, 111,
		46, 109, 105, 107, 101, 46, 116, 101, 115, 116, 110, 101, 116, 34, 44, 34, 99, 97,
		108, 108, 98, 97, 99, 107, 95, 109, 101, 116, 104, 111, 100, 34, 58, 34, 116, 111,
		107, 101, 110, 95, 112, 114, 105, 99, 101, 95, 99, 97, 108, 108, 98, 97, 99, 107,
		34, 44, 34, 100, 97, 116, 97, 34, 58, 34, 81, 107, 70, 85, 34, 44, 34, 112, 97, 121,
		109, 101, 110, 116, 34, 58, 49, 48, 44, 34, 101, 120, 112, 105, 114, 97, 116, 105,
		111, 110, 34, 58, 49, 57, 48, 54, 50, 57, 51, 52, 50, 55, 50, 52, 54, 51, 48, 54,
		55, 48, 48, 125, 125, 44, 123, 34, 110, 111, 110, 99, 101, 34, 58, 34, 50, 34, 44,
		34, 114, 101, 113, 117, 101, 115, 116, 34, 58, 123, 34, 99, 97, 108, 108, 101, 114,
		95, 97, 99, 99, 111, 117, 110, 116, 34, 58, 34, 99, 108, 105, 101, 110, 116, 46, 111,
		46, 109, 105, 107, 101, 46, 116, 101, 115, 116, 110, 101, 116, 34, 44, 34, 114, 101,
		113, 117, 101, 115, 116, 95, 115, 112, 101, 99, 34, 58, 34, 100, 87, 53, 112, 99, 88,
		86, 108, 73, 72, 78, 119, 90, 87, 77, 103, 97, 87, 81, 61, 34, 44, 34, 99, 97, 108,
		108, 98, 97, 99, 107, 95, 97, 100, 100, 114, 101, 115, 115, 34, 58, 34, 99, 108, 105,
		101, 110, 116, 46, 111, 46, 109, 105, 107, 101, 46, 116, 101, 115, 116, 110, 101, 116,
		34, 44, 34, 99, 97, 108, 108, 98, 97, 99, 107, 95, 109, 101, 116, 104, 111, 100, 34,
		58, 34, 116, 111, 107, 101, 110, 95, 112, 114, 105, 99, 101, 95, 99, 97, 108, 108, 98,
		97, 99, 107, 34, 44, 34, 100, 97, 116, 97, 34, 58, 34, 81, 107, 70, 85, 34, 44, 34,
		112, 97, 121, 109, 101, 110, 116, 34, 58, 49, 48, 44, 34, 101, 120, 112, 105, 114, 97,
		116, 105, 111, 110, 34, 58, 49, 57, 48, 54, 50, 57, 51, 52, 50, 55, 50, 52, 54, 51,
		48, 54, 55, 48, 48, 125, 125, 93, 44, 34, 111, 46, 109, 105, 107, 101, 46, 116, 101,
		115, 116, 110, 101, 116, 34, 58, 91, 123, 34, 110, 111, 110, 99, 101, 34, 58, 34, 49,
		34, 44, 34, 114, 101, 113, 117, 101, 115, 116, 34, 58, 123, 34, 99, 97, 108, 108, 101,
		114, 95, 97, 99, 99, 111, 117, 110, 116, 34, 58, 34, 111, 46, 109, 105, 107, 101, 46,
		116, 101, 115, 116, 110, 101, 116, 34, 44, 34, 114, 101, 113, 117, 101, 115, 116, 95,
		115, 112, 101, 99, 34, 58, 34, 100, 87, 53, 112, 99, 88, 86, 108, 73, 72, 78, 119, 90,
		87, 77, 103, 97, 87, 81, 61, 34, 44, 34, 99, 97, 108, 108, 98, 97, 99, 107, 95, 97, 100,
		100, 114, 101, 115, 115, 34, 58, 34, 99, 108, 105, 101, 110, 116, 46, 111, 46, 109, 105,
		107, 101, 46, 116, 101, 115, 116, 110, 101, 116, 34, 44, 34, 99, 97, 108, 108, 98, 97,
		99, 107, 95, 109, 101, 116, 104, 111, 100, 34, 58, 34, 116, 111, 107, 101, 110, 95, 112,
		114, 105, 99, 101, 95, 99, 97, 108, 108, 98, 97, 99, 107, 34, 44, 34, 100, 97, 116, 97,
		34, 58, 34, 81, 107, 70, 85, 34, 44, 34, 112, 97, 121, 109, 101, 110, 116, 34, 58, 49,
		48, 44, 34, 101, 120, 112, 105, 114, 97, 116, 105, 111, 110, 34, 58, 49, 57, 48, 54,
		50, 57, 51, 52, 50, 55, 50, 52, 54, 51, 48, 54, 55, 48, 48, 125, 125, 44, 123, 34, 110,
		111, 110, 99, 101, 34, 58, 34, 51, 34, 44, 34, 114, 101, 113, 117, 101, 115, 116, 34,
		58, 123, 34, 99, 97, 108, 108, 101, 114, 95, 97, 99, 99, 111, 117, 110, 116, 34, 58,
		34, 111, 46, 109, 105, 107, 101, 46, 116, 101, 115, 116, 110, 101, 116, 34, 44, 34,
		114, 101, 113, 117, 101, 115, 116, 95, 115, 112, 101, 99, 34, 58, 34, 100, 87, 53, 112,
		99, 88, 86, 108, 73, 72, 78, 119, 90, 87, 77, 103, 97, 87, 81, 61, 34, 44, 34, 99, 97,
		108, 108, 98, 97, 99, 107, 95, 97, 100, 100, 114, 101, 115, 115, 34, 58, 34, 99, 108,
		105, 101, 110, 116, 46, 111, 46, 109, 105, 107, 101, 46, 116, 101, 115, 116, 110, 101,
		116, 34, 44, 34, 99, 97, 108, 108, 98, 97, 99, 107, 95, 109, 101, 116, 104, 111, 100,
		34, 58, 34, 116, 111, 107, 101, 110, 95, 112, 114, 105, 99, 101, 95, 99, 97, 108, 108,
		98, 97, 99, 107, 34, 44, 34, 100, 97, 116, 97, 34, 58, 34, 81, 107, 70, 85, 34, 44, 34,
		112, 97, 121, 109, 101, 110, 116, 34, 58, 49, 48, 44, 34, 101, 120, 112, 105, 114, 97,
		116, 105, 111, 110, 34, 58, 49, 57, 48, 54, 50, 57, 51, 52, 50, 55, 50, 52, 54, 51, 48,
		54, 55, 48, 48, 125, 125, 44, 123, 34, 110, 111, 110, 99, 101, 34, 58, 34, 49, 57, 34,
		44, 34, 114, 101, 113, 117, 101, 115, 116, 34, 58, 123, 34, 99, 97, 108, 108, 101, 114,
		95, 97, 99, 99, 111, 117, 110, 116, 34, 58, 34, 111, 46, 109, 105, 107, 101, 46, 116, 101,
		115, 116, 110, 101, 116, 34, 44, 34, 114, 101, 113, 117, 101, 115, 116, 95, 115, 112, 101,
		99, 34, 58, 34, 100, 87, 53, 112, 99, 88, 86, 108, 73, 72, 78, 119, 90, 87, 77, 103, 97,
		87, 81, 61, 34, 44, 34, 99, 97, 108, 108, 98, 97, 99, 107, 95, 97, 100, 100, 114, 101, 115,
		115, 34, 58, 34, 99, 108, 105, 101, 110, 116, 46, 111, 46, 109, 105, 107, 101, 46, 116,
		101, 115, 116, 110, 101, 116, 34, 44, 34, 99, 97, 108, 108, 98, 97, 99, 107, 95, 109, 101,
		116, 104, 111, 100, 34, 58, 34, 116, 111, 107, 101, 110, 95, 112, 114, 105, 99, 101, 95,
		99, 97, 108, 108, 98, 97, 99, 107, 34, 44, 34, 100, 97, 116, 97, 34, 58, 34, 81, 107, 70,
		85, 34, 44, 34, 112, 97, 121, 109, 101, 110, 116, 34, 58, 49, 48, 44, 34, 101, 120, 112,
		105, 114, 97, 116, 105, 111, 110, 34, 58, 49, 57, 48, 54, 50, 57, 51, 52, 50, 55, 50, 52,
		54, 51, 48, 54, 55, 48, 48, 125, 125, 93, 125}

	var oracleRequests map[string][]NEAROracleRequest
	err := json.Unmarshal(result, &oracleRequests)
	require.Nil(t, err)
	assert.NotNil(t, oracleRequests)
	assert.Equal(t, 2, len(oracleRequests))
}

func Test_nearManager_ParseResponse(t *testing.T) {
	// TODO: test
}
