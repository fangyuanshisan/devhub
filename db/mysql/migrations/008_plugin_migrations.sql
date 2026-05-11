-- Plugin migrations registry for v1.3.2 plugin governance.
-- This file is safe to run repeatedly.

CREATE TABLE IF NOT EXISTS plugin_migrations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  version VARCHAR(32) NOT NULL DEFAULT '',
  migration_name VARCHAR(128) NOT NULL,
  checksum VARCHAR(128) NOT NULL DEFAULT '',
  status ENUM('pending','running','success','failed') NOT NULL DEFAULT 'pending',
  executed_at DATETIME NULL,
  execution_time_ms INT NOT NULL DEFAULT 0,
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_migrations_unique (plugin_code, version, migration_name),
  KEY idx_plugin_migrations_plugin (plugin_code),
  KEY idx_plugin_migrations_status (status),
  KEY idx_plugin_migrations_executed (executed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE plugin_migrations
  MODIFY COLUMN status ENUM('pending','running','success','failed') NOT NULL DEFAULT 'pending';

SET @devhub_plugin_migrations_updated_at_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'plugin_migrations'
    AND COLUMN_NAME = 'updated_at'
);

SET @devhub_plugin_migrations_updated_at_sql := IF(
  @devhub_plugin_migrations_updated_at_exists = 0,
  'ALTER TABLE plugin_migrations ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at',
  'SELECT 1'
);

PREPARE devhub_plugin_migrations_updated_at_stmt FROM @devhub_plugin_migrations_updated_at_sql;
EXECUTE devhub_plugin_migrations_updated_at_stmt;
DEALLOCATE PREPARE devhub_plugin_migrations_updated_at_stmt;
