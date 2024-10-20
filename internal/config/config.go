package config

import (
	"strings"
	"time"

	"google.golang.org/grpc"
)

// Config holds the configuration settings for the distributed cache system.
// It defines parameters such as network settings, cache behavior, and gRPC options.
type Config struct {
	Addr           string        // Address on which the gRPC server listens.
	Peers          []string      // List of peer addresses in the distributed system.
	NumShards      int           // Number of shards used to partition the cache.
	Capacity       int           // Maximum number of cache entries across all shards.
	TTL            time.Duration // Time-to-live (TTL) for cache entries.
	MaxRecvMsgSize int           // Maximum size of a received gRPC message (in bytes).
	MaxSendMsgSize int           // Maximum size of a sent gRPC message (in bytes).
	RateLimit      int           // Rate limit for incoming requests per second.
	RateLimitBurst int           // Maximum burst size for rate-limited requests.
}

// Creates and initializes a new Config struct by loading configuration values from environment variables.
func New() (*Config, error) {
	numShards := getInt("NUM_SHARDS", 1)
	capacity := getInt("CAPACITY", 1000)
	TTL := getInt("TTL", 3600)
	maxRecvMsgSize := getInt("MAX_RECV_MSG_SIZE", 4194304)
	maxSendMsgSize := getInt("MAX_SEND_MSG_SIZE", 4194304)
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
		RateLimit:      rateLimit,
		RateLimitBurst: rateLimitBurst,
	}, nil
}

// GrpcServerOptions returns a slice of gRPC server options configured based on the current Config values.
// These options include message size limits for receiving and sending gRPC messages.
func (c *Config) GrpcServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(c.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(c.MaxSendMsgSize),
	}
}
