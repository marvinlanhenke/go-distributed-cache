package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/marvinlanhenke/go-distributed-cache/internal/server"
)

var (
	host  string
	port  string
	peers string
)

func main() {
	flag.StringVar(&host, "host", "http://localhost", "HTTP server port")
	flag.StringVar(&port, "port", ":8080", "HTTP server port")
	flag.StringVar(&peers, "peers", "", "Comma-separated list of peer addresses")
	flag.Parse()

	addr := host + port
	peerList := strings.Split(peers, ",")
	capacity := 10

	cs := server.New(addr, peerList, capacity)
	http.HandleFunc("/set", cs.SetHandler)
	http.HandleFunc("/get", cs.GetHandler)

	log.Println("server is starting...")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
