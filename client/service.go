package client

import (
	"encoding/json"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/services"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/tidwall/gjson"
)

var (
	promActiveSubscriptions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ei_subscriptions_active",
		Help: "The total number of active subscriptions",
	}, []string{"endpoint"})
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
	SaveJobSpec(arg *store.JobSpec) error
	LoadJobSpec(jobid string) (*store.JobSpec, error)
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
	common.ExpectsMock = config.ExpectsMock

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
		KeeperBlockCooldown:    config.KeeperBlockCooldown,
		FMAdapterTimeout:       config.FMAdapterTimeout,
		FMAdapterRetryAttempts: config.FMAdapterRetryAttempts,
		FMAdapterRetryDelay:    config.FMAdapterRetryDelay,
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

	go RunWebserver(config.ChainlinkToInitiatorAccessKey, config.ChainlinkToInitiatorSecret, srv, config.Port)

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
		err = srv.subscribe(sub)
		if err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func closeSubscription(sub *activeSubscription) {
	sub.onStop.Do(func() {
		sub.Service.Stop()
	})
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
	Service      services.Service

	onStop sync.Once
}

func (srv Service) jobTriggerListener(job string, ch <-chan subscriber.Event) {
	// Add a second of delay to let services (Chainlink core)
	// sync up before sending the first job run trigger.
	time.Sleep(1 * time.Second)

	for {
		trigger := <-ch
		go func() {
			err := srv.clNode.TriggerJob(job, trigger)
			if err != nil {
				logger.Errorw("Failed sending job run trigger: ", err, "job", job)
			}
		}()
	}
}

func (srv *Service) subscribe(sub store.Subscription) error {
	if _, ok := srv.subscriptions[sub.Job]; ok {
		return errors.New("already subscribed to this jobid")
	}

	var service services.Service
	var err error
	triggerJobRun := make(chan subscriber.Event, 100)
	js, err := srv.store.LoadJobSpec(sub.Job)
	if err != nil || gjson.GetBytes(js.Spec, "fluxmonitor").Raw == "null" {
		service, err = srv.subscribeRunlog(sub, triggerJobRun, js)
	} else {
		service, err = srv.subscribeFluxmonitor(sub, triggerJobRun, js)
	}
	if err != nil {
		return err
	}

	srv.subscriptions[sub.Job] = &activeSubscription{
		Subscription: &sub,
		Service:      service,
	}

	go srv.jobTriggerListener(sub.Job, triggerJobRun)

	return nil
}

func (srv Service) subscribeRunlog(sub store.Subscription, ch chan subscriber.Event, _ *store.JobSpec) (services.Service, error) {
	blockchainManager, err := blockchain.CreateRunlogManager(sub)
	if err != nil {
		return nil, err
	}

	return services.NewRunlog(sub.Job, ch, blockchainManager)
}

func (srv *Service) subscribeFluxmonitor(sub store.Subscription, ch chan subscriber.Event, js *store.JobSpec) (services.Service, error) {
	logger.Info("Starting FluxMonitor for Job: ", js.Job)

	fmConfig, err := services.ParseFMSpec(js.Spec, srv.runtimeConfig)
	if err != nil {
		return nil, err
	}

	blockchainManager, err := blockchain.CreateFluxMonitorManager(sub)
	if err != nil {
		return nil, err
	}

	return services.NewFluxMonitor(js.Job, fmConfig, ch, blockchainManager)
}

func (srv *Service) SaveJobSpec(arg *store.JobSpec) error {
	return srv.store.SaveJobSpec(arg)
}

func (srv *Service) LoadJobSpec(jobid string) (*store.JobSpec, error) {
	jobspec, err := srv.store.LoadJobSpec(jobid)
	if err != nil {
		return nil, err
	}
	return jobspec, nil
}

// SaveSubscription tests, stores and subscribes to the store.Subscription
// provided.
func (srv *Service) SaveSubscription(arg *store.Subscription) error {
	if err := srv.store.SaveSubscription(arg); err != nil {
		return errors.Wrap(err, "unable to store subscription")
	}

	return srv.subscribe(*arg)
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

func normalizeLocalhost(endpoint string) string {
	if strings.HasPrefix(endpoint, "localhost") {
		return "http://" + endpoint
	}
	return endpoint
}
