package utils

import (
	"context"
	"log/slog"
	"os"

	"github.com/lucas-remigio/wallet-tracker/config"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// InitLogger initializes the global slog logger based on the environment
func InitLogger() {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if config.Envs.IsProduction {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// GetRequestID returns the request ID from the context if it exists
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// LogWithContext returns a logger with the request ID from the context
func LogWithContext(ctx context.Context) *slog.Logger {
	id := GetRequestID(ctx)
	if id != "" {
		return slog.With("request_id", id)
	}
	return slog.Default()
}
