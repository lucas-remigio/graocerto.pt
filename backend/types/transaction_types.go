package types

import "time"

type TransactionStore interface {
	GetTransactionsByAccountToken(accountToken string, month, year *int) ([]*Transaction, error)
	GetTransactionsDTOByAccountToken(accountToken string, month, year *int) ([]*TransactionDTO, error)
	GetTransactionDTOById(id int) (*TransactionDTO, error)
	CreateTransaction(transaction *Transaction, userId int) (*Transaction, error)
	CreateTransactionAndReturn(transaction *Transaction, userId int) (*TransactionChangeResponse, error)
	UpdateTransaction(transaction *UpdateTransactionPayload, userId int) (*Transaction, error)
	UpdateTransactionAndReturn(payload *UpdateTransactionPayload, userId int) (*TransactionChangeResponse, error)
	DeleteTransaction(transactionId int, userId int) (balance *float64, err error)
	DeleteTransactionAndReturn(transactionId int, userId int) (*TransactionChangeResponse, error)
	GetAvailableTransactionMonthsByAccountToken(accountToken string) ([]*MonthYear, error)
	CalculateTransactionTotals(transactions []*TransactionDTO) (*TransactionTotals, error)
	GetTransactionStatistics(accountToken string, month, year *int) (*TransactionStatistics, error)
}

type CreateTransactionPayload struct {
	AccountToken string  `json:"account_token" validate:"required,min=1,max=255"`
	CategoryID   int     `json:"category_id" validate:"numeric,min=1,max=999999999"`
	Amount       float64 `json:"amount" validate:"required,numeric,gte=0,lte=999999999"`
	Description  string  `json:"description" validate:"max=255"`
	Date         string  `json:"date" validate:"required"`
}

type UpdateTransactionPayload struct {
	// id not required as it is sent on the url
	ID          int     `json:"id" validate:"numeric"`
	Amount      float64 `json:"amount" validate:"required,numeric,gte=0,lte=999999999"`
	CategoryID  int     `json:"category_id" validate:"numeric,min=1,max=999999999"`
	Description string  `json:"description" validate:"max=255"`
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

type TransactionChangeResponse struct {
	Transaction    *TransactionDTO `json:"transaction"`
	AccountBalance *float64        `json:"account_balance,omitempty"`
	Months         []*MonthYear    `json:"months"`
}

type TransactionsResponse struct {
	Transactions []*TransactionDTO `json:"transactions"`
}

type MonthYear struct {
	Month int `json:"month"`
	Year  int `json:"year"`
	Count int `json:"count"`
}

type TransactionTotals struct {
	Debit      float64 `json:"debit"`
	Credit     float64 `json:"credit"`
	Difference float64 `json:"difference"`
}

type CategoryStatistic struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Total      float64 `json:"total"`
	Percentage float64 `json:"percentage"`
	Color      string  `json:"color"`
}

type DailyTotal struct {
	Date       string  `json:"date"`
	Credit     float64 `json:"credit"`
	Debit      float64 `json:"debit"`
	Difference float64 `json:"difference"`
}

type TransactionStatistics struct {
	TotalTransactions       int                  `json:"total_transactions"`
	LargestDebit            float64              `json:"largest_debit"`
	LargestCredit           float64              `json:"largest_credit"`
	CreditCategoryBreakdown []*CategoryStatistic `json:"credit_category_breakdown"`
	DebitCategoryBreakdown  []*CategoryStatistic `json:"debit_category_breakdown"`
	Totals                  *TransactionTotals   `json:"totals"`
	DailyTotals             []*DailyTotal        `json:"daily_totals"`
	StartDate               string               `json:"start_date"` // Format: YYYY-MM-DD
	EndDate                 string               `json:"end_date"`   // Format: YYYY-MM-DD
}
