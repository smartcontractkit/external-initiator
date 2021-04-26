package services

import (
	"context"
	"sync"

	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/smartcontractkit/chainlink/core/logger"
	"go.uber.org/zap"
)

type Runlog struct {
	blockchain common.RunlogManager

	quitOnce  sync.Once
	ctxCancel context.CancelFunc

	chJobTrigger chan subscriber.Event
	chClose      chan struct{}

	logger *zap.SugaredLogger
}

func NewRunlog(job string, triggerJobRun chan subscriber.Event, blockchainManager common.RunlogManager) (*Runlog, error) {
	ctx, cancel := context.WithCancel(context.Background())
	run := Runlog{
		blockchain:   blockchainManager,
		ctxCancel:    cancel,
		chJobTrigger: triggerJobRun,
		logger:       logger.Default.With("job", job),
		chClose:      make(chan struct{}),
	}
	run.logger.Infof("New Runlog job")

	runlogEvents := make(chan common.RunlogRequest)
	err := run.blockchain.SubscribeEvents(ctx, runlogEvents)
	if err != nil {
		run.Stop()
		return nil, err
	}

	go run.listenForEvents(runlogEvents)

	return nil, nil
}

func (r *Runlog) Stop() {
	r.quitOnce.Do(func() {
		close(r.chClose)
		r.ctxCancel()
		r.blockchain.Stop()
	})
}

func (r *Runlog) listenForEvents(ch <-chan common.RunlogRequest) {
	for {
		select {
		case req := <-ch:
			r.handleRequest(req)
		case <-r.chClose:
			return
		}
	}
}

func (r *Runlog) handleRequest(reqs ...common.RunlogRequest) {
	for _, req := range reqs {
		r.chJobTrigger <- r.blockchain.CreateJobRun(req)
	}
}
