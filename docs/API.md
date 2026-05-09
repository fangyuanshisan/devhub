# DevHub API 文档

[返回文档大纲](README.md)

更新时间：2026-05-09

本文档记录当前仓库真实实现。若需求描述与代码不一致，以本文档中的实际路径、字段和完成度为准。

后续 Codex / AI Agent 更新 API 前，应先阅读 `docs/AGENT_RULES.md`，再核对 `internal/transport/httpapi/router.go`。接口存在但字段尚未展开时，可以标注“待补充字段细节”；不要把计划中的接口写成已完成。

## v1.2.0 API 范围

v1.2.0 API 按以下模块归档：Auth、Communities、Community stats、Community categories、Community moderators、Topics、Comments、Search、Tags、Reactions、Favorites、Follows、Activities、Notifications、Reports、Governance、Moderator Workspace、Moderators、Admin communities、Admin categories、Admin tags、Audit logs、SEO endpoints 和 Compatibility APIs。本版本新增标签详情、标签内容聚合、标签建议、后台标签 CRUD、标签启用 / 禁用和标签关联内容接口。

## 模块索引

- Auth：前台登录、注册、刷新、退出、当前用户。
- Communities：子站列表、详情、概览、统计、板块、标签、公开版主。
- Categories：子站板块读取和后台板块管理。
- Topics：Topic 列表、详情、发布、编辑、删除。
- Comments：评论列表、评论、回复、问答采纳。
- Search：Topic 搜索、筛选、未解决、精华。
- Tags：标签列表、热门标签、子站标签、标签详情、标签内容聚合、标签建议、标签关注。
- Reactions：点赞和兼容 reaction toggle。
- Favorites：收藏和我的收藏。
- Follows：关注 user/community/tag/topic。
- Activities：动态和我的动态。
- Notifications：通知、已读、全部已读。
- Reports：举报创建、举报管理、单条 / 批量处理。
- Governance：精华、置顶、隐藏、恢复、评论锁定、评论隐藏、批量治理。
- Moderator Workspace：独立版主工作台 API，使用前台 `users` token 和 `community_moderators` scope。
- Moderators：后台版主管理和子站范围权限。
- Admin：后台登录、概览、内容、用户、角色、权限、设置、通知。
- Audit Logs：治理审计日志。
- SEO Endpoints：`/topics/:id`、`/c/:slug`、sitemap、robots。
- Compatibility APIs：`sites/posts` 兼容 API。

## 通用规则

- API 前缀：`/api/v1`。
- 认证方式：支持 `Authorization: Bearer <access_token>`。
- 错误响应：`{"error":"错误信息"}`。
- 分页参数：`page`、`page_size`，`page_size` 最大为 `50`。
- `sites/posts` 兼容 API 继续保留，不作为本轮互动和评论闭环的主接口。
- 后台接口需要后台 admin token，或明确允许的子站版主 user token，并继续按 `community_id` / 子站范围校验。

## 身份与认证边界

DevHub v1.1.1 明确三类身份：

- `users`：前台社区用户。用于注册、登录、发帖、评论、点赞、收藏、关注、举报、我的动态和我的通知。
- `admin_users`：后台人员。用于登录 `/admin-next`、后台管理、全局治理、子站配置、用户管理、系统配置和审计日志。
- `community_moderators`：子站版主授权表。版主本质仍是 `users`，通过 `community_id + user_id + status=1` 获得指定子站治理权限。

Token / storage 规则：

- 前台 access token：`token_type=user`，`aud=devhub_frontend`。
- 后台 access token：`token_type=admin`，`aud=devhub_admin`。
- 前台推荐 localStorage key：`devhub_user_token`、`devhub_user_refresh_token`；当前前台兼容读取旧 `devhub_access_token`、`devhub_refresh_token`。
- 后台 sessionStorage key：`devhub_admin_token`、`devhub_admin_refresh_token`、`devhub_admin_user`。
- 前台 token 不能直接访问特权后台接口；后台 token 不能被 `/api/v1/auth/me` 或前台写接口当作 `users` 身份。
- `/api/v1/moderator/*` 只接受前台 user token，并要求当前 `users.id` 在 `community_moderators` 中有启用授权；普通用户返回 403。
- MemoryStore 仍使用 demo seed 用户支持本地开发；该策略只作为开发兜底，不是生产权限规则。

## Communities / Categories / Tags / Topics

社区与板块：

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

`GET /api/v1/communities` 和 `GET /api/v1/communities/:slug` 返回 v1.1.0 增强字段：

```json
{
  "id": 1,
  "name": "PHP",
  "slug": "php",
  "logo": "PHP",
  "cover_image": "",
  "slogan": "PHP 工程实践与 Laravel 生态",
  "description": "PHP 子站简介",
  "theme_color": "#4f46e5",
  "seo_title": "PHP 技术社区",
  "seo_description": "PHP 子站 SEO 描述",
  "seo_keywords": "PHP,Laravel,Hyperf,Swoole",
  "sort_order": 1,
  "status": 1,
  "follower_count": 20,
  "topic_count": 100,
  "comment_count": 200,
  "hot_score": 500,
  "announcement_title": "公告标题",
  "announcement_content": "公告内容",
  "announcement_url": "/topics/1/"
}
```

子站统计：

```http
GET /api/v1/communities/:slug/stats
```

响应：

```json
{
  "topic_count": 100,
  "comment_count": 200,
  "question_count": 30,
  "unsolved_count": 5,
  "follower_count": 20,
  "today_topic_count": 3,
  "today_comment_count": 8,
  "moderator_count": 2,
  "hot_score": 500
}
```

子站公开版主：

```http
GET /api/v1/communities/:slug/moderators
```

返回启用状态版主，包含 `community_id/community_slug/community_name/user_id/user_name/user_nickname/role/status`。

子站板块：

```http
GET /api/v1/communities/:slug/categories
```

返回字段包含 `content_type`、`visible`、`nav_visible`、`postable`、`seo_title`、`seo_description`、`status`。前台子站页只展示启用且可见的板块导航。

Topic：

```http
GET    /api/v1/topics
GET    /api/v1/topics/:id
POST   /api/v1/topics
PUT    /api/v1/topics/:id
DELETE /api/v1/topics/:id
```

标签能力：

```http
GET /api/v1/tags
GET /api/v1/tags/hot
GET /api/v1/tags/suggestions
GET /api/v1/tags/:tag
GET /api/v1/tags/:tag/topics
GET /api/v1/communities/:slug/tags
```

说明：

- 普通 Topic 列表和搜索默认过滤隐藏 / 删除内容。
- `POST /api/v1/topics` 发布后，`/topics/:id` 可立即由 Go 动态 SEO 页面访问，不需要重新前端构建。
- v1.2.0 已完成标签详情 SEO 页和后台标签管理；标签合并、标签别名和标签趋势统计仍是后续规划。

标签详情响应包含：

```json
{
  "id": 1,
  "site": "php",
  "community_id": 1,
  "community_slug": "php",
  "community_name": "PHP",
  "name": "Laravel",
  "slug": "laravel",
  "description": "Laravel 相关内容",
  "status": "enable",
  "sort_order": 1,
  "topic_count": 12,
  "follower_count": 3,
  "seo_title": "Laravel 相关内容",
  "seo_description": "DevHub Laravel 标签聚合",
  "seo_keywords": "Laravel,PHP"
}
```

标签建议：

```http
GET /api/v1/tags/suggestions?community_slug=php&q=lar&limit=20
```

返回当前子站启用标签，用于发布页选择。发布 Topic 最多 5 个标签，后端会校验标签属于当前子站。

标签内容聚合：

```http
GET /api/v1/tags/laravel/topics?community_slug=php&sort=latest&page=1&page_size=12
```

`sort` 支持 `latest`、`hot`、`active`、`featured`。

## v1.1.0 已完成子站接口

以下接口是 v1.1.0 子站增强已落地能力；字段细节以上文真实接口说明为准：

- `GET /api/v1/communities/:slug/stats`：子站统计。
- `GET /api/v1/communities/:slug/moderators`：公开子站版主。
- 后台 communities CRUD：`GET/POST/GET by id/PUT /api/v1/admin/communities`。
- 后台 community categories CRUD：`GET/POST /api/v1/admin/communities/:id/categories`、`PUT /api/v1/admin/categories/:id`。
- 后台启用 / 禁用 / 排序子站和板块。

## 认证

```http
POST /api/v1/auth/login
POST /api/v1/auth/register
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

登录成功返回 `access_token`、`refresh_token` 和 `user`。MemoryStore 注册当前返回演示会话；MySQLStore 会创建普通用户。

前台认证说明：

- `/api/v1/auth/login` 校验 `users` 并返回前台 user token。
- `/api/v1/auth/refresh` 只刷新 `token_type=user` 的 refresh token。
- `/api/v1/auth/me` 只接受前台 user token。
- 前台写操作必须使用 users 身份，包括发帖、评论、点赞、收藏、关注、举报和我的通知已读。

## Admin 基础接口

```http
POST /api/v1/admin/login
POST /api/v1/admin/refresh
POST /api/v1/admin/logout
GET  /api/v1/admin/me
GET  /api/v1/admin/overview
GET  /api/v1/admin/communities
POST /api/v1/admin/communities
GET  /api/v1/admin/communities/:id
PUT  /api/v1/admin/communities/:id
POST /api/v1/admin/communities/:id/enable
POST /api/v1/admin/communities/:id/disable
POST /api/v1/admin/communities/reorder
GET  /api/v1/admin/communities/:id/categories
POST /api/v1/admin/communities/:id/categories
PUT  /api/v1/admin/categories/:id
POST /api/v1/admin/categories/:id/enable
POST /api/v1/admin/categories/:id/disable
POST /api/v1/admin/categories/reorder
GET  /api/v1/admin/sites
POST /api/v1/admin/sites
PUT  /api/v1/admin/sites/:site
GET  /api/v1/admin/users
PUT  /api/v1/admin/users/:id/status
GET  /api/v1/admin/roles
GET  /api/v1/admin/permissions
GET  /api/v1/admin/tags
GET  /api/v1/admin/tags/:id
GET  /api/v1/admin/tags/:id/topics
POST /api/v1/admin/tags
PUT  /api/v1/admin/tags/:id
POST /api/v1/admin/tags/:id/enable
POST /api/v1/admin/tags/:id/disable
GET  /api/v1/admin/settings
PUT  /api/v1/admin/settings
GET  /api/v1/admin/logs
POST /api/v1/admin/notifications
```

后台认证说明：

- `/api/v1/admin/login` 校验 `admin_users` 并返回后台 admin token。
- `/api/v1/admin/refresh` 只刷新 `token_type=admin` 的 refresh token。
- `/api/v1/admin/*` 默认需要后台 admin token。
- 已启用的子站版主可以使用前台 user token 进入允许的治理类后台接口，但会被限制在 `community_moderators` 授权的子站 scope 内。
- 普通前台用户不能访问 `/api/v1/admin/*`。
- 版主不能管理后台人员、系统设置、版主分配、全局子站新增 / 排序等管理员能力。

后台种子账号见 `README.md`。后台接口需要 `Authorization: Bearer <access_token>`，并按 RBAC permission 和站点 scope 校验。

### Admin Tags

```http
GET  /api/v1/admin/tags?site=php&status=all&q=lar
POST /api/v1/admin/tags
GET  /api/v1/admin/tags/:id
PUT  /api/v1/admin/tags/:id
GET  /api/v1/admin/tags/:id/topics
POST /api/v1/admin/tags/:id/enable
POST /api/v1/admin/tags/:id/disable
```

新增 / 更新请求体：

```json
{
  "site": "php",
  "name": "Laravel",
  "slug": "laravel",
  "description": "Laravel 生态、实践和线上问题复盘。",
  "status": "enable",
  "sort_order": 1,
  "seo_title": "Laravel 相关内容",
  "seo_description": "DevHub Laravel 标签聚合，汇总相关文章、问答、项目和文档。",
  "seo_keywords": "Laravel,PHP,框架"
}
```

规则：

- `site` 使用子站 slug；`portal` 表示总站标签。
- `status=enable` 启用，`status=disable` 禁用。
- 禁用标签不能被公开详情 API 读取，不进入 sitemap。
- `GET /api/v1/admin/tags/:id/topics` 返回该标签关联的 Topic，用于后台查看标签关联内容。
- 新增、更新、启用和禁用标签写入 `admin_logs`。

### Admin Communities

```http
GET  /api/v1/admin/communities
POST /api/v1/admin/communities
GET  /api/v1/admin/communities/:id
PUT  /api/v1/admin/communities/:id
POST /api/v1/admin/communities/:id/enable
POST /api/v1/admin/communities/:id/disable
POST /api/v1/admin/communities/reorder
```

新增 / 更新请求体：

```json
{
  "name": "Rust",
  "slug": "rust",
  "logo": "RS",
  "cover_image": "",
  "slogan": "Rust 工程实践",
  "description": "Rust 子站简介",
  "theme_color": "#b45309",
  "seo_title": "Rust 技术社区",
  "seo_description": "Rust 子站 SEO 描述",
  "seo_keywords": "Rust,Tokio,Axum",
  "sort_order": 6,
  "status": 1,
  "announcement_title": "欢迎加入 Rust 子站",
  "announcement_content": "这里讨论 Rust 工程实践。",
  "announcement_url": "/c/rust/"
}
```

规则：

- 管理员可以新增、排序所有子站。
- 管理员和具备该子站管理范围的后台用户可以编辑当前子站配置。
- `status=1` 启用，`status=0` 禁用，`status=2` 归档。
- 禁用或归档子站不会进入 sitemap，前台 `/c/:slug/` 返回带 `noindex,follow` 的不可用页面。
- 新增子站时 MemoryStore / MySQLStore 会创建默认板块。
- 操作会写入 `admin_logs`，可在 `/admin-next/audit-logs` 查看。

排序请求体：

```json
{
  "ids": [1, 2, 3, 4, 5]
}
```

### Admin Community Categories

```http
GET  /api/v1/admin/communities/:id/categories
POST /api/v1/admin/communities/:id/categories
PUT  /api/v1/admin/categories/:id
POST /api/v1/admin/categories/:id/enable
POST /api/v1/admin/categories/:id/disable
POST /api/v1/admin/categories/reorder
```

新增 / 更新请求体：

```json
{
  "name": "问答中心",
  "slug": "questions",
  "content_type": "question",
  "description": "当前子站的问答板块",
  "icon": "QuestionFilled",
  "sort_order": 2,
  "visible": true,
  "nav_visible": true,
  "postable": true,
  "seo_title": "问答中心",
  "seo_description": "问答板块 SEO 描述",
  "status": 1
}
```

`content_type` 支持 `article`、`question`、`project`、`ai_work`、`job`、`wiki`、`doc`。禁用板块不会出现在前台子站导航中；已有内容不会因为板块禁用而丢失。排序请求体同样使用 `{"ids":[...]}`。

## Topic 互动

### 点赞 / 取消点赞

```http
POST /api/v1/topics/:id/like
```

第一次调用点赞，再次调用取消点赞。响应：

```json
{
  "liked": true,
  "like_count": 10,
  "count": 10,
  "hot_score": 120
}
```

规则：

- 写入或删除 `reactions`，唯一键为 `user_id + target_type + target_id + reaction_type`。
- `like_count` 不会小于 `0`。
- `hot_score = view_count + comment_count*5 + like_count*3 + favorite_count*4`。
- 点赞写入 `activities.action=liked`。
- 非作者点赞会创建 `topic_liked` 通知；自己点赞自己不通知。

### 收藏 / 取消收藏

```http
POST /api/v1/topics/:id/favorite
```

第一次调用收藏，再次调用取消收藏。响应：

```json
{
  "favorited": true,
  "favorite_count": 6,
  "hot_score": 130
}
```

规则：

- 写入或删除 `favorites`，唯一键为 `user_id + target_type + target_id`。
- `favorite_count` 不会小于 `0`。
- 收藏写入 `activities.action=favorited`。
- 非作者收藏会创建 `topic_favorited` 通知；自己收藏自己不通知。

兼容接口：

```http
POST /api/v1/favorites/toggle
```

请求体：

```json
{
  "target_type": "topic",
  "target_id": 1
}
```

### 关注 / 取消关注

```http
POST /api/v1/follows/toggle
Content-Type: application/json
```

请求体：

```json
{
  "target_type": "community",
  "target_id": 1
}
```

响应：

```json
{
  "followed": true,
  "target_type": "community",
  "target_id": 1
}
```

支持 `user`、`community`、`tag`、`topic`。关注写入 `activities.action=followed`；关注用户时触发 `user_followed` 通知，自己关注自己不通知。

### 当前用户互动状态

```http
GET /api/v1/topics/:id/interaction
```

响应：

```json
{
  "liked": true,
  "favorited": false,
  "followed": true,
  "like_count": 10,
  "favorite_count": 6,
  "hot_score": 130
}
```

`GET /api/v1/topics/:id` 也会补充 `liked`、`favorited`、`followed`、`can_edit`、`can_delete`。

## 我的页面接口

```http
GET /api/v1/me/favorites?page=1&page_size=20&target_type=topic
GET /api/v1/me/follows?page=1&page_size=20&target_type=community
GET /api/v1/me/activities?page=1&page_size=20&action=commented
GET /api/v1/me/notifications?page=1&page_size=20&is_read=0
POST /api/v1/me/notifications/:id/read
POST /api/v1/me/notifications/read-all
```

`/me/notifications` 返回：

```json
{
  "items": [],
  "unread_count": 0,
  "total": 0,
  "page": 1,
  "page_size": 20,
  "has_more": false
}
```

兼容通知接口仍保留：

```http
GET  /api/v1/notifications
GET  /api/v1/notifications/unread-count
POST /api/v1/notifications/:id/read
POST /api/v1/notifications/read-all
```

## Compatibility APIs

`sites/posts` 兼容 API 继续保留：

```http
GET    /api/v1/sites
GET    /api/v1/sites/:site
GET    /api/v1/sites/:site/overview
GET    /api/v1/boards
GET    /api/v1/posts
GET    /api/v1/posts/:id
POST   /api/v1/posts
PUT    /api/v1/posts/:id
DELETE /api/v1/posts/:id
POST   /api/v1/posts/:id/like
GET    /api/v1/posts/:id/comments
POST   /api/v1/posts/:id/comments
GET    /api/v1/search
GET    /api/v1/hot
```

这些接口用于旧调用方兼容；新前台主线优先使用 `communities/topics/search`。

## 评论系统

### 获取评论列表

```http
GET /api/v1/topics/:id/comments?page=1&page_size=20&sort=best
```

`sort` 支持：

- `best`：最佳答案优先。
- `latest`：最新优先。
- `oldest`：最早优先。

响应：

```json
{
  "items": [
    {
      "id": 1,
      "post_id": 1,
      "topic_id": 1,
      "parent_id": 0,
      "reply_to_user_id": 0,
      "user_id": 1,
      "user_name": "Demo 用户",
      "author": "Demo 用户",
      "text": "评论内容",
      "content": "评论内容",
      "content_html": "",
      "status": "normal",
      "likes": 0,
      "like_count": 0,
      "is_best": false,
      "created_at": "2026-05-09 12:00:00",
      "updated_at": "2026-05-09 12:00:00",
      "replies": []
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "has_more": false,
  "filters": {
    "sort": "best"
  }
}
```

### 发表评论

```http
POST /api/v1/topics/:id/comments
Content-Type: application/json
```

请求体：

```json
{
  "content": "这是一条评论"
}
```

响应：

```json
{
  "comment": {},
  "item": {},
  "topic": {}
}
```

规则：

- Topic 必须存在。
- `content` 必填，长度为 `2` 到 `5000` 个字符。
- 创建后 `topics.comment_count + 1`，更新 `last_active_at` / `updated_at`，并按统一公式刷新 `hot_score`。
- 写入 `activities.action=commented`，`target_type=topic`。
- 非作者评论 Topic 时创建 `topic_commented` 通知；自己评论自己不通知。
- MySQLStore 中评论写入、统计更新、动态和通知尽量在同一事务中完成，避免计数与评论记录不一致。

### 回复评论

```http
POST /api/v1/topics/:id/comments/:commentId/replies
Content-Type: application/json
```

请求体：

```json
{
  "content": "这是一条回复"
}
```

规则：

- Topic 必须存在。
- 被回复评论必须存在且属于当前 Topic。
- `parent_id` 指向被回复评论，`reply_to_user_id` 指向被回复评论作者。
- 创建后同样更新 `comment_count`、`last_active_at`、`hot_score`。
- 写入 `activities.action=commented`，`target_type=comment`，`target_id` 为被回复评论 ID。
- 非本人回复时创建 `comment_replied` 通知。

## 问答采纳

```http
POST /api/v1/topics/:id/comments/:commentId/accept
```

兼容接口：

```http
POST /api/v1/topics/:id/solve
Content-Type: application/json

{"comment_id": 1}
```

规则：

- 仅 `content_type=question` 的 Topic 可采纳。
- 评论必须存在、属于当前 Topic，且不是隐藏或删除状态。
- 只有 Topic 作者或管理员可以采纳；demo 阶段未登录用户使用 `user_id=1`，seed Topic 作者也是 `1`，因此可完成本地演示。
- 支持采纳和更换最佳答案；暂不支持取消已解决状态。
- 采纳后：
  - `topics.is_solved = true`
  - `topics.best_comment_id = comment.id`
  - 当前评论 `is_best = true`
  - 同 Topic 其他评论 `is_best = false`
  - 更新 `last_active_at` / `updated_at`
  - 写入 `activities.action=accepted_answer`，actor 为实际采纳操作者
  - 非本人时创建 `answer_accepted` 通知
- MySQLStore 中采纳会在事务内清理同 Topic 其他 `is_best`、设置新最佳答案、更新 Topic、写动态和通知。

响应：

```json
{
  "accepted": true,
  "solved": true,
  "best_comment_id": 1,
  "topic": {}
}
```

## 搜索与未解决筛选

```http
GET /api/v1/search/topics
GET /api/v1/search/topics?sort=unsolved
GET /api/v1/search/topics?sort=featured
```

真实行为：

- 返回 `content_type = question` 且 `is_solved = 0` 的 Topic。
- 采纳最佳答案后，该 Topic 不再出现在未解决筛选中。
- `sort=featured` 只返回 `is_featured=true` 且 `status=1` 的 Topic。
- 普通搜索和普通 Topic 列表不会返回隐藏或删除内容。

列表接口也支持：

```http
GET /api/v1/topics?sort=unsolved
GET /api/v1/topics?content_type=question&is_solved=0
```

## 举报和治理

### 创建举报

```http
POST /api/v1/reports
Content-Type: application/json
```

请求体：

```json
{
  "target_type": "topic",
  "target_id": 1,
  "reason_type": "spam",
  "reason_text": "广告内容"
}
```

规则：

- `target_type` 支持 `topic`、`comment`、`user`、`wiki`。
- `target_id` 必须存在。
- `reason_type` 必填。
- `reason_text` 可选，最多 500 字。
- 接口会优先使用当前登录用户；未登录时使用 demo `user_id=1`。
- 创建成功后 `status=pending`。
- 同一用户对同一对象已有 `pending` 举报时会返回 `同一对象已有待处理举报，请勿重复提交`；举报处理完成后可以再次提交。

响应：

```json
{
  "report": {
    "id": 1,
    "reporter_user_id": 1,
    "target_type": "topic",
    "target_id": 1,
    "community_id": 1,
    "topic_id": 1,
    "reason_type": "spam",
    "reason_text": "广告内容",
    "status": "pending",
    "target_title": "主题标题",
    "target_url": "/topics/1/",
    "created_at": "2026-05-09 12:00:00"
  },
  "item": {}
}
```

### 举报管理

```http
GET /api/v1/admin/reports?status=pending&target_type=topic&community_slug=php&page=1&page_size=20
GET /api/v1/admin/reports/:id
POST /api/v1/admin/reports/:id/handle
Authorization: Bearer <access_token>
```

处理请求体：

```json
{
  "status": "accepted",
  "handle_note": "确认违规，已隐藏目标内容"
}
```

规则：

- `status` 只能是 `accepted` 或 `rejected`。
- `admin_users` 中的 `super_admin` 可以查看和处理全部举报。
- 后台站点管理员只能处理自己授权子站内举报。
- `community_moderators` 中启用的版主使用 `users` 身份，只能查看和处理自己负责子站的举报。
- `topic` / `comment` 举报会携带 `community_id`，用于版主范围判断。
- `user` / `wiki` 等无子站归属的全局举报仅管理员可处理。
- `accepted` 会联动隐藏目标 topic 或 comment；`rejected` 只更新举报状态。
- 接受 comment 举报时，如果目标评论是最佳答案，会返回错误，避免破坏问答采纳闭环。
- `handled_by` 使用当前登录用户，`handled_at` 使用处理时间。

### Topic 治理

```http
GET /api/v1/admin/posts
POST /api/v1/admin/posts
PUT /api/v1/admin/posts/:id
DELETE /api/v1/admin/posts/:id
POST /api/v1/admin/topics/:id/feature
POST /api/v1/admin/topics/:id/pin
POST /api/v1/admin/topics/:id/hide
POST /api/v1/admin/topics/:id/restore
POST /api/v1/admin/topics/:id/lock-comments
POST /api/v1/admin/topics/:id/unlock-comments
POST /api/v1/admin/topics/batch
Authorization: Bearer <access_token>
```

规则：

- `/api/v1/admin/posts` 路径保持后台兼容命名，但真实读写 `topics`；公开 `/api/v1/posts` 兼容 API 仍读写 legacy `posts`。
- `admin_users.super_admin` 可操作所有 Topic。
- 后台站点管理员和子站版主只能操作自己负责 `community_id` 下的 Topic。
- `feature` 为 toggle 接口，切换 `is_featured`。
- `pin` 为 toggle 接口，切换 `is_pinned`。
- `hide` 设置 `status=0`，`restore` 设置 `status=1`。
- `lock-comments` 设置 `comment_locked=true`，`unlock-comments` 设置 `false`。
- 被隐藏 Topic 不进入普通列表、搜索和 sitemap；后台内容管理仍可看到。
- 被锁定 Topic 后端会拒绝普通评论和回复创建，前端也显示“评论已锁定”并禁用提交入口。

批量请求体：

```json
{
  "ids": [1, 2, 3],
  "action": "hide",
  "note": "批量下架违规内容"
}
```

`action` 支持 `feature`、`unfeature`、`pin`、`unpin`、`hide`、`restore`、`lock-comments`、`unlock-comments`、`delete`。后台页面中的 `lock_comments` / `unlock_comments` 会映射为后端真实 action。接口逐条校验权限并返回每条成功或失败原因；`note` 会进入批量治理审计日志摘要。

响应：

```json
{
  "topic": {},
  "changed": true
}
```

### Comment 治理

```http
POST /api/v1/admin/comments/:id/hide
POST /api/v1/admin/comments/:id/restore
POST /api/v1/admin/comments/batch
Authorization: Bearer <access_token>
```

规则：

- `admin_users.super_admin` 可操作所有评论。
- 后台站点管理员只能操作自己授权子站内评论。
- 子站版主只能操作自己负责子站下 Topic 的评论。
- `hide` 设置评论 `status=hidden`，`restore` 设置 `status=normal`。
- 普通评论列表过滤 `hidden`、`deleted` 状态。
- 当前版本禁止隐藏最佳答案评论；需要先更换最佳答案或后续治理轮次补取消采纳。
- 批量 `action` 支持 `hide`、`restore`、`delete`，逐条返回成功或失败原因。
- 批量请求体支持可选 `note`，会进入批量治理审计日志摘要。

响应：

```json
{
  "comment": {},
  "changed": true
}
```

### Moderator Workspace API

v1.1.3 新增独立版主工作台 API。它们不使用 `admin_users` 身份，只接受前台 `users` token，并通过 `community_moderators.user_id + community_id + status=1` 校验子站范围。

```http
GET  /api/v1/moderator/communities
GET  /api/v1/moderator/dashboard
GET  /api/v1/moderator/reports?community_id=1&status=pending&page=1&page_size=20
POST /api/v1/moderator/reports/:id/handle
GET  /api/v1/moderator/topics?community_id=1&status=all&content_type=all&keyword=go&page=1&page_size=20
POST /api/v1/moderator/topics/:id/feature
POST /api/v1/moderator/topics/:id/unfeature
POST /api/v1/moderator/topics/:id/pin
POST /api/v1/moderator/topics/:id/unpin
POST /api/v1/moderator/topics/:id/hide
POST /api/v1/moderator/topics/:id/restore
POST /api/v1/moderator/topics/:id/lock-comments
POST /api/v1/moderator/topics/:id/unlock-comments
GET  /api/v1/moderator/comments?community_id=1&status=all&keyword=foo&page=1&page_size=20
POST /api/v1/moderator/comments/:id/hide
POST /api/v1/moderator/comments/:id/restore
GET  /api/v1/moderator/audit-logs?community_id=1&action=hide&target_type=topics&page=1&page_size=20
Authorization: Bearer <frontend_user_access_token>
```

规则：

- 普通 `users` 访问返回 403。
- 后台 admin token 不作为版主工作台身份；super_admin 使用 admin API 和 `/admin-next`。
- 不传 `community_id` 时，列表接口返回当前版主负责的所有子站数据；传入未授权 `community_id` 返回 403。
- 单项治理会先反查 report/topic/comment 所属 `community_id`，再校验当前用户是否为该子站启用版主。
- 版主不能通过这些接口管理 `admin_users`、`community_moderators`、`communities`、categories 或系统设置。
- 治理操作写入 `admin_logs`，`actor_type=moderator`，`actor_id` / `actor_user_id` 对应 `users.id`。
- `GET /api/v1/moderator/audit-logs` 只返回当前版主负责子站的日志。

`GET /api/v1/moderator/dashboard` 响应包含：

```json
{
  "managed_communities": [],
  "pending_report_count": 0,
  "topic_count": 0,
  "comment_count": 0,
  "today_action_count": 0,
  "recent_reports": [],
  "recent_audit_logs": []
}
```

### 版主管理

```http
GET /api/v1/admin/moderators?community_slug=php&status=1&page=1&page_size=20
POST /api/v1/admin/moderators
PUT /api/v1/admin/moderators/:id
DELETE /api/v1/admin/moderators/:id
Authorization: Bearer <access_token>
```

新增 / 更新请求体：

```json
{
  "community_slug": "php",
  "user_id": 2,
  "role": "moderator",
  "status": 1
}
```

规则：

- `role` 支持 `moderator`、`owner`。
- `status=1` 启用，`status=0` 停用。
- 当前版本只有后台管理员可以新增、更新和停用版主。
- `DELETE` 为软停用，不删除 `community_moderators` 行。
- 版主列表会返回 `community_slug/community_name/user_name/user_nickname`，便于后台展示。

### 批量举报处理

```http
POST /api/v1/admin/reports/batch-handle
Authorization: Bearer <access_token>
```

请求体：

```json
{
  "ids": [1, 2],
  "status": "accepted",
  "handle_note": "批量确认违规，已隐藏目标内容"
}
```

`status` 只能是 `accepted` 或 `rejected`。接口逐条校验管理员 / 版主子站权限，`accepted` 沿用单条处理逻辑隐藏目标内容。

### 治理审计日志

```http
GET /api/v1/admin/audit-logs?site=portal&type=audit&actor_type=admin_user&action=批量&target_type=topics&actor=admin&actor_user_id=1&community_id=1&page=1&page_size=20
Authorization: Bearer <access_token>
```

规则：

- `type` 支持 `all`、`audit`、`operation`、`system`、`login` 等当前日志类型。
- `action`、`actor`、`target` 为模糊筛选。
- `actor_type` 支持 `admin_user`、`moderator`、`system`。
- `target_type` 从 `target` 文本前缀派生，支持如 `topics`、`comments`、`reports`、`community_moderators`。
- `actor_user_id` / `actor_id` 指向操作者 ID；`admin_user` 对应 `admin_users.id`，`moderator` 对应 `users.id`。
- `community_id` 当前通过日志 `site_key` 与子站 ID 映射筛选。
- 非全局后台仍按当前站点 scope 返回日志。
- 新增/更新版主、内容 CRUD、举报处理、批量 topic/comment/report 治理都会写入 `admin_logs`。
- 新增 / 编辑 / 启用 / 禁用 / 排序子站，以及新增 / 编辑 / 启用 / 禁用 / 排序子站板块，也会写入 `admin_logs`。
- 后台独立入口为 `/admin-next/audit-logs`；系统设置页仍保留基础日志区。
- 旧 `GET /api/v1/admin/logs` 仍保留，返回当前站点日志列表。

响应字段至少包含：

```json
{
  "items": [
    {
      "id": 1,
      "site": "portal",
      "type": "audit",
      "actor": "admin",
      "actor_type": "admin_user",
      "actor_user_id": 1,
      "actor_id": 1,
      "role": "super_admin",
      "action": "批量治理主题",
      "target": "topics:hide:2/2",
      "target_type": "topics",
      "target_id": 0,
      "community_id": 0,
      "ip": "127.0.0.1",
      "created_at": "2026-05-09 18:00:00"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "has_more": false
}
```

## SEO endpoints

```http
GET /c/:slug
GET /c/:slug/
GET /site/:slug
GET /site/:slug/
GET /tags/:tag
GET /tags/:tag/
GET /topics/:id
GET /topics/:id/
GET /posts/:id
GET /sitemap.xml
GET /sitemap-index.xml
GET /robots.txt
```

说明：

- `/c/:slug` 和 `/c/:slug/` 由 Go 动态输出百度友好的子站 SEO HTML，包含 title、description、canonical、h1、子站简介、板块链接、Topic 链接、标签链接、版主和公告。
- `/site/:slug` 和 `/site/:slug/` 301 跳转到 canonical `/c/:slug/`。
- 禁用或归档子站返回带 `noindex,follow` 的不可用 HTML，不进入 sitemap。
- `/tags/:tag` 和 `/tags/:tag/` 由 Go 动态输出百度友好的标签聚合 SEO HTML，包含 title、description、canonical、h1、标签说明、内容链接和相关标签链接。
- 禁用标签返回不可用页面，不进入 sitemap。
- `/topics/:id` 和 `/topics/:id/` 由 Go 动态输出百度友好的 SEO HTML，不是纯 CSR 空壳。
- `/posts/:id` 301 跳转到 `/topics/:id/`，保留旧入口兼容。
- `/sitemap.xml` 输出启用子站、启用标签和已发布且未隐藏的 Topic；隐藏 Topic、禁用 / 归档子站、禁用标签不进入 sitemap。
- `/robots.txt` 动态输出 sitemap 地址。

## 通知类型与动态类型

当前通知类型：

- `topic_liked`
- `topic_favorited`
- `user_followed`
- `topic_commented`
- `comment_replied`
- `answer_accepted`

已预留但评论体系未完整覆盖的类型：

- `comment_liked`：表结构和 `reactions` 可支持，前台评论点赞本轮未实现。

当前动态 action：

- `created_topic`
- `liked`
- `favorited`
- `followed`
- `commented`
- `accepted_answer`

## 数据结构

当前真实表和字段：

- `comments`：`id`、`post_id`、`topic_id`、`parent_id`、`reply_to_user_id`、`user_id`、`author`、`to_author`、`text`、`content_html`、`status`、`likes`、`is_best`、`created_at`、`updated_at`、`deleted_at`。
- `tags`：`id`、`site_key`、`name`、`slug`、`description`、`status`、`sort_order`、`use_count`、`follower_count`、`seo_title`、`seo_description`、`seo_keywords`、`created_at`、`updated_at`。
- `communities`：`logo`、`cover_image`、`slogan`、`theme_color`、`seo_title`、`seo_description`、`seo_keywords`、`sort_order`、`status`、`follower_count`、`topic_count`、`comment_count`、`hot_score`、`announcement_title`、`announcement_content`、`announcement_url`。
- `categories`：`community_id`、`slug`、`type`、`visible`、`nav_visible`、`postable`、`seo_title`、`seo_description`、`status`。
- `topics`：`content_type`、`status`、`is_pinned`、`is_featured`、`is_solved`、`comment_locked`、`best_comment_id`、`comment_count`、`last_active_at`、`hot_score`。
- `reports`：`reporter_id`、`target_type`、`target_id`、`community_id`、`topic_id`、`reason_type`、`reason_text`、`status`、`handled_by`、`handled_at`、`handle_note`。
- `community_moderators`：`community_id`、`user_id`、`role`、`status`。
- `admin_logs`：`site_key`、`log_type`、`actor`、`actor_type`、`actor_id`、`role_code`、`action`、`target`、`ip`、`created_at`。
- `refresh_tokens`：`user_id`、`token_hash`、`token_type`、`expires_at`、`revoked_at`；当前 `user_id` 按 `token_type=user/admin` 分别指向 `users.id` 或 `admin_users.id`，不再使用单一 `users` 外键约束。
- `activities`：`topic_id`、`action`、`target_type`、`target_id`、`metadata`。
- `notifications`：`actor_user_id`、`type`、`target_type`、`target_id`、`topic_id`、`comment_id`、`is_read`、`read_at`。

`comments.post_id` 仍保留用于 `posts` 兼容 API；第六轮 Topic 评论实际使用 `topic_id`，新 schema 不再给 `comments.post_id` 加 `posts` 外键，避免 Topic ID 与 legacy Post ID 不一致时写入失败。

## 常见错误

- `400 {"error":"ID 不合法"}`：路径 ID 非正整数。
- `400 {"error":"评论内容至少 2 个字符"}`。
- `400 {"error":"评论内容最多 5000 个字符"}`。
- `400 {"error":"只有问答主题可以采纳答案"}`。
- `400 {"error":"举报对象类型不合法"}`。
- `400 {"error":"举报原因不能为空"}`。
- `400 {"error":"举报说明最多 500 字"}`。
- `400 {"error":"处理状态不合法"}`。
- `400 {"error":"评论已锁定"}`。
- `400 {"error":"主题已隐藏"}`。
- `400 {"error":"最佳答案不能隐藏"}`。
- `400 {"error":"同一对象已有待处理举报，请勿重复提交"}`。
- `400 {"error":"版主角色不合法"}`。
- `400 {"error":"不支持的批量主题操作"}`。
- `401 {"error":"未登录"}`：未携带前台 user token。
- `403 {"error":"当前用户不是启用状态子站版主"}`：普通用户访问版主工作台 API。
- `403 {"error":"只有主题作者或管理员可以采纳答案"}`。
- `403 {"error":"无权管理该子站内容"}`。
- `403 {"error":"版主只能治理具体子站"}`：版主工作台尝试处理无子站归属目标。
- `403 {"error":"只有管理员可以管理全局举报"}`。
- `404 {"error":"主题不存在"}`。
- `404 {"error":"父评论不存在"}`。
- `404 {"error":"举报不存在"}`。
- `404 {"error":"通知不存在"}`。

## 部分完成 / 后续完善

- 评论点赞本轮未实现，仅保留 `comments.likes` / `like_count` 字段和旧 `POST /api/v1/comments/:id/like`。
- 采纳支持更换最佳答案，暂不支持取消已解决状态。
- 最佳答案当前通过前端运行时展示，不强制进入 `/topics/:id` 初始 SEO HTML。
- 标签关注已在 v1.2.0 标签页接入，使用 `target_type=tag`。
- 版主工作台是 v1.1.3 MVP，不包含复杂 RBAC、权限点矩阵、版主任期或绩效统计。
- 标签合并、标签别名和标签趋势统计仍是后续增强，v1.2.0 未实现。
