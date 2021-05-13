package conflux

import (
	"errors"

	"github.com/smartcontractkit/external-initiator/blockchain/evm"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

const Name = "conflux"

// The manager implements the common.Manager interface and allows
// for interacting with CFX nodes over RPC.
type manager struct {
	fq           *filterQuery
	endpointName string
	jobid        string

	subscriber subscriber.ISubscriber
}

// createManager creates a new instance of cfxManager with the provided
// connection type and store.CfxSubscription config.
func createManager(sub store.Subscription) (*manager, error) {
	fq, err := createFilterQuery(sub.Job, sub.Conflux.Addresses)
	if err != nil {
		return nil, err
	}

	conn, err := subscriber.NewSubscriber(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	return &manager{
		fq:           &fq,
		endpointName: sub.EndpointName,
		jobid:        sub.Job,
		subscriber:   conn,
	}, nil
}

func (m manager) Stop() {
	// TODO: Implement
}

type cfxLogResponse struct {
	LogIndex         string             `json:"logIndex"`
	EpochNumber      string             `json:"epochNumber"`
	BlockHash        eth.Hash           `json:"blockHash"`
	TransactionHash  eth.Hash           `json:"transactionHash"`
	TransactionIndex string             `json:"transactionIndex"`
	Address          cfxaddress.Address `json:"address"`
	Data             string             `json:"data"`
	Topics           []eth.Hash         `json:"topics"`
}

// cfx2EthResponse converts cfxLogResponse type to eth.Log type
func cfx2EthResponse(cfx cfxLogResponse) (models.Log, error) {
	blockNumber, err := hexutil.DecodeUint64(cfx.EpochNumber)
	if err != nil {
		return models.Log{}, err
	}

	txIndex, err := hexutil.DecodeUint64(cfx.TransactionIndex)
	if err != nil {
		return models.Log{}, err
	}

	index, err := hexutil.DecodeUint64(cfx.LogIndex)
	if err != nil {
		return models.Log{}, err
	}

	data := eth.Hex2Bytes(cfx.Data[2:])
	return models.Log{
		Address:     eth.HexToAddress(cfx.Address.GetHexAddress()),
		Topics:      cfx.Topics,
		Data:        data,
		BlockNumber: blockNumber,
		TxHash:      cfx.TransactionHash,
		TxIndex:     uint(txIndex),
		BlockHash:   cfx.BlockHash,
		Index:       uint(index),
	}, nil
}

type filterQuery struct {
	BlockHash *eth.Hash            // used by cfx_getLogs, return logs only from block with this hash
	FromEpoch string               // beginning of the queried range, nil means genesis block
	ToEpoch   string               // end of the range, nil means latest block
	Addresses []cfxaddress.Address // restricts matches to events created by specific contracts

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	Topics [][]eth.Hash
}

func createFilterQuery(jobid string, strAddresses []string) (filterQuery, error) {
	var addresses []cfxaddress.Address
	for _, a := range strAddresses {
		base32Addr, err := cfxaddress.NewFromBase32(a)
		if err != nil {
			return filterQuery{}, err
		}
		addresses = append(addresses, base32Addr)
	}

	// Hard-set the topics to match the OracleRequest()
	// event emitted by the oracle contract provided.
	topics := [][]eth.Hash{{
		models.RunLogTopic20190207withoutIndexes,
	}, {
		evm.StringToBytes32(jobid),
	}}

	return filterQuery{
		Addresses: addresses,
		Topics:    topics,
	}, nil
}

func (q filterQuery) toMapInterface() (map[string]interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromEpoch != "" || q.ToEpoch != "" {
			return nil, errors.New("cannot specify both BlockHash and FromEpoch/ToEpoch")
		}
	} else {
		if q.FromEpoch == "" {
			arg["fromEpoch"] = "0x0"
		} else {
			arg["fromEpoch"] = q.FromEpoch
		}
		if q.ToEpoch == "" {
			arg["toEpoch"] = "latest_state"
		} else {
			arg["toEpoch"] = q.ToEpoch
		}
	}
	return arg, nil
}
