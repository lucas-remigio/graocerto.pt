package transaction

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
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

func (s *Store) getAccountBalances(accountToken string, userID int) (balance float64, pendingBalance float64, err error) {
	account, err := s.accountStore.GetAccountByToken(accountToken, userID)
	if err != nil {
		return 0, 0, err
	}

	return account.Balance, account.PendingBalance, nil
}

// Scanner functions for use with db utilities
func scanTransaction(rows *sql.Rows) (*types.Transaction, error) {
	t := new(types.Transaction)
	err := rows.Scan(
		&t.ID,
		&t.AccountToken,
		&t.TransactionTypeId,
		&t.CategoryId,
		&t.Amount,
		&t.Description,
		&t.Date,
		&t.Balance,
		&t.IsPending,
		&t.TransferGroupId,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func scanTransactionRow(row *sql.Row) (*types.Transaction, error) {
	t := new(types.Transaction)
	err := row.Scan(
		&t.ID,
		&t.AccountToken,
		&t.TransactionTypeId,
		&t.CategoryId,
		&t.Amount,
		&t.Description,
		&t.Date,
		&t.Balance,
		&t.IsPending,
		&t.TransferGroupId,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanTransactionDTOFromScanner(s scanner) (*types.TransactionDTO, error) {
	t := new(types.TransactionDTO)
	t.Category = &types.CategoryDTO{}
	t.Category.TransactionType = &types.TransactionType{}
	t.TransactionType = &types.TransactionType{}

	// Nullable parent category fields
	var parentCategoryID sql.NullInt64
	var parentCategoryName sql.NullString
	var parentCategoryColor sql.NullString

	err := s.Scan(
		&t.ID,
		&t.AccountToken,
		&t.Amount,
		&t.Description,
		&t.Date,
		&t.Balance,
		&t.IsPending,
		&t.TransferGroupId,
		&t.CreatedAt,
		&t.Category.ID,
		&t.Category.ParentCategoryId,
		&t.Category.CategoryName,
		&t.Category.Color,
		&t.Category.CreatedAt,
		&t.Category.UpdatedAt,
		&t.Category.Budget,
		&t.Category.TransactionType.ID,
		&t.Category.TransactionType.TypeName,
		&t.Category.TransactionType.TypeSlug,
		&t.TransactionType.ID,
		&t.TransactionType.TypeName,
		&t.TransactionType.TypeSlug,
		&parentCategoryID,
		&parentCategoryName,
		&parentCategoryColor,
	)
	if err != nil {
		return nil, err
	}

	// Populate parent category if it exists (minimal DTO with just needed fields)
	if parentCategoryID.Valid {
		t.Category.ParentCategory = &types.CategoryDTO{
			ID:           int(parentCategoryID.Int64),
			CategoryName: parentCategoryName.String,
			Color:        parentCategoryColor.String,
		}
	}

	return t, nil
}

// For *sql.Rows
func scanTransactionsDTOs(rows *sql.Rows) (*types.TransactionDTO, error) {
	return scanTransactionDTOFromScanner(rows)
}

// For *sql.Row
func scanTransactionDTO(row *sql.Row) (*types.TransactionDTO, error) {
	return scanTransactionDTOFromScanner(row)
}

func (s *Store) CreateTransaction(transaction *types.Transaction, userId int) (*types.Transaction, error) {
	catStore := category.NewStore(s.db)
	category, err := catStore.GetCategoryById(transaction.CategoryId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Set transaction type from category
	transaction.TransactionTypeId = category.TransactionTypeId

	// do not allow transfers here, they demand a different logic
	if transaction.TransactionTypeId == int(types.TransferTransactionType) {
		return nil, fmt.Errorf("transfers are not allowed here, use /transactions/transfer endpoint")
	}

	account, err := s.accountStore.GetAccountByToken(transaction.AccountToken, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// check if the user is the owner of the account
	if err := db.ValidateOwnership(account.UserID, userId, "account"); err != nil {
		return nil, err
	}

	// category transaction type id == 1 means credit
	// if category.TransactionTypeID == 2 means debit
	amount := transaction.Amount
	if category.TransactionTypeId == (int)(types.DebitTransactionType) {
		amount = amount * -1
	}
	newBalance := account.Balance + amount

	var insertedId int
	err = s.db.QueryRow(
		`INSERT INTO transactions 
        (account_token, transaction_type_id, category_id, amount, description, date, balance, transfer_group_id, is_pending) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false) RETURNING id`,
		transaction.AccountToken,
		transaction.TransactionTypeId,
		transaction.CategoryId,
		transaction.Amount,
		transaction.Description,
		transaction.Date,
		newBalance,
		transaction.TransferGroupId,
	).Scan(&insertedId)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// update user account balance
	_, err = db.ExecWithValidation(s.db, "UPDATE accounts SET balance = $1 WHERE token = $2", newBalance, transaction.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	slog.Info("transaction created", "id", insertedId, "account_token", transaction.AccountToken, "amount", transaction.Amount)

	transaction.ID = insertedId
	transaction.Balance = newBalance

	return transaction, nil
}

func (s *Store) CreateTransactionAndReturn(transaction *types.Transaction, userId int) (*types.TransactionChangeResponse, error) {
	createdTransaction, err := s.CreateTransaction(transaction, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	createdDTO, err := s.GetTransactionDTOById(createdTransaction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created transaction DTO: %w", err)
	}

	// Get available months for the account token
	availableMonths, err := s.GetAvailableTransactionMonthsByAccountToken(createdTransaction.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get available months: %w", err)
	}

	return &types.TransactionChangeResponse{
		Transaction:           createdDTO,
		AccountBalance:        &createdDTO.Balance,
		AccountPendingBalance: &createdDTO.Balance,
		Months:                availableMonths,
	}, nil
}

func (s *Store) CreateTransfer(payload *types.CreateTransferPayload, userId int) (*types.TransferResponse, error) {
	catStore := category.NewStore(s.db)

	// Verify debit category exists and is a debit type
	debitCategory, err := catStore.GetCategoryById(payload.DebitCategoryID, userId)
	if err != nil {
		return nil, fmt.Errorf("invalid debit category: %w", err)
	}
	if debitCategory.TransactionTypeId != int(types.DebitTransactionType) {
		return nil, fmt.Errorf("debit category must be a debit type")
	}

	// Verify credit category exists and is a credit type
	creditCategory, err := catStore.GetCategoryById(payload.CreditCategoryID, userId)
	if err != nil {
		return nil, fmt.Errorf("invalid credit category: %w", err)
	}
	if creditCategory.TransactionTypeId != int(types.CreditTransactionType) {
		return nil, fmt.Errorf("credit category must be a credit type")
	}

	// Verify both accounts exist and belong to user
	sourceAccount, err := s.accountStore.GetAccountByToken(payload.SourceAccountToken, userId)
	if err != nil {
		return nil, fmt.Errorf("invalid source account: %w", err)
	}

	destAccount, err := s.accountStore.GetAccountByToken(payload.DestinationAccountToken, userId)
	if err != nil {
		return nil, fmt.Errorf("invalid destination account: %w", err)
	}

	// Verify ownership
	if err := db.ValidateOwnership(sourceAccount.UserID, userId, "source account"); err != nil {
		return nil, err
	}
	if err := db.ValidateOwnership(destAccount.UserID, userId, "destination account"); err != nil {
		return nil, err
	}

	// Generate unique transfer group ID
	transferGroupID, err := utils.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate transfer group ID: %w", err)
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create DEBIT transaction on source account (money going out)
	debitBalance := sourceAccount.Balance - payload.Amount
	var debitTxId int
	err = tx.QueryRow(
		`INSERT INTO transactions 
        (account_token, transaction_type_id, category_id, amount, description, date, balance, transfer_group_id, is_pending) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false) RETURNING id`,
		payload.SourceAccountToken,
		int(types.DebitTransactionType),
		payload.DebitCategoryID, // Use debit category
		payload.Amount,
		payload.Description,
		payload.Date,
		debitBalance,
		transferGroupID,
	).Scan(&debitTxId)
	if err != nil {
		return nil, fmt.Errorf("failed to create debit transaction: %w", err)
	}

	// Update source account balance
	_, err = tx.Exec("UPDATE accounts SET balance = $1 WHERE token = $2", debitBalance, payload.SourceAccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update source account balance: %w", err)
	}

	// Create CREDIT transaction on destination account (money coming in)
	creditBalance := destAccount.Balance + payload.Amount
	var creditTxId int
	err = tx.QueryRow(
		`INSERT INTO transactions 
        (account_token, transaction_type_id, category_id, amount, description, date, balance, transfer_group_id, is_pending) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false) RETURNING id`,
		payload.DestinationAccountToken,
		int(types.CreditTransactionType),
		payload.CreditCategoryID, // Use credit category
		payload.Amount,
		payload.Description,
		payload.Date,
		creditBalance,
		transferGroupID,
	).Scan(&creditTxId)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit transaction: %w", err)
	}

	// Update destination account balance
	_, err = tx.Exec("UPDATE accounts SET balance = $1 WHERE token = $2", creditBalance, payload.DestinationAccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update destination account balance: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transfer: %w", err)
	}

	slog.Info("transfer completed", "group_id", transferGroupID, "source", payload.SourceAccountToken, "dest", payload.DestinationAccountToken, "amount", payload.Amount)

	// Get the created transaction DTOs
	debitTxDTO, err := s.GetTransactionDTOById(debitTxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get debit transaction: %w", err)
	}

	creditTxDTO, err := s.GetTransactionDTOById(creditTxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit transaction: %w", err)
	}

	// Get available months for source account
	sourceMonths, err := s.GetAvailableTransactionMonthsByAccountToken(payload.SourceAccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get source account months: %w", err)
	}

	// Get available months for destination account
	destMonths, err := s.GetAvailableTransactionMonthsByAccountToken(payload.DestinationAccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination account months: %w", err)
	}

	return &types.TransferResponse{
		TransferGroupID:           transferGroupID,
		DebitTransaction:          debitTxDTO,
		CreditTransaction:         creditTxDTO,
		SourceAccountBalance:      debitBalance,
		DestinationAccountBalance: creditBalance,
		SourceAccountMonths:       sourceMonths,
		DestinationAccountMonths:  destMonths,
	}, nil
}

func (s *Store) GetTransactionsByAccountToken(accountToken string, month, year *int) ([]*types.Transaction, error) {
	var query string
	var args []interface{}

	baseQuery := `
        SELECT id, account_token, transaction_type_id, category_id, amount, description, date, balance, is_pending, transfer_group_id, created_at 
        FROM transactions 
        WHERE account_token = $1`

	args = append(args, accountToken)

	if month != nil && year != nil {
		query = baseQuery + " AND EXTRACT(MONTH FROM date) = $2 AND EXTRACT(YEAR FROM date) = $3" +
			" ORDER BY date DESC, id DESC"
		args = append(args, *month, *year)
	} else {
		query = baseQuery + " ORDER BY date DESC, id DESC"
	}

	return db.QueryList(s.db, query, scanTransaction, args...)
}

func (s *Store) GetTransactionsDTOByAccountToken(accountToken string, month, year *int) ([]*types.TransactionDTO, error) {
	var query string
	var args []interface{}

	baseQuery := "SELECT " +
		"t.id, t.account_token, t.amount, t.description, t.date, t.balance, t.is_pending, t.transfer_group_id, t.created_at, " +
		"c.id, c.parent_category_id, c.category_name, c.color, c.created_at, c.updated_at, c.budget, " +
		"c_tt.id, c_tt.type_name, c_tt.type_slug, " +
		"tt.id, tt.type_name, tt.type_slug, " +
		"pc.id, pc.category_name, pc.color " + // Parent category fields
		"FROM transactions t " +
		"JOIN transaction_types tt ON t.transaction_type_id = tt.id " +
		"JOIN categories c ON t.category_id = c.id " +
		"JOIN transaction_types c_tt ON c.transaction_type_id = c_tt.id " +
		"LEFT JOIN categories pc ON c.parent_category_id = pc.id " + // LEFT JOIN for parent
		"WHERE t.account_token = $1 "

	args = append(args, accountToken)

	if month != nil && year != nil {
		query = baseQuery + "AND EXTRACT(MONTH FROM t.date) = $2 AND EXTRACT(YEAR FROM t.date) = $3 " +
			"ORDER BY t.date DESC, t.id DESC"
		args = append(args, *month, *year)
	} else {
		query = baseQuery + "ORDER BY t.date DESC, t.id DESC"
	}

	return db.QueryList(s.db, query, scanTransactionsDTOs, args...)
}

func (s *Store) GetTransactionDTOById(id int) (*types.TransactionDTO, error) {
	query := `
        SELECT 
            t.id, t.account_token, t.amount, t.description, t.date, t.balance, t.is_pending, t.transfer_group_id, t.created_at,
            c.id, c.parent_category_id, c.category_name, c.color, c.created_at, c.updated_at, c.budget,
            c_tt.id, c_tt.type_name, c_tt.type_slug,
            tt.id, tt.type_name, tt.type_slug,
            pc.id, pc.category_name, pc.color
        FROM transactions t
        JOIN transaction_types tt ON t.transaction_type_id = tt.id
        JOIN categories c ON t.category_id = c.id
        JOIN transaction_types c_tt ON c.transaction_type_id = c_tt.id
        LEFT JOIN categories pc ON c.parent_category_id = pc.id
        WHERE t.id = $1`

	return db.QuerySingle(s.db, query, scanTransactionDTO, id)
}

func (s *Store) GetTransactionById(id int) (*types.Transaction, error) {
	query := "SELECT id, account_token, transaction_type_id, category_id, amount, description, date, balance, is_pending, transfer_group_id, created_at FROM transactions WHERE id = $1"
	return db.QuerySingle(s.db, query, scanTransactionRow, id)
}

func (s *Store) UpdateTransaction(transaction *types.UpdateTransactionPayload, userId int) (*types.Transaction, error) {
	// get the current transaction before the update
	tx, err := s.GetTransactionById(transaction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// get the account
	account, err := s.accountStore.GetAccountByToken(tx.AccountToken, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// check if the user is the owner of the account
	if err := db.ValidateOwnership(account.UserID, userId, "account"); err != nil {
		return nil, err
	}

	if tx.IsPending {
		return nil, fmt.Errorf("cannot edit a pending transaction")
	}

	// there are a lot of things that can happen here
	// most simple case: from credit to credit. if it was 100 and now is 130, we add 30 to the balance
	// if it was debit to debit, if it was 100 and now is 70, we add 30 to the balance
	// if it was credit to debit, if it was 100 and now is 70, we subtract 30 from the balance
	// if it was debit to credit, if it was 100 and now is 130, we subtract 30 from the balance

	// For now, we cannot change the transaction type. If it is debit, it will remain a debit.
	// We only need to calculate the difference in amount and update the balance accordingly.

	// get the current category
	catStore := category.NewStore(s.db)
	currentCategory, err := catStore.GetCategoryById(tx.CategoryId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous category: %w", err)
	}

	newCategory, err := catStore.GetCategoryById(transaction.CategoryID, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get new category: %w", err)
	}

	// Update transaction type based on new category
	newTransactionTypeId := newCategory.TransactionTypeId

	// get the current balance
	currentBalance := account.Balance

	// get the new balance

	// So for a credit, if the user had 200 registered and now is 300, we add 100 to the balance
	// If the user has 200 registered and now is 100, we subtract 100 from the balance
	// For a debit, if the user had 200 registered and now is 100, we add 100 to the balance
	// If the user has 200 registered and now is 300, we subtract 100
	// Having in mind, in the database, the amount is always positive
	// Get the current transaction amount (positive for both credit and debit)
	currentAmount := tx.Amount
	if currentCategory.TransactionTypeId == (int)(types.DebitTransactionType) {
		currentAmount = currentAmount * -1 // Negate for debit
	}

	// Get the new transaction amount (positive for both credit and debit)
	newAmount := transaction.Amount
	if newCategory.TransactionTypeId == (int)(types.DebitTransactionType) {
		newAmount = newAmount * -1 // Negate for debit
	}

	// Calculate the new balance
	amountDifference := newAmount - currentAmount
	newBalance := currentBalance + amountDifference

	_, err = db.ExecWithValidation(s.db,
		"UPDATE transactions SET transaction_type_id = $1, amount = $2, category_id = $3, description = $4, date = $5, balance = $6 WHERE id = $7",
		newTransactionTypeId,
		transaction.Amount,
		transaction.CategoryID,
		transaction.Description,
		transaction.Date,
		newBalance,
		transaction.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// update the account balance
	_, err = db.ExecWithValidation(s.db, "UPDATE accounts SET balance = $1 WHERE token = $2", newBalance, tx.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	// Get the updated transaction
	updatedTransaction := &types.Transaction{
		ID:                tx.ID,
		AccountToken:      tx.AccountToken,
		TransactionTypeId: newTransactionTypeId,
		CategoryId:        transaction.CategoryID,
		Amount:            transaction.Amount,
		Description:       transaction.Description,
		Date:              transaction.Date,
		Balance:           newBalance,
		TransferGroupId:   tx.TransferGroupId,
		CreatedAt:         tx.CreatedAt,
	}

	return updatedTransaction, nil
}

func (s *Store) UpdateTransactionAndReturn(payload *types.UpdateTransactionPayload, userId int) (*types.TransactionChangeResponse, error) {
	updatedTx, err := s.UpdateTransaction(payload, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	transactionDTO, err := s.GetTransactionDTOById(updatedTx.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated transaction DTO: %w", err)
	}

	// Get available months for the account token
	availableMonths, err := s.GetAvailableTransactionMonthsByAccountToken(transactionDTO.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get available months: %w", err)
	}

	return &types.TransactionChangeResponse{
		Transaction:           transactionDTO,
		AccountBalance:        &transactionDTO.Balance,
		AccountPendingBalance: &transactionDTO.Balance,
		Months:                availableMonths,
	}, nil
}

func (s *Store) DeleteTransaction(transactionId int, userId int) (balance *float64, err error) {
	// get the transaction
	tx, err := s.GetTransactionById(transactionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// get the account
	account, err := s.accountStore.GetAccountByToken(tx.AccountToken, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// check if the user is the owner of the account
	if err := db.ValidateOwnership(userId, account.UserID, "transaction"); err != nil {
		return nil, err
	}

	if tx.IsPending {
		// Pending transfers are two linked rows; drop both so no orphan side remains.
		// Pending rows are not yet applied to balances, so no balance adjustment is needed.
		if tx.TransferGroupId != nil && *tx.TransferGroupId != "" {
			_, err = db.ExecWithValidation(s.db,
				`DELETE FROM transactions
				 WHERE transfer_group_id = $1
				   AND account_token IN (SELECT token FROM accounts WHERE user_id = $2)`,
				*tx.TransferGroupId, userId)
			if err != nil {
				return nil, fmt.Errorf("failed to delete pending transfer: %w", err)
			}
			return &account.Balance, nil
		}
		_, err = db.ExecWithValidation(s.db, "DELETE FROM transactions WHERE id = $1", tx.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete pending transaction: %w", err)
		}
		return &account.Balance, nil
	}

	// Check if this is a transfer transaction
	if tx.TransferGroupId != nil && *tx.TransferGroupId != "" {
		// This is a transfer - need to delete both transactions
		result, err := s.deleteTransferPair(*tx.TransferGroupId, userId)
		if err != nil {
			return nil, err
		}
		return &result.primaryBalance, nil
	}

	// Not a transfer - delete single transaction normally
	return s.deleteSingleTransaction(tx, account, userId)
}

func (s *Store) deleteSingleTransaction(tx *types.Transaction, account *types.Account, userId int) (*float64, error) {
	// get the transaction category
	catStore := category.NewStore(s.db)
	category, err := catStore.GetCategoryById(tx.CategoryId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// if the transaction was a credit, we must remove that amount
	amount := tx.Amount
	if category.TransactionTypeId == int(types.CreditTransactionType) {
		amount = amount * -1
	}

	// get the current balance
	currentBalance := account.Balance

	// get the new balance
	newBalance := currentBalance + amount

	_, err = db.ExecWithValidation(s.db, "DELETE FROM transactions WHERE id = $1", tx.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete transaction: %w", err)
	}

	// update the account balance
	_, err = db.ExecWithValidation(s.db, "UPDATE accounts SET balance = $1 WHERE token = $2", newBalance, tx.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	return &newBalance, nil
}

type deleteTransferResult struct {
	primaryAccountToken   string
	primaryBalance        float64
	secondaryAccountToken string
	secondaryBalance      float64
}

func (s *Store) deleteTransferPair(transferGroupID string, userId int) (*deleteTransferResult, error) {
	// Get both transactions in the transfer
	query := `
        SELECT t.id, t.account_token, t.transaction_type_id, t.category_id, t.amount, t.description, 
               t.date, t.balance, t.is_pending, t.transfer_group_id, t.created_at
        FROM transactions t
        INNER JOIN accounts a ON t.account_token = a.token
        WHERE t.transfer_group_id = $1 AND a.user_id = $2
    `

	rows, err := s.db.Query(query, transferGroupID, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*types.Transaction
	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	if len(transactions) != 2 {
		return nil, fmt.Errorf("invalid transfer: expected 2 transactions, found %d", len(transactions))
	}

	// Start database transaction for atomic deletion
	dbTx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	result := &deleteTransferResult{}

	// Delete both transactions and update both account balances
	for i, tx := range transactions {
		// Get the account
		account, err := s.accountStore.GetAccountByToken(tx.AccountToken, userId)
		if err != nil {
			return nil, fmt.Errorf("failed to get account: %w", err)
		}

		// Calculate balance adjustment
		amount := tx.Amount
		if tx.TransactionTypeId == int(types.CreditTransactionType) {
			amount = amount * -1
		}

		newBalance := account.Balance + amount

		// Delete transaction
		_, err = dbTx.Exec("DELETE FROM transactions WHERE id = $1", tx.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete transaction %d: %w", tx.ID, err)
		}

		// Update account balance
		_, err = dbTx.Exec("UPDATE accounts SET balance = $1 WHERE token = $2", newBalance, tx.AccountToken)
		if err != nil {
			return nil, fmt.Errorf("failed to update account balance: %w", err)
		}

		// Store both account info
		if i == 0 {
			result.primaryAccountToken = tx.AccountToken
			result.primaryBalance = newBalance
		} else {
			result.secondaryAccountToken = tx.AccountToken
			result.secondaryBalance = newBalance
		}
	}

	if err := dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func (s *Store) DeleteTransactionAndReturn(transactionId int, userId int) (*types.TransactionChangeResponse, error) {
	transactionDTO, err := s.GetTransactionDTOById(transactionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction DTO: %w", err)
	}

	// Pending rows are not yet applied to balances, so deleting one is a rejection:
	// route it through the pending path, which drops both sides of a transfer
	// without the confirmed-transfer balance adjustments.
	if transactionDTO.IsPending {
		return s.RejectPendingTransactionAndReturn(transactionId, userId)
	}

	// Check if this is a transfer
	isTransfer := transactionDTO.TransferGroupId != nil && *transactionDTO.TransferGroupId != ""

	if isTransfer {
		// Handle transfer deletion - returns both accounts' info
		result, err := s.deleteTransferPair(*transactionDTO.TransferGroupId, userId)
		if err != nil {
			return nil, fmt.Errorf("failed to delete transfer: %w", err)
		}

		// Get available months for primary account
		primaryMonths, err := s.GetAvailableTransactionMonthsByAccountToken(result.primaryAccountToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get primary account months: %w", err)
		}

		// Get available months for secondary account
		secondaryMonths, err := s.GetAvailableTransactionMonthsByAccountToken(result.secondaryAccountToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get secondary account months: %w", err)
		}

		// Determine which is the paired account (not the one being deleted)
		var pairedToken string
		var pairedBalance float64
		var pairedPendingBalance float64
		var pairedMonths []*types.MonthYear
		var currentBalance float64
		var currentPendingBalance float64
		var currentMonths []*types.MonthYear

		if transactionDTO.AccountToken == result.primaryAccountToken {
			pairedToken = result.secondaryAccountToken
			pairedBalance = result.secondaryBalance
			pairedMonths = secondaryMonths

			currentBalance = result.primaryBalance
			currentMonths = primaryMonths
		} else {
			pairedToken = result.primaryAccountToken
			pairedBalance = result.primaryBalance
			pairedMonths = primaryMonths

			currentBalance = result.secondaryBalance
			currentMonths = secondaryMonths
		}

		_, currentPendingBalance, err = s.getAccountBalances(transactionDTO.AccountToken, userId)
		if err != nil {
			return nil, fmt.Errorf("failed to get current account pending balance: %w", err)
		}
		_, pairedPendingBalance, err = s.getAccountBalances(pairedToken, userId)
		if err != nil {
			return nil, fmt.Errorf("failed to get paired account pending balance: %w", err)
		}

		return &types.TransactionChangeResponse{
			Transaction:                 transactionDTO,
			AccountBalance:              &currentBalance,
			AccountPendingBalance:       &currentPendingBalance,
			Months:                      currentMonths,
			IsTransfer:                  true,
			PairedAccountToken:          &pairedToken,
			PairedAccountBalance:        &pairedBalance,
			PairedAccountPendingBalance: &pairedPendingBalance,
			PairedAccountMonths:         pairedMonths,
		}, nil
	}

	// Not a transfer - handle normally
	balance, err := s.DeleteTransaction(transactionId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete transaction: %w", err)
	}

	availableMonths, err := s.GetAvailableTransactionMonthsByAccountToken(transactionDTO.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get available months: %w", err)
	}

	_, pendingBalance, err := s.getAccountBalances(transactionDTO.AccountToken, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get account pending balance: %w", err)
	}

	return &types.TransactionChangeResponse{
		Transaction:           transactionDTO,
		AccountBalance:        balance,
		AccountPendingBalance: &pendingBalance,
		Months:                availableMonths,
		IsTransfer:            false,
	}, nil
}

func (s *Store) ApprovePendingTransactionAndReturn(transactionID int, userID int) (*types.TransactionChangeResponse, error) {
	txData, err := s.GetTransactionById(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if !txData.IsPending {
		return nil, fmt.Errorf("transaction is not pending")
	}

	account, err := s.accountStore.GetAccountByToken(txData.AccountToken, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if err := db.ValidateOwnership(account.UserID, userID, "transaction"); err != nil {
		return nil, err
	}

	// Transfers are two linked transactions; approve both sides together.
	if txData.TransferGroupId != nil && *txData.TransferGroupId != "" {
		return s.approvePendingTransferAndReturn(txData, userID)
	}

	catStore := category.NewStore(s.db)
	txCategory, err := catStore.GetCategoryById(txData.CategoryId, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction category: %w", err)
	}

	amount := txData.Amount
	if txCategory.TransactionTypeId == int(types.DebitTransactionType) {
		amount *= -1
	}

	newBalance := account.Balance + amount

	dbTx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin approval transaction: %w", err)
	}
	defer dbTx.Rollback()

	result, err := dbTx.Exec(
		"UPDATE transactions SET is_pending = false, balance = $1 WHERE id = $2 AND is_pending = true",
		newBalance, txData.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to approve pending transaction: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil, fmt.Errorf("transaction is no longer pending")
	}

	_, err = dbTx.Exec("UPDATE accounts SET balance = $1 WHERE token = $2", newBalance, txData.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	if err := dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit pending approval: %w", err)
	}

	dto, err := s.GetTransactionDTOById(txData.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch approved transaction: %w", err)
	}
	availableMonths, err := s.GetAvailableTransactionMonthsByAccountToken(txData.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get available months: %w", err)
	}

	return &types.TransactionChangeResponse{
		Transaction:           dto,
		AccountBalance:        &newBalance,
		AccountPendingBalance: &newBalance,
		Months:                availableMonths,
	}, nil
}

// approvePendingTransferAndReturn approves both pending sides of a transfer
// atomically, applying each leg to its account balance, and returns the change
// for the requested side plus its paired account (mirrors DeleteTransactionAndReturn).
func (s *Store) approvePendingTransferAndReturn(current *types.Transaction, userID int) (*types.TransactionChangeResponse, error) {
	transferGroupID := *current.TransferGroupId

	query := `
        SELECT t.id, t.account_token, t.transaction_type_id, t.category_id, t.amount, t.description,
               t.date, t.balance, t.is_pending, t.transfer_group_id, t.created_at
        FROM transactions t
        INNER JOIN accounts a ON t.account_token = a.token
        WHERE t.transfer_group_id = $1 AND a.user_id = $2
    `
	rows, err := s.db.Query(query, transferGroupID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*types.Transaction
	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}
	if len(transactions) != 2 {
		return nil, fmt.Errorf("invalid transfer: expected 2 transactions, found %d", len(transactions))
	}

	dbTx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin approval transaction: %w", err)
	}
	defer dbTx.Rollback()

	var currentToken, pairedToken string
	var currentBalance, pairedBalance float64

	for _, tx := range transactions {
		if !tx.IsPending {
			return nil, fmt.Errorf("transaction is no longer pending")
		}

		account, err := s.accountStore.GetAccountByToken(tx.AccountToken, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account: %w", err)
		}

		amount := tx.Amount
		if tx.TransactionTypeId == int(types.DebitTransactionType) {
			amount *= -1
		}
		newBalance := account.Balance + amount

		result, err := dbTx.Exec(
			"UPDATE transactions SET is_pending = false, balance = $1 WHERE id = $2 AND is_pending = true",
			newBalance, tx.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to approve pending transaction: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return nil, fmt.Errorf("transaction is no longer pending")
		}

		_, err = dbTx.Exec("UPDATE accounts SET balance = $1 WHERE token = $2", newBalance, tx.AccountToken)
		if err != nil {
			return nil, fmt.Errorf("failed to update account balance: %w", err)
		}

		if tx.ID == current.ID {
			currentToken = tx.AccountToken
			currentBalance = newBalance
		} else {
			pairedToken = tx.AccountToken
			pairedBalance = newBalance
		}
	}

	if err := dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit pending approval: %w", err)
	}

	dto, err := s.GetTransactionDTOById(current.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch approved transaction: %w", err)
	}
	currentMonths, err := s.GetAvailableTransactionMonthsByAccountToken(currentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get available months: %w", err)
	}
	pairedMonths, err := s.GetAvailableTransactionMonthsByAccountToken(pairedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get paired account months: %w", err)
	}

	_, currentPendingBalance, err := s.getAccountBalances(currentToken, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current account pending balance: %w", err)
	}
	_, pairedPendingBalance, err := s.getAccountBalances(pairedToken, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get paired account pending balance: %w", err)
	}

	return &types.TransactionChangeResponse{
		Transaction:                 dto,
		AccountBalance:              &currentBalance,
		AccountPendingBalance:       &currentPendingBalance,
		Months:                      currentMonths,
		IsTransfer:                  true,
		PairedAccountToken:          &pairedToken,
		PairedAccountBalance:        &pairedBalance,
		PairedAccountPendingBalance: &pairedPendingBalance,
		PairedAccountMonths:         pairedMonths,
	}, nil
}

func (s *Store) RejectPendingTransactionAndReturn(transactionID int, userID int) (*types.TransactionChangeResponse, error) {
	dto, err := s.GetTransactionDTOById(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction DTO: %w", err)
	}
	if !dto.IsPending {
		return nil, fmt.Errorf("transaction is not pending")
	}

	// Transfers reject both sides; resolve the paired account before deletion so
	// the response can update it too (only pending balances change, not balances).
	var pairedToken string
	isTransfer := dto.TransferGroupId != nil && *dto.TransferGroupId != ""
	if isTransfer {
		pairedToken, err = s.getPairedAccountToken(*dto.TransferGroupId, dto.AccountToken, userID)
		if err != nil {
			return nil, err
		}
	}

	balance, err := s.DeleteTransaction(transactionID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to reject pending transaction: %w", err)
	}

	availableMonths, err := s.GetAvailableTransactionMonthsByAccountToken(dto.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get available months: %w", err)
	}

	_, pendingBalance, err := s.getAccountBalances(dto.AccountToken, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account pending balance: %w", err)
	}

	response := &types.TransactionChangeResponse{
		Transaction:           dto,
		AccountBalance:        balance,
		AccountPendingBalance: &pendingBalance,
		Months:                availableMonths,
		IsTransfer:            isTransfer,
	}

	if isTransfer && pairedToken != "" {
		pairedBalance, pairedPendingBalance, err := s.getAccountBalances(pairedToken, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get paired account pending balance: %w", err)
		}
		pairedMonths, err := s.GetAvailableTransactionMonthsByAccountToken(pairedToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get paired account months: %w", err)
		}
		response.PairedAccountToken = &pairedToken
		response.PairedAccountBalance = &pairedBalance
		response.PairedAccountPendingBalance = &pairedPendingBalance
		response.PairedAccountMonths = pairedMonths
	}

	return response, nil
}

// getPairedAccountToken returns the other account token in a transfer group for
// the given user (the side that is not currentToken).
func (s *Store) getPairedAccountToken(transferGroupID, currentToken string, userID int) (string, error) {
	query := `
        SELECT t.account_token
        FROM transactions t
        INNER JOIN accounts a ON t.account_token = a.token
        WHERE t.transfer_group_id = $1 AND a.user_id = $2 AND t.account_token <> $3
        LIMIT 1
    `
	var pairedToken string
	err := s.db.QueryRow(query, transferGroupID, userID, currentToken).Scan(&pairedToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to resolve paired account token: %w", err)
	}
	return pairedToken, nil
}

// Store implementation
func (s *Store) GetAvailableTransactionMonthsByAccountToken(accountToken string) ([]*types.MonthYear, error) {
	query := `
        SELECT 
            year,
            month,
            count
        FROM (
            SELECT 
                DATE_PART('year', date)::int as year,
                DATE_PART('month', date)::int as month,
                COUNT(*) as count
            FROM transactions 
            WHERE account_token = $1 
            GROUP BY DATE_PART('year', date), DATE_PART('month', date)
        ) subquery
        ORDER BY year DESC, month DESC
    `

	return db.QueryList(s.db, query, scanMonthYear, accountToken)
}

func scanMonthYear(rows *sql.Rows) (*types.MonthYear, error) {
	m := new(types.MonthYear)
	err := rows.Scan(&m.Year, &m.Month, &m.Count)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Store) CalculateTransactionTotals(transactions []*types.TransactionDTO) (*types.TransactionTotals, error) {
	if transactions == nil {
		return nil, fmt.Errorf("transactions cannot be nil")
	}

	total := &types.TransactionTotals{
		Debit:      0,
		Credit:     0,
		Difference: 0,
	}

	// Early return for empty slice
	if len(transactions) == 0 {
		return total, nil
	}

	for _, tx := range transactions {
		// Skip transactions without category info
		if tx.IsPending || tx.Category == nil || tx.Category.TransactionType == nil {
			continue
		}

		switch tx.Category.TransactionType.ID {
		case int(types.CreditTransactionType):
			total.Credit += tx.Amount
		case int(types.DebitTransactionType):
			total.Debit += tx.Amount
		}
	}

	total.Credit = utils.Round(total.Credit, 2)
	total.Debit = utils.Round(total.Debit, 2)
	difference := total.Credit - total.Debit
	total.Difference = utils.Round(difference, 2)
	return total, nil
}

// Helper function to get absolute value for float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Process transactions and calculate largest amounts and daily breakdowns
func (s *Store) calculateLargestAmountsAndDailyTotals(
	transactions []*types.TransactionDTO,
) (largestCredit, largestDebit float64, dailyTotals map[string]*types.DailyTotal) {
	dailyTotals = make(map[string]*types.DailyTotal)

	for _, tx := range transactions {
		if tx.IsPending || tx.Category == nil {
			continue
		}

		date := tx.Date.Format("2006-01-02")

		// Initialize daily total if it doesn't exist
		if dailyTotals[date] == nil {
			dailyTotals[date] = &types.DailyTotal{
				Date:       date,
				Credit:     0,
				Debit:      0,
				Difference: 0,
			}
		}

		switch tx.Category.TransactionType.ID {
		case int(types.DebitTransactionType):
			dailyTotals[date].Debit += tx.Amount
			dailyTotals[date].Difference -= tx.Amount // Debit reduces the difference
			if absAmount := abs(tx.Amount); absAmount > largestDebit {
				largestDebit = absAmount
			}
		case int(types.CreditTransactionType):
			dailyTotals[date].Credit += tx.Amount
			dailyTotals[date].Difference += tx.Amount // Credit increases the difference
			if tx.Amount > largestCredit {
				largestCredit = tx.Amount
			}
		}
	}

	// Calculate differences and round all values
	for _, daily := range dailyTotals {
		daily.Credit = utils.Round(daily.Credit, 2)
		daily.Debit = utils.Round(daily.Debit, 2)
		daily.Difference = utils.Round(daily.Credit-daily.Debit, 2)
	}

	largestCredit = utils.Round(largestCredit, 2)
	largestDebit = utils.Round(largestDebit, 2)
	return largestCredit, largestDebit, dailyTotals
}

// Build category breakdown maps from transactions
// Build category breakdown maps from transactions
func (s *Store) buildCategoryBreakdowns(transactions []*types.TransactionDTO) (
	creditCategoryMap, debitCategoryMap map[int]*types.CategoryStatistic) {

	creditCategoryMap = make(map[int]*types.CategoryStatistic)
	debitCategoryMap = make(map[int]*types.CategoryStatistic)

	for _, tx := range transactions {
		if tx.IsPending || tx.Category == nil {
			continue
		}

		categoryID := tx.Category.ID
		categoryName := tx.Category.CategoryName
		categoryColor := tx.Category.Color
		categoryBudget := tx.Category.Budget
		parentID := tx.Category.ParentCategoryId

		absAmount := abs(tx.Amount)

		// Process based on transaction type
		switch tx.Category.TransactionType.ID {
		case int(types.CreditTransactionType):
			s.updateCategoryMapByID(creditCategoryMap, categoryID, parentID, categoryName, categoryColor, categoryBudget, absAmount)
		case int(types.DebitTransactionType):
			s.updateCategoryMapByID(debitCategoryMap, categoryID, parentID, categoryName, categoryColor, categoryBudget, absAmount)
		}
	}

	return creditCategoryMap, debitCategoryMap
}

// Helper to update category map using category ID as key
func (s *Store) updateCategoryMapByID(
	categoryMap map[int]*types.CategoryStatistic,
	categoryID int,
	parentID *int,
	categoryName, categoryColor string,
	budget *int,
	amount float64) {

	if _, exists := categoryMap[categoryID]; !exists {
		categoryMap[categoryID] = &types.CategoryStatistic{
			ID:               categoryID,
			ParentID:         parentID,
			Name:             categoryName,
			Count:            0,
			Total:            0,
			Percentage:       0,
			Color:            categoryColor,
			Budget:           budget,
			BudgetPercentage: 0,
			Subcategories:    []types.CategoryStatistic{},
		}
	}
	categoryMap[categoryID].Count++
	categoryMap[categoryID].Total += amount
}

// Calculate percentages, build hierarchy, and convert map to slice
func (s *Store) processCategoryBreakdown(
	categoryMap map[int]*types.CategoryStatistic,
	totalAmount float64) []*types.CategoryStatistic {

	// First pass: calculate percentages for all categories
	for _, categoryStat := range categoryMap {
		if totalAmount > 0 {
			percentage := (categoryStat.Total / totalAmount) * 100
			categoryStat.Percentage = utils.Round(percentage, 2)
		}

		if categoryStat.Budget != nil && *categoryStat.Budget > 0 {
			budgetPct := (categoryStat.Total / float64(*categoryStat.Budget)) * 100
			categoryStat.BudgetPercentage = utils.Round(budgetPct, 2)
		} else {
			categoryStat.BudgetPercentage = 0
		}

		categoryStat.Total = utils.Round(categoryStat.Total, 2)
	}

	// Second pass: fetch missing parents from database
	s.addMissingParents(categoryMap)

	// Third pass: build hierarchy
	parentMap := make(map[int]*types.CategoryStatistic)
	var rootCategories []*types.CategoryStatistic

	for id, cat := range categoryMap {
		if cat.ParentID == nil {
			parentMap[id] = cat
			rootCategories = append(rootCategories, cat)
		}
	}

	// Attach children to parents and aggregate totals
	for _, cat := range categoryMap {
		if cat.ParentID != nil {
			if parent, exists := parentMap[*cat.ParentID]; exists {
				parent.Subcategories = append(parent.Subcategories, *cat)
				parent.Total += cat.Total
				parent.Count += cat.Count
			}
		}
	}

	// Recalculate parent percentages and budget after aggregation
	for _, parent := range rootCategories {
		if len(parent.Subcategories) > 0 {
			parent.Total = utils.Round(parent.Total, 2)

			if totalAmount > 0 {
				parent.Percentage = utils.Round((parent.Total/totalAmount)*100, 2)
			}

			if parent.Budget != nil && *parent.Budget > 0 {
				parent.BudgetPercentage = utils.Round((parent.Total/float64(*parent.Budget))*100, 2)
			}

			// Sort subcategories by total (descending)
			sort.Slice(parent.Subcategories, func(i, j int) bool {
				return parent.Subcategories[i].Total > parent.Subcategories[j].Total
			})
		}
	}

	// Sort parents by budget_percentage (descending), then by total
	sort.Slice(rootCategories, func(i, j int) bool {
		if rootCategories[i].BudgetPercentage != rootCategories[j].BudgetPercentage {
			return rootCategories[i].BudgetPercentage > rootCategories[j].BudgetPercentage
		}
		return rootCategories[i].Total > rootCategories[j].Total
	})

	return rootCategories
}

// Add missing parent categories to the map
func (s *Store) addMissingParents(categoryMap map[int]*types.CategoryStatistic) {
	var missingIDs []int
	for _, cat := range categoryMap {
		if cat.ParentID != nil {
			if _, exists := categoryMap[*cat.ParentID]; !exists {
				missingIDs = append(missingIDs, *cat.ParentID)
			}
		}
	}

	if len(missingIDs) == 0 {
		return
	}

	placeholders := make([]string, len(missingIDs))
	args := make([]interface{}, len(missingIDs))
	for i, id := range missingIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(
		"SELECT id, category_name, color, budget FROM categories WHERE id IN (%s)",
		strings.Join(placeholders, ","),
	)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, color string
		var budget *int

		if err := rows.Scan(&id, &name, &color, &budget); err != nil {
			continue
		}

		categoryMap[id] = &types.CategoryStatistic{
			ID:               id,
			ParentID:         nil,
			Name:             name,
			Color:            color,
			Budget:           budget,
			Count:            0,
			Total:            0,
			Percentage:       0,
			BudgetPercentage: 0,
			Subcategories:    []types.CategoryStatistic{},
		}
	}
}

func (s *Store) GetTransactionStatistics(accountToken string, month, year *int) (*types.TransactionStatistics, error) {
	// Get transactions for the specified period
	transactions, err := s.GetTransactionsDTOByAccountToken(accountToken, month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Calculate totals
	totals, err := s.CalculateTransactionTotals(transactions)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate totals: %w", err)
	}

	// Initialize statistics with early return for empty transactions
	confirmedCount := 0
	for _, tx := range transactions {
		if !tx.IsPending {
			confirmedCount++
		}
	}

	stats := &types.TransactionStatistics{
		TotalTransactions:       confirmedCount,
		LargestDebit:            0,
		LargestCredit:           0,
		CreditCategoryBreakdown: []*types.CategoryStatistic{},
		DebitCategoryBreakdown:  []*types.CategoryStatistic{},
		Totals:                  totals,
	}

	if len(transactions) == 0 {
		return stats, nil
	}

	// Calculate largest amounts and daily totals
	var dailyTotalsMap map[string]*types.DailyTotal
	stats.LargestCredit, stats.LargestDebit, dailyTotalsMap = s.calculateLargestAmountsAndDailyTotals(transactions)

	// Convert daily totals map to a slice and sort by date
	for _, dailyTotal := range dailyTotalsMap {
		stats.DailyTotals = append(stats.DailyTotals, dailyTotal)
	}

	// Build category breakdowns
	creditCategoryMap, debitCategoryMap := s.buildCategoryBreakdowns(transactions)

	// Process breakdowns with percentages and sorting
	stats.CreditCategoryBreakdown = s.processCategoryBreakdown(creditCategoryMap, totals.Credit)
	stats.DebitCategoryBreakdown = s.processCategoryBreakdown(debitCategoryMap, totals.Debit)

	if month != nil && year != nil {
		stats.StartDate, stats.EndDate = getMonthDateRange(month, year)
	} else if len(stats.DailyTotals) > 0 {
		// Find min and max dates from daily totals
		minDate := stats.DailyTotals[0].Date
		for _, dt := range stats.DailyTotals {
			if dt.Date < minDate {
				minDate = dt.Date
			}
		}
		stats.StartDate = minDate
		stats.EndDate = time.Now().Format("2006-01-02")
	} else {
		stats.StartDate = ""
		stats.EndDate = ""
	}

	return stats, nil
}

// GetCategoryMonthlyTrends returns a per-category monthly time series over the
// last `months` months for one transaction type (spending or income). Everything
// is aligned to a shared month axis so the frontend can plot it directly.
func (s *Store) GetCategoryMonthlyTrends(accountToken string, months, transactionTypeID int) (*types.CategoryTrendsResponse, error) {
	if months < 1 {
		months = 12
	}
	if months > 24 {
		months = 24
	}

	axis := buildMonthAxis(time.Now().UTC(), months)
	// index by year*100+month for O(1) row placement onto the axis.
	indexByKey := make(map[int]int, len(axis))
	for i, tm := range axis {
		indexByKey[tm.Year*100+tm.Month] = i
	}
	startDate := fmt.Sprintf("%04d-%02d-01", axis[0].Year, axis[0].Month)

	// Group by the actual (leaf) category but also carry its root, so we can build
	// both the rolled-up parent series and each subcategory's own series in one pass.
	// Root = the parent (or the category itself when it has no parent), matching the
	// statistics breakdown.
	const query = `
		SELECT
			EXTRACT(YEAR FROM t.date)::int  AS yr,
			EXTRACT(MONTH FROM t.date)::int AS mo,
			c.id, c.category_name, c.color,
			root.id, root.category_name, root.color,
			SUM(ABS(t.amount)) AS total
		FROM transactions t
		JOIN categories c    ON c.id = t.category_id
		JOIN categories root ON root.id = COALESCE(c.parent_category_id, c.id)
		WHERE t.account_token = $1
		  AND t.is_pending = false
		  AND t.transaction_type_id = $2
		  AND t.date >= $3::date
		GROUP BY yr, mo, c.id, c.category_name, c.color, root.id, root.category_name, root.color`

	rows, err := s.db.Query(query, accountToken, transactionTypeID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query category trends: %w", err)
	}
	defer rows.Close()

	totals := make([]float64, len(axis))
	roots := make(map[int]*types.CategoryTrend)
	subs := make(map[int]*types.CategoryTrend) // leaf id -> subcategory series
	subParent := make(map[int]int)             // leaf id -> root id

	for rows.Next() {
		var yr, mo, catID, rootID int
		var catName, catColor, rootName, rootColor string
		var total float64
		if err := rows.Scan(&yr, &mo, &catID, &catName, &catColor, &rootID, &rootName, &rootColor, &total); err != nil {
			return nil, fmt.Errorf("failed to scan category trend row: %w", err)
		}

		idx, ok := indexByKey[yr*100+mo]
		if !ok {
			continue // outside the requested window (e.g. a future-dated transaction)
		}

		rounded := utils.Round(total, 2)

		// Roll every leaf (direct-on-parent or subcategory) up into the root series.
		root := roots[rootID]
		if root == nil {
			root = &types.CategoryTrend{ID: rootID, Name: rootName, Color: rootColor, Totals: make([]float64, len(axis))}
			roots[rootID] = root
		}
		root.Totals[idx] = utils.Round(root.Totals[idx]+rounded, 2)
		root.Total = utils.Round(root.Total+rounded, 2)

		// Track actual subcategories (a leaf distinct from its root) as their own series.
		if catID != rootID {
			sub := subs[catID]
			if sub == nil {
				sub = &types.CategoryTrend{ID: catID, Name: catName, Color: catColor, Totals: make([]float64, len(axis))}
				subs[catID] = sub
				subParent[catID] = rootID
			}
			sub.Totals[idx] = utils.Round(sub.Totals[idx]+rounded, 2)
			sub.Total = utils.Round(sub.Total+rounded, 2)
		}

		totals[idx] = utils.Round(totals[idx]+rounded, 2)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate category trend rows: %w", err)
	}

	for leafID, sub := range subs {
		if root := roots[subParent[leafID]]; root != nil {
			root.Subcategories = append(root.Subcategories, sub)
		}
	}

	// Biggest first so the picker lists the most relevant series on top.
	categorySlice := make([]*types.CategoryTrend, 0, len(roots))
	for _, root := range roots {
		if len(root.Subcategories) > 1 {
			sort.SliceStable(root.Subcategories, func(i, j int) bool {
				return root.Subcategories[i].Total > root.Subcategories[j].Total
			})
		}
		categorySlice = append(categorySlice, root)
	}
	sort.SliceStable(categorySlice, func(i, j int) bool {
		return categorySlice[i].Total > categorySlice[j].Total
	})

	var windowTotal float64
	for _, v := range totals {
		windowTotal += v
	}
	windowTotal = utils.Round(windowTotal, 2)
	var monthlyAverage float64
	if len(axis) > 0 {
		monthlyAverage = utils.Round(windowTotal/float64(len(axis)), 2)
	}

	// Income is always the credit side, independent of the selected type, so any
	// category can be shown as a share of income (e.g. a savings category's % of
	// income == the savings rate). It's a separate query because the main pass is
	// filtered to a single transaction_type_id.
	income, err := s.queryMonthlyIncome(accountToken, indexByKey, len(axis), startDate)
	if err != nil {
		return nil, err
	}
	var windowIncome float64
	for _, v := range income {
		windowIncome += v
	}
	windowIncome = utils.Round(windowIncome, 2)

	return &types.CategoryTrendsResponse{
		AccountToken:    accountToken,
		TransactionType: transactionTypeID,
		Months:          axis,
		Totals:          totals,
		Income:          income,
		WindowTotal:     windowTotal,
		WindowIncome:    windowIncome,
		MonthlyAverage:  monthlyAverage,
		Categories:      categorySlice,
		Movers:          computeCategoryMovers(categorySlice, len(axis)),
	}, nil
}

// monthlyAmount is one (year, month) bucket of a pre-summed amount, as scanned
// from an aggregate query. Kept separate from the axis so bucketing onto it can
// be unit-tested without a database.
type monthlyAmount struct {
	year   int
	month  int
	amount float64
}

// bucketMonthlyAmounts places pre-summed monthly amounts onto the shared axis via
// indexByKey, dropping any month outside the window. Pure (no DB) so the income
// aggregation is unit-testable.
func bucketMonthlyAmounts(rows []*monthlyAmount, indexByKey map[int]int, n int) []float64 {
	out := make([]float64, n)
	for _, r := range rows {
		idx, ok := indexByKey[r.year*100+r.month]
		if !ok {
			continue // outside the requested window
		}
		out[idx] = utils.Round(out[idx]+utils.Round(r.amount, 2), 2)
	}
	return out
}

// queryMonthlyIncome returns per-month credit totals aligned to the axis. Incoming
// transfers (rows carrying a transfer_group_id) are excluded so money moved between
// the user's own accounts is not counted as income.
func (s *Store) queryMonthlyIncome(accountToken string, indexByKey map[int]int, n int, startDate string) ([]float64, error) {
	const query = `
		SELECT
			EXTRACT(YEAR FROM date)::int  AS yr,
			EXTRACT(MONTH FROM date)::int AS mo,
			SUM(ABS(amount))             AS total
		FROM transactions
		WHERE account_token = $1
		  AND is_pending = false
		  AND transaction_type_id = $2
		  AND transfer_group_id IS NULL
		  AND date >= $3::date
		GROUP BY yr, mo`

	rows, err := db.QueryList(s.db, query, func(r *sql.Rows) (*monthlyAmount, error) {
		m := &monthlyAmount{}
		if err := r.Scan(&m.year, &m.month, &m.amount); err != nil {
			return nil, err
		}
		return m, nil
	}, accountToken, int(types.CreditTransactionType), startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly income: %w", err)
	}

	return bucketMonthlyAmounts(rows, indexByKey, n), nil
}

// computeCategoryMovers ranks root categories by how much their recent-half
// average moved versus the earlier half of the window, keeping only meaningful
// movers (>|2%|, or brand-new spend), biggest first, capped at six.
func computeCategoryMovers(roots []*types.CategoryTrend, n int) []*types.CategoryMover {
	movers := make([]*types.CategoryMover, 0)
	if n < 2 {
		return movers
	}
	half := n / 2
	avg := func(arr []float64, from, to int) float64 {
		if to <= from {
			return 0
		}
		var sum float64
		for i := from; i < to; i++ {
			sum += arr[i]
		}
		return sum / float64(to-from)
	}

	for _, root := range roots {
		priorAvg := avg(root.Totals, 0, half)
		recentAvg := avg(root.Totals, n-half, n)

		var pct *float64
		direction := ""
		switch {
		case priorAvg == 0 && recentAvg > 0:
			direction = "new"
		case priorAvg == 0:
			continue // no earlier spend and none now
		default:
			p := ((recentAvg - priorAvg) / priorAvg) * 100
			switch {
			case p > 2:
				direction = "up"
			case p < -2:
				direction = "down"
			default:
				continue // flat — not a mover
			}
			rounded := utils.Round(p, 0)
			pct = &rounded
		}

		movers = append(movers, &types.CategoryMover{
			ID:        root.ID,
			Name:      root.Name,
			Color:     root.Color,
			Totals:    root.Totals,
			Pct:       pct,
			Direction: direction,
		})
	}

	// Biggest movers first; brand-new categories float to the top.
	rank := func(m *types.CategoryMover) float64 {
		if m.Direction == "new" {
			return 1e18
		}
		if m.Pct == nil {
			return 0
		}
		return abs(*m.Pct)
	}
	sort.SliceStable(movers, func(i, j int) bool {
		return rank(movers[i]) > rank(movers[j])
	})

	if len(movers) > 6 {
		movers = movers[:6]
	}
	return movers
}

// buildMonthAxis returns `months` consecutive TrendMonths ending with the month
// of `now` (chronological order).
func buildMonthAxis(now time.Time, months int) []types.TrendMonth {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -(months - 1), 0)
	axis := make([]types.TrendMonth, 0, months)
	for i := 0; i < months; i++ {
		d := start.AddDate(0, i, 0)
		axis = append(axis, types.TrendMonth{Month: int(d.Month()), Year: d.Year()})
	}
	return axis
}

// Returns start and end date (YYYY-MM-DD) for a given month/year
func getMonthDateRange(month, year *int) (startDate, endDate string) {
	loc := time.UTC
	start := time.Date(*year, time.Month(*month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, -1)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}
