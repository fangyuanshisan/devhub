# DevHub 测试文档

[返回文档大纲](README.md)

更新时间：2026-05-09

本文档用于当前真实实现的手工验收。完成代码变更后，优先执行自动检查，再按页面、接口、业务闭环、SEO 顺序回归。

## 启动检查

```bash
bash -n dev.sh
go test ./...
cd web/frontend-app && npm run build
cd web/admin-app && npm run build
```

修改 Go 代码后需要重启：

```bash
./dev.sh --restart
```

默认端口仍是 `8090`，发布 Topic、发表评论、采纳最佳答案都不需要重新 `npm run build`。

### 2026-05-09 第六轮收尾实测

本次收尾只做启动、验收和文档补充，不再改业务功能。

- 残留进程：`lsof -i :8090` 初始无监听；发现并清理了残留 `git@gitee.com git-upload-pack 'OAK_cloud_master/devhub.git'` 子进程。
- Go 模块环境：`GOPROXY=https://goproxy.cn,direct`，`GOPRIVATE` / `GONOSUMDB` 为空，`GOSUMDB=sum.golang.org`；仓库 `go.mod` / `go.sum` 未发现 Gitee 私有依赖。
- 构建：此前 `./dev.sh --local-go --restart` 已完成 Astro 前台和 Vue 后台构建；本机没有 `npm` 时由 Docker Node 完成构建。
- 启动：`go build -o .devhub/devhub .` 成功；最终使用 `setsid` 后台启动二进制：`PORT=8090 CMS_STORE=memory ./.devhub/devhub`。
- 稳定性：`curl -I /`、`GET /api/v1/health`、`GET /api/v1/topics` 均返回 200。

## 基础页面

应可打开：

- `/`
- `/admin-next`
- `/search`
- `/search?sort=unsolved`
- `/topics/new`
- `/c/php`
- `/c/go`
- `/c/java`
- `/c/ai`
- `/c/frontend`
- `/topics/:id`
- `/me/favorites`
- `/me/follows`
- `/me/activities`
- `/notifications`
- `/me/notifications`
- `/sitemap.xml`
- `/robots.txt`

验收要点：

- `/topics/:id` 刷新不 404。
- `/api/*` 不被前端 fallback 吃掉。
- `/posts/:id` 301 跳转到 `/topics/:id/`。
- `/admin`、`/admin/:site` 只兼容重定向到 `/admin-next`。

## 互动回归

点赞：

```bash
curl -X POST http://127.0.0.1:8090/api/v1/topics/1/like
curl -X POST http://127.0.0.1:8090/api/v1/topics/1/like
```

应验证：

- 第一次返回 `liked=true`，再次返回 `liked=false`。
- `like_count` 增加后可减少。
- 不重复计数，且不出现负数。
- `hot_score` 按 `view_count + comment_count*5 + like_count*3 + favorite_count*4` 更新。
- 点赞写入 `activities.liked`，非作者点赞生成 `topic_liked` 通知。

收藏：

```bash
curl -X POST http://127.0.0.1:8090/api/v1/topics/1/favorite
curl -X POST http://127.0.0.1:8090/api/v1/topics/1/favorite
```

应验证：

- 第一次返回 `favorited=true`，再次返回 `favorited=false`。
- `favorite_count` 增加后可减少。
- 收藏后出现在 `/me/favorites` 和 `GET /api/v1/me/favorites`。
- 收藏写入 `activities.favorited`，非作者收藏生成 `topic_favorited` 通知。

关注：

```bash
curl -X POST http://127.0.0.1:8090/api/v1/follows/toggle \
  -H "Content-Type: application/json" \
  -d '{"target_type":"community","target_id":1}'
```

应验证：

- 支持 `user`、`community`、`tag`、`topic`。
- 再次调用取消关注。
- 关注后出现在 `/me/follows` 和 `GET /api/v1/me/follows`。
- 关注写入 `activities.followed`。
- 关注用户生成 `user_followed` 通知；自己关注自己不通知。

## 评论系统

评论列表：

```bash
curl "http://127.0.0.1:8090/api/v1/topics/1/comments?page=1&page_size=20&sort=best"
```

应验证：

- 返回 `items`、`total`、`page`、`page_size`、`has_more`。
- `sort=latest`、`sort=oldest`、`sort=best` 可用。
- 超过一页时详情页显示“加载更多评论”并可继续请求下一页。
- 空评论时前台显示空状态，不白屏。

发表评论：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/topics/1/comments" \
  -H "Content-Type: application/json" \
  -d '{"content":"这是一条测试评论"}'
```

应验证：

- `content` 为空、少于 2 个字符、超过 5000 个字符会返回明确错误。
- 成功后 `comment_count + 1`。
- `last_active_at` 更新。
- `hot_score` 随评论数按权重 `*5` 更新。
- 详情页评论列表刷新。
- 写入 `activities.commented`。
- 非作者评论生成 `topic_commented` 通知；自己评论自己的 Topic 不通知。

回复评论：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/topics/1/comments/1/replies" \
  -H "Content-Type: application/json" \
  -d '{"content":"这是一条测试回复"}'
```

应验证：

- 被回复评论不存在时返回错误。
- 被回复评论不属于当前 Topic 时返回错误。
- 成功后写入 `parent_id`、`reply_to_user_id`。
- 成功后 `comment_count`、`last_active_at`、`hot_score` 更新。
- 写入 `activities.commented`，`target_type=comment`，`target_id` 指向被回复评论。
- 非本人回复生成 `comment_replied` 通知；自己回复自己不通知。

## 问答采纳

采纳最佳答案：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/topics/1/comments/1/accept"
```

也可使用兼容接口：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/topics/1/solve" \
  -H "Content-Type: application/json" \
  -d '{"comment_id":1}'
```

应验证：

- 只有 `content_type=question` 可以采纳。
- 非 question 类型返回错误。
- 只有 Topic 作者或管理员可以采纳；demo 阶段 `user_id=1` 可采纳 seed 作者为 `1` 的问题。
- 采纳后 `topic.is_solved=true`。
- 采纳后 `topic.best_comment_id` 指向评论 ID。
- 被采纳评论 `is_best=true`，其他评论 `is_best=false`。
- 支持更换最佳答案；当前暂不支持取消已解决状态。
- 采纳后详情页显示“最佳答案”。
- 写入 `activities.accepted_answer`。
- `accepted_answer` 动态 actor 为实际采纳操作者。
- 非本人时生成 `answer_accepted` 通知。
- 刷新详情页后最佳答案状态仍存在。

## 未解决筛选

```bash
curl "http://127.0.0.1:8090/api/v1/search/topics?sort=unsolved"
curl "http://127.0.0.1:8090/api/v1/topics?sort=unsolved"
```

应验证：

- 只返回 `content_type=question` 且 `is_solved=false` 的 Topic。
- 已采纳的问题不再出现在未解决筛选中。
- 搜索页 `/search?sort=unsolved` 显示未解决问答。
- 搜索页、子站页、首页卡片对 question 显示“未解决 / 已解决”、回答数和最后活跃时间。

## 我的页面和通知

接口：

```bash
curl "http://127.0.0.1:8090/api/v1/me/favorites"
curl "http://127.0.0.1:8090/api/v1/me/follows"
curl "http://127.0.0.1:8090/api/v1/me/activities"
curl "http://127.0.0.1:8090/api/v1/me/notifications"
curl -X POST "http://127.0.0.1:8090/api/v1/me/notifications/1/read"
curl -X POST "http://127.0.0.1:8090/api/v1/me/notifications/read-all"
```

应验证：

- 收藏、关注、评论、采纳等动作进入我的动态。
- 动态按时间倒序。
- 通知返回 `unread_count`。
- 单条已读和全部已读能更新未读数量。
- 点击通知能跳转 Topic 详情页，当前可不带 comment anchor。

## SEO 回归

```bash
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/'
curl -s http://127.0.0.1:8090/robots.txt
```

必须保持：

- `/topics/:id` 由 Go 动态输出 HTML。
- 源码中有 `<title>`、`meta name="description"`、`<h1>`、正文、标签链接、发布时间、子站名称、分类名称。
- 评论区可以运行时加载，评论内容和最佳答案本轮不强制进入初始 HTML。
- 新发布 Topic、新评论、采纳答案后，不需要重新 build。
- `/sitemap.xml` 包含已发布 Topic。
- `/robots.txt` 不屏蔽必要 CSS / JS / 图片资源。

## Memory / MySQL 模式

Memory 模式应验证：

- 点赞、取消点赞、收藏、取消收藏、关注、取消关注。
- 评论列表、发表评论、回复评论。
- `comment_count`、`last_active_at`、`hot_score` 更新。
- 采纳最佳答案、`is_solved`、`best_comment_id` 和 `is_best` 正确。
- 我的收藏、我的关注、我的动态、通知、已读逻辑可用。
- 当前进程内刷新页面后状态仍正确；不要求磁盘持久化。

MySQL 模式应验证：

- `reactions`、`favorites`、`follows`、`activities`、`notifications`、`comments` 表读写。
- Toggle 操作不重复计数。
- 评论创建和采纳写入数据库字段正确。
- `comment_count`、`last_active_at`、`hot_score`、`is_solved`、`best_comment_id` 正确。
- 评论创建和采纳事务失败时不应留下半更新状态。
- 分页 total 不因 replies join 重复。
- 兼容 MySQL 8。

## 数据结构回归

第六轮必须保持：

- `comments.topic_id`
- `comments.parent_id`
- `comments.reply_to_user_id`
- `comments.user_id`
- `comments.content_html`
- `comments.is_best`
- `comments.deleted_at`
- `topics.comment_count`
- `topics.last_active_at`
- `topics.is_solved`
- `topics.best_comment_id`
- `activities.topic_id`
- `notifications.actor_user_id/type/target_type/target_id/topic_id/comment_id/read_at`

## 当前部分完成

- 评论点赞入口没有纳入第六轮，只保留旧 `POST /api/v1/comments/:id/like` 和字段。
- 采纳支持更换最佳答案，暂不支持取消已解决状态。
- 最佳答案当前通过详情页运行时评论区展示，不进入初始 SEO HTML。
- 举报、版主治理、评论锁定后续轮次实现。

## 第六轮接口实测记录

2026-05-09 收尾验收结果：

- `GET /api/v1/topics`：200。
- `GET /api/v1/search/topics?sort=unsolved`：200；采纳 topic 2 后返回 ID 为 `27,21,14,8` 的未解决 question，topic 2 已移除。
- `GET /api/v1/topics/1/comments`：200，返回评论列表和 replies。
- `POST /api/v1/topics/1/comments`：201，新增评论 ID `9`，`comment_count`、`last_active_at`、`hot_score` 已更新。
- `POST /api/v1/topics/1/comments/9/replies`：201，新增回复 ID `10`，`parent_id=9`，并更新 Topic 统计。
- `POST /api/v1/topics/1/comments/1/accept`：400，符合非 question 不能采纳的规则。
- `POST /api/v1/topics/2/comments`：201，新增可采纳回答 ID `11`。
- `POST /api/v1/topics/2/comments/11/accept`：200，`is_solved=true`，`best_comment_id=11`。
- `GET /api/v1/topics/2/comments?sort=best`：200，评论 ID `11` 的 `is_best=true`。
- `GET /api/v1/me/activities`：200，出现 `commented` 和 `accepted_answer`。
- `GET /api/v1/me/notifications`：200；memory 模式当前认证统一落到 demo/user 1，本次自评论、自回复、自采纳未新增自通知，符合“不通知自己”规则。非本人通知由 MemoryStore / MySQLStore 代码路径支持，MySQL 多用户场景需单独验收。
- `/topics/1` 源码：包含 `<title>`、`meta description`、`<h1>`、`<article>`、正文内容和标签链接。
- `/sitemap.xml`、`/robots.txt`：200。
- 页面 URL：`/`、`/search?sort=unsolved`、`/topics/1`、`/me/activities`、`/notifications`、`/admin-next`、`/topics/new`、`/c/php` 均返回 200。
