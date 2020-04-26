package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink/core/logger"
	"net/http"
)

type JsonRpcRequest struct {
	Version string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type JsonRpcResponse struct {
	Id     string          `json:"id"`
	Error  int64           `json:"error"`
	Desc   string          `json:"desc"`
	Result json.RawMessage `json:"result"`
}

func setOntRoutes(router *gin.Engine) {
	router.POST("/http/ont", handleOntRpcRequest)
}

func handleOntRpcRequest(c *gin.Context) {
	var req JsonRpcRequest
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	resp, err := HandleOntRequest(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func HandleOntRequest(msg JsonRpcRequest) (JsonRpcResponse, error) {
	switch msg.Method {
	case "getblockcount":
		return handleGetBlockCount(msg)
	case "getsmartcodeevent":
		return handleGetSmartCodeEvent(msg)
	}

	return JsonRpcResponse{}, errors.New(fmt.Sprint("unexpected method: ", msg.Method))
}

func handleGetBlockCount(msg JsonRpcRequest) (JsonRpcResponse, error) {
	r, _ := json.Marshal(1)
	return JsonRpcResponse{
			Id:      msg.Id,
			Result:  r,
	}, nil
}

type ExecuteNotify struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []NotifyEventInfo
}

type NotifyEventInfo struct {
	ContractAddress string
	States          interface{}
}

func handleGetSmartCodeEvent(msg JsonRpcRequest) (JsonRpcResponse, error) {
	eInfos := make([]*ExecuteNotify, 0)
	data, err := json.Marshal(eInfos)
	if err != nil {
		return JsonRpcResponse{}, err
	}

	return JsonRpcResponse{
			Id:      msg.Id,
			Result:  data,
	}, nil
}
