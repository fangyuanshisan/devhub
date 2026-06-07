-- DevHub v1.3.2 patch: extend plugin lifecycle/status model.
-- This migration keeps enabled/disabled compatible while allowing P0 governance states.

ALTER TABLE plugins
  MODIFY COLUMN status ENUM('discovered','installed','migrated','configured','enabled','disabled','running','archived','config_invalid','migration_pending','migration_failed','dependency_missing') NOT NULL DEFAULT 'enabled';
