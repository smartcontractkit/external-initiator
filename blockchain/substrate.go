package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

// Substrate is the identifier of this
// blockchain integration.
const Substrate = "substrate"

var (
	ErrorResultIsNull = errors.New("result is null")
)

type substrateManager struct {
	endpointName string

	meta      *types.Metadata
	feedId    FeedId
	accountId types.AccountID

	subscriber subscriber.ISubscriber
}

func createSubstrateManager(sub store.Subscription) (*substrateManager, error) {
	feedId := types.NewU32(sub.Substrate.FeedId)
	accountId, err := types.NewAddressFromHexAccountID(sub.Substrate.AccountId)
	if err != nil {
		return nil, err
	}

	conn, err := subscriber.NewSubscriber(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	return &substrateManager{
		endpointName: sub.EndpointName,
		feedId:       FeedId(feedId),
		accountId:    accountId.AsAccountID,
		subscriber:   conn,
	}, nil
}

func (sm substrateManager) Request(t string) (interface{}, error) {
	switch t {
	case FMRequestState:
		return sm.getFluxState()
	}
	return nil, errors.New("request type is not implemented")
}

func (sm substrateManager) Subscribe(t string, ch chan<- interface{}) error {
	switch t {
	case FMSubscribeEvents:
		return sm.SubscribeToFluxMonitor(ch)
	}
	return errors.New("subscribe type is not implemented")
}

func (sm substrateManager) CreateJobRun(t string, params interface{}) (map[string]interface{}, error) {
	switch t {
	case FMJobRun:
		fmState, ok := params.(FluxAggregatorState)
		if !ok {
			return nil, errors.New("params is not FluxAggregatorState")
		}

		return map[string]interface{}{
			"request_type": "fluxmonitor",
			"feed_id":      fmt.Sprintf("%d", sm.feedId),
			"round_id":     fmt.Sprintf("%d", fmState.CurrentRoundID),
		}, nil
	}

	return nil, errors.New("job run type not implemented")
}

type substrateSubscribeResponse struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
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
	var subRes substrateSubscribeResponse
	err := json.Unmarshal(data, &subRes)
	if err != nil {
		return nil, err
	}

	var changeSet types.StorageChangeSet
	err = json.Unmarshal(subRes.Result, &changeSet)
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
		err = types.EventRecordsRaw(change.StorageData).DecodeEventRecords(meta, &eventRecords)
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

// LEGACY: Getting the metadata

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

func subscribeToStorage(meta *types.Metadata, prefix, method string, args ...interface{}) (key types.StorageKey, m string, params json.RawMessage, err error) {
	m = "state_subscribeStorage"

	key, err = getStorageKey(meta, prefix, method, args...)
	if err != nil {
		return
	}

	keys := [][]string{{key.Hex()}}
	params, err = json.Marshal(keys)
	return
}

func (sm substrateManager) queryState(prefix, method string, t interface{}, args ...interface{}) error {
	key, rpcMethod, params, err := subscribeToStorage(sm.meta, prefix, method, args...)
	if err != nil {
		return err
	}

	responses := make(chan []byte)
	unsubscribe, err := sm.subscriber.Subscribe(rpcMethod, params, responses)
	if err != nil {
		return err
	}
	defer unsubscribe()

	for {
		response := <-responses
		err = parseChange(key, response, t)
		if err == ErrorResultIsNull {
			return err
		}
		if err != nil {
			logger.Error(err)
			continue
		}

		return nil
	}
}

func (sm substrateManager) getFluxState() (*FluxAggregatorState, error) {
	// Call chainlinkFeed.feeds(FeedId) to get the latest round
	// Return back `latest_round`
	// - payment
	// - timeout
	// - restart_delay
	// - submission bounds

	// "method": "state_subscribeStorage"
	var feedConfig FeedConfig
	err := sm.queryState("ChainlinkFeed", "Feeds", &feedConfig, sm.feedId)
	if err != nil {
		return nil, err
	}

	var round Round
	// TODO: Can be nil?
	err = sm.queryState("ChainlinkFeed", "Rounds", &round, sm.feedId, feedConfig.Latest_Round)
	if err != nil {
		return nil, err
	}

	roundId := int32(feedConfig.Latest_Round)
	var latestAnswer *decimal.Decimal
	if round.Answer.IsSome() {
		val := decimal.NewFromBigInt(round.Answer.value.Int, 10)
		latestAnswer = &val
	}

	minSubmission := decimal.NewFromBigInt(feedConfig.Submission_Value_Bounds.From.Int, 10)
	maxSubmission := decimal.NewFromBigInt(feedConfig.Submission_Value_Bounds.To.Int, 10)
	canSubmit := sm.oracleIsEligibleToSubmit()

	payment := feedConfig.Payment_Amount.Int
	timeout := uint32(feedConfig.Timeout)
	restartDelay := int32(feedConfig.Restart_Delay)

	return &FluxAggregatorState{
		CurrentRoundID: &roundId,
		LatestAnswer:   latestAnswer,
		MinSubmission:  &minSubmission,
		MaxSubmission:  &maxSubmission,
		Payment:        payment,
		Timeout:        &timeout,
		RestartDelay:   &restartDelay,
		CanSubmit:      &canSubmit,
	}, nil
}

func (sm substrateManager) oracleIsEligibleToSubmit() bool {
	var oracleStatus OracleStatus
	err := sm.queryState("ChainlinkFeed", "OracleStati", &oracleStatus, sm.feedId, sm.accountId)
	if err == ErrorResultIsNull {
		return false
	}
	if err != nil {
		logger.Error(err)
		return false
	}

	return oracleStatus.Ending_Round.IsNone()
}

func (sm substrateManager) SubscribeToFluxMonitor(ch chan<- interface{}) error {
	// Subscribe to events and watch for NewRound
	// This should increment the round ID on FM – but it should also return if this account was the initiator of the round
	// Also watch for OraclePermissionsUpdated - this tells if the oracle loses permission to submit new answers
	// Also needs to watch for RoundDetailsUpdated – this tells if timeout/restart delay/payment changes

	return noneShouldError(
		sm.subscribeNewRounds(ch),
		sm.subscribeAnswerUpdated(ch),
		sm.subscribeOraclePermissions(ch),
		sm.subscribeRoundDetailsUpdate(ch),
	)
}

func noneShouldError(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func (sm substrateManager) subscribe(method string, handler func(event EventRecords)) error {
	key, rpcMethod, params, err := subscribeToStorage(sm.meta, "ChainlinkFeed", method)
	if err != nil {
		return err
	}

	responses := make(chan []byte)
	unsubscribe, err := sm.subscriber.Subscribe(rpcMethod, params, responses)
	if err != nil {
		return err
	}

	go func() {
		defer unsubscribe()

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

func (sm substrateManager) subscribeNewRounds(ch chan<- interface{}) error {
	return sm.subscribe("NewRound", func(event EventRecords) {
		for _, round := range event.ChainlinkFeeds_NewRound {
			if round.FeedId != sm.feedId {
				continue
			}
			roundId := int32(round.RoundId)
			ch <- FluxAggregatorState{
				CurrentRoundID: &roundId,
			}
		}
	})
}

func (sm substrateManager) subscribeAnswerUpdated(ch chan<- interface{}) error {
	return sm.subscribe("AnswerUpdated", func(event EventRecords) {
		for _, update := range event.ChainlinkFeeds_AnswerUpdated {
			if update.FeedId != sm.feedId {
				continue
			}
			latestAnswer := decimal.NewFromBigInt(update.Value.Int, 10)
			ch <- FluxAggregatorState{
				LatestAnswer: &latestAnswer,
			}
		}
	})
}

func (sm substrateManager) subscribeOraclePermissions(ch chan<- interface{}) error {
	return sm.subscribe("OraclePermissionsUpdated", func(event EventRecords) {
		for _, update := range event.ChainlinkFeeds_OraclePermissionsUpdated {
			if update.FeedId != sm.feedId || update.AccountId != sm.accountId {
				continue
			}
			// TODO: Verify is correct
			canSubmit := bool(update.Bool)
			ch <- FluxAggregatorState{
				CanSubmit: &canSubmit,
			}
		}
	})
}

func (sm substrateManager) subscribeRoundDetailsUpdate(ch chan<- interface{}) error {
	return sm.subscribe("RoundDetailsUpdated", func(event EventRecords) {
		for _, update := range event.ChainlinkFeeds_RoundDetailsUpdated {
			if update.FeedId != sm.feedId {
				continue
			}
			// TODO: Anything to do here?
		}
	})
}
