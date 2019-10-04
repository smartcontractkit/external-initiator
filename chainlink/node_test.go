package chainlink

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var (
	clMockUrl    = ""
	accessKey    = "abc"
	accessSecret = "def"
	jobId        = "123"
)

func TestMain(m *testing.M) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != fmt.Sprintf("/v2/specs/%s/runs", jobId) {
			w.WriteHeader(http.StatusNotFound)
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
		jobId string
	}

	u, err := url.Parse(clMockUrl)
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
			"does a successfuly POST request",
			fields{
				AccessKey:    accessKey,
				AccessSecret: accessSecret,
				Endpoint:     *u,
			},
			args{jobId: jobId},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := Node{
				AccessKey:    tt.fields.AccessKey,
				AccessSecret: tt.fields.AccessSecret,
				Endpoint:     tt.fields.Endpoint,
			}
			if err := cl.TriggerJob(tt.args.jobId); (err != nil) != tt.wantErr {
				t.Errorf("TriggerJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
