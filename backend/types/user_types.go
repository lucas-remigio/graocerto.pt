package types

import (
	"time"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserById(id int) (*User, error)
	CreateUser(user *User) error
	ValidatePassword(password string) error
	DeleteUser(userId int) error
}

type RegisterUserPayload struct {
	FirstName string `json:"first_name" validate:"required,max=32"`
	LastName  string `json:"last_name" validate:"required,max=32"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=64"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
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
