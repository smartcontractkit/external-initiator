package chainlink

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	externalInitiatorAccessKeyHeader = "X-Chainlink-EA-AccessKey"
	externalInitiatorSecretHeader    = "X-Chainlink-EA-Secret"
)

type Node struct {
	AccessKey    string
	AccessSecret string
	Endpoint     url.URL
}

func (cl Node) TriggerJob(jobId string) error {
	fmt.Printf("Sending a job run trigger to %s for job %s\n", cl.Endpoint.String(), jobId)

	u := cl.Endpoint
	u.Path = fmt.Sprintf("/v2/specs/%s/runs", jobId)

	request, err := http.NewRequest(http.MethodPost, u.String(), nil)
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

	if r.StatusCode > 201 {
		return fmt.Errorf("received faulty status code: %d", r.StatusCode)
	}

	return nil
}
