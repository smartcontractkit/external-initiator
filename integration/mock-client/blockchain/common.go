package blockchain

import (
	"fmt"

	"github.com/smartcontractkit/external-initiator/blockchain/common"

	"github.com/smartcontractkit/external-initiator/blockchain/common"

	"github.com/gin-gonic/gin"
)

// JsonrpcMessage declares JSON-RPC message type
type JsonrpcMessage = common.JsonrpcMessage

func HandleRequest(conn, platform string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	cannedResponse, ok := GetCannedResponse(platform, msg)
	if ok {
		return cannedResponse, nil
	}

	switch platform {
	case "eth":
		return handleEthRequest(conn, msg)
	case "hmy":
		return handleHmyRequest(conn, msg)
	case "ont":
		return handleOntRequest(msg)
	case "near":
		return handleNEARRequest(conn, msg)
	case "cfx":
		return handleCfxRequest(conn, msg)
	case "keeper":
		return handleKeeperRequest(conn, msg)
	case "substrate":
		return handleSubstrateRequest(conn, msg)
	default:
		return nil, fmt.Errorf("unexpected platform: %v", platform)
	}
}

func SetHttpRoutes(router *gin.Engine) {
	setXtzRoutes(router)
	setBSNIritaRoutes(router)
}
