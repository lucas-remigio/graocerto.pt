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

type Store struct {
	db *sql.DB
}

func NewStore(dbConn *sql.DB) *Store {
	return &Store{db: dbConn}
}

func (s *Store) GetNotificationsByUserID(userID int) ([]*types.Notification, error) {
	query := `SELECT id, user_id, type, account_token, target_date, notify_days_ahead, debit_count, total_debit, credit_count, total_credit, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY is_read ASC, created_at DESC
		LIMIT 100`
	return db.QueryList(s.db, query, scanNotificationsRows, userID)
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

	query := `SELECT user_id, enabled, notify_days_ahead, min_total_debit, updated_at
		FROM notification_preferences
		WHERE user_id = $1`
	return db.QuerySingle(s.db, query, scanNotificationPreferencesRow, userID)
}

func (s *Store) UpdateNotificationPreferences(userID int, payload *types.UpdateNotificationPreferencesPayload) (*types.NotificationPreferences, error) {
	_, err := db.ExecWithValidation(
		s.db,
		`INSERT INTO notification_preferences (user_id, enabled, notify_days_ahead, min_total_debit, updated_at)
		 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		 ON CONFLICT (user_id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   notify_days_ahead = EXCLUDED.notify_days_ahead,
		   min_total_debit = EXCLUDED.min_total_debit,
		   updated_at = CURRENT_TIMESTAMP`,
		userID,
		payload.Enabled,
		payload.NotifyDaysAhead,
		payload.MinTotalDebit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update notification preferences: %w", err)
	}

	return s.GetNotificationPreferences(userID)
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

func scanNotificationsRows(rows *sql.Rows) (*types.Notification, error) {
	n := new(types.Notification)
	var accountToken sql.NullString
	var targetDate sql.NullTime
	var createdAt time.Time
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
		&createdAt,
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
	return n, nil
}

func scanNotificationPreferencesRow(row *sql.Row) (*types.NotificationPreferences, error) {
	p := new(types.NotificationPreferences)
	var updatedAt time.Time
	if err := row.Scan(&p.UserID, &p.Enabled, &p.NotifyDaysAhead, &p.MinTotalDebit, &updatedAt); err != nil {
		return nil, err
	}
	p.UpdatedAt = updatedAt.Format(time.RFC3339)
	return p, nil
}
