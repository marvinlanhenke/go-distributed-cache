package main

import (
	"net"

	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"github.com/marvinlanhenke/go-distributed-cache/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err)
	}

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", cfg.Addr).Msg("failed to listen")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCacheServiceServer(grpcServer, server.New(cfg))
	log.Info().Str("addr", cfg.Addr).Msg("server is starting")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err)
	}
}
