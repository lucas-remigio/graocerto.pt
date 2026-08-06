package user

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/types"
)

func (s *Store) LinkTelegramChatID(userID int, chatID string) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE users SET telegram_chat_id = $1 WHERE id = $2",
		chatID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to link telegram chat: %w", err)
	}

	slog.Info("telegram chat linked", "userID", userID)
	return nil
}

func (s *Store) GetUserByTelegramChatID(chatID string) (*types.User, error) {
	user, err := db.QueryFirstFromRows(s.db,
		userSelectColumns+" FROM users WHERE telegram_chat_id = $1",
		scanRowIntoUser, chatID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (s *Store) UnlinkTelegram(userID int) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE users SET telegram_chat_id = NULL WHERE id = $1",
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to unlink telegram chat: %w", err)
	}

	slog.Info("telegram chat unlinked", "userID", userID)
	return nil
}
