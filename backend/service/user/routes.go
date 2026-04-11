package user

import (
	"encoding/json"

	"fmt"
	"net/http"
	"time"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/service/mailer"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

// testing v
func NewHandlerForTesting(userStore types.UserStore) *Handler {
	return &Handler{
		store:            userStore,
		authTokenStore:   nil,
		mailer:           nil,
		accountStore:     nil, // Not needed for basic user tests
		categoryStore:    nil,
		transactionStore: nil,
	}
}

type Handler struct {
	store            types.UserStore
	authTokenStore   types.AuthTokenStore
	mailer           mailer.Mailer
	accountStore     types.AccountStore
	categoryStore    types.CategoryStore
	transactionStore types.TransactionStore
}

func NewHandler(store types.UserStore, authTokenStore types.AuthTokenStore, mailer mailer.Mailer, accountStore types.AccountStore, categoryStore types.CategoryStore, transactionStore types.TransactionStore) *Handler {
	return &Handler{
		store:            store,
		authTokenStore:   authTokenStore,
		mailer:           mailer,
		accountStore:     accountStore,
		categoryStore:    categoryStore,
		transactionStore: transactionStore,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/login", h.handleLogin)
	router.HandleFunc("/auth/google", h.handleGoogleLogin)
	router.HandleFunc("/register", h.handleRegister)
	router.HandleFunc("/auth/verify-email", h.handleVerifyEmail)
	router.HandleFunc("/auth/resend-verification", h.handleResendVerification)
	router.HandleFunc("/auth/verify-login-otp", h.handleVerifyLoginOTP)
	router.HandleFunc("/auth/forgot-password", h.handleForgotPassword)
	router.HandleFunc("/auth/reset-password", h.handleResetPassword)
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

	if h.authTokenStore == nil || h.mailer == nil || user.MfaMethod == types.MfaMethodOff {
		h.issueAuthToken(w, r, user)
		return
	}

	if !user.EmailVerified {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("email verification required"))
		return
	}

	if user.MfaMethod == types.MfaMethodTOTP {
		utils.WriteError(w, http.StatusNotImplemented, fmt.Errorf("totp mfa is not implemented yet"))
		return
	}

	challenge, err := h.createLoginOTPChallenge(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusAccepted, challenge)
}

func (h *Handler) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload types.GoogleLoginPayload
	if !middleware.ValidatePayloadAndRespond(w, r, &payload) {
		return
	}

	clientID := config.Envs.GoogleClientID

	claims, err := h.verifyGoogleTokenInfo(payload.Token, clientID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid google token: %v", err))
		return
	}

	user, err := h.getOrCreateGoogleUser(claims)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.issueAuthToken(w, r, user)
}

type googleClaims struct {
	Aud           string `json:"aud"`
	Email         string `json:"email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	EmailVerified string `json:"email_verified"`
	Error         string `json:"error"`
}

func (h *Handler) verifyGoogleTokenInfo(tokenString, expectedClientID string) (*googleClaims, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to contact google tokeninfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google rejected the token (status: %d)", resp.StatusCode)
	}

	var claims googleClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse google response: %w", err)
	}

	// Validate Audience (Client ID)
	if claims.Aud != expectedClientID {
		return nil, fmt.Errorf("unauthorized audience: expected %s, got %s", expectedClientID, claims.Aud)
	}

	// Validate Email
	if claims.EmailVerified != "true" && claims.EmailVerified != "1" {
		return nil, fmt.Errorf("google email is not verified")
	}

	return &claims, nil
}

func (h *Handler) getOrCreateGoogleUser(claims *googleClaims) (*types.User, error) {
	// 3. Find User in DB or Create if missing
	user, err := h.store.GetUserByEmail(claims.Email)
	if err == nil {
		if !user.EmailVerified {
			_ = h.store.MarkEmailVerified(user.ID, true)
		}
		if user.MfaMethod == "" {
			_ = h.store.UpdateMfaMethod(user.ID, types.MfaMethodEmailOTP)
		}
		return user, nil
	}

	// User not found, create a new one
	firstName := claims.GivenName
	lastName := claims.FamilyName

	randPass := utils.GenerateSecureRandomPassword()

	hashedPassword, err := auth.HashPassword(randPass)
	if err != nil {
		return nil, err
	}

	err = h.store.CreateUser(&types.User{
		FirstName:     firstName,
		LastName:      lastName,
		Email:         claims.Email,
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the created user to ensure ID is populated properly
	return h.store.GetUserByEmail(claims.Email)
}

func (h *Handler) issueAuthToken(w http.ResponseWriter, r *http.Request, user *types.User) {
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

	utils.WriteJson(w, http.StatusOK, map[string]string{"token": token, "created_at": user.CreatedAt})
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
	existingUser, err := h.store.GetUserByEmail(payload.Email)
	if err == nil {
		if !existingUser.EmailVerified && h.authTokenStore != nil && h.mailer != nil {
			if err := h.sendVerificationEmail(existingUser); err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}
		}

		utils.WriteJson(w, http.StatusAccepted, types.MessageResponse{Message: genericAuthResponse})
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
		FirstName:     payload.FirstName,
		LastName:      payload.LastName,
		Email:         payload.Email,
		Password:      hashedPassword,
		EmailVerified: false,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	if err != nil {
		fmt.Println("Error during user creation:", err) // Debugging
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if h.authTokenStore != nil && h.mailer != nil {
		createdUser, err := h.store.GetUserByEmail(payload.Email)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if err := h.sendVerificationEmail(createdUser); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	utils.WriteJson(w, http.StatusAccepted, types.MessageResponse{Message: genericAuthResponse})
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
