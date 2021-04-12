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
	blockchain common.Manager

	quitOnce  sync.Once
	ctxCancel context.CancelFunc

	chJobTrigger chan subscriber.Event
	chClose      chan struct{}

	logger *zap.SugaredLogger
}

func NewRunlog(job string, triggerJobRun chan subscriber.Event, blockchainManager common.Manager) (*Runlog, error) {
	ctx, cancel := context.WithCancel(context.Background())
	run := Runlog{
		blockchain:   blockchainManager,
		ctxCancel:    cancel,
		chJobTrigger: triggerJobRun,
		logger:       logger.Default.With("job", job),
	}
	run.logger.Infof("New Runlog job")

	backfilledRequests, err := run.blockchain.Request(common.RunlogBackfill)
	if err != nil {
		run.Stop()
		return nil, err
	}
	go run.handleRequest(backfilledRequests.([]interface{})...)

	runlogEvents := make(chan interface{})
	err = run.blockchain.Subscribe(ctx, common.RunlogSubscribe, runlogEvents)
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
	})
}

func (r *Runlog) listenForEvents(ch <-chan interface{}) {
	for {
		select {
		case req := <-ch:
			r.handleRequest(req)
		case <-r.chClose:
			return
		}
	}
}

func (r *Runlog) handleRequest(reqs ...interface{}) {
	for _, req := range reqs {
		kv, err := r.blockchain.CreateJobRun(common.RunlogJobRun, req)
		if err != nil {
			logger.Error(err)
			continue
		}

		r.chJobTrigger <- kv
	}
}
