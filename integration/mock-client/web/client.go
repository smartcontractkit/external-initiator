package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/smartcontractkit/external-initiator/integration/mock-client/blockchain"
	"log"
	"net/http"
)

// RunWebserver starts a new web server using the access key
// and secret as provided on protected routes.
func RunWebserver() {
	srv := NewHTTPService()
	err := srv.Router.Run()
	if err != nil {
		fmt.Println(err)
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
	srv := HttpService{
		Router: gin.Default(),
	}
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
	r.GET("/xtz/monitor/heads/:chain_id", srv.HandleXtzMonitor)
	r.GET("/xtz/chains/main/blocks/:block_id/operations", srv.HandleXtzOperations)
	r.POST("/:platform", srv.HandleRpc)
	r.GET("/ws/:platform", srv.HandleWs)

	srv.Router = r
}

func (srv *HttpService) HandleXtzMonitor(c *gin.Context) {
	resp, err := blockchain.HandleXtzMonitorRequest(c.Param("chain_id"))

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (srv *HttpService) HandleXtzOperations(c *gin.Context) {
	resp, err := blockchain.HandleXtzOperationsRequest(c.Param("block_id"))

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateSubscription expects a CreateSubscriptionReq payload,
// validates the request and subscribes to the job.
func (srv *HttpService) HandleRpc(c *gin.Context) {
	var req blockchain.JsonrpcMessage
	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	resp, err := blockchain.HandleRequest("http", c.Param("platform"), req)
	if len(resp) == 0 || err != nil {
		var response blockchain.JsonrpcMessage
		response.ID = req.ID
		response.Version = req.Version
		if err != nil {
			log.Println(err)
			var errintf interface{}
			errintf = err.Error()
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
		log.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	defer conn.Close()

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var req blockchain.JsonrpcMessage
		err = json.Unmarshal(message, &req)
		if err != nil {
			log.Println("unmarshal:", err)
			continue
		}

		resp, err := blockchain.HandleRequest("ws", c.Param("platform"), req)
		if err != nil {
			log.Println("handle request:", err)
			continue
		}

		for _, msg := range resp {
			bz, err := json.Marshal(msg)
			if err != nil {
				log.Println("marshal:", err)
				continue
			}

			err = conn.WriteMessage(mt, bz)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}
