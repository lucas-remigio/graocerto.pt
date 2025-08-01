package types

type AccountStore interface {
	GetAccountsByUserId(userId int) ([]*Account, error)
	GetAccountByToken(token string, userId int) (*Account, error)
	CreateAccount(account *Account) error
	UpdateAccount(account *Account, userId int) error
	DeleteAccount(token string, userId int) error
	GetAccountFeedbackMonthly(userId int, accountToken, language string, month, year int) (*MonthlyFeedback, error)
	ReorderAccounts(userId int, accounts []ReorderAccount) error
	FavoriteAccount(token string, userId int, isFavorite bool) error
}

type CreateAccountPayload struct {
	AccountName string   `json:"account_name" validate:"required,min=3,max=50"`
	Balance     *float64 `json:"balance" validate:"required,gte=0,lt=100000000"`
}

type UpdateAccountPayload struct {
	AccountName string   `json:"account_name" validate:"required,min=3,max=50"`
	Balance     *float64 `json:"balance" validate:"required,gte=0,lt=100000000"`
}

type ReorderAccountsPayload struct {
	Accounts []ReorderAccount `json:"accounts" validate:"required,dive"`
}

type ReorderAccount struct {
	Token      string `json:"token" validate:"required"`
	OrderIndex int    `json:"order_index" validate:"required,gte=0"`
}

type FavoriteAccountPayload struct {
	IsFavorite bool `json:"is_favorite"`
}

type Account struct {
	ID          int     `json:"id"`
	Token       string  `json:"token"`
	UserID      int     `json:"user_id"`
	AccountName string  `json:"account_name"`
	Balance     float64 `json:"balance"`
	CreatedAt   string  `json:"created_at"`
	OrderIndex  int     `json:"order_index"`
	IsFavorite  bool    `json:"is_favorite"`
}
