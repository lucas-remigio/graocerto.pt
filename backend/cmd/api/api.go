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
	return http.ListenAndServe(s.addr, corsMiddleware(router))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from your SvelteKit app
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

		// Allow credentials (cookies) to be sent with the request
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Allow specific HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
