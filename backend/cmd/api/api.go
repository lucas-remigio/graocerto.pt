package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/service/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := http.NewServeMux()

	apiV1Router := http.NewServeMux()

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(apiV1Router)

	// Register a handler for paths starting with /api/v1
	router.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1Router))

	log.Println("Server is running on", s.addr)
	return http.ListenAndServe(s.addr, router)
}
