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
- `/admin-next/plugins` 支持查看 `config_schema` 和 `resolved_config`，全局插件配置使用 JSON Editor + Ajv 做客户端基础校验，后端保存时继续按简化 `config_schema` 二次校验。
- `/admin-next/plugins` 全局启用 / 禁用有确认提示，并明确 disabled 不影响历史内容详情和 SEO。
- `/admin-next/communities` 的插件配置抽屉支持启用 / 禁用、全局 / 子站双状态、全局禁用原因、`config_json` 编辑、schema 参考、JSON Editor、Ajv 基础校验和数字排序保存。
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

## 后台 E2E Docker Runner

后台 Playwright E2E 使用项目内固定 Docker 镜像，不依赖宿主机 Node/npm：

```bash
docker compose build admin-e2e
docker compose run --rm admin-e2e npm run build
docker compose run --rm admin-e2e
```

说明：

- `admin-e2e` 使用 `web/admin-app/Dockerfile.e2e`，基础镜像固定为 `mcr.microsoft.com/playwright:v1.59.1-noble`。
- 镜像构建阶段执行 `npm ci`；运行阶段复用镜像依赖，并通过 Docker volume 避免把 `node_modules` 写入仓库。
- 默认测试目标是 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8090`，运行 E2E 前需要先启动 DevHub 后端服务。
- E2E 前建议先执行 `docker compose run --rm admin-e2e npm run build`，让 Go 服务读取最新 `web/admin-vue` 静态产物；不要和 E2E 并行写 `web/admin-vue`。
- 首次构建会拉取较大的 Playwright 基础镜像，后续重复执行会复用本地 `sns-admin-e2e` 镜像。

当前已自动化覆盖：

- 后台登录与保护页面会话边界。
- `/admin-next/content` 内容管理打开与标题搜索。
- `/admin-next/comments` 评论管理打开与筛选。
- `/admin-next/communities` 子站管理打开与搜索。
- `/admin-next/tags` 标签管理打开与搜索。
- `/admin-next/audit-logs` 治理审计打开与动作筛选。
- `/admin-next/plugins` 打开与搜索筛选。
- 插件详情 Tabs、`config_schema` / `resolved_config` 展示和 schema 错误提示。
- 全局禁用确认、impact 提示和全局 disabled 后子站启用限制。
- 子站插件配置抽屉、JSON Editor 与 Ajv 错误提示。
- 通用 `PluginContent` 页入口、子站筛选和状态筛选。

## 前台 E2E Docker Runner

前台 Playwright E2E 使用项目内固定 Docker 镜像，不依赖宿主机 Node/npm：

```bash
docker compose build frontend-e2e
docker compose run --rm frontend-e2e npm run build
docker compose run --rm frontend-e2e
```

说明：

- `frontend-e2e` 使用 `web/frontend-app/Dockerfile.e2e`，基础镜像固定为 `mcr.microsoft.com/playwright:v1.59.1-noble`。
- 镜像构建阶段执行 `npm ci`；为适应慢网环境，Dockerfile 为 `npm ci` 增加了 fetch retry / timeout 参数。
- 运行阶段通过 Docker volume 复用依赖，不把 `node_modules` 写入仓库。
- 默认测试目标是 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8090`，运行 E2E 前需要先启动 DevHub 后端服务。
- E2E 前建议先执行 `docker compose run --rm frontend-e2e npm run build`，让 Go 服务读取最新 `web/frontend` 静态产物。

当前已自动化覆盖：

- `/` 总站首页打开，基础导航可见，游客看不到总后台入口。
- `/c/php/` 和 `/c/go/` 子站首页打开，并检查 canonical 基础 SEO 元素。
- `/search/` 关键词搜索提交，结果区域可见。
- `/topics/1/` 动态 Topic 详情打开，包含 h1、article 和 JSON-LD。
- `/topics/new/` 未登录发布拦截，提交后提示登录。
- `/tags/go/` 与 `/c/php/tags/laravel/` 标签页打开，并检查 canonical 基础 SEO 元素。

## 已实现但后续补测

- 插件全局 `config_json` 和子站 `config_json` 接口已有自动化覆盖，仍需要真实 admin token 做联调补测。
- 插件声明里的 `config_schema`、`dependencies`、`min_core_version` 和 `hooks` 已有结构测试，仍需要继续补测前后台展示或消费场景；`config_schema` 强校验当前不作为通过项。
- 子站插件排序接口已有自动化覆盖，仍需要真实 admin token 和浏览器做联调补测。
- 后台插件治理中心和核心后台页面已有 Docker 化 Playwright E2E runner；更大矩阵（多浏览器、多账号、多子站和视觉细节）仍需要按下方手工矩阵或后续扩展 E2E 补测。
- 前台已有 Docker 化 Playwright 冒烟 E2E runner；登录互动、发布成功、插件启停联动、版主工作台、多账号权限边界仍待后续自动化。
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

以下步骤用于补充 Docker 化 Playwright 最小 E2E 尚未覆盖的浏览器矩阵；执行后需要把结果回填到当前任务验收记录：

1. 使用后台管理员登录 `/admin-next`，进入“系统插件”，确认顶部说明卡片说明禁用不影响历史内容。
2. 在 `/admin-next/plugins` 顶部确认统计卡片展示：全部、enabled、disabled、system、有 schema 的插件数量（与列表筛选一致）。
3. 在 `/admin-next/plugins` 使用筛选工具栏分别验证：
   - 按 code/name 搜索
   - 按 status 筛选 enabled/disabled
   - 按 content_type 筛选（列表只保留包含该 content_type 的插件）
   - 按 system 筛选
   - 按 schema 筛选
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

## 插件详情抽屉 Tabs + JSON Editor 手工验收

以下步骤用于验收“插件详情抽屉 Tabs + JSON 配置编辑器”增强。基础路径已由 `admin-e2e` 覆盖，复制、保存持久化和视觉细节仍需手工补测：

1. 后台管理员登录 `/admin-next/plugins`，点击任一插件“详情”，确认抽屉打开，并包含 Tabs：
   - 概览 / 内容类型 / 权限 / 菜单 / 配置 / Hooks / 路由。
2. 点击列表“权限/菜单/配置”快捷按钮，确认能直接打开抽屉并切换到对应 Tab。
3. 在“权限”Tab：
   - 输入权限码关键字搜索，表格过滤生效；
   - 点击“复制”能复制权限码，或在不支持剪贴板时有明确提示。
4. 在“菜单”Tab：确认显示 area/title/path/permission/sort_order，并提示“菜单展示仍受全局/子站状态和权限过滤影响”。
5. 在“配置”Tab：确认同时展示：
   - `config_schema`（只读 JSON）
   - `config_json`（JSON Editor 可编辑）
   - `resolved_config`（只读 JSON）
6. 在 JSON Editor 中修改 `config_json` 为合法 JSON（例如 `{}` 或 `{ "example": true }`），点击保存：
   - 若 `config_schema` 存在：应先通过 Ajv 校验；
   - 保存成功后关闭抽屉再打开，配置值应保持；
   - 后端应仍做二次校验并写审计日志（见 admin_logs）。
7. 模拟 schema 校验失败（例如 schema 要求必填字段但配置缺失）：
   - 页面应显示 schema 错误列表；
   - “保存”按钮应被禁用或保存失败有明确提示。
8. 验证辅助操作：
   - 点击“格式化”保持 JSON 值可读；
   - 点击“清空为 {}”把配置重置为空对象；
   - 点击“复制”复制当前 JSON。
9. 在 “Hooks”Tab：确认不伪造 handler 状态；“平台调用点”列仅在可确认接入时显示“存在”，其余显示“未知/未覆盖”。

## 子站插件配置抽屉升级 手工验收

以下步骤用于验收“子站插件配置抽屉升级”（`/admin-next/communities` -> 子站 -> “插件”）：

1. 后台管理员登录 `/admin-next/communities`，打开任一子站的“插件配置”抽屉，确认顶部展示：
   - 子站名称、slug；
   - 子站 enabled 插件数量、子站 disabled 插件数量、全局 disabled 插件数量。
2. 验证筛选能力：
   - 切换：全部 / 子站已启用 / 子站未启用 / 全局已禁用；
   - 按 name/code 搜索；
   - 按 content_type 筛选（只保留包含该 content_type 的插件）。
3. 列表字段检查：确认展示插件名称、code、全局状态、子站状态、content_types、sort_order、配置覆盖状态与操作按钮。
4. 全局 disabled 限制：对全局 disabled 的插件，“启用”按钮应禁用；并在说明中明确“该插件已被全局禁用，不能在子站启用”。
5. 子站 disabled 提示：对子站 disabled 的插件，说明中应提示“当前子站未启用，不能新发布对应内容”。
6. 配置编辑：点击某插件“配置”，确认弹窗内：
   - `config_schema` 只读展示；
   - `config_json` 使用 JSON Editor；
   - `resolved_config` 只读展示（如接口返回）。
7. Ajv 校验：当 schema 校验失败时应显示错误，并禁止保存；修复错误后可保存。
8. 保存配置后刷新：保存成功后，关闭弹窗并刷新列表；再次打开该插件配置，配置值应保持。
9. JSON 辅助操作：点击“清空为 {}”应将配置置为空对象；然后点击保存应写入后端。
10. 排序能力：通过 input-number 或上移/下移调整顺序，点击“保存排序”，刷新后顺序保持。

## 插件治理中心：影响分析与审计入口 手工验收

1. 在 `/admin-next/plugins` 全局禁用任一插件（测试用），确认禁用确认弹窗中：
   - 若 impact 接口可用，会显示“影响子站/影响板块/已有内容/审核中内容/菜单声明”等计数；
   - 若 impact 接口不可用，会显示“影响范围统计待后端接口支持或当前环境暂不可用”，不出现假数字。
2. 在 `/admin-next/communities` -> 子站 -> “插件配置”中禁用某个子站插件，确认禁用确认弹窗中：
   - 若 impact 接口可用，会显示该子站范围内的板块数/已有内容数等计数；
   - 不影响历史内容与 SEO 的提示仍存在。
3. 在 `/admin-next/plugins` 打开任一插件详情抽屉，进入“审计”Tab：
   - 默认按 `plugins#<plugin_code>` 前缀筛选；
   - 可按“动作关键字”进一步筛选；
   - 可输入 `community_id` 限定子站范围；
   - 能查看 old_value/new_value/metadata_json（如为空表示该条日志未结构化，不伪造）。

## PluginContent（插件内容页）手工验收

1. 在 `/admin-next/plugins` 点击某插件的“管理”按钮进入插件内容页（`/qa`、`/docs` 等）。
2. 页面顶部显示：
   - 当前 plugin_code、content_type 与插件状态 badge；
   - 禁用不影响历史内容访问的提示。
3. 子站筛选：
   - 选择一个子站后，列表应切换到该子站内容（通过 `admin/posts?site=<slug>`）。
4. 状态筛选：
   - 切换 publish/hidden/pending，列表筛选应生效（取决于后端 `admin/posts` 支持情况）。
5. 列表中展示 `plugin_code` / `content_type` 字段（若后端返回为空则显示 `-`）。

## 2026-05-10 插件治理中心专项验收记录

本轮执行了命令行、API 和 SEO 层面的集中验收。当时尚未接入自动化浏览器 runner，因此插件治理中心 UI 的点击矩阵按上方手工步骤保留；后续已新增 Docker 化 Playwright 最小 E2E runner。

已执行并通过：

- `go test ./...`
- `go build -o .devhub/devhub .`
- Docker Node 后台构建：`docker run --rm -v "$PWD/web/admin-app":/app -w /app node:20-alpine sh -lc "npm ci && npm run build"`
- 临时服务：`PORT=18090 CMS_STORE=memory ./.devhub/devhub`
- `/admin-next`、`/admin-next/plugins`、`/admin-next/qa` 均返回 200。
- `GET /api/v1/admin/plugins` 可返回插件声明、状态、权限、菜单、`config_schema` 和 `resolved_config`。
- `GET /api/v1/admin/plugins/qa/impact` 返回全局 impact 轻量计数。
- `GET /api/v1/admin/communities/1/plugins/qa/impact` 返回子站 impact 轻量计数。
- `PUT /api/v1/admin/plugins/qa/config` 合法配置返回 200，缺少 required 字段返回 400。
- `PUT /api/v1/admin/communities/1/plugins/qa/config` 类型错误返回 400。
- 全局禁用 `qa` 后，子站启用 `qa` 返回 400，提示插件全局未启用；验收后已恢复 `qa` enabled。
- `GET /api/v1/admin/audit-logs?target=plugins%23qa` 返回插件启停审计，并包含 `old_value`、`new_value`、`metadata_json`。
- 首页源码未暴露 `/admin-next` 总后台入口；版主入口为登录后按权限显示的隐藏入口。
- `GET /api/v1/moderator/plugin-menus?community_slug=php` 使用前台 user token 返回当前子站可见插件菜单。
- `/topics/1/` 返回 SEO HTML，包含 title、description、h1、article、标签链接和 Article JSON-LD。
- 全局禁用 `qa` 后访问已有 question `/topics/2/` 仍返回 SEO HTML；验收后已恢复 `qa` enabled。
- `/c/php/` 返回子站 SEO HTML，包含 title、description、canonical、h1、topic 链接和热门标签。
- `/sitemap.xml` 和 `/robots.txt` 正常返回。

失败 / 跳过项：

- `cd web/admin-app && npm run build` 在宿主机失败，原因是当前环境没有 `npm`；已按项目规则用 Docker Node 完成构建。
- 未执行 `cd web/frontend-app && npm run build`：本轮未修改前台代码。
- 未执行真实浏览器点击矩阵：当时仓库没有 Playwright/Cypress 等自动化 runner；后续已新增 `admin-e2e` 最小 E2E，完整多账号/多子站矩阵仍需扩展。

发现的问题：

- P0：未发现。
- P1：未发现会阻塞插件治理中心基础可用性的后端/API/SEO 问题。
- P2：后台构建存在 Vite chunk size warning，主要来自 `PluginJsonEditor` 等大 chunk；后续可考虑按需加载或手动拆包。

## 2026-05-11：固定 DevHub 后台 E2E Docker 镜像

本轮目标是为后台 E2E 固定项目内 Docker 镜像，避免重复临时拉取大型 Playwright 镜像，并保证本地与 CI 可以复用同一条命令链路。

已执行并通过：

- `docker compose build admin-e2e`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e`

验收结果：

- `admin-e2e` 首次构建会拉取 `mcr.microsoft.com/playwright:v1.59.1-noble`，后续复用本地 `sns-admin-e2e` 镜像。
- 后台构建在容器内完成，不依赖宿主机 `npm` / `node`。
- Playwright E2E 当前 5 条用例全部通过，覆盖插件治理中心核心路径。
- `node_modules`、`web/admin-vue`、`playwright-report` 和 `test-results` 均为忽略产物，不提交仓库。

注意：

- E2E 依赖已启动的 DevHub 服务，默认访问 `http://host.docker.internal:8090`。
- E2E 前需要先用 `admin-e2e` 构建后台静态产物；不要和 E2E 并行写 `web/admin-vue`，否则可能读到短暂不完整的静态资源。
