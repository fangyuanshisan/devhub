-- Add real site scope to notifications and admin operation logs.

ALTER TABLE notifications
  ADD COLUMN site_key VARCHAR(64) NOT NULL DEFAULT 'portal' AFTER id,
  ADD KEY idx_notifications_site_read_created (site_key, is_read, created_at);

UPDATE notifications
SET site_key = CASE
  WHEN scope IN ('php','go','java') THEN scope
  ELSE 'portal'
END
WHERE site_key = 'portal';

ALTER TABLE admin_logs
  ADD COLUMN site_key VARCHAR(64) NOT NULL DEFAULT 'portal' AFTER id,
  ADD COLUMN role_code VARCHAR(64) NOT NULL DEFAULT '' AFTER actor,
  ADD KEY idx_admin_logs_site_created (site_key, created_at);
