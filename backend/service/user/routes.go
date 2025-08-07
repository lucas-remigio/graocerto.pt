package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

// testing v
func NewHandlerForTesting(userStore types.UserStore) *Handler {
	return &Handler{
		store:            userStore,
		accountStore:     nil, // Not needed for basic user tests
		categoryStore:    nil,
		transactionStore: nil,
	}
}

type Handler struct {
	store            types.UserStore
	accountStore     types.AccountStore
	categoryStore    types.CategoryStore
	transactionStore types.TransactionStore
}

func NewHandler(store types.UserStore, accountStore types.AccountStore, categoryStore types.CategoryStore, transactionStore types.TransactionStore) *Handler {
	return &Handler{
		store:            store,
		accountStore:     accountStore,
		categoryStore:    categoryStore,
		transactionStore: transactionStore,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/login", h.handleLogin)
	router.HandleFunc("/register", h.handleRegister)
	router.HandleFunc("/verify-token", middleware.AuthMiddleware(h.verifyToken))
	router.HandleFunc("/auth/delete-account", middleware.AuthMiddleware(h.handleDeleteAccount))
	router.HandleFunc("/auth/export-data", middleware.AuthMiddleware(h.handleExportData))
}

func (h *Handler) verifyToken(w http.ResponseWriter, r *http.Request) {
	// If we reach here, the middleware has already verified the token
	// and the user is authenticated
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse and validate JSON payload
	var payload types.LoginUserPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// get the user from the store
	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("not found, invalid email or password"))
		return
	}

	// check if the password is correct
	if !auth.CheckPasswordHash([]byte(payload.Password), user.Password) {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("not found, invalid email or password"))
		return
	}

	secret := []byte(config.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	isSecure := r.TLS != nil
	jwtExpiration := config.Envs.JWTExpirationInSeconds

	// Set the authToken as a secure, HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,                    // Prevents client-side JavaScript from accessing the cookie
		Secure:   isSecure,                // Only send the cookie over HTTPS
		SameSite: http.SameSiteStrictMode, // Prevents CSRF attacks
		MaxAge:   int(jwtExpiration),      // Token expires at the same time as the JWT
	})

	utils.WriteJson(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse and validate JSON payload
	var payload types.RegisterUserPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	// check if the user exists
	_, err := h.store.GetUserByEmail(payload.Email)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	err = h.store.ValidatePassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// create a new user
	err = h.store.CreateUser(&types.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
	})

	if err != nil {
		fmt.Println("Error during user creation:", err) // Debugging
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.WriteCreatedResponse(w)
}

func (h *Handler) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the user ID from the middleware (set by AuthMiddleware)
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// Delete all user data (this should cascade delete related data)
	err := h.store.DeleteUser(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete account: %v", err))
		return
	}

	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete the cookie
	})

	utils.WriteJson(w, http.StatusOK, map[string]string{"message": "Account deleted successfully"})
}

func (h *Handler) handleExportData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the user ID from the middleware
	userId, ok := middleware.RequireAuth(w, r)
	if !ok {
		return
	}

	// Get all user data using the handler method (not store method)
	userData, err := h.getUserDataExport(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to export data: %v", err))
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=grao-certo-data-%s.json", time.Now().Format("2006-01-02")))

	// Write JSON response
	utils.WriteJson(w, http.StatusOK, userData)
}

func (h *Handler) getUserDataExport(userID int) (*types.ExportData, error) {
	result := &types.ExportData{ExportedAt: time.Now()}

	// Get user using user store
	user, err := h.store.GetUserById(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	result.User = user

	// Get accounts using account store
	accounts, err := h.accountStore.GetAccountsByUserId(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}
	result.Accounts = accounts

	// Get categories using category store
	categories, err := h.categoryStore.GetCategoriesDtoByUserId(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %v", err)
	}
	result.Categories = categories

	// Get transactions using transaction store
	var allTransactions []*types.TransactionDTO
	for _, account := range accounts {
		transactions, err := h.transactionStore.GetTransactionsDTOByAccountToken(account.Token, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get transactions for account %s: %v", account.Token, err)
		}
		allTransactions = append(allTransactions, transactions...)
	}
	result.Transactions = allTransactions

	return result, nil
}
