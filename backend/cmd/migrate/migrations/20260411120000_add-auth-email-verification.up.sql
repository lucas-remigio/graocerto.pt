ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS mfa_method VARCHAR(32) NOT NULL DEFAULT 'email_otp';

UPDATE users
SET email_verified = TRUE,
    mfa_method = 'email_otp'
WHERE email_verified = FALSE;

CREATE TABLE IF NOT EXISTS auth_tokens (
    id VARCHAR(64) PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose VARCHAR(32) NOT NULL,
    secret_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_purpose
    ON auth_tokens (user_id, purpose);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_secret_hash
    ON auth_tokens (purpose, secret_hash);
