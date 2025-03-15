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
	amount := transaction.Amount
	if category.TransactionTypeID == 2 {
		amount = amount * -1
	}
	newBalance := account.Balance + amount

	_, err = s.db.Exec(
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
		"WHERE t.account_token = ? "+
		"ORDER BY t.date DESC, t.id DESC", accountToken)

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

func (s *Store) GetTransactionById(id int) (*types.Transaction, error) {
	rows, err := s.db.Query("SELECT id, account_token, category_id, amount, description, date, balance, created_at FROM transactions WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("transaction not found")
	}

	transaction, err := scanRowIntoTransaction(rows)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *Store) UpdateTransaction(transaction *types.UpdateTransactionPayload) error {
	// get the current transaction before the update
	tx, err := s.GetTransactionById(transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// get the account
	accStore := account.NewStore(s.db)
	account, err := accStore.GetAccountByToken(tx.AccountToken)
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

	_, err = s.db.Exec("UPDATE transactions SET amount = ?, category_id = ?, description = ?, date = ? WHERE id = ?",
		transaction.Amount,
		transaction.CategoryID,
		transaction.Description,
		parsedDate.Format("2006-01-02"),
		transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// update the account balance
	_, err = s.db.Exec("UPDATE accounts SET balance = ? WHERE token = ?", newBalance, tx.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	return nil
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

func (s *Store) DeleteTransaction(transactionId int, userId int) error {
	// get the transaction
	tx, err := s.GetTransactionById(transactionId)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// get the account
	accStore := account.NewStore(s.db)
	account, err := accStore.GetAccountByToken(tx.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// check if the user is the owner of the account
	if account.UserID != userId {
		return fmt.Errorf("user does not have permission to delete this transaction")
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

	_, err = s.db.Exec("DELETE FROM transactions WHERE id = ?", transactionId)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// update the account balance
	_, err = s.db.Exec("UPDATE accounts SET balance = ? WHERE token = ?", newBalance, tx.AccountToken)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	return nil
}
