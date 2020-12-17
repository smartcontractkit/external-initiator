package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/external-initiator/blockchain"
)

const expectedStorageKey = "0x26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7"

func TestSubstrateMock_state_getMetadata(t *testing.T) {
	metadata, err := getMetadata()
	require.NoError(t, err)
	require.NotNil(t, metadata)

	key, err := types.CreateStorageKey(metadata, "System", "Events", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, expectedStorageKey, key.Hex())
}

func getMetadata() (*types.Metadata, error) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "state_getMetadata",
	}

	resp, ok := GetCannedResponse("substrate", req)
	if !ok {
		return nil, errors.New("Request for canned response did not return ok")
	}

	var result string
	err := json.Unmarshal(resp[0].Result, &result)
	if err != nil {
		return nil, err
	}

	var metadata types.Metadata
	err = types.DecodeFromHexString(result, &metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

type subscribeResponseParams struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

func TestSubstrateMock_state_subscribeStorage(t *testing.T) {
	req := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage("1"),
		Method:  "state_subscribeStorage",
	}

	resp, ok := GetCannedResponse("substrate", req)
	require.True(t, ok)

	// assert 2+ responses (subscription confirmation is the first one)
	assert.GreaterOrEqual(t, len(resp), 2)

	// get the subscription id number from the first response (subscription confirmation)
	var subscriptionNum string
	err := json.Unmarshal(resp[0].Result, &subscriptionNum)
	require.NoError(t, err)

	// require metadata for decoding
	metadata, err := getMetadata()
	require.NoError(t, err)

	for i := 1; i < len(resp); i++ {
		testName := fmt.Sprintf("Test JSON-RPC response #%d", i)
		t.Run(testName, func(t *testing.T) {
			// assert that subscription id numbers are consistent across responses
			var params subscribeResponseParams
			err = json.Unmarshal(resp[i].Params, &params)
			require.NoError(t, err)
			assert.Equal(t, subscriptionNum, params.Subscription)

			// assert that the response is for an expected StorageKey
			var changeSet types.StorageChangeSet
			err = json.Unmarshal(params.Result, &changeSet)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(changeSet.Changes), 1)
			assert.True(t, includesKeyInChanges(expectedStorageKey, changeSet.Changes))

			testEventRecordsDecoding(t, metadata, changeSet.Changes)
		})
	}
}

func includesKeyInChanges(expectedKey string, changes []types.KeyValueOption) bool {
	for _, change := range changes {
		if change.StorageKey.Hex() == expectedKey {
			return true
		}
	}
	return false
}

func testEventRecordsDecoding(t *testing.T, metadata *types.Metadata, changes []types.KeyValueOption) {
	for _, change := range changes {
		testName := fmt.Sprintf("Test decoding storage change %x", change.StorageKey)
		t.Run(testName, func(t *testing.T) {
			events := blockchain.EventRecords{}
			err := types.EventRecordsRaw(change.StorageData).DecodeEventRecords(metadata, &events)
			require.NoError(t, err)
			assert.NotNil(t, events)
		})
	}
}
