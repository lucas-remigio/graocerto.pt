package telegram

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/types"
)

// PendingTTL is how long a half-answered conversation stays alive. Short on
// purpose: an abandoned parse must not resurface hours later and get confirmed.
const PendingTTL = 15 * time.Minute

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// UpsertPendingParse replaces whatever the chat had pending. Expiry is computed
// with the database clock so reads and writes never disagree about "now".
func (s *Store) UpsertPendingParse(p *types.PendingParse) error {
	items, err := json.Marshal(p.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal pending items: %w", err)
	}

	_, err = db.ExecWithValidation(s.db,
		`INSERT INTO telegram_pending_transactions
			(user_id, chat_id, transactions, account_token, expires_at)
		 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP + make_interval(secs => $5))
		 ON CONFLICT (chat_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			transactions = EXCLUDED.transactions,
			account_token = EXCLUDED.account_token,
			created_at = CURRENT_TIMESTAMP,
			expires_at = EXCLUDED.expires_at`,
		p.UserID, p.ChatID, items, p.AccountToken, PendingTTL.Seconds(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert pending parse: %w", err)
	}

	return nil
}

// GetPendingParse returns nil when the chat has no parse or it already expired.
// Expiry is a read-time filter, so a stale row is simply invisible until the
// next parse overwrites it.
func (s *Store) GetPendingParse(chatID string) (*types.PendingParse, error) {
	pending, err := db.QueryFirstFromRows(s.db,
		`SELECT id, user_id, chat_id, transactions, account_token, created_at, expires_at
		 FROM telegram_pending_transactions
		 WHERE chat_id = $1 AND expires_at > CURRENT_TIMESTAMP`,
		scanRowIntoPendingParse, chatID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pending parse: %w", err)
	}

	return pending, nil
}

func (s *Store) DeletePendingParse(chatID string) error {
	_, err := db.ExecWithValidation(s.db,
		"DELETE FROM telegram_pending_transactions WHERE chat_id = $1",
		chatID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete pending parse: %w", err)
	}

	return nil
}

// DeletePendingParsesByUserID is used on unlink, so an in-flight conversation
// cannot outlive the link that authorised it.
func (s *Store) DeletePendingParsesByUserID(userID int) error {
	_, err := db.ExecWithValidation(s.db,
		"DELETE FROM telegram_pending_transactions WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete pending parses for user: %w", err)
	}

	return nil
}

func scanRowIntoPendingParse(rows *sql.Rows) (*types.PendingParse, error) {
	pending := new(types.PendingParse)
	var items []byte
	var accountToken sql.NullString

	if err := rows.Scan(
		&pending.ID,
		&pending.UserID,
		&pending.ChatID,
		&items,
		&accountToken,
		&pending.CreatedAt,
		&pending.ExpiresAt,
	); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(items, &pending.Items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pending items: %w", err)
	}

	if accountToken.Valid {
		pending.AccountToken = &accountToken.String
	}

	return pending, nil
}
