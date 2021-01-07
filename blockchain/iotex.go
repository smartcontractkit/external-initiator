package blockchain

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"

	"github.com/facebookgo/clock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

const (
	IOTX              = "iotex"
	iotexScanInterval = 5 * time.Second
)

type iotexConnection struct {
	sync.RWMutex

	endpoint   string
	secureConn bool
	api        iotexapi.APIServiceClient
	grpcConn   *grpc.ClientConn
}

func (io *iotexConnection) connect() error {
	io.Lock()
	defer io.Unlock()
	// Check if the existing connection is good.
	if io.grpcConn != nil && io.grpcConn.GetState() != connectivity.Shutdown {
		return nil
	}
	var opts []grpc.DialOption
	if io.secureConn {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	grpcConn, err := grpc.Dial(io.endpoint, opts...)
	if err != nil {
		return err
	}
	io.grpcConn = grpcConn
	io.api = iotexapi.NewAPIServiceClient(io.grpcConn)
	return err
}

func createIoTeXSubscriber(sub store.Subscription) (*iotexSubscriber, error) {
	u, err := url.Parse(sub.Endpoint.Url)
	if err != nil {
		return nil, err
	}

	return &iotexSubscriber{
		conn: &iotexConnection{
			endpoint:   u.Host,
			secureConn: u.Scheme == "https",
		},
		filter: createIoTeXLogFilter(sub.Job, sub.Ethereum.Addresses),
	}, nil
}

type iotexSubscriber struct {
	conn   *iotexConnection
	filter *iotexapi.LogsFilter
}

func (io *iotexSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ store.RuntimeConfig) (subscriber.ISubscription, error) {
	ctx, cancel := context.WithCancel(context.Background())
	sub := io.newSubscription(channel, cancel, clock.New())
	sub.run(ctx)
	return sub, nil
}

func (io *iotexSubscriber) newSubscription(channel chan<- subscriber.Event, cancel context.CancelFunc, clk clock.Clock) *iotexSubscription {
	return &iotexSubscription{
		conn:         io.conn,
		interval:     iotexScanInterval,
		cancel:       cancel,
		eventChannel: channel,
		filter:       io.filter,
		clock:        clk,
	}
}

func (io *iotexSubscriber) Test() error {
	if err := io.conn.connect(); err != nil {
		return err
	}
	_, err := io.conn.api.GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
	return err
}

type iotexSubscription struct {
	conn         *iotexConnection
	interval     time.Duration
	cancel       context.CancelFunc
	eventChannel chan<- subscriber.Event
	filter       *iotexapi.LogsFilter
	clock        clock.Clock

	ticker          *clock.Ticker
	requestedHeight uint64
}

func (io *iotexSubscription) run(ctx context.Context) {
	io.ticker = io.clock.Ticker(io.interval)
	go func() {
		for {
			select {
			case <-io.ticker.C:
				io.poll(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (io *iotexSubscription) poll(ctx context.Context) {
	if err := io.conn.connect(); err != nil {
		logger.Error("failed to connect to iotex server:", err)
		return
	}
	cm, err := io.conn.api.GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
	if err != nil {
		logger.Error("failed to get iotex chain meta:", err)
		return
	}
	currentHeight := cm.GetChainMeta().GetHeight()
	fromHeight := currentHeight
	count := uint64(1)
	blocks := currentHeight - io.requestedHeight
	if blocks == 0 {
		return
	} else if blocks < 1000 {
		count = blocks
		fromHeight = io.requestedHeight + 1
	}
	req := &iotexapi.GetLogsRequest{
		Filter: io.filter,
		Lookup: &iotexapi.GetLogsRequest_ByRange{
			ByRange: &iotexapi.GetLogsByRange{
				FromBlock: fromHeight,
				Count:     count,
			},
		},
	}
	resp, err := io.conn.api.GetLogs(ctx, req)
	if err != nil {
		logger.Error("failed to get iotex event logs:", err)
		return
	}

	// set sub height
	io.requestedHeight = currentHeight

	events, err := iotexLogEventToSubscriberEvents(resp.GetLogs())
	if err != nil {
		logger.Error("failed to convert iotex event logs to subscriber events:", err)
		return
	}

	go func() {
		for _, event := range events {
			io.eventChannel <- event
		}
	}()
}

func createIoTeXLogFilter(jobid string, addresses []string) *iotexapi.LogsFilter {
	topic := StringToBytes32(jobid)
	return &iotexapi.LogsFilter{
		Address: addresses,
		Topics: []*iotexapi.Topics{
			{
				Topic: [][]byte{
					models.RunLogTopic20190207withoutIndexes[:],
					topic[:],
				},
			},
		},
	}
}

func iotexLogEventToSubscriberEvents(logs []*iotextypes.Log) ([]subscriber.Event, error) {
	events := make([]subscriber.Event, 0, len(logs))
	for _, log := range logs {
		cborData, dataPrefixBytes, err := logDataParse(log.GetData())
		if err != nil {
			return nil, err
		}
		js, err := models.ParseCBOR(cborData)
		if err != nil {
			return nil, fmt.Errorf("error parsing CBOR: %v", err)
		}
		js, err = js.MultiAdd(models.KV{
			"address":          log.GetContractAddress(),
			"dataPrefix":       bytesToHex(dataPrefixBytes),
			"functionSelector": models.OracleFulfillmentFunctionID20190128withoutCast,
		})
		if err != nil {
			return nil, fmt.Errorf("error add json fields: %v", err)
		}
		event, err := json.Marshal(js)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (io *iotexSubscription) Unsubscribe() {
	if io.ticker != nil {
		io.ticker.Stop()
	}
	io.cancel()
}
