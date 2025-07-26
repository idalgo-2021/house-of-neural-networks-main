package grpc_clients

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"house-of-neural-networks/pkg/logger"
	"net/http"

	pb "house-of-neural-networks/pkg/api/auth"

	"google.golang.org/grpc"
)

type AuthClient struct {
	client pb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AuthService: %w", err)
	}
	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	response, err := c.client.SignUp(ctx, req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(
			ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
	}
	return response, err
}

func (c *AuthClient) LogIn(ctx context.Context, req *pb.LogInRequest) (*pb.LogInResponse, error) {
	response, err := c.client.LogIn(ctx, req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(
			ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
	}
	return response, err
}

func (c *AuthClient) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	response, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(
			ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
	}
	return response, err
}
