package client

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

// startService runs the service in the background and gracefully stops when a
// SIGINT is received.
func startService(
	config Config,
	dbClient *store.Client,
	args []string,
) {
	clUrl, err := url.Parse(normalizeLocalhost(config.Chainlink))
	if err != nil {
		log.Fatal(err)
	}

	srv := newService(dbClient, chainlink.Node{
		AccessKey:    config.ChainlinkAccessKey,
		AccessSecret: config.ChainlinkSecret,
		Endpoint:     *clUrl,
	})

	for _, e := range args {
		var endpoint store.Endpoint
		err := json.Unmarshal([]byte(e), &endpoint)
		if err != nil {
			continue
		}
		err = srv.SaveEndpoint(&endpoint)
		if err != nil {
			fmt.Println(err)
		}
	}

	go func() {
		err := srv.run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go runWebserver(srv)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Println("Shutting down...")
	srv.close()
	os.Exit(0)
}

type service struct {
	clNode        chainlink.Node
	store         *store.Client
	subscriptions map[string]*ActiveSubscription
}

func validateEndpoint(endpoint store.Endpoint) error {
	switch endpoint.Type {
	case blockchain.ETH:
		// Do nothing, valid blockchain
	default:
		return errors.New("Missing or invalid endpoint blockchain")
	}

	if len(endpoint.Name) == 0 {
		return errors.New("Missing endpoint name")
	}

	return nil
}

func newService(
	dbClient *store.Client,
	clNode chainlink.Node,
) *service {
	return &service{
		store:         dbClient,
		clNode:        clNode,
		subscriptions: make(map[string]*ActiveSubscription),
	}
}

func (srv *service) run() error {
	subs, err := srv.store.LoadSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subs {
		iSubscriber, err := srv.getAndTestSubscription(&sub)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = srv.subscribe(&sub, iSubscriber)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (srv *service) getAndTestSubscription(sub *store.Subscription) (subscriber.ISubscriber, error) {
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

func (srv *service) close() {
	for _, sub := range srv.subscriptions {
		sub.Interface.Unsubscribe()
		close(sub.Events)
	}
	err := srv.store.Close()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("All connections closed. Bye!")
}

type ActiveSubscription struct {
	Subscription *store.Subscription
	Interface    subscriber.ISubscription
	Events       chan subscriber.Event
	Node         chainlink.Node
}

func (srv *service) subscribe(sub *store.Subscription, iSubscriber subscriber.ISubscriber) error {
	if _, ok := srv.subscriptions[sub.Job]; ok {
		return errors.New("already subscribed to this jobid")
	}

	events := make(chan subscriber.Event)

	subscription, err := iSubscriber.SubscribeToEvents(events)
	if err != nil {
		log.Fatal(err)
	}

	as := &ActiveSubscription{
		Subscription: sub,
		Interface:    subscription,
		Events:       events,
		Node:         srv.clNode,
	}
	srv.subscriptions[sub.Job] = as

	go func() {
		for {
			_, ok := <-as.Events
			if !ok {
				return
			}
			if err := as.Node.TriggerJob(as.Subscription.Job); err != nil {
				fmt.Println(err)
			}
		}
	}()

	return nil
}

func (srv *service) SaveSubscription(arg *store.Subscription) error {
	sub, err := srv.getAndTestSubscription(arg)
	if err != nil {
		return err
	}

	if err := srv.store.SaveSubscription(arg); err != nil {
		return err
	}

	return srv.subscribe(arg, sub)
}

func (srv *service) DeleteJob(jobid string) error {
	sub, ok := srv.subscriptions[jobid]
	if !ok {
		return errors.New("subscription not found")
	}

	sub.Interface.Unsubscribe()
	close(sub.Events)
	err := srv.store.DeleteSubscription(sub.Subscription)
	delete(srv.subscriptions, jobid)
	return err
}

func (srv *service) GetEndpoint(name string) (*store.Endpoint, error) {
	endpoint, err := srv.store.LoadEndpoint(name)
	if err != nil {
		return nil, err
	}
	if endpoint.Name != name {
		return nil, nil
	}
	return &endpoint, nil
}

func (srv *service) SaveEndpoint(e *store.Endpoint) error {
	if err := validateEndpoint(*e); err != nil {
		return err
	}
	return srv.store.SaveEndpoint(e)
}

func getConnectionType(rawUrl string) (subscriber.Type, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return 0, err
	}

	if strings.HasPrefix(u.Scheme, "ws") {
		return subscriber.WS, nil
	} else if strings.HasPrefix(u.Scheme, "http") {
		return subscriber.RPC, nil
	}

	return 0, errors.New("unknown connection scheme")
}

func getManager(sub store.Subscription, p subscriber.Type) (subscriber.Manager, error) {
	switch sub.Endpoint.Type {
	case blockchain.ETH:
		return blockchain.CreateEthManager(p, sub.Ethereum), nil
	}

	return nil, errors.New("unknown blockchain type")
}

func getSubscriber(sub store.Subscription) (subscriber.ISubscriber, error) {
	connType, err := getConnectionType(sub.Endpoint.Url)
	if err != nil {
		return nil, err
	}

	manager, err := getManager(sub, connType)
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
