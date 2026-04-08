package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealthz(t *testing.T) {
	server := &APIServer{}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	server.handleHealthz(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHandleReadyz(t *testing.T) {
	tests := []struct {
		name       string
		dbPing     func(context.Context) error
		wantStatus int
	}{
		{
			name: "returns ok when db is reachable",
			dbPing: func(context.Context) error {
				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns service unavailable when db is down",
			dbPing: func(context.Context) error {
				return errors.New("db unavailable")
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := &APIServer{dbPing: tc.dbPing}
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rr := httptest.NewRecorder()

			server.handleReadyz(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}
