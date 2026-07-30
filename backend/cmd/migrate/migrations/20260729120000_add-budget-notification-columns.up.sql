-- Budget-threshold notifications reuse the existing notifications table.
-- category_id  : which budgeted category the alert is about.
-- threshold_pct: which threshold was crossed (80 or 100), so 80% and 100% are distinct rows.
-- target_date  : reused as the first day of the budget period (month), so each
--                (category, threshold) fires at most once per month.
ALTER TABLE notifications
ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE;

ALTER TABLE notifications
ADD COLUMN IF NOT EXISTS threshold_pct SMALLINT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_notifications_budget_threshold
ON notifications (user_id, type, category_id, threshold_pct, target_date)
WHERE type = 'budget_threshold';
