package utils

import (
	"context"
	"log/slog"
	"os"

	"github.com/lucas-remigio/wallet-tracker/config"
	"go.opentelemetry.io/otel/trace"
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

// LogWithContext returns a logger with the request ID and OTel trace context
func LogWithContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()

	// Add request ID if available
	if id := GetRequestID(ctx); id != "" {
		logger = logger.With("request_id", id)
	}

	// Add OTel trace context if available
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		logger = logger.With(
			"trace_id", span.TraceID().String(),
			"span_id", span.SpanID().String(),
		)
	}

	return logger
}
