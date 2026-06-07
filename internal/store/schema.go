package store

// mySQLSchema defines the authoritative MySQL 8 DDL used by the application auto-migration.
const mySQLSchema = `
-- DevHub MySQL 8 schema.
-- This file is safe to run repeatedly; application demo data is seeded by Go when sites is empty.

CREATE TABLE IF NOT EXISTS sites (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  site_key VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  logo VARCHAR(64) NOT NULL DEFAULT '',
  title VARCHAR(255) NOT NULL DEFAULT '',
  subtitle VARCHAR(255) NOT NULL DEFAULT '',
  pub_text VARCHAR(128) NOT NULL DEFAULT '',
  description TEXT NULL,
  color VARCHAR(32) NOT NULL DEFAULT '',
  status ENUM('enable', 'disable') NOT NULL DEFAULT 'enable',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_sites_site_key (site_key),
  KEY idx_sites_status_sort (status, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS boards (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  board_key VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  site_key VARCHAR(64) NOT NULL DEFAULT 'all',
  sort_order INT NOT NULL DEFAULT 0,
  visible TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_boards_board_key (board_key),
  KEY idx_boards_site_visible_sort (site_key, visible, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tags (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  site_key VARCHAR(64) NOT NULL DEFAULT 'portal',
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  description TEXT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'enable',
  merged_to_id BIGINT UNSIGNED NULL,
  sort_order INT NOT NULL DEFAULT 0,
  use_count INT UNSIGNED NOT NULL DEFAULT 0,
  follower_count INT UNSIGNED NOT NULL DEFAULT 0,
  hot_score INT UNSIGNED NOT NULL DEFAULT 0,
  seo_title VARCHAR(255) NOT NULL DEFAULT '',
  seo_description VARCHAR(500) NOT NULL DEFAULT '',
  seo_keywords VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tags_site_slug (site_key, slug),
  KEY idx_tags_site_status_sort (site_key, status, sort_order),
  KEY idx_tags_use_count (use_count),
  KEY idx_tags_merged_to (merged_to_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tag_aliases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tag_id BIGINT UNSIGNED NOT NULL,
  site_key VARCHAR(64) NOT NULL DEFAULT 'portal',
  alias VARCHAR(128) NOT NULL,
  alias_slug VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tag_aliases_site_alias_slug (site_key, alias_slug),
  KEY idx_tag_aliases_tag_id (tag_id),
  CONSTRAINT fk_tag_aliases_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS posts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  site_key VARCHAR(64) NOT NULL,
  board_key VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  summary TEXT NULL,
  content MEDIUMTEXT NULL,
  author VARCHAR(128) NOT NULL DEFAULT '',
  status ENUM('draft', 'review', 'publish', 'offline', 'rejected') NOT NULL DEFAULT 'publish',
  pinned TINYINT(1) NOT NULL DEFAULT 0,
  recommended TINYINT(1) NOT NULL DEFAULT 0,
  reject_reason TEXT NULL,
  offline_reason TEXT NULL,
  views INT UNSIGNED NOT NULL DEFAULT 0,
  likes INT UNSIGNED NOT NULL DEFAULT 0,
  comments INT UNSIGNED NOT NULL DEFAULT 0,
  tags_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_posts_site_status_created (site_key, status, created_at),
  KEY idx_posts_board_status_created (board_key, status, created_at),
  KEY idx_posts_hot (views, likes, comments),
  KEY idx_posts_pinned_recommended (pinned, recommended),
  FULLTEXT KEY ft_posts_search (title, summary, content),
  CONSTRAINT fk_posts_site FOREIGN KEY (site_key) REFERENCES sites (site_key) ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT fk_posts_board FOREIGN KEY (board_key) REFERENCES boards (board_key) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS comments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  post_id BIGINT UNSIGNED NOT NULL,
  topic_id BIGINT UNSIGNED NULL,
  parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reply_to_user_id BIGINT UNSIGNED NULL,
  user_id BIGINT UNSIGNED NULL,
  author VARCHAR(128) NOT NULL DEFAULT '',
  to_author VARCHAR(128) NOT NULL DEFAULT '',
  text TEXT NULL,
  content_html MEDIUMTEXT NULL,
  status ENUM('normal', 'illegal', 'deleted', 'hidden', 'pending', 'rejected') NOT NULL DEFAULT 'normal',
  likes INT UNSIGNED NOT NULL DEFAULT 0,
  is_best TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  KEY idx_comments_topic_created (topic_id, created_at),
  KEY idx_comments_post_created (post_id, created_at),
  KEY idx_comments_parent (parent_id),
  KEY idx_comments_user_created (user_id, created_at),
  KEY idx_comments_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  site_key VARCHAR(64) NOT NULL DEFAULT 'portal',
  actor_user_id BIGINT UNSIGNED NULL,
  type VARCHAR(64) NOT NULL DEFAULT '',
  target_type VARCHAR(32) NOT NULL DEFAULT '',
  target_id BIGINT UNSIGNED NULL,
  topic_id BIGINT UNSIGNED NULL,
  comment_id BIGINT UNSIGNED NULL,
  title VARCHAR(255) NOT NULL,
  content TEXT NULL,
  scope VARCHAR(64) NOT NULL DEFAULT 'all',
  user_id BIGINT UNSIGNED NULL,
  is_read TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  read_at DATETIME NULL,
  PRIMARY KEY (id),
  KEY idx_notifications_read_created (is_read, created_at),
  KEY idx_notifications_site_read_created (site_key, is_read, created_at),
  KEY idx_notifications_user_created (user_id, created_at),
  KEY idx_notifications_type_target (type, target_type, target_id),
  KEY idx_notifications_scope_created (scope, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  nickname VARCHAR(128) NOT NULL DEFAULT '',
  email VARCHAR(128) NOT NULL DEFAULT '',
  phone VARCHAR(32) NOT NULL DEFAULT '',
  password_hash VARCHAR(255) NOT NULL,
  avatar VARCHAR(255) NOT NULL DEFAULT '',
  status ENUM('normal', 'forbidden') NOT NULL DEFAULT 'normal',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  last_login_at DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email (email),
  KEY idx_users_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  builtin TINYINT(1) NOT NULL DEFAULT 0,
  description TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_roles_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(128) NOT NULL,
  module VARCHAR(64) NOT NULL,
  action VARCHAR(64) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_permissions_code (code),
  KEY idx_permissions_module_action (module, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id BIGINT UNSIGNED NOT NULL,
  permission_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (role_id, permission_id),
  CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
  CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  site_key VARCHAR(64) NOT NULL DEFAULT '*',
  status ENUM('normal', 'disabled') NOT NULL DEFAULT 'normal',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_roles_scope (user_id, role_id, site_key),
  KEY idx_user_roles_site (site_key, status),
  CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  token_hash CHAR(64) NOT NULL,
  token_type VARCHAR(32) NOT NULL DEFAULT 'user',
  expires_at DATETIME NOT NULL,
  revoked_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_refresh_tokens_hash (token_hash),
  KEY idx_refresh_tokens_user (user_id, revoked_at),
  KEY idx_refresh_tokens_type_user (token_type, user_id, revoked_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS admin_roles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  builtin TINYINT(1) NOT NULL DEFAULT 0,
  description TEXT NULL,
  permissions_json JSON NULL,
  user_count INT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_roles_name (name),
  KEY idx_admin_roles_builtin (builtin)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  nickname VARCHAR(128) NOT NULL DEFAULT '',
  avatar VARCHAR(255) NOT NULL DEFAULT '',
  phone VARCHAR(32) NOT NULL DEFAULT '',
  email VARCHAR(128) NOT NULL DEFAULT '',
  status ENUM('normal', 'forbidden') NOT NULL DEFAULT 'normal',
  role_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  role_name VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_login_at DATETIME NULL,
  violation_note TEXT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_users_username (username),
  KEY idx_admin_users_status_created (status, created_at),
  KEY idx_admin_users_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS admin_settings (
  id BIGINT UNSIGNED NOT NULL,
  site_name VARCHAR(128) NOT NULL,
  copyright VARCHAR(255) NOT NULL,
  default_page_size INT UNSIGNED NOT NULL,
  review_timeout_hour INT UNSIGNED NOT NULL,
  password_rule VARCHAR(255) NOT NULL,
  captcha_enabled TINYINT(1) NOT NULL,
  search_default VARCHAR(64) NOT NULL,
  search_sort VARCHAR(64) NOT NULL,
  hot_view_weight INT UNSIGNED NOT NULL,
  hot_like_weight INT UNSIGNED NOT NULL,
  hot_comment_weight INT UNSIGNED NOT NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS system_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  namespace VARCHAR(64) NOT NULL,
  setting_key VARCHAR(128) NOT NULL,
  value_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_system_settings_namespace_key (namespace, setting_key),
  KEY idx_system_settings_updated (namespace, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS admin_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  site_key VARCHAR(64) NOT NULL DEFAULT 'portal',
  log_type VARCHAR(64) NOT NULL,
  actor VARCHAR(128) NOT NULL,
  actor_type VARCHAR(32) NOT NULL DEFAULT 'admin_user',
  actor_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  role_code VARCHAR(64) NOT NULL DEFAULT '',
  action VARCHAR(255) NOT NULL,
  target VARCHAR(255) NOT NULL,
  old_value JSON NULL,
  new_value JSON NULL,
  metadata_json JSON NULL,
  ip VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_admin_logs_site_created (site_key, created_at),
  KEY idx_admin_logs_type_created (log_type, created_at),
  KEY idx_admin_logs_actor_created (actor, created_at),
  KEY idx_admin_logs_actor_type_created (actor_type, created_at),
  KEY idx_admin_logs_target (target)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- New tables for DevHub generic community system
-- Communities (子站表) - 替代 sites，但保留 sites 表用于兼容
CREATE TABLE IF NOT EXISTS communities (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL COMMENT '子站名称，如 PHP、Go、Java',
  slug VARCHAR(64) NOT NULL UNIQUE COMMENT '子站标识，如 php、go、java',
  logo VARCHAR(255) NOT NULL DEFAULT '',
  cover_image VARCHAR(500) NOT NULL DEFAULT '',
  slogan VARCHAR(255) NOT NULL DEFAULT '',
  description TEXT NULL,
  theme_color VARCHAR(32) NOT NULL DEFAULT '',
  seo_title VARCHAR(255) NOT NULL DEFAULT '',
  seo_description VARCHAR(500) NOT NULL DEFAULT '',
  seo_keywords VARCHAR(500) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 0禁用 2归档',
  follower_count INT UNSIGNED NOT NULL DEFAULT 0,
  topic_count INT UNSIGNED NOT NULL DEFAULT 0,
  comment_count INT UNSIGNED NOT NULL DEFAULT 0,
  hot_score INT UNSIGNED NOT NULL DEFAULT 0,
  announcement_title VARCHAR(255) NOT NULL DEFAULT '',
  announcement_content TEXT NULL,
  announcement_url VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_communities_slug (slug),
  KEY idx_communities_status_sort (status, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugins (系统插件表)
CREATE TABLE IF NOT EXISTS plugins (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  version VARCHAR(32) NOT NULL DEFAULT '',
  status ENUM('discovered','installed','migrated','configured','enabled','disabled','running','archived','config_invalid','migration_pending','migration_failed','dependency_missing') NOT NULL DEFAULT 'enabled',
  description VARCHAR(500) NOT NULL DEFAULT '',
  source_type VARCHAR(32) NOT NULL DEFAULT 'builtin',
  manifest_json JSON NULL,
  manifest_checksum VARCHAR(128) NOT NULL DEFAULT '',
  package_checksum VARCHAR(128) NOT NULL DEFAULT '',
  compatible_core_version VARCHAR(64) NOT NULL DEFAULT '',
  config_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugins_code (plugin_code),
  KEY idx_plugins_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Community Plugins (子站插件启用关系)
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

-- Plugin Migrations (插件迁移执行记录)
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

-- Plugin Config Versions (插件配置版本历史)
CREATE TABLE IF NOT EXISTS plugin_config_versions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  scope VARCHAR(32) NOT NULL DEFAULT 'global',
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  version_no INT NOT NULL DEFAULT 0,
  config_json JSON NULL,
  config_hash VARCHAR(128) NOT NULL DEFAULT '',
  changed_keys_json JSON NULL,
  diff_json JSON NULL,
  source VARCHAR(32) NOT NULL DEFAULT 'manual',
  operator_type VARCHAR(32) NOT NULL DEFAULT 'admin_user',
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_name VARCHAR(128) NOT NULL DEFAULT '',
  reason VARCHAR(255) NOT NULL DEFAULT '',
  previous_version_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  rollback_from_version_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  metadata_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_config_versions_lookup (plugin_code, scope, community_id, version_no),
  KEY idx_plugin_config_versions_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Approval Requests (插件安装/升级审批)
CREATE TABLE IF NOT EXISTS plugin_approval_requests (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  request_no VARCHAR(64) NOT NULL DEFAULT '',
  action VARCHAR(32) NOT NULL,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  plugin_name VARCHAR(128) NOT NULL DEFAULT '',
  current_version VARCHAR(32) NOT NULL DEFAULT '',
  target_version VARCHAR(32) NOT NULL DEFAULT '',
  package_path VARCHAR(500) NOT NULL DEFAULT '',
  package_checksum_status VARCHAR(32) NOT NULL DEFAULT '',
  package_risk_level VARCHAR(32) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  reason VARCHAR(1000) NOT NULL DEFAULT '',
  requested_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  requested_by_name VARCHAR(128) NOT NULL DEFAULT '',
  requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  reviewed_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reviewed_by_name VARCHAR(128) NOT NULL DEFAULT '',
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(1000) NOT NULL DEFAULT '',
  executed_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  executed_at DATETIME NULL,
  execute_result_json JSON NULL,
  manifest_json JSON NULL,
  dry_run_json JSON NULL,
  risk_report_json JSON NULL,
  dependency_summary_json JSON NULL,
  compatibility_json JSON NULL,
  changed_keys_json JSON NULL,
  diff_json JSON NULL,
  error_code VARCHAR(64) NOT NULL DEFAULT '',
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  metadata_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_approvals_status_created (status, created_at),
  KEY idx_plugin_approvals_action_created (action, created_at),
  KEY idx_plugin_approvals_plugin_created (plugin_code, created_at),
  KEY idx_plugin_approvals_requested (requested_by, requested_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Package Uploads (zip 上传包生命周期)
CREATE TABLE IF NOT EXISTS plugin_package_uploads (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  upload_id VARCHAR(80) NOT NULL,
  original_filename VARCHAR(255) NOT NULL DEFAULT '',
  uploaded_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  uploaded_by_name VARCHAR(128) NOT NULL DEFAULT '',
  uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status VARCHAR(32) NOT NULL DEFAULT 'uploaded',
  package_code VARCHAR(64) NOT NULL DEFAULT '',
  package_name VARCHAR(128) NOT NULL DEFAULT '',
  package_version VARCHAR(32) NOT NULL DEFAULT '',
  upload_path VARCHAR(500) NOT NULL DEFAULT '',
  staging_path VARCHAR(500) NOT NULL DEFAULT '',
  package_path VARCHAR(500) NOT NULL DEFAULT '',
  promoted_path VARCHAR(500) NOT NULL DEFAULT '',
  compressed_size BIGINT NOT NULL DEFAULT 0,
  uncompressed_size BIGINT NOT NULL DEFAULT 0,
  file_count INT NOT NULL DEFAULT 0,
  checksum_status VARCHAR(32) NOT NULL DEFAULT '',
  signature_status VARCHAR(32) NOT NULL DEFAULT '',
  publisher_id VARCHAR(128) NOT NULL DEFAULT '',
  trust_status VARCHAR(32) NOT NULL DEFAULT '',
  risk_level VARCHAR(32) NOT NULL DEFAULT '',
  risk_report_json JSON NULL,
  zip_scan_json JSON NULL,
  file_scan_json JSON NULL,
  manifest_validation_json JSON NULL,
  install_dry_run_json JSON NULL,
  approval_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  install_approval_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  expires_at DATETIME NULL,
  deleted_at DATETIME NULL,
  error_code VARCHAR(80) NOT NULL DEFAULT '',
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  metadata_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_package_uploads_upload_id (upload_id),
  KEY idx_plugin_package_uploads_status_created (status, created_at),
  KEY idx_plugin_package_uploads_package (package_code, package_version),
  KEY idx_plugin_package_uploads_uploaded_by (uploaded_by, uploaded_at),
  KEY idx_plugin_package_uploads_risk (risk_level, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Package Downloads (远程插件包安全下载到 staging)
CREATE TABLE IF NOT EXISTS plugin_package_downloads (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  version VARCHAR(64) NOT NULL DEFAULT '',
  source_url VARCHAR(1000) NOT NULL DEFAULT '',
  final_url VARCHAR(1000) NOT NULL DEFAULT '',
  signature_url VARCHAR(1000) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  file_name VARCHAR(255) NOT NULL DEFAULT '',
  staging_path VARCHAR(500) NOT NULL DEFAULT '',
  file_size BIGINT NOT NULL DEFAULT 0,
  sha256_expected VARCHAR(128) NOT NULL DEFAULT '',
  sha256_actual VARCHAR(128) NOT NULL DEFAULT '',
  content_type VARCHAR(255) NOT NULL DEFAULT '',
  error_code VARCHAR(80) NOT NULL DEFAULT '',
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  downloaded_at DATETIME NULL,
  deleted_at DATETIME NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_package_downloads_status_created (status, created_at),
  KEY idx_plugin_package_downloads_plugin_version (plugin_code, version),
  KEY idx_plugin_package_downloads_created_by (created_by, created_at),
  KEY idx_plugin_package_downloads_sha (sha256_actual)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS plugin_package_prechecks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  package_download_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  version VARCHAR(64) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'passed',
  manifest_json JSON NULL,
  package_path VARCHAR(500) NOT NULL DEFAULT '',
  staging_path VARCHAR(500) NOT NULL DEFAULT '',
  checksum_status VARCHAR(32) NOT NULL DEFAULT '',
  error_code VARCHAR(80) NOT NULL DEFAULT '',
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_package_prechecks_status_created (status, created_at),
  KEY idx_plugin_package_prechecks_plugin_version (plugin_code, version),
  KEY idx_plugin_package_prechecks_download (package_download_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS plugin_package_compat_checks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  package_download_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  package_precheck_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  version VARCHAR(64) NOT NULL DEFAULT '',
  status VARCHAR(64) NOT NULL DEFAULT 'pending',
  can_install TINYINT(1) NOT NULL DEFAULT 0,
  core_version VARCHAR(64) NOT NULL DEFAULT '',
  compatible_core_version VARCHAR(128) NOT NULL DEFAULT '',
  dependency_result_json JSON NULL,
  conflict_result_json JSON NULL,
  permission_result_json JSON NULL,
  route_result_json JSON NULL,
  menu_result_json JSON NULL,
  hook_result_json JSON NULL,
  config_schema_result_json JSON NULL,
  migration_result_json JSON NULL,
  warnings_json JSON NULL,
  errors_json JSON NULL,
  summary_json JSON NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_package_compat_status_created (status, created_at),
  KEY idx_plugin_package_compat_precheck (package_precheck_id),
  KEY idx_plugin_package_compat_plugin_version (plugin_code, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS plugin_package_signatures (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  package_download_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  package_precheck_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  package_compat_check_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  version VARCHAR(64) NOT NULL DEFAULT '',
  publisher_id VARCHAR(128) NOT NULL DEFAULT '',
  key_id VARCHAR(128) NOT NULL DEFAULT '',
  algorithm VARCHAR(32) NOT NULL DEFAULT '',
  status VARCHAR(64) NOT NULL DEFAULT 'pending',
  signature_url VARCHAR(1000) NOT NULL DEFAULT '',
  signature_file_path VARCHAR(500) NOT NULL DEFAULT '',
  package_sha256 VARCHAR(128) NOT NULL DEFAULT '',
  manifest_sha256 VARCHAR(128) NOT NULL DEFAULT '',
  signature_payload_json JSON NULL,
  signature_base64 VARCHAR(512) NOT NULL DEFAULT '',
  verified_at DATETIME NULL,
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  warnings_json JSON NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_package_signatures_status_created (status, created_at),
  KEY idx_plugin_package_signatures_plugin_version (plugin_code, version),
  KEY idx_plugin_package_signatures_precheck (package_precheck_id),
  KEY idx_plugin_package_signatures_download (package_download_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Enable Prechecks (启用前安全检查)
CREATE TABLE IF NOT EXISTS plugin_enable_prechecks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  version VARCHAR(64) NOT NULL DEFAULT '',
  plugin_install_task_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  plugin_installation_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(64) NOT NULL DEFAULT 'pending',
  can_enable TINYINT(1) NOT NULL DEFAULT 0,
  core_version VARCHAR(64) NOT NULL DEFAULT '',
  installed_path VARCHAR(500) NOT NULL DEFAULT '',
  manifest_sha256 VARCHAR(128) NOT NULL DEFAULT '',
  file_integrity_result_json JSON NULL,
  manifest_result_json JSON NULL,
  dependency_result_json JSON NULL,
  config_result_json JSON NULL,
  migration_result_json JSON NULL,
  permission_result_json JSON NULL,
  menu_result_json JSON NULL,
  route_result_json JSON NULL,
  hook_result_json JSON NULL,
  content_type_result_json JSON NULL,
  runtime_result_json JSON NULL,
  warnings_json JSON NULL,
  errors_json JSON NULL,
  summary_json JSON NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_enable_prechecks_status_created (status, created_at),
  KEY idx_plugin_enable_prechecks_plugin_version (plugin_code, version),
  KEY idx_plugin_enable_prechecks_install_task (plugin_install_task_id),
  KEY idx_plugin_enable_prechecks_installation (plugin_installation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Enable Tasks (启用与运行时注册)
CREATE TABLE IF NOT EXISTS plugin_enable_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  version VARCHAR(64) NOT NULL DEFAULT '',
  plugin_install_task_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  plugin_enable_precheck_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(64) NOT NULL DEFAULT 'pending',
  previous_status VARCHAR(64) NOT NULL DEFAULT '',
  new_status VARCHAR(64) NOT NULL DEFAULT '',
  registered_content_types_json JSON NULL,
  registered_permissions_json JSON NULL,
  registered_menus_json JSON NULL,
  registered_routes_json JSON NULL,
  registered_hooks_json JSON NULL,
  effective_config_json JSON NULL,
  errors_json JSON NULL,
  warnings_json JSON NULL,
  rollback_log_json JSON NULL,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  duration_ms BIGINT NOT NULL DEFAULT 0,
  enabled_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_enable_tasks_status_created (status, created_at),
  KEY idx_plugin_enable_tasks_plugin_version (plugin_code, version),
  KEY idx_plugin_enable_tasks_precheck (plugin_enable_precheck_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Webhook Governance (for external plugin services, non_blocking only for now)
CREATE TABLE IF NOT EXISTS webhook_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(64) NOT NULL,
  event_name VARCHAR(128) NOT NULL DEFAULT '',
  event_type VARCHAR(64) NOT NULL DEFAULT '',
  plugin_code VARCHAR(64) NOT NULL,
  hook_name VARCHAR(128) NOT NULL DEFAULT '',
  mode VARCHAR(32) NOT NULL DEFAULT 'non_blocking',
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  actor_type VARCHAR(32) NOT NULL DEFAULT '',
  actor_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  resource_type VARCHAR(32) NOT NULL DEFAULT '',
  resource_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  request_id VARCHAR(64) NOT NULL DEFAULT '',
  payload_json JSON NULL,
  metadata_json JSON NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/delivering/delivered/failed/skipped/circuit_open',
  occurred_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_webhook_events_event_id (event_id),
  KEY idx_webhook_events_plugin_hook (plugin_code, hook_name),
  KEY idx_webhook_events_status_created (status, created_at),
  KEY idx_webhook_events_request (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS webhook_deliveries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  delivery_id VARCHAR(64) NOT NULL,
  event_id VARCHAR(64) NOT NULL,
  plugin_code VARCHAR(64) NOT NULL,
  hook_name VARCHAR(128) NOT NULL DEFAULT '',
  target_url VARCHAR(1000) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/sending/success/failed/retry_scheduled/retry_exhausted/skipped/circuit_open',
  attempt INT NOT NULL DEFAULT 1,
  max_attempts INT NOT NULL DEFAULT 5,
  signature_alg VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'HMAC-SHA256',
  secret_ref VARCHAR(128) NOT NULL DEFAULT '',
  body_sha256 VARCHAR(128) NOT NULL DEFAULT '',
  signature_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'signed|secret_missing|secret_disabled|secret_revoked|secret_expired|sign_failed',
  signed_at DATETIME NULL,
  signature_error VARCHAR(500) NOT NULL DEFAULT '',
  next_retry_at DATETIME NULL,
  retry_reason VARCHAR(64) NOT NULL DEFAULT '',
  request_headers_json JSON NULL,
  request_body_sha256 VARCHAR(128) NOT NULL DEFAULT '',
  response_status INT NOT NULL DEFAULT 0,
  response_body_excerpt VARCHAR(2000) NOT NULL DEFAULT '',
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  duration_ms BIGINT NOT NULL DEFAULT 0,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_webhook_deliveries_delivery_id (delivery_id),
  KEY idx_webhook_deliveries_event (event_id),
  KEY idx_webhook_deliveries_plugin_hook (plugin_code, hook_name),
  KEY idx_webhook_deliveries_status_retry (status, next_retry_at),
  KEY idx_webhook_deliveries_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS plugin_webhook_secrets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  target_url VARCHAR(512) NOT NULL,
  secret_ref VARCHAR(128) NOT NULL,
  secret_ciphertext VARCHAR(512) NOT NULL DEFAULT '',
  secret_hash VARCHAR(128) NOT NULL DEFAULT '',
  version INT NOT NULL DEFAULT 1,
  status ENUM('active','previous','disabled','revoked','expired') NOT NULL DEFAULT 'active',
  rotation_group VARCHAR(128) NOT NULL DEFAULT '',
  previous_secret_ref VARCHAR(128) NOT NULL DEFAULT '',
  active_from DATETIME NULL,
  active_until DATETIME NULL,
  grace_until DATETIME NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  rotated_at DATETIME NULL,
  revoked_at DATETIME NULL,
  last_used_at DATETIME NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_webhook_secrets_ref (secret_ref),
  KEY idx_plugin_webhook_secrets_plugin_target (plugin_code, target_url),
  KEY idx_plugin_webhook_secrets_status (status),
  KEY idx_plugin_webhook_secrets_lookup (plugin_code, target_url, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS webhook_circuit_breakers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  target_url VARCHAR(512) NOT NULL,
  status ENUM('closed','open','half_open') NOT NULL DEFAULT 'closed',
  failure_count INT NOT NULL DEFAULT 0,
  success_count INT NOT NULL DEFAULT 0,
  opened_at DATETIME NULL,
  closed_at DATETIME NULL,
  next_probe_at DATETIME NULL,
  last_error_message VARCHAR(1000) NOT NULL DEFAULT '',
  last_failure_at DATETIME NULL,
  last_success_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_webhook_circuit_breakers_plugin_target (plugin_code, target_url),
  KEY idx_webhook_circuit_breakers_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin callback tokens for external plugin services (v1.7.7)
CREATE TABLE IF NOT EXISTS plugin_callback_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL,
  plugin_installation_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  publisher_id VARCHAR(128) NOT NULL DEFAULT '',
  token_ref VARCHAR(128) NOT NULL,
  token_hash VARCHAR(128) NOT NULL,
  name VARCHAR(255) NOT NULL DEFAULT '',
  status ENUM('active','disabled','revoked','expired') NOT NULL DEFAULT 'active',
  scopes_json JSON NULL,
  community_scope_json JSON NULL,
  expires_at DATETIME NULL,
  last_used_at DATETIME NULL,
  last_used_ip VARCHAR(64) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  rotated_at DATETIME NULL,
  revoked_at DATETIME NULL,
  revoked_reason VARCHAR(500) NOT NULL DEFAULT '',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_callback_tokens_ref (token_ref),
  UNIQUE KEY uk_plugin_callback_tokens_hash (token_hash),
  KEY idx_plugin_callback_tokens_plugin_status (plugin_code, status),
  KEY idx_plugin_callback_tokens_status_updated (status, updated_at),
  KEY idx_plugin_callback_tokens_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin callback request logs (v1.7.7)
CREATE TABLE IF NOT EXISTS plugin_callback_requests (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  request_id VARCHAR(64) NOT NULL,
  plugin_code VARCHAR(64) NOT NULL,
  token_ref VARCHAR(128) NOT NULL DEFAULT '',
  api_path VARCHAR(255) NOT NULL,
  method VARCHAR(16) NOT NULL,
  scope_required VARCHAR(64) NOT NULL DEFAULT '',
  status ENUM('accepted','rejected','failed') NOT NULL DEFAULT 'accepted',
  response_status INT NOT NULL DEFAULT 0,
  error_code VARCHAR(64) NOT NULL DEFAULT '',
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  actor_type VARCHAR(32) NOT NULL DEFAULT '',
  actor_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  ip_address VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(255) NOT NULL DEFAULT '',
  duration_ms BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_callback_requests_request_id (request_id),
  KEY idx_plugin_callback_requests_plugin_created (plugin_code, created_at),
  KEY idx_plugin_callback_requests_request_id (request_id),
  KEY idx_plugin_callback_requests_status_created (status, created_at),
  KEY idx_plugin_callback_requests_token_ref (token_ref)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Upgrade Tasks (基于 compat-check 的升级任务记录)
CREATE TABLE IF NOT EXISTS plugin_upgrade_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plugin_code VARCHAR(64) NOT NULL DEFAULT '',
  old_version VARCHAR(64) NOT NULL DEFAULT '',
  new_version VARCHAR(64) NOT NULL DEFAULT '',
  old_plugin_installation_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  new_package_download_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  new_package_precheck_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  new_package_compat_check_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(64) NOT NULL DEFAULT 'pending',
  previous_plugin_status VARCHAR(64) NOT NULL DEFAULT '',
  new_plugin_status VARCHAR(64) NOT NULL DEFAULT '',
  backup_path VARCHAR(500) NOT NULL DEFAULT '',
  old_install_path VARCHAR(500) NOT NULL DEFAULT '',
  new_install_path VARCHAR(500) NOT NULL DEFAULT '',
  manifest_diff_json JSON NULL,
  config_diff_json JSON NULL,
  permission_diff_json JSON NULL,
  menu_diff_json JSON NULL,
  route_diff_json JSON NULL,
  hook_diff_json JSON NULL,
  content_type_diff_json JSON NULL,
  migration_diff_json JSON NULL,
  impact_json JSON NULL,
  errors_json JSON NULL,
  warnings_json JSON NULL,
  rollback_log_json JSON NULL,
  failure_stage VARCHAR(64) NOT NULL DEFAULT '',
  failure_reason VARCHAR(1000) NOT NULL DEFAULT '',
  next_step VARCHAR(1000) NOT NULL DEFAULT '',
  reason VARCHAR(1000) NOT NULL DEFAULT '',
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  duration_ms BIGINT NOT NULL DEFAULT 0,
  requested_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plugin_upgrade_tasks_status_created (status, created_at),
  KEY idx_plugin_upgrade_tasks_plugin_version (plugin_code, new_version),
  KEY idx_plugin_upgrade_tasks_compat (new_package_compat_check_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Operation Snapshots (安装/升级保护与失败恢复)
CREATE TABLE IF NOT EXISTS plugin_operation_snapshots (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  operation_id VARCHAR(80) NOT NULL,
  operation_type VARCHAR(32) NOT NULL,
  plugin_code VARCHAR(64) NOT NULL,
  from_version VARCHAR(32) NOT NULL DEFAULT '',
  to_version VARCHAR(32) NOT NULL DEFAULT '',
  package_path VARCHAR(500) NOT NULL DEFAULT '',
  package_source VARCHAR(32) NOT NULL DEFAULT '',
  approval_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  before_plugin_json JSON NULL,
  before_manifest_json JSON NULL,
  before_config_json JSON NULL,
  before_config_version_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  before_migrations_json JSON NULL,
  before_permissions_json JSON NULL,
  before_menus_json JSON NULL,
  before_routes_json JSON NULL,
  before_dependencies_json JSON NULL,
  before_status VARCHAR(32) NOT NULL DEFAULT '',
  after_manifest_json JSON NULL,
  dry_run_json JSON NULL,
  risk_report_json JSON NULL,
  diff_json JSON NULL,
  checksum_summary_json JSON NULL,
  signature_summary_json JSON NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'created',
  error_code VARCHAR(80) NOT NULL DEFAULT '',
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  metadata_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_operation_snapshots_operation_id (operation_id),
  KEY idx_plugin_operation_snapshots_plugin (plugin_code, created_at),
  KEY idx_plugin_operation_snapshots_status (status, created_at),
  KEY idx_plugin_operation_snapshots_type (operation_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Plugin Trusted Publishers (Ed25519 公钥信任源)
CREATE TABLE IF NOT EXISTS plugin_trusted_publishers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  publisher_id VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL DEFAULT '',
  homepage VARCHAR(255) NOT NULL DEFAULT '',
  email VARCHAR(255) NOT NULL DEFAULT '',
  public_key_id VARCHAR(128) NOT NULL,
  public_key_algorithm VARCHAR(32) NOT NULL DEFAULT 'ed25519',
  public_key TEXT NOT NULL,
  fingerprint VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'trusted',
  notes VARCHAR(1000) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  revoked_at DATETIME NULL,
  blocked_at DATETIME NULL,
  expires_at DATETIME NULL,
  metadata_json JSON NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_trusted_publishers_key (publisher_id, public_key_id),
  KEY idx_plugin_trusted_publishers_status (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS plugin_remote_indexes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  source_id VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL DEFAULT '',
  index_url VARCHAR(500) NOT NULL,
  homepage VARCHAR(255) NOT NULL DEFAULT '',
  description VARCHAR(1000) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  trust_policy VARCHAR(32) NOT NULL DEFAULT 'readonly',
  last_fetch_status VARCHAR(32) NOT NULL DEFAULT '',
  last_fetch_at DATETIME NULL,
  last_error_code VARCHAR(128) NOT NULL DEFAULT '',
  last_error_message VARCHAR(1000) NOT NULL DEFAULT '',
  last_index_hash VARCHAR(128) NOT NULL DEFAULT '',
  metadata_json JSON NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plugin_remote_indexes_source (source_id),
  KEY idx_plugin_remote_indexes_status (status, updated_at),
  KEY idx_plugin_remote_indexes_fetch (last_fetch_status, last_fetch_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Hook Executions (内置插件 HookBus 执行记录)
CREATE TABLE IF NOT EXISTS hook_executions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  hook_name VARCHAR(128) NOT NULL,
  plugin_code VARCHAR(64) NOT NULL,
  service_type VARCHAR(32) NOT NULL DEFAULT '',
  endpoint_url VARCHAR(1000) NOT NULL DEFAULT '',
  mode VARCHAR(32) NOT NULL DEFAULT 'non_blocking',
  content_type VARCHAR(64) NOT NULL DEFAULT '',
  content_id BIGINT UNSIGNED NULL,
  community_id BIGINT UNSIGNED NULL,
  category_id BIGINT UNSIGNED NULL,
  actor_type VARCHAR(32) NOT NULL DEFAULT '',
  actor_id BIGINT UNSIGNED NULL,
  user_id BIGINT UNSIGNED NULL,
  admin_user_id BIGINT UNSIGNED NULL,
  request_id VARCHAR(128) NOT NULL DEFAULT '',
  started_at DATETIME NOT NULL,
  finished_at DATETIME NULL,
  duration_ms INT NOT NULL DEFAULT 0,
  success TINYINT(1) NOT NULL DEFAULT 1,
  error_message TEXT NULL,
  blocking TINYINT(1) NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT '',
  response_status INT NOT NULL DEFAULT 0,
  response_body_excerpt VARCHAR(2000) NOT NULL DEFAULT '',
  request_body_sha256 VARCHAR(128) NOT NULL DEFAULT '',
  error_code VARCHAR(128) NOT NULL DEFAULT '',
  metadata_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_hook_executions_plugin_hook (plugin_code, hook_name),
  KEY idx_hook_executions_success (plugin_code, success, started_at),
  KEY idx_hook_executions_content (content_id),
  KEY idx_hook_executions_community (community_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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

-- Core SecretCenter secret refs (v1.8.4-S14)
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

-- Categories (板块表) - 替代 boards，支持 content_type
CREATE TABLE IF NOT EXISTS categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0表示全站通用板块',
  name VARCHAR(64) NOT NULL,
  slug VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL DEFAULT 'article' COMMENT 'article/question/project/ai_work/job/wiki_page/document/news',
  plugin_code VARCHAR(64) NOT NULL DEFAULT 'core',
  allowed_content_types JSON NULL,
  description TEXT NULL,
  icon VARCHAR(100) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0,
  visible TINYINT NOT NULL DEFAULT 1,
  nav_visible TINYINT NOT NULL DEFAULT 1,
  postable TINYINT NOT NULL DEFAULT 1,
  seo_title VARCHAR(255) NOT NULL DEFAULT '',
  seo_description VARCHAR(500) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_categories_community_slug (community_id, slug),
  KEY idx_categories_community_type (community_id, type),
  KEY idx_categories_visible_sort (visible, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Topics (内容主题表) - 替代 posts，支持多种内容类型
CREATE TABLE IF NOT EXISTS topics (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL,
  category_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(220) NULL,
  plugin_code VARCHAR(64) NOT NULL DEFAULT 'core',
  content_type VARCHAR(32) NOT NULL DEFAULT 'article' COMMENT 'article/question/project/ai_work/job/wiki_page/document/news',
  summary VARCHAR(500) NULL,
  content MEDIUMTEXT NOT NULL,
  ai_summary TEXT NULL,
  cover_image VARCHAR(255) NULL,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 0隐藏 2审核中 3已删除',
  is_pinned TINYINT NOT NULL DEFAULT 0,
  is_featured TINYINT NOT NULL DEFAULT 0,
  is_solved TINYINT NOT NULL DEFAULT 0,
  comment_locked TINYINT NOT NULL DEFAULT 0,
  reject_reason TEXT NULL,
  offline_reason TEXT NULL,
  best_comment_id BIGINT UNSIGNED NULL,
  view_count INT UNSIGNED NOT NULL DEFAULT 0,
  comment_count INT UNSIGNED NOT NULL DEFAULT 0,
  like_count INT UNSIGNED NOT NULL DEFAULT 0,
  favorite_count INT UNSIGNED NOT NULL DEFAULT 0,
  hot_score INT UNSIGNED NOT NULL DEFAULT 0,
  last_active_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  KEY idx_topics_plugin_type_status (plugin_code, content_type, status),
  KEY idx_topics_community_type_status (community_id, content_type, status),
  KEY idx_topics_category_status (category_id, status),
  KEY idx_topics_hot_score (hot_score),
  KEY idx_topics_last_active (last_active_at),
  KEY idx_topics_pinned_featured (is_pinned, is_featured),
  FULLTEXT KEY ft_topics_search (title, summary, content)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- QA plugin tables
CREATE TABLE IF NOT EXISTS qa_questions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  topic_id BIGINT UNSIGNED NOT NULL,
  is_solved TINYINT NOT NULL DEFAULT 0,
  best_answer_id BIGINT UNSIGNED NULL,
  accepted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_qa_questions_topic (topic_id),
  KEY idx_qa_questions_solved (is_solved),
  CONSTRAINT fk_qa_questions_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS qa_answers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  topic_id BIGINT UNSIGNED NOT NULL,
  comment_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  is_accepted TINYINT NOT NULL DEFAULT 0,
  accepted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_qa_answers_comment (comment_id),
  KEY idx_qa_answers_topic (topic_id, is_accepted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Docs plugin tables
CREATE TABLE IF NOT EXISTS docs_spaces (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  description TEXT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_docs_spaces_community_slug (community_id, slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS docs_documents (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  space_id BIGINT UNSIGNED NULL,
  topic_id BIGINT UNSIGNED NOT NULL,
  parent_id BIGINT UNSIGNED NULL,
  sort_order INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_docs_documents_topic (topic_id),
  KEY idx_docs_documents_space_parent (space_id, parent_id, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Wiki plugin tables
CREATE TABLE IF NOT EXISTS wiki_spaces (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  description TEXT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_wiki_spaces_community_slug (community_id, slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS wiki_pages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  space_id BIGINT UNSIGNED NULL,
  topic_id BIGINT UNSIGNED NOT NULL,
  current_version_id BIGINT UNSIGNED NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_wiki_pages_topic (topic_id),
  KEY idx_wiki_pages_space (space_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS wiki_page_versions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  wiki_page_id BIGINT UNSIGNED NOT NULL,
  topic_id BIGINT UNSIGNED NOT NULL,
  editor_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  version_no INT NOT NULL DEFAULT 1,
  title VARCHAR(200) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  change_note VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_wiki_versions_page_no (wiki_page_id, version_no),
  KEY idx_wiki_versions_topic (topic_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Topic Tags (主题标签关联表)
CREATE TABLE IF NOT EXISTS topic_tags (
  topic_id BIGINT UNSIGNED NOT NULL,
  tag_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (topic_id, tag_id),
  KEY idx_topic_tags_tag (tag_id),
  CONSTRAINT fk_topic_tags_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
  CONSTRAINT fk_topic_tags_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Reactions (点赞表)
CREATE TABLE IF NOT EXISTS reactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'topic/comment/wiki',
  target_id BIGINT UNSIGNED NOT NULL,
  reaction_type VARCHAR(32) NOT NULL DEFAULT 'like',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_reactions_user_target (user_id, target_type, target_id, reaction_type),
  KEY idx_reactions_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Favorites (收藏表)
CREATE TABLE IF NOT EXISTS favorites (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'topic/wiki/project',
  target_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_favorites_user_target (user_id, target_type, target_id),
  KEY idx_favorites_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Follows (关注表)
CREATE TABLE IF NOT EXISTS follows (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'user/topic/community/tag',
  target_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_follows_user_target (user_id, target_type, target_id),
  KEY idx_follows_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Activities (动态表)
CREATE TABLE IF NOT EXISTS activities (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  community_id BIGINT UNSIGNED NULL,
  topic_id BIGINT UNSIGNED NULL,
  action VARCHAR(64) NOT NULL COMMENT 'created_topic/commented/liked/followed/favorited',
  target_type VARCHAR(32) NOT NULL,
  target_id BIGINT UNSIGNED NOT NULL,
  remark VARCHAR(500) NULL,
  metadata TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_activities_user_created (user_id, created_at),
  KEY idx_activities_community_created (community_id, created_at),
  KEY idx_activities_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Reports (举报表)
CREATE TABLE IF NOT EXISTS reports (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  reporter_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'topic/comment/user/wiki',
  target_id BIGINT UNSIGNED NOT NULL,
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  topic_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reason_type VARCHAR(64) NOT NULL,
  reason_text VARCHAR(1000) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/accepted/rejected',
  handled_by BIGINT UNSIGNED NULL,
  handled_at DATETIME NULL,
  handle_note VARCHAR(1000) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_reports_status (status),
  KEY idx_reports_target (target_type, target_id),
  KEY idx_reports_community_status (community_id, status),
  KEY idx_reports_reporter_target_status (reporter_id, target_type, target_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Community Moderators (子站版主表)
CREATE TABLE IF NOT EXISTS community_moderators (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  role VARCHAR(32) NOT NULL DEFAULT 'moderator' COMMENT 'moderator/owner',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_community_moderators (community_id, user_id),
  KEY idx_community_moderators_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

`
