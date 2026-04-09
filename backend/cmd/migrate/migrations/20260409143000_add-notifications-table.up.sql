CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    min_total_debit NUMERIC(15, 2),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO notification_preferences (user_id, enabled, min_total_debit)
SELECT id, TRUE, NULL
FROM users
ON CONFLICT (user_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(64) NOT NULL,
    account_token VARCHAR(255) REFERENCES accounts(token) ON DELETE SET NULL,
    target_date DATE,
    debit_count INTEGER NOT NULL DEFAULT 0,
    total_debit NUMERIC(15, 2) NOT NULL DEFAULT 0,
    credit_count INTEGER NOT NULL DEFAULT 0,
    total_credit NUMERIC(15, 2) NOT NULL DEFAULT 0,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created
ON notifications (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
ON notifications (user_id, is_read, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_notifications_recurring_due_tomorrow
ON notifications (user_id, type, account_token, target_date)
WHERE type = 'recurring_due_tomorrow';
