package message

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	client "house-of-neural-networks/pkg/api/message"
	"house-of-neural-networks/pkg/logger"
	"net/http"
)

type Service interface {
	ProcessMessage(ctx context.Context, userID, modelID, versionID int64, inputs []*client.Input) ([]string, error)
	GetMessages(ctx context.Context, userID, modelID int64) ([]*client.Message, error)
}

type MessageService struct {
	client.UnimplementedMessageServiceServer
	service Service
	ctx     context.Context
}

func NewMessageService(ctx context.Context, srv Service) *MessageService {
	return &MessageService{service: srv, ctx: ctx}
}

func (s *MessageService) SendMessage(ctx context.Context, req *client.SendMessageRequest) (*client.SendMessageResponse, error) {
	results, err := s.service.ProcessMessage(ctx, req.GetUserId(), req.GetModelId(), req.GetVersionId(), req.GetInputs())
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, status.Errorf(codes.Unknown, "SendMessage: %s", err)
	}

	return &client.SendMessageResponse{
		Results: results,
	}, nil
}

func (s *MessageService) GetMessages(ctx context.Context, req *client.GetMessagesRequest) (*client.GetMessagesResponse, error) {
	messages, err := s.service.GetMessages(ctx, req.GetUserId(), req.GetModelId())
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, status.Errorf(codes.Unknown, "GetMessages: %s", err)
	}

	return &client.GetMessagesResponse{
		Messages: messages,
	}, nil
}
