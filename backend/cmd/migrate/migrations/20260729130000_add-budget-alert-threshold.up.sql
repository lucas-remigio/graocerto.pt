-- Per-user budget warning threshold (percentage). The generator always also
-- alerts at 100% (over budget); this column is the earlier warning level.
ALTER TABLE notification_preferences
ADD COLUMN IF NOT EXISTS budget_alert_threshold SMALLINT NOT NULL DEFAULT 80;
