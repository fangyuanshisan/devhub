-- Plugin migrations registry for v1.3.2 plugin governance.
-- This file is safe to run repeatedly.

CREATE TABLE IF NOT EXISTS plugin_migrations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  version VARCHAR(32) NOT NULL DEFAULT '',
  migration_name VARCHAR(128) NOT NULL,
  checksum VARCHAR(128) NOT NULL DEFAULT '',
  status ENUM('pending','success','failed') NOT NULL DEFAULT 'pending',
  executed_at DATETIME NULL,
  execution_time_ms INT NOT NULL DEFAULT 0,
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_migrations_unique (plugin_code, version, migration_name),
  KEY idx_plugin_migrations_plugin (plugin_code),
  KEY idx_plugin_migrations_status (status),
  KEY idx_plugin_migrations_executed (executed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

