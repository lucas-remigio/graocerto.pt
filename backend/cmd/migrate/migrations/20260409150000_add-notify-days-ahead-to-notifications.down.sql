ALTER TABLE notifications
DROP COLUMN IF EXISTS notify_days_ahead;

ALTER TABLE notification_preferences
DROP COLUMN IF EXISTS notify_days_ahead;
