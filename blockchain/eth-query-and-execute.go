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
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/status-im/keycard-go/hexutils"
)

const (
	ETH_QAE            = "eth-query-and-execute"
	defaultResponseKey = "value"
)

type ethQaeSubscriber struct {
	Endpoint    string
	Address     common.Address
	ABI         abi.ABI
	MethodName  string
	Job         string
	ResponseKey string
	Connection  subscriber.Type
}

func createEthQaeSubscriber(sub store.Subscription) (ethQaeSubscriber, error) {
	abiBytes := sub.EthQae.ABI
	// Add a check to convert stringified JSON to JSON object
	var s string
	if json.Unmarshal(abiBytes, &s) == nil {
		abiBytes = []byte(s)
	}

	contractAbi, err := abi.JSON(bytes.NewReader(abiBytes))
	if err != nil {
		return ethQaeSubscriber{}, err
	}

	var t subscriber.Type
	if strings.HasPrefix(sub.Endpoint.Url, "ws") {
		t = subscriber.WS
	} else if strings.HasPrefix(sub.Endpoint.Url, "http") {
		t = subscriber.RPC
	} else {
		return ethQaeSubscriber{}, fmt.Errorf("unknown endpoint protocol: %+v", sub.Endpoint.Url)
	}

	return ethQaeSubscriber{
		Endpoint:    strings.TrimSuffix(sub.Endpoint.Url, "/"),
		Address:     common.HexToAddress(sub.EthQae.Address),
		ABI:         contractAbi,
		Job:         sub.Job,
		ResponseKey: sub.EthQae.ResponseKey,
		MethodName:  sub.EthQae.MethodName,
		Connection:  t,
	}, nil
}

type ethQaeSubscription struct {
	endpoint string
	events   chan<- subscriber.Event
	address  common.Address
	abi      abi.ABI
	method   string
	isDone   bool
	jobid    string
	key      string
}

func (ethQae ethQaeSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ ...interface{}) (subscriber.ISubscription, error) {
	sub := ethQaeSubscription{
		endpoint: ethQae.Endpoint,
		events:   channel,
		jobid:    ethQae.Job,
		address:  ethQae.Address,
		abi:      ethQae.ABI,
		method:   ethQae.MethodName,
		key:      ethQae.ResponseKey,
	}

	switch ethQae.Connection {
	case subscriber.RPC:
		go sub.readMessagesWithRetry()
	case subscriber.WS:
		sub.subscribeToNewHeads()
	}

	return sub, nil
}

func (ethQae ethQaeSubscriber) Test() error {
	switch ethQae.Connection {
	case subscriber.RPC:
		return ethQae.TestRPC()
	case subscriber.WS:
		return ethQae.TestWS()
	}
	return errors.New("unknown connection type")
}

func (ethQae ethQaeSubscriber) TestRPC() error {
	resp, err := sendEthQaePost(ethQae.Endpoint, ethQae.GetTestJson())
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (ethQae ethQaeSubscriber) TestWS() error {
	c, _, err := websocket.DefaultDialer.Dial(ethQae.Endpoint, nil)
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

	err = c.WriteMessage(websocket.BinaryMessage, ethQae.GetTestJson())
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
		return ethQae.ParseTestResponse(body)
	}
}

func (ethQae ethQaeSubscriber) GetTestJson() []byte {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "eth_blockNumber",
	}
	payload, _ := json.Marshal(msg)
	return payload
}

func (ethQae ethQaeSubscriber) ParseTestResponse(resp []byte) error {
	if len(resp) == 0 {
		return errors.New("unexpected empty response")
	}

	return nil
}

type ethCall struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to"`
	Gas      string `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Value    string `json:"value,omitempty"`
	Data     string `json:"data,omitempty"`
}

func (ethQae ethQaeSubscription) getCallPayload() ([]byte, error) {
	data, err := ethQae.abi.Pack(ethQae.method)
	if err != nil {
		return nil, err
	}

	call := ethCall{
		To:   ethQae.address.Hex(),
		Data: hexutil.Encode(data[:]),
	}

	callBz, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}

	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "eth_call",
		Params:  json.RawMessage(`[` + string(callBz) + `,"latest"]`),
	}
	return json.Marshal(msg)
}

func (ethQae ethQaeSubscription) getSubscribePayload() ([]byte, error) {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "eth_subscribe",
		Params:  json.RawMessage(`["newHeads"]`),
	}
	return json.Marshal(msg)
}

func (ethQae ethQaeSubscription) readMessagesWithRetry() {
	for {
		if ethQae.isDone {
			return
		}
		ethQae.readMessages()
		time.Sleep(monitorRetryInterval)
	}
}

func (ethQae ethQaeSubscription) readMessages() {
	payload, err := ethQae.getCallPayload()
	if err != nil {
		logger.Error("Unable to get ETH QAE payload:", err)
		return
	}

	resp, err := sendEthQaePost(ethQae.endpoint, payload)
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

	ethQae.parseResponse(response)
}

func sendEthQaePost(endpoint string, payload []byte) (*http.Response, error) {
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
		return nil, fmt.Errorf("Unexpected status code %v from endpoint %s", resp.StatusCode, endpoint)
	}
	return resp, nil
}

func (ethQae ethQaeSubscription) messageReader(conn *websocket.Conn, callPayload []byte) {
	defer func() {
		ethQae.isDone = true
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
			ethQae.parseResponse(msg)
			requestedEthCall = false
		}
	}
}

func (ethQae ethQaeSubscription) subscribeToNewHeads() {
	logger.Infof("Connecting to WS endpoint: %s", ethQae.endpoint)

	callPayload, err := ethQae.getCallPayload()
	if err != nil {
		logger.Error(err)
		return
	}

	subscribePayload, err := ethQae.getSubscribePayload()
	if err != nil {
		logger.Error(err)
		return
	}

	c, _, err := websocket.DefaultDialer.Dial(ethQae.endpoint, nil)
	if err != nil {
		logger.Error(err)
		ethQae.isDone = true
		return
	}

	go ethQae.messageReader(c, callPayload)

	logger.Debug(string(subscribePayload))

	err = c.WriteMessage(websocket.TextMessage, subscribePayload)
	if err != nil {
		logger.Error(err)
		ethQae.isDone = true
		c.Close()
		return
	}

	logger.Infof("Connected to %s", ethQae.endpoint)
}

func (ethQae ethQaeSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from ETH QAE endpoint", ethQae.endpoint)
	ethQae.isDone = true
}

func (ethQae ethQaeSubscription) parseResponse(response JsonrpcMessage) {
	var result string
	err := json.Unmarshal(response.Result, &result)
	if err != nil {
		logger.Error(err)
		return
	}

	// Remove 0x prefix
	if strings.HasPrefix(result, "0x") {
		result = result[2:]
	}

	resultData := hexutils.HexToBytes(result)
	b, err := unpackResultIntoBool(ethQae.abi, ethQae.method, resultData)
	if err == nil {
		if *b == true {
			ethQae.events <- subscriber.Event{}
		}
		return
	}

	res, err := unpackResultIntoAddresses(ethQae.abi, ethQae.method, resultData)
	if err != nil {
		logger.Error(err)
		return
	}

	for _, r := range *res {
		event := map[string]interface{}{
			ethQae.key: r,
		}
		bz, err := json.Marshal(event)
		if err != nil {
			logger.Error(err)
			continue
		}
		ethQae.events <- bz
	}
}
