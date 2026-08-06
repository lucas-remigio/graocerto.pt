package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/service/mailer"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type memoryUserStore struct {
	telegramUserStoreStub
	users      map[int]*types.User
	emailIndex map[string]int
	nextID     int
}

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{
		users:      map[int]*types.User{},
		emailIndex: map[string]int{},
		nextID:     1,
	}
}

func (s *memoryUserStore) GetUserByEmail(email string) (*types.User, error) {
	id, ok := s.emailIndex[email]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return s.users[id], nil
}

func (s *memoryUserStore) GetUserById(id int) (*types.User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *memoryUserStore) CreateUser(ctx context.Context, user *types.User) error {
	user.ID = s.nextID
	s.nextID++
	s.users[user.ID] = user
	s.emailIndex[user.Email] = user.ID
	return nil
}

func (s *memoryUserStore) ValidatePassword(password string) error {
	return validatePasswordForTest(password)
}

func (s *memoryUserStore) DeleteUser(userId int) error { return nil }

func (s *memoryUserStore) MarkEmailVerified(userID int, verified bool) error {
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.EmailVerified = verified
	return nil
}

func (s *memoryUserStore) UpdateMfaMethod(userID int, method types.MfaMethod) error {
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.MfaMethod = method
	return nil
}

func (s *memoryUserStore) UpdatePassword(userID int, hashedPassword string) error {
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.Password = hashedPassword
	return nil
}

type memoryAuthTokenStore struct {
	tokens map[string]*types.AuthToken
}

func newMemoryAuthTokenStore() *memoryAuthTokenStore {
	return &memoryAuthTokenStore{tokens: map[string]*types.AuthToken{}}
}

func (s *memoryAuthTokenStore) CreateAuthToken(token *types.AuthToken) error {
	copyToken := *token
	if copyToken.CreatedAt.IsZero() {
		copyToken.CreatedAt = time.Now()
	}
	s.tokens[token.ID] = &copyToken
	return nil
}

func (s *memoryAuthTokenStore) GetAuthTokenByID(id string) (*types.AuthToken, error) {
	token, ok := s.tokens[id]
	if !ok {
		return nil, fmt.Errorf("auth token not found")
	}
	return token, nil
}

func (s *memoryAuthTokenStore) GetAuthTokenByPurposeAndSecret(purpose types.AuthTokenPurpose, secret string) (*types.AuthToken, error) {
	secretHash := auth.HashSecret(secret)
	for _, token := range s.tokens {
		if token.Purpose == purpose && token.SecretHash == secretHash {
			return token, nil
		}
	}
	return nil, fmt.Errorf("auth token not found")
}

func (s *memoryAuthTokenStore) GetLatestAuthTokenByUserAndPurpose(userID int, purpose types.AuthTokenPurpose) (*types.AuthToken, error) {
	var latest *types.AuthToken
	for _, token := range s.tokens {
		if token.UserID != userID || token.Purpose != purpose {
			continue
		}

		if latest == nil || token.CreatedAt.After(latest.CreatedAt) {
			latest = token
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("auth token not found")
	}

	return latest, nil
}

func (s *memoryAuthTokenStore) ConsumeAuthToken(id string) error {
	token, ok := s.tokens[id]
	if !ok {
		return fmt.Errorf("auth token not found")
	}
	now := time.Now()
	token.ConsumedAt = &now
	return nil
}

func (s *memoryAuthTokenStore) IncrementAuthTokenAttempts(id string) error {
	token, ok := s.tokens[id]
	if !ok {
		return fmt.Errorf("auth token not found")
	}
	token.Attempts++
	return nil
}

func (s *memoryAuthTokenStore) DeleteAuthTokensByUserAndPurpose(userID int, purpose types.AuthTokenPurpose) error {
	for id, token := range s.tokens {
		if token.UserID == userID && token.Purpose == purpose && token.ConsumedAt == nil {
			delete(s.tokens, id)
		}
	}
	return nil
}

type memoryMailer struct {
	messages []mailer.Message
}

func (m *memoryMailer) Send(message mailer.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

func performJSONRequest(handler http.HandlerFunc, method, path string, body any) *httptest.ResponseRecorder {
	marshalled, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(marshalled))
	rr := httptest.NewRecorder()
	router := http.NewServeMux()
	router.HandleFunc(path, handler)
	router.ServeHTTP(rr, req)
	return rr
}

func TestHandleLoginCreatesOtpChallenge(t *testing.T) {
	userStore := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	mailer := &memoryMailer{}
	handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)

	rr := performJSONRequest(handler.handleLogin, http.MethodPost, "/login", types.LoginUserPayload{
		Email:    "john@example.com",
		Password: "Password1!",
	})

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	var response types.AuthChallengeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.ChallengeID == "" {
		t.Fatal("expected a challenge id")
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.messages))
	}
}

func TestHandleVerifyEmailMarksUserVerified(t *testing.T) {
	userStore := newMemoryUserStore()
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      "hashed",
		EmailVerified: false,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	token := "verification-token"
	_ = authStore.CreateAuthToken(&types.AuthToken{
		ID:          token,
		UserID:      1,
		Purpose:     types.AuthTokenPurposeEmailVerification,
		SecretHash:  auth.HashSecret(token),
		ExpiresAt:   time.Now().Add(time.Hour),
		Attempts:    0,
		MaxAttempts: 3,
	})

	handler := NewHandler(userStore, authStore, &memoryMailer{}, nil, nil, nil)
	rr := performJSONRequest(handler.handleVerifyEmail, http.MethodPost, "/auth/verify-email", types.VerifyEmailPayload{
		Token: token,
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	user, _ := userStore.GetUserById(1)
	if !user.EmailVerified {
		t.Fatal("expected email to be marked verified")
	}
}

func TestHandleResetPasswordUpdatesPassword(t *testing.T) {
	userStore := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	token := "reset-token"
	_ = authStore.CreateAuthToken(&types.AuthToken{
		ID:          token,
		UserID:      1,
		Purpose:     types.AuthTokenPurposePasswordReset,
		SecretHash:  auth.HashSecret(token),
		ExpiresAt:   time.Now().Add(time.Hour),
		Attempts:    0,
		MaxAttempts: 3,
	})

	handler := NewHandler(userStore, authStore, &memoryMailer{}, nil, nil, nil)
	rr := performJSONRequest(handler.handleResetPassword, http.MethodPost, "/auth/reset-password", types.ResetPasswordPayload{
		Token:    token,
		Password: "NewPassword1!",
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	user, _ := userStore.GetUserById(1)
	if !auth.CheckPasswordHash([]byte("NewPassword1!"), user.Password) {
		t.Fatal("expected password to be updated")
	}
}

func TestHandleRegisterSendsVerificationEmailWithFrontendRoute(t *testing.T) {
	userStore := newMemoryUserStore()
	authStore := newMemoryAuthTokenStore()
	mailer := &memoryMailer{}
	handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)

	prevFrontendURL := config.Envs.FrontendUrl
	config.Envs.FrontendUrl = "https://example.com/"
	t.Cleanup(func() {
		config.Envs.FrontendUrl = prevFrontendURL
	})

	rr := performJSONRequest(handler.handleRegister, http.MethodPost, "/register", types.RegisterUserPayload{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "Password1!",
	})

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.messages))
	}

	body := mailer.messages[0].Body
	if !strings.Contains(body, "https://example.com/verify-email?token=") {
		t.Fatalf("expected verification link to use /verify-email, got %q", body)
	}

	if strings.Contains(body, "/auth/verify-email?token=") {
		t.Fatalf("expected verification link not to use /auth/verify-email, got %q", body)
	}
}

func TestHandleRegisterReturnsGenericResponse(t *testing.T) {
	cases := []struct {
		name           string
		existingUser   *types.User
		expectedEmails int
	}{
		{
			name:           "new user",
			expectedEmails: 1,
		},
		{
			name: "verified existing user",
			existingUser: &types.User{
				FirstName:     "Jane",
				LastName:      "Doe",
				Email:         "jane@example.com",
				Password:      "hashed",
				EmailVerified: true,
				MfaMethod:     types.MfaMethodEmailOTP,
			},
		},
		{
			name: "unverified existing user",
			existingUser: &types.User{
				FirstName:     "John",
				LastName:      "Doe",
				Email:         "john@example.com",
				Password:      "hashed",
				EmailVerified: false,
				MfaMethod:     types.MfaMethodEmailOTP,
			},
			expectedEmails: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userStore := newMemoryUserStore()
			authStore := newMemoryAuthTokenStore()
			mailer := &memoryMailer{}
			handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)

			if tc.existingUser != nil {
				_ = userStore.CreateUser(context.Background(), tc.existingUser)
			}

			email := "john@example.com"
			if tc.existingUser != nil {
				email = tc.existingUser.Email
			}

			rr := performJSONRequest(handler.handleRegister, http.MethodPost, "/register", types.RegisterUserPayload{
				FirstName: "First",
				LastName:  "Last",
				Email:     email,
				Password:  "Password1!",
			})

			if rr.Code != http.StatusAccepted {
				t.Fatalf("expected 202, got %d", rr.Code)
			}

			var response types.MessageResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if response.Message != genericAuthResponse {
				t.Fatalf("unexpected response message: %q", response.Message)
			}

			if len(mailer.messages) != tc.expectedEmails {
				t.Fatalf("expected %d emails, got %d", tc.expectedEmails, len(mailer.messages))
			}
		})
	}
}

func TestSendVerificationEmailRespectsCooldown(t *testing.T) {
	userStore := newMemoryUserStore()
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      "hashed",
		EmailVerified: false,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	mailer := &memoryMailer{}
	handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)
	user, _ := userStore.GetUserById(1)

	if err := handler.sendVerificationEmail(user); err != nil {
		t.Fatalf("first send failed: %v", err)
	}
	if err := handler.sendVerificationEmail(user); err != nil {
		t.Fatalf("second send failed: %v", err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected one verification email, got %d", len(mailer.messages))
	}
}

func TestSendPasswordResetEmailRespectsCooldown(t *testing.T) {
	userStore := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	mailer := &memoryMailer{}
	handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)
	user, _ := userStore.GetUserById(1)

	if err := handler.sendPasswordResetEmail(user); err != nil {
		t.Fatalf("first send failed: %v", err)
	}
	if err := handler.sendPasswordResetEmail(user); err != nil {
		t.Fatalf("second send failed: %v", err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected one reset email, got %d", len(mailer.messages))
	}
}

func TestHandleForgotPasswordReturnsGenericResponse(t *testing.T) {
	userStore := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	mailer := &memoryMailer{}
	handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)

	knownRR := performJSONRequest(handler.handleForgotPassword, http.MethodPost, "/auth/forgot-password", types.ForgotPasswordPayload{
		Email: "john@example.com",
	})
	if knownRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for known email, got %d", knownRR.Code)
	}

	unknownRR := performJSONRequest(handler.handleForgotPassword, http.MethodPost, "/auth/forgot-password", types.ForgotPasswordPayload{
		Email: "unknown@example.com",
	})
	if unknownRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for unknown email, got %d", unknownRR.Code)
	}

	var knownResponse types.MessageResponse
	var unknownResponse types.MessageResponse
	if err := json.Unmarshal(knownRR.Body.Bytes(), &knownResponse); err != nil {
		t.Fatalf("failed to parse known response: %v", err)
	}
	if err := json.Unmarshal(unknownRR.Body.Bytes(), &unknownResponse); err != nil {
		t.Fatalf("failed to parse unknown response: %v", err)
	}

	if knownResponse.Message != genericAuthResponse || unknownResponse.Message != genericAuthResponse {
		t.Fatalf("expected generic responses, got %q and %q", knownResponse.Message, unknownResponse.Message)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected one reset email for known user, got %d", len(mailer.messages))
	}
}

func TestHandleResendVerificationReturnsGenericResponse(t *testing.T) {
	userStore := newMemoryUserStore()
	authStore := newMemoryAuthTokenStore()
	mailer := &memoryMailer{}
	handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)

	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      "hashed",
		EmailVerified: false,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	knownRR := performJSONRequest(handler.handleResendVerification, http.MethodPost, "/auth/resend-verification", types.ResendVerificationPayload{
		Email: "john@example.com",
	})
	if knownRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for known email, got %d", knownRR.Code)
	}

	unknownRR := performJSONRequest(handler.handleResendVerification, http.MethodPost, "/auth/resend-verification", types.ResendVerificationPayload{
		Email: "unknown@example.com",
	})
	if unknownRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for unknown email, got %d", unknownRR.Code)
	}

	var knownResponse types.MessageResponse
	var unknownResponse types.MessageResponse
	if err := json.Unmarshal(knownRR.Body.Bytes(), &knownResponse); err != nil {
		t.Fatalf("failed to parse known response: %v", err)
	}
	if err := json.Unmarshal(unknownRR.Body.Bytes(), &unknownResponse); err != nil {
		t.Fatalf("failed to parse unknown response: %v", err)
	}

	if knownResponse.Message != genericAuthResponse || unknownResponse.Message != genericAuthResponse {
		t.Fatalf("expected generic responses, got %q and %q", knownResponse.Message, unknownResponse.Message)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected one verification email for known user, got %d", len(mailer.messages))
	}
}

func TestHandleVerifyLoginOTPRejectsWrongAndExhaustedCodes(t *testing.T) {
	userStore := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	mailer := &memoryMailer{}
	handler := NewHandler(userStore, authStore, mailer, nil, nil, nil)
	user, _ := userStore.GetUserById(1)
	challenge, err := handler.createLoginOTPChallenge(user)
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}

	wrong := performJSONRequest(handler.handleVerifyLoginOTP, http.MethodPost, "/auth/verify-login-otp", types.VerifyLoginOTPPayload{
		ChallengeID: challenge.ChallengeID,
		Code:        "000000",
	})
	if wrong.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong code, got %d", wrong.Code)
	}

	token, err := authStore.GetAuthTokenByID(challenge.ChallengeID)
	if err != nil {
		t.Fatalf("failed to load challenge token: %v", err)
	}
	if token.Attempts != 1 {
		t.Fatalf("expected attempts to increment, got %d", token.Attempts)
	}

	_ = authStore.CreateAuthToken(&types.AuthToken{
		ID:          "exhausted-challenge",
		UserID:      1,
		Purpose:     types.AuthTokenPurposeLoginOTP,
		SecretHash:  auth.HashSecret("123456"),
		ExpiresAt:   time.Now().Add(time.Minute),
		Attempts:    loginOTPAttempts,
		MaxAttempts: loginOTPAttempts,
	})

	exhausted := performJSONRequest(handler.handleVerifyLoginOTP, http.MethodPost, "/auth/verify-login-otp", types.VerifyLoginOTPPayload{
		ChallengeID: "exhausted-challenge",
		Code:        "123456",
	})
	if exhausted.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for exhausted challenge, got %d", exhausted.Code)
	}
}

func TestHandleVerifyAndResetRejectInvalidExpiredAndConsumedTokens(t *testing.T) {
	userStore := newMemoryUserStore()
	hashedPassword, _ := auth.HashPassword("Password1!")
	_ = userStore.CreateUser(context.Background(), &types.User{
		FirstName:     "John",
		LastName:      "Doe",
		Email:         "john@example.com",
		Password:      hashedPassword,
		EmailVerified: true,
		MfaMethod:     types.MfaMethodEmailOTP,
	})

	authStore := newMemoryAuthTokenStore()
	handler := NewHandler(userStore, authStore, &memoryMailer{}, nil, nil, nil)

	verifyRR := performJSONRequest(handler.handleVerifyEmail, http.MethodPost, "/auth/verify-email", types.VerifyEmailPayload{
		Token: "missing-token",
	})
	if verifyRR.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid verification token, got %d", verifyRR.Code)
	}

	_ = authStore.CreateAuthToken(&types.AuthToken{
		ID:          "expired-token",
		UserID:      1,
		Purpose:     types.AuthTokenPurposeEmailVerification,
		SecretHash:  auth.HashSecret("expired-token"),
		ExpiresAt:   time.Now().Add(-time.Minute),
		Attempts:    0,
		MaxAttempts: 3,
	})

	expiredVerifyRR := performJSONRequest(handler.handleVerifyEmail, http.MethodPost, "/auth/verify-email", types.VerifyEmailPayload{
		Token: "expired-token",
	})
	if expiredVerifyRR.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for expired verification token, got %d", expiredVerifyRR.Code)
	}

	consumedToken := &types.AuthToken{
		ID:          "consumed-reset-token",
		UserID:      1,
		Purpose:     types.AuthTokenPurposePasswordReset,
		SecretHash:  auth.HashSecret("consumed-reset-token"),
		ExpiresAt:   time.Now().Add(time.Hour),
		Attempts:    0,
		MaxAttempts: 3,
	}
	_ = authStore.CreateAuthToken(consumedToken)
	_ = authStore.ConsumeAuthToken(consumedToken.ID)

	resetRR := performJSONRequest(handler.handleResetPassword, http.MethodPost, "/auth/reset-password", types.ResetPasswordPayload{
		Token:    consumedToken.ID,
		Password: "NewPassword1!",
	})
	if resetRR.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for consumed reset token, got %d", resetRR.Code)
	}
}
