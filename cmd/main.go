package main

import (
	"log"
	"net/http"

	"github.com/marvinlanhenke/go-distributed-cache/internal/server"
)

func main() {
	cs := server.New(10)
	http.HandleFunc("/set", cs.SetHandler)
	http.HandleFunc("/get", cs.GetHandler)

	log.Println("server is starting...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
