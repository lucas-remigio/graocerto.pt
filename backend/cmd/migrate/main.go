package main

import (
	"log"
	"net/url"
	"os"
	"strings"

	mysqlConfig "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/db"
)

func main() {
	log.Println("Starting migration")
	var dbConfig *mysqlConfig.Config

	// Check if we should use remote database
	if config.Envs.IsProduction {
		dbConfig = parseDatabaseUrl(config.Envs.RemoteDBUrl, true)
		log.Println("Using remote database connection")
	} else {
		dbConfig = parseDatabaseUrl(config.Envs.DatabaseUrl, false)
		log.Println("Using local database connection")
	}

	db, err := db.NewMySqlStorage(dbConfig)

	if err != nil {
		log.Fatal(err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
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
		"mysql",
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

// Parse  database URL into MySQL config
func parseDatabaseUrl(dbUrl string, isRemote bool) *mysqlConfig.Config {
	// Parse the URL
	u, err := url.Parse(dbUrl)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}

	// Extract username and password
	userInfo := u.User
	username := userInfo.Username()
	password, _ := userInfo.Password()

	// Extract host and port
	hostPort := u.Host

	// Extract database name
	dbName := strings.TrimPrefix(u.Path, "/")

	// Create MySQL config
	config := &mysqlConfig.Config{
		User:                 username,
		Passwd:               password,
		Addr:                 hostPort,
		DBName:               dbName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	if isRemote {
		// Set the TLS config for remote connections
		config.TLSConfig = "skip-verify"
	}

	return config
}
