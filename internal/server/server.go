package server

import "github.com/marvinlanhenke/go-distributed-cache/internal/pb"

type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	addr  string
	peers []string
}

func New(addr string, peers []string) *cacheServer {
	cs := &cacheServer{
		addr:  addr,
		peers: peers,
	}

	return cs
}
