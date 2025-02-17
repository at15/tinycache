package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/at15/tinycache/cache"
	"github.com/at15/tinycache/server"
)

func main() {
	// Start http server on port 8080
	metrics := cache.NewPrometheusMetrics()
	cache := cache.NewLRUCache(10, 500*time.Millisecond, metrics)
	server := server.NewHTTPServer(cache, metrics)

	// Listen to ctrl c to stop the server in background
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		server.Start(context.Background(), "localhost", 8080)
	}()
	<-ch
	log.Println("Stopping server...")
	cache.Stop()
	server.Stop(context.Background())
}
