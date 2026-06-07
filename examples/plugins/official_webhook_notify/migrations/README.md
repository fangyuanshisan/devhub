# Migrations

本示例插件没有真实数据库迁移。

保留 `migrations/` 目录是为了符合 DevHub 插件包结构规范：如后续需要迁移，SQL 只能放在 `migrations/` 下；根目录 SQL、package scripts 和外部 SQL 均不会执行。
