package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
	"github.com/marvinlanhenke/go-distributed-cache/internal/hashring"
)

const (
	replicationHeader = "X-Replication-Request"
	forwardedHeader   = "X-Forwarded-For"
)

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CacheServer struct {
	mu       sync.Mutex
	cache    *cache.Cache
	hashRing *hashring.HashRing
	peers    []string
	addr     string
}

func New(addr string, peers []string, capacity int) *CacheServer {
	cs := &CacheServer{
		cache:    cache.New(capacity),
		hashRing: hashring.New(),
		peers:    peers,
		addr:     addr,
	}

	for _, peer := range peers {
		if peer != "" {
			cs.hashRing.Add(&hashring.Node{ID: peer, Addr: peer})
		}
	}
	cs.hashRing.Add(&hashring.Node{ID: addr, Addr: addr})

	return cs
}

func (cs *CacheServer) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req SetRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isReplication := r.Header.Get(replicationHeader) != ""
	if isReplication {
		cs.cache.Set(req.Key, req.Value, time.Hour*1)
		w.WriteHeader(http.StatusOK)
		return
	}

	targetNode, ok := cs.hashRing.Get(req.Key)
	if !ok {
		http.Error(w, "no target node found", http.StatusInternalServerError)
		return
	}

	if targetNode.Addr == cs.addr {
		cs.cache.Set(req.Key, req.Value, time.Hour*1)

		go cs.replicateSet(req.Key, req.Value)

		w.WriteHeader(http.StatusOK)
		return
	}

	cs.forwardRequest(w, r, targetNode, body)
}

func (cs *CacheServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	targetNode, ok := cs.hashRing.Get(key)
	if !ok {
		http.Error(w, "no target node found", http.StatusInternalServerError)
		return
	}

	if targetNode.Addr == cs.addr {
		value, ok := cs.cache.Get(key)
		if !ok {
			http.NotFound(w, r)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]string{"value": value}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	originalSender := r.Header.Get(forwardedHeader)
	if originalSender == cs.addr {
		http.Error(w, "forward loop detected", http.StatusBadRequest)
		return
	}

	r.Header.Set(forwardedHeader, cs.addr)
	cs.forwardRequest(w, r, targetNode, nil)
}

func (cs *CacheServer) replicateSet(key, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	req := SetRequest{Key: key, Value: value}
	data, err := json.Marshal(req)
	if err != nil {
		log.Printf("failed to marshal request to json: %v", err)
		return
	}

	for _, peer := range cs.peers {
		if peer == "" {
			continue
		}

		go func(peer string) {
			client := &http.Client{}
			req, err := http.NewRequest("POST", peer+"/set", bytes.NewReader(data))
			if err != nil {
				log.Printf("failed to create replication request: %v", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(replicationHeader, "true")

			if _, err := client.Do(req); err != nil {
				log.Printf("failed to replicate to peer %s: %v", peer, err)
				return
			}

			log.Println("replicated successful to", peer)
		}(peer)
	}
}

func (cs *CacheServer) forwardRequest(w http.ResponseWriter, r *http.Request, targetNode *hashring.Node, body []byte) {
	log.Printf("forwarding request to node %s", targetNode.Addr)
	client := &http.Client{}

	var req *http.Request
	var err error

	if r.Method == http.MethodGet {
		url := fmt.Sprintf("%s%s?%s", targetNode.Addr, r.URL.Path, r.URL.RawQuery)
		req, err = http.NewRequest(r.Method, url, nil)
	} else if r.Method == http.MethodPost {
		url := fmt.Sprintf("%s%s", targetNode.Addr, r.URL.Path)
		req, err = http.NewRequest(r.Method, url, bytes.NewReader(body))
	}

	if err != nil {
		log.Printf("failed to create forward request: %v", err)
		http.Error(w, "failed to create forward request", http.StatusInternalServerError)
		return
	}

	req.Header = r.Header
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to forward request to node %s: %v", targetNode.Addr, err)
		http.Error(w, "failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("failed to write response from node %s: %v", targetNode.Addr, err)
	}
}
