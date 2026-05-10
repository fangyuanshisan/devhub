# DevHub 测试文档

[返回文档入口](README.md)

更新时间：2026-05-10

本文档只记录当前 v1.3.x 必测项和后续补测项。历史版本测试只保留必要回归，不再展开旧版本完整矩阵。

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
- `GET /api/v1/plugins` 和 `GET /api/v1/communities/:slug/plugins` 不暴露 `config_json` / `resolved_config`。
- `GET /api/v1/admin/plugins` 返回 `qa`、`docs`、`wiki` 的全局状态。
- `PUT /api/v1/admin/plugins/:code/config` 可以保存合法 JSON，非法 JSON 应返回 400，并写入审计日志。
- 插件启停、全局配置、子站启停、子站配置和子站排序审计日志应包含 `old_value`、`new_value`、`metadata_json` 结构化字段。
- `POST /api/v1/admin/plugins/:code/disable` 可以禁用全局插件，并写入审计日志。
- `POST /api/v1/admin/plugins/:code/enable` 可以启用全局插件，并写入审计日志。
- `GET /api/v1/communities/:slug/plugins` 只返回该子站全局 enabled 且子站 enabled 的插件。
- `GET /api/v1/admin/communities/:id/plugins` 返回某个子站的插件状态列表。
- `POST /api/v1/admin/communities/:id/plugins/:code/disable` 可以禁用某个子站插件。
- `POST /api/v1/admin/communities/:id/plugins/:code/enable` 可以启用某个子站插件。
- `PUT /api/v1/admin/communities/:id/plugins/:code/config` 可以保存合法 JSON，非法 JSON 应返回 400，并写入审计日志。
- `PUT /api/v1/admin/communities/:id/plugins/sort` 可以调整排序，并写入审计日志。
- 全局 disabled 插件不能被子站启用。
- `GET /api/v1/moderator/plugin-menus` 只返回全局 enabled、子站 enabled 且当前用户有权限的插件菜单。
- 插件 manifest 结构一致性测试覆盖 `code/name/version/is_system/content_types/permissions/menus/routes/config_schema/min_core_version`。
- `doc -> document`、`wiki -> wiki_page` 和 `content_type -> plugin_code` 映射测试应通过。
- 插件配置合并测试覆盖默认配置、全局配置、子站配置和 `effective` 生效值。

发布与板块：

- `Service.CreatePost` 不再能绕过插件校验；业务创建必须走 `Service.CreateTopic`。
- `POST /api/v1/posts`、`PUT /api/v1/posts/:id`、`DELETE /api/v1/posts/:id` 写接口返回 `410 Gone` 或明确废弃。
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

后台兼容入口：

- `POST /api/v1/admin/posts` 创建 `question` 时需要 `qa.question.create`。
- `POST /api/v1/admin/posts` 创建 `document` 时需要 `docs.document.create`。
- `POST /api/v1/admin/posts` 创建 `wiki_page` 时需要 `wiki.page.create`。
- 缺少对应插件权限时返回 `403`。
- 禁用 `qa` 后，`POST /api/v1/admin/posts` 不能创建 `question`。
- 禁用 `docs` 后，`POST /api/v1/admin/posts` 不能创建 `document`。
- 禁用 `wiki` 后，`POST /api/v1/admin/posts` 不能创建 `wiki_page`。
- 后台编辑内容不能绕过 `allowed_content_types`。
- 后台编辑内容不能绕过插件 enabled 状态。
- v1.3.1 起禁止后台普通编辑修改内容归属：修改 `site`、`board`、`content_type` 或 `plugin_code` 应失败。
- 已有内容列表、详情、评论、标签、收藏和关注不受影响。

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

- `/admin-next/plugins` 可查看插件 name/code/version/status/content_types/permissions/menus/schema 摘要。
- `/admin-next/plugins` 支持打开插件详情抽屉，按基础信息、内容类型、权限、菜单、配置、路由和 Hooks 分区展示。
- `/admin-next/plugins` 支持查看 `config_schema` 和 `resolved_config`，并明确提示当前 `config_schema` 仅展示 / 预留，保存配置时只做 JSON 合法性校验。
- `/admin-next/plugins` 全局启用 / 禁用有确认提示，并明确 disabled 不影响历史内容详情和 SEO。
- `/admin-next/communities` 的插件配置抽屉支持启用 / 禁用、全局 / 子站双状态、全局禁用原因、`config_json` 编辑、schema 参考、JSON 格式化、JSON 合法性拦截和数字排序保存。
- `/moderator` 可按当前子站显示插件治理入口；全局 disabled 或子站 disabled 插件不显示。

HookBus 最小检查：

- 创建 Topic 时触发 `BeforeCreateContent` 和 `AfterCreateContent`。
- 更新 Topic、置顶、精选、隐藏 / 恢复、评论锁时触发 `BeforeUpdateContent` 和 `AfterUpdateContent`。
- 删除 Topic 时触发 `BeforeDeleteContent` 和 `AfterDeleteContent`。
- 创建评论时触发 `AfterCreateComment`。
- 搜索、通知、Topic SEO 读取分别触发 `OnSearchIndex`、`OnNotificationBuild`、`OnSEOBuild`；当前只要求事件可派发，不要求存在复杂插件处理器。

SEO 源码检查：

```bash
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1|<article|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/'
curl -s http://127.0.0.1:8090/robots.txt
```

## 已实现但后续补测

- 插件全局 `config_json` 和子站 `config_json` 接口已有自动化覆盖，仍需要真实 admin token 做联调补测。
- 插件声明里的 `config_schema`、`dependencies`、`min_core_version` 和 `hooks` 已有结构测试，仍需要继续补测前后台展示或消费场景；`config_schema` 强校验当前不作为通过项。
- 子站插件排序接口已有自动化覆盖，仍需要真实 admin token 和浏览器做联调补测。
- 后台插件管理 UI 已增强为详情抽屉 + 子站配置抽屉；本仓库当前未引入自动浏览器测试，需要按下方手工矩阵继续验收。
- 前台发布页按子站插件状态收口需要浏览器验收。
- 子站导航按插件状态显示需要继续做多子站浏览器验收。
- 版主菜单按子站插件状态和权限过滤需要多子站、多版主账号矩阵补测。
- MySQL 老库执行 `db/mysql/migrations/004_community_plugins.sql` 后，需要补测历史板块与历史内容兼容。

## 待实现后补测

- 更细粒度的权限体系补测：例如 Core 兼容类型 `article` / `news` 的细分权限码、按子站/板块维度配置权限矩阵与更明确的错误码（当前发布链路已实现最小权限码校验）。
- Projects / Jobs / AI Works 的专属扩展表、专属管理页和完整业务流程。
- P0：HookBus 的完整业务处理器、关键 Hook 事务回滚、非关键 Hook 统一错误日志和重试策略。
- P0：`config_schema` 基础强校验。
- P1：`config_schema` 配置表单自动渲染。
- Docs 文档树专用编辑 UI、拖拽排序和批量排序。
- Wiki 版本回滚和协作编辑交互。
- QA 取消采纳最佳答案。

## 完整插件系统平台验收矩阵

P0 已实现或必测：

- Manifest 字段一致性：`code/name/version/is_system/content_types/permissions/menus/routes/config_schema/min_core_version/hooks`。
- `content_type -> plugin_code` 映射：`question/docs/wiki_page/project/job/ai_work` 均映射到对应插件，`doc/wiki` 能归一。
- 插件全局 enabled / disabled：全局 disabled 后不能新发布对应内容，历史详情仍可访问。
- 子站 enabled / disabled：子站 disabled 后仅该子站不能新发布对应内容，其他子站不受影响。
- 板块绑定：`categories.plugin_code` 与 `allowed_content_types` 不匹配时拒绝发布。
- 权限码校验：发布时使用 ContentTypeDefinition 中的 create_permission。
- ActorContext 来源可信：由服务端 token / admin / moderator scope 计算，客户端请求体不能伪造。
- 菜单过滤：前台、后台、版主菜单按全局状态、子站状态、权限和 scope 过滤。
- `config_json` 合法性：非法 JSON 保存失败。
- `admin_logs` old/new/metadata：插件启停、配置、排序写入 `old_value`、`new_value`、`metadata_json`。
- disabled 插件历史内容访问：`/topics/:id` 不返回 404。
- `/topics/:id` SEO 动态 HTML：title、description、h1、article、标签、JSON-LD 不丢失。
- migration 新装 / 老库升级：`001_schema.sql`、`internal/store/schema.go` 和 migrations 字段口径一致。

P0 待实现 / 待补测：

- `config_schema` 基础校验：已接入（简化 JSON Schema），仍需用真实浏览器矩阵补测错误提示与边界值。
- HookBus 业务处理器：Create / Update / Delete / Search / Notification / SEO 不仅能派发事件，还要具备插件处理器、错误日志和失败策略验收。
- 插件 migration runner：当前新增 `plugin_migrations` 表用于记录执行状态，完整 runner 与后台一键执行仍待后续。
- 完整真实 token 验收矩阵：全局禁用、子站禁用、跨子站发布、版主菜单、历史 SEO。

P1 / P2 / P3 后续验收：

- P1：schema 自动表单、插件 SDK 文档、插件生成模板、依赖检查、版本兼容检查、插件搜索 / 通知 / SEO 扩展。
- P2：本地插件包、插件安装、插件升级、soft uninstall、插件 migration runner、插件包签名校验、插件市场雏形。
- P3：远程插件市场、在线更新、动态加载能力评估、插件沙箱、插件权限隔离。

## 必要历史回归

- 普通前台会员不能看到总后台入口。
- 前台 user token 与后台 admin token 不能混用。
- `/api/v1/moderator/*` 只接受前台 user token 和有效子站版主授权。
- 标签 alias / merged URL 不进入 sitemap，并能跳转或 canonical 到主标签。
- disabled / merged 标签不进入 sitemap。
- `sites/posts` 兼容 API 继续可用。
- 隐藏 Topic 不进入 sitemap，隐藏详情页带 `noindex,follow`。

## 后台插件管理 UI 手工验收矩阵

当前没有自动浏览器测试 runner，以下步骤需要在 `/admin-next/plugins` 和 `/admin-next/communities` 手工执行，并把结果回填到当前任务验收记录：

1. 使用后台管理员登录 `/admin-next`，进入“系统插件”，确认顶部说明卡片说明禁用不影响历史内容。
2. 确认全局插件列表展示插件名称、code、版本、system 标识、全局状态、内容类型、权限数量、菜单数量和 schema 状态。
3. 点击“详情”，确认插件详情抽屉可打开，并能查看基础信息、内容类型、权限、菜单、配置、路由和 Hooks tabs。
4. 在详情抽屉的“配置”tab 中确认 `config_schema` 与 `resolved_config` 使用 JSON 代码块展示，复制按钮可用或有明确失败提示。
5. 点击全局“配置”，输入合法 JSON，例如 `{"mode":"test"}`，保存后刷新列表，确认全局配置持久化。
6. 在全局“配置”中输入非法 JSON，例如 `{bad`，确认前端阻止保存并提示 `config_json 不是合法 JSON`。
7. 全局禁用任一测试插件，确认二次确认文案说明所有子站不能新发布、入口隐藏、历史内容和 SEO 不受影响；再启用恢复。
8. 进入“子站管理”，打开任一子站的“插件配置”抽屉，确认列表展示插件名称、code、版本、全局状态、子站状态、内容类型和 `sort_order`。
9. 禁用子站插件，确认二次确认文案说明只影响当前子站的新发布、导航、发布入口和版主菜单，不影响历史内容和 SEO。
10. 启用子站插件，确认全局 enabled 时可恢复子站 enabled；若先在全局插件页禁用该插件，再回到子站启用应被禁用并显示“该插件已被全局禁用，不能在子站启用”。
11. 打开子站“配置”，确认能看到 schema 参考；输入合法 JSON 并保存，刷新后确认 `config_json` 持久化。
12. 再次打开子站“配置”，输入非法 JSON，确认前端阻止保存并提示 `config_json 不是合法 JSON`。
13. 使用上移 / 下移或数字排序调整 `sort_order`，点击“保存排序”，刷新后确认顺序保持。
14. 到对应子站前台和发布页确认禁用插件入口不展示；直接强传对应 `content_type` 发布应由接口拒绝。
