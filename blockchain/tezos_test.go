package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/smartcontractkit/external-initiator/eitest"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Regex of the paths being mocked in tests
func pathContractScript(contract string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("/chains/main/blocks/[A-z0-9]+/context/contracts/%s/script/normalized", contract))
}

func TestCreateTezosSubscriber(t *testing.T) {
	t.Run("creates tezosSubscriber from subscription",
		func(t *testing.T) {
			sub := store.Subscription{
				Endpoint: store.Endpoint{
					Url: "http://example.com/api",
				},
				Tezos: store.TezosSubscription{
					Addresses: []string{"foobar", "baz"},
				},
			}
			tezosSubscriber := createTezosSubscriber(sub)
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
			tezosSubscriber := createTezosSubscriber(sub)
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
	addresses := []string{"KT1TuXZYMD9tf5Bbv99cpZMB9sy1kbcZJKTk", "KT1UVaDQ2LuTatPXdvvZDAawmp7qi4Y45mdt"}
	wd, _ := os.Getwd()

	ui := path.Join(wd, "testdata/tezos/block_operations_without_internal_calls.json")
	userInitiatedSampleFile, err := os.Open(ui)
	require.NoError(t, err)
	userInitiatedSampleJSON, err := ioutil.ReadAll(userInitiatedSampleFile)
	require.NoError(t, err)
	defer eitest.MustClose(userInitiatedSampleFile)

	sci := path.Join(wd, "testdata/tezos/block_operations_with_internal_calls.json")
	scInitiatedSampleFile, err := os.Open(sci)
	require.NoError(t, err)
	scInitiatedSampleJSON, err := ioutil.ReadAll(scInitiatedSampleFile)
	require.NoError(t, err)
	defer eitest.MustClose(scInitiatedSampleFile)

	scriptNormalized := path.Join(wd, "testdata/tezos/contract_script_normalized.json")
	scriptNormalizedFile, err := os.Open(scriptNormalized)
	require.NoError(t, err)
	scriptNormalizedJSON, err := ioutil.ReadAll(scriptNormalizedFile)
	require.NoError(t, err)
	defer eitest.MustClose(scriptNormalizedFile)

	scriptNormalized2 := path.Join(wd, "testdata/tezos/contract_script_normalized2.json")
	scriptNormalizedFile2, err := os.Open(scriptNormalized2)
	require.NoError(t, err)
	scriptNormalizedJSON2, err := ioutil.ReadAll(scriptNormalizedFile2)
	require.NoError(t, err)
	defer eitest.MustClose(scriptNormalizedFile2)

	requestHandlers := []RequestHandler{
		{
			RegPath:  pathContractScript(addresses[0]),
			Response: scriptNormalizedJSON,
		},
		{
			RegPath:  pathContractScript(addresses[1]),
			Response: scriptNormalizedJSON2,
		},
	}
	mockServer := httptest.NewServer(mockRequestHandler(requestHandlers))
	defer mockServer.Close()

	tzs := tezosSubscription{
		endpoint: mockServer.URL,
	}

	t.Run("returns error if json is invalid",
		func(t *testing.T) {
			json := []byte("{")
			_, err := tzs.extractEventsFromBlock(json, addresses, "test123")
			assert.NotNil(t, err)
		})
	t.Run("returns error if json is in unexpected shape",
		func(t *testing.T) {
			json := []byte("[[]]")
			_, err := tzs.extractEventsFromBlock(json, addresses, "test123")
			assert.NotNil(t, err)

		})
	t.Run("returns no events if the address doesn't match",
		func(t *testing.T) {
			json := userInitiatedSampleJSON
			events, err := tzs.extractEventsFromBlock(json, []string{"notAnAddress"}, "test123")
			assert.Nil(t, err)
			assert.Len(t, events, 0)
		})
	t.Run("extract request parameters from internal call (on_token_tranfer|create_request)",
		func(t *testing.T) {
			js := scInitiatedSampleJSON
			events, err := tzs.extractEventsFromBlock(js, addresses, "mock")
			require.NoError(t, err)
			require.Len(t, events, 2)
			assert.IsType(t, []subscriber.Event{}, events)

			assert.Equal(t, "XTZ", gjson.GetBytes(events[0], "parameters.from").Str)
			assert.Equal(t, "EUR", gjson.GetBytes(events[0], "parameters.to").Str)
			assert.Equal(t, "a43617fdec5ed080d581dc46be5a929167d44f69bc39252ee618768a2733969f", gjson.GetBytes(events[0], requestFieldID).Str)

			assert.Equal(t, "EUR", gjson.GetBytes(events[1], "parameters.from").Str)
			assert.Equal(t, "XTZ", gjson.GetBytes(events[1], "parameters.to").Str)
			assert.Equal(t, "f6ab2c039df8049391a4c09effa97861d8de2400aea24c6d1e55961e389fd726", gjson.GetBytes(events[1], requestFieldID).Str)
		})
}

func Test_getRequestType(t *testing.T) {
	wd, _ := os.Getwd()

	scriptNormalized := path.Join(wd, "testdata/tezos/requests_type_normalized.json")
	scriptNormalizedFile, err := os.Open(scriptNormalized)
	require.NoError(t, err)
	scriptNormalizedJSON, err := ioutil.ReadAll(scriptNormalizedFile)
	require.NoError(t, err)
	defer eitest.MustClose(scriptNormalizedFile)

	requestTypeNormalized := path.Join(wd, "testdata/tezos/request_type_normalized.json")
	requestTypeNormalizedFile, err := os.Open(requestTypeNormalized)
	require.NoError(t, err)
	requestTypeNormalizedJSON, err := ioutil.ReadAll(requestTypeNormalizedFile)
	require.NoError(t, err)
	defer eitest.MustClose(requestTypeNormalizedFile)

	t.Run("Valid",
		func(t *testing.T) {
			expr, err := unmarshalExpression(json.RawMessage(scriptNormalizedJSON))
			require.NoError(t, err)
			types, err := getRequestType(expr)
			require.NoError(t, err)
			require.JSONEq(t, string(requestTypeNormalizedJSON), prettifyJSON(types))
		})
	t.Run("Invalid",
		func(t *testing.T) {
			expr, err := unmarshalExpression(json.RawMessage(requestTypeNormalizedJSON))
			require.NoError(t, err)
			_, err = getRequestType(expr)
			require.Error(t, err)
		})
}

func Test_extractMapKeyValues(t *testing.T) {
	wd, _ := os.Getwd()

	t.Run("Valid", func(t *testing.T) {
		validMapValue := path.Join(wd, "testdata/tezos/map_value.json")
		validMapValueFile, err := os.Open(validMapValue)
		require.NoError(t, err)
		validMapValueJSON, err := ioutil.ReadAll(validMapValueFile)
		require.NoError(t, err)
		defer eitest.MustClose(validMapValueFile)
		expression, err := convertToJSONArray(json.RawMessage(validMapValueJSON))
		require.NoError(t, err)

		keyValues, err := extractMapKeyValues(expression)
		assert.NoError(t, err)
		assert.Equal(t, "XTZ", keyValues["to"])
		assert.Equal(t, "EUR", keyValues["from"])
	})

	t.Run("Invalid", func(t *testing.T) {
		invalidMapValue := path.Join(wd, "testdata/tezos/invalid_map_value.json")
		invalidMapValueFile, err := os.Open(invalidMapValue)
		require.NoError(t, err)
		invalidMapValueJSON, err := ioutil.ReadAll(invalidMapValueFile)
		require.NoError(t, err)
		defer eitest.MustClose(invalidMapValueFile)
		expression, err := convertToJSONArray(json.RawMessage(invalidMapValueJSON))
		require.NoError(t, err)

		keyValues, err := extractMapKeyValues(expression)
		assert.Error(t, err)
		assert.Equal(t, 0, len(keyValues))
	})
}

func Test_getAnnotation(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		annotation, err := getAnnotation([]string{"%valid_annotation"})
		assert.NoError(t, err)
		assert.EqualValues(t, "valid_annotation", annotation)

		annotation, err = getAnnotation([]string{"%_valid_annotation"})
		assert.NoError(t, err)
		assert.EqualValues(t, "_valid_annotation", annotation)

		annotation, err = getAnnotation([]string{"%a"})
		assert.NoError(t, err)
		assert.Equal(t, "a", annotation)
	})
	t.Run("Invalid", func(t *testing.T) {
		annotation, err := getAnnotation([]string{""})
		assert.Error(t, err)
		assert.EqualValues(t, "", annotation)

		annotation, err = getAnnotation([]string{"@invalid_annotation"})
		assert.Error(t, err)
		assert.Equal(t, "", annotation)
	})
}

func Test_extractVariantValue(t *testing.T) {
	t.Run("Bool Variants", func(t *testing.T) {
		expression := michelsonExpression{
			Prim: "True",
		}
		value, err := extractVariantValue(expression)
		assert.NoError(t, err)
		assert.Equal(t, true, value)

		expression = michelsonExpression{
			Prim: "False",
		}
		value, err = extractVariantValue(expression)
		assert.NoError(t, err)
		assert.Equal(t, false, value)
	})
	t.Run("String Variant", func(t *testing.T) {
		expression := michelsonExpression{
			Prim: "Right",
			Args: []json.RawMessage{
				json.RawMessage(`
					{
						"prim": "Right",
						"args": [
							{
								"string": "XTZ"
							}
						]
					}
				`),
			},
		}
		value, err := extractVariantValue(expression)
		assert.NoError(t, err)
		assert.Equal(t, "XTZ", value)
	})
	t.Run("Int Variant", func(t *testing.T) {
		expression := michelsonExpression{
			Prim: "Right",
			Args: []json.RawMessage{
				json.RawMessage(`{ "int": "10" }`),
			},
		}
		value, err := extractVariantValue(expression)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), value)
	})
	t.Run("Bytes Variant", func(t *testing.T) {
		expression := michelsonExpression{
			Prim: "Right",
			Args: []json.RawMessage{
				json.RawMessage(`
					{
						"prim": "Right",
						"args": [
							{
								"prim": "Right",
								"args": [
									{
										"bytes": "74657a6f73"
									}
								]
							}
						]
					}
				`),
			},
		}
		value, err := extractVariantValue(expression)
		assert.NoError(t, err)
		assert.Equal(t, "74657a6f73", value)
	})
	t.Run("Invalid", func(t *testing.T) {
		expression := michelsonExpression{}
		value, err := extractVariantValue(expression)
		assert.Error(t, err)
		assert.Equal(t, nil, value)
	})
}

// ---------
//  Helpers
// ---------
type RequestHandler struct {
	RegPath  *regexp.Regexp
	Response []byte
}

func mockRequestHandler(rh []RequestHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, req := range rh {
			if req.RegPath.MatchString(r.URL.String()) {
				w.Write(req.Response)
				return
			}
		}
	})
}

func prettifyJSON(o interface{}) string {
	prettyJSON, _ := json.MarshalIndent(o, "", "  ")
	return string(prettyJSON)
}
