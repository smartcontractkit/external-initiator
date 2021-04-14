package substrate

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/external-initiator/blockchain/common"

	"github.com/smartcontractkit/chainlink/core/logger"
)

func (sm *substrateManager) getFluxState() (*common.FluxAggregatorState, error) {
	var feedConfig FeedConfig
	err := sm.queryState("ChainlinkFeed", "Feeds", &feedConfig, sm.feedId)
	if err != nil {
		return nil, err
	}

	var round Round
	err = sm.queryState("ChainlinkFeed", "Rounds", &round, sm.feedId, feedConfig.Latest_Round)
	if err != nil && err != ErrorResultIsNull {
		return nil, err
	}

	var latestAnswer big.Int
	if round.Answer.IsSome() {
		latestAnswer = *round.Answer.value.Int
	} else {
		latestAnswer = *big.NewInt(0)
	}

	return &common.FluxAggregatorState{
		RoundID:       uint32(feedConfig.Latest_Round),
		LatestAnswer:  latestAnswer,
		MinSubmission: *feedConfig.Submission_Value_Bounds.From.Int,
		MaxSubmission: *feedConfig.Submission_Value_Bounds.To.Int,
		Payment:       *feedConfig.Payment_Amount.Int,
		Timeout:       uint32(feedConfig.Timeout),
		RestartDelay:  int32(feedConfig.Restart_Delay),
		CanSubmit:     sm.oracleIsEligibleToSubmit(),
	}, nil
}

func (sm *substrateManager) oracleIsEligibleToSubmit() bool {
	var oracleStatus OracleStatus
	err := sm.queryState("ChainlinkFeed", "OracleStatuses", &oracleStatus, sm.feedId, sm.accountId)
	if err == ErrorResultIsNull {
		return false
	}
	if err != nil {
		logger.Error(err)
		return false
	}

	return oracleStatus.Ending_Round.IsNone()
}

func (sm *substrateManager) SubscribeToFluxMonitor(ctx context.Context, ch chan<- interface{}) error {
	return sm.subscribe(ctx, "System", "Events", func(event EventRecords) {
		for _, round := range event.ChainlinkFeed_NewRound {
			if round.FeedId != sm.feedId {
				continue
			}
			ch <- common.FMEventNewRound{
				RoundID:         uint32(round.RoundId),
				OracleInitiated: round.AccountId == sm.accountId && !common.ExpectsMock,
			}
		}
		for _, update := range event.ChainlinkFeed_AnswerUpdated {
			if update.FeedId != sm.feedId {
				continue
			}
			ch <- common.FMEventAnswerUpdated{
				LatestAnswer: *update.Value.Int,
			}
		}
		for _, update := range event.ChainlinkFeed_OraclePermissionsUpdated {
			if update.FeedId != sm.feedId || update.AccountId != sm.accountId {
				continue
			}
			ch <- common.FMEventPermissionsUpdated{
				CanSubmit: bool(update.Bool),
			}
		}
		for _, update := range event.ChainlinkFeed_RoundDetailsUpdated {
			if update.FeedId != sm.feedId {
				continue
			}
			// TODO: Anything to do here?
		}
	})
}
