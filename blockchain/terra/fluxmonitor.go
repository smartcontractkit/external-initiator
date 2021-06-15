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
	var config FluxAggregatorConfig
	err := fm.query(ctx, fm.contractAddress, `{"get_aggregator_config":{}}`, &config)
	if err != nil {
		return nil, err
	}

	var round RoundData
	err = fm.query(ctx, fm.contractAddress, `{"get_latest_round_data":{}}`, &round)
	if err != nil {
		return nil, err
	}

	var latestAnswer big.Int
	if round.Answer.IsSome() {
		latestAnswer = round.Answer.Int
	} else {
		latestAnswer = *big.NewInt(0)
	}

	// TODO! move out parsing logic from here
	payment := new(big.Int)
	payment, _ = payment.SetString(config.PaymentAmount, 10)
	min := new(big.Int)
	min, _ = min.SetString(config.MinSubmissionValue, 10)
	max := new(big.Int)
	max, _ = max.SetString(config.MaxSubmissionValue, 10)
	state := &common.FluxAggregatorState{
		RoundID:       round.RoundId,
		LatestAnswer:  latestAnswer,
		Payment:       *payment,
		Timeout:       config.Timeout,
		RestartDelay:  int32(config.RestartDelay),
		MinSubmission: *min,
		MaxSubmission: *max,
		CanSubmit:     fm.oracleIsEligibleToSubmit(ctx),
	}

	return state, nil
}

func (fm fluxMonitorManager) oracleIsEligibleToSubmit(ctx context.Context) bool {
	var status OracleStatus
	query := fmt.Sprintf(`{"get_oracle_status":{"oracle":"%s"}}`, fm.accountAddress)
	err := fm.query(ctx, fm.contractAddress, query, &status)
	if err != nil {
		logger.Error(err)
		return false
	}

	return status.EndingRound > 0
}

func (fm fluxMonitorManager) SubscribeEvents(ctx context.Context, ch chan<- interface{}) error {
	filter := fmt.Sprintf("tm.event='Tx' AND execute_contract.contract_address='%s'", fm.contractAddress)
	return fm.subscribe(ctx, filter, func(event EventRecords) {
		for _, round := range event.NewRound {
			ch <- common.FMEventNewRound{
				RoundID:         uint32(round.RoundId),
				OracleInitiated: string(round.StartedBy) == fm.accountAddress,
			}
		}
		for _, update := range event.AnswerUpdated {
			ch <- common.FMEventAnswerUpdated{
				LatestAnswer: update.Value,
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
		// for _, update := range event.RoundDetailsUpdated {
		// 	if update.FeedId != fm.feedId {
		// 		continue
		// 	}
		// }
	})
}

func (fm fluxMonitorManager) CreateJobRun(roundId uint32) map[string]interface{} {
	return map[string]interface{}{
		"endpoint": "fluxmonitor",
		"address":  fm.contractAddress,
		"round_id": fmt.Sprintf("%d", roundId),
	}
}
