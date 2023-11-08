// Package chainlink implements functions to interact
// with a Chainlink node.
package chainlink

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/avast/retry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/smartcontractkit/chainlink/core/logger"
)

var (
	promJobrunsTotalErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ei_jobruns_total_errored_attempts",
		Help: "The total number of jobrun trigger attempts that errored",
	})
	promJobrunsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ei_jobruns_failed",
		Help: "The number of failed jobruns",
	}, []string{"jobid"})
	promJobrunsSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ei_jobruns_success",
		Help: "The number of successful jobruns",
	}, []string{"jobid"})
)

const (
	externalInitiatorAccessKeyHeader = "X-Chainlink-EA-AccessKey"
	externalInitiatorSecretHeader    = "X-Chainlink-EA-Secret"
)

type RetryConfig struct {
	Timeout  time.Duration
	Attempts uint
	Delay    time.Duration
}

// Node encapsulates all the configuration
// necessary to interact with a Chainlink node.
type Node struct {
	AccessKey    string
	AccessSecret string
	Endpoint     url.URL
	Retry        RetryConfig
}

// TriggerJob wil send a job run trigger for the
// provided jobId.
func (cl Node) TriggerJob(jobId string, data []byte) error {
	logger.Infof("Sending a job run trigger to %s for job %s\n", cl.Endpoint.String(), jobId)

	err := cl.sendJobrunTrigger(jobId, data)
	if err != nil {
		promJobrunsFailed.With(prometheus.Labels{"jobid": jobId}).Inc()
		return err
	}

	promJobrunsSuccess.With(prometheus.Labels{"jobid": jobId}).Inc()
	return nil
}

func (cl Node) sendJobrunTrigger(jobId string, data []byte) error {
	u := cl.Endpoint
	u.Path = fmt.Sprintf("/v2/jobs/%s/runs", jobId)

	request, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Add(externalInitiatorAccessKeyHeader, cl.AccessKey)
	request.Header.Add(externalInitiatorSecretHeader, cl.AccessSecret)

	_, statusCode, err := cl.Retry.withRetry(&http.Client{}, request)
	if err != nil {
		return err
	}

	if statusCode >= 400 {
		return fmt.Errorf("received faulty status code: %v", statusCode)
	}

	return nil
}

func (config RetryConfig) withRetry(client *http.Client, request *http.Request) (responseBody []byte, statusCode int, err error) {
	err = retry.Do(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
			defer cancel()
			requestWithTimeout := request.Clone(ctx)

			start := time.Now()

			r, e := client.Do(requestWithTimeout)
			if e != nil {
				logger.Errorf("job run trigger error making request: %v", e.Error())
				return e
			}
			defer logger.ErrorIfCalling(r.Body.Close)
			statusCode = r.StatusCode
			elapsed := time.Since(start)
			logger.Debugw(fmt.Sprintf("job run trigger got %v in %s", statusCode, elapsed), "statusCode", statusCode, "timeElapsedSeconds", elapsed)

			bz, e := ioutil.ReadAll(r.Body)
			if e != nil {
				logger.Errorf("job run trigger error reading body: %v", err.Error())
				return e
			}
			elapsed = time.Since(start)
			logger.Debugw(fmt.Sprintf("job run trigger finished after %s", elapsed), "statusCode", statusCode, "timeElapsedSeconds", elapsed)

			responseBody = bz

			// Retry on 5xx since this might give a different result
			if 500 <= r.StatusCode && r.StatusCode < 600 {
				e = fmt.Errorf("remote server error: %v\nResponse body: %v", r.StatusCode, string(responseBody))
				logger.Error(e)
				return e
			}

			return nil
		},
		retry.Delay(config.Delay),
		retry.Attempts(config.Attempts),
		retry.OnRetry(func(n uint, err error) {
			promJobrunsTotalErrors.Inc()
			logger.Debugw("job run trigger error, will retry", "error", err.Error(), "attempt", n, "timeout", config.Timeout)
		}),
	)
	if err != nil {
		promJobrunsTotalErrors.Inc()
	}

	return responseBody, statusCode, err
}
