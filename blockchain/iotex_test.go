package blockchain

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/facebookgo/clock"
	"github.com/golang/mock/gomock"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotexapi/mock_iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/smartcontractkit/external-initiator/eitest"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestCreateIoTeXLogFilter(t *testing.T) {
	tests := []struct {
		jobID     string
		addresses []string
	}{
		{
			jobID:     "",
			addresses: []string{},
		},
		{
			jobID:     "468bba3012fb4e43b5399f0d55f1a18e",
			addresses: []string{"io1r63nasmzw72xvn23em5365azy5vu5jn5rea903"},
		},
		{
			jobID: "468bba3012fb4e43b5399f0d55f1a18e",
			addresses: []string{
				"io1r63nasmzw72xvn23em5365azy5vu5jn5rea903",
				"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw",
			},
		},
	}

	for i, tt := range tests {
		filter := createIoTeXLogFilter(tt.jobID, tt.addresses)
		require.NotNilf(t, filter, "test #%d", i)
		assert.EqualValuesf(t, tt.addresses, filter.GetAddress(), "test #%d", i)
		assert.NotZerof(t, len(filter.GetTopics()), "test #%d", i)
	}
}

func TestCreateIoTeXSubscriber(t *testing.T) {
	sub := store.Subscription{
		Job: "468bba3012fb4e43b5399f0d55f1a18e",
		Endpoint: store.Endpoint{
			Url: "http://example.com",
		},
		Ethereum: store.EthSubscription{Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
	}
	s, err := createIoTeXSubscriber(sub)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NotNil(t, s.conn)
	assert.Equal(t, sub.Endpoint.Url, "http://"+s.conn.endpoint)
	require.NotNil(t, s.filter)
	assert.EqualValues(t, sub.Ethereum.Addresses, s.filter.GetAddress())
	assert.NotZero(t, len(s.filter.GetTopics()))
}

func TestIoTeXSubscriberTest(t *testing.T) {
	serv, cancel := newIoTeXMockServer(t)
	defer cancel()

	s := store.Subscription{
		Job: "468bba3012fb4e43b5399f0d55f1a18e",
		Endpoint: store.Endpoint{
			Url: iotexMockServerEndpoint(),
		},
		Ethereum: store.EthSubscription{Addresses: []string{"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw"}},
	}
	sub, err := createIoTeXSubscriber(s)
	require.NoError(t, err)

	serv.EXPECT().
		GetChainMeta(gomock.Any(), gomock.AssignableToTypeOf(&iotexapi.GetChainMetaRequest{})).
		Return(&iotexapi.GetChainMetaResponse{ChainMeta: &iotextypes.ChainMeta{Height: 10000}}, nil).Times(1)
	assert.NoError(t, sub.Test())
}

func TestIoTeXLogEventToSubscriberEvents(t *testing.T) {
	t.Run("zero event", func(t *testing.T) {
		out, err := iotexLogEventToSubscriberEvents(nil)
		require.Empty(t, out)
		require.Nil(t, err)
	})
	t.Run("one event", testIoTeXLogEventToSubscriberEventsOneEvent)
}

func testIoTeXLogEventToSubscriberEventsOneEvent(t *testing.T) {
	tests := []struct {
		contract string
		data     string
		isError  bool
		out      string
	}{
		{
			isError: true,
		},
		{
			contract: "io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw",
			data:     "0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
			isError:  false,
			out:      `{"address":"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`,
		},
		{
			contract: "io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw",
			data:     "0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
			isError:  false,
			out:      `{"address":"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`,
		},
	}

	for i, tt := range tests {
		in := []*iotextypes.Log{
			&iotextypes.Log{
				ContractAddress: tt.contract,
				Data:            common.Hex2Bytes(tt.data),
			},
		}
		out, err := iotexLogEventToSubscriberEvents(in)
		if tt.isError {
			require.Errorf(t, err, "test #%d", i)
			continue
		}
		require.NoErrorf(t, err, "test #%d", i)
		require.NotZerof(t, len(out), "test#%d", i)
		require.EqualValuesf(t, tt.out, string(out[0]), "test #%d", i)
	}
}

func TestIoTeXSubscriptionPoll(t *testing.T) {
	serv, cancel := newIoTeXMockServer(t)
	defer cancel()
	ctx, ctxcancel := context.WithCancel(context.Background())

	channel := make(chan subscriber.Event)
	s := store.Subscription{
		Job: "468bba3012fb4e43b5399f0d55f1a18e",
		Endpoint: store.Endpoint{
			Url: iotexMockServerEndpoint(),
		},
		Ethereum: store.EthSubscription{Addresses: []string{"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw"}},
	}
	suber, err := createIoTeXSubscriber(s)
	require.NoError(t, err)
	sub := suber.newSubscription(channel, ctxcancel, clock.New())

	// 1st poll, expect to poll 1 block data, return 1 event log
	serv.EXPECT().
		GetChainMeta(gomock.Any(), gomock.AssignableToTypeOf(&iotexapi.GetChainMetaRequest{})).
		Return(&iotexapi.GetChainMetaResponse{ChainMeta: &iotextypes.ChainMeta{Height: 10000}}, nil).Times(1)
	serv.EXPECT().GetLogs(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req *iotexapi.GetLogsRequest) (*iotexapi.GetLogsResponse, error) {
		assert.EqualValues(t, sub.filter.GetAddress(), req.GetFilter().GetAddress())
		assert.EqualValues(t, len(sub.filter.GetTopics()), len(req.GetFilter().GetTopics()))
		assert.Equal(t, uint64(10000), req.GetByRange().GetFromBlock())
		assert.Equal(t, uint64(1), req.GetByRange().GetCount())
		return &iotexapi.GetLogsResponse{Logs: []*iotextypes.Log{
			&iotextypes.Log{
				ContractAddress: "io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw",
				Data:            common.Hex2Bytes("0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864"),
			},
		}}, nil
	}).Times(1)
	sub.poll(ctx)
	event, ok := <-channel
	assert.True(t, ok)
	assert.Equal(t, `{"address":"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`, string(event))
	assert.Equal(t, uint64(10000), sub.requestedHeight)

	// 2nd poll, reutrn same block height, expect to poll no data
	serv.EXPECT().
		GetChainMeta(gomock.Any(), gomock.AssignableToTypeOf(&iotexapi.GetChainMetaRequest{})).
		Return(&iotexapi.GetChainMetaResponse{ChainMeta: &iotextypes.ChainMeta{Height: 10000}}, nil).Times(1)
	sub.poll(ctx)
	select {
	case <-channel:
		assert.Fail(t, "channel has event")
	default:
	}

	// 3rd poll, reutrn same block height, expect to poll 5 blocks data, return 0 event logs
	serv.EXPECT().
		GetChainMeta(gomock.Any(), gomock.AssignableToTypeOf(&iotexapi.GetChainMetaRequest{})).
		Return(&iotexapi.GetChainMetaResponse{ChainMeta: &iotextypes.ChainMeta{Height: 10005}}, nil).Times(1)

	serv.EXPECT().GetLogs(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req *iotexapi.GetLogsRequest) (*iotexapi.GetLogsResponse, error) {
		assert.EqualValues(t, sub.filter.GetAddress(), req.GetFilter().GetAddress())
		assert.EqualValues(t, len(sub.filter.GetTopics()), len(req.GetFilter().GetTopics()))
		assert.Equal(t, uint64(10001), req.GetByRange().GetFromBlock())
		assert.Equal(t, uint64(5), req.GetByRange().GetCount())
		return &iotexapi.GetLogsResponse{Logs: []*iotextypes.Log{}}, nil
	}).Times(1)

	sub.poll(ctx)
	select {
	case <-channel:
		assert.Fail(t, "channel has event")
	default:
	}

}

func TestIoTeXSubscriptionRun(t *testing.T) {
	serv, cancel := newIoTeXMockServer(t)
	defer cancel()
	ctx, ctxcancel := context.WithCancel(context.Background())
	ck := clock.NewMock()

	channel := make(chan subscriber.Event)
	sub := &iotexSubscription{
		conn:         &iotexConnection{endpoint: iotexMockServerHost()},
		interval:     iotexScanInterval,
		cancel:       ctxcancel,
		eventChannel: channel,
		filter:       createIoTeXLogFilter("468bba3012fb4e43b5399f0d55f1a18e", []string{"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw"}),
		clock:        ck,
	}
	sub.run(ctx)
	select {
	case <-channel:
		assert.Fail(t, "channel has event")
	default:
	}

	serv.EXPECT().
		GetChainMeta(gomock.Any(), gomock.AssignableToTypeOf(&iotexapi.GetChainMetaRequest{})).
		Return(&iotexapi.GetChainMetaResponse{ChainMeta: &iotextypes.ChainMeta{Height: 10000}}, nil).Times(1)
	serv.EXPECT().GetLogs(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req *iotexapi.GetLogsRequest) (*iotexapi.GetLogsResponse, error) {
		return &iotexapi.GetLogsResponse{Logs: []*iotextypes.Log{
			&iotextypes.Log{
				ContractAddress: "io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw",
				Data:            common.Hex2Bytes("0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864"),
			},
		}}, nil
	}).Times(1)

	ck.Add(6 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			assert.Fail(t, "channel has not event")
			return
		case <-channel:
			return
		}
	}
}

func TestIoTeXSubscriptionUnSubscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ck := clock.NewMock()

	channel := make(chan subscriber.Event)
	sub := &iotexSubscription{
		interval:     iotexScanInterval,
		cancel:       cancel,
		eventChannel: channel,
		filter:       createIoTeXLogFilter("468bba3012fb4e43b5399f0d55f1a18e", []string{"io1uzfy7aa920thkm7tqdf73sexcljzkhqv55kpyw"}),
		clock:        ck,
	}
	sub.run(ctx)
	sub.Unsubscribe()
	ck.Add(6 * time.Second)
	_, ok := <-ctx.Done()
	assert.False(t, ok)
}

func iotexMockServerEndpoint() string { return "http://" + iotexMockServerHost() }
func iotexMockServerHost() string     { return "localhost:14016" }
func newIoTeXMockServer(t *testing.T) (*mock_iotexapi.MockAPIServiceServer, context.CancelFunc) {
	serv := mock_iotexapi.NewMockAPIServiceServer(gomock.NewController(t))
	server := grpc.NewServer()
	iotexapi.RegisterAPIServiceServer(server, serv)
	listener, err := net.Listen("tcp", iotexMockServerHost())
	require.NoError(t, err)
	go func() {
		eitest.Must(server.Serve(listener))
	}()
	return serv, func() {
		server.Stop()
	}
}
