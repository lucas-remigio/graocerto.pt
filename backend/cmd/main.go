package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/lucas-remigio/wallet-tracker/cmd/api"
	"github.com/lucas-remigio/wallet-tracker/config"
)

func main() {
	var dbURL string

	if config.Envs.IsProduction {
		dbURL = config.Envs.RemoteDBUrl
		log.Println("Using remote database connection")
	} else {
		dbURL = config.Envs.DatabaseUrl
		log.Println("Using local database connection")
	}

	// Open the Postgres database connection
	pgdb, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()

	initStorage(pgdb)

	port := config.Envs.Port
	if port == "" {
		port = "8080"
		log.Println("No port specified, using default port 8080")
	}
	addr := ":" + port

	server := api.NewAPIServer(addr, pgdb)

	// Use the server's Run method which will handle HTTP/HTTPS internally
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB) {
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("DB: Successfully connected")
}
