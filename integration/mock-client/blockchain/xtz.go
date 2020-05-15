package blockchain

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink/core/logger"
)

type xtzMonitorResponse struct {
	Hash           string   `json:"hash"`
	Level          int      `json:"level"`
	Proto          int      `json:"proto"`
	Predecessor    string   `json:"predecessor"`
	TimeStamp      string   `json:"timestamp"`
	ValidationPass int      `json:"validation_pass"`
	OperationsHash string   `json:"operations_hash"`
	Fitness        []string `json:"fitness"`
	Context        string   `json:"context"`
	ProtocolData   string   `json:"protocol_data"`
}

type xtzTransaction struct {
	Protocol string                  `json:"protocol"`
	ChainID  string                  `json:"chain_id"`
	Hash     string                  `json:"hash"`
	Branch   string                  `json:"branch"`
	Contents []xtzTransactionContent `json:"contents"`
}

type xtzTransactionContent struct {
	Kind         string                        `json:"kind"`
	Source       string                        `json:"source"`
	Fee          string                        `json:"fee"`
	Counter      string                        `json:"counter"`
	GasLimit     string                        `json:"gas_limit"`
	StorageLimit string                        `json:"storage_limit"`
	Amount       string                        `json:"amount"`
	Destination  string                        `json:"destination"`
	Parameters   interface{}                   `json:"parameters"`
	Metadata     xtzTransactionContentMetadata `json:"metadata"`
}

type xtzTransactionContentMetadata struct {
	BalanceUpdates           []interface{}                 `json:"balance_updates"`
	OperationResult          interface{}                   `json:"operation_result"`
	InternalOperationResults *[]xtzInternalOperationResult `json:"internal_operation_results"`
}

type xtzInternalOperationResult struct {
	Kind        string      `json:"kind"`
	Source      string      `json:"source"`
	Nonce       int         `json:"nonce"`
	Amount      string      `json:"amount"`
	Destination string      `json:"destination"`
	Parameters  interface{} `json:"parameters"`
	Result      interface{} `json:"result"`
}

func setXtzRoutes(router *gin.Engine) {
	router.GET("/http/xtz/monitor/heads/:chain_id", handleXtzMonitorRequest)
	router.GET("/http/xtz/chains/main/blocks/:block_id/operations", handleXtzOperationsRequest)
}

func handleXtzMonitorRequest(c *gin.Context) {
	resp, err := getXtzMonitorResponse(c.Param("chain_id"))

	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func handleXtzOperationsRequest(c *gin.Context) {
	resp, err := getXtzOperationsResponse(c.Param("block_id"))

	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func getXtzMonitorResponse(chainId string) (xtzMonitorResponse, error) {
	return xtzMonitorResponse{
		Hash:           "8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D",
		Level:          0,
		Proto:          0,
		Predecessor:    "BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3",
		ValidationPass: 0,
		OperationsHash: "00000000000000000000000000000000000000000000000000000000",
		Context:        "00000000000000000000000000000000000000000000000000000000",
		ProtocolData:   "",
	}, nil
}

func getXtzOperationsResponse(blockId string) ([][]xtzTransaction, error) {
	subscriptionAddress := os.Getenv("SUBSCRIBED_ADDRESS")

	transactionContents := []xtzTransactionContent{
		{
			Kind:         "transaction",
			Source:       "BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3",
			Fee:          "666",
			Counter:      "666",
			GasLimit:     "6666",
			StorageLimit: "42",
			Amount:       "66666",
			Destination:  subscriptionAddress,
			Metadata:     xtzTransactionContentMetadata{},
		},
	}
	transactions := [][]xtzTransaction{
		{},
		{},
		{},
		{
			{
				Protocol: "nonsense",
				ChainID:  "nonsense",
				Hash:     "8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D",
				Branch:   "8BADF00D",
				Contents: transactionContents,
			},
		},
	}

	return transactions, nil
}
