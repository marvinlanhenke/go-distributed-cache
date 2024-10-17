package main

import (
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/joho/godotenv"
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
	grpcServer := app.mount()

	lis, err := net.Listen("tcp", app.config.Port)
	if err != nil {
		log.Fatal().Err(err).Str("port", app.config.Port).Msg("failed to listen")
	}

	log.Info().Str("port", app.config.Port).Msg("server starting...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err)
	}
}

func (app *application) mount() *grpc.Server {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	interceptorOpts := grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(server.InterceptorLogger(log.Logger), loggingOpts...),
	)

	opts := app.config.GrpcServerOptions()
	opts = append(opts, interceptorOpts)

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterCacheServiceServer(grpcServer, server.New(app.config))
	reflection.Register(grpcServer)

	go server.GracefulShutdown(grpcServer, app.config)

	return grpcServer
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Warn().Err(err).Msg("failed loading .env file")
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create config")
	}

	app := NewApplication(cfg)

	app.run()
}
