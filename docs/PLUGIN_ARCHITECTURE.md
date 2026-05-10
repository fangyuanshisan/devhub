# DevHub 插件架构说明

[返回文档入口](README.md)

更新时间：2026-05-10

## 版本定位

`v1.3.x` 是 Core + Plugins 架构拆分与插件平台收口阶段。问答、文档、Wiki 不再作为 Core 硬编码业务类型描述，而是由 `qa`、`docs`、`wiki` 三个内置系统插件注册内容类型、菜单、权限和路由描述；`v1.3.1` 进一步封口旧写入口，并补强后台内容创建 / 更新时的插件权限边界。

DevHub 的长期目标不是只支持内置 `qa/docs/wiki`，而是形成完整插件平台。Core 只提供通用社区底座，业务能力都应通过插件声明、插件状态、插件权限、插件菜单、插件配置、插件 Hook、插件 migration、插件 API、插件 SEO、插件通知和插件搜索扩展。

完整插件系统是当前最高优先级长期主线。下一阶段完整目标、生命周期和验收标准以 [完整插件系统长期完善路线图](PLUGIN_SYSTEM_ROADMAP.md) 为准。插件市场、插件包、远程安装、在线更新和动态加载不在当前代码实现范围内，但不再作为永久排除项；它们进入后续阶段路线，并必须在安全边界和 SEO 红线内推进。

## Core 边界

Core 只保留通用社区能力：

- 用户、前台认证、后台认证、子站版主授权。
- 子站、板块、通用内容、评论、标签、搜索、通知、收藏、关注、举报、审计日志。
- 权限、后台 RBAC、版主 scope 校验。
- `/topics/:id`、子站页、标签页、sitemap 和 robots 的 SEO 兜底。
- 插件注册、插件状态、插件菜单、插件权限描述和基础分发。

当前兼容命名：

- `topics` 是当前通用内容表，对应目标架构中的 `contents`。
- `categories` 是当前通用板块表，对应目标架构中的 `channels`。

## 完整插件系统路线

本节是当前架构文档中的阶段摘要；更完整的目标流程、治理能力、后台能力、运行时能力、审计能力和 E2E 要求见 [完整插件系统长期完善路线图](PLUGIN_SYSTEM_ROADMAP.md)。

P0：插件平台收口

- Manifest 契约稳定。
- Registry 稳定。
- ActorContext 稳定。
- 权限码平台化。
- 全局插件状态。
- 子站插件状态。
- 板块绑定。
- 发布校验。
- 菜单过滤。
- `config_json`。
- `config_schema` 基础校验。
- HookBus 全调用点。
- `admin_logs` 结构化审计。
- migration 边界。
- 测试矩阵。

P1：插件平台增强

- schema 自动表单。
- 插件 SDK 文档。
- 插件生成模板。
- 插件依赖检查。
- 插件版本兼容检查。
- 插件事件和通知模板。
- 插件搜索索引扩展。
- 插件 SEO 扩展。

P2：插件分发能力

- 本地插件包。
- 插件安装。
- 插件升级。
- 插件禁用。
- soft uninstall。
- 插件 migration runner。
- 插件包签名校验。
- 插件市场雏形。

P3：高级能力

- 远程插件市场。
- 在线更新。
- 动态加载能力评估。
- 插件沙箱。
- 插件权限隔离。

安全红线：

- 禁用插件不能影响历史内容访问。
- 禁用插件不能破坏 `/topics/:id` SEO 动态 HTML。
- Core 表不能被插件随意破坏。
- 插件写操作必须走权限校验。
- ActorContext 必须由服务端生成，不能由客户端伪造。
- 前台 user token、后台 admin token、版主 user token + scope 不能混用。
- 插件配置必须至少保证 JSON 合法。
- 插件 migration 必须有备份和回滚说明。
- 未实现能力不能写成已完成。
- 预留、部分完成、待验收能力必须明确标注。

## 内置系统插件

- `qa`：问答插件，提供 `question` 内容类型，并承载问题、回答、采纳和已解决状态。
- `docs`：文档插件，提供 `document` 内容类型，并预留文档空间、文档树和文档详情能力。
- `wiki`：Wiki 插件，提供 `wiki_page` 内容类型，并预留页面版本、回滚和协作编辑能力。

`project`、`job`、`ai_work` 已按内置系统插件拆分为：

- `projects -> project`
- `jobs -> job`
- `ai_works -> ai_work`

历史 URL 和 `/topics/:id` SEO 输出保持不变；迁移只影响 `plugin_code` 归属和插件状态治理。

## Manifest 规范

当前内置系统插件统一使用一套 `PluginManifest` 风格的声明结构，不再允许每个插件各自定义一套字段。

统一字段包括：

- `code`
- `name`
- `version`
- `description`
- `is_system`
- `content_types`
- `content_type_definitions`
- `permissions`
- `menus`
- `routes`
- `config_schema`
- `dependencies`
- `min_core_version`
- `hooks`

说明：

- manifest 只描述能力和元数据，不直接承载业务执行流程。
- `qa/docs/wiki/projects/jobs/ai_works` 当前都通过统一 registry 返回相同结构。
- `config_schema` 当前是预留元数据，方便后续为 `plugins.config_json` 和 `community_plugins.config_json` 增加更稳定的 schema 校验与表单 UI。

## 内容类型声明

每个内容类型当前统一描述为 `ContentTypeDefinition`，至少包含：

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

当前内置插件映射：

- `question -> qa`
- `document -> docs`
- `wiki_page -> wiki`
- `project -> projects`
- `job -> jobs`
- `ai_work -> ai_works`

兼容归一仍集中在 registry：

- `doc -> document`
- `wiki -> wiki_page`

当前 Core 兼容内容类型主要是 `article`、`news` 等尚未拆分插件的通用内容，默认使用：

- `create_permission = core.topic.create`
- `edit_permission = post.update`
- `delete_permission = post.delete`

## 权限声明

插件权限当前统一描述为 `PermissionDefinition`，至少包含：

- `code`
- `name`
- `description`
- `scope`
- `plugin_code`

scope 语义当前约定为：

- `global`
- `community`
- `category`
- `own`

当前发布链路已接入最小权限码校验；更细粒度的权限矩阵仍是后续任务。

## 菜单声明

插件菜单当前统一描述为 `MenuDefinition`，至少包含：

- `code`
- `title`
- `path`
- `location`
- `permission`
- `plugin_code`
- `sort_order`

菜单展示必须经过以下过滤：

- 插件全局 `enabled`
- 插件在当前子站 `enabled`
- 当前用户具备菜单权限
- 当前请求作用域匹配（`admin` / `moderator` / `frontend`）

当前实现中：

- 后台左侧导航只保留“系统插件”入口
- 插件业务菜单通过系统插件列表或版主插件菜单返回

## HookBus 与 Hook

HookBus 完整化属于插件平台 P0 收口任务。当前只服务内置系统插件扩展点，不承载第三方动态执行；第三方插件执行、沙箱和动态加载进入 P3 评估，不是当前代码实现范围。

建议 Hook 名称统一为：

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

当前状态：

- `HookDefinition` 是 manifest 声明层，描述插件希望参与的扩展点。
- `v1.3.2` 起 HookBus 作为插件平台能力收口到 `internal/plugins`，并在 Service 的内容创建/更新/删除、评论创建、搜索、通知与 SEO 构建等流程中派发 Hook 事件。
- HookBus 当前仅注册内置系统插件 Hook handlers（编译期内置注册，不支持第三方动态加载）。
- 当前没有第三方动态注册，也没有插件包运行时加载；HookBus 仅服务内置系统插件和后续 Core 内部扩展。
- 搜索、通知和 SEO 当前是最小调用点：已能派发事件，但还没有复杂索引、通知模板或结构化数据插件处理器。
- 完整插件业务处理器、统一失败日志、重试策略和跨 Store 事务边界属于 P0/P1 继续收口项，不能降级为低优先级优化。

失败策略约定：

- 关键 Hook：`BeforeCreateContent`、`BeforeUpdateContent`、`BeforeDeleteContent` 失败会阻断当前操作；当前没有跨 Store 事务回滚封装，后续如 Hook 写外部资源需单独设计事务边界。
- 非关键 Hook：`AfterCreateContent`、`AfterUpdateContent`、`AfterDeleteContent`、`AfterCreateComment`、`OnSearchIndex`、`OnNotificationBuild`、`OnSEOBuild` 当前不阻断主流程；后续需要补统一日志和重试策略。

## 配置优先级

插件配置优先级当前定义为：

1. 默认配置 / `config_schema` 约定
2. `plugins.config_json` 全局配置
3. `community_plugins.config_json` 子站配置

子站配置优先级最高。

当前真实实现说明：

- `plugins.config_json` 已落库，并可通过后台插件页和 `PUT /api/v1/admin/plugins/:code/config` 管理。
- `community_plugins.config_json` 已落地，并可通过后台子站插件配置和 `PUT /api/v1/admin/communities/:id/plugins/:code/config` 管理。
- API 返回的 `resolved_config` 以 `default`、`global`、`community`、`effective` 四段表达当前合并视图。
- 当前已完成 JSON 合法性校验与简化 `config_schema` 基础校验；后台插件配置使用 JSON Editor + Ajv 做客户端校验，后端保存时仍会二次校验。后台自动表单渲染和更完整 JSON Schema 支持是 P1 任务。

## 两层插件状态

插件状态分两层：

- `plugins.status`：全局插件状态，表示插件是否在系统层面可用。
- `community_plugins.status`：子站插件状态，表示某个子站是否启用该插件。

状态值：

- `installed`：已安装 / 已注册，但未启用。
- `enabled`：已启用。
- `disabled`：已禁用。

插件在某个子站可用必须同时满足：

- 插件已注册。
- `plugins.status = enabled`。
- `community_plugins.status = enabled`。
- 当前板块 `categories.plugin_code` 匹配插件。
- 当前 `content_type` 在 `categories.allowed_content_types` 内。
- 当前用户具备对应权限。

特殊说明：

- `core` 是兼容内置能力，Service 层视为始终可用，不要求写入 `plugins` 或 `community_plugins`。
- 禁用全局插件会影响所有子站的新发布、导航、菜单和管理入口。
- 禁用子站插件只影响该子站的新发布、导航、菜单和后台管理入口。
- 禁用插件不影响历史内容访问，尤其不能破坏 `/topics/:id` SEO 动态 HTML。

## 发布校验流程

发布 Topic 时应走统一插件校验：

1. 解析 community。
2. 解析 category，并校验 category 属于 community。
3. 归一 `content_type`：`doc -> document`，`wiki -> wiki_page`。
4. 根据 `content_type` 推断 `plugin_code`：`question -> qa`，`document -> docs`，`wiki_page -> wiki`，`project -> projects`，`job -> jobs`，`ai_work -> ai_works`，其他 Core 兼容类型 -> `core`。
5. 校验全局插件状态。
6. 校验当前子站插件状态。
7. 校验 `category.plugin_code` 是否匹配。
8. 校验 `content_type` 是否在 `category.allowed_content_types` 内。
9. 校验当前用户权限。
10. 写入归一后的 `topics.content_type` 和 `topics.plugin_code`。

当前真实状态：

- 步骤 1-8 已在 `ValidateTopicPluginAccess` 和 Store 层板块校验中落地。
- 步骤 9 已接入最小权限码校验：
  - `question -> qa.question.create`
  - `document -> docs.document.create`
  - `wiki_page -> wiki.page.create`
  - `project -> projects.project.create`
  - `job -> jobs.job.create`
  - `ai_work -> ai_works.work.create`
  - Core 兼容类型 `article`、`news` 当前仍为粗粒度 `core.topic.create`（兼容旧 `post.create`）。
- `Service.CreateTopic` 是业务创建安全入口；`Service.CreatePost` 已封口，不再允许旧 posts 业务链路绕过插件校验。
- `repo.CreateTopic` / `repo.CreatePost` 属于仓储层裸写入或兼容能力，可以保留防御性归一，但不应被 HTTP / Service 常规业务链路当作权限入口。

## 后台内容更新边界

v1.3.1 采用稳妥策略：后台编辑已存在内容时禁止修改归属和内容类型。

禁止修改：

- `community_id` / `site`
- `category_id` / `board`
- `content_type`
- `plugin_code`

允许修改：

- 标题、摘要、正文、标签、状态、置顶、精华和 SEO 等非归属字段。

原因：

- 跨子站、跨板块或跨插件迁移需要同时校验目标子站、目标板块、全局插件状态、子站插件状态、`allowed_content_types`、当前用户权限和历史扩展表一致性。
- 当前后台编辑不承担迁移职责；后续如需要，应新增迁移专项 API，而不是复用普通编辑接口。

## 兼容权限桥

`core.topic.create` 是 Core 兼容内容类型的当前创建权限。为了兼容历史后台和老角色配置，当前 `post.create` 仍可作为 `core.topic.create` 的过渡兼容权限。

当前口径：

- `post.create` 不是长期主权限。
- 插件内容类型必须使用自己的 create 权限，例如 `qa.question.create`、`docs.document.create`、`wiki.page.create`。
- `article` / `news` 后续要么明确作为 Core 内容定义继续存在，要么拆为插件，再逐步移除 `post.create` 兼容桥。

## 数据结构

插件相关表：

- `plugins`
- `community_plugins`
- `qa_questions`
- `qa_answers`
- `docs_spaces`
- `docs_documents`
- `wiki_spaces`
- `wiki_pages`
- `wiki_page_versions`

增强字段：

- `topics.plugin_code`
- `categories.plugin_code`
- `categories.allowed_content_types`

## API 与菜单

已注册的插件 API 包括：

- `GET /api/v1/plugins`
- `GET /api/v1/communities/:slug/plugins`
- `GET /api/v1/admin/plugins`
- `POST /api/v1/admin/plugins/:code/enable`
- `POST /api/v1/admin/plugins/:code/disable`
- `PUT /api/v1/admin/plugins/:code/config`
- `GET /api/v1/admin/plugin-menus`
- `GET /api/v1/admin/communities/:id/plugins`
- `POST /api/v1/admin/communities/:id/plugins/:code/enable`
- `POST /api/v1/admin/communities/:id/plugins/:code/disable`
- `PUT /api/v1/admin/communities/:id/plugins/:code/config`
- `PUT /api/v1/admin/communities/:id/plugins/sort`
- `GET /api/v1/moderator/plugin-menus`

菜单策略：

- 后台左侧导航只保留“系统插件”入口。
- 插件业务管理页通过系统插件列表进入，避免 qa / docs / wiki 直接散落到左侧导航。
- 版主插件菜单必须同时满足全局 enabled、子站 enabled、当前用户是该子站版主、当前用户具备菜单权限。

后台插件管理体验：

- `/admin-next/plugins` 展示全局插件列表、状态 badge、系统插件标识、内容类型、权限数量、菜单数量和 `config_schema` 摘要。
- 插件详情使用抽屉分区展示基础信息、内容类型、权限、菜单、配置、路由和 Hooks，避免把 JSON 直接堆在表格中。
- 全局插件配置已升级为 JSON Editor（`json-editor-vue`），并使用 Ajv 做 `config_schema` 基础校验；后续仍可增强为更完整的 schema 强校验与自动表单渲染。
- `/admin-next/communities` 的子站插件配置抽屉展示全局状态和子站状态双 badge，并支持子站启用 / 禁用、`config_json` 编辑、JSON 格式化、数字排序和禁用原因提示。
- 全局禁用和子站禁用都有二次确认，并明确 disabled 只影响新发布、导航、菜单和管理入口，不影响历史内容详情页和 SEO。
- 插件影响范围统计已提供轻量 impact 计数接口：
  - `GET /api/v1/admin/plugins/:code/impact`
  - `GET /api/v1/admin/communities/:id/plugins/:code/impact`
  UI 在接口不可用时必须显示“待接口支持/暂不可用”，不得伪造数字。
  同时插件详情抽屉提供“审计”Tab，复用 `GET /api/v1/admin/audit-logs`，按 `target=plugins#<code>` 前缀筛选展示。

## 当前限制与阶段边界

- 插件市场、插件包上传、本地/远程安装、在线更新和 Go 动态插件加载不是当前 P0 代码实现范围；它们分别进入 P2 / P3 路线，后续推进时必须满足安全红线、权限隔离、migration 备份回滚和 SEO 不退化。
- 插件路由当前是注册描述 + Core 分发；动态路由加载和动态执行环境进入 P2 / P3 路线评估。
- Docs / Wiki 的专用编辑体验仍是部分完成。
- 子站插件配置和排序已有 API 与增强后的后台 UI，但仍需继续做真实浏览器矩阵验收。
- 插件治理审计已新增 `admin_logs.old_value`、`admin_logs.new_value` 和 `admin_logs.metadata_json` 结构化字段，同时保留 `target` 文本摘要兼容旧展示；非插件历史日志可能仍没有结构化 diff。
- 新装库已在 `db/mysql/001_schema.sql` 和 `internal/store/schema.go` 包含结构化审计字段；老库升级使用 `db/mysql/migrations/007_admin_logs_structured_plugin_audit.sql`，启动迁移辅助也会尝试补齐这些列。
- `plugins.config_json` 与 `community_plugins.config_json` 已可写，但当前仅做 JSON 格式校验；`config_schema` 基础校验属于 P0，自动表单渲染属于 P1。
- HookBus 当前是最小内部调度器；调用点已覆盖内容创建、更新、删除、评论、搜索、通知和 SEO，但搜索 / 通知 / SEO 仍是预留级事件派发，完整业务处理器和日志策略属于 P0/P1。
