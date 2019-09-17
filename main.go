package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/eth"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Config struct {
	Job  string
	Node chainlink.Node
}

func main() {
	fmt.Println("Starting external initiator")

	_ = godotenv.Load()

	jobid, clUrl, endpointUrl, err := loadRequiredEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	conf := Config{
		Job: jobid,
		Node: chainlink.Node{
			Endpoint:     *clUrl,
			AccessKey:    os.Getenv("CL_ACCESS_KEY"),
			AccessSecret: os.Getenv("CL_ACCESS_SECRET"),
		},
	}

	ws := flag.Bool("ws", true, "use websockets endpoint")

	sub, err := getSubscriber(*endpointUrl, *ws)
	if err != nil {
		log.Fatal(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	subscribeToEvents(sub, interrupt, conf.PublishUpdates)
}

func getSubscriber(endpointUrl url.URL, ws bool) (subscriber.ISubscriber, error) {
	var sub subscriber.ISubscriber
	if ws {
		fmt.Println("Using WS")
		if endpointUrl.Scheme != "ws" && endpointUrl.Scheme != "wss" {
			return nil, errors.New("invalid endpoint scheme for websocket")
		}

		sub = subscriber.WebsocketSubscriber{Endpoint: endpointUrl}
	} else {
		fmt.Println("Using RPC")
		sub = subscriber.RpcSubscriber{Endpoint: endpointUrl, Interval: 10 * time.Second}
	}
	return sub, nil
}

func loadRequiredEnvVars() (jobid string, clUrl *url.URL, endpointUrl *url.URL, err error) {
	endpoint := os.Getenv("ENDPOINT")
	if len(endpoint) == 0 {
		err = errors.New("missing ENDPOINT")
		return
	}

	clEndpoint := os.Getenv("CHAINLINK_URL")
	if len(clEndpoint) == 0 {
		err = errors.New("missing CHAINLINK_URL")
		return
	}

	jobid = os.Getenv("JOBID")
	if len(jobid) == 0 {
		err = errors.New("missing JOBID")
		return
	}

	clUrl, err = url.Parse(clEndpoint)
	if err != nil {
		return
	}

	endpointUrl, err = url.Parse(endpoint)
	return
}

func subscribeToEvents(sub subscriber.ISubscriber, int chan os.Signal, publisher func(event subscriber.Event)) {
	ch := make(chan subscriber.Event)

	filter := eth.CreateFilterMessage(os.Getenv("ADDRESS"), strings.Split(os.Getenv("TOPICS"), ","))

	subscription, err := sub.SubscribeToEvents(ch, filter)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-ch:
			go publisher(event)
		case <-int:
			fmt.Println("Got interrupt, quitting")
			subscription.Unsubscribe()
			return
		}
	}
}

func (config Config) PublishUpdates(event subscriber.Event) {
	err := config.Node.TriggerJob(config.Job)
	if err != nil {
		fmt.Println(err)
	}
}
