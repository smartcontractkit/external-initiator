package blockchain

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
		Endpoint:     strings.TrimSuffix(sub.Endpoint.Url, "/"),
		EndpointName: sub.EndpointName,
		Addresses:    sub.Tezos.Addresses,
		JobID:        sub.Job,
	}
}

type tezosSubscriber struct {
	Endpoint     string
	EndpointName string
	Addresses    []string
	JobID        string
}

type tezosSubscription struct {
	endpoint     string
	endpointName string
	events       chan<- subscriber.Event
	addresses    []string
	monitorResp  *http.Response
	isDone       bool
	jobid        string
}

func (tz tezosSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ store.RuntimeConfig) (subscriber.ISubscription, error) {
	logger.Infof("Using Tezos RPC endpoint: %s\nListening for events on addresses: %v", tz.Endpoint, tz.Addresses)

	tzs := tezosSubscription{
		endpoint:     tz.Endpoint,
		endpointName: tz.EndpointName,
		events:       channel,
		addresses:    tz.Addresses,
		jobid:        tz.JobID,
	}

	go tzs.readMessagesWithRetry()

	return tzs, nil
}

func (tz tezosSubscriber) Test() error {
	resp, err := monitor(tz.Endpoint)
	if err != nil {
		return err
	}
	return resp.Body.Close()
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
	defer logger.ErrorIfCalling(resp.Body.Close)
	logger.Debugf("Connected to RPC endpoint at %s, waiting for heads...\n", tzs.endpoint)

	reader := bufio.NewReader(resp.Body)

	lines := make(chan []byte)
	go tzs.readLines(lines, reader)

	for {
		line, ok := <-lines
		if !ok {
			return
		}
		promLastSourcePing.With(prometheus.Labels{"endpoint": tzs.endpointName, "jobid": tzs.jobid}).SetToCurrentTime()

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

		events, err := tzs.extractEventsFromBlock(blockJSON, tzs.addresses, tzs.jobid)
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

func monitor(endpoint string) (*http.Response, error) {
	resp, err := http.Get(fmt.Sprintf("%s/monitor/heads/main", endpoint))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 400 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s returned 400. This endpoint may not support calls to /monitor", endpoint)
	}
	if resp.StatusCode != 200 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %v from endpoint %s", resp.StatusCode, endpoint)
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
	defer logger.ErrorIfCalling(resp.Body.Close)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
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
	Action  string `json:"action"`
	BigMap  string `json:"big_map"`
	KeyHash string `json:"key_hash"`
	// (key, value) can be a JSON array ([]) or a JSON object ({})
	Key   michelsonExpression `json:"key"`
	Value json.RawMessage     `json:"value"`
}

type xtzInternalOperationParameters struct {
	Entrypoint string          `json:"entrypoint"`
	Value      json.RawMessage `json:"value"`
}

const (
	requestFieldID            string = "id"
	requestFieldJobID         string = "jobId"
	requestFieldOracleAddress string = "oracleAddress"
	requestFieldParameters    string = "parameters"
)

// Core data types (https://tezos.gitlab.io/alpha/michelson.html#core-data-types-and-notations)
const (
	primTypeInt       = "int"       // "prim":"int"
	primTypeNat       = "nat"       // "prim":"nat"
	primTypeTimestamp = "timestamp" // "prim":"timestamp"
	primTypeAddress   = "address"   // "prim":"address"
	primTypeString    = "string"    // "prim":"string"
	primTypeBytes     = "bytes"     // "prim":"bytes"
	primTypeBool      = "bool"      // "prim":"bool"
	primTypePair      = "pair"      // "prim":"pair"
	primTypeMap       = "map"       // "prim":"map"
	primTypeBigMap    = "big_map"   // "prim":"big_map"
)

// https://tezos.gitlab.io/alpha/michelson.html#full-grammar
const (
	primValuePair  = "Pair"  // "prim":"Pair"
	primValueElt   = "Elt"   // "prim":"Elt"
	primValueLeft  = "Left"  // "prim":"Left"
	primValueRight = "Right" // "prim":"Right"
	primValueTrue  = "True"  // "prim":"True"
	primValueFalse = "False" // "prim":"False"
)

type michelsonExpression struct {
	Bytes  string `json:"bytes,omitempty"`
	Int    string `json:"int,omitempty"`
	String string `json:"string,omitempty"`
	// Generic prim (can have args and annots)
	Prim string `json:"prim,omitempty"`
	// Args cannot be of type []michelsonExpression,
	// since it can contain arrays and objects as items
	Args   []json.RawMessage `json:"args,omitempty"`
	Annots []string          `json:"annots,omitempty"`
}

// Get contract storage type and value
func (tzs tezosSubscription) getContractStorage(contractAddress string) (michelsonExpression, json.RawMessage, error) {
	var typeExpr michelsonExpression
	var valueRaw json.RawMessage

	// Get contract script in "Optimized" form
	postBody, _ := json.Marshal(map[string]string{
		"unparsing_mode": "Optimized",
	})
	resp, err := http.Post(
		fmt.Sprintf("%s/chains/main/blocks/head/context/contracts/%s/script/normalized", tzs.endpoint, contractAddress),
		"application/json",
		bytes.NewBuffer(postBody),
	)
	if err != nil {
		return typeExpr, valueRaw, err
	}
	defer logger.ErrorIfCalling(resp.Body.Close)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return typeExpr, valueRaw, err
	}

	// Get the storage type
	// { "code": [{...}, {"prim": "storage", args [{}]}, {...}], "storage": {} }
	// The line below gets this item --------------^
	result := gjson.GetBytes(body, `code.#(prim=="storage").args.0`)
	var raw []byte
	if result.Index > 0 {
		raw = body[result.Index : result.Index+len(result.Raw)]
	} else {
		raw = []byte(result.Raw)
	}
	typeExpr, err = unmarshalExpression(raw)
	if err != nil {
		return typeExpr, valueRaw, err
	}

	// Get the storage value
	// { "code": [], "storage": {} }
	// -------------------------^
	result = gjson.GetBytes(body, "storage")
	valueRaw = json.RawMessage(result.Raw)

	return typeExpr, valueRaw, nil
}

func (tzs tezosSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from Tezos endpoint", tzs.endpoint)
	tzs.isDone = true
	if tzs.monitorResp != nil {
		logger.ErrorIf(tzs.monitorResp.Body.Close())
	}
}

func (tzs tezosSubscription) extractEventsFromBlock(data []byte, addresses []string, jobID string) ([]subscriber.Event, error) {
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
		 > there is a concept of operation batch where a list of multiple operations is signed at once by the same key
		 > wallets use it to issue reveal+transaction or reveal+origination etc, bakers use it for batch payouts
		 > AFAIK the list cannot be empty

		 Note that this list only contains user-initiated calls. (implicit account (tz1) calls)

		 User initiated transactions can originate "internal calls" that
		 can be found under metadata->internal_operation_results
		*/
		for _, content := range t.Contents {
			// Check to see if this is a successful oracle request to
			// one of the oracle addresses we monitor.
			ops := getSuccessfulRequestCalls(content, addresses)

			// A transaction could originate multiple requests
			for _, op := range ops {

				// Get the request type from the oracle storage type, it contains
				// the label(called annotation) and the type of each value
				typeExpr, valueRaw, err := tzs.getContractStorage(op.Destination)
				if err != nil {
					return nil, err
				}

				// Extract "%requests" big_map ID to allow custom storage structs as long as mandatory annotations
				// are present in the type expression. (%requests, %id, %jobID, %parameters)
				bigMapID, err := extractBigMapID(typeExpr, valueRaw)
				if err != nil {
					return nil, err
				}

				// Get big_map_diff associated with the big_map ID abtained above
				bigMapDiff, err := getBigMapDiff(bigMapID, op.Result.BigMapDiff)
				if err != nil {
					return nil, err
				}

				// Get the big_map type expression
				bigMapTypeExpr, err := getRequestType(typeExpr)
				if err != nil {
					return nil, err
				}

				// Extract the new request from the big_map_diff
				values, err := extractValuesFromExpression(bigMapTypeExpr, bigMapDiff.Value)
				if err != nil {
					return nil, err
				}

				if !matchesJobID(jobID, values[requestFieldJobID].(string)) {
					continue
				}

				// Extract request id from the big_map_diff
				requestId, err := extractVariantValue(bigMapDiff.Key)
				if err != nil {
					return nil, err
				}

				// Add request arguments (a clean map with only the mandatory arguments)
				requestArguments := make(map[string]interface{})
				requestArguments[requestFieldParameters] = values[requestFieldParameters]
				requestArguments[requestFieldJobID] = values[requestFieldJobID]
				requestArguments[requestFieldOracleAddress] = op.Destination
				requestArguments[requestFieldID] = requestId

				event, err := json.Marshal(requestArguments)
				if err != nil {
					return nil, err
				}
				events = append(events, event)
			}
		}
	}
	return events, nil
}

// Get all internal operations that were:
// - successful applied;
// - destination address is in oracleAddressess
// - called entrypoint is "on_token_transfer" or "create_request"
func getSuccessfulRequestCalls(content xtzTransactionContent, oracleAddresses []string) []xtzInternalOperationResult {
	internalOps := []xtzInternalOperationResult{}

	if content.Metadata.OperationResult.Status != "applied" {
		// Transaction did not succeed
		return internalOps
	}

	for _, op := range content.Metadata.InternalOperationResults {
		// Check if any transaction called the expected entry-points
		ep := op.Parameters.Entrypoint
		if ep != "create_request" && ep != "on_token_transfer" {
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
				internalOps = append(internalOps, op)
			}
		}
	}

	return internalOps
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

// Extract the request values map(key => value) from the expression,
// where (key) is the annotation obtained from the storage type
func extractValuesFromExpression(typeExpr michelsonExpression, valueRaw json.RawMessage) (map[string]interface{}, error) {
	req := make(map[string]interface{})

	valueExpr, err := convertRawToMichelsonExpression(valueRaw)
	if err != nil {
		// Raw was not a valid michelson expression
		return req, err
	}

	switch typeExpr.Prim {
	case primTypeInt, primTypeNat, primTypeTimestamp, primTypeBigMap,
		primTypeString, primTypeBytes, primTypeAddress, primTypeBool:
		// Extract field name from the annotation
		field, err := getAnnotation(typeExpr.Annots)
		if err != nil {
			return req, err
		}
		if req[field], err = extractVariantValue(valueExpr); err != nil {
			return req, err
		}
	case primTypeMap:
		// Extract field name from the annotation
		field, err := getAnnotation(typeExpr.Annots)
		if err != nil {
			return nil, err
		}

		// Extract key and values from the map
		keyValues, err := extractMapKeyValues(valueExpr.Args)
		if err != nil {
			return nil, err
		}

		req[field] = keyValues
	case primTypePair:

		// Normalize type and value
		typeExpr, valueExpr, err := normalizeTypeAndValue(typeExpr, valueExpr)
		if err != nil {
			return nil, err
		}

		for i, reqType := range typeExpr.Args {
			// Get type expression
			tExpr, err := unmarshalExpression(reqType)
			if err != nil {
				return nil, err
			}

			keyValues, err := extractValuesFromExpression(tExpr, valueExpr.Args[i])
			if err != nil {
				return nil, err
			}
			for k, v := range keyValues {
				req[k] = v
			}
		}
	default:
		return req, fmt.Errorf("unexpected expression type %s", typeExpr.Prim)
	}

	return req, nil
}

// Normalize the type the value (the number or args should be equal on type and value)
func normalizeTypeAndValue(typeExpr michelsonExpression, valueExpr michelsonExpression) (michelsonExpression, michelsonExpression, error) {
	var err error
	if len(typeExpr.Args) < len(valueExpr.Args) {
		// If valueExpr is more optimized than the typeExpr (then optimize typeExpr)
		// In this situations, typeExpr is a pair and not a tuple with more than 2 items
		typeExpr, err = optimizeExpression(typeExpr)
	} else if len(valueExpr.Args) < len(typeExpr.Args) {
		// If typeExpr is more optimized than the valueExpr (then optimize valueExpr)
		// In this situations, valueExpr is a pair and not a tuple with more than 2 items
		valueExpr, err = optimizeExpression(valueExpr)
	}

	if err != nil {
		return typeExpr, valueExpr, err
	}

	return typeExpr, valueExpr, nil
}

// Optimize expression (spread inner pairs to the parent pair)
func optimizeExpression(expr michelsonExpression) (michelsonExpression, error) {
	// [Optimized]			: the shortest representation is used. It depends on the number of arguments:
	//							n < 4 	=> { Pair(a1, a2), Pair(a3, Pair(a4, a5)) }
	//							n >= 4	=> Pair(a1, a2, ... an)
	// [Optimized_legacy]	: nested pairs are always used: { Pair(a1, a2), Pair(a3, Pair(a4, a5)) }
	aux := expr
	aux.Args = []json.RawMessage{}
	for _, arg := range expr.Args {
		innerExpr, err := unmarshalExpression(arg)
		if err != nil {
			return expr, err
		}

		if innerExpr.Prim == primTypePair || innerExpr.Prim == primValuePair {
			aux.Args = append(aux.Args, innerExpr.Args...)
		} else {
			aux.Args = append(aux.Args, arg)
		}
	}

	return aux, nil
}

// Extract key values from map
func extractMapKeyValues(vExpr []json.RawMessage) (map[string]interface{}, error) {
	// The following types are enforced:
	// * [key] => string
	// * [value] => (string|int64)
	values := make(map[string]interface{})

	// For each map entry, get the associated key and value
	for _, expr := range vExpr {
		expression, err := unmarshalExpression(expr)
		if err != nil {
			return nil, err
		}

		// "prim":"Elt" means "element" and every map entry is identified by an element
		if expression.Prim != primValueElt && len(expression.Args) == 2 {
			return values, fmt.Errorf("expected element (Elt) with 2 args, but instead got %v", expression.Prim)
		}

		// Get key expression
		keyExpr, err := unmarshalExpression(expression.Args[0])
		if err != nil {
			return nil, err
		}

		// Get value expression
		valueExpr, err := unmarshalExpression(expression.Args[1])
		if err != nil {
			return nil, err
		}

		// Extract the normalized value from michelson expression
		value, err := extractVariantValue(valueExpr)
		if err != nil {
			return nil, err
		}

		values[keyExpr.String] = value
	}

	return values, nil
}

// Lookup a given annotation and return the respective expression
func extractExpressionByAnnotation(expr michelsonExpression, annot string) (michelsonExpression, error) {
	// INPUTS:
	// -- expr: { prim: "pair", annots: ["%a"], args: [{ prim: "int", annots: ["%b"] }, {}] }
	// -- annot: "%b"
	//
	// OUTPUT: { prim: "int", annots: ["%b"] }

	if len(expr.Annots) > 0 && expr.Annots[0] == annot {
		return expr, nil
	}

	for _, arg := range expr.Args {
		exp, err := unmarshalExpression(arg)
		if err != nil {
			return exp, err
		}

		if exp, err = extractExpressionByAnnotation(exp, annot); err == nil {
			return exp, err
		}
	}

	return expr, fmt.Errorf("expression doesn't contain the annotation %s", annot)
}

// Get value type of big map "requests", the type is used to map values to their respective field names
func getRequestType(typeExpr michelsonExpression) (michelsonExpression, error) {
	// Extract requests big_map type from contract storage type
	typeExpr, err := extractExpressionByAnnotation(typeExpr, "%requests")
	if err != nil {
		return typeExpr, err
	}

	// Only the "%requests" value type is needed
	typeExpr, err = unmarshalExpression(typeExpr.Args[1])
	return typeExpr, err
}

// Get big_map_diff associated with oracle "requests" big_map
func getBigMapDiff(bigMapID string, bigMapDiffs []xtzBigMapDiff) (xtzBigMapDiff, error) {
	for _, diff := range bigMapDiffs {
		if diff.BigMap == bigMapID && (diff.Action == "alloc" || diff.Action == "update") {
			return diff, nil
		}
	}
	return xtzBigMapDiff{}, fmt.Errorf("could not find a big_map_diff update associated with big map %s", bigMapID)
}

// Extract big_map ID, it is used to find big_map diffs
func extractBigMapID(tExpr michelsonExpression, vExpr json.RawMessage) (string, error) {
	storageValues, err := extractValuesFromExpression(tExpr, vExpr)
	if err != nil {
		return "", err
	}

	oracleRequests := storageValues["requests"]
	if oracleRequests == nil {
		return "", fmt.Errorf("the contract storage is not compatible, did not find any (requests) annotation")
	}
	return fmt.Sprintf("%v", oracleRequests), err
}

// Get field from type annotation (["%field"] => "field")
func getAnnotation(annot []string) (string, error) {
	l := len(annot)
	if l != 1 {
		return "", fmt.Errorf("expected one annotation, but received %d", l)
	}
	if !strings.HasPrefix(annot[0], "%") {
		return "", fmt.Errorf(`invalid annotation format (%s), expected prefix (%c)`, annot[0], '%')
	}
	return strings.TrimPrefix(annot[0], "%"), nil
}

// Extract value from michelson expression
func extractVariantValue(expr michelsonExpression) (interface{}, error) {
	if expr.String != "" {
		return expr.String, nil
	} else if expr.Int != "" {
		// Convert string to int64
		return strconv.ParseInt(expr.Int, 10, 64)
	} else if expr.Bytes != "" {
		// In hexadecimal
		return expr.Bytes, nil
	} else if expr.Prim == primValueTrue {
		return true, nil
	} else if expr.Prim == primValueFalse {
		return false, nil
	} else if len(expr.Args) > 0 {
		expr, err := unmarshalExpression(expr.Args[0])
		if err != nil {
			return nil, err
		}
		return extractVariantValue(expr)
	}

	return nil, fmt.Errorf("could not get value from variant")
}

// Convert json.RawMessage to michelsonExpression
// valueRaw can be an object or an array
func convertRawToMichelsonExpression(valueRaw json.RawMessage) (michelsonExpression, error) {
	var vExpression michelsonExpression
	err := json.Unmarshal(valueRaw, &vExpression)
	// If we cannot unmarshal to an object, check if we can unmarshal to an array of objects
	// This can happen because of the "unparsing_mode" used in "internal transactions",
	// which currently is (Optimized mode)
	//
	// [Readable]			: comb pairs are always written Pair(a1, a2, ... an)
	// [Optimized]			: the shortest representation is used. It depends on the number of arguments:
	//							n < 4 	=> { Pair(a1, a2), Pair(a3, Pair(a4, a5)) }
	//							n >= 4	=> Pair(a1, a2, ... an)
	// [Optimized_legacy]	: nested pairs are always used: { Pair(a1, a2), Pair(a3, Pair(a4, a5)) }
	if err != nil {
		vExpressions, err := convertToJSONArray(valueRaw)
		if err != nil {
			// This was neither an object or an array of objects, return error
			return vExpression, err
		}

		// vExpr is an array of expressions ([ expression, ... ])
		// It must either be a sequence or an optimized pair
		vExpression = michelsonExpression{
			Args: vExpressions,
		}
	}

	return vExpression, nil
}

// Convert json.RawMessage to []json.RawMessage
func convertToJSONArray(expr json.RawMessage) ([]json.RawMessage, error) {
	var expressions []json.RawMessage
	err := json.Unmarshal(expr, &expressions)
	return expressions, err
}

// Convert json.RawMessage to michelsonExpression
func unmarshalExpression(expr json.RawMessage) (michelsonExpression, error) {
	var expression michelsonExpression
	err := json.Unmarshal(expr, &expression)
	return expression, err
}
