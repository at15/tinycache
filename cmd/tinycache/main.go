package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/at15/tinycache/cache"
	"github.com/at15/tinycache/server"
)

var (
	useGRPC bool
	port    int
	host    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tinycache",
		Short: "A tiny cache server supporting both HTTP and gRPC protocols",
		Run:   run,
	}

	// Add flags
	rootCmd.Flags().BoolVar(&useGRPC, "grpc", false, "Use gRPC server instead of HTTP")
	rootCmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")
	rootCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host address to bind to")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	metrics := cache.NewPrometheusMetrics()
	cache := cache.NewLRUCache(10, 500*time.Millisecond, metrics)

	var srv server.Server
	if useGRPC {
		srv = server.NewGRPCServer(cache, metrics)
		log.Printf("Starting gRPC server on %s:%d", host, port)
	} else {
		srv = server.NewHTTPServer(cache, metrics)
		log.Printf("Starting HTTP server on %s:%d", host, port)
	}

	// Listen to ctrl c to stop the server in background
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		if err := srv.Start(context.Background(), host, port); err != nil {
			log.Printf("Server error: %v", err)
			ch <- os.Interrupt // Trigger shutdown
		}
	}()

	<-ch
	log.Println("Stopping server...")
	cache.Stop()
	srv.Stop(context.Background())
}
