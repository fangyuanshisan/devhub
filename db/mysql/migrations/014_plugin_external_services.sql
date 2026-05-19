-- v1.8.3-S11 external_service runtime preparation.
-- Add a governed external_service config table and extend hook_executions
-- with non-sensitive HTTP health/check observability fields.

CREATE TABLE IF NOT EXISTS plugin_external_services (
  plugin_code VARCHAR(64) NOT NULL,
  service_type VARCHAR(32) NOT NULL DEFAULT 'external_service',
  endpoint_url VARCHAR(1000) NOT NULL DEFAULT '',
  health_check_path VARCHAR(255) NOT NULL DEFAULT '/health',
  timeout_ms INT NOT NULL DEFAULT 3000,
  failure_policy VARCHAR(32) NOT NULL DEFAULT 'warn',
  auth_type VARCHAR(32) NOT NULL DEFAULT 'none',
  token_ref VARCHAR(128) NOT NULL DEFAULT '',
  token_ciphertext VARCHAR(1024) NOT NULL DEFAULT '',
  token_hash VARCHAR(128) NOT NULL DEFAULT '',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  last_health_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  last_checked_at DATETIME NULL,
  last_success_at DATETIME NULL,
  last_failure_at DATETIME NULL,
  failure_count INT NOT NULL DEFAULT 0,
  warning_threshold INT NOT NULL DEFAULT 3,
  error_threshold INT NOT NULL DEFAULT 5,
  last_error_message VARCHAR(1000) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (plugin_code),
  KEY idx_plugin_external_services_status (status, updated_at),
  KEY idx_plugin_external_services_health (last_health_status, last_checked_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hook_executions' AND COLUMN_NAME = 'service_type'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE hook_executions ADD COLUMN service_type VARCHAR(32) NOT NULL DEFAULT '''' AFTER plugin_code',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hook_executions' AND COLUMN_NAME = 'endpoint_url'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE hook_executions ADD COLUMN endpoint_url VARCHAR(1000) NOT NULL DEFAULT '''' AFTER service_type',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hook_executions' AND COLUMN_NAME = 'status'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE hook_executions ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT '''' AFTER blocking',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hook_executions' AND COLUMN_NAME = 'response_status'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE hook_executions ADD COLUMN response_status INT NOT NULL DEFAULT 0 AFTER status',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hook_executions' AND COLUMN_NAME = 'response_body_excerpt'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE hook_executions ADD COLUMN response_body_excerpt VARCHAR(2000) NOT NULL DEFAULT '''' AFTER response_status',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hook_executions' AND COLUMN_NAME = 'request_body_sha256'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE hook_executions ADD COLUMN request_body_sha256 VARCHAR(128) NOT NULL DEFAULT '''' AFTER response_body_excerpt',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hook_executions' AND COLUMN_NAME = 'error_code'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE hook_executions ADD COLUMN error_code VARCHAR(128) NOT NULL DEFAULT '''' AFTER request_body_sha256',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
