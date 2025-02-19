package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/at15/tinycache/cache"
	"github.com/at15/tinycache/proto"
)

type grpcServer struct {
	proto.UnimplementedTinyCacheServer

	cache   cache.Cache
	metrics cache.MetricsExporter
	server  *grpc.Server
}

func NewGRPCServer(cache cache.Cache, metrics cache.MetricsExporter) Server {
	return &grpcServer{
		cache:   cache,
		metrics: metrics,
		// server will be initalized in Start
	}
}

func (s *grpcServer) Start(ctx context.Context, addr string, port int) error {
	addr = fmt.Sprintf("%s:%d", addr, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	proto.RegisterTinyCacheServer(s.server, s)

	return s.server.Serve(lis)
}

func (s *grpcServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	s.server.GracefulStop()
	return nil
}

func (s *grpcServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	b, err := s.cache.Get(req.Bucket, req.Key, cache.Options{})
	if err != nil {
		return nil, err
	}

	return &proto.GetResponse{Value: b}, nil
}

func (s *grpcServer) Set(ctx context.Context, req *proto.SetRequest) (*proto.EmptyResponse, error) {
	err := s.cache.Set(req.Bucket, req.Key, req.Value, cache.Options{
		TTL: time.Duration(req.TtlMs) * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}

	return &proto.EmptyResponse{}, nil
}

func (s *grpcServer) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.EmptyResponse, error) {
	err := s.cache.Delete(req.Bucket, req.Key)
	if err != nil {
		return nil, err
	}

	return &proto.EmptyResponse{}, nil
}
