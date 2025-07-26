package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"house-of-neural-networks/internal/models"
	"house-of-neural-networks/pkg/db/postgres"
)

type MessageRepository struct {
	db *postgres.DB
}

func NewMessageRepository(db *postgres.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) SaveMessage(ctx context.Context, msg models.Message) error {
	result, err := squirrel.Insert("messages").Columns("user_id", "model_id", "version_id", "input1", "input2", "results", "created_at").
		Values(msg.UserID, msg.ModelID, msg.VersionID, msg.Input1, msg.Input2, pq.Array(msg.Results), msg.CreatedAt).
		PlaceholderFormat(squirrel.Dollar).RunWith(r.db.Db).ExecContext(ctx)

	if err != nil {
		return status.Errorf(codes.Internal, "repository.SaveMessage: %s", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return status.Errorf(codes.Internal, "repository.SaveMessage: %s", err)
	}

	if rowsAffected == 0 {
		return status.Error(codes.Internal, "repository.SaveMessage: Can't insert row")
	}
	return nil
}

func (r *MessageRepository) GetMessages(ctx context.Context, userID, modelID int64) ([]models.Message, error) {
	rows, err := squirrel.Select("id", "user_id", "model_id", "version_id", "input1", "input2", "results", "created_at").
		From("messages").
		Where(squirrel.And{squirrel.Eq{"user_id": userID}, squirrel.Eq{"model_id": modelID}}).
		PlaceholderFormat(squirrel.Dollar).RunWith(r.db.Db).QueryContext(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "repository.GetMessages: %s", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %s", err)
		}
	}(rows)

	var result []models.Message

	for rows.Next() {
		var msg models.Message
		var results pq.StringArray
		if err = rows.Scan(&msg.ID, &msg.UserID, &msg.ModelID, &msg.VersionID, &msg.Input1, &msg.Input2, &results, &msg.CreatedAt); err != nil {
			return nil, status.Errorf(codes.Internal, "repository.GetMessages: %s", err)
		}
		msg.Results = results
		result = append(result, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "repository.GetMessages: %s", err)
	}

	return result, nil
}

func (s *MessageRepository) GetModelName(ctx context.Context, model models.Model) (string, error) {
	var name string
	err := squirrel.Select("name").
		From("models").
		Where(squirrel.Eq{"id": model.ID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryRow().
		Scan(&name)

	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("repository.GetModelName: %s", err))
	}

	return name, nil
}

func (s *MessageRepository) GetVersionNumber(ctx context.Context, version models.Version) (int, error) {
	var number int
	err := squirrel.Select("number").
		From("versions").
		Where(squirrel.Eq{"id": version.ID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.db.Db).
		QueryRow().
		Scan(&number)

	if err != nil {
		return number, status.Error(codes.Internal, fmt.Sprintf("repository.GetModelName: %s", err))
	}
	return number, nil
}
