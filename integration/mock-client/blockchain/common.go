package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

type JsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *interface{}    `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func HandleRequest(conn, platform string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch platform {
	case "eth":
		return HandleEthRequest(conn, msg)
	default:
		return nil, errors.New(fmt.Sprint("unexpected platform: ", platform))
	}
}

type MockResponse struct {
	Path   string          `json:"path"`
	Method string          `json:"method"`
	Code   int             `json:"code"`
	Body   json.RawMessage `json:"body"`
}

func SetHttpRoutesFromJSON(routerGroup *gin.RouterGroup) error {
	mockResponsesPath := os.Getenv("MOCK_RESPONSES_DIR")
	if len(mockResponsesPath) == 0 {
		wd, _ := os.Getwd()
		mockResponsesPath = path.Join(wd, "mock-responses")
	}

	httpResponsesPath := mockResponsesPath + "/http"
	files, err := ioutil.ReadDir(httpResponsesPath)

	if err != nil {
		return err
	}

	for _, f := range files {
		path := path.Join(httpResponsesPath, f.Name())
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)

		var resp MockResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			return err
		}
		routerGroup.Handle(strings.ToUpper(resp.Method), resp.Path, func(c *gin.Context) {
			c.JSON(resp.Code, resp.Body)
		})
	}

	return nil
}
