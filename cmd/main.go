package main

import (
	"log"
	"net"

	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"github.com/marvinlanhenke/go-distributed-cache/internal/server"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to provide a valid config: %v", err)
	}

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", cfg.Addr, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCacheServiceServer(grpcServer, server.New(cfg))
	log.Printf("server starting at: %s\n", cfg.Addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve grpc server: %v", err)
	}
}
