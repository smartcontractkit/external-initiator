package terra

import (
	"math/big"
)

type Value big.Int

func (v *Value) UnmarshalJSON(data []byte) error {
	var i big.Int
	*v = Value(*i.SetBytes(data))

	return nil
}

type Addr string

type option struct {
	hasValue bool
}

func (o option) IsNone() bool {
	return !o.hasValue
}

func (o option) IsSome() bool {
	return o.hasValue
}

type OptionBigInt struct {
	option
	Value
}

type OptionU32 struct {
	option
	uint32
}

type OptionU64 struct {
	option
	uint64
}

type OptionAddr struct {
	option
	Addr
}

type FluxAggregatorConfig struct {
	Link               Addr   `json:"link"`
	Validator          Addr   `json:"validator"`
	PaymentAmount      Value  `json:"payment_amount"`
	MaxSubmissionCount uint32 `json:"max_submission_count"`
	MinSubmissionCount uint32 `json:"min_submission_count"`
	RestartDelay       uint32 `json:"restart_delay"`
	Timeout            uint32 `json:"timeout"`
	Decimals           uint8  `json:"decimals"`
	Description        string `json:"description"`
	MaxSubmissionValue Value  `json:"max_submission_value"`
	MinSubmissionValue Value  `json:"min_submission_value"`
}

type RoundData struct {
	RoundId         uint32
	Answer          OptionBigInt
	StartedAt       OptionU64
	UpdatedAt       OptionU64
	AnsweredInRound uint32
}

type OracleStatus struct {
	Withdrawable      Value        `json:"withdrawable"`
	StartingRound     uint32       `json:"starting_round"`
	EndingRound       uint32       `json:"ending_round"`
	LastReportedRound OptionU32    `json:"last_reported_round,omitempty"`
	LastStartedRound  OptionU32    `json:"last_started_round,omitempty"`
	LatestSubmission  OptionBigInt `json:"latest_submission,omitempty"`
	Index             uint16       `json:"index"`
	Admin             Addr         `json:"admin"`
	PendingAdmin      OptionAddr   `json:"pending_admin,omitempty"`
}

// Events
type EventNewRound struct {
	RoundId   uint32
	StartedBy Addr
	StartedAt OptionU64
}

type EventRoundDetailsUpdated struct {
	PaymentAmount  Value
	MinSubmissions uint32
	MaxSubmissions uint32
	RestartDelay   uint32
	Timeout        uint32
}

type EventOraclePermissionsUpdated struct {
	Oracle Addr
	Bool   bool
}

type EventAnswerUpdated struct {
	Value   Value
	RoundId uint32
}

type EventSubmissionReceived struct {
	Submission Value
	RoundId    uint32
	Oracle     Addr
}

type QueryResponse struct {
	Height string
	Result interface{}
}

type EventRecords struct {
	// FluxMonitor requests
	NewRound                 []EventNewRound
	RoundDetailsUpdated      []EventRoundDetailsUpdated
	OraclePermissionsUpdated []EventOraclePermissionsUpdated
	AnswerUpdated            []EventAnswerUpdated
	SubmissionReceived       []EventSubmissionReceived
}
