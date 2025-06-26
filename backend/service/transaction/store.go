package transaction

import (
	"database/sql"
	"fmt"
	"sort"
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
		if tx.Category == nil || tx.Category.TransactionType == nil {
			continue
		}

		switch tx.Category.TransactionType.ID {
		case int(types.CreditTransactionType):
			total.Credit += tx.Amount
		case int(types.DebitTransactionType):
			total.Debit += tx.Amount
		}
	}

	total.Difference = total.Credit - total.Debit
	return total, nil
}

// Helper function to get absolute value for float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Process transactions and calculate largest amounts
func (s *Store) calculateLargestAmounts(transactions []*types.TransactionDTO) (largestCredit, largestDebit float64) {
	for _, tx := range transactions {
		if tx.Category == nil {
			continue
		}

		switch tx.Category.TransactionType.ID {
		case int(types.DebitTransactionType):
			if absAmount := abs(tx.Amount); absAmount > largestDebit {
				largestDebit = absAmount
			}
		case int(types.CreditTransactionType):
			if tx.Amount > largestCredit {
				largestCredit = tx.Amount
			}
		}
	}
	return largestCredit, largestDebit
}

// Build category breakdown maps from transactions
func (s *Store) buildCategoryBreakdowns(transactions []*types.TransactionDTO) (
	creditCategoryMap, debitCategoryMap map[string]*types.CategoryStatistic) {

	creditCategoryMap = make(map[string]*types.CategoryStatistic)
	debitCategoryMap = make(map[string]*types.CategoryStatistic)

	for _, tx := range transactions {
		categoryName := "Unknown"
		categoryColor := "#6b7280" // Default gray color

		if tx.Category != nil {
			categoryName = tx.Category.CategoryName
			categoryColor = tx.Category.Color
		}

		absAmount := abs(tx.Amount)

		// Process based on transaction type
		if tx.Category != nil {
			switch tx.Category.TransactionType.ID {
			case int(types.CreditTransactionType):
				s.updateCategoryMap(creditCategoryMap, categoryName, categoryColor, absAmount)
			case int(types.DebitTransactionType):
				s.updateCategoryMap(debitCategoryMap, categoryName, categoryColor, absAmount)
			}
		}
	}

	return creditCategoryMap, debitCategoryMap
}

// Helper to update category map (reduces code duplication)
func (s *Store) updateCategoryMap(categoryMap map[string]*types.CategoryStatistic,
	categoryName, categoryColor string, amount float64) {

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
	categoryMap[categoryName].Total += amount
}

// Calculate percentages and convert map to slice
func (s *Store) processCategoryBreakdown(categoryMap map[string]*types.CategoryStatistic,
	totalAmount float64) []*types.CategoryStatistic {

	breakdown := make([]*types.CategoryStatistic, 0, len(categoryMap))

	for _, categoryStat := range categoryMap {
		if totalAmount > 0 {
			categoryStat.Percentage = (categoryStat.Total / totalAmount) * 100
		}
		breakdown = append(breakdown, categoryStat)
	}

	// Sort by total amount (descending) - O(n log n)
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].Total > breakdown[j].Total
	})

	return breakdown
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
	stats := &types.TransactionStatistics{
		TotalTransactions:       len(transactions),
		LargestDebit:            0,
		LargestCredit:           0,
		CreditCategoryBreakdown: []*types.CategoryStatistic{},
		DebitCategoryBreakdown:  []*types.CategoryStatistic{},
		Totals:                  totals,
	}

	if len(transactions) == 0 {
		return stats, nil
	}

	// Calculate largest amounts
	stats.LargestCredit, stats.LargestDebit = s.calculateLargestAmounts(transactions)

	// Build category breakdowns
	creditCategoryMap, debitCategoryMap := s.buildCategoryBreakdowns(transactions)

	// Process breakdowns with percentages and sorting
	stats.CreditCategoryBreakdown = s.processCategoryBreakdown(creditCategoryMap, totals.Credit)
	stats.DebitCategoryBreakdown = s.processCategoryBreakdown(debitCategoryMap, totals.Debit)

	return stats, nil
}

func (s *Store) GetGroupedTransactionsDTOByAccountToken(accountToken string, month, year *int) (*types.GroupedTransactionsResponse, error) {
	// Get transactions for the account with optional month/year filtering
	transactions, err := s.GetTransactionsDTOByAccountToken(accountToken, month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Calculate totals
	totals, err := s.CalculateTransactionTotals(transactions)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate totals: %w", err)
	}

	// Group transactions by month
	groups := make(map[int]*types.TransactionGroup)

	for _, tx := range transactions {
		// Create month key as integer (1-12)
		monthKey := int(tx.Date.Month())

		if _, exists := groups[monthKey]; !exists {
			groups[monthKey] = &types.TransactionGroup{
				Month:        monthKey,
				Year:         tx.Date.Year(),
				Transactions: []*types.TransactionDTO{},
			}
		}

		groups[monthKey].Transactions = append(groups[monthKey].Transactions, tx)
	}

	// Convert map to sorted slice (most recent first)
	var groupSlice []*types.TransactionGroup
	for _, group := range groups {
		// Sort transactions within each group by date (most recent first)
		sort.Slice(group.Transactions, func(i, j int) bool {
			return group.Transactions[i].Date.After(group.Transactions[j].Date)
		})

		groupSlice = append(groupSlice, group)
	}

	// Sort groups by month key (most recent first)
	sort.Slice(groupSlice, func(i, j int) bool {
		// Sort by year first, then by month (most recent first)
		if groupSlice[i].Year != groupSlice[j].Year {
			return groupSlice[i].Year > groupSlice[j].Year
		}
		return groupSlice[i].Month > groupSlice[j].Month
	})

	return &types.GroupedTransactionsResponse{
		Groups: groupSlice,
		Totals: totals,
	}, nil
}
