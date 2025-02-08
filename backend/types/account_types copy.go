package types

type AccountStore interface {
	GetAccountsByUserId(userId int) ([]*Account, error)
	CreateAccount(account *Account) error
}

type CreateAccountPayload struct {
	AccountName string  `json:"account_name" validate:"required,min=3"`
	Balance     float64 `json:"balance" validate:"required,gte=0"`
}

type Account struct {
	ID          int     `json:"id"`
	Token       string  `json:"token"`
	UserID      int     `json:"user_id"`
	AccountName string  `json:"account_name"`
	Balance     float64 `json:"balance"`
	CreatedAt   string  `json:"created_at"`
}
