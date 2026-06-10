package utils

import (
	"context"
	"log/slog"
	"os"

	"github.com/lucas-remigio/wallet-tracker/config"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// InitLogger initializes the global slog logger based on the environment
func InitLogger() {
	var stdoutHandler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// 1. Stdout Handler (for Docker logs / CLI)
	if config.Envs.IsProduction {
		stdoutHandler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		stdoutHandler = slog.NewTextHandler(os.Stdout, opts)
	}

	// 2. OTel Handler (for Loki)
	otelHandler := otelslog.NewHandler("wallet-tracker")

	// 3. Combined Multi-Handler
	logger := slog.New(newMultiHandler(stdoutHandler, otelHandler))
	slog.SetDefault(logger)
}

// multiHandler implements slog.Handler and fans out to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range m.handlers {
		_ = h.Handle(ctx, record)
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
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
