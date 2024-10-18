package config

import (
	"strings"
	"time"

	"google.golang.org/grpc"
)

type Config struct {
	Addr           string
	Peers          []string
	NumShards      int
	Capacity       int
	TTL            time.Duration
	MaxRecvMsgSize int
	MaxSendMsgSize int
	RPCTimeout     int
	RateLimit      int
	RateLimitBurst int
}

func New() (*Config, error) {
	numShards := getInt("NUM_SHARDS", 1)
	capacity := getInt("CAPACITY", 1000)
	TTL := getInt("TTL", 3600)
	maxRecvMsgSize := getInt("MAX_RECV_MSG_SIZE", 4194304)
	maxSendMsgSize := getInt("MAX_SEND_MSG_SIZE", 4194304)
	rpcTimeout := getInt("RPC_TIMEOUT", 5)
	rateLimit := getInt("RATE_LIMIT", 10)
	rateLimitBurst := getInt("RATE_LIMIT_BURST", 100)

	addr := getString("ADDR", "localhost:8080")
	peersEnv := getString("PEERS", "")
	peers := strings.Split(peersEnv, ",")

	return &Config{
		Addr:           addr,
		Peers:          peers,
		NumShards:      numShards,
		Capacity:       capacity,
		TTL:            time.Duration(TTL) * time.Second,
		MaxRecvMsgSize: maxRecvMsgSize,
		MaxSendMsgSize: maxSendMsgSize,
		RPCTimeout:     rpcTimeout,
		RateLimit:      rateLimit,
		RateLimitBurst: rateLimitBurst,
	}, nil
}

func (c *Config) GrpcServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(c.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(c.MaxSendMsgSize),
	}
}
