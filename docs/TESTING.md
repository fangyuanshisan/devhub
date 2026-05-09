# DevHub 测试文档

[返回文档大纲](README.md)

更新时间：2026-05-09

本文档用于当前真实实现的手工验收。完成代码变更后，优先执行自动检查，再按页面、接口、业务闭环、SEO 顺序回归。

## v1.1.0 测试矩阵

基础启动：

- `./dev.sh --restart`
- `CMS_STORE=memory ./dev.sh --restart`
- `./dev.sh --mysql --restart`
- `go build -o .devhub/devhub . && PORT=8090 CMS_STORE=memory ./.devhub/devhub`
- `lsof -i :8090`
- `/` 前台可访问
- `/admin-next` 后台可访问

前台页面：

- `/`
- `/c/php`
- `/c/go`
- `/c/java`
- `/c/ai`
- `/c/frontend`
- `/site/php`
- `/search`
- `/topics/new`
- `/c/php/topics/new`
- `/topics/:id`
- `/me/favorites`
- `/me/follows`
- `/me/activities`
- `/notifications`

核心 API：

- `GET /api/v1/communities`
- `GET /api/v1/communities/php`
- `GET /api/v1/communities/php/stats`
- `GET /api/v1/communities/php/categories`
- `GET /api/v1/communities/php/tags`
- `GET /api/v1/communities/php/moderators`
- `GET /api/v1/topics`
- `GET /api/v1/topics/:id`
- `GET /api/v1/search/topics`
- `POST /api/v1/topics`
- `GET /api/v1/topics/:id/comments`
- `POST /api/v1/topics/:id/comments`
- `POST /api/v1/topics/:id/comments/:commentId/replies`
- `POST /api/v1/topics/:id/comments/:commentId/accept`
- `POST /api/v1/topics/:id/like`
- `POST /api/v1/topics/:id/favorite`
- `POST /api/v1/follows/toggle`
- `GET /api/v1/me/favorites`
- `GET /api/v1/me/follows`
- `GET /api/v1/me/activities`
- `GET /api/v1/me/notifications`
- `POST /api/v1/reports`

后台 API：

- `GET /api/v1/admin/reports`
- `POST /api/v1/admin/reports/:id/handle`
- `POST /api/v1/admin/reports/batch-handle`
- `GET /api/v1/admin/moderators`
- `POST /api/v1/admin/moderators`
- `PUT /api/v1/admin/moderators/:id`
- `DELETE /api/v1/admin/moderators/:id`
- `POST /api/v1/admin/topics/:id/feature`
- `POST /api/v1/admin/topics/:id/pin`
- `POST /api/v1/admin/topics/:id/hide`
- `POST /api/v1/admin/topics/:id/restore`
- `POST /api/v1/admin/topics/:id/lock-comments`
- `POST /api/v1/admin/topics/:id/unlock-comments`
- `POST /api/v1/admin/topics/batch`
- `POST /api/v1/admin/comments/batch`
- `GET /api/v1/admin/communities`
- `POST /api/v1/admin/communities`
- `GET /api/v1/admin/communities/:id`
- `PUT /api/v1/admin/communities/:id`
- `POST /api/v1/admin/communities/:id/enable`
- `POST /api/v1/admin/communities/:id/disable`
- `POST /api/v1/admin/communities/reorder`
- `GET /api/v1/admin/communities/:id/categories`
- `POST /api/v1/admin/communities/:id/categories`
- `PUT /api/v1/admin/categories/:id`
- `POST /api/v1/admin/categories/:id/enable`
- `POST /api/v1/admin/categories/:id/disable`
- `POST /api/v1/admin/categories/reorder`
- `GET /api/v1/admin/audit-logs`

SEO 回归：

- `/c/:slug` HTML 源码有 `<title>`。
- `/c/:slug` HTML 源码有 `meta name="description"`。
- `/c/:slug` HTML 源码有 canonical。
- `/c/:slug` HTML 源码有 `<h1>`。
- `/c/:slug` HTML 源码有子站简介、板块链接、Topic 链接和标签链接。
- `/site/:slug` 301 到 `/c/:slug/`。
- `/topics/:id` HTML 源码有 `<title>`。
- `/topics/:id` HTML 源码有 `meta name="description"`。
- `/topics/:id` HTML 源码有 `<h1>`。
- `/topics/:id` HTML 源码有正文摘要或正文。
- `/topics/:id` HTML 源码有标签链接。
- `/topics/:id` 不是纯 CSR 空壳。
- `/sitemap.xml` 可访问。
- `/robots.txt` 可访问。
- 隐藏 Topic 不进入 sitemap。
- 隐藏 Topic 返回 noindex 或隐藏状态页面。

后台页面：

- `/admin-next`
- `/admin-next/content`
- `/admin-next/comments`
- `/admin-next/reports`
- `/admin-next/moderators`
- `/admin-next/audit-logs`
- `/admin-next/communities`
- `/admin-next/sites`
- `/admin-next/users`
- `/admin-next/system`

文档对账：

- `README.md`
- `docs/README.md`
- `docs/API.md`
- `docs/TESTING.md`
- `docs/DEPLOYMENT.md`
- `docs/SEO.md`
- `docs/PROJECT_PROGRESS.md`
- `docs/BACKUP_AND_ROLLBACK.md`
- `CHANGELOG.md`
- `docs/releases/v1.1.0.md`
- `docs/releases/v1.0.0.md`

## v1.1.1 身份边界测试清单

前台 / 后台登录态：

- 前台登录 `POST /api/v1/auth/login` 返回 `token_type=user`、`aud=devhub_frontend`。
- 后台登录 `POST /api/v1/admin/login` 返回 `token_type=admin`、`aud=devhub_admin`。
- 前台 localStorage 使用 `devhub_user_token`、`devhub_user_refresh_token`，兼容旧 `devhub_access_token`、`devhub_refresh_token`。
- 后台 sessionStorage 使用 `devhub_admin_token`、`devhub_admin_refresh_token`。
- 前台退出只清理前台 token；后台退出只清理后台 token。

API 边界：

- 前台登录后调用前台接口成功。
- 前台 token 调用 `/api/v1/admin/users`、`/api/v1/admin/settings` 等特权后台接口应失败。
- 后台登录后调用后台接口成功。
- 后台 token 调用 `/api/v1/auth/me` 或前台写操作不应被识别为 `users` 身份。
- 普通 user 不能隐藏 topic。
- 普通 user 不能处理 report。
- `super_admin` 可以治理全站。

版主 scope：

- 已启用 PHP 版主 user token 可以查看和治理 PHP 子站举报 / 内容。
- PHP 版主不能治理 Go 子站。
- Go 版主不能治理 PHP 子站。
- 版主不能新增、更新、停用版主。
- 版主不能管理后台人员。
- 版主不能修改系统设置。

审计日志：

- 后台管理员操作写入 `actor_type=admin_user`，`actor_id` 对应 `admin_users.id`。
- 子站版主治理操作写入 `actor_type=moderator`，`actor_id` 对应 `users.id`。
- 系统自动或 seed 日志可写入 `actor_type=system`。
- `GET /api/v1/admin/audit-logs?actor_type=admin_user` 可筛选后台人员操作。
- `GET /api/v1/admin/audit-logs?actor_type=moderator` 可筛选版主操作。
- demo user 规则仅在开发模式或文档标注范围内生效。

## v1.1.3 独立版主工作台测试清单

页面：

- 普通用户访问 `/moderator` 显示无权限或需要版主身份。
- PHP 版主访问 `/moderator` 成功。
- PHP 版主访问 `/moderator/reports` 成功。
- PHP 版主访问 `/moderator/topics` 成功。
- PHP 版主访问 `/moderator/comments` 成功。
- PHP 版主访问 `/moderator/audit-logs` 成功。
- 页面不要白屏，浏览器 Console 不应有明显错误。
- `/topics/:id` SEO 不受影响。
- `/c/:slug` SEO 不受影响。

API：

```bash
curl "http://127.0.0.1:8090/api/v1/moderator/communities" -H "Authorization: Bearer <user_token>"
curl "http://127.0.0.1:8090/api/v1/moderator/dashboard" -H "Authorization: Bearer <user_token>"
curl "http://127.0.0.1:8090/api/v1/moderator/reports?community_id=1&status=pending" -H "Authorization: Bearer <user_token>"
curl "http://127.0.0.1:8090/api/v1/moderator/topics?community_id=1" -H "Authorization: Bearer <user_token>"
curl "http://127.0.0.1:8090/api/v1/moderator/comments?community_id=1" -H "Authorization: Bearer <user_token>"
curl "http://127.0.0.1:8090/api/v1/moderator/audit-logs?community_id=1" -H "Authorization: Bearer <user_token>"
```

权限：

- 普通 user 调用 `/api/v1/moderator/*` 返回 403。
- PHP 版主只能看到 PHP 子站。
- PHP 版主看不到 Go 子站数据。
- Go 版主只能看到 Go 子站。
- 多子站版主可以切换自己负责的子站。
- PHP 版主处理 PHP report 成功。
- PHP 版主处理 Go report 返回 403。
- PHP 版主隐藏 PHP topic 成功。
- PHP 版主隐藏 Go topic 返回 403。
- PHP 版主隐藏 PHP comment 成功。
- PHP 版主隐藏 Go comment 返回 403。
- 版主不能访问 admin_users 管理。
- 版主不能访问 moderators 管理。
- 版主不能访问 system settings。
- 版主操作写入 audit logs。
- audit logs `actor_type=moderator`。
- audit logs `community_id` 正确。

## v1.2.0 标签系统测试清单

页面：

- `/tags/laravel/` 可访问，不是纯 CSR 空壳。
- `/tags/gin/` 可访问，不是纯 CSR 空壳。
- 标签页显示名称、说明、内容数、关注数、所属子站、最新内容、热门内容、精华内容和相关标签。
- 标签页关注按钮可点击；未登录时按当前前台登录规则返回提示或 401。
- `/admin-next/tags` 可打开，不白屏。
- 发布页 `/topics/new/` 选择子站后能加载标签建议。

公开 API：

```bash
curl "http://127.0.0.1:8090/api/v1/tags"
curl "http://127.0.0.1:8090/api/v1/tags/hot"
curl "http://127.0.0.1:8090/api/v1/tags/suggestions?community_slug=php&q=lar&limit=20"
curl "http://127.0.0.1:8090/api/v1/tags/laravel?community_slug=php"
curl "http://127.0.0.1:8090/api/v1/tags/laravel/topics?community_slug=php&sort=latest"
```

后台 API：

```bash
curl "http://127.0.0.1:8090/api/v1/admin/tags" -H "Authorization: Bearer <admin_token>"
curl "http://127.0.0.1:8090/api/v1/admin/tags/1" -H "Authorization: Bearer <admin_token>"
curl "http://127.0.0.1:8090/api/v1/admin/tags/1/topics" -H "Authorization: Bearer <admin_token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/tags/1/disable" -H "Authorization: Bearer <admin_token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/tags/1/enable" -H "Authorization: Bearer <admin_token>"
```

SEO：

- `/tags/:tag/` 源码包含 `<title>`。
- `/tags/:tag/` 源码包含 `meta name="description"`。
- `/tags/:tag/` 源码包含 canonical。
- `/tags/:tag/` 源码包含 `<h1>`。
- `/tags/:tag/` 源码包含标签说明、Topic 链接和相关标签链接。
- 禁用标签不进入 `/sitemap.xml`。
- `/topics/:id` SEO 不受影响。
- `/c/:slug` SEO 不受影响。

sitemap：

```bash
curl -s "http://127.0.0.1:8090/sitemap.xml" | rg "/tags/"
```

不做项确认：

- v1.2.0 不测试标签合并。
- v1.2.0 不测试标签别名。
- v1.2.0 不测试标签趋势统计。

## 启动检查

```bash
bash -n dev.sh
go test ./...
go build -o .devhub/devhub .
cd web/frontend-app && npm run build
cd web/admin-app && npm run build
```

本机没有 `npm` 时，使用 Docker Node 等价构建：

```bash
docker run --rm -e NPM_CONFIG_REGISTRY=https://registry.npmmirror.com -e FRONTEND_SITE_URL=http://127.0.0.1:8090 -v "$PWD/web/frontend-app:/app" -v "$PWD/web/frontend:/frontend" -w /app node:20-alpine sh -lc 'if [ ! -d node_modules ]; then npm install --registry=https://registry.npmmirror.com --prefer-offline --no-audit --progress=false; fi; npm run build'
docker run --rm -e NPM_CONFIG_REGISTRY=https://registry.npmmirror.com -v "$PWD/web/admin-app:/app" -v "$PWD/web/admin-vue:/admin-vue" -w /app node:20-alpine sh -lc 'if [ ! -d node_modules ]; then npm install --registry=https://registry.npmmirror.com --prefer-offline --no-audit --progress=false; fi; npm run build'
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
- `/site/php`
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
- `/site/:slug` 301 跳转到 `/c/:slug/`。
- `/admin`、`/admin/:site` 只兼容重定向到 `/admin-next`。

## v1.1.0 子站增强

公开子站 API：

```bash
curl "http://127.0.0.1:8090/api/v1/communities"
curl "http://127.0.0.1:8090/api/v1/communities/php"
curl "http://127.0.0.1:8090/api/v1/communities/php/stats"
curl "http://127.0.0.1:8090/api/v1/communities/php/categories"
curl "http://127.0.0.1:8090/api/v1/communities/php/tags"
curl "http://127.0.0.1:8090/api/v1/communities/php/moderators"
```

应验证：

- `communities` 返回 logo、cover_image、slogan、theme_color、SEO、统计和公告字段。
- `stats` 返回 topic_count、comment_count、question_count、unsolved_count、follower_count、today_topic_count、today_comment_count、moderator_count 和 hot_score。
- `categories` 返回启用状态、nav_visible、postable 和 content_type。
- `moderators` 只返回启用版主。

子站页面和 SEO：

```bash
curl -s "http://127.0.0.1:8090/c/php/" | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'
curl -I "http://127.0.0.1:8090/site/php"
curl "http://127.0.0.1:8090/c/go/"
curl "http://127.0.0.1:8090/c/java/"
curl "http://127.0.0.1:8090/c/ai/"
curl "http://127.0.0.1:8090/c/frontend/"
```

应验证：

- `/c/php/` 源码有 title、meta description、canonical、h1、简介、板块链接、Topic 链接和标签链接。
- `/site/php` 返回 301，Location 指向 `/c/php/`。
- 子站不存在返回友好 404。
- 禁用或归档子站返回 noindex 不可用页面，并不进入 sitemap。
- `/topics/:id` SEO 不受影响。

子站关注：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/follows/toggle" \
  -H "Content-Type: application/json" \
  -d '{"target_type":"community","target_id":1}'
```

应验证：

- 第一次返回 `followed=true`，再次返回 `followed=false`。
- follower_count 增加 / 减少且不小于 0。
- 关注写入 `activities.action=followed`。
- 我的关注接口和页面可看到 community 关注记录。

后台子站和板块管理：

```bash
curl "http://127.0.0.1:8090/api/v1/admin/communities" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/communities" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Rust","slug":"rust","slogan":"Rust 工程实践","description":"Rust 子站","theme_color":"#b45309","status":1}'
curl -X PUT "http://127.0.0.1:8090/api/v1/admin/communities/<id>" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Rust","slug":"rust","seo_title":"Rust 技术社区","announcement_title":"欢迎","announcement_content":"欢迎加入 Rust 子站","status":1}'
curl "http://127.0.0.1:8090/api/v1/admin/communities/<id>/categories" -H "Authorization: Bearer <token>"
```

应验证：

- `/admin-next/communities` 可以打开。
- 可以新增、编辑、启用 / 禁用子站。
- 可以设置子站 SEO 和公告。
- 可以查看、新增、编辑、启用 / 禁用、排序子站板块。
- 子站板块启用 / 禁用会影响 `/c/:slug/` 前台导航。
- 操作写入 `/admin-next/audit-logs`。

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

## 第七轮举报和治理

举报 Topic：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/reports" \
  -H "Content-Type: application/json" \
  -d '{"target_type":"topic","target_id":1,"reason_type":"spam","reason_text":"测试举报"}'
```

举报 Comment：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/reports" \
  -H "Content-Type: application/json" \
  -d '{"target_type":"comment","target_id":1,"reason_type":"abuse","reason_text":"测试评论举报"}'
```

应验证：

- 成功返回 `report.status=pending`。
- 非法 `target_type`、不存在 `target_id`、空 `reason_type`、超过 500 字说明会返回明确错误。
- Topic 详情页有举报入口，评论区每条评论有举报入口。
- 举报成功前台显示提示，失败显示错误且不白屏。

后台登录获取 token：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/admin/login" \
  -H "Content-Type: application/json" \
  -d '{"account":"admin","password":"admin123"}'
```

举报管理：

```bash
curl "http://127.0.0.1:8090/api/v1/admin/reports" \
  -H "Authorization: Bearer <token>"

curl -X POST "http://127.0.0.1:8090/api/v1/admin/reports/1/handle" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"accepted","handle_note":"确认违规，已隐藏目标内容"}'
```

应验证：

- 管理员可以查看和处理全部举报。
- PHP 版主 user 2 只能处理 PHP 子站举报，Go 版主 user 3 只能处理 Go 子站举报。
- 无子站归属的 user/wiki 举报只有管理员可处理。
- `accepted` 会隐藏目标 topic 或 comment；`rejected` 只更新举报状态。
- 目标评论是最佳答案时，接受举报会被拒绝并保持原状态。

Topic 治理：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/feature" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/pin" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/lock-comments" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/unlock-comments" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/hide" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/restore" -H "Authorization: Bearer <token>"
```

应验证：

- `feature` / `pin` 是 toggle，再次调用可取消。
- 精华内容可通过 `GET /api/v1/search/topics?sort=featured` 和 `/search?sort=featured` 查询。
- 置顶内容在普通列表中优先展示，并显示“置顶”。
- 隐藏后普通 `GET /api/v1/topics`、`GET /api/v1/search/topics` 不再返回该 Topic。
- 隐藏后 `/sitemap.xml` 不包含该 Topic。
- 恢复后普通列表、搜索和 sitemap 可再次出现。

评论锁定：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/lock-comments" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/topics/1/comments" \
  -H "Content-Type: application/json" \
  -d '{"content":"锁定后这条评论应被拒绝"}'
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/unlock-comments" -H "Authorization: Bearer <token>"
```

应验证：

- 锁定后后端返回 `评论已锁定`。
- 锁定后回复评论同样被拒绝。
- 前台详情页显示“评论已锁定”，提交入口禁用。
- 解锁后普通评论和回复恢复可用。

Comment 治理：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/admin/comments/1/hide" -H "Authorization: Bearer <token>"
curl "http://127.0.0.1:8090/api/v1/topics/1/comments"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/comments/1/restore" -H "Authorization: Bearer <token>"
```

应验证：

- 隐藏评论后普通评论列表不返回该评论。
- 恢复后评论重新出现在普通评论列表。
- 隐藏最佳答案返回 `最佳答案不能隐藏`。

后台页面：

- `/admin-next` 可打开。
- `/admin-next/reports` 可打开并显示举报列表。
- 内容管理的更多菜单包含精华、置顶、隐藏 / 恢复、锁定 / 解锁评论。
- 评论管理包含隐藏 / 恢复评论入口。

## 第八轮 admin-next CRUD、版主、批量治理和审计

后台内容 CRUD：

```bash
curl "http://127.0.0.1:8090/api/v1/admin/posts" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/posts" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"site":"php","board":"community","title":"第八轮后台新增内容","summary":"后台 CRUD 验收","content":"这是一条通过 admin-next 后台 API 写入 topics 的内容。","status":"publish","tags":["Laravel"]}'
curl -X PUT "http://127.0.0.1:8090/api/v1/admin/posts/<id>" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"第八轮后台编辑内容","pinned":true,"recommended":true}'
curl -X DELETE "http://127.0.0.1:8090/api/v1/admin/posts/<id>" -H "Authorization: Bearer <token>"
```

应验证：后台 `/admin/posts` 新增、编辑、删除真实作用于 `topics`；公开 `/api/v1/posts` 兼容 API 仍可用且未删除。

版主管理：

```bash
curl "http://127.0.0.1:8090/api/v1/admin/moderators" -H "Authorization: Bearer <token>"
curl -X POST "http://127.0.0.1:8090/api/v1/admin/moderators" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"community_slug":"java","user_id":2,"role":"moderator","status":1}'
curl -X PUT "http://127.0.0.1:8090/api/v1/admin/moderators/<id>" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"role":"owner","status":1}'
curl -X DELETE "http://127.0.0.1:8090/api/v1/admin/moderators/<id>" -H "Authorization: Bearer <token>"
```

应验证：新增和更新仅管理员可操作；删除为停用，`status=0`；停用后不再拥有对应子站治理权限。

批量治理：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/batch" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"ids":[1,2],"action":"feature"}'

curl -X POST "http://127.0.0.1:8090/api/v1/admin/comments/batch" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"ids":[1,2],"action":"hide"}'

curl -X POST "http://127.0.0.1:8090/api/v1/admin/reports/batch-handle" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"ids":[1,2],"status":"rejected","handle_note":"批量驳回"}'
```

应验证：返回 `updated/failed/items`；每条失败原因清晰；版主只能批量处理自己子站内对象。

举报频率限制：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/reports" \
  -H "Content-Type: application/json" \
  -d '{"target_type":"topic","target_id":1,"reason_type":"spam","reason_text":"第一次举报"}'
curl -X POST "http://127.0.0.1:8090/api/v1/reports" \
  -H "Content-Type: application/json" \
  -d '{"target_type":"topic","target_id":1,"reason_type":"spam","reason_text":"重复举报"}'
```

应验证：第二次返回 `同一对象已有待处理举报，请勿重复提交`；处理该举报后允许再次提交。

治理审计：

```bash
curl "http://127.0.0.1:8090/api/v1/admin/audit-logs?type=audit&page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

应验证：内容 CRUD、版主管理、举报处理、批量治理均写入 `admin_logs`；后台 `/admin-next/audit-logs` 可以筛选审计日志。

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
- `/sitemap.xml` 不包含 `status=0` 隐藏 Topic。
- `/sitemap.xml` 包含启用子站 `/c/php/` 等 canonical URL。
- `/sitemap.xml` 不包含 `/site/:slug` 兼容 URL。
- `/sitemap.xml` 不包含禁用或归档子站。
- `/robots.txt` 不屏蔽必要 CSS / JS / 图片资源。

隐藏 Topic SEO：

```bash
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/hide" -H "Authorization: Bearer <token>"
curl -s http://127.0.0.1:8090/topics/1/ | rg '内容已隐藏|noindex'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/1/' || true
curl -X POST "http://127.0.0.1:8090/api/v1/admin/topics/1/restore" -H "Authorization: Bearer <token>"
```

应验证：隐藏详情页仍由 Go 动态输出，不退化为纯 CSR；页面不输出原正文，包含 `noindex,follow`。

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
- `reports` 和 `community_moderators` 读写、举报分页、举报处理、版主权限判断可用。
- 版主管理 CRUD、后台内容 CRUD、批量 topic/comment/report 治理、举报 pending 去重和审计日志筛选可用。
- 子站增强字段、子站统计、后台子站 CRUD、后台子站板块 CRUD、子站关注 follower_count 和 sitemap 子站过滤可用。
- Topic 精华、置顶、隐藏、恢复、评论锁定字段更新正确。
- Comment 隐藏、恢复字段更新正确。
- 普通列表、搜索、评论列表过滤隐藏内容。
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
- `topics.is_pinned`
- `topics.is_featured`
- `topics.status`
- `topics.comment_locked`
- `reports.community_id`
- `reports.topic_id`
- `reports.handle_note`
- `community_moderators.community_id/user_id/role/status`
- `communities.logo/cover_image/slogan/theme_color/seo_title/seo_description/seo_keywords/sort_order/status/follower_count/topic_count/comment_count/hot_score/announcement_title/announcement_content/announcement_url`
- `categories.visible/nav_visible/postable/seo_title/seo_description/status`
- `activities.topic_id`
- `notifications.actor_user_id/type/target_type/target_id/topic_id/comment_id/read_at`

## 当前部分完成

- 评论点赞入口没有纳入第六轮，只保留旧 `POST /api/v1/comments/:id/like` 和字段。
- 采纳支持更换最佳答案，暂不支持取消已解决状态。
- 最佳答案当前通过详情页运行时评论区展示，不进入初始 SEO HTML。
- 标签高级能力、标签后台和标签 SEO 聚合页不属于 v1.1.0 主线，计划放入 v1.2.0。

## CI 回归

GitHub Actions 配置位于 `.github/workflows/ci.yml`，执行：

- `go test ./...`
- `go build -o /tmp/devhub .`
- `cd web/frontend-app && npm ci && FRONTEND_SITE_URL=http://127.0.0.1:8090 npm run build`
- `cd web/admin-app && npm ci && npm run build`
- 检查 `db/mysql/001_schema.sql` 和核心文档文件存在。

CI 不依赖本地私有 token 或私有代理；前端和后台构建使用 package lock。

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

## 第八轮接口实测记录

2026-05-09 已完成第八轮本地验收：

- `go test -count=1 ./...`：通过。
- `npm run build`：本机无 `npm`，预期失败；已使用 Docker Node 构建前台和 admin-next 并通过。
- Docker Node 前台构建：通过。
- Docker Node 后台构建：通过，仅有 Vite chunk size warning；产物包含 `AuditLogs-*.js`。
- `./dev.sh --restart`：已执行，完成 Astro 前台和 Vue 后台构建，并启动 Docker Go 服务。
- 稳定服务：为避免 `dev.sh` 前台进程悬挂，最终停止 Docker Go 容器，执行 `go build -o .devhub/devhub .`，并以 `PORT=8090 CMS_STORE=memory ./.devhub/devhub` 后台启动；`lsof -i :8090` 显示 `devhub` 正在监听。
- `GET /api/v1/health`：200，`store=memory`。
- 后台内容 CRUD：`POST /api/v1/admin/posts` 成功创建 topic，`PUT /api/v1/admin/posts/:id` 成功更新标题、置顶和精华；后台路径真实写入 `topics`。
- 批量 Topic 治理：`POST /api/v1/admin/topics/batch` 的 `lock-comments` / `unlock-comments` 成功；锁定后普通评论返回 `评论已锁定`。
- 批量 Comment 治理：使用真实 comment ID 验证 `POST /api/v1/admin/comments/batch` 的 `hide` / `restore`，均返回 `updated=1, failed=0`。
- 举报频率限制：首次 `POST /api/v1/reports` 返回 `pending`，同一用户同一对象重复 pending 举报返回 `同一对象已有待处理举报，请勿重复提交`。
- 批量举报处理：`POST /api/v1/admin/reports/batch-handle` 的 `rejected` 返回 `updated=1, failed=0`。
- 版主管理 CRUD：`POST /api/v1/admin/moderators` 新增 Java 版主，`PUT /api/v1/admin/moderators/:id` 更新为 `owner`，`DELETE /api/v1/admin/moderators/:id` 返回 `disabled=true`；稳定服务重启后 seed 版主列表正常返回 PHP/Go 两条。
- 治理审计：`GET /api/v1/admin/audit-logs` 返回登录、批量治理、举报处理、版主管理等日志，包含派生的 `actor_user_id`、`target_type`、`target_id`、`community_id`；稳定服务重启后 seed 和新登录日志正常返回。
- 页面 URL：`/`、`/search?sort=unsolved`、`/topics/1`、`/admin-next`、`/admin-next/content`、`/admin-next/comments`、`/admin-next/reports`、`/admin-next/moderators`、`/admin-next/audit-logs`、`/admin-next/system` 均返回 200。
- SEO 回归：`/topics/1` 源码保留 `<title>`、`meta description`、`<h1>`、`<article>`、正文/摘要、标签链接和 Article JSON-LD；`/sitemap.xml`、`/robots.txt` 均返回 200。

注意：MemoryStore 重启后运行时创建的临时 topic/report/comment/moderator 不持久化，这是当前 memory 模式预期行为。

## 第九轮 v1.0.0 归档实测记录

2026-05-09 已完成 v1.0.0 归档验收：

- `bash -n dev.sh`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- 前台 Docker Node 构建：通过，Astro 输出到 `web/frontend`。
- 后台 Docker Node 构建：通过，Vite 仅提示 chunk size warning，产物包含 `AuditLogs-*.js`。
- `./dev.sh --local-go restart --no-build`：已执行，健康检查通过；随后为避免 `go run` 临时进程链路，切换为 `.devhub/devhub` 二进制后台常驻。
- 当前 8090：`lsof -i :8090` 显示 `./.devhub/devhub` 监听，`GET /api/v1/health` 返回 `store=memory`。
- 基础 URL：`/`、`/admin-next`、`/api/v1/communities`、`/api/v1/topics`、`/api/v1/search/topics?keyword=go`、`/topics/1`、`/sitemap.xml`、`/robots.txt` 均返回 200。
- 后台页面：`/admin-next/content`、`/admin-next/comments`、`/admin-next/reports`、`/admin-next/moderators`、`/admin-next/audit-logs`、`/admin-next/sites`、`/admin-next/users`、`/admin-next/system` 均返回 200。
- 后台 API：`GET /api/v1/admin/reports`、`GET /api/v1/admin/moderators`、`GET /api/v1/admin/audit-logs` 均返回 200。
- 批量治理：`POST /api/v1/admin/topics/batch` 的 `feature/unfeature` 返回 `updated=1, failed=0`；`POST /api/v1/admin/comments/batch` 的 `hide/restore` 返回 `updated=1, failed=0`。
- 批量举报：创建测试举报后，`POST /api/v1/admin/reports/batch-handle` 的 `rejected` 返回 `updated=1, failed=0`。
- SEO 回归：`/topics/1` 源码包含 `<title>`、`meta description`、`<h1>`、`<article>`、正文、标签链接、发布时间和 Article JSON-LD，且不是纯 CSR 空壳。
- 隐藏内容 SEO：隐藏 Topic 1 时详情页返回“内容已隐藏”/`noindex`，严格匹配 sitemap 不含 `/topics/1/`；恢复后 `/topics/1` SEO 正常，sitemap 重新包含 `/topics/1/`。
- 文档对账：`README.md`、`docs/README.md`、`docs/API.md`、`docs/TESTING.md`、`docs/DEPLOYMENT.md`、`docs/SEO.md`、`docs/PROJECT_PROGRESS.md`、`docs/BACKUP_AND_ROLLBACK.md`、`CHANGELOG.md`、`docs/releases/v1.0.0.md` 已同步 v1.0.0 状态。

## v1.1.0 子站增强实测记录

2026-05-09 已完成 v1.1.0 子站增强验收：

- `bash -n dev.sh`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- 本机 `npm run build`：失败，原因是本机没有 `npm`，符合当前环境预期。
- 前台 Docker Node 构建：通过，Astro 输出到 `web/frontend`。
- 后台 Docker Node 构建：通过，Vite 仅提示 chunk size warning，产物包含 `Communities-*.js`。
- `./dev.sh --restart`：通过，完成前后台构建并启动 Docker Go 服务。
- 稳定服务：最终按文档切换为 `PORT=8090 CMS_STORE=memory ./.devhub/devhub` 二进制后台常驻；`lsof -i :8090` 显示 `./.devhub/devhub` 监听。
- 健康检查：`GET /api/v1/health` 返回 200，`store=memory`。
- 子站 API：`GET /api/v1/communities`、`/api/v1/communities/php`、`/stats`、`/categories`、`/tags`、`/moderators` 均返回 200，并返回 v1.1.0 增强字段。
- 子站页面：`/c/php/`、`/c/go/`、`/c/java/`、`/c/ai/`、`/c/frontend/` 均返回 200。
- 兼容入口：`GET /site/php` 和 `HEAD /site/php` 均返回 301，Location 为 `/c/php/`。
- 子站 SEO：`/c/php/` 源码包含 `<title>`、`meta description`、canonical、`<h1>`、子站简介、公告、Topic 链接和标签链接。
- Topic SEO 回归：`/topics/1` 源码包含 `<title>`、`meta description`、`<h1>`、`<article>`、正文、标签链接和 Article JSON-LD。
- sitemap：`/sitemap.xml` 包含 `/c/php/`、`/c/go/`、`/c/java/`、`/c/ai/`、`/c/frontend/` 和 `/topics/1/`，不包含 `/site/php`。
- robots：`/robots.txt` 返回 200。
- 后台页面：`/admin-next`、`/admin-next/communities`、`/admin-next/moderators`、`/admin-next/audit-logs` 均返回 200。
- 子站关注：`POST /api/v1/follows/toggle` 使用 `target_type=community,target_id=1` 时第一次返回 `followed=true`，`follower_count` 从 0 到 1；第二次返回 `followed=false`，`follower_count` 回到 0；动态中出现 `action=followed,target_type=community`。
- 子站后台 CRUD：使用临时 memory 子站验证新增、编辑 SEO / 公告、禁用、启用均成功。
- 子站板块 CRUD：使用临时 memory 子站验证新增板块、禁用、启用均成功。
- 禁用子站 SEO：临时子站禁用后 `/c/:slug/` 返回 404 且包含 `noindex`，不进入 sitemap；启用后页面恢复 200。
- 审计日志：子站新增 / 更新 / 启用 / 禁用、子站板块新增 / 启用 / 禁用均写入 `/api/v1/admin/audit-logs`。
