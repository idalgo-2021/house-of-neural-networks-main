package tests

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"house-of-neural-networks/internal/repository"
	"house-of-neural-networks/internal/service"
	"house-of-neural-networks/internal/transport/grpc/auth"
	client "house-of-neural-networks/pkg/api/auth"
	"house-of-neural-networks/pkg/db/postgres"
	"log"
	"regexp"
	"testing"
)

func TestLogIn_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	password, _ := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	rows := sqlmock.NewRows([]string{"id", "username", "password"}).
		AddRow(1, "test user", password)
	mock.ExpectQuery("SELECT id, username, password FROM users WHERE username = \\$1").
		WithArgs("test user").
		WillReturnRows(rows)

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewAuthRepository(&postgres.DB{Db: db})
	serv := service.NewAuthService(repo, "very-secret-key")
	authService := auth.NewAuthService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := authService.LogIn(context.Background(), &client.LogInRequest{Username: "test user", Password: "123"})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogIn_UserNotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectQuery("SELECT id, username, password FROM users WHERE username = \\$1").
		WithArgs("test user").
		WillReturnError(sql.ErrNoRows)

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewAuthRepository(&postgres.DB{Db: db})
	serv := service.NewAuthService(repo, "very-secret-key")
	authService := auth.NewAuthService(ctx, serv)

	t.Run("User not found", func(t *testing.T) {
		resp, err := authService.LogIn(context.Background(), &client.LogInRequest{Username: "test user", Password: "123"})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogIn_PasswordDoesNotMatch(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	password, _ := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	rows := sqlmock.NewRows([]string{"id", "username", "password"}).
		AddRow(1, "test user", password)
	mock.ExpectQuery("SELECT id, username, password FROM users WHERE username = \\$1").
		WithArgs("test user").
		WillReturnRows(rows)

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewAuthRepository(&postgres.DB{Db: db})
	serv := service.NewAuthService(repo, "very-secret-key")
	authService := auth.NewAuthService(ctx, serv)

	t.Run("Password does not match", func(t *testing.T) {
		resp, err := authService.LogIn(context.Background(), &client.LogInRequest{Username: "test user", Password: "wrong_password"})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "password does not match")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogIn_IncorrectData(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewAuthRepository(&postgres.DB{Db: db})
	serv := service.NewAuthService(repo, "very-secret-key")
	authService := auth.NewAuthService(ctx, serv)

	t.Run("Incorrect data", func(t *testing.T) {
		resp, err := authService.LogIn(context.Background(), &client.LogInRequest{Username: "", Password: ""})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "username or password is empty")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSignUp_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users`)).
		WithArgs("test user", sqlmock.AnyArg(), "test@test.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewAuthRepository(&postgres.DB{Db: db})
	serv := service.NewAuthService(repo, "very-secret-key")
	authService := auth.NewAuthService(ctx, serv)

	t.Run("Success", func(t *testing.T) {
		resp, err := authService.SignUp(context.Background(), &client.SignUpRequest{Username: "test user", Password: "123", Email: "test@test.com"})
		require.NoError(t, err)
		assert.Equal(t, true, resp.GetSuccess())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSignUp_IncorrectData(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	repo := repository.NewAuthRepository(&postgres.DB{Db: db})
	serv := service.NewAuthService(repo, "very-secret-key")
	authService := auth.NewAuthService(ctx, serv)

	t.Run("Incorrect data", func(t *testing.T) {
		resp, err := authService.SignUp(context.Background(), &client.SignUpRequest{Username: "", Password: "", Email: ""})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "username or password is empty")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
