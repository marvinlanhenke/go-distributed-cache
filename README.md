# Go Distributed Cache

Go Distributed Cache is a highly scalable, fault-tolerant distributed caching system written in Go. It supports sharded caches, gRPC-based communication, and strong consistency using quorum-based replication. The system features dynamic membership with nodes joining and leaving the cluster without disruption, making it horizontally scalable.

Disclaimer: This project is not intended to be used in production.

## Overview

This project provides a distributed in-memory cache system designed to handle high availability, strong consistency, and scalability.

By leveraging sharded caches and a consistent hash ring, the system ensures that data is distributed efficiently across all nodes in the cluster.

The project uses gRPC for inter-node communication and membership protocols to manage dynamic cluster membership.

Data replication across nodes is done using a quorum-based approach to ensure strong consistency during read and write operations.

## How It Works

- **Sharded Cache**: The cache is divided into multiple shards to reduce contention and improve performance.
- **gRPC Communication**: Nodes communicate with each other using gRPC for efficiency, providing fast and reliable inter-node communication.
- **Quorum-Based Replication**: Each key-value pair is replicated to a majority (quorum) of nodes. This ensures strong consistency even in the event of node failures.
- **Dynamic Membership**: Nodes can join and leave the cluster dynamically, and the system adjusts the distribution of keys accordingly using consistent hashing.
- **Graceful Shutdown:** The system ensures that nodes gracefully leave the cluster, completing in-progress operations before exiting.
- **Structured Logging:** For fast structured logging, _zerolog_ is used.

## Environment Variables

The following environment variables can be configured to customize the system:

- `ADDR`: The address (host
  ) on which the node will listen for gRPC requests (default: localhost:8080).
- `PEERS`: Comma-separated list of peer node addresses to join the cluster.
- `NUM_SHARDS`: Number of cache shards (default: 1).
- `CAPACITY`: Total cache capacity across all shards (default: 1000).
- `TTL`: Time-to-live for cache entries, in seconds (default: 3600).
- `MAX_RECV_MSG_SIZE`: Maximum size (in bytes) for incoming gRPC messages (default: 4194304).
- `MAX_SEND_MSG_SIZE`: Maximum size (in bytes) for outgoing gRPC messages (default: 4194304).
- `RPC_TIMEOUT`: Timeout duration (in seconds) for inter-node gRPC calls (default: 5).
- `RATE_LIMIT`: Maximum number of incoming requests per second (default: 10).
- `RATE_LIMIT_BURST`: Maximum burst size for rate-limited requests (default: 100).

## Installation

### Binary

1. Clone the repository

```shell
git clone https://github.com/your-username/go-distributed-cache.git
cd go-distributed-cache
```

2. Build the project

```shell
go build -o distributed-cache .
```

3. Run the binary

```shell
./distributed-cache
```

### Docker

1. Build the Docker image

```shell
docker build -t distributed-cache .
```

2. Run the container

```shell
docker run -d -p 8080:8080 -e ADDR=localhost:8080 distributed-cache
```

## Basic Examples

### Starting a Single Node

You can start a single node by running the following command

```shell
./distributed-cache
```

### Starting Multiple Nodes

To run multiple nodes, each node should be started with its own address and a list of peers

```shell
./distributed-cache -e ADDR=node1:8080 -e PEERS=node2:8080,node3:8080
./distributed-cache -e ADDR=node2:8080 -e PEERS=node1:8080,node3:8080
./distributed-cache -e ADDR=node3:8080 -e PEERS=node1:8080,node2:8080
```

The system will automatically adjust and distribute cache entries across the nodes using consistent hashing.

### Set and Get Example

To set a value:

```shell
grpcurl -plaintext -d '{"key":"foo", "value":"bar"}' localhost:8080 pb.CacheService/Set
```

To get the value:

```shell
grpcurl -plaintext -d '{"key":"foo"}' localhost:8080 pb.CacheService/Get
```

## Missing Features / Trade-Offs

- **Anti-Entropy Mechanism**: Re-distribution and replication is currently not handled when the hash ring changes (e.g. a node has left). This could be done using a `merkle-tree` in order to detect differences between nodes quicky and efficiently.
- **Last-Write-Wins**: Currently the `last-write-wins` strategy is used for conflict resolution. This is done via a monotonically increasing ID (version). While this approach is simple and easy to understand, it is vulnerable to data-loss.

## License

This project ist licensed under the Apache License, Version 2.0
