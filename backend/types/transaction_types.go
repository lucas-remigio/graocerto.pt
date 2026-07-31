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
	GetCategoryMonthlyTrends(accountToken string, months, transactionTypeID, compareBase, compareCurrent int) (*CategoryTrendsResponse, error)
}

// TrendMonth is one point on the trends x-axis.
type TrendMonth struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

// CategoryTrend is one category's spend/income per month, aligned index-for-index
// with CategoryTrendsResponse.Months. Total is the window sum (used for ordering
// the picker so the biggest categories surface first).
type CategoryTrend struct {
	ID     int       `json:"id"`
	Name   string    `json:"name"`
	Color  string    `json:"color"`
	Totals []float64 `json:"totals"`
	Total  float64   `json:"total"`
	// Subcategories carry their own series; only populated on root categories.
	// The parent's Totals already include these (rolled up), matching the stats view.
	Subcategories []*CategoryTrend `json:"subcategories,omitempty"`
}

// CategoryMover is a category whose recent-half average moved meaningfully vs the
// earlier half of the window. Pct is nil for a "new" category (no earlier spend).
// Direction is one of: "up", "down", "new".
type CategoryMover struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Totals    []float64 `json:"totals"`
	Pct       *float64  `json:"pct"`
	Direction string    `json:"direction"`
}

// CategoryTrendsResponse is a chart-ready time series over the last N months:
// a shared month axis, the per-month grand total, and a per-category series
// (subcategories rolled up into their parent, matching the statistics view).
// WindowTotal/MonthlyAverage and Movers are computed server-side so the client
// only renders.
type CategoryTrendsResponse struct {
	AccountToken    string           `json:"account_token"`
	TransactionType int              `json:"transaction_type"`
	Months          []TrendMonth     `json:"months"`
	Totals          []float64        `json:"totals"`
	// Income is per-month credit total (transfers-in excluded), aligned to Months,
	// regardless of the selected TransactionType — it's always the credit side so the
	// client can express any category as a share of income. WindowIncome is its sum.
	Income         []float64        `json:"income"`
	WindowTotal    float64          `json:"window_total"`
	WindowIncome   float64          `json:"window_income"`
	MonthlyAverage float64          `json:"monthly_average"`
	Categories     []*CategoryTrend `json:"categories"`
	Movers         []*CategoryMover `json:"movers"`
	// CompareBase/CompareCurrent are the axis indices of the two months the movers
	// compare. Echoed back so the client's month pickers reflect the resolved
	// default (or the pair the client requested).
	CompareBase    int `json:"compare_base"`
	CompareCurrent int `json:"compare_current"`
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
