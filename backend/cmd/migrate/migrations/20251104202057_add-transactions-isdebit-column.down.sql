DROP INDEX IF EXISTS idx_transactions_transaction_type;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS fk_transactions_transaction_type;

ALTER TABLE transactions DROP COLUMN IF EXISTS transaction_type_id;