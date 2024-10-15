package main

import (
	"flag"
	"log"
	"net"
	"strings"

	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"github.com/marvinlanhenke/go-distributed-cache/internal/server"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "server address <host>:<port>")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses")
	flag.Parse()

	peerList := strings.Split(*peers, ",")

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", *addr, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCacheServiceServer(grpcServer, server.New(*addr, peerList))
	log.Printf("grpc server starting at: %s\n", *addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve grpc server: %v", err)
	}
}
