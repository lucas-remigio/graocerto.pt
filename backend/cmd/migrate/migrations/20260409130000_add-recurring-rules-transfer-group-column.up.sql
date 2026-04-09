ALTER TABLE recurring_rules
ADD COLUMN IF NOT EXISTS recurring_transfer_group_id VARCHAR(36);

CREATE INDEX IF NOT EXISTS idx_recurring_rules_transfer_group
ON recurring_rules (user_id, recurring_transfer_group_id);
