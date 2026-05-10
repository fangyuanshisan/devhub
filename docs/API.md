# DevHub API 文档

[返回文档入口](README.md)

更新时间：2026-05-10

本文档只记录当前仓库真实可用 API。接口路径以 `internal/transport/httpapi/router.go` 为准；未实现能力集中放在“规划 / 未完成”小节，不写入当前真实 API 主体。

## 通用规则

- API 前缀：`/api/v1`。
- 认证方式：`Authorization: Bearer <access_token>`。
- 前台用户 token：`token_type=user`，用于发帖、评论、关注、举报、用户中心和 `/api/v1/moderator/*`。
- 后台管理员 token：`token_type=admin`，用于 `/api/v1/admin/*`。
- 错误响应：`{"error":"错误信息"}`。
- 分页参数：`page`、`page_size`，默认按接口实现处理，建议 `page_size <= 50`。

## 插件 API

说明：后台全局插件管理页、子站插件配置抽屉和版主插件菜单均继续使用本节现有接口；本轮后台体验增强未新增插件 API，也未改变返回字段语义。

### 全局插件 API

`GET /api/v1/plugins`

- 认证：不需要。
- 返回：只返回全局 `enabled` 插件。
- 用途：前台判断系统可用插件能力。
- 说明：当前返回的是内置系统插件统一 manifest 视图，包括内容类型、权限、菜单、路由声明与 `config_schema` 预留字段。
- 安全处理：公共接口不返回 `config_json` 和 `resolved_config`。

响应示例：

```json
{
  "items": [
    {
      "code": "qa",
      "plugin_code": "qa",
      "name": "问答",
      "version": "1.0.0",
      "status": "enabled",
      "content_types": ["question"]
    }
  ]
}
```

`GET /api/v1/admin/plugins`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 返回：全部注册插件，包括 `installed`、`enabled`、`disabled`、`config_schema`、`config_json` 和 `resolved_config`。

`POST /api/v1/admin/plugins/:code/enable`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 路径参数：`code` 为插件 code，例如 `qa`、`docs`、`wiki`。
- 返回：更新后的插件对象。
- 审计：写入插件状态变更审计日志。

`POST /api/v1/admin/plugins/:code/disable`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 影响：禁用全局插件后，所有子站都不能继续发布该插件内容；已有内容详情不应受影响。
- 返回：更新后的插件对象。
- 审计：写入插件状态变更审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

常见错误：

- `400 {"error":"插件不存在"}`
- `400 {"error":"插件状态不合法"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

`PUT /api/v1/admin/plugins/:code/config`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 请求：

```json
{
  "config_json": {
    "example": true
  }
}
```

- 返回：更新后的插件对象，包含 `config_json` 和 `resolved_config`。
- 校验：当前只校验 `config_json` 是合法 JSON，暂不做 `config_schema` 强校验。
- 审计：写入插件全局配置审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

常见错误：

- `400 {"error":"插件不存在"}`
- `400 {"error":"config_json 必须是合法 JSON"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

### 前台子站插件展示 API

`GET /api/v1/communities/:slug/plugins`

- 认证：不需要。
- 路径参数：`slug` 为子站 slug。
- 返回：当前子站可用插件，只包含“全局 enabled + 子站 enabled”的插件。
- 安全处理：不返回后台配置字段 `config_json`，也不暴露 `global_status` / `community_status`。
- 用途：前台子站首页、发布页和导航收口。

常见错误：

- `404 {"error":"子站不存在"}`
- `400 {"error":"..."}`：读取子站插件状态失败。

### 后台子站插件 API

`GET /api/v1/admin/communities/:id/plugins`

- 认证：后台 admin token。
- 权限：`site.read`，并经过子站管理范围校验。
- 返回：某个子站的插件列表，包含全局状态叠加后的子站状态、内容类型、菜单、权限、`config_json` 和 `resolved_config`。

`POST /api/v1/admin/communities/:id/plugins/:code/enable`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 规则：全局 disabled 插件不能被子站启用。
- 返回：更新后的插件对象。
- 审计：写入子站插件状态变更审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

`POST /api/v1/admin/communities/:id/plugins/:code/disable`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 影响：只影响该子站的新发布、导航、菜单和管理入口，不影响历史内容访问。
- 返回：更新后的插件对象。
- 审计：写入子站插件状态变更审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

`PUT /api/v1/admin/communities/:id/plugins/:code/config`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 请求：

```json
{
  "config_json": {
    "example": true
  }
}
```

- 返回：更新后的插件对象。
- 校验：当前只校验 `config_json` 是合法 JSON，暂不做 `config_schema` 强校验。
- 审计：写入子站插件配置审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

`PUT /api/v1/admin/communities/:id/plugins/sort`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 请求：

```json
{
  "codes": ["qa", "docs", "wiki"]
}
```

- 返回：

```json
{
  "updated": 3
}
```

- 审计：写入子站插件排序审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

常见错误：

- `400 {"error":"插件不存在"}`
- `400 {"error":"插件全局未启用，不能在子站启用"}`
- `400 {"error":"插件状态不合法"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`
- `404 {"error":"子站不存在"}`

### 插件菜单 API

`GET /api/v1/admin/plugin-menus`

- 认证：后台 admin token。
- 返回：全局 enabled 插件的 `admin` 菜单，并按当前后台用户权限过滤。

`GET /api/v1/moderator/plugin-menus`

- 认证：前台 user token。
- 授权：当前用户必须是启用状态子站版主。
- 查询参数：可选 `community_slug` 或 `community_id`。不传时按当前用户可治理子站范围汇总去重。
- 返回：同时满足全局 enabled、子站 enabled、当前用户有权限的 `moderator` 插件菜单。

常见错误：

- `401 {"error":"未登录"}`
- `403 {"error":"当前用户不是启用状态子站版主"}`
- `403 {"error":"无权管理该子站内容"}`

## 发布 API 与插件校验

`POST /api/v1/topics`

- 认证：前台 user token。
- 请求字段：`community_id` / `community_slug`、`category_id`、`content_type`、`title`、`summary`、`content`、`tags` 等。
- 写入：保存归一后的 `topics.content_type` 和 `topics.plugin_code`。

插件校验流程：

1. 解析 community。
2. 解析 category，并校验 category 属于 community。
3. 归一 `content_type`：`doc -> document`，`wiki -> wiki_page`。
4. 根据内容类型推断插件：`question -> qa`，`document -> docs`，`wiki_page -> wiki`，`project -> projects`，`job -> jobs`，`ai_work -> ai_works`，其他 Core 兼容类型 -> `core`。
5. 校验全局插件是否 enabled。
6. 校验当前子站插件是否 enabled。
7. 校验 `category.plugin_code` 是否匹配。
8. 校验 `content_type` 是否在 `category.allowed_content_types` 内。
9. 校验当前用户是否具备内容类型对应的发布权限码。

常见错误：

- `400 {"error":"内容类型不能为空"}`
- `400 {"error":"内容类型不合法"}`
- `400 {"error":"插件全局未启用"}`
- `400 {"error":"当前子站未启用该插件"}`
- `400 {"error":"板块不存在"}`
- `400 {"error":"当前板块未绑定对应插件"}`
- `400 {"error":"内容类型与板块不匹配"}`

当前限制：

- 发布链路已校验插件状态、子站插件状态和板块约束。
- 发布链路已增加插件权限码校验：
  - `question -> qa.question.create`
  - `document -> docs.document.create`
  - `wiki_page -> wiki.page.create`
  - `project -> projects.project.create`
  - `job -> jobs.job.create`
  - `ai_work -> ai_works.work.create`
  - `article`、`news` 当前仍使用粗粒度 `core.topic.create`（兼容旧 `post.create`）。
- 当前内容类型权限码来自统一 `ContentTypeDefinition` 声明，而不是散落在各个 handler 中。
- `projects/jobs/ai_works` 当前已完成插件归属、发布校验、权限码和菜单声明；专属扩展表与完整业务流程仍是后续任务。

`POST /api/v1/admin/posts`

- 认证：后台 admin token。
- 基础权限：`post.create`，作为历史后台内容创建兼容权限。
- 动态权限：接口会先将请求转换为 Topic 创建请求，归一 `content_type` 并推断 `plugin_code`，再叠加校验真实内容类型对应的插件 create 权限。
- 写入链路：内部调用 `Service.CreateTopic`，继续执行全局插件状态、子站插件状态、板块绑定、`allowed_content_types` 和发布权限码校验。

动态 create 权限：

- `question -> qa.question.create`
- `document -> docs.document.create`
- `wiki_page -> wiki.page.create`
- `project -> projects.project.create`
- `job -> jobs.job.create`
- `ai_work -> ai_works.work.create`
- `article/news -> core.topic.create`，并兼容旧 `post.create`

常见错误：

- `403 {"error":"缺少权限 qa.question.create，不能创建该类型内容"}`
- `403 {"error":"缺少权限 docs.document.create，不能创建该类型内容"}`
- `403 {"error":"缺少权限 wiki.page.create，不能创建该类型内容"}`
- `400 {"error":"插件全局未启用"}`
- `400 {"error":"当前子站未启用该插件"}`
- `400 {"error":"内容类型与板块不匹配"}`

说明：`admin/posts` 是后台兼容入口，不是绕过插件体系的独立写入口。`post.create` 只是第一层兼容基础权限，真实内容类型权限由插件权限码决定。

`PUT /api/v1/admin/posts/:id`

- 认证：后台 admin token。
- 权限：`post.update`，并经过当前后台用户的子站治理范围校验。
- 当前策略：禁止修改内容归属和内容类型。
- 允许更新：标题、摘要、正文、标签、状态、置顶、精华等非归属字段。
- 禁止更新：`site/community_id`、`board/category_id`、`content_type`、`plugin_code`。

常见错误：

- `400 {"error":"后台编辑不允许修改内容归属子站，请通过迁移专项处理"}`
- `400 {"error":"后台编辑不允许修改内容板块或内容类型，请通过迁移专项处理"}`

说明：如果后续需要后台迁移内容子站、板块或插件类型，应新增迁移专项接口，并逐条校验插件全局状态、子站插件状态、板块绑定、`allowed_content_types` 和对应插件权限码。

## 插件声明约定

当前内置插件统一按 manifest 风格组织：

- `PluginManifest`
- `ContentTypeDefinition`
- `PermissionDefinition`
- `MenuDefinition`
- `RouteDefinition`
- `HookDefinition`

当前说明：

- 这是当前内置系统插件规范；完整插件系统是最高优先级长期主线。插件市场、插件包、远程安装、在线更新和动态加载进入 P2 / P3 路线，但不是当前真实 API。
- `HookDefinition` 当前是扩展点声明；`Service` 已有最小内部 HookBus，当前调用点覆盖 `BeforeCreateContent`、`AfterCreateContent`、`BeforeUpdateContent`、`AfterUpdateContent`、`BeforeDeleteContent`、`AfterDeleteContent`、`AfterCreateComment`、`OnSearchIndex`、`OnNotificationBuild` 和 `OnSEOBuild`。
- Search / Notification / SEO 当前只是最小事件派发，尚未实现复杂索引、通知模板或结构化 SEO 插件处理器。
- 配置优先级按“默认配置 -> `plugins.config_json` -> `community_plugins.config_json`”合并；API 用 `resolved_config.default/global/community/effective` 表达合并视图。
- 当前配置已完成 JSON 合法性校验；`config_schema` 基础校验是 P0 插件平台任务，后台自动表单渲染是 P1 插件平台任务。

## 当前真实 API 索引

认证：

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
POST /api/v1/admin/login
POST /api/v1/admin/refresh
POST /api/v1/admin/logout
GET  /api/v1/admin/me
```

子站与板块：

```http
GET /api/v1/communities
GET /api/v1/communities/:slug
GET /api/v1/communities/:slug/home
GET /api/v1/communities/:slug/overview
GET /api/v1/communities/:slug/stats
GET /api/v1/communities/:slug/categories
GET /api/v1/communities/:slug/tags
GET /api/v1/communities/:slug/moderators
```

Topic 与搜索：

```http
GET    /api/v1/topics
GET    /api/v1/topics/:id
GET    /api/v1/topics/:id/qa
GET    /api/v1/topics/:id/docs
GET    /api/v1/topics/:id/wiki/versions
POST   /api/v1/topics
PUT    /api/v1/topics/:id
DELETE /api/v1/topics/:id
GET    /api/v1/search/topics
```

评论与问答：

```http
GET  /api/v1/topics/:id/comments
POST /api/v1/topics/:id/comments
POST /api/v1/topics/:id/comments/:commentId/replies
POST /api/v1/topics/:id/comments/:commentId/accept
POST /api/v1/topics/:id/solve
```

插件扩展只读接口：

- `GET /api/v1/topics/:id/qa`
  - 仅适用于 `question`
  - 返回 `qa_questions` 扩展状态和 `qa_answers` 列表
- `GET /api/v1/topics/:id/docs`
  - 仅适用于 `document`
  - 返回 `docs_documents` 扩展行和当前文档空间的基础文档树
- `GET /api/v1/topics/:id/wiki/versions`
  - 仅适用于 `wiki_page`
  - 返回 `wiki_pages` 扩展行和 `wiki_page_versions` 列表

标签：

```http
GET    /api/v1/tags
GET    /api/v1/tags/hot
GET    /api/v1/tags/suggestions
GET    /api/v1/tags/suggest
GET    /api/v1/tags/by-slug/:tag
GET    /api/v1/tags/:tag
GET    /api/v1/tags/:tag/topics
GET    /api/v1/communities/:slug/tags/:tag
GET    /api/v1/communities/:slug/tags/:tag/topics
GET    /api/v1/admin/tags
GET    /api/v1/admin/tags/:id
GET    /api/v1/admin/tags/:id/topics
GET    /api/v1/admin/tags/:id/aliases
POST   /api/v1/admin/tags
PUT    /api/v1/admin/tags/:id
POST   /api/v1/admin/tags/:id/aliases
DELETE /api/v1/admin/tags/:id/aliases/:aliasId
POST   /api/v1/admin/tags/:id/enable
POST   /api/v1/admin/tags/:id/disable
POST   /api/v1/admin/tags/:id/merge
POST   /api/v1/admin/tags/:id/recalculate
POST   /api/v1/admin/tags/recalculate-all
```

互动与用户中心：

```http
POST /api/v1/topics/:id/like
POST /api/v1/topics/:id/favorite
GET  /api/v1/topics/:id/interaction
POST /api/v1/actions/toggle
POST /api/v1/reactions/toggle
POST /api/v1/favorites/toggle
POST /api/v1/follows/toggle
GET  /api/v1/me/favorites
GET  /api/v1/me/follows
GET  /api/v1/me/activities
GET  /api/v1/me/notifications
POST /api/v1/me/notifications/read-all
POST /api/v1/me/notifications/:id/read
```

举报、治理和审计：

```http
POST /api/v1/reports
GET  /api/v1/admin/reports
GET  /api/v1/admin/reports/:id
POST /api/v1/admin/reports/:id/handle
POST /api/v1/admin/reports/batch-handle
GET  /api/v1/admin/comments
POST /api/v1/admin/comments/batch
GET  /api/v1/admin/audit-logs
```

`GET /api/v1/admin/audit-logs` 返回的插件治理日志会包含结构化字段：

- `old_value`：操作前 JSON diff，例如旧状态或旧配置。
- `new_value`：操作后 JSON diff，例如新状态或新配置。
- `metadata_json`：操作上下文，例如 `plugin_code`、`community_id`、`operation`。

说明：非插件历史日志可能仍只有 `target` 文本摘要。

版主工作台：

```http
GET  /api/v1/moderator/communities
GET  /api/v1/moderator/plugin-menus
GET  /api/v1/moderator/dashboard
GET  /api/v1/moderator/reports
POST /api/v1/moderator/reports/:id/handle
GET  /api/v1/moderator/topics
POST /api/v1/moderator/topics/:id/feature
POST /api/v1/moderator/topics/:id/unfeature
POST /api/v1/moderator/topics/:id/pin
POST /api/v1/moderator/topics/:id/unpin
POST /api/v1/moderator/topics/:id/hide
POST /api/v1/moderator/topics/:id/restore
POST /api/v1/moderator/topics/:id/lock-comments
POST /api/v1/moderator/topics/:id/unlock-comments
GET  /api/v1/moderator/comments
POST /api/v1/moderator/comments/:id/hide
POST /api/v1/moderator/comments/:id/restore
GET  /api/v1/moderator/audit-logs
```

兼容 API：

```http
GET    /api/v1/sites
GET    /api/v1/sites/:site
GET    /api/v1/sites/:site/overview
GET    /api/v1/boards
GET    /api/v1/posts
GET    /api/v1/posts/:id
```

说明：

- `POST/PUT/DELETE /api/v1/posts*` 写接口已废弃，当前返回 `410 Gone`，请使用 `/api/v1/topics`。
- `Service.CreatePost` 已在业务层封口，不再直接调用仓储裸写入；`repo.CreatePost` 仅作为 legacy / seed / migration 或兼容层内部能力保留，不作为业务写入口。

SEO 端点：

```http
GET /topics/:id/
GET /c/:slug/
GET /tags/:tag/
GET /c/:slug/tags/:tag/
GET /sitemap.xml
GET /robots.txt
```

## 规划 / 未完成

以下内容不是当前真实可用 API：

- P0：`config_schema` 基础校验 API / 错误结构。
- P0/P1：HookBus 业务处理器、统一错误日志、插件搜索 / 通知 / SEO 扩展 API。
- P1：插件 SDK / 开发规范、插件生成模板、插件依赖检查、插件版本兼容检查。
- P2：本地插件包、插件安装、插件升级、soft uninstall、插件 migration runner、插件包签名校验和插件市场雏形。
- P3：远程插件市场、在线更新、动态加载能力评估、插件沙箱和插件权限隔离。
- 插件权限配置 API、按子站 / 板块维护细粒度权限矩阵，以及 Core 兼容类型 `article/news` 的长期权限收口。
- Docs 文档树专用编辑 API 的完整形态。
- Wiki 版本历史、版本对比和回滚 API 的完整形态。
- 取消最佳答案 / 取消已解决状态接口。
- 标签趋势统计和标签运营分析 API。
