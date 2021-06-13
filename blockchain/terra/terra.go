// might cover all cosmos/tendermint chains later

package terra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tidwall/gjson"
)

const Name = "terra"

type manager struct {
	endpointName    string
	contractAddress string
	accountId       string // TODO! accountAddress should be a more appropriate name
	subscriber      subscriber.ISubscriber
}

func createManager(sub store.Subscription) (*manager, error) {
	conn, err := subscriber.NewSubscriber(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	return &manager{
		endpointName:    sub.EndpointName,
		contractAddress: sub.Terra.ContractAddress,
		accountId:       sub.Terra.AccountId,
		subscriber:      conn,
	}, nil
}

func (tm *manager) Stop() {
	// TODO!
}

func (tm *manager) query(ctx context.Context, address, query string, t interface{}) error {
	// TODO! remove hardcoded url; potentially use Tendermint http client
	url := fmt.Sprintf("http://localhost:1317/wasm/contracts/%s/store?query_msg=%s", address, query)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var decoded map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}

	if err := json.Unmarshal(decoded["result"], &t); err != nil {
		return err
	}

	return nil
}

func (tm *manager) subscribe(ctx context.Context, queryFilter string, handler func(event EventRecords)) error {
	responses := make(chan json.RawMessage)
	params, err := json.Marshal(queryFilter)
	if err != nil {
		return err
	}

	err = tm.subscriber.Subscribe(ctx, "subscribe", "unsubscribe", params, responses)
	if err != nil {
		return err
	}

	// TODO!
	go func() {
		for {
			select {
			case resp, ok := <-responses:
				if !ok {
					return
				}

				events, err := extractEvents(resp)
				if err != nil {
					logger.Error(err)
					continue
				}
				attributes := extractCustomAttributes(events)
				event, err := attributesToEvent(attributes)
				if err != nil {
					logger.Error(err)
					continue
				}
				if event != nil {
					handler(*event)
				}
			}
		}
	}()

	return nil
}

func extractEvents(data json.RawMessage) ([]types.Event, error) {
	value := gjson.Get(string(data), "result.data.value.TxResult.result.events") // TODO! this parsing should be improved

	var events []types.Event
	err := json.Unmarshal([]byte(value.Raw), &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func extractCustomAttributes(events []types.Event) []types.EventAttribute {
	for _, event := range events {
		if event.Type == "from_contract" {
			return event.Attributes
		}
	}

	return nil
}

func attributesToEvent(attributes []types.EventAttribute) (*EventRecords, error) {
	action, err := getAttributeValue(attributes, "action")
	if err != nil {
		return nil, err
	}

	switch action {
	case "new_round":
		// TODO!
		// roundId, err := getAttributeValue(attributes, "round_id")
		// if err != nil {
		// 	return nil, err
		// }
		// startedBy, err := getAttributeValue(attributes, "started_by")
		// if err != nil {
		// 	return nil, err
		// }
		// startedAt, err := getAttributeValue(attributes, "started_at")
		// if err != nil {
		// 	return nil, err
		// }
		// return &EventRecords{
		// 	NewRound: []EventNewRound{
		// 		{RoundId: roundId},
		// 	},
		// }, nil
	case "answer_updated":

	case "oracle_permissions_updated":
	}

	return nil, nil
}

func getAttributeValue(attributes []types.EventAttribute, attributeKey string) (string, error) {
	for _, attr := range attributes {
		if string(attr.Key) == attributeKey {
			return string(attr.Value), nil
		}
	}

	return "", fmt.Errorf("attribute key %s does not exist", attributeKey)
}
