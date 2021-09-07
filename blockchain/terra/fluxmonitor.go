package terra

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/store"
)

type fluxMonitorManager struct {
	*manager
}

func CreateFluxMonitorManager(sub store.Subscription) (*fluxMonitorManager, error) {
	manager, err := createManager(sub)
	if err != nil {
		return nil, err
	}

	return &fluxMonitorManager{manager}, nil
}

func (fm fluxMonitorManager) GetState(ctx context.Context) (*common.FluxAggregatorState, error) {
	// get aggregator configs
	var config FluxAggregatorConfig
	err := fm.query(ctx, fm.contractAddress, `{"get_aggregator_config":{}}`, &config)
	if err != nil {
		return nil, err
	}

	// get last completed round data
	var round RoundData
	err = fm.query(ctx, fm.contractAddress, `{"get_latest_round_data":{}}`, &round)
	if err != nil {
		return nil, err
	}

	// get last oracle submission data (used to be oracleIsEligibleToSubmit())
	var status OracleStatus
	var canSubmit bool
	query := fmt.Sprintf(`{"get_oracle_status":{"oracle":"%s"}}`, fm.accountAddress)
	err = fm.query(ctx, fm.contractAddress, query, &status)
	if err != nil {
		logger.Error(err)
		canSubmit = false
	}
	canSubmit = status.EndingRound == 0xffffffff // uint32 max
	logger.Debugf("[terra/GetState/get_oracle_status] %+v", status)

	// if node has reported to a newer incomplete round use it's latest submission & round
	latestAnswer := round.Answer
	latestRound := round.RoundId
	if round.RoundId < status.LastReportedRound {
		latestRound = status.LastReportedRound
		latestAnswer = status.LatestSubmission
	}
	// if zero then set big.Int(0)
	if latestAnswer.String() == big.NewInt(0).String() {
		latestAnswer = Value{*big.NewInt(0)}
	}

	state := &common.FluxAggregatorState{
		RoundID:       latestRound,
		LatestAnswer:  latestAnswer.Int,
		Payment:       config.PaymentAmount.Int,
		Timeout:       config.Timeout,
		RestartDelay:  int32(config.RestartDelay),
		MinSubmission: config.MinSubmissionValue.Int,
		MaxSubmission: config.MaxSubmissionValue.Int,
		CanSubmit:     canSubmit,
		Decimals:      config.Decimals,
	}

	logger.Debugf("[terra/GetState] %+v", state)

	return state, nil
}

func (fm fluxMonitorManager) SubscribeEvents(ctx context.Context, ch chan<- interface{}) error {
	filter := fmt.Sprintf(`tm.event='Tx' AND execute_contract.contract_address='%s'`, fm.contractAddress)
	return fm.subscribe(ctx, filter, func(event EventRecords) {
		for _, round := range event.NewRound {
			ch <- common.FMEventNewRound{
				RoundID:         uint32(round.RoundId),
				OracleInitiated: string(round.StartedBy) == fm.accountAddress,
			}
		}
		for _, round := range event.SubmissionReceived {
			if round.Oracle != Addr(fm.accountAddress) {
				continue
			}
			ch <- common.FMSubmissionReceived{
				RoundID:      uint32(round.RoundId),
				LatestAnswer: round.Submission.Int,
			}
		}
		for _, update := range event.AnswerUpdated {
			ch <- common.FMEventAnswerUpdated{
				LatestAnswer: update.Value.Int,
			}
		}
		for _, update := range event.OraclePermissionsUpdated {
			if update.Oracle != Addr(fm.accountAddress) {
				continue
			}
			ch <- common.FMEventPermissionsUpdated{
				CanSubmit: update.Bool,
			}
		}
		for _, update := range event.RoundDetailsUpdated {
			ch <- common.FMEventRoundDetailsUpdated{
				Payment:      update.PaymentAmount.Int,
				Timeout:      update.Timeout,
				RestartDelay: int32(update.RestartDelay),
			}
		}
	})
}

func (fm fluxMonitorManager) CreateJobRun(roundId uint32) map[string]interface{} {
	return map[string]interface{}{
		"endpoint": "fluxmonitor",
		"address":  fm.contractAddress,
		"round_id": roundId,
	}
}
