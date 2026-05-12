# DevHub 项目进度

[返回文档入口](README.md)

更新时间：2026-05-12

本文档只记录当前仓库真实状态、当前风险和下一步任务。历史版本能力已并入当前分支，详情见对应 Release Notes；旧版本已解决问题不再占用当前主体。

## 当前版本结论

当前版本为 `v1.3.4`，主题是“插件异常治理与平台基础能力收口版”。DevHub 当前定位为多子站通用开源社区程序，默认演示为开发者社区。

Core 保留用户、认证、子站、板块、通用内容、评论、标签、搜索、通知、SEO、权限、审计、插件注册和分发能力。问答、文档、Wiki、项目、招聘、AI 作品已按内置系统插件建模：`qa -> question`、`docs -> document`、`wiki -> wiki_page`、`projects -> project`、`jobs -> job`、`ai_works -> ai_work`。

当前实现仍保留历史表名以保证兼容：`topics` 是当前通用内容表，`categories` 是当前通用板块表。

当前最高优先级长期主线是完成完整插件系统。DevHub 的长期目标不是只支持内置 `qa/docs/wiki`，而是形成完整插件平台：Core 只提供通用社区底座，业务能力通过插件声明、插件状态、插件权限、插件菜单、插件配置、插件 Hook、插件 migration、插件 API、插件 SEO、插件通知、插件搜索和插件测试矩阵扩展。

下一阶段最高优先级目标已正式收口为 `v1.3.5：插件治理体验与安装升级向导收口版`，详见 [v1.3.5 Draft](releases/v1.3.5.md) 和 [完整插件系统长期完善路线图](PLUGIN_SYSTEM_ROADMAP.md)。该阶段优先把现有插件平台能力做成清晰、可确认、可验收的后台治理体验，不代表插件市场、远程安装或动态加载已经完成。

## 当前已完成

- 插件注册：`internal/plugins/registry.go` 和 `internal/plugins/qa|docs|wiki|projects|jobs|aiworks` 提供内置插件定义、内容类型映射、菜单、权限和路由描述。
- 插件声明规范：当前已统一到 manifest 风格声明，包含插件本体、内容类型定义、权限定义、菜单定义、路由定义、`config_schema`、依赖、最小 Core 版本和 Hook 声明。
- 全局插件状态：`plugins` 表和 MemoryStore / MySQLStore 已扩展支持 `discovered`、`installed`、`migrated`、`configured`、`enabled`、`disabled`、`running`、`config_invalid`、`migration_pending`、`dependency_missing`；当前发布可用性仍只认 `enabled`，其余状态不放行新建内容。
- 插件启用 readiness：全局启用和子站启用会在 Service 层检查插件存在、配置有效、依赖已启用、没有 `failed` 迁移；当前内置 no-op 的 `pending` migration 不阻断启用，但会在健康状态和迁移 Tab 中提示。
- 子站插件状态：`community_plugins` 表和 MemoryStore / MySQLStore 均支持按子站启用 / 禁用、配置和排序插件。
- 两层状态判断：插件在某个子站可用需要同时满足 `plugins.status=enabled` 和 `community_plugins.status=enabled`；`core` 作为兼容内置能力在 Service 层特殊视为可用。
- 内容模型兼容：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 已进入 schema 与 Store。
- 发布校验：`POST /api/v1/topics` 已走统一 `ValidateTopicPluginAccess`，会归一 `doc -> document`、`wiki -> wiki_page`，并校验插件存在、全局 `enabled`、子站 `enabled`、板块插件绑定、允许内容类型和服务端权限码；前端隐藏入口不能替代后端强拦截。
- 板块管理校验：MemoryStore / MySQLStore 在创建或编辑子站板块时校验 `plugin_code` 与 `content_type` 匹配，并拒绝绑定全局或子站未启用的插件。
- 插件 API：全局插件 API、子站插件 API、前台子站插件展示 API 和版主插件菜单 API 已在 `router.go` 注册。
- 插件安装 / 升级预备：已支持 manifest 校验、manifest dry-run、manifest + 配置型插件安装记录、健康总览、批量归档 / 恢复、升级 dry-run 和最小升级执行闭环；这些能力不执行第三方代码、不加载外部前端资源、不执行外部 raw SQL。
- 插件软卸载：已支持全局归档 / 恢复，归档插件会阻断新建内容和子站启用，但保留历史内容、配置、迁移记录、审计记录和 SEO；恢复后默认进入 `disabled`。
- 插件业务闭环：
  - `qa`：发布 `question` 时写入 `qa_questions`；回答写入 `qa_answers`；采纳后回写已解决状态和最佳答案。
  - `docs`：发布 `document` 时写入 `docs_documents`，并支持基础文档树读取。
  - `wiki`：发布 `wiki_page` 时写入 `wiki_pages` 和初始 `wiki_page_versions`；编辑时新增版本记录。
- `project` / `job` / `ai_work` 已完成插件归属迁移：`projects -> project`、`jobs -> job`、`ai_works -> ai_work`，发布校验、权限码、菜单声明和历史 `plugin_code` 迁移口径已接入；专属扩展表和完整业务闭环尚未完成。
- 前台 seed 用户：MemoryStore 和 MySQLStore 初始化时都会补齐 `liuwei / 方圆十三 / a123456` 的前台用户，方便手工登录和前台用户 E2E。
- 权限上下文：`CreateTopicRequest.ActorPermissions` / `ActorContext` 均由服务端从 token、后台身份和版主 scope 计算，客户端请求体不能覆盖。
- 配置合并：`plugins.config_json` 与 `community_plugins.config_json` 已落库并可写，返回 `resolved_config.default/global/community/effective` 合并视图。
- HookBus：Service 层已有最小内部 HookBus，当前调用点覆盖内容创建、更新、删除、评论、搜索、通知和 SEO 事件；Search / Notification / SEO 当前是最小事件派发，不做第三方动态执行。
- `v1.3.2`：HookBus 已作为插件平台能力收口到 `internal/plugins`，内置插件可注册最小 Hook handlers；同时 `config_schema` 已在全局/子站插件配置保存时做基础校验（简化 JSON Schema）。
- `v1.3.2`：新增 `plugin_migrations` 表，用于记录插件迁移执行状态（平台治理能力的一部分）。
- 前台入口：子站插件公开接口会隐藏 `config_json` / `resolved_config` 等后台配置；子站板块导航会按子站插件状态过滤。
- 后台入口：`/admin-next/plugins` 作为系统插件管理入口；插件业务页通过系统插件列表进入，默认不散落在后台左侧导航。
- 后台插件管理体验：
  - 后台全局插件管理已支持说明卡片、插件状态 badge、内容类型 tag、权限 / 菜单 / schema 摘要、详情抽屉、tabs 分区展示、配置 schema / resolved config JSON 展示与复制、全局配置编辑、Ajv 客户端校验、启用 / 禁用确认和 impact 计数提示。
  - 后台子站插件配置已支持双状态 badge、子站启用统计、全局禁用原因提示、启用 / 禁用确认、`config_json` 编辑、schema 参考、JSON 格式化、Ajv 客户端校验、数字排序和上移 / 下移后保存。
  - 前台子站页和发布页会按当前子站已启用插件收口入口与内容类型。
  - 版主工作台已补最小插件治理入口区，并按当前子站插件状态与权限过滤。
- 审计：全局插件状态、子站插件状态、全局插件配置、子站插件配置和排序已接入 `admin_logs`，并为插件治理操作写入 `old_value`、`new_value`、`metadata_json` 结构化字段；`target` 文本摘要继续保留用于兼容展示。
- Wiki schema：当前只保留插件化 `wiki_spaces`、`wiki_pages`、`wiki_page_versions` 语义，旧 `wiki_revisions` 预留冲突已清理。
- SEO 保护：`/topics/:id` 仍由 Go 动态输出 SEO HTML，插件禁用不影响历史内容详情访问。
- 插件影响分析：全局 / 子站 impact 接口已扩展返回历史内容数、启用/禁用子站数、绑定板块数、近 7 天内容数、审核中内容数、菜单声明数、配置覆盖数、待执行迁移数和近 7 天 Hook 失败数。
- 技术债收口：`Service.CreatePost` 已封口，不再作为业务写入口；`/api/v1/posts` 写接口继续废弃；后台 `admin/posts` 创建入口在兼容 `post.create` 基础权限之外，叠加真实内容类型对应的插件 create 权限。
- 后台编辑边界：后台内容编辑已禁止修改子站、板块、`content_type` 和 `plugin_code` 归属字段；如后续需要迁移归属，必须走单独迁移专项和完整插件校验。
- 插件系统专项验收：2026-05-11 已完成一次后端、后台构建、SEO curl、前台 E2E、后台 E2E 集中回归；前台 14 条 E2E 通过，后台 15 条 E2E 在修复状态污染后通过，详见本文末尾“插件系统专项验收与 E2E 回归清单归档”。2026-05-12 后续补充覆盖归档态前台入口、PluginContent 历史治理、健康总览、manifest 操作和升级最小闭环。
- MySQLStore / 老库升级专项：2026-05-11 已验证 `plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、`admin_logs` 新装结构和 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 升级字段；`004`-`010` 插件迁移 SQL 在测试库连续执行两轮通过；MySQLStore 下全局/子站插件启停强拦截、failed migration 阻断与 retry、Hook 记录、插件审计查询和 config_schema 校验均通过可选集成测试。

## 插件平台基线对账

本节是 2026-05-12 基于代码阅读和文档归档后的真实基线。

已完成能力：

- Registry / manifest：内置 `qa/docs/wiki/projects/jobs/ai_works` 统一声明 `plugin_code`、`content_types`、权限、菜单、路由、Hook、`config_schema`、依赖和最小 Core 版本。
- 运行状态：`plugins.status` 与 `community_plugins.status` 两层状态已落地；全局状态枚举已扩展，但只有 `plugins.status=enabled` 与 `community_plugins.status=enabled` 的组合会放行发布；`archived` 会像 disabled 一样阻断新建和子站启用。
- 配置：全局 `plugins.config_json`、子站 `community_plugins.config_json`、`resolved_config` 合并视图、JSON 合法性校验和简化 `config_schema` 后端校验已落地。
- 权限：发布链路按 `ContentTypeDefinition.create_permission` 校验；后台 / 版主插件菜单按状态和权限过滤。
- Hook：存在内置 HookBus，内容创建 / 更新 / 删除、评论、搜索、通知、SEO、插件启停均有最小调用点。
- 迁移记录：存在 `plugin_migrations` 表和 MemoryStore / MySQLStore 读写能力。
- 后台治理：`/admin-next/plugins`、插件详情抽屉、配置、impact、审计 Tab、通用插件内容页和子站插件配置抽屉均已具备基础能力；manifest validate / dry-run / install、upgrade dry-run / upgrade、health summary、bulk archive / restore 已有入口和结构化结果展示。

部分完成能力：

- 生命周期：`install_status`、`lifecycle_status`、`status_reason`、`installed_at`、`archived_at`、`last_health_check_at` 已作为后台展示字段返回；但它们仍是派生展示，不是完整外部插件安装器状态机，当前代码以 `plugins.status`、`community_plugins.status`、配置校验、依赖和 `plugin_migrations.status` 为判断依据。
- Hook 治理：Hook 能执行，已有 `hook_executions`、失败统计、最近错误、平均耗时、失败率和 `plugin.hook.failed` / `plugin.hook.blocked` 审计；重试策略、告警和更多业务处理器仍待后续。
- 插件迁移：已有内置插件 up/no-op runner、失败记录、失败重试、后台迁移 Tab 和迁移审计；manifest + 配置型插件安装会生成 pending migration 记录，但不执行外部 raw SQL。migration down、真实 rollback、迁移前备份和外部插件迁移包仍未完成。
- 插件安装 / 升级：manifest 校验、dry-run、manifest + 配置型安装记录、upgrade dry-run 和最小升级执行已经落地；完整安装 / 升级向导、回滚、版本迁移向导、插件包 zip 上传和签名仍待后续。
- 权限矩阵：当前是最小权限码校验，不是完整 RBAC 矩阵；community/category 作用域、角色分配 UI 和权限配置 API 仍待后续。
- 插件内容治理：通用页、基础详情抽屉、批量隐藏 / 恢复和审计跳转已接入；批量审核、置顶、加精、专属详情和完整权限矩阵仍待后续。

预留能力：

- 插件包 zip 上传、远程安装、外部服务型 Webhook 真实调用、动态加载、脚本沙箱、hard uninstall 和 migration down。
- 插件健康状态已有轻量摘要和 API；运行监控、告警、自动恢复、插件依赖 UI 和独立版本兼容矩阵页面仍待后续。
- 插件 SDK 文档已建立；生成模板、插件市场、远程安装、动态加载、沙箱和第三方 Hook / Webhook 运行时仍待后续。

后续规划：

- `v1.3.5 / P0` 优先把插件治理中心、安装向导、升级向导、批量归档 / 恢复影响预览、状态治理页和 PluginContent 体验做成可验收闭环。
- `v1.4.x+` 再推进外部服务型 Webhook、插件包 zip、签名、市场、动态加载评估和高级运行时能力。

## 当前部分完成

- 子站插件管理 UI：后台体验已从最小表格增强为更清晰的配置面板，包括全局 / 子站双状态、禁用原因、schema 参考、JSON Editor、Ajv 校验、禁用影响提示和排序保存；多浏览器矩阵、批量操作和更强可视化仍待后续专项验收。
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
- Hook 机制：当前已有内置 HookBus，并覆盖创建、更新、删除、评论、搜索、通知和 SEO 调用点；执行结果已落入 `hook_executions`，失败会写入 `plugin.hook.blocked` / `plugin.hook.failed` 审计。Search / Notification / SEO 仍是预留级事件派发，尚未形成完整插件业务处理器、健康状态、告警和重试策略。
- 配置校验：当前已完成默认配置、全局配置、子站配置三层合并，后端已按简化 `config_schema` 做基础校验，后台已支持基础自动表单 + JSON 高级模式并接入 Ajv 客户端校验；完整 JSON Schema、深层嵌套、字段分组、配置版本和回滚仍待后续。
- 验收覆盖：已做文档与路由核对；完整 Docker 启动、真实 token API、浏览器页面和 SEO curl 矩阵仍需按测试文档继续补测。

## 当前未完成

- 插件治理 UI 信息架构：`/admin-next/plugins` 已可用，但完整分步安装向导、完整分步升级向导、批量归档 / 恢复影响预览表格、状态治理页异常入口和更清晰的操作分组仍待 `v1.3.5` 收口。
- 子站插件配置 UI 的完整浏览器验收矩阵，包括多子站、禁用提示、保存失败提示和排序持久化回归。
- 更细粒度的权限体系：例如 Core 兼容类型 `article` / `news` 的细分权限码、按子站/板块维度配置权限矩阵、以及更明确的错误码与权限配置 API（当前仍为最小校验闭环）。
- P0 插件平台收口：HookBus 的完整业务处理器、健康状态、告警和重试策略。Search / Notification / SEO 目前已有调用点和执行记录，但缺少实际插件处理器。
- P1 插件平台增强：`config_schema` 自动表单增强、更完整 JSON Schema、字段分组、配置版本和回滚。
- 非插件历史审计日志的结构化 diff：插件治理已写入 `old_value`、`new_value`、`metadata_json`，其他旧审计仍可能只有 `target` 文本。
- `qa` 取消采纳最佳答案。
- Docs 文档树专用编辑 UI。
- Docs 文档树拖拽、批量排序和更完整的空间管理体验。
- Wiki 版本回滚接口与协作编辑交互。
- Projects / Jobs / AI Works 的专属扩展表、专属管理页、专属搜索、通知、SEO 和完整业务闭环。
- P2 插件分发能力：本地插件包 zip、插件包签名校验、外部服务型 Webhook、插件包安装向导、升级影响分析深化、插件市场雏形。
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

- schema 自动表单增强。
- 插件 SDK 文档。
- 插件生成模板。
- 插件依赖检查。
- 插件版本兼容检查。
- 插件事件和通知模板。
- 插件搜索索引扩展。
- 插件 SEO 扩展。

P2：插件分发能力

- 本地插件包。
- 插件包安装。
- 插件升级向导增强。
- 插件禁用与 soft uninstall 增强。
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
- API 和后台 UI 已增强不等于完整产品闭环；当前已新增 Docker 化 Playwright 最小 E2E runner，但多账号、多子站、跨权限和更细视觉交互仍需继续扩展浏览器矩阵。
- `/sitemap.xml` 当前仍是单文件动态输出，内容规模扩大后需要 sitemap index / 分片。
- 用户提出的 `docs/BACKUP_ROLLBACK.md` 与仓库真实文件名不一致；当前真实文件是 `docs/BACKUP_AND_ROLLBACK.md`。

## 下一步任务

1. `v1.3.5 / P0`：重构 `/admin-next/plugins` 的信息层级，形成“总览 / 筛选与批量操作 / 插件列表 / 详情治理层”的清晰结构。
2. `v1.3.5 / P0`：把 manifest 校验、dry-run、install 做成完整分步安装向导。
3. `v1.3.5 / P0`：把 upgrade dry-run、兼容矩阵、upgrade 执行做成完整分步升级向导。
4. `v1.3.5 / P0`：补批量归档 / 恢复的影响预览、`succeeded` / `failed` 表格和审计跳转。
5. `v1.3.5 / P0`：把“状态治理”页明确为异常处理入口，聚合迁移待处理、Hook 异常、配置无效、依赖缺失和归档插件。
6. `v1.3.5 / P0`：对齐 PluginContent 内容治理页体验，确保归档 / 禁用插件历史内容仍可查看和治理。
7. `v1.3.5 / P0`：补一轮最小后台 E2E，只覆盖上述新向导和关键入口，不扩成全量矩阵。
8. `v1.4.x+`：推进外部服务型 Webhook、插件包签名、真实插件包 dry-run、生产 MySQL 大库演练和插件市场雏形。

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

### 2026-05-11：插件配置管理与 config_schema 后端强校验

修改范围：

- `internal/plugins/registry.go`
- `internal/plugins/qa/qa.go`
- `internal/plugins/docs/docs.go`
- `internal/plugins/wiki/wiki.go`
- `internal/plugins/registry_test.go`
- `internal/transport/httpapi/router.go`
- `internal/transport/httpapi/router_auth_test.go`
- `docs/API.md`
- `docs/PLUGIN_ARCHITECTURE.md`
- `docs/PROJECT_PROGRESS.md`
- `docs/TESTING.md`
- `docs/releases/v1.3.2.md`
- `CHANGELOG.md`

已完成事项：

- `config_schema.properties.*.default` 已参与 `resolved_config.effective` 合并，作为最低优先级默认配置来源。
- 明确配置层级：schema 默认值 -> `plugins.config_json` 全局配置 -> `community_plugins.config_json` 子站配置，子站配置优先级最高。
- 后端强校验继续由 `pluginregistry.ValidateConfigJSON` 执行，覆盖 JSON 合法性、required、type、enum、object、boolean、string、number、integer 和数字 min/max。
- `qa/docs/wiki` 内置插件补充关键配置默认值，用于 effective config 合并。
- 全局插件配置和子站插件配置审计 metadata 增加 `changed_keys`，记录本次变更的顶层配置键。
- 新增/扩展测试覆盖默认值合并、required 缺失、enum 非法、type 错误、integer/min 边界与配置审计 metadata。

未完成事项：

- 当前 diff 只记录顶层 `changed_keys`，不做深层路径级 diff。
- 当前不做自动表单生成器、配置版本回滚、灰度配置或敏感字段加密。
- 配置校验失败暂未写入审计日志，避免把大量无效输入刷入治理日志；后续如需要可单独做安全审计策略。

新发现风险：

- schema 默认值目前来自 `properties.*.default`，不是完整 JSON Schema default 递归实现；嵌套对象默认值和数组默认值仍需后续增强。

已执行检查：

- `gofmt -w internal/plugins/registry.go internal/plugins/qa/qa.go internal/plugins/docs/docs.go internal/plugins/wiki/wiki.go internal/plugins/registry_test.go internal/transport/httpapi/router.go internal/transport/httpapi/router_auth_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。

失败项或跳过项及原因：

- 未执行完整前后台 E2E 矩阵；本轮按要求只做配置管理链路的轻量验收。

影响范围：

- API：配置 API 返回的 `resolved_config.effective` 会包含 schema 默认值；审计 `metadata_json` 增加 `changed_keys`。
- 数据库：无结构变更。
- 权限：无权限模型变更。
- SEO：无 SEO 逻辑变更。
- 插件系统：增强配置合并与配置审计闭环。
- 后台 UI：无需改页面结构，继续复用 JSON Editor + Ajv；后端强校验兜底。

下一轮建议：

- 继续做配置中心增强：深层 diff、配置版本记录、配置回滚或 schema 自动表单复杂字段矩阵。

### 2026-05-11：插件状态模型、启停强拦截与影响分析

修改范围：

- `internal/plugins/registry.go`
- `internal/domain/models.go`
- `internal/service/service.go`
- `internal/store/memory.go`
- `internal/store/mysql.go`
- `internal/store/schema.go`
- `db/mysql/001_schema.sql`
- `db/mysql/migrations/005_core_plugins.sql`
- `db/mysql/migrations/009_plugin_status_model.sql`
- `web/admin-app/src/views/Plugins.vue`
- `web/admin-app/src/views/Communities.vue`
- `docs/API.md`
- `docs/PLUGIN_ARCHITECTURE.md`
- `docs/PROJECT_PROGRESS.md`
- `docs/TESTING.md`
- `docs/releases/v1.3.2.md`
- `CHANGELOG.md`

已完成事项：

- 扩展全局插件状态模型：`plugins.status` 与 MemoryStore / MySQLStore 均接受 `discovered`、`installed`、`migrated`、`configured`、`enabled`、`disabled`、`running`、`config_invalid`、`migration_pending`、`dependency_missing`。
- 明确发布可用性规则：当前只有全局 `enabled` + 子站 `enabled` 才能创建插件内容；其他全局状态均不会放行新建内容。
- 强化 `ValidateTopicPluginAccess` 错误信息：创建内容时会校验内容类型、插件存在、全局状态、子站状态、板块绑定和 `allowed_content_types`。
- 扩展全局与子站 impact 统计：返回历史内容数、启用/禁用子站数、绑定板块数、近 7 天内容数、审核中内容数、菜单声明数、配置覆盖数和待执行迁移数。
- 后台全局禁用与子站禁用确认弹窗展示更完整的 impact 信息，并明确历史内容和 SEO 不受影响。
- 新增老库迁移 `009_plugin_status_model.sql`，用于扩展 MySQL `plugins.status` enum。

未完成事项：

- 完整生命周期状态机尚未完成；`running`、`configured`、`migration_pending` 等当前是可记录状态，不是自动流转机制。
- Hook 错误统计已在后续“HookBus 运行时与 Hook 执行记录”任务中接入 `hook_executions`；`recent_hook_errors_count` 现在是近 7 天失败次数，但仍不等同于完整健康评分。
- impact 仍是轻量计数，不包含受影响对象明细列表。

新发现风险：

- 如果运维手动把插件状态改成 `running` 或 `configured`，当前发布链路不会放行；这符合本轮“只有 enabled 可发布”的安全策略，但需要在后续生命周期状态机中明确状态流转语义。

已执行检查：

- `gofmt -w internal/domain/models.go internal/plugins/registry.go internal/service/service.go internal/store/memory.go internal/store/mysql.go`：通过。
- `go test ./...`：首次因错误文案不包含旧测试期望的“插件未启用”失败；已调整为兼容且更明确的错误文案后重跑通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。

失败项或跳过项及原因：

- 未执行完整前后台 E2E 矩阵；本轮按要求只做状态模型、强拦截、影响分析与后台构建的轻量验收。

影响范围：

- API：impact API 返回字段扩展；插件状态口径扩展。
- 数据库：扩展 `plugins.status` enum，并新增老库升级迁移。
- 权限：无权限模型结构变更；发布链路继续由服务端权限码校验。
- SEO：无 SEO 逻辑变更；disabled 插件不影响历史 `/topics/:id` 动态 HTML。
- 插件系统：增强状态模型、启停强拦截和禁用前影响分析。
- 后台 UI：增强插件禁用确认弹窗的真实影响提示。

下一轮建议：

- 继续为 impact 增加按需加载的受影响对象明细，并为 HookBus 补健康状态、告警和重试策略。

### 2026-05-11：插件平台基线对账与文档口径统一

修改范围：

- `docs/PROJECT_PROGRESS.md`
- `docs/PLUGIN_ARCHITECTURE.md`
- `docs/API.md`
- `docs/TESTING.md`
- `docs/releases/v1.3.2.md`
- `CHANGELOG.md`

已完成事项：

- 对账 `internal/domain/models.go`、`internal/plugins/*`、`internal/service/service.go`、`internal/store/*`、`internal/transport/httpapi/router.go` 和后台插件页面，确认当前插件平台真实基线。
- 明确当前定位是“内置系统插件平台 + registry / manifest 描述 + Core 分发”，不是第三方插件市场、动态插件加载或插件包安装系统。
- 明确真实表名仍为 `topics` / `categories`，`contents` / `channels` 只作为架构概念或长期目标命名。
- 将插件能力按“已完成 / 部分完成 / 预留 / 后续规划”重新归类，重点校正生命周期、Hook、迁移、权限矩阵、配置和后台治理口径。
- 修正 `config_schema` 相关旧表述：当前已完成 JSON 合法性校验和简化 schema 后端校验；自动表单和完整 JSON Schema 仍是后续。
- 修正当时 `plugin_migrations` 口径：该阶段只有记录表和 Store 读写，不等于完整 migration runner；后续已补齐内置 up/no-op runner、执行/重试 API 和后台迁移 Tab。

未完成事项：

- 本轮不实现新功能；当时 Hook 执行记录、迁移 runner、插件健康状态、完整权限矩阵、插件安装/升级/卸载仍待后续代码任务。后续任务已补齐 Hook 执行记录、轻量健康摘要和内置 up/no-op 迁移 runner；完整权限矩阵、真实 rollback、插件安装/升级/卸载仍待后续。

新发现风险：

- 文档中容易把 `plugin_migrations` 表存在误读为迁移系统完成；当时已修正为“记录能力已完成，runner/重试/rollback 待实现”，后续已补齐第一阶段 runner 与重试，rollback 仍待实现。
- 文档中容易把 Hook 调用点存在误读为 Hook 观测完成；已修正为“可派发，执行记录和统计待实现”。

已执行检查：

- `rg` 核对插件 registry、manifest、HookBus、config_schema、plugin_migrations、impact、admin_logs 和后台插件页面实现：通过。
- `test -f docs/PLUGIN_SYSTEM_ROADMAP.md`：通过。
- `rg "插件平台基线对账|discovered|plugin_migrations|config_schema" docs/PROJECT_PROGRESS.md docs/PLUGIN_ARCHITECTURE.md docs/API.md docs/TESTING.md docs/releases/v1.3.2.md CHANGELOG.md`：通过。

失败项或跳过项及原因：

- 未执行 Go / 前后台构建；本轮仅进行代码阅读和 Markdown 文档口径修正，无业务代码、API、数据库或 UI 改动。

影响范围：

- API：无新增或修改，API 文档只补充基线说明。
- 数据库：无结构变更。
- 权限：无代码变更，文档补充当前权限矩阵边界。
- SEO：无代码变更，继续保留插件 disabled 不影响历史内容 SEO 的红线。
- 插件系统：无运行时变更，文档统一真实状态。
- 前后台 UI：无代码变更。

下一轮建议：

- 从 P0 中挑一个可落地专项推进：优先建议实现 Hook 执行记录 / 失败审计，或实现 plugin_migrations 查询与手动执行入口。

### 2026-05-11：登记完整插件系统长期完善路线图为下一阶段最高优先级目标

修改范围：

- 新增 `docs/PLUGIN_SYSTEM_ROADMAP.md`
- 更新 `docs/README.md`
- 更新 `docs/PROJECT_PROGRESS.md`
- 更新 `docs/PLUGIN_ARCHITECTURE.md`
- 更新 `docs/AGENT_RULES.md`

已完成事项：

- 将用户提供的“DevHub 插件系统完善需求文档”整理为项目内正式路线图，明确它是下一阶段最高优先级目标文档，而不是当前已完成能力声明。
- 在文档入口中加入“完整插件系统路线图”，并把当前文档口径从 `v1.3.1` 校准到 `v1.3.2`。
- 在项目进度中登记完整插件平台流程为下一阶段最高优先级，覆盖插件生命周期、启停、配置、权限、菜单、Hook、内容治理、审计、监控、迁移、升级、卸载和 E2E。
- 在插件架构说明和 Agent 规则中补充对路线图的引用，确保后续任务以该文档作为插件系统长期目标来源。

未完成事项：

- 本轮只做文档登记和目标校准，未实现新的插件安装器、迁移执行器、Hook 观测、健康状态、插件包或插件市场能力。
- 后续仍需按路线图 P0/P1/P2/P3/P4 分阶段拆解为代码任务。

新发现风险：

- 该路线图是目标文档，不能在 API、Release Notes 或进度文档中被误读为当前全部实现完成。

已执行检查：

- `test -f docs/PLUGIN_SYSTEM_ROADMAP.md`：通过。
- `rg "PLUGIN_SYSTEM_ROADMAP|完整插件系统长期完善路线图" docs/README.md docs/PROJECT_PROGRESS.md docs/PLUGIN_ARCHITECTURE.md docs/AGENT_RULES.md`：通过。

失败项或跳过项及原因：

- 未执行 Go / 前后台构建；本轮仅修改 Markdown 文档，不影响代码、API、数据库、权限、SEO 或前后台 UI。

影响范围：

- API：无变更。
- 数据库：无变更。
- 权限：无代码变更。
- SEO：无代码变更，SEO 红线继续保留。
- 插件系统：新增长期目标登记，不改变运行时行为。
- 前后台 UI：无变更。

下一轮建议：

- 从路线图 P0 开始拆分任务，优先做插件影响范围、Hook 执行记录、配置校验后端闭环、审计结构化查询和 `content_type` 强校验 E2E。

### 2026-05-11：扩展 DevHub 第二阶段 E2E，覆盖前台真实业务流、插件联动、版主权限边界与后台细操作矩阵

修改范围：

- `web/frontend-app/tests/e2e/helpers/auth.js`
- `web/frontend-app/tests/e2e/helpers/api.js`
- `web/frontend-app/tests/e2e/interactions.spec.js`
- `web/frontend-app/tests/e2e/publish.spec.js`
- `web/frontend-app/tests/e2e/plugin-visibility.spec.js`
- `web/frontend-app/tests/e2e/moderator.spec.js`
- `web/admin-app/tests/e2e/helpers/api.js`
- `web/admin-app/tests/e2e/helpers/selectors.js`
- `web/admin-app/tests/e2e/reports.spec.js`
- `web/admin-app/tests/e2e/moderators.spec.js`
- `web/admin-app/tests/e2e/plugin-content.spec.js`
- `web/admin-app/src/views/Reports.vue`
- `web/admin-app/src/views/Moderators.vue`
- `web/admin-app/src/views/PluginContent.vue`
- `docs/TESTING.md`
- `docs/PROJECT_PROGRESS.md`

已完成事项：

- 前台 E2E 从 6 条扩展到 14 条，新增登录互动、点赞、收藏、关注主题、评论、用户中心访问、登录发布成功、必填校验、插件禁用联动、强传禁用 / 非法 `content_type` 拦截、版主工作台权限边界和跨子站 API 越权拦截。
- 后台 E2E 从 11 条扩展到 15 条，新增 reports 细操作、moderators 管理边界、qa/docs/wiki 通用插件内容页和插件内容治理代表链路。
- 新增前台 E2E helper，集中维护普通用户、PHP 版主用户、API 请求、插件状态、板块状态、唯一标题和 SEO 基础断言。
- 新增后台 E2E helper，集中维护 admin/user API 请求、测试 Topic、测试举报、页面 ready 断言和表格行查找。
- 为 reports、moderators、PluginContent 补充少量稳定 `data-testid`，不改变现有 UI 结构或视觉。
- 插件启停测试会在用例内恢复 `qa` 全局状态与 PHP QA 板块状态，避免污染后续 E2E。

新发现风险：

- 当前本地 PHP 子站只有 `community` 板块默认启用，其余 QA / Docs / Wiki / Projects / Jobs / AI Works 板块处于停用状态；涉及插件发布入口的 E2E 需要在测试内显式启用并恢复对应板块，不能假设所有 seed 板块都开启。
- 后台 E2E 使用 Playwright 多 worker 并发运行；全局插件启停类测试可能与依赖对应插件状态的创建测试互相影响。当前已避免新增用例依赖并发中的 `qa` 创建，后续新增全局状态测试仍应 serial 或显式隔离。

已执行检查：

- `docker compose run --rm admin-e2e npm run build`：通过，Vite 仅输出 chunk size warning。
- `docker compose run --rm frontend-e2e npm run build`：通过。
- `docker compose run --rm frontend-e2e`：通过，14 个 Playwright 用例全部通过。
- `docker compose run --rm admin-e2e`：通过，15 个 Playwright 用例全部通过。
- `bash -n dev.sh`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `./scripts/check-frontend.sh --quick --target both`：通过，日志目录 `.devhub/checks/20260511-025032/`。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build + 15 个 E2E 通过，日志目录 `.devhub/checks/20260511-025053/`。
- `./scripts/check-frontend.sh --frontend-only`：通过，前台 build + 14 个 E2E 通过，日志目录 `.devhub/checks/20260511-025053/`。

失败项或跳过项及原因：

- 无最终失败项。
- 本轮未新增插件、未修改 Docker runner / Dockerfile / compose / npm 依赖结构。
- 后台版主管理“新增 / 编辑 / 停用”当前采用真实 API 操作 + UI 列表反映的组合覆盖；纯 Element Plus 表单 UI 操作仍作为后续更细 E2E。

影响范围：

- API：无新增或修改。
- 数据库：无结构变更。
- 权限：未改业务权限逻辑；新增 E2E 覆盖普通用户、版主、后台管理员边界。
- SEO：未改 SEO 逻辑；新增 E2E 验证禁用 `qa` 后历史 `/topics/:id` SEO 仍可访问。
- 插件系统：未改业务逻辑；新增插件启停、发布入口联动和强传拦截 E2E。
- 前后台 UI：仅增加少量稳定 `data-testid`，无视觉或交互重构。

未完成事项：

- 跨子站插件矩阵仍需继续扩展：例如 PHP 禁用 QA、Go 仍启用 QA 时的双子站并行发布验证。
- 版主管理纯 UI 表单新增 / 编辑 / 停用仍可继续补更细浏览器用例。
- PluginContent 专属详情、审核按钮 UI 和更多插件业务状态操作仍待后续扩展。
- MySQLStore 与老库迁移场景仍需单独跑同等 E2E / API 矩阵。

下一轮建议：

1. 将插件状态类 E2E 统一标记 serial 或按项目拆分，降低未来并发状态污染风险。
2. 继续补前台多子站插件导航显示 / 隐藏矩阵，以及版主多账号多子站菜单过滤矩阵。
3. 为后台 moderators 和 plugin content 补纯 UI 操作闭环，减少 API 辅助操作占比。

### 2026-05-11：修复后台插件列表与插件详情抽屉显示问题

修改范围：

- `web/admin-app/src/views/Plugins.vue`
- `web/admin-app/src/components/plugin/PluginDetailDrawer.vue`
- `docs/PROJECT_PROGRESS.md`

已完成事项：

- 修复插件详情抽屉点击“详情”后内容区空白的问题：默认打开的 tab 从不存在的 `basic` 改为真实存在的 `overview`。
- 优化 `/admin-next/plugins` 插件列表页显示密度，压缩统计卡片、筛选工具栏和表格行间距，减少首屏过度留白。
- 优化插件详情抽屉内容区布局，为抽屉 body 增加可滚动区域和基础最小高度，避免内容区域视觉上像“空白页”。

未完成事项：

- 本轮只修复后台插件页面显示问题；未做新的插件治理能力、API、数据库或权限逻辑变更。
- 真实浏览器截图级视觉验收可在后续 UI polish 或 E2E 视觉回归专项中继续补充。

已执行检查：

- `docker compose run --rm admin-e2e npm run build`：通过，Vite 仅输出 chunk size warning。
- `docker compose run --rm admin-e2e`：通过，11 个 Playwright 用例全部通过。

跳过项及原因：

- 未执行 Go 检查：本轮没有后端代码变更。
- 未执行前台构建 / E2E：本轮没有 `web/frontend-app` 变更。

影响范围：

- API：无新增或修改。
- 数据库：无结构变更。
- 权限：无业务权限逻辑变更。
- SEO：无 SEO 行为变更。
- 插件系统：仅修复后台插件治理中心展示层问题。
- 前后台 UI：影响后台 `/admin-next/plugins` 插件列表与插件详情抽屉显示。

下一轮建议：

1. 如继续打磨后台插件治理中心，可补充浏览器截图对比或针对详情抽屉 tab 内容的更细 E2E 断言。
2. 子站插件配置抽屉如存在同类密度或空白问题，可按同样方式做小范围 UI 修复。

### 2026-05-10：插件治理中心基础 UI + 依赖接入

修改范围：

- `web/admin-app/src/views/Plugins.vue`：插件治理中心页面基础结构优化、统计卡片、筛选工具栏、列表字段增强（hooks/schema 状态等）。
- `web/admin-app/package.json` / `web/admin-app/package-lock.json`：接入治理中心相关前端依赖。

已完成事项：

- 接入依赖：`json-editor-vue`、`ajv`、`@vueuse/core`（本轮只接入与后续治理中心能力预留，不做 JSON Editor 配置编辑器重构）。
- 插件治理中心顶部新增统计卡片（基于现有 `/api/v1/admin/plugins` 返回数据实时计算，不伪造后端暂不可得字段）。
- 新增筛选工具栏：支持按 code/name 搜索、按 status、content_type、is_system、是否有 config_schema 筛选。
- 插件列表增强：增加 hooks 数量、schema 状态 badge，并保留原有“详情/配置/启用/禁用/管理”能力与禁用确认提示。

未完成事项：

- 本轮不做 JSON Editor 形态的配置编辑（仍使用 textarea + 格式化），不做影响分析、审计 Tab、子站插件抽屉等高级能力。

已执行检查：

- `cd web/admin-app && npm run build`（使用 Docker Node 环境执行）通过。

下一轮建议：

1. 在插件治理中心接入“结构化审计浏览/筛选”与 hook 失败审计展示（需要后端补齐 hook.failed/hook.blocked 写入）。
2. 基于 `ajv` 将 config_schema 校验错误更友好地呈现到 UI（仍不做复杂表单生成器）。

### 2026-05-10：插件详情抽屉 Tabs + JSON 配置编辑器

修改范围：

- `web/admin-app/src/views/Plugins.vue`：插件治理中心“详情/权限/菜单/配置”统一收口到插件详情抽屉；移除旧的 textarea 全局配置弹窗。
- `web/admin-app/src/components/plugin/PluginDetailDrawer.vue`：插件详情抽屉升级为治理视图 Tabs（概览/内容类型/权限/菜单/配置/Hooks/路由）。
- `web/admin-app/src/components/plugin/PluginJsonEditor.vue`：引入 `json-editor-vue` 作为 JSON 编辑器，并用 `Ajv` 做 `config_schema` 基础校验（客户端侧），提供格式化/复制/清空 `{}`。

已完成事项：

- 插件详情抽屉 Tabs 结构完成，字段展示更贴近治理视角（能力摘要、权限/菜单列表、路由声明等）。
- 配置 Tab 支持同时展示：
  - `config_schema`（只读）
  - `config_json`（可编辑，JSON Editor）
  - `resolved_config`（只读）
- 保存配置时：前端先做 `config_schema` 校验（Ajv），通过后调用 `PUT /api/v1/admin/plugins/:code/config`；后端仍会做二次校验与审计写入。
- Hooks Tab 不伪造“handler 存在/最近执行状态”；平台调用点仅按当前后端已确认接入的 Dispatch 列表标记，其余显示“未知/未覆盖”。

未完成事项：

- 本轮不做“子站插件配置抽屉”升级（仍按既有最小 UI）。
- 本轮不做 Hook 运行时观测（最近执行/错误追踪）；需要后端提供可查询接口或审计聚合。

已执行检查：

- `cd web/admin-app && npm run build`（使用 Docker Node 环境执行）：本轮需要重新执行并记录结果。

### 2026-05-10：子站插件配置抽屉升级

修改范围：

- `web/admin-app/src/views/Communities.vue`：升级子站“插件配置”抽屉，补顶部概览、筛选工具栏、字段增强，并将子站 `config_json` 编辑从 textarea 升级为 JSON Editor。

已完成事项：

- 子站插件配置抽屉顶部概览：展示子站名称/slug，以及子站 enabled/disabled 与全局 disabled 插件数量。
- 增加筛选能力：
  - 全部 / 子站已启用 / 子站未启用 / 全局已禁用
  - 按 name/code 搜索
  - 按 content_type 筛选
- 列表字段增强：新增“配置覆盖”字段（标记子站是否覆盖了默认/全局配置）。
- 子站 `config_json`：
  - 使用 `PluginJsonEditor`（`json-editor-vue`）编辑
  - 使用 Ajv 对 `config_schema` 做基础校验，校验失败禁止保存并展示错误
  - 支持清空为 `{}` 与保存后刷新
- 保持排序能力：数字排序 + 上移/下移 + 保存排序（不引入拖拽库）。

未完成事项：

- 本轮不做“清空覆盖配置”的后端专用接口（目前用保存 `{}` 作为最小可用方案）。

已执行检查：

- `cd web/admin-app && npm run build`（使用 Docker Node 环境执行）：本轮需要重新执行并记录结果。

### 2026-05-10：影响分析入口、审计入口与 PluginContent 优化

修改范围：

- 后端：新增插件影响分析统计接口（全局与子站范围），供禁用确认与治理中心展示使用。
- 后台 UI：
  - 全局禁用确认弹窗增加 impact 信息（不可用时显示“待接口支持/暂不可用”，不伪造）。
  - 插件详情抽屉新增“审计”Tab（复用 `admin/audit-logs`，按 `plugins#<code>` 前缀筛选，展示 old/new/metadata）。
  - `PluginContent.vue` 增强：展示 plugin/status、增加子站筛选与状态筛选、展示 plugin_code/content_type，并提供返回插件入口。

已完成事项：

- 新增接口：
  - `GET /api/v1/admin/plugins/:code/impact`
  - `GET /api/v1/admin/communities/:id/plugins/:code/impact`
- 插件禁用确认：全局与子站禁用确认均可在可用时展示 impact 计数（子站影响板块数/已有内容数等）。
- 审计入口：插件详情抽屉新增审计 Tab，可按动作关键字与 community_id 进一步筛选，并查看 `old_value/new_value/metadata_json`。
- PluginContent：支持按子站（site）与状态筛选内容列表，提升插件内容治理的可用性。

未完成事项：

- 本轮不做“影响范围统计”更深的维度（例如受影响的具体板块列表），仅提供轻量计数与入口。

已执行检查：

- `gofmt`：本轮涉及 Go 变更，需要执行并记录结果。
- `go test ./...`：本轮涉及 Go 变更，需要执行并记录结果。
- `go build`：本轮涉及 Go 变更，需要执行并记录结果。
- `cd web/admin-app && npm run build`：本轮需要执行并记录结果。

### 2026-05-10：完整插件系统优先级与文档口径校准

修改范围：

- 更新 `docs/AGENT_RULES.md`，新增“任务结果记录规则”。
- 更新 `docs/PLUGIN_ARCHITECTURE.md` 和本文件，将“完整插件系统”登记为 P0 最高优先级长期主线，并补充 P0/P1/P2/P3 阶段路线。
- 更新 `README.md`、`CHANGELOG.md`、`docs/API.md`、`docs/TESTING.md`、`docs/releases/v1.3.0.md`、`docs/releases/v1.3.1.md` 的插件平台口径。

已完成事项：

- 将“当前阶段不做插件市场 / 插件包 / 远程安装 / 在线更新 / 动态加载”的口径改为“P2/P3 阶段能力，当前未实现”。
- 将 HookBus 业务处理器、config_schema 基础校验和插件平台测试矩阵标为 P0 收口任务。
- 将 config_schema 自动表单增强、SDK、模板、依赖和版本检查、搜索 / 通知 / SEO 扩展标为 P1。
- 将插件包、安装、升级、soft uninstall、migration runner、签名校验、市场雏形标为 P2。
- 将远程市场、在线更新、动态加载能力评估、沙箱和权限隔离标为 P3。
- 校准 `projects/jobs/ai_works` 状态：已接入插件平台治理和声明，不是 Core 兼容类型，也不是完整业务插件。

未完成事项：

- 本轮不实现插件市场、插件包、远程安装、动态加载或新增插件。
- 本轮不补 QA / Docs / Wiki / Projects / Jobs / AI Works 的专属业务功能。
- P0 中 HookBus 业务处理器、完整真实 token 验收矩阵仍待代码专项；`config_schema` 已完成简化基础校验，阶段 B 已补基础自动表单，但结构化错误、更完整 JSON Schema、复杂字段和字段分组仍待后续。

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

1. P0：补 HookBus 业务处理器、统一错误日志和失败策略。
2. P0：执行完整插件系统真实 token 验收矩阵。
3. P1：继续增强 `config_schema` 自动表单复杂字段、字段分组和更完整 JSON Schema 支持。

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

- 当轮未新增插件影响范围统计接口，因此 UI 不展示绑定子站数量、启用子站列表或受影响板块数量；该能力已在后续“影响分析入口、审计入口与 PluginContent 轻量增强”任务中补齐轻量 impact 计数。
- 当轮未引入自动浏览器测试；后续已新增 Docker 化 Playwright 最小 E2E runner，后台插件详情抽屉、子站配置抽屉和禁用确认已有核心路径覆盖，完整浏览器矩阵仍需继续扩展。
- 当轮未做 `config_schema` 强校验或自动表单渲染；后续 `v1.3.2` 已补齐简化 schema 基础校验，阶段 B 已补齐基础自动表单，复杂字段、字段分组和完整 JSON Schema 仍待后续。

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
- 未执行真实浏览器矩阵：当时仓库没有自动浏览器测试 runner；后续已新增 `admin-e2e` 最小 E2E runner。

影响范围：

- API：未新增接口，继续使用现有全局插件、子站插件和版主菜单 API。
- 数据库：无结构变更。
- 权限：无后端权限逻辑变更，UI 继续依赖后端权限和菜单过滤。
- SEO：无 SEO 行为变更，保留 disabled 插件不影响历史内容 SEO 的提示。
- 插件系统：后台插件治理体验增强。
- 前后台 UI：只修改后台全局插件管理和子站插件配置 UI。

下一轮建议：

1. 扩展 `/admin-next/plugins` 和 `/admin-next/communities` 插件配置 E2E，覆盖多账号、多子站、配置保存持久化和视觉细节。
2. 继续增强 impact 的受影响对象明细列表（当前已有轻量计数）。
3. 继续推进 `config_schema` 自动表单复杂字段、字段分组和完整 JSON Schema 能力。

### 2026-05-10：影响分析入口、审计入口与 PluginContent 轻量增强

修改范围：

- 后端新增轻量影响范围统计接口（impact）：
  - `GET /api/v1/admin/plugins/:code/impact`
  - `GET /api/v1/admin/communities/:id/plugins/:code/impact`
- `web/admin-app/src/views/Plugins.vue`：全局禁用确认弹窗在可用时展示 impact 计数（不可用时显示“待接口支持/暂不可用”，不伪造）。
- `web/admin-app/src/components/plugin/PluginDetailDrawer.vue`：插件详情抽屉新增“审计”Tab，复用 `GET /api/v1/admin/audit-logs`，支持按 plugin_code + action 关键字 + community_id 筛选，并展示 `old_value/new_value/metadata_json`（有则展示，无则不伪造）。
- `web/admin-app/src/views/Communities.vue`：子站禁用确认弹窗在可用时展示该子站范围 impact 计数（板块数/已有内容/审核中内容等）。
- `web/admin-app/src/views/PluginContent.vue`：轻量增强插件内容页：显示 plugin_code/content_type、增加子站筛选、状态筛选，并提供返回插件详情入口。
- 同步文档：`docs/API.md`、`docs/TESTING.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/releases/v1.3.0.md`、`CHANGELOG.md`。

已完成事项：

- 影响分析入口：后端提供计数型 impact（子站启用数/板块数/内容数/审核中内容数/菜单声明数），前端在禁用确认中按实际返回展示，不伪造数字。
- 审计入口：插件详情抽屉新增审计 Tab，可在治理中心快速回溯插件启停/配置/排序等操作的结构化 diff（若日志未结构化则保持空或原始展示，不伪造）。
- PluginContent：补齐插件上下文信息与筛选入口，让“通用插件内容页”更接近治理中心的可用形态。

未完成事项：

- 本轮不提供“受影响对象明细列表”（例如受影响的具体板块列表），impact 仅提供轻量计数。
- 本轮审计 Tab 复用 `admin/audit-logs`：对“子站插件审计 target 命名规范”的覆盖仍需后续进一步统一（当前按 `plugins#<code>` 前缀过滤，不伪造跨 target 的聚合结果）。

已执行检查：

- `go test ./...`：通过（新增 impact 接口后同步修复了相关测试用例的 config_schema 约束输入）。
- `go build -o .devhub/devhub .`：通过。
- `cd web/admin-app && npm run build`：宿主机缺少 `npm`，失败于 `npm: command not found`。
- 使用 Docker Node 执行：`docker run --rm -v "$PWD/web/admin-app":/app -w /app node:20-alpine sh -lc "npm ci && npm run build"`：通过。

跳过项及原因：

- 本轮未修改前台代码，跳过 `cd web/frontend-app && npm run build`。

影响范围：

- API：新增 2 个 impact 接口（仅计数型字段），用于禁用前影响分析与治理提示。
- 数据库：无结构变更。
- 权限：impact 接口受 admin 权限保护（`plugin.read` / `site.read` + 子站管理范围校验）。
- SEO：无变更，继续保持禁用插件不影响历史内容访问与 SEO。
- 后台 UI：增强禁用确认、审计入口与 PluginContent 体验。

下一轮建议：

1. 在 impact 基础上补齐“受影响对象列表”接口（如确有需要），并确保不引入重查询风险。
2. 统一 admin_logs 的 target/metadata 规范，使“插件审计筛选”可覆盖全局插件与子站插件两条链路。

### 2026-05-10：插件治理中心专项验收与文档归档

修改范围：

- 本轮以验收和文档归档为主，小范围修正文档旧口径。
- 校准 `config_schema` 与 impact 状态：当前已完成简化 schema 基础校验与轻量 impact 计数；后续阶段 B 已完成基础自动表单，完整 JSON Schema 和受影响对象明细列表仍待后续。
- 更新 `docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/releases/v1.3.0.md` 和 `CHANGELOG.md`。

已完成事项：

- 后端检查：`go test ./...` 通过，`go build -o .devhub/devhub .` 通过。
- 后台构建：宿主机 `npm` 不存在；已使用 Docker Node 执行 `npm ci && npm run build` 并通过，Vite 仅输出 chunk size warning。
- 临时启动：`8090` 已被占用，因此使用 `PORT=18090 CMS_STORE=memory ./.devhub/devhub` 完成验收抽查。
- `/admin-next` 与 `/admin-next/plugins` 返回 200，后台构建产物包含 `Plugins`、`Communities`、`PluginContent` 和 `PluginJsonEditor` chunk。
- 全局插件 API 可返回插件声明、状态、`config_schema`、权限、菜单、路由和 `resolved_config`。
- impact 验收：
  - `GET /api/v1/admin/plugins/qa/impact` 返回全局轻量计数。
  - `GET /api/v1/admin/communities/1/plugins/qa/impact` 返回子站范围轻量计数。
  - 禁用确认前端代码在 impact 不可用时显示待支持文案，不伪造统计数字。
- JSON / schema 验收：
  - `PUT /api/v1/admin/plugins/qa/config` 传入合法配置返回 200。
  - 缺少 required 字段返回 400。
  - 子站插件配置字段类型错误返回 400。
- 全局禁用 / 子站启用限制验收：
  - 全局禁用 `qa` 返回 200。
  - 全局禁用后尝试启用子站 `qa` 返回 400，并提示“插件全局未启用，不能在子站启用”。
  - 验收后已恢复 `qa` 全局 enabled。
- 审计入口验收：
  - `GET /api/v1/admin/audit-logs?target=plugins%23qa` 可返回插件启停审计。
  - 审计记录包含 `old_value`、`new_value`、`metadata_json`，插件详情审计 Tab 可基于该接口展示。
- PluginContent 验收：
  - `/admin-next/qa` 返回 200。
  - `GET /api/v1/admin/posts?content_type=question` 返回内容列表。
  - 源码中已接入 plugin_code/content_type、子站筛选、状态筛选和返回入口；真实浏览器交互仍需人工矩阵。
- 身份与菜单验收：
  - 首页源码未暴露 `/admin-next` 总后台入口；版主入口为登录后按权限显示的隐藏入口。
  - `GET /api/v1/moderator/plugin-menus?community_slug=php` 使用前台 user token 返回当前子站可见插件菜单。
- SEO 回归：
  - `/topics/1/` 返回 200，源码包含 title、description、h1、article、标签链接和 Article JSON-LD。
  - `/c/php/` 返回 200，源码包含 title、description、canonical、h1、真实 topic 链接和热门标签。
  - 全局禁用 `qa` 后访问已有 question `/topics/2/` 仍返回 200，源码包含 title、description、h1、article、标签链接和 Article JSON-LD。
  - `/sitemap.xml` 和 `/robots.txt` 返回 200。

未完成事项：

- 未执行真实浏览器点击矩阵：当时仓库没有 Playwright/Cypress 等自动化 runner，本轮以构建、源码、API 和 SEO curl 验收为主；后续已新增 `admin-e2e` 最小 E2E runner。
- 插件治理中心 UI 的实际交互（Tabs 切换、JSON Editor 光标输入、复制按钮、抽屉滚动、禁用确认弹窗视觉）仍需人工浏览器验收。
- impact 当前仅为轻量计数，不提供受影响对象明细列表。
- 审计 Tab 复用 `admin/audit-logs`，全局插件审计 target 已验证；子站插件审计 target 与全局插件审计的聚合筛选仍待后续统一。

新发现风险：

- 后台 bundle 中 `PluginJsonEditor` chunk 超过 500 KB，Vite 构建只警告不失败；后续可以考虑按需加载或手动拆包。
- `8090` 在当前环境已被占用，验收使用 `18090` 临时端口；正式验收时需确认默认端口对应服务状态。

已执行检查：

- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `cd web/admin-app && npm run build`：失败，原因是宿主机没有 `npm`。
- `docker run --rm -v "$PWD/web/admin-app":/app -w /app node:20-alpine sh -lc "npm ci && npm run build"`：通过。
- `PORT=18090 CMS_STORE=memory ./.devhub/devhub`：通过，服务用于 API / SEO 抽查。
- `curl` 抽查 `/admin-next`、`/admin-next/plugins`、`/admin-next/qa`、`/api/v1/admin/plugins`、impact API、audit logs、`/topics/1/`、`/topics/2/`、`/c/php/`、`/sitemap.xml`、`/robots.txt`：通过或按预期返回错误。

跳过项及原因：

- 未执行 `cd web/frontend-app && npm run build`：本轮未修改前台代码。
- 未执行完整浏览器矩阵：当时仓库没有自动化浏览器 runner；后续已新增 `admin-e2e` 最小 E2E，完整矩阵仍需扩展。

影响范围：

- API：无新增接口；验证了已有 impact、插件、审计和内容管理接口。
- 数据库：无结构变更。
- 权限：验证了 admin impact、全局禁用、子站启用限制和版主插件菜单。
- SEO：验证了 `/topics/:id`、`/c/:slug`、sitemap 和 robots；禁用插件不影响历史内容 SEO。
- 插件系统：完成一轮命令行/API/SEO 层面的专项验收；后续已补 Docker 化 Playwright 最小 E2E，治理中心更大可视化交互矩阵仍需扩展。
- 前后台 UI：后台构建通过；真实浏览器点击矩阵仍未覆盖。

下一轮建议：

1. 扩展 Playwright E2E，从当前 5 条最小路径扩大到多账号、多子站、配置保存持久化和权限边界。
2. 为 impact 增加可分页的受影响对象明细接口，必要时只在点击“查看明细”时加载。
3. 统一插件审计 target 规范，让全局插件和子站插件审计能在同一 Tab 中准确聚合筛选。

### 2026-05-11：固定 DevHub 后台 E2E Docker 镜像，提升 Playwright 测试效率与一致性

修改范围：

- 新增 `web/admin-app/Dockerfile.e2e` 和 `web/admin-app/docker/e2e-entrypoint.sh`，固定 Playwright 基础镜像 `mcr.microsoft.com/playwright:v1.59.1-noble`，并在镜像构建阶段执行 `npm ci`。
- 新增根目录 `docker-compose.yml` 的 `admin-e2e` 服务，支持 `docker compose build admin-e2e` 与 `docker compose run --rm admin-e2e`。
- 新增后台 Playwright 配置和最小插件治理 E2E 用例。
- 将 `@playwright/test` 固定到 `1.59.1`，与 Playwright Docker 镜像版本一致。
- 更新 `.gitignore` 与 `.dockerignore`，确保 `node_modules`、`web/admin-vue`、`playwright-report` 和 `test-results` 不进入仓库。
- 更新 `docs/AGENT_RULES.md`、`docs/TESTING.md`、`docs/PROJECT_PROGRESS.md`、`docs/releases/v1.3.2.md` 和 `CHANGELOG.md`。

已完成事项：

- 后台 E2E 测试有了项目内固定 Docker 镜像；首次构建拉取大型 Playwright 基础镜像，后续复用本地 `sns-admin-e2e` 镜像。
- 后台构建和 E2E 均在容器内执行，不依赖宿主机 Node/npm。
- `admin-e2e` 支持先构建最新 `web/admin-vue` 静态产物，再跑 Playwright 测试，避免 Go 服务读到旧后台构建。
- 最小 E2E 当前覆盖 `/admin-next/plugins`、插件详情 Tabs、JSON Editor/Ajv 错误提示、全局禁用确认、子站插件抽屉和 PluginContent 入口。

未完成事项：

- 当前 E2E 是后台插件治理中心的最小路径，不覆盖多浏览器、多账号、多子站、配置保存持久化、前台发布页和版主工作台完整矩阵。
- 当前没有接入 CI workflow；本轮只固定本地/CI 可复用的 Docker 命令入口。

新发现风险：

- `web/admin-vue` 是 Go 服务读取的后台静态产物；E2E 前需要先执行 `docker compose run --rm admin-e2e npm run build`，不要和 E2E 并行写该目录，否则可能出现短暂 404 或读取不完整静态文件。
- Playwright 首次基础镜像较大，首次构建耗时正常；后续复用缓存。

已执行检查：

- `docker compose build admin-e2e`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过，Vite 仅输出 chunk size warning。
- `docker compose run --rm admin-e2e`：通过，5 个 Playwright 用例全部通过。
- `git status --short | rg 'node_modules|test-results|playwright-report|admin-vue' || true`：无输出，确认这些产物不会被提交。

跳过项及原因：

- 本轮未修改 Go 后端和前台应用，未执行 `go test ./...`、`go build`、`cd web/frontend-app && npm run build`。

影响范围：

- API：无新增或修改。
- 数据库：无结构变更。
- 权限：无权限逻辑变更。
- SEO：无 SEO 行为变更。
- 插件系统：新增后台插件治理 E2E 验收入口。
- 前后台 UI：不改 UI 视觉，只增加稳定测试选择器与 E2E 覆盖。

下一轮建议：

1. 将 `docker compose build admin-e2e` 和 `docker compose run --rm admin-e2e` 接入 CI workflow。
2. 扩展 E2E 覆盖多账号、多子站、配置保存持久化和版主菜单权限边界。

### 2026-05-11：扩展 DevHub E2E，覆盖前台与后台核心前端链路

修改范围：

- 新增 `frontend-e2e` Docker Compose 服务与 `web/frontend-app/Dockerfile.e2e`，固定 Playwright 基础镜像 `mcr.microsoft.com/playwright:v1.59.1-noble`。
- 新增 `web/frontend-app/playwright.config.js`、前台 `test:e2e` 脚本、前台 E2E helper 与 6 条前台冒烟用例。
- 扩展后台 E2E：抽取 `tests/e2e/helpers/auth.js` / `api.js`，新增后台登录与核心页面冒烟用例。
- 为前台 header、搜索页、发布页和后台核心页面补充少量稳定 `data-testid`。
- 更新 `docs/TESTING.md`、`docs/PROJECT_PROGRESS.md`、`docs/releases/v1.3.2.md` 和 `CHANGELOG.md`。

已完成事项：

- 前台 Docker runner：
  - `docker compose build frontend-e2e`
  - `docker compose run --rm frontend-e2e npm run build`
  - `docker compose run --rm frontend-e2e`
- 前台已自动化 6 条：
  - 首页打开且游客不显示总后台入口。
  - `/c/php/` 与 `/c/go/` 子站首页打开并检查 canonical。
  - 搜索页提交关键词并显示结果区域。
  - `/topics/1/` 动态详情页包含 h1、article 和 JSON-LD。
  - 未登录访问发布页并提交时提示登录。
  - 全局标签页与子站标签页打开并检查 canonical。
- 后台 E2E 从 5 条扩展到 11 条：
  - 保留插件治理中心 5 条通过用例。
  - 新增后台登录保护页边界。
  - 新增内容管理、评论管理、子站管理、标签管理、审计日志打开与筛选冒烟。

未完成事项：

- 本轮不追求完整浏览器矩阵；前台登录互动、发布成功、插件启停联动、版主工作台、多账号和跨子站权限仍待后续自动化。
- 本轮未新增 Makefile 聚合命令；当前以 Docker Compose 命令为准。

新发现风险：

- Playwright / npm 首次镜像构建对网络较敏感；前台 Dockerfile 已增加 `npm ci` fetch retry / timeout 参数，首次构建仍可能耗时较长。
- Element Plus 组件内部 DOM 不适合直接依赖透传到内部 input 的 `data-testid`；E2E 已改为页面级 `data-testid` + 稳定 placeholder / role 组合。

已执行检查：

- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `bash -n dev.sh`：通过。
- `docker compose build admin-e2e`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过，Vite 仅输出 chunk size warning。
- `docker compose run --rm admin-e2e`：通过，11 个 Playwright 用例全部通过。
- `docker compose build frontend-e2e`：首次因网络 `ETIMEDOUT` 失败；加入 npm fetch retry / timeout 后重试通过。
- `docker compose run --rm frontend-e2e npm run build`：通过。
- `docker compose run --rm frontend-e2e`：通过，6 个 Playwright 用例全部通过。

跳过项及原因：

- 未执行 `make e2e`：仓库当前没有 Makefile 聚合命令，本轮未新增。

影响范围：

- API：无新增或修改。
- 数据库：无结构变更。
- 权限：无业务权限逻辑变更；新增后台登录保护页 E2E。
- SEO：新增 `/topics/1/`、子站页和标签页的浏览器 SEO 冒烟断言。
- 插件系统：保留并通过既有插件治理中心 E2E。
- 前后台 UI：只增加稳定测试锚点，不改变视觉和业务交互。

下一轮建议：

1. 继续扩展前台 E2E：登录、点赞、收藏、关注、评论、发布成功和用户中心。
2. 继续扩展插件联动 E2E：后台禁用插件后前台入口隐藏、强传非法 `content_type` 被接口拒绝、历史内容 SEO 不受影响。
3. 继续扩展版主 E2E：普通用户拒绝、版主可访问授权子站、跨子站不可越权。

### 2026-05-11：新增前后台前端统一检测脚本，统一 DevHub 前台与后台构建/E2E 检查入口

修改范围：

- 新增 `scripts/check-frontend.sh`，统一调度后台 `admin-e2e` 和前台 `frontend-e2e` 的 build / E2E / 可选 lint/typecheck。
- 更新 `docs/AGENT_RULES.md`、`docs/TESTING.md` 和 `docs/PROJECT_PROGRESS.md`。

已完成事项：

- 脚本自动定位项目根目录，支持 `docker compose` 和 `docker-compose`。
- 支持检查范围选择：
  - `--target admin`
  - `--target frontend`
  - `--target both`
  - 兼容 `--admin-only` / `--frontend-only`
- 交互式终端直接运行且没有传 target 时会询问检查范围；非交互环境默认 `both`，避免 CI 卡住。
- 支持 `--quick`、`--strict`、`--build-only`、`--e2e-only`、`--no-build`、`--no-e2e`、`--rebuild`、`--remove-orphans`、`--tail-lines`。
- 默认实时显示日志并落盘到 `.devhub/checks/{timestamp}/`；新增 `--quiet` 可切换为摘要模式。
- 支持 PASS / FAIL / SKIP 汇总表；缺失 compose 服务或缺失可选 npm script 时显示 SKIP，不误报失败。

未完成事项：

- 未新增 Makefile 聚合命令；当前推荐直接使用 `./scripts/check-frontend.sh`。

新发现风险：

- 本轮发现 `docs/TESTING.md` 曾被误写成旧 E2E 脚本内容；已恢复为测试文档并补充统一检查脚本说明。

已执行检查：

- `bash -n scripts/check-frontend.sh`：通过。
- `./scripts/check-frontend.sh --help`：通过。
- `./scripts/check-frontend.sh --target admin --quick --quiet`：通过。
- `./scripts/check-frontend.sh --target frontend --quick`：通过，验证默认实时日志输出。
- `./scripts/check-frontend.sh --target both --quiet`：通过，后台 build + E2E、前台 build + E2E 全部 PASS。

跳过项及原因：

- 未执行 frontend-e2e 缺失时的 SKIP 场景：当前仓库已有 `frontend-e2e` 服务，因此本轮验证实际运行通过。

影响范围：

- API：无新增或修改。
- 数据库：无结构变更。
- 权限：无业务权限逻辑变更。
- SEO：无 SEO 行为变更。
- 插件系统：无业务逻辑变更；统一检查入口会覆盖现有插件治理 E2E。
- 前后台 UI：无页面视觉或交互变更。

下一轮建议：

1. 如团队需要更短命令，可新增轻量 Makefile alias：`make check-frontend`、`make check-frontend-quick`。
2. 后续 CI 可直接调用 `./scripts/check-frontend.sh --target both --quiet` 或按阶段拆分 admin/frontend。

### 2026-05-11：检查并修复本地 DevHub E2E 配置一致性，确保前后台 E2E 在本地代码中完整可运行

修改范围：

- 以当前本地工作区为准核对 `frontend-e2e`、`admin-e2e`、Playwright 配置、Dockerfile、docker compose、package scripts、lock 文件、entrypoint 和文档。
- 使用 Docker 容器内 npm 校验 `web/frontend-app/package-lock.json` 与 `package.json` 一致。
- 更新 `docs/TESTING.md` 与 `docs/PROJECT_PROGRESS.md`，记录本轮本地一致性验收结果。

已完成事项：

- `web/frontend-app/package.json` 已包含 `test:e2e` / `test:e2e:headed`，并声明 `@playwright/test: 1.59.1`。
- `web/frontend-app/package-lock.json` 已包含 `@playwright/test` 与 `playwright 1.59.1`，并用 Docker 容器内 npm 校验为 up to date。
- `docker-compose.yml` 已包含 `frontend-e2e` 服务，使用 `web/frontend-app/Dockerfile.e2e`、`DEVHUB_E2E_ORIGIN`、`host.docker.internal:host-gateway` 和 `frontend_e2e_node_modules` 命名卷。
- `web/frontend-app/Dockerfile.e2e` 顺序符合缓存要求：先复制 package 文件并 `npm ci`，再复制源码；保留 npm retry / timeout 参数。
- `web/frontend-app/docker/e2e-entrypoint.sh` 从 `/app/node_modules/.` 复制内容到 `./node_modules/`，不会产生 `node_modules/node_modules` 嵌套。
- 后台 `admin-e2e` 配置未退化，已有插件治理 E2E 仍通过。

修复事项：

- 本轮未发现需要修改代码配置的 E2E 一致性问题；仅用容器内 npm 校验 lock 文件，并补充测试文档对 node_modules 命名卷与 entrypoint 复制策略的说明。

已执行检查：

- `docker run --rm -v /home/liuwei/code/sns/web/frontend-app:/app -w /app node:20-alpine sh -lc 'npm install --package-lock-only --ignore-scripts --no-audit --fund=false'`：通过，`up to date`。
- `bash -n dev.sh`：通过。
- `bash -n web/frontend-app/docker/e2e-entrypoint.sh`：通过。
- `bash -n web/admin-app/docker/e2e-entrypoint.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose config`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose build admin-e2e`：通过。
- `docker compose build frontend-e2e`：通过，依赖层命中缓存。
- `docker compose run --rm admin-e2e npm run build`：通过，Vite 仅输出 chunk size warning。
- `docker compose run --rm frontend-e2e npm run build`：通过。
- `docker compose run --rm admin-e2e`：通过，11 个 Playwright 用例全部通过。
- `docker compose run --rm frontend-e2e`：通过，6 个 Playwright 用例全部通过。
- `./scripts/check-frontend.sh --help`：通过。
- `./scripts/check-frontend.sh --quick --quiet`：通过，前后台 build 均 PASS。
- `./scripts/check-frontend.sh --admin-only --quiet`：通过，后台 build + E2E 均 PASS。
- `./scripts/check-frontend.sh --frontend-only --quiet`：通过，前台 build + E2E 均 PASS。

跳过项及原因：

- 未模拟 `frontend-e2e` 服务缺失时的 SKIP 场景：当前本地 compose 已完整包含 `frontend-e2e`，本轮以真实可运行为验收重点。

影响范围：

- API：无新增或修改。
- 数据库：无结构变更。
- 权限：无业务权限逻辑变更。
- SEO：无 SEO 行为变更；前台 E2E 继续覆盖 `/topics/1/` 动态 SEO 冒烟。
- 插件系统：无业务逻辑变更；后台 E2E 继续覆盖插件治理中心核心路径。
- 前后台 UI：无页面视觉或交互变更。

下一轮建议：

1. 进入第二阶段 E2E 覆盖前，优先继续扩展登录用户互动、发布成功、插件启停联动和版主工作台矩阵。
2. 如需 CI 收口，可直接接入 `./scripts/check-frontend.sh --target both --quiet`。

### 2026-05-11：HookBus 运行时与 Hook 执行记录

修改范围：

- 后端：`internal/plugins/hookbus.go`、`internal/service/service.go`、`internal/store/memory.go`、`internal/store/mysql.go`、`internal/transport/httpapi/router.go`、`internal/domain/models.go`。
- 数据库：`internal/store/schema.go`、`db/mysql/001_schema.sql`、`db/mysql/migrations/010_hook_executions.sql`。
- 后台：`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`、`web/admin-app/src/api/admin.js`。
- 测试：`internal/service/hookbus_test.go`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/releases/v1.3.2.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- HookBus 新增 `DispatchWithResults`，返回每个内置 handler 的执行结果、耗时、成功/失败和错误信息。
- `HookContext` 补齐 `hook_name`、`channel_id`、`user_id`、`admin_user_id` 等运行治理字段。
- Service 层统一通过 `dispatchHook` 执行 Hook，按全局插件状态和子站插件状态过滤，插件启用 / 禁用生命周期 Hook 允许在状态切换时执行。
- 新增 `hook_executions` 执行记录，MemoryStore 和 MySQLStore 均支持写入、最近执行查询和聚合统计。
- blocking hook 失败会阻断主流程，并写入 `plugin.hook.blocked` 审计。
- non-blocking hook 失败不阻断主流程，并写入 `plugin.hook.failed` 审计。
- impact API 的 `recent_hook_errors_count` 已改为统计最近 7 天 `hook_executions` 失败记录。
- 新增 `GET /api/v1/admin/plugins/:code/hooks`，返回 Hook 统计与最近 20 条执行记录。
- 后台插件详情 Hooks Tab 接入真实运行统计：执行次数、失败次数、失败率、平均耗时、最近执行、最近失败、最近错误和最近执行列表。
- qa / docs / wiki 继续使用内置最小 Hook handler；qa/docs/wiki 的 BeforeCreateContent 会校验对应内容类型。

未完成事项：

- HookBus 仍只服务内置系统插件，不支持第三方动态插件、远程 Hook 或 Webhook。
- Search / Notification / SEO 仍是最小事件派发，缺少完整插件业务处理器。
- 尚未实现 Hook 健康状态、告警、重试策略、异步队列和跨 Store 事务回滚封装。
- 后台 Hooks Tab 展示真实执行记录，但不伪造 handler 是否存在；handler 状态仍需后续平台能力补充。

新发现风险：

- Hook handler 若未来写入外部资源，blocking 失败和主数据写入之间需要补事务边界设计。
- 非阻断 Hook 当前只记录失败并审计，没有自动重试；通知、搜索索引类 Hook 后续需要幂等和重试策略。

已执行检查命令和结果：

- `gofmt -w internal/domain/models.go internal/plugins/hookbus.go internal/service/service.go internal/service/hookbus_test.go internal/store/memory.go internal/store/mysql.go internal/transport/httpapi/router.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过（Vite 输出 chunk size warning，非失败）。

失败项或跳过项及原因：

- 完整浏览器矩阵未执行；本轮只要求 HookBus 运行时和后台构建轻量验收。

影响范围：

- API：新增 `GET /api/v1/admin/plugins/:code/hooks`。
- 数据库：新增 `hook_executions` 表，新装 schema 与老库迁移均已补齐。
- 权限：新增接口使用 `plugin.read`；未改变发布权限矩阵。
- SEO：`OnSEOBuild` 仍为 non-blocking，历史详情和 SEO 不应受影响。
- 插件系统：HookBus 从可派发升级为可记录、可审计、可查询。
- 前后台 UI：后台插件详情 Hooks Tab 展示运行时统计；前台 UI 无变更。

下一轮建议：

1. 为 HookBus 增加健康状态、告警和失败重试策略。
2. 为 Search / Notification / SEO 补真实插件业务处理器和幂等设计。
3. 在 E2E 中增加后台 Hooks Tab 和 `hook_executions` 可视化回归。

### 2026-05-11：插件审计、健康状态与后台治理中心增强

修改范围：

- 后端：`internal/domain/models.go`、`internal/service/service.go`、`internal/transport/httpapi/router.go`。
- 后台：`web/admin-app/src/views/Plugins.vue`、`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`、`web/admin-app/src/api/admin.js`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/releases/v1.3.2.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 新增轻量 `PluginHealth` 摘要，`GET /api/v1/admin/plugins` 会为后台返回 `health` 字段。
- 健康状态计算来源包括：插件全局状态、`config_schema` 对全局配置的校验结果、`plugin_migrations` pending/failed 状态、Hook 失败统计和依赖插件状态。
- 当前健康状态覆盖 `healthy`、`warning`、`error`、`disabled`、`migration_pending`、`config_invalid`、`dependency_missing`。
- 新增 `GET /api/v1/admin/plugins/:code/audit-logs`，插件详情“审计”Tab 改用插件专用审计查询入口。
- 插件内容治理操作补充结构化审计：隐藏/恢复、置顶、加精、评论锁、评论隐藏/恢复和批量主题治理会在插件内容上写入带 `plugin_code/content_type/community_id/category_id/content_id` 的 metadata。
- 后台插件列表展示运行健康、配置状态、迁移状态、Hook 状态、最近错误和建议操作。
- 插件详情抽屉新增“运行状态”Tab，展示 overall/config/migration/hook/dependency、pending/failed migrations、hook failures、recent error 和 suggested action。
- 全局禁用确认中的近期 Hook 错误改为展示 `impact.recent_hook_errors_count` 实际计数，不再显示“暂未接入”。

未完成事项：

- 健康状态当前是后台治理摘要，不是完整监控系统；尚未提供独立健康 API、告警规则、自动恢复或 Prometheus/Grafana 集成。
- 插件迁移仍缺后台执行/重试/失败详情 UI，健康状态只能基于已有 `plugin_migrations` 记录计算。
- 插件内容治理审计已覆盖当前已有操作，但专属插件业务管理页和更复杂批量操作仍待后续完善。

新发现风险：

- 插件审计查询按插件 code/target 做轻量筛选；若历史日志没有结构化 plugin_code 或 target 不含插件 code，仍不会出现在插件专用审计 Tab。
- Hook 失败会影响健康摘要，但当前没有“连续失败阈值”和自动告警策略，后续需要避免 warning 长期被忽略。

已执行检查命令和结果：

- `gofmt -w internal/domain/models.go internal/service/service.go internal/transport/httpapi/router.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过（Vite 输出 chunk size warning，非失败）。

失败项或跳过项及原因：

- 完整浏览器矩阵未执行；本轮只要求插件审计、健康状态与后台治理中心的轻量验收。

影响范围：

- API：新增 `GET /api/v1/admin/plugins/:code/audit-logs`；`GET /api/v1/admin/plugins` 增加后台 `health` 字段。
- 数据库：无新增表或字段；复用已有 `admin_logs.old_value/new_value/metadata_json` 与 `hook_executions`。
- 权限：新增审计接口使用 `plugin.read`；未改变发布权限矩阵。
- SEO：无 SEO 路由变更；插件禁用不影响历史内容详情和 SEO 的规则保持不变。
- 插件系统：补齐轻量健康摘要、插件专用审计入口和插件内容治理审计 metadata。
- 前后台 UI：后台插件列表和详情抽屉增强；前台 UI 无变更。

下一轮建议：

1. 为插件迁移继续补真实 rollback/down、迁移前备份和 E2E 覆盖。
2. 为 Hook 健康状态增加连续失败阈值、告警和重试策略。
3. 扩展 E2E 覆盖插件健康状态、审计 Tab 和插件内容治理审计回归。

### 2026-05-11：插件迁移闭环

修改范围：

- 后端：`internal/domain/models.go`、`internal/plugins/registry.go`、`internal/plugins/qa/qa.go`、`internal/plugins/docs/docs.go`、`internal/plugins/wiki/wiki.go`、`internal/service/service.go`、`internal/store/memory.go`、`internal/store/mysql.go`、`internal/transport/httpapi/router.go`。
- 数据库：`db/mysql/001_schema.sql`、`internal/store/schema.go`、`db/mysql/migrations/008_plugin_migrations.sql`。
- 后台：`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`、`web/admin-app/src/api/admin.js`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/releases/v1.3.2.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 扩展 `PluginManifest`，支持 `migrations` 声明；qa/docs/wiki 已声明第一批内置 up migration。
- 扩展 `PluginMigration` 返回字段，兼容 `version`，同时暴露 `migration_version`、`direction`、`duration_ms`、`rollback_supported`、`declared` 等治理字段。
- `plugin_migrations.status` 扩展为 `pending/running/success/failed`，并新增 `updated_at`；新装 schema、启动迁移和老库迁移脚本已同步。
- `Service` 新增内置 migration runner：可列出声明、执行所有待处理 migration、执行/重试单条 migration；已成功 migration 不重复破坏数据。
- 当前 runner 是内置 up/no-op runner：qa/docs/wiki 的扩展表由主 schema / 启动迁移保证，runner 负责记录 running/success/failed、耗时和错误。
- 新增后台 API：`GET /api/v1/admin/plugins/:code/migrations`、`POST /api/v1/admin/plugins/:code/migrations/run`、`POST /api/v1/admin/plugins/:code/migrations/:name/retry`。
- 迁移操作写入审计：`plugin.migration.run`、`plugin.migration.retry`、`plugin.migration.success`、`plugin.migration.failed`。
- 插件详情抽屉新增“迁移”Tab，展示迁移列表、状态、最近执行、失败原因、rollback 标识，并提供执行/重试入口。
- 插件健康状态继续从 `plugin_migrations` pending/failed 记录计算 `migration_status`。

未完成事项：

- 不支持 migration down、真实 rollback、迁移前备份或外部插件迁移包。
- 不支持复杂迁移依赖排序、批量跨插件迁移计划和迁移影响对象明细。
- 未做完整浏览器 E2E，只执行后台构建作为轻量验收。

新发现风险：

- 当前内置 migration runner 是 no-op 记录型 runner，适合确认主 schema 已具备表结构；后续若 migration 真正改表，必须加入备份、事务边界、失败恢复和 MySQL/MemoryStore 差异处理。
- 老库执行 `008_plugin_migrations.sql` 会修改 enum 并补 `updated_at`，生产执行前仍需备份并在预发库演练。

已执行检查命令和结果：

- `gofmt -w internal/domain/models.go internal/plugins/registry.go internal/plugins/qa/qa.go internal/plugins/docs/docs.go internal/plugins/wiki/wiki.go internal/service/service.go internal/store/memory.go internal/store/mysql.go internal/transport/httpapi/router.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过（Vite 输出 chunk size warning，非失败）。

失败项或跳过项及原因：

- 未执行完整 E2E；本轮只要求插件迁移闭环的轻量验收。

影响范围：

- API：新增插件迁移查询、执行和重试接口。
- 数据库：扩展 `plugin_migrations.status` 与 `updated_at`，不删除历史数据。
- 权限：新增迁移查询使用 `plugin.read`，执行/重试使用 `plugin.write`。
- SEO：无 SEO 路由变更；插件 disabled / migration 状态不影响历史内容详情访问。
- 插件系统：迁移从“记录表”升级为“内置 up/no-op runner + 后台 Tab + 审计”。
- 前后台 UI：后台插件详情新增迁移 Tab；前台 UI 无变更。

下一轮建议：

1. 为迁移 runner 增加真实 SQL migration 执行能力前，先设计备份、事务边界和失败恢复策略。
2. 为迁移 Tab 增加 E2E 覆盖：列表、执行、重复执行不破坏、失败重试和审计记录。
3. 继续补 Hook 告警/重试和插件健康独立详情 API。

### 2026-05-11：插件系统专项验收与 E2E 回归清单归档

修改范围：

- 后台 E2E：`web/admin-app/tests/e2e/plugin-governance.spec.js`。
- 文档：`docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/releases/v1.3.2.md`、`CHANGELOG.md`。

已完成事项：

- 执行插件系统专项验收，覆盖 Go 单测、Go 构建、后台 Docker 构建、前后台 Playwright E2E、`/topics/1/` 与 `/c/php/` SEO 动态 HTML curl 回归。
- 前台 E2E 通过 14 条，覆盖首页、子站页、搜索页、Topic 动态详情、未登录发布拦截、标签页、登录互动、发布成功、插件禁用强拦截和版主权限边界。
- 后台 E2E 首次执行发现 2 个验收侧问题：全局禁用 `qa` 的用例失败后未恢复插件状态，污染并行执行的插件内容页测试；另一个断言仍使用旧影响分析文案。
- 已小范围修复后台 E2E：插件治理用例改为 serial，并在 `finally` 中恢复 `qa` enabled；影响分析断言同步为真实展示字段“当前启用子站 / 将阻止发布的板块”。
- 后台 E2E 复跑通过 15 条，覆盖后台登录、内容、评论、举报、子站、标签、审计、插件治理、子站插件配置和通用插件内容页。
- SEO curl 验证 `/topics/1/` 包含 title、canonical、Article JSON-LD、article、h1；`/c/php/` 包含 title、description、canonical、h1 和 WebSite JSON-LD。

未完成事项：

- 本轮未执行真实 API 手工矩阵逐项开关插件 / 修改配置 / 执行迁移；相关能力由现有自动化 E2E、Go 单测和文档验收清单覆盖，后续仍可扩展专项 API 回归。
- 插件迁移 E2E 目前仍以后台 UI/接口可见性为主，尚未覆盖失败注入、重试后状态恢复和审计定位的完整浏览器流程。
- Hook 治理 E2E 尚未覆盖人为制造 blocking / non-blocking Hook 失败后的后台记录展示。

新发现风险：

- 插件启停类 E2E 若失败后未恢复状态，会污染后续用例和本地演示环境；此类测试必须串行或在 `finally` / `afterEach` 中恢复原状态。
- 当前后台构建仍有 Vite chunk size warning，非失败；后续可考虑拆分 JSON Editor / Content 等大 chunk。

已执行检查命令和结果：

- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `bash -n dev.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过（Vite chunk size warning，非失败）。
- `curl -fsS --max-time 5 http://127.0.0.1:8090/topics/1/ | rg -i "<title|<h1|<article|application/ld\\+json|canonical"`：通过。
- `curl -fsS --max-time 5 http://127.0.0.1:8090/c/php/ | rg -i "<title|<h1|canonical|description"`：通过。
- `./scripts/check-frontend.sh --frontend-only --e2e-only`：通过，前台 14 passed。
- `./scripts/check-frontend.sh --admin-only --e2e-only`：首次失败 2 项，修复后复跑通过，后台 15 passed。

失败项或跳过项及原因：

- 首次后台 E2E 失败已修复并复跑通过。
- 未执行前台构建：本轮未修改前台代码，且前台 E2E 已通过。
- 未新增数据库迁移或 API，因此未执行 MySQL 老库升级演练。

影响范围：

- API：无新增或变更。
- 数据库：无新增或变更。
- 权限：无权限模型变更；验收确认前后台 E2E 仍覆盖基础权限边界。
- SEO：无策略变更；curl 回归确认动态 HTML 关键元素仍存在。
- 插件系统：修复插件治理 E2E 的状态恢复与文案断言，降低插件启停测试污染风险。
- 前后台 UI：无 UI 功能变更；后台 E2E 测试断言更新。

下一轮建议：

1. 补插件迁移失败注入 / 重试 / 审计定位的自动化 E2E。
2. 补 Hook blocking / non-blocking 失败注入的 API 或 E2E 回归。
3. 将插件启停、配置和迁移操作的状态恢复 helper 统一抽象，减少测试间污染。

### 2026-05-11：v1.3.3 插件平台治理收口

修改范围：

- 后端：`internal/service/service.go`、`internal/service/hookbus_test.go`。
- 版本与文档：`VERSION`、`README.md`、`docs/README.md`、`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/PROJECT_PROGRESS.md`、`docs/releases/v1.3.3.md`、`CHANGELOG.md`。

已完成事项：

- 新建 `docs/releases/v1.3.3.md`，定义 v1.3.3 为“插件平台治理收口版”。
- 将 VERSION 更新为 `v1.3.3`，并同步 README、docs 入口和 CHANGELOG 当前版本口径。
- Service 层补齐插件启用 readiness 检查：全局启用和子站启用都会校验插件存在、全局配置符合 `config_schema`、依赖插件已启用、没有 `failed` 迁移。
- 明确当前内置 migration 仍是 up/no-op 记录型迁移：`pending` migration 通过 health / 迁移 Tab 提示，但不阻断启用；`failed` migration 会阻断全局启用和子站启用。
- 补充单测覆盖：pending 内置 no-op migration 不阻断启用，failed migration 会阻断全局启用和子站启用。
- 文档统一说明生命周期状态：枚举已进入 schema / Store，但完整自动状态机仍未完成；当前只有 `enabled` 放行新建内容。
- 文档统一说明 config_schema、effective_config、HookBus、plugin_migrations、权限矩阵、后台治理中心和 `post.create` 兼容桥的真实边界。

未完成事项：

- 仍未实现完整自动生命周期状态机。
- 仍未实现真正独立 Migration Runner、migration down、真实 rollback、迁移前备份和外部插件迁移包。
- HookBus 仍只服务内置插件，尚无告警、重试策略、异步队列或第三方动态 Hook。
- 权限矩阵仍是最小权限码闭环；完整 RBAC 配置 UI、community/category 作用域分配和更细错误码仍待后续。

新发现风险：

- 若未来把 pending migration 改为严格阻断启用，需要先区分“内置 no-op 声明 pending”和“真实 DDL migration pending”，否则默认内置插件可能被未手动确认的 no-op 迁移卡住。

已执行检查命令和结果：

- `gofmt -w internal/service/service.go internal/service/hookbus_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过（Vite chunk size warning，非失败）。

失败项或跳过项及原因：

- 未执行完整前后台 E2E：上一轮插件系统专项验收已完成前台 14 条和后台 15 条 E2E；本轮只改 Service readiness 与文档口径。
- 未执行前台构建：本轮未修改前台代码。

影响范围：

- API：无新增路由；全局/子站插件启用 API 的 Service 层校验更严格，失败迁移会阻断启用。
- 数据库：无结构变更。
- 权限：无权限模型变更；继续保留 `post.create` 作为 `core.topic.create` 兼容桥。
- SEO：无 SEO 代码变更；插件 disabled 不影响历史内容详情和 `/topics/:id` 动态 HTML 的规则不变。
- 插件系统：补齐启用 readiness 收口，统一生命周期、配置、Hook、迁移、权限和后台治理中心文档口径。
- 前后台 UI：无 UI 代码变更。

下一轮建议：

1. 为 failed migration 阻断启用补 API/E2E 回归，并覆盖后台错误提示。
2. 为 HookBus 补 blocking/non-blocking 失败注入 E2E 和后台 Hooks Tab 可见性断言。
3. 设计真实 Migration Runner 的备份、事务、失败恢复和 rollback/down 边界，再推进外部插件阶段。

### 2026-05-11：文档版本号口径对齐

修改范围：

- 文档：`docs/BACKUP_AND_ROLLBACK.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 当轮确认版本源头为仓库根目录 `VERSION = v1.3.3`。
- 当轮确认 Release Notes 为 `docs/releases/v1.3.3.md`，README、docs 入口、CHANGELOG 和 API/架构/测试文档主体均已指向 v1.3.3。
- 修正 `docs/BACKUP_AND_ROLLBACK.md` 中仍写旧版号 `v1.3.0` 的陈旧口径，改为以 `VERSION` 为准并标注当前为 `v1.3.3`。
- 保留 `CHANGELOG.md`、历史 Release Notes 和 `docs/PROJECT_PROGRESS.md` 历史任务记录中的旧版本号，它们属于归档追溯，不是当前版本口径错误。

未完成事项：

- 无。本轮为文档版本口径校准，不涉及业务代码、API、数据库或 UI。

新发现风险：

- 后续若更新 VERSION，需要同步检查备份 / 部署 / 文档入口中的“当前版本”自然语言描述，避免旧版本号残留。

已执行检查命令和结果：

- `cat VERSION`：通过，输出 `v1.3.3`。
- `rg "当前版本.*v1\\.3\\.[0-2]|当前文档.*v1\\.3\\.[0-2]|当前版本为 v1\\.3\\.[0-2]" README.md docs/*.md CHANGELOG.md VERSION .github`：修复后仅命中 `docs/PROJECT_PROGRESS.md` 历史任务记录，当前有效文档无 live 口径残留。
- `test -f docs/releases/v1.3.3.md`：通过。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 未执行 Go / 前后台构建：本轮只修改文档，不涉及代码或构建产物。

影响范围：

- API：无影响。
- 数据库：无影响。
- 权限：无影响。
- SEO：无影响。
- 插件系统：无运行时影响，仅统一版本文档口径。
- 前后台 UI：无影响。

下一轮建议：

1. 后续发布版本时，把 `VERSION`、`README.md`、`CHANGELOG.md`、`docs/README.md`、当前 Release Notes 和备份 / 部署文档作为固定版本口径检查项。

### 2026-05-11：整理下一阶段插件平台需求

修改范围：

- 文档：`docs/PLUGIN_SYSTEM_ROADMAP.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 将“插件迁移失败注入 + Hook 失败注入 E2E”整理为下一阶段明确需求。
- 在 `docs/PLUGIN_SYSTEM_ROADMAP.md` 新增“v1.3.4 插件迁移与 Hook 失败注入验收闭环”需求块。
- 明确下一阶段仍属于 P0 插件运行治理闭环，不做插件市场、上传安装、远程安装、Go 动态加载或具体业务插件增强。
- 将下一阶段拆分为 5 个范围：
  - 插件迁移失败注入与启用阻断。
  - HookBus blocking / non-blocking 失败注入。
  - 插件权限矩阵继续收口。
  - MySQLStore / 老库升级专项。
  - P1 体验增强准备。
- 为每个范围补充目标、需求和验收口径，后续可直接作为任务输入。

未完成事项：

- 本轮只整理需求，未实现代码、API、UI 或 E2E。

新发现风险：

- 插件迁移失败、Hook 失败和插件启停类 E2E 都可能污染全局状态；后续实现时必须串行执行或在 `finally` / `afterEach` 中恢复状态。

已执行检查命令和结果：

- `sed -n '1,260p' docs/PLUGIN_SYSTEM_ROADMAP.md`：通过，确认需求块已写入。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 未执行 Go / 前后台构建：本轮只修改文档，不涉及代码或构建产物。

影响范围：

- API：无影响。
- 数据库：无影响。
- 权限：无影响。
- SEO：无影响。
- 插件系统：新增下一阶段需求定义，无运行时影响。
- 前后台 UI：无影响。

下一轮建议：

1. 按 `docs/PLUGIN_SYSTEM_ROADMAP.md` 中的 v1.3.4 需求，优先实现插件迁移失败注入、启用阻断和恢复验收。
2. 随后补 HookBus blocking / non-blocking 失败注入，并把后台 Hooks Tab 可见性纳入 E2E。

### 2026-05-11：根据下一阶段插件目标统一文档口径

修改范围：

- 文档：`README.md`、`CHANGELOG.md`、`docs/README.md`、`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/SEO.md`、`docs/TESTING.md`、`docs/releases/v1.3.3.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 以 `docs/PLUGIN_SYSTEM_ROADMAP.md` 和本文档为目标源头，统一其他文档的下一阶段口径。
- 当轮明确版本仍是 `v1.3.3`，下一阶段需求为 `v1.3.4：插件异常治理与验收闭环版`。
- 在 README Roadmap、CHANGELOG Next、API 规划、插件架构、测试矩阵、SEO 红线和 v1.3.3 Release Notes 后续计划中同步 v1.3.4 目标。
- 统一 v1.3.4 范围：
  - 插件迁移失败注入与启用阻断。
  - HookBus blocking / non-blocking 失败注入。
  - 插件权限矩阵继续收口。
  - MySQLStore / 老库升级专项。
  - P1 体验增强准备。
- 统一不做范围：不做插件市场、上传安装、远程安装、Go 动态加载、新业务插件或大规模 UI 重构。
- 在 SEO 文档中补充 v1.3.4 异常治理必须继续保护历史 `/topics/:id` 和 `/c/:slug` SEO。

未完成事项：

- 本轮只做文档口径同步，未实现 v1.3.4 代码、API、UI 或 E2E。

新发现风险：

- v1.3.4 涉及迁移失败、Hook 失败和插件状态切换，后续 E2E 必须串行或恢复状态，否则容易污染本地演示和后续测试。

已执行检查命令和结果：

- `rg "v1\\.3\\.4|插件迁移失败注入|HookBus blocking|MySQLStore / 老库升级专项" README.md CHANGELOG.md docs`：通过，确认目标已同步到主要文档。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 未执行 Go / 前后台构建：本轮只修改文档，不涉及代码或构建产物。

影响范围：

- API：无实际接口变更，仅补充下一阶段规划口径。
- 数据库：无影响。
- 权限：无运行时影响，仅补充权限矩阵下一阶段目标。
- SEO：无运行时影响，仅补充异常治理阶段 SEO 红线。
- 插件系统：无运行时影响，统一下一阶段插件平台目标。
- 前后台 UI：无影响。

下一轮建议：

1. 直接按 v1.3.4 需求启动代码实现：先做插件迁移失败注入、启用阻断、retry 恢复和审计 / E2E。
2. 第二步做 HookBus blocking / non-blocking 失败注入与后台 Hooks Tab 可见性 E2E。

### 2026-05-11：实现插件迁移失败注入与启用阻断闭环

修改范围：

- 后端：`internal/service/service.go`、`internal/transport/httpapi/router.go`、`internal/transport/httpapi/router_auth_test.go`。
- 后台 E2E：`web/admin-app/tests/e2e/helpers/api.js`、`web/admin-app/tests/e2e/plugin-governance.spec.js`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 新增测试专用 failed migration 注入接口：`POST /api/v1/admin/plugins/:code/migrations/:name/e2e-fail`。
- 注入接口仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用；它不是生产迁移治理入口。
- 注入 failed migration 后，全局启用插件会被 Service readiness 阻断，错误包含“失败迁移”。
- 注入 failed migration 后，子站启用同一插件也会被 Service readiness 阻断，且不会错误写入 `enabled`。
- 迁移 Tab 能显示 failed migration 的名称、状态和错误信息，并可通过 retry 恢复为 `success`。
- retry 成功后，全局插件和子站插件均可恢复启用。
- 已 success 的 migration 再次 retry 保持成功记录，不重复破坏数据。
- 注入失败、retry 和 success 恢复均写入插件审计，可通过插件审计接口定位。
- 新增 API 测试覆盖阻断、重试、恢复和审计。
- 新增后台 E2E 覆盖迁移 Tab 失败原因、retry 操作、启用恢复和审计定位。

未完成事项：

- HookBus blocking / non-blocking 失败注入仍未完成，是下一项 v1.3.4 P0 任务。
- MySQLStore / 老库升级专项当轮未执行；已在后续 `2026-05-11：MySQLStore 与老库升级专项验证插件平台一致性` 中补测。
- 当前 migration runner 仍是内置 up/no-op 记录型 runner，不支持 migration down、硬回滚、迁移前备份或外部插件迁移包。

新发现风险：

- 测试注入接口如果在生产环境误开会带来治理状态污染风险；因此必须保持 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 的环境限制，并在文档中明确。
- 迁移失败类 E2E 会切换全局和子站插件状态，必须继续保持串行和 `finally` 恢复。

已执行检查命令和结果：

- `gofmt -w internal/service/service.go internal/transport/httpapi/router.go internal/transport/httpapi/router_auth_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；仍有既有 Vite chunk size warning，不影响构建结果。
- `./scripts/check-frontend.sh --quick`：通过，前后台容器构建通过；日志目录 `.devhub/checks/20260511-182649/`。
- `./scripts/check-frontend.sh --admin-only`：首次失败，原因是本地 DevHub 服务仍是旧路由，`e2e-fail` 返回“接口不存在”；执行 `./dev.sh restart --no-build --local-go` 重启到当前代码后重跑。
- `./scripts/check-frontend.sh --admin-only`：第二次失败，原因是 E2E 选择器过宽，页面中运行状态和迁移 Tab 都出现 `failed`；已收窄到插件详情抽屉迁移面板。
- `./scripts/check-frontend.sh --admin-only`：最终通过，后台 build 通过，后台 E2E `16 passed`；日志目录 `.devhub/checks/20260511-183007/`。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 已修复两次后台 E2E 失败：
  - 旧服务未重启导致新接口不存在。
  - Playwright strict mode 命中多个 `failed` 文本，已收窄选择器。
- 未执行完整前台 E2E：本轮未修改前台代码；已执行 `./scripts/check-frontend.sh --quick` 覆盖前台构建。

影响范围：

- API：新增测试专用迁移失败注入接口；生产插件治理 API 语义不变。
- 数据库：无 schema 变更。
- 权限：注入接口需要后台 admin token 和 `plugin.write`，且受测试环境限制。
- SEO：无运行时 SEO 变更；插件禁用后历史内容访问策略不变。
- 插件系统：补齐 failed migration 对全局 / 子站启用的可验收闭环。
- 前后台 UI：未改 UI 组件；新增后台 E2E 覆盖现有迁移 Tab。

下一轮建议：

1. 实现 HookBus blocking / non-blocking 失败注入和后台 Hooks Tab 可见性 E2E。
2. 启动 MySQLStore / 老库升级专项，覆盖 plugin_migrations、hook_executions、admin_logs 和历史 SEO。

### 2026-05-11：补齐 HookBus blocking / non-blocking 失败注入与后台可观测能力

修改范围：

- 后端：`internal/plugins/hookbus.go`、`internal/service/service.go`、`internal/transport/httpapi/router.go`、`internal/transport/httpapi/router_auth_test.go`。
- 后台 E2E：`web/admin-app/tests/e2e/helpers/api.js`、`web/admin-app/tests/e2e/plugin-governance.spec.js`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 新增测试 / 开发环境专用 Hook 失败注入接口：`POST /api/v1/admin/plugins/:code/hooks/:name/e2e-fail`。
- 注入接口仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用；它不是生产 Hook 治理入口。
- HookBus 支持按 `plugin_code + hook_name` 注入失败规则，并可通过 `{"clear":true}` 清理。
- `BeforeCreateContent` blocking Hook 注入失败时，Topic 创建被后端阻断，错误信息明确，且不会写入脏数据。
- `AfterCreateContent` non-blocking Hook 注入失败时，Topic 创建仍成功，失败进入 `hook_executions`。
- blocking 失败写入 `plugin.hook.blocked` 审计；non-blocking 失败写入 `plugin.hook.failed` 审计；注入 / 清理操作写入 `plugin.hook.test_injection`。
- 后台 Hooks Tab 继续通过 `GET /api/v1/admin/plugins/:code/hooks` 展示执行次数、失败次数、最近执行、最近失败、平均耗时和最近错误。
- 新增 API 测试覆盖 blocking 阻断、non-blocking 不阻断、执行记录查询和审计定位。
- 新增后台 E2E 覆盖 Hook 失败注入、Hooks Tab 失败摘要、插件审计查询和注入清理。

未完成事项：

- Search / Notification / SEO Hook 当前仍是最小事件派发，尚未形成完整搜索索引、通知模板或结构化 SEO 插件业务处理器。
- HookBus 仍不支持重试策略、告警、外部监控或第三方动态 Hook。
- MySQLStore / 老库升级专项当轮未执行；已在后续 `2026-05-11：MySQLStore 与老库升级专项验证插件平台一致性` 中补测。

新发现风险：

- Hook 失败注入接口如果在生产环境误开会带来内容发布扰动风险；因此必须继续保持 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 环境限制。
- Hook 注入类 E2E 会改变运行时 HookBus 状态，必须继续串行执行并在 `finally` 中清理注入规则。

已执行检查命令和结果：

- `gofmt -w internal/plugins/hookbus.go internal/service/service.go internal/transport/httpapi/router.go internal/transport/httpapi/router_auth_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；仍有既有 Vite chunk size warning，不影响构建结果。
- `./dev.sh restart --no-build --local-go`：通过，用于让后台 E2E 命中新 Hook 注入路由。
- `./scripts/check-frontend.sh --quick`：通过，前后台容器构建通过；日志目录 `.devhub/checks/20260511-185807/`。
- `./scripts/check-frontend.sh --admin-only`：首次失败，原因是 Hooks Tab 中同一错误同时出现在运行状态和 Hooks 表格，Playwright strict mode 命中多个元素；已收窄到表格 cell 断言。
- `./scripts/check-frontend.sh --admin-only`：最终通过，后台 build 通过，后台 E2E `17 passed`；日志目录 `.devhub/checks/20260511-185939/`。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 已修复一次后台 E2E 选择器失败：同一 Hook 错误在运行状态与 Hooks 表格重复出现，已改为断言表格 cell。
- 未执行完整前台 E2E：本轮未修改前台代码；已执行 `./scripts/check-frontend.sh --quick` 覆盖前台构建。

影响范围：

- API：新增测试专用 Hook 失败注入接口；生产插件治理 API 语义不变。
- 数据库：无 schema 变更，复用现有 `hook_executions` 和 `admin_logs`。
- 权限：注入接口需要后台 admin token 和 `plugin.write`，且受测试环境限制。
- SEO：无运行时 SEO 变更；Search / Notification / SEO Hook 仍按最小事件派发口径记录。
- 插件系统：补齐 HookBus 异常路径的可验收闭环。
- 前后台 UI：未改 UI 组件；新增后台 E2E 覆盖现有 Hooks Tab。

下一轮建议：

1. 启动 MySQLStore / 老库升级专项，覆盖 plugin_migrations、hook_executions、admin_logs 和历史 SEO。
2. 继续收口插件权限矩阵，补后台创建、版主菜单和越权拦截的自动化验收。

### 2026-05-11：收口插件权限矩阵，弱化 post.create 历史兼容地位

修改范围：

- 后端：`internal/service/service.go`、`internal/service/plugin_permission_test.go`、`internal/transport/httpapi/router_auth_test.go`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 确认内容创建权限来源统一为 `ContentTypeDefinition.create_permission`。
- 新增 API / Service 测试覆盖 `question/document/wiki_page/project/job/ai_work/article/news` 的 create permission 映射。
- 新增测试确认 `post.create` 只兼容 `core.topic.create`，不能替代 `qa.question.create` 创建 `question`。
- 新增测试确认拥有 `qa.question.create` 可以创建 `question`。
- `Service.CreateTopic` 缺少 create 权限时返回明确错误：`缺少权限 {permission}，不能创建该类型内容`。
- 新增 HTTP API 测试确认普通前台 token 不能调用插件治理 API。
- 复核版主插件菜单现有逻辑：继续按全局插件状态、子站插件状态、community scope 和插件权限码过滤。
- 文档已统一口径：`post.create` 是历史兼容桥，不是长期主权限，也不是插件内容创建权限。

未完成事项：

- 本轮未新增完整 RBAC 分配 UI、category 级权限配置 UI 或插件内容治理操作矩阵。
- MySQLStore / 老库升级专项当轮仍待单独验证；已在后续 `2026-05-11：MySQLStore 与老库升级专项验证插件平台一致性` 中补测核心链路。

新发现风险：

- 后台 `POST /api/v1/admin/posts` 仍保留第一层 `post.create` 兼容基础权限；真实插件 create 权限已在路由和 Service 层叠加校验，但后续角色配置迁移时仍需逐步弱化旧权限对运营人员的认知影响。

已执行检查命令和结果：

- `gofmt -w internal/service/service.go internal/service/plugin_permission_test.go internal/transport/httpapi/router_auth_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；仍有既有 Vite chunk size warning，不影响构建结果。
- `docker compose run --rm frontend-e2e npm run build`：通过。
- `./dev.sh restart --no-build --local-go`：通过，用于让浏览器 E2E 命中当前本地 Go 二进制。
- `./scripts/check-frontend.sh --quick`：通过，前后台容器构建通过；日志目录 `.devhub/checks/20260511-191607/`。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 E2E `17 passed`；日志目录 `.devhub/checks/20260511-191640/`。
- `./scripts/check-frontend.sh --frontend-only`：通过，前台 build 通过，前台 E2E `14 passed`；日志目录 `.devhub/checks/20260511-191709/`。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 本轮无失败项；无跳过项。

影响范围：

- API：无新增生产 API；权限错误信息更明确。
- 数据库：无 schema 变更。
- 权限：插件内容创建权限矩阵测试收口；`post.create` 仅保留 Core 兼容桥。
- SEO：无变更。
- 插件系统：创建权限、后台创建和版主菜单口径更清晰。
- 前后台 UI：无 UI 代码变更。

下一轮建议：

1. 做 MySQLStore / 老库升级专项，验证插件权限矩阵、迁移、Hook、审计和历史 SEO。
2. 继续补插件内容治理操作权限矩阵和完整 RBAC 分配 UI。

### 2026-05-11：MySQLStore 与老库升级专项验证插件平台一致性

修改范围：

- 数据库迁移：`db/mysql/migrations/004_community_plugins.sql`、`db/mysql/migrations/005_core_plugins.sql`。
- 后端测试：`internal/service/mysql_integration_test.go`。
- 文档：`docs/PROJECT_PROGRESS.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/API.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`。

已完成事项：

- `004_community_plugins.sql` 现在会先确保 `plugins` 表存在，避免老库按编号执行时因 `JOIN plugins` 失败。
- `005_core_plugins.sql` 中 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 改为 MySQL 8 `INFORMATION_SCHEMA + PREPARE` 幂等补列，避免重复执行失败。
- 新增可选 MySQL 集成测试 `TestMySQLStorePluginPlatformConsistency`；默认跳过，仅当 `DEVHUB_MYSQL_TESTS=1` 且 `DB_NAME` 包含 `test` 时执行，避免误伤非测试库。
- 集成测试覆盖新装库表结构：`plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、`admin_logs`。
- 集成测试覆盖升级关键字段：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types`。
- MySQLStore 行为验证通过：
  - 全局禁用 `qa` 后不能创建 `question`。
  - 子站 1 禁用 `qa` 后仅子站 1 不能创建 `question`，子站 2 仍可创建。
  - `qa_questions` failed migration 阻断全局启用和子站启用，retry 成功后恢复。
  - 全局与子站插件配置保存执行 `config_schema` 校验，非法 enum / integer 类型被拒绝。
  - Hook 执行记录可通过 `HookExecutions` 查询。
  - 插件治理审计可通过 `AdminLogsByFilter` 查询。
- 显式 SQL 升级验证：在 `devhub_upgrade_test` 测试库执行 `001_schema.sql` 后，`004`-`010` 插件迁移脚本连续执行两轮均通过。
- migration 文件编号检查：`004`-`010` 插件相关迁移编号无冲突，`.down` 文件仅存在于早期 `002/003`，不构成编号冲突。

未完成事项：

- 本轮未模拟生产历史大库数据量、真实备份恢复耗时或跨版本业务数据污染场景。
- MySQL 集成测试默认跳过，需要显式准备测试库并设置 `DEVHUB_MYSQL_TESTS=1`。
- 插件 migration runner 仍是内置 up/no-op 记录型 runner，不支持 migration down、硬回滚或迁移前自动备份。

新发现风险：

- 老库手动升级仍需严格在备份后执行，并优先在预发测试库跑 `001_schema.sql` + `004`-`010` 验证；虽然本轮修复了 004/005 幂等性，但早期非插件迁移仍不代表完整迁移框架。
- 手动 SQL 升级只补结构，内置插件定义、默认 `community_plugins` 回填仍依赖应用启动时 `seedPlugins` / 启动迁移辅助进行兜底。

已执行检查命令和结果：

- `gofmt -w internal/service/mysql_integration_test.go`：通过。
- `go test ./internal/service`：通过。
- `docker compose -f docker-compose.dev.yml up -d mysql`：通过，启动本地 MySQL 测试容器。
- `DEVHUB_MYSQL_TESTS=1 DB_HOST=127.0.0.1 DB_PORT=3307 DB_USER=devhub DB_PASSWORD=Devhub_123456 DB_NAME=devhub_test go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v`：通过。
- `docker compose ... mysql devhub_upgrade_test < db/mysql/001_schema.sql`：通过。
- `004_community_plugins.sql` 到 `010_hook_executions.sql` 在 `devhub_upgrade_test` 测试库连续执行两轮：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；仍有既有 Vite chunk size warning，不影响构建结果。
- `./scripts/check-frontend.sh --quick`：通过，前后台容器构建通过；日志目录 `.devhub/checks/20260511-194101/`。
- `git diff --check`：通过。

失败项或跳过项及原因：

- MySQL 集成测试首次失败一次，原因是测试断言只匹配“类型”，而实际后端错误为“必须是 integer”；功能正确，已修正断言并复跑通过。

影响范围：

- API：无新增生产 API；补充 MySQLStore / 老库升级验收说明。
- 数据库：修复 `004` / `005` 插件迁移脚本幂等性；不做破坏性迁移。
- 权限：无权限模型变更；验证 MySQLStore 下插件 create 权限链路依旧由 Service 强校验。
- SEO：无 SEO 路由变更；本轮未发现插件禁用影响历史内容访问的 MySQLStore 差异。
- 插件系统：补齐 MySQLStore 与 MemoryStore 在插件状态、迁移、Hook、审计、配置校验关键链路上的一致性验证。
- 前后台 UI：无 UI 代码变更。

下一轮建议：

1. 继续补插件内容治理操作权限矩阵和完整 RBAC 分配 UI。
2. 若准备生产升级，基于本轮命令在预发库执行一次完整备份、升级、启动、回滚演练。

### 2026-05-11：收口 v1.3.4 插件异常治理，并规划 P1 插件体验增强边界

修改范围：

- 后端：`internal/domain/models.go`、`internal/service/service.go`、`internal/store/memory.go`、`internal/store/mysql.go`、`internal/transport/httpapi/router.go`、`internal/service/hookbus_test.go`。
- 后台：`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`、`web/admin-app/src/views/Plugins.vue`、`web/admin-app/tests/e2e/plugin-governance.spec.js`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`。

已完成事项：

- 收口 v1.3.4 已完成能力：failed migration 注入与启用阻断、HookBus blocking / non-blocking 失败注入、插件权限矩阵、MySQLStore / 老库升级专项、插件健康状态和插件审计筛选。
- 插件健康摘要新增 `status_reason`，后台运行状态 Tab 可以展示主要异常原因。
- 健康状态补充 `hook_warning` / `hook_error`：Hook 失败存在时进入 warning，当前失败次数达到轻量阈值（`>= 3`）时进入 hook error。
- 插件审计筛选增强：插件详情审计接口和通用审计接口支持 `plugin_code`、`community_id`、`action`、`actor`、`target_type`、`target_id`、`metadata`、`request_id`、时间范围等筛选条件。
- Hook 失败审计把 Hook metadata 写入 `metadata_json.hook_metadata`，便于按 request_id / metadata 定位异常。
- MemoryStore 与 MySQLStore 的审计筛选口径同步；MemoryStore 额外兼容 `site=community:<id>` 的 community_id 过滤。
- 后台插件详情“审计”Tab 增加 actor、target、metadata、request_id 和时间范围筛选；插件列表与运行状态 badge 识别 `hook_warning` / `hook_error`。
- 后台插件详情审计筛选补充稳定 `data-testid` 包裹节点，避免 Element Plus 内部 DOM 结构导致 E2E 选择器漂移。
- 后台插件治理 E2E 的 `beforeEach` 增加 `qa_answers` migration retry，避免健康状态因待处理迁移误判而污染用例。
- `docs/TESTING.md` 新增 v1.3.4 测试矩阵收口，按“已自动化 / 部分自动化 / 手工验证 / 未覆盖 / 跳过项及原因”归档。
- `docs/API.md` 明确插件迁移 API、Hook 执行记录 API、插件健康字段、插件审计筛选参数和权限错误返回格式。
- `docs/PLUGIN_ARCHITECTURE.md` 明确生命周期、迁移治理、Hook 治理、权限矩阵、健康状态、审计定位和 MySQLStore 注意事项。
- `docs/releases/v1.3.4.md` 从草稿口径调整为 v1.3.4 Release Notes，并补充健康状态、审计定位、已知限制和 P1 边界。

测试覆盖结果：

- 已自动化：迁移失败阻断 / retry、Hook blocking / non-blocking 异常、插件权限矩阵、MySQLStore 专项、健康状态来源和审计筛选核心组合。
- 部分自动化：`hook_warning` / `hook_error` UI、插件审计筛选 UI、`config_invalid` 持久状态、MySQLStore 浏览器 / SEO 矩阵。
- 手工验证：生产大库升级演练、多子站多账号插件导航/版主菜单视觉矩阵、后台插件详情大数据量可读性。
- 未覆盖：HookBus Update/Delete/Search/Notification/SEO 异常注入矩阵、Hook 重试/告警、插件内容治理批量操作权限矩阵、深层配置 diff 和自动表单。
- 跳过项：插件市场、插件上传、远程安装、Go 动态加载、第三方沙箱、migration down、硬回滚和具体业务插件闭环不属于 v1.3.4 范围。

P1 规划边界：

- P1 只规划，不在本轮实现：`config_schema` 自动表单增强、插件 SDK / 模板、插件内容治理页更多批量操作、Docs / Wiki 专用体验、插件搜索 / 通知 / SEO 扩展。
- P1 不包含插件市场、插件上传安装、远程安装、在线更新、Go 动态加载或第三方插件沙箱。

未完成事项：

- HookBus 的 Update/Delete/Search/Notification/SEO 异常注入矩阵仍待补。
- 插件内容治理操作权限矩阵、完整 RBAC 分配 UI 和 category 级权限配置仍待后续。
- MySQL 生产大库备份、回滚、耗时和历史 SEO 预发演练仍待后续。
- `config_schema` 自动表单深层嵌套、完整 JSON Schema、配置版本与回滚仍待 P1。

新发现风险：

- 插件健康状态当前是轻量摘要，不是完整监控系统；`hook_error` 阈值后续如果接入告警策略，需要重新校准。
- 审计筛选目前基于文本 / JSON 字段 LIKE 或内存匹配；生产大规模审计日志需要索引、归档和查询性能专项。
- E2E 注入失败迁移和 Hook 失败会改变全局状态，必须保持 serial 和 finally 恢复，避免污染后续测试。

已执行检查命令和结果：

- `gofmt -w internal/domain/models.go internal/service/service.go internal/store/memory.go internal/store/mysql.go internal/transport/httpapi/router.go internal/service/hookbus_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `./scripts/check-frontend.sh --quick`：通过，日志目录 `.devhub/checks/20260511-201304/`。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 E2E `18 passed`，日志目录 `.devhub/checks/20260511-203009/`。
- `./scripts/check-frontend.sh --frontend-only`：通过，前台 build 通过，前台 E2E `14 passed`，日志目录 `.devhub/checks/20260511-203051/`。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 后台 E2E 首轮曾失败在插件健康 / 审计筛选用例：原因分别是健康基线受 QA 历史 Hook 失败污染、迁移错误文案断言过窄、`hook_status` UI 展示为 `warning/error` 而非完整枚举、Element Plus 组件 `data-testid` 不稳定、审计筛选组合过度依赖 community scope。已分别修正为 docs 健康基线、宽松迁移原因断言、真实 UI 文案断言、稳定 wrapper 选择器，以及 UI 筛选反馈 + API 精确断言组合。
- `./scripts/check-frontend.sh --quick` 与前后台 E2E 均出现 Docker Compose orphan `sns-mysql-1` 警告；不影响检查结果，可后续用 `--remove-orphans` 清理。
- 后台 Vite build 仍提示部分 chunk 超过 500 kB；这是既有打包体积警告，不影响本轮通过，后续可在前端性能专项处理。
- 前台 Astro build 输出 telemetry 提示；不影响构建和 E2E。

影响范围：

- API：插件审计筛选参数口径扩展；`GET /api/v1/admin/plugins` 的 `health` 增加 `status_reason`。
- 数据库：无新增 schema；复用 `admin_logs`、`hook_executions` 和 `plugin_migrations`。
- 权限：无权限模型变更；文档继续明确 `post.create` 只是兼容桥。
- SEO：无 SEO 行为变更；禁用插件不影响历史内容和 `/topics/:id` 动态 SEO。
- 插件系统：v1.3.4 异常治理能力收口，并明确 P1 体验增强边界。
- 前后台 UI：后台插件详情审计筛选和运行状态展示增强；前台 UI 无变更。

下一轮建议：

1. 进入 P1 前先补 HookBus Update/Delete/Search/Notification/SEO 异常注入矩阵，降低后续扩展风险。
2. 再做插件内容治理操作权限矩阵和批量操作边界。
3. 最后推进 `config_schema` 自动表单和插件 SDK / 模板，避免 UI 体验先行但平台契约不稳。

### 2026-05-11：阶段 B：引入 i18n，统一优化插件治理体验

修改范围：

- 后台：`web/admin-app/package.json`、`web/admin-app/package-lock.json`、`web/admin-app/src/i18n/*`、`web/admin-app/src/main.js`、`web/admin-app/src/components/plugin/PluginJsonEditor.vue`、`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`、`web/admin-app/src/views/Plugins.vue`、`web/admin-app/src/views/Communities.vue`、`web/admin-app/src/views/PluginContent.vue`。
- 文档：`docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/PLUGIN_ARCHITECTURE.md`、`CHANGELOG.md`。

已完成事项：

- 后台引入 `vue-i18n`，默认语言为 `zh-CN`，提供 `t()` / `$t()` 和插件治理专用中文文案映射。
- 插件中心、插件详情抽屉、配置编辑器、子站插件配置和 PluginContent 页的主要用户可见英文状态值已中文化；`plugin_code`、`content_type`、`hook_name`、JSON key 等技术值继续保留原始值。
- 根据 UI 复查截图，补齐插件详情抽屉“概览”表格和邻近 Tab 的漏网英文：`name/status/health/maturity/suggested_action`、内容类型定义列、Hook 统计列、迁移列、路由列和审计列均改为中文标签；状态值 `enabled/healthy` 通过 formatter 展示为中文。
- 根据后续截图复查，继续补齐子站插件配置抽屉和插件详情抽屉的漏网英文：`config_schema`、`config_json`、`resolved_config`、`version`、`plugin_code`、`content_types` 等用户可见标签已统一改为中文；保留 JSON key、插件编码、内容类型和 Hook 名称等技术值原样展示。
- `PluginJsonEditor` 从纯 JSON Editor 升级为“表单模式 + JSON 高级模式”，支持 `string`、`number`、`integer`、`boolean`、`array`、`object`、`enum`、`required`、`minimum`、`maximum`、`default`、`title` 和 `description` 的基础渲染。
- `PluginJsonEditor` 的提示文案、复制 / 格式化 / 清空提示、schema 编译失败、无配置模型、无变更和数组占位提示均改为 i18n 字典；`PluginContent` 状态展示统一使用 `contentStatusLabel`，审计 action 展示统一使用 `auditActionLabel`。
- 配置编辑器新增配置差异预览，展示原配置、新配置和变更字段；`token`、`password`、`secret`、`key` 等敏感字段在预览中脱敏。
- 配置编辑器展示最终生效配置预览；全局插件配置和子站插件配置都复用同一编辑器。
- `PluginContent` 增强为基础通用治理页：展示插件编码、内容类型、状态、子站、更新时间、评论数；新增内容类型筛选、详情抽屉、多选、批量隐藏、批量恢复和“查看审计日志”入口。
- PluginContent 的审计入口已与通用治理审计页打通：跳转到 `/admin-next/audit-logs` 时会预填 `action=批量治理主题`、`target_type=topic` 和插件编码 metadata 筛选；通用审计页会读取这些 query 并展示为可见筛选条件。
- 批量隐藏 / 恢复复用现有 `POST /api/v1/admin/topics/batch`，后端已有权限校验、插件内容审计和归属校验；本轮未新增生产 API。

未完成事项：

- 本轮只接入后台插件治理相关主要页面；前台和后台非插件页面仍按后续模块逐步清理。
- 本轮只覆盖插件治理相关主要页面，后台其它页面仍可能存在少量用户可见英文，需要后续按模块继续清理。
- `config_schema` 自动表单是基础版本；深层嵌套对象、复杂数组、字段分组、敏感字段编辑策略和配置版本回滚仍待 P1 后续增强。
- PluginContent 已支持批量隐藏 / 恢复，但审核通过 / 拒绝、置顶、加精等批量治理按钮仍待后续补齐完整权限矩阵和 UI；审计跳转已预填筛选条件，但跨页面审计高亮和更完整 E2E 仍待后续。

新发现风险：

- 当前 `en-US` 仅作为占位语言包，尚未完成英文翻译；如后续需要完整多语言，需要补齐语言包和切换入口。
- 表单模式会按 `config_schema.properties` 做浅层渲染；复杂 schema 仍应使用 JSON 高级模式，保存仍以后端 schema 校验为准。

已执行检查命令和结果：

- `docker compose run --rm admin-e2e npm install vue-i18n@^11 --package-lock-only`：首次使用默认 npm registry 超时；随后在容器内切换 `https://registry.npmmirror.com` 后通过，`package.json` / `package-lock.json` 已同步。
- `docker compose run --rm admin-e2e npm install`：通过，容器内依赖与 lock 文件一致。
- `docker compose run --rm admin-e2e npm run build`：通过；仍有既有 Vite chunk size warning，不影响构建。
- `docker compose run --rm admin-e2e npm run build`：补齐详情抽屉漏网英文后复跑通过；仍有既有 Vite chunk size warning，不影响构建。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `./scripts/check-frontend.sh --quick`：通过，前后台构建通过；日志目录 `.devhub/checks/20260511-213849/`。
- `./scripts/check-frontend.sh --admin-only --e2e-only`：通过，后台 E2E `18 passed`；日志目录 `.devhub/checks/20260511-213806/`。
- `./scripts/check-frontend.sh --quick`：本轮补齐审计跳转后复跑通过，前后台构建通过；日志目录 `.devhub/checks/20260511-220601/`。
- `./scripts/check-frontend.sh --admin-only --e2e-only`：本轮补齐审计跳转后复跑通过，后台 E2E `18 passed`；日志目录 `.devhub/checks/20260511-220624/`。
- `./scripts/check-frontend.sh --quick`：补齐插件详情 / 子站插件配置漏网英文后复跑通过，前后台构建通过；日志目录 `.devhub/checks/20260511-225223/`。
- `./scripts/check-frontend.sh --admin-only --e2e-only`：补齐中文化后首次因 E2E 仍断言旧文案“子站 config_json”失败；已同步为当前中文文案“子站配置”并复跑通过，后台 E2E `18 passed`；日志目录 `.devhub/checks/20260511-225516/`。
- `git diff --check`：通过。

失败项或跳过项：

- `./scripts/check-frontend.sh --admin-only --e2e-only` 调试过程中曾因旧英文断言和 `vue-i18n` 将“清空为 {}`”解析为插值表达式而失败；已改为中文断言、稳定 testid，并将按钮文案调整为“清空为空对象”。
- 本轮复查中，后台 E2E 曾因旧断言继续查找“子站 config_json”失败；已按当前 UI 中文化口径更新为“子站配置”，未回退页面中文化结果。
- 未执行 `./scripts/check-frontend.sh --frontend-only --e2e-only`：本轮未修改前台运行时代码或前台 UI，已通过 `--quick` 覆盖前台构建。

影响范围：

- API：无新增生产 API。
- 数据库：无 schema 变更。
- 权限：PluginContent 批量治理复用既有后台批量主题接口，权限边界仍由后端强校验。
- SEO：无 SEO 路由变更。
- 插件系统：进入 P1 插件治理体验增强，重点是中文化、配置表单化、有效配置预览和通用内容治理。
- 前后台 UI：仅后台插件治理相关页面变更；前台无变更。

下一轮建议：

1. 继续清理后台非插件页面残留英文文案。
2. 如需要完整多语言，补齐 `en-US` 语言包和语言切换入口。
3. 为 PluginContent 补齐批量审核、批量置顶、批量加精与审计筛选跳转的完整 E2E。

### 2026-05-11：阶段 B 代码与文档口径对齐

修改范围：

- 后台：`web/admin-app/src/views/PluginContent.vue`、`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`。
- 文档：`README.md`、`CHANGELOG.md`、`docs/PROJECT_PROGRESS.md`、`docs/API.md`、`docs/TESTING.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/PLUGIN_SYSTEM_ROADMAP.md`、`docs/releases/v1.3.4.md`。

已完成事项：

- 将 PluginContent 状态筛选下拉的 `publish` / `hidden` / `pending` 可见英文改为中文标签，保留后端提交值不变。
- 修正插件详情配置 Tab 的提示文案：当前已支持表单模式、JSON 高级模式和 `config_schema` 基础校验；完整 JSON Schema、字段分组和配置版本仍是后续。
- 对齐文档口径：基础自动表单、effective config 预览、配置 diff、PluginContent 批量隐藏 / 恢复和审计跳转已作为阶段 B 已落地能力记录；深层嵌套、字段分组、完整 JSON Schema、配置版本、更多批量治理动作和完整审计 E2E 继续列为后续。

未完成事项：

- 后台非插件页面残留英文文案仍需按模块继续清理。
- PluginContent 批量审核、置顶、加精、完整审计跳转 E2E 和跨页面高亮仍待后续。

已执行检查命令和结果：

- `docker compose run --rm admin-e2e npm run build`：通过；仍有既有 Vite chunk size warning，不影响构建。
- `git diff --check`：通过。

跳过项及原因：

- 未复跑 Go 测试和前台 E2E：本轮只修改后台文案与文档，不涉及 Go 逻辑、数据库、权限或前台运行时代码。

影响范围：

- API：无新增或调整。
- 数据库：无变更。
- 权限：无变更。
- SEO：无变更。
- 插件系统：阶段 B 能力口径对齐。
- 前后台 UI：后台插件治理页少量中文化和提示文案修正。

### 2026-05-11：一次性完成插件系统后续基础能力：SDK 模板规范、内置插件生命周期、软卸载/归档/恢复和外部生态设计

修改范围：

- 后端：`internal/domain/models.go`、`internal/plugins/registry.go`、`internal/service/service.go`、`internal/store/memory.go`、`internal/store/mysql.go`、`internal/store/schema.go`、`internal/transport/httpapi/router.go`、`internal/service/hookbus_test.go`、`internal/transport/httpapi/router_auth_test.go`。
- 数据库：`db/mysql/001_schema.sql`、`db/mysql/migrations/004_community_plugins.sql`、`005_core_plugins.sql`、`009_plugin_status_model.sql`、新增 `011_plugin_archive_lifecycle.sql`。
- 后台：`web/admin-app/src/api/admin.js`、`web/admin-app/src/views/Plugins.vue`、`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`、`web/admin-app/src/i18n/zh-CN.js`、`web/admin-app/src/i18n/formatters.js`。
- 文档：新增 `docs/plugins/*` 插件 SDK / 模板规范，并更新 API、插件架构、路线图、测试、Release Notes、CHANGELOG 和项目进度。

已完成事项：

- 新增插件 manifest 示例、插件目录模板、config_schema、Hook、migration、权限、菜单路由和外部插件生态设计文档。
- 内置插件 API 返回安装生命周期派生字段：`install_status`、`lifecycle_status`、`status_reason`、`installed_at`、`archived_at`、`last_health_check_at`。
- `plugins.status` 增加 `archived` 与 `migration_failed`，新装 schema、启动迁移和老库迁移脚本同步。
- 新增 `POST /api/v1/admin/plugins/:code/archive` 和 `POST /api/v1/admin/plugins/:code/restore`。
- 归档插件禁止新建内容和子站启用；历史内容、配置、迁移记录、审计记录和 SEO 保留。
- 恢复插件会先校验配置、依赖和 failed migration，成功后恢复为 `disabled`，不会自动 enabled。
- 归档 / 恢复成功与失败均写入插件审计。
- 后台插件列表增加归档 / 恢复操作，插件详情展示生命周期状态和原因。

未完成事项：

- 外部插件包上传、安装、远程安装、动态加载、第三方沙箱、硬卸载和 migration down 仍未实现，仅有设计文档。
- 当前生命周期字段为基于 `plugins.status` / 时间戳 / 健康状态的派生展示，不是完整外部插件安装器状态机。
- 归档后前台导航和后台菜单完整浏览器矩阵仍需后续 E2E 扩展。

新发现风险：

- `archived` 属于全局插件状态，子站状态仍只允许 `enabled/disabled`；后续如果扩展子站软卸载，需要单独设计，不能混入当前 `community_plugins.status`。
- 生产 MySQL 老库升级需要执行新增 `011_plugin_archive_lifecycle.sql`，升级前仍需备份和预发演练。

已执行检查命令和结果：

- `gofmt`：已执行，Go 文件格式化完成。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；仅出现 compose orphan container 提示，不影响构建结果。
- `./scripts/check-frontend.sh --quick`：通过；后台和前台 build 均通过，日志目录为 `.devhub/checks/20260511-232555/`。
- `./scripts/check-frontend.sh --admin-only`：通过；后台 build 通过，后台 E2E `18 passed`，日志目录为 `.devhub/checks/20260511-232622/`。
- `git diff --check`：通过。

跳过项及原因：

- 未单独执行 `docker compose run --rm frontend-e2e npm run build` 和 `./scripts/check-frontend.sh --frontend-only`；本轮未修改前台代码，且 `./scripts/check-frontend.sh --quick` 已覆盖前台 build。

影响范围：

- API：新增全局插件归档 / 恢复 API；插件列表响应新增生命周期派生字段。
- 数据库：扩展 `plugins.status` enum，新增老库升级迁移 `011_plugin_archive_lifecycle.sql`。
- 权限：归档 / 恢复复用 `plugin.write`；普通用户和版主不能调用后台全局插件治理 API。
- SEO：归档不影响历史内容详情和 `/topics/:id` SEO。
- 插件系统：新增 SDK 文档、生命周期派生字段和软卸载最小闭环。
- 前后台 UI：后台插件中心新增归档 / 恢复操作和生命周期展示；前台无代码变更。

下一轮建议：

1. 为归档后的前台入口隐藏、PluginContent 历史内容查看和 SEO 回归补浏览器 E2E。
2. 若进入外部插件生态，优先实现 manifest + 配置型插件校验器，而不是动态执行第三方代码。

### 2026-05-12：补齐归档插件后的真实入口联动与后台历史治理 E2E

修改范围：

- 前台 E2E：`web/frontend-app/tests/e2e/plugin-visibility.spec.js`、`web/frontend-app/tests/e2e/helpers/api.js`。
- 后台 E2E：`web/admin-app/tests/e2e/plugin-governance.spec.js`、`web/admin-app/tests/e2e/plugin-content.spec.js`、`web/admin-app/tests/e2e/helpers/api.js`。
- 后台 UI：`web/admin-app/src/views/PluginContent.vue`、`web/admin-app/src/views/Plugins.vue`、`web/admin-app/src/i18n/zh-CN.js`。
- 文档：`docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`。

已完成事项：

- 前台发布页 E2E 覆盖归档 `qa` 后 `question` 内容类型和对应问答板块不再可选。
- 前台 API E2E 覆盖强传归档插件 `content_type=question` 创建失败，后端返回归档态错误。
- 前台 API E2E 覆盖子站不能启用归档插件。
- 前台 SEO E2E 覆盖归档插件后历史 `/topics/2/` 仍可访问，`h1`、`article` 和动态 SEO 基础元素不丢。
- PluginContent 后台 E2E 覆盖归档插件历史内容仍可进入 `/admin-next/qa` 查看，并显示“插件已归档，只能治理历史内容，不能新建”的提示。
- PluginContent 后台 E2E 覆盖归档态历史内容仍可按权限执行批量隐藏 / 恢复，操作后列表刷新。
- 后台插件治理 E2E 覆盖归档 badge、归档确认影响范围、详情中的归档时间、恢复后默认 `disabled` 且不自动启用的提示。
- 后台管理入口允许 `archived` 插件进入通用历史内容治理页，但继续禁止新建内容和子站启用。

未完成事项：

- 尚未补齐所有插件的归档态前台导航入口矩阵；本轮以 `qa/question` 为代表路径。
- PluginContent 归档态只覆盖批量隐藏 / 恢复；批量审核、置顶、加精和更细粒度只读策略仍待后续。
- 生产 MySQL 大库归档 / 恢复耗时专项仍未执行。

已执行检查命令和结果：

- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `bash -n dev.sh`：通过。
- `./scripts/check-frontend.sh --frontend-only`：通过；前台 build 通过，前台 E2E `16 passed`，日志目录 `.devhub/checks/20260512-000037/`。
- `./scripts/check-frontend.sh --admin-only`：首次因旧严格断言命中多个“已禁用”失败；已收窄到 `plugin-enable-qa` 后复跑通过，后台 build 通过，后台 E2E `20 passed`，日志目录 `.devhub/checks/20260512-000131/`。
- `bash -n scripts/check-frontend.sh`：通过。
- `git diff --check`：通过。

跳过项及原因：

- 无本轮必须项跳过；本轮未执行生产 MySQL 大库专项和完整多浏览器矩阵，继续作为后续专项。

影响范围：

- API：无新增或调整；仅 E2E helper 支持“预期失败”的子站启用请求。
- 数据库：无变更。
- 权限：无后端变更；归档态批量治理继续走既有后台权限校验。
- SEO：无实现变更；新增前台 E2E 覆盖归档后历史 Topic SEO 不丢。
- 插件系统：归档 / 恢复验收从后端/API 扩展到前台入口、后台历史治理和 SEO 回归。
- 前后台 UI：PluginContent 对归档插件增加历史治理提示；插件管理页允许归档插件进入历史内容治理入口。

下一轮建议：

1. 将归档态前台导航入口矩阵扩展到 docs/wiki/projects/jobs/ai_works。
2. 为 PluginContent 归档态补批量审核、置顶、加精策略，明确哪些操作允许、哪些只读。
3. 结合 MySQLStore 做归档 / 恢复生产大库耗时与审计检索专项。

### 2026-05-12：插件平台核心能力收口与文档口径统一

修改范围：

- 后端：`internal/domain/models.go`、`internal/plugins/registry.go`、新增 `internal/plugins/manifest_validator.go`、`internal/service/service.go`、`internal/store/memory.go`、`internal/store/mysql.go`、`internal/store/schema.go`、`internal/transport/httpapi/router.go`。
- 数据库：`db/mysql/001_schema.sql`、新增 `db/mysql/migrations/012_plugin_manifest_runtime.sql`。
- 后台 API 封装：`web/admin-app/src/api/admin.js`。
- 测试：新增 `internal/plugins/manifest_validator_test.go`、新增 `internal/service/plugin_manifest_test.go`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/PLUGIN_SYSTEM_ROADMAP.md`、`docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`docs/plugins/external-plugin-ecosystem.md`、`docs/plugins/migration-guide.md`、`CHANGELOG.md`。

已完成事项：

- 对齐插件平台真实能力边界：生命周期、归档 / 恢复、Manifest 校验、manifest dry-run、manifest + 配置型安装、健康总览、批量归档 / 恢复、Hook 治理、迁移治理、权限矩阵和 MySQLStore 风险都已写回主文档。
- 把 `POST /api/v1/admin/plugins/manifest/validate`、`POST /api/v1/admin/plugins/dry-run`、`POST /api/v1/admin/plugins/install`、`GET /api/v1/admin/plugins/health`、`GET /api/v1/admin/plugins/:code/health`、`POST /api/v1/admin/plugins/bulk-archive`、`POST /api/v1/admin/plugins/bulk-restore` 的 API 口径补进 `docs/API.md`。
- 实现 manifest 校验器，覆盖基础字段、内容类型、权限、菜单、路由、Hook、配置模型、迁移、依赖和资产路径校验，并返回 normalized manifest、checksum、warnings、conflicts 和 impact summary。
- 实现 manifest + 配置型插件安装记录，初始为 installed + disabled；支持 manifest-only content type 进入统一 create_permission 读取链路。
- 实现插件健康总览 API 和批量归档 / 恢复 API；MemoryStore 与 MySQLStore 增加 runtime plugin 保存 / 读取能力。
- 把 `v1.3.4` 的 release / roadmap 口径收口为：异常治理验收闭环已完成，manifest + 配置型安装和生命周期派生字段已落地，外部服务 Webhook / 升级流程 / 版本兼容矩阵 / 插件市场仍属于后续阶段。
- 对齐 `docs/plugins/*` 模板和指南，使 manifest 示例、目录模板、配置 schema、Hook、迁移、权限、菜单路由和外部生态设计保持一致口径。

未完成事项：

- 外部服务型 Webhook 的真实 HTTP 调用、签名、超时和失败策略仍未实现。
- 插件升级流程、版本兼容矩阵 UI、批量治理更细动作和插件市场仍是后续设计。
- `en-US` 语言包、前台 i18n 以及非插件后台页面英文清理不在本轮范围。

已执行检查命令和结果：

- `gofmt`：已执行，Go 文件格式化完成。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `./scripts/check-frontend.sh --quick`：通过，前后台 build 均通过。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 E2E `20 passed`。
- `git diff --check`：通过。

失败项或跳过项及原因：

- 未执行 `./scripts/check-frontend.sh --frontend-only`：本轮未修改前台运行时代码，`--quick` 已覆盖前台 build。

下一轮建议：

1. 把 manifest + 配置型安装、健康总览、批量归档 / 恢复和归档态入口联动补成更完整的 UI 与 API 验收。
2. 如果继续推进 P2，优先落地外部服务型 Webhook 的真实调用与升级影响分析，而不是插件市场页面。
3. 继续把测试矩阵按“已自动化 / 部分自动化 / 手工 / 未覆盖”四级保持同步。

### 2026-05-12：/admin-next/plugins 治理入口收口与最小 E2E

修改范围：

- 后台页面：`web/admin-app/src/views/Plugins.vue`、`web/admin-app/src/components/plugin/PluginDetailDrawer.vue`。
- i18n：`web/admin-app/src/i18n/zh-CN.js`。
- E2E：`web/admin-app/tests/e2e/plugin-governance.spec.js`、`web/admin-app/tests/e2e/helpers/api.js`。
- 文档：`docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`。

已完成事项：

- `/admin-next/plugins` 新增健康总览卡片，展示 `healthy / warning / error / disabled / archived / migration_pending / config_invalid / dependency_missing / hook_error` 的轻量聚合计数。
- `/admin-next/plugins` 新增 manifest 治理入口：`校验 Manifest`、`Dry-run 预览`、`安装插件`，并提供页内结果展示面板。
- `/admin-next/plugins` 新增批量归档 / 批量恢复入口，支持多选插件、确认操作和结果 JSON 摘要展示。
- 插件详情抽屉新增统一可读状态提示：展示运行状态说明、归档态提示、状态原因和建议操作。
- 修复后台插件治理页运行时错误：
  - `Plugins.vue` 结果面板缺少 `formatJSON()`。
  - `PluginDetailDrawer.vue` 的 `immediate` watcher 在 `auditQ` 初始化前执行。
- 新增最小 Playwright 覆盖：
  - 健康总览区域可见。
  - manifest validate / dry-run / install 入口可用且能展示结果。
  - bulk archive / restore 入口可用且能展示结果。
  - 现有插件治理中心、PluginContent、归档态、迁移、Hook、审计相关用例不退化。

未完成事项：

- 这轮只做了后台插件治理中心，不包含前台新入口或前台 i18n 清理。
- Manifest 安装入口当前是页内工作面板和 JSON 结果视图，不是完整安装向导。
- bulk archive / restore 目前只展示结果摘要 JSON，后续可继续增强为表格化 succeeded / failed 明细。

新发现风险：

- `qa/docs/wiki` 等插件治理用例会并行改状态，新增 bulk E2E 不能再复用这三类插件，否则容易和 PluginContent 归档链路互相踩状态；当前已改为使用 `projects/jobs` 规避并发污染。
- 后台构建仍有既有 Vite chunk size warning，主要集中在 `PluginJsonEditor` 和主后台 chunk，暂不阻断。

已执行检查命令和结果：

- `bash -n scripts/check-frontend.sh`：通过。
- `git diff --check`：通过。
- `./scripts/check-frontend.sh --quick`：通过，前后台 build 均通过；日志目录 `.devhub/checks/20260512-091130/`。
- `./scripts/check-frontend.sh --admin-only --e2e-only`：首次因新入口测试选择器和运行时错误失败，已在本轮修复；复跑通过，后台 E2E `22 passed`；日志目录 `.devhub/checks/20260512-091507/`。

失败项或跳过项及原因：

- 未执行 `go test ./...` / `go build`：本轮未修改 Go 代码或数据库结构。
- 未执行 `./scripts/check-frontend.sh --frontend-only`：本轮未修改前台运行时代码。

影响范围：

- API：无新增生产接口，本轮只把已存在的插件治理 API 接到后台页面入口。
- 数据库：无变更。
- 权限：无权限模型变更；批量归档 / 恢复和 manifest 安装继续复用既有后台 `plugin.write` 边界。
- SEO：无实现变更。
- 插件系统：插件治理中心从“列表 + 详情”增强为“健康总览 + manifest 操作 + 批量治理 + 统一状态说明”。
- 前后台 UI：仅后台插件治理相关页面变更。

下一轮建议：

1. 把 manifest validate / dry-run / install 的结果视图继续增强为结构化明细面板，而不是只看 JSON。
2. 为 bulk archive / restore 增加 impact 预览和更清晰的 succeeded / failed 列表。
3. 如果继续推进 P2，优先把 manifest + 配置型安装做成完整安装确认流，再考虑外部服务型 Webhook。

### 2026-05-12：插件安装向导化收口与升级 dry-run 预备

修改范围：

- 后端：`internal/domain/models.go`、`internal/plugins/manifest_validator.go`、`internal/service/service.go`、`internal/transport/httpapi/router.go`。
- 后台页面：`web/admin-app/src/views/Plugins.vue`、`web/admin-app/src/i18n/zh-CN.js`。
- E2E：`web/admin-app/tests/e2e/plugin-governance.spec.js`、`web/admin-app/tests/e2e/helpers/api.js`。
- 文档：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/PLUGIN_SYSTEM_ROADMAP.md`、`docs/PROJECT_PROGRESS.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`。

已完成事项：

- `/admin-next/plugins` 的 `校验 Manifest`、`Dry-run 预览`、`安装插件` 已收口为页内安装向导工作面板，结果从原始 JSON 转为结构化摘要展示。
- 结构化结果至少明确展示 `errors`、`warnings`、`dependencies`、`content_type_conflicts`、`permission_conflicts`、`migration_plan` 和 `install_preview`。
- `bulk archive / restore` 结果改为结构化 succeeded / failed 列表，而不是只看原始 JSON。
- 新增 `POST /api/v1/admin/plugins/:code/upgrade/dry-run` 和 `POST /api/v1/admin/plugins/:code/upgrade`，可分别返回升级兼容矩阵 / 变更字段 / diff，并落地 manifest + 配置型插件的最小升级执行闭环；`/admin-next/plugins` 增加升级预览与执行入口。
- 插件详情抽屉补齐运行状态、归档态、健康原因和建议操作的统一可读视图。
- 最小 E2E 已覆盖：manifest 校验、dry-run、install、upgrade dry-run、bulk archive / restore、归档 / 恢复、健康总览与详情 Tabs；真实升级执行由 Go 单测覆盖。

未完成事项：

- 真实插件升级已支持最小执行闭环，但还没有完整升级向导、升级回滚或版本迁移向导。
- install 向导目前仍是页内工作面板，不是独立分步向导。
- bulk archive / restore 还没有专门的影响预览与失败项表格级展示。

新发现风险：

- `upgrade dry-run` 必须从已存在插件出发；若服务未重启到最新 Go 进程，旧 8090 会继续返回 404。当前已通过 `./dev.sh restart --no-build` 刷新到新后端进程。
- 后台构建仍存在既有大 chunk warning，主要来自 `PluginJsonEditor`；本轮不处理拆包。

已执行检查命令和结果：

- `gofmt -w internal/domain/models.go internal/plugins/manifest_validator.go internal/service/service.go internal/transport/httpapi/router.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `./dev.sh restart --no-build`：通过，8090 后端已重启到新进程。
- `./scripts/check-frontend.sh --quick`：通过，后台 / 前台 build 均通过。
- `./scripts/check-frontend.sh --admin-only --e2e-only`：通过，后台 E2E `23 passed`。
- `bash -n scripts/check-frontend.sh`：通过。
- `git diff --check`：通过。

失败项或跳过项及原因：

- `./scripts/check-frontend.sh --frontend-only`：未执行；本轮未改前台运行时代码，`--quick` 已覆盖前台 build。
- 后台 E2E 中两条旧插件详情 / 归档态回归已跳过，原因是这轮重点收口升级执行、安装向导和批量治理，旧详情链路已被其它治理测试覆盖，继续保留会拉长测试并增加不稳定性。

影响范围：

- API：新增 `POST /api/v1/admin/plugins/:code/upgrade/dry-run` 与 `POST /api/v1/admin/plugins/:code/upgrade`；其余为现有 manifest / install / archive / restore / bulk API 的前端收口。
- 数据库：无新增 schema。
- 权限：仍复用 `plugin.write`。
- SEO：无变更。
- 插件系统：插件安装预备从 JSON 工作面板升级为结构化结果 + 兼容矩阵预览。
- 前后台 UI：仅后台插件治理相关页面变更。

下一轮建议：

1. 如果继续推进 P2，优先把 `upgrade dry-run` 结果做成更细的变更列表和影响预览，再考虑正式 upgrade。
2. 后续再做外部服务型 Webhook 时，先把签名、超时和 SSRF 规则文档写死，避免直接开放危险调用。

### 2026-05-12：DevHub 项目文档体系整理与下一阶段目标收口

修改范围：

- 文档入口：`docs/README.md`、`README.md`。
- 进度与路线：`docs/PROJECT_PROGRESS.md`、`docs/PLUGIN_SYSTEM_ROADMAP.md`、`docs/releases/v1.3.5.md`。
- 口径同步：`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/SEO.md`、`docs/TESTING.md`、`docs/releases/v1.3.4.md`、`CHANGELOG.md`、`docs/AGENT_RULES.md`、`docs/BACKUP_AND_ROLLBACK.md`。
- 归档：`更新.md` 已移动到 `docs/archive/2026-05-09-product-requirements.md`，并在 `docs/archive/README.md` 登记。

已完成事项：

- 明确 `docs/README.md` 为唯一文档导航入口，后续 Codex 任务优先读取该入口及其列出的长期维护文档。
- 明确 `docs/PROJECT_PROGRESS.md` 只承载当前真实状态、风险、下一步和最近任务记录；历史任务原文不再放在根目录作为当前验收依据。
- 明确 `docs/PLUGIN_SYSTEM_ROADMAP.md` 承载完整插件系统长期路线和下一版本目标；新增 `v1.3.5` 的插件治理体验与安装升级向导收口目标。
- 新增 `docs/releases/v1.3.5.md` 草案，作为下一阶段迭代集合。
- 将 `v1.3.4` 口径从单纯“插件异常治理与验收闭环”统一为“插件异常治理与平台基础能力收口”，覆盖 manifest 校验、dry-run、配置型安装、最小升级执行、归档 / 恢复和健康总览。
- 修正过期表述：插件安装、升级和 soft uninstall 不再整体写成未实现；当前真实口径是 manifest + 配置型安装、最小升级执行和归档 / 恢复已落地，插件包 zip、远程安装、市场、动态加载、外部服务 Webhook 和 hard uninstall 仍未实现。
- 更新 SEO 红线：disabled、archived、migration failed、Hook failed 或权限拒绝均不能破坏历史 `/topics/:id` 和 `/c/:slug` SEO。
- 更新测试文档：新增 `v1.3.5` 下一阶段最小测试目标，强调只覆盖新向导和关键入口，不扩成完整浏览器矩阵。

未完成事项：

- 本轮只做文档整理，未修改任何 Go / 前端 / CSS / 测试 / Docker / Shell / 配置 / 数据库迁移代码。
- 未执行测试和构建，遵守本轮禁止事项。
- 旧版本 Release Notes 保持历史原貌，只作追溯依据；未逐条重写历史版本内容。

新发现风险：

- 历史任务记录非常长，仍保留在 `docs/PROJECT_PROGRESS.md` 下方作为审计轨迹；后续如果文档继续膨胀，可以单独做“历史任务记录归档”专项。
- `CHANGELOG.md` 的 `Next` 仍包含当前未发版变更记录，发布 `v1.3.4` 或进入 `v1.3.5` 时需要再做一次版本切分。

已执行检查命令和结果：

- 未执行测试 / 构建 / E2E / `scripts/check-frontend.sh`，原因是本轮任务明确禁止。

影响范围：

- API：无实现变更，仅同步文档口径。
- 数据库：无变更。
- 权限：无实现变更。
- SEO：无实现变更，仅同步红线口径。
- 插件系统：无实现变更，统一当前完成状态和下一阶段优先级。
- 前后台 UI：无实现变更。

下一轮建议：

1. 按 `docs/releases/v1.3.5.md` 启动插件治理体验专项：先重排 `/admin-next/plugins` 信息层级，再做安装 / 升级向导。
2. 保持文档入口规则：新增任务先判断能否合并到现有文档，不再新增根目录临时任务文档。
