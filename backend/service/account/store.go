package account

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Store struct {
	db                *sql.DB
	categoryStore     types.CategoryStore
	openAiStore       types.OpenAIStore
	transactionsStore types.TransactionStore
}

func NewStore(db *sql.DB, categoryStore types.CategoryStore, openAiStore types.OpenAIStore) *Store {
	return &Store{
		db:            db,
		categoryStore: categoryStore,
		openAiStore:   openAiStore,
	}
}

func (s *Store) SetTransactionStore(transactionsStore types.TransactionStore) {
	s.transactionsStore = transactionsStore
}

const accountColumns = `
    id, token, user_id, account_name, balance, created_at, order_index, is_favorite
`

func (s *Store) GetAccountsByUserId(userId int) ([]*types.Account, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM accounts WHERE user_id = ? ORDER BY order_index`,
		accountColumns,
	)
	return db.QueryList(
		s.db,
		query,
		scanRowsIntoAccount,
		userId,
	)
}

func (s *Store) GetAccountByToken(token string, userId int) (*types.Account, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM accounts WHERE token = ? AND user_id = ?`,
		accountColumns,
	)
	return db.QuerySingle(
		s.db,
		query,
		scanRowIntoAccount,
		token, userId,
	)
}

func (s *Store) GetAccountById(id int, userId int) (*types.Account, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM accounts WHERE id = ? AND user_id = ?`,
		accountColumns,
	)
	return db.QuerySingle(
		s.db,
		query,
		scanRowIntoAccount,
		id, userId,
	)
}

func (s *Store) CreateAccount(account *types.Account) error {
	token, err := utils.GenerateToken(16)
	if err != nil {
		return err
	}
	account.Token = token

	var maxOrderIndex int
	err = s.db.QueryRow("SELECT COALESCE(MAX(order_index), 0) FROM accounts WHERE user_id = ?", account.UserID).Scan(&maxOrderIndex)
	if err != nil {
		return err
	}
	account.OrderIndex = maxOrderIndex + 1

	err = db.ExecWithValidation(s.db,
		"INSERT INTO accounts (token, user_id, account_name, balance, order_index) VALUES (?, ?, ?, ?, ?)",
		account.Token, account.UserID, account.AccountName, account.Balance, account.OrderIndex,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateAccount(account *types.Account, userId int) error {
	// first get the current account so that we can check if the user is the owner of the account
	currentAccount, err := s.GetAccountById(account.ID, userId)
	if err != nil {
		return err
	}

	if err := db.ValidateOwnership(account.UserID, currentAccount.UserID, "account"); err != nil {
		return err
	}

	return db.ExecWithValidation(s.db,
		"UPDATE accounts SET account_name = ?, balance = ? WHERE id = ?",
		account.AccountName, account.Balance, account.ID)
}

func (s *Store) GetAccountFeedbackMonthly(userId int, accountToken, language string, month, year int) (*types.MonthlyFeedback, error) {
	// check if the account belongs to the user
	account, err := s.GetAccountByToken(accountToken, userId)
	if err != nil {
		return nil, err
	}
	if account.UserID != userId {
		return nil, fmt.Errorf("user does not have permission to access this account")
	}

	// Get the transactions for the account using the transaction store
	transactions, err := s.transactionsStore.GetTransactionsByAccountToken(accountToken, &month, &year)
	if err != nil {
		return nil, fmt.Errorf("error getting transactions: %v", err)
	}

	categories, err := s.categoryStore.GetCategoriesByUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("error getting categories: %v", err)
	}

	// Create a map to store category names by ID
	categoryMap := make(map[int]*types.Category)
	for _, category := range categories {
		categoryMap[category.ID] = category
	}

	// Format transactions for the prompt
	var transactionsData strings.Builder

	for _, tx := range transactions {
		// Determine transaction type based on transaction type ID
		txType := "DEBIT"
		categoryName := "Uncategorized"

		// Get category info if available
		if category, exists := categoryMap[tx.CategoryId]; exists {
			categoryName = category.CategoryName

			// Determine transaction type based on category's transaction type ID
			switch category.TransactionTypeID {
			case int(types.CreditTransactionType):
				txType = "CREDIT"
			case int(types.DebitTransactionType):
				txType = "DEBIT"
			case int(types.TransferTransactionType):
				txType = "TRANSFER"
			}
		}

		// get the transaction type by the category transaction type id

		// Format the transaction line
		transactionsData.WriteString(fmt.Sprintf("- Date: %s | Description: %s | Amount: %.2f | Type: %s | Category: %s\n",
			tx.Date,
			tx.Description,
			tx.Amount,
			txType,
			categoryName))
	}

	// Read the prompt template
	promptTemplate, err := os.ReadFile("prompts/monthlyFeedback.txt")
	if err != nil {
		return nil, fmt.Errorf("error reading prompt template: %v", err)
	}

	// Combine template with transactions data
	feedbackLanguage := fmt.Sprintf("\n\n\n Give the feedback in the following language: %s", language)
	fullPrompt := string(promptTemplate) + "\n" + transactionsData.String() + feedbackLanguage

	log.Println("Full prompt:", fullPrompt)

	// Call the OpenAI API to get the feedback
	message, err := s.openAiStore.GenerateGPT4Response(fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("error generating feedback: %v", err)
	}

	log.Println("Generated message:", message)

	// unmarshal the message to get the feedback
	feedback := new(types.MonthlyFeedback)
	err = json.Unmarshal([]byte(message), feedback)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling feedback: %v", err)
	}

	return feedback, nil
}

func (s *Store) DeleteAccount(token string, userId int) error {
	// first get the account so that we can check if the user is the owner of the account
	account, err := s.GetAccountByToken(token, userId)
	if err != nil {
		return err
	}

	if account.UserID != userId {
		return fmt.Errorf("user does not have permission to delete this account")
	}

	// delete all transactions associated with the account
	err = db.ExecWithValidation(s.db, "DELETE FROM transactions WHERE account_token = ?", token)
	if err != nil {
		return err
	}

	// delete the account
	err = db.ExecWithValidation(s.db, "DELETE FROM accounts WHERE token = ? AND user_id = ?", token, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ReorderAccounts(userId int, accounts []types.ReorderAccount) error {
	// Check: all tokens belong to the user
	tokens := make([]any, len(accounts))
	for i, account := range accounts {
		tokens[i] = account.Token
	}
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM accounts WHERE user_id = ? AND token IN (%s)",
		strings.TrimRight(strings.Repeat("?,", len(tokens)), ","),
	)
	args := append([]any{userId}, tokens...)
	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to verify account ownership: %w", err)
	}
	if count != len(accounts) {
		return fmt.Errorf("one or more accounts do not belong to the user")
	}

	// Check: all order indexes are unique
	orderIndexes := make(map[int]bool)
	for _, account := range accounts {
		if orderIndexes[account.OrderIndex] {
			return fmt.Errorf("duplicate order_index found: %d", account.OrderIndex)
		}
		orderIndexes[account.OrderIndex] = true
	}

	// Proceed with update
	for _, account := range accounts {
		err := db.ExecWithValidation(s.db, "UPDATE accounts SET order_index = ? WHERE token = ? AND user_id = ?", account.OrderIndex, account.Token, userId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) FavoriteAccount(token string, userId int, isFavorite bool) error {
	// first get the account so that we can check if the user is the owner of the account
	account, err := s.GetAccountByToken(token, userId)
	if err != nil {
		return err
	}

	if account.UserID != userId {
		return fmt.Errorf("user does not have permission to favorite this account")
	}

	return db.ExecWithValidation(s.db,
		"UPDATE accounts SET is_favorite = ? WHERE token = ? AND user_id = ?",
		isFavorite, token, userId)
}

func scanRowIntoAccount(row *sql.Row) (*types.Account, error) {
	a := new(types.Account)
	err := row.Scan(
		&a.ID,
		&a.Token,
		&a.UserID,
		&a.AccountName,
		&a.Balance,
		&a.CreatedAt,
		&a.OrderIndex,
		&a.IsFavorite,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func scanRowsIntoAccount(rows *sql.Rows) (*types.Account, error) {
	a := new(types.Account)
	err := rows.Scan(
		&a.ID,
		&a.Token,
		&a.UserID,
		&a.AccountName,
		&a.Balance,
		&a.CreatedAt,
		&a.OrderIndex,
		&a.IsFavorite,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}
