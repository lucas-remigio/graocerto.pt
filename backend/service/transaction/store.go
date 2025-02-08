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

func (s *Store) GetTransactionsDTOByAccountToken(accountToken string) ([]*types.TransactionDTO, error) {

	rows, err := s.db.Query("SELECT "+
		"t.id, t.account_token, t.amount, t.description, t.date, t.balance, t.created_at, "+
		"c.id, c.category_name, c.color, c.created_at, c.updated_at, "+
		"tt.id, tt.type_name, tt.type_slug "+
		"FROM transactions t "+
		"JOIN categories c ON t.category_id = c.id "+
		"JOIN transaction_types tt ON c.transaction_type_id = tt.id "+
		"WHERE t.account_token = ?", accountToken)

	if err != nil {
		return nil, err
	}

	transactions := make([]*types.TransactionDTO, 0)
	for rows.Next() {
		transaction, err := scanRowIntoTransactionDTO(rows)
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

func scanRowIntoTransactionDTO(rows *sql.Rows) (*types.TransactionDTO, error) {
	a := new(types.TransactionDTO)

	// Initialize nested structs so they're not nil
	a.Category = &types.CategoryDTO{}
	a.Category.TransactionType = &types.TransactionType{}

	err := rows.Scan(
		&a.ID, &a.AccountToken, &a.Amount, &a.Description, &a.Date, &a.Balance, &a.CreatedAt,
		&a.Category.ID, &a.Category.CategoryName, &a.Category.Color, &a.Category.CreatedAt, &a.Category.UpdatedAt,
		&a.Category.TransactionType.ID, &a.Category.TransactionType.TypeName, &a.Category.TransactionType.TypeSlug,
	)

	if err != nil {
		return nil, err
	}

	return a, nil
}
