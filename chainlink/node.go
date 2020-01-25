// Package chainlink implements functions to interact
// with a Chainlink node.
package chainlink

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

const (
	externalInitiatorAccessKeyHeader = "X-Chainlink-EA-AccessKey"
	externalInitiatorSecretHeader    = "X-Chainlink-EA-Secret"
)

// Node encapsulates all the configuration
// necessary to interact with a Chainlink node.
type Node struct {
	AccessKey    string
	AccessSecret string
	Endpoint     url.URL
}

// TriggerJob wil send a job run trigger for the
// provided jobId.
func (cl Node) TriggerJob(jobId string, data []byte) error {
	fmt.Printf("Sending a job run trigger to %s for job %s\n", cl.Endpoint.String(), jobId)

	u := cl.Endpoint
	u.Path = fmt.Sprintf("/v2/specs/%s/runs", jobId)

	request, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Add(externalInitiatorAccessKeyHeader, cl.AccessKey)
	request.Header.Add(externalInitiatorSecretHeader, cl.AccessSecret)

	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		return fmt.Errorf("received faulty status code: %d", r.StatusCode)
	}

	return nil
}
