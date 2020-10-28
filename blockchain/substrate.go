package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/scale"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

// Substrate is the identifier of this
// blockchain integration.
const Substrate = "substrate"

type substrateFilter struct {
	JobID   types.Text
	Address []types.Address
}

type substrateManager struct {
	filter substrateFilter
	meta   *types.Metadata
	key    types.StorageKey
}

func createSubstrateManager(t subscriber.Type, conf store.Subscription) (*substrateManager, error) {
	if t != subscriber.WS {
		return nil, errors.New("only WS connections are allowed for Substrate")
	}

	var addresses []types.Address
	for _, id := range conf.Substrate.AccountIds {
		address, err := types.NewAddressFromHexAccountID(id)
		if err != nil {
			logger.Error(err)
			continue
		}
		addresses = append(addresses, address)
	}

	return &substrateManager{
		filter: substrateFilter{
			JobID:   types.NewText(conf.Job),
			Address: addresses,
		},
	}, nil
}

func (sm *substrateManager) GetTriggerJson() []byte {
	if sm.meta == nil {
		return nil
	}

	if len(sm.key) == 0 {
		key, err := types.CreateStorageKey(sm.meta, "System", "Events", nil, nil)
		if err != nil {
			logger.Error(err)
			return nil
		}
		sm.key = key
	}

	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "state_subscribeStorage",
	}

	keys := [][]string{{sm.key.Hex()}}
	params, err := json.Marshal(keys)
	if err != nil {
		logger.Error(err)
		return nil
	}
	msg.Params = params

	data, _ := json.Marshal(msg)
	return data
}

// SubstrateRequestParams allows for decoding a scale hex string into
// a byte array, which is then encoded back to a scale encoded byte array,
// to be decoded into a string array. This solves issues where decoding
// directly into a string array would read past the end of the array.
type SubstrateRequestParams []string

func (a *SubstrateRequestParams) Decode(decoder scale.Decoder) error {
	// Decode hex string into a byte array.
	// This allows us to stop reading where the
	// intended byte array stops.
	var bz types.Bytes
	err := decoder.Decode(&bz)
	if err != nil {
		return err
	}

	// Encode byte array into a scale encoded byte array
	encoded, err := types.EncodeToBytes(bz)
	if err != nil {
		return err
	}

	// Decode byte array into a string array
	var strings []string
	err = types.DecodeFromBytes(encoded, &strings)
	if err != nil {
		return err
	}

	*a = strings

	return nil
}

func (a SubstrateRequestParams) Encode(_ scale.Encoder) error {
	return nil
}

// EventChainlinkOracleRequest is the event structure we expect
// to be emitted from the Chainlink pallet
type EventChainlinkOracleRequest struct {
	Phase              types.Phase
	OracleAccountID    types.AccountID
	SpecIndex          types.Text
	RequestIdentifier  types.U64
	RequesterAccountID types.AccountID
	DataVersion        types.U64
	Bytes              SubstrateRequestParams
	Callback           types.Text
	Payment            types.U32
	Topics             []types.Hash
}

type EventChainlinkOracleAnswer struct {
	Phase              types.Phase
	OracleAccountID    types.AccountID
	RequestIdentifier  types.U64
	RequesterAccountID types.AccountID
	Bytes              types.Text
	Payment            types.U32
	Topics             []types.Hash
}

type EventChainlinkOperatorRegistered struct {
	Phase     types.Phase
	AccountID types.AccountID
	Topics    []types.Hash
}

type EventChainlinkOperatorUnregistered struct {
	Phase     types.Phase
	AccountID types.AccountID
	Topics    []types.Hash
}

type EventChainlinkKillRequest struct {
	Phase             types.Phase
	RequestIdentifier types.U64
	Topics            []types.Hash
}

type EventRecords struct {
	types.EventRecords
	Chainlink_OracleRequest        []EventChainlinkOracleRequest        //nolint:stylecheck,golint
	Chainlink_OracleAnswer         []EventChainlinkOracleAnswer         //nolint:stylecheck,golint
	Chainlink_OperatorRegistered   []EventChainlinkOperatorRegistered   //nolint:stylecheck,golint
	Chainlink_OperatorUnregistered []EventChainlinkOperatorUnregistered //nolint:stylecheck,golint
	Chainlink_KillRequest          []EventChainlinkKillRequest          //nolint:stylecheck,golint
}

type substrateSubscribeResponse struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

func (sm *substrateManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	var msg JsonrpcMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		logger.Error("Failed parsing JSON-RPC message:", err)
		return nil, false
	}

	var subRes substrateSubscribeResponse
	err = json.Unmarshal(msg.Params, &subRes)
	if err != nil {
		logger.Error("Failed parsing substrateSubscribeResponse:", err)
		return nil, false
	}

	var changes types.StorageChangeSet
	err = json.Unmarshal(subRes.Result, &changes)
	if err != nil {
		logger.Error("Failed parsing StorageChangeSet:", err)
		return nil, false
	}

	var subEvents []subscriber.Event
	for _, change := range changes.Changes {
		if !types.Eq(change.StorageKey, sm.key) || !change.HasStorageData {
			logger.Error("Does not match storage")
			continue
		}

		events := EventRecords{}
		err = types.EventRecordsRaw(change.StorageData).DecodeEventRecords(sm.meta, &events)
		if err != nil {
			logger.Errorw("Failed parsing EventRecords:",
				"err", err,
				"change.StorageData", change.StorageData,
				"sm.key", sm.key,
				"types.EventRecordsRaw", types.EventRecordsRaw(change.StorageData))
			continue
		}

		for _, request := range events.Chainlink_OracleRequest {
			// Check if our jobID matches
			jobID := fmt.Sprint(sm.filter.JobID)
			specIndex := fmt.Sprint(request.SpecIndex)
			if !matchesJobID(jobID, specIndex) {
				logger.Errorf("Does not match job : expected %s, requested %s", jobID, specIndex)
				continue
			}

			// Check if request is being sent from correct
			// oracle address
			found := false
			for _, address := range sm.filter.Address {
				if request.OracleAccountID == address.AsAccountID {
					found = true
					break
				}
			}
			if !found {
				logger.Errorf("Does not match OracleAccountID, requested is %s", request.OracleAccountID)
				continue
			}

			requestParams := convertStringArrayToKV(request.Bytes)
			requestParams["function"] = string(request.Callback)
			requestParams["request_id"] = fmt.Sprint(request.RequestIdentifier)
			requestParams["payment"] = fmt.Sprint(request.Payment)
			event, err := json.Marshal(requestParams)
			if err != nil {
				logger.Error(err)
				continue
			}
			subEvents = append(subEvents, event)
		}
	}

	return subEvents, true
}

func (sm *substrateManager) GetTestJson() []byte {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "state_getMetadata",
	}
	data, _ := json.Marshal(msg)
	return data
}

func (sm *substrateManager) ParseTestResponse(data []byte) error {
	var msg JsonrpcMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	var res string
	err = json.Unmarshal(msg.Result, &res)
	if err != nil {
		return err
	}

	var metadata types.Metadata
	err = types.DecodeFromHexString(res, &metadata)
	if err != nil {
		return err
	}

	sm.meta = &metadata
	return nil
}
