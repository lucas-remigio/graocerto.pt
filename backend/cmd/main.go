package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/lucas-remigio/wallet-tracker/cmd/api"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/db"
)

func main() {
	db, err := db.NewMYSQlStorage(&mysql.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 config.Envs.DBAddress,
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})

	if err != nil {
		log.Fatal(err)
	}

	dbHost := os.Getenv("DB_HOST")
	// print every info from config.envs
	fmt.Println("DB_HOST is:", dbHost)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", config.Envs.DBUser, config.Envs.DBPassword, config.Envs.DBAddress, config.Envs.DBName)
	// print the dsn

	fmt.Println("DSN is:", dsn)

	initStorage(db)

	server := api.NewAPIServer(":8080", db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected")
}
