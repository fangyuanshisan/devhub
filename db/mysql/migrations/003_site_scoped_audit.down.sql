-- Rollback for site-scoped audit fields.

ALTER TABLE admin_logs
  DROP KEY idx_admin_logs_site_created,
  DROP COLUMN role_code,
  DROP COLUMN site_key;

ALTER TABLE notifications
  DROP KEY idx_notifications_site_read_created,
  DROP COLUMN site_key;
