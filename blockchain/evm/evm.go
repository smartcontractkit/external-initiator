package evm

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/smartcontractkit/external-initiator/blockchain/common"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/chainlink/core/utils"
)

const (
	evmWordSize      = utils.EVMWordByteLen
	requesterSize    = evmWordSize
	idSize           = evmWordSize
	paymentSize      = evmWordSize
	versionSize      = evmWordSize
	callbackAddrSize = evmWordSize
	callbackFuncSize = evmWordSize
	expirationSize   = evmWordSize
	dataLocationSize = evmWordSize
	dataLengthSize   = evmWordSize
)

func CreateEvmFilterQuery(jobid string, strAddresses []string) *FilterQuery {
	var addresses []eth.Address
	for _, a := range strAddresses {
		addresses = append(addresses, eth.HexToAddress(a))
	}

	// Hard-set the topics to match the OracleRequest()
	// event emitted by the oracle contract provided.
	topics := [][]eth.Hash{{
		models.RunLogTopic20190207withoutIndexes,
	}, {
		StringToBytes32(jobid),
	}}

	return &FilterQuery{
		Addresses: addresses,
		Topics:    topics,
	}
}

type FilterQuery struct {
	BlockHash *eth.Hash     // used by eth_getLogs, return logs only from block with this hash
	FromBlock string        // beginning of the queried range, nil means genesis block
	ToBlock   string        // end of the range, nil means latest block
	Addresses []eth.Address // restricts matches to events created by specific contracts

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

func (q FilterQuery) ToMapInterface() (map[string]interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != "" || q.ToBlock != "" {
			return nil, errors.New("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == "" {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = q.FromBlock
		}
		if q.ToBlock == "" {
			arg["toBlock"] = "latest"
		} else {
			arg["toBlock"] = q.ToBlock
		}
	}
	return arg, nil
}

func StringToBytes32(str string) eth.Hash {
	value := eth.RightPadBytes([]byte(str), utils.EVMWordByteLen)
	hx := utils.RemoveHexPrefix(hexutil.Encode(value))

	if len(hx) > utils.EVMWordHexLen {
		hx = hx[:utils.EVMWordHexLen]
	}

	hxStr := utils.AddHexPrefix(hx)
	return eth.HexToHash(hxStr)
}

func LogEventToOracleRequest(log models.Log) (common.RunlogRequest, error) {
	cborData, dataPrefixBytes, err := LogDataParse(log.Data)
	if err != nil {
		return common.RunlogRequest{}, err
	}
	js, err := models.ParseCBOR(cborData)
	if err != nil {
		return common.RunlogRequest{}, fmt.Errorf("error parsing CBOR: %v", err)
	}

	request, err := js.AsMap()
	if err != nil {
		return common.RunlogRequest{}, err
	}

	return common.MergeMaps(request, map[string]interface{}{
		"address":          log.Address.String(),
		"dataPrefix":       BytesToHex(dataPrefixBytes),
		"functionSelector": models.OracleFulfillmentFunctionID20190128withoutCast,
	}), nil
}

func LogDataParse(data []byte) (cborData []byte, dataPrefixBytes []byte, rerr error) {
	idStart := requesterSize
	expirationEnd := idStart + idSize + paymentSize + callbackAddrSize + callbackFuncSize + expirationSize

	dataLengthStart := expirationEnd + versionSize + dataLocationSize
	cborStart := dataLengthStart + dataLengthSize

	if len(data) < dataLengthStart+32 {
		return nil, nil, errors.New("malformed data")
	}

	dataLengthBytes, err := models.UntrustedBytes(data).SafeByteSlice(dataLengthStart, dataLengthStart+32)
	if err != nil {
		return nil, nil, err
	}
	dataLength := utils.EVMBytesToUint64(dataLengthBytes)

	if len(data) < cborStart+int(dataLength) {
		return nil, nil, errors.New("cbor too short")
	}

	cborData, err = models.UntrustedBytes(data).SafeByteSlice(cborStart, cborStart+int(dataLength))
	if err != nil {
		return nil, nil, err
	}

	dataPrefixBytes, err = models.UntrustedBytes(data).SafeByteSlice(idStart, expirationEnd)
	if err != nil {
		return nil, nil, err
	}
	return cborData, dataPrefixBytes, nil
}

func BytesToHex(data []byte) string {
	return utils.AddHexPrefix(hex.EncodeToString(data))
}

type NewHeadsResponseParams struct {
	Subscription string                 `json:"subscription"`
	Result       map[string]interface{} `json:"result"`
}

func ParseBlocknumberFromNewHeads(msg common.JsonrpcMessage) (*big.Int, error) {
	var params NewHeadsResponseParams
	err := json.Unmarshal(msg.Params, &params)
	if err != nil {
		return nil, err
	}
	number, ok := params.Result["number"]
	if !ok {
		return nil, errors.New("newHeads result is missing block number")
	}
	return hexutil.DecodeBig(fmt.Sprint(number))
}

func ParseBlockNumberResult(data []byte) (uint64, error) {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(str, 0, 64)
}
