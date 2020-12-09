package blockchain

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/tendermint/tendermint/rpc/jsonrpc/types"
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
			errintf = types.RPCError{
				Message: err.Error(),
			}

			response.Error = &errintf
		}

		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, rsp[0])
	return
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
	var params []json.RawMessage
	err := json.Unmarshal(msg.Params, &params)
	if err != nil {
		return nil, err
	}

	if len(params) != 4 {
		return nil, fmt.Errorf("incorrect length of params array: %v", len(params))
	}

	var path string
	err = json.Unmarshal(params[0], &path)
	if err != nil {
		return nil, err
	}

	if path == "/custom/service/request" {
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
