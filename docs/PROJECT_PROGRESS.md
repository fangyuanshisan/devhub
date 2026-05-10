# DevHub 项目进度

[返回文档入口](README.md)

更新时间：2026-05-10

本文档只记录当前仓库真实状态、当前风险和下一步任务。历史版本能力已并入当前分支，详情见对应 Release Notes；旧版本已解决问题不再占用当前主体。

## 当前版本结论

当前版本为 `v1.3.1`，主题是“插件化关键入口封口与权限校验补强版”。DevHub 当前定位为多子站通用开源社区程序，默认演示为开发者社区。

Core 保留用户、认证、子站、板块、通用内容、评论、标签、搜索、通知、SEO、权限、审计、插件注册和分发能力。问答、文档、Wiki、项目、招聘、AI 作品已按内置系统插件建模：`qa -> question`、`docs -> document`、`wiki -> wiki_page`、`projects -> project`、`jobs -> job`、`ai_works -> ai_work`。

当前实现仍保留历史表名以保证兼容：`topics` 是当前通用内容表，`categories` 是当前通用板块表。

当前最高优先级长期主线是完成完整插件系统。DevHub 的长期目标不是只支持内置 `qa/docs/wiki`，而是形成完整插件平台：Core 只提供通用社区底座，业务能力通过插件声明、插件状态、插件权限、插件菜单、插件配置、插件 Hook、插件 migration、插件 API、插件 SEO、插件通知、插件搜索和插件测试矩阵扩展。

## 当前已完成

- 插件注册：`internal/plugins/registry.go` 和 `internal/plugins/qa|docs|wiki|projects|jobs|aiworks` 提供内置插件定义、内容类型映射、菜单、权限和路由描述。
- 插件声明规范：当前已统一到 manifest 风格声明，包含插件本体、内容类型定义、权限定义、菜单定义、路由定义、`config_schema`、依赖、最小 Core 版本和 Hook 声明。
- 全局插件状态：`plugins` 表和 MemoryStore / MySQLStore 均支持 `installed`、`enabled`、`disabled`，并提供全局插件列表、启用和禁用 API。
- 子站插件状态：`community_plugins` 表和 MemoryStore / MySQLStore 均支持按子站启用 / 禁用、配置和排序插件。
- 两层状态判断：插件在某个子站可用需要同时满足 `plugins.status=enabled` 和 `community_plugins.status=enabled`；`core` 作为兼容内置能力在 Service 层特殊视为可用。
- 内容模型兼容：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 已进入 schema 与 Store。
- 发布校验：`POST /api/v1/topics` 已走统一 `ValidateTopicPluginAccess`，会归一 `doc -> document`、`wiki -> wiki_page`，并校验插件状态、子站插件状态、板块插件绑定和允许内容类型。
- 板块管理校验：MemoryStore / MySQLStore 在创建或编辑子站板块时校验 `plugin_code` 与 `content_type` 匹配，并拒绝绑定全局或子站未启用的插件。
- 插件 API：全局插件 API、子站插件 API、前台子站插件展示 API 和版主插件菜单 API 已在 `router.go` 注册。
- 插件业务闭环：
  - `qa`：发布 `question` 时写入 `qa_questions`；回答写入 `qa_answers`；采纳后回写已解决状态和最佳答案。
  - `docs`：发布 `document` 时写入 `docs_documents`，并支持基础文档树读取。
  - `wiki`：发布 `wiki_page` 时写入 `wiki_pages` 和初始 `wiki_page_versions`；编辑时新增版本记录。
- `project` / `job` / `ai_work` 已完成插件归属迁移：`projects -> project`、`jobs -> job`、`ai_works -> ai_work`，发布校验、权限码、菜单声明和历史 `plugin_code` 迁移口径已接入；专属扩展表和完整业务闭环尚未完成。
- 权限上下文：`CreateTopicRequest.ActorPermissions` / `ActorContext` 均由服务端从 token、后台身份和版主 scope 计算，客户端请求体不能覆盖。
- 配置合并：`plugins.config_json` 与 `community_plugins.config_json` 已落库并可写，返回 `resolved_config.default/global/community/effective` 合并视图。
- HookBus：Service 层已有最小内部 HookBus，当前调用点覆盖内容创建、更新、删除、评论、搜索、通知和 SEO 事件；Search / Notification / SEO 当前是最小事件派发，不做第三方动态执行。
- 前台入口：子站插件公开接口会隐藏 `config_json` / `resolved_config` 等后台配置；子站板块导航会按子站插件状态过滤。
- 后台入口：`/admin-next/plugins` 作为系统插件管理入口；插件业务页通过系统插件列表进入，默认不散落在后台左侧导航。
- 后台插件管理体验：
  - 后台全局插件管理已支持说明卡片、插件状态 badge、内容类型 tag、权限 / 菜单 / schema 摘要、详情抽屉、tabs 分区展示、配置 schema / resolved config JSON 展示与复制、全局配置编辑、启用 / 禁用确认。
  - 后台子站插件配置已支持双状态 badge、子站启用统计、全局禁用原因提示、启用 / 禁用确认、`config_json` 编辑、schema 参考、JSON 格式化、JSON 合法性拦截、数字排序和上移 / 下移后保存。
  - 前台子站页和发布页会按当前子站已启用插件收口入口与内容类型。
  - 版主工作台已补最小插件治理入口区，并按当前子站插件状态与权限过滤。
- 审计：全局插件状态、子站插件状态、全局插件配置、子站插件配置和排序已接入 `admin_logs`，并为插件治理操作写入 `old_value`、`new_value`、`metadata_json` 结构化字段；`target` 文本摘要继续保留用于兼容展示。
- Wiki schema：当前只保留插件化 `wiki_spaces`、`wiki_pages`、`wiki_page_versions` 语义，旧 `wiki_revisions` 预留冲突已清理。
- SEO 保护：`/topics/:id` 仍由 Go 动态输出 SEO HTML，插件禁用不影响历史内容详情访问。
- 技术债收口：`Service.CreatePost` 已封口，不再作为业务写入口；`/api/v1/posts` 写接口继续废弃；后台 `admin/posts` 创建入口在兼容 `post.create` 基础权限之外，叠加真实内容类型对应的插件 create 权限。
- 后台编辑边界：后台内容编辑已禁止修改子站、板块、`content_type` 和 `plugin_code` 归属字段；如后续需要迁移归属，必须走单独迁移专项和完整插件校验。

## 当前部分完成

- 子站插件管理 UI：后台体验已从最小表格增强为更清晰的配置面板，包括全局 / 子站双状态、禁用原因、schema 参考、JSON 格式化、禁用影响提示和排序保存；多浏览器矩阵、批量操作和更强可视化仍待后续专项验收。
- 插件权限：后台菜单和版主菜单已按权限过滤；发布链路已按内容类型做最小权限码校验：
  - `question -> qa.question.create`
  - `document -> docs.document.create`
  - `wiki_page -> wiki.page.create`
  - `project -> projects.project.create`
  - `job -> jobs.job.create`
  - `ai_work -> ai_works.work.create`
  - Core 兼容类型 `article`、`news` 当前仍为粗粒度 `core.topic.create`（兼容旧 `post.create`，后续可按内容类型细化）。
- 权限兼容桥：`post.create` 仍作为 `core.topic.create` 的过渡兼容权限存在；它不是长期主权限，后续需要随 Core 内容定义或 article/news 插件化一起收口。
- Docs / Wiki 业务体验：当前已具备基础空间、文档树读取、版本列表等最小闭环；拖拽排序、完整回滚 UI、协作锁和专用编辑体验仍待后续。
- Projects / Jobs / AI Works 业务体验：当前完成插件归属、发布校验、权限码和菜单声明；专属扩展表、专属管理页和完整业务流程仍待后续。
- 插件路由：当前是注册描述 + Core 分发，不是真正动态运行时加载器。
- Hook 机制：当前已有最小内部 HookBus，并覆盖创建、更新、删除、评论、搜索、通知和 SEO 调用点；Search / Notification / SEO 仍是预留级事件派发，尚未形成完整插件业务处理器、统一错误日志和重试策略。
- 配置校验：当前已完成默认配置、全局配置、子站配置三层合并和 JSON 格式校验；`config_schema` 强校验仍待后续。
- 验收覆盖：已做文档与路由核对；完整 Docker 启动、真实 token API、浏览器页面和 SEO curl 矩阵仍需按测试文档继续补测。

## 当前未完成

- 子站插件配置 UI 的完整浏览器验收矩阵，包括多子站、禁用提示、保存失败提示和排序持久化回归。
- 更细粒度的权限体系：例如 Core 兼容类型 `article` / `news` 的细分权限码、按子站/板块维度配置权限矩阵、以及更明确的错误码与权限配置 API（当前仍为最小校验闭环）。
- P0 插件平台收口：HookBus 的完整业务处理器与日志策略。Search / Notification / SEO 目前已有调用点，但缺少实际插件处理器、统一失败日志和重试策略。
- P0 插件平台收口：`config_schema` 基础校验。
- P1 插件平台增强：`config_schema` 后台自动表单渲染。
- 非插件历史审计日志的结构化 diff：插件治理已写入 `old_value`、`new_value`、`metadata_json`，其他旧审计仍可能只有 `target` 文本。
- `qa` 取消采纳最佳答案。
- Docs 文档树专用编辑 UI。
- Docs 文档树拖拽、批量排序和更完整的空间管理体验。
- Wiki 版本回滚接口与协作编辑交互。
- Projects / Jobs / AI Works 的专属扩展表、专属管理页、专属搜索、通知、SEO 和完整业务闭环。
- P2 插件分发能力：本地插件包、插件安装、插件升级、soft uninstall、插件 migration runner、插件包签名校验和插件市场雏形。
- P3 高级能力：远程插件市场、在线更新、动态加载能力评估、插件沙箱和插件权限隔离。

## 完整插件系统路线

P0：插件平台收口

- Manifest 契约稳定。
- Registry 稳定。
- ActorContext 稳定。
- 权限码平台化。
- 全局插件状态。
- 子站插件状态。
- 板块绑定。
- 发布校验。
- 菜单过滤。
- `config_json`。
- `config_schema` 基础校验。
- HookBus 全调用点和业务处理器。
- `admin_logs` 结构化审计。
- migration 边界。
- 测试矩阵。

P1：插件平台增强

- schema 自动表单。
- 插件 SDK 文档。
- 插件生成模板。
- 插件依赖检查。
- 插件版本兼容检查。
- 插件事件和通知模板。
- 插件搜索索引扩展。
- 插件 SEO 扩展。

P2：插件分发能力

- 本地插件包。
- 插件安装。
- 插件升级。
- 插件禁用。
- soft uninstall。
- 插件 migration runner。
- 插件包签名校验。
- 插件市场雏形。

P3：高级能力

- 远程插件市场。
- 在线更新。
- 动态加载能力评估。
- 插件沙箱。
- 插件权限隔离。

安全红线：

- 禁用插件不能影响历史内容访问。
- 禁用插件不能破坏 `/topics/:id` SEO 动态 HTML。
- Core 表不能被插件随意破坏。
- 插件写操作必须走权限校验。
- ActorContext 必须由服务端生成，不能由客户端伪造。
- 前台 user token、后台 admin token、版主 user token + scope 不能混用。
- 插件配置必须至少保证 JSON 合法。
- 插件 migration 必须有备份和回滚说明。
- 未实现能力不能写成已完成。
- 预留、部分完成、待验收能力必须明确标注。

## 当前风险

- 历史数据可能存在 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 或 `community_plugins` 缺失 / 不一致，生产升级前需要迁移演练和抽样校验。
- 子站插件禁用后已有内容应继续可读；后续改发布、列表或 SEO 时要避免把禁用插件误当作历史内容 404 条件。
- API 和后台 UI 已增强不等于完整产品闭环；子站插件配置、排序、插件详情抽屉和插件入口仍需继续做真实浏览器矩阵验收。
- `/sitemap.xml` 当前仍是单文件动态输出，内容规模扩大后需要 sitemap index / 分片。
- 用户提出的 `docs/BACKUP_ROLLBACK.md` 与仓库真实文件名不一致；当前真实文件是 `docs/BACKUP_AND_ROLLBACK.md`。

## 下一步任务

1. P0：补 `config_schema` 基础校验，至少覆盖 `type`、`enum`、`required` 等最小规则。
2. P0：补 HookBus 真实业务处理器和统一错误日志，优先覆盖 Search / Notification / SEO。
3. P0：用真实 admin/user token 补测全局插件、子站插件、版主菜单和跨子站发布矩阵。
4. P0：完成后台插件管理 UI 浏览器验收：全局插件详情抽屉、全局启用 / 禁用确认、子站 `config_json`、排序、禁用影响提示、失败提示和保存后刷新。
5. P0：补跑 `/topics/:id` SEO curl 检查，确认插件禁用后历史内容源码不退化。
6. P1：细化插件权限矩阵，尤其是 Core 兼容类型 `article` / `news` 的权限码策略。
7. P1：规划插件 SDK 文档和插件生成模板。
8. 如需要后台迁移内容子站、板块或类型，设计单独迁移 API，并逐条校验插件状态、子站插件状态、板块绑定、allowed_content_types 和权限码。

## 当前验收清单

- [ ] `go test ./...`
- [ ] `go build` 或 `go build -buildvcs=false ./...`
- [ ] `cd web/frontend-app && npm run build`
- [ ] `cd web/admin-app && npm run build`
- [ ] `GET /api/v1/plugins` 只返回全局 enabled 插件。
- [ ] `GET /api/v1/communities/:slug/plugins` 只返回全局 enabled 且子站 enabled 插件。
- [ ] 管理员可以查看、启用和禁用全局插件。
- [ ] 管理员可以查看、启用和禁用某个子站插件。
- [ ] 子站禁用 `qa` 后，该子站不能发布 `question`；其他启用 `qa` 的子站不受影响。
- [ ] 子站禁用 `docs` 后，该子站不能发布 `document`。
- [ ] 子站禁用 `wiki` 后，该子站不能发布 `wiki_page`。
- [ ] 板块不能绑定当前子站未启用的插件。
- [ ] 前台发布页只展示当前子站可发布的内容类型。
- [ ] 版主插件菜单只返回全局 enabled、子站 enabled 且当前用户有权限的插件菜单。
- [ ] 禁用插件后，已有 `/topics/:id` 详情页仍可访问并保留 SEO HTML。
- [ ] `/sitemap.xml` 和 `/robots.txt` 正常返回。
- [ ] `Service.CreatePost` 不能绕过插件发布校验。
- [ ] `POST/PUT/DELETE /api/v1/posts*` 写接口返回 `410 Gone` 或明确废弃。
- [ ] `POST /api/v1/admin/posts` 创建 `question/document/wiki_page` 时分别需要 `qa.question.create`、`docs.document.create`、`wiki.page.create`。
- [ ] 后台编辑内容不能修改子站、板块、`content_type` 或 `plugin_code`。

## 最近任务记录

### 2026-05-10：完整插件系统优先级与文档口径校准

修改范围：

- 更新 `docs/AGENT_RULES.md`，新增“任务结果记录规则”。
- 更新 `docs/PLUGIN_ARCHITECTURE.md` 和本文件，将“完整插件系统”登记为 P0 最高优先级长期主线，并补充 P0/P1/P2/P3 阶段路线。
- 更新 `README.md`、`CHANGELOG.md`、`docs/API.md`、`docs/TESTING.md`、`docs/releases/v1.3.0.md`、`docs/releases/v1.3.1.md` 的插件平台口径。

已完成事项：

- 将“当前阶段不做插件市场 / 插件包 / 远程安装 / 在线更新 / 动态加载”的口径改为“P2/P3 阶段能力，当前未实现”。
- 将 HookBus 业务处理器、config_schema 基础校验和插件平台测试矩阵标为 P0 收口任务。
- 将 config_schema 自动表单、SDK、模板、依赖和版本检查、搜索 / 通知 / SEO 扩展标为 P1。
- 将插件包、安装、升级、soft uninstall、migration runner、签名校验、市场雏形标为 P2。
- 将远程市场、在线更新、动态加载能力评估、沙箱和权限隔离标为 P3。
- 校准 `projects/jobs/ai_works` 状态：已接入插件平台治理和声明，不是 Core 兼容类型，也不是完整业务插件。

未完成事项：

- 本轮不实现插件市场、插件包、远程安装、动态加载或新增插件。
- 本轮不补 QA / Docs / Wiki / Projects / Jobs / AI Works 的专属业务功能。
- P0 中 `config_schema` 基础校验、HookBus 业务处理器、完整真实 token 验收矩阵仍待代码专项。

新发现风险：

- 旧文档中的“当前阶段不做”容易被后续任务误读为永久排除；已在当前主文档中改为分阶段路线，历史 Release Notes 保持归档属性。

已执行检查：

- `rg` 搜索限制性表述和插件相关口径，确认当前主文档已改为阶段路线。

跳过项及原因：

- 未执行 `go test ./...` / `go build`：本轮只调整文档口径和 Roadmap，没有修改 Go 代码结构。
- 未执行前后台构建：本轮没有修改前后台源码。

影响范围：

- API：仅文档口径更新，未新增真实接口。
- 数据库：无结构变更。
- 权限：无代码变更，文档明确 ActorContext 和插件写操作安全红线。
- SEO：无代码变更，保留 `/topics/:id` 动态 SEO 红线。
- 插件系统：提升为最高优先级长期主线，并补充 P0-P3 路线。
- 前后台 UI：无代码变更。

下一轮建议：

1. P0：实现 `config_schema` 基础校验。
2. P0：补 HookBus 业务处理器、统一错误日志和失败策略。
3. P0：执行完整插件系统真实 token 验收矩阵。

### 2026-05-10：后台插件管理界面体验增强

修改范围：

- 更新 `web/admin-app/src/views/Plugins.vue`。
- 更新 `web/admin-app/src/views/Communities.vue`。
- 更新 `docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/releases/v1.3.0.md`、`docs/releases/v1.3.1.md` 和 `CHANGELOG.md`。

已完成事项：

- 全局插件页新增插件系统说明卡片、状态 badge、内容类型 tag、权限 / 菜单 / schema 摘要和 loading / empty 状态。
- 全局插件详情从简单弹窗升级为抽屉，并按“基础信息 / 内容类型 / 权限 / 菜单 / 配置 / 路由 / Hooks”分区展示。
- 全局插件配置弹窗展示 `config_schema` 参考，支持 JSON 格式化、复制 schema，并继续在保存前校验 JSON 合法性。
- 全局启用 / 禁用增加确认文案，明确禁用只影响新发布、导航、菜单和管理入口，不影响历史内容和 SEO。
- 子站插件配置抽屉新增启用统计、全局 / 子站双状态 badge、全局禁用原因、子站未启用说明、schema 参考、JSON 格式化和更清晰的排序操作。
- 修正子站插件上移 / 下移后未同步 `sort_order` 导致保存排序可能按旧数字排序的问题。
- 版主插件菜单本轮未改代码；现有 `/moderator` 工作台已通过 `GET /api/v1/moderator/plugin-menus` 展示过滤后的插件治理入口。

未完成事项：

- 本轮未新增插件影响范围统计接口，因此 UI 不展示绑定子站数量、启用子站列表或受影响板块数量，避免伪造数据。
- 本轮未引入自动浏览器测试；后台插件详情抽屉、子站配置抽屉、禁用确认、配置保存和排序仍需真实浏览器矩阵验收。
- 本轮未做 `config_schema` 强校验或自动表单渲染，仍只校验 JSON 合法性。

新发现风险：

- 后台 UI 依赖现有插件接口字段；如果未来后端将 `config_json` 从字符串改为对象，页面已做基础兼容，但仍建议 API 文档保持字段类型稳定。

已执行检查：

- `gofmt -w internal/domain/models.go internal/service/service.go internal/store/mysql.go internal/store/schema.go internal/transport/httpapi/router.go internal/transport/httpapi/router_auth_test.go`：通过；这些 Go 文件为工作区既有改动，格式化后无报错。
- `go test ./...`：通过。
- `go build`：通过；构建产生的根目录临时二进制已清理。
- `cd web/admin-app && npm run build`：宿主机缺少 `npm`，失败于 `npm: command not found`。
- `docker run --rm -v "$PWD/web/admin-app":/app -w /app node:20-alpine sh -c "npm run build"`：通过；Vite 输出 chunk size warning，但构建成功。

跳过项及原因：

- 本轮未修改前台代码，跳过 `cd web/frontend-app && npm run build`。
- 未执行真实浏览器矩阵：仓库当前没有自动浏览器测试 runner，需要后续手工验收或引入专项测试工具。

影响范围：

- API：未新增接口，继续使用现有全局插件、子站插件和版主菜单 API。
- 数据库：无结构变更。
- 权限：无后端权限逻辑变更，UI 继续依赖后端权限和菜单过滤。
- SEO：无 SEO 行为变更，保留 disabled 插件不影响历史内容 SEO 的提示。
- 插件系统：后台插件治理体验增强。
- 前后台 UI：只修改后台全局插件管理和子站插件配置 UI。

下一轮建议：

1. 用真实 admin token 完成 `/admin-next/plugins` 和 `/admin-next/communities` 插件配置浏览器矩阵。
2. 增加插件影响范围统计接口后，再在禁用确认中展示启用子站、绑定板块和可能受影响入口。
3. 继续实现 P0 `config_schema` 基础校验。
