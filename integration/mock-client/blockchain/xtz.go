package blockchain

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/integration/mock-client/blockchain/static"
)

func setXtzRoutes(router *gin.Engine) {
	router.GET("/http/xtz/monitor/heads/:chain_id", handleXtzMonitorRequest)
	router.GET("/http/xtz/chains/main/blocks/:block_id/operations", handleXtzOperationsRequest)
}

type xtzResponses map[string]interface{}

func getXtzResponse(method string) (interface{}, error) {
	bz, err := static.Get("xtz")
	if err != nil {
		return nil, err
	}

	var responses xtzResponses
	err = json.Unmarshal(bz, &responses)
	if err != nil {
		return nil, err
	}

	response, ok := responses[method]
	if !ok {
		return nil, errors.New("method not found")
	}

	return response, nil
}

func handleXtzMonitorRequest(c *gin.Context) {
	resp, err := getXtzResponse("monitor")
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func handleXtzOperationsRequest(c *gin.Context) {
	resp, err := getXtzResponse("operations")
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}
