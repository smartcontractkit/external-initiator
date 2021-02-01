package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/keeper/keeper_registry_contract"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"go.uber.org/atomic"
)

const (
	checkUpkeep        = "checkUpkeep"
	performUpkeep      = "performUpkeep"
	executionQueueSize = 10
)

var (
	upkeepRegistryABI = mustGetABI(keeper_registry_contract.KeeperRegistryContractABI)
	performUpkeepHex  = utils.AddHexPrefix(common.Bytes2Hex(upkeepRegistryABI.Methods[performUpkeep].ID))
	refreshInterval   = 5 * time.Second
)

type UpkeepExecuter interface {
	Start() error
	Stop()
}

func NewUpkeepExecuter(dbClient *gorm.DB, clNode chainlink.Node, config store.RuntimeConfig) UpkeepExecuter {
	return upkeepExecuter{
		blockHeight:         atomic.NewUint64(0),
		chainlinkNode:       clNode,
		endpoint:            config.KeeperEthEndpoint,
		registrationManager: NewRegistrationManager(dbClient, uint64(config.KeeperBlockCooldown)),
		executionQueue:      make(chan struct{}, executionQueueSize),
		chDone:              make(chan struct{}),
		chSignalRun:         make(chan struct{}, 1),
	}
}

type upkeepExecuter struct {
	blockHeight         *atomic.Uint64
	chainlinkNode       chainlink.Node
	connectionType      subscriber.Type
	endpoint            string
	ethClient           *ethclient.Client
	registrationManager RegistrationManager

	executionQueue chan struct{}
	chDone         chan struct{}
	chSignalRun    chan struct{}
}

var _ UpkeepExecuter = upkeepExecuter{} // upkeepExecuter satisfies UpkeepExecuter

func (executer upkeepExecuter) Start() error {
	ethClient, err := ethclient.Dial(executer.endpoint)
	if err != nil {
		return err
	}
	executer.ethClient = ethClient

	if strings.HasPrefix(executer.endpoint, "ws") {
		go executer.setRunsOnHeadSubscription()
	} else if strings.HasPrefix(executer.endpoint, "http") {
		go executer.setRunsOnTimer()
	} else {
		return fmt.Errorf("unknown endpoint protocol: %+v", executer.endpoint)
	}

	go executer.run()

	return nil
}

func (executer upkeepExecuter) Stop() {
	close(executer.chDone)
}

func (executer upkeepExecuter) run() {
	for {
		select {
		case <-executer.chDone:
			return
		case <-executer.chSignalRun:
			executer.processActiveRegistrations()
		}
	}
}

func (executer upkeepExecuter) processActiveRegistrations() {
	// TODO - RYAN - this should be batched to avoid congestgion
	logger.Debug("received new block, running checkUpkeep for keeper registrations")

	activeRegistrations, err := executer.registrationManager.Active(executer.blockHeight.Load())
	if err != nil {
		logger.Errorf("unable to load active registrations: %v", err)
		return
	}

	for _, reg := range activeRegistrations {
		executer.concurrentExecute(reg)
	}
}

func (executer upkeepExecuter) concurrentExecute(registration upkeepRegistration) {
	for {
		select {
		case executer.executionQueue <- struct{}{}:
			go executer.execute(registration)
			return
		default:
		}
	}
}

// execute will call checkForUpkeep and, if it succeeds, triger a job on the CL node
func (executer upkeepExecuter) execute(registration upkeepRegistration) {
	// pop queue when done executing
	defer func() {
		<-executer.executionQueue
	}()

	checkPayload, err := upkeepRegistryABI.Pack(checkUpkeep, big.NewInt(registration.UpkeepID), registration.Registry.From)
	if err != nil {
		logger.Error(err)
		return
	}

	msg := ethereum.CallMsg{
		From: utils.ZeroAddress,
		To:   &registration.Registry.Address,
		Gas:  20000000, // TODO - RYAN - https://www.pivotaltracker.com/story/show/176391267
		Data: checkPayload,
	}

	result, err := executer.ethClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		// don't log anything as this is extremely common and would be too noisey
		// error just signifies "inelligible to performUpkeep"
		return
	}

	res, err := upkeepRegistryABI.Unpack(checkUpkeep, result)
	if err != nil {
		logger.Error(err)
		return
	}

	performData, ok := res[0].([]byte)
	if !ok {
		logger.Error("checkupkeep payload not as expected")
		return
	}

	logger.Debugf("Performing upkeep on registry %s, ID %d", registration.Registry.Address.Hex(), registration.UpkeepID)

	performPayload, err := upkeepRegistryABI.Pack(performUpkeep, big.NewInt(registration.UpkeepID), performData)
	if err != nil {
		logger.Error(err)
		return
	}

	performPayloadString := utils.AddHexPrefix(common.Bytes2Hex(performPayload[4:]))

	// TODO - RYAN - need to include gas here
	// https://www.pivotaltracker.com/story/show/176391267
	chainlinkPayloadJSON := map[string]interface{}{
		"format":           "preformatted",
		"address":          registration.Registry.Address.Hex(),
		"functionSelector": performUpkeepHex,
		"result":           performPayloadString,
		"fromAddresses":    []string{registration.Registry.From.Hex()},
	}

	chainlinkPayload, err := json.Marshal(chainlinkPayloadJSON)
	if err != nil {
		logger.Error(err)
		return
	}

	err = executer.chainlinkNode.TriggerJob(registration.Registry.JobID.String(), chainlinkPayload)
	if err != nil {
		logger.Errorf("Unable to trigger job on chainlink node: %v", err)
	}

	executer.registrationManager.SetRanAt(registration, executer.blockHeight.Load())
}

func (executer upkeepExecuter) setRunsOnHeadSubscription() {
	logger.Debug("setting UpkeepExecuter to run on new heads")

	headers := make(chan *types.Header)
	sub, err := executer.ethClient.SubscribeNewHead(context.Background(), headers)
	defer sub.Unsubscribe()
	if err != nil {
		logger.Fatal(err)
	}

	for {
		select {
		case <-executer.chDone:
			return
		case err := <-sub.Err():
			logger.Errorf("error in keeper header subscription: %v", err)
		case head := <-headers:
			executer.blockHeight.Store(uint64(head.Number.Int64()))
			executer.signalRun()
		}
	}
}

func (executer upkeepExecuter) setRunsOnTimer() {
	logger.Debug("setting UpkeepExecuter to run on timer")
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-executer.chDone:
			return
		case <-ticker.C:
			height, err := executer.ethClient.BlockNumber(context.TODO()) // TODO - RYAN
			if err != nil {
				logger.Errorf("unable to determine latest block height: %v", err)
				continue
			}
			if executer.blockHeight.Load() < height {
				executer.blockHeight.Store(height)
				executer.signalRun()
			}
		}
	}
}

func (executer upkeepExecuter) signalRun() {
	// avoid blocking if signal already in buffer
	select {
	case executer.chSignalRun <- struct{}{}:
	default:
	}
}

func mustGetABI(json string) abi.ABI {
	abi, err := abi.JSON(strings.NewReader(json))
	if err != nil {
		panic("could not parse ABI: " + err.Error())
	}
	return abi
}
