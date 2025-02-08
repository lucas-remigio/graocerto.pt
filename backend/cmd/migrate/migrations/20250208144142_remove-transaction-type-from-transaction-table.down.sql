-- Step 1: Add the column back.
ALTER TABLE transactions
    ADD COLUMN transaction_type_id INT UNSIGNED NOT NULL AFTER account_token;

-- Step 2: Re-add the foreign key constraint.
ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_transaction_type
    FOREIGN KEY (transaction_type_id)
    REFERENCES transaction_types(id)
    ON DELETE RESTRICT;