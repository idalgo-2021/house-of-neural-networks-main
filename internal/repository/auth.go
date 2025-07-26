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

type AuthRepository struct {
	db *postgres.DB
}

func NewAuthRepository(db *postgres.DB) *AuthRepository {
	return &AuthRepository{db}
}

func (s *AuthRepository) CreateUser(ctx context.Context, user models.User) (bool, error) {
	result, err := squirrel.Insert("users").
		Columns("username", "password", "email").
		Values(user.Username, user.Password, user.Email).
		Suffix(`ON CONFLICT (username) DO NOTHING`).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		ExecContext(ctx)

	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("repository.CreateUser: %s", err.Error()))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("repository.CreateUser: %s", err.Error()))
	}

	if rowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (s *AuthRepository) GetUser(ctx context.Context, user models.User) (models.User, error) {
	var result models.User
	err := squirrel.Select("id", "username", "password").
		From("users").
		Where(squirrel.Eq{"username": user.Username}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryRowContext(ctx).
		Scan(&result.ID, &result.Username, &result.Password)

	if err != nil {
		return models.User{}, status.Error(codes.Internal, fmt.Sprintf("repository.GetUser: %s", err.Error()))
	}
	return result, nil
}
