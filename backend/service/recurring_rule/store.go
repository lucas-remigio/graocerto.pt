package recurring_rule

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/service/category"
	"github.com/lucas-remigio/wallet-tracker/types"
	"github.com/lucas-remigio/wallet-tracker/utils"
)

type Store struct {
	db *sql.DB
}

const (
	transactionTypeCredit = 1
	transactionTypeDebit  = 2
)

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateRecurringRule(ctx context.Context, rule *types.RecurringRule) (*types.RecurringRule, error) {
	if err := s.validateOwnership(rule.UserID, rule.AccountToken, rule.CategoryID); err != nil {
		return nil, err
	}
	if err := s.validateCategoryTransactionType(rule.UserID, rule.CategoryID, rule.TransactionTypeID); err != nil {
		return nil, err
	}

	var id int
	err := s.db.QueryRow(
		`INSERT INTO recurring_rules
		(user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::date, $11)
		RETURNING id`,
		rule.UserID,
		rule.AccountToken,
		rule.CategoryID,
		rule.TransactionTypeID,
		rule.RecurringTransferGroupID,
		rule.Amount,
		rule.Description,
		string(rule.Frequency),
		rule.IntervalValue,
		rule.NextRunDate,
		rule.Active,
	).Scan(&id)

	if err != nil {
		utils.LogWithContext(ctx).Error("failed to create recurring rule", "user_id", rule.UserID, "error", err)
		return nil, fmt.Errorf("failed to create recurring rule: %w", err)
	}

	utils.LogWithContext(ctx).Info("recurring rule created", "user_id", rule.UserID, "id", id)

	return s.GetRecurringRuleByID(id, rule.UserID)
}

func (s *Store) GetRecurringRulesByUserID(userID int) ([]*types.RecurringRule, error) {
	query := `SELECT id, user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
		FROM recurring_rules
		WHERE user_id = $1
		ORDER BY active DESC, next_run_date ASC, id DESC`
	return db.QueryList(s.db, query, scanRecurringRuleRows, userID)
}

func (s *Store) GetRecurringForecast(userID int, accountToken string, days int) (*types.RecurringForecastResponse, error) {
	if days <= 0 {
		days = 30
	}

	if err := s.validateAccountOwnership(userID, accountToken); err != nil {
		return nil, err
	}

	query := `SELECT id, user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
		FROM recurring_rules
		WHERE user_id = $1 AND account_token = $2 AND active = true
		ORDER BY next_run_date ASC, id ASC`
	rules, err := db.QueryList(s.db, query, scanRecurringRuleRows, userID, accountToken)
	if err != nil {
		return nil, fmt.Errorf("failed to load recurring rules for forecast: %w", err)
	}

	windowStart := time.Now().UTC().Truncate(24 * time.Hour)
	windowEnd := windowStart.AddDate(0, 0, days)

	items := make([]*types.RecurringForecastItem, 0)
	summary := &types.RecurringForecastSummary{}

	for _, rule := range rules {
		nextDate, err := time.Parse("2006-01-02", rule.NextRunDate)
		if err != nil {
			return nil, fmt.Errorf("invalid next run date for rule %d: %w", rule.ID, err)
		}

		// Protect against malformed loops in case of unexpected frequency/date issues.
		for iteration := 0; iteration < 2000; iteration++ {
			if nextDate.After(windowEnd) {
				break
			}

			if !nextDate.Before(windowStart) {
				items = append(items, &types.RecurringForecastItem{
					RecurringRuleID:          rule.ID,
					AccountToken:             rule.AccountToken,
					CategoryID:               rule.CategoryID,
					TransactionTypeID:        rule.TransactionTypeID,
					RecurringTransferGroupID: rule.RecurringTransferGroupID,
					Amount:                   rule.Amount,
					Description:              rule.Description,
					Date:                     nextDate.Format("2006-01-02"),
				})

				switch rule.TransactionTypeID {
				case transactionTypeCredit:
					summary.Credit += rule.Amount
					summary.Difference += rule.Amount
				case transactionTypeDebit:
					summary.Debit += rule.Amount
					summary.Difference -= rule.Amount
				}
			}

			nextDateStr, err := calculateNextRunDate(nextDate.Format("2006-01-02"), rule.Frequency, rule.IntervalValue)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate forecast next run date for rule %d: %w", rule.ID, err)
			}
			nextDate, err = time.Parse("2006-01-02", nextDateStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse forecast next run date for rule %d: %w", rule.ID, err)
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Date != items[j].Date {
			return items[i].Date < items[j].Date
		}
		return items[i].RecurringRuleID < items[j].RecurringRuleID
	})

	return &types.RecurringForecastResponse{
		AccountToken: accountToken,
		Days:         days,
		Items:        items,
		Summary:      summary,
	}, nil
}

func (s *Store) GetRecurringRuleByID(id int, userID int) (*types.RecurringRule, error) {
	query := `SELECT id, user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
		FROM recurring_rules
		WHERE id = $1 AND user_id = $2`
	return db.QuerySingle(s.db, query, scanRecurringRuleRow, id, userID)
}

func (s *Store) UpdateRecurringRule(ctx context.Context, rule *types.RecurringRule, userID int) (*types.RecurringRule, error) {
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
		utils.LogWithContext(ctx).Error("failed to update recurring rule", "id", rule.ID, "error", err)
		return nil, fmt.Errorf("failed to update recurring rule: %w", err)
	}

	utils.LogWithContext(ctx).Info("recurring rule updated", "id", rule.ID, "user_id", userID)

	return s.GetRecurringRuleByID(rule.ID, userID)
}

func (s *Store) CreateRecurringTransfer(ctx context.Context, payload *types.CreateRecurringTransferPayload, userID int) ([]*types.RecurringRule, error) {
	if payload.SourceAccountToken == payload.DestinationAccountToken {
		return nil, fmt.Errorf("source and destination accounts must be different")
	}

	active := true
	if payload.Active != nil {
		active = *payload.Active
	}

	nextRunDate := calculateInitialNextRunDate(payload.Frequency, payload.ExecutionDay, payload.ExecutionWeekday)
	groupID, err := newRecurringTransferGroupID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate recurring transfer group id: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin recurring transfer creation: %w", err)
	}
	defer tx.Rollback()

	debitRule, err := s.createRecurringTransferSide(ctx, tx, &types.RecurringRule{
		UserID:                   userID,
		AccountToken:             payload.SourceAccountToken,
		CategoryID:               payload.DebitCategoryID,
		TransactionTypeID:        transactionTypeDebit,
		RecurringTransferGroupID: &groupID,
		Amount:                   payload.Amount,
		Description:              payload.Description,
		Frequency:                payload.Frequency,
		IntervalValue:            payload.IntervalValue,
		NextRunDate:              nextRunDate,
		Active:                   active,
	})
	if err != nil {
		return nil, err
	}

	creditRule, err := s.createRecurringTransferSide(ctx, tx, &types.RecurringRule{
		UserID:                   userID,
		AccountToken:             payload.DestinationAccountToken,
		CategoryID:               payload.CreditCategoryID,
		TransactionTypeID:        transactionTypeCredit,
		RecurringTransferGroupID: &groupID,
		Amount:                   payload.Amount,
		Description:              payload.Description,
		Frequency:                payload.Frequency,
		IntervalValue:            payload.IntervalValue,
		NextRunDate:              nextRunDate,
		Active:                   active,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		utils.LogWithContext(ctx).Error("failed to commit recurring transfer", "error", err)
		return nil, fmt.Errorf("failed to commit recurring transfer creation: %w", err)
	}

	utils.LogWithContext(ctx).Info("recurring transfer created", "user_id", userID, "group_id", groupID)

	return []*types.RecurringRule{debitRule, creditRule}, nil
}

func (s *Store) UpdateRecurringTransfer(ctx context.Context, groupID string, payload *types.UpdateRecurringTransferPayload, userID int) ([]*types.RecurringRule, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group id is required")
	}
	if payload.SourceAccountToken == payload.DestinationAccountToken {
		return nil, fmt.Errorf("source and destination accounts must be different")
	}

	nextRunDate := payload.NextRunDate
	if nextRunDate == "" {
		nextRunDate = calculateInitialNextRunDate(payload.Frequency, payload.ExecutionDay, payload.ExecutionWeekday)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin recurring transfer update: %w", err)
	}
	defer tx.Rollback()

	var ownerCount int
	err = tx.QueryRow(
		`SELECT COUNT(1)
		 FROM recurring_rules
		 WHERE user_id = $1 AND recurring_transfer_group_id = $2`,
		userID,
		groupID,
	).Scan(&ownerCount)
	if err != nil {
		return nil, fmt.Errorf("failed to validate recurring transfer ownership: %w", err)
	}
	if ownerCount == 0 {
		return nil, fmt.Errorf("recurring transfer not found")
	}

	if err := s.validateOwnership(userID, payload.SourceAccountToken, payload.DebitCategoryID); err != nil {
		return nil, err
	}
	if err := s.validateOwnership(userID, payload.DestinationAccountToken, payload.CreditCategoryID); err != nil {
		return nil, err
	}
	if err := s.validateCategoryTransactionType(userID, payload.DebitCategoryID, transactionTypeDebit); err != nil {
		return nil, err
	}
	if err := s.validateCategoryTransactionType(userID, payload.CreditCategoryID, transactionTypeCredit); err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		`UPDATE recurring_rules
		 SET account_token = $1,
		     category_id = $2,
		     transaction_type_id = $3,
		     amount = $4,
		     description = $5,
		     frequency = $6,
		     interval_value = $7,
		     next_run_date = $8::date,
		     active = $9,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE user_id = $10
		   AND recurring_transfer_group_id = $11
		   AND transaction_type_id = $12`,
		payload.SourceAccountToken,
		payload.DebitCategoryID,
		transactionTypeDebit,
		payload.Amount,
		payload.Description,
		string(payload.Frequency),
		payload.IntervalValue,
		nextRunDate,
		payload.Active,
		userID,
		groupID,
		transactionTypeDebit,
	)
	if err != nil {
		utils.LogWithContext(ctx).Error("failed to update debit side of recurring transfer", "group_id", groupID, "error", err)
		return nil, fmt.Errorf("failed to update debit recurring transfer rule: %w", err)
	}

	_, err = tx.Exec(
		`UPDATE recurring_rules
		 SET account_token = $1,
		     category_id = $2,
		     transaction_type_id = $3,
		     amount = $4,
		     description = $5,
		     frequency = $6,
		     interval_value = $7,
		     next_run_date = $8::date,
		     active = $9,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE user_id = $10
		   AND recurring_transfer_group_id = $11
		   AND transaction_type_id = $12`,
		payload.DestinationAccountToken,
		payload.CreditCategoryID,
		transactionTypeCredit,
		payload.Amount,
		payload.Description,
		string(payload.Frequency),
		payload.IntervalValue,
		nextRunDate,
		payload.Active,
		userID,
		groupID,
		transactionTypeCredit,
	)
	if err != nil {
		utils.LogWithContext(ctx).Error("failed to update credit side of recurring transfer", "group_id", groupID, "error", err)
		return nil, fmt.Errorf("failed to update credit recurring transfer rule: %w", err)
	}

	rows, err := tx.Query(
		`SELECT id, user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
		 FROM recurring_rules
		 WHERE user_id = $1 AND recurring_transfer_group_id = $2
		 ORDER BY transaction_type_id DESC`,
		userID,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated recurring transfer rules: %w", err)
	}
	defer rows.Close()

	updatedRules := []*types.RecurringRule{}
	for rows.Next() {
		rule, scanErr := scanRecurringRuleRows(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		updatedRules = append(updatedRules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		utils.LogWithContext(ctx).Error("failed to commit recurring transfer update", "group_id", groupID, "error", err)
		return nil, fmt.Errorf("failed to commit recurring transfer update: %w", err)
	}

	utils.LogWithContext(ctx).Info("recurring transfer updated", "user_id", userID, "group_id", groupID)

	return updatedRules, nil
}

func (s *Store) DeleteRecurringRule(ctx context.Context, id int, userID int) error {
	current, err := s.GetRecurringRuleByID(id, userID)
	if err != nil {
		return fmt.Errorf("failed to get recurring rule: %w", err)
	}
	if err := db.ValidateOwnership(current.UserID, userID, "recurring rule"); err != nil {
		return err
	}

	if current.RecurringTransferGroupID != nil && *current.RecurringTransferGroupID != "" {
		_, err = db.ExecWithValidation(
			s.db,
			"DELETE FROM recurring_rules WHERE user_id = $1 AND recurring_transfer_group_id = $2",
			userID,
			*current.RecurringTransferGroupID,
		)
	} else {
		_, err = db.ExecWithValidation(s.db, "DELETE FROM recurring_rules WHERE id = $1", id)
	}

	if err != nil {
		utils.LogWithContext(ctx).Error("failed to delete recurring rule", "id", id, "error", err)
		return fmt.Errorf("failed to delete recurring rule: %w", err)
	}

	utils.LogWithContext(ctx).Info("recurring rule deleted", "id", id, "user_id", userID)
	return nil
}

func (s *Store) GeneratePendingTransactionsForDueRules() error {
	const query = `SELECT id, user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at
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

func calculateInitialNextRunDate(frequency types.RecurringFrequency, executionDay *int, executionWeekday *int) string {
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

	if frequency == types.RecurringWeekly && executionWeekday != nil {
		weekday := *executionWeekday
		if weekday < 0 {
			weekday = 0
		}
		if weekday > 6 {
			weekday = 6
		}

		todayStart := now.Truncate(24 * time.Hour)
		currentWeekday := int(todayStart.Weekday())
		offsetDays := weekday - currentWeekday
		if offsetDays < 0 {
			offsetDays += 7
		}
		candidate := todayStart.AddDate(0, 0, offsetDays)
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
	if err := s.validateAccountOwnership(userID, accountToken); err != nil {
		return err
	}

	var categoryOwnerID int
	err := s.db.QueryRow("SELECT user_id FROM categories WHERE id = $1", categoryID).Scan(&categoryOwnerID)
	if err != nil {
		return fmt.Errorf("failed to get category ownership: %w", err)
	}
	if err := db.ValidateOwnership(categoryOwnerID, userID, "category"); err != nil {
		return err
	}

	return nil
}

func (s *Store) validateAccountOwnership(userID int, accountToken string) error {
	var accountOwnerID int
	err := s.db.QueryRow("SELECT user_id FROM accounts WHERE token = $1", accountToken).Scan(&accountOwnerID)
	if err != nil {
		return fmt.Errorf("failed to get account ownership: %w", err)
	}
	if err := db.ValidateOwnership(accountOwnerID, userID, "account"); err != nil {
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
	var recurringTransferGroupID sql.NullString
	err := rows.Scan(
		&r.ID,
		&r.UserID,
		&r.AccountToken,
		&r.CategoryID,
		&r.TransactionTypeID,
		&recurringTransferGroupID,
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
	if recurringTransferGroupID.Valid {
		r.RecurringTransferGroupID = &recurringTransferGroupID.String
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
	var recurringTransferGroupID sql.NullString
	err := row.Scan(
		&r.ID,
		&r.UserID,
		&r.AccountToken,
		&r.CategoryID,
		&r.TransactionTypeID,
		&recurringTransferGroupID,
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
	if recurringTransferGroupID.Valid {
		r.RecurringTransferGroupID = &recurringTransferGroupID.String
	}
	r.NextRunDate = nextRunDate.Format("2006-01-02")
	r.CreatedAt = createdAt.Format(time.RFC3339)
	r.UpdatedAt = updatedAt.Format(time.RFC3339)
	return r, nil
}

func (s *Store) createRecurringTransferSide(ctx context.Context, tx *sql.Tx, rule *types.RecurringRule) (*types.RecurringRule, error) {
	if err := s.validateOwnership(rule.UserID, rule.AccountToken, rule.CategoryID); err != nil {
		return nil, err
	}
	if err := s.validateCategoryTransactionType(rule.UserID, rule.CategoryID, rule.TransactionTypeID); err != nil {
		return nil, err
	}

	var insertedRule types.RecurringRule
	var nextRunDate time.Time
	var createdAt time.Time
	var updatedAt time.Time
	var recurringTransferGroupID sql.NullString

	err := tx.QueryRow(
		`INSERT INTO recurring_rules
		 (user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::date, $11)
		 RETURNING id, user_id, account_token, category_id, transaction_type_id, recurring_transfer_group_id, amount, description, frequency, interval_value, next_run_date, active, created_at, updated_at`,
		rule.UserID,
		rule.AccountToken,
		rule.CategoryID,
		rule.TransactionTypeID,
		rule.RecurringTransferGroupID,
		rule.Amount,
		rule.Description,
		string(rule.Frequency),
		rule.IntervalValue,
		rule.NextRunDate,
		rule.Active,
	).Scan(
		&insertedRule.ID,
		&insertedRule.UserID,
		&insertedRule.AccountToken,
		&insertedRule.CategoryID,
		&insertedRule.TransactionTypeID,
		&recurringTransferGroupID,
		&insertedRule.Amount,
		&insertedRule.Description,
		&insertedRule.Frequency,
		&insertedRule.IntervalValue,
		&nextRunDate,
		&insertedRule.Active,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		utils.LogWithContext(ctx).Error("failed to insert recurring transfer rule side", "user_id", rule.UserID, "type_id", rule.TransactionTypeID, "error", err)
		return nil, fmt.Errorf("failed to insert recurring transfer rule: %w", err)
	}

	if recurringTransferGroupID.Valid {
		insertedRule.RecurringTransferGroupID = &recurringTransferGroupID.String
	}
	insertedRule.NextRunDate = nextRunDate.Format("2006-01-02")
	insertedRule.CreatedAt = createdAt.Format(time.RFC3339)
	insertedRule.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &insertedRule, nil
}

func newRecurringTransferGroupID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
