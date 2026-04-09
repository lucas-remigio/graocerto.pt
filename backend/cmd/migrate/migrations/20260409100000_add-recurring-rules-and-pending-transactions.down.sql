DROP INDEX IF EXISTS idx_recurring_rules_user;
DROP INDEX IF EXISTS idx_recurring_rules_due;
DROP TABLE IF EXISTS recurring_rules;

DROP INDEX IF EXISTS idx_transactions_account_pending_date;

ALTER TABLE transactions
DROP COLUMN IF EXISTS is_pending;
