package blockchain

import (
	"encoding/hex"
	"fmt"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"log"
	"time"
)

const (
	ONT          = "ontology"
	scanInterval = 5 * time.Second
)

func createOntSubscriber(sub store.Subscription) (*OntSubscriber, error) {
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress(sub.Endpoint.Url)
	return &OntSubscriber{
		Sdk:       sdk,
		Addresses: sub.Ontology.Addresses,
		JobId:     sub.Job,
	}, nil
}

type OntSubscriber struct {
	Sdk       *ontology_go_sdk.OntologySdk
	Addresses []string
	JobId     string
}

type OntSubscription struct {
	sdk       *ontology_go_sdk.OntologySdk
	events    chan<- subscriber.Event
	addresses map[string]bool
	jobId     string
	height    uint32
	isDone    bool
}

func (ot OntSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ ...interface{}) (subscriber.ISubscription, error) {
	logger.Error("Using Ontology RPC endpoint: Listening for events on addresses: %v\n", ot.Addresses)
	addresses := make(map[string]bool)
	for _, a := range ot.Addresses {
		addresses[a] = true
	}
	ontSubscription := OntSubscription{
		sdk:       ot.Sdk,
		events:    channel,
		addresses: addresses,
		jobId:     ot.JobId,
	}

	go ontSubscription.scanWithRetry()

	return ontSubscription, nil
}

func (ot OntSubscriber) Test() error {
	_, err := ot.Sdk.GetCurrentBlockHeight()
	if err != nil {
		return err
	}
	return nil
}

func (ots OntSubscription) scanWithRetry() {
	for {
		ots.scan()
		if !ots.isDone {
			time.Sleep(scanInterval)
			continue
		}
		return
	}
}

func (ots OntSubscription) scan() {
	currentHeight, err := ots.sdk.GetCurrentBlockHeight()
	if err != nil {
		logger.Error("ont scan, get current block height error:", err)
	}
	if ots.height == 0 {
		ots.height = currentHeight
	}
	for h := ots.height; h < currentHeight+1; h++ {
		err := ots.parseOntEvent(h)
		if err != nil {
			logger.Error("ont scan, parse ont event error:", err)
		}
	}
	ots.height = currentHeight + 1
}

func (ots OntSubscription) parseOntEvent(height uint32) error {
	ontEvents, err := ots.sdk.GetSmartContractEventByBlock(height)
	log.Printf("parseOntEvent, start to parse ont block %d", height)
	if err != nil {
		return fmt.Errorf("parseOntEvent, get smartcontract event by block error:%s", err)
	}

	for _, e := range ontEvents {
		for _, notify := range e.Notify {
			states, ok := notify.States.([]interface{})
			if !ok {
				continue
			}
			_, ok = ots.addresses[notify.ContractAddress]
			if !ok {
				continue
			}
			name := fmt.Sprint(states[0])
			if name == hex.EncodeToString([]byte("oracleRequest")) {
				jobId := states[1].(string)
				if jobId != ots.jobId {
					continue
				}
				logger.Error("parseOntEvent, found tracked job: %s", jobId)

				requestID := fmt.Sprint(states[3])
				p := fmt.Sprint(states[4])
				callbackAddress := fmt.Sprint(states[5])
				function := fmt.Sprint(states[6])
				expiration := fmt.Sprint(states[7])
				data := fmt.Sprint(states[9])
				dataBytes, err := hex.DecodeString(data)
				if err != nil {
					logger.Error("parseOntEvent, date from hex to bytes error:", err)
				}
				js, err := models.ParseCBOR(dataBytes)
				if err != nil {
					logger.Error("parseOntEvent, date from bytes to JSON error:", err)
				}
				js, err = js.Add("address", notify.ContractAddress)
				if err != nil {
					logger.Error("parseOntEvent, date JSON add address error:", err)
				}
				js, err = js.Add("requestID", requestID)
				if err != nil {
					logger.Error("parseOntEvent, date JSON add requestID error:", err)
				}
				js, err = js.Add("payment", p)
				if err != nil {
					logger.Error("parseOntEvent, date JSON add payment error:", err)
				}
				js, err = js.Add("callbackAddress", callbackAddress)
				if err != nil {
					logger.Error("parseOntEvent, date JSON add callbackAddress error:", err)
				}
				js, err = js.Add("callbackFunction", function)
				if err != nil {
					logger.Error("parseOntEvent, date JSON add callbackFunction error:", err)
				}
				js, err = js.Add("expiration", expiration)
				if err != nil {
					logger.Error("parseOntEvent, date JSON add expiration error:", err)
				}
				event, _ := js.MarshalJSON()
				ots.events <- event
			}
		}
	}
	return nil
}

func (ots OntSubscription) Unsubscribe() {
	logger.Error("Unsubscribing from Ontology endpoint")
	ots.isDone = true
}
