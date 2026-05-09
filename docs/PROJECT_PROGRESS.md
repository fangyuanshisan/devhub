# DevHub 项目进度

[返回文档大纲](README.md)

更新时间：2026-05-09

本文档记录当前仓库的真实实现状态。需求来源以根目录 `更新.md` 为准；旧规划文档仅作背景参考。

## 当前结论

DevHub 当前是 “Go API + Astro 前台 + Vue 后台” 的多子站社区 CMS。默认入口和端口保持不变：

```text
前台：http://127.0.0.1:8090/
后台：http://127.0.0.1:8090/admin-next
```

首页、子站页、搜索页和用户中心使用 Astro 静态壳 + 运行时 API；Topic 详情页 `/topics/:id` 仍由 Go 动态输出 SEO HTML。第五轮互动闭环已完成基础实现：点赞、收藏、关注、我的收藏、我的关注、我的动态、通知列表和已读逻辑在 MemoryStore 与 MySQLStore 中均可用。第六轮已完成评论列表、发表评论、回复评论、问答采纳、未解决筛选、评论动态和评论通知的基础闭环。第七轮已完成举报、版主范围治理、精华、置顶、隐藏、评论锁定和后台最小治理入口。第八轮补丁后，admin-next 后台内容 CRUD、版主管理 CRUD、批量治理、批量举报处理和治理审计日志均有真实页面入口和 API 封装。

## 近期迭代摘要

- 第一轮：同步通用社区 schema，补齐 PHP、Go、Java、AI、Frontend 五个子站 seed 数据。
- 第二轮：前台逐步迁移到 `communities/topics/search` API，并保留 `sites/posts` 兼容 API。
- 第三轮：完成发布入口、发布页、`POST /api/v1/topics`，MemoryStore / MySQLStore 均支持创建 Topic。
- 第四轮：增强 `/api/v1/search/topics`，支持关键词、范围、子站、分类、标签、排序、分页组合筛选；新增 `/api/v1/tags/hot`。
- SEO 整改：`/topics/:id` 改由 Go 动态输出 SEO 友好的 HTML；`/sitemap.xml` 和 `/robots.txt` 改为 Go 动态输出。
- 前台登录完善：首页登录/注册接入 `/api/v1/auth/*`，导航支持会话恢复、refresh token 刷新和退出登录。
- 第五轮互动联动：补齐点赞、收藏、关注、我的收藏、我的关注、我的动态、通知中心和详情页互动状态；新增用户中心页面入口。
- 第六轮评论问答：补齐 `GET/POST /api/v1/topics/:id/comments`、回复、采纳最佳答案、`sort=unsolved` 未解决筛选、`commented/accepted_answer` 动态和 `topic_commented/comment_replied/answer_accepted` 通知。
- 第七轮社区治理：补齐 topic/comment 举报、举报后台处理、版主子站权限、精华、置顶、隐藏、恢复、评论锁定、评论隐藏和 sitemap 隐藏过滤。
- 第八轮后台治理：补齐 admin-next 内容 CRUD 写入 `topics`、版主管理 CRUD、topic/comment/report 批量治理、举报重复 pending 限制、`/admin-next/moderators` 和 `/admin-next/audit-logs`。

## 第五轮互动联动状态

按 2026-05-09 当前代码，状态如下：

- 点赞 / 取消点赞：已完成。真实接口为 `POST /api/v1/topics/:id/like`，也保留 `POST /api/v1/reactions/toggle`。第一次调用写入 `reactions`，再次调用删除；返回 `liked`、`like_count`、`hot_score`。
- 收藏 / 取消收藏：已完成。真实接口为 `POST /api/v1/topics/:id/favorite`，兼容 `POST /api/v1/favorites/toggle`。第一次调用写入 `favorites`，再次调用删除；返回 `favorited`、`favorite_count`、`hot_score`。
- 关注 / 取消关注：已完成。真实接口为 `POST /api/v1/follows/toggle`，支持 `user`、`community`、`tag`、`topic`；再次调用取消关注。
- 我的收藏页面：已完成。前台路由 `/me/favorites`，接口 `GET /api/v1/me/favorites`。
- 我的关注页面：已完成。前台路由 `/me/follows`，接口 `GET /api/v1/me/follows`。
- 我的动态页面：已完成。前台路由 `/me/activities`，接口 `GET /api/v1/me/activities`；兼容 `GET /api/v1/activities`。
- 通知页面：已完成。前台路由 `/notifications`，`/me/notifications` 为同一静态壳别名；接口 `GET /api/v1/me/notifications`。
- 通知单条已读：已完成。真实接口为 `POST /api/v1/me/notifications/:id/read`；兼容旧 `POST /api/v1/notifications/:id/read`。
- 通知全部已读：已完成。真实接口为 `POST /api/v1/me/notifications/read-all`；兼容旧 `POST /api/v1/notifications/read-all`。
- `hot_score` 更新逻辑：已完成基础公式。Topic 互动后使用 `view_count + comment_count*5 + like_count*3 + favorite_count*4` 重新计算；MemoryStore 运行时计算，MySQLStore 写回 `topics.hot_score`。
- MemoryStore 支持情况：已完成本轮基础能力。支持点赞、取消点赞、收藏、取消收藏、关注、取消关注、我的收藏、我的关注、我的动态、通知列表、单条已读、全部已读；内存模式不持久化，当前进程内状态正确。
- MySQLStore 支持情况：已完成本轮基础能力。支持 `reactions`、`favorites`、`follows`、`activities`、`notifications` 读写；toggle 避免重复计数；取消操作使用 `GREATEST(...,0)` 避免负数。
- 是否写入 `activities`：已完成。发布写入 `created_topic`；点赞写入 `liked`；收藏写入 `favorited`；关注写入 `followed`。
- 是否写入 `notifications`：已完成基础规则。点赞、收藏会给 Topic 作者生成通知；关注用户会给被关注用户生成通知；评论 Topic 生成 `topic_commented`，回复评论生成 `comment_replied`，回答被采纳生成 `answer_accepted`；自己操作自己的内容不通知自己。
- 是否保持 `/topics/:id` 百度 SEO HTML 不被破坏：已保持。详情页仍由 Go 动态输出 title、meta description、h1、正文、标签链接、发布时间和 Article JSON-LD；点赞、收藏、关注、评论作为运行时增强。
- 第七轮已补齐：举报、版主范围治理、精华、置顶、隐藏和评论锁定。
- 尚未完成的问题：评论点赞本轮未实现；采纳支持更换最佳答案但暂不支持取消已解决状态；复杂权限继续沿用当前轻量 RBAC 和 demo user 方案。

## 第六轮评论系统和问答采纳状态

按 2026-05-09 当前代码，状态如下：

- 评论列表：已完成。真实接口为 `GET /api/v1/topics/:id/comments`，支持 `page`、`page_size`、`sort=best/latest/oldest`；详情页评论区支持加载更多。
- 发表评论：已完成。真实接口为 `POST /api/v1/topics/:id/comments`，请求字段为 `content`，长度限制为 2 到 5000 个字符。
- 回复评论：已完成。真实接口为 `POST /api/v1/topics/:id/comments/:commentId/replies`，会校验父评论必须属于当前 Topic。
- 评论统计联动：已完成。评论和回复都会更新 `comment_count`、`last_active_at`、`updated_at`，并按统一公式刷新 `hot_score`。
- 评论动态：已完成。评论 Topic 写入 `activities.action=commented,target_type=topic`；回复写入 `activities.action=commented,target_type=comment`，`target_id` 指向被回复评论。
- 评论通知：已完成。非作者评论 Topic 生成 `topic_commented`；非本人回复评论生成 `comment_replied`；自己评论 / 回复自己不通知。
- 问答状态：已完成。`topics.content_type=question` 使用 `is_solved` 和 `best_comment_id` 表示已解决状态。
- 采纳最佳答案：已完成。真实接口为 `POST /api/v1/topics/:id/comments/:commentId/accept`，兼容 `POST /api/v1/topics/:id/solve`。
- 更换最佳答案：已完成。新采纳评论 `is_best=true`，同 Topic 其他评论 `is_best=false`；暂不支持取消已解决状态。
- 采纳动态和通知：已完成。采纳写入 `accepted_answer` 动态；使用实际操作者作为 actor；非本人时生成 `answer_accepted` 通知。
- 未解决筛选：已完成。`GET /api/v1/search/topics?sort=unsolved` 和 `GET /api/v1/topics?sort=unsolved` 只返回 `content_type=question AND is_solved=0`。
- MemoryStore 支持情况：已完成第六轮基础能力，当前进程内支持评论、回复、采纳、未解决筛选、动态和通知。
- MySQLStore 支持情况：已完成第六轮基础能力，支持 `comments` 表读写、统计更新、动态、通知、采纳状态和未解决筛选；评论创建、统计更新、活动和通知尽量在事务内完成，采纳最佳答案同样使用事务更新 Topic、评论、动态和通知。
- SEO 是否保持不被破坏：已保持。`/topics/:id` 仍由 Go 动态输出 SEO HTML；评论区和最佳答案为运行时增强，当前不强制进入初始 SEO HTML。
- 尚未完成事项：评论点赞未纳入本轮；取消已解决状态未实现；评论 anchor 跳转当前可后续优化。

## 第七轮举报和社区治理状态

按 2026-05-09 当前代码，状态如下：

- 用户举报：已完成。真实接口为 `POST /api/v1/reports`，支持举报 `topic`、`comment`、`user`、`wiki`；本轮前台在 Topic 详情页和评论区提供举报入口。
- 举报管理：已完成最小后台。真实接口为 `GET /api/v1/admin/reports`、`GET /api/v1/admin/reports/:id`、`POST /api/v1/admin/reports/:id/handle`；`accepted` 会联动隐藏 topic 或 comment，`rejected` 只更新举报状态。
- 版主权限：已完成最小可用判断。`currentUserID` 优先使用登录用户，未登录兜底 demo `user_id=1`；后台管理员 user 1 / `super_admin` 可管理全部，`community_moderators` 中启用用户只能管理对应子站。seed 中 user 2 是 PHP 版主，user 3 是 Go 版主。
- Topic 精华：已完成。`POST /api/v1/admin/topics/:id/feature` toggle `topics.is_featured`，前台列表和详情数据展示“精华”，`sort=featured` 只返回精华内容。
- Topic 置顶：已完成。`POST /api/v1/admin/topics/:id/pin` toggle `topics.is_pinned`，普通列表默认置顶优先展示，前台显示“置顶”。
- Topic 隐藏 / 恢复：已完成。`POST /api/v1/admin/topics/:id/hide` 设置 `status=0`，`restore` 设置 `status=1`；普通列表、搜索和 sitemap 过滤隐藏 topic；后台内容管理仍可看到隐藏内容。
- 评论锁定：已完成。`POST /api/v1/admin/topics/:id/lock-comments` 和 `unlock-comments` 更新 `comment_locked`；前台详情页显示“评论已锁定”，后端创建评论和回复时强制拦截。
- Comment 隐藏 / 恢复：已完成。`POST /api/v1/admin/comments/:id/hide` 设置隐藏，`restore` 恢复正常；普通评论列表过滤隐藏评论。隐藏最佳答案当前被禁止，避免破坏问答采纳闭环。
- 举报处理联动：已完成。接受 topic 举报会隐藏 topic；接受 comment 举报会隐藏评论；如果目标评论是最佳答案，处理会返回错误并保持原状态。
- 后台入口：已完成最小入口。`/admin-next/reports` 提供举报列表、筛选和接受 / 驳回操作；内容管理提供精华、置顶、隐藏、锁定评论操作；评论管理提供隐藏 / 恢复评论操作。
- MemoryStore 支持情况：已完成创建举报、查询举报、处理举报、版主判断、精华、置顶、隐藏 / 恢复 topic、锁定 / 解锁评论、隐藏 / 恢复 comment、普通列表过滤隐藏内容和评论。
- MySQLStore 支持情况：已完成 `reports`、`community_moderators`、`topics.comment_locked` 等 schema 和迁移补齐；支持举报分页、处理、权限判断、topic/comment 治理字段更新、普通列表和搜索过滤隐藏内容，兼容 MySQL 8。
- SEO 保护：已保持。正常 `/topics/:id` 仍由 Go 动态输出 title、description、h1、正文、标签和发布时间；隐藏 topic 返回带 `noindex,follow` 的“内容已隐藏”动态 HTML，不输出原正文；`/sitemap.xml` 不包含隐藏 topic。
- 尚未完成事项：第七轮隐藏最佳答案采用禁止策略；复杂登录 / 权限系统留到后续。

## 第八轮 admin-next 后台 CRUD、版主管理、批量治理和治理审计状态

按 2026-05-09 当前代码，状态如下：

- 后台内容 CRUD：已完成。`GET/POST/PUT/DELETE /api/v1/admin/posts` 路径保持兼容命名，但后台真实读写 `topics`；公开 `sites/posts` 兼容 API 未删除。
- 版主管理 CRUD：已完成。真实接口为 `GET/POST/PUT/DELETE /api/v1/admin/moderators`；删除采用停用策略，将 `community_moderators.status` 置为 `0`。
- 版主权限：已保持轻量实现。`super_admin` 可管理全部；启用版主可按 `community_id` 范围查看和处理自己子站的举报与内容。版主分配目前要求管理员操作。
- 批量 Topic 治理：已完成。`POST /api/v1/admin/topics/batch` 支持 `feature/unfeature/pin/unpin/hide/restore/lock-comments/unlock-comments/delete`。
- 批量 Comment 治理：已完成。`POST /api/v1/admin/comments/batch` 支持 `hide/restore/delete`；隐藏最佳答案仍然禁止。
- 批量举报处理：已完成。`POST /api/v1/admin/reports/batch-handle` 支持批量 `accepted/rejected`，逐条返回成功和失败原因。
- 举报频率限制：已完成最小限制。同一用户对同一对象已有 `pending` 举报时，新举报返回 `同一对象已有待处理举报，请勿重复提交`；已处理后可再次举报。
- 治理审计日志：已完成。新增 `GET /api/v1/admin/audit-logs`，支持 `site/type/action/target/target_type/actor/actor_user_id/community_id/page/page_size`；新增/更新版主、批量治理、举报处理、内容 CRUD 都会写入 `admin_logs`。当前表结构保存 `actor/action/target/site` 文本字段，接口返回的 `actor_user_id`、`target_type`、`target_id`、`community_id` 为派生字段。
- admin-next 页面：已完成。内容管理支持新增、编辑、删除和批量治理；评论管理支持勾选批量处理；举报管理支持批量接受 / 驳回；`/admin-next/moderators` 提供版主管理；`/admin-next/audit-logs` 提供独立治理审计入口；系统设置页仍保留基础日志区。
- MemoryStore 支持情况：已完成第八轮能力。支持版主 CRUD、后台内容 CRUD、批量治理、举报 pending 去重、审计日志筛选分页。
- MySQLStore 支持情况：已完成第八轮能力。支持 `community_moderators` CRUD、`topics` 后台 CRUD、`reports` pending 去重索引、批量治理依赖的单项更新和 `admin_logs` 筛选分页，兼容 MySQL 8。
- SEO 保护：已保持。第八轮未改变 `/topics/:id` Go 动态 SEO HTML；隐藏内容仍不进入普通列表、搜索和 sitemap。
- 第八轮补丁收口：已完成 `admin.js` 版主管理 / 批量治理 / 审计日志封装对账；`Content.vue`、`Comments.vue`、`Reports.vue` 已接入真实批量接口并支持备注；后台 Vite dev proxy、`main.go` 默认端口和文档口径均统一为 `8090`。
- 尚未完成事项：复杂登录系统、细粒度版主分配审批、评论点赞、取消已解决状态不在本轮主线。标签关注和用户关注后端已支持，但前台入口仍作为后续增强。

## 第六轮收尾验收记录

2026-05-09 已完成第六轮收尾验收：

- 启动方式：清理残留进程后，使用 `go build -o .devhub/devhub .` 产出二进制，并以 `PORT=8090 CMS_STORE=memory` 通过 `setsid` 后台启动。
- 8090 状态：`curl -I /`、`GET /api/v1/health`、`GET /api/v1/topics` 均返回 200，当前 8090 有稳定 DevHub 进程监听。
- Gitee 卡住问题：仓库依赖未发现 Gitee 私有模块；已清理残留 `git-upload-pack` 子进程，最终避开 `go run`，使用二进制启动完成验收。
- Go 测试：本轮已执行 `go test ./...` 并通过。
- 前端构建：`./dev.sh --local-go --restart` 已完成 Astro 前台和 Vue 后台构建；本机无 `npm` 时由 Docker Node 构建。
- 评论验收：`GET /api/v1/topics/1/comments` 返回正常；`POST /api/v1/topics/1/comments` 新增评论 ID `9`；`POST /api/v1/topics/1/comments/9/replies` 新增回复 ID `10`；`comment_count`、`last_active_at`、`hot_score` 均更新。
- 采纳验收：topic 1 是 article，采纳返回 400；topic 2 是 question，新增回答 ID `11` 后采纳成功，`is_solved=true`，`best_comment_id=11`，评论 `is_best=true`。
- 未解决筛选：采纳 topic 2 后，`GET /api/v1/search/topics?sort=unsolved` 不再返回 topic 2，只返回未解决 question。
- 动态和通知：`GET /api/v1/me/activities` 返回 `commented` 与 `accepted_answer`；`GET /api/v1/me/notifications` 返回正常。memory 模式当前认证统一为 demo/user 1，本次自操作未新增自通知，符合“不通知自己”规则；非本人通知路径已在 Store 实现中保留。
- SEO 回归：`/topics/1` 源码保留 `<title>`、`meta description`、`<h1>`、`<article>`、正文和标签链接；`/sitemap.xml`、`/robots.txt` 均返回 200。
- 页面回归：`/`、`/search?sort=unsolved`、`/topics/1`、`/me/activities`、`/notifications`、`/admin-next`、`/topics/new`、`/c/php` 均返回 200。

## 已完成

### 项目结构

- 技术栈：Go + Gin、Astro + Vue Islands、Vue 3 + Element Plus、MySQL 8。
- 后端按 `domain / store / service / transport/httpapi` 拆分。
- 保留 `MemoryStore` 和 `MySQLStore` 两种数据仓库。
- 前台入口 `/`，后台入口 `/admin-next`；`/admin` 和 `/admin/:site` 只做兼容重定向。

### 启动脚本

- `dev.sh` 默认端口固定为 `8090`。
- 支持内存模式和 MySQL 模式。
- 无本机 `npm` 时可使用 Docker Node 构建前后台。
- 修改 Go 代码后应使用 `./dev.sh --restart` 或 `./dev.sh restart --no-build` 重启服务。

### 数据模型

当前 Go 内置 schema 和 `db/mysql/001_schema.sql` 包含：

- 兼容模型：`sites`、`boards`、`posts`、`comments`、`tags`、`notifications`。
- 用户和权限模型：`users`、`roles`、`permissions`、`role_permissions`、`user_roles`、`refresh_tokens`。
- 通用社区模型：`communities`、`categories`、`topics`、`topic_tags`、`reactions`、`favorites`、`follows`、`activities`、`reports`、`community_moderators`、`wiki_pages`、`wiki_revisions`。
- 已扩展字段：`comments.topic_id/reply_to_user_id/user_id/content_html/is_best/updated_at/deleted_at`，`topics.is_solved/best_comment_id/comment_count/last_active_at`，`notifications.actor_user_id/type/target_type/target_id/topic_id/comment_id/read_at`，`activities.topic_id/metadata`，`reactions/favorites/follows.updated_at`。

### 后端 API

第五轮新增或增强：

- `POST /api/v1/topics/:id/like`
- `POST /api/v1/topics/:id/favorite`
- `GET /api/v1/topics/:id/interaction`
- `POST /api/v1/follows/toggle`
- `GET /api/v1/me/favorites`
- `GET /api/v1/me/follows`
- `GET /api/v1/me/activities`
- `GET /api/v1/me/notifications`
- `POST /api/v1/me/notifications/:id/read`
- `POST /api/v1/me/notifications/read-all`
- `GET /api/v1/topics/:id/comments`
- `POST /api/v1/topics/:id/comments`
- `POST /api/v1/topics/:id/comments/:commentId/replies`
- `POST /api/v1/topics/:id/comments/:commentId/accept`
- `GET /api/v1/search/topics?sort=unsolved`
- `POST /api/v1/reports`
- `GET /api/v1/admin/reports`
- `GET /api/v1/admin/reports/:id`
- `POST /api/v1/admin/reports/:id/handle`
- `POST /api/v1/admin/topics/:id/feature`
- `POST /api/v1/admin/topics/:id/pin`
- `POST /api/v1/admin/topics/:id/hide`
- `POST /api/v1/admin/topics/:id/restore`
- `POST /api/v1/admin/topics/:id/lock-comments`
- `POST /api/v1/admin/topics/:id/unlock-comments`
- `POST /api/v1/admin/comments/:id/hide`
- `POST /api/v1/admin/comments/:id/restore`

保留兼容接口：

- `POST /api/v1/reactions/toggle`
- `POST /api/v1/favorites/toggle`
- `GET /api/v1/activities`
- `GET /api/v1/notifications`
- `GET /api/v1/notifications/unread-count`
- `POST /api/v1/notifications/:id/read`
- `POST /api/v1/notifications/read-all`
- `sites/posts` 兼容 API 未删除。

### 前台

- 首页、子站页、搜索页、发布页、标签页可用。
- `/topics/:id` 由 Go 动态输出 SEO HTML；交互按钮通过运行时 API 增强。
- 新增 `/me/favorites`、`/me/follows`、`/me/activities`、`/notifications` 页面。
- `/me/notifications` 作为通知中心静态壳别名由 Go 路由 fallback 到同一产物。
- 导航用户菜单新增我的动态、我的收藏、我的关注、通知中心入口。
- 子站页新增关注子站按钮。
- 详情页点赞、收藏、关注主题按钮初始化状态正确，点击后更新状态和计数。
- 详情页评论区支持运行时加载评论、加载更多、发表评论、回复评论；question 类型支持采纳按钮并展示“最佳答案”，非作者默认不显示采纳入口。
- 详情页支持举报 Topic 和举报评论；评论锁定时显示锁定提示并禁用普通提交入口。
- 列表页和搜索页展示置顶、精华、问答状态，隐藏内容不进入普通列表。

### 后台

- `/admin-next` 和 `/admin-next/...` 深层路由可用。
- `/admin`、`/admin/:site` 保持兼容重定向。
- `/admin-next/reports` 提供举报管理入口，支持举报筛选、列表、单条接受 / 驳回和批量接受 / 驳回。
- `/admin-next/content` 提供内容 CRUD、精华、置顶、隐藏 / 恢复、锁定 / 解锁评论和批量治理入口。
- `/admin-next/comments` 提供评论隐藏 / 恢复 / 删除和批量治理入口。
- `/admin-next/moderators` 提供版主管理 CRUD，删除为停用。
- `/admin-next/audit-logs` 提供治理审计列表和筛选。

## 部分完成 / 已预留

- 评论点赞：部分完成。数据字段和旧评论点赞接口保留，本轮未把评论点赞接入详情页运行时评论区。
- 取消已解决状态：已预留。当前支持采纳和更换最佳答案，暂不支持取消已解决。
- 标签关注：后端 `follows` 支持 `target_type=tag`，我的关注页可展示；前台标签区域尚未提供批量关注按钮，后续可在标签聚合页增强。
- 用户关注：后端支持 `target_type=user` 并触发 `user_followed` 通知；前台作者信息区域的关注入口后续完善。

## 已知风险

- 如果后端代码已修改但 `8090` 上已有服务正在运行，普通 `./dev.sh` 可能复用旧服务；应使用 `./dev.sh --restart`。
- Docker Go / Docker Node 构建依赖本机 Docker 权限和网络。
- `/sitemap.xml` 当前最多输出 5000 条 Topic；内容量上来后需要拆分 sitemap index。

## 下一步

1. 第九轮：测试、部署、备份、回滚文档。
2. 后续优化：评论点赞、取消已解决状态、通知跳转 comment anchor。
3. 后台增强：更细粒度角色授权、版主任期记录、治理统计看板。

## 验收清单

- [x] 项目名称保持 DevHub。
- [x] 默认端口固定为 `8090`。
- [x] `main.go` 直接运行默认端口已统一为 `8090`，`PORT` 仍可覆盖。
- [x] 前台 `/` 可打开。
- [x] 后台 `/admin-next` 可打开。
- [x] `/topics/:id` 由 Go 动态输出 SEO HTML。
- [x] 新发布 Topic 后无需重新 build 即可访问 `/topics/:id`。
- [x] `sites/posts` 兼容 API 未删除。
- [x] `POST /api/v1/topics/:id/like` 返回 `liked/like_count/hot_score`。
- [x] `POST /api/v1/topics/:id/favorite` 返回 `favorited/favorite_count/hot_score`。
- [x] `POST /api/v1/follows/toggle` 返回 `followed/target_type/target_id`。
- [x] `/me/favorites` 页面和 `GET /api/v1/me/favorites` 可用。
- [x] `/me/follows` 页面和 `GET /api/v1/me/follows` 可用。
- [x] `/me/activities` 页面和 `GET /api/v1/me/activities` 可用。
- [x] `/notifications` 页面和 `GET /api/v1/me/notifications` 可用。
- [x] 通知单条已读、全部已读可用。
- [x] 点赞、收藏、关注写入动态。
- [x] 点赞、收藏、用户关注触发基础通知。
- [x] MemoryStore 支持本轮互动闭环。
- [x] MySQLStore 支持本轮互动闭环。
- [x] SEO 文档已记录百度 SEO 保护要求。
- [x] `GET /api/v1/topics/:id/comments` 已记录第六轮评论列表。
- [x] `POST /api/v1/topics/:id/comments` 已记录发表评论。
- [x] `POST /api/v1/topics/:id/comments/:commentId/replies` 已记录回复评论。
- [x] `POST /api/v1/topics/:id/comments/:commentId/accept` 已记录采纳最佳答案。
- [x] 未解决筛选 `sort=unsolved` 可用。
- [x] 评论和采纳写入动态与通知。
- [x] `/topics/:id` 评论功能未破坏百度 SEO HTML。
- [x] `POST /api/v1/reports` 可创建举报。
- [x] `/admin-next/reports` 可查看和处理举报。
- [x] `/admin-next/moderators` 可查看、新增、编辑和停用版主。
- [x] `/admin-next/content` 可执行批量 Topic 治理。
- [x] `/admin-next/comments` 可执行批量 Comment 治理。
- [x] `/admin-next/audit-logs` 可查看治理审计日志。
- [x] 管理员可处理全部举报，版主仅可处理自己子站举报。
- [x] Topic 精华、置顶、隐藏、恢复、评论锁定、解锁可用。
- [x] Comment 隐藏、恢复可用，隐藏最佳答案会被拒绝。
- [x] 隐藏 Topic 不进入普通列表、搜索和 sitemap。
- [x] 隐藏 Topic 详情页返回 noindex 的“内容已隐藏”动态 HTML。
