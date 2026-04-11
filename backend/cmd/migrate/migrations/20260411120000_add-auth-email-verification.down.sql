DROP TABLE IF EXISTS auth_tokens;

ALTER TABLE users
    DROP COLUMN IF EXISTS mfa_method,
    DROP COLUMN IF EXISTS email_verified;
