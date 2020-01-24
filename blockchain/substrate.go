package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

// Substrate is the identifier of this
// blockchain integration.
const Substrate = "substrate"

type substrateFilter struct {
	Address []types.Address
}

type SubstrateManager struct {
	filter substrateFilter
	meta   *types.Metadata
	key    types.StorageKey
}

func CreateSubstrateManager(conf store.SubstrateSubscription) *SubstrateManager {
	var addresses []types.Address
	for _, id := range conf.AccountIds {
		address, err := types.NewAddressFromHexAccountID(id)
		if err != nil {
			fmt.Println(err)
			continue
		}
		addresses = append(addresses, address)
	}

	return &SubstrateManager{
		filter: substrateFilter{
			Address: addresses,
		},
	}
}

func (sm *SubstrateManager) GetTriggerJson() []byte {
	if sm.meta == nil {
		return nil
	}

	if len(sm.key) == 0 {
		key, err := types.CreateStorageKey(sm.meta, "System", "Events", nil, nil)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		sm.key = key
	}

	msg := jsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "state_subscribeStorage",
	}

	keys := [][]string{{sm.key.Hex()}}
	params, err := json.Marshal(keys)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	msg.Params = params

	data, _ := json.Marshal(msg)
	return data
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EventChainlinkOracleRequest is the event structure we expect
// to be emitted from the Chainlink pallet
type EventChainlinkOracleRequest struct {
	Phase             types.Phase
	SpecIndex         types.U32
	RequestIdentifier types.U64
	AccountID         types.AccountID
	DataVersion       types.U64
	Bytes             []types.Bytes
	Topics            []types.Hash
}

type EventRecords struct {
	types.EventRecords
	Chainlink_OracleRequest []EventChainlinkOracleRequest //nolint:stylecheck,golint
}

type substrateSubscribeResponse struct {
	Subscription int             `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

func (sm *SubstrateManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	var msg jsonrpcMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println("Failed parsing JSON-RPC message:", err)
		return nil, false
	}

	var subRes substrateSubscribeResponse
	err = json.Unmarshal(msg.Params, &subRes)
	if err != nil {
		fmt.Println("Failed parsing substrateSubscribeResponse:", err)
		return nil, false
	}

	var changes types.StorageChangeSet
	err = json.Unmarshal(subRes.Result, &changes)
	if err != nil {
		fmt.Println("Failed parsing StorageChangeSet:", err)
		return nil, false
	}

	var subEvents []subscriber.Event
	for _, change := range changes.Changes {
		if !types.Eq(change.StorageKey, sm.key) || !change.HasStorageData {
			continue
		}

		events := EventRecords{}
		err = types.EventRecordsRaw(change.StorageData).DecodeEventRecords(sm.meta, &events)
		if err != nil {
			fmt.Println("Failed parsing EventRecords:", err)
			// Do not stop after this error because it could still work
			// TODO: investigate this error: "unable to decode Phase for event #2: EOF"
			// continue
		}

		for _, request := range events.Chainlink_OracleRequest {
			found := false
			for _, address := range sm.filter.Address {
				if request.AccountID == address.AsAccountID {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			requestParams := convertByteArraysToKV(request.Bytes)
			event, err := json.Marshal(requestParams)
			if err != nil {
				fmt.Println(err)
				continue
			}
			subEvents = append(subEvents, event)
		}
	}

	return subEvents, true
}

func (sm *SubstrateManager) GetTestJson() []byte {
	msg := jsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "state_getMetadata",
	}
	data, _ := json.Marshal(msg)
	return data
}

func (sm *SubstrateManager) ParseTestResponse(data []byte) error {
	var msg jsonrpcMessage
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

func convertByteArraysToKV(data []types.Bytes) []KeyValue {
	var result []KeyValue
	var keyValue KeyValue

	for i := 0; i < len(data); i++ {
		val := string(data[i])
		if i%2 == 0 {
			if val == "" {
				i++
				continue
			}
			keyValue.Key = val
		} else {
			keyValue.Value = val
			result = append(result, keyValue)
			keyValue = KeyValue{}
		}
	}

	return result
}
