package user

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/service/auth"
	"github.com/lucas-remigio/wallet-tracker/types"
)

var errTrustedDeviceNotFound = errors.New("trusted device not found")

const trustedDeviceColumns = "id, user_id, token_hash, label, created_at, last_used_at, expires_at"

func (s *Store) CreateTrustedDevice(device *types.TrustedDevice) error {
	// Prune this user's expired rows on write so the table doesn't need a cleanup job.
	if _, err := db.ExecWithValidation(s.db,
		"DELETE FROM trusted_devices WHERE user_id = $1 AND expires_at < CURRENT_TIMESTAMP",
		device.UserID,
	); err != nil {
		return fmt.Errorf("prune expired trusted devices: %w", err)
	}

	_, err := db.ExecWithValidation(s.db,
		"INSERT INTO trusted_devices (user_id, token_hash, label, expires_at) VALUES ($1, $2, $3, $4)",
		device.UserID, device.TokenHash, device.Label, device.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create trusted device: %w", err)
	}
	return nil
}

// GetTrustedDeviceByToken returns the device only while it is still valid.
func (s *Store) GetTrustedDeviceByToken(token string) (*types.TrustedDevice, error) {
	device, err := db.QueryFirstFromRows(s.db,
		"SELECT "+trustedDeviceColumns+" FROM trusted_devices WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP",
		scanRowIntoTrustedDevice, auth.HashSecret(token),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errTrustedDeviceNotFound
	}
	return device, err
}

func (s *Store) ListTrustedDevices(userID int) ([]*types.TrustedDevice, error) {
	return db.QueryList(s.db,
		"SELECT "+trustedDeviceColumns+" FROM trusted_devices WHERE user_id = $1 AND expires_at > CURRENT_TIMESTAMP ORDER BY last_used_at DESC",
		scanRowIntoTrustedDevice, userID,
	)
}

func (s *Store) TouchTrustedDevice(id int) error {
	_, err := db.ExecWithValidation(s.db,
		"UPDATE trusted_devices SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1",
		id,
	)
	return err
}

// DeleteTrustedDevice is scoped by user so one user can never revoke another's device.
func (s *Store) DeleteTrustedDevice(userID, id int) error {
	res, err := db.ExecWithValidation(s.db,
		"DELETE FROM trusted_devices WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete trusted device: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete trusted device: %w", err)
	}
	if affected == 0 {
		return errTrustedDeviceNotFound
	}
	return nil
}

func (s *Store) DeleteTrustedDevicesByUser(userID int) error {
	_, err := db.ExecWithValidation(s.db, "DELETE FROM trusted_devices WHERE user_id = $1", userID)
	return err
}

func scanRowIntoTrustedDevice(rows *sql.Rows) (*types.TrustedDevice, error) {
	device := new(types.TrustedDevice)
	err := rows.Scan(
		&device.ID,
		&device.UserID,
		&device.TokenHash,
		&device.Label,
		&device.CreatedAt,
		&device.LastUsedAt,
		&device.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return device, nil
}
