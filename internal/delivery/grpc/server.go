package grpc

import (
	"github.com/itsLeonB/gerpc"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/delivery/grpc/server"
	"github.com/itsLeonB/stortr/internal/provider"
	"google.golang.org/grpc"
)

func Setup(configs config.Config) *gerpc.GrpcServer {
	providers := provider.All(configs)
	servers := server.ProvideServers(providers.Services)

	// Middlewares/Interceptors
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			gerpc.NewLoggingInterceptor(providers.Logger),
			gerpc.NewErrorInterceptor(providers.Logger),
		),
	}

	return gerpc.NewGrpcServer().
		WithLogger(providers.Logger).
		WithAddress(":" + configs.App.Port).
		WithOpts(opts...).
		WithRegisterSrvFunc(servers.Register).
		WithShutdownFunc(providers.Shutdown)
}
