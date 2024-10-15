package server

import (
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
)

type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	config *config.Config
}

func New(cfg *config.Config) *cacheServer {
	cs := &cacheServer{
		config: cfg,
	}

	return cs
}
