package telegram

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

func newLinkServiceForTest(users *memoryUserStore) (*LinkService, *memoryTokenStore, *memoryPendingStore) {
	tokens := newMemoryTokenStore()
	pending := newMemoryPendingStore()
	return NewLinkService(users, tokens, pending), tokens, pending
}

func TestIssueCode(t *testing.T) {
	users := newMemoryUserStore(&types.User{ID: 1, FirstName: "Lucas"})
	links, tokens, _ := newLinkServiceForTest(users)

	code, expiresAt, err := links.IssueCode(1)
	if err != nil {
		t.Fatalf("IssueCode returned error: %v", err)
	}

	if len(code) != linkCodeLength {
		t.Fatalf("expected a %d character code, got %q", linkCodeLength, code)
	}

	if expiresAt.Before(time.Now()) {
		t.Fatalf("code expires in the past: %v", expiresAt)
	}

	if len(tokens.tokens) != 1 {
		t.Fatalf("expected exactly 1 stored token, got %d", len(tokens.tokens))
	}

	for _, token := range tokens.tokens {
		if token.SecretHash == code {
			t.Fatal("code was stored in plain text")
		}
		if token.SecretHash != auth.HashSecret(code) {
			t.Fatal("stored hash does not match the issued code")
		}
		if token.Purpose != types.AuthTokenPurposeTelegramLink {
			t.Fatalf("unexpected token purpose: %s", token.Purpose)
		}
	}
}

func TestIssueCodeSupersedesPreviousCode(t *testing.T) {
	users := newMemoryUserStore(&types.User{ID: 1})
	links, tokens, _ := newLinkServiceForTest(users)

	firstCode, _, err := links.IssueCode(1)
	if err != nil {
		t.Fatalf("first IssueCode returned error: %v", err)
	}

	secondCode, _, err := links.IssueCode(1)
	if err != nil {
		t.Fatalf("second IssueCode returned error: %v", err)
	}

	if len(tokens.tokens) != 1 {
		t.Fatalf("expected the old code to be dropped, got %d tokens", len(tokens.tokens))
	}

	if _, err := links.Redeem("chat-1", firstCode); !errors.Is(err, ErrInvalidLinkCode) {
		t.Fatalf("superseded code should be invalid, got %v", err)
	}

	if _, err := links.Redeem("chat-1", secondCode); err != nil {
		t.Fatalf("newest code should still redeem: %v", err)
	}
}

func TestRedeemLinksChatAndConsumesCode(t *testing.T) {
	user := &types.User{ID: 7, FirstName: "Lucas"}
	users := newMemoryUserStore(user)
	links, tokens, _ := newLinkServiceForTest(users)

	code, _, err := links.IssueCode(7)
	if err != nil {
		t.Fatalf("IssueCode returned error: %v", err)
	}

	linked, err := links.Redeem("chat-42", code)
	if err != nil {
		t.Fatalf("Redeem returned error: %v", err)
	}

	if linked.ID != 7 {
		t.Fatalf("expected user 7, got %d", linked.ID)
	}

	if user.TelegramChatID == nil || *user.TelegramChatID != "chat-42" {
		t.Fatalf("chat was not linked, got %v", user.TelegramChatID)
	}

	for _, token := range tokens.tokens {
		if token.ConsumedAt == nil {
			t.Fatal("code was not consumed")
		}
	}

	// Single use: the same code must not link a second chat.
	if _, err := links.Redeem("chat-99", code); !errors.Is(err, ErrInvalidLinkCode) {
		t.Fatalf("expected consumed code to be rejected, got %v", err)
	}
}

func TestRedeemNormalizesUserTypedCode(t *testing.T) {
	users := newMemoryUserStore(&types.User{ID: 1})
	links, _, _ := newLinkServiceForTest(users)

	code, _, err := links.IssueCode(1)
	if err != nil {
		t.Fatalf("IssueCode returned error: %v", err)
	}

	messy := "  " + strings.ToLower(code) + "\n"
	if _, err := links.Redeem("chat-1", messy); err != nil {
		t.Fatalf("expected lowercase padded code to redeem, got %v", err)
	}
}

func TestRedeemRejectsUnknownAndExpiredCodes(t *testing.T) {
	users := newMemoryUserStore(&types.User{ID: 1})
	links, tokens, _ := newLinkServiceForTest(users)

	if _, err := links.Redeem("chat-1", "NOTACODE"); !errors.Is(err, ErrInvalidLinkCode) {
		t.Fatalf("expected unknown code to be invalid, got %v", err)
	}

	code, _, err := links.IssueCode(1)
	if err != nil {
		t.Fatalf("IssueCode returned error: %v", err)
	}

	for _, token := range tokens.tokens {
		token.ExpiresAt = time.Now().Add(-time.Minute)
	}

	if _, err := links.Redeem("chat-1", code); !errors.Is(err, ErrInvalidLinkCode) {
		t.Fatalf("expected expired code to be invalid, got %v", err)
	}
}

// The unique index is what really stops a chat serving two accounts; when it
// fires, the code must stay usable so the user can retry after unlinking.
func TestRedeemKeepsCodeUsableWhenLinkFails(t *testing.T) {
	users := newMemoryUserStore(&types.User{ID: 1})
	links, tokens, _ := newLinkServiceForTest(users)

	code, _, err := links.IssueCode(1)
	if err != nil {
		t.Fatalf("IssueCode returned error: %v", err)
	}

	users.linkErr = fmt.Errorf("duplicate key value violates unique constraint")

	if _, err := links.Redeem("chat-1", code); err == nil {
		t.Fatal("expected an error when linking fails")
	}

	for _, token := range tokens.tokens {
		if token.ConsumedAt != nil {
			t.Fatal("code must not be consumed when linking failed")
		}
	}

	users.linkErr = nil
	if _, err := links.Redeem("chat-1", code); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
}

func TestUnlinkClearsChatAndPendingConversation(t *testing.T) {
	chatID := "chat-1"
	user := &types.User{ID: 3, TelegramChatID: &chatID}
	users := newMemoryUserStore(user)
	links, _, pending := newLinkServiceForTest(users)

	if err := pending.UpsertPendingParse(&types.PendingParse{UserID: 3, ChatID: chatID}); err != nil {
		t.Fatalf("failed to seed pending parse: %v", err)
	}

	if err := links.Unlink(3); err != nil {
		t.Fatalf("Unlink returned error: %v", err)
	}

	if user.TelegramChatID != nil {
		t.Fatalf("expected chat id to be cleared, got %v", *user.TelegramChatID)
	}

	parse, _ := pending.GetPendingParse(chatID)
	if parse != nil {
		t.Fatal("pending conversation outlived the link")
	}
}

func TestIsLinked(t *testing.T) {
	chatID := "chat-1"
	empty := ""
	users := newMemoryUserStore(
		&types.User{ID: 1, TelegramChatID: &chatID},
		&types.User{ID: 2},
		&types.User{ID: 3, TelegramChatID: &empty},
	)
	links, _, _ := newLinkServiceForTest(users)

	tests := []struct {
		userID int
		want   bool
	}{
		{userID: 1, want: true},
		{userID: 2, want: false},
		{userID: 3, want: false}, // empty string is not a link
	}

	for _, tc := range tests {
		got, err := links.IsLinked(tc.userID)
		if err != nil {
			t.Fatalf("IsLinked(%d) returned error: %v", tc.userID, err)
		}
		if got != tc.want {
			t.Fatalf("IsLinked(%d) = %v, want %v", tc.userID, got, tc.want)
		}
	}
}
