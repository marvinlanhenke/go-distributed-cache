package config

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	Addr     string
	Peers    []string
	Capacity int
}

func New() (*Config, error) {
	addr := flag.String("addr", "localhost:8080", "server address <host>:<port>")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses")
	capacity := flag.Int("capacity", 100, "Cache capacity (number of items)")
	flag.Parse()

	peerList := strings.Split(*peers, ",")

	if *capacity <= 0 {
		return nil, fmt.Errorf("capacity must be a positive integer")
	}

	return &Config{
		Addr:     *addr,
		Peers:    peerList,
		Capacity: *capacity,
	}, nil
}
