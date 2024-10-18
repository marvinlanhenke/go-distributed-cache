package server

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/memberlist"
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
	cache      *cache.Cache
	hashRing   *hashring.HashRing
	memberlist *memberlist.Memberlist
	connPool   *grpcConnPool
	config     *config.Config
	limiter    *rate.Limiter
}

func New(cfg *config.Config) *cacheServer {
	cs := &cacheServer{
		cache:    cache.New(cfg.NumShards, cfg.Capacity, cfg.TTL),
		hashRing: hashring.New(),
		connPool: newGrpcConnPool(),
		config:   cfg,
		limiter:  rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateLimitBurst),
	}
	cs.memberlist = newMemberlist(cs, cfg)
	cs.hashRing.Add(&hashring.Node{ID: cfg.Addr, Addr: cfg.Addr})

	return cs
}

func (cs *cacheServer) Set(ctx context.Context, req *pb.SetRequest) (*empty.Empty, error) {
	isForwarded := req.SourceNode != ""
	if isForwarded {
		cs.cache.Set(req)
		return &empty.Empty{}, nil
	}
	req.SourceNode = cs.config.Addr

	nodes, ok := cs.hashRing.GetNodes(req.Key)
	if !ok {
		return nil, status.Errorf(codes.Internal, "not enough nodes available to achieve write quorum")
	}

	var wg sync.WaitGroup
	wg.Add(len(nodes))
	var writeSuccess int32

	for _, node := range nodes {
		if node.Addr == cs.config.Addr {
			go func() {
				defer wg.Done()
				atomic.AddInt32(&writeSuccess, 1)
			}()
		} else {
			go func(target string) {
				defer wg.Done()
				if err := cs.forwardSet(req, node.Addr); err == nil {
					atomic.AddInt32(&writeSuccess, 1)
				}
			}(node.Addr)
		}
	}

	wg.Wait()
	if atomic.LoadInt32(&writeSuccess) < int32(cs.hashRing.Replication) {
		log.Error().Str("addr", cs.config.Addr).Msg("no write quorum achieved")
		return nil, status.Errorf(codes.Internal, "no write quorum achived")
	}

	cs.cache.Set(req)
	return &empty.Empty{}, nil
}

func (cs *cacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	isForwarded := req.SourceNode != ""
	if isForwarded {
		resp, ok := cs.cache.Get(req)
		if !ok {
			return nil, status.Errorf(codes.NotFound, "no entry for key %q found", req.Key)
		}
		return resp, nil
	}
	req.SourceNode = cs.config.Addr

	nodes, ok := cs.hashRing.GetNodes(req.Key)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no entry for key %q found", req.Key)
	}

	var wg sync.WaitGroup
	wg.Add(len(nodes))
	var readSuccess int32
	responseCh := make(chan *pb.GetResponse, len(nodes))

	for _, node := range nodes {
		if node.Addr == cs.config.Addr {
			go func() {
				defer wg.Done()
				item, ok := cs.cache.Get(req)
				if ok {
					atomic.AddInt32(&readSuccess, 1)
					responseCh <- item
				}
			}()
		} else {
			go func(target string) {
				defer wg.Done()
				item, err := cs.forwardGet(req, target)
				if err == nil {
					atomic.AddInt32(&readSuccess, 1)
					responseCh <- item
				}
			}(node.Addr)
		}
	}

	wg.Wait()
	close(responseCh)

	if int(atomic.LoadInt32(&readSuccess)) < cs.hashRing.Replication {
		return nil, status.Errorf(codes.Internal, "not enough nodes available to achieve read quorum")
	}

	var response *pb.GetResponse
	var maxVersion uint32 = 0

	for resp := range responseCh {
		if resp.Version >= maxVersion {
			maxVersion = resp.Version
			response = resp
		}
	}

	if response == nil {
		return nil, status.Errorf(codes.NotFound, "no entry for key %q found", req.Key)
	}

	return response, nil
}

func (cs *cacheServer) forwardSet(in *pb.SetRequest, target string) error {
	log.Info().Str("addr", target).Msg("forwarding set request to target node")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := cs.connPool.get(target)
	if err != nil {
		log.Error().Err(err).Msg("failed to create grpc client while forwarding request")
		return err
	}

	if _, err := client.Set(ctx, in); err != nil {
		log.Error().Err(err).Str("addr", target).Msg("failed to forward set request")
		return err
	}

	return nil
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
