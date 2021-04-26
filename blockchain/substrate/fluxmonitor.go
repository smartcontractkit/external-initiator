package substrate

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/store"

	"github.com/smartcontractkit/chainlink/core/logger"
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

func (fm fluxMonitorManager) CreateJobRun(roundId uint32) map[string]interface{} {
	return map[string]interface{}{
		"request_type": "fluxmonitor",
		"feed_id":      fmt.Sprintf("%d", fm.feedId),
		"round_id":     fmt.Sprintf("%d", roundId),
	}
}

func (fm fluxMonitorManager) GetState(ctx context.Context) (*common.FluxAggregatorState, error) {
	var feedConfig FeedConfig
	err := fm.queryState(ctx, "ChainlinkFeed", "Feeds", &feedConfig, fm.feedId)
	if err != nil {
		return nil, err
	}

	var round Round
	err = fm.queryState(ctx, "ChainlinkFeed", "Rounds", &round, fm.feedId, feedConfig.Latest_Round)
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
		CanSubmit:     fm.oracleIsEligibleToSubmit(),
	}, nil
}

func (fm fluxMonitorManager) oracleIsEligibleToSubmit() bool {
	var oracleStatus OracleStatus
	err := fm.queryState(context.TODO(), "ChainlinkFeed", "OracleStatuses", &oracleStatus, fm.feedId, fm.accountId)
	if err == ErrorResultIsNull {
		return false
	}
	if err != nil {
		logger.Error(err)
		return false
	}

	return oracleStatus.Ending_Round.IsNone()
}

func (fm fluxMonitorManager) SubscribeEvents(ctx context.Context, ch chan<- interface{}) error {
	return fm.subscribe(ctx, "System", "Events", func(event EventRecords) {
		for _, round := range event.ChainlinkFeed_NewRound {
			if round.FeedId != fm.feedId {
				continue
			}
			ch <- common.FMEventNewRound{
				RoundID:         uint32(round.RoundId),
				OracleInitiated: round.AccountId == fm.accountId && !common.ExpectsMock,
			}
		}
		for _, update := range event.ChainlinkFeed_AnswerUpdated {
			if update.FeedId != fm.feedId {
				continue
			}
			ch <- common.FMEventAnswerUpdated{
				LatestAnswer: *update.Value.Int,
			}
		}
		for _, update := range event.ChainlinkFeed_OraclePermissionsUpdated {
			if update.FeedId != fm.feedId || update.AccountId != fm.accountId {
				continue
			}
			ch <- common.FMEventPermissionsUpdated{
				CanSubmit: bool(update.Bool),
			}
		}
		for _, update := range event.ChainlinkFeed_RoundDetailsUpdated {
			if update.FeedId != fm.feedId {
				continue
			}
			// TODO: Anything to do here?
		}
	})
}
