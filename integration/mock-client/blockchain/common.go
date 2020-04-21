package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

type RpcMockRequestMetadata struct {
	Path       string          `json:"path"`
	Code       int             `json:"code"`
	Result     json.RawMessage `json:"result"`
	RpcRequest JsonrpcMessage  `json:"rpc_request"`
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

func SetRpcRoutesFromJSON(routerGroup *gin.RouterGroup) error {
	mockResponsesPath := os.Getenv("MOCK_RESPONSES_DIR")
	if len(mockResponsesPath) == 0 {
		wd, _ := os.Getwd()
		mockResponsesPath = path.Join(wd, "mock-responses")
	}

	rpcResponsesPath := mockResponsesPath + "/rpc"
	files, err := ioutil.ReadDir(rpcResponsesPath)

	if err != nil {
		return err
	}

	// responses[path][rpc_method][params] = response
	responses := make(map[string]map[string]map[string]RpcMockRequestMetadata)

	for _, f := range files {
		path := path.Join(rpcResponsesPath, f.Name())
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)

		var resp RpcMockRequestMetadata
		err = json.Unmarshal(data, &resp)
		if err != nil {
			return err
		}

		normalizedParams := NormalizeParams(resp.RpcRequest.Params)
		method := resp.RpcRequest.Method

		if _, ok := responses[resp.Path]; !ok {
			pathMap := make(map[string]map[string]RpcMockRequestMetadata)
			responses[resp.Path] = pathMap
		}
		if _, ok := responses[resp.Path][method]; !ok {
			methodMap := make(map[string]RpcMockRequestMetadata)
			responses[resp.Path][method] = methodMap
		}
		responses[resp.Path][method][normalizedParams] = resp
	}

	for path, respMap := range responses {
		routerGroup.Handle("POST", path, func(c *gin.Context) {
			var req JsonrpcMessage

			if err := c.BindJSON(&req); err != nil {
				log.Println(err)
				c.JSON(http.StatusBadRequest, nil)
				return
			}

			normalizedParams := NormalizeParams(req.Params)
			resp, ok := respMap[req.Method][normalizedParams]

			if ok {
				rpcResp := JsonrpcMessage{
					ID:      req.ID,
					Version: req.Version,
					Method:  req.Method,
					Result:  resp.Result,
				}
				c.JSON(resp.Code, rpcResp)
			} else {
				c.JSON(http.StatusBadRequest, "Expected to find matching mock response for request")
			}
		})
	}
	return nil
}

func NormalizeParams(params json.RawMessage) string {
	// TODO!
	return "param"
}
