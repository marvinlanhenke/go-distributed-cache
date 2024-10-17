package server

import (
	"context"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	cache    *cache.Cache
	hashRing *hashring.HashRing
	connPool *grpcConnPool
	config   *config.Config
	limiter  *rate.Limiter
}

func New(cfg *config.Config) *cacheServer {
	cs := &cacheServer{
		cache:    cache.New(cfg.NumShards, cfg.Capacity, cfg.TTL),
		hashRing: hashring.New(),
		connPool: newGrpcConnPool(),
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

	isForwarded := req.SourceNode != ""
	if isForwarded {
		cs.cache.Set(req)
		return &empty.Empty{}, nil
	}
	req.SourceNode = cs.config.Addr

	targetNode, ok := cs.hashRing.Get(req.Key)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "a node for the requested key %s was not found", req.Key)
	}

	if targetNode.Addr == cs.config.Addr {
		log.Info().Msg("handling set request on current node")
		cs.cache.Set(req)
		go cs.replicate(req)
		return &empty.Empty{}, nil
	}

	go cs.forwardSet(req, targetNode.Addr)
	return &empty.Empty{}, nil
}

func (cs *cacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	targetNode, ok := cs.hashRing.Get(req.Key)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "a node for the requested key %s was not found", req.Key)
	}

	if targetNode.Addr == cs.config.Addr {
		log.Info().Msg("handling get request on current node")
		value, ok := cs.cache.Get(req)
		if !ok {
			return nil, status.Errorf(codes.NotFound, "an entry for key %s does not exist", req.Key)
		}
		return value, nil
	}

	if req.SourceNode == cs.config.Addr {
		return nil, status.Errorf(codes.Internal, "forward loop detected")
	}

	req.SourceNode = cs.config.Addr
	return cs.forwardGet(req, targetNode.Addr)
}

func (cs *cacheServer) replicate(in *pb.SetRequest) {
	for _, peer := range cs.config.Peers {
		if peer == "" {
			continue
		}
		log.Info().Str("addr", peer).Msg("sending replication request to peer")
		go cs.forwardSet(in, peer)
	}
}

func (cs *cacheServer) forwardSet(in *pb.SetRequest, target string) {
	log.Info().Str("addr", target).Msg("forwarding set request on target node")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := cs.connPool.get(target)
	if err != nil {
		log.Error().Err(err).Msg("failed to create grpc client while forwarding request")
		return
	}

	if _, err := client.Set(ctx, in); err != nil {
		log.Error().Err(err).Str("addr", target).Msg("failed to forward set request")
		return
	}

	log.Info().Str("addr", target).Msg("successfully forwarded request")
}

func (cs *cacheServer) forwardGet(in *pb.GetRequest, target string) (*pb.GetResponse, error) {
	log.Info().Str("addr", target).Msg("forwarding get request on target node")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := cs.connPool.get(target)
	if err != nil {
		log.Error().Err(err).Msg("failed to create grpc client while forwarding get request")
		return nil, status.Errorf(codes.Internal, "failed to forward request")
	}

	return client.Get(ctx, in)
}
