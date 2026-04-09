ALTER TABLE recurring_rules
DROP CONSTRAINT IF EXISTS recurring_rules_transaction_type_id_fkey;

ALTER TABLE recurring_rules
DROP COLUMN IF EXISTS transaction_type_id;
