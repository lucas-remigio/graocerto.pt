-- totp_secret / totp_pending_secret hold AES-GCM ciphertext (see TOTP_ENC_KEY), never plaintext.
-- totp_last_step is the last accepted 30s time step, so a code can't be replayed.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS totp_secret TEXT,
    ADD COLUMN IF NOT EXISTS totp_pending_secret TEXT,
    ADD COLUMN IF NOT EXISTS totp_last_step BIGINT NOT NULL DEFAULT 0;
