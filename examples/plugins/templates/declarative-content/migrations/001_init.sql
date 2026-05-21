-- DevHub declarative-content template migration.
-- dry-run only builds a plan; install/upgrade never executes root SQL or package scripts.
CREATE TABLE IF NOT EXISTS official_links_template_items (
  id BIGINT PRIMARY KEY,
  title VARCHAR(120) NOT NULL,
  url VARCHAR(500) NOT NULL,
  community_id BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
