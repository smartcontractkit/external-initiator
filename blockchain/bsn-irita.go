package blockchain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmrpc "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const (
	BIRITA = "bsn-irita"

	DefaultScannerInterval = 5
	ClientTimeout          = 10

	ServiceRequestEventType = "new_batch_request_provider"
)

type biritaSubscriber struct {
	Client      *tmrpc.HTTP
	Interval    time.Duration
	JobID       string
	Addresses   []string
	ServiceName string
}

type biritaSubscription struct {
	client      *tmrpc.HTTP
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

type BIritaServiceRequest struct {
	ID          string `json:"id"`
	ServiceName string `json:"service_name"`
	Provider    string `json:"provider"`
	Input       string `json:"input"`
}

func createBSNIritaSubscriber(sub store.Subscription) (*biritaSubscriber, error) {
	client, err := tmrpc.NewWithTimeout(sub.Endpoint.Url, "", ClientTimeout)
	if err != nil {
		return nil, err
	}

	interval := sub.Endpoint.RefreshInt
	if interval <= 0 {
		interval = DefaultScannerInterval
	}

	return &biritaSubscriber{
		Client:      client,
		Interval:    time.Duration(interval) * time.Second,
		JobID:       sub.Job,
		Addresses:   sub.BSNIrita.Addresses,
		ServiceName: sub.BSNIrita.ServiceName,
	}, nil
}

func (bs *biritaSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ store.RuntimeConfig) (subscriber.ISubscription, error) {
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
	_, err := bs.Client.Status(context.Background())
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
	res, err := bs.client.Status(context.Background())
	if err != nil {
		return -1, err
	}

	return res.SyncInfo.LatestBlockHeight, nil
}

func (bs *biritaSubscription) scanByRange(startHeight int64, endHeight int64) {
	for h := startHeight; h <= endHeight; {
		blockResult, err := bs.client.BlockResults(context.Background(), &h)
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

func (bs *biritaSubscription) queryServiceRequest(requestID string) (request BIritaServiceRequest, err error) {
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

	res, err := bs.client.ABCIQuery(context.Background(), "/custom/service/request", bz)
	if err != nil {
		return request, err
	}

	err = tmjson.Unmarshal(res.Response.Value, &request)
	if err != nil {
		return request, err
	}

	return
}

func (bs *biritaSubscription) onServiceRequest(request BIritaServiceRequest) {
	logger.Infof("BSN-IRITA: service request received: %s", request.ID)

	event, err := bs.buildTriggerEvent(request)
	if err != nil {
		logger.Errorf("BSN-IRITA: failed to build the event to trigger job run: %s", err)
		return
	}

	bs.events <- event
}

func (bs *biritaSubscription) buildTriggerEvent(request BIritaServiceRequest) (subscriber.Event, error) {
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
	triggerEvent.RequestID = request.ID
	triggerEvent.RequestBody = requestBody

	event, err := json.Marshal(triggerEvent)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (bs *biritaSubscription) checkServiceRequest(request BIritaServiceRequest) error {
	if len(request.ID) == 0 {
		return fmt.Errorf("missing request id")
	}

	if len(request.Input) == 0 {
		return fmt.Errorf("missing request input")
	}

	if len(bs.serviceName) > 0 && request.ServiceName != bs.serviceName {
		return fmt.Errorf("service name does not match")
	}

	if _, ok := bs.addresses[request.Provider]; !ok {
		return fmt.Errorf("provider address does not match")
	}

	return nil
}

func (bs *biritaSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from BSN-IRITA endpoint")
	bs.done = true
}

func getAttributeValue(event abci.Event, attributeKey string) (string, error) {
	for _, attr := range event.Attributes {
		if string(attr.Key) == attributeKey {
			return string(attr.Value), nil
		}
	}

	return "", fmt.Errorf("attribute key %s does not exist", attributeKey)
}

type dummySM2 []byte

func (dummySM2) Address() crypto.Address {
	return nil
}

func (dummySM2) Bytes() []byte {
	return nil
}

func (dummySM2) VerifySignature(msg []byte, sig []byte) bool {
	return true
}

func (dummySM2) Equals(other crypto.PubKey) bool {
	return false
}

func (dummySM2) Type() string {
	return ""
}

func init() {
	tmjson.RegisterType(dummySM2{}, "tendermint/PubKeySm2")
}
