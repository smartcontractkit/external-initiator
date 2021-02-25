package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

// Agoric is the identifier of this
// blockchain integration.
const Agoric = "agoric"

// linkDecimals is the number of decimal places in $LINK
const linkDecimals = 18

// linkAgoricDecimals is the number of decimal places in a uaglink token
// FIXME: Ideally the same as linkDecimals.
const linkAgoricDecimals = 6

type agoricFilter struct {
	JobID string
}

type agoricManager struct {
	endpointName string
	filter       agoricFilter
}

func init() {
	if linkAgoricDecimals > linkDecimals {
		panic(fmt.Errorf("linkAgoricDecimals %d must be less than or equal to linkDecimals %d", linkAgoricDecimals, linkDecimals))
	}
}

func createAgoricManager(t subscriber.Type, conf store.Subscription) (*agoricManager, error) {
	if t != subscriber.WS {
		return nil, errors.New("only WS connections are allowed for Agoric")
	}

	return &agoricManager{
		endpointName: conf.EndpointName,
		filter: agoricFilter{
			JobID: conf.Job,
		},
	}, nil
}

func (sm agoricManager) GetTriggerJson() []byte {
	return nil
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
	JobID  string                 `json:"jobId"`
	Params map[string]interface{} `json:"params"`
}

func (sm *agoricManager) ParseResponse(data []byte) ([]subscriber.Event, bool) {
	promLastSourcePing.With(prometheus.Labels{"endpoint": sm.endpointName, "jobid": string(sm.filter.JobID)}).SetToCurrentTime()

	var agEvent agoricEvent
	err := json.Unmarshal(data, &agEvent)
	if err != nil {
		logger.Error("Failed parsing agoricEvent:", err)
		return nil, false
	}

	var subEvents []subscriber.Event

	switch agEvent.Type {
	case "oracleServer/onQuery":
		// Do this below.
		break
	case "oracleServer/onError":
	case "oracleServer/onReply":
		return nil, false
	default:
		// We don't need something so noisy.
		// logger.Error("Unimplemented message type:", agEvent.Type)
		return nil, false
	}

	var onQueryData agoricOnQueryData
	err = json.Unmarshal(agEvent.Data, &onQueryData)
	if err != nil {
		logger.Error("Failed parsing queryData:", err)
		return nil, false
	}

	var query chainlinkQuery
	err = json.Unmarshal(onQueryData.Query, &query)
	if err != nil {
		logger.Error("Failed parsing chainlink query:", err)
		return nil, false
	}

	// Check that the job ID matches.
	if query.JobID != sm.filter.JobID {
		return subEvents, true
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

	event, err := json.Marshal(requestParams)
	if err != nil {
		logger.Error(err)
		return nil, false
	}
	subEvents = append(subEvents, event)

	return subEvents, true
}

func (sm *agoricManager) GetTestJson() []byte {
	return nil
}

func (sm *agoricManager) ParseTestResponse(data []byte) error {
	return nil
}
