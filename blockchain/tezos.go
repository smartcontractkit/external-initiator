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
		JobID:     sub.Job,
	}
}

type tezosSubscriber struct {
	Endpoint  string
	Addresses []string
	JobID     string
}

type tezosSubscription struct {
	endpoint    string
	events      chan<- subscriber.Event
	addresses   []string
	monitorResp *http.Response
	isDone      bool
	jobid       string
}

func (tz tezosSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ store.RuntimeConfig) (subscriber.ISubscription, error) {
	logger.Infof("Using Tezos RPC endpoint: %s\nListening for events on addresses: %v", tz.Endpoint, tz.Addresses)

	tzs := tezosSubscription{
		endpoint:  tz.Endpoint,
		events:    channel,
		addresses: tz.Addresses,
		jobid:     tz.JobID,
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
	logger.Debugf("Connected to RPC endpoint at %s, waiting for heads...\n", tzs.endpoint)

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

			logger.Debugf("Got new Tezos head: %s\n", blockID)
			blockJSON, err := tzs.getBlock(blockID)
			if err != nil {
				logger.Error(err)
				return
			}

			events, err := extractEventsFromBlock(blockJSON, tzs.addresses, tzs.jobid)
			if err != nil {
				logger.Error(err)
				return
			}

			logger.Debugf("%v events matching addresses %v\n", len(events), tzs.addresses)

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
		return nil, fmt.Errorf("%s returned 400. This endpoint may not support calls to /monitor", endpoint)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("Unexpected status code %v from endpoint %s", resp.StatusCode, endpoint)
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

func extractEventsFromBlock(data []byte, addresses []string, jobID string) ([]subscriber.Event, error) {
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
			// Check to see if this is a successful oracle request to
			// one of the oracle addresses we monitor.
			op, ok := getSuccessfulRequestCall(content, addresses)
			if !ok {
				continue
			}

			var args xtzArgs
			err := json.Unmarshal(op.Parameters.Value, &args)
			if err != nil {
				return nil, err
			}

			vals, err := args.GetValues()
			if err != nil {
				return nil, err
			}

			// Check if our jobID matches
			if !matchesXtzJobid(vals, jobID) {
				continue
			}

			params, err := getXtzKeyValues(vals)
			if err != nil {
				return nil, err
			}
			// Set the address to the oracle address.
			// The adapter will use this to fulfill the request.
			params["address"] = op.Destination
			params["request_id"], err = op.Result.GetRequestId()
			if err != nil {
				return nil, err
			}

			event, err := json.Marshal(params)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	}
	return events, nil
}

func matchesXtzJobid(values []string, expected string) bool {
	if len(values) < 4 {
		// Not enough params
		return false
	}

	jobID := values[3]
	return matchesJobID(expected, jobID)
}

func getSuccessfulRequestCall(content xtzTransactionContent, oracleAddresses []string) (xtzInternalOperationResult, bool) {
	if content.Metadata.OperationResult.Status != "applied" {
		// Transaction did not succeed
		return xtzInternalOperationResult{}, false
	}

	for _, op := range content.Metadata.InternalOperationResults {
		// Check for oracle request
		if op.Parameters.Entrypoint != "create_request" {
			continue
		}

		// Check if internal operation succeeded
		if op.Result.Status != "applied" {
			continue
		}

		// Check for the call from the Link token to the Oracle contract
		for _, address := range oracleAddresses {
			if op.Destination == address {
				// The destination is an oracle address we monitor,
				// this is a successful oracle request.
				return op, true
			}
		}
	}

	return xtzInternalOperationResult{}, false
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
	BalanceUpdates           []interface{}                `json:"balance_updates"`
	OperationResult          xtzOperationResult           `json:"operation_result"`
	InternalOperationResults []xtzInternalOperationResult `json:"internal_operation_results"`
}

type xtzInternalOperationResult struct {
	Kind        string                         `json:"kind"`
	Source      string                         `json:"source"`
	Nonce       int                            `json:"nonce"`
	Amount      string                         `json:"amount"`
	Destination string                         `json:"destination"`
	Parameters  xtzInternalOperationParameters `json:"parameters"`
	Result      xtzOperationResult             `json:"result"`
}

type xtzOperationResult struct {
	Status              string          `json:"status"`
	Errors              []interface{}   `json:"errors"`
	Storage             json.RawMessage `json:"storage"`
	BalanceUpdates      []interface{}   `json:"balance_updates"`
	ConsumedGas         string          `json:"consumed_gas"`
	StorageSize         string          `json:"storage_size"`
	PaidStorageSizeDiff string          `json:"paid_storage_size_diff"`
	BigMapDiff          []xtzBigMapDiff `json:"big_map_diff"`
}

type xtzBigMapDiff struct {
	Action  string      `json:"action"`
	BigMap  string      `json:"big_map"`
	KeyHash string      `json:"key_hash"`
	Key     interface{} `json:"key"`
	Value   xtzValue    `json:"value"`
}

// Since there is no event-based request_id, we need to check
// the storage changes in the operation to see the oracle
// request id.
func (op xtzOperationResult) GetRequestId() (string, error) {
	var arg xtzArgs
	err := json.Unmarshal(op.Storage, &arg)
	if err != nil {
		return "", err
	}

	vals, err := arg.GetValues()
	if err != nil {
		return "", err
	}

	// We expect to get the big_map index in the
	// 7th value from storage changes.
	if len(vals) < 7 {
		return "", errors.New("expected at least 7 storage changes")
	}

	bigMapIndex := vals[6]

	for _, bmd := range op.BigMapDiff {
		if bmd.BigMap != bigMapIndex {
			continue
		}
		return bmd.Value.Int, nil
	}

	return "", errors.New("unable to find request_id")
}

type xtzInternalOperationParameters struct {
	Entrypoint string          `json:"entrypoint"`
	Value      json.RawMessage `json:"value"`
}

func getXtzKeyValues(vals []string) (map[string]string, error) {
	if len(vals) < 7 {
		return nil, errors.New("not enough values provided")
	}

	// Values #1, #2, #3, #4 are client, amount, (client) request_id and job_id.
	// The last two values are target and timeout.
	// We ignore these when converting to key-value arrays,
	// then we add the necessary values with correct keys.
	kv := convertStringArrayToKV(vals[4 : len(vals)-2])
	kv["payment"] = vals[1]
	kv["request_id"] = vals[2]
	return kv, nil
}

type xtzValue struct {
	Bytes  string `json:"bytes,omitempty"`
	Int    string `json:"int,omitempty"`
	String string `json:"string,omitempty"`
}

type xtzArgs struct {
	xtzValue
	Prim string            `json:"prim,omitempty"`
	Args []json.RawMessage `json:"args,omitempty"`
}

func (a xtzArgs) GetValue() string {
	if a.Bytes != "" {
		return a.Bytes
	} else if a.Int != "" {
		return a.Int
	} else if a.String != "" {
		return a.String
	}
	return ""
}

func (a xtzArgs) GetValues() ([]string, error) {
	// If we don't have any args, return the value
	if len(a.Args) < 1 {
		return []string{a.GetValue()}, nil
	}

	args, err := a.GetArgs()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, arg := range args {
		values, err := arg.GetValues()
		if err != nil {
			return nil, err
		}
		result = append(result, values...)
	}

	return result, nil
}

func (a xtzArgs) GetArgs() ([]xtzArgs, error) {
	if len(a.Args) < 1 {
		return nil, nil
	}

	var args []xtzArgs
	// Iterate the array of objects _or_ arrays
	for _, bz := range a.Args {
		var arg xtzArgs
		err := json.Unmarshal(bz, &arg)
		// If we cannot unmarshal to an object, check if we can unmarshal to an array of objects
		if err != nil {
			var argArgs []xtzArgs
			err = json.Unmarshal(bz, &argArgs)
			if err != nil {
				// This was neither an object or an array of objects, return error
				return nil, err
			}
			// Unmarshal succeeded, add the objects to our result list
			args = append(args, argArgs...)
			continue
		}
		// Our initial object unmarshal succeeded, add the object to result list
		args = append(args, arg)
	}

	return args, nil
}
