package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/cmd/api/middlewares"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/service/account"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/service/investment_calculator"
	"github.com/lucas-remigio/wallet-tracker/service/openai"
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

	transactionTypesStore := transaction_types.NewStore(s.db)
	transactionTypesHandler := transaction_types.NewHandler(transactionTypesStore)
	transactionTypesHandler.RegisterRoutes(apiV1Router)

	categoryStore := category.NewStore(s.db)
	categoryHandler := category.NewHandler(categoryStore)
	categoryHandler.RegisterRoutes(apiV1Router)

	openAiStore := openai.NewClient()

	accountStore := account.NewStore(s.db, categoryStore, openAiStore)
	accountHandler := account.NewHandler(accountStore)
	accountHandler.RegisterRoutes(apiV1Router)

	transactionStore := transaction.NewStore(s.db, accountStore)
	transactionHandler := transaction.NewHandler(transactionStore)
	transactionHandler.RegisterRoutes(apiV1Router)

	accountStore.SetTransactionStore(transactionStore)

	investmentCalculatorStore := investment_calculator.NewStore()
	investmentCalculatorHandler := investment_calculator.NewHandler(investmentCalculatorStore)
	investmentCalculatorHandler.RegisterRoutes(apiV1Router)

	// Set up rate limiting middleware
	// Allow 2 requests per second, with a burst of 5 requests, and a
	rateLimiter := middlewares.NewClientRateLimiter(2, 5, 2*time.Minute)

	// Register a handler for paths starting with /api/v1
	apiHandlerChain := chainMiddleware(
		http.StripPrefix("/api/v1", apiV1Router),
		middlewares.RateLimitMiddleware(rateLimiter),
	)
	router.Handle("/api/v1/", apiHandlerChain)

	log.Println("Server is running on", s.addr)
	log.Printf("Starting HTTP server on %s", s.addr)
	return http.ListenAndServe(s.addr, corsMiddleware(router))
}

// Define a helper function to chain middlewares
func chainMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
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
