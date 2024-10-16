package server

import (
	"context"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"golang.org/x/time/rate"
)

type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	cache   *cache.Cache
	config  *config.Config
	limiter *rate.Limiter
}

func New(cfg *config.Config) *cacheServer {
	cs := &cacheServer{
		cache:   cache.New(cfg.Capacity, cfg.TTL*time.Second),
		config:  cfg,
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateLimitBurst),
	}

	return cs
}

func (cs *cacheServer) Set(ctx context.Context, req *pb.SetRequest) (*empty.Empty, error) {
	cs.cache.Set(req)
	return &empty.Empty{}, nil
}

func (cs *cacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	resp, ok := cs.cache.Get(req)
	if !ok {
		return nil, nil
	}
	return resp, nil
}
