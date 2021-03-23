package blockchain

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/eitest"
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
		Nonces:     nil,
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"returns JSON when using RPC",
			args{filter: filter, connectionType: subscriber.RPC},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"query","params":{"request_type":"call_function","finality":"final","account_id":"oracle.chainlink.testnet","method_name":"get_nonces","args_base64":""}}`),
		},
		{
			"returns empty when using WS",
			args{filter: filter, connectionType: subscriber.WS},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: &filter, connectionType: tt.args.connectionType}
			if got := m.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", string(got), string(tt.want))
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
		Nonces:     nil,
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
			m := nearManager{filter: &filter, connectionType: tt.fields.connectionType}
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
		Nonces:     nil,
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
			m := nearManager{filter: &tt.args.filter, connectionType: tt.args.connectionType}
			if got := m.GetTriggerJson(); (got != nil) != tt.want {
				t.Errorf("GetTriggerJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

// readTestGetAllRequestsResult will read JSON-RPC result for `get_all_requests` call from a test file
func readTestGetAllRequestsResult() ([]byte, error) {
	wd, _ := os.Getwd()
	ui := path.Join(wd, "testdata/near_test_oracle_get_all_requests.json")
	file, err := os.Open(ui)
	if err != nil {
		return nil, err
	}
	defer eitest.MustClose(file)

	resultJSON, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return resultJSON, nil
}

func readTestJSONRPCMessage() (*JsonrpcMessage, error) {
	resultRaw, err := readTestGetAllRequestsResult()
	if err != nil {
		return nil, err
	}

	var msg JsonrpcMessage
	err = json.Unmarshal(resultRaw, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func Test_NEAROracleRequest_Unmarshal(t *testing.T) {
	msg, err := readTestJSONRPCMessage()
	require.NoError(t, err)

	var queryResult NEARQueryResult
	err = json.Unmarshal(msg.Result, &queryResult)
	require.NoError(t, err)

	var oracleRequests map[string][]NEAROracleRequest
	err = json.Unmarshal(queryResult.Result, &oracleRequests)
	require.NoError(t, err)
	assert.NotNil(t, oracleRequests)
	assert.Equal(t, 2, len(oracleRequests))
}

func Test_ParseNEAROracleRequestsMap(t *testing.T) {
	msg, err := readTestJSONRPCMessage()
	require.NoError(t, err)
	oracleRequestsMap, err := ParseNEAROracleRequestsMap(*msg)
	require.NoError(t, err)
	assert.NotNil(t, oracleRequestsMap)
	assert.Equal(t, 2, len(oracleRequestsMap))
}

// Helper test function, to test ParseResponse for test JON-RPC Message using different filters
func testParseResponse(t *testing.T, filter nearFilter, n int) {
	type fields struct {
		filter         nearFilter
		connectionType subscriber.Type
	}
	type args struct {
		data []byte
	}

	msg, err := readTestJSONRPCMessage()
	require.NoError(t, err)
	msgBytes, err := json.Marshal(msg)

	tests := []struct {
		name   string
		fields fields
		args   args
		wantOk bool
	}{
		{
			"fails unmarshal payload",
			fields{filter: filter, connectionType: subscriber.RPC},
			args{[]byte(`error`)},
			false,
		},
		{
			"fails unmarshal result",
			fields{filter: filter, connectionType: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			false,
		},
		{
			"correctly generates 5 events",
			fields{filter: filter, connectionType: subscriber.RPC},
			args{msgBytes},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: &filter, connectionType: tt.fields.connectionType}
			events, ok := m.ParseResponse(tt.args.data)
			if ok != tt.wantOk {
				t.Errorf("ParseResponse() ok = %v, wantOk %v", ok, tt.wantOk)
			}
			if ok {
				assert.NotNil(t, events)
				assert.Equal(t, n, len(events))

				for _, e := range events {
					// check that we are able to unmarshal these bytes
					var data map[string]interface{}
					err = json.Unmarshal(e, &data)
					require.NoError(t, err)
					assert.NotNil(t, data)
					// check that every event holds five arguments
					assert.Equal(t, 5, len(data))
					assert.Contains(t, data, "account")
					assert.Contains(t, data, "nonce")
					assert.Contains(t, data, "get")
					assert.Contains(t, data, "path")
					assert.Contains(t, data, "times")
				}
			}
		})
	}
}

func Test_nearManager_ParseResponse_NoNoncesFilter(t *testing.T) {
	filter := nearFilter{
		JobID:      "mock",
		AccountIDs: []string{"oracle.chainlink.testnet"},
		Nonces:     NEAROracleNonces{},
	}
	testParseResponse(t, filter, 5)
}

func Test_nearManager_ParseResponse_ZeroNoncesFilter(t *testing.T) {
	filter := nearFilter{
		JobID:      "mock",
		AccountIDs: []string{"oracle.chainlink.testnet"},
		Nonces: NEAROracleNonces{
			"oracle.testnet":        "0",
			"client.oracle.testnet": "0",
		},
	}
	testParseResponse(t, filter, 5)
}

func Test_nearManager_ParseResponse_SomeNoncesFilter(t *testing.T) {
	filter := nearFilter{
		JobID:      "mock",
		AccountIDs: []string{"oracle.chainlink.testnet"},
		Nonces: NEAROracleNonces{
			"oracle.testnet":        "1",
			"client.oracle.testnet": "1",
		},
	}
	testParseResponse(t, filter, 3)
}
