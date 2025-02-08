package account

import (
	"database/sql"

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

func (s *Store) CreateAccount(account *types.Account) error {
	token, err := utils.GenerateToken(16)
	if err != nil {
		return err
	}
	account.Token = token

	_, err = s.db.Exec("INSERT INTO accounts (token, user_id, account_name, balance) VALUES (?, ?, ?, ?)", account.Token, account.UserID, account.AccountName, account.Balance)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAccountsByUserId(userId int) ([]*types.Account, error) {
	rows, err := s.db.Query("SELECT id, token, user_id, account_name, balance, created_at FROM accounts WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}

	accounts := make([]*types.Account, 0)
	for rows.Next() {
		account, err := scanRowIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanRowIntoAccount(rows *sql.Rows) (*types.Account, error) {
	a := new(types.Account)

	err := rows.Scan(&a.ID, &a.Token, &a.UserID, &a.AccountName, &a.Balance, &a.CreatedAt)

	if err != nil {
		return nil, err
	}

	return a, nil
}
