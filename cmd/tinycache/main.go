package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/at15/tinycache/cache"
	"github.com/at15/tinycache/proto"
	"github.com/at15/tinycache/server"
)

var (
	// server flags
	useGRPC bool
	port    int
	host    string

	// client flags
	clientHost string
	clientPort int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tinycache",
		Short: "A tiny cache server supporting both HTTP and gRPC protocols",
	}

	// Server command
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the cache server",
		Run:   runServer,
	}

	// Client command
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Start an interactive client",
		Run:   runClient,
	}

	// Server flags
	serverCmd.Flags().BoolVar(&useGRPC, "grpc", false, "Use gRPC server instead of HTTP")
	serverCmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")
	serverCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host address to bind to")

	// Client flags
	clientCmd.Flags().StringVar(&clientHost, "host", "localhost", "Server host to connect to")
	clientCmd.Flags().IntVar(&clientPort, "port", 8080, "Server port to connect to")

	rootCmd.AddCommand(serverCmd, clientCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	metrics := cache.NewPrometheusMetrics()
	cache := cache.NewLRUCache(10, 500*time.Millisecond, metrics)

	var srv server.Server
	if useGRPC {
		// TODO: expose promtheus metrics for gRPC server
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

func runClient(cmd *cobra.Command, args []string) {
	addr := fmt.Sprintf("%s:%d", clientHost, clientPort)
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := proto.NewTinyCacheClient(conn)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("TinyCache CLI (type 'help' for commands, 'exit' to quit)")
	fmt.Printf("Connected to %s\n", addr)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		args := strings.Fields(input)
		cmd := strings.ToLower(args[0])

		switch cmd {
		case "exit", "quit":
			return
		case "help":
			printHelp()
		case "get":
			if len(args) != 3 {
				fmt.Println("Usage: get <bucket> <key>")
				continue
			}
			handleGet(client, args[1], args[2])
		case "set":
			if len(args) < 4 {
				fmt.Println("Usage: set <bucket> <key> <value> [ttl_ms]")
				continue
			}
			var ttl int64 = 0
			if len(args) > 4 {
				var err error
				ttl, err = strconv.ParseInt(args[4], 10, 64)
				if err != nil {
					fmt.Printf("Invalid TTL: %v\n", err)
					continue
				}
			}
			handleSet(client, args[1], args[2], args[3], ttl)
		case "del", "delete":
			if len(args) != 3 {
				fmt.Println("Usage: del <bucket> <key>")
				continue
			}
			handleDelete(client, args[1], args[2])
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			printHelp()
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  get <bucket> <key>                    Get value by bucket and key")
	fmt.Println("  set <bucket> <key> <value> [ttl_ms]  Set value with optional TTL in milliseconds")
	fmt.Println("  del <bucket> <key>                    Delete value by bucket and key")
	fmt.Println("  help                                  Show this help message")
	fmt.Println("  exit                                  Exit the client")
}

func handleGet(client proto.TinyCacheClient, bucket, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.Get(ctx, &proto.GetRequest{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("%s\n", resp.Value)
}

func handleSet(client proto.TinyCacheClient, bucket, key, value string, ttlMs int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.Set(ctx, &proto.SetRequest{
		Bucket: bucket,
		Key:    key,
		Value:  []byte(value),
		TtlMs:  int32(ttlMs),
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("OK")
}

func handleDelete(client proto.TinyCacheClient, bucket, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.Delete(ctx, &proto.DeleteRequest{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("OK")
}
