ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS is_pending BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_transactions_account_pending_date
ON transactions (account_token, is_pending, date DESC);

CREATE TABLE IF NOT EXISTS recurring_rules (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_token VARCHAR(255) NOT NULL REFERENCES accounts(token) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    description VARCHAR(255) NOT NULL DEFAULT '',
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'every_x_days')),
    interval_value INTEGER NOT NULL DEFAULT 1 CHECK (interval_value >= 1),
    next_run_date DATE NOT NULL,
    last_generated_date DATE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recurring_rules_due
ON recurring_rules (active, next_run_date);

CREATE INDEX IF NOT EXISTS idx_recurring_rules_user
ON recurring_rules (user_id, active);
