package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lucas-remigio/wallet-tracker/cmd/api"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

func main() {
	utils.InitLogger()

	var dbURL string

	if config.Envs.IsProduction {
		dbURL = config.Envs.RemoteDBUrl + "?sslmode=verify-ca&sslrootcert=db/prod-ca-2021.crt"
		slog.Info("Using remote database connection")
	} else {
		dbURL = config.Envs.DatabaseUrl + "?sslmode=disable"
		slog.Info("Using local database connection")
	}

	// Open the Postgres database connection
	pgdb, err := sql.Open("pgx", dbURL)
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
