package substrate

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
)

func MustDecodeHex(s string) []byte {
	data, _ := hex.DecodeString(s)
	return data
}

func Test_getChanges(t *testing.T) {
	type args struct {
		key  types.StorageKey
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []types.KeyValueOption
		wantErr bool
	}{
		{
			name: "gets a value",
			args: args{
				key:  types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607")),
				data: []byte(`{"result":{"block":"0x0400bf57e54d9043203d8b8766dca498e7a86b48b048fe65c7e6a8750532e5ce","changes":[["0x1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607","0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d"]]},"subscription":"Vhdet0j80btCi4nz"}`),
			},
			want: []types.KeyValueOption{
				{
					StorageKey:     types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607")),
					HasStorageData: true,
					StorageData:    types.NewStorageDataRaw(MustDecodeHex("d43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d")),
				},
			},
			wantErr: false,
		},
		{
			name: "gets a result with no change",
			args: args{
				key:  types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d33c112c4b12e2be4c71e6501a2d50d7e40")),
				data: []byte(`{"result":{"block":"0x6374c6e27838a2edff58550a82d43b7e0f209b30eabc455acbf1284bc4ec129f","changes":[["0x1ae70894dea2956c24d9c19ac4d15d33c112c4b12e2be4c71e6501a2d50d7e40",null]]},"subscription":"T2EcUQJkmaR3YGL0"}`),
			},
			want: []types.KeyValueOption{
				{
					StorageKey:     types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d33c112c4b12e2be4c71e6501a2d50d7e40")),
					HasStorageData: true,
					StorageData:    types.NewStorageDataRaw([]byte{}),
				},
			},
			wantErr: false,
		},
		{
			name: "only gets the correct storage key",
			args: args{
				key:  types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607")),
				data: []byte(`{"result":{"block":"0x0400bf57e54d9043203d8b8766dca498e7a86b48b048fe65c7e6a8750532e5ce","changes":[["0x1ae70894dea2956c24d9c19ac4d15d33c112c4b12e2be4c71e6501a2d50d7e40","0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27e"],["0x1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607","0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d"]]},"subscription":"Vhdet0j80btCi4nz"}`),
			},
			want: []types.KeyValueOption{
				{
					StorageKey:     types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607")),
					HasStorageData: true,
					StorageData:    types.NewStorageDataRaw(MustDecodeHex("d43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d")),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getChanges(tt.args.key, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("getChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getChanges() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseChange(t *testing.T) {
	var accountId types.AccountID
	var feedId FeedId
	var feedConfig FeedConfig
	var round Round

	type args struct {
		key  types.StorageKey
		data []byte
		t    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "parses account id",
			args: args{
				key:  types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607")),
				data: []byte(`{"result":{"block":"0x0400bf57e54d9043203d8b8766dca498e7a86b48b048fe65c7e6a8750532e5ce","changes":[["0x1ae70894dea2956c24d9c19ac4d15d334f9828de31b1944b243d7cfa8f863607","0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d"]]},"subscription":"Vhdet0j80btCi4nz"}`),
				t:    &accountId,
			},
			wantErr: false,
		},
		{
			name: "parses feed id",
			args: args{
				key:  types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d33c112c4b12e2be4c71e6501a2d50d7e40")),
				data: []byte(`{"result":{"block":"0x0109640cb073a8c704938224424bf127b851ddaefe9a47b9acd2557c30eaec14","changes":[["0x1ae70894dea2956c24d9c19ac4d15d33c112c4b12e2be4c71e6501a2d50d7e40","0x01000000"]]},"subscription":"GK6frSf4qIDtxmPl"}`),
				t:    &feedId,
			},
			wantErr: false,
		},
		{
			name: "parses feed config",
			args: args{
				key:  types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d33407d94c38db8ad2e4aa005c8942ac2c5b4def25cfda6ef3a00000000")),
				data: []byte(`{"result":{"block":"0xc0b664cdf2c0e0f49aeddd0f3aa272c49310d5a75efce39c0e0c9ee977ae8bec","changes":[["0x1ae70894dea2956c24d9c19ac4d15d33407d94c38db8ad2e4aa005c8942ac2c5b4def25cfda6ef3a00000000","0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d0000000000000000000000000000000000ffc99a3b000000000000000000000000010000000100000000e876481700000000000000000000000100000002080123000000000600000006000000010100000001000000"]]},"subscription":"gVK1dqmAOifBAKrw"}`),
				t:    &feedConfig,
			},
			wantErr: false,
		},
		{
			name: "parses round data",
			args: args{
				key:  types.NewStorageKey(MustDecodeHex("1ae70894dea2956c24d9c19ac4d15d337b45e7c782afb68af72fc5ebb26b974bb4def25cfda6ef3a000000005153cb1f00942ff401000000")),
				data: []byte(`{"result":{"block":"0x7eb1c3626a76510323f0f8174881de62761e06b668b506631a4ee73a9f051560","changes":[["0x1ae70894dea2956c24d9c19ac4d15d337b45e7c782afb68af72fc5ebb26b974bb4def25cfda6ef3a000000005153cb1f00942ff401000000","0xc102000001d204000000000000000000000000000001c10200000101000000"]]},"subscription":"GRMvxdpwlBiTjbUa"}`),
				t:    &round,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseChange(tt.args.key, tt.args.data, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("parseChange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
