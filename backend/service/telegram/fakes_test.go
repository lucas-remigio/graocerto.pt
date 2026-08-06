package telegram

import (
	"fmt"
	"time"

	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

// In-memory doubles shared by every test in this package. They implement the
// narrow interfaces from deps.go, which is why they stay this small.

type memoryUserStore struct {
	users     map[int]*types.User
	linkErr   error // forced failure, e.g. the unique index rejecting a chat
	unlinkErr error
}

func newMemoryUserStore(users ...*types.User) *memoryUserStore {
	store := &memoryUserStore{users: map[int]*types.User{}}
	for _, u := range users {
		store.users[u.ID] = u
	}
	return store
}

func (s *memoryUserStore) GetUserById(id int) (*types.User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *memoryUserStore) GetUserByTelegramChatID(chatID string) (*types.User, error) {
	for _, user := range s.users {
		if user.TelegramChatID != nil && *user.TelegramChatID == chatID {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (s *memoryUserStore) LinkTelegramChatID(userID int, chatID string) error {
	if s.linkErr != nil {
		return s.linkErr
	}
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.TelegramChatID = &chatID
	return nil
}

func (s *memoryUserStore) UnlinkTelegram(userID int) error {
	if s.unlinkErr != nil {
		return s.unlinkErr
	}
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.TelegramChatID = nil
	return nil
}

type memoryTokenStore struct {
	tokens map[string]*types.AuthToken
}

func newMemoryTokenStore() *memoryTokenStore {
	return &memoryTokenStore{tokens: map[string]*types.AuthToken{}}
}

func (s *memoryTokenStore) CreateAuthToken(token *types.AuthToken) error {
	copied := *token
	s.tokens[token.ID] = &copied
	return nil
}

func (s *memoryTokenStore) GetAuthTokenByPurposeAndSecret(purpose types.AuthTokenPurpose, secret string) (*types.AuthToken, error) {
	hash := auth.HashSecret(secret)
	for _, token := range s.tokens {
		if token.Purpose == purpose && token.SecretHash == hash {
			return token, nil
		}
	}
	return nil, fmt.Errorf("auth token not found")
}

func (s *memoryTokenStore) ConsumeAuthToken(id string) error {
	token, ok := s.tokens[id]
	if !ok {
		return fmt.Errorf("auth token not found")
	}
	now := time.Now()
	token.ConsumedAt = &now
	return nil
}

func (s *memoryTokenStore) DeleteAuthTokensByUserAndPurpose(userID int, purpose types.AuthTokenPurpose) error {
	for id, token := range s.tokens {
		if token.UserID == userID && token.Purpose == purpose && token.ConsumedAt == nil {
			delete(s.tokens, id)
		}
	}
	return nil
}

type stubCategoryReader struct {
	categories []*types.CategoryDTO
	err        error
}

func (s *stubCategoryReader) GetCategoriesDtoByUserId(userId int) ([]*types.CategoryDTO, error) {
	return s.categories, s.err
}

type stubAccountReader struct {
	accounts []*types.Account
	err      error
}

func (s *stubAccountReader) GetAccountsByUserId(userId int) ([]*types.Account, error) {
	return s.accounts, s.err
}

// recordingTransactionWriter captures what would have been written, and can be
// capped so later writes fail — the partial-failure case.
type recordingTransactionWriter struct {
	created []*types.Transaction
	userIDs []int
	// successLimit is how many writes succeed before the store starts failing.
	// -1 means every write succeeds.
	successLimit int
	balance      float64
}

func newRecordingTransactionWriter(balance float64) *recordingTransactionWriter {
	return &recordingTransactionWriter{successLimit: -1, balance: balance}
}

func (w *recordingTransactionWriter) CreateTransactionAndReturn(transaction *types.Transaction, userId int) (*types.TransactionChangeResponse, error) {
	if w.successLimit >= 0 && len(w.created) >= w.successLimit {
		return nil, fmt.Errorf("insert failed")
	}

	w.created = append(w.created, transaction)
	w.userIDs = append(w.userIDs, userId)

	balance := w.balance
	return &types.TransactionChangeResponse{AccountBalance: &balance}, nil
}

type memoryPendingStore struct {
	parses map[string]*types.PendingParse
}

func newMemoryPendingStore() *memoryPendingStore {
	return &memoryPendingStore{parses: map[string]*types.PendingParse{}}
}

func (s *memoryPendingStore) UpsertPendingParse(p *types.PendingParse) error {
	copied := *p
	s.parses[p.ChatID] = &copied
	return nil
}

func (s *memoryPendingStore) GetPendingParse(chatID string) (*types.PendingParse, error) {
	parse, ok := s.parses[chatID]
	if !ok {
		return nil, nil
	}
	return parse, nil
}

func (s *memoryPendingStore) DeletePendingParse(chatID string) error {
	delete(s.parses, chatID)
	return nil
}

func (s *memoryPendingStore) DeletePendingParsesByUserID(userID int) error {
	for chatID, parse := range s.parses {
		if parse.UserID == userID {
			delete(s.parses, chatID)
		}
	}
	return nil
}
