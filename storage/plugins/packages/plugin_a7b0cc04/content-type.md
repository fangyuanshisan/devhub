# 内容数据类型

`plugin_a7b0cc04` 是声明型内容插件，业务数据复用 Core 内容表：

- `topics.plugin_code = plugin_a7b0cc04`
- `topics.content_type = feishu_link`

manifest 中同时声明：

- `content_types: ["feishu_link"]`
- `content_type_definitions[0].type = feishu_link`
- 创建权限：`plugin_a7b0cc04.link.create`
- 治理权限：`plugin_a7b0cc04.link.audit`

发布链路要求插件全局启用、目标子站启用插件，并且目标板块绑定 `plugin_code=plugin_a7b0cc04` / `content_type=feishu_link` 或允许 `allowed_content_types=["feishu_link"]`。
