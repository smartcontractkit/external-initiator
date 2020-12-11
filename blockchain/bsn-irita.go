package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	servicesdk "github.com/irisnet/service-sdk-go"
	"github.com/irisnet/service-sdk-go/service"
	"github.com/irisnet/service-sdk-go/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const (
	BIRITA                 = "bsn-irita"
	DefaultScannerInterval = 5 * time.Second

	ServiceRequestEventType = "new_batch_request_provider"
)

type biritaSubscriber struct {
	Client      servicesdk.ServiceClient
	Interval    time.Duration
	JobID       string
	Addresses   []string
	ServiceName string
}

type biritaSubscription struct {
	client      servicesdk.ServiceClient
	events      chan<- subscriber.Event
	interval    time.Duration
	jobID       string
	addresses   map[string]bool
	serviceName string
	lastHeight  int64
	done        bool
}

type biritaTriggerEvent struct {
	RequestID   string      `json:"request_id"`
	RequestBody models.JSON `json:"request_body"`
}

type biritaQueryServiceRequestParams struct {
	RequestID []byte
}

func createBSNIritaSubscriber(sub store.Subscription) *biritaSubscriber {
	cfg := types.ClientConfig{
		NodeURI: sub.Endpoint.Url,
	}
	serviceClient := servicesdk.NewServiceClient(cfg)

	interval := time.Duration(sub.Endpoint.RefreshInt) * time.Second
	if interval <= 0 {
		interval = DefaultScannerInterval
	}

	return &biritaSubscriber{
		Client:      serviceClient,
		Interval:    interval,
		JobID:       sub.Job,
		Addresses:   sub.BSNIrita.Addresses,
		ServiceName: sub.BSNIrita.ServiceName,
	}
}

func (bs *biritaSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ ...interface{}) (subscriber.ISubscription, error) {
	logger.Infof("Subscribe to BSN-IRITA service requests, provider addresses: %v, service name: %s\n", bs.Addresses, bs.ServiceName)

	addressMap := make(map[string]bool)
	for _, addr := range bs.Addresses {
		addressMap[addr] = true
	}

	biritaSubscription := &biritaSubscription{
		client:      bs.Client,
		events:      channel,
		interval:    bs.Interval,
		jobID:       bs.JobID,
		addresses:   addressMap,
		serviceName: bs.ServiceName,
	}

	go biritaSubscription.start()

	return biritaSubscription, nil
}

func (bs *biritaSubscriber) Test() error {
	_, err := bs.Client.Status()
	if err != nil {
		return err
	}

	return nil
}

func (bs *biritaSubscription) start() {
	for {
		bs.scan()

		if bs.done {
			return
		}

		time.Sleep(bs.interval)
	}
}

func (bs *biritaSubscription) scan() {
	currentHeight, err := bs.getLatestHeight()
	if err != nil {
		logger.Errorf("BSN-IRITA: failed to retrieve the latest block height: %s", err)
		return
	}

	if bs.lastHeight == 0 {
		bs.lastHeight = currentHeight - 1
	}

	if currentHeight <= bs.lastHeight {
		return
	}

	bs.scanByRange(bs.lastHeight+1, currentHeight)
}

func (bs biritaSubscription) getLatestHeight() (int64, error) {
	res, err := bs.client.Status()
	if err != nil {
		return -1, err
	}

	return res.SyncInfo.LatestBlockHeight, nil
}

func (bs *biritaSubscription) scanByRange(startHeight int64, endHeight int64) {
	for h := startHeight; h <= endHeight; {
		blockResult, err := bs.client.BlockResults(&h)
		if err != nil {
			logger.Errorf("BSN-IRITA: failed to retrieve the block result, height: %d, err: %s", h, err)
			continue
		}

		bs.parseServiceRequests(blockResult.EndBlockEvents)

		bs.lastHeight = h
		h++
	}
}

func (bs *biritaSubscription) parseServiceRequests(events []abci.Event) {
	for _, e := range events {
		if bs.validServiceRequestEvent(e) {
			requestIDArr, err := getAttributeValue(e, "requests")
			if err != nil {
				logger.Errorf("BSN-IRITA: failed to parse service request ids, event: %s, err: %s", e.String(), err)
				return
			}

			var requestIDs []string
			err = json.Unmarshal([]byte(requestIDArr), &requestIDs)
			if err != nil {
				logger.Errorf("BSN-IRITA: failed to unmarshal service request ids: %s", err)
				return
			}

			for _, id := range requestIDs {
				request, err := bs.queryServiceRequest(id)
				if err != nil {
					logger.Errorf("BSN-IRITA: failed to query the service request %s: %s", id, err)
					continue
				}

				bs.onServiceRequest(request)
			}
		}
	}
}

func (bs *biritaSubscription) validServiceRequestEvent(event abci.Event) bool {
	if event.Type != ServiceRequestEventType {
		return false
	}

	if len(bs.serviceName) > 0 {
		serviceName, err := getAttributeValue(event, "service_name")
		if err != nil || serviceName != bs.serviceName {
			return false
		}
	}

	providerAddr, err := getAttributeValue(event, "provider")
	if err != nil {
		return false
	}

	if _, ok := bs.addresses[providerAddr]; !ok {
		return false
	}

	return true
}

func (bs *biritaSubscription) queryServiceRequest(requestID string) (request service.Request, err error) {
	requestIDBz, err := hex.DecodeString(requestID)
	if err != nil {
		return request, err
	}

	params := biritaQueryServiceRequestParams{
		RequestID: requestIDBz,
	}

	bz, err := json.Marshal(params)
	if err != nil {
		return request, err
	}

	res, err := bs.client.ABCIQuery("/custom/service/request", bz)
	if err != nil {
		return request, err
	}

	err = tmjson.Unmarshal(res.Response.Value, &request)
	if err != nil {
		return request, err
	}

	return
}

func (bs *biritaSubscription) onServiceRequest(request service.Request) {
	logger.Infof("BSN-IRITA: service request received: %s", request.Id.String())

	event, err := bs.buildTriggerEvent(request)
	if err != nil {
		logger.Errorf("BSN-IRITA: failed to build the event to trigger job run: %s", err)
		return
	}

	bs.events <- event
}

func (bs *biritaSubscription) buildTriggerEvent(request service.Request) (subscriber.Event, error) {
	err := bs.checkServiceRequest(request)
	if err != nil {
		return nil, err
	}

	requestBodyStr := gjson.Get(request.Input, "body").String()

	requestBody, err := models.ParseJSON([]byte(requestBodyStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse request body %s to JSON: %s", requestBodyStr, err)
	}

	var triggerEvent biritaTriggerEvent
	triggerEvent.RequestID = request.Id.String()
	triggerEvent.RequestBody = requestBody

	event, err := json.Marshal(triggerEvent)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (bs *biritaSubscription) checkServiceRequest(request service.Request) error {
	if len(request.Id) == 0 {
		return fmt.Errorf("missing request id")
	}

	if len(request.Input) == 0 {
		return fmt.Errorf("missing request input")
	}

	if len(bs.serviceName) > 0 && request.ServiceName != bs.serviceName {
		return fmt.Errorf("service name does not match")
	}

	if _, ok := bs.addresses[request.Provider.String()]; !ok {
		return fmt.Errorf("provider address does not match")
	}

	return nil
}

func (bs *biritaSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from BSN-IRITA endpoint")
	bs.done = true
}

func getAttributeValue(event abci.Event, attributeKey string) (string, error) {
	stringEvents := types.StringifyEvents([]abci.Event{event})
	return stringEvents.GetValue(event.Type, attributeKey)
}
