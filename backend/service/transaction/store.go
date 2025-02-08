package transaction

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas-remigio/wallet-tracker/service/account"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreateTransaction(transaction *types.Transaction) error {
	catStore := category.NewStore(s.db)
	category, err := catStore.GetCategoryById(transaction.CategoryId)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	// do not allow transfers here, they demand a different logic
	if category.TransactionTypeID == 3 {
		return fmt.Errorf("transfers are not allowed here")
	}

	accStore := account.NewStore(s.db)
	account, err := accStore.GetAccountByToken(transaction.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// category transaction type id == 1 means credit
	// if category.TransactionTypeID == 2 means debit
	// transaction type 3 means transfer
	if category.TransactionTypeID == 2 {
		transaction.Amount = transaction.Amount * -1
	}
	newBalance := account.Balance + transaction.Amount

	date := time.Now().Format("2006-01-02")

	_, err = s.db.Exec(
		"INSERT INTO transactions (account_token, category_id, amount, description, date, balance) VALUES (?, ?, ?, ?, ?, ?)",
		transaction.AccountToken,
		transaction.CategoryId,
		transaction.Amount,
		transaction.Description,
		date,
		newBalance,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// update user account balance
	_, err = s.db.Exec("UPDATE accounts SET balance = ? WHERE token = ?", newBalance, transaction.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	return nil
}

func (s *Store) GetTransactionsByAccountToken(accountToken string) ([]*types.Transaction, error) {
	rows, err := s.db.Query("SELECT id, account_token, category_id, amount, description, date, balance, created_at FROM transactions WHERE account_token = ?", accountToken)
	if err != nil {
		return nil, err
	}

	transactions := make([]*types.Transaction, 0)
	for rows.Next() {
		transaction, err := scanRowIntoTransaction(rows)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func scanRowIntoTransaction(rows *sql.Rows) (*types.Transaction, error) {
	a := new(types.Transaction)

	err := rows.Scan(&a.ID, &a.AccountToken, &a.CategoryId, &a.Amount, &a.Description, &a.Date, &a.Balance, &a.CreatedAt)

	if err != nil {
		return nil, err
	}

	return a, nil
}
