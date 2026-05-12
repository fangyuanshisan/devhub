# 插件迁移指南

DevHub 插件迁移用于追踪插件表结构和初始化数据的状态。当前阶段内置插件迁移以记录型 / no-op 确认为主，主 schema 仍负责创建内置插件表。

manifest + 配置型插件安装时可以声明 migration plan，系统会记录为待处理迁移，但当前不执行外部插件 raw SQL，也不会做 destructive migration。

## 命名规范

- 文件名建议：`001_plugin_feature.sql`。
- `migration_version` 使用插件版本或迁移批次，例如 `1.0.0`。
- `migration_name` 使用稳定蛇形名称，例如 `qa_questions`。
- 当前只支持 `direction = up`。

## 状态

- `pending`：待执行。
- `running`：执行中。
- `success`：执行成功。
- `failed`：执行失败，可重试。

成功迁移不得重复破坏数据。失败迁移必须保留错误信息，并阻断插件启用 / 恢复。

## 幂等要求

- 使用 `CREATE TABLE IF NOT EXISTS`。
- 添加字段前先检查字段是否存在。
- 不做破坏性删除。
- 不默认修改历史内容语义。

## 审计

迁移操作必须写入审计：

- `plugin.migration.run`
- `plugin.migration.success`
- `plugin.migration.failed`
- `plugin.migration.retry`

## MySQLStore 注意事项

- 新装库以 `db/mysql/001_schema.sql` 为准。
- 老库升级按 `db/mysql/migrations/` 编号顺序执行。
- 当前不引入外部迁移框架，执行前必须备份。

## 当前限制

- 不支持 migration down。
- 不支持真实硬回滚。
- 不做迁移前自动备份。
- 外部插件独立 Migration Runner 属于后续阶段。
- manifest + 配置型插件当前只记录迁移计划；真实外部 DDL runner、checksum 风险处理和依赖排序仍是后续能力。
