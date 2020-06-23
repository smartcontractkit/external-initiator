package blockchain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/whisper/whisperv6"
	"github.com/smartcontractkit/chainlink/core/eth"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/chainlink/core/utils"
)

const (
	evmWordSize      = common.HashLength
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

func StringToBytes32(jobid string) string {
	value := common.RightPadBytes([]byte(jobid), utils.EVMWordByteLen)
	hx := utils.RemoveHexPrefix(hexutil.Encode(value))

	if len(hx) > utils.EVMWordHexLen {
		hx = hx[:utils.EVMWordHexLen]
	}

	return utils.AddHexPrefix(hx)
}

func logEventToOracleRequest(log eth.Log) (models.JSON, error) {
	data := log.Data
	idStart := requesterSize
	expirationEnd := idStart + idSize + paymentSize + callbackAddrSize + callbackFuncSize + expirationSize

	dataLengthStart := expirationEnd + versionSize + dataLocationSize
	cborStart := dataLengthStart + dataLengthSize

	if len(log.Data) < dataLengthStart+32 {
		return models.JSON{}, errors.New("malformed data")
	}

	dataLengthBytes, err := data.SafeByteSlice(dataLengthStart, dataLengthStart+32)
	if err != nil {
		return models.JSON{}, err
	}
	dataLength := whisperv6.BytesToUintBigEndian(dataLengthBytes)

	if len(log.Data) < cborStart+int(dataLength) {
		return models.JSON{}, errors.New("cbor too short")
	}

	cborData, err := data.SafeByteSlice(cborStart, cborStart+int(dataLength))
	if err != nil {
		return models.JSON{}, err
	}

	js, err := models.ParseCBOR(cborData)
	if err != nil {
		return js, fmt.Errorf("error parsing CBOR: %v", err)
	}

	dataPrefixBytes, err := data.SafeByteSlice(idStart, expirationEnd)
	if err != nil {
		return models.JSON{}, err
	}

	return js.MultiAdd(models.KV{
		"address":          log.Address.String(),
		"dataPrefix":       bytesToHex(dataPrefixBytes),
		"functionSelector": models.OracleFulfillmentFunctionID20190128withoutCast,
	})
}

func bytesToHex(data []byte) string {
	return utils.AddHexPrefix(hex.EncodeToString(data))
}
