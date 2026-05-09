# DevHub SEO 文档

[返回文档大纲](README.md)

更新时间：2026-05-09

DevHub 当前 SEO 重点面向百度。核心原则是：动态 Topic 详情页必须由 Go 输出可被搜索引擎直接读取的 HTML，互动功能只能作为运行时增强，不能破坏源码中的核心内容。

## 当前策略

- `/topics/:id` 和 `/topics/:id/` 由 Go 的 `topicSEOPage` 动态输出 SEO HTML。
- `/topics/:id` 不能退化成纯 CSR 空壳，源码中必须直接包含标题、描述、正文、标签等核心内容。
- `/topics/:id` 的动态 HTML 会读取当前 Astro 构建产物中的 `/_astro/*.css` 样式链接，避免详情页引用过期 CSS hash 后无样式。
- 首页、子站页、搜索页、用户中心类页面可以使用 Astro 静态壳 + 浏览器运行时 API。
- 点赞、收藏、关注、通知、评论等互动功能属于前端运行时增强。
- 互动按钮、评论运行时加载失败时，也不能影响 `/topics/:id` HTML 源码里的 SEO 内容。
- 第六轮评论区、回复表单、加载更多和最佳答案展示均由浏览器运行时加载；当前不要求评论内容或最佳答案进入初始 SEO HTML。
- 第七轮举报入口、评论锁定提示、精华 / 置顶状态属于运行时或详情页增强，不能改变正常 Topic 的 SEO 主体输出。
- 第八轮 admin-next 内容 CRUD、版主管理、批量治理和审计日志只改变后台 API / 页面；正常 Topic 详情 SEO 仍由 Go 动态输出。
- 隐藏 Topic 详情页由 Go 输出“内容已隐藏”HTML，并带 `meta name="robots" content="noindex,follow"`；隐藏页不输出原正文。

## Topic 详情页源码要求

`/topics/:id` HTML 源码中应包含：

- `<title>`：当前为 `Topic 标题 - DevHub`。
- `meta name="description"`：来自 `summary` 或正文摘要。
- `<h1>`：Topic 标题。
- 子站名称：来自 `communities`。
- 分类名称：来自 `categories` 或内容类型兜底文案。
- 标签链接：当前指向 `/search/?tag=...&scope=community&community_slug=...`。
- 发布时间：当前输出 `created_at`。
- 正文摘要或正文：正文由 Go 转成段落 HTML。
- Article JSON-LD：当前输出 `application/ld+json`。
- canonical：当前指向 `/topics/:id/`。

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
- `/sitemap.xml` 由 Go 动态输出，包含已发布且 `status=1` 的 Topic。
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

可以使用 Astro 静态壳 + 运行时 API：

- `/`
- `/c/:site`
- `/site/:site`
- `/search`
- `/tags/:tag`
- `/topics/new`
- 后续用户中心页面，如 `/me/favorites`、`/me/follows`、`/me/activities`、`/notifications`

说明：列表页和搜索页当前仍可浏览，但百度 SEO 的重点兜底放在详情页。若后续要求列表页强 SEO，可继续由 Go 输出更多服务端 HTML。

## 回归命令

```bash
curl -s http://127.0.0.1:8090/topics/101/ | rg '<title>|description|<h1|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/'
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
