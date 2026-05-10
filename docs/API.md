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

### 全局插件 API

`GET /api/v1/plugins`

- 认证：不需要。
- 返回：只返回全局 `enabled` 插件。
- 用途：前台判断系统可用插件能力。
- 说明：当前返回的是内置系统插件统一 manifest 视图，包括内容类型、权限、菜单、路由声明与 `config_schema` 预留字段。

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
- 返回：全部注册插件，包括 `installed`、`enabled`、`disabled`。

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
  当前通过 `admin_logs.target` 记录 `plugin_code` 与 old/new 状态摘要。

常见错误：

- `400 {"error":"插件不存在"}`
- `400 {"error":"插件状态不合法"}`
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
- 返回：某个子站的插件列表，包含全局状态叠加后的子站状态、内容类型、菜单、权限、`config_json`。

`POST /api/v1/admin/communities/:id/plugins/:code/enable`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 规则：全局 disabled 插件不能被子站启用。
- 返回：更新后的插件对象。
- 审计：写入子站插件状态变更审计日志。
  当前通过 `admin_logs.target` 记录 `community_id`、`plugin_code` 与 old/new 状态摘要。

`POST /api/v1/admin/communities/:id/plugins/:code/disable`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 影响：只影响该子站的新发布、导航、菜单和管理入口，不影响历史内容访问。
- 返回：更新后的插件对象。
- 审计：写入子站插件状态变更审计日志。
  当前通过 `admin_logs.target` 记录 `community_id`、`plugin_code` 与 old/new 状态摘要。

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
- 审计：写入子站插件配置审计日志。
  当前通过 `admin_logs.target` 记录 `community_id`、`plugin_code` 与 old/new 配置摘要。

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
  当前通过 `admin_logs.target` 记录 `community_id` 与排序前后摘要。

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

## 插件声明约定

当前内置插件统一按 manifest 风格组织：

- `PluginManifest`
- `ContentTypeDefinition`
- `PermissionDefinition`
- `MenuDefinition`
- `RouteDefinition`
- `HookDefinition`

当前说明：

- 这是内置系统插件规范，不是插件市场，也不是第三方动态插件机制。
- `HookDefinition` 当前主要作为扩展点声明，不代表已经存在通用运行时 Hook 执行器。
- 配置优先级按“默认配置 -> 全局插件配置预留 -> 子站插件配置”约定，其中 v1.3.x 当前真实可写的是 `community_plugins.config_json`。

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

- 插件市场、插件包上传、远程插件安装、在线更新和 Go 动态插件加载。
- 发布时按插件权限码做细粒度拒绝的独立错误码和权限配置 API。
- Docs 文档树专用编辑 API 的完整形态。
- Wiki 版本历史、版本对比和回滚 API 的完整形态。
- 取消最佳答案 / 取消已解决状态接口。
- 标签趋势统计和标签运营分析 API。
