DROP INDEX IF EXISTS idx_recurring_rules_transfer_group;

ALTER TABLE recurring_rules
DROP COLUMN IF EXISTS recurring_transfer_group_id;
