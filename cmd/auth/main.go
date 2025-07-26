package main

import (
	"context"
	"house-of-neural-networks/internal/config"
	"house-of-neural-networks/internal/repository"
	"house-of-neural-networks/internal/service"
	"house-of-neural-networks/internal/transport/grpc/auth"
	"house-of-neural-networks/pkg/db/postgres"
	"house-of-neural-networks/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

const (
	serviceName = "auth"
)

func main() {
	ctx := context.Background()
	mainLogger := logger.New(serviceName)
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	cfg := config.New()
	if cfg == nil {
		mainLogger.Fatal(ctx, "failed to initialize config") // аналог panic()
		return
	}

	db, err := postgres.New(cfg.Config)
	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
	}

	repo := repository.NewAuthRepository(db)
	serv := service.NewAuthService(repo, cfg.JWTSecret)

	grpcServer, err := auth.New(ctx, cfg.GRPCServerPort, serv)
	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = grpcServer.Start(ctx); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	<-graceCh
	grpcServer.Stop(ctx)
	mainLogger.Info(ctx, "Server Stopped")
}
