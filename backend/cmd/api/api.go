package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/service/account"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/service/transaction"
	"github.com/lucas-remigio/wallet-tracker/service/transaction_types"
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

	accountStore := account.NewStore(s.db)
	accountHandler := account.NewHandler(accountStore)
	accountHandler.RegisterRoutes(apiV1Router)

	transactionTypesStore := transaction_types.NewStore(s.db)
	transactionTypesHandler := transaction_types.NewHandler(transactionTypesStore)
	transactionTypesHandler.RegisterRoutes(apiV1Router)

	categoryStore := category.NewStore(s.db)
	categoryHandler := category.NewHandler(categoryStore)
	categoryHandler.RegisterRoutes(apiV1Router)

	transactionStore := transaction.NewStore(s.db)
	transactionHandler := transaction.NewHandler(transactionStore)
	transactionHandler.RegisterRoutes(apiV1Router)

	// Register a handler for paths starting with /api/v1
	router.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1Router))

	log.Println("Server is running on", s.addr)
	return http.ListenAndServe(s.addr, corsMiddleware(router))
}

func corsMiddleware(next http.Handler) http.Handler {
	// Define allowed origins
	allowedOrigins := map[string]bool{
		"http://localhost":      true,
		config.Envs.FrontendUrl: true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false

		// Check if origin starts with any allowed prefix
		for allowedOrigin := range allowedOrigins {
			if strings.HasPrefix(origin, allowedOrigin) {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Common CORS headers
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
