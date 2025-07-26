package grpc_clients

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"house-of-neural-networks/pkg/logger"
	"net/http"

	pb "house-of-neural-networks/pkg/api/model"

	"google.golang.org/grpc"
)

type ModelClient struct {
	client pb.ModelServiceClient
}

func NewModelClient(addr string) (*ModelClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ModelService: %w", err)
	}
	return &ModelClient{
		client: pb.NewModelServiceClient(conn),
	}, nil
}

func (c *ModelClient) GetModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error) {
	response, err := c.client.GetModel(ctx, req)
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

func (c *ModelClient) ListModels(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	response, err := c.client.ListModels(ctx, req)
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

func (c *ModelClient) UploadModel(ctx context.Context, req *pb.UploadModelRequest) (*pb.UploadModelResponse, error) {
	response, err := c.client.UploadModel(ctx, req)
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

func (c *ModelClient) UploadVersion(ctx context.Context, req *pb.UploadVersionRequest) (*pb.UploadVersionResponse, error) {
	response, err := c.client.UploadVersion(ctx, req)
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

func (c *ModelClient) UnloadModel(ctx context.Context, req *pb.UnloadModelRequest) (*pb.UnloadModelResponse, error) {
	response, err := c.client.UnloadModel(ctx, req)
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
