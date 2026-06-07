-- DevHub v1.3.0 Core + Plugins compatibility migration.
-- The Go startup migration performs the same changes defensively; this file is
-- provided for operators who prefer applying SQL migrations explicitly.

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

SET @topics_plugin_code_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'topics' AND COLUMN_NAME = 'plugin_code'
);
SET @sql := IF(@topics_plugin_code_exists = 0,
  'ALTER TABLE topics ADD COLUMN plugin_code VARCHAR(64) NOT NULL DEFAULT ''core'' AFTER slug',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @categories_plugin_code_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'categories' AND COLUMN_NAME = 'plugin_code'
);
SET @sql := IF(@categories_plugin_code_exists = 0,
  'ALTER TABLE categories ADD COLUMN plugin_code VARCHAR(64) NOT NULL DEFAULT ''core'' AFTER type',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @categories_allowed_content_types_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'categories' AND COLUMN_NAME = 'allowed_content_types'
);
SET @sql := IF(@categories_allowed_content_types_exists = 0,
  'ALTER TABLE categories ADD COLUMN allowed_content_types JSON NULL AFTER plugin_code',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE topics SET content_type='document', plugin_code='docs' WHERE content_type='doc';
UPDATE topics SET content_type='wiki_page', plugin_code='wiki' WHERE content_type='wiki';
UPDATE topics SET plugin_code='qa' WHERE content_type='question';
UPDATE topics SET plugin_code='projects' WHERE content_type='project';
UPDATE topics SET plugin_code='jobs' WHERE content_type='job';
UPDATE topics SET plugin_code='ai_works' WHERE content_type='ai_work';

UPDATE categories SET type='document', plugin_code='docs', allowed_content_types=JSON_ARRAY('document','doc') WHERE type='doc' OR slug='docs';
UPDATE categories SET type='wiki_page', plugin_code='wiki', allowed_content_types=JSON_ARRAY('wiki_page','wiki') WHERE type='wiki' OR slug='wiki';
UPDATE categories SET plugin_code='qa', allowed_content_types=JSON_ARRAY('question') WHERE type='question';
UPDATE categories SET plugin_code='projects', allowed_content_types=JSON_ARRAY('project') WHERE type='project' OR slug='opensource';
UPDATE categories SET plugin_code='jobs', allowed_content_types=JSON_ARRAY('job') WHERE type='job' OR slug='jobs';
UPDATE categories SET plugin_code='ai_works', allowed_content_types=JSON_ARRAY('ai_work') WHERE type='ai_work' OR slug='ai';
UPDATE categories SET plugin_code='core', allowed_content_types=JSON_ARRAY(type) WHERE plugin_code='core' AND allowed_content_types IS NULL;

CREATE TABLE IF NOT EXISTS qa_questions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  topic_id BIGINT UNSIGNED NOT NULL,
  is_solved TINYINT NOT NULL DEFAULT 0,
  best_answer_id BIGINT UNSIGNED NULL,
  accepted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_qa_questions_topic (topic_id)
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
  UNIQUE KEY uk_qa_answers_comment (comment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
  UNIQUE KEY uk_docs_documents_topic (topic_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
  UNIQUE KEY uk_wiki_pages_topic (topic_id)
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
  UNIQUE KEY uk_wiki_versions_page_no (wiki_page_id, version_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
