package config

import (
	"flag"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
)

type Config struct {
	CacheConfig `toml:"cache" validate:"required"`
	GRPCConfig  `toml:"grpc" validate:"required"`
}

type CacheConfig struct {
	Addr      string        `toml:"addr" validate:"required"`
	Peers     []string      `toml:"peers"`
	NumShards int           `toml:"num_shards" validate:"required;gt=0"`
	Capacity  int           `toml:"capacity" validate:"required,gt=0"`
	TTL       time.Duration `toml:"ttl" validate:"required,gt=0"`
}

type GRPCConfig struct {
	MaxRecvMsgSize int `toml:"max_recv_msg_size" validate:"required,gt=0"`
	MaxSendMsgSize int `toml:"max_send_msg_size" validate:"required,gt=0"`
	RPCTimeout     int `toml:"rpc_timeout" validate:"required,gt=0"`
	RateLimit      int `toml:"rate_limit" validate:"required,gt=0"`
	RateLimitBurst int `toml:"rate_limit_burst" validate:"required,gt=0"`
}

func New() (*Config, error) {
	path := flag.String("config", "config.toml", "TOML config filepath")
	flag.Parse()

	var cfg Config

	if _, err := toml.DecodeFile(*path, &cfg); err != nil {
		return nil, err
	}

	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) GrpcServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(c.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(c.MaxSendMsgSize),
	}
}
