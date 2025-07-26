package model

import (
	"context"
	"github.com/AlekSi/pointer"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"house-of-neural-networks/internal/models"
	client "house-of-neural-networks/pkg/api/model"
	"house-of-neural-networks/pkg/logger"
	"net/http"
)

type Service interface {
	CreateModel(ctx context.Context, model models.Model, filename string, content []byte) (*models.Model, error)
	GetModel(ctx context.Context, model models.Model) (*models.Model, error)
	CreateVersion(ctx context.Context, version models.Version, files []models.File) (*models.Version, error)
	DeleteModel(ctx context.Context, model models.Model) (bool, error)
	ListModels(ctx context.Context, userID int64) ([]*models.Model, error)
}

type ModelService struct {
	client.UnimplementedModelServiceServer
	service Service
	ctx     context.Context
}

func NewModelService(ctx context.Context, srv Service) *ModelService {
	return &ModelService{service: srv, ctx: ctx}
}

func (s *ModelService) UploadModel(ctx context.Context, req *client.UploadModelRequest) (*client.UploadModelResponse, error) {
	resp, err := s.service.CreateModel(ctx, models.Model{
		Name:   req.GetName(),
		UserID: req.GetUserId(),
	}, req.GetConfig().GetFilename(), req.GetConfig().GetContent())
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, status.Errorf(codes.Internal, "UploadModel: %s", err)
	}
	r := pointer.Get(resp)
	return &client.UploadModelResponse{
		Id: r.ID,
	}, nil
}

func (s *ModelService) UploadVersion(ctx context.Context, req *client.UploadVersionRequest) (*client.UploadVersionResponse, error) {
	files := make([]models.File, 0)
	for _, file := range req.GetFiles() {
		filename := file.GetFilename()
		content := file.GetContent()

		files = append(files, models.File{Filename: filename, Content: content})
	}

	resp, err := s.service.CreateVersion(ctx, models.Version{
		Number:  req.GetNumber(),
		ModelID: req.GetModelId(),
	}, files)
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, status.Errorf(codes.Internal, "UploadVersion: %s", err)
	}
	r := pointer.Get(resp)
	return &client.UploadVersionResponse{
		Id: r.ID,
	}, nil
}

func (s *ModelService) GetModel(ctx context.Context, req *client.GetModelRequest) (*client.GetModelResponse, error) {
	resp, err := s.service.GetModel(ctx, models.Model{
		ID: req.GetId(),
	})
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, status.Errorf(codes.NotFound, "GetModel: %s", err)
	}
	r := pointer.Get(resp)

	versions := make([]*client.Version, 0)
	for _, version := range resp.Versions {
		v := pointer.Get(version)
		versions = append(versions, &client.Version{
			Id:      v.ID,
			Number:  v.Number,
			ModelId: v.ModelID,
		})
	}

	return &client.GetModelResponse{
		Model: &client.Model{
			Id:       r.ID,
			Name:     r.Name,
			UserId:   r.UserID,
			Versions: versions,
		},
	}, nil
}

func (s *ModelService) UnloadModel(ctx context.Context, req *client.UnloadModelRequest) (*client.UnloadModelResponse, error) {
	resp, err := s.service.DeleteModel(ctx, models.Model{
		ID: req.GetId(),
	})
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, status.Errorf(codes.NotFound, "UnloadModel: %s", err)
	}

	return &client.UnloadModelResponse{
		Success: resp,
	}, nil
}

func (s *ModelService) ListModels(ctx context.Context, req *client.ListModelsRequest) (*client.ListModelsResponse, error) {
	resp, err := s.service.ListModels(ctx, req.GetUserId())
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(
			s.ctx,
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusInternalServerError)),
		)
		return nil, status.Errorf(codes.Internal, "ListModels: %s", err)
	}

	result := make([]*client.Model, 0)
	for _, model := range resp {
		r := pointer.Get(model)
		result = append(result, &client.Model{
			Id:     r.ID,
			Name:   r.Name,
			UserId: r.UserID,
		})
	}

	return &client.ListModelsResponse{
		Models: result,
	}, nil
}
