package recurring_rule

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateRecurringRule(rule *types.RecurringRule) (*types.RecurringRule, error) {
	if err := s.validateOwnership(rule.UserID, rule.AccountToken, rule.CategoryID); err != nil {
		return nil, err
	}
	if err := s.validateCategoryTransactionType(rule.UserID, rule.CategoryID, rule.TransactionTypeID); err != nil {
		return nil, err
	}

	var id int
	err := s.db.QueryRow(
		`INSERT INTO recurring_rules
		(user_id, account_token, category_id, transaction_type_id, amount, description, frequency, interval_value, next_run_date, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::date, $10)
		RETURNING id`,
		rule.UserID,
		rule.AccountToken,
		rule.CategoryID,
		rule.TransactionTypeID,
		rule.Amount,
		rule.Description,
		string(rule.Frequency),
		rule.IntervalValue,
		rule.NextRunDate,
		rule.Active,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create recurring rule: %w", err)
	}

	return s.GetRecurringRuleByID(id, rule.UserID)
}

func (s *Store) GetRecurringRulesByUserID(userID int) ([]*types.RecurringRule, error) {
	query := `SELECT id, user_id, account_token, category_id, transaction_type_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
		FROM recurring_rules
		WHERE user_id = $1
		ORDER BY active DESC, next_run_date ASC, id DESC`
	return db.QueryList(s.db, query, scanRecurringRuleRows, userID)
}

func (s *Store) GetRecurringRuleByID(id int, userID int) (*types.RecurringRule, error) {
	query := `SELECT id, user_id, account_token, category_id, transaction_type_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
		FROM recurring_rules
		WHERE id = $1 AND user_id = $2`
	return db.QuerySingle(s.db, query, scanRecurringRuleRow, id, userID)
}

func (s *Store) UpdateRecurringRule(rule *types.RecurringRule, userID int) (*types.RecurringRule, error) {
	current, err := s.GetRecurringRuleByID(rule.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring rule: %w", err)
	}
	if err := db.ValidateOwnership(current.UserID, userID, "recurring rule"); err != nil {
		return nil, err
	}

	if err := s.validateOwnership(userID, rule.AccountToken, rule.CategoryID); err != nil {
		return nil, err
	}
	if err := s.validateCategoryTransactionType(userID, rule.CategoryID, rule.TransactionTypeID); err != nil {
		return nil, err
	}

	_, err = db.ExecWithValidation(
		s.db,
		`UPDATE recurring_rules
		SET account_token = $1, category_id = $2, transaction_type_id = $3, amount = $4, description = $5,
			frequency = $6, interval_value = $7, next_run_date = $8::date, active = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $10`,
		rule.AccountToken,
		rule.CategoryID,
		rule.TransactionTypeID,
		rule.Amount,
		rule.Description,
		string(rule.Frequency),
		rule.IntervalValue,
		rule.NextRunDate,
		rule.Active,
		rule.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update recurring rule: %w", err)
	}

	return s.GetRecurringRuleByID(rule.ID, userID)
}

func (s *Store) DeleteRecurringRule(id int, userID int) error {
	current, err := s.GetRecurringRuleByID(id, userID)
	if err != nil {
		return fmt.Errorf("failed to get recurring rule: %w", err)
	}
	if err := db.ValidateOwnership(current.UserID, userID, "recurring rule"); err != nil {
		return err
	}

	_, err = db.ExecWithValidation(s.db, "DELETE FROM recurring_rules WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete recurring rule: %w", err)
	}
	return nil
}

func (s *Store) GeneratePendingTransactionsForDueRules() error {
	const query = `SELECT id, user_id, account_token, category_id, transaction_type_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
		FROM recurring_rules
		WHERE active = true AND next_run_date <= CURRENT_DATE
		ORDER BY next_run_date ASC, id ASC
		LIMIT 500`

	rules, err := db.QueryList(s.db, query, scanRecurringRuleRows)
	if err != nil {
		return fmt.Errorf("failed to query due recurring rules: %w", err)
	}

	for _, rule := range rules {
		if err := s.generatePendingTransactionForRule(rule); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) generatePendingTransactionForRule(rule *types.RecurringRule) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin recurring transaction generation: %w", err)
	}
	defer tx.Rollback()

	var existingCount int
	err = tx.QueryRow(
		`SELECT COUNT(1)
		 FROM transactions
		 WHERE account_token = $1
		   AND category_id = $2
		   AND is_pending = true
		   AND DATE(date) = $3::date
		   AND description = $4
		   AND amount = $5`,
		rule.AccountToken,
		rule.CategoryID,
		rule.NextRunDate,
		rule.Description,
		rule.Amount,
	).Scan(&existingCount)
	if err != nil {
		return fmt.Errorf("failed to check pending transaction idempotency: %w", err)
	}

	if existingCount == 0 {
		_, err = tx.Exec(
			`INSERT INTO transactions
			(account_token, transaction_type_id, category_id, amount, description, date, balance, transfer_group_id, is_pending)
			SELECT $1, c.transaction_type_id, $2, $3, $4, $5::date, 0, NULL, true
			FROM categories c
			WHERE c.id = $2`,
			rule.AccountToken,
			rule.CategoryID,
			rule.Amount,
			rule.Description,
			rule.NextRunDate,
		)
		if err != nil {
			return fmt.Errorf("failed to insert pending transaction for recurring rule %d: %w", rule.ID, err)
		}
	}

	nextDate, err := calculateNextRunDate(rule.NextRunDate, rule.Frequency, rule.IntervalValue)
	if err != nil {
		return fmt.Errorf("failed to calculate next run date: %w", err)
	}

	_, err = tx.Exec(
		`UPDATE recurring_rules
		SET last_generated_date = $1::date, next_run_date = $2::date, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`,
		rule.NextRunDate,
		nextDate,
		rule.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update recurring next run date: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit recurring generation: %w", err)
	}
	return nil
}

func calculateNextRunDate(currentDate string, frequency types.RecurringFrequency, interval int) (string, error) {
	parsed, err := time.Parse("2006-01-02", currentDate)
	if err != nil {
		return "", err
	}

	if interval < 1 {
		interval = 1
	}

	var next time.Time
	switch frequency {
	case types.RecurringDaily:
		next = parsed.AddDate(0, 0, interval)
	case types.RecurringWeekly:
		next = parsed.AddDate(0, 0, 7*interval)
	case types.RecurringMonthly:
		next = addMonthsWithMonthEndFallback(parsed, interval)
	case types.RecurringEveryXDays:
		next = parsed.AddDate(0, 0, interval)
	default:
		return "", fmt.Errorf("invalid recurring frequency: %s", frequency)
	}

	return next.Format("2006-01-02"), nil
}

func calculateInitialNextRunDate(frequency types.RecurringFrequency, executionDay *int) string {
	now := time.Now().UTC()

	if frequency == types.RecurringMonthly && executionDay != nil {
		day := *executionDay
		if day < 1 {
			day = 1
		}
		if day > 31 {
			day = 31
		}

		year, month, _ := now.Date()
		candidate := dateWithMonthEndFallback(year, month, day)
		if candidate.Before(now.Truncate(24 * time.Hour)) {
			nextMonth := now.AddDate(0, 1, 0)
			y2, m2, _ := nextMonth.Date()
			candidate = dateWithMonthEndFallback(y2, m2, day)
		}
		return candidate.Format("2006-01-02")
	}

	return now.Format("2006-01-02")
}

func addMonthsWithMonthEndFallback(source time.Time, months int) time.Time {
	targetBase := source.AddDate(0, months, 0)
	year, month, _ := targetBase.Date()
	return dateWithMonthEndFallback(year, month, source.Day())
}

func dateWithMonthEndFallback(year int, month time.Month, day int) time.Time {
	lastDay := daysInMonth(year, month)
	if day > lastDay {
		day = lastDay
	}
	if day < 1 {
		day = 1
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
	// day 0 of next month returns last day of current month
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (s *Store) validateOwnership(userID int, accountToken string, categoryID int) error {
	var accountOwnerID int
	err := s.db.QueryRow("SELECT user_id FROM accounts WHERE token = $1", accountToken).Scan(&accountOwnerID)
	if err != nil {
		return fmt.Errorf("failed to get account ownership: %w", err)
	}
	if err := db.ValidateOwnership(accountOwnerID, userID, "account"); err != nil {
		return err
	}

	var categoryOwnerID int
	err = s.db.QueryRow("SELECT user_id FROM categories WHERE id = $1", categoryID).Scan(&categoryOwnerID)
	if err != nil {
		return fmt.Errorf("failed to get category ownership: %w", err)
	}
	if err := db.ValidateOwnership(categoryOwnerID, userID, "category"); err != nil {
		return err
	}

	return nil
}

func (s *Store) validateCategoryTransactionType(userID int, categoryID int, transactionTypeID int) error {
	catStore := category.NewStore(s.db)
	cat, err := catStore.GetCategoryById(categoryID, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch category: %w", err)
	}
	if cat.TransactionTypeId != transactionTypeID {
		return fmt.Errorf("category does not match selected transaction type")
	}
	return nil
}

func scanRecurringRuleRows(rows *sql.Rows) (*types.RecurringRule, error) {
	r := new(types.RecurringRule)
	var nextRunDate time.Time
	var createdAt time.Time
	var updatedAt time.Time
	err := rows.Scan(
		&r.ID,
		&r.UserID,
		&r.AccountToken,
		&r.CategoryID,
		&r.TransactionTypeID,
		&r.Amount,
		&r.Description,
		&r.Frequency,
		&r.IntervalValue,
		&nextRunDate,
		&r.Active,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	r.NextRunDate = nextRunDate.Format("2006-01-02")
	r.CreatedAt = createdAt.Format(time.RFC3339)
	r.UpdatedAt = updatedAt.Format(time.RFC3339)
	return r, nil
}

func scanRecurringRuleRow(row *sql.Row) (*types.RecurringRule, error) {
	r := new(types.RecurringRule)
	var nextRunDate time.Time
	var createdAt time.Time
	var updatedAt time.Time
	err := row.Scan(
		&r.ID,
		&r.UserID,
		&r.AccountToken,
		&r.CategoryID,
		&r.TransactionTypeID,
		&r.Amount,
		&r.Description,
		&r.Frequency,
		&r.IntervalValue,
		&nextRunDate,
		&r.Active,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	r.NextRunDate = nextRunDate.Format("2006-01-02")
	r.CreatedAt = createdAt.Format(time.RFC3339)
	r.UpdatedAt = updatedAt.Format(time.RFC3339)
	return r, nil
}
