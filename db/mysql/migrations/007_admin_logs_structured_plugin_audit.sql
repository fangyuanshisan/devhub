-- DevHub v1.3.1 patch: structured audit diff fields for plugin governance.
-- MySQL 8 compatible, safe to run repeatedly.

SET @devhub_admin_logs_old_value_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'admin_logs'
    AND COLUMN_NAME = 'old_value'
);

SET @devhub_admin_logs_old_value_sql := IF(
  @devhub_admin_logs_old_value_exists = 0,
  'ALTER TABLE admin_logs ADD COLUMN old_value JSON NULL AFTER target',
  'SELECT 1'
);

PREPARE devhub_admin_logs_old_value_stmt FROM @devhub_admin_logs_old_value_sql;
EXECUTE devhub_admin_logs_old_value_stmt;
DEALLOCATE PREPARE devhub_admin_logs_old_value_stmt;

SET @devhub_admin_logs_new_value_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'admin_logs'
    AND COLUMN_NAME = 'new_value'
);

SET @devhub_admin_logs_new_value_sql := IF(
  @devhub_admin_logs_new_value_exists = 0,
  'ALTER TABLE admin_logs ADD COLUMN new_value JSON NULL AFTER old_value',
  'SELECT 1'
);

PREPARE devhub_admin_logs_new_value_stmt FROM @devhub_admin_logs_new_value_sql;
EXECUTE devhub_admin_logs_new_value_stmt;
DEALLOCATE PREPARE devhub_admin_logs_new_value_stmt;

SET @devhub_admin_logs_metadata_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'admin_logs'
    AND COLUMN_NAME = 'metadata_json'
);

SET @devhub_admin_logs_metadata_sql := IF(
  @devhub_admin_logs_metadata_exists = 0,
  'ALTER TABLE admin_logs ADD COLUMN metadata_json JSON NULL AFTER new_value',
  'SELECT 1'
);

PREPARE devhub_admin_logs_metadata_stmt FROM @devhub_admin_logs_metadata_sql;
EXECUTE devhub_admin_logs_metadata_stmt;
DEALLOCATE PREPARE devhub_admin_logs_metadata_stmt;
