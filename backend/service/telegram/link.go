package telegram

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

const (
	// LinkCodeTTL is deliberately short: the code is typed into a third-party
	// chat app, so its useful life should be minutes, not days.
	LinkCodeTTL = 15 * time.Minute
	// linkCodeLength balances typeability against brute force. See
	// auth.GenerateReadableCode.
	linkCodeLength = 8
	// linkCodeMaxAttempts caps redemption attempts against a known code id.
	linkCodeMaxAttempts = 5
)

// ErrInvalidLinkCode covers unknown, expired and already-used codes. They are
// deliberately indistinguishable to whoever is typing codes into the bot.
var ErrInvalidLinkCode = errors.New("invalid or expired link code")

// LinkService owns the lifecycle of the chat <-> user link. It has no HTTP or
// Telegram knowledge, so both the JWT handler and the bot can drive it.
type LinkService struct {
	users   UserLinkStore
	tokens  LinkTokenStore
	pending types.TelegramStore
}

func NewLinkService(users UserLinkStore, tokens LinkTokenStore, pending types.TelegramStore) *LinkService {
	return &LinkService{
		users:   users,
		tokens:  tokens,
		pending: pending,
	}
}

// IssueCode creates a fresh link code for the user, superseding any code that
// is still outstanding so only the newest one can ever be redeemed.
func (s *LinkService) IssueCode(userID int) (string, time.Time, error) {
	if err := s.tokens.DeleteAuthTokensByUserAndPurpose(userID, types.AuthTokenPurposeTelegramLink); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to clear previous link codes: %w", err)
	}

	code, err := auth.GenerateReadableCode(linkCodeLength)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate link code: %w", err)
	}

	// The row id is independent of the code: the user types the code, and only
	// its hash is stored.
	id, err := auth.GenerateOneTimeToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate link token id: %w", err)
	}

	expiresAt := time.Now().Add(LinkCodeTTL)
	token := &types.AuthToken{
		ID:          id,
		UserID:      userID,
		Purpose:     types.AuthTokenPurposeTelegramLink,
		SecretHash:  auth.HashSecret(code),
		ExpiresAt:   expiresAt,
		MaxAttempts: linkCodeMaxAttempts,
	}

	if err := s.tokens.CreateAuthToken(token); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to store link code: %w", err)
	}

	return code, expiresAt, nil
}

// Redeem links chatID to the user that owns the code and returns that user.
// The unique index on users.telegram_chat_id is what actually guarantees a chat
// cannot be linked to two accounts; a link failure here surfaces as an error.
func (s *LinkService) Redeem(chatID, code string) (*types.User, error) {
	token, err := s.tokens.GetAuthTokenByPurposeAndSecret(types.AuthTokenPurposeTelegramLink, normalizeCode(code))
	if err != nil || token == nil {
		return nil, ErrInvalidLinkCode
	}

	if token.ConsumedAt != nil || time.Now().After(token.ExpiresAt) {
		return nil, ErrInvalidLinkCode
	}

	// Link before consuming: if the link fails the code stays usable, so the
	// user can simply try again instead of asking for a new one.
	if err := s.users.LinkTelegramChatID(token.UserID, chatID); err != nil {
		return nil, fmt.Errorf("failed to link telegram chat: %w", err)
	}

	if err := s.tokens.ConsumeAuthToken(token.ID); err != nil {
		return nil, fmt.Errorf("failed to consume link code: %w", err)
	}

	user, err := s.users.GetUserById(token.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load linked user: %w", err)
	}

	slog.Info("telegram chat linked to user", "userID", user.ID)
	return user, nil
}

// Unlink drops the link and any conversation it authorised.
func (s *LinkService) Unlink(userID int) error {
	if err := s.users.UnlinkTelegram(userID); err != nil {
		return err
	}

	// A pending conversation must not outlive the link that authorised it.
	return s.pending.DeletePendingParsesByUserID(userID)
}

// IsLinked reports whether the user currently has a chat linked.
func (s *LinkService) IsLinked(userID int) (bool, error) {
	user, err := s.users.GetUserById(userID)
	if err != nil {
		return false, err
	}

	return user.TelegramChatID != nil && *user.TelegramChatID != "", nil
}

// normalizeCode forgives the formatting a chat app or a human adds around a code.
func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
