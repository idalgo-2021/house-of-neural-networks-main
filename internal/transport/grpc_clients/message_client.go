package grpc_clients

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"house-of-neural-networks/pkg/logger"
	"net/http"

	pb "house-of-neural-networks/pkg/api/message"

	"google.golang.org/grpc"
)

type MessageClient struct {
	client pb.MessageServiceClient
}

func NewMessageClient(addr string) (*MessageClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Message-service: %w", err)
	}
	return &MessageClient{
		client: pb.NewMessageServiceClient(conn),
	}, nil
}

func (c *MessageClient) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	resp, err := c.client.SendMessage(ctx, req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(
			ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
	}
	return resp, err
}

//func (c *MessageClient) GetMSG(ctx context.Context, req *pb.GetMSGRequest) (*pb.GetMSGResponse, error) {
//	return c.client.GetMSG(ctx, req)
//}

func (c *MessageClient) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	resp, err := c.client.GetMessages(ctx, req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(
			ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
	}
	return resp, err
}
