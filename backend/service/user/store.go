package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	user, err := db.QueryFirstFromRows(s.db,
		"SELECT id, first_name, last_name, email, password, email_verified, mfa_method, created_at FROM users WHERE email = $1",
		scanRowIntoUser, email)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Store) CreateUser(ctx context.Context, user *types.User) error {
	_, err := db.ExecWithValidation(s.db,
		"INSERT INTO users (first_name, last_name, email, password, email_verified, mfa_method) VALUES ($1, $2, $3, $4, $5, $6)",
		user.FirstName, user.LastName, user.Email, user.Password, user.EmailVerified, string(user.MfaMethod))

	if err != nil {
		utils.LogWithContext(ctx).Error("failed to create user", "email", user.Email, "error", err)
		return err
	}

	utils.LogWithContext(ctx).Info("new user created", "email", user.Email, "first_name", user.FirstName)
	return nil
}

func (s *Store) GetUserById(id int) (*types.User, error) {
	user, err := db.QueryFirstFromRows(s.db,
		"SELECT id, first_name, last_name, email, password, email_verified, mfa_method, created_at FROM users WHERE id = $1",
		scanRowIntoUser, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Store) DeleteUser(userId int) error {
	// Optional: Check if user exists first
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", userId).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("user not found")
	}

	// Delete the user (will cascade to all related data)
	_, err = db.ExecWithValidation(s.db,
		"DELETE FROM users WHERE id = $1",
		userId)

	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}

func scanRowIntoUser(rows *sql.Rows) (*types.User, error) {
	user := new(types.User)

	err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.EmailVerified, &user.MfaMethod, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil

}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isSpecial(c rune) bool {
	return !isUpper(c) && !isLower(c) && !isDigit(c)
}

func (s *Store) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		if isUpper(c) {
			hasUpper = true
		} else if isLower(c) {
			hasLower = true
		} else if isDigit(c) {
			hasDigit = true
		} else if isSpecial(c) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
