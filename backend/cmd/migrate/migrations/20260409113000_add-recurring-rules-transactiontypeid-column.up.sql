ALTER TABLE recurring_rules
ADD COLUMN IF NOT EXISTS transaction_type_id INTEGER;

UPDATE recurring_rules rr
SET transaction_type_id = c.transaction_type_id
FROM categories c
WHERE c.id = rr.category_id
  AND rr.transaction_type_id IS NULL;

ALTER TABLE recurring_rules
ALTER COLUMN transaction_type_id SET NOT NULL;

ALTER TABLE recurring_rules
ADD CONSTRAINT recurring_rules_transaction_type_id_fkey
FOREIGN KEY (transaction_type_id) REFERENCES transaction_types(id) ON DELETE RESTRICT;
