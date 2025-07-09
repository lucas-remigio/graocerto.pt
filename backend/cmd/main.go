package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/lucas-remigio/wallet-tracker/cmd/api"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/db"
)

func main() {
	var dbConfig *mysql.Config

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

	initStorage(db)

	port := config.Envs.Port
	if port == "" {
		port = "8080"
		log.Println("No port specified, using default port 8080")
	}
	addr := ":" + port

	server := api.NewAPIServer(addr, db)

	// Use the server's Run method which will handle HTTP/HTTPS internally
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
		// Load your CA certificate
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile("db/DigiCertGlobalRootCA.crt.pem")
		if err != nil {
			log.Fatalf("Failed to read CA cert: %v", err)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			log.Fatal("Failed to append CA cert")
		}
		tlsConfig := &tls.Config{
			RootCAs: rootCertPool,
		}
		mysql.RegisterTLSConfig("custom", tlsConfig)
		config.TLSConfig = "custom"
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
