package client

import (
	"encoding/json"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

type storeInterface interface {
	DeleteAllEndpointsExcept(names []string) error
	LoadSubscriptions() ([]store.Subscription, error)
	LoadSubscription(jobid string) (*store.Subscription, error)
	LoadEndpoint(name string) (store.Endpoint, error)
	Close() error
	SaveSubscription(arg *store.Subscription) error
	DeleteSubscription(subscription *store.Subscription) error
	SaveEndpoint(e *store.Endpoint) error
}

// startService runs the Service in the background and gracefully stops when a
// SIGINT is received.
func startService(
	config Config,
	dbClient *store.Client,
	args []string,
) {
	logger.Info("Starting External Initiator")

	// Set the mocking status before we start anything else
	blockchain.ExpectsMock = config.ExpectsMock

	clUrl, err := url.Parse(normalizeLocalhost(config.ChainlinkURL))
	if err != nil {
		logger.Fatal(err)
	}

	srv := NewService(dbClient, chainlink.Node{
		AccessKey:    config.InitiatorToChainlinkAccessKey,
		AccessSecret: config.InitiatorToChainlinkSecret,
		Endpoint:     *clUrl,
		Retry: chainlink.RetryConfig{
			Timeout:  config.ChainlinkTimeout,
			Attempts: config.ChainlinkRetryAttempts,
			Delay:    config.ChainlinkRetryDelay,
		},
	}, store.RuntimeConfig{
		KeeperBlockCooldown: config.KeeperBlockCooldown,
	})

	var names []string
	for _, e := range args {
		var endpoint store.Endpoint
		err = json.Unmarshal([]byte(e), &endpoint)
		if err != nil {
			continue
		}
		err = srv.SaveEndpoint(&endpoint)
		if err != nil {
			logger.Error(err)
		}

		names = append(names, endpoint.Name)
	}

	// Any endpoint that's not provided on startup
	// should be deleted
	if len(names) > 0 {
		err = srv.store.DeleteAllEndpointsExcept(names)
		if err != nil {
			logger.Error(err)
		}
	}

	go func() {
		err := srv.Run()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	go RunWebserver(config.ChainlinkToInitiatorAccessKey, config.ChainlinkToInitiatorSecret, srv)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	logger.Info("Shutting down...")
	srv.Close()
	os.Exit(0)
}

// Service holds the main process for running
// the external initiator.
type Service struct {
	clNode        chainlink.Node
	store         storeInterface
	subscriptions map[string]*activeSubscription
	runtimeConfig store.RuntimeConfig
}

func validateEndpoint(endpoint store.Endpoint) error {
	validBlockchain := blockchain.ValidBlockchain(endpoint.Type)
	if !validBlockchain {
		return errors.New("Missing or invalid endpoint blockchain")
	}

	if len(endpoint.Name) == 0 {
		return errors.New("Missing endpoint name")
	}

	if _, err := url.Parse(endpoint.Url); err != nil {
		return errors.New("Invalid endpoint URL")
	}

	return nil
}

// NewService returns a new instance of Service, using
// the provided database client and Chainlink node config.
func NewService(
	dbClient storeInterface,
	clNode chainlink.Node,
	runtimeConfig store.RuntimeConfig,
) *Service {
	return &Service{
		store:         dbClient,
		clNode:        clNode,
		subscriptions: make(map[string]*activeSubscription),
		runtimeConfig: runtimeConfig,
	}
}

// Run loads subscriptions, validates and subscribes to them.
func (srv *Service) Run() error {
	subs, err := srv.store.LoadSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subs {
		iSubscriber, err := srv.getAndTestSubscription(&sub)
		if err != nil {
			logger.Error(err)
			continue
		}

		err = srv.subscribe(&sub, iSubscriber)
		if err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func (srv *Service) getAndTestSubscription(sub *store.Subscription) (subscriber.ISubscriber, error) {
	endpoint, err := srv.store.LoadEndpoint(sub.EndpointName)
	if err != nil {
		return nil, errors.Wrap(err, "Failed loading endpoint")
	}
	sub.Endpoint = endpoint

	iSubscriber, err := getSubscriber(*sub)
	if err != nil {
		return nil, err
	}

	if err := iSubscriber.Test(); err != nil {
		return nil, errors.Wrap(err, "Failed testing subscriber")
	}

	return iSubscriber, nil
}

func closeSubscription(sub *activeSubscription) {
	if sub.Interface != nil {
		sub.Interface.Unsubscribe()
	}
	if sub.Events != nil {
		close(sub.Events)
	}
}

// Close shuts down any open subscriptions and closes
// the database client.
func (srv *Service) Close() {
	for _, sub := range srv.subscriptions {
		closeSubscription(sub)
	}

	err := srv.store.Close()
	if err != nil {
		logger.Error(err)
	}

	logger.Info("All connections closed. Bye!")
}

type activeSubscription struct {
	Subscription *store.Subscription
	Interface    subscriber.ISubscription
	Events       chan subscriber.Event
	Node         chainlink.Node
}

func (srv *Service) subscribe(sub *store.Subscription, iSubscriber subscriber.ISubscriber) error {
	if _, ok := srv.subscriptions[sub.Job]; ok {
		return errors.New("already subscribed to this jobid")
	}

	events := make(chan subscriber.Event)

	subscription, err := iSubscriber.SubscribeToEvents(events, srv.runtimeConfig)
	if err != nil {
		return err
	}

	as := &activeSubscription{
		Subscription: sub,
		Interface:    subscription,
		Events:       events,
		Node:         srv.clNode,
	}
	srv.subscriptions[sub.Job] = as

	go func() {
		// Add a second of delay to let services (Chainlink core)
		// sync up before sending the first job run trigger.
		time.Sleep(1 * time.Second)

		for {
			event, ok := <-as.Events
			if !ok {
				return
			}
			go func() {
				err := as.Node.TriggerJob(as.Subscription.Job, event)
				if err != nil {
					logger.Error("Failed sending job run trigger: ", err)
				}
			}()
		}
	}()

	return nil
}

// SaveSubscription tests, stores and subscribes to the store.Subscription
// provided.
func (srv *Service) SaveSubscription(arg *store.Subscription) error {
	sub, err := srv.getAndTestSubscription(arg)
	if err != nil {
		return err
	}

	if err := srv.store.SaveSubscription(arg); err != nil {
		return err
	}

	return srv.subscribe(arg, sub)
}

// DeleteJob unsubscribes (if applicable) and deletes
// the subscription associated with the jobId provided.
func (srv *Service) DeleteJob(jobid string) error {
	var sub *store.Subscription
	activeSub, ok := srv.subscriptions[jobid]
	if ok {
		closeSubscription(activeSub)
		defer delete(srv.subscriptions, jobid)
		sub = activeSub.Subscription
	} else {
		dbSub, err := srv.store.LoadSubscription(jobid)
		if err != nil {
			return err
		}
		sub = dbSub
	}

	return srv.store.DeleteSubscription(sub)
}

// GetEndpoint returns an instance of store.Endpoint that
// matches the endpoint name provided.
func (srv *Service) GetEndpoint(name string) (*store.Endpoint, error) {
	endpoint, err := srv.store.LoadEndpoint(name)
	if err != nil {
		return nil, err
	}
	if endpoint.Name != name {
		return nil, errors.New("endpoint name mismatch")
	}
	return &endpoint, nil
}

// SaveEndpoint validates and stores the store.Endpoint provided.
func (srv *Service) SaveEndpoint(e *store.Endpoint) error {
	if err := validateEndpoint(*e); err != nil {
		return err
	}
	return srv.store.SaveEndpoint(e)
}

func getSubscriber(sub store.Subscription) (subscriber.ISubscriber, error) {
	connType, err := blockchain.GetConnectionType(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	if connType == subscriber.Client {
		return blockchain.CreateClientManager(sub)
	}

	manager, err := blockchain.CreateJsonManager(connType, sub)
	if err != nil {
		return nil, err
	}

	switch connType {
	case subscriber.WS:
		return subscriber.WebsocketSubscriber{Endpoint: sub.Endpoint.Url, Manager: manager}, nil
	case subscriber.RPC:
		return subscriber.RpcSubscriber{Endpoint: sub.Endpoint.Url, Interval: time.Duration(sub.Endpoint.RefreshInt) * time.Second, Manager: manager}, nil
	}

	return nil, errors.New("unknown Endpoint type")
}

func normalizeLocalhost(endpoint string) string {
	if strings.HasPrefix(endpoint, "localhost") {
		return "http://" + endpoint
	}
	return endpoint
}
