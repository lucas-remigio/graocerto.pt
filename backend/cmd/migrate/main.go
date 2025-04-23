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
	if remoteUrl := config.Envs.RemoteDBUrl; remoteUrl != "" {
		dbConfig = parseRemoteDbUrl(remoteUrl)
		log.Println("Using remote database connection")
	} else {
		dbConfig = &mysqlConfig.Config{
			User:                 config.Envs.DBUser,
			Passwd:               config.Envs.DBPassword,
			Addr:                 config.Envs.DBAddress,
			DBName:               config.Envs.DBName,
			Net:                  "tcp",
			AllowNativePasswords: true,
			ParseTime:            true,
		}
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

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
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

// Parse remote database URL into MySQL config
func parseRemoteDbUrl(remoteUrl string) *mysqlConfig.Config {
	log.Println("Parsing remote DB URL:", remoteUrl)
	// Parse the URL
	u, err := url.Parse(remoteUrl)
	if err != nil {
		log.Fatalf("Failed to parse remote DB URL: %v", err)
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
		TLSConfig:            "skip-verify", // Enable SSL
	}

	return config
}
