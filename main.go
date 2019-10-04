package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"time"
)

var opts struct {
	Endpoints []map[string]string `short:"e" long:"endpoint" description:"A list of endpoints as name:url"`
}

func main() {
	fmt.Println("Starting external initiator")

	_ = godotenv.Load()

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	ei, err := loadExternalInitiator()
	if err != nil {
		log.Fatal(err)
	}

	if len(ei.Endpoints) == 0 {
		log.Fatal("No qualified endpoints provided")
	}

	ei.Run()
}

func loadExternalInitiator() (ExternalInitiator, error) {
	ei := ExternalInitiator{
		Endpoints:           map[string]store.Endpoint{},
		Subscriptions:       map[string]store.Subscription{},
		ActiveSubscriptions: []*ActiveSubscription{},
	}

	db, err := store.ConnectToDb()
	if err != nil {
		return ei, err
	}
	defer db.Close()

	endpoints, err := db.LoadEndpoints()
	if err != nil {
		return ei, err
	}

	for _, endpoint := range endpoints {
		fmt.Printf("Loading endpoint from DB: %s - %s\n", endpoint.Blockchain, endpoint.Url.String())
		ei.Endpoints[endpoint.Blockchain] = endpoint
	}

	subscriptions, err := db.LoadSubscriptions()
	if err != nil {
		return ei, err
	}

	for _, subscription := range subscriptions {
		ei.Subscriptions[subscription.Id] = subscription
	}

	for _, e := range opts.Endpoints {
		for key, val := range e {
			u, err := url.Parse(val)
			if err != nil {
				fmt.Println("Error parsing endpoint URL: ", err)
				continue
			}

			switch key {
			case "eth":
				fmt.Printf("Adding ETH endpoint: %s\n", u.String())
				endpoint := store.Endpoint{
					Url:        *u,
					Type:       subscriber.WS,
					Blockchain: "eth",
				}
				ei.Endpoints["eth"] = endpoint
				err := db.SaveEndpoint(endpoint)
				if err != nil {
					fmt.Println(err)
				}
			default:
				fmt.Println("Unknown blockchain for endpoint: ", key)
			}
		}
	}

	return ei, nil
}

type ExternalInitiator struct {
	Endpoints           map[string]store.Endpoint
	Subscriptions       map[string]store.Subscription
	ActiveSubscriptions []*ActiveSubscription
}

func (ei ExternalInitiator) listenForShutdown(interrupt chan os.Signal) {
	<-interrupt
	fmt.Println("Shutting down...")
	for _, sub := range ei.ActiveSubscriptions {
		sub.Interface.Unsubscribe()
		close(sub.Events)
	}
	fmt.Println("All subscriptions closed. Bye!")
	os.Exit(0)
}

func (ei ExternalInitiator) Run() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go ei.listenForShutdown(interrupt)

	for _, sub := range ei.Subscriptions {
		err := ei.subscribe(sub)
		if err != nil {
			fmt.Println(err)
		}
	}

	ei.listenOnPort(8080)
}

func (ei ExternalInitiator) listenOnPort(port int) {
	http.HandleFunc("/", ei.handler)

	fmt.Printf("Ready and listening on port :%v\n", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

type RequestBody struct {
	JobID        string   `json:"job_id"`
	Addresses    []string `json:"addresses"`
	Topics       []string `json:"topics"`
	RefreshInt   int      `json:"refresh_interval"`
	NodeURL      string   `json:"node_url"`
	AccessKey    string   `json:"access_key"`
	AccessSecret string   `json:"access_secret"`
}

func (ei ExternalInitiator) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var t RequestBody
	err := decoder.Decode(&t)
	if err != nil {
		fmt.Println("Error parsing request: ", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Could not parse request")
		return
	}

	p := r.URL.Path
	base := path.Base(p)

	err = ei.validateRequest(t, base)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	e, _ := url.Parse(t.NodeURL)
	endpoint := ei.Endpoints[base]

	sub := store.Subscription{
		Id:        uuid.New().String(),
		Job:       t.JobID,
		Addresses: t.Addresses,
		Topics:    t.Topics,
		Node: chainlink.Node{
			AccessKey:    t.AccessKey,
			AccessSecret: t.AccessSecret,
			Endpoint:     *e,
		},
		Endpoint:   endpoint,
		RefreshInt: t.RefreshInt,
	}

	err = ei.saveAndSubscribe(sub)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, sub.Id)
}

func (ei ExternalInitiator) validateRequest(t RequestBody, base string) error {
	if base == "." || base == "/" {
		return errors.New("missing endpoint name")
	}

	_, ok := ei.Endpoints[base]
	if !ok {
		return errors.New("endpoint not found")
	}

	validations := [5]int{
		len(t.JobID),
		len(t.Addresses) + len(t.Topics),
		len(t.NodeURL),
		len(t.AccessKey),
		len(t.AccessSecret),
	}

	for _, v := range validations {
		if v == 0 {
			return errors.New("missing required field(s)")
		}
	}

	_, err := url.Parse(t.NodeURL)
	if err != nil {
		return err
	}

	return nil
}

func (ei ExternalInitiator) saveAndSubscribe(sub store.Subscription) error {
	db, err := store.ConnectToDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.SaveSubscription(sub)
	if err != nil {
		return err
	}

	return ei.subscribe(sub)
}

func (ei ExternalInitiator) subscribe(sub store.Subscription) error {
	events := make(chan subscriber.Event)
	iSubscriber := getSubscriber(sub.Endpoint.Url, sub.RefreshInt)

	var filter subscriber.Filter
	if sub.Endpoint.Blockchain == "blockchain" {
		filter = blockchain.CreateEthFilterMessage(sub.Addresses, sub.Topics)
	} else {
		return errors.New(fmt.Sprintf("Unable to subscribe to blockchain %s", sub.Endpoint.Blockchain))
	}

	subscription, err := iSubscriber.SubscribeToEvents(events, filter)
	if err != nil {
		log.Fatal(err)
	}

	as := &ActiveSubscription{
		Subscription: sub,
		Interface:    subscription,
		Events:       events,
	}
	ei.ActiveSubscriptions = append(ei.ActiveSubscriptions, as)

	go ei.ListenForUpdates(as)

	return nil
}

type ActiveSubscription struct {
	Subscription store.Subscription
	Interface    subscriber.ISubscription
	Events       chan subscriber.Event
}

func (as ActiveSubscription) publishUpdates(event subscriber.Event) {
	err := as.Subscription.Node.TriggerJob(as.Subscription.Job)
	if err != nil {
		fmt.Println(err)
	}
}

func (ei ExternalInitiator) ListenForUpdates(sub *ActiveSubscription) {
	for {
		event, ok := <-sub.Events
		if !ok {
			return
		}
		sub.publishUpdates(event)
	}
}

func getSubscriber(endpointUrl url.URL, interval int) subscriber.ISubscriber {
	var sub subscriber.ISubscriber
	if endpointUrl.Scheme == "ws" || endpointUrl.Scheme == "wss" {
		sub = subscriber.WebsocketSubscriber{Endpoint: endpointUrl}
	} else {
		sub = subscriber.RpcSubscriber{Endpoint: endpointUrl, Interval: time.Duration(interval) * time.Second}
	}
	return sub
}
