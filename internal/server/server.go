package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/marvinlanhenke/go-distributed-cache/internal/cache"
)

type CacheServer struct {
	cache *cache.Cache
}

func New(capacity int) *CacheServer {
	return &CacheServer{cache: cache.New(capacity)}
}

func (cs *CacheServer) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cs.cache.Set(req.Key, req.Value, time.Hour*1)

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
