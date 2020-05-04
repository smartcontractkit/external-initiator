package chainlink

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	clMockUrl     = ""
	accessKey     = "abc"
	accessSecret  = "def"
	jobId         = "123"
	jobIdWPayload = "123payload"
	testPayload   = []byte(`{"somekey":"somevalue"}`)
)

func TestMain(m *testing.M) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.Header.Get(externalInitiatorAccessKeyHeader) != accessKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if r.Header.Get(externalInitiatorSecretHeader) != accessSecret {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if r.URL.Path == fmt.Sprintf("/v2/specs/%s/runs", jobIdWPayload) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !reflect.DeepEqual(body, testPayload) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else if r.URL.Path != fmt.Sprintf("/v2/specs/%s/runs", jobId) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		fmt.Println("created...")
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	clMockUrl = ts.URL

	code := m.Run()
	os.Exit(code)
}

func TestNode_TriggerJob(t *testing.T) {
	type fields struct {
		AccessKey    string
		AccessSecret string
		Endpoint     url.URL
	}
	type args struct {
		jobId   string
		payload []byte
	}

	u, err := url.Parse(clMockUrl)
	if err != nil {
		log.Fatal(err)
	}

	fakeU, err := url.Parse("http://fakeurl:6688/")
	if err != nil {
		log.Fatal(err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"is missing credentials",
			fields{
				Endpoint: *u,
			},
			args{jobId: jobId},
			true,
		},
		{
			"is missing access key",
			fields{
				AccessKey:    "",
				AccessSecret: accessSecret,
				Endpoint:     *u,
			},
			args{jobId: jobId},
			true,
		},
		{
			"is missing access secret",
			fields{
				AccessKey:    accessKey,
				AccessSecret: "",
				Endpoint:     *u,
			},
			args{jobId: jobId},
			true,
		},
		{
			"is missing job id",
			fields{
				AccessKey:    accessKey,
				AccessSecret: accessSecret,
				Endpoint:     *u,
			},
			args{jobId: ""},
			true,
		},
		{
			"does a successful POST request",
			fields{
				AccessKey:    accessKey,
				AccessSecret: accessSecret,
				Endpoint:     *u,
			},
			args{jobId: jobId},
			false,
		},
		{
			"cannot reach endpoint",
			fields{Endpoint: *fakeU},
			args{jobId: jobId},
			true,
		},
		{
			"does a successful POST request with payload",
			fields{
				AccessKey:    accessKey,
				AccessSecret: accessSecret,
				Endpoint:     *u,
			},
			args{jobId: jobIdWPayload, payload: testPayload},
			false,
		},
		{
			"does a POST request with invalid payload",
			fields{
				AccessKey:    accessKey,
				AccessSecret: accessSecret,
				Endpoint:     *u,
			},
			args{jobId: jobIdWPayload, payload: []byte(`weird payload`)},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := Node{
				AccessKey:    tt.fields.AccessKey,
				AccessSecret: tt.fields.AccessSecret,
				Endpoint:     tt.fields.Endpoint,
				Retry: RetryConfig{
					Timeout:  2 * time.Second,
					Attempts: 3,
					Delay:    100 * time.Millisecond,
				},
			}
			if err := cl.TriggerJob(tt.args.jobId, tt.args.payload); (err != nil) != tt.wantErr {
				t.Errorf("TriggerJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
