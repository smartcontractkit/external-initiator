package blockchain

import (
	"encoding/hex"
	"fmt"
	"time"

	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/common"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const (
	ONT          = "ontology"
	scanInterval = 5 * time.Second
)

func createOntSubscriber(sub store.Subscription) *ontSubscriber {
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress(sub.Endpoint.Url)
	return &ontSubscriber{
		Sdk:       sdk,
		Addresses: sub.Ontology.Addresses,
		JobId:     sub.Job,
	}
}

type ontSubscriber struct {
	Sdk       *ontology_go_sdk.OntologySdk
	Addresses []string
	JobId     string
}

type ontSubscription struct {
	sdk       *ontology_go_sdk.OntologySdk
	events    chan<- subscriber.Event
	addresses map[string]bool
	jobId     string
	height    uint32
	isDone    bool
}

func (ot *ontSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ store.RuntimeConfig) (subscriber.ISubscription, error) {
	logger.Infof("Using Ontology RPC endpoint: Listening for events on addresses: %v\n", ot.Addresses)
	addresses := make(map[string]bool)
	for _, a := range ot.Addresses {
		addresses[a] = true
	}
	ontSubscription := &ontSubscription{
		sdk:       ot.Sdk,
		events:    channel,
		addresses: addresses,
		jobId:     ot.JobId,
	}

	go ontSubscription.scanWithRetry()

	return ontSubscription, nil
}

func (ot *ontSubscriber) Test() error {
	_, err := ot.Sdk.GetCurrentBlockHeight()
	if err != nil {
		return err
	}
	return nil
}

func (ots *ontSubscription) scanWithRetry() {
	for {
		ots.scan()
		if !ots.isDone {
			time.Sleep(scanInterval)
			continue
		}
		return
	}
}

func (ots *ontSubscription) scan() {
	currentHeight, err := ots.sdk.GetCurrentBlockHeight()
	if err != nil {
		logger.Error("ont scan, get current block height error:", err)
		return
	}
	if ots.height == 0 {
		ots.height = currentHeight
	}
	for h := ots.height; h < currentHeight+1; h++ {
		err := ots.parseOntEvent(h)
		if err != nil {
			logger.Error("ont scan, parse ont event error:", err)
			return
		}
	}
	ots.height = currentHeight + 1
}

func (ots *ontSubscription) parseOntEvent(height uint32) error {
	ontEvents, err := ots.sdk.GetSmartContractEventByBlock(height)
	logger.Debugf("parseOntEvent, start to parse ont block %d", height)
	if err != nil {
		return fmt.Errorf("parseOntEvent, get smartcontract event by block error:%s", err)
	}

	for _, e := range ontEvents {
		for _, notify := range e.Notify {
			event, ok := ots.notifyTrigger(notify)
			if ok {
				ots.events <- event
			}
		}
	}
	return nil
}

func (ots *ontSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from Ontology endpoint")
	ots.isDone = true
}

func (ots *ontSubscription) notifyTrigger(notify *common.NotifyEventInfo) ([]byte, bool) {
	states, ok := notify.States.([]interface{})
	if !ok {
		return nil, false
	}
	_, ok = ots.addresses[notify.ContractAddress]
	if !ok {
		return nil, false
	}
	if len(states) < 11 {
		return nil, false
	}
	name := fmt.Sprint(states[0])
	if name == hex.EncodeToString([]byte("oracleRequest")) {
		jobId := fmt.Sprint(states[1])
		// Check if our jobID matches
		if !matchesJobID(ots.jobId, jobId) {
			return nil, false
		}
		logger.Debugf("parseOntEvent, found tracked job: %s", jobId)

		requestID := fmt.Sprint(states[3])
		p := fmt.Sprint(states[4])
		callbackAddress := fmt.Sprint(states[5])
		function := fmt.Sprint(states[6])
		expiration := fmt.Sprint(states[7])
		data := fmt.Sprint(states[9])
		dataBytes, err := hex.DecodeString(data)
		if err != nil {
			logger.Error("parseOntEvent, date from hex to bytes error:", err)
			return nil, false
		}
		js, err := models.ParseCBOR(dataBytes)
		if err != nil {
			logger.Error("parseOntEvent, date from bytes to JSON error:", err)
			return nil, false
		}
		js, err = js.Add("address", notify.ContractAddress)
		if err != nil {
			logger.Error("parseOntEvent, date JSON add address error:", err)
			return nil, false
		}
		js, err = js.Add("requestID", requestID)
		if err != nil {
			logger.Error("parseOntEvent, date JSON add requestID error:", err)
			return nil, false
		}
		js, err = js.Add("payment", p)
		if err != nil {
			logger.Error("parseOntEvent, date JSON add payment error:", err)
			return nil, false
		}
		js, err = js.Add("callbackAddress", callbackAddress)
		if err != nil {
			logger.Error("parseOntEvent, date JSON add callbackAddress error:", err)
			return nil, false
		}
		js, err = js.Add("callbackFunction", function)
		if err != nil {
			logger.Error("parseOntEvent, date JSON add callbackFunction error:", err)
			return nil, false
		}
		js, err = js.Add("expiration", expiration)
		if err != nil {
			logger.Error("parseOntEvent, date JSON add expiration error:", err)
			return nil, false
		}
		event, _ := js.MarshalJSON()
		return event, true
	}
	return nil, false
}
