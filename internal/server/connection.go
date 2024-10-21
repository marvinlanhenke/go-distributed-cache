package server

import (
	"sync"

	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Wraps a gRPC client connection and the generated CacheServiceClient.
// It provides both the client for interacting with the cache service and the underlying connection.
type grpcConn struct {
	pb.CacheServiceClient // gRPC client for cache service interaction.
	*grpc.ClientConn      // Underlying gRPC client connection.
}

// A thread-safe pool for managing and reusing gRPC connections to different server addresses.
// It maintains a map of active connections and provides methods for accessing them.
type grpcConnPool struct {
	mu    sync.Mutex           // Mutex to synchronize access to the connections map.
	conns map[string]*grpcConn // Map of server addresses to their corresponding gRPC connections.
}

// Creates and initializes a new grpcConnPool instance.
// It returns a pointer to the newly created connection pool, ready to manage gRPC connections.
func newGrpcConnPool() *grpcConnPool {
	return &grpcConnPool{conns: make(map[string]*grpcConn)}
}

// Retrieves an existing gRPC connection to the specified address, or establishes a new one if it doesn't exist.
//
// If a new connection is established, it is added to the pool for future reuse.
// Returns a pointer to the grpcConn and any error encountered during connection establishment.
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
