package transaction

import (
	"database/sql"
	"fmt"
	"sort"
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

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanTransactionDTOFromScanner(s scanner) (*types.TransactionDTO, error) {
	t := new(types.TransactionDTO)
	t.Category = &types.CategoryDTO{}
	t.Category.TransactionType = &types.TransactionType{}

	err := s.Scan(
		&t.ID, &t.AccountToken, &t.Amount, &t.Description, &t.Date, &t.Balance, &t.CreatedAt,
		&t.Category.ID, &t.Category.CategoryName, &t.Category.Color, &t.Category.CreatedAt, &t.Category.UpdatedAt,
		&t.Category.TransactionType.ID, &t.Category.TransactionType.TypeName, &t.Category.TransactionType.TypeSlug,
	)
	if err != nil {
		return nil, err
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

	// do not allow transfers here, they demand a different logic
	if category.TransactionTypeID == 3 {
		return nil, fmt.Errorf("transfers are not allowed here")
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
	if category.TransactionTypeID == (int)(types.DebitTransactionType) {
		amount = amount * -1
	}
	newBalance := account.Balance + amount

	var insertedId int
	err = s.db.QueryRow(
		"INSERT INTO transactions (account_token, category_id, amount, description, date, balance) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		transaction.AccountToken,
		transaction.CategoryId,
		transaction.Amount,
		transaction.Description,
		transaction.Date,
		newBalance,
	).Scan(&insertedId)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// update user account balance
	_, err = db.ExecWithValidation(s.db, "UPDATE accounts SET balance = $1 WHERE token = $2", newBalance, transaction.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

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
		Transaction:    createdDTO,
		AccountBalance: &createdDTO.Balance,
		Months:         availableMonths,
	}, nil
}

func (s *Store) GetTransactionsByAccountToken(accountToken string, month, year *int) ([]*types.Transaction, error) {
	var query string
	var args []interface{}

	baseQuery := `
        SELECT id, account_token, category_id, amount, description, date, balance, created_at 
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
		"t.id, t.account_token, t.amount, t.description, t.date, t.balance, t.created_at, " +
		"c.id, c.category_name, c.color, c.created_at, c.updated_at, " +
		"tt.id, tt.type_name, tt.type_slug " +
		"FROM transactions t " +
		"JOIN categories c ON t.category_id = c.id " +
		"JOIN transaction_types tt ON c.transaction_type_id = tt.id " +
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
			t.id, t.account_token, t.amount, t.description, t.date, t.balance, t.created_at,
			c.id, c.category_name, c.color, c.created_at, c.updated_at,
			tt.id, tt.type_name, tt.type_slug
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		JOIN transaction_types tt ON c.transaction_type_id = tt.id
		WHERE t.id = $1`

	return db.QuerySingle(s.db, query, scanTransactionDTO, id)
}

func (s *Store) GetTransactionById(id int) (*types.Transaction, error) {
	query := "SELECT id, account_token, category_id, amount, description, date, balance, created_at FROM transactions WHERE id = $1"
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
	if currentCategory.TransactionTypeID == (int)(types.DebitTransactionType) {
		currentAmount = currentAmount * -1 // Negate for debit
	}

	// Get the new transaction amount (positive for both credit and debit)
	newAmount := transaction.Amount
	if newCategory.TransactionTypeID == (int)(types.DebitTransactionType) {
		newAmount = newAmount * -1 // Negate for debit
	}

	// Calculate the new balance
	amountDifference := newAmount - currentAmount
	newBalance := currentBalance + amountDifference

	_, err = db.ExecWithValidation(s.db, "UPDATE transactions SET amount = $1, category_id = $2, description = $3, date = $4, balance = $5 WHERE id = $6",
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
		ID:           tx.ID,
		AccountToken: tx.AccountToken,
		CategoryId:   transaction.CategoryID,
		Amount:       transaction.Amount,
		Description:  transaction.Description,
		Date:         transaction.Date,
		Balance:      newBalance,
		CreatedAt:    tx.CreatedAt,
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
		Transaction:    transactionDTO,
		AccountBalance: &transactionDTO.Balance,
		Months:         availableMonths,
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

	// get the transaction category
	catStore := category.NewStore(s.db)
	category, err := catStore.GetCategoryById(tx.CategoryId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
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

	_, err = db.ExecWithValidation(s.db, "DELETE FROM transactions WHERE id = $1", transactionId)
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

func (s *Store) DeleteTransactionAndReturn(transactionId int, userId int) (*types.TransactionChangeResponse, error) {
	transactionDTO, err := s.GetTransactionDTOById(transactionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction DTO: %w", err)
	}

	balance, err := s.DeleteTransaction(transactionId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete transaction: %w", err)
	}

	transactionDTO.Balance = *balance

	// Get available months for the account token
	availableMonths, err := s.GetAvailableTransactionMonthsByAccountToken(transactionDTO.AccountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get available months: %w", err)
	}

	return &types.TransactionChangeResponse{
		Transaction:    transactionDTO,
		AccountBalance: &transactionDTO.Balance,
		Months:         availableMonths,
	}, nil
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

// Process transactions and calculate largest amounts
func (s *Store) calculateLargestAmountsAndDailyTotals(
	transactions []*types.TransactionDTO,
) (largestCredit, largestDebit float64, dailyTotals map[string]float64) {
	dailyTotals = make(map[string]float64)

	for _, tx := range transactions {
		if tx.Category == nil {
			continue
		}

		date := tx.Date.Format("2006-01-02")

		switch tx.Category.TransactionType.ID {
		case int(types.DebitTransactionType):
			dailyTotals[date] -= tx.Amount
			if absAmount := abs(tx.Amount); absAmount > largestDebit {
				largestDebit = absAmount
			}
		case int(types.CreditTransactionType):
			dailyTotals[date] += tx.Amount
			if tx.Amount > largestCredit {
				largestCredit = tx.Amount
			}
		}
	}
	largestCredit = utils.Round(largestCredit, 2)
	largestDebit = utils.Round(largestDebit, 2)
	return largestCredit, largestDebit, dailyTotals
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
			percentage := (categoryStat.Total / totalAmount) * 100
			categoryStat.Percentage = utils.Round(percentage, 2)
		}
		categoryStat.Total = utils.Round(categoryStat.Total, 2)
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
	var dailyTotals map[string]float64
	stats.LargestCredit, stats.LargestDebit, dailyTotals = s.calculateLargestAmountsAndDailyTotals(transactions)

	// Convert daily totals map to a slice
	for date, total := range dailyTotals {
		stats.DailyTotals = append(stats.DailyTotals, &types.DailyTotal{
			Date:  date,
			Total: utils.Round(total, 2),
		})
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

// Returns start and end date (YYYY-MM-DD) for a given month/year
func getMonthDateRange(month, year *int) (startDate, endDate string) {
	loc := time.UTC
	start := time.Date(*year, time.Month(*month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, -1)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}
