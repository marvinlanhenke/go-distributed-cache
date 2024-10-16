package server

import (
	"context"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	cache    *cache.Cache
	hashRing *hashring.HashRing
	config   *config.Config
	limiter  *rate.Limiter
}

func New(cfg *config.Config) *cacheServer {
	cs := &cacheServer{
		cache:    cache.New(cfg.Capacity, cfg.TTL*time.Second),
		hashRing: hashring.New(),
		config:   cfg,
		limiter:  rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateLimitBurst),
	}

	for _, peer := range cfg.Peers {
		if peer != "" {
			cs.hashRing.Add(&hashring.Node{ID: peer, Addr: peer})
		}
	}
	cs.hashRing.Add(&hashring.Node{ID: cfg.Addr, Addr: cfg.Addr})

	return cs
}

func (cs *cacheServer) Set(ctx context.Context, req *pb.SetRequest) (*empty.Empty, error) {
	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: you must provide a valid key")
	}

	isReplication := req.SourceNode != ""
	if isReplication {
		cs.cache.Set(req)
		return &empty.Empty{}, nil
	}

	targetNode, ok := cs.hashRing.Get(req.Key)
	if !ok {
		return &empty.Empty{}, nil
	}

	if targetNode.Addr == cs.config.Addr {
		cs.cache.Set(req)
		go cs.replicate(req)
		return &empty.Empty{}, nil
	}
	// TODO: forward
	return &empty.Empty{}, nil
}

func (cs *cacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	resp, ok := cs.cache.Get(req)
	if !ok {
		return nil, nil
	}
	return resp, nil
}

func (cs *cacheServer) replicate(req *pb.SetRequest) {
	// TODO: impl client pool
}
