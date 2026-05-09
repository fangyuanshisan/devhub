# DevHub API 文档

[返回文档大纲](README.md)

更新时间：2026-05-09

本文档记录当前仓库真实实现。若需求描述与代码不一致，以本文档中的实际路径、字段和完成度为准。

## 通用规则

- API 前缀：`/api/v1`。
- 认证方式：支持 `Authorization: Bearer <access_token>`。本阶段前台互动接口会优先使用当前登录用户；未登录时使用 demo `user_id=1`，后续可替换为正式登录态。
- 错误响应：`{"error":"错误信息"}`。
- 分页参数：`page`、`page_size`，`page_size` 最大为 `50`。
- `sites/posts` 兼容 API 继续保留，不作为本轮互动和评论闭环的主接口。

## 认证

```http
POST /api/v1/auth/login
POST /api/v1/auth/register
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

登录成功返回 `access_token`、`refresh_token` 和 `user`。MemoryStore 注册当前返回演示会话；MySQLStore 会创建普通用户。

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
- 管理员 user 1 / `super_admin` 可以查看和处理全部举报。
- `community_moderators` 中启用的版主只能查看和处理自己负责子站的举报。
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
- 管理员可操作所有 Topic。
- 子站版主只能操作自己负责 `community_id` 下的 Topic。
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

- 管理员可操作所有评论。
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
- 当前版本只有管理员可以新增、更新和停用版主。
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
GET /api/v1/admin/audit-logs?site=portal&type=audit&action=批量&target_type=topics&actor=admin&actor_user_id=1&community_id=1&page=1&page_size=20
Authorization: Bearer <access_token>
```

规则：

- `type` 支持 `all`、`audit`、`operation`、`system`、`login` 等当前日志类型。
- `action`、`actor`、`target` 为模糊筛选。
- `target_type` 从 `target` 文本前缀派生，支持如 `topics`、`comments`、`reports`、`community_moderators`。
- `actor_user_id` 当前通过 `actor` 文本和用户表派生匹配。
- `community_id` 当前通过日志 `site_key` 与子站 ID 映射筛选。
- 非全局后台仍按当前站点 scope 返回日志。
- 新增/更新版主、内容 CRUD、举报处理、批量 topic/comment/report 治理都会写入 `admin_logs`。
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
      "actor_user_id": 1,
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
- `topics`：`content_type`、`status`、`is_pinned`、`is_featured`、`is_solved`、`comment_locked`、`best_comment_id`、`comment_count`、`last_active_at`、`hot_score`。
- `reports`：`reporter_id`、`target_type`、`target_id`、`community_id`、`topic_id`、`reason_type`、`reason_text`、`status`、`handled_by`、`handled_at`、`handle_note`。
- `community_moderators`：`community_id`、`user_id`、`role`、`status`。
- `admin_logs`：`site_key`、`log_type`、`actor`、`role_code`、`action`、`target`、`ip`、`created_at`。
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
- `403 {"error":"只有主题作者或管理员可以采纳答案"}`。
- `403 {"error":"无权管理该子站内容"}`。
- `403 {"error":"只有管理员可以管理全局举报"}`。
- `404 {"error":"主题不存在"}`。
- `404 {"error":"父评论不存在"}`。
- `404 {"error":"举报不存在"}`。
- `404 {"error":"通知不存在"}`。

## 部分完成 / 后续完善

- 评论点赞本轮未实现，仅保留 `comments.likes` / `like_count` 字段和旧 `POST /api/v1/comments/:id/like`。
- 采纳支持更换最佳答案，暂不支持取消已解决状态。
- 最佳答案当前通过前端运行时展示，不强制进入 `/topics/:id` 初始 SEO HTML。
- 标签关注后端和我的关注页可展示，前台标签区域的关注按钮后续增强。
