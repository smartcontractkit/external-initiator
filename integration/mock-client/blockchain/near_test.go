package blockchain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNEARMock_status(t *testing.T) {
	status, err := getStatus()
	require.NoError(t, err)
	require.NotNil(t, status)
}

func TestNEARMock_status_Unmarshal(t *testing.T) {
	status, err := getStatus()
	require.NoError(t, err)
	require.NotNil(t, status)

	//TODO: Unmarshal to near.Status type
}

func getStatus() ([]JsonrpcMessage, error) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "status",
	}

	resp, ok := GetCannedResponse("near", req)
	if !ok {
		return nil, errors.New("Request for canned response did not return ok")
	}

	return resp, nil
}
