DROP INDEX IF EXISTS uq_notifications_recurring_due_tomorrow;
DROP INDEX IF EXISTS idx_notifications_user_unread;
DROP INDEX IF EXISTS idx_notifications_user_created;

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_preferences;
