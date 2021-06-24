package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const HEDERA = "hedera"

func createHederaSubscriber(sub store.Subscription) hederaSubscriber {
	return hederaSubscriber{
		Endpoint:  strings.TrimSuffix(sub.Endpoint.Url, "/"),
		AccountIds: sub.Hedera.AccountIds,
		JobID:     sub.Job,
	}
}

type hederaSubscriber struct {
	Endpoint  string
	AccountIds []string
	JobID     string
}

type hederaSubscription struct {
	endpoint    string
	events      chan<- subscriber.Event
	accountIds   []string
	monitorResp *http.Response
	isDone      bool
	jobid       string
}

func (hSubscr hederaSubscriber) SubscribeToEvents(channel chan<- subscriber.Event, _ store.RuntimeConfig) (subscriber.ISubscription, error) {

	logger.Infof("Using Hedera Mirror endpoint: %s\nListening for events on account ids: %v", hSubscr.Endpoint, hSubscr.AccountIds)

	hederaSubscription := hederaSubscription{
		endpoint:   hSubscr.Endpoint,
		events:     channel,
		accountIds: hSubscr.AccountIds,
		jobid:      hSubscr.JobID,
	}

	//todo implement logic to poll hedera mirror node - interval and accountid needed
	//todo see how to use timestamps to request transaction for period after our last request
	//todo parse result, extract info from memo
	//todo trigger jobs on chainlink node
	//todo map somehow task id from memo to a different jobspec? is that possible? what is jobID in hederaSubscription?
	//todo check if payment was okay
	//todo create tests - hedera_test.go & also elsewhere where we changed the code + db migration

	//todo see where we lose the /
	var url = hederaSubscription.endpoint + "/"


	//go tzs.readMessagesWithRetry()

	var client = NewClient(url, 5)

	client.GetAccountCreditTransactionsAfterTimestamp("0.0.1943014", 324324234234)

	return hederaSubscription, nil
}

func (hSubscr hederaSubscriber) Test() error {
	return nil
}

func (hederaSubscription hederaSubscription) Unsubscribe() {
	logger.Info("Unsubscribing from Hedera endpoint", hederaSubscription.endpoint)
	hederaSubscription.isDone = true
	if hederaSubscription.monitorResp != nil {
		hederaSubscription.monitorResp.Body.Close()
	}
}

//todo see what we need from here and tidy-up somehow
// COPIED FROM HEDERA EVM BRIDGE

type Client struct {
	mirrorAPIAddress string
	httpClient       *http.Client
	pollingInterval  time.Duration
}

type (
	// Transaction struct used by the Hedera Mirror node REST API
	Transaction struct {
		ConsensusTimestamp   string     `json:"consensus_timestamp"`
		EntityId             string     `json:"entity_id"`
		TransactionHash      string     `json:"transaction_hash"`
		ValidStartTimestamp  string     `json:"valid_start_timestamp"`
		ChargedTxFee         int        `json:"charged_tx_fee"`
		MemoBase64           string     `json:"memo_base64"`
		Result               string     `json:"result"`
		Name                 string     `json:"name"`
		MaxFee               string     `json:"max_fee"`
		ValidDurationSeconds string     `json:"valid_duration_seconds"`
		Node                 string     `json:"node"`
		Scheduled            bool       `json:"scheduled"`
		TransactionID        string     `json:"transaction_id"`
		Transfers            []Transfer `json:"transfers"`
		TokenTransfers       []Transfer `json:"token_transfers"`
	}
	// Transfer struct used by the Hedera Mirror node REST API
	Transfer struct {
		Account string `json:"account"`
		Amount  int64  `json:"amount"`
		// When retrieving ordinary hbar transfers, this field does not get populated
		Token string `json:"token_id"`
	}
	// Response struct used by the Hedera Mirror node REST API and returned once
	// account transactions are queried
	Response struct {
		Transactions []Transaction
		Status       `json:"_status"`
	}
	ErrorMessage struct {
		Message string `json:"message"`
	}
	Status struct {
		Messages []ErrorMessage
	}
)

func NewClient(mirrorNodeAPIAddress string, pollingInterval time.Duration) *Client {
	return &Client{
		mirrorAPIAddress: mirrorNodeAPIAddress,
		pollingInterval:  pollingInterval,
		httpClient:       &http.Client{},
	}
}

func (c Client) GetAccountCreditTransactionsAfterTimestamp(accountID string, from int64) (*Response, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc&transactiontype=cryptotransfer",
		accountID,
		String(from))
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

const (
	nanosInSecond = 1000000000
)

// FromString parses a string in the format `{seconds}.{nanos}` into int64 timestamp
func FromString(timestamp string) (int64, error) {
	var err error
	stringTimestamp := strings.Split(timestamp, ".")

	seconds, err := strconv.ParseInt(stringTimestamp[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid timestamp seconds provided")
	}
	nano, err := strconv.ParseInt(stringTimestamp[1], 10, 64)
	if err != nil {
		return 0, errors.New("invalid timestamp nanos provided")
	}

	return seconds*nanosInSecond + nano, nil
}

// String parses int64 timestamp into `{seconds}.{nanos}` string
func String(timestamp int64) string {
	seconds := timestamp / nanosInSecond
	nano := timestamp % nanosInSecond
	return fmt.Sprintf("%d.%d", seconds, nano)
}

// ToHumanReadable converts the timestamp into human readable string
func ToHumanReadable(timestampNanos int64) string {
	parsed := time.Unix(timestampNanos/nanosInSecond, timestampNanos&nanosInSecond)
	return parsed.Format(time.RFC3339Nano)
}

func (c Client) getTransactionsByQuery(query string) (*Response, error) {
	transactionsQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions", query)
	httpResponse, e := c.get(transactionsQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, e := readResponseBody(httpResponse)
	if e != nil {
		return nil, e
	}

	var response *Response
	e = json.Unmarshal(bodyBytes, &response)
	if e != nil {
		return nil, e
	}
	if httpResponse.StatusCode >= 400 {
		return response, errors.New(fmt.Sprintf(`Failed to execute query: [%s]. Error: [%s]`, query, response.Status.String()))
	}

	return response, nil
}

func (c Client) get(query string) (*http.Response, error) {
	return c.httpClient.Get(query)
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

// String converts the Status struct to human readable string
func (s *Status) String() string {
	r := "["
	for i, m := range s.Messages {
		r += m.String()
		if i != len(s.Messages)-1 {
			r += ", "
		}
	}
	r += "]"
	return r
}

// String converts ErrorMessage struct to human readable string
func (m *ErrorMessage) String() string {
	return fmt.Sprintf("message: %s", m.Message)
}