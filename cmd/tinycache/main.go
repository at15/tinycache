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
	cache := cache.NewLRUCache(10, 1*time.Second)
	server := server.NewHTTPServer(cache)

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
