package grpc

import (
	"net"

	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/integration/mock-client/blockchain"
	"google.golang.org/grpc"
)

func RunServer() {
	/* #nosec */
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		logger.Error(err)
		return
	}
	grpcServer := grpc.NewServer()

	attachServers(grpcServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(err)
			return
		}
	}()
}

func attachServers(grpcServer *grpc.Server) {
	// iotex
	iotexapi.RegisterAPIServiceServer(grpcServer, &blockchain.MockIoTeXServer{})
}
