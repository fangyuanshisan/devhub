# DevHub 测试文档

[返回文档入口](README.md)

更新时间：2026-05-10

本文档只记录当前 v1.3.0 必测项和后续补测项。历史版本测试只保留必要回归，不再展开旧版本完整矩阵。

## 已实现必测

基础检查：

- `go test ./...`
- `go build` 或 `go build -buildvcs=false ./...`
- `cd web/frontend-app && npm run build`
- `cd web/admin-app && npm run build`
- `bash -n dev.sh`

插件 API：

- `GET /api/v1/plugins` 只返回全局 enabled 插件。
- `GET /api/v1/plugins` 返回统一 manifest 风格的插件声明结构，包括内容类型、权限、菜单、路由和 `config_schema` 预留字段。
- `GET /api/v1/admin/plugins` 返回 `qa`、`docs`、`wiki` 的全局状态。
- `POST /api/v1/admin/plugins/:code/disable` 可以禁用全局插件，并写入审计日志。
- `POST /api/v1/admin/plugins/:code/enable` 可以启用全局插件，并写入审计日志。
- `GET /api/v1/communities/:slug/plugins` 只返回该子站全局 enabled 且子站 enabled 的插件。
- `GET /api/v1/admin/communities/:id/plugins` 返回某个子站的插件状态列表。
- `POST /api/v1/admin/communities/:id/plugins/:code/disable` 可以禁用某个子站插件。
- `POST /api/v1/admin/communities/:id/plugins/:code/enable` 可以启用某个子站插件。
- 全局 disabled 插件不能被子站启用。
- `GET /api/v1/moderator/plugin-menus` 只返回全局 enabled、子站 enabled 且当前用户有权限的插件菜单。

发布与板块：

- `POST /api/v1/topics` 写入归一后的 `content_type` 和 `plugin_code`。
- `doc` 参数保存为 `document`，`wiki` 参数保存为 `wiki_page`。
- 发布 `question` 后可通过 `GET /api/v1/topics/:id/qa` 看到 `qa_questions` 扩展状态。
- 发布 `document` 后可通过 `GET /api/v1/topics/:id/docs` 看到 `docs_documents` 扩展行与基础文档树。
- 发布 `wiki_page` 后可通过 `GET /api/v1/topics/:id/wiki/versions` 看到 `wiki_pages` 扩展行与版本列表。
- 发布权限码映射来自统一 `ContentTypeDefinition` 声明。
- 发布权限校验：
  - `question` 需要 `qa.question.create`
  - `document` 需要 `docs.document.create`
  - `wiki_page` 需要 `wiki.page.create`
  - `project` 需要 `projects.project.create`
  - `job` 需要 `jobs.job.create`
  - `ai_work` 需要 `ai_works.work.create`
- 子站禁用 `qa` 后，该子站不能发布 `question`。
- 其他仍启用 `qa` 的子站可以继续发布 `question`。
- 子站禁用 `docs` 后，该子站不能发布 `document`。
- 子站禁用 `wiki` 后，该子站不能发布 `wiki_page`。
- 板块不能绑定当前子站未启用的插件。
- `category.plugin_code` 与 `content_type` 对应插件不匹配时拒绝发布。
- `content_type` 不在 `category.allowed_content_types` 内时拒绝发布。
- 采纳回答后，`GET /api/v1/topics/:id/qa` 中的 `is_resolved`、`accepted_answer_id` 和回答接受状态会更新。
- 编辑 `wiki_page` 后，`GET /api/v1/topics/:id/wiki/versions` 返回的版本数应增加。

页面与 SEO：

- `/` 可访问。
- `/search/` 可访问。
- `/topics/new/` 可访问。
- `/c/php/` 可访问。
- `/c/php/` 子站页插件导航会按当前子站 enabled 插件显示 / 隐藏问答、文档、Wiki 入口。
- `/topics/:id/` 由 Go 动态输出 SEO HTML。
- `/sitemap.xml` 正常返回。
- `/robots.txt` 正常返回。
- 插件 disabled 后，已有内容详情页仍可访问。

后台与版主体验：

- `/admin-next/plugins` 可查看插件 code/name/version/status/content_types/permissions。
- `/admin-next/plugins` 支持查看插件声明与 `config_schema`。
- `/admin-next/communities` 的插件配置抽屉支持启用 / 禁用、`config_json` 编辑和数字排序。
- `/moderator` 可按当前子站显示插件治理入口；全局 disabled 或子站 disabled 插件不显示。

SEO 源码检查：

```bash
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1|<article|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/'
curl -s http://127.0.0.1:8090/robots.txt
```

## 已实现但后续补测

- 子站插件 `config_json` 接口需要使用真实 admin token 补测。
- 插件声明里的 `config_schema`、`dependencies`、`min_core_version` 和 `hooks` 当前主要是元数据预留，需要继续补测前后台展示或消费场景。
- 子站插件排序接口需要使用真实 admin token 补测。
- 后台子站插件配置 UI 的 `config_json` 编辑、JSON 校验、数字排序、禁用确认和保存后刷新已经有最小实现，需要继续做浏览器矩阵验收。
- 前台发布页按子站插件状态收口需要浏览器验收。
- 子站导航按插件状态显示需要继续做多子站浏览器验收。
- 版主菜单按子站插件状态和权限过滤需要多子站、多版主账号矩阵补测。
- MySQL 老库执行 `db/mysql/migrations/004_community_plugins.sql` 后，需要补测历史板块与历史内容兼容。

## 待实现后补测

- 更细粒度的权限体系补测：例如 Core 兼容类型 `article` / `news` 的细分权限码、按子站/板块维度配置权限矩阵与更明确的错误码（当前发布链路已实现最小权限码校验）。
- Projects / Jobs / AI Works 的专属扩展表、专属管理页和完整业务流程。
- 通用运行时 Hook 调度器，以及关键 Hook 回滚 / 非关键 Hook 记录日志的执行策略。
- Docs 文档树专用编辑 UI、拖拽排序和批量排序。
- Wiki 版本回滚和协作编辑交互。
- QA 取消采纳最佳答案。

## 必要历史回归

- 普通前台会员不能看到总后台入口。
- 前台 user token 与后台 admin token 不能混用。
- `/api/v1/moderator/*` 只接受前台 user token 和有效子站版主授权。
- 标签 alias / merged URL 不进入 sitemap，并能跳转或 canonical 到主标签。
- disabled / merged 标签不进入 sitemap。
- `sites/posts` 兼容 API 继续可用。
- 隐藏 Topic 不进入 sitemap，隐藏详情页带 `noindex,follow`。
