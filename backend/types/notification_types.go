package types

type NotificationStore interface {
	GetNotificationsByUserID(userID int) ([]*Notification, error)
	GetUnreadNotificationCount(userID int) (int, error)
	MarkNotificationAsRead(notificationID int, userID int) error
	GetNotificationPreferences(userID int) (*NotificationPreferences, error)
	UpdateNotificationPreferences(userID int, payload *UpdateNotificationPreferencesPayload) (*NotificationPreferences, error)
	GenerateRecurringDueTomorrowNotifications() error
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
	CreatedAt       string  `json:"created_at"`
}

type NotificationsResponse struct {
	Notifications []*Notification `json:"notifications"`
}

type UnreadNotificationCountResponse struct {
	Count int `json:"count"`
}

type NotificationPreferences struct {
	UserID          int      `json:"user_id"`
	Enabled         bool     `json:"enabled"`
	NotifyDaysAhead int      `json:"notify_days_ahead"`
	MinTotalDebit   *float64 `json:"min_total_debit,omitempty"`
	UpdatedAt       string   `json:"updated_at"`
}

type NotificationPreferencesResponse struct {
	Preferences *NotificationPreferences `json:"preferences"`
}

type UpdateNotificationPreferencesPayload struct {
	Enabled         bool     `json:"enabled"`
	NotifyDaysAhead int      `json:"notify_days_ahead" validate:"required,gte=1,lte=30"`
	MinTotalDebit   *float64 `json:"min_total_debit" validate:"omitempty,gte=0,lte=999999999"`
}
