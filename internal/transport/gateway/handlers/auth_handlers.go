package handlers

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"house-of-neural-networks/internal/transport/grpc_clients"
	"house-of-neural-networks/pkg/logger"
	"net/http"
	"time"

	pb "house-of-neural-networks/pkg/api/auth"
)

type AuthHandlers struct {
	client *grpc_clients.AuthClient
}

func NewAuthHandlers(client *grpc_clients.AuthClient) *AuthHandlers {
	return &AuthHandlers{client: client}
}

// SignUp handles user signUp.
// @Summary Регистрация пользователя
// @Description Регистрирует новых пользователей
// @Tags Auth service
// @Accept json
// @Produce json
// @Param register body models.SignUpRequest true "SignUp data"
// @Success 200 {object} models.SignUpResponse
// @Router /signup [post]
func (h *AuthHandlers) SignUp(w http.ResponseWriter, r *http.Request) {
	var req pb.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusBadRequest)),
		)
		return
	}
	req.RequestId = r.Context().Value(logger.RequestID).(string)
	resp, err := h.client.SignUp(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error calling Auth-Service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LogIn handles user logIn.
// @Summary Авторизация пользователя
// @Description Авторизует пользователей
// @Tags Auth service
// @Accept json
// @Produce json
// @Param register body models.LogInRequest true "LogIn data"
// @Success 200 {object} models.LogInResponse
// @Router /login [post]
func (h *AuthHandlers) LogIn(w http.ResponseWriter, r *http.Request) {
	var req pb.LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		logger.GetLoggerFromCtx(r.Context()).Error(
			r.Context(),
			err.Error(),
			zap.String("Function", logger.GetFunctionName()),
			zap.String("Status", http.StatusText(http.StatusBadRequest)),
		)
		return
	}
	req.RequestId = r.Context().Value(logger.RequestID).(string)
	resp, err := h.client.LogIn(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling Auth-Service, LogIn: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    resp.Jwt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    fmt.Sprint(resp.UserId),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
