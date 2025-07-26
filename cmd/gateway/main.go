package main

import (
	"context"
	"fmt"
	_ "house-of-neural-networks/docs"
	"house-of-neural-networks/internal/config"
	"house-of-neural-networks/internal/transport/gateway"
	"house-of-neural-networks/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	serviceName = "gateway"
)

// @title House of neural networks API
// @version 1.0
// @description This is the API documentation for the services.
// @host localhost:80
// @BasePath /
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mainLogger := logger.New(serviceName)
	if mainLogger == nil {
		panic("failed to create logger")
	}

	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	cfg := config.New()
	if cfg == nil {
		mainLogger.Fatal(ctx, "failed to load config")
		return
	}

	gtw, err := gateway.New(ctx, cfg)
	if err != nil {
		mainLogger.Fatal(ctx, fmt.Sprintf("failed to init gateway: %s", err.Error()))
		return
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mainLogger.Info(ctx, "Starting gateway...")
		mainLogger.Info(ctx, fmt.Sprintf("Listening on :%d", cfg.HTTPServerPort))
		if err = gtw.Run(); err != nil {
			mainLogger.Error(ctx, fmt.Sprintf("gateway stopped with error: %s", err.Error()))
		}
		cancel()
	}()

	<-graceCh
	mainLogger.Info(ctx, "Shutting down gracefully...")

	// Таймаут для graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err = gtw.Shutdown(shutdownCtx); err != nil {
		mainLogger.Error(ctx, fmt.Sprintf("failed to shutdown gateway gracefully: %s", err.Error()))
	} else {
		mainLogger.Info(ctx, "Gateway shutdown completed.")
	}
}
