package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
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
	"strings"
	"time"
)

func main() {
	fmt.Println("Starting external initiator")

	_ = godotenv.Load()

	ei, err := loadExternalInitiator()
	if err != nil {
		log.Fatal(err)
	}

	ei.Run()
}

func loadExternalInitiator() (ExternalInitiator, error) {
	u, err := url.Parse(os.Getenv("CL_URL"))
	if err != nil {
		return ExternalInitiator{}, err
	}

	ei := ExternalInitiator{
		Subscriptions:       map[string]store.Subscription{},
		ActiveSubscriptions: []*ActiveSubscription{},
		Node: chainlink.Node{
			AccessKey:    os.Getenv("CL_ACCESS_KEY"),
			AccessSecret: os.Getenv("CL_ACCESS_SECRET"),
			Endpoint:     *u,
		},
	}

	db, err := store.ConnectToDb()
	if err != nil {
		return ei, err
	}
	defer db.Close()

	subscriptions, err := db.LoadSubscriptions()
	if err != nil {
		return ei, err
	}

	for _, subscription := range subscriptions {
		ei.Subscriptions[subscription.Id] = subscription
	}

	return ei, nil
}

type ExternalInitiator struct {
	Subscriptions       map[string]store.Subscription
	ActiveSubscriptions []*ActiveSubscription
	Node                chainlink.Node
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
	JobID      string   `json:"job_id"`
	Addresses  []string `json:"addresses"`
	Topics     []string `json:"topics"`
	RefreshInt int      `json:"refresh_interval"`
	Type       string   `json:"type"`
	Endpoint   string   `json:"endpoint"`
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

	err = ei.validateRequest(t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	u, _ := url.Parse(t.Endpoint)

	urlType := subscriber.RPC
	if strings.HasPrefix(u.Scheme, "ws") {
		urlType = subscriber.WS
	}

	sub := store.Subscription{
		Id:        uuid.New().String(),
		Job:       t.JobID,
		Addresses: t.Addresses,
		Topics:    t.Topics,
		Endpoint: store.Endpoint{
			Url:        *u,
			Type:       urlType,
			Blockchain: t.Type,
		},
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

func (ei ExternalInitiator) validateRequest(t RequestBody) error {
	validations := [...]int{
		len(t.JobID),
		len(t.Addresses) + len(t.Topics),
		len(t.Endpoint),
		len(t.Type),
	}

	for _, v := range validations {
		if v == 0 {
			return errors.New("missing required field(s)")
		}
	}

	_, err := url.Parse(t.Endpoint)
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
	if err := iSubscriber.Test(); err != nil {
		return err
	}

	var filter subscriber.Filter
	switch sub.Endpoint.Blockchain {
	case "eth":
		filter = blockchain.CreateEthFilterMessage(sub.Addresses, sub.Topics)
	default:
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
		Node:         ei.Node,
	}
	ei.ActiveSubscriptions = append(ei.ActiveSubscriptions, as)

	go ei.ListenForUpdates(as)

	return nil
}

type ActiveSubscription struct {
	Subscription store.Subscription
	Interface    subscriber.ISubscription
	Events       chan subscriber.Event
	Node         chainlink.Node
}

func (as ActiveSubscription) publishUpdates(event subscriber.Event) {
	err := as.Node.TriggerJob(as.Subscription.Job)
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
