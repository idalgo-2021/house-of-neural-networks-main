package logger

import (
	"context"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

const (
	LoggerKey   = "logger"
	RequestID   = "requestID"
	ServiceName = "service"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
}

type logger struct {
	serviceName string
	logger      *zap.Logger
}

func (l logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}

	l.logger.Info(msg, fields...)
}

func (l logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}

	l.logger.Error(msg, fields...)
}

func (l logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}

	l.logger.Fatal(msg, fields...)
}

func New(serviceName string) Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	return &logger{
		serviceName: serviceName,
		logger:      zapLogger,
	}
}

func GetLoggerFromCtx(ctx context.Context) Logger {
	log := ctx.Value(LoggerKey)
	if log != nil {
		return log.(Logger)
	}
	return New("Test")
}

func GetFunctionName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}
	fullName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(fullName, "/")
	return parts[len(parts)-1]
}
