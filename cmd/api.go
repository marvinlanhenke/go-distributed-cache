package main

import (
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"github.com/marvinlanhenke/go-distributed-cache/internal/server"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type application struct {
	config *config.Config
}

func NewApplication(config *config.Config) *application {
	return &application{config: config}
}

func (app *application) run() {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	lis, err := net.Listen("tcp", app.config.Addr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", app.config.Addr).Msg("failed to listen")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(server.InterceptorLogger(log.Logger), opts...),
		),
	)
	pb.RegisterCacheServiceServer(grpcServer, server.New(app.config))

	reflection.Register(grpcServer)

	go server.GracefulShutdown(grpcServer, app.config)

	log.Info().Str("addr", app.config.Addr).Msg("server starting...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err)
	}
}
