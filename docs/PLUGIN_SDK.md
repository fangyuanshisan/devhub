# DevHub Plugin SDK

[返回文档入口](README.md)

更新时间：2026-05-12

本文档说明 DevHub 当前插件声明契约、安装校验流程和安全边界。它面向“声明型 / 配置型插件骨架”，不是第三方动态运行时 SDK。

## 插件开发边界

当前支持：

- 内置系统插件：随 DevHub 源码编译发布，通过代码 registry 注册。
- Manifest + 配置型插件：通过 `manifest.json` 描述内容类型、权限、菜单、路由、配置模型、Hook 声明和 migration 声明；可走后台 validate / dry-run / install / upgrade 流程。
- 本地插件包 dry-run 导入预览：按 `docs/PLUGIN_PACKAGE.md` 规范扫描目录、校验与预览（不安装、不执行代码/SQL、不动态加载前端资产）。

当前不支持：

- 插件市场。
- zip 插件包上传。
- 远程安装或在线更新。
- Go 动态加载。
- JS / WASM / Lua 脚本沙箱。
- 第三方本地代码执行。
- 第三方动态 Hook 处理器。
- 远程 Webhook 执行。

`HookDefinition` 当前是声明和内部 HookBus 治理边界，不是第三方动态 Hook 运行时。

## 生命周期状态

- `discovered`：系统识别到插件声明；当前主要是路线状态，不代表可发布。
- `installed`：manifest + 配置型插件已记录；默认不启用，不允许新建内容。
- `disabled`：插件已禁用；不允许新建内容，不影响历史内容访问和 `/topics/:id` SEO。
- `enabled`：全局插件可用；具体子站还需要 `community_plugins.status=enabled` 才能发布对应内容。
- `archived`：插件已软卸载 / 归档；禁止新建内容和子站启用，保留历史内容、配置、迁移记录、审计记录和 SEO。
- `config_invalid`：配置未通过 `config_schema` 校验；不应放行启用或发布。
- `migration_pending`：存在待处理迁移；当前内置 no-op pending 不阻断启用，但会在健康治理中提示。
- `migration_failed`：存在失败迁移；阻断全局启用和子站启用。
- `dependency_missing`：依赖插件缺失或未启用；阻断 readiness。
- `hook_warning`：存在 Hook 失败记录但未达到错误阈值。
- `hook_error`：Hook 失败达到当前轻量阈值。

## Manifest 字段

- `code`：插件唯一编码，只能使用小写字母、数字、下划线，并以字母开头。
- `name`：插件名称。
- `version`：语义化版本，例如 `1.0.0`。
- `description`、`author`、`homepage`、`license`：展示和治理元信息。
- `min_core_version`：声明最低 Core 版本。
- `compatible_core_version`：当前 validator 支持的兼容范围字段，例如 `>=1.4.0`。
- `is_system`：是否内置系统插件；外部声明型模板应为 `false`。
- `content_types`：内容类型列表，当前也兼容内容类型对象数组。
- `content_type_definitions`：内容类型完整定义。
- `permissions`：权限声明。
- `menus`：admin / moderator / frontend 菜单声明。
- `routes`：声明型路由元信息，不代表动态加载页面或后端 handler。
- `config_schema`：简化配置模型，用于后端保存校验和后台自动表单。
- `default_config`：当前 `PluginManifest` 结构未落地该字段；默认值应写在 `config_schema.properties.*.default`。
- `dependencies`：依赖声明数组，兼容旧字符串数组；对象字段包括 `code`、`version`、`required`、`reason`。required 默认 `true`。
- `hooks`：Hook 声明。
- `migrations`：migration 声明；当前只支持 up/no-op 记录型边界。
- `assets`：资产路径声明数组；当前仅校验安全路径，不动态加载前端资产。
- `external_service`：预留字段；当前不执行远程 Webhook。


## 依赖与版本兼容

`dependencies` 推荐结构：

```json
[
  {
    "code": "qa",
    "version": ">=1.0.0",
    "required": true,
    "reason": "需要问答内容类型作为数据来源"
  },
  {
    "code": "docs",
    "version": ">=1.0.0 <2.0.0",
    "required": false,
    "reason": "启用后可关联文档内容"
  }
]
```

支持字段：`code`、`version`、`required`、`reason`。旧格式 `["qa"]` 仍兼容，并按 required 处理。

版本约束只支持数字 `x.y.z`：精确版本、`>=1.2.0`、`>1.2.0`、`<=1.2.0`、`<2.0.0`、空格组合范围（如 `>=1.2.0 <2.0.0`）。当前不支持 `^`、`~`、`||`、预发布标签或 npm 完整语法。

依赖检查状态：`satisfied`、`missing`、`disabled`、`archived`、`migration_failed`、`config_invalid`、`version_mismatch`、`circular_dependency`、`self_dependency`、`optional_missing`。required 不满足会阻断 validate / dry-run / install / upgrade dry-run / upgrade / enable；optional 缺失只 warning。自依赖和循环依赖当前一律阻断，包括 optional 循环。

Core 版本兼容：`min_core_version` 缺失会 warning；高于当前 Core 时阻断。`compatible_core_version` 存在但当前 Core 不满足时阻断。当前 Core 版本来自项目 `VERSION`。

后台安装向导、升级向导和插件详情 Dependencies 区域会展示依赖总览、required / optional、当前版本、要求版本、状态、阻断原因、Core 兼容状态和依赖变更 diff；不会自动安装依赖，也不会远程下载依赖。

## 内容类型声明

`ContentTypeDefinition` 至少包含：

- `type`
- `name`
- `plugin_code`
- `create_permission`
- `edit_permission`
- `delete_permission`
- `audit_permission`
- `default_status`
- `allow_comment`
- `allow_like`
- `allow_favorite`
- `seo_type`

发布链路仍由后端校验插件全局状态、子站插件状态、板块绑定、`allowed_content_types` 和 `create_permission`。前端隐藏入口不能替代后端强校验。

## 权限声明

推荐命名：

- `{plugin_code}.{resource}.create`
- `{plugin_code}.{resource}.edit`
- `{plugin_code}.{resource}.delete`
- `{plugin_code}.{resource}.audit`
- `{plugin_code}.manage`
- `{plugin_code}.configure`

权限码用于内容创建、后台治理、版主菜单和配置入口。所有写操作必须由后端按权限码校验。

## 菜单声明

菜单可声明到：

- admin menu
- moderator menu
- frontend menu

菜单显示必须同时受以下条件影响：

- 插件全局状态。
- 子站插件状态。
- community scope。
- category scope。
- permission code。

菜单声明只决定入口元信息，不授予权限。

### 前台菜单声明字段（v1.4.0-P1-11）

前台菜单建议按统一结构声明（旧 manifest 缺字段保持兼容；未声明前台入口的插件不会自动暴露入口）：

- `code`：菜单唯一编码（可选）。
- `title`：显示标题（必填）。
- `description`：说明文案（可选）。
- `plugin_code`：来源插件（必填/由系统补齐）。
- `route`：前台跳转路由（可选；缺省使用 `path`）。
- `path`：兼容字段（必填；必须以 `/` 开头）。
- `location`：入口位置（建议值：`global_nav` / `community_nav` / `category_nav` / `user_center` / `create_menu` / `sidebar`）。
- `content_type`：关联内容类型（可选；用于分类绑定和创建入口治理）。
- `permission`：所需权限码（可选；只影响入口展示，后端仍会强校验）。
- `require_login`：是否要求登录（可选）。
- `require_community_enabled`：是否要求子站已启用插件（可选；默认建议为 true）。
- `require_category_binding`：是否要求当前子站已绑定对应内容类型板块（可选；要求为 true 时必须设置 `content_type`）。
- `visible_when`：可见性条件（可选数组；支持 `plugin_enabled` / `community_enabled` / `dependency_satisfied` / `config_valid`）。
- `order`：排序权重（可选；缺省回退到 `sort_order`）。
- `icon` / `badge`：展示增强字段（可选）。

限制与安全边界：

- 菜单声明仅影响“入口是否显示/是否可点击”，不绕过后端鉴权、依赖、迁移、配置等强校验。
- 插件 `disabled/archived/dependency_missing/config_invalid/migration_failed` 时必须隐藏“新建入口”；历史内容访问与 `/topics/:id/` SEO 兜底不受影响。

## config_schema

当前后端简化校验支持：

- `type`
- `required`
- `enum`
- `object`
- `boolean`
- `string`
- `number`
- `integer`
- `array`
- `items`
- `properties`
- `min` / `max`
- `default`
- `additionalProperties`

unknown fields 策略：object 默认拒绝未知字段；只有 `additionalProperties: true` 时允许。

后台配置体验：

- 表单模式：基于 `config_schema` 自动生成基础表单。
- JSON 高级模式：保留原始 JSON 编辑能力。
- `resolved_config.default/global/community/effective`：展示默认配置、全局配置、子站覆盖配置和最终生效配置。
- 敏感字段：后台表单和 diff 做展示脱敏；这不是加密存储。

当前限制：

- 不是完整 JSON Schema。
- 深层嵌套只做基础体验，不是复杂配置建模器。
- 字段分组只影响 UI。
- 不支持配置版本回滚。
- 不支持灰度配置。
- 不做敏感字段加密存储。

## Hook 声明

当前已知 Hook 名称：

- `BeforeCreateContent`
- `AfterCreateContent`
- `BeforeUpdateContent`
- `AfterUpdateContent`
- `BeforeDeleteContent`
- `AfterDeleteContent`
- `AfterCreateComment`
- `OnSearchIndex`
- `OnNotificationBuild`
- `OnSEOBuild`

治理语义：

- blocking Hook 失败会阻断主流程。
- non-blocking Hook 失败不阻断主流程。
- 两者都会写入 `hook_executions` 和审计。
- blocking 失败写入 `plugin.hook.blocked`。
- non-blocking 失败写入 `plugin.hook.failed`。

当前不支持第三方动态 Hook 处理器，也不支持远程 Webhook 执行。

## Migration 声明

当前 migration 是内置 up/no-op 与记录型迁移边界。模板可以生成 `migrations` 示例声明，但不会执行外部 SQL。

当前不支持：

- 外部 raw SQL 执行。
- migration down。
- hard rollback。
- 自动备份。
- 外部插件迁移包。

## 安装与校验流程

- `POST /api/v1/admin/plugins/manifest/validate`：校验 manifest，不写入数据，不执行代码；返回 `dependency_summary`、逐项 `dependencies` 和 `compatibility`。
- `POST /api/v1/admin/plugins/dry-run`：预览安装影响，不写入数据，不执行代码；required 依赖不满足时返回 blocked / invalid。
- `POST /api/v1/admin/plugins/install`：写入 manifest + 配置型插件记录，初始为 installed / disabled，不执行第三方代码；required 依赖或 Core 兼容不满足时拒绝。
- `POST /api/v1/admin/plugins/:code/upgrade/dry-run`：预览升级差异，不写入数据，不执行代码；返回依赖新增 / 删除 / 版本约束变化和 Core 兼容矩阵。
- `POST /api/v1/admin/plugins/:code/upgrade`：更新 manifest / version / checksum，保留历史内容、配置、迁移和审计，不执行第三方代码；required 新依赖不满足、Core 不兼容、降级或同版本重复升级会被拒绝。

这些接口均使用后台 admin token；写入类接口需要插件写权限。
