# Migration 声明

当前 migration 是内置 up/no-op 与记录型迁移边界。模板中的 migrations 仅为声明示例，不会执行外部 SQL。

当前不支持：

- 外部 raw SQL 执行。
- migration down。
- hard rollback。
- 自动备份。
- 外部插件迁移包。
