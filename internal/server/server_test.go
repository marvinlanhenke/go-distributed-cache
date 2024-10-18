package server

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
	"github.com/marvinlanhenke/go-distributed-cache/internal/pb"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func createHashRing(addrs []string, replication int) *hashring.HashRing {
	hashRing := hashring.New()
	for _, addr := range addrs {
		hashRing.Add(&hashring.Node{ID: addr, Addr: addr})
	}

	hashRing.Replication = replication

	return hashRing
}

func startServer(port string, hashRing *hashring.HashRing) (*cacheServer, *grpc.Server) {
	config, _ := config.New()
	config.Addr = port

	srv := &cacheServer{
		cache:    cache.New(10, 100, time.Second*3600),
		hashRing: hashRing,
		connPool: newGrpcConnPool(),
		config:   config,
		limiter:  rate.NewLimiter(rate.Limit(10), 100),
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCacheServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)

	go func(port string, srv *grpc.Server) {
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve grpc server: %v", err)
		}
	}(port, grpcServer)

	return srv, grpcServer
}

func TestServerSetSuccess(t *testing.T) {
	addrs := []string{":8080", ":8081", ":8082"}
	hashRing := createHashRing(addrs, 2)
	srv1, grpc1 := startServer(":8080", hashRing)
	_, grpc2 := startServer(":8081", hashRing)
	_, grpc3 := startServer(":8082", hashRing)
	defer grpc1.Stop()
	defer grpc2.Stop()
	defer grpc3.Stop()

	ctx := context.Background()
	req := &pb.SetRequest{
		Key:   "test-key",
		Value: "test-value",
	}

	_, err := srv1.Set(ctx, req)
	require.NoError(t, err, "expected no error, instead got %v", err)
}

func TestServerSetNoWriteQuorum(t *testing.T) {
	addrs := []string{":8080", ":8081", ":8082"}
	hashRing := createHashRing(addrs, 2)
	srv1, grpc1 := startServer(":8080", hashRing)
	defer grpc1.Stop()

	ctx := context.Background()
	req := &pb.SetRequest{
		Key:   "test-key",
		Value: "test-value",
	}

	_, err := srv1.Set(ctx, req)
	require.Error(t, err, "expected an error, instead got %v", err)
}

func TestServerGetSuccess(t *testing.T) {
	addrs := []string{":8080", ":8081", ":8082"}
	hashRing := createHashRing(addrs, 2)
	srv1, grpc1 := startServer(":8080", hashRing)
	_, grpc2 := startServer(":8081", hashRing)
	_, grpc3 := startServer(":8082", hashRing)
	defer grpc1.Stop()
	defer grpc2.Stop()
	defer grpc3.Stop()

	ctx := context.Background()
	setReq := &pb.SetRequest{
		Key:   "test-key",
		Value: "test-value",
	}

	_, err := srv1.Set(ctx, setReq)
	require.NoError(t, err, "expected no error, instead got %v", err)

	getReq := &pb.GetRequest{Key: "test-key"}
	expected := &pb.GetResponse{Value: "test-value", Version: uint32(0)}
	result, err := srv1.Get(ctx, getReq)
	require.NoError(t, err, "expected no error, instead got %v", err)
	require.Equal(t, expected.Value, result.Value, "expected %v, instead got %v", expected.Value, result.Value)
	require.Equal(t, expected.Version, result.Version, "expected %v, instead got %v", expected.Version, result.Version)
}

func TestServerGetNoReadQuorum(t *testing.T) {
	addrs := []string{":8080", ":8081", ":8082"}
	hashRing := createHashRing(addrs, 2)
	srv1, grpc1 := startServer(":8080", hashRing)
	_, grpc2 := startServer(":8081", hashRing)
	_, grpc3 := startServer(":8082", hashRing)
	defer grpc1.Stop()

	ctx := context.Background()
	setReq := &pb.SetRequest{
		Key:   "test-key",
		Value: "test-value",
	}

	_, err := srv1.Set(ctx, setReq)
	require.NoError(t, err, "expected no error, instead got %v", err)

	grpc2.Stop()
	grpc3.Stop()

	getReq := &pb.GetRequest{Key: "test-key"}
	_, err = srv1.Get(ctx, getReq)
	require.Error(t, err, "expected error, instead got %v", err)
}
