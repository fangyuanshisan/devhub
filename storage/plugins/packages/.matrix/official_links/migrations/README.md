# official_links migrations

`official_links` 是 DevHub 官方声明型内容插件，业务数据复用 Core 的 `topics` / `categories` / `plugin_configs` / `admin_logs`：

- `topics.content_type = friend_link`
- `topics.plugin_code = official_links`
- `topics.title` 对应友情链接标题
- `topics.content` 对应链接 URL
- `topics.summary` 对应描述
- `topics.status` 对应发布 / 隐藏 / 下线状态
- `topics.pinned` / `topics.recommended` 可作为排序和推荐治理标记

本目录是唯一迁移入口。`001_init.sql` 当前只作为 no-op 迁移计划文件，说明 `official_links` 使用 Core 内容表；当前版本不需要插件私有业务表，不包含根目录 SQL，不执行 package scripts。dry-run 只生成计划，不执行 SQL。
