# 飞书链接插件

这是 DevHub 声明型插件模板。它只包含 manifest、配置、文档和示例，不会被系统自动扫描或动态加载。

## 使用方式

1. 打开 manifest.json。
2. 将 JSON 复制到后台 /admin-next/plugins 的 Manifest 校验 / dry-run / install 流程。
3. 先执行 validate，再执行 dry-run；确认 impact、权限、菜单、Hook 和迁移声明后再 install。

## 当前目录内容

- manifest.json：插件声明。
- config.example.json：可用于全局或子站配置的示例配置。
- content-type.md：内容类型声明说明。
- permissions.md：权限声明说明。
- hooks.md：Hook 声明边界。
- migrations.md：迁移声明边界。
- docs/registry-example.md：内置系统插件接入示例说明，不会被动态加载。

## 如何定义插件能力

- 内容类型：编辑 manifest.json 的 content_types 与 content_type_definitions。
- 权限：编辑 permissions，权限码需要和内容类型中的 create/edit/delete/audit permission 对应。
- 菜单：编辑 menus；菜单展示仍会受插件状态、子站状态、scope 和 permission 共同过滤。
- 配置：编辑 config_schema，并用 config.example.json 验证示例配置。
- Hook：编辑 hooks；当前只是声明和治理元信息，或 external_service non-blocking 投递声明，不会加载第三方处理器。
- 迁移：编辑 migrations；当前只是声明示例，不执行外部 SQL。

## 安全边界

- 不执行动态代码。
- 不上传插件包。
- 不接入插件市场。
- 不执行外部 SQL。
- 不支持 migration down。
- 不执行远程 Webhook。
- 不动态加载前端资产。
- 不破坏 Core 表。
- 不绕过权限校验。
- 不影响历史内容访问。
- 不破坏 /topics/:id SEO。
- 不在示例配置中保存明文敏感字段。
