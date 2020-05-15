package blockchain

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/tidwall/gjson"
)

const (
	XTZ                  = "tezos"
	monitorRetryInterval = 5 * time.Second
)

func createTezosSubscriber(sub store.Subscription) tezosSubscriber {
	return tezosSubscriber{
		Endpoint:  strings.TrimSuffix(sub.Endpoint.Url, "/"),
		Addresses: sub.Tezos.Addresses,
	}
}

type tezosSubscriber struct {
	Endpoint  string
	Addresses []string
}

type tezosSubscription struct {
	endpoint    string
	events      chan<- subscriber.Event
	addresses   []string
	monitorResp *http.Response
	isDone      bool
}

func (tz tezosSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ ...interface{}) (subscriber.ISubscription, error) {
	logger.Infof("Using Tezos RPC endpoint: %s\nListening for events on addresses: %v\n", tz.Endpoint, tz.Addresses)

	tzs := tezosSubscription{
		endpoint:  tz.Endpoint,
		events:    channel,
		addresses: tz.Addresses,
	}

	go tzs.readMessagesWithRetry()

	return tzs, nil
}

func (tz tezosSubscriber) Test() error {
	resp, err := monitor(tz.Endpoint)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (tzs tezosSubscription) readMessagesWithRetry() {
	for {
		tzs.readMessages()
		if !tzs.isDone {
			time.Sleep(monitorRetryInterval)
			continue
		}
		return
	}
}

func (tzs tezosSubscription) readMessages() {
	resp, err := monitor(tzs.endpoint)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	logger.Infof("Connected to RPC endpoint at %s, waiting for heads...\n", tzs.endpoint)

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
				logger.Error(err)
				return
			}

			logger.Infof("Got new Tezos head: %s\n", blockID)
			blockJSON, err := tzs.getBlock(blockID)
			if err != nil {
				logger.Error(err)
				return
			}

			events, err := extractEventsFromBlock(blockJSON, tzs.addresses)
			if err != nil {
				logger.Error(err)
				return
			}

			logger.Infof("%v events matching addresses %v\n", len(events), tzs.addresses)

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

func (tzs tezosSubscription) readLines(lines chan []byte, reader *bufio.Reader) {
	defer close(lines)
	for {
		line, err := reader.ReadBytes('\n')
		if tzs.isDone {
			return
		}
		if err == io.EOF {
			lines <- line
			logger.Warnf("Lost connection to Tezos RPC node, retrying in %v...\n", monitorRetryInterval)
			return
		}
		if err != nil {
			logger.Error(err)
			return
		}
		lines <- line
	}
}

func (tzs tezosSubscription) getBlock(blockID string) ([]byte, error) {
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

func (tzs tezosSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from Tezos endpoint", tzs.endpoint)
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
	var transactions []xtzTransaction
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

func hasDestinationInAddresses(content xtzTransactionContent, addresses []string) bool {
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
	var header xtzHeader
	err := json.Unmarshal(data, &header)
	if err != nil {
		return "", err
	}
	if header.Hash == "" {
		return "", errors.New("could not extract block ID")
	}

	return header.Hash, nil
}

func (t xtzTransaction) toEvent() (subscriber.Event, error) {
	event, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return event, nil
}

type xtzHeader struct {
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

type xtzTransaction struct {
	Protocol string                  `json:"protocol"`
	ChainID  string                  `json:"chain_id"`
	Hash     string                  `json:"hash"`
	Branch   string                  `json:"branch"`
	Contents []xtzTransactionContent `json:"contents"`
}

type xtzTransactionContent struct {
	Kind         string                        `json:"kind"`
	Source       string                        `json:"source"`
	Fee          string                        `json:"fee"`
	Counter      string                        `json:"counter"`
	GasLimit     string                        `json:"gas_limit"`
	StorageLimit string                        `json:"storage_limit"`
	Amount       string                        `json:"amount"`
	Destination  string                        `json:"destination"`
	Parameters   interface{}                   `json:"parameters"`
	Metadata     xtzTransactionContentMetadata `json:"metadata"`
}

type xtzTransactionContentMetadata struct {
	BalanceUpdates           []interface{}                 `json:"balance_updates"`
	OperationResult          interface{}                   `json:"operation_result"`
	InternalOperationResults *[]xtzInternalOperationResult `json:"internal_operation_results"`
}

type xtzInternalOperationResult struct {
	Kind        string      `json:"kind"`
	Source      string      `json:"source"`
	Nonce       int         `json:"nonce"`
	Amount      string      `json:"amount"`
	Destination string      `json:"destination"`
	Parameters  interface{} `json:"parameters"`
	Result      interface{} `json:"result"`
}
