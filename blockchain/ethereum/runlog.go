package ethereum

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	. "github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/blockchain/evm"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

const RpcRequestTimeout = 5 * time.Second

type runlogManager struct {
	*manager
}

func CreateRunlogManager(sub store.Subscription) (*runlogManager, error) {
	manager, err := createManager(sub)
	if err != nil {
		return nil, err
	}

	return &runlogManager{
		manager: manager,
	}, nil
}

func (rm runlogManager) SubscribeEvents(ctx context.Context, ch chan<- RunlogRequest) error {
	switch rm.subscriber.Type() {
	case subscriber.RPC:
		return rm.getEventsRPC(ctx, ch)
	case subscriber.WS:
		return rm.getEventsWS(ctx, ch)
	}

	return errors.New("got unexpected subscriber type")
}

func (rm runlogManager) getEventsRPC(ctx context.Context, ch chan<- RunlogRequest) error {
	newBlocks := make(chan uint64)
	err := rm.subscribeNewBlocks(ctx, newBlocks)
	if err != nil {
		return err
	}

	var startingBlockNum uint64
	select {
	case startingBlockNum = <-newBlocks:
	case <-ctx.Done():
		return nil
	}

	requests, err := rm.getRecentEventsRPC(ctx, startingBlockNum)
	if err != nil {
		return err
	}

	go func() {
		// We expect ch to be blocked until the job has been created,
		// so we wait with writing until we're in a new goroutine
		for _, request := range requests {
			ch <- request
		}

		for {
			select {
			case block := <-newBlocks:
				requests, err := rm.getRecentEventsRPC(ctx, startingBlockNum)
				if err != nil {
					logger.Error(err)
					continue
				}
				// If the request was successful, update the last
				// block number we should query from
				startingBlockNum = block + 1
				for _, request := range requests {
					ch <- request
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (rm runlogManager) getRecentEventsRPC(ctx context.Context, fromBlock uint64) ([]RunlogRequest, error) {
	method := "eth_getLogs"
	fq, err := rm.getFilterQuery(fmt.Sprintf("%d", fromBlock))
	if err != nil {
		return nil, err
	}

	params, err := json.Marshal([]interface{}{fq})
	if err != nil {
		return nil, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, RpcRequestTimeout)
	defer cancel()

	resp, err := rm.subscriber.Request(ctxWithTimeout, method, params)
	if err != nil {
		return nil, err
	}

	return parseEthLogsResponse(resp)
}

func parseEthLogsResponse(result json.RawMessage) ([]RunlogRequest, error) {
	var events []models.Log
	if err := json.Unmarshal(result, &events); err != nil {
		return nil, err
	}

	var requests []RunlogRequest
	for _, evt := range events {
		if evt.Removed {
			continue
		}

		request, err := evm.LogEventToOracleRequest(evt)
		if err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	return requests, nil
}

func parseEthLogResponse(result json.RawMessage) (RunlogRequest, error) {
	var event models.Log
	if err := json.Unmarshal(result, &event); err != nil {
		return nil, err
	}

	if event.Removed {
		return nil, errors.New("event was removed")
	}

	return evm.LogEventToOracleRequest(event)
}

func (rm runlogManager) subscribeNewBlocks(ctx context.Context, ch chan<- uint64) error {
	listener := make(chan json.RawMessage)
	err := rm.subscriber.Subscribe(ctx, "eth_blockNumber", "", nil, listener)
	if err != nil {
		return err
	}

	var latestBlockNumber uint64

	go func() {
		for {
			select {
			case resp := <-listener:
				num, err := parseBlockNumberResult(resp)
				if err != nil {
					logger.Error(err)
					continue
				}

				if num <= latestBlockNumber {
					continue
				}

				// Update the recorded block height
				// and notify channel about new height
				latestBlockNumber = num
				ch <- num
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func parseBlockNumberResult(data []byte) (uint64, error) {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(str, 0, 64)
}

func (rm runlogManager) getFilterQuery(fromBlock string) (interface{}, error) {
	fq := *rm.fq

	if fromBlock != "" {
		fq.FromBlock = fromBlock
	} else {
		fq.FromBlock = "latest"
	}

	return fq.ToMapInterface()
}

func (rm runlogManager) getEventsWS(ctx context.Context, ch chan<- RunlogRequest) error {
	if rm.fq.FromBlock == "" {
		rm.fq.FromBlock = "latest"
	}

	filter, err := rm.fq.ToMapInterface()
	if err != nil {
		return err
	}

	params, err := json.Marshal([]interface{}{"logs", filter})
	if err != nil {
		return err
	}

	responses := make(chan json.RawMessage)
	err = rm.subscriber.Subscribe(ctx, "eth_subscribe", "eth_unsubscribe", params, responses)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case resp := <-responses:
				request, err := parseEthLogResponse(resp)
				if err != nil {
					logger.Error(err)
					continue
				}
				ch <- request
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (rm runlogManager) CreateJobRun(request RunlogRequest) map[string]interface{} {
	// This implementation does not need to make any changes
	// to the request payload.
	return request
}
