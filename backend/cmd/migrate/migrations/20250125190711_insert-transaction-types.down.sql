-- Remove predefined transaction types
DELETE FROM transaction_types
WHERE type_name IN ('Credit', 'Debit', 'Transfer')
  AND type_slug IN ('credit', 'debit', 'transfer');