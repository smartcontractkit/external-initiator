package conflux

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/blockchain/common"
)

func Test_parseEthLogResponse(t *testing.T) {
	type args struct {
		result json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    common.RunlogRequest
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEthLogResponse(tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEthLogResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEthLogResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEthLogsResponse(t *testing.T) {
	type args struct {
		result json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    []common.RunlogRequest
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEthLogsResponse(tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEthLogsResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEthLogsResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runlogManager_CreateJobRun(t *testing.T) {
	type fields struct {
		manager *manager
	}
	type args struct {
		request common.RunlogRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := runlogManager{
				manager: tt.fields.manager,
			}
			if got := rm.CreateJobRun(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateJobRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runlogManager_getFilterQuery(t *testing.T) {
	type fields struct {
		manager *manager
	}
	type args struct {
		fromBlock string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := runlogManager{
				manager: tt.fields.manager,
			}
			got, err := rm.getFilterQuery(tt.args.fromBlock)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFilterQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFilterQuery() got = %v, want %v", got, tt.want)
			}
		})
	}
}
