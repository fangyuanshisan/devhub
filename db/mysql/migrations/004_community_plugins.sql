-- v1.3.0 patch: per-community plugin enablement.
-- Safe to run repeatedly.

-- This migration can be applied to old databases before 005_core_plugins.sql.
-- Keep a minimal plugins table definition here so the default backfill below
-- does not fail when operators follow numeric migration order.
CREATE TABLE IF NOT EXISTS plugins (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  version VARCHAR(32) NOT NULL DEFAULT '',
  status ENUM('discovered','installed','migrated','configured','enabled','disabled','running','archived','config_invalid','migration_pending','migration_failed','dependency_missing') NOT NULL DEFAULT 'enabled',
  description VARCHAR(500) NOT NULL DEFAULT '',
  config_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugins_code (plugin_code),
  KEY idx_plugins_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS community_plugins (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL,
  plugin_code VARCHAR(64) NOT NULL,
  status ENUM('enabled','disabled') NOT NULL DEFAULT 'enabled',
  sort_order INT NOT NULL DEFAULT 0,
  config_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_community_plugins_community_code (community_id, plugin_code),
  KEY idx_community_plugins_plugin (plugin_code),
  KEY idx_community_plugins_community (community_id),
  CONSTRAINT fk_community_plugins_community FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Backfill defaults: enable all globally-enabled built-in plugins for all communities.
-- You can later disable per community via admin API.
INSERT IGNORE INTO community_plugins (community_id, plugin_code, status, sort_order, config_json, created_at, updated_at)
SELECT c.id, p.plugin_code, 'enabled', 0, NULL, NOW(), NOW()
FROM communities c
JOIN plugins p ON p.status='enabled';
