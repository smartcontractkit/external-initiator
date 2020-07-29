package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/external-initiator/blockchain"
)

func TestNEARMock_status(t *testing.T) {
	status, err := getStatus()
	require.NoError(t, err)
	require.NotNil(t, status)
}

func TestNEARMock_status_Unmarshal(t *testing.T) {
	msg, err := getStatus()
	require.NoError(t, err)
	require.NotNil(t, msg)

	// Unmarshal to NEAR NEARStatus type
	var res blockchain.NEARStatus
	err = json.Unmarshal(msg.Result, &res)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func getStatus() (JsonrpcMessage, error) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "status",
	}

	resp, ok := GetCannedResponse("near", req)
	if !ok {
		return JsonrpcMessage{}, errors.New("request for canned response did not return ok")
	}

	return resp[0], nil
}

func Test_handleNEARRequest_unexpected_connection(t *testing.T) {
	type args struct {
		conn string
		msg  JsonrpcMessage
	}
	testRPCMsg := JsonrpcMessage{ID: []byte(`123`), Method: "query"}
	tests := []struct {
		name string
		args args
	}{
		{
			"returns an error for json-rpc connection",
			args{"json-rpc", testRPCMsg},
		},
		{
			"returns an error for ws connection",
			args{"ws", testRPCMsg},
		},
		{
			"returns an error for unknown connection",
			args{"unknown", testRPCMsg},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handleNEARRequest(tt.args.conn, tt.args.msg)
			expectedErrorString := fmt.Sprintf("unexpected connection: %s", tt.args.conn)
			require.EqualError(t, err, expectedErrorString)
		})
	}
}

func Test_handleNEARRequest_unexpected_method(t *testing.T) {
	type args struct {
		conn string
		msg  JsonrpcMessage
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"returns an error for json-rpc method: view",
			args{"rpc", JsonrpcMessage{ID: []byte(`123`), Method: "view"}},
		},
		{
			"returns an error for json-rpc method: metadata",
			args{"rpc", JsonrpcMessage{ID: []byte(`123`), Method: "metadata"}},
		},
		{
			"returns an error for json-rpc method: withdraw",
			args{"rpc", JsonrpcMessage{ID: []byte(`123`), Method: "withdraw"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handleNEARRequest(tt.args.conn, tt.args.msg)
			expectedErrorString := fmt.Sprintf("unexpected method: %s", tt.args.msg.Method)
			require.EqualError(t, err, expectedErrorString)
		})
	}
}

func Test_handleNEARRequest_error_MethodNotFound(t *testing.T) {
	type args struct {
		conn string
		msg  JsonrpcMessage
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"returns an json-rpc error MethodNotFound: get_nonexistant",
			args{"rpc", JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_nonexistant"}`)}},
		},
		{
			"returns an json-rpc error MethodNotFound: get_123",
			args{"rpc", JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_123"}`)}},
		},
		{
			"returns an json-rpc error MethodNotFound: get_something",
			args{"rpc", JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_something"}`)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handleNEARRequest(tt.args.conn, tt.args.msg)
			require.NotNil(t, resp)
			require.NoError(t, err)
			assert.Equal(t, len(resp), 1)
			assert.Contains(t, string(resp[0].Result), "MethodNotFound")
		})
	}
}

func Test_buildResponseID(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"returns an error, fails to build ID from JSON-RPC message",
			args{JsonrpcMessage{ID: []byte(`123`)}},
			"",
			true,
		},
		{
			"returns an error, fails to parse JSON-RPC Params",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "status", Params: []byte(`!#$`)}},
			"",
			true,
		},
		{
			"returns a 'status' as ID",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "status", Params: []byte(`{}`)}},
			"status",
			false,
		},
		{
			"returns a 'query_get_nonces' as ID",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_nonces"}`)}},
			"query_get_nonces",
			false,
		},
		{
			"returns a 'query_get_requests' as ID",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_requests"}`)}},
			"query_get_requests",
			false,
		},
		{
			"returns a 'query_get_all_requests' as ID",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_all_requests"}`)}},
			"query_get_all_requests",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := buildResponseID(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildResponseID() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, resp)
		})
	}
}

func Test_handleNEARRequest(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"returns an error, fails to build ID from JSON-RPC message",
			args{JsonrpcMessage{ID: []byte(`123`)}},
			true,
		},
		{
			"returns an error, fails to parse JSON-RPC Params",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "status", Params: []byte(`!#$`)}},
			true,
		},
		{
			"returns a 'query_get_nonces' canned response",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_nonces"}`)}},
			false,
		},
		{
			"returns a 'query_get_requests' canned response",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_requests"}`)}},
			false,
		},
		{
			"returns a 'query_get_all_requests' canned response",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_all_requests"}`)}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handleNEARRequest("rpc", tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleNEARRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleNEARRequest_query_get_nonces(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"returns a get_nonces query method result with the correct ID",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_nonces"}`)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handleNEARRequest("rpc", tt.args.msg)
			require.NotNil(t, resp)
			require.NoError(t, err)
			assert.Equal(t, len(resp), 1)
			assert.Equal(t, resp[0].ID, tt.args.msg.ID)

			// Unmarshal and check result
			nonces, err := blockchain.ParseNEARNEAROracleNonces(resp[0])
			require.NoError(t, err)
			assert.NotNil(t, nonces)
			assert.Equal(t, 2, len(nonces))
		})
	}
}

func Test_handleNEARRequest_query_get_all_requests(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"returns a get_all_requests query method result with the correct ID",
			args{JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_all_requests"}`)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handleNEARRequest("rpc", tt.args.msg)
			require.NotNil(t, resp)
			require.NoError(t, err)
			assert.Equal(t, len(resp), 1)
			assert.Equal(t, resp[0].ID, tt.args.msg.ID)

			// Unmarshal and check result
			oracleRequestsMap, err := blockchain.ParseNEAROracleRequestsMap(resp[0])
			require.NoError(t, err)
			assert.NotNil(t, oracleRequestsMap)
			assert.Equal(t, 2, len(oracleRequestsMap))
		})
	}
}
