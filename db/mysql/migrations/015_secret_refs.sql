-- v1.8.4-S14: Core SecretCenter secret refs storage.

CREATE TABLE IF NOT EXISTS secret_refs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  ref VARCHAR(255) NOT NULL,
  namespace VARCHAR(64) NOT NULL,
  name VARCHAR(255) NOT NULL,
  key_id VARCHAR(64) NOT NULL DEFAULT '',
  encrypted_value VARCHAR(2048) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  description VARCHAR(500) NOT NULL DEFAULT '',
  last_used_at DATETIME NULL,
  usage_count INT NOT NULL DEFAULT 0,
  rotated_at DATETIME NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_secret_refs_ref (ref),
  KEY idx_secret_refs_namespace_updated (namespace, updated_at),
  KEY idx_secret_refs_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

