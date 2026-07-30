package types

import "time"

type TransactionStore interface {
	GetTransactionsByAccountToken(accountToken string, month, year *int) ([]*Transaction, error)
	GetTransactionsDTOByAccountToken(accountToken string, month, year *int) ([]*TransactionDTO, error)
	GetTransactionDTOById(id int) (*TransactionDTO, error)
	CreateTransaction(transaction *Transaction, userId int) (*Transaction, error)
	CreateTransactionAndReturn(transaction *Transaction, userId int) (*TransactionChangeResponse, error)
	CreateTransfer(payload *CreateTransferPayload, userId int) (*TransferResponse, error)
	UpdateTransaction(transaction *UpdateTransactionPayload, userId int) (*Transaction, error)
	UpdateTransactionAndReturn(payload *UpdateTransactionPayload, userId int) (*TransactionChangeResponse, error)
	DeleteTransaction(transactionId int, userId int) (balance *float64, err error)
	DeleteTransactionAndReturn(transactionId int, userId int) (*TransactionChangeResponse, error)
	ApprovePendingTransactionAndReturn(transactionID int, userID int) (*TransactionChangeResponse, error)
	RejectPendingTransactionAndReturn(transactionID int, userID int) (*TransactionChangeResponse, error)
	GetAvailableTransactionMonthsByAccountToken(accountToken string) ([]*MonthYear, error)
	CalculateTransactionTotals(transactions []*TransactionDTO) (*TransactionTotals, error)
	GetTransactionStatistics(accountToken string, month, year *int) (*TransactionStatistics, error)
	GetStatisticsComparison(accountToken string, month, year int) (*StatisticsComparison, error)
}

// CategoryDelta is one category's month-over-month change. PercentageDelta is
// nil when it can't be expressed honestly (no prior spend, or the category is
// new/removed) — the frontend renders those as "new"/"—" instead of a
// misleading ±100%.
type CategoryDelta struct {
	ID              int      `json:"id"`
	ParentID        *int     `json:"parent_id,omitempty"`
	Name            string   `json:"name"`
	Color           string   `json:"color"`
	CurrentTotal    float64  `json:"current_total"`
	PreviousTotal   float64  `json:"previous_total"`
	AbsoluteDelta   float64  `json:"absolute_delta"`
	PercentageDelta *float64 `json:"percentage_delta"`
	// Status is one of: "new", "removed", "increased", "decreased", "unchanged".
	Status string `json:"status"`
}

// StatisticsComparison diffs a month against the previous one (with year
// rollover for January → previous December).
type StatisticsComparison struct {
	AccountToken         string             `json:"account_token"`
	CurrentMonth         int                `json:"current_month"`
	CurrentYear          int                `json:"current_year"`
	PreviousMonth        int                `json:"previous_month"`
	PreviousYear         int                `json:"previous_year"`
	CurrentTotals        *TransactionTotals `json:"current_totals"`
	PreviousTotals       *TransactionTotals `json:"previous_totals"`
	TotalsDelta          *TransactionTotals `json:"totals_delta"`
	DebitCategoryDeltas  []*CategoryDelta   `json:"debit_category_deltas"`
	CreditCategoryDeltas []*CategoryDelta   `json:"credit_category_deltas"`
}

type CreateTransactionPayload struct {
	AccountToken string  `json:"account_token" validate:"required,min=1,max=255"`
	CategoryID   int     `json:"category_id" validate:"numeric,min=1,max=999999999"`
	Amount       float64 `json:"amount" validate:"required,numeric,gte=0,lte=999999999"`
	Description  string  `json:"description" validate:"max=255"`
	Date         string  `json:"date" validate:"required"`
}
type CreateTransferPayload struct {
	SourceAccountToken      string  `json:"source_account_token" validate:"required,min=1,max=255"`
	DestinationAccountToken string  `json:"destination_account_token" validate:"required,min=1,max=255"`
	DebitCategoryID         int     `json:"debit_category_id" validate:"required,numeric,min=1"`
	CreditCategoryID        int     `json:"credit_category_id" validate:"required,numeric,min=1"`
	Amount                  float64 `json:"amount" validate:"required,numeric,gt=0,lte=999999999"`
	Description             string  `json:"description" validate:"max=255"`
	Date                    string  `json:"date" validate:"required"`
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
	ID                int     `json:"id"`
	AccountToken      string  `json:"account_token"`
	TransactionTypeId int     `json:"transaction_type_id"`
	CategoryId        int     `json:"category_id"`
	Amount            float64 `json:"amount"`
	Description       string  `json:"description"`
	Date              string  `json:"date"`
	Balance           float64 `json:"balance"`
	IsPending         bool    `json:"is_pending"`
	TransferGroupId   *string `json:"transfer_group_id,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

type TransactionDTO struct {
	ID              int              `json:"id"`
	AccountToken    string           `json:"account_token"`
	Amount          float64          `json:"amount"`
	Description     string           `json:"description"`
	Date            time.Time        `json:"date"`
	Balance         float64          `json:"balance"`
	IsPending       bool             `json:"is_pending"`
	CreatedAt       time.Time        `json:"created_at"`
	Category        *CategoryDTO     `json:"category,omitempty"`
	TransferGroupId *string          `json:"transfer_group_id,omitempty"`
	TransactionType *TransactionType `json:"transaction_type,omitempty"`
}

type TransactionChangeResponse struct {
	Transaction                 *TransactionDTO `json:"transaction"`
	AccountBalance              *float64        `json:"account_balance,omitempty"`
	AccountPendingBalance       *float64        `json:"account_pending_balance,omitempty"`
	Months                      []*MonthYear    `json:"months"`
	PairedAccountToken          *string         `json:"paired_account_token,omitempty"`
	PairedAccountBalance        *float64        `json:"paired_account_balance,omitempty"`
	PairedAccountPendingBalance *float64        `json:"paired_account_pending_balance,omitempty"`
	PairedAccountMonths         []*MonthYear    `json:"paired_account_months,omitempty"`
	IsTransfer                  bool            `json:"is_transfer"`
}
type TransferResponse struct {
	TransferGroupID           string          `json:"transfer_group_id"`
	DebitTransaction          *TransactionDTO `json:"debit_transaction"`
	CreditTransaction         *TransactionDTO `json:"credit_transaction"`
	SourceAccountBalance      float64         `json:"source_account_balance"`
	DestinationAccountBalance float64         `json:"destination_account_balance"`
	SourceAccountMonths       []*MonthYear    `json:"source_account_months"`
	DestinationAccountMonths  []*MonthYear    `json:"destination_account_months"`
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
	ID               int                 `json:"id"`
	ParentID         *int                `json:"parent_id,omitempty"`
	Name             string              `json:"name"`
	Count            int                 `json:"count"`
	Total            float64             `json:"total"`
	Percentage       float64             `json:"percentage"`
	Color            string              `json:"color"`
	Budget           *int                `json:"budget"`
	BudgetPercentage float64             `json:"budget_percentage"`
	Subcategories    []CategoryStatistic `json:"subcategories,omitempty"`
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
