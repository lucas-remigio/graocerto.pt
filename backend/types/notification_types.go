package types

import "errors"

var ErrInvalidSubscription = errors.New("invalid push subscription")

type NotificationStore interface {
	GetNotificationsByUserID(userID int) ([]*Notification, error)
	GetUnreadNotificationCount(userID int) (int, error)
	MarkNotificationAsRead(notificationID int, userID int) error
	GetNotificationPreferences(userID int) (*NotificationPreferences, error)
	UpdateNotificationPreferences(userID int, payload *UpdateNotificationPreferencesPayload) (*NotificationPreferences, error)
	GenerateRecurringDueTomorrowNotifications() error
	GenerateBudgetThresholdNotifications() error

	// Push Subscriptions
	CreatePushSubscription(userID int, sub *PushSubscription) error
	DeletePushSubscription(userID int, endpoint string) error
	GetPushSubscriptionsByUserID(userID int) ([]*PushSubscription, error)
	GetUnpushedNotifications() ([]*Notification, error)
	MarkNotificationAsPushed(notificationID int) error
}

type PushSubscription struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Endpoint  string `json:"endpoint"`
	P256dh    string `json:"p256dh"`
	Auth      string `json:"auth"`
	CreatedAt string `json:"created_at"`
}

type Notification struct {
	ID              int     `json:"id"`
	UserID          int     `json:"user_id"`
	Type            string  `json:"type"`
	AccountToken    *string `json:"account_token,omitempty"`
	TargetDate      *string `json:"target_date,omitempty"`
	NotifyDaysAhead int     `json:"notify_days_ahead"`
	DebitCount      int     `json:"debit_count"`
	TotalDebit      float64 `json:"total_debit"`
	CreditCount     int     `json:"credit_count"`
	TotalCredit     float64 `json:"total_credit"`
	IsRead          bool    `json:"is_read"`
	Pushed          bool    `json:"pushed"`
	CreatedAt       string  `json:"created_at"`
	// Budget-threshold notifications only. CategoryName is populated on read via
	// a join; for this type TotalDebit is the amount spent and TotalCredit is the
	// category budget.
	CategoryID   *int    `json:"category_id,omitempty"`
	ThresholdPct *int    `json:"threshold_pct,omitempty"`
	CategoryName *string `json:"category_name,omitempty"`
}

type NotificationsResponse struct {
	Notifications []*Notification `json:"notifications"`
}

type UnreadNotificationCountResponse struct {
	Count int `json:"count"`
}

type NotificationPreferences struct {
	UserID               int      `json:"user_id"`
	Enabled              bool     `json:"enabled"`
	NotifyDaysAhead      int      `json:"notify_days_ahead"`
	MinTotalDebit        *float64 `json:"min_total_debit,omitempty"`
	BudgetAlertThreshold int      `json:"budget_alert_threshold"`
	PushEndpoints        []string `json:"push_endpoints"`
	UpdatedAt            string   `json:"updated_at"`
}

type NotificationPreferencesResponse struct {
	Preferences *NotificationPreferences `json:"preferences"`
}

type UpdateNotificationPreferencesPayload struct {
	Enabled              bool     `json:"enabled"`
	NotifyDaysAhead      int      `json:"notify_days_ahead" validate:"required,gte=1,lte=30"`
	MinTotalDebit        *float64 `json:"min_total_debit" validate:"omitempty,gte=0,lte=999999999"`
	BudgetAlertThreshold int      `json:"budget_alert_threshold" validate:"required,gte=1,lte=100"`
}
