package grpc

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/grpc/middleware"
	"google.golang.org/grpc"
)

func NewServerAuth(opts ...grpc.ServerOption) *grpc.Server {
	return newServerWithMiddleware(opts...)
}

func NewServerOrchestrator(opts ...grpc.ServerOption) *grpc.Server {
	return newServerWithMiddleware(opts...)
}

func newServerWithMiddleware(opts ...grpc.ServerOption) *grpc.Server {
	chainedUnary := grpc.ChainUnaryInterceptor(
		middleware.UnaryServerRecovery(),
		middleware.UnaryServerLogging(),
		middleware.UnaryServerError(),
	)

	chainedStream := grpc.ChainStreamInterceptor(
		middleware.StreamServerRecovery(),
		middleware.StreamServerLogging(),
		middleware.StreamServerError(),
	)

	serverOpts := append([]grpc.ServerOption{
		chainedUnary,
		chainedStream,
	}, opts...)

	return grpc.NewServer(serverOpts...)
}
