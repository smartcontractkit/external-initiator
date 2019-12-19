package blockchain

import (
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestCreateTezosSubscriber(t *testing.T) {
	t.Run("creates TezosSubscriber from subscription",
		func(t *testing.T) {
			sub := store.Subscription{
				Endpoint: store.Endpoint{
					Url: "http://example.com/api",
				},
				Tezos: store.TezosSubscription{
					Addresses: []string{"foobar", "baz"},
				},
			}
			tezosSubscriber := CreateTezosSubscriber(sub)
			assert.Equal(t, "http://example.com/api", tezosSubscriber.Endpoint)
			assert.Equal(t, []string{"foobar", "baz"}, tezosSubscriber.Addresses)
		})
	t.Run("trims trailing slash from endpoint",
		func(t *testing.T) {
			sub := store.Subscription{
				Endpoint: store.Endpoint{
					Url: "https://example.com/api/",
				},
			}
			tezosSubscriber := CreateTezosSubscriber(sub)
			assert.Equal(t, "https://example.com/api", tezosSubscriber.Endpoint)
		})
}

func Test_extractBlockIDFromHeaderJSON(t *testing.T) {
	t.Run("extracts block ID from valid header JSON",
		func(t *testing.T) {
			json := []byte(`{"hash":"theBlockID","level":136875,"proto":1,"predecessor":"BLjyuxQa8QGEpXAJ5kdfYuqqL49jRs4bUPDq1Ye2PA27C4zdyGM","timestamp":"2019-12-16T20:55:42Z","validation_pass":4,"operations_hash":"LLoaRmpaxjeV1QsrczSVuLK5ddDfaSZ7xZt1BJMZMzPoS591TsXwu","fitness":["01","00000000000216aa"],"context":"CoUrZrMSmff6NYSSSg9xHqDvwKbCMQMmaVBQ8N7Bc1xXiu9MSh1K","protocol_data":"0000e11a790239180200002143b97eee6f034c1f06e4ddb0833799ad5820da57bfae68987c90e3bd61579e0733173a429c89b7415f11f8822ee715254e23a789c52a858ac52337252eef0f"}`)

			blockID, err := extractBlockIDFromHeaderJSON(json)
			assert.Nil(t, err)
			assert.Equal(t, "theBlockID", blockID)
		})
	t.Run("returns error when header JSON is invalid",
		func(t *testing.T) {
			json := []byte(`{`)

			blockID, err := extractBlockIDFromHeaderJSON(json)
			assert.NotNil(t, err)
			assert.Equal(t, "", blockID)
		})
	t.Run("returns error when header JSON is in an unexpected format",
		func(t *testing.T) {
			json := []byte(`{"foo":42}`)

			blockID, err := extractBlockIDFromHeaderJSON(json)
			assert.NotNil(t, err)
			assert.Equal(t, "", blockID)
		})
}

func Test_extractEventsFromBlock(t *testing.T) {
	addresses := []string{"KT1Address", "KT2Address"}
	wd, _ := os.Getwd()
	ui := path.Join(wd, "testdata/tezos_test_block_operations_user_initiated.json")
	userInitiatedSampleFile, err := os.Open(ui)
	if err != nil {
		require.NoError(t, err)
	}
	defer userInitiatedSampleFile.Close()

	sci := path.Join(wd, "testdata/tezos_test_block_operations_sc_initiated.json")
	scInitiatedSampleFile, err := os.Open(sci)
	if err != nil {
		require.NoError(t, err)
	}
	defer scInitiatedSampleFile.Close()

	userInitiatedSampleJSON, err := ioutil.ReadAll(userInitiatedSampleFile)
	scInitiatedSampleJSON, err := ioutil.ReadAll(scInitiatedSampleFile)

	t.Run("returns error if json is invalid",
		func(t *testing.T) {
			json := []byte("{")

			_, err := extractEventsFromBlock(json, addresses)
			assert.NotNil(t, err)

		})
	t.Run("returns error if json is in unexpected shape",
		func(t *testing.T) {
			json := []byte("[[]]")

			_, err := extractEventsFromBlock(json, addresses)
			assert.NotNil(t, err)

		})
	t.Run("returns no events if the address doesn't match",
		func(t *testing.T) {
			json := userInitiatedSampleJSON

			events, err := extractEventsFromBlock(json, []string{"notAnAddress"})
			assert.Nil(t, err)
			assert.Len(t, events, 0)
		})
	t.Run("extracts user-initiated calls to matching addresses",
		func(t *testing.T) {
			json := userInitiatedSampleJSON

			events, err := extractEventsFromBlock(json, addresses)
			assert.Nil(t, err)
			assert.Len(t, events, 1)
			assert.IsType(t, []subscriber.Event{}, events)

			expectedAmount := "42"
			actualAmount := gjson.GetBytes(events[0], "contents.0.amount").Str

			assert.Equal(t, expectedAmount, actualAmount)
		})
	t.Run("extracts SC-initiated calls to matching addresses",
		func(t *testing.T) {
			json := scInitiatedSampleJSON

			events, err := extractEventsFromBlock(json, addresses)
			assert.Nil(t, err)
			assert.Len(t, events, 1)
			assert.IsType(t, []subscriber.Event{}, events)

			expectedAmount := "1000000"
			actualAmount := gjson.GetBytes(events[0], "contents.0.amount").Str

			assert.Equal(t, expectedAmount, actualAmount)
		})
}

func Test_toEvent(t *testing.T) {
	t.Run("marshals TzTransaction to event JSON",
		func(t *testing.T) {
			expected := []byte(`{"protocol":"foo","chain_id":"foo","hash":"foo","branch":"foo","contents":[]}`)
			transaction := XtzTransaction{
				Protocol: "foo",
				ChainID:  "foo",
				Hash:     "foo",
				Branch:   "foo",
				Contents: []XtzTransactionContent{},
			}
			event, err := transaction.toEvent()
			assert.Nil(t, err)
			assert.Equal(t, subscriber.Event(expected), event)
		})
}
