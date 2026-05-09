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
- 写入 `activities.action=commented`，`target_type=comment`。
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
  - 写入 `activities.action=accepted_answer`
  - 非本人时创建 `answer_accepted` 通知

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
```

真实行为：

- 返回 `content_type = question` 且 `is_solved = 0` 的 Topic。
- 采纳最佳答案后，该 Topic 不再出现在未解决筛选中。

列表接口也支持：

```http
GET /api/v1/topics?sort=unsolved
GET /api/v1/topics?content_type=question&is_solved=0
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

第六轮相关真实表和字段：

- `comments`：`id`、`post_id`、`topic_id`、`parent_id`、`reply_to_user_id`、`user_id`、`author`、`to_author`、`text`、`content_html`、`status`、`likes`、`is_best`、`created_at`、`updated_at`、`deleted_at`。
- `topics`：`content_type`、`is_solved`、`best_comment_id`、`comment_count`、`last_active_at`、`hot_score`。
- `activities`：`topic_id`、`action`、`target_type`、`target_id`、`metadata`。
- `notifications`：`actor_user_id`、`type`、`target_type`、`target_id`、`topic_id`、`comment_id`、`is_read`、`read_at`。

`comments.post_id` 仍保留用于 `posts` 兼容 API；第六轮 Topic 评论实际使用 `topic_id`，新 schema 不再给 `comments.post_id` 加 `posts` 外键，避免 Topic ID 与 legacy Post ID 不一致时写入失败。

## 常见错误

- `400 {"error":"ID 不合法"}`：路径 ID 非正整数。
- `400 {"error":"评论内容至少 2 个字符"}`。
- `400 {"error":"评论内容最多 5000 个字符"}`。
- `400 {"error":"只有问答主题可以采纳答案"}`。
- `403 {"error":"只有主题作者或管理员可以采纳答案"}`。
- `404 {"error":"主题不存在"}`。
- `404 {"error":"父评论不存在"}`。
- `404 {"error":"通知不存在"}`。

## 部分完成 / 后续完善

- 评论点赞本轮未实现，仅保留 `comments.likes` / `like_count` 字段和旧 `POST /api/v1/comments/:id/like`。
- 采纳支持更换最佳答案，暂不支持取消已解决状态。
- 最佳答案当前通过前端运行时展示，不强制进入 `/topics/:id` 初始 SEO HTML。
- 标签关注后端和我的关注页可展示，前台标签区域的关注按钮后续增强。
- 举报按钮已预留，完整举报和版主治理后续轮次实现。
