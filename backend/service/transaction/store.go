package transaction

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type Store struct {
	db           *sql.DB
	accountStore types.AccountStore
}

func NewStore(db *sql.DB, accountStore types.AccountStore) *Store {
	return &Store{
		db:           db,
		accountStore: accountStore,
	}
}

// Scanner functions for use with db utilities
func scanTransaction(rows *sql.Rows) (*types.Transaction, error) {
	t := new(types.Transaction)
	err := rows.Scan(&t.ID, &t.AccountToken, &t.CategoryId, &t.Amount, &t.Description, &t.Date, &t.Balance, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func scanTransactionRow(row *sql.Row) (*types.Transaction, error) {
	t := new(types.Transaction)
	err := row.Scan(&t.ID, &t.AccountToken, &t.CategoryId, &t.Amount, &t.Description, &t.Date, &t.Balance, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func scanTransactionDTO(rows *sql.Rows) (*types.TransactionDTO, error) {
	t := new(types.TransactionDTO)
	t.Category = &types.CategoryDTO{}
	t.Category.TransactionType = &types.TransactionType{}

	err := rows.Scan(
		&t.ID, &t.AccountToken, &t.Amount, &t.Description, &t.Date, &t.Balance, &t.CreatedAt,
		&t.Category.ID, &t.Category.CategoryName, &t.Category.Color, &t.Category.CreatedAt, &t.Category.UpdatedAt,
		&t.Category.TransactionType.ID, &t.Category.TransactionType.TypeName, &t.Category.TransactionType.TypeSlug,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
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

	account, err := s.accountStore.GetAccountByToken(transaction.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// category transaction type id == 1 means credit
	// if category.TransactionTypeID == 2 means debit
	amount := transaction.Amount
	if category.TransactionTypeID == 2 {
		amount = amount * -1
	}
	newBalance := account.Balance + amount

	err = db.ExecWithValidation(s.db,
		"INSERT INTO transactions (account_token, category_id, amount, description, date, balance) VALUES (?, ?, ?, ?, ?, ?)",
		transaction.AccountToken,
		transaction.CategoryId,
		transaction.Amount,
		transaction.Description,
		transaction.Date,
		newBalance,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// update user account balance
	err = db.ExecWithValidation(s.db, "UPDATE accounts SET balance = ? WHERE token = ?", newBalance, transaction.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	return nil
}

func (s *Store) GetTransactionsByAccountToken(accountToken string) ([]*types.Transaction, error) {
	query := "SELECT id, account_token, category_id, amount, description, date, balance, created_at FROM transactions WHERE account_token = ?"
	return db.QueryList(s.db, query, scanTransaction, accountToken)
}

func (s *Store) GetTransactionsByAccountTokenAndMonth(accountToken string, month, year int) ([]*types.Transaction, error) {
	query := `
        SELECT id, account_token, category_id, amount, description, date, balance, created_at 
        FROM transactions 
        WHERE account_token = ? 
        AND MONTH(date) = ? 
        AND YEAR(date) = ? 
        ORDER BY date DESC, id DESC
    `
	return db.QueryList(s.db, query, scanTransaction, accountToken, month, year)
}

func (s *Store) GetTransactionsDTOByAccountToken(accountToken string) ([]*types.TransactionDTO, error) {
	query := "SELECT " +
		"t.id, t.account_token, t.amount, t.description, t.date, t.balance, t.created_at, " +
		"c.id, c.category_name, c.color, c.created_at, c.updated_at, " +
		"tt.id, tt.type_name, tt.type_slug " +
		"FROM transactions t " +
		"JOIN categories c ON t.category_id = c.id " +
		"JOIN transaction_types tt ON c.transaction_type_id = tt.id " +
		"WHERE t.account_token = ? " +
		"ORDER BY t.date DESC, t.id DESC"

	return db.QueryList(s.db, query, scanTransactionDTO, accountToken)
}

func (s *Store) GetTransactionById(id int) (*types.Transaction, error) {
	query := "SELECT id, account_token, category_id, amount, description, date, balance, created_at FROM transactions WHERE id = ?"
	return db.QuerySingle(s.db, query, scanTransactionRow, id)
}

func (s *Store) UpdateTransaction(transaction *types.UpdateTransactionPayload) error {
	// get the current transaction before the update
	tx, err := s.GetTransactionById(transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// get the account
	account, err := s.accountStore.GetAccountByToken(tx.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// there are a lot of things that can happen here
	// most simple case: from credit to credit. if it was 100 and now is 130, we add 30 to the balance
	// if it was debit to debit, if it was 100 and now is 70, we add 30 to the balance
	// if it was credit to debit, if it was 100 and now is 70, we subtract 30 from the balance
	// if it was debit to credit, if it was 100 and now is 130, we subtract 30 from the balance

	// get the current category
	catStore := category.NewStore(s.db)
	currentCategory, err := catStore.GetCategoryById(tx.CategoryId)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	// get the new category
	newCategory, err := catStore.GetCategoryById(transaction.CategoryID)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	// get the current balance
	currentBalance := account.Balance

	// get the new balance
	amount := transaction.Amount
	if currentCategory.TransactionTypeID == (int)(types.DebitTransactionType) {
		amount = amount * -1
	}
	amountDifference := amount - tx.Amount
	var newBalance float64 = currentBalance

	// Helper function to check transaction type
	isTransactionType := func(categoryTypeID int, expectedType types.TransactionTypeID) bool {
		return categoryTypeID == int(expectedType)
	}

	// Both transactions are of the same type (both credit or both debit)
	if isTransactionType(currentCategory.TransactionTypeID, types.CreditTransactionType) ==
		isTransactionType(newCategory.TransactionTypeID, types.CreditTransactionType) {
		newBalance = currentBalance + amountDifference
	} else {
		// Transaction types are different (switching between credit and debit)
		newBalance = currentBalance - amountDifference
	}

	parsedDate, err := time.Parse(time.RFC3339, transaction.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	err = db.ExecWithValidation(s.db, "UPDATE transactions SET amount = ?, category_id = ?, description = ?, date = ? WHERE id = ?",
		transaction.Amount,
		transaction.CategoryID,
		transaction.Description,
		parsedDate.Format("2006-01-02"),
		transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// update the account balance
	err = db.ExecWithValidation(s.db, "UPDATE accounts SET balance = ? WHERE token = ?", newBalance, tx.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	return nil
}

func (s *Store) DeleteTransaction(transactionId int, userId int) error {
	// get the transaction
	tx, err := s.GetTransactionById(transactionId)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// get the account
	account, err := s.accountStore.GetAccountByToken(tx.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// check if the user is the owner of the account
	if err := db.ValidateOwnership(userId, account.UserID, "transaction"); err != nil {
		return err
	}

	// get the transaction category
	catStore := category.NewStore(s.db)
	category, err := catStore.GetCategoryById(tx.CategoryId)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	// check if the category is a transfer
	// if the transaction was a credit, we must remove that amount
	amount := tx.Amount
	if category.TransactionTypeID == int(types.CreditTransactionType) {
		amount = amount * -1
	}

	// get the current balance
	currentBalance := account.Balance

	// get the new balance
	newBalance := currentBalance + amount

	err = db.ExecWithValidation(s.db, "DELETE FROM transactions WHERE id = ?", transactionId)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// update the account balance
	err = db.ExecWithValidation(s.db, "UPDATE accounts SET balance = ? WHERE token = ?", newBalance, tx.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	return nil
}
