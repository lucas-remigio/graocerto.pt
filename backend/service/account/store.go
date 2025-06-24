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

func (s *Store) CreateAccount(account *types.Account) error {
	token, err := utils.GenerateToken(16)
	if err != nil {
		return err
	}
	account.Token = token

	err = db.ExecWithValidation(s.db,
		"INSERT INTO accounts (token, user_id, account_name, balance) VALUES (?, ?, ?, ?)",
		account.Token, account.UserID, account.AccountName, account.Balance,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAccountsByUserId(userId int) ([]*types.Account, error) {
	return db.QueryList(s.db,
		"SELECT id, token, user_id, account_name, balance, created_at FROM accounts WHERE user_id = ?",
		scanRowsIntoAccount, userId)
}

func (s *Store) GetAccountByToken(token string) (*types.Account, error) {
	return db.QuerySingle(s.db,
		"SELECT id, token, user_id, account_name, balance, created_at FROM accounts WHERE token = ?",
		scanRowIntoAccount, token)
}

func (s *Store) GetAccountById(id int) (*types.Account, error) {
	return db.QuerySingle(s.db,
		"SELECT id, token, user_id, account_name, balance, created_at FROM accounts WHERE id = ?",
		scanRowIntoAccount, id)
}

func (s *Store) UpdateAccount(account *types.Account) error {
	// first get the current account so that we can check if the user is the owner of the account
	currentAccount, err := s.GetAccountById(account.ID)
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

func (s *Store) GetAccountFeedbackMonthly(userId int, accountToken string, month, year int) (*types.MonthlyFeedback, error) {

	// check if the account belongs to the user
	account, err := s.GetAccountByToken(accountToken)
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
	fullPrompt := string(promptTemplate) + "\n" + transactionsData.String()

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
	account, err := s.GetAccountByToken(token)
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

func scanRowIntoAccount(row *sql.Row) (*types.Account, error) {
	a := new(types.Account)

	err := row.Scan(&a.ID, &a.Token, &a.UserID, &a.AccountName, &a.Balance, &a.CreatedAt)

	if err != nil {
		return nil, err
	}

	return a, nil
}

func scanRowsIntoAccount(rows *sql.Rows) (*types.Account, error) {
	a := new(types.Account)

	err := rows.Scan(&a.ID, &a.Token, &a.UserID, &a.AccountName, &a.Balance, &a.CreatedAt)

	if err != nil {
		return nil, err
	}

	return a, nil
}
