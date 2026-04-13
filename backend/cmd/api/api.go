package api

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/cmd/api/middlewares"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/service/account"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/service/investment_calculator"
	"github.com/lucas-remigio/wallet-tracker/service/mailer"
	"github.com/lucas-remigio/wallet-tracker/service/notification"
	"github.com/lucas-remigio/wallet-tracker/service/openai"
	"github.com/lucas-remigio/wallet-tracker/service/recurring_rule"
	"github.com/lucas-remigio/wallet-tracker/service/transaction"
	"github.com/lucas-remigio/wallet-tracker/service/transaction_types"
	"github.com/lucas-remigio/wallet-tracker/service/user"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type APIServer struct {
	addr   string
	db     *sql.DB
	dbPing func(context.Context) error
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
	apiV1Router.HandleFunc("/healthz", s.handleHealthz)
	apiV1Router.HandleFunc("/readyz", s.handleReadyz)

	router.Handle("/metrics", promhttp.Handler())

	// Initialize all stores first
	userStore := user.NewStore(s.db)
	authTokenStore := userStore
	transactionTypesStore := transaction_types.NewStore(s.db)
	categoryStore := category.NewStore(s.db)
	recurringRuleStore := recurring_rule.NewStore(s.db)
	notificationStore := notification.NewStore(s.db)
	openAiStore := openai.NewClient()
	accountStore := account.NewStore(s.db, categoryStore, openAiStore)
	transactionStore := transaction.NewStore(s.db, accountStore)
	var authMailer mailer.Mailer = mailer.NoopMailer{}
	if config.Envs.SMTPHost != "" && config.Envs.SMTPFromEmail != "" {
		authMailer = mailer.NewSMTPMailer(
			config.Envs.SMTPHost,
			config.Envs.SMTPPort,
			config.Envs.SMTPUsername,
			config.Envs.SMTPPassword,
			config.Envs.SMTPFromEmail,
			config.Envs.SMTPFromName,
		)
	}

	// Now initialize handlers with the stores they need
	userHandler := user.NewHandler(userStore, authTokenStore, authMailer, accountStore, categoryStore, transactionStore)
	userHandler.RegisterRoutes(apiV1Router)

	transactionTypesHandler := transaction_types.NewHandler(transactionTypesStore)
	transactionTypesHandler.RegisterRoutes(apiV1Router)

	categoryHandler := category.NewHandler(categoryStore)
	categoryHandler.RegisterRoutes(apiV1Router)

	accountHandler := account.NewHandler(accountStore)
	accountHandler.RegisterRoutes(apiV1Router)

	transactionHandler := transaction.NewHandler(transactionStore)
	transactionHandler.RegisterRoutes(apiV1Router)

	recurringRuleHandler := recurring_rule.NewHandler(recurringRuleStore)
	recurringRuleHandler.RegisterRoutes(apiV1Router)

	notificationHandler := notification.NewHandler(notificationStore)
	notificationHandler.RegisterRoutes(apiV1Router)

	accountStore.SetTransactionStore(transactionStore)

	investmentCalculatorStore := investment_calculator.NewStore()
	investmentCalculatorHandler := investment_calculator.NewHandler(investmentCalculatorStore)
	investmentCalculatorHandler.RegisterRoutes(apiV1Router)

	go s.runRecurringRuleScheduler(recurringRuleStore, notificationStore)

	// Set up rate limiting middleware
	// Allow 2 requests per second, with a burst of 10 requests, and a
	rateLimiter := middlewares.NewClientRateLimiter(2, 10, 2*time.Minute)

	// Register a handler for paths starting with /api/v1
	apiHandlerChain := chainMiddleware(
		http.StripPrefix("/api/v1", apiV1Router),
		middlewares.RateLimitMiddleware(rateLimiter),
		middlewares.PrometheusMetricsMiddleware,
		middlewares.StructuredLoggerMiddleware,
		middlewares.RequestIDMiddleware,
	)
	router.Handle("/api/v1/", apiHandlerChain)

	slog.Info("Server is running", "addr", s.addr)
	return http.ListenAndServe(s.addr, corsMiddleware(router))
}

func (s *APIServer) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *APIServer) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := s.pingDatabase(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"degraded","db":"unreachable"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","db":"reachable"}`))
}

func (s *APIServer) pingDatabase(ctx context.Context) error {
	if s.dbPing != nil {
		return s.dbPing(ctx)
	}

	if s.db == nil {
		return errors.New("database connection is not configured")
	}

	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.db.PingContext(checkCtx)
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) runRecurringRuleScheduler(recurringRuleStore *recurring_rule.Store, notificationStore *notification.Store) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	run := func() {
		if err := recurringRuleStore.GeneratePendingTransactionsForDueRules(); err != nil {
			slog.Error("recurring rule scheduler error", "error", err)
		}
		if err := notificationStore.GenerateRecurringDueTomorrowNotifications(); err != nil {
			slog.Error("notification scheduler error", "error", err)
		}
	}

	run()
	for range ticker.C {
		run()
	}
}
