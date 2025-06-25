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

func (s *Store) GetTransactionsByAccountToken(accountToken string, month, year *int) ([]*types.Transaction, error) {
	query := `
		SELECT id, account_token, category_id, amount, description, date, balance, created_at 
		FROM transactions 
		WHERE account_token = ?`

	if month != nil && year != nil {
		query += " AND MONTH(date) = ? AND YEAR(date) = ?"
	}

	query += " ORDER BY date DESC, id DESC"

	if month != nil && year != nil {
		return db.QueryList(s.db, query, scanTransaction, accountToken, *month, *year)
	}

	return db.QueryList(s.db, query, scanTransaction, accountToken)
}

func (s *Store) GetTransactionsDTOByAccountToken(accountToken string, month, year *int) ([]*types.TransactionDTO, error) {
	query := "SELECT " +
		"t.id, t.account_token, t.amount, t.description, t.date, t.balance, t.created_at, " +
		"c.id, c.category_name, c.color, c.created_at, c.updated_at, " +
		"tt.id, tt.type_name, tt.type_slug " +
		"FROM transactions t " +
		"JOIN categories c ON t.category_id = c.id " +
		"JOIN transaction_types tt ON c.transaction_type_id = tt.id " +
		"WHERE t.account_token = ? "

	if month != nil && year != nil {
		query += "AND MONTH(t.date) = ? AND YEAR(t.date) = ? "
	}

	query += "ORDER BY t.date DESC, t.id DESC"

	if month != nil && year != nil {
		return db.QueryList(s.db, query, scanTransactionDTO, accountToken, *month, *year)
	}

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

// Store implementation
func (s *Store) GetAvailableTransactionMonthsByAccountToken(accountToken string) ([]*types.MonthYear, error) {
	query := `
        SELECT 
            YEAR(date) as year,
            MONTH(date) as month,
            COUNT(*) as count
        FROM transactions 
        WHERE account_token = ? 
        GROUP BY YEAR(date), MONTH(date)
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
	total := &types.TransactionTotals{
		Debit:      0,
		Credit:     0,
		Difference: 0,
	}

	if transactions == nil {
		return total, fmt.Errorf("transactions cannot be nil")
	}

	for _, tx := range transactions {
		if tx.Category.TransactionType.ID == int(types.CreditTransactionType) {
			total.Credit += tx.Amount
		} else if tx.Category.TransactionType.ID == int(types.DebitTransactionType) {
			total.Debit += tx.Amount
		}
	}

	total.Difference = total.Credit - total.Debit

	return total, nil
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

	// Initialize statistics
	stats := &types.TransactionStatistics{
		TotalTransactions:       len(transactions),
		LargestDebit:            0,
		LargestCredit:           0,
		CreditCategoryBreakdown: []*types.CategoryStatistic{},
		DebitCategoryBreakdown:  []*types.CategoryStatistic{},
		Totals:                  totals,
	}

	// If no transactions, return empty stats
	if len(transactions) == 0 {
		return stats, nil
	}

	// Calculate basic statistics
	var totalAmount float64
	var largestDebit, largestCredit float64 // These will store positive values for display
	categoryMap := make(map[string]*types.CategoryStatistic)
	creditCategoryMap := make(map[string]*types.CategoryStatistic)
	debitCategoryMap := make(map[string]*types.CategoryStatistic)

	creditCount := 0
	debitCount := 0

	for _, tx := range transactions {
		absAmount := tx.Amount
		if absAmount < 0 {
			absAmount = -absAmount
		}
		totalAmount += absAmount

		// Track largest amounts
		if tx.Category != nil && tx.Category.TransactionType.ID == int(types.DebitTransactionType) {
			// For debits, we want the largest absolute value (most negative becomes largest positive)
			if absAmount > largestDebit {
				largestDebit = absAmount
			}
		}
		if tx.Category != nil && tx.Category.TransactionType.ID == int(types.CreditTransactionType) {
			// For credits, we want the largest positive value
			if tx.Amount > largestCredit {
				largestCredit = tx.Amount
			}
		}

		// Category breakdown
		categoryName := "Unknown"
		categoryColor := "#6b7280" // Default gray color
		if tx.Category != nil {
			categoryName = tx.Category.CategoryName
			categoryColor = tx.Category.Color
		}

		// Overall category breakdown
		if _, exists := categoryMap[categoryName]; !exists {
			categoryMap[categoryName] = &types.CategoryStatistic{
				Name:       categoryName,
				Count:      0,
				Total:      0,
				Percentage: 0,
				Color:      categoryColor,
			}
		}
		categoryMap[categoryName].Count++
		categoryMap[categoryName].Total += absAmount

		// Separate breakdowns by transaction type
		if tx.Category != nil && tx.Category.TransactionType.ID == int(types.CreditTransactionType) {
			creditCount++
			if _, exists := creditCategoryMap[categoryName]; !exists {
				creditCategoryMap[categoryName] = &types.CategoryStatistic{
					Name:       categoryName,
					Count:      0,
					Total:      0,
					Percentage: 0,
					Color:      categoryColor,
				}
			}
			creditCategoryMap[categoryName].Count++
			creditCategoryMap[categoryName].Total += absAmount
		} else if tx.Category != nil && tx.Category.TransactionType.ID == int(types.DebitTransactionType) {
			debitCount++
			if _, exists := debitCategoryMap[categoryName]; !exists {
				debitCategoryMap[categoryName] = &types.CategoryStatistic{
					Name:       categoryName,
					Count:      0,
					Total:      0,
					Percentage: 0,
					Color:      categoryColor,
				}
			}
			debitCategoryMap[categoryName].Count++
			debitCategoryMap[categoryName].Total += absAmount
		}
	}

	// Process credit category breakdown
	for _, categoryStat := range creditCategoryMap {
		if totals.Credit > 0 {
			categoryStat.Percentage = (categoryStat.Total / totals.Credit) * 100
		}
		stats.CreditCategoryBreakdown = append(stats.CreditCategoryBreakdown, categoryStat)
	}

	// Process debit category breakdown
	for _, categoryStat := range debitCategoryMap {
		if totals.Debit > 0 {
			categoryStat.Percentage = (categoryStat.Total / totals.Debit) * 100
		}
		stats.DebitCategoryBreakdown = append(stats.DebitCategoryBreakdown, categoryStat)
	}

	// Sort all breakdowns by total amount (descending)
	sortByTotal := func(breakdown []*types.CategoryStatistic) {
		for i := 0; i < len(breakdown)-1; i++ {
			for j := i + 1; j < len(breakdown); j++ {
				if breakdown[i].Total < breakdown[j].Total {
					breakdown[i], breakdown[j] = breakdown[j], breakdown[i]
				}
			}
		}
	}

	sortByTotal(stats.CreditCategoryBreakdown)
	sortByTotal(stats.DebitCategoryBreakdown)
	// Set final calculations
	stats.LargestDebit = largestDebit
	stats.LargestCredit = largestCredit

	return stats, nil
}
