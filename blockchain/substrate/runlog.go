package substrate

import (
	"context"
	"fmt"
	"github.com/smartcontractkit/external-initiator/store"

	"github.com/smartcontractkit/external-initiator/blockchain/common"

	"github.com/smartcontractkit/chainlink/core/logger"
)

type runlogManager struct {
	*manager
}

func CreateRunlogManager(sub store.Subscription) (*runlogManager, error) {
	manager, err := createManager(sub)
	if err != nil {
		return nil, err
	}

	return &runlogManager{manager}, nil
}

func (rm runlogManager) SubscribeEvents(ctx context.Context, ch chan<- common.RunlogRequest) error {
	return rm.subscribe(ctx, "System", "Events", func(event EventRecords) {
		for _, request := range event.Chainlink_OracleRequest {
			// Check if our jobID matches
			jobID := fmt.Sprint(rm.jobId)
			specIndex := fmt.Sprint(request.SpecIndex)
			if !common.MatchesJobID(jobID, specIndex) {
				logger.Errorf("Does not match job : expected %s, requested %s", jobID, specIndex)
				continue
			}

			// Check if request is being sent from correct
			// oracle address
			if request.OracleAccountID != rm.accountId {
				logger.Errorf("Does not match OracleAccountID, requested is %s", request.OracleAccountID)
				continue
			}

			ch <- common.MergeMaps(common.ConvertStringArrayToKV(request.Bytes), map[string]interface{}{
				"payment":    fmt.Sprint(request.Payment),
				"request_id": fmt.Sprint(request.RequestIdentifier),
				"function":   string(request.Callback),
			})
		}
	})
}

func (rm runlogManager) CreateJobRun(request common.RunlogRequest) map[string]interface{} {
	return common.MergeMaps(request, map[string]interface{}{
		"request_type": "runlog",
	})
}
