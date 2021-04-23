package substrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
)

// Name is the identifier of this
// blockchain integration.
const Name = "substrate"

var (
	ErrorResultIsNull = errors.New("result is null")
)

type Params struct {
	FeedId     *uint32  `json:"feed_id"`
	AccountId  string   `json:"account_id"`
	AccountIds []string `json:"accountIds"`
}

type manager struct {
	endpointName string

	meta      *types.Metadata
	feedId    FeedId
	accountId types.AccountID
	jobId     types.Text

	subscriber subscriber.ISubscriber
}

func createManager(sub store.Subscription) (*manager, error) {
	feedId := types.NewU32(sub.Substrate.FeedId)
	accountId, err := types.NewAddressFromHexAccountID(sub.Substrate.AccountId)
	if err != nil {
		return nil, err
	}

	conn, err := subscriber.NewSubscriber(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	return &manager{
		endpointName: sub.EndpointName,
		feedId:       FeedId(feedId),
		accountId:    accountId.AsAccountID,
		subscriber:   conn,
	}, nil
}

func (sm *manager) Stop() {
	// TODO: Implement me
}

func decodeStorageData(sd types.StorageDataRaw, t interface{}) error {
	// ensure t is a pointer
	ttyp := reflect.TypeOf(t)
	if ttyp.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer, but is " + fmt.Sprint(ttyp))
	}
	// ensure t is not a nil pointer
	tval := reflect.ValueOf(t)
	if tval.IsNil() {
		return errors.New("target is a nil pointer")
	}
	val := tval.Elem()
	typ := val.Type()
	// ensure val can be set
	if !val.CanSet() {
		return fmt.Errorf("unsettable value %v", typ)
	}

	decoder := scale.NewDecoder(bytes.NewReader(sd))
	return decoder.Decode(t)
}

func getChanges(key types.StorageKey, data []byte) ([]types.KeyValueOption, error) {
	var changeSet types.StorageChangeSet
	err := json.Unmarshal(data, &changeSet)
	if err != nil {
		return nil, err
	}

	var changes []types.KeyValueOption
	for _, change := range changeSet.Changes {
		if !types.Eq(change.StorageKey, key) {
			logger.Debugw("Does not match storage",
				"key", change.StorageKey.Hex(),
				"expects", key.Hex())
			continue
		}
		changes = append(changes, change)
	}

	return changes, nil
}

func parseChange(key types.StorageKey, data []byte, t interface{}) error {
	changes, err := getChanges(key, data)
	if err != nil {
		return err
	}

	if len(changes) != 1 {
		return errors.New("number of changes is not 1")
	}

	if len(changes[0].StorageData) == 0 {
		return ErrorResultIsNull
	}

	return decodeStorageData(changes[0].StorageData, t)
}

func parseEvents(meta *types.Metadata, key types.StorageKey, data []byte) ([]EventRecords, error) {
	changes, err := getChanges(key, data)
	if err != nil {
		return nil, err
	}

	var events []EventRecords
	for _, change := range changes {
		eventRecords := EventRecords{}
		raw := types.EventRecordsRaw(change.StorageData)
		err = DecodeEventRecords(meta, raw, &eventRecords)
		if err != nil {
			logger.Errorw("Failed parsing EventRecords:",
				"err", err,
				"change.StorageData", change.StorageData,
				"key", key.Hex(),
				"types.EventRecordsRaw", types.EventRecordsRaw(change.StorageData))
			continue
		}

		events = append(events, eventRecords)
	}

	return events, nil
}

func (sm *manager) getMetadata() (*types.Metadata, error) {
	if sm.meta != nil {
		return sm.meta, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := sm.subscriber.Request(ctx, "state_getMetadata", nil)
	if err != nil {
		return nil, err
	}

	var res string
	err = json.Unmarshal(response, &res)
	if err != nil {
		return nil, err
	}

	var metadata types.Metadata
	err = types.DecodeFromHexString(res, &metadata)
	if err != nil {
		return nil, err
	}

	sm.meta = &metadata

	return sm.meta, nil
}

// Events to send to FM
// - A new round has started
// - Oracle is now eligible to submit
// - Oracle is no longer eligible to submit

func getStorageKey(meta *types.Metadata, prefix, method string, args ...interface{}) (types.StorageKey, error) {
	if len(args) > 2 {
		return types.StorageKey{}, errors.New("too many arguments given")
	}

	var err error
	encoded := make([][]byte, 2)
	for i, arg := range args {
		encoded[i], err = types.EncodeToBytes(arg)
		if err != nil {
			return types.StorageKey{}, err
		}
	}

	return types.CreateStorageKey(meta, prefix, method, encoded[0], encoded[1])
}

func subscribeToStorage(meta *types.Metadata, prefix, method string, args ...interface{}) (key types.StorageKey, m, um string, params json.RawMessage, err error) {
	m = "state_subscribeStorage"
	um = "state_unsubscribeStorage"

	key, err = getStorageKey(meta, prefix, method, args...)
	if err != nil {
		return
	}

	keys := [][]string{{key.Hex()}}
	params, err = json.Marshal(keys)
	return
}

func (sm *manager) queryState(ctx context.Context, prefix, method string, t interface{}, args ...interface{}) error {
	meta, err := sm.getMetadata()
	if err != nil {
		return err
	}
	key, rpcMethod, unsubscribeMethod, params, err := subscribeToStorage(meta, prefix, method, args...)
	if err != nil {
		return err
	}

	responses := make(chan json.RawMessage)
	err = sm.subscriber.Subscribe(ctx, rpcMethod, unsubscribeMethod, params, responses)
	if err != nil {
		return err
	}

	for {
		select {
		case response := <-responses:
			err = parseChange(key, response, t)
			if err == ErrorResultIsNull {
				return err
			}
			if err != nil {
				logger.Error(err)
				continue
			}

			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (sm *manager) subscribe(ctx context.Context, prefix, method string, handler func(event EventRecords)) error {
	meta, err := sm.getMetadata()
	if err != nil {
		return err
	}
	key, rpcMethod, unsubscribeMethod, params, err := subscribeToStorage(meta, prefix, method)
	if err != nil {
		return err
	}

	responses := make(chan json.RawMessage)
	err = sm.subscriber.Subscribe(ctx, rpcMethod, unsubscribeMethod, params, responses)
	if err != nil {
		return err
	}

	go func() {
		for {
			response, ok := <-responses
			if !ok {
				return
			}

			events, err := parseEvents(sm.meta, key, response)
			if err != nil {
				logger.Error(err)
				continue
			}

			for _, event := range events {
				handler(event)
			}
		}
	}()

	return nil
}
