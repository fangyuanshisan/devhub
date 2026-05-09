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
  sort_order INT NOT NULL DEFAULT 0,
  use_count INT UNSIGNED NOT NULL DEFAULT 0,
  follower_count INT UNSIGNED NOT NULL DEFAULT 0,
  seo_title VARCHAR(255) NOT NULL DEFAULT '',
  seo_description VARCHAR(500) NOT NULL DEFAULT '',
  seo_keywords VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tags_site_slug (site_key, slug),
  KEY idx_tags_site_status_sort (site_key, status, sort_order),
  KEY idx_tags_use_count (use_count)
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

-- Categories (板块表) - 替代 boards，支持 content_type
CREATE TABLE IF NOT EXISTS categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0表示全站通用板块',
  name VARCHAR(64) NOT NULL,
  slug VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL DEFAULT 'article' COMMENT 'article/question/project/ai_work/job/wiki/doc/news',
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
  content_type VARCHAR(32) NOT NULL DEFAULT 'article' COMMENT 'article/question/project/ai_work/job/wiki/doc/news',
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
  KEY idx_topics_community_type_status (community_id, content_type, status),
  KEY idx_topics_category_status (category_id, status),
  KEY idx_topics_hot_score (hot_score),
  KEY idx_topics_last_active (last_active_at),
  KEY idx_topics_pinned_featured (is_pinned, is_featured),
  FULLTEXT KEY ft_topics_search (title, summary, content)
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

-- Wiki Pages (Wiki 表，预留)
CREATE TABLE IF NOT EXISTS wiki_pages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL,
  category_id BIGINT UNSIGNED NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(220) NULL,
  summary VARCHAR(500) NULL,
  content MEDIUMTEXT NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  view_count INT UNSIGNED NOT NULL DEFAULT 0,
  like_count INT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  KEY idx_wiki_pages_community (community_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Wiki Revisions (Wiki 版本表，预留)
CREATE TABLE IF NOT EXISTS wiki_revisions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  wiki_page_id BIGINT UNSIGNED NOT NULL,
  editor_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(200) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  change_note VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_wiki_revisions_page (wiki_page_id),
  CONSTRAINT fk_wiki_revisions_page FOREIGN KEY (wiki_page_id) REFERENCES wiki_pages(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

`
