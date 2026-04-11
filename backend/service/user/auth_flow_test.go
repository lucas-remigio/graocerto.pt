package user

import (
	"bytes"
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

func (s *memoryUserStore) CreateUser(user *types.User) error {
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
	_ = userStore.CreateUser(&types.User{
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
	_ = userStore.CreateUser(&types.User{
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
	_ = userStore.CreateUser(&types.User{
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

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
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
