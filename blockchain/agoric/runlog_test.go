package agoric

import (
	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"reflect"
	"testing"
)

func TestRunlogManager_ParseRequests(t *testing.T) {
	type fields struct {
		jobid string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    common.RunlogRequest
		wantErr bool
	}{
		{
			"fails parsing invalid payload",
			fields{},
			args{data: []byte(`invalid`)},
			nil,
			true,
		},
		{
			"fails parsing invalid WS body",
			fields{},
			args{data: []byte(`{}`)},
			nil,
			true,
		},
		{
			"fails parsing invalid WS type",
			fields{},
			args{data: []byte(`{"type":"oracleServer/wrongType"}`)},
			nil,
			true,
		},
		{
			"successfully parses WS Oracle request",
			fields{jobid: "9999"},
			args{data: []byte(`{"type":"oracleServer/onQuery","data":{"query":{"jobID":"9999","params":{"path":"foo"}},"queryId":"123","fee":191919}}`)},
			common.RunlogRequest{
				"path":       "foo",
				"payment":    "191919000000000000",
				"request_id": "123",
			},
			false,
		},
		{
			"skips unfiltered WS Oracle request",
			fields{jobid: "Z9999"},
			args{data: []byte(`{"type":"oracleServer/onQuery","data":{"query":{"jobID":"9999","params":{"path":"foo"}},"queryId":"123","fee":191919}}`)},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := runlogManager{
				manager: &manager{
					jobid: tt.fields.jobid,
				},
			}
			got, err := e.parseRequests(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRequests() got = %s, want %s", got, tt.want)
			}
		})
	}
}
