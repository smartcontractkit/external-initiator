package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
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
)

const (
	Keeper        = "keeper"
	checkMethod   = "checkForUpkeep"
	executeMethod = "performUpkeep"
)

const UpkeepRegistryInterface = `[
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "upkeepId",
				"type": "uint256"
			}
		],
		"name": "checkForUpkeep",
		"outputs": [
			{
				"internalType": "bool",
				"name": "canPerform",
				"type": "bool"
			},
			{
				"internalType": "bytes",
				"name": "performData",
				"type": "bytes"
			},
			{
				"internalType": "uint256",
				"name": "maxLinkPayment",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "gasLimit",
				"type": "uint256"
			},
			{
				"internalType": "int256",
				"name": "gasWei",
				"type": "int256"
			},
			{
				"internalType": "int256",
				"name": "linkEth",
				"type": "int256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "id",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "performData",
				"type": "bytes"
			}
		],
		"name": "performUpkeep",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

type keeperSubscriber struct {
	Endpoint   string
	Address    common.Address
	Abi        abi.ABI
	UpkeepID   *big.Int
	JobID      string
	Connection subscriber.Type
	Interval   time.Duration
}

func createKeeperSubscriber(sub store.Subscription) (*keeperSubscriber, error) {
	abiBytes := []byte(UpkeepRegistryInterface)
	contractAbi, err := abi.JSON(bytes.NewReader(abiBytes))
	if err != nil {
		return nil, err
	}

	upkeepId := new(big.Int)
	_, err = fmt.Sscan(sub.Keeper.UpkeepID, upkeepId)
	if err != nil {
		return nil, err
	}

	var t subscriber.Type
	if strings.HasPrefix(sub.Endpoint.Url, "ws") {
		t = subscriber.WS
	} else if strings.HasPrefix(sub.Endpoint.Url, "http") {
		t = subscriber.RPC
	} else {
		return nil, fmt.Errorf("unknown endpoint protocol: %+v", sub.Endpoint.Url)
	}

	return &keeperSubscriber{
		Endpoint:   strings.TrimSuffix(sub.Endpoint.Url, "/"),
		Address:    common.HexToAddress(sub.Keeper.Address),
		Abi:        contractAbi,
		UpkeepID:   upkeepId,
		JobID:      sub.Job,
		Connection: t,
		Interval:   time.Duration(sub.Endpoint.RefreshInt) * time.Second,
	}, nil
}

type keeperSubscription struct {
	endpoint         string
	events           chan<- subscriber.Event
	address          common.Address
	abi              abi.ABI
	upkeepId         *big.Int
	isDone           bool
	jobID            string
	cooldown         *big.Int
	lastInitiatedRun *big.Int
	blockHeight      *big.Int
}

func (keeper keeperSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, runtimeConfig store.RuntimeConfig) (subscriber.ISubscription, error) {
	sub := keeperSubscription{
		endpoint:         keeper.Endpoint,
		events:           channel,
		jobID:            keeper.JobID,
		address:          keeper.Address,
		abi:              keeper.Abi,
		upkeepId:         keeper.UpkeepID,
		cooldown:         big.NewInt(runtimeConfig.KeeperBlockCooldown),
		lastInitiatedRun: big.NewInt(0),
		blockHeight:      big.NewInt(0),
	}

	switch keeper.Connection {
	case subscriber.RPC:
		go sub.queryUntilDone(keeper.Interval)
	case subscriber.WS:
		go sub.subscribeToNewHeadsWithRetry()
	default:
		return nil, ErrConnectionType
	}

	return &sub, nil
}

func (keeper keeperSubscriber) Test() error {
	switch keeper.Connection {
	case subscriber.RPC:
		return keeper.TestRPC()
	case subscriber.WS:
		return keeper.TestWS()
	default:
		return ErrConnectionType
	}
}

func (keeper keeperSubscriber) TestRPC() error {
	resp, err := sendEthNodePost(keeper.Endpoint, keeper.GetTestJson())
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (keeper keeperSubscriber) TestWS() error {
	c, _, err := websocket.DefaultDialer.Dial(keeper.Endpoint, nil)
	if err != nil {
		return err
	}
	defer logger.ErrorIfCalling(c.Close)

	resp := make(chan []byte)

	go func() {
		var body []byte
		_, body, err = c.ReadMessage()
		if err != nil {
			close(resp)
		}
		resp <- body
	}()

	err = c.WriteMessage(websocket.BinaryMessage, keeper.GetTestJson())
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
		return keeper.ParseTestResponse(body)
	}
}

func (keeper keeperSubscriber) GetTestJson() []byte {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "eth_blockNumber",
	}
	payload, _ := json.Marshal(msg)
	return payload
}

func (keeper keeperSubscriber) ParseTestResponse(resp []byte) error {
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

func (keeper keeperSubscription) getCallPayload() ([]byte, error) {
	data, err := keeper.abi.Pack(checkMethod, keeper.upkeepId)
	if err != nil {
		return nil, err
	}

	call := ethCallMessage{
		To:   keeper.address.Hex(),
		Data: bytesToHex(data),
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

func (keeper keeperSubscription) getSubscribePayload() ([]byte, error) {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "eth_subscribe",
		Params:  json.RawMessage(`["newHeads"]`),
	}
	return json.Marshal(msg)
}

func (keeper keeperSubscription) queryUntilDone(interval time.Duration) {
	for {
		if keeper.isDone {
			return
		}
		keeper.query()
		time.Sleep(interval)
	}
}

func (keeper keeperSubscription) getBlockHeightPost() (*big.Int, error) {
	payload, err := GetBlockNumberPayload()
	if err != nil {
		return nil, err
	}

	resp, err := sendEthNodePost(keeper.endpoint, payload)
	if err != nil {
		return nil, err
	}
	defer logger.ErrorIfCalling(resp.Body.Close)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response JsonrpcMessage
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	var blockNum string
	err = json.Unmarshal(response.Result, &blockNum)
	if err != nil {
		return nil, err
	}

	return hexutil.DecodeBig(blockNum)
}

func (keeper *keeperSubscription) updateLastInitiatedRun() {
	derefHeight := *keeper.blockHeight
	keeper.lastInitiatedRun = &derefHeight
}

func (keeper *keeperSubscription) query() {
	blockHeight, err := keeper.getBlockHeightPost()
	if err != nil {
		logger.Error("Unable to get the current block height:", err)
		return
	}
	if blockHeight.Cmp(keeper.blockHeight) < 1 {
		// No new blocks...
		return
	}

	logger.Debugw("Keeper subscription got new block header", "blockHeight", blockHeight.String())
	keeper.blockHeight = blockHeight

	if !keeper.isCooldownDone() {
		return
	}

	payload, err := keeper.getCallPayload()
	if err != nil {
		logger.Error("Unable to get Keeper ETH payload:", err)
		return
	}

	resp, err := sendEthNodePost(keeper.endpoint, payload)
	if err != nil {
		logger.Error(err)
		return
	}
	defer logger.ErrorIfCalling(resp.Body.Close)

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

	events, err := keeper.parseResponse(response)
	if err != nil {
		logger.Error("failed parseResponse:", err)
		return
	}

	if len(events) > 0 {
		keeper.updateLastInitiatedRun()
	}

	for _, event := range events {
		keeper.events <- event
	}
}

func (keeper keeperSubscription) isCooldownDone() bool {
	difference := &big.Int{}
	if keeper.cooldown.Cmp(difference.Sub(keeper.blockHeight, keeper.lastInitiatedRun)) > 0 {
		logger.Debugw("initiated a run too recently, waiting...",
			"cooldown", keeper.cooldown.String(),
			"lastInitiatedRun", keeper.lastInitiatedRun.String(),
			"blockHeight", keeper.blockHeight.String())
		return false
	}
	return true
}

func (keeper *keeperSubscription) handleWsSubscriptionMessage(msg JsonrpcMessage) (bool, error) {
	blockNum, err := ParseBlocknumberFromNewHeads(msg)
	if err != nil {
		return false, err
	}
	logger.Debugw("Keeper subscription got new block header", "blockHeight", blockNum.String())
	keeper.blockHeight = blockNum

	if !keeper.isCooldownDone() {
		return false, err
	}

	return true, nil
}

func (keeper *keeperSubscription) handleWsMessage(msg JsonrpcMessage) error {
	events, err := keeper.parseResponse(msg)
	if err != nil {
		return err
	}
	if len(events) > 0 {
		keeper.updateLastInitiatedRun()
	}
	for _, event := range events {
		keeper.events <- event
	}

	return nil
}

func (keeper keeperSubscription) subscribeToNewHeads() {
	logger.Infof("Connecting to Keeper WS endpoint: %s", keeper.endpoint)

	callPayload, err := keeper.getCallPayload()
	if err != nil {
		logger.Error(err)
		return
	}

	subscribePayload, err := keeper.getSubscribePayload()
	if err != nil {
		logger.Error(err)
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial(keeper.endpoint, nil)
	if err != nil {
		logger.Error(err)
		return
	}
	defer func() {
		logger.Infof("Disconnecting from Keeper WS endpoint: %s", keeper.endpoint)
		logger.ErrorIf(conn.Close())
	}()

	logger.Infof("Connected to Keeper WS endpoint: %s", keeper.endpoint)

	err = conn.WriteMessage(websocket.TextMessage, subscribePayload)
	if err != nil {
		logger.Error(err)
		return
	}

	first := true

	for {
		if keeper.isDone {
			return
		}

		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			logger.Error(errors.Wrap(err, "failed reading messages"))
			return
		}

		var msg JsonrpcMessage
		err = json.Unmarshal(rawMsg, &msg)
		if err != nil {
			logger.Error("error unmarshalling Keeper WS message:", err)
			continue
		}

		// The first message will be the subscription ID.
		// Since we don't anticipate any other subscriptions,
		// we just skip this ID.
		if first {
			first = false
			continue
		}

		shouldRequestEthCall := false
		switch msg.Method {
		case "eth_subscription":
			shouldRequestEthCall, err = keeper.handleWsSubscriptionMessage(msg)
		default:
			err = keeper.handleWsMessage(msg)
		}
		if err != nil {
			logger.Error(err)
			continue
		}

		if shouldRequestEthCall {
			err = conn.WriteMessage(websocket.TextMessage, callPayload)
			if err != nil {
				logger.Error("failed writing to WS connection:", err)
				return
			}
		}
	}
}

func (keeper keeperSubscription) subscribeToNewHeadsWithRetry() {
	for {
		if keeper.isDone {
			return
		}

		keeper.subscribeToNewHeads()
		if !keeper.isDone {
			logger.Debugf("Waiting 5s to reconnect to Keeper WS endpoint")
			time.Sleep(5 * time.Second)
		}
	}
}

func (keeper *keeperSubscription) Unsubscribe() {
	logger.Info("Stopping Keeper subscription on endpoint", keeper.endpoint)
	keeper.isDone = true
}

func (keeper keeperSubscription) parseResponse(response JsonrpcMessage) ([]subscriber.Event, error) {
	var data string
	err := json.Unmarshal(response.Result, &data)
	if err != nil {
		return nil, err
	}

	dataNoPrefix := strings.TrimPrefix(data, "0x")
	encb, err := hex.DecodeString(dataNoPrefix)
	if err != nil {
		return nil, err
	}

	res, err := keeper.abi.Unpack(checkMethod, encb)
	if err != nil {
		return nil, err
	}

	// If there is no data returned, we have no jobs to initiate
	if len(res) == 0 {
		return nil, errors.New("ethCall returned no results")
	}

	canPerform, ok := res[0].(bool)
	if !ok {
		return nil, errors.New("unable to determine if canPerform == true")
	}

	if !canPerform {
		return nil, nil
	}

	executeData, err := keeper.abi.Pack(executeMethod, keeper.upkeepId, res[1].([]byte))
	if err != nil {
		return nil, err
	}

	event := map[string]interface{}{
		"format":           "preformatted",
		"address":          keeper.address.String(),
		"functionSelector": bytesToHex(executeData[:4]),
		"result":           bytesToHex(executeData[4:]),
	}

	eventBz, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return []subscriber.Event{eventBz}, nil
}
