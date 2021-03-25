package blockchain

import (
	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
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
	value *types.AccountID
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
	value types.BlockNumber
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
	Timeout                 types.BlockNumber
	Decimals                types.U8
	Description             types.Text
	Restart_Delay           RoundID
	Reporting_Round         RoundID
	Latest_Round            RoundID
	First_Valid_Round       OptionRoundID
	Oracle_Count            types.U32
}

type Round struct {
	Started_At        types.BlockNumber
	Answer            OptionValue
	Updated_At        OptionBlockNumber
	Answered_In_Round OptionRoundID
}

type RoundDetails struct {
	Submissions             []Value
	Submission_Count_Bounds TupleU32
	Payment_Amount          Balance
	Timeout                 types.BlockNumber
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
	Started_At        types.BlockNumber
	Answer            Value
	Updated_At        types.BlockNumber
	Answered_In_Round RoundID
}

type EventNewRound struct {
	Phase       types.Phase
	FeedId      FeedId
	RoundId     RoundID
	AccountId   types.AccountID
	BlockNumber types.BlockNumber
	Topics      []types.Hash
}

type EventRoundDetailsUpdated struct {
	Phase            types.Phase
	FeedId           FeedId
	Balance          Balance
	SubmissionBounds TupleU32
	RoundId          RoundID
	BlockNumber      types.BlockNumber
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
	Value       Value
	BlockNumber types.BlockNumber
	Topics      []types.Hash
}

type EventRecords struct {
	types.EventRecords
	// FluxMonitor requests
	ChainlinkFeeds_NewRound                 []EventNewRound                 //nolint:stylecheck,golint
	ChainlinkFeeds_RoundDetailsUpdated      []EventRoundDetailsUpdated      //nolint:stylecheck,golint
	ChainlinkFeeds_OraclePermissionsUpdated []EventOraclePermissionsUpdated //nolint:stylecheck,golint
	ChainlinkFeeds_AnswerUpdated            []EventAnswerUpdated            //nolint:stylecheck,golint
}
