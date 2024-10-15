package server

import (
	"context"

	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/types/known/emptypb"
)

type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	config  *config.Config
	limiter *rate.Limiter
}

func New(cfg *config.Config) *cacheServer {
	cs := &cacheServer{
		config:  cfg,
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateLimitBurst),
	}

	return cs
}

func (cs *cacheServer) Set(ctx context.Context, req *pb.SetRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (cs *cacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	return nil, nil
}
