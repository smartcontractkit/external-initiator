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
		return JsonrpcMessage{}, errors.New("Request for canned response did not return ok")
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
			require.Nil(t, err)
			assert.Equal(t, len(resp), 1)
			assert.Contains(t, string(resp[0].Result), "MethodNotFound")
		})
	}
}

func Test_handleNEARRequest_query_get_requests(t *testing.T) {
	type args struct {
		msg JsonrpcMessage
	}
	tests := []struct {
		name string
		args args
		want []JsonrpcMessage
	}{
		{
			"returns a get_requests query method result with the correct ID",
			args{
				JsonrpcMessage{ID: []byte(`123`), Method: "query", Params: []byte(`{"method_name": "get_requests"}`)},
			},
			[]JsonrpcMessage{
				{
					Version: "2.0",
					ID:      []byte(`123`),
					Result:  []byte(`"0x0"`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handleNEARRequest("rpc", tt.args.msg)
			require.NotNil(t, resp)
			require.Nil(t, err)
			assert.Equal(t, len(resp), 1)
			assert.Equal(t, resp[0].ID, tt.args.msg.ID)

			// Unmarshal and check result
			var result blockchain.NEARQueryResult
			err = json.Unmarshal(resp[0].Result, &result)
			require.NotNil(t, result)
			require.Nil(t, err)
		})
	}
}
