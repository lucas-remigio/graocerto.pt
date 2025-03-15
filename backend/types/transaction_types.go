package types

import "time"

type TransactionStore interface {
	GetTransactionsByAccountToken(accountToken string) ([]*Transaction, error)
	CreateTransaction(transaction *Transaction) error
	GetTransactionsDTOByAccountToken(accountToken string) ([]*TransactionDTO, error)
	UpdateTransaction(transaction *UpdateTransactionPayload) error
	DeleteTransaction(transactionId int, userId int) error
}

type CreateTransactionPayload struct {
	AccountToken string  `json:"account_token" validate:"required"`
	CategoryID   int     `json:"category_id" validate:"numeric,min=1"`
	Amount       float64 `json:"amount" validate:"required,numeric"`
	Description  string  `json:"description"`
	Date         string  `json:"date" validate:"required"`
}

type UpdateTransactionPayload struct {
	ID          int     `json:"id" validate:"required,numeric"`
	Amount      float64 `json:"amount" validate:"required,numeric"`
	CategoryID  int     `json:"category_id" validate:"numeric,min=1"`
	Description string  `json:"description"`
	Date        string  `json:"date" validate:"required"`
}

type Transaction struct {
	ID           int     `json:"id"`
	AccountToken string  `json:"account_token"`
	CategoryId   int     `json:"category_id"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Date         string  `json:"date"`
	Balance      float64 `json:"balance"`
	CreatedAt    string  `json:"created_at"`
}

type TransactionDTO struct {
	ID           int          `json:"id"`
	AccountToken string       `json:"account_token"`
	Amount       float64      `json:"amount"`
	Description  string       `json:"description"`
	Date         time.Time    `json:"date"`
	Balance      float64      `json:"balance"`
	CreatedAt    time.Time    `json:"created_at"`
	Category     *CategoryDTO `json:"category,omitempty"`
}
