package gateway

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
	"house-of-neural-networks/internal/transport/gateway/handlers"
	"house-of-neural-networks/internal/transport/grpc_clients"
	pb "house-of-neural-networks/pkg/api/auth"
	"house-of-neural-networks/pkg/logger"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Router struct {
	muxRouter   *mux.Router
	authClient  *grpc_clients.AuthClient
	modelClient *grpc_clients.ModelClient
	ctx         context.Context
}

func NewRouter(ctx context.Context, messageClient *grpc_clients.MessageClient, authClient *grpc_clients.AuthClient, modelClient *grpc_clients.ModelClient) *Router {
	muxRouter := mux.NewRouter()
	r := &Router{muxRouter: muxRouter, ctx: ctx, authClient: authClient, modelClient: modelClient}
	r.muxRouter.Use(r.RequestIDMiddleware, r.loggingMiddleware, r.authMiddleware)

	r.muxRouter.PathPrefix("/docs/").Handler(httpSwagger.WrapHandler)

	// Auth-service routes
	authHandlers := handlers.NewAuthHandlers(authClient)
	r.muxRouter.HandleFunc("/signup", authHandlers.SignUp).Methods(http.MethodPost)
	r.muxRouter.HandleFunc("/login", authHandlers.LogIn).Methods(http.MethodPost)

	// Model-service routes
	modelHandlers := handlers.NewModelHandlers(modelClient)
	r.muxRouter.HandleFunc("/models/{id:[0-9]+}", modelHandlers.GetModel).Methods(http.MethodGet)
	r.muxRouter.HandleFunc("/models", modelHandlers.ListModels).Methods(http.MethodGet)
	r.muxRouter.HandleFunc("/models", modelHandlers.UploadModel).Methods(http.MethodPost)
	r.muxRouter.HandleFunc("/models/version", modelHandlers.UploadVersion).Methods(http.MethodPost)
	r.muxRouter.HandleFunc("/models", modelHandlers.UnloadModel).Methods(http.MethodDelete)

	// Message-service routes
	messageHandlers := handlers.NewMessageHandlers(messageClient)
	r.muxRouter.HandleFunc("/chat/{model_id:[0-9]+}", messageHandlers.GetMessages).Methods("GET")
	r.muxRouter.HandleFunc("/chat/{model_id:[0-9]+}/{version_id:[0-9]+}", messageHandlers.SendMessage).Methods("POST")
	//muxRouter.HandleFunc("/chat", messageHandlers.ListChats).Methods("GET")

	return r
}

func (s *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.GetLoggerFromCtx(r.Context()).Info(r.Context(), fmt.Sprintf("Incoming request: %s %s", r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
	})
}

func (s *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if parts[1] == "signup" || parts[1] == "login" || parts[1] == "docs" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusUnauthorized)
			return
		}
		result, err := s.authClient.ValidateToken(s.ctx, &pb.ValidateTokenRequest{Jwt: cookie.Value})
		if !result.GetValid() {
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				Path:     "/",
			})
			http.Redirect(w, r, "/login", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Router) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), logger.RequestID, requestID)
		ctx = context.WithValue(ctx, logger.LoggerKey, s.ctx.Value(logger.LoggerKey))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
