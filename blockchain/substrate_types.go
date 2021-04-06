package blockchain

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
)

type option struct {
	hasValue bool
}

func (o option) IsNone() bool {
	return !o.hasValue
}

func (o option) IsSome() bool {
	return o.hasValue
}

type OptionAccountID struct {
	option
	value types.AccountID
}

func (o *OptionAccountID) Decode(decoder scale.Decoder) error {
	return decoder.DecodeOption(&o.hasValue, &o.value)
}

type FeedId types.U32
type RoundID types.U32

type OptionRoundID struct {
	option
	value RoundID
}

func (o *OptionRoundID) Decode(decoder scale.Decoder) error {
	return decoder.DecodeOption(&o.hasValue, &o.value)
}

type Value types.U128

func (v *Value) Decode(decoder scale.Decoder) error {
	var u128 types.U128
	err := decoder.Decode(&u128)
	if err != nil {
		return err
	}

	*v = Value(u128)
	return nil
}

type OptionValue struct {
	option
	value Value
}

func (o *OptionValue) Decode(decoder scale.Decoder) error {
	return decoder.DecodeOption(&o.hasValue, &o.value)
}

type Balance types.U128

func (b *Balance) Decode(decoder scale.Decoder) error {
	var u128 types.U128
	err := decoder.Decode(&u128)
	if err != nil {
		return err
	}

	*b = Balance(u128)
	return nil
}

type OptionBlockNumber struct {
	option
	value types.U32
}

func (o *OptionBlockNumber) Decode(decoder scale.Decoder) error {
	return decoder.DecodeOption(&o.hasValue, &o.value)
}

type TupleValue struct {
	From Value
	To   Value
}

type TupleU32 struct {
	From types.U32
	To   types.U32
}

type FeedConfig struct {
	Owner                   types.AccountID
	Pending_Owner           OptionAccountID
	Submission_Value_Bounds TupleValue
	Submission_Count_Bounds TupleU32
	Payment_Amount          Balance
	Timeout                 types.U32
	Decimals                types.U8
	Description             types.Bytes
	Restart_Delay           RoundID
	Reporting_Round         RoundID
	Latest_Round            RoundID
	First_Valid_Round       OptionRoundID
	Oracle_Count            types.U32
}

type Round struct {
	Started_At        types.U32
	Answer            OptionValue
	Updated_At        OptionBlockNumber
	Answered_In_Round OptionRoundID
}

type RoundDetails struct {
	Submissions             []Value
	Submission_Count_Bounds TupleU32
	Payment_Amount          Balance
	Timeout                 types.U32
}

type OracleMeta struct {
	Withdrawable  Balance
	Admin         types.AccountID
	Pending_Admin OptionAccountID
}

type OracleStatus struct {
	Starting_Round      RoundID
	Ending_Round        OptionRoundID
	Last_Reported_Round OptionRoundID
	Last_Started_Round  OptionRoundID
	Latest_Submission   OptionValue
}

type Requester struct {
	Delay              RoundID
	Last_Started_Round OptionRoundID
}

type RoundData struct {
	Started_At        types.U32
	Answer            Value
	Updated_At        types.U32
	Answered_In_Round RoundID
}

type EventNewRound struct {
	Phase       types.Phase
	FeedId      FeedId
	RoundId     RoundID
	AccountId   types.AccountID
	BlockNumber types.U32
	Topics      []types.Hash
}

type EventRoundDetailsUpdated struct {
	Phase            types.Phase
	FeedId           FeedId
	Balance          types.U128
	SubmissionBounds TupleU32
	RoundId          RoundID
	BlockNumber      types.U32
	Topics           []types.Hash
}

type EventOraclePermissionsUpdated struct {
	Phase     types.Phase
	FeedId    FeedId
	AccountId types.AccountID
	Bool      types.Bool
	Topics    []types.Hash
}

type EventAnswerUpdated struct {
	Phase       types.Phase
	FeedId      FeedId
	RoundId     RoundID
	Value       types.U128
	BlockNumber types.U32
	Topics      []types.Hash
}

type EventSubmissionReceived struct {
	Phase     types.Phase
	FeedId    FeedId
	RoundId   RoundID
	Value     types.U128
	AccountId types.AccountID
	Topics    []types.Hash
}

type EventRecords struct {
	types.EventRecords
	// FluxMonitor requests
	ChainlinkFeed_NewRound                 []EventNewRound                 //nolint:stylecheck,golint
	ChainlinkFeed_RoundDetailsUpdated      []EventRoundDetailsUpdated      //nolint:stylecheck,golint
	ChainlinkFeed_OraclePermissionsUpdated []EventOraclePermissionsUpdated //nolint:stylecheck,golint
	ChainlinkFeed_AnswerUpdated            []EventAnswerUpdated            //nolint:stylecheck,golint
	ChainlinkFeed_SubmissionReceived       []EventSubmissionReceived       //nolint:stylecheck,golint
}

// Copied from https://github.com/centrifuge/go-substrate-rpc-client/blob/904cb0b931a9949b08e731fa14c24ed62226a748/types/event_record.go#L221
// Changed so that any events that are not have defined types will just be skipped instead of erroring out prematurely.
func DecodeEventRecords(m *types.Metadata, e types.EventRecordsRaw, t interface{}) error {
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
	// ensure val points to a struct
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct, but is " + fmt.Sprint(typ))
	}

	decoder := scale.NewDecoder(bytes.NewReader(e))

	// determine number of events
	n, err := decoder.DecodeUintCompact()
	if err != nil {
		return err
	}

	// iterate over events
	for i := uint64(0); i < n.Uint64(); i++ {
		// decode Phase
		phase := types.Phase{}
		err := decoder.Decode(&phase)
		if err != nil {
			return fmt.Errorf("unable to decode Phase for event #%v: %v", i, err)
		}

		// decode EventID
		id := types.EventID{}
		err = decoder.Decode(&id)
		if err != nil {
			return fmt.Errorf("unable to decode EventID for event #%v: %v", i, err)
		}

		// ask metadata for method & event name for event
		moduleName, eventName, err := m.FindEventNamesForEventID(id)
		// moduleName, eventName, err := "System", "ExtrinsicSuccess", nil
		if err != nil {
			logger.Errorf("Unable to find %s (%s, %s, %s)", id, moduleName, eventName, err)
			continue
		}

		// check whether name for eventID exists in t
		field := val.FieldByName(fmt.Sprintf("%v_%v", moduleName, eventName))
		if !field.IsValid() {
			logger.Errorf("Event name is not valid (%v_%v): %s", moduleName, eventName, field.String())
			continue
		}

		// create a pointer to with the correct type that will hold the decoded event
		holder := reflect.New(field.Type().Elem())

		// ensure first field is for Phase, last field is for Topics
		numFields := holder.Elem().NumField()
		if numFields < 2 {
			return fmt.Errorf("expected event #%v with EventID %v, field %v_%v to have at least 2 fields "+
				"(for Phase and Topics), but has %v fields", i, id, moduleName, eventName, numFields)
		}
		phaseField := holder.Elem().FieldByIndex([]int{0})
		if phaseField.Type() != reflect.TypeOf(phase) {
			return fmt.Errorf("expected the first field of event #%v with EventID %v, field %v_%v to be of type "+
				"types.Phase, but got %v", i, id, moduleName, eventName, phaseField.Type())
		}
		topicsField := holder.Elem().FieldByIndex([]int{numFields - 1})
		if topicsField.Type() != reflect.TypeOf([]types.Hash{}) {
			return fmt.Errorf("expected the last field of event #%v with EventID %v, field %v_%v to be of type "+
				"[]types.Hash for Topics, but got %v", i, id, moduleName, eventName, topicsField.Type())
		}

		// set the phase we decoded earlier
		phaseField.Set(reflect.ValueOf(phase))

		// set the remaining fields
		for j := 1; j < numFields; j++ {
			err = decoder.Decode(holder.Elem().FieldByIndex([]int{j}).Addr().Interface())
			if err != nil {
				return fmt.Errorf("unable to decode field %v event #%v with EventID %v, field %v_%v: %v", j, i, id, moduleName,
					eventName, err)
			}
		}

		// add the decoded event to the slice
		field.Set(reflect.Append(field, holder.Elem()))
	}

	return nil
}
