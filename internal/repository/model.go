package repository

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"house-of-neural-networks/internal/models"
	"house-of-neural-networks/pkg/db/postgres"
)

type ModelRepository struct {
	db *postgres.DB
}

func NewModelRepository(db *postgres.DB) *ModelRepository {
	return &ModelRepository{db}
}

func (s *ModelRepository) CreateModel(ctx context.Context, model models.Model) (*models.Model, error) {
	var result models.Model
	err := squirrel.Insert("models").
		Columns("name", "user_id").
		Values(model.Name, model.UserID).
		Suffix("returning *").
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryRowContext(ctx).
		Scan(&result.ID, &result.Name, &result.UserID)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("repository.CreateModel: %s", err.Error()))
	}

	return &result, nil
}

func (s *ModelRepository) GetModel(ctx context.Context, model models.Model) (*models.Model, error) {
	rows, err := squirrel.Select("models.id", "models.name", "models.user_id", "versions.id as version_id", "versions.number as version_number", "versions.model_id as version_model_id").
		From("models").
		LeftJoin("versions ON models.id = versions.model_id").
		Where(squirrel.Eq{"models.id": model.ID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("repository.GetModel: %s", err.Error()))
	}
	defer rows.Close()

	var result models.Model

	for rows.Next() {
		var version models.Version
		var versionID, VersionModelId *int64
		var versionNumber *int32
		if err = rows.Scan(&result.ID, &result.Name, &result.UserID, &versionID, &versionNumber, &VersionModelId); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("repository.GetModel: %s", err.Error()))
		}
		if versionID != nil && versionNumber != nil && VersionModelId != nil {
			version.ID = *versionID
			version.Number = *versionNumber
			version.ModelID = *VersionModelId
			result.Versions = append(result.Versions, &version)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("repository.GetModel: %s", err.Error()))
	}

	return &result, nil

}

func (s *ModelRepository) GetModelName(ctx context.Context, model models.Model) (string, error) {
	var result string
	err := squirrel.Select("name").
		From("models").
		Where(squirrel.Eq{"id": model.ID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryRow().
		Scan(&result)

	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("repository.GetModel: %s", err.Error()))
	}

	return result, nil
}

func (s *ModelRepository) DeleteModel(ctx context.Context, model models.Model) (bool, error) {
	_, err := squirrel.Delete("messages").
		Where(squirrel.Eq{"model_id": model.ID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		ExecContext(ctx)

	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("repository.DeleteModel: failed to delete related messages: %s", err.Error()))
	}

	result, err := squirrel.Delete("models").
		Where(squirrel.Eq{"id": model.ID}).
		PlaceholderFormat(squirrel.Dollar).
		Suffix("returning *").
		RunWith(s.db.Db).
		ExecContext(ctx)

	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("repository.DeleteModel: %s", err.Error()))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("repository.DeleteModel: %s", err.Error()))
	}
	if rowsAffected == 0 {
		return false, status.Error(codes.Internal, fmt.Sprintf("repository.DeleteModel: order (id %d) not found", model.ID))
	}

	return true, nil
}

func (s *ModelRepository) CreateVersion(ctx context.Context, version models.Version) (*models.Version, error) {
	var result models.Version
	err := squirrel.Insert("versions").
		Columns("number", "model_id").
		Values(version.Number, version.ModelID).
		Suffix("returning *").
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryRowContext(ctx).
		Scan(&result.ID, &result.Number, &result.ModelID)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("repository.CreateVersion: %s", err.Error()))
	}

	return &result, nil
}

func (s *ModelRepository) ListModels(ctx context.Context, userID int64) ([]*models.Model, error) {
	var result []*models.Model
	rows, err := squirrel.Select("*").
		From("models").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryContext(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("repository.ListModels: %s", err.Error()))
	}
	defer rows.Close()

	for rows.Next() {
		var model models.Model
		if err = rows.Scan(&model.ID, &model.Name, &model.UserID); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("repository.ListModels: %s", err.Error()))
		}
		result = append(result, &model)
	}

	if err = rows.Err(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("repository.ListModels: %s", err.Error()))
	}

	return result, nil
}
