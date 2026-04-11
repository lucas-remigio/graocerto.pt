package user

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

func (s *Store) MarkEmailVerified(userID int, verified bool) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE users SET email_verified = $1 WHERE id = $2",
		verified, userID,
	)
	return err
}

func (s *Store) UpdateMfaMethod(userID int, method types.MfaMethod) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE users SET mfa_method = $1 WHERE id = $2",
		string(method), userID,
	)
	return err
}

func (s *Store) UpdatePassword(userID int, hashedPassword string) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE users SET password = $1 WHERE id = $2",
		hashedPassword, userID,
	)
	return err
}

func (s *Store) CreateAuthToken(token *types.AuthToken) error {
	_, err := db.ExecWithValidation(s.db,
		`INSERT INTO auth_tokens
			(id, user_id, purpose, secret_hash, expires_at, attempts, max_attempts)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		token.ID, token.UserID, string(token.Purpose), token.SecretHash, token.ExpiresAt, token.Attempts, token.MaxAttempts,
	)
	return err
}

func (s *Store) GetAuthTokenByID(id string) (*types.AuthToken, error) {
	token, err := db.QueryFirstFromRows(s.db,
		`SELECT id, user_id, purpose, secret_hash, expires_at, consumed_at, attempts, max_attempts, created_at
		 FROM auth_tokens WHERE id = $1`,
		scanRowIntoAuthToken, id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("auth token not found")
		}
		return nil, err
	}
	return token, nil
}

func (s *Store) GetAuthTokenByPurposeAndSecret(purpose types.AuthTokenPurpose, secret string) (*types.AuthToken, error) {
	secretHash := auth.HashSecret(secret)
	token, err := db.QueryFirstFromRows(s.db,
		`SELECT id, user_id, purpose, secret_hash, expires_at, consumed_at, attempts, max_attempts, created_at
		 FROM auth_tokens WHERE purpose = $1 AND secret_hash = $2`,
		scanRowIntoAuthToken, string(purpose), secretHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("auth token not found")
		}
		return nil, err
	}
	return token, nil
}

func (s *Store) ConsumeAuthToken(id string) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE auth_tokens SET consumed_at = CURRENT_TIMESTAMP WHERE id = $1",
		id,
	)
	return err
}

func (s *Store) IncrementAuthTokenAttempts(id string) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE auth_tokens SET attempts = attempts + 1 WHERE id = $1",
		id,
	)
	return err
}

func (s *Store) DeleteAuthTokensByUserAndPurpose(userID int, purpose types.AuthTokenPurpose) error {
	_, err := db.ExecWithValidation(s.db,
		"DELETE FROM auth_tokens WHERE user_id = $1 AND purpose = $2 AND consumed_at IS NULL",
		userID, string(purpose),
	)
	return err
}

func scanRowIntoAuthToken(rows *sql.Rows) (*types.AuthToken, error) {
	token := new(types.AuthToken)
	var consumedAt sql.NullTime

	if err := rows.Scan(
		&token.ID,
		&token.UserID,
		&token.Purpose,
		&token.SecretHash,
		&token.ExpiresAt,
		&consumedAt,
		&token.Attempts,
		&token.MaxAttempts,
		&token.CreatedAt,
	); err != nil {
		return nil, err
	}

	if consumedAt.Valid {
		token.ConsumedAt = &consumedAt.Time
	}

	return token, nil
}

func isExpiredAuthToken(token *types.AuthToken) bool {
	return time.Now().After(token.ExpiresAt)
}
