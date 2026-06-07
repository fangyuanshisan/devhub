-- DevHub v1.3.1 patch: global plugin config placeholder.
-- MySQL 8 new/upgrade schema alignment for plugins.config_json.
-- Go startup migration also applies this defensively for older databases.

SET @devhub_plugins_config_json_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'plugins'
    AND COLUMN_NAME = 'config_json'
);

SET @devhub_plugins_config_json_sql := IF(
  @devhub_plugins_config_json_exists = 0,
  'ALTER TABLE plugins ADD COLUMN config_json JSON NULL AFTER description',
  'SELECT 1'
);

PREPARE devhub_plugins_config_json_stmt FROM @devhub_plugins_config_json_sql;
EXECUTE devhub_plugins_config_json_stmt;
DEALLOCATE PREPARE devhub_plugins_config_json_stmt;
