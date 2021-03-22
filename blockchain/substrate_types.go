package blockchain

import (
	"fmt"

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

var opt types.OptionBytes

type OptionAccountID struct {
	option
	value *types.AccountID
}

func (o OptionAccountID) Encode(encoder scale.Encoder) error {
	return encoder.EncodeOption(o.hasValue, o.value)
}

func (o *OptionAccountID) Decode(decoder scale.Decoder) error {
	fmt.Println("Decoding OptionAccountID")
	return decoder.DecodeOption(&o.hasValue, &o.value)
}

func (o *OptionAccountID) SetSome(value types.AccountID) {
	o.hasValue = true
	o.value = &value
}

func (o *OptionAccountID) SetNone() {
	o.hasValue = false
	o.value = nil
}

func (o OptionAccountID) Unwrap() (ok bool, value types.AccountID) {
	return o.hasValue, *o.value
}

type FeedId types.U32
type RoundID types.U32

type OptionRoundID struct {
	option
	value RoundID
}

func (o OptionRoundID) Encode(encoder scale.Encoder) error {
	return encoder.EncodeOption(o.hasValue, o.value)
}

func (o *OptionRoundID) Decode(decoder scale.Decoder) error {
	fmt.Println("Decoding OptionRoundID")
	return decoder.DecodeOption(&o.hasValue, &o.value)
}

func (o *OptionRoundID) SetSome(value RoundID) {
	o.hasValue = true
	o.value = value
}

func (o *OptionRoundID) SetNone() {
	o.hasValue = false
	o.value = RoundID(0)
}

func (o OptionRoundID) Unwrap() (ok bool, value RoundID) {
	return o.hasValue, o.value
}

type Value types.U128
type Balance types.U128

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

type FeedConfigOf FeedConfig

type OptionFeedConfigOf struct {
	option
	value FeedConfigOf
}

func (o OptionFeedConfigOf) Encode(encoder scale.Encoder) error {
	return encoder.EncodeOption(o.hasValue, o.value)
}

func (o *OptionFeedConfigOf) Decode(decoder scale.Decoder) error {
	fmt.Println("Decoding OptionFeedConfigOf")
	return decoder.DecodeOption(&o.hasValue, &o.value)
}

func (o *OptionFeedConfigOf) SetSome(value FeedConfigOf) {
	o.hasValue = true
	o.value = value
}

func (o *OptionFeedConfigOf) SetNone() {
	o.hasValue = false
	o.value = FeedConfigOf{}
}

func (o OptionFeedConfigOf) Unwrap() (ok bool, value FeedConfigOf) {
	return o.hasValue, o.value
}

type Round struct {
	Started_At        types.BlockNumber
	Answer            *Value
	Updated_At        *types.BlockNumber
	Answered_In_Round *RoundID
}

type RoundOf Round

type RoundDetails struct {
	Submissions             []Value
	Submission_Count_Bounds []types.U32
	Payment_Amount          Balance
	Timeout                 types.BlockNumber
}

type RoundDetailsOf RoundDetails

type OracleMeta struct {
	Withdrawable  Balance
	Admin         types.AccountID
	Pending_Admin *types.AccountID
}

type OracleMetaOf OracleMeta

type OracleStatus struct {
	Starting_Round      RoundID
	Ending_Round        *RoundID
	Last_Reported_Round *RoundID
	Last_Started_Round  *RoundID
	Latest_Submission   *Value
}

type OracleStatusOf OracleStatus

type Requester struct {
	Delay              RoundID
	Last_Started_Round *RoundID
}

type RequesterOf Requester

type RoundData struct {
	Started_At        types.BlockNumber
	Answer            Value
	Updated_At        types.BlockNumber
	Answered_In_Round RoundID
}

type RoundDataOf RoundData

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
	SubmissionBounds []types.U32
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
