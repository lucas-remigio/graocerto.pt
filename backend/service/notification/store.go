package notification

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/types"
)

const recurringDueTomorrowType = "recurring_due_tomorrow"

// BudgetThresholdType marks a notification created when a category's month-to-date
// spending crosses a budget threshold.
const BudgetThresholdType = "budget_threshold"

// overBudgetThreshold is the fixed "you've exceeded the budget" alert that always
// fires in addition to the user's configurable warning threshold.
const overBudgetThreshold = 100

// defaultBudgetAlertThreshold mirrors the DB default for the warning level.
const defaultBudgetAlertThreshold = 80

type Store struct {
	db *sql.DB
}

func NewStore(dbConn *sql.DB) *Store {
	return &Store{db: dbConn}
}

// notificationSelectColumns is shared by the notification read queries. The
// LEFT JOIN on categories is what populates category_name for budget alerts.
const notificationSelectColumns = `n.id, n.user_id, n.type, n.account_token, n.target_date, n.notify_days_ahead,
		n.debit_count, n.total_debit, n.credit_count, n.total_credit, n.is_read, n.pushed, n.created_at,
		n.category_id, n.threshold_pct, c.category_name`

func (s *Store) GetNotificationsByUserID(userID int) ([]*types.Notification, error) {
	query := `SELECT ` + notificationSelectColumns + `
		FROM notifications n
		LEFT JOIN categories c ON c.id = n.category_id
		WHERE n.user_id = $1
		ORDER BY n.is_read ASC, n.created_at DESC
		LIMIT 100`
	return db.QueryList(s.db, query, scanNotificationsRows, userID)
}

func (s *Store) GetUnpushedNotifications() ([]*types.Notification, error) {
	query := `SELECT ` + notificationSelectColumns + `
		FROM notifications n
		LEFT JOIN categories c ON c.id = n.category_id
		WHERE n.pushed = false AND n.is_read = false
		LIMIT 100`
	return db.QueryList(s.db, query, scanNotificationsRows)
}

func (s *Store) MarkNotificationAsPushed(notificationID int) error {
	_, err := db.ExecWithValidation(
		s.db,
		`UPDATE notifications SET pushed = true WHERE id = $1`,
		notificationID,
	)
	return err
}

func (s *Store) GetUnreadNotificationCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	var count int
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}
	return count, nil
}

func (s *Store) MarkNotificationAsRead(notificationID int, userID int) error {
	_, err := db.ExecWithValidation(
		s.db,
		`UPDATE notifications
		 SET is_read = true
		 WHERE id = $1 AND user_id = $2`,
		notificationID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

func (s *Store) GetNotificationPreferences(userID int) (*types.NotificationPreferences, error) {
	if _, err := s.db.Exec(
		`INSERT INTO notification_preferences (user_id, enabled, notify_days_ahead, min_total_debit)
		 VALUES ($1, true, 1, NULL)
		 ON CONFLICT (user_id) DO NOTHING`,
		userID,
	); err != nil {
		return nil, fmt.Errorf("failed to ensure notification preferences: %w", err)
	}

	query := `SELECT user_id, enabled, notify_days_ahead, min_total_debit, budget_alert_threshold, updated_at
		FROM notification_preferences
		WHERE user_id = $1`
	prefs, err := db.QuerySingle(s.db, query, scanNotificationPreferencesRow, userID)
	if err != nil {
		return nil, err
	}

	// Populate push endpoints
	subs, err := s.GetPushSubscriptionsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch push subscriptions for preferences: %w", err)
	}

	prefs.PushEndpoints = make([]string, len(subs))
	for i, sub := range subs {
		prefs.PushEndpoints[i] = sub.Endpoint
	}

	return prefs, nil
}

func (s *Store) UpdateNotificationPreferences(userID int, payload *types.UpdateNotificationPreferencesPayload) (*types.NotificationPreferences, error) {
	_, err := db.ExecWithValidation(
		s.db,
		`INSERT INTO notification_preferences (user_id, enabled, notify_days_ahead, min_total_debit, budget_alert_threshold, updated_at)
		 VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		 ON CONFLICT (user_id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   notify_days_ahead = EXCLUDED.notify_days_ahead,
		   min_total_debit = EXCLUDED.min_total_debit,
		   budget_alert_threshold = EXCLUDED.budget_alert_threshold,
		   updated_at = CURRENT_TIMESTAMP`,
		userID,
		payload.Enabled,
		payload.NotifyDaysAhead,
		payload.MinTotalDebit,
		payload.BudgetAlertThreshold,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update notification preferences: %w", err)
	}

	return s.GetNotificationPreferences(userID)
}

func (s *Store) CreatePushSubscription(userID int, sub *types.PushSubscription) error {
	_, err := db.ExecWithValidation(
		s.db,
		`INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, endpoint) DO UPDATE SET
		   p256dh = EXCLUDED.p256dh,
		   auth = EXCLUDED.auth`,
		userID,
		sub.Endpoint,
		sub.P256dh,
		sub.Auth,
	)
	if err != nil {
		return fmt.Errorf("failed to create push subscription: %w", err)
	}
	return nil
}

func (s *Store) DeletePushSubscription(userID int, endpoint string) error {
	_, err := db.ExecWithValidation(
		s.db,
		`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`,
		userID,
		endpoint,
	)
	if err != nil {
		return fmt.Errorf("failed to delete push subscription: %w", err)
	}
	return nil
}

func (s *Store) GetPushSubscriptionsByUserID(userID int) ([]*types.PushSubscription, error) {
	query := `SELECT id, user_id, endpoint, p256dh, auth, created_at FROM push_subscriptions WHERE user_id = $1`
	return db.QueryList(s.db, query, scanPushSubscriptionRows, userID)
}

func (s *Store) GenerateRecurringDueTomorrowNotifications() error {
	query := `
		INSERT INTO notifications (user_id, type, account_token, target_date, notify_days_ahead, debit_count, total_debit, credit_count, total_credit, is_read)
		SELECT
			r.user_id,
			$1,
			r.account_token,
			(CURRENT_DATE + (COALESCE(p.notify_days_ahead, 1) * INTERVAL '1 day'))::date,
			COALESCE(p.notify_days_ahead, 1),
			COUNT(*) FILTER (WHERE r.transaction_type_id = 2) AS debit_count,
			COALESCE(SUM(r.amount) FILTER (WHERE r.transaction_type_id = 2), 0) AS total_debit,
			COUNT(*) FILTER (WHERE r.transaction_type_id = 1) AS credit_count,
			COALESCE(SUM(r.amount) FILTER (WHERE r.transaction_type_id = 1), 0) AS total_credit,
			false
		FROM recurring_rules r
		LEFT JOIN notification_preferences p ON p.user_id = r.user_id
		WHERE r.active = true
		  AND r.next_run_date = (CURRENT_DATE + (COALESCE(p.notify_days_ahead, 1) * INTERVAL '1 day'))::date
		  AND r.transaction_type_id IN (1, 2)
		  AND COALESCE(p.enabled, true) = true
		GROUP BY r.user_id, r.account_token, p.notify_days_ahead, p.min_total_debit
		HAVING COALESCE(SUM(r.amount) FILTER (WHERE r.transaction_type_id = 2), 0) >= COALESCE(p.min_total_debit, 0)
		   AND (COUNT(*) FILTER (WHERE r.transaction_type_id = 2) > 0 OR COUNT(*) FILTER (WHERE r.transaction_type_id = 1) > 0)
		ON CONFLICT (user_id, type, account_token, target_date)
		WHERE type = 'recurring_due_tomorrow'
		DO UPDATE SET
		  notify_days_ahead = EXCLUDED.notify_days_ahead,
		  debit_count = EXCLUDED.debit_count,
		  total_debit = EXCLUDED.total_debit,
		  credit_count = EXCLUDED.credit_count,
		  total_credit = EXCLUDED.total_credit,
		  is_read = CASE 
			WHEN notifications.total_debit != EXCLUDED.total_debit 
			  OR notifications.total_credit != EXCLUDED.total_credit 
			  OR notifications.debit_count != EXCLUDED.debit_count 
			  OR notifications.credit_count != EXCLUDED.credit_count 
			THEN false 
			ELSE notifications.is_read 
		  END
	`

	if res, err := s.db.Exec(query, recurringDueTomorrowType); err != nil {
		slog.Error("failed to generate recurring notifications", "error", err)
		return fmt.Errorf("failed to generate recurring due tomorrow notifications: %w", err)
	} else {
		rows, _ := res.RowsAffected()
		if rows > 0 {
			slog.Info("recurring notifications generated/updated", "count", rows)
		}
	}

	return nil
}

// budgetSpend is the month-to-date debit spend for one budgeted category, along
// with the owning user's configured warning threshold.
type budgetSpend struct {
	userID        int
	categoryID    int
	spent         float64
	budget        int
	warnThreshold int
}

// budgetThresholdsCrossed returns which budget thresholds the given spend has
// reached: the user's configurable warning level plus the fixed 100% over-budget
// alert (deduped, so a 100% warning level yields a single alert).
func budgetThresholdsCrossed(spent float64, budget int, warnThreshold int) []int {
	if budget <= 0 || spent <= 0 {
		return nil
	}
	pct := (spent / float64(budget)) * 100
	crossed := make([]int, 0, 2)
	seen := make(map[int]bool, 2)
	for _, th := range []int{warnThreshold, overBudgetThreshold} {
		if th <= 0 || seen[th] {
			continue
		}
		seen[th] = true
		if pct >= float64(th) {
			crossed = append(crossed, th)
		}
	}
	return crossed
}

// GenerateBudgetThresholdNotifications creates a notification the first time a
// budgeted debit category's month-to-date spend crosses each threshold. Spend of
// a category includes its subcategories, matching the statistics view. The dedup
// index on (user_id, type, category_id, threshold_pct, target_date) makes it
// idempotent within a month.
func (s *Store) GenerateBudgetThresholdNotifications() error {
	spends, err := s.getMonthlyBudgetSpend()
	if err != nil {
		return err
	}

	now := time.Now()
	periodMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	insertQuery := `
		INSERT INTO notifications
			(user_id, type, category_id, threshold_pct, target_date, notify_days_ahead, debit_count, total_debit, credit_count, total_credit, is_read)
		VALUES ($1, $2, $3, $4, $5, 0, 1, $6, 0, $7, false)
		ON CONFLICT (user_id, type, category_id, threshold_pct, target_date)
		WHERE type = 'budget_threshold'
		DO NOTHING`

	created := 0
	for _, bs := range spends {
		for _, th := range budgetThresholdsCrossed(bs.spent, bs.budget, bs.warnThreshold) {
			res, err := s.db.Exec(insertQuery, bs.userID, BudgetThresholdType, bs.categoryID, th, periodMonth, bs.spent, bs.budget)
			if err != nil {
				return fmt.Errorf("failed to insert budget threshold notification: %w", err)
			}
			if rows, _ := res.RowsAffected(); rows > 0 {
				created += int(rows)
			}
		}
	}

	if created > 0 {
		slog.Info("budget threshold notifications generated", "count", created)
	}
	return nil
}

// getMonthlyBudgetSpend returns, per budgeted debit category, the month-to-date
// spend across all of the user's accounts (subcategory spend rolls up to the
// budgeted parent). Only fires for users whose notifications are enabled.
func (s *Store) getMonthlyBudgetSpend() ([]*budgetSpend, error) {
	query := `
		SELECT c.user_id, c.id, COALESCE(SUM(t.amount), 0) AS spent, c.budget,
			COALESCE(p.budget_alert_threshold, $1) AS warn_threshold
		FROM categories c
		JOIN accounts a ON a.user_id = c.user_id
		JOIN transactions t
			ON t.account_token = a.token
		   AND t.transaction_type_id = 2
		   AND t.is_pending = false
		   AND date_trunc('month', t.date) = date_trunc('month', CURRENT_DATE)
		JOIN categories tc
			ON tc.id = t.category_id
		   AND (tc.id = c.id OR tc.parent_category_id = c.id)
		LEFT JOIN notification_preferences p ON p.user_id = c.user_id
		WHERE c.budget IS NOT NULL
		  AND c.budget > 0
		  AND c.deleted_at IS NULL
		  AND c.transaction_type_id = 2
		  AND COALESCE(p.enabled, true) = true
		GROUP BY c.user_id, c.id, c.budget, p.budget_alert_threshold`

	return db.QueryList(s.db, query, scanBudgetSpendRow, defaultBudgetAlertThreshold)
}

func scanBudgetSpendRow(rows *sql.Rows) (*budgetSpend, error) {
	bs := new(budgetSpend)
	if err := rows.Scan(&bs.userID, &bs.categoryID, &bs.spent, &bs.budget, &bs.warnThreshold); err != nil {
		return nil, err
	}
	return bs, nil
}

func scanNotificationsRows(rows *sql.Rows) (*types.Notification, error) {
	n := new(types.Notification)
	var accountToken sql.NullString
	var targetDate sql.NullTime
	var createdAt time.Time
	var categoryID sql.NullInt64
	var thresholdPct sql.NullInt64
	var categoryName sql.NullString
	if err := rows.Scan(
		&n.ID,
		&n.UserID,
		&n.Type,
		&accountToken,
		&targetDate,
		&n.NotifyDaysAhead,
		&n.DebitCount,
		&n.TotalDebit,
		&n.CreditCount,
		&n.TotalCredit,
		&n.IsRead,
		&n.Pushed,
		&createdAt,
		&categoryID,
		&thresholdPct,
		&categoryName,
	); err != nil {
		return nil, err
	}

	if accountToken.Valid {
		n.AccountToken = &accountToken.String
	}
	if targetDate.Valid {
		formatted := targetDate.Time.Format("2006-01-02")
		n.TargetDate = &formatted
	}
	n.CreatedAt = createdAt.Format(time.RFC3339)
	if categoryID.Valid {
		v := int(categoryID.Int64)
		n.CategoryID = &v
	}
	if thresholdPct.Valid {
		v := int(thresholdPct.Int64)
		n.ThresholdPct = &v
	}
	if categoryName.Valid {
		n.CategoryName = &categoryName.String
	}
	return n, nil
}

func scanNotificationPreferencesRow(row *sql.Row) (*types.NotificationPreferences, error) {
	p := new(types.NotificationPreferences)
	var updatedAt time.Time
	if err := row.Scan(&p.UserID, &p.Enabled, &p.NotifyDaysAhead, &p.MinTotalDebit, &p.BudgetAlertThreshold, &updatedAt); err != nil {
		return nil, err
	}
	p.UpdatedAt = updatedAt.Format(time.RFC3339)
	return p, nil
}

func scanPushSubscriptionRows(rows *sql.Rows) (*types.PushSubscription, error) {
	s := new(types.PushSubscription)
	var createdAt time.Time
	if err := rows.Scan(
		&s.ID,
		&s.UserID,
		&s.Endpoint,
		&s.P256dh,
		&s.Auth,
		&createdAt,
	); err != nil {
		return nil, err
	}
	s.CreatedAt = createdAt.Format(time.RFC3339)
	return s, nil
}
