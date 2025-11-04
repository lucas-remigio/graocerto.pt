ALTER TABLE transactions ADD COLUMN transfer_group_id VARCHAR(36);

CREATE INDEX idx_transactions_transfer_group ON transactions(transfer_group_id);