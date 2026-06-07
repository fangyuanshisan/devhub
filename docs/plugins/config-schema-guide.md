

# config_schema 开发指南

DevHub 插件配置以 `config_schema` 声明结构，并由后台表单 / JSON 高级模式展示。保存配置时以后端 schema 校验为准，前端校验只用于提前提示。

## 当前支持

- `type`：`object`、`string`、`number`、`integer`、`boolean`、`array`。
- `required`：对象必填字段。
- `enum`：枚举值。
- `minimum` / `maximum`：数值范围。
- `default`：默认值，用于生成默认配置。
- `title`：后台表单字段名。
- `description`：后台表单字段说明。
- `sensitive`：敏感字段展示和 diff 脱敏提示。

## 配置层级

```text
插件默认配置 < 全局插件配置 < 子站插件配置 = effective_config
```

- 插件默认配置来自 `config_schema.properties.*.default`。
- 全局配置来自 `plugins.config_json`。
- 子站配置来自 `community_plugins.config_json`。
- 子站配置优先级最高。

## 敏感字段

以下字段应被视为敏感：

- `sensitive: true`
- `format: "password"`
- 字段名包含 `token`、`password`、`secret`、`key`

敏感字段在配置差异、审计展示和后台详情中应脱敏。后续阶段再评估加密存储。

## 当前限制

- 当前不是完整 JSON Schema 实现，不支持复杂 `oneOf`、`anyOf`、条件校验和字段分组。
- 自动表单为后台体验能力，最终安全边界仍是后端 schema 校验。
- 配置版本、配置回滚、灰度配置和敏感字段加密仍是后续规划。
