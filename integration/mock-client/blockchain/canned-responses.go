package blockchain

import (
	"encoding/json"

	"github.com/smartcontractkit/chainlink/core/logger"
	mockresponses "github.com/smartcontractkit/external-initiator/integration/mock-client/blockchain/mock-responses"
)

type cannedResponse map[string][]JsonrpcMessage

func GetCannedResponse(platform string, msg JsonrpcMessage) ([]JsonrpcMessage, bool) {
	bz, err := mockresponses.Get(platform)
	if err != nil {
		logger.Debug(err)
		return nil, false
	}

	var responses cannedResponse
	err = json.Unmarshal(bz, &responses)
	if err != nil {
		logger.Error(err)
		return nil, false
	}

	responseList, ok := responses[msg.Method]
	if !ok {
		return nil, false
	}

	return setJsonrpcId(msg.ID, responseList), true
}

func setJsonrpcId(id json.RawMessage, msgs []JsonrpcMessage) []JsonrpcMessage {
	for i := 0; i < len(msgs); i++ {
		msgs[i].ID = id
	}
	return msgs
}
