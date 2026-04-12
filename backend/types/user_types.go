package types

import (
	"context"
	"time"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserById(id int) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	ValidatePassword(password string) error
	DeleteUser(userId int) error
	MarkEmailVerified(userID int, verified bool) error
	UpdateMfaMethod(userID int, method MfaMethod) error
	UpdatePassword(userID int, hashedPassword string) error
}

type User struct {
	ID           int        `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Email        string     `json:"email"`
	Password     string     `json:"-"`
	EmailVerified bool      `json:"email_verified"`
	MfaMethod    MfaMethod  `json:"mfa_method"`
	CreatedAt    string     `json:"created_at"`
}

/* ==============================
* GDPR Export Data Structures
* ============================== */

type ExportData struct {
	User         *User             `json:"user"`
	Accounts     []*Account        `json:"accounts"`
	Categories   []*CategoryDTO    `json:"categories"`
	Transactions []*TransactionDTO `json:"transactions"`
	ExportedAt   time.Time         `json:"exported_at"`
}
