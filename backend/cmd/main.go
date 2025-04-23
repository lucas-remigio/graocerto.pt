package main

import (
	"database/sql"
	"log"
	"net/url"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/lucas-remigio/wallet-tracker/cmd/api"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/db"
)

func main() {
	var dbConfig *mysql.Config

	// Check if we should use remote database
	if remoteUrl := config.Envs.RemoteDBUrl; remoteUrl != "" {
		dbConfig = parseDatabaseUrl(remoteUrl, true)
		log.Println("Using remote database connection")
	} else {
		dbConfig = parseDatabaseUrl(config.Envs.DatabaseUrl, false)
		log.Println("Using local database connection")
	}

	db, err := db.NewMySqlStorage(dbConfig)

	if err != nil {
		log.Fatal(err)
	}

	initStorage(db)

	server := api.NewAPIServer(":8080", db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

// Parse  database URL into MySQL config
func parseDatabaseUrl(dbUrl string, isRemote bool) *mysql.Config {
	log.Println("Parsing database URL:", dbUrl)
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
	config := &mysql.Config{
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

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected")
}
