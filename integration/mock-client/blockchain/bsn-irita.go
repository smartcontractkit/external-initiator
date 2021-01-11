package blockchain

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmrpc "github.com/tendermint/tendermint/rpc/jsonrpc/types"

	"github.com/smartcontractkit/chainlink/core/logger"
)

func setBSNIritaRoutes(router *gin.Engine) {
	router.POST("/", handleBSNIritaRPC)
}

func handleBSNIritaRPC(c *gin.Context) {
	var req JsonrpcMessage
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	rsp, err := handleBSNIritaRequest(req)
	if len(rsp) == 0 || err != nil {
		var response JsonrpcMessage
		response.ID = req.ID
		response.Version = req.Version

		if err != nil {
			logger.Error(err)

			var errintf interface{}
			errintf = tmrpc.RPCError{
				Message: err.Error(),
			}

			response.Error = &errintf
		}

		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, rsp[0])
}

func handleBSNIritaRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch msg.Method {
	case "status", "block_results":
		rsp, ok := GetCannedResponse("birita", msg)
		if !ok {
			return nil, fmt.Errorf("failed to handle BSN-IRITA request for method %s", msg.Method)
		}

		return rsp, nil

	case "abci_query":
		return handleQueryABCI(msg)

	default:
		return nil, fmt.Errorf("unexpected method: %v", msg.Method)
	}
}

func handleQueryABCI(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	var params abci.RequestQuery
	err := tmjson.Unmarshal(msg.Params, &params)
	if err != nil {
		return nil, err
	}

	if params.Path == "/custom/service/request" {
		return handleQueryServiceRequest(msg)
	}

	return []JsonrpcMessage{
		{
			ID:     msg.ID,
			Result: []byte{},
		},
	}, nil
}

func handleQueryServiceRequest(msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	msg.Method = "abci_query_service_request"

	rsp, ok := GetCannedResponse("birita", msg)
	if !ok {
		return nil, fmt.Errorf("failed to handle BSN-IRITA request for service request query")
	}

	return rsp, nil
}
