package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

func performRequest(handler http.HandlerFunc, method, path string, body interface{}) *httptest.ResponseRecorder {
	marshalled, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(marshalled))
	rr := httptest.NewRecorder()
	router := http.NewServeMux()
	router.HandleFunc(path, handler)
	router.ServeHTTP(rr, req)
	return rr
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name       string
		store      types.UserStore
		payload    interface{}
		wantStatus int
	}{
		{
			name:  "invalid email",
			store: &mockUserStore{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "invalid.pt", Password: "1234341233",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing fields",
			store:      &mockUserStore{},
			payload:    map[string]string{"FirstName": "John", "Email": "john.doe@example.com"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "user already exists",
			store: &mockUserStoreDuplicate{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "existing@mail.pt", Password: "password123",
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:  "internal server error",
			store: &mockUserStoreError{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "john.doe@example.com", Password: "Password1!",
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:  "success",
			store: &mockUserStore{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "lucas@mail.pt", Password: "Password1!",
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:  "password too short",
			store: &mockUserStore{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "john.short@mail.pt", Password: "Ab1!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "password missing uppercase",
			store: &mockUserStore{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "john.noupper@mail.pt", Password: "password1!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "password missing lowercase",
			store: &mockUserStore{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "john.nolower@mail.pt", Password: "PASSWORD1!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "password missing digit",
			store: &mockUserStore{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "john.nodigit@mail.pt", Password: "Password!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "password missing special",
			store: &mockUserStore{},
			payload: types.RegisterUserPayload{
				FirstName: "John", LastName: "Doe", Email: "john.nospecial@mail.pt", Password: "Password1",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandlerForTesting(tc.store) // Changed this line
			rr := performRequest(handler.handleRegister, http.MethodPost, "/register", tc.payload)
			if rr.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		store      types.UserStore
		payload    interface{}
		wantStatus int
		checkToken bool
	}{
		{
			name:       "fail if password is incorrect",
			store:      &mockUserStoreLogin{},
			payload:    types.LoginUserPayload{Email: "test@mail.pt", Password: "wrong_password"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "fail if email does not exist",
			store:      &mockUserStore{},
			payload:    types.LoginUserPayload{Email: "nonexistent@mail.pt", Password: "password123"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "success",
			store:      &mockUserStoreLogin{},
			payload:    types.LoginUserPayload{Email: "test@mail.pt", Password: "correct_password"},
			wantStatus: http.StatusOK,
			checkToken: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandlerForTesting(tc.store)
			rr := performRequest(handler.handleLogin, http.MethodPost, "/login", tc.payload)
			if rr.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}
			if tc.checkToken && rr.Code == http.StatusOK {
				var responseBody map[string]string
				_ = json.Unmarshal(rr.Body.Bytes(), &responseBody)
				if _, ok := responseBody["token"]; !ok {
					t.Error("expected a token in the response, got none")
				}
			}
		})
	}
}

type mockUserStore struct{}

func (m *mockUserStore) GetUserByEmail(email string) (*types.User, error) {
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserStore) GetUserById(id int) (*types.User, error) {
	return &types.User{}, nil
}

func (m *mockUserStore) CreateUser(ctx context.Context, user *types.User) error {
	return nil
}

func (m *mockUserStore) ValidatePassword(password string) error {
	return validatePasswordForTest(password)
}

func (m *mockUserStore) DeleteUser(userId int) error {
	return nil
}

func (m *mockUserStore) MarkEmailVerified(userID int, verified bool) error { return nil }

func (m *mockUserStore) UpdateMfaMethod(userID int, method types.MfaMethod) error {
	return nil
}

func (m *mockUserStore) UpdatePassword(userID int, hashedPassword string) error {
	return nil
}

type mockUserStoreDuplicate struct{}

func (m *mockUserStoreDuplicate) GetUserByEmail(email string) (*types.User, error) {
	return &types.User{Email: email}, nil // Simulate existing user
}

func (m *mockUserStoreDuplicate) CreateUser(ctx context.Context, user *types.User) error {
	return nil
}

func (m *mockUserStoreDuplicate) GetUserById(id int) (*types.User, error) {
	return &types.User{}, nil
}

func (m *mockUserStoreDuplicate) ValidatePassword(password string) error {
	return validatePasswordForTest(password)
}

func (m *mockUserStoreDuplicate) DeleteUser(userId int) error {
	return nil
}

func (m *mockUserStoreDuplicate) MarkEmailVerified(userID int, verified bool) error {
	return nil
}

func (m *mockUserStoreDuplicate) UpdateMfaMethod(userID int, method types.MfaMethod) error {
	return nil
}

func (m *mockUserStoreDuplicate) UpdatePassword(userID int, hashedPassword string) error {
	return nil
}

type mockUserStoreError struct{}

func (m *mockUserStoreError) GetUserByEmail(email string) (*types.User, error) {
	return nil, fmt.Errorf("user already exists")
}

func (m *mockUserStoreError) CreateUser(ctx context.Context, user *types.User) error {
	return fmt.Errorf("internal server error")
}

func (m *mockUserStoreError) GetUserById(id int) (*types.User, error) {
	return nil, fmt.Errorf("internal server error")
}

func (m *mockUserStoreError) ValidatePassword(password string) error {
	return validatePasswordForTest(password)
}

func (m *mockUserStoreError) DeleteUser(userId int) error {
	return fmt.Errorf("internal server error")
}

func (m *mockUserStoreError) MarkEmailVerified(userID int, verified bool) error {
	return nil
}

func (m *mockUserStoreError) UpdateMfaMethod(userID int, method types.MfaMethod) error {
	return nil
}

func (m *mockUserStoreError) UpdatePassword(userID int, hashedPassword string) error {
	return nil
}

type mockUserStoreLogin struct{}

func (m *mockUserStoreLogin) GetUserByEmail(email string) (*types.User, error) {
	// Use the same auth.HashPassword function as production for consistency
	hashed, _ := auth.HashPassword("correct_password")
	return &types.User{
		Email:    email,
		Password: hashed,
	}, nil
}

func (m *mockUserStoreLogin) CreateUser(ctx context.Context, user *types.User) error            { return nil }
func (m *mockUserStoreLogin) GetUserById(id int) (*types.User, error) { return &types.User{}, nil }
func (m *mockUserStoreLogin) ValidatePassword(password string) error {
	return validatePasswordForTest(password)
}

func (m *mockUserStoreLogin) DeleteUser(userId int) error {
	return nil
}

func (m *mockUserStoreLogin) MarkEmailVerified(userID int, verified bool) error { return nil }

func (m *mockUserStoreLogin) UpdateMfaMethod(userID int, method types.MfaMethod) error {
	return nil
}

func (m *mockUserStoreLogin) UpdatePassword(userID int, hashedPassword string) error {
	return nil
}

type mockUserStoreSuccess struct{}

func (m *mockUserStoreSuccess) GetUserByEmail(email string) (*types.User, error) {
	hashedPassword, _ := auth.HashPassword("123123") // Simulate matching hash
	return &types.User{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
	}, nil
}
func (m *mockUserStoreSuccess) CreateUser(ctx context.Context, user *types.User) error {
	return nil
}

func (m *mockUserStoreSuccess) GetUserById(id int) (*types.User, error) {
	return &types.User{}, nil
}

func (m *mockUserStoreSuccess) ValidatePassword(password string) error {
	return validatePasswordForTest(password)
}

func validatePasswordForTest(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

func (m *mockUserStoreSuccess) DeleteUser(userId int) error {
	return nil
}

func (m *mockUserStoreSuccess) MarkEmailVerified(userID int, verified bool) error { return nil }

func (m *mockUserStoreSuccess) UpdateMfaMethod(userID int, method types.MfaMethod) error {
	return nil
}

func (m *mockUserStoreSuccess) UpdatePassword(userID int, hashedPassword string) error {
	return nil
}
