package auth

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"house-of-neural-networks/internal/models"
	client "house-of-neural-networks/pkg/api/auth"
	"house-of-neural-networks/pkg/logger"
	"net/http"
)

type Service interface {
	SignUp(ctx context.Context, user models.User) (bool, error)
	LogIn(ctx context.Context, user models.User) (string, int64, error)
	ValidateToken(jwt string) (bool, error)
}

type AuthService struct {
	client.UnimplementedAuthServiceServer
	service Service
	ctx     context.Context
}

func NewAuthService(ctx context.Context, srv Service) *AuthService {
	return &AuthService{service: srv, ctx: ctx}
}

func (s *AuthService) SignUp(ctx context.Context, req *client.SignUpRequest) (*client.SignUpResponse, error) {
	success, err := s.service.SignUp(ctx, models.User{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Email:    req.GetEmail(),
	})
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, fmt.Errorf("CreateUser: %w", err)
	}

	return &client.SignUpResponse{
		Success: success,
	}, nil
}

func (s *AuthService) LogIn(ctx context.Context, req *client.LogInRequest) (*client.LogInResponse, error) {
	token, userId, err := s.service.LogIn(ctx, models.User{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	})
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, fmt.Errorf("ValidateUser: %w", err)
	}
	return &client.LogInResponse{
		Jwt:    token,
		UserId: userId,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, req *client.ValidateTokenRequest) (*client.ValidateTokenResponse, error) {
	valid, err := s.service.ValidateToken(req.GetJwt())
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, fmt.Errorf("ValidateToken: %w", err)
	}

	return &client.ValidateTokenResponse{
		Valid: valid,
	}, nil
}
