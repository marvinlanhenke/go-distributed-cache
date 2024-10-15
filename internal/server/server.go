package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
)

const replicationHeader = "X-Replication-Request"

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CacheServer struct {
	mu    sync.Mutex
	cache *cache.Cache
	peers []string
}

func New(peers []string, capacity int) *CacheServer {
	return &CacheServer{cache: cache.New(capacity), peers: peers}
}

func (cs *CacheServer) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req SetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cs.cache.Set(req.Key, req.Value, time.Hour*1)

	if r.Header.Get(replicationHeader) == "" {
		go cs.replicateSet(req.Key, req.Value)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (cs *CacheServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	value, ok := cs.cache.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"value": value}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
