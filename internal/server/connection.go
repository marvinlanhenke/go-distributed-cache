package server

import (
	"sync"

	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcConn struct {
	pb.CacheServiceClient
	*grpc.ClientConn
}

type grpcConnPool struct {
	mu    sync.Mutex
	conns map[string]*grpcConn
}

func newGrpcConnPool() *grpcConnPool {
	return &grpcConnPool{conns: make(map[string]*grpcConn)}
}

func (c *grpcConnPool) get(addr string) (*grpcConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, ok := c.conns[addr]
	if ok {
		return conn, nil
	}

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	cc, err := grpc.NewClient(addr, opts)
	if err != nil {
		return nil, err
	}
	conn = &grpcConn{pb.NewCacheServiceClient(cc), cc}
	c.conns[addr] = conn

	return conn, nil
}
