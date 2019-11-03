package client

import (
	"fmt"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	r.POST("/job", srv.CreateSubscription)
	srv.router = r
}

type CreateSubscriptionReq struct {
	JobID  string `json:"jobId"`
	Type   string `json:"type"`
	Params struct {
		Type   string `json:"type"`
		Config struct {
			Endpoint   string `json:"endpoint"`
			ChainId    string `json:"chainId"`
			RefreshInt int    `json:"refreshInterval"`
		} `json:"config"`
		Addresses []string `json:"addresses"`
		Topics    []string `json:"topics"`
	} `json:"params"`
}

func validateRequest(t *CreateSubscriptionReq) error {
	validations := []int{
		len(t.JobID),
		len(t.Params.Type),
	}

	switch t.Params.Type {
	case blockchain.ETH:
		validations = append(validations,
			len(t.Params.Addresses)+len(t.Params.Topics),
			len(t.Params.Config.Endpoint),
		)
	default:
		return errors.New("unknown blockchain")
	}

	for _, v := range validations {
		if v == 0 {
			return errors.New("missing required field(s)")
		}
	}

	if t.Params.Type == blockchain.ETH {
		_, err := url.Parse(t.Params.Config.Endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func newSubscriptionFromReq(arg *CreateSubscriptionReq) *store.Subscription {
	urlType := subscriber.RPC
	if strings.HasPrefix(arg.Params.Config.Endpoint, "ws") {
		urlType = subscriber.WS
	}

	sub := &store.Subscription{
		ReferenceId: uuid.New().String(),
		Job:         arg.JobID,
		Addresses:   arg.Params.Addresses,
		Topics:      arg.Params.Topics,
		Endpoint: store.Endpoint{
			Url:        arg.Params.Config.Endpoint,
			Type:       int(urlType),
			Blockchain: arg.Params.Type,
		},
		RefreshInt: arg.Params.Config.RefreshInt,
	}

	return sub
}

func (srv *httpService) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionReq
	type resp struct {
		ID string `json:"id"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := validateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	sub := newSubscriptionFromReq(&req)

	if err := srv.store.SaveSubscription(sub); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: sub.ReferenceId})
}

func (srv *httpService) ShowHealth(c *gin.Context) {
	c.JSON(200, gin.H{"chainlink": true})
}
