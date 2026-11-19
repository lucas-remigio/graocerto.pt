-- Fall back to email codes so nobody is locked out when the columns go away.
UPDATE users SET mfa_method = 'email_otp' WHERE mfa_method = 'totp';

ALTER TABLE users
    DROP COLUMN IF EXISTS totp_secret,
    DROP COLUMN IF EXISTS totp_pending_secret,
    DROP COLUMN IF EXISTS totp_last_step;
