package agoric

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/store"

	"github.com/smartcontractkit/chainlink/core/logger"
)

// linkDecimals is the number of decimal places in $LINK
// This value must be greater than linkAgoricDecimals
const linkDecimals = 18

// linkAgoricDecimals is the number of decimal places in a uaglink token
// FIXME: Ideally the same as linkDecimals.
// This value must be lower than linkDecimals
const linkAgoricDecimals = 6

var (
	errNoJobMatch = errors.New("event did not match a job")
)

type runlogManager struct {
	*manager
}

func CreateRunlogManager(sub store.Subscription) (*runlogManager, error) {
	manager, err := createManager(sub)
	if err != nil {
		return nil, err
	}

	return &runlogManager{
		manager: manager,
	}, nil
}

func (r runlogManager) SubscribeEvents(ctx context.Context, ch chan<- common.RunlogRequest) error {
	msgs := make(chan []byte)
	go r.conn.Read(msgs)

	go func() {
		for {
			select {
			case msg := <-msgs:
				req, err := r.parseRequests(msg)
				if err == errNoJobMatch {
					continue
				} else if err != nil {
					logger.Error(err)
					continue
				}
				ch <- req
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r runlogManager) CreateJobRun(request common.RunlogRequest) map[string]interface{} {
	// This implementation does not need to make any changes
	// to the request payload.
	return request
}

type agoricEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type agoricOnQueryData struct {
	QueryID string          `json:"queryId"`
	Query   json.RawMessage `json:"query"`
	Fee     int64           `json:"fee"`
}

type chainlinkQuery struct {
	JobID  string                 `json:"jobid"`
	Params map[string]interface{} `json:"params"`
}

func (r runlogManager) parseRequests(data []byte) (common.RunlogRequest, error) {
	var agEvent agoricEvent
	err := json.Unmarshal(data, &agEvent)
	if err != nil {
		return nil, err
	}

	if agEvent.Type != "oracleServer/onQuery" {
		return nil, errNoJobMatch
	}

	var onQueryData agoricOnQueryData
	err = json.Unmarshal(agEvent.Data, &onQueryData)
	if err != nil {
		return nil, err
	}

	var query chainlinkQuery
	err = json.Unmarshal(onQueryData.Query, &query)
	if err != nil {
		return nil, err
	}

	// Check that the job ID matches.
	if query.JobID != r.jobid {
		return nil, errNoJobMatch
	}

	var requestParams map[string]interface{}
	if query.Params == nil {
		requestParams = make(map[string]interface{})
	} else {
		requestParams = query.Params
	}
	requestParams["request_id"] = onQueryData.QueryID
	requestParams["payment"] = fmt.Sprint(onQueryData.Fee) +
		strings.Repeat("0", linkDecimals-linkAgoricDecimals)

	return requestParams, nil
}
