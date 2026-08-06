ALTER TABLE users
    ADD COLUMN IF NOT EXISTS telegram_chat_id VARCHAR(64);

-- A telegram chat may only ever be linked to a single user.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_chat_id
    ON users (telegram_chat_id)
    WHERE telegram_chat_id IS NOT NULL;

-- Holds the in-flight slot-filling conversation for a chat. One row per chat:
-- a new parse replaces the previous one. The awaited question is derived from
-- the data (first item missing a category, then a missing account), never stored.
CREATE TABLE IF NOT EXISTS telegram_pending_transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chat_id VARCHAR(64) NOT NULL UNIQUE,
    transactions JSONB NOT NULL,
    account_token VARCHAR(255) REFERENCES accounts(token) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_telegram_pending_user
    ON telegram_pending_transactions (user_id);
