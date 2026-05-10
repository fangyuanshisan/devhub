-- v1.3.0 patch: per-community plugin enablement.
-- Safe to run repeatedly.

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
