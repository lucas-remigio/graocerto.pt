package main

import (
	"log"
	"os"

	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/lucas-remigio/wallet-tracker/config"
)

func main() {
	log.Println("Starting migration")
	// Choose the correct database URL
	var dbURL string
	if config.Envs.IsProduction {
		dbURL = config.Envs.RemoteDBUrl + "?sslmode=verify-ca&sslrootcert=db/prod-ca-2021.crt"
		log.Println("Using remote database connection")
	} else {
		dbURL = config.Envs.DatabaseUrl + "?sslmode=disable"
		log.Println("Using local database connection")
	}

	// Open the database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Get migrations path from environment variable or use default
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://cmd/migrate/migrations"
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	cmd := os.Args[(len(os.Args) - 1)]

	if cmd == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}

	if cmd == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}
}
