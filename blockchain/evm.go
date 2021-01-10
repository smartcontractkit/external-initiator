package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
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

func createEvmFilterQuery(jobid string, strAddresses []string) *filterQuery {
	var addresses []common.Address
	for _, a := range strAddresses {
		addresses = append(addresses, common.HexToAddress(a))
	}

	// Hard-set the topics to match the OracleRequest()
	// event emitted by the oracle contract provided.
	topics := [][]common.Hash{{
		models.RunLogTopic20190207withoutIndexes,
	}, {
		StringToBytes32(jobid),
	}}

	return &filterQuery{
		Addresses: addresses,
		Topics:    topics,
	}
}

type filterQuery struct {
	BlockHash *common.Hash     // used by eth_getLogs, return logs only from block with this hash
	FromBlock string           // beginning of the queried range, nil means genesis block
	ToBlock   string           // end of the range, nil means latest block
	Addresses []common.Address // restricts matches to events created by specific contracts

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
	Topics [][]common.Hash
}

func (q filterQuery) toMapInterface() (interface{}, error) {
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

func StringToBytes32(str string) common.Hash {
	value := common.RightPadBytes([]byte(str), utils.EVMWordByteLen)
	hx := utils.RemoveHexPrefix(hexutil.Encode(value))

	if len(hx) > utils.EVMWordHexLen {
		hx = hx[:utils.EVMWordHexLen]
	}

	hxStr := utils.AddHexPrefix(hx)
	return common.HexToHash(hxStr)
}

func logEventToOracleRequest(log models.Log) (models.JSON, error) {
	cborData, dataPrefixBytes, err := logDataParse(log.Data)
	if err != nil {
		return models.JSON{}, err
	}
	js, err := models.ParseCBOR(cborData)
	if err != nil {
		return js, fmt.Errorf("error parsing CBOR: %v", err)
	}
	return js.MultiAdd(models.KV{
		"address":          log.Address.String(),
		"dataPrefix":       bytesToHex(dataPrefixBytes),
		"functionSelector": models.OracleFulfillmentFunctionID20190128withoutCast,
	})
}

func logDataParse(data []byte) (cborData []byte, dataPrefixBytes []byte, rerr error) {
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

func bytesToHex(data []byte) string {
	return utils.AddHexPrefix(hex.EncodeToString(data))
}

type newHeadsResponseParams struct {
	Subscription string                 `json:"subscription"`
	Result       map[string]interface{} `json:"result"`
}

func ParseBlocknumberFromNewHeads(msg JsonrpcMessage) (*big.Int, error) {
	var params newHeadsResponseParams
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

func GetBlockNumberPayload() ([]byte, error) {
	msg := JsonrpcMessage{
		Version: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "eth_blockNumber",
	}
	return json.Marshal(msg)
}

func sendEthNodePost(endpoint string, payload []byte) (*http.Response, error) {
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("%s returned 400. This endpoint may not support calls to /monitor", endpoint)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %v from endpoint %s", resp.StatusCode, endpoint)
	}
	return resp, nil
}
