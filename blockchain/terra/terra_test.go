package terra

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/tendermint/tendermint/abci/types"
)

func Test_UnmarshalJSONValue(t *testing.T) {
	var i big.Int
	err := json.Unmarshal([]byte("10000"), &i)
	if err != nil {
		t.Errorf("Errored: %s", err.Error())
	} else {
		var expected big.Int
		expected.SetUint64(10000)
		if i.String() != expected.String() {
			t.Errorf("Incorrect value: %s", i.String())
		}
	}
}

func Test_extractEvents(t *testing.T) {
	result, err := extractEvents(testData)
	if err != nil {
		t.Errorf("Failed with %s", err.Error())
	} else {
		count := len(result)
		if count != 8 {
			t.Errorf("Incorrect events count %d", count)
		}
	}
}

func Test_getRequiredAttributes(t *testing.T) {
	event := types.Event{
		Type: "coin_spent",
		Attributes: []types.EventAttribute{
			{
				Key:   []byte("spender"),
				Value: []byte("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"),
				Index: true,
			},
			{
				Key:   []byte("amount"),
				Value: []byte("149032000000uluna"),
				Index: true,
			},
		},
	}
	expected := map[string]string{
		"spender": "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v",
		"amount":  "149032000000uluna",
	}

	attributes, _ := getRequiredAttributes(event, []string{"amount", "spender"})
	if !reflect.DeepEqual(attributes, expected) {
		t.Error("Incorrect attributes extracted")
	}

	event = types.Event{
		Type: "coin_spent",
		Attributes: []types.EventAttribute{
			{
				Key:   []byte("amount"),
				Value: []byte("149032000000uluna"),
				Index: true,
			},
		},
	}

	attributes, err := getRequiredAttributes(event, []string{"amount", "spender"})
	if attributes != nil && err == nil {
		t.Error("Error expected")
	}
}

func Test_parseNewRoundEvent(t *testing.T) {
	event := types.Event{
		Type: "wasm-new_round",
		Attributes: []types.EventAttribute{
			{
				Key:   []byte("round_id"),
				Value: []byte("32"),
				Index: true,
			},
			{
				Key:   []byte("started_by"),
				Value: []byte("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"),
				Index: true,
			},
		},
	}

	expected := EventNewRound{
		RoundId:   32,
		StartedBy: "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v",
		StartedAt: 0,
	}

	res, _ := parseNewRoundEvent(event)
	if !reflect.DeepEqual(*res, expected) {
		t.Error("Incorrect parsing")
	}
}

func Test_parseSubmissionReceivedEvent(t *testing.T) {
	event := types.Event{
		Type: "wasm-submission_received",
		Attributes: []types.EventAttribute{
			{
				Key:   []byte("round_id"),
				Value: []byte("32"),
				Index: true,
			},
			{
				Key:   []byte("oracle"),
				Value: []byte("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"),
				Index: true,
			},
			{
				Key:   []byte("submission"),
				Value: []byte("100000000"),
				Index: true,
			},
		},
	}

	expected := &EventSubmissionReceived{
		RoundId:    32,
		Oracle:     "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v",
		Submission: Value{*big.NewInt(100000000)},
	}

	res, _ := parseSubmissionReceivedEvent(event)
	if !reflect.DeepEqual(res, expected) {
		t.Error("Incorrect parsing")
	}
}

func Test_parseAnswerUpdatedEvent(t *testing.T) {
	event := types.Event{
		Type: "wasm-answer_updated",
		Attributes: []types.EventAttribute{
			{
				Key:   []byte("round_id"),
				Value: []byte("32"),
				Index: true,
			},
			{
				Key:   []byte("current"),
				Value: []byte("100000000"),
				Index: true,
			},
		},
	}

	expected := &EventAnswerUpdated{
		RoundId: 32,
		Value:   Value{*big.NewInt(100000000)},
	}

	res, _ := parseAnswerUpdatedEvent(event)
	if !reflect.DeepEqual(res, expected) {
		t.Error("Incorrect parsing")
	}
}

func Test_parseOraclePermissionsUpdatedEvent(t *testing.T) {
	event := types.Event{
		Type: "wasm-permissions_updated",
		Attributes: []types.EventAttribute{
			{
				Key:   []byte("added"),
				Value: []byte(`["addr1", "addr2"]`),
				Index: true,
			},
			{
				Key:   []byte("removed"),
				Value: []byte(`["addr3", "addr4"]`),
				Index: true,
			},
		},
	}

	expected := []EventOraclePermissionsUpdated{
		{Oracle: "addr1", Bool: true},
		{Oracle: "addr2", Bool: true},
		{Oracle: "addr3", Bool: false},
		{Oracle: "addr4", Bool: false},
	}

	res, _ := parseOraclePermissionsUpdatedEvent(event)
	if !reflect.DeepEqual(res, expected) {
		t.Error("Incorrect parsing")
	}
}

var testData = []byte(`{
       "query": "execute_contract.contract_address='terra183rx7pqzjwj4mj7rxrrgv589zsfl22yeagalc0'",
       "data": {
           "type": "tendermint/event/Tx",
           "value": {
               "TxResult": {
                   "height": "24717",
                   "tx": "CtoBCtcBCiYvdGVycmEud2FzbS52MWJldGExLk1zZ0V4ZWN1dGVDb250cmFjdBKsAQosdGVycmExeDQ2cnFheTRkM2Nzc3E4Z3h4dnF6OHh0Nm53bHo0dGQyMGszOHYSLHRlcnJhMTgzcng3cHF6andqNG1qN3J4cnJndjU4OXpzZmwyMnllYWdhbGMwGk57InNldF92YWxpZGF0b3IiOnsidmFsaWRhdG9yIjoidGVycmExODNyeDdwcXpqd2o0bWo3cnhycmd2NTg5enNmbDIyeWVhZ2FsYzAifX0SbQpOCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohAjszqFJDRAYbEjZMuiD+ChqzbUSGq/RRu3zr0R6iJB5bEgQKAgh/EhsKFQoFdWx1bmESDDE0OTAzMjAwMDAwMBCojAkaQBi3KefjVSPKrlOC+8MNov4QNmv7YmOZFmgppHnCEIGMVbyZlqS+IHlWkC6ee15J75zvp1u76P/BGbNal0w23Fc=",
                   "result": {
                       "data": "CigKJi90ZXJyYS53YXNtLnYxYmV0YTEuTXNnRXhlY3V0ZUNvbnRyYWN0",
                       "log": "[{\"events\":[{\"type\":\"execute_contract\",\"attributes\":[{\"key\":\"sender\",\"value\":\"terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v\"},{\"key\":\"contract_address\",\"value\":\"terra183rx7pqzjwj4mj7rxrrgv589zsfl22yeagalc0\"}]},{\"type\":\"from_contract\",\"attributes\":[{\"key\":\"contract_address\",\"value\":\"terra183rx7pqzjwj4mj7rxrrgv589zsfl22yeagalc0\"},{\"key\":\"action\",\"value\":\"validator_updated\"},{\"key\":\"previous\",\"value\":\"terra1dqrzwx9trx3uhx5k6cm7dxm3dfgmsy58epswlc\"},{\"key\":\"new\",\"value\":\"terra183rx7pqzjwj4mj7rxrrgv589zsfl22yeagalc0\"}]},{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/terra.wasm.v1beta1.MsgExecuteContract\"},{\"key\":\"module\",\"value\":\"wasm\"},{\"key\":\"sender\",\"value\":\"terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v\"}]}]}]",
                       "gas_wanted": "149032",
                       "gas_used": "117586",
                       "events": [
                           {
                               "type": "coin_spent",
                               "attributes": [
                                   {
                                       "key": "c3BlbmRlcg==",
                                       "value": "dGVycmExeDQ2cnFheTRkM2Nzc3E4Z3h4dnF6OHh0Nm53bHo0dGQyMGszOHY=",
                                       "index": true
                                   },
                                   {
                                       "key": "YW1vdW50",
                                       "value": "MTQ5MDMyMDAwMDAwdWx1bmE=",
                                       "index": true
                                   }
                               ]
                           },
                           {
                               "type": "coin_received",
                               "attributes": [
                                   {
                                       "key": "cmVjZWl2ZXI=",
                                       "value": "dGVycmExN3hwZnZha20yYW1nOTYyeWxzNmY4NHoza2VsbDhjNWxrYWVxZmE=",
                                       "index": true
                                   },
                                   {
                                       "key": "YW1vdW50",
                                       "value": "MTQ5MDMyMDAwMDAwdWx1bmE=",
                                       "index": true
                                   }
                               ]
                           },
                           {
                               "type": "transfer",
                               "attributes": [
                                   {
                                       "key": "cmVjaXBpZW50",
                                       "value": "dGVycmExN3hwZnZha20yYW1nOTYyeWxzNmY4NHoza2VsbDhjNWxrYWVxZmE=",
                                       "index": true
                                   },
                                   {
                                       "key": "c2VuZGVy",
                                       "value": "dGVycmExeDQ2cnFheTRkM2Nzc3E4Z3h4dnF6OHh0Nm53bHo0dGQyMGszOHY=",
                                       "index": true
                                   },
                                   {
                                       "key": "YW1vdW50",
                                       "value": "MTQ5MDMyMDAwMDAwdWx1bmE=",
                                       "index": true
                                   }
                               ]
                           },
                           {
                               "type": "message",
                               "attributes": [
                                   {
                                       "key": "c2VuZGVy",
                                       "value": "dGVycmExeDQ2cnFheTRkM2Nzc3E4Z3h4dnF6OHh0Nm53bHo0dGQyMGszOHY=",
                                       "index": true
                                   }
                               ]
                           },
                           {
                               "type": "message",
                               "attributes": [
                                   {
                                       "key": "YWN0aW9u",
                                       "value": "L3RlcnJhLndhc20udjFiZXRhMS5Nc2dFeGVjdXRlQ29udHJhY3Q=",
                                       "index": true
                                   }
                               ]
                           },
                           {
                               "type": "execute_contract",
                               "attributes": [
                                   {
                                       "key": "c2VuZGVy",
                                       "value": "dGVycmExeDQ2cnFheTRkM2Nzc3E4Z3h4dnF6OHh0Nm53bHo0dGQyMGszOHY=",
                                       "index": true
                                   },
                                   {
                                       "key": "Y29udHJhY3RfYWRkcmVzcw==",
                                       "value": "dGVycmExODNyeDdwcXpqd2o0bWo3cnhycmd2NTg5enNmbDIyeWVhZ2FsYzA=",
                                       "index": true
                                   }
                               ]
                           },
                           {
                               "type": "message",
                               "attributes": [
                                   {
                                       "key": "bW9kdWxl",
                                       "value": "d2FzbQ==",
                                       "index": true
                                   },
                                   {
                                       "key": "c2VuZGVy",
                                       "value": "dGVycmExeDQ2cnFheTRkM2Nzc3E4Z3h4dnF6OHh0Nm53bHo0dGQyMGszOHY=",
                                       "index": true
                                   }
                               ]
                           },
                           {
                               "type": "from_contract",
                               "attributes": [
                                   {
                                       "key": "Y29udHJhY3RfYWRkcmVzcw==",
                                       "value": "dGVycmExODNyeDdwcXpqd2o0bWo3cnhycmd2NTg5enNmbDIyeWVhZ2FsYzA=",
                                       "index": true
                                   },
                                   {
                                       "key": "YWN0aW9u",
                                       "value": "dmFsaWRhdG9yX3VwZGF0ZWQ=",
                                       "index": true
                                   },
                                   {
                                       "key": "cHJldmlvdXM=",
                                       "value": "dGVycmExZHFyend4OXRyeDN1aHg1azZjbTdkeG0zZGZnbXN5NThlcHN3bGM=",
                                       "index": true
                                   },
                                   {
                                       "key": "bmV3",
                                       "value": "dGVycmExODNyeDdwcXpqd2o0bWo3cnhycmd2NTg5enNmbDIyeWVhZ2FsYzA=",
                                       "index": true
                                   }
                               ]
                           }
                       ]
                   }
               }
           }
       },
       "events": {
           "transfer.amount": [
               "149032000000uluna"
           ],
           "message.action": [
               "/terra.wasm.v1beta1.MsgExecuteContract"
           ],
           "from_contract.action": [
               "validator_updated"
           ],
           "tx.height": [
               "24717"
           ],
           "coin_spent.spender": [
               "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"
           ],
           "coin_spent.amount": [
               "149032000000uluna"
           ],
           "execute_contract.contract_address": [
               "terra183rx7pqzjwj4mj7rxrrgv589zsfl22yeagalc0"
           ],
           "tm.event": [
               "Tx"
           ],
           "transfer.recipient": [
               "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa"
           ],
           "transfer.sender": [
               "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"
           ],
           "execute_contract.sender": [
               "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"
           ],
           "coin_received.receiver": [
               "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa"
           ],
           "coin_received.amount": [
               "149032000000uluna"
           ],
           "message.sender": [
               "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v",
               "terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"
           ],
           "message.module": [
               "wasm"
           ],
           "from_contract.contract_address": [
               "terra183rx7pqzjwj4mj7rxrrgv589zsfl22yeagalc0"
           ],
           "from_contract.previous": [
               "terra1dqrzwx9trx3uhx5k6cm7dxm3dfgmsy58epswlc"
           ],
           "from_contract.new": [
               "terra183rx7pqzjwj4mj7rxrrgv589zsfl22yeagalc0"
           ],
           "tx.hash": [
               "A56465B67861C9C46D49B22CA42F5EC3FE2035D1D2A4A91C31AE0158B65AC439"
           ]
       }
   }`)
