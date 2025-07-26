package model

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	interceptor "house-of-neural-networks/internal/transport/grpc"
	client "house-of-neural-networks/pkg/api/model"
	"house-of-neural-networks/pkg/logger"
	"log"
	"net"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func New(ctx context.Context, port int, service Service) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.ContextWithLogger(logger.GetLoggerFromCtx(ctx))),
	}
	grpcServer := grpc.NewServer(opts...)
	client.RegisterModelServiceServer(grpcServer, NewModelService(ctx, service))

	return &Server{grpcServer: grpcServer, listener: lis}, nil
}

func (s *Server) Start(ctx context.Context) error {
	eg := errgroup.Group{}

	eg.Go(func() error {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "starting gRPC server", zap.Int("port", s.listener.Addr().(*net.TCPAddr).Port))
		return s.grpcServer.Serve(s.listener)
	})

	return eg.Wait()
}

func (s *Server) Stop(ctx context.Context) {
	s.grpcServer.GracefulStop()
	l := logger.GetLoggerFromCtx(ctx)
	if l != nil {
		l.Info(ctx, "gRPC server stopped")
	}
}
