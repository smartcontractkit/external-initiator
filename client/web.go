package client

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/store"
	"log"
	"net/http"
)

type subscriptionStorer interface {
	SaveSubscription(sub *store.Subscription) error
	DeleteJob(jobid string) error
	GetEndpoint(name string) (*store.Endpoint, error)
	SaveEndpoint(endpoint *store.Endpoint) error
}

func runWebserver(
	store subscriptionStorer,
) {
	srv := newHTTPService(store)
	err := srv.router.Run()
	if err != nil {
		fmt.Println(err)
	}
}

type httpService struct {
	router *gin.Engine

	store subscriptionStorer
}

func newHTTPService(
	store subscriptionStorer,
) *httpService {
	srv := httpService{
		router: gin.Default(),
		store:  store,
	}
	srv.createRouter()
	return &srv
}

func (srv *httpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.router.ServeHTTP(w, r)
}

func (srv *httpService) createRouter() {
	r := gin.Default()
	r.GET("/health", srv.ShowHealth)
	r.POST("/jobs", srv.CreateSubscription)
	r.DELETE("/jobs/:jobid", srv.DeleteSubscription)
	r.POST("/config", srv.CreateEndpoint)

	srv.router = r
}

type CreateSubscriptionReq struct {
	JobID  string `json:"jobId"`
	Type   string `json:"type"`
	Params struct {
		Endpoint  string   `json:"endpoint"`
		Addresses []string `json:"addresses"`
		Topics    []string `json:"eventTopics"`
	} `json:"params"`
}

func validateRequest(t *CreateSubscriptionReq, endpointType string) error {
	validations := []int{
		len(t.JobID),
	}

	switch endpointType {
	case blockchain.ETH:
		validations = append(validations,
			len(t.Params.Addresses)+len(t.Params.Topics),
		)
	}

	for _, v := range validations {
		if v == 0 {
			return errors.New("missing required field(s)")
		}
	}

	return nil
}

type resp struct {
	ID string `json:"id"`
}

func (srv *httpService) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionReq

	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	endpoint, err := srv.store.GetEndpoint(req.Params.Endpoint)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	if endpoint == nil {
		log.Println("unknown endpoint provided")
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := validateRequest(&req, endpoint.Type); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	sub := &store.Subscription{
		ReferenceId:  uuid.New().String(),
		Job:          req.JobID,
		EndpointName: req.Params.Endpoint,
	}

	switch endpoint.Type {
	case blockchain.ETH:
		sub.Ethereum = store.EthSubscription{
			Addresses: req.Params.Addresses,
			Topics:    req.Params.Topics,
		}
	}

	if err := srv.store.SaveSubscription(sub); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: sub.ReferenceId})
}

func (srv *httpService) DeleteSubscription(c *gin.Context) {
	jobid := c.Param("jobid")
	if err := srv.store.DeleteJob(jobid); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, resp{ID: jobid})
}

func (srv *httpService) ShowHealth(c *gin.Context) {
	c.JSON(200, gin.H{"chainlink": true})
}

func (srv *httpService) CreateEndpoint(c *gin.Context) {
	var config store.Endpoint
	err := c.BindJSON(&config)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := srv.store.SaveEndpoint(&config); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, resp{ID: config.Name})
}
