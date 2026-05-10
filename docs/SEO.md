# DevHub SEO 文档

[返回文档大纲](README.md)

更新时间：2026-05-10

DevHub 当前 SEO 重点面向百度。核心原则是：动态 Topic 详情页必须由 Go 输出可被搜索引擎直接读取的 HTML，互动功能只能作为运行时增强，不能破坏源码中的核心内容。

后续 Codex / AI Agent 修改 SEO、前台路由或详情页时，应先阅读 `docs/AGENT_RULES.md`。`/topics/:id` 是阻塞级保护项，不允许退化为纯 CSR 空壳。

## 当前策略

- `/topics/:id` 和 `/topics/:id/` 由 Go 的 `topicSEOPage` 动态输出 SEO HTML。
- `/topics/:id` 不能退化成纯 CSR 空壳，源码中必须直接包含标题、描述、正文、标签等核心内容。
- `/topics/:id` 的动态 HTML 会读取当前 Astro 构建产物中的 `/_astro/*.css` 样式链接，避免详情页引用过期 CSS hash 后无样式。
- 首页、搜索页、用户中心类页面可以使用 Astro 静态壳 + 浏览器运行时 API。
- v1.1.0 起 `/c/:slug` 和 `/c/:slug/` 由 Go 的 `communitySEOPage` 动态输出子站 SEO HTML；`/site/:slug` 兼容 301 到 canonical `/c/:slug/`。
- 点赞、收藏、关注、通知、评论等互动功能属于前端运行时增强。
- 互动按钮、评论运行时加载失败时，也不能影响 `/topics/:id` HTML 源码里的 SEO 内容。
- 第六轮评论区、回复表单、加载更多和最佳答案展示均由浏览器运行时加载；当前不要求评论内容或最佳答案进入初始 SEO HTML。
- 第七轮举报入口、评论锁定提示、精华 / 置顶状态属于运行时或详情页增强，不能改变正常 Topic 的 SEO 主体输出。
- 第八轮 admin-next 内容 CRUD、版主管理、批量治理和审计日志只改变后台 API / 页面；正常 Topic 详情 SEO 仍由 Go 动态输出。
- v1.1.1 前后台身份边界整理只调整认证、token、权限和审计 actor，不改变 `/topics/:id` 或 `/c/:slug` 的 SEO 输出结构。
- v1.1.3 独立版主工作台只新增 `/moderator` 运行时治理页面和 `/api/v1/moderator/*`，不改变 `/topics/:id` 或 `/c/:slug` 的 Go 动态 SEO 输出结构。
- v1.1.4 只修复前台登录态、关注请求、版主入口、后台菜单和发布类型匹配；`/topics/:id` 与 `/c/:slug` 的 SEO 主体仍由 Go 动态输出。
- v1.2.0 起 `/tags/:tag`、`/tags/:tag/`、`/c/:slug/tags/:tag` 和 `/c/:slug/tags/:tag/` 由 Go 动态输出标签聚合 SEO HTML。
- v1.2.1 起，alias URL 和 merged source URL 会优先 301 到 canonical 主标签 URL；disabled / merged / alias URL 不进入 sitemap。
- v1.0.0 归档后，任何上线前回归都必须把 `/topics/:id` SEO 源码检查作为阻塞项。
- 隐藏 Topic 详情页由 Go 输出“内容已隐藏”HTML，并带 `meta name="robots" content="noindex,follow"`；隐藏页不输出原正文。

## Topic 详情页源码要求

`/topics/:id` HTML 源码中应包含：

- `<title>`：当前为 `Topic 标题 - DevHub`。
- `meta name="description"`：来自 `summary` 或正文摘要。
- `<h1>`：Topic 标题。
- 子站名称：来自 `communities`。
- 分类名称：来自 `categories` 或内容类型兜底文案。
- 标签链接：子站 Topic 中的标签指向 `/c/:slug/tags/:tag/`，总站上下文中的标签指向 `/tags/:tag/`。
- 发布时间：当前输出 `created_at`。
- 正文摘要或正文：正文由 Go 转成段落 HTML。
- Article JSON-LD：当前输出 `application/ld+json`。
- canonical：当前指向 `/topics/:id/`。

## 子站页源码要求

v1.1.0 起，`/c/:slug` 是子站 canonical 地址，`/site/:slug` 只做兼容跳转。`/c/:slug` HTML 源码中应包含：

- `<title>`：来自 `communities.seo_title` 或“子站名 技术社区 - DevHub”。
- `meta name="description"`：来自 `seo_description`、简介或 slogan。
- `meta name="keywords"`：来自 `seo_keywords`。
- canonical：指向 `/c/:slug/`。
- `<h1>`：子站标题。
- 子站简介或 slogan。
- 子站板块链接：来自启用且可见的 `categories`。
- 子站内容链接：置顶、精华、最新、热门和未解决问答中的真实 `<a href="/topics/:id/">`。
- 热门标签链接：指向 `/c/:slug/tags/:tag/`。
- 发帖入口链接：`/c/:slug/topics/new/`。
- 版主列表和公告区域：有数据时输出。

禁用或归档子站返回 404 状态码的不可用 HTML，并带 `meta name="robots" content="noindex,follow"`；禁用或归档子站不进入 sitemap。

## 标签页源码要求

v1.2.0 起，`/tags/:tag/` 是全站标签 canonical 地址，`/c/:slug/tags/:tag/` 是子站标签 canonical 地址，均由 Go 动态输出。HTML 源码中应包含：

- `<title>`：来自 `tags.seo_title` 或“标签名 相关内容 - DevHub”。
- `meta name="description"`：来自 `seo_description`、标签说明或默认标签聚合描述。
- `meta name="keywords"`：来自 `seo_keywords`。
- canonical：全站标签指向 `/tags/:slug/`，子站标签指向 `/c/:communitySlug/tags/:slug/`。
- `<h1>`：标签名称。
- 标签说明。
- 标签内容链接：真实 `<a href="/topics/:id/">`。
- 相关标签链接：全站上下文使用 `/tags/:tag/`，子站上下文使用 `/c/:communitySlug/tags/:tag/`。
- 所属子站链接：有子站时指向 `/c/:slug/`。
- 发布入口链接：全站标签使用 `/topics/new/`，子站标签使用 `/c/:slug/topics/new/`。

禁用标签返回不可用页面或 404，不进入 sitemap。v1.2.1 起，alias URL 和 merged source URL 优先 301 到主标签 canonical URL，不作为独立 SEO 页。标签页可以通过 JS 增强关注状态，但初始 SEO HTML 不依赖浏览器 JS。

## 互动功能边界

当前 Topic 详情页 HTML 中有点赞、收藏、评论、举报等操作入口，但它们是运行时增强：

- 点赞真实接口为 `POST /api/v1/topics/:id/like`。
- 收藏真实接口为 `POST /api/v1/topics/:id/favorite`，兼容 `POST /api/v1/favorites/toggle`。
- 评论列表通过 `GET /api/v1/topics/:id/comments` 运行时加载。
- 发表评论使用 `POST /api/v1/topics/:id/comments`，回复使用 `POST /api/v1/topics/:id/comments/:commentId/replies`。
- 问答采纳使用 `POST /api/v1/topics/:id/comments/:commentId/accept`，最佳答案徽标在运行时评论区展示。
- 举报使用 `POST /api/v1/reports`，Topic 和 Comment 举报表单都在浏览器运行时提交。
- 评论锁定由后端 `comment_locked` 字段控制，详情页会显示锁定提示并禁用普通提交入口，但初始正文 SEO 内容不依赖该交互。
- 通知和关注不会影响详情页 HTML 的 SEO 主体内容。

后续修改互动功能时，必须先确认 `curl /topics/:id/` 仍能看到 title、description、h1、正文和标签链接。

## 发布与构建

- 新发布 Topic 后，不需要重新 `npm run build`。
- 新评论、回复、采纳最佳答案后，也不需要重新 `npm run build`；评论区由运行时 API 获取最新状态。
- Go 会通过 API / Store 读取 Topic，并动态输出 `/topics/:id/`。
- `/sitemap.xml` 由 Go 动态输出，包含启用状态子站、启用全站标签、启用子站标签和已发布且 `status=1` 的 Topic。
- 启用子站以 `/c/:slug/` 进入 sitemap，`/site/:slug` 不作为 canonical sitemap URL。
- 启用标签以 `/tags/:slug/` 进入 sitemap；有子站归属的启用标签同时以 `/c/:slug/tags/:tag/` 进入 sitemap。
- disabled 标签、merged 标签和 alias URL 不进入 sitemap。
- 被隐藏的 `status=0` Topic 不进入 `/sitemap.xml`。
- 当前 sitemap 最多输出 5000 条 Topic，内容量继续增长后需要拆分 sitemap index。

## 隐藏内容 SEO

第七轮治理支持隐藏 Topic。当前真实行为：

- 普通列表和搜索只返回 `status=1` 的 Topic。
- `/sitemap.xml` 通过 `TopicsByFilter(..., status=1)` 输出，不包含隐藏 Topic。
- `/topics/:id` 如果读取到隐藏 Topic，不返回纯 CSR，也不输出原正文；Go 会动态输出“内容已隐藏”页面。
- 隐藏页包含 `title`、`meta description`、`h1` 和 `noindex,follow`，用于避免搜索引擎继续索引违规正文。
- 管理员 / 版主查看隐藏内容通过后台 API 和 `/admin-next` 完成，不依赖前台 SEO 页面。

回归隐藏内容时建议执行：

```bash
curl -s http://127.0.0.1:8090/topics/<隐藏ID>/ | rg '内容已隐藏|noindex'
curl -s http://127.0.0.1:8090/sitemap.xml | rg "/topics/<隐藏ID>/" || true
```

## Robots

`/robots.txt` 由 Go 动态输出：

```text
User-agent: *
Allow: /

Sitemap: <站点绝对地址>/sitemap.xml
```

robots 不应屏蔽必要的 CSS、JS、图片资源。当前 Go 静态托管路径包括：

- `/_astro`
- `/frontend-assets`
- `/admin-next/assets`

## 页面类型

需要 SEO 兜底：

- `/topics/:id`
- `/c/:site`
- `/tags/:tag`
- `/c/:slug/tags/:tag`

可以使用 Astro 静态壳 + 运行时 API：

- `/`
- `/search`
- `/topics/new`
- 后续用户中心页面，如 `/me/favorites`、`/me/follows`、`/me/activities`、`/notifications`

说明：列表页和搜索页当前仍可浏览，但百度 SEO 的重点兜底放在详情页。若后续要求列表页强 SEO，可继续由 Go 输出更多服务端 HTML。

## 回归命令

```bash
curl -s http://127.0.0.1:8090/topics/101/ | rg '<title>|description|<h1|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/c/php/ | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'
curl -s http://127.0.0.1:8090/tags/laravel/ | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'
curl -s http://127.0.0.1:8090/c/php/tags/laravel/ | rg '<title>|description|canonical|<h1|/topics/|/c/php/'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/c/php/'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/tags/'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/c/php/tags/'
curl -s http://127.0.0.1:8090/robots.txt
```

发布新 Topic 后回归：

```bash
curl -s http://127.0.0.1:8090/api/v1/topics?page_size=5
curl -s http://127.0.0.1:8090/topics/<新ID>/
curl -s http://127.0.0.1:8090/sitemap.xml | rg "/topics/<新ID>/"
```

## 禁止退化项

- 不要把 `/topics/:id` 改成只返回前端 app 空壳。
- 不要让详情页必须依赖浏览器 JS 才能看到标题、描述、正文和标签。
- 不要让互动接口异常阻断详情 HTML 输出。
- 不要把收藏、关注、通知、评论、最佳答案等运行时互动内容写成详情页 SEO 依赖项。
- 不要让隐藏内容继续进入 sitemap。
- 不要在隐藏页输出被隐藏 Topic 的原正文。
- 不要把 `/tags/:tag` 改回 Astro 预生成列表，也不要依赖 `getStaticPaths` 生成标签详情页。

## 后续 SEO 规划

- sitemap 分片和 sitemap index。
- 更完整的列表页服务端摘要输出。
- 标签趋势统计和运营分析页的 SEO 策略。
