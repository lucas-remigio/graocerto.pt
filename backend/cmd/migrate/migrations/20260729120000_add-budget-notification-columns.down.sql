DROP INDEX IF EXISTS uq_notifications_budget_threshold;

ALTER TABLE notifications DROP COLUMN IF EXISTS threshold_pct;

ALTER TABLE notifications DROP COLUMN IF EXISTS category_id;
