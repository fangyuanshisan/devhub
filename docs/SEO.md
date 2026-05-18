# DevHub SEO 文档

[返回文档入口](README.md)

更新时间：2026-05-13

DevHub 当前 SEO 重点面向百度。核心原则是：`/topics/:id` 必须由 Go 动态输出可被搜索引擎直接读取的 HTML，插件化、前端运行时增强和后台治理都不能把详情页退化成纯 CSR 空壳。

## SEO 在 Core + 插件架构中的定位

SEO 是 DevHub Core 默认内容能力的一部分。当前 Core 负责 `/topics/:id`、`/c/:slug`、标签页、sitemap 和 robots 的稳定输出；后续插件可以扩展 SEO 能力，例如结构化数据、sitemap 扩展、统计代码、站点验证和垂直内容 SEO，但不能绕过 Core 的权限、安全、生命周期和历史内容访问边界。

当前 SEO 仍以默认社区内容页为主。插件 disabled、soft-uninstalled、升级或配置异常都不能破坏历史内容详情页和 Core SEO 兜底。

## 当前策略

- `/topics/:id` 和 `/topics/:id/` 由 Go 动态输出 SEO HTML。
- `/c/:slug` 和 `/c/:slug/` 由 Go 动态输出子站 SEO HTML；`/site/:slug` 兼容 301 到 `/c/:slug/`。
- `/tags/:tag`、`/tags/:tag/`、`/c/:slug/tags/:tag` 和 `/c/:slug/tags/:tag/` 由 Go 动态输出标签 SEO HTML。
- `/sitemap.xml` 和 `/robots.txt` 由 Go 动态输出。
- v1.3.0 起，`qa`、`docs`、`wiki` 作为内置系统插件注册内容类型；插件 disabled 不影响历史 `/topics/:id` SEO 访问。
- v1.3.4 已完成插件迁移失败注入、Hook 失败注入、权限矩阵收口、MySQLStore 升级专项和归档态入口联动验收；插件 disabled、archived、migration failed、Hook failed 或权限拒绝都不能导致历史 `/topics/:id` 详情 404，也不能破坏 `/c/:slug` 子站 SEO。
- v1.3.5 的插件治理 UI、安装向导和升级向导优化仍必须保持同一 SEO 红线。
- v1.4.0-P1-11 的前台入口与菜单可见性治理（navigation/create-options）只影响入口展示与发布可创建选项，不允许破坏历史 `/topics/:id` 与 `/c/:slug` 的 SEO 输出。
- 首页、搜索页、发布页和用户中心类页面可以使用 Astro 静态壳 + 运行时 API。

## Topic 详情页源码要求

`/topics/:id` HTML 源码中必须包含：

- `<title>`
- `meta name="description"`
- `<h1>`
- 子站名称和板块 / 内容类型信息
- 正文摘要或正文
- 标签链接
- canonical `/topics/:id/`
- Article JSON-LD

隐藏 Topic 的详情页应输出“内容已隐藏”HTML，并带 `meta name="robots" content="noindex,follow"`；隐藏页不输出原正文。

## 子站页源码要求

`/c/:slug/` HTML 源码中应包含：

- `<title>`
- `meta name="description"`
- canonical `/c/:slug/`
- `<h1>`
- 子站简介或 slogan
- 启用且可见的板块链接
- 真实 Topic 链接
- 热门标签链接
- 发帖入口链接

## v1.8.1-S2 补充：official_announcement 在子站页的挂载（不破坏 SEO）

说明：`official_announcement` 作为内置官方插件，在 `/c/:slug` 子站 SEO 动态页中通过 **Host + iframe** 方式最小挂载公告展示区，用于验证前端挂载模型。

约束：

1. SEO 关键元素必须保留：`<title>`、`<link rel="canonical">`、JSON-LD、`<h1>` 和主体内容。
2. 插件挂载仅在浏览器运行时执行，不改变服务端输出的核心 SEO HTML 兜底。
3. iframe 仅允许仓库内置页面：`/plugins/official-announcement/iframe`，不允许远程 URL。
4. iframe 使用 `sandbox="allow-scripts"`。
5. 浏览器侧不暴露 callback token / webhook secret。

禁用或归档子站返回不可用 HTML / 404，并带 `noindex,follow`；禁用或归档子站不进入 sitemap。

## 标签页源码要求

标签 SEO 页应包含：

- `<title>`
- `meta name="description"`
- canonical
- `<h1>`
- 标签说明
- 真实 Topic 链接
- 相关标签链接

alias URL 和 merged source URL 优先 301 到主标签 canonical URL，不作为独立 SEO 页。disabled / merged 标签和 alias URL 不进入 sitemap。

## Sitemap 与 Robots

`/sitemap.xml` 当前包含：

- 已发布且 `status=1` 的 Topic。
- 启用子站 canonical `/c/:slug/`。
- 启用全站标签 canonical `/tags/:slug/`。
- 启用子站标签 canonical `/c/:slug/tags/:tag/`。

`/sitemap.xml` 当前不包含：

- 被隐藏或删除的 Topic。
- 禁用或归档子站。
- disabled 标签。
- merged 标签。
- alias URL。

`/robots.txt` 应允许正常抓取，并指向 sitemap。

## 插件化 SEO 边界

- 禁用全局插件或子站插件只影响新发布、导航、菜单和管理入口。
- 禁用插件不能导致历史内容详情页 404。
- `question`、`document`、`wiki_page` 类型的历史 Topic 仍由 Core `/topics/:id` 动态页输出 SEO HTML。
- 插件专用 UI 后续增强时，不得替换或破坏 `/topics/:id` 的 Go 动态 SEO 输出。

当前增强与限制：

- `question` 当前仍复用通用 Article JSON-LD；`QAPage` 结构化数据属于后续增强项。
- `document` 当前仍复用通用 Article JSON-LD；`BreadcrumbList` 属于后续增强项。
- `wiki_page` 当前仍复用通用 Article JSON-LD；`TechArticle` 与版本元信息属于后续增强项。

## 回归命令

```bash
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1|<article|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/c/php/ | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'
curl -s http://127.0.0.1:8090/tags/laravel/ | rg '<title>|description|canonical|<h1|/topics/'
curl -s http://127.0.0.1:8090/c/php/tags/laravel/ | rg '<title>|description|canonical|<h1|/topics/|/c/php/'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/'
curl -s http://127.0.0.1:8090/robots.txt
```

## 禁止退化项

- 不要把 `/topics/:id` 改成只返回前端 app 空壳。
- 不要让详情页必须依赖浏览器 JS 才能看到标题、描述、正文和标签。
- 不要让插件 disabled 导致历史内容详情页不可访问。
- 不要让隐藏内容继续进入 sitemap。
- 不要在隐藏页输出被隐藏 Topic 的原正文。

## v1.6.0-P0-04 远程插件索引 SEO 影响确认

本轮只新增后台远程插件索引只读镜像能力，不修改前台内容页、分类页、搜索页、路由渲染或动态 SEO 逻辑。`/topics/:id/`、`/c/:slug/`、canonical、Article JSON-LD 与历史插件内容访问边界保持 v1.5/v1.6 既有口径。

## v1.6.0 插件包上传与分发前置能力收口 SEO 验收

v1.6.0 主要改动集中在后台插件包治理和插件分发前置能力，不应改变前台 SEO 基础边界。总验收仍需确认：

- `/topics/1/` 保持动态 HTML，包含 title、canonical、Article JSON-LD、article / h1。
- `/c/php/` 保持分类页 SEO，包含 title、description、canonical、h1、内容列表或标签聚合信息。
- 插件 disabled / archived / config_invalid / migration_failed / dependency_missing 不应破坏历史内容详情访问。
- hidden / deleted / pending / rejected 内容仍不能公开访问。

本轮实际 curl 结果记录在 `docs/TESTING.md` 与 `docs/releases/v1.6.0.md` 的 v1.6.0-P1-10 总验收章节。
