package blockchain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const (
	Algorand         = "algorand"
	AlgorandInterval = 5
)

var AlgorandLastAmount uint64 = 0

type algorandSubscriber struct {
	Endpoint     string
	EndpointName string
	Daemon       algod.Client
	Address      string
	Interval     time.Duration
}

func createAlgoSubscriber(sub store.Subscription) (*algorandSubscriber, error) {
	commonClient, err := common.MakeClient(sub.Endpoint.Url, "X-API-Key", sub.Algorand.AlgoAPIToken)
	if err != nil {
		logger.Error("failed to make common client: %s\n", err)
		return nil, err
	}

	algodClient := (*algod.Client)(commonClient)

	interval := sub.Endpoint.RefreshInt
	if interval <= 0 {
		interval = AlgorandInterval
	}

	return &algorandSubscriber{
		Endpoint:     sub.Endpoint.Url,
		EndpointName: Algorand,
		Daemon:       *algodClient,
		Address:      sub.Algorand.Address,
		Interval:     time.Duration(interval) * time.Second,
	}, nil
}

type algorandSubscription struct {
	endpoint     string
	endpointName string
	daemon       algod.Client
	events       chan<- subscriber.Event
	address      string
	isDone       bool
}

func (a algorandSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ store.RuntimeConfig) (subscriber.ISubscription, error) {
	logger.Infof("Using Algorand endpoint : %s", a.Endpoint)
	logger.Infof("Checking Address        : %v", a.Address)

	sub := algorandSubscription{
		endpoint:     a.Endpoint,
		endpointName: a.EndpointName,
		daemon:       a.Daemon,
		events:       channel,
		address:      a.Address,
		isDone:       false,
	}

	go sub.queryUntilDone(a.Interval)

	return &sub, nil
}

func (a algorandSubscription) queryUntilDone(interval time.Duration) {
	for {
		if a.isDone {
			return
		}
		a.query()
		time.Sleep(interval)
	}
}

func (a *algorandSubscription) query() {
	accountInfo, err := a.daemon.AccountInformation(a.address).Do(context.Background())
	if err != nil {
		logger.Error("Error getting account info: %s", err)
		return
	}
	logger.Debugf("%s balance: %d", a.address, accountInfo.Amount)

	if accountInfo.Amount != AlgorandLastAmount {
		AlgorandLastAmount = accountInfo.Amount
		events, err := a.parseResponse(accountInfo)
		if err != nil {
			logger.Error("failed parseResponse:", err)
			return
		}

		for _, event := range events {
			a.events <- event
		}
	}
}

func (a algorandSubscription) parseResponse(info models.Account) ([]subscriber.Event, error) {
	event := map[string]interface{}{
		"address": info.Address,
		"amount":  info.Amount,
		"round":   info.Round,
		"status":  info.Status,
	}

	eventBz, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return []subscriber.Event{eventBz}, nil
}

func (a *algorandSubscription) Unsubscribe() {
	logger.Info("Stopping Algorand subscription on endpoint", a.endpoint)
	a.isDone = true
}

func (a *algorandSubscriber) Test() error {
	_, err := a.Daemon.Status().Do(context.Background())
	return err
}
