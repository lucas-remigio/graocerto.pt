package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucas-remigio/wallet-tracker/middleware"
	"github.com/lucas-remigio/wallet-tracker/types"
)

// requestAs builds a request that already carries what AuthMiddleware would
// have put in the context, so handlers can be exercised on their own.
func requestAs(userID int, method string) *http.Request {
	req := httptest.NewRequest(method, "/telegram", nil)
	if userID == 0 {
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
}

func newHandlerForTest(users *memoryUserStore) (*Handler, *memoryPendingStore) {
	links, _, pending := newLinkServiceForTest(users)
	return NewHandler(links), pending
}

func TestCreateLinkTokenRequiresAuth(t *testing.T) {
	handler, _ := newHandlerForTest(newMemoryUserStore(&types.User{ID: 1}))
	rr := httptest.NewRecorder()

	handler.CreateLinkToken(rr, requestAs(0, http.MethodPost))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestCreateLinkTokenReturnsCode(t *testing.T) {
	handler, _ := newHandlerForTest(newMemoryUserStore(&types.User{ID: 1}))
	rr := httptest.NewRecorder()

	handler.CreateLinkToken(rr, requestAs(1, http.MethodPost))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body types.TelegramLinkTokenResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Token) != linkCodeLength {
		t.Fatalf("expected a %d character token, got %q", linkCodeLength, body.Token)
	}

	if body.ExpiresAt.IsZero() {
		t.Fatal("expected an expiry in the response")
	}
}

func TestGetStatusReflectsLink(t *testing.T) {
	chatID := "chat-1"
	users := newMemoryUserStore(
		&types.User{ID: 1, TelegramChatID: &chatID},
		&types.User{ID: 2},
	)
	handler, _ := newHandlerForTest(users)

	tests := []struct {
		userID int
		want   bool
	}{
		{userID: 1, want: true},
		{userID: 2, want: false},
	}

	for _, tc := range tests {
		rr := httptest.NewRecorder()
		handler.GetStatus(rr, requestAs(tc.userID, http.MethodGet))

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}

		var body types.TelegramStatusResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if body.Linked != tc.want {
			t.Fatalf("user %d: linked = %v, want %v", tc.userID, body.Linked, tc.want)
		}
	}
}

func TestUnlinkClearsLink(t *testing.T) {
	chatID := "chat-1"
	user := &types.User{ID: 1, TelegramChatID: &chatID}
	handler, pending := newHandlerForTest(newMemoryUserStore(user))

	if err := pending.UpsertPendingParse(&types.PendingParse{UserID: 1, ChatID: chatID}); err != nil {
		t.Fatalf("failed to seed pending parse: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.Unlink(rr, requestAs(1, http.MethodDelete))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if user.TelegramChatID != nil {
		t.Fatal("expected the chat link to be cleared")
	}

	parse, _ := pending.GetPendingParse(chatID)
	if parse != nil {
		t.Fatal("expected the pending conversation to be cleared")
	}
}

func TestUnlinkRequiresAuth(t *testing.T) {
	handler, _ := newHandlerForTest(newMemoryUserStore(&types.User{ID: 1}))
	rr := httptest.NewRecorder()

	handler.Unlink(rr, requestAs(0, http.MethodDelete))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
