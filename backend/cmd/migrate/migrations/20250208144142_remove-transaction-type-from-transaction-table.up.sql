ALTER TABLE transactions
    DROP FOREIGN KEY transactions_ibfk_2;
    
ALTER TABLE transactions
    DROP COLUMN transaction_type_id;