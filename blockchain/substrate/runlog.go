package substrate

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/external-initiator/blockchain/common"

	"github.com/smartcontractkit/chainlink/core/logger"
)

func (sm substrateManager) Backfill() (interface{}, error) {
	// Currently not implemented
	return nil, nil
}

func (sm *substrateManager) SubscribeToRunlog(ctx context.Context, ch chan<- interface{}) error {
	return sm.subscribe(ctx, "System", "Events", func(event EventRecords) {
		for _, request := range event.Chainlink_OracleRequest {
			// Check if our jobID matches
			jobID := fmt.Sprint(sm.jobId)
			specIndex := fmt.Sprint(request.SpecIndex)
			if !common.MatchesJobID(jobID, specIndex) {
				logger.Errorf("Does not match job : expected %s, requested %s", jobID, specIndex)
				continue
			}

			// Check if request is being sent from correct
			// oracle address
			if request.OracleAccountID != sm.accountId {
				logger.Errorf("Does not match OracleAccountID, requested is %s", request.OracleAccountID)
				continue
			}

			ch <- common.RunlogRequest{
				Params:           common.ConvertStringArrayToKV(request.Bytes),
				Payment:          fmt.Sprint(request.Payment),
				RequestId:        fmt.Sprint(request.RequestIdentifier),
				CallbackFunction: string(request.Callback),
			}
		}
	})
}
