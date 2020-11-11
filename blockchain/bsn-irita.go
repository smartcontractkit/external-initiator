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

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const (
	BIRITA          = "bsn-irita"
	ScannerInterval = 5 * time.Second
)

type biritaSubscriber struct {
	Client       servicesdk.ServiceClient
	ServiceName  string
	ProviderAddr string
}

type biritaSubscription struct {
	client       servicesdk.ServiceClient
	events       chan<- subscriber.Event
	interval     time.Duration
	serviceName  string
	providerAddr string
	lastHeight   int64
	done         bool
}

type biritaTriggerEvent struct {
	RequestID   string `json:"request_id"`
	RequestBody string `json:"request_body"`
}

func createBSNIritaSubscriber(sub store.Subscription) *biritaSubscriber {
	cfg := types.ClientConfig{
		NodeURI: sub.Endpoint.Url,
	}
	serviceClient := servicesdk.NewServiceClient(cfg)

	return &biritaSubscriber{
		Client:       serviceClient,
		ServiceName:  sub.BSNIrita.ServiceName,
		ProviderAddr: sub.BSNIrita.ProviderAddr,
	}
}

func (bs *biritaSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ ...interface{}) (subscriber.ISubscription, error) {
	logger.Infof("Subscribe to BSN-IRITA service requests, service name: %s, provider address: %s\n", bs.ServiceName, bs.ProviderAddr)

	biritaSubscription := &biritaSubscription{
		client:       bs.Client,
		events:       channel,
		interval:     ScannerInterval,
		serviceName:  bs.ServiceName,
		providerAddr: bs.ProviderAddr,
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

		if !bs.done {
			time.Sleep(bs.interval)
			continue
		}

		return
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
	for h := startHeight; h <= endHeight; h++ {
		blockResult, err := bs.client.BlockResults(&h)
		if err != nil {
			logger.Errorf("BSN-IRITA: failed to retrieve the block result, height: %d, err: %s", h, err)
			continue
		}

		if len(blockResult.EndBlockEvents) > 0 {
			bs.parseServiceRequests(types.StringifyEvents(blockResult.EndBlockEvents))
		}
	}

	bs.lastHeight = endHeight
}

func (bs *biritaSubscription) parseServiceRequests(events types.StringEvents) {
	requestIDArr, err := events.GetValue("new_batch_request_provider", "requests")
	if err != nil {
		logger.Errorf("BSN-IRITA: failed to parse service requests: %s", err)
		return
	}

	var requestIDs []string
	err = json.Unmarshal([]byte(requestIDArr), &requestIDs)
	if err != nil {
		logger.Errorf("BSN-IRITA: failed to parse service requests: %s", err)
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

func (bs *biritaSubscription) queryServiceRequest(requestID string) (request service.Request, err error) {
	requestIDBz, err := hex.DecodeString(requestID)
	if err != nil {
		return request, err
	}

	data := service.QueryRequestRequest{
		RequestId: requestIDBz,
	}

	err = bs.client.QueryWithResponse("/custom/service/request", &data, &request)
	return
}

func (bs *biritaSubscription) onServiceRequest(request service.Request) {
	event, err := bs.buildTriggerEvent(request)
	if err != nil {
		logger.Errorf("BSN-IRITA: failed to build the event to trigger job run: %s", err)
		return
	}

	bs.events <- event
}

func (bs *biritaSubscription) buildTriggerEvent(request service.Request) (subscriber.Event, error) {
	if len(request.Id) == 0 {
		return nil, fmt.Errorf("missing request id")
	}

	if len(request.Input) == 0 {
		return nil, fmt.Errorf("missing request input")
	}

	if request.ServiceName != bs.serviceName {
		return nil, fmt.Errorf("service name does not match")
	}

	if request.Provider.String() != bs.providerAddr {
		return nil, fmt.Errorf("provider address does not match")
	}

	var triggerEvent biritaTriggerEvent

	triggerEvent.RequestID = request.Id.String()
	triggerEvent.RequestBody = gjson.Get(request.Input, "body").String()

	event, err := json.Marshal(triggerEvent)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (bs *biritaSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from BSN-IRITA endpoint")
	bs.done = true
}
