package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/at15/tinycache/cache"
)

type httpServer struct {
	cache  cache.Cache
	server *http.Server
}

func NewHTTPServer(cache cache.Cache) Server {
	return &httpServer{
		cache: cache,
		// server will be initalized in Start
	}
}

func (s *httpServer) Start(ctx context.Context, addr string, port int) error {
	mux := http.NewServeMux()
	// https://go.dev/blog/routing-enhancements
	mux.HandleFunc("GET /cache/{bucket}/{key}", requireBucketAndKey(s.handleGet))
	mux.HandleFunc("PUT /cache/{bucket}/{key}", requireBucketAndKey(s.handleSet))
	mux.HandleFunc("DELETE /cache/{bucket}/{key}", requireBucketAndKey(s.handleDelete))
	mux.HandleFunc("/stats", s.handleStats)

	addr = fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	s.server = server

	log.Printf("Starting HTTP server on %s", addr)
	return server.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

type kvHandler func(bucket, key string, body []byte) ([]byte, error)

func requireBucketAndKey(handler kvHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.PathValue("key")
		if bucket == "" || key == "" {
			http.Error(w, "Invalid bucket or key", http.StatusBadRequest)
			return
		}
		// TODO: read operations from request parameters for ttl and policy
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		b, err := handler(bucket, key, body)
		if err != nil {
			// TODO: check error type, though we only return error when not found
			// so it's always 404...
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Write(b)
	}
}

func (s *httpServer) handleGet(bucket, key string, body []byte) ([]byte, error) {
	return s.cache.Get(bucket, key, cache.Options{})
}

func (s *httpServer) handleSet(bucket, key string, body []byte) ([]byte, error) {
	return nil, s.cache.Set(bucket, key, body, cache.Options{})
}

func (s *httpServer) handleDelete(bucket, key string, body []byte) ([]byte, error) {
	return nil, s.cache.Delete(bucket, key)
}

func (s *httpServer) handleStats(w http.ResponseWriter, r *http.Request) {
	// TODO: stats on cache hit, evictions etc.
}
