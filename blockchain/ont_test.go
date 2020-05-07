package blockchain

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology-go-sdk/common"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateOntSubscriber(t *testing.T) {
	t.Run("creates ontSubscriber from subscription",
		func(t *testing.T) {
			sub := store.Subscription{
				Job: "test",
				Ontology: store.OntSubscription{
					Addresses: []string{"foobar", "baz"},
				},
			}
			ontSubscriber := createOntSubscriber(sub)
			assert.Equal(t, "test", ontSubscriber.JobId)
			assert.Equal(t, []string{"foobar", "baz"}, ontSubscriber.Addresses)
		})
}

func TestNotifyTrigger(t *testing.T) {
	tests := []struct {
		name     string
		args     *common.NotifyEventInfo
		wantPass bool
		want     []byte
	}{
		{
			"standard notify",
			&common.NotifyEventInfo{
				ContractAddress: "b54dd842fadc8b04f0c58b1ea921f49bf54d04f0",
				States: []interface{}{
					"6f7261636c6552657175657374",
					"9b17104a15424c2288cd01d7016a8b7e",
					"d5be50a52d69a5b4596883b93edd1ff0f191540a",
					"e7d5067e33586630c4402c962b1bafab13e857073798486b909fbe15444e33a7",
					"000064a7b3b6e00d",
					"60e60caf46a2acd97ca338e5450988b9e146aadb",
					"66756c66696c6c",
					"fe94a75e",
					"01",
					"63676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
					"6f7261636c6552657175657374",
				},
			},
			true,
			[]byte(`{"address":"b54dd842fadc8b04f0c58b1ea921f49bf54d04f0","callbackAddress":"60e60caf46a2acd97ca338e5450988b9e146aadb","callbackFunction":"66756c66696c6c","expiration":"fe94a75e","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","payment":"000064a7b3b6e00d","requestID":"e7d5067e33586630c4402c962b1bafab13e857073798486b909fbe15444e33a7","times":100}`),
		},
		{
			"notify lack of some fields",
			&common.NotifyEventInfo{
				ContractAddress: "b54dd842fadc8b04f0c58b1ea921f49bf54d04f0",
				States: []interface{}{
					"6f7261636c6552657175657374",
					"9b17104a15424c2288cd01d7016a8b7e",
					"d5be50a52d69a5b4596883b93edd1ff0f191540a",
					"e7d5067e33586630c4402c962b1bafab13e857073798486b909fbe15444e33a7",
					"000064a7b3b6e00d",
					"60e60caf46a2acd97ca338e5450988b9e146aadb",
					"66756c66696c6c",
					"fe94a75e",
					"63676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
				},
			},
			false,
			nil,
		},
		{
			"notify that contract address doesn't match",
			&common.NotifyEventInfo{
				ContractAddress: "004dd842fadc8b04f0c58b1ea921f49bf54d04f0",
				States: []interface{}{
					"6f7261636c6552657175657374",
					"9b17104a15424c2288cd01d7016a8b7e",
					"d5be50a52d69a5b4596883b93edd1ff0f191540a",
					"e7d5067e33586630c4402c962b1bafab13e857073798486b909fbe15444e33a7",
					"000064a7b3b6e00d",
					"60e60caf46a2acd97ca338e5450988b9e146aadb",
					"66756c66696c6c",
					"fe94a75e",
					"01",
					"63676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
					"6f7261636c6552657175657374",
				},
			},
			false,
			nil,
		},
		{
			"notify that job id doesn't match",
			&common.NotifyEventInfo{
				ContractAddress: "b54dd842fadc8b04f0c58b1ea921f49bf54d04f0",
				States: []interface{}{
					"6f7261636c6552657175657374",
					"0017104a15424c2288cd01d7016a8b7e",
					"d5be50a52d69a5b4596883b93edd1ff0f191540a",
					"e7d5067e33586630c4402c962b1bafab13e857073798486b909fbe15444e33a7",
					"000064a7b3b6e00d",
					"60e60caf46a2acd97ca338e5450988b9e146aadb",
					"66756c66696c6c",
					"fe94a75e",
					"01",
					"63676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
					"6f7261636c6552657175657374",
				},
			},
			false,
			nil,
		},
		{
			"notify that name doesn't match",
			&common.NotifyEventInfo{
				ContractAddress: "b54dd842fadc8b04f0c58b1ea921f49bf54d04f0",
				States: []interface{}{
					"0x1",
					"9b17104a15424c2288cd01d7016a8b7e",
					"d5be50a52d69a5b4596883b93edd1ff0f191540a",
					"e7d5067e33586630c4402c962b1bafab13e857073798486b909fbe15444e33a7",
					"000064a7b3b6e00d",
					"60e60caf46a2acd97ca338e5450988b9e146aadb",
					"66756c66696c6c",
					"fe94a75e",
					"01",
					"63676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864",
					"6f7261636c6552657175657374",
				},
			},
			false,
			nil,
		},
		{
			"notify that contains illegal CBOR field",
			&common.NotifyEventInfo{
				ContractAddress: "b54dd842fadc8b04f0c58b1ea921f49bf54d04f0",
				States: []interface{}{
					"6f7261636c6552657175657374",
					"9b17104a15424c2288cd01d7016a8b7e",
					"01",
					"02",
					"03",
					"04",
					"05",
					"06",
					"07",
					"08",
					"09",
				},
			},
			false,
			nil,
		},
		{
			"notify that have legal necessary fields",
			&common.NotifyEventInfo{
				ContractAddress: "b54dd842fadc8b04f0c58b1ea921f49bf54d04f0",
				States: []interface{}{
					"6f7261636c6552657175657374",
					"9b17104a15424c2288cd01d7016a8b7e",
					"01",
					"02",
					"03",
					"04",
					"05",
					"06",
					"07",
					"",
					"09",
				},
			},
			true,
			[]byte(`{"address":"b54dd842fadc8b04f0c58b1ea921f49bf54d04f0","callbackAddress":"04","callbackFunction":"05","expiration":"06","payment":"03","requestID":"02"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ots := &ontSubscription{
				addresses: map[string]bool{"b54dd842fadc8b04f0c58b1ea921f49bf54d04f0": true},
				jobId:     "9b17104a15424c2288cd01d7016a8b7e",
			}
			js, ok := ots.notifyTrigger(tt.args)
			if ok != tt.wantPass {
				t.Errorf("notifyTrigger pass = %v, wantPass %v", ok, tt.wantPass)
			}
			if !bytes.Equal(js, tt.want) {
				t.Errorf("ParseResponse() got = %s, want %s", js, tt.want)
			}
		})
	}
}
