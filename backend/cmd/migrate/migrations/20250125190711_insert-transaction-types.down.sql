-- Remove predefined transaction types
DELETE FROM transaction_types
WHERE (type_name, type_slug) IN (('Credit', 'credit'), ('Debit', 'debit'), ('Transfer', 'transfer'));