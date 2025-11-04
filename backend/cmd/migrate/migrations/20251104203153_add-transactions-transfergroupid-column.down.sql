DROP INDEX IF EXISTS idx_transactions_transfer_group;

ALTER TABLE transactions DROP COLUMN IF EXISTS transfer_group_id;