package types

type RecurringFrequency string

const (
	RecurringDaily      RecurringFrequency = "daily"
	RecurringWeekly     RecurringFrequency = "weekly"
	RecurringMonthly    RecurringFrequency = "monthly"
	RecurringEveryXDays RecurringFrequency = "every_x_days"
)

type RecurringRuleStore interface {
	CreateRecurringRule(rule *RecurringRule) (*RecurringRule, error)
	CreateRecurringTransfer(payload *CreateRecurringTransferPayload, userID int) ([]*RecurringRule, error)
	UpdateRecurringRule(rule *RecurringRule, userID int) (*RecurringRule, error)
	UpdateRecurringTransfer(groupID string, payload *UpdateRecurringTransferPayload, userID int) ([]*RecurringRule, error)
	DeleteRecurringRule(id int, userID int) error
	GetRecurringRuleByID(id int, userID int) (*RecurringRule, error)
	GetRecurringRulesByUserID(userID int) ([]*RecurringRule, error)
	GeneratePendingTransactionsForDueRules() error
}

type RecurringRule struct {
	ID                       int                `json:"id"`
	UserID                   int                `json:"user_id"`
	AccountToken             string             `json:"account_token"`
	CategoryID               int                `json:"category_id"`
	TransactionTypeID        int                `json:"transaction_type_id,omitempty"`
	RecurringTransferGroupID *string            `json:"recurring_transfer_group_id,omitempty"`
	Amount                   float64            `json:"amount"`
	Description              string             `json:"description"`
	Frequency                RecurringFrequency `json:"frequency"`
	IntervalValue            int                `json:"interval_value"`
	NextRunDate              string             `json:"next_run_date"`
	Active                   bool               `json:"active"`
	CreatedAt                string             `json:"created_at"`
	UpdatedAt                string             `json:"updated_at"`
}

type CreateRecurringRulePayload struct {
	AccountToken      string             `json:"account_token" validate:"required,min=1,max=255"`
	CategoryID        int                `json:"category_id" validate:"required,min=1"`
	TransactionTypeID int                `json:"transaction_type_id" validate:"required,oneof=1 2"`
	Amount            float64            `json:"amount" validate:"required,gt=0,lte=999999999"`
	Description       string             `json:"description" validate:"max=255"`
	Frequency         RecurringFrequency `json:"frequency" validate:"required,oneof=daily weekly monthly every_x_days"`
	IntervalValue     int                `json:"interval_value" validate:"required,min=1,max=365"`
	ExecutionDay      *int               `json:"execution_day" validate:"omitempty,min=1,max=31"`
	Active            *bool              `json:"active"`
}

type UpdateRecurringRulePayload struct {
	AccountToken      string             `json:"account_token" validate:"required,min=1,max=255"`
	CategoryID        int                `json:"category_id" validate:"required,min=1"`
	TransactionTypeID int                `json:"transaction_type_id" validate:"required,oneof=1 2"`
	Amount            float64            `json:"amount" validate:"required,gt=0,lte=999999999"`
	Description       string             `json:"description" validate:"max=255"`
	Frequency         RecurringFrequency `json:"frequency" validate:"required,oneof=daily weekly monthly every_x_days"`
	IntervalValue     int                `json:"interval_value" validate:"required,min=1,max=365"`
	ExecutionDay      *int               `json:"execution_day" validate:"omitempty,min=1,max=31"`
	NextRunDate       string             `json:"next_run_date" validate:"omitempty"`
	Active            bool               `json:"active"`
}

type RecurringRulesResponse struct {
	RecurringRules []*RecurringRule `json:"recurring_rules"`
}

type CreateRecurringTransferPayload struct {
	SourceAccountToken      string             `json:"source_account_token" validate:"required,min=1,max=255"`
	DestinationAccountToken string             `json:"destination_account_token" validate:"required,min=1,max=255"`
	DebitCategoryID         int                `json:"debit_category_id" validate:"required,min=1"`
	CreditCategoryID        int                `json:"credit_category_id" validate:"required,min=1"`
	Amount                  float64            `json:"amount" validate:"required,gt=0,lte=999999999"`
	Description             string             `json:"description" validate:"max=255"`
	Frequency               RecurringFrequency `json:"frequency" validate:"required,oneof=daily weekly monthly every_x_days"`
	IntervalValue           int                `json:"interval_value" validate:"required,min=1,max=365"`
	ExecutionDay            *int               `json:"execution_day" validate:"omitempty,min=1,max=31"`
	Active                  *bool              `json:"active"`
}

type UpdateRecurringTransferPayload struct {
	SourceAccountToken      string             `json:"source_account_token" validate:"required,min=1,max=255"`
	DestinationAccountToken string             `json:"destination_account_token" validate:"required,min=1,max=255"`
	DebitCategoryID         int                `json:"debit_category_id" validate:"required,min=1"`
	CreditCategoryID        int                `json:"credit_category_id" validate:"required,min=1"`
	Amount                  float64            `json:"amount" validate:"required,gt=0,lte=999999999"`
	Description             string             `json:"description" validate:"max=255"`
	Frequency               RecurringFrequency `json:"frequency" validate:"required,oneof=daily weekly monthly every_x_days"`
	IntervalValue           int                `json:"interval_value" validate:"required,min=1,max=365"`
	ExecutionDay            *int               `json:"execution_day" validate:"omitempty,min=1,max=31"`
	NextRunDate             string             `json:"next_run_date" validate:"omitempty"`
	Active                  bool               `json:"active"`
}
