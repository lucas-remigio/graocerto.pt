package user

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/types"
)

func (s *Store) GetTOTPState(userID int) (*types.TOTPState, error) {
	state := new(types.TOTPState)
	var secret, pending sql.NullString

	err := s.db.QueryRow(
		"SELECT totp_secret, totp_pending_secret, totp_last_step FROM users WHERE id = $1",
		userID,
	).Scan(&secret, &pending, &state.LastStep)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get totp state: %w", err)
	}

	state.Secret = secret.String
	state.PendingSecret = pending.String
	return state, nil
}

func (s *Store) SetPendingTOTPSecret(userID int, encryptedSecret string) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE users SET totp_pending_secret = $1 WHERE id = $2",
		encryptedSecret, userID,
	)
	return err
}

func (s *Store) EnableTOTP(userID int, lastStep int64) error {
	res, err := db.ExecWithValidation(s.db,
		`UPDATE users
		 SET totp_secret = totp_pending_secret, totp_pending_secret = NULL,
		     totp_last_step = $1, mfa_method = $2
		 WHERE id = $3 AND totp_pending_secret IS NOT NULL`,
		lastStep, string(types.MfaMethodTOTP), userID,
	)
	if err != nil {
		return fmt.Errorf("enable totp: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return fmt.Errorf("enable totp: no pending secret")
	}
	return nil
}

func (s *Store) DisableTOTP(userID int) error {
	_, err := db.ExecWithValidation(s.db,
		`UPDATE users
		 SET totp_secret = NULL, totp_pending_secret = NULL, totp_last_step = 0, mfa_method = $1
		 WHERE id = $2`,
		string(types.MfaMethodEmailOTP), userID,
	)
	return err
}

// AdvanceTOTPStep is a single conditional UPDATE so two concurrent logins with
// the same code can't both succeed.
func (s *Store) AdvanceTOTPStep(userID int, step int64) (bool, error) {
	res, err := db.ExecWithValidation(s.db,
		"UPDATE users SET totp_last_step = $1 WHERE id = $2 AND totp_last_step < $1",
		step, userID,
	)
	if err != nil {
		return false, fmt.Errorf("advance totp step: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("advance totp step: %w", err)
	}
	return affected == 1, nil
}
