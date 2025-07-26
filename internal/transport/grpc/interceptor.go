package grpc

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"house-of-neural-networks/pkg/logger"
)

func ContextWithLogger(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		l.Info(ctx, "request started", zap.String("method", info.FullMethod))
		return handler(ctx, req)
	}
}
