package tests

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"house-of-neural-networks/internal/repository"
	"house-of-neural-networks/internal/service"
	"house-of-neural-networks/internal/transport/grpc/model"
	client "house-of-neural-networks/pkg/api/model"
	"house-of-neural-networks/pkg/db/postgres"
	"log"
	"regexp"
	"testing"
)

func TestGetModel_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	rows := sqlmock.NewRows([]string{"models.id", "models.name", "models.user_id", "versions.id as version_id", "versions.number as version_number", "versions.model_id as version_model_id"}).
		AddRow(1, "simple model", 1, 0, 1, 1)
	mock.ExpectQuery("SELECT models.id, models.name, models.user_id, versions.id as version_id, versions.number as version_number, versions.model_id as version_model_id FROM models LEFT JOIN versions ON models.id = versions.model_id WHERE models.id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := modelService.GetModel(context.Background(), &client.GetModelRequest{Id: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.GetModel().GetId())
		assert.Equal(t, "simple model", resp.GetModel().GetName())
		assert.Equal(t, int64(1), resp.GetModel().GetUserId())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetModel_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectQuery("SELECT models.id, models.name, models.user_id, versions.id as version_id, versions.number as version_number, versions.model_id as version_model_id FROM models LEFT JOIN versions ON models.id = versions.model_id WHERE models.id = \\$1").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Not Found", func(t *testing.T) {
		resp, err := modelService.GetModel(context.Background(), &client.GetModelRequest{Id: 1})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadModel_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectQuery("INSERT INTO models").
		WithArgs("test model", int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "user_id"}).
			AddRow(1, "test model", 1))

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := modelService.UploadModel(context.Background(), &client.UploadModelRequest{Name: "test model", UserId: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.GetId())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadModel_IncorrectData(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Incorrect data", func(t *testing.T) {
		resp, err := modelService.UploadModel(context.Background(), &client.UploadModelRequest{})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "name or user_id is empty")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListModels_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "user_id"}).
		AddRow(1, "test model 1", 1).
		AddRow(2, "test model 2", 1).
		AddRow(3, "test model 3", 1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM models WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnRows(rows)

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := modelService.ListModels(context.Background(), &client.ListModelsRequest{UserId: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.GetModels()[0].GetId())
		assert.Equal(t, "test model 1", resp.GetModels()[0].GetName())
		assert.Equal(t, int64(1), resp.GetModels()[0].GetUserId())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListModels_IncorrectData(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := modelService.ListModels(context.Background(), &client.ListModelsRequest{})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user_id is empty")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteModel_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM models WHERE id = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := modelService.UnloadModel(context.Background(), &client.UnloadModelRequest{Id: 1})
		require.NoError(t, err)
		assert.Equal(t, true, resp.GetSuccess())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteModel_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM models WHERE id = $1`)).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Not found", func(t *testing.T) {
		resp, err := modelService.UnloadModel(context.Background(), &client.UnloadModelRequest{Id: 1})
		require.Error(t, err)
		assert.Equal(t, false, resp.GetSuccess())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteModel_IncorrectData(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Incorrect data", func(t *testing.T) {
		resp, err := modelService.UnloadModel(context.Background(), &client.UnloadModelRequest{})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "id is empty")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadVersion_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectQuery("INSERT INTO versions").
		WithArgs(1, int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "number", "model_id"}).
			AddRow(1, 1, 1))

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := modelService.UploadVersion(context.Background(), &client.UploadVersionRequest{Number: 1, ModelId: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.GetId())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadVersion_IncorrectData(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewModelRepository(&postgres.DB{Db: db})
	serv := service.NewModelService(repo, true)
	modelService := model.NewModelService(ctx, serv)

	t.Run("Incorrect data", func(t *testing.T) {
		resp, err := modelService.UploadVersion(context.Background(), &client.UploadVersionRequest{})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "number or model_id is empty")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
