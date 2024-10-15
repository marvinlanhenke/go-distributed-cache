package config

import (
	"flag"
	"fmt"
	"strings"

	"google.golang.org/grpc"
)

type Config struct {
	Addr           string
	Peers          []string
	Capacity       int
	MaxRecvMsgSize int
	MaxSendMsgSize int
	RPCTimeout     int
}

func New() (*Config, error) {
	addr := flag.String("addr", "localhost:8080", "server address <host>:<port>")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses")
	capacity := flag.Int("capacity", 100, "Cache capacity (number of items)")
	maxRecvMsgSize := flag.Int("max_recv_msg_size", 4*1024*1024, "Maximum message size the server can receive in bytes (default 4MB)")
	maxSendMsgSize := flag.Int("max_send_msg_size", 4*1024*1024, "Maximum message size the server can send in bytes (default 4MB)")

	flag.Parse()

	peerList := strings.Split(*peers, ",")

	if *capacity <= 0 {
		return nil, fmt.Errorf("capacity must be a positive integer")
	}
	if *maxRecvMsgSize <= 0 {
		return nil, fmt.Errorf("max_recv_msg_size must be a positive integer")
	}
	if *maxSendMsgSize <= 0 {
		return nil, fmt.Errorf("max_send_msg_size must be a positive integer")
	}

	return &Config{
		Addr:           *addr,
		Peers:          peerList,
		Capacity:       *capacity,
		MaxRecvMsgSize: *maxRecvMsgSize,
		MaxSendMsgSize: *maxSendMsgSize,
	}, nil
}

func (c *Config) GrpcServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(c.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(c.MaxSendMsgSize),
	}
}
