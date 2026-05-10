# DevHub 插件架构说明

[返回文档大纲](README.md)

更新时间：2026-05-10

## 版本定位

`v1.3.0` 是 Core + Plugins 架构拆分版。Core 不再把问答、文档、Wiki 作为硬编码核心类型处理，而是通过内置系统插件注册内容类型、菜单、权限和路由描述。

## 系统目标

- Core 只处理通用社区能力，并提供“插件注册与分发”基础设施。
- Plugins 负责扩展内容类型与专属业务能力；Core 通过 `plugin_code + content_type` 分发与约束发布。

## Core 边界

Core 保留通用能力：

- 用户、前台认证、后台认证、子站版主授权。
- 子站、板块、通用内容、评论、标签、搜索、通知、收藏、关注、举报、审计日志。
- 后台 RBAC、版主 scope 校验、SEO 动态页和 sitemap / robots。
- 插件注册、插件状态、插件菜单、插件权限描述。

当前物理表兼容历史命名：

- `topics` 是 Core 内容表，对应需求中的 `contents`。
- `categories` 是 Core 板块表，对应需求中的 `channels`。

## 插件边界

内置插件目录：

- `internal/plugins/qa`：注册 `qa` 插件和 `question` 内容类型。
- `internal/plugins/docs`：注册 `docs` 插件和 `document` 内容类型。
- `internal/plugins/wiki`：注册 `wiki` 插件和 `wiki_page` 内容类型。

插件状态：

- `installed`：已安装但未启用。
- `enabled`：已启用，可发布对应内容。
- `disabled`：已禁用，对应板块不能继续发布新内容。

插件状态分层（全局 + 子站）：

- `plugins.status`：系统层插件状态，决定插件是否全局可用。
- `community_plugins.status`：子站层插件状态，决定某个子站是否启用该插件。
- 只有当插件同时满足“全局 enabled + 子站 enabled”时，才能在该子站绑定板块、展示菜单、发布新内容。
- 禁用只影响新发布与入口展示，不影响已有内容 `/topics/:id` 的访问与 SEO。

插件生命周期（概念层）：

- `registered`：代码内置定义层的插件（`internal/plugins/*`），不一定落库。
- `installed` / `enabled` / `disabled`：运行时状态（落库到 `plugins` 表），覆盖静态定义。

插件定义结构（静态定义 + 运行时覆盖）：

- `code` / `plugin_code`：插件唯一标识（当前两者等价，优先以 `code` 为准）。
- `name` / `version` / `description`：展示信息。
- `status`：运行时状态。
- `content_types`：插件拥有的内容类型（例如 `qa -> question`）。
- `menus` / `permissions` / `routes`：后台/版主菜单、权限码和路由描述。

## 发布校验

`POST /api/v1/topics` 会执行以下校验：

- 当前 `category` 必须属于目标 `community`。
- `content_type` 先归一：历史 `doc` / `wiki` 会归一为 `document` / `wiki_page`。
- 根据 `content_type` 判断 `plugin_code`（例如 `question -> qa`，否则为 `core`）。
- `category.plugin_code` 必须匹配 `content_type` 对应插件（历史空值兼容为 `core`）。
- 插件必须全局 `enabled`，且当前子站 `community_plugins` 也为 `enabled`（禁用后只限制新发布，不影响已有内容阅读与 SEO）。
- `content_type` 必须在 `category.allowed_content_types` 中（允许 legacy alias）。
- 通过后写入 `topics.content_type`（归一后）与 `topics.plugin_code`。

前台发布页接入策略：

- 内容类型选择会按“启用插件 + 当前子站板块 `allowed_content_types`”收口；禁用插件的内容类型不会出现在下拉中。

## 数据表

新增：

- `plugins`
- `community_plugins`
- `qa_questions`
- `qa_answers`
- `docs_spaces`
- `docs_documents`
- `wiki_spaces`
- `wiki_pages`
- `wiki_page_versions`

增强：

- `topics.plugin_code`
- `categories.plugin_code`
- `categories.allowed_content_types`

## 当前限制

- 暂未做插件市场、压缩包上传安装、远程更新。
- 插件路由当前以注册描述和 Core 分发为主，没有引入独立运行时加载器。
- 文档树、Wiki 协作编辑和版本回滚已具备表结构与基础注册，完整专用编辑 UI 留到后续迭代。
- admin-next 侧边栏默认只展示“系统插件”入口；插件业务管理页（qa/docs/wiki）通过系统插件列表进入，避免插件菜单散落在左侧导航中。
