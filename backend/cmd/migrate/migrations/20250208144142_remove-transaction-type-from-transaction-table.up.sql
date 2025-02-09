ALTER TABLE transactions
    DROP FOREIGN KEY transactions_ibfk_2,
    DROP COLUMN transaction_type_id;