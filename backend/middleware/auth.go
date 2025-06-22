package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const UserIDKey ContextKey = "user_id"

// Common error messages
const (
	ErrUserNotAuthenticated = "user not authenticated"
	ErrMissingAuthHeader    = "missing authorization header"
	ErrInvalidPathParam     = "invalid path parameter"
)

// Standard response helpers
func WriteSuccessResponse(w http.ResponseWriter) {
	utils.WriteJson(w, http.StatusOK, map[string]string{"status": "success"})
}

func WriteCreatedResponse(w http.ResponseWriter) {
	utils.WriteJson(w, http.StatusCreated, map[string]string{"status": "created"})
}

func WriteDataResponse(w http.ResponseWriter, data interface{}) {
	utils.WriteJson(w, http.StatusOK, data)
}

// RequireAuth is a helper that checks authentication and returns user ID
// Returns (userID, true) if authenticated, (0, false) if not (and handles response)
func RequireAuth(w http.ResponseWriter, r *http.Request) (int, bool) {
	userId, ok := GetUserIDFromContext(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, errors.New(ErrUserNotAuthenticated))
		return 0, false
	}
	return userId, true
}

// AuthMiddleware extracts and validates the user ID from the Authorization header
// and adds it to the request context
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		if authToken == "" {
			utils.WriteError(w, http.StatusUnauthorized, errors.New(ErrMissingAuthHeader))
			return
		}

		userId, err := auth.GetUserIdFromToken(authToken)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, err)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userId)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userId, ok := r.Context().Value(UserIDKey).(int)
	return userId, ok
}

// PayloadValidator is a generic function that parses and validates request payloads
func ParseAndValidatePayload[T any](r *http.Request, payload *T) error {
	// Parse JSON payload
	if err := utils.ParseJson(r, payload); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("validation failed: %v", validationErrors)
		}
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// ValidatePayloadAndRespond is a helper that parses, validates, and handles errors automatically
func ValidatePayloadAndRespond[T any](w http.ResponseWriter, r *http.Request, payload *T) bool {
	if err := ParseAndValidatePayload(r, payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

// ExtractPathParam extracts a parameter from URL path segments
// For URL "/users/123/posts/456", GetPathParam(r, 1) returns "123", GetPathParam(r, 3) returns "456"
func ExtractPathParam(r *http.Request, index int) (string, error) {
	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if index < 0 || index >= len(segments) {
		return "", fmt.Errorf("path segment at index %d not found", index)
	}

	param := segments[index]
	if param == "" {
		return "", fmt.Errorf("empty path segment at index %d", index)
	}

	return param, nil
}

// ExtractPathParamAsInt extracts a path parameter and converts it to integer
func ExtractPathParamAsInt(r *http.Request, index int) (int, error) {
	param, err := ExtractPathParam(r, index)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf("invalid integer format in path segment: %w", err)
	}

	return id, nil
}

// ExtractPathParamAndRespond extracts a string path parameter and handles errors automatically
func ExtractPathParamAndRespond(w http.ResponseWriter, r *http.Request, index int) (string, bool) {
	param, err := ExtractPathParam(r, index)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf(ErrInvalidPathParam+": %w", err))
		return "", false
	}
	return param, true
}

// ExtractPathParamAsIntAndRespond extracts an integer path parameter and handles errors automatically
func ExtractPathParamAsIntAndRespond(w http.ResponseWriter, r *http.Request, index int) (int, bool) {
	id, err := ExtractPathParamAsInt(r, index)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf(ErrInvalidPathParam+": %w", err))
		return 0, false
	}
	return id, true
}

// MethodRouter handles multiple HTTP methods for a single endpoint
func MethodRouter(handlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, exists := handlers[r.Method]; exists {
			handler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
