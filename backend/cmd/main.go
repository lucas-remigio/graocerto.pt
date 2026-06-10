package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lucas-remigio/wallet-tracker/cmd/api"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/utils"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	utils.InitLogger()

	// Handle SIGINT (CTRL+C) and SIGTERM gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize OpenTelemetry
	otelShutdown, err := utils.InitOTel(ctx, "wallet-tracker", config.Envs.OTelCollectorEndpoint)
	if err != nil {
		slog.Error("Failed to initialize OpenTelemetry", "error", err)
		// We continue even if OTel fails, but in production, you might want to exit.
	} else {
		defer func() {
			slog.Info("Shutting down OpenTelemetry")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := otelShutdown(shutdownCtx); err != nil {
				slog.Error("Failed to shutdown OpenTelemetry", "error", err)
			}
		}()
	}

	var dbURL string

	if config.Envs.IsProduction {
		dbURL = config.Envs.RemoteDBUrl + "?sslmode=verify-ca&sslrootcert=db/prod-ca-2021.crt"
		slog.Info("Using remote database connection")
	} else {
		dbURL = config.Envs.DatabaseUrl + "?sslmode=disable"
		slog.Info("Using local database connection")
	}

	// Open the Postgres database connection
	pgdb, err := otelsql.Open("pgx", dbURL,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName("wallet_tracker"),
	)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer pgdb.Close()

	initStorage(pgdb)

	port := config.Envs.Port
	if port == "" {
		port = "8080"
		slog.Warn("No port specified, using default port 8080")
	}
	addr := ":" + port

	server := api.NewAPIServer(addr, pgdb)

	// Use the server's Run method which will handle HTTP/HTTPS internally
	if err := server.Run(); err != nil {
		slog.Error("Server failed to run", "error", err)
		os.Exit(1)
	}
}

func initStorage(db *sql.DB) {
	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}

	// Check if SSL is being used
	var sslStatus string
	err := db.QueryRow("SHOW ssl").Scan(&sslStatus)
	if err != nil {
		slog.Warn("Could not check SSL status", "error", err)
	} else {
		slog.Info("Database SSL Status", "status", sslStatus)
	}

	slog.Info("DB: Successfully connected")
}
