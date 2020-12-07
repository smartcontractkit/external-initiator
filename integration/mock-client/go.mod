module github.com/smartcontractkit/external-initiator/integration/mock-client

go 1.14

require (
	github.com/centrifuge/go-substrate-rpc-client v1.1.0
	github.com/ethereum/go-ethereum v1.9.22
	github.com/gin-gonic/gin v1.6.0
	github.com/gorilla/websocket v1.4.2
	github.com/iotexproject/iotex-proto v0.4.3
	github.com/irisnet/service-sdk-go v0.0.0-20201030091855-7f57f83f8c6c
	github.com/smartcontractkit/chainlink v0.9.5
	github.com/smartcontractkit/external-initiator v0.0.0-20200710101835-de7d82ec7e0c
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tendermint v0.33.4
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.31.1
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
	// Useful for local development (TODO: comment out when not needed)
	github.com/smartcontractkit/external-initiator v0.0.0-20200710101835-de7d82ec7e0c => ../../
	github.com/tendermint/tendermint => github.com/bianjieai/tendermint v0.34.0-irita-200930
)
