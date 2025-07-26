package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"house-of-neural-networks/internal/models"
	"house-of-neural-networks/internal/triton"
	"os"
)

type ModelRepo interface {
	CreateModel(ctx context.Context, model models.Model) (*models.Model, error)
	GetModel(ctx context.Context, model models.Model) (*models.Model, error)
	DeleteModel(ctx context.Context, model models.Model) (bool, error)
	CreateVersion(ctx context.Context, version models.Version) (*models.Version, error)
	ListModels(ctx context.Context, userID int64) ([]*models.Model, error)
}

type ModelService struct {
	Repo         ModelRepo
	TritonClient *triton.TritonClient
}

func NewModelService(repo ModelRepo, tritonClient *triton.TritonClient) *ModelService {
	return &ModelService{repo, tritonClient}
}

func (s *ModelService) CreateModel(ctx context.Context, model models.Model, filename string, content []byte) (*models.Model, error) {
	res, err := s.Repo.CreateModel(ctx, model)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(fmt.Sprintf("/models/%s", res.Name), os.ModePerm)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("service.UploadModel: %s", err.Error()))
	}

	err = os.WriteFile(fmt.Sprintf("/models/%s/%s", res.Name, filename), content, 0644)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("service.UploadModel: failed to save config %s: %v", filename, err))
	}

	return res, nil
}

func (s *ModelService) GetModel(ctx context.Context, model models.Model) (*models.Model, error) {
	return s.Repo.GetModel(ctx, model)
}

func (s *ModelService) CreateVersion(ctx context.Context, version models.Version, files []models.File) (*models.Version, error) {
	res, err := s.Repo.CreateVersion(ctx, version)
	if err != nil {
		return nil, err
	}
	model, err := s.Repo.GetModel(ctx, models.Model{ID: version.ModelID})
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		err = os.MkdirAll(fmt.Sprintf("/models/%s/%d", model.Name, version.Number), os.ModePerm)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("service.UploadVersion: %s", err.Error()))
		}

		err = os.WriteFile(fmt.Sprintf("/models/%s/%d/%s", model.Name, version.Number, file.Filename), file.Content, 0644)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("service.UploadVersion: failed to save file %s: %v", file.Filename, err))
		}
	}

	return res, nil
}

func (s *ModelService) DeleteModel(ctx context.Context, model models.Model) (bool, error) {
	respModel, err := s.Repo.GetModel(ctx, model)
	if err != nil {
		return false, err
	}
	ready, err := triton.ModelReadyRequest(s.TritonClient.Client, respModel.Name, "")
	if err != nil {
		return false, err
	}
	if ready {
		err = triton.UnloadModelRequest(s.TritonClient.Client, respModel.Name)
		if err != nil {
			return false, err
		}
	}
	return s.Repo.DeleteModel(ctx, model)
}

func (s *ModelService) ListModels(ctx context.Context, userID int64) ([]*models.Model, error) {
	return s.Repo.ListModels(ctx, userID)
}
