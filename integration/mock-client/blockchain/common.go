package blockchain

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/external-initiator/blockchain"
)

// JsonrpcMessage declares JSON-RPC message type
type JsonrpcMessage = blockchain.JsonrpcMessage

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
	case "binance-smart-chain":
		return handleBscRequest(conn, msg)
	case "near":
		return handleNEARRequest(conn, msg)
	case "cfx":
		return handleCfxRequest(conn, msg)
	case "keeper":
		return handleKeeperRequest(conn, msg)
	default:
		return nil, fmt.Errorf("unexpected platform: %v", platform)
	}
}

func SetHttpRoutes(router *gin.Engine) {
	setXtzRoutes(router)
	setBSNIritaRoutes(router)
}
