-- DevHub v1.3.2 patch: extend plugin lifecycle/status model.
-- This migration keeps enabled/disabled compatible while allowing P0 governance states.

ALTER TABLE plugins
  MODIFY COLUMN status ENUM('discovered','installed','migrated','configured','enabled','disabled','running','config_invalid','migration_pending','dependency_missing') NOT NULL DEFAULT 'enabled';
