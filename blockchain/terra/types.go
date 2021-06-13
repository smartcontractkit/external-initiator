package terra

import (
	"math/big"
)

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
	big.Int
}

type OptionU64 struct {
	option
	uint64
}

type FluxAggregatorConfig struct {
	Link               Addr   `json:"link"`
	Validator          Addr   `json:"validator"`
	PaymentAmount      string `json:"payment_amount"`
	MaxSubmissionCount uint32 `json:"max_submission_count"`
	MinSubmissionCount uint32 `json:"min_submission_count"`
	RestartDelay       uint32 `json:"restart_delay"`
	Timeout            uint32 `json:"timeout"`
	Decimals           uint8  `json:"decimals"`
	Description        string `json:"description"`
	MaxSubmissionValue string `json:"max_submission_value"`
	MinSubmissionValue string `json:"min_submission_value"`
}

type RoundData struct {
	RoundId         uint32
	Answer          OptionBigInt
	StartedAt       OptionU64
	UpdatedAt       OptionU64
	AnsweredInRound uint32
}

// Events
type EventNewRound struct {
	RoundId   uint32
	StartedBy Addr
	StartedAt OptionU64
}

type EventRoundDetailsUpdated struct {
	PaymentAmount  big.Int
	MinSubmissions uint32
	MaxSubmissions uint32
	RestartDelay   uint32
	Timeout        uint32
}

type EventOraclePermissionsUpdated struct {
	// TODO!
	bool
}

type EventAnswerUpdated struct {
	Value   big.Int
	RoundId uint32
}

type EventSubmissionReceived struct {
	Submission big.Int
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
