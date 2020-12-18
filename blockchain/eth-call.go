package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/status-im/keycard-go/hexutils"
)

const (
	ETH_CALL           = "eth-call"
	defaultResponseKey = "result"
)

var emptyFunctionSelector = [4]byte{0, 0, 0, 0}

type solFunctionHelper struct {
	abi              abi.ABI
	functionSelector models.FunctionSelector
	methodName       string
}

func NewSolFunctionHelper(abiJson json.RawMessage, methodName string, funcSelector models.FunctionSelector, returnType string) (*solFunctionHelper, error) {
	if (len(abiJson) == 0 || methodName == "") && (len(funcSelector.Bytes()) == 0 || returnType == "") {
		return nil, errors.New("missing ABI & methodName or functionSelector & returnType")
	}

	helper := solFunctionHelper{}

	var err error
	// If ABI is included, we set up the helper based on ABI + methodName
	if len(abiJson) > 0 {
		helper.abi, err = abi.JSON(bytes.NewReader(abiJson))
		if err != nil {
			return nil, err
		}
		helper.methodName = methodName

		var funcSelectorBytes []byte
		funcSelectorBytes, err = helper.abi.Pack(methodName)
		if err != nil {
			return nil, err
		}

		helper.functionSelector = models.BytesToFunctionSelector(funcSelectorBytes)

		return &helper, nil
	}

	// With no ABI included, we set up the helper based on funcSelector + returnType
	t, err := abi.NewType(returnType, "", nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	helper.functionSelector = funcSelector
	helper.methodName = "arbitrary"
	helper.abi.Methods = map[string]abi.Method{
		helper.methodName: {
			Name:    helper.methodName,
			RawName: helper.methodName,
			Type:    abi.Function,
			Outputs: abi.Arguments{
				{
					Type: t,
				},
			},
		},
	}

	return &helper, nil
}

func (helper solFunctionHelper) FunctionSelector() string {
	if helper.functionSelector == emptyFunctionSelector {
		return "0x"
	}
	return helper.functionSelector.String()
}

func (helper solFunctionHelper) Unpack(str string) ([]interface{}, error) {
	// Remove 0x prefix
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}
	data := hexutils.HexToBytes(str)
	return helper.abi.Unpack(helper.methodName, data)
}

type ethCallSubscriber struct {
	Endpoint       string
	Address        common.Address
	FunctionHelper solFunctionHelper
	JobID          string
	ResponseKey    string
	Connection     subscriber.Type
	Interval       time.Duration
}

func createEthCallSubscriber(sub store.Subscription) (*ethCallSubscriber, error) {
	abiBytes := sub.EthCall.ABI
	// Add a check to convert stringified JSON to JSON object
	var s string
	if json.Unmarshal(abiBytes, &s) == nil {
		abiBytes = []byte(s)
	}

	helper, err := NewSolFunctionHelper([]byte(abiBytes), sub.EthCall.MethodName, sub.EthCall.FunctionSelector, sub.EthCall.ReturnType)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating solidity function helper")
	}

	var t subscriber.Type
	if strings.HasPrefix(sub.Endpoint.Url, "ws") {
		t = subscriber.WS
	} else if strings.HasPrefix(sub.Endpoint.Url, "http") {
		t = subscriber.RPC
	} else {
		return nil, fmt.Errorf("unknown endpoint protocol: %+v", sub.Endpoint.Url)
	}

	return &ethCallSubscriber{
		Endpoint:       strings.TrimSuffix(sub.Endpoint.Url, "/"),
		Address:        common.HexToAddress(sub.EthCall.Address),
		FunctionHelper: *helper,
		JobID:          sub.Job,
		ResponseKey:    sub.EthCall.ResponseKey,
		Connection:     t,
		Interval:       time.Duration(sub.Endpoint.RefreshInt) * time.Second,
	}, nil
}

type ethCallSubscription struct {
	endpoint string
	events   chan<- subscriber.Event
	address  common.Address
	helper   solFunctionHelper
	isDone   bool
	jobID    string
	key      string
}

func (ethCall ethCallSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ ...interface{}) (subscriber.ISubscription, error) {
	sub := ethCallSubscription{
		endpoint: ethCall.Endpoint,
		events:   channel,
		jobID:    ethCall.JobID,
		address:  ethCall.Address,
		helper:   ethCall.FunctionHelper,
		key:      ethCall.ResponseKey,
	}

	switch ethCall.Connection {
	case subscriber.RPC:
		go sub.queryUntilDone(ethCall.Interval)
	case subscriber.WS:
		sub.subscribeToNewHeads()
	}

	return sub, nil
}

func (ethCall ethCallSubscriber) Test() error {
	switch ethCall.Connection {
	case subscriber.RPC:
		return ethCall.TestRPC()
	case subscriber.WS:
		return ethCall.TestWS()
	}
	return errors.New("unknown connection type")
}

func (ethCall ethCallSubscriber) TestRPC() error {
	resp, err := sendEthCallPost(ethCall.Endpoint, ethCall.GetTestJson())
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (ethCall ethCallSubscriber) TestWS() error {
	c, _, err := websocket.DefaultDialer.Dial(ethCall.Endpoint, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	resp := make(chan []byte)

	go func() {
		var body []byte
		_, body, err = c.ReadMessage()
		if err != nil {
			close(resp)
		}
		resp <- body
	}()

	err = c.WriteMessage(websocket.BinaryMessage, ethCall.GetTestJson())
	if err != nil {
		return err
	}

	// Set timeout for response to 5 seconds
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()

	select {
	case <-t.C:
		return errors.New("timeout from test payload")
	case body, ok := <-resp:
		if !ok {
			return errors.New("failed reading test response from WS endpoint")
		}
		return ethCall.ParseTestResponse(body)
	}
}

func (ethCall ethCallSubscriber) GetTestJson() []byte {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "eth_blockNumber",
	}
	payload, _ := json.Marshal(msg)
	return payload
}

func (ethCall ethCallSubscriber) ParseTestResponse(resp []byte) error {
	if len(resp) == 0 {
		return errors.New("unexpected empty response")
	}

	return nil
}

type ethCallMessage struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to"`
	Gas      string `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Value    string `json:"value,omitempty"`
	Data     string `json:"data,omitempty"`
}

func (ethCall ethCallSubscription) getCallPayload() ([]byte, error) {
	call := ethCallMessage{
		To:   ethCall.address.Hex(),
		Data: ethCall.helper.FunctionSelector(),
	}

	var params []interface{}
	params = append(params, call)
	params = append(params, "latest")
	paramsBz, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "eth_call",
		Params:  paramsBz,
	}
	return json.Marshal(msg)
}

func (ethCall ethCallSubscription) getSubscribePayload() ([]byte, error) {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "eth_subscribe",
		Params:  json.RawMessage(`["newHeads"]`),
	}
	return json.Marshal(msg)
}

func (ethCall ethCallSubscription) queryUntilDone(interval time.Duration) {
	for {
		if ethCall.isDone {
			return
		}
		ethCall.query()
		time.Sleep(interval)
	}
}

func (ethCall ethCallSubscription) query() {
	payload, err := ethCall.getCallPayload()
	if err != nil {
		logger.Error("Unable to get ETH QAE payload:", err)
		return
	}

	resp, err := sendEthCallPost(ethCall.endpoint, payload)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return
	}

	var response JsonrpcMessage
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Error(err)
		return
	}

	events, err := ethCall.parseResponse(response)
	if err != nil {
		logger.Error("failed parseResponse:", err)
		return
	}

	for _, event := range events {
		ethCall.events <- event
	}
}

func sendEthCallPost(endpoint string, payload []byte) (*http.Response, error) {
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("%s returned 400. This endpoint may not support calls to /monitor", endpoint)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %v from endpoint %s", resp.StatusCode, endpoint)
	}
	return resp, nil
}

func (ethCall ethCallSubscription) messageReader(conn *websocket.Conn, callPayload []byte) {
	defer func() {
		ethCall.isDone = true
		_ = conn.Close()
		logger.Debug("closing WS subscription")
	}()

	requestedEthCall := false

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			logger.Error("failed reading messages:", err)
			return
		}

		var msg JsonrpcMessage
		err = json.Unmarshal(rawMsg, &msg)
		if err != nil {
			logger.Error("error unmarshalling WS message:", err)
			continue
		}

		if msg.Method == "eth_subscription" {
			err = conn.WriteMessage(websocket.TextMessage, callPayload)
			if err != nil {
				logger.Error("failed writing to WS connection:", err)
				return
			}
			requestedEthCall = true
		} else if requestedEthCall {
			events, err := ethCall.parseResponse(msg)
			if err != nil {
				logger.Error("Failed parsing response:", err)
				continue
			}
			for _, event := range events {
				ethCall.events <- event
			}
			requestedEthCall = false
		}
	}
}

func (ethCall ethCallSubscription) subscribeToNewHeads() {
	logger.Infof("Connecting to WS endpoint: %s", ethCall.endpoint)

	callPayload, err := ethCall.getCallPayload()
	if err != nil {
		logger.Error(err)
		return
	}

	subscribePayload, err := ethCall.getSubscribePayload()
	if err != nil {
		logger.Error(err)
		return
	}

	c, _, err := websocket.DefaultDialer.Dial(ethCall.endpoint, nil)
	if err != nil {
		logger.Error(err)
		ethCall.isDone = true
		return
	}

	go ethCall.messageReader(c, callPayload)

	err = c.WriteMessage(websocket.TextMessage, subscribePayload)
	if err != nil {
		logger.Error(err)
		ethCall.isDone = true
		c.Close()
		return
	}

	logger.Infof("Connected to %s", ethCall.endpoint)
}

func (ethCall ethCallSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from ETH QAE endpoint", ethCall.endpoint)
	ethCall.isDone = true
}

func (ethCall ethCallSubscription) parseResponse(response JsonrpcMessage) ([]subscriber.Event, error) {
	var data string
	err := json.Unmarshal(response.Result, &data)
	if err != nil {
		return nil, err
	}

	res, err := ethCall.helper.Unpack(data)
	if err != nil {
		return nil, err
	}

	// If there is no data returned, we have no jobs to initiate
	if res == nil || len(res) == 0 {
		return nil, nil
	}

	result := res[0]
	var events []subscriber.Event
	switch result.(type) {
	// Add cases for special types
	case bool:
		if result.(bool) == true {
			events = append(events, subscriber.Event{})
		}
	// For any other types, figure out if we have an array or a single value
	default:
		slice, err := interfaceToSlice(result)
		// bytes32 is an array, but we don't want to initiate a job run for each byte
		_, isbytes := result.([32]byte)
		if err == nil && !isbytes {
			for _, r := range slice {
				event := map[string]interface{}{
					ethCall.key: r,
				}
				bz, err := json.Marshal(event)
				if err != nil {
					return nil, err
				}
				events = append(events, bz)
			}
		} else {
			// If we have bytes32, we want the raw bytes32 instead of decoding it
			if isbytes {
				result = fmt.Sprintf("%s", data)
			}
			event := map[string]interface{}{
				ethCall.key: result,
			}
			bz, err := json.Marshal(event)
			if err != nil {
				return nil, err
			}
			events = append(events, bz)
		}
	}

	return events, nil
}

func interfaceToSlice(data interface{}) ([]interface{}, error) {
	bz, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var slice []interface{}
	return slice, json.Unmarshal(bz, &slice)
}
