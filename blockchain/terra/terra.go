// might cover all cosmos/tendermint chains later

package terra

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tidwall/gjson"
)

const Name = "terra"

type TerraParams struct {
	ContractAddress string `json:"contract_address"`
	AccountAddress  string `json:"account_address"`
	// FcdUrl          string
}

type manager struct {
	endpointName    string
	contractAddress string
	accountAddress  string
	// fcdUrl          string
	subscriber subscriber.ISubscriber
}

func createManager(sub store.Subscription) (*manager, error) {
	conn, err := subscriber.NewSubscriber(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	return &manager{
		endpointName:    sub.EndpointName,
		contractAddress: sub.Terra.ContractAddress,
		accountAddress:  sub.Terra.AccountAddress,
		subscriber:      conn,
	}, nil
}

func (tm *manager) Stop() {
	// TODO!
}

func (tm *manager) query(ctx context.Context, address, query string, t interface{}) error {
	// TODO! remove hardcoded url; potentially use Tendermint http client
	url := fmt.Sprintf("%s/wasm/contracts/%s/store?query_msg=%s", os.Getenv("TERRA_URL"), address, url.QueryEscape(query))
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

	go func() {
		for {
			resp, ok := <-responses
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

	// TODO! improve event extraction
	switch action {
	case "new_round":
		roundIdStr, err := getAttributeValue(attributes, "round_id")
		if err != nil {
			return nil, err
		}
		roundId, _ := strconv.Atoi(roundIdStr)
		startedBy, err := getAttributeValue(attributes, "started_by")
		if err != nil {
			return nil, err
		}
		var startedAt OptionU64
		startedAtStr, err := getAttributeValue(attributes, "started_at")
		if err == nil {
			value, err := strconv.Atoi(startedAtStr)
			if err != nil {
				return nil, err
			}
			startedAt.hasValue = true
			startedAt.uint64 = uint64(value)
		}
		return &EventRecords{
			NewRound: []EventNewRound{
				{
					RoundId:   uint32(roundId),
					StartedBy: Addr(startedBy),
					StartedAt: startedAt,
				},
			},
		}, nil
	case "answer_updated":
		roundIdStr, err := getAttributeValue(attributes, "round_id")
		if err != nil {
			return nil, err
		}
		roundId, _ := strconv.Atoi(roundIdStr)
		valueStr, err := getAttributeValue(attributes, "current")
		if err != nil {
			return nil, err
		}
		var value *big.Int
		value, _ = value.SetString(valueStr, 10)
		return &EventRecords{
			AnswerUpdated: []EventAnswerUpdated{
				{
					Value:   *value,
					RoundId: uint32(roundId),
				},
			},
		}, nil
	case "oracle_permissions_updated":
		var permissionChanges []EventOraclePermissionsUpdated

		addedStr, err := getAttributeValue(attributes, "added")
		if err != nil {
			return nil, err
		}
		var added []string
		err = json.Unmarshal([]byte(addedStr), &added)
		if err != nil {
			return nil, err
		}
		for _, oracle := range added {
			permissionChanges = append(permissionChanges, EventOraclePermissionsUpdated{
				Oracle: Addr(oracle),
				Bool:   true,
			})
		}

		removedStr, err := getAttributeValue(attributes, "removed")
		if err != nil {
			return nil, err
		}
		var removed []string
		err = json.Unmarshal([]byte(removedStr), &removed)
		if err != nil {
			return nil, err
		}
		for _, oracle := range removed {
			permissionChanges = append(permissionChanges, EventOraclePermissionsUpdated{
				Oracle: Addr(oracle),
				Bool:   false,
			})
		}

		return &EventRecords{
			OraclePermissionsUpdated: permissionChanges,
		}, nil
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
