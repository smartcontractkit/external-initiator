package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func handleSubstrateRequest(conn string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	if conn == "ws" {
		switch msg.Method {
		case "state_subscribeStorage":
			return handleSubstrateSubscribeStorage(msg)
		case "state_unsubscribeStorage":
			// TODO: implement
			return nil, nil
		}
	}

	return nil, fmt.Errorf("unexpected method: %v", msg.Method)
}

func handleSubstrateSubscribeStorage(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	var contents [][]string
	err := json.Unmarshal(msg.Params, &contents)
	if err != nil {
		return nil, err
	}

	if len(contents) != 1 || len(contents[0]) != 1 {
		return nil, fmt.Errorf("possibly incorrect length of params array: %v", len(contents))
	}

	subId := uuid.New()
	result := ethSubscribeResponse{
		Subscription: subId.String(),
		Result:       getParamsFromStorageKey(contents[0][0]),
	}
	resultBz, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return []JsonrpcMessage{
		{
			Version: "2.0",
			ID:      msg.ID,
			Result:  []byte(fmt.Sprintf(`"%s"`, subId.String())),
		},
		{
			Version: "2.0",
			Method:  "state_storage",
			Params:  resultBz,
		},
	}, nil
}

func getParamsFromStorageKey(key string) json.RawMessage {
	if key == "0x26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7" {
		return generateEvents()
	}

	var change string
	switch key[:82] {
	// Feeds()
	case "0x1ae70894dea2956c24d9c19ac4d15d33407d94c38db8ad2e4aa005c8942ac2c5b4def25cfda6ef3a":
		change = "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d0000000000000000000000000000000000ffc99a3b000000000000000000000000010000000100000000e876481700000000000000000000000100000002080123000000000600000006000000010100000001000000"
	// Round()
	case "0x1ae70894dea2956c24d9c19ac4d15d337b45e7c782afb68af72fc5ebb26b974bb4def25cfda6ef3a":
		change = "0x1a0000000100000000000000000000000000000000011a0000000100000000"
	// OracleStati()
	case "0x1ae70894dea2956c24d9c19ac4d15d33d6e0ca5ff50f58afe0ab51e6398dda4fb4def25cfda6ef3a":
		change = "0x0000000000010600000001060000000138150000000000000000000000000000"
	default:
		fmt.Println("unknown key:", key)
		return nil
	}

	return generateChanges(key, change)
}

func generateChanges(key, change string) json.RawMessage {
	block := "0xc0b664cdf2c0e0f49aeddd0f3aa272c49310d5a75efce39c0e0c9ee977ae8bec"
	contents := fmt.Sprintf(`{"block":"%s","changes":[["%s","%s"]]}`, block, key, change)
	return json.RawMessage(contents)
}

func generateEvents() json.RawMessage {
	block := "0x52be2d315d20a3909c7b551469a446a76d4e85ea4b30b986044fefef6bc726d8"
	contents := fmt.Sprintf(`{"block":"%s","changes":[["0x26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7","0x1400000000000000482d7c09000000000200000001000000080100000000080000007c522c8273973e7bcf4a5dbfcc745dba4a3ab08c1e410167d7b1bdf9cb924f6c630e000000000100000008020000000008000000c91400000000000000000000000000007c522c8273973e7bcf4a5dbfcc745dba4a3ab08c1e410167d7b1bdf9cb924f6c00000100000008030000000008000000c9140000000000000000000000000000630e000000000100000000006400000000000000000000"]]}`, block)
	return json.RawMessage(contents)
}
