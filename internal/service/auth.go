package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"house-of-neural-networks/internal/models"
	"time"
)

type AuthRepo interface {
	CreateUser(ctx context.Context, user models.User) (bool, error)
	GetUser(ctx context.Context, user models.User) (models.User, error)
}

type AuthService struct {
	Repo      AuthRepo
	JWTSecret string
}

type CustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthService(repo AuthRepo, secret string) *AuthService {
	return &AuthService{Repo: repo, JWTSecret: secret}
}

func NewCustomClaims(userId int64, claims jwt.RegisteredClaims) *CustomClaims {
	return &CustomClaims{UserID: userId, RegisteredClaims: claims}
}

func (s *AuthService) SignUp(ctx context.Context, user models.User) (bool, error) {
	if user.Password == "" || user.Username == "" {
		return false, status.Error(codes.InvalidArgument, "username or password is empty")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("service.SingUp: %s", err.Error()))
	}
	user.Password = string(bytes)
	return s.Repo.CreateUser(ctx, user)
}

func (s *AuthService) LogIn(ctx context.Context, user models.User) (string, int64, error) {
	if user.Password == "" || user.Username == "" {
		return "", 0, status.Error(codes.InvalidArgument, "username or password is empty")
	}

	result, err := s.Repo.GetUser(ctx, user)
	if err != nil {
		return "", 0, status.Error(codes.Internal, fmt.Sprintf("service.LogIn: %s", err.Error()))
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		return "", 0, status.Error(codes.Unauthenticated, "password does not match")
	}

	token, err := s.GenerateToken(result)
	if err != nil {
		return "", 0, status.Error(codes.Internal, fmt.Sprintf("service.LogIn: %s", err.Error()))
	}

	return token, result.ID, nil
}

func (s *AuthService) GenerateToken(user models.User) (string, error) {
	claims := NewCustomClaims(
		user.ID,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "auth-service",
		})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("service.GenerateToken: %s", err.Error()))
	}

	return signedToken, nil
}

func (s *AuthService) ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.JWTSecret), nil
	})

	if err != nil {
		return false, status.Error(codes.Internal, fmt.Sprintf("service.ValidateToken: %s", err.Error()))
	}

	_, ok := token.Claims.(*CustomClaims)

	if !ok {
		return false, status.Error(codes.Unauthenticated, "invalid token claims")
	}

	if !token.Valid {
		return false, status.Error(codes.Unauthenticated, "invalid token")
	}

	return true, nil
}
