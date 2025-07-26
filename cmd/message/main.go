package main

import (
	"context"
	"fmt"
	"house-of-neural-networks/internal/config"
	"house-of-neural-networks/internal/repository"
	"house-of-neural-networks/internal/service"
	"house-of-neural-networks/internal/transport/grpc/message"
	"house-of-neural-networks/internal/triton"
	"house-of-neural-networks/pkg/db/postgres"
	"house-of-neural-networks/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

const (
	serviceName = "message"
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

	tritonClient := triton.NewTritonClient(cfg.TritonConfig.Host, cfg.TritonConfig.Port)
	serverLiveResponse := triton.ServerLiveRequest(tritonClient.Client)
	mainLogger.Info(ctx, fmt.Sprintf("Triton Health - Live: %v", serverLiveResponse.Live))
	serverReadyResponse := triton.ServerReadyRequest(tritonClient.Client)
	mainLogger.Info(ctx, fmt.Sprintf("Triton Health - Ready: %v", serverReadyResponse.Ready))

	repo := repository.NewMessageRepository(db)
	serv := service.NewMessageService(repo, tritonClient)

	grpcServer, err := message.New(ctx, cfg.GRPCServerPort, serv)
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
