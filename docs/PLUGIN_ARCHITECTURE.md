# DevHub 插件架构说明

[返回文档入口](README.md)

更新时间：2026-05-10

## 版本定位

`v1.3.1` 是 Core + Plugins 架构拆分后的技术债收口版。问答、文档、Wiki 不再作为 Core 硬编码业务类型描述，而是由 `qa`、`docs`、`wiki` 三个内置系统插件注册内容类型、菜单、权限和路由描述；本版本进一步封口旧写入口，并补强后台内容创建 / 更新时的插件权限边界。

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

## HookBus 与 Hook 预留

当前只定义内置插件扩展点，不做第三方动态执行机制，也不做 Go 动态插件加载。

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
- `Service` 已有最小内部 `HookBus`，当前调用点包括 `BeforeCreateContent`、`AfterCreateContent` 和 `AfterCreateComment`。
- 当前没有第三方动态注册，也没有插件包运行时加载；HookBus 仅服务内置系统插件和后续 Core 内部扩展。
- 目前尚未接入搜索索引、通知构建和 SEO 构建的完整 Hook 调度。

失败策略约定：

- 关键 Hook：失败应阻断主流程并由调用方回滚，例如创建前/更新前校验类 Hook。
- 非关键 Hook：失败记录日志，不阻断主流程，例如 SEO、搜索索引、通知构建类 Hook。

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
- 当前只校验 JSON 格式，暂不做 `config_schema` 强校验。

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

## 当前限制

- 当前阶段不做插件市场。
- 当前阶段不做插件包上传。
- 当前阶段不做远程插件下载或在线更新。
- 当前阶段不做 Go 动态插件加载。
- 插件路由当前是注册描述 + Core 分发，不是真正动态运行时路由加载。
- Docs / Wiki 的专用编辑体验仍是部分完成。
- 子站插件配置和排序已有 API 与最小后台 UI，但仍需继续做浏览器矩阵验收。
- 插件治理审计当前通过 `admin_logs.target` 记录 `plugin_code`、`community_id` 与 old/new 摘要；尚未拆出独立 JSON diff 字段。
- `plugins.config_json` 与 `community_plugins.config_json` 已可写，但当前仅做 JSON 格式校验，尚未做 schema 强校验。
- HookBus 当前是最小内部调度器；尚未覆盖 Update/Delete/Search/Notification/SEO 等完整调用点，也尚未形成独立结构化日志策略。
