// might cover all cosmos/tendermint chains later

package terra

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
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
	url := fmt.Sprintf("%s/wasm/contracts/%s/store?query_msg=%s", os.Getenv("TERRA_URL"), address, query)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("[terra/manager/query] query failed: %s %+v", url, resp)
	}

	defer resp.Body.Close()

	var decoded map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}

	if err := json.Unmarshal(decoded["result"], &t); err != nil {
		return err
	}
	logger.Debug("[terra/manager/query]", url, t)

	return nil
}

func (tm *manager) subscribe(ctx context.Context, queryFilter string, handler func(event EventRecords)) error {
	responses := make(chan json.RawMessage)
	filter := []string{queryFilter}
	params, err := json.Marshal(filter)
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
			eventRecords, err := parseEvents(events)
			if err != nil {
				logger.Error(err)
				continue
			}
			if eventRecords != nil {
				handler(*eventRecords)
			}
		}
	}()

	return nil
}

func extractEvents(data json.RawMessage) ([]types.Event, error) {
	value := gjson.Get(string(data), "data.value.TxResult.result.events")

	var events []types.Event
	err := json.Unmarshal([]byte(value.Raw), &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func parseEvents(events []types.Event) (*EventRecords, error) {
	var eventRecords EventRecords
	for _, event := range events {
		switch event.Type {
		case "wasm-new_round":
			newRound, err := parseNewRoundEvent(event)
			if err != nil {
				return nil, err
			}
			eventRecords.NewRound = append(eventRecords.NewRound, *newRound)
		case "wasm-submission_received":
			submission, err := parseSubmissionReceivedEvent(event)
			if err != nil {
				return nil, err
			}
			eventRecords.SubmissionReceived = append(eventRecords.SubmissionReceived, *submission)
		case "wasm-answer_updated":
			answerUpdated, err := parseAnswerUpdatedEvent(event)
			if err != nil {
				return nil, err
			}
			eventRecords.AnswerUpdated = append(eventRecords.AnswerUpdated, *answerUpdated)
		case "wasm-oracle_permissions_updated":
			permissionsUpdated, err := parseOraclePermissionsUpdatedEvent(event)
			if err != nil {
				return nil, err
			}
			eventRecords.OraclePermissionsUpdated = append(eventRecords.OraclePermissionsUpdated, permissionsUpdated...)
		case "wasm-round_details_updated":
			detailsUpdated, err := parseRoundDetailsUpdatedEvent(event)
			if err != nil {
				return nil, err
			}
			eventRecords.RoundDetailsUpdated = append(eventRecords.RoundDetailsUpdated, *detailsUpdated)
		}
	}

	return &eventRecords, nil
}

func parseNewRoundEvent(event types.Event) (*EventNewRound, error) {
	attributes, err := getRequiredAttributes(event, []string{"round_id", "started_by"})
	if err != nil {
		return nil, err
	}
	roundId, err := strconv.Atoi(attributes["round_id"])
	if err != nil {
		return nil, err
	}
	var startedAt uint64
	startedAtStr, err := getAttributeValue(event, "started_at")
	if err == nil {
		value, err := strconv.Atoi(startedAtStr)
		if err != nil {
			return nil, err
		}
		startedAt = uint64(value)
	}
	return &EventNewRound{
		RoundId:   uint32(roundId),
		StartedBy: Addr(attributes["started_by"]),
		StartedAt: startedAt,
	}, nil
}

func parseSubmissionReceivedEvent(event types.Event) (*EventSubmissionReceived, error) {
	attributes, err := getRequiredAttributes(event, []string{"submission", "round_id", "oracle"})
	if err != nil {
		return nil, err
	}

	submission := new(big.Int)
	submission, _ = submission.SetString(attributes["submission"], 10)

	roundId, err := strconv.Atoi(attributes["round_id"])
	if err != nil {
		return nil, err
	}

	return &EventSubmissionReceived{
		Oracle:     Addr(attributes["oracle"]),
		Submission: Value{*submission},
		RoundId:    uint32(roundId),
	}, nil
}

func parseAnswerUpdatedEvent(event types.Event) (*EventAnswerUpdated, error) {
	attributes, err := getRequiredAttributes(event, []string{"round_id", "current"})
	if err != nil {
		return nil, err
	}

	roundId, err := strconv.Atoi(attributes["round_id"])
	if err != nil {
		return nil, err
	}

	value := new(big.Int)
	value, _ = value.SetString(attributes["current"], 10)

	return &EventAnswerUpdated{
		Value:   Value{*value},
		RoundId: uint32(roundId),
	}, nil
}

func parseOraclePermissionsUpdatedEvent(event types.Event) (events []EventOraclePermissionsUpdated, err error) {
	attributes, err := getRequiredAttributes(event, []string{"added", "removed"})
	if err != nil {
		return nil, err
	}

	var added []string
	err = json.Unmarshal([]byte(attributes["added"]), &added)
	if err != nil {
		return nil, err
	}
	for _, oracle := range added {
		events = append(events, EventOraclePermissionsUpdated{
			Oracle: Addr(oracle),
			Bool:   true,
		})
	}

	var removed []string
	err = json.Unmarshal([]byte(attributes["removed"]), &removed)
	if err != nil {
		return nil, err
	}
	for _, oracle := range removed {
		events = append(events, EventOraclePermissionsUpdated{
			Oracle: Addr(oracle),
			Bool:   false,
		})
	}

	return
}

func parseRoundDetailsUpdatedEvent(event types.Event) (*EventRoundDetailsUpdated, error) {
	attributes, err := getRequiredAttributes(event, []string{"payment_amount", "restart_delay", "timeout"})
	if err != nil {
		return nil, err
	}

	payment := new(big.Int)
	payment, _ = payment.SetString(attributes["payment_amount"], 10)

	delay, err := strconv.Atoi(attributes["restart_delay"])
	if err != nil {
		return nil, err
	}

	timeout, err := strconv.Atoi(attributes["timeout"])
	if err != nil {
		return nil, err
	}

	return &EventRoundDetailsUpdated{
		PaymentAmount: Value{*payment},
		RestartDelay:  uint32(delay),
		Timeout:       uint32(timeout),
	}, nil
}

func getAttributeValue(event types.Event, attributeKey string) (string, error) {
	for _, attr := range event.Attributes {
		if string(attr.Key) == attributeKey {
			return string(attr.Value), nil
		}
	}

	return "", fmt.Errorf("attribute key %s does not exist", attributeKey)
}

func getRequiredAttributes(event types.Event, attributes []string) (map[string]string, error) {
	var attrs = make(map[string]string)
	for _, attr := range attributes {
		value, err := getAttributeValue(event, attr)
		if err != nil {
			return nil, err
		}

		attrs[attr] = value
	}
	return attrs, nil
}
