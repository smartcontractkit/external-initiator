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
	srv.router = r
}

type CreateSubscriptionReq struct {
	JobID  string `json:"jobId"`
	Type   string `json:"type"`
	Params struct {
		Type      string   `json:"type"`
		Endpoint  string   `json:"endpoint"`
		Addresses []string `json:"addresses"`
		Topics    []string `json:"initiatorTopics"`
	} `json:"params"`
}

func validateRequest(t *CreateSubscriptionReq) error {
	validations := []int{
		len(t.JobID),
		len(t.Params.Type),
		len(t.Params.Endpoint),
	}

	switch t.Params.Type {
	case blockchain.ETH:
		validations = append(validations,
			len(t.Params.Addresses)+len(t.Params.Topics),
		)
	default:
		return errors.New("unknown blockchain")
	}

	for _, v := range validations {
		if v == 0 {
			return errors.New("missing required field(s)")
		}
	}

	return nil
}

func (srv *httpService) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionReq
	type resp struct {
		ID string `json:"id"`
	}

	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := validateRequest(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	sub := &store.Subscription{
		ReferenceId:  uuid.New().String(),
		Job:          req.JobID,
		Addresses:    req.Params.Addresses,
		Topics:       req.Params.Topics,
		EndpointName: req.Params.Endpoint,
	}

	if err := srv.store.SaveSubscription(sub); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: sub.ReferenceId})
}

func (srv *httpService) ShowHealth(c *gin.Context) {
	c.JSON(200, gin.H{"chainlink": true})
}
