-- v1.3.4+ plugin lifecycle archive statuses.
-- This is additive: it only extends the plugins.status enum for old databases.

ALTER TABLE plugins
  MODIFY COLUMN status ENUM('discovered','installed','migrated','configured','enabled','disabled','running','archived','config_invalid','migration_pending','migration_failed','dependency_missing') NOT NULL DEFAULT 'enabled';
