package main

import (
	"log/slog"
	"os"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	pgxv5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lucas-remigio/wallet-tracker/config"
)

func main() {
	slog.Info("Starting migration")
	// Choose the correct database URL
	var dbURL string
	if config.Envs.IsProduction {
		dbURL = config.Envs.RemoteDBUrl + "?sslmode=verify-ca&sslrootcert=db/prod-ca-2021.crt"
		slog.Info("Using remote database connection")
	} else {
		dbURL = config.Envs.DatabaseUrl + "?sslmode=disable"
		slog.Info("Using local database connection")
	}

	// Open the database connection
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create the migration driver
	driver, err := pgxv5.WithInstance(db, &pgxv5.Config{})
	if err != nil {
		slog.Error("failed to create migration driver", "error", err)
		os.Exit(1)
	}

	// Get migrations path from environment variable or use default
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://cmd/migrate/migrations"
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"pgx",
		driver,
	)
	if err != nil {
		slog.Error("failed to create migration instance", "error", err)
		os.Exit(1)
	}

	cmd := os.Args[(len(os.Args) - 1)]

	if cmd == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			slog.Error("failed to run migration up", "error", err)
			os.Exit(1)
		}
	}

	if cmd == "down" {
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			slog.Error("failed to run migration down", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("migration completed successfully", "command", cmd)
}
