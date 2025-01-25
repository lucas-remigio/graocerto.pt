package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

func TestUserServiceHandlers(t *testing.T) {
	userStore := &mockUserStore{}
	handler := NewHandler(userStore)

	t.Run("should fail if the user payload is invalid", func(t *testing.T) {
		payload := types.RegisterUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "invalid.pt",
			Password:  "1234341233",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))

		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()

		// handle /register func
		router.HandleFunc("/register", handler.handleRegister)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("Should correctly register the user", func(t *testing.T) {
		payload := types.RegisterUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "lucas@mail.pt",
			Password:  "1234341233",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))

		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()

		// handle /register func
		router.HandleFunc("/register", handler.handleRegister)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("should fail if required fields are missing", func(t *testing.T) {
		payload := map[string]string{
			"FirstName": "John",
			"Email":     "john.doe@example.com",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()
		router.HandleFunc("/register", handler.handleRegister)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("should fail if user already exists", func(t *testing.T) {
		userStore := &mockUserStoreDuplicate{}
		handler := NewHandler(userStore)

		payload := types.RegisterUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "existing@mail.pt",
			Password:  "password123",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()
		router.HandleFunc("/register", handler.handleRegister)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("should return 500 if user creation fails", func(t *testing.T) {
		userStore := &mockUserStoreError{}
		handler := NewHandler(userStore)

		payload := types.RegisterUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()
		router.HandleFunc("/register", handler.handleRegister)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
	})

	t.Run("should fail if password is incorrect", func(t *testing.T) {

		userStore := &mockUserStoreLogin{}
		handler := NewHandler(userStore)

		payload := types.LoginUserPayload{
			Email:    "test@mail.pt",
			Password: "wrong_password",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()
		router.HandleFunc("/login", handler.handleLogin)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("should fail if email does not exist", func(t *testing.T) {
		userStore := &mockUserStore{}
		handler := NewHandler(userStore)

		payload := types.LoginUserPayload{
			Email:    "nonexistent@mail.pt",
			Password: "password123",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()
		router.HandleFunc("/login", handler.handleLogin)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("should return token on successful login", func(t *testing.T) {

		userStore := &mockUserStoreSuccess{}
		handler := NewHandler(userStore)

		payload := types.LoginUserPayload{
			Email:    "lucas@mail.pt",
			Password: "123123",
		}

		marshalled, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := http.NewServeMux()
		router.HandleFunc("/login", handler.handleLogin)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var responseBody map[string]string
		_ = json.Unmarshal(rr.Body.Bytes(), &responseBody)

		if _, ok := responseBody["token"]; !ok {
			t.Error("expected a token in the response, got none")
		}
	})
}

type mockUserStore struct{}

func (m *mockUserStore) GetUserByEmail(email string) (*types.User, error) {
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserStore) GetUserById(id int) (*types.User, error) {
	return &types.User{}, nil
}

func (m *mockUserStore) CreateUser(user *types.User) error {
	return nil
}

type mockUserStoreDuplicate struct{}

func (m *mockUserStoreDuplicate) GetUserByEmail(email string) (*types.User, error) {
	return &types.User{Email: email}, nil // Simulate existing user
}

func (m *mockUserStoreDuplicate) CreateUser(user *types.User) error {
	return nil
}

func (m *mockUserStoreDuplicate) GetUserById(id int) (*types.User, error) {
	return &types.User{}, nil
}

type mockUserStoreError struct{}

func (m *mockUserStoreError) GetUserByEmail(email string) (*types.User, error) {
	return nil, fmt.Errorf("user already exists")
}

func (m *mockUserStoreError) CreateUser(user *types.User) error {
	return fmt.Errorf("internal server error")
}

func (m *mockUserStoreError) GetUserById(id int) (*types.User, error) {
	return nil, fmt.Errorf("internal server error")
}

type mockUserStoreLogin struct{}

func (m *mockUserStoreLogin) GetUserByEmail(email string) (*types.User, error) {
	return &types.User{
		Email:    email,
		Password: "hashed_password", // Simulate stored password
	}, nil
}

func (m *mockUserStoreLogin) CreateUser(*types.User) error {
	return nil
}

func (m *mockUserStoreLogin) GetUserById(id int) (*types.User, error) {
	return &types.User{}, nil
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
func (m *mockUserStoreSuccess) CreateUser(*types.User) error {
	return nil
}

func (m *mockUserStoreSuccess) GetUserById(id int) (*types.User, error) {
	return &types.User{}, nil
}
