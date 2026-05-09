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

首页、子站页、搜索页和用户中心使用 Astro 静态壳 + 运行时 API；Topic 详情页 `/topics/:id` 仍由 Go 动态输出 SEO HTML。第五轮互动闭环已完成基础实现：点赞、收藏、关注、我的收藏、我的关注、我的动态、通知列表和已读逻辑在 MemoryStore 与 MySQLStore 中均可用。第六轮已完成评论列表、发表评论、回复评论、问答采纳、未解决筛选、评论动态和评论通知的基础闭环。

## 近期迭代摘要

- 第一轮：同步通用社区 schema，补齐 PHP、Go、Java、AI、Frontend 五个子站 seed 数据。
- 第二轮：前台逐步迁移到 `communities/topics/search` API，并保留 `sites/posts` 兼容 API。
- 第三轮：完成发布入口、发布页、`POST /api/v1/topics`，MemoryStore / MySQLStore 均支持创建 Topic。
- 第四轮：增强 `/api/v1/search/topics`，支持关键词、范围、子站、分类、标签、排序、分页组合筛选；新增 `/api/v1/tags/hot`。
- SEO 整改：`/topics/:id` 改由 Go 动态输出 SEO 友好的 HTML；`/sitemap.xml` 和 `/robots.txt` 改为 Go 动态输出。
- 前台登录完善：首页登录/注册接入 `/api/v1/auth/*`，导航支持会话恢复、refresh token 刷新和退出登录。
- 第五轮互动联动：补齐点赞、收藏、关注、我的收藏、我的关注、我的动态、通知中心和详情页互动状态；新增用户中心页面入口。
- 第六轮评论问答：补齐 `GET/POST /api/v1/topics/:id/comments`、回复、采纳最佳答案、`sort=unsolved` 未解决筛选、`commented/accepted_answer` 动态和 `topic_commented/comment_replied/answer_accepted` 通知。

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
- 尚未完成的问题：评论点赞本轮未实现；采纳支持更换最佳答案但暂不支持取消已解决状态；举报入口仅预留；复杂权限、版主治理、后台 CRUD 不在本轮范围。
- 下一轮建议：第七轮优先做举报、版主、精华、置顶、隐藏、评论锁定；第八轮补齐 admin-next 后台 CRUD。

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

### 后台

- `/admin-next` 和 `/admin-next/...` 深层路由可用。
- `/admin`、`/admin/:site` 保持兼容重定向。
- 本轮未扩展后台 CRUD。

## 部分完成 / 已预留

- 评论点赞：部分完成。数据字段和旧评论点赞接口保留，本轮未把评论点赞接入详情页运行时评论区。
- 取消已解决状态：已预留。当前支持采纳和更换最佳答案，暂不支持取消已解决。
- 举报入口：已预留。详情页显示举报按钮，但完整举报、版主、隐藏、锁定评论属于后续治理轮次。
- 标签关注：后端 `follows` 支持 `target_type=tag`，我的关注页可展示；前台标签区域尚未提供批量关注按钮，后续可在标签聚合页增强。
- 用户关注：后端支持 `target_type=user` 并触发 `user_followed` 通知；前台作者信息区域的关注入口后续完善。

## 已知风险

- 如果后端代码已修改但 `8090` 上已有服务正在运行，普通 `./dev.sh` 可能复用旧服务；应使用 `./dev.sh --restart`。
- Docker Go / Docker Node 构建依赖本机 Docker 权限和网络。
- `/sitemap.xml` 当前最多输出 5000 条 Topic；内容量上来后需要拆分 sitemap index。

## 下一步

1. 第七轮：举报、版主、精华、置顶、隐藏、评论锁定。
2. 第八轮：admin-next 后台 CRUD 补齐。
3. 第九轮：测试、部署、备份、回滚文档。
4. 后续优化：评论点赞、取消已解决状态、通知跳转 comment anchor。

## 验收清单

- [x] 项目名称保持 DevHub。
- [x] 默认端口固定为 `8090`。
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
- [ ] 举报、版主治理后续完善。
