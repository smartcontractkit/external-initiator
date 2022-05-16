package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/integration/mock-client/blockchain"
)

// RunWebserver starts a new web server using the access key
// and secret as provided on protected routes.
func RunWebserver() {
	srv := NewHTTPService()
	err := srv.Router.Run(":8080")
	if err != nil {
		logger.Error(err)
	}
}

// HttpService encapsulates router, EI service
// and access credentials.
type HttpService struct {
	Router *gin.Engine
}

// NewHTTPService creates a new HttpService instance
// with the default router.
func NewHTTPService() *HttpService {
	srv := HttpService{}
	srv.createRouter()
	return &srv
}

// ServeHTTP calls ServeHTTP on the underlying router,
// which conforms to the http.Handler interface.
func (srv *HttpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.Router.ServeHTTP(w, r)
}

func (srv *HttpService) createRouter() {
	r := gin.Default()
	r.Use(gin.Recovery(), loggerFunc())

	blockchain.SetHttpRoutes(r)
	r.GET("/ws/:platform", srv.HandleWs)
	r.POST("/rpc/:platform", srv.HandleRpc)

	srv.Router = r
}

// CreateSubscription expects a CreateSubscriptionReq payload,
// validates the request and subscribes to the job.
func (srv *HttpService) HandleRpc(c *gin.Context) {
	var req blockchain.JsonrpcMessage
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	resp, err := blockchain.HandleRequest("rpc", c.Param("platform"), req)
	if len(resp) == 0 || err != nil {
		var response blockchain.JsonrpcMessage
		response.ID = req.ID
		response.Version = req.Version
		if err != nil {
			logger.Error(err)
			errintf := interface{}(err.Error())
			response.Error = &errintf
		}
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp[0])
}

var upgrader = websocket.Upgrader{}

func (srv *HttpService) HandleWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	defer logger.ErrorIfCalling(conn.Close)

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			logger.Error("read:", err)
			break
		}

		var req blockchain.JsonrpcMessage
		err = json.Unmarshal(message, &req)
		if err != nil {
			logger.Error("unmarshal:", err)
			continue
		}

		resp, err := blockchain.HandleRequest("ws", c.Param("platform"), req)
		if err != nil {
			logger.Error("handle request:", err)
			continue
		}

		for _, msg := range resp {
			bz, err := json.Marshal(msg)
			if err != nil {
				logger.Error("marshal:", err)
				continue
			}

			err = conn.WriteMessage(mt, bz)
			if err != nil {
				logger.Error("write:", err)
				break
			}
		}
	}
}

// Inspired by https://github.com/gin-gonic/gin/issues/961
func loggerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("Web request log error: ", err.Error())
			// Implicitly relies on limits.RequestSizeLimiter
			// overriding of c.Request.Body to abort gin's Context
			// inside ioutil.ReadAll.
			// Functions as we would like, but horrible from an architecture
			// and design pattern perspective.
			if !c.IsAborted() {
				c.AbortWithStatus(http.StatusBadRequest)
			}
			return
		}
		rdr := bytes.NewBuffer(buf)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

		start := time.Now()
		c.Next()
		end := time.Now()

		logger.Infow(fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			"method", c.Request.Method,
			"status", c.Writer.Status(),
			"path", c.Request.URL.Path,
			"query", c.Request.URL.Query(),
			"body", readBody(rdr),
			"clientIP", c.ClientIP(),
			"errors", c.Errors.String(),
			"servedAt", end.Format("2006-01-02 15:04:05"),
			"latency", fmt.Sprint(end.Sub(start)),
		)
	}
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		logger.Warn("unable to read from body for sanitization: ", err)
		return "*FAILED TO READ BODY*"
	}

	if buf.Len() == 0 {
		return ""
	}

	s, err := readSanitizedJSON(buf)
	if err != nil {
		logger.Warn("unable to sanitize json for logging: ", err)
		return "*FAILED TO READ BODY*"
	}
	return s
}

func readSanitizedJSON(buf *bytes.Buffer) (string, error) {
	var dst map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &dst)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(dst)
	if err != nil {
		return "", err
	}
	return string(b), err
}
