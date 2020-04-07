package blockchain

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	XTZ                  = "tezos"
	monitorRetryInterval = 5 * time.Second
)

func createTezosSubscriber(sub store.Subscription) TezosSubscriber {
	return TezosSubscriber{
		Endpoint:  strings.TrimSuffix(sub.Endpoint.Url, "/"),
		Addresses: sub.Tezos.Addresses,
	}
}

type TezosSubscriber struct {
	Endpoint  string
	Addresses []string
}

type TezosSubscription struct {
	endpoint    string
	events      chan<- subscriber.Event
	addresses   []string
	monitorResp *http.Response
	isDone      bool
}

func (tz TezosSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ ...interface{}) (subscriber.ISubscription, error) {
	log.Printf("Using Tezos RPC endpoint: %s\nListening for events on addresses: %v\n", tz.Endpoint, tz.Addresses)

	tzs := TezosSubscription{
		endpoint:  tz.Endpoint,
		events:    channel,
		addresses: tz.Addresses,
	}

	go tzs.readMessagesWithRetry()

	return tzs, nil
}

func (tz TezosSubscriber) Test() error {
	resp, err := monitor(tz.Endpoint)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (tzs TezosSubscription) readMessagesWithRetry() {
	for {
		tzs.readMessages()
		if !tzs.isDone {
			time.Sleep(monitorRetryInterval)
			continue
		}
		return
	}
}

func (tzs TezosSubscription) readMessages() {
	resp, err := monitor(tzs.endpoint)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Connected to RPC endpoint at %s, waiting for heads...\n", tzs.endpoint)

	reader := bufio.NewReader(resp.Body)

	lines := make(chan []byte)
	go tzs.readLines(lines, reader)

	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return
			}

			blockID, err := extractBlockIDFromHeaderJSON(line)
			if err != nil {
				log.Println(err)
				return
			}

			log.Printf("Got new Tezos head: %s\n", blockID)
			blockJSON, err := tzs.getBlock(blockID)
			if err != nil {
				log.Println(err)
				return
			}

			events, err := extractEventsFromBlock(blockJSON, tzs.addresses)
			if err != nil {
				log.Println(err)
				return
			}

			log.Printf("%v events matching addresses %v\n", len(events), tzs.addresses)

			for _, event := range events {
				tzs.events <- event
			}
		}
	}
}

func monitor(endpoint string) (*http.Response, error) {
	resp, err := http.Get(fmt.Sprintf("%s/monitor/heads/main", endpoint))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 400 {
		resp.Body.Close()
		return nil, errors.New(fmt.Sprintf("%s returned 400. This endpoint may not support calls to /monitor", endpoint))
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, errors.New(fmt.Sprintf("Unexpected status code %v from endpoint %s", resp.StatusCode, endpoint))
	}
	return resp, nil
}

func (tzs TezosSubscription) readLines(lines chan []byte, reader *bufio.Reader) {
	defer close(lines)
	for {
		line, err := reader.ReadBytes('\n')
		if tzs.isDone {
			return
		}
		if err == io.EOF {
			log.Printf("Lost connection to Tezos RPC node, retrying in %v...\n", monitorRetryInterval)
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
		lines <- line
	}
}

func (tzs TezosSubscription) getBlock(blockID string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/chains/main/blocks/%s/operations", tzs.endpoint, blockID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (tzs TezosSubscription) Unsubscribe() {
	log.Println("Unsubscribing from Tezos endpoint", tzs.endpoint)
	tzs.isDone = true
	if tzs.monitorResp != nil {
		tzs.monitorResp.Body.Close()
	}
}

func extractEventsFromBlock(data []byte, addresses []string) ([]subscriber.Event, error) {
	if !gjson.ValidBytes(data) {
		return nil, errors.New("got invalid JSON object from Tezos RPC endpoint")
	}

	/*
		This data structure is not documented well. From some guy in the Tezos slack:

		> There are around 10 types of operations and they are grouped into four groups.
		> Endorsements (consensus operations ) are in the first group, transactions into
		> the 4th group. So in that block there are no operations for the second group
		> (here there would operations related to voting) and 3rd group.
		> The groups are called "validation passes". See for instance, the next RPC in that list.
		> You can see the group of each type of operations here: https://gitlab.com/nomadic-labs/tezos/blob/master/src/proto_alpha/lib_protocol/operation_repr.ml#L670.
		> I guess this is documented somewhere but I don't where...

		Tezos is self-modifying so it's entirely possible (though unlikely) that
		this could change in future versions of the protocol.
	*/
	managerOps := gjson.GetBytes(data, "3")

	raw := data[managerOps.Index : managerOps.Index+len(managerOps.Raw)]
	var transactions []XtzTransaction
	err := json.Unmarshal(raw, &transactions)
	if err != nil {
		return nil, err
	}
	var events []subscriber.Event
	for _, t := range transactions {
		/*
		 There is no official documentation on this, but according to Alex Eichhorn:
		 > there is a concept of batch transactions (not officially called that way) where a list of multiple operations is signed at once by the same key
		 > wallets use it to issue reveal+transaction or reveal+origination etc, bakers use it for batch payouts
		 > AFAIK the list cannot be empty

		 Note that this list only contains user-initiated calls.
		 SC-initiated calls are buried deep inside the call to that SC, as a callback (return value of that SC).
		 You can find this under metadata->internal_operation_results
		*/
		for _, content := range t.Contents {
			if hasDestinationInAddresses(content, addresses) {
				event, err := t.toEvent()
				if err != nil {
					return nil, err
				}
				events = append(events, event)
			}
		}
	}
	return events, nil
}

func hasDestinationInAddresses(content XtzTransactionContent, addresses []string) bool {
	for _, address := range addresses {
		if address == content.Destination {
			return true
		}
		if content.Metadata.InternalOperationResults != nil {
			for _, internalOperationResult := range *content.Metadata.InternalOperationResults {
				if address == internalOperationResult.Destination {
					return true
				}
			}
		}
	}
	return false
}

func extractBlockIDFromHeaderJSON(data []byte) (string, error) {
	var header XtzHeader
	err := json.Unmarshal(data, &header)
	if err != nil {
		return "", err
	}
	if header.Hash == "" {
		return "", errors.New("could not extract block ID")
	}

	return header.Hash, nil
}

func (t XtzTransaction) toEvent() (subscriber.Event, error) {
	event, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return event, nil
}

type XtzHeader struct {
	Hash           string   `json:"hash"`
	Level          int      `json:"level"`
	Proto          int      `json:"proto"`
	Predecessor    string   `json:"predecessor"`
	Timestamp      string   `json:"timestamp"`
	ValidationPass int      `json:"validation_pass"`
	OperationsHash string   `json:"operations_hash"`
	Fitness        []string `json:"fitness"`
	Context        string   `json:"context"`
	ProtocolData   string   `json:"protocol_data"`
}

type XtzTransaction struct {
	Protocol string                  `json:"protocol"`
	ChainID  string                  `json:"chain_id"`
	Hash     string                  `json:"hash"`
	Branch   string                  `json:"branch"`
	Contents []XtzTransactionContent `json:"contents"`
}

type XtzTransactionContent struct {
	Kind         string                        `json:"kind"`
	Source       string                        `json:"source"`
	Fee          string                        `json:"fee"`
	Counter      string                        `json:"counter"`
	GasLimit     string                        `json:"gas_limit"`
	StorageLimit string                        `json:"storage_limit"`
	Amount       string                        `json:"amount"`
	Destination  string                        `json:"destination"`
	Parameters   interface{}                   `json:"parameters"`
	Metadata     XtzTransactionContentMetadata `json:"metadata"`
}

type XtzTransactionContentMetadata struct {
	BalanceUpdates           []interface{}                 `json:"balance_updates"`
	OperationResult          interface{}                   `json:"operation_result"`
	InternalOperationResults *[]XtzInternalOperationResult `json:"internal_operation_results"`
}

type XtzInternalOperationResult struct {
	Kind        string      `json:"kind"`
	Source      string      `json:"source"`
	Nonce       int         `json:"nonce"`
	Amount      string      `json:"amount"`
	Destination string      `json:"destination"`
	Parameters  interface{} `json:"parameters"`
	Result      interface{} `json:"result"`
}
