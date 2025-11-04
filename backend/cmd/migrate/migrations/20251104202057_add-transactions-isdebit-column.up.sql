ALTER TABLE transactions ADD COLUMN transaction_type_id INTEGER;

UPDATE transactions t
SET transaction_type_id = c.transaction_type_id
FROM categories c
WHERE t.category_id = c.id;

ALTER TABLE transactions ALTER COLUMN transaction_type_id SET NOT NULL;

ALTER TABLE transactions 
ADD CONSTRAINT fk_transactions_transaction_type 
FOREIGN KEY (transaction_type_id) REFERENCES transaction_types(id);

CREATE INDEX idx_transactions_transaction_type ON transactions(transaction_type_id);