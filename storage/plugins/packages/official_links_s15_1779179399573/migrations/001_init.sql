-- DevHub S15 declarative plugin fixture migration.
-- dry-run 不执行 SQL；install 只允许基于 migrations/ 计划处理。
CREATE TABLE IF NOT EXISTS official_links_items_s15_1779179399573 (
  id BIGINT PRIMARY KEY,
  title VARCHAR(120) NOT NULL,
  url VARCHAR(500) NOT NULL,
  community_id BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
