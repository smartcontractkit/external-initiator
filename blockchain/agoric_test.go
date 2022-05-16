package blockchain

import (
	"errors"
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

func TestCreateAgoricFilterMessage(t *testing.T) {
	tests := []struct {
		name string
		args store.AgoricSubscription
		p    subscriber.Type
		want []byte
		err  error
	}{
		{
			"empty",
			store.AgoricSubscription{},
			subscriber.WS,
			nil,
			nil,
		},
		{
			"address only",
			store.AgoricSubscription{},
			subscriber.WS,
			nil,
			nil,
		},
		{
			"empty RPC",
			store.AgoricSubscription{},
			subscriber.RPC,
			nil,
			errors.New("only WS connections are allowed for Agoric"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := createAgoricManager(tt.p, store.Subscription{Agoric: tt.args})
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("createAgoricManager.err = %s, want %s", err, tt.err)
			}
			if err == nil {
				if got := mgr.GetTriggerJson(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
				}
			}
		})
	}

	t.Run("has invalid filter query", func(t *testing.T) {
		got := agoricManager{filter: agoricFilter{JobID: "1919"}}.GetTriggerJson()
		if got != nil {
			t.Errorf("GetTriggerJson() = %s, want nil", got)
		}
	})
}

func TestAgoricManager_GetTestJson(t *testing.T) {
	type fields struct {
		filter agoricFilter
		p      subscriber.Type
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			"returns empty when using RPC",
			fields{
				p: subscriber.RPC,
			},
			nil,
		},
		{
			"returns empty when using WS",
			fields{
				p: subscriber.WS,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := agoricManager{
				filter: tt.fields.filter,
			}
			if got := e.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgoricManager_ParseTestResponse(t *testing.T) {
	type fields struct {
		f agoricFilter
		p subscriber.Type
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"does nothing for WS",
			fields{f: agoricFilter{}, p: subscriber.WS},
			args{},
			false,
		},
		{
			"parses RPC responses",
			fields{f: agoricFilter{}, p: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`)},
			false,
		},
		{
			"fails unmarshal payload",
			fields{f: agoricFilter{}, p: subscriber.RPC},
			args{[]byte(`error`)},
			false,
		},
		{
			"fails unmarshal result",
			fields{f: agoricFilter{}, p: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := agoricManager{
				filter: tt.fields.f,
			}
			if err := e.ParseTestResponse(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgoricManager_ParseResponse(t *testing.T) {
	type fields struct {
		filter agoricFilter
		p      subscriber.Type
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []subscriber.Event
		want1  bool
	}{
		{
			"fails parsing invalid payload",
			fields{filter: agoricFilter{}, p: subscriber.WS},
			args{data: []byte(`invalid`)},
			nil,
			false,
		},
		{
			"fails parsing invalid WS body",
			fields{filter: agoricFilter{}, p: subscriber.WS},
			args{data: []byte(`{}`)},
			nil,
			false,
		},
		{
			"fails parsing invalid WS type",
			fields{filter: agoricFilter{}, p: subscriber.WS},
			args{data: []byte(`{"type":"oracleServer/wrongType"}`)},
			nil,
			false,
		},
		{
			"successfully parses WS Oracle request",
			fields{filter: agoricFilter{JobID: "9999"}, p: subscriber.WS},
			args{data: []byte(`{"type":"oracleServer/onQuery","data":{"query":{"jobID":"9999","params":{"path":"foo"}},"queryId":"123","fee":"191919000000000000000"}}`)},
			[]subscriber.Event{[]byte(`{"path":"foo","payment":"191919000000000000000","request_id":"123"}`)},
			true,
		},
		{
			"skips unfiltered WS Oracle request",
			fields{filter: agoricFilter{JobID: "Z9999"}, p: subscriber.WS},
			args{data: []byte(`{"type":"oracleServer/onQuery","data":{"query":{"jobID":"9999","params":{"path":"foo"}},"queryId":"123","fee":"191919"}}`)},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := agoricManager{
				filter: tt.fields.filter,
			}
			got, got1 := e.ParseResponse(tt.args.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseResponse() got = %s, want %s", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseResponse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
