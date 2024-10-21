package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// GracefulShutdown listens for system interrupt signals (e.g., SIGINT, SIGTERM) and gracefully shuts down the gRPC server.
// This function ensures that the server stops accepting new connections and allows in-progress requests to complete before shutting down.
func GracefulShutdown(srv *grpc.Server, cfg *config.Config) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Info().Str("addr", cfg.Addr).Msg("server shutting down...")
	srv.GracefulStop()
}
