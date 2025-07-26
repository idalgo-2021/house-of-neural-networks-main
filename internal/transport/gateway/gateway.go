package gateway

import (
	"context"
	"errors"
	"fmt"
	"house-of-neural-networks/internal/config"
	"house-of-neural-networks/internal/transport/grpc_clients"
	"net/http"
	"strconv"
)

type Gateway struct {
	httpServer *http.Server
}

func New(ctx context.Context, cfg *config.Config) (*Gateway, error) {
	authClient, err := grpc_clients.NewAuthClient(cfg.AuthServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}

	modelClient, err := grpc_clients.NewModelClient(cfg.ModelServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create model client: %w", err)
	}

	messageClient, err := grpc_clients.NewMessageClient(cfg.MessageServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create message client: %w", err)
	}

	r := NewRouter(ctx, messageClient, authClient, modelClient)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", strconv.Itoa(cfg.HTTPServerPort)),
		Handler: r.muxRouter,
	}

	return &Gateway{
		httpServer: httpServer,
	}, nil
}

func (g *Gateway) Run() error {
	if err := g.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to run HTTP server: %w", err)
	}
	return nil
}

func (g *Gateway) Shutdown(ctx context.Context) error {
	if g.httpServer != nil {
		return g.httpServer.Shutdown(ctx)
	}
	return nil
}
