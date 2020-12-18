module github.com/smartcontractkit/external-initiator/integration/mock-client

go 1.15

require (
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/centrifuge/go-substrate-rpc-client v2.0.0+incompatible
	github.com/elastic/gosigar v0.10.4 // indirect
	github.com/ethereum/go-ethereum v1.9.24
	github.com/gin-gonic/gin v1.6.0
	github.com/gorilla/websocket v1.4.2
	github.com/iotexproject/iotex-proto v0.4.3
	github.com/smartcontractkit/chainlink v0.9.5-0.20201214122441-66aaea171293
	github.com/smartcontractkit/external-initiator v0.0.0-20200710101835-de7d82ec7e0c
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.27.0
)

// Useful for local development (TODO: comment out when not needed)
replace github.com/smartcontractkit/external-initiator v0.0.0-20200710101835-de7d82ec7e0c => ../../
