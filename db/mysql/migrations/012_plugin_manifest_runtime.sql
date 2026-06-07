-- v1.3.4+ plugin manifest / lifecycle runtime metadata.
-- This migration is additive and safe to run repeatedly on MySQL 8.

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'plugins' AND COLUMN_NAME = 'source_type'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE plugins ADD COLUMN source_type VARCHAR(32) NOT NULL DEFAULT ''builtin'' AFTER description',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'plugins' AND COLUMN_NAME = 'manifest_json'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE plugins ADD COLUMN manifest_json JSON NULL AFTER source_type',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'plugins' AND COLUMN_NAME = 'manifest_checksum'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE plugins ADD COLUMN manifest_checksum VARCHAR(128) NOT NULL DEFAULT '''' AFTER manifest_json',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'plugins' AND COLUMN_NAME = 'package_checksum'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE plugins ADD COLUMN package_checksum VARCHAR(128) NOT NULL DEFAULT '''' AFTER manifest_checksum',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'plugins' AND COLUMN_NAME = 'compatible_core_version'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE plugins ADD COLUMN compatible_core_version VARCHAR(64) NOT NULL DEFAULT '''' AFTER package_checksum',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
