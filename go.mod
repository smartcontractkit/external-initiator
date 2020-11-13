module github.com/smartcontractkit/external-initiator

go 1.12

require (
	github.com/FactomProject/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/centrifuge/go-substrate-rpc-client v0.0.4-0.20200117100327-4dc63dc6b2e6
	github.com/cmars/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/ethereum/go-ethereum v1.9.12
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a
	github.com/gin-gonic/gin v1.6.0
	github.com/golang/mock v1.4.4
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/iotexproject/iotex-proto v0.4.3
	github.com/irisnet/service-sdk-go v0.0.0-20201030091855-7f57f83f8c6c
	github.com/jinzhu/gorm v1.9.11
	github.com/magiconair/properties v1.8.1
	github.com/ontio/ontology-go-sdk v1.11.1
	github.com/pierrec/xxHash v0.1.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/smartcontractkit/chainlink v0.8.2
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tendermint v0.33.4
	github.com/tidwall/gjson v1.6.0
	go.uber.org/zap v1.14.1
	golang.org/x/tools v0.0.0-20200513201620-d5fe73897c97 // indirect
	google.golang.org/grpc v1.31.1
	gopkg.in/gormigrate.v1 v1.6.0
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
	github.com/tendermint/tendermint => github.com/bianjieai/tendermint v0.34.0-irita-200930
	launchpad.net/gocheck => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
)
