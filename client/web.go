package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/keeper"
	"github.com/smartcontractkit/external-initiator/store"
)

const (
	externalInitiatorAccessKeyHeader = "X-Chainlink-EA-AccessKey"
	externalInitiatorSecretHeader    = "X-Chainlink-EA-Secret"
)

type subscriptionStorer interface {
	SaveSubscription(sub *store.Subscription) error
	DeleteJob(jobid string) error
	GetEndpoint(name string) (*store.Endpoint, error)
	SaveEndpoint(endpoint *store.Endpoint) error
	DB() *gorm.DB
}

// RunWebserver starts a new web server using the access key
// and secret as provided on protected routes.
func RunWebserver(
	accessKey, secret string,
	store subscriptionStorer,
	port int,
) {
	srv := NewHTTPService(accessKey, secret, store)
	addr := fmt.Sprintf(":%v", port)
	err := srv.Router.Run(addr)
	if err != nil {
		logger.Error(err)
	}
}

// HttpService encapsulates router, EI service
// and access credentials.
type HttpService struct {
	Router    *gin.Engine
	AccessKey string
	Secret    string
	Store     subscriptionStorer
}

// NewHTTPService creates a new HttpService instance
// with the default router.
func NewHTTPService(
	accessKey, secret string,
	store subscriptionStorer,
) *HttpService {
	srv := HttpService{
		AccessKey: accessKey,
		Secret:    secret,
		Store:     store,
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
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggerFunc())
	r.GET("/health", srv.ShowHealth)

	auth := r.Group("/")
	auth.Use(authenticate(srv.AccessKey, srv.Secret))
	{
		auth.POST("/jobs", srv.CreateSubscription)
		auth.DELETE("/jobs/:jobid", srv.DeleteSubscription)
		auth.POST("/config", srv.CreateEndpoint)
	}

	srv.Router = r
}

func authenticate(accessKey, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqAccessKey := c.GetHeader(externalInitiatorAccessKeyHeader)
		reqSecret := c.GetHeader(externalInitiatorSecretHeader)
		if reqAccessKey == accessKey && reqSecret == secret {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// CreateSubscriptionReq holds the payload expected for job POSTs
// from the Chainlink node.
type CreateSubscriptionReq struct {
	JobID  string            `json:"jobId"`
	Type   string            `json:"type"`
	Params blockchain.Params `json:"params"`
}

func validateRequest(t *CreateSubscriptionReq, endpointType string) error {
	validations := append([]int{
		len(t.JobID),
	}, blockchain.GetValidations(endpointType, t.Params)...)

	for _, v := range validations {
		if v < 1 {
			return errors.New("missing required field(s)")
		}
	}

	return nil
}

type resp struct {
	ID string `json:"id"`
}

// CreateSubscription expects a CreateSubscriptionReq payload,
// validates the request and subscribes to the job.
func (srv *HttpService) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionReq

	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	// HACK - making an exception to the normal workflow for keepers
	// since they will be removed from EI at a later date
	if req.Params.Endpoint == "keeper" {
		srv.createKeeperSubscription(req, c)
		return
	}

	endpoint, err := srv.Store.GetEndpoint(req.Params.Endpoint)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	if endpoint == nil {
		logger.Error("unknown endpoint provided")
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := validateRequest(&req, endpoint.Type); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	sub := &store.Subscription{
		ReferenceId:  uuid.New().String(),
		Job:          req.JobID,
		EndpointName: req.Params.Endpoint,
		Endpoint:     *endpoint,
	}

	blockchain.CreateSubscription(sub, req.Params)

	if err := srv.Store.SaveSubscription(sub); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: sub.ReferenceId})
}

// DeleteSubscription deletes any job with the jobid
// provided as parameter in the request.
func (srv *HttpService) DeleteSubscription(c *gin.Context) {
	jobid := c.Param("jobid")
	if err := srv.Store.DeleteJob(jobid); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, resp{ID: jobid})
}

// ShowHealth returns the following when online:
//  {"chainlink": true}
func (srv *HttpService) ShowHealth(c *gin.Context) {
	c.JSON(200, gin.H{"chainlink": true})
}

// CreateEndpoint saves the endpoint configuration provided
// as payload.
func (srv *HttpService) CreateEndpoint(c *gin.Context) {
	var config store.Endpoint
	err := c.BindJSON(&config)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := srv.Store.SaveEndpoint(&config); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: config.Name})
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

func (srv *HttpService) createKeeperSubscription(req CreateSubscriptionReq, c *gin.Context) {
	if err := validateKeeperRequest(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	jobID, err := models.NewIDFromString(req.JobID)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	address := common.HexToAddress(req.Params.Address)
	from := common.HexToAddress(req.Params.From)
	reg := keeper.NewRegistry(address, from, jobID)
	err = srv.Store.DB().Create(&reg).Error
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: reg.ReferenceID})
}

func validateKeeperRequest(req *CreateSubscriptionReq) error {
	if req.Params.Address == "" || req.Params.From == "" || req.JobID == "" {
		return errors.New("missing required fields")
	}
	return nil
}
