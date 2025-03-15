package account

import (
	"database/sql"
	"fmt"

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

func (s *Store) GetAccountByToken(token string) (*types.Account, error) {
	row := s.db.QueryRow("SELECT id, token, user_id, account_name, balance, created_at FROM accounts WHERE token = ?", token)
	account := new(types.Account)
	if err := row.Scan(&account.ID, &account.Token, &account.UserID, &account.AccountName, &account.Balance, &account.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}
	return account, nil
}

func (s *Store) GetAccountById(id int) (*types.Account, error) {
	row := s.db.QueryRow("SELECT id, token, user_id, account_name, balance, created_at FROM accounts WHERE id = ?", id)
	account := new(types.Account)
	if err := row.Scan(&account.ID, &account.Token, &account.UserID, &account.AccountName, &account.Balance, &account.CreatedAt); err != nil {
		return nil, err
	}
	return account, nil
}

func scanRowIntoAccount(rows *sql.Rows) (*types.Account, error) {
	a := new(types.Account)

	err := rows.Scan(&a.ID, &a.Token, &a.UserID, &a.AccountName, &a.Balance, &a.CreatedAt)

	if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) UpdateAccount(account *types.Account) error {
	// first get the current account so that we can check if the user is the owner of the account
	currentAccount, err := s.GetAccountById(account.ID)
	if err != nil {
		return err
	}

	if currentAccount.UserID != account.UserID {
		return fmt.Errorf("user does not have permission to update this account")
	}

	_, err = s.db.Exec("UPDATE accounts SET account_name = ?, balance = ? WHERE id = ?", account.AccountName, account.Balance, account.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteAccount(token string, userId int) error {
	// first get the account so that we can check if the user is the owner of the account
	account, err := s.GetAccountByToken(token)
	if err != nil {
		return err
	}

	if account.UserID != userId {
		return fmt.Errorf("user does not have permission to delete this account")
	}

	// delete all transactions associated with the account
	_, err = s.db.Exec("DELETE FROM transactions WHERE account_token = ?", token)
	if err != nil {
		return err
	}

	// delete the account
	_, err = s.db.Exec("DELETE FROM accounts WHERE token = ? AND user_id = ?", token, userId)
	if err != nil {
		return err
	}

	return nil
}
