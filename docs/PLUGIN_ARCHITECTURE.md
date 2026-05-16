# DevHub 插件架构说明

[返回文档入口](README.md)

更新时间：2026-05-14

## 版本定位

`v1.3.x` 是 Core + Plugins 架构拆分与插件平台收口阶段。问答、文档、Wiki 不再作为 Core 硬编码业务类型描述，而是由 `qa`、`docs`、`wiki` 三个内置系统插件注册内容类型、菜单、权限和路由描述；`v1.3.1` 进一步封口旧写入口，并补强后台内容创建 / 更新时的插件权限边界。

DevHub 的长期目标不是只支持内置 `qa/docs/wiki`，而是形成完整插件平台。Core 只提供通用社区底座，业务能力都应通过插件声明、插件状态、插件权限、插件菜单、插件配置、插件 Hook、插件 migration、插件 API、插件 SEO、插件通知和插件搜索扩展。

完整插件系统是当前最高优先级长期主线。下一阶段完整目标、生命周期和验收标准以 [完整插件系统长期完善路线图](PLUGIN_SYSTEM_ROADMAP.md) 为准。插件市场、插件包、远程安装、在线更新和动态加载不在当前代码实现范围内，但不再作为永久排除项；它们进入后续阶段路线，并必须在安全边界和 SEO 红线内推进。

## Core 边界

Core 只保留通用社区能力：

- 用户、前台认证、后台认证、子站版主授权。
- 子站、板块、通用内容、评论、标签、搜索、通知、收藏、关注、举报、审计日志。
- 权限、后台 RBAC、版主 scope 校验。
- `/topics/:id`、子站页、标签页、sitemap 和 robots 的 SEO 兜底。
- 插件注册、插件状态、插件菜单、插件权限描述和基础分发。

当前兼容命名：

- `topics` 是当前通用内容表，对应目标架构中的 `contents`。
- `categories` 是当前通用板块表，对应目标架构中的 `channels`。

## 当前插件平台基线对账

本节是 2026-05-11 的真实代码基线，用于避免把预留能力误写成已完成能力。

专项验收补充：2026-05-11 已完成一次插件系统专项验收与 E2E 回归归档。后端单测、Go 构建、后台 Docker 构建、前台 E2E、后台 E2E 和 `/topics/:id` / `/c/:slug` SEO curl 回归均已执行；后台 E2E 首次暴露的插件启停测试状态污染和影响分析旧文案断言已修复并复跑通过。该验收不改变架构边界：插件市场、插件上传、远程安装和 Go 动态加载仍是后续阶段能力，不属于当前已完成范围。

MySQL 专项补充：2026-05-11 已完成 MySQLStore 与老库升级专项验证。`db/mysql/001_schema.sql` 与 `internal/store/schema.go` 包含 `plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、结构化 `admin_logs` 以及 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types`。`db/mysql/migrations/004_community_plugins.sql` 到 `010_hook_executions.sql` 已在测试库连续执行两轮通过；MySQLStore 可选集成测试覆盖全局 / 子站插件启停强拦截、failed migration 阻断与 retry、Hook 记录查询、插件治理审计查询和 `config_schema` 校验。该专项不是完整外部迁移框架，仍不包含 migration down、硬回滚或迁移前自动备份。

当前定位：

- DevHub 当前是内置系统插件平台，并开始支持安全的 manifest + 配置型插件安装预备形态：插件通过代码 registry 或 manifest 声明接入，由 Core 负责状态、权限、菜单、配置、Hook 元信息、发布校验和后台治理分发。
- 当前不是第三方插件市场；已支持管理员 zip 插件包上传到安全沙箱并 promote 到本地插件仓库，但不支持远程安装、在线更新、Go 动态插件加载或执行第三方本地代码。
- 当前真实表名仍是 `topics` / `categories`；`contents` / `channels` 是架构概念或长期目标命名，不能在本阶段强行改表。

已完成：

- 插件注册：`qa`、`docs`、`wiki`、`projects`、`jobs`、`ai_works` 均通过统一 registry 暴露 `code`、`plugin_code`、`content_types`、权限、菜单、路由、Hook 声明、`config_schema`、依赖和最小 Core 版本。
- 内容类型归一：`doc -> document`、`wiki -> wiki_page` 和 `content_type -> plugin_code` 映射集中在 registry。
- 全局状态：`plugins.status` 已扩展为插件运行治理状态模型，当前 schema / Store 接受 `discovered`、`installed`、`migrated`、`configured`、`enabled`、`disabled`、`running`、`archived`、`config_invalid`、`migration_pending`、`migration_failed`、`dependency_missing`。
- 发布可用性：当前只有 `plugins.status=enabled` 会放行新建内容；`running`、`configured` 等状态先作为生命周期 / 健康治理预留，不等价于可发布。`archived` 表示软卸载 / 归档，必须和 disabled 一样阻断新发布和子站启用，但不影响历史内容访问和 SEO。
- 启用 readiness：`v1.3.3` 起，全局启用和子站启用都会在 Service 层检查插件存在、全局配置有效、依赖插件已启用、没有 `failed` 迁移记录；当前内置 up/no-op 的 `pending` migration 不阻断启用，只通过健康状态和迁移 Tab 提示。
- 子站状态：`community_plugins.status` 支持子站级 `enabled` / `disabled`，并叠加 `sort_order`、`config_json`。
- 配置：`plugins.config_json`、`community_plugins.config_json`、`resolved_config.default/global/community/effective` 已落地；后端保存时执行简化 `config_schema` 校验，后台用 JSON Editor + Ajv 做客户端基础校验。
- 权限：插件权限来自 manifest / registry；发布链路按内容类型读取 `create_permission`；菜单按全局状态、子站状态、权限和 scope 过滤。
- 前台入口治理：新增 `navigation/create-options/menus preview` 接口用于统一前台导航与发布入口可见性判断；插件 `disabled/archived/dependency_missing/config_invalid/migration_failed` 时隐藏新建入口但不影响历史内容与 SEO。
- HookBus：已有内置 HookBus 和最小 handler 注册；创建、更新、删除、评论、搜索、通知、SEO 以及插件启停会派发 Hook 事件。
- 影响分析：已有轻量 impact API，返回启用子站数、板块数、内容数、待审核内容数和菜单数等计数。
- 健康摘要：`GET /api/v1/admin/plugins`、`GET /api/v1/admin/plugins/health` 和 `GET /api/v1/admin/plugins/:code/health` 返回轻量 `health`，由全局状态、配置校验、迁移记录、依赖状态和 Hook 失败统计计算；当前额外返回 `status_reason` 解释主要异常原因。
- 审计：插件启停、全局配置、子站启停、子站配置、排序、Hook 失败和带 plugin_code 的插件内容治理操作写入 `admin_logs.old_value`、`admin_logs.new_value`、`admin_logs.metadata_json`。
- 迁移治理：`plugin_migrations` 表、MemoryStore / MySQLStore 读写能力、内置插件 migration 声明、up/no-op runner、失败记录、失败重试和迁移审计已存在；成功迁移不会重复执行。
- Manifest 校验与 dry-run：`PluginManifestValidator` 已可校验 manifest JSON 的基础字段、内容类型、权限、菜单、路由、Hook、配置模型、迁移、依赖和资产路径，并返回 errors / warnings / checksum / impact summary；`dry-run` 不写入插件记录。
- Manifest + 配置型安装：`POST /api/v1/admin/plugins/install` 可安装只含声明与配置的插件，初始为 installed + disabled；不执行第三方代码、不动态加载前端资源、不执行外部 SQL。
- SDK / 模板：已新增 `docs/PLUGIN_SDK.md`、`docs/PLUGIN_TEMPLATE.md`、`docs/examples/plugin-manifest-example.json` 和 `go run ./cmd/devhub plugin:new ...`，用于生成声明型插件骨架；生成器复用 `PluginManifestValidator` 和当前简化 `config_schema` 校验。
- 最小升级执行：`POST /api/v1/admin/plugins/:code/upgrade/dry-run` 和 `POST /api/v1/admin/plugins/:code/upgrade` 已支持 manifest + 配置型插件的预览和最小执行闭环；后台已提供抽屉分步升级向导；回滚、migration down、外部 SQL 和插件包升级仍未实现。
- 安装/升级审批流：新增 `plugin_approval_requests` 与审批 API（`/api/v1/admin/plugins/approvals*`），后台提供 `/admin-next/plugins/approvals` 审批中心；审批通过后执行时会重新 dry-run 校验，审批不等于绕过后端强校验。
- 软卸载 / 归档 / 恢复：插件可归档、恢复和批量归档 / 恢复；归档后禁止新建内容和子站启用，但保留历史内容、配置、迁移记录、审计记录和 SEO。
- 后台：`/admin-next/plugins` 已具备插件列表、详情抽屉、配置、impact 提示、审计 Tab、迁移 Tab 和通用插件内容页入口；`/admin-next/communities` 已具备子站插件配置抽屉。

部分完成：

- 生命周期：状态枚举已扩展，并已派生 `install_status`、`runtime_status`、`health_status`、`lifecycle_status`、`status_reason`、`installed_at`、`archived_at`、`last_health_check_at` 等后台展示字段。当前还不是完整外部插件包安装器，运行判断仍以 `plugins.status=enabled`、`community_plugins.status=enabled`、`plugin_migrations.status` 和健康摘要为准。
- Hook 治理：Hook 可以执行，blocking hook 可阻断；`hook_executions` 已记录执行结果、最近错误、失败次数、平均耗时和失败率，失败会写入 `plugin.hook.failed` / `plugin.hook.blocked` 审计。当前已有轻量健康摘要；重试策略、告警和复杂业务处理器仍待后续。
- 插件迁移：当前 runner 只支持内置插件 up/no-op 执行记录、失败记录和重试；尚无 migration down、真实 rollback、迁移前备份、外部插件迁移包或复杂迁移依赖排序。
- 权限矩阵：发布和菜单已做最小权限码校验；角色可分配、按 community / category 作用域细分的完整权限矩阵和配置 UI 仍未完成。
- 插件内容治理：通用 `PluginContent` 页已接入头部状态、精确过滤、详情抽屉、批量隐藏 / 恢复、批量审核、批量置顶、批量加精和审计跳转；专属详情、完整权限矩阵和跨页面审计高亮仍待后续。
- Projects / Jobs / AI Works：已接入 plugin_code、content_type、权限、菜单、状态和发布校验；专属扩展表、专属搜索、通知、SEO 和业务闭环未完成。

预留：

- 插件包分发：本地插件包 **dry-run 导入预览**、“本地插件仓库扫描（discovered packages）”、“已安装声明型插件导出为本地插件包”、“zip 上传安全沙箱 + promote 到本地仓库”与“上传包生命周期治理”已落地（含 `plugin_package_uploads`、导入审批、rescan、cancel/delete/cleanup、`checksums.json` sha256 校验、`risk_report` 风险报告，以及Ed25519 真实签名验签与可信发布者管理：`publisher.json`/`signature.json` + 后台 `plugin_trusted_publishers`，见 `docs/PLUGIN_PACKAGE.md`）；远程安装/市场仍处于预留阶段。
- 插件健康状态：`healthy`、`warning`、`error`、`disabled`、`migration_pending`、`config_invalid`、`dependency_missing`、`hook_warning`、`hook_error` 已有轻量计算；`hook_error` 当前基于 Hook 失败次数阈值（当前为 `>= 3`）判断。告警、自动恢复、重试队列和 Prometheus/Grafana 式可观测指标仍是后续能力。
- 插件依赖解析、版本兼容检查深化、远程插件索引、插件包签名生态和市场分发。
- 外部服务型 Webhook、动态路由加载、动态执行环境、沙箱和第三方 Hook 运行时。

## 完整插件系统路线

本节是当前架构文档中的阶段摘要；更完整的目标流程、治理能力、后台能力、运行时能力、审计能力和 E2E 要求见 [完整插件系统长期完善路线图](PLUGIN_SYSTEM_ROADMAP.md)。

当前 `VERSION` 为 `v1.6.0`，主题是“插件包上传与分发前置能力收口版”。历史阶段 `v1.4.0` / `v1.5.0` 已完成插件内容治理、本地插件包治理、配置版本、敏感配置加密、审批流和本地目录包导出；在此基础上，`v1.6.0` 补齐 zip 上传安全沙箱、上传包生命周期、Ed25519 真实验签、可信发布者、远程索引只读镜像、版本仓库、升级差异、操作恢复预览、配置密钥轮换和后台插件治理 UI 分组。远程插件市场、远程包自动下载、在线更新、动态加载、脚本沙箱、第三方代码执行、外部 raw SQL、hard uninstall 和 migration down 仍不属于当前实现。

下一阶段进入 `v1.7.x / P0` 插件分发前置增强：优先评估远程插件包下载到 staging、下载安全校验、远程索引缓存刷新、trusted publishers 只读同步草案、zip export 下载和安全 fixture 生成器；仍不引入动态执行环境。

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
- HookBus 全调用点与执行记录。
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

## 内置系统插件

- `qa`：问答插件，提供 `question` 内容类型，并承载问题、回答、采纳和已解决状态。
- `docs`：文档插件，提供 `document` 内容类型，并预留文档空间、文档树和文档详情能力。
- `wiki`：Wiki 插件，提供 `wiki_page` 内容类型，并预留页面版本、回滚和协作编辑能力。

`project`、`job`、`ai_work` 已按内置系统插件拆分为：

- `projects -> project`
- `jobs -> job`
- `ai_works -> ai_work`

历史 URL 和 `/topics/:id` SEO 输出保持不变；迁移只影响 `plugin_code` 归属和插件状态治理。

## Manifest 规范

当前内置系统插件统一使用一套 `PluginManifest` 风格的声明结构，不再允许每个插件各自定义一套字段。

统一字段包括：

- `code`
- `name`
- `version`
- `description`
- `is_system`
- `content_types`
- `content_type_definitions`
- `permissions`
- `menus`
- `routes`
- `config_schema`
- `dependencies`
- `min_core_version`
- `hooks`
- `migrations`
- `assets`

说明：

- manifest 只描述能力和元数据，不直接承载业务执行流程。
- `qa/docs/wiki/projects/jobs/ai_works` 当前都通过统一 registry 返回相同结构。
- `config_schema` 当前已经用于全局 / 子站插件配置的简化后端校验，并供后台基础自动表单、JSON 高级模式和 Ajv 做客户端基础校验；v1.5.0 已补齐“配置版本历史 + diff（脱敏）+ 回滚 dry-run 预览（不写入）”。完整 JSON Schema、字段分组、更复杂嵌套矩阵、真实回滚与敏感配置加密仍是后续任务。

## 内容类型声明

每个内容类型当前统一描述为 `ContentTypeDefinition`，至少包含：

- `type`
- `name`
- `plugin_code`
- `create_permission`
- `edit_permission`
- `delete_permission`
- `audit_permission`
- `default_status`
- `allow_comment`
- `allow_like`
- `allow_favorite`
- `seo_type`

当前内置插件映射：

- `question -> qa`
- `document -> docs`
- `wiki_page -> wiki`
- `project -> projects`
- `job -> jobs`
- `ai_work -> ai_works`

兼容归一仍集中在 registry：

- `doc -> document`
- `wiki -> wiki_page`

当前 Core 兼容内容类型主要是 `article`、`news` 等尚未拆分插件的通用内容，默认使用：

- `create_permission = core.topic.create`
- `edit_permission = post.update`
- `delete_permission = post.delete`

## 权限声明

插件权限当前统一描述为 `PermissionDefinition`，至少包含：

- `code`
- `name`
- `description`
- `scope`
- `plugin_code`

scope 语义当前约定为：

- `global`
- `community`
- `category`
- `own`

当前发布链路已接入最小权限码校验；更细粒度的权限矩阵仍是后续任务。

## 菜单声明

插件菜单当前统一描述为 `MenuDefinition`，至少包含：

- `code`
- `title`
- `path`
- `location`
- `permission`
- `plugin_code`
- `sort_order`

菜单展示必须经过以下过滤：

- 插件全局 `enabled`
- 插件在当前子站 `enabled`
- 当前用户具备菜单权限
- 当前请求作用域匹配（`admin` / `moderator` / `frontend`）

当前实现中：

- 后台左侧导航只保留“系统插件”入口
- 插件业务菜单通过系统插件列表或版主插件菜单返回

## HookBus 与 Hook

HookBus 完整化属于插件平台 P0 收口任务。当前只服务内置系统插件扩展点，不承载第三方动态执行；第三方插件执行、沙箱和动态加载进入 P3 评估，不是当前代码实现范围。

建议 Hook 名称统一为：

- `BeforeCreateContent`
- `AfterCreateContent`
- `BeforeUpdateContent`
- `AfterUpdateContent`
- `BeforeDeleteContent`
- `AfterDeleteContent`
- `AfterCreateComment`
- `OnSearchIndex`
- `OnNotificationBuild`
- `OnSEOBuild`

当前状态：

- `HookDefinition` 是 manifest 声明层，描述插件希望参与的扩展点。
- `v1.3.2` 起 HookBus 作为插件平台能力收口到 `internal/plugins`，并在 Service 的内容创建/更新/删除、评论创建、搜索、通知与 SEO 构建等流程中派发 Hook 事件。
- HookBus 当前仅注册内置系统插件 Hook handlers（编译期内置注册，不支持第三方动态加载）。
- `v1.3.2` 起 HookBus 执行结果会落入 `hook_executions`，后台可查询每个插件 Hook 的执行次数、失败次数、平均耗时、最近执行、最近失败和最近错误。
- blocking hook 失败会阻断主流程，并写入 `plugin.hook.blocked` 审计；non-blocking hook 失败不阻断主流程，但会写入 `plugin.hook.failed` 审计。
- `v1.3.4` 起提供测试 / 开发环境专用 Hook 失败注入接口：`POST /api/v1/admin/plugins/:code/hooks/:name/e2e-fail`。该接口仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用，用于自动化验证 blocking 失败阻断主流程、non-blocking 失败不阻断主流程、`hook_executions` 和审计可追踪；它不是生产治理接口。
- 当前没有第三方动态注册，也没有插件包运行时加载；HookBus 仅服务内置系统插件和后续 Core 内部扩展。
- 搜索、通知和 SEO 当前是最小调用点：已能派发事件，但还没有复杂索引、通知模板或结构化数据插件处理器。
- 完整插件业务处理器、统一失败日志、重试策略和跨 Store 事务边界属于 P0/P1 继续收口项，不能降级为低优先级优化。

失败策略约定：

- 关键 Hook：`BeforeCreateContent`、`BeforeUpdateContent`、`BeforeDeleteContent` 失败会阻断当前操作；当前没有跨 Store 事务回滚封装，后续如 Hook 写外部资源需单独设计事务边界。
- 非关键 Hook：`AfterCreateContent`、`AfterUpdateContent`、`AfterDeleteContent`、`AfterCreateComment`、`OnSearchIndex`、`OnNotificationBuild`、`OnSEOBuild` 当前不阻断主流程；当前会记录失败与审计，后续需要补重试策略、告警和健康评分。
- manifest / registry 可预留 `failure_policy`、`timeout_ms` 和 `failure_threshold` 字段；其中 `block`、`log`、`retry_later` 是治理语义，`retry_later` 当前只记录待重试状态，不代表已有异步队列。

## 配置优先级

插件配置优先级当前定义为：

1. 默认配置 / `config_schema` 约定
2. `plugins.config_json` 全局配置
3. `community_plugins.config_json` 子站配置

子站配置优先级最高。

当前真实实现说明：

- `config_schema.properties.*.default` 会参与 `resolved_config.effective` 合并，作为最低优先级默认配置来源。
- `plugins.config_json` 已落库，并可通过后台插件页和 `PUT /api/v1/admin/plugins/:code/config` 管理。
- `community_plugins.config_json` 已落地，并可通过后台子站插件配置和 `PUT /api/v1/admin/communities/:id/plugins/:code/config` 管理。
- API 返回的 `resolved_config` 以 `default`、`global`、`community`、`effective` 四段表达当前合并视图。
- 当前已完成 JSON 合法性校验与简化 `config_schema` 后端强校验；支持 `type`、`required`、`enum`、`object`、`boolean`、`string`、`number`、`integer`、`default` 和数字 `min/max`。后台插件配置支持基础自动表单 + JSON 高级模式，并用 Ajv 做客户端校验，后端保存时仍会二次校验；同时已补齐配置版本历史、版本详情与稳定 diff（脱敏），并提供“回滚 dry-run 预览”（不写入）。完整 JSON Schema、字段分组、更复杂嵌套矩阵、真实回滚和敏感字段加密是后续任务。
- 配置审计记录 `old_value`、`new_value` 和 `metadata_json.changed_keys`；配置版本 diff 支持嵌套路径（object 展开），array 以整体值对比输出稳定 diff。

## 两层插件状态

插件状态分两层：

- `plugins.status`：全局插件状态，表示插件是否在系统层面可用。
- `community_plugins.status`：子站插件状态，表示某个子站是否启用该插件。

状态值：

- `discovered`：系统识别到插件声明，当前主要是生命周期预留状态。
- `installed`：已安装 / 已注册，但未启用。
- `migrated`：迁移已完成的生命周期预留状态，当前尚无自动流转。
- `configured`：配置已完成的生命周期预留状态，当前尚无自动流转。
- `enabled`：已启用，也是当前唯一允许新建内容的全局状态。
- `disabled`：已禁用。
- `running`：运行中生命周期预留状态，当前不等价于可发布。
- `config_invalid`：配置无效治理状态，当前不放行新建内容。
- `migration_pending`：迁移待处理治理状态，当前不放行新建内容；但内置 no-op 迁移声明产生的 pending 记录通过 health/迁移 Tab 提示，不会在启用前阻断。
- `dependency_missing`：依赖缺失治理状态，当前不放行新建内容。

插件在某个子站可用必须同时满足：

- 插件已注册。
- `plugins.status = enabled`。
- `community_plugins.status = enabled`。
- 当前板块 `categories.plugin_code` 匹配插件。
- 当前 `content_type` 在 `categories.allowed_content_types` 内。
- 当前用户具备对应权限。

特殊说明：

- `core` 是兼容内置能力，Service 层视为始终可用，不要求写入 `plugins` 或 `community_plugins`。
- 禁用全局插件会影响所有子站的新发布、导航、菜单和管理入口。
- 禁用子站插件只影响该子站的新发布、导航、菜单和后台管理入口。
- 禁用插件不影响历史内容访问，尤其不能破坏 `/topics/:id` SEO 动态 HTML。
- `post.create` 是 `core.topic.create` 的历史兼容桥，不是长期主权限；插件内容创建必须使用对应 content type 的 create permission。

## 发布校验流程

发布 Topic 时应走统一插件校验：

1. 解析 community。
2. 解析 category，并校验 category 属于 community。
3. 归一 `content_type`：`doc -> document`，`wiki -> wiki_page`。
4. 根据 `content_type` 推断 `plugin_code`：`question -> qa`，`document -> docs`，`wiki_page -> wiki`，`project -> projects`，`job -> jobs`，`ai_work -> ai_works`，其他 Core 兼容类型 -> `core`。
5. 校验全局插件状态。
6. 校验当前子站插件状态。
7. 校验 `category.plugin_code` 是否匹配。
8. 校验 `content_type` 是否在 `category.allowed_content_types` 内。
9. 校验当前用户权限。
10. 写入归一后的 `topics.content_type` 和 `topics.plugin_code`。

当前真实状态：

- 步骤 1-8 已在 `ValidateTopicPluginAccess` 和 Store 层板块校验中落地。
- 步骤 9 已接入插件权限矩阵校验，权限来源统一为 `ContentTypeDefinition.create_permission`：
  - `question -> qa.question.create`
  - `document -> docs.document.create`
  - `wiki_page -> wiki.page.create`
  - `project -> projects.project.create`
  - `job -> jobs.job.create`
  - `ai_work -> ai_works.work.create`
  - Core 兼容类型 `article`、`news` 当前仍为粗粒度 `core.topic.create`（兼容旧 `post.create`）。
- `post.create` 只能作为 `core.topic.create` 的历史兼容桥；不能替代任何插件 create 权限。
- `Service.CreateTopic` 是业务创建安全入口；`Service.CreatePost` 已封口，不再允许旧 posts 业务链路绕过插件校验。
- `repo.CreateTopic` / `repo.CreatePost` 属于仓储层裸写入或兼容能力，可以保留防御性归一，但不应被 HTTP / Service 常规业务链路当作权限入口。

## 后台内容更新边界

v1.3.1 采用稳妥策略：后台编辑已存在内容时禁止修改归属和内容类型。

禁止修改：

- `community_id` / `site`
- `category_id` / `board`
- `content_type`
- `plugin_code`

允许修改：

- 标题、摘要、正文、标签、状态、置顶、精华和 SEO 等非归属字段。

原因：

- 跨子站、跨板块或跨插件迁移需要同时校验目标子站、目标板块、全局插件状态、子站插件状态、`allowed_content_types`、当前用户权限和历史扩展表一致性。
- 当前后台编辑不承担迁移职责；后续如需要，应新增迁移专项 API，而不是复用普通编辑接口。

## 兼容权限桥

`core.topic.create` 是 Core 兼容内容类型的当前创建权限。为了兼容历史后台和老角色配置，当前 `post.create` 仍可作为 `core.topic.create` 的过渡兼容权限。

当前口径：

- `post.create` 不是长期主权限。
- 插件内容类型必须使用自己的 create 权限，例如 `qa.question.create`、`docs.document.create`、`wiki.page.create`。
- API 测试已覆盖：仅拥有 `post.create` / `core.topic.create` 时不能创建 `question`；拥有 `qa.question.create` 时才能创建 `question`；`post.create` 继续兼容 Core `article`。
- `article` / `news` 后续要么明确作为 Core 内容定义继续存在，要么拆为插件，再逐步移除 `post.create` 兼容桥。

## 数据结构

插件相关表：

- `plugins`
- `community_plugins`
- `plugin_migrations`
- `qa_questions`
- `qa_answers`
- `docs_spaces`
- `docs_documents`
- `wiki_spaces`
- `wiki_pages`
- `wiki_page_versions`

增强字段：

- `topics.plugin_code`
- `categories.plugin_code`
- `categories.allowed_content_types`

`plugin_migrations` 当前是插件迁移治理记录表。v1.3.2 已支持内置插件 migration 声明、查询、up/no-op 执行、失败记录、失败重试和审计；qa/docs/wiki 的第一批 migration 用于确认扩展表已经由主 schema / 启动迁移创建，不会重复破坏数据。v1.3.4 已补齐测试专用 failed migration 注入和启用阻断验收：当某插件存在 `failed` migration 时，全局启用和子站启用都会在 Service readiness 阶段失败，必须 retry 成功后才能恢复启用。migration down、真实 rollback、迁移前备份和外部插件迁移包仍是后续任务。

测试 helper：

- `POST /api/v1/admin/plugins/:code/migrations/:name/e2e-fail` 仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用。
- 该接口只用于自动化测试构造 failed migration，不是生产迁移入口。
- 注入操作写入 `plugin.migration.failed` 审计，metadata 中标记 `test_injection=true`。

## API 与菜单

已注册的插件 API 包括：

- `GET /api/v1/plugins`
- `GET /api/v1/communities/:slug/plugins`
- `GET /api/v1/admin/plugins`
- `GET /api/v1/admin/plugins/health`
- `GET /api/v1/admin/plugins/:code/health`
- `POST /api/v1/admin/plugins/manifest/validate`
- `POST /api/v1/admin/plugins/dry-run`
- `POST /api/v1/admin/plugins/install`
- `POST /api/v1/admin/plugins/:code/enable`
- `POST /api/v1/admin/plugins/:code/disable`
- `POST /api/v1/admin/plugins/:code/archive`
- `POST /api/v1/admin/plugins/:code/restore`
- `POST /api/v1/admin/plugins/bulk-archive`
- `POST /api/v1/admin/plugins/bulk-restore`
- `PUT /api/v1/admin/plugins/:code/config`
- `GET /api/v1/admin/plugin-menus`
- `GET /api/v1/admin/communities/:id/plugins`
- `POST /api/v1/admin/communities/:id/plugins/:code/enable`
- `POST /api/v1/admin/communities/:id/plugins/:code/disable`
- `PUT /api/v1/admin/communities/:id/plugins/:code/config`
- `PUT /api/v1/admin/communities/:id/plugins/sort`
- `GET /api/v1/moderator/plugin-menus`

菜单策略：

- 后台左侧导航只保留“系统插件”入口。
- 插件业务管理页通过系统插件列表进入，避免 qa / docs / wiki 直接散落到左侧导航。
- 版主插件菜单必须同时满足全局 enabled、子站 enabled、当前用户是该子站版主、当前用户具备菜单权限。

后台插件管理体验：

- 后台“系统插件”按功能分页呈现：`/admin-next/plugins` 重定向到 `/admin-next/plugins/overview`，二级页面包括 `overview/list/content/install/config/dependencies/hooks/events/search-index/navigation/permissions/audit/developer`。
- 插件列表页（`/admin-next/plugins/list`）展示全局插件列表、状态 badge、系统插件标识、内容类型、权限数量、菜单数量和 `config_schema` 摘要，并提供筛选、批量归档/恢复与详情抽屉入口。
- 安装升级页（`/admin-next/plugins/install`）提供 manifest validate / dry-run / install、`upgrade dry-run` 与 `upgrade` 入口，用于展示版本兼容矩阵、变更字段和 diff，并执行 manifest + 配置型插件的最小升级闭环。
- 插件详情使用抽屉分区展示基础信息、内容类型、权限、菜单、配置、路由和 Hooks，避免把 JSON 直接堆在表格中。
- 插件详情抽屉当前增加统一可读状态提示，展示运行状态说明、归档态提示、状态原因和建议操作。
- 全局插件配置已升级为基础自动表单 + JSON 高级模式（`json-editor-vue`），并使用 Ajv 做 `config_schema` 基础校验；后续仍可增强为更完整的 JSON Schema、深层嵌套和字段分组。
- `/admin-next/communities` 的子站插件配置抽屉展示全局状态和子站状态双 badge，并支持子站启用 / 禁用、`config_json` 编辑、JSON 格式化、数字排序和禁用原因提示。
- 全局禁用和子站禁用都有二次确认，并明确 disabled 只影响新发布、导航、菜单和管理入口，不影响历史内容详情页和 SEO。
- 插件影响范围统计已提供轻量 impact 计数接口：
  - `GET /api/v1/admin/plugins/:code/impact`
  - `GET /api/v1/admin/communities/:id/plugins/:code/impact`
  UI 在接口不可用时必须显示“待接口支持/暂不可用”，不得伪造数字。
  当前返回字段包括历史内容数、启用/禁用子站数、绑定板块数、近 7 天内容数、审核中内容数、菜单声明数、配置覆盖数、待执行迁移数和近 7 天 Hook 失败数；`recent_hook_errors_count` 来自 `hook_executions`，仍只是轻量提示。
  同时插件详情抽屉提供“审计”Tab，使用 `GET /api/v1/admin/plugins/:code/audit-logs` 展示插件启停、配置、Hook 失败和带 plugin_code 的内容治理审计；筛选支持 action、community_id、actor、target_type、target_id、metadata、request_id 和时间范围。
  插件列表与详情抽屉的“运行状态”展示轻量 `health` 摘要，覆盖 overall、status_reason、config、migration、Hook、dependency、recent_error 与 suggested_action；这不是完整监控系统。

## 当前限制与阶段边界

- 当前已支持 manifest + 配置型插件的校验、dry-run、安装记录、升级预览和最小升级执行闭环，也支持 zip 上传安全沙箱、上传包生命周期治理与 promote 到本地插件仓库；仍不支持远程市场安装、在线更新、Go 动态插件加载或执行第三方本地代码。这些能力分别进入 P2 / P3 路线，后续推进时必须满足安全红线、权限隔离、migration 备份回滚和 SEO 不退化。
- 插件路由当前是注册描述 + Core 分发；动态路由加载和动态执行环境进入 P2 / P3 路线评估。
- Docs / Wiki 的专用编辑体验仍是部分完成。
- 子站插件配置和排序已有 API 与增强后的后台 UI，但仍需继续做真实浏览器矩阵验收。
- 插件治理审计已新增 `admin_logs.old_value`、`admin_logs.new_value` 和 `admin_logs.metadata_json` 结构化字段，同时保留 `target` 文本摘要兼容旧展示；非插件历史日志可能仍没有结构化 diff。
- 新装库已在 `db/mysql/001_schema.sql` 和 `internal/store/schema.go` 包含结构化审计字段；老库升级使用 `db/mysql/migrations/007_admin_logs_structured_plugin_audit.sql`，启动迁移辅助也会尝试补齐这些列。
- `plugins.config_json` 与 `community_plugins.config_json` 已可写，并已做 JSON 格式校验和简化 `config_schema` 基础校验；基础自动表单、配置 diff UI、effective config 预览与配置版本历史/回滚预览已接入后台插件治理体验；更完整 JSON Schema、字段分组、更复杂嵌套矩阵与真实回滚属于后续任务。
- v1.5.0 起，插件配置支持“敏感字段加密存储”最小闭环：保存配置时会按 `config_schema` 标记与字段名规则识别敏感字段，并对敏感字段值使用 AES-256-GCM 加密后入库；v1.6.0-P1-07 起新增密钥环（current/old keys）与密钥轮换预检/受控 re-encrypt，新写入密文为 `enc:v2:<key_id>:<nonce_b64>:<cipher_b64>`，同时兼容读取旧 `enc:v1:<nonce_b64>:<cipher_b64>`。API/审计/版本历史/diff/回滚预览只返回脱敏占位（例如 `[REDACTED]` / `[ENCRYPTED]`），不会返回明文或密文。当前不支持 KMS/Vault、自动定时轮换与历史明文批量迁移（需后续专项）。
- v1.5.0-P2-10 起，已安装声明型插件可导出为本地插件包目录：导出 `manifest.json`、自动 README、脱敏 `config.example.json`、`checksums.json` 与可选 docs/migrations/publisher/signature 草案，输出目录固定在 `storage/plugins/exports/`。导出不会包含真实全局/子站配置、敏感明文或 `enc:v1:` 密文，不包含用户数据、Hook 历史、审计原始明细、搜索索引、运行时代码或外部 SQL；导出后会自动执行 package dry-run 自检。
- HookBus 当前是内置插件运行时调度器；调用点已覆盖内容创建、更新、删除、评论、搜索、通知和 SEO，并记录执行结果与失败审计。搜索 / 通知 / SEO 仍是预留级事件派发，完整业务处理器、重试策略和健康状态属于 P0/P1。
- 插件生命周期当前已能派生安装 / 运行 / 健康状态，但仍不是完整外部插件包安装器状态机；代码真实运行门禁仍以 `plugins.status`、`community_plugins.status`、`plugin_migrations.status`、依赖检查和配置校验为准。
- `v1.3.4` 的架构重点不是扩展具体业务插件，而是验证平台治理：failed migration 必须阻断启用并可 retry 恢复，blocking Hook 必须能阻断主流程，non-blocking Hook 必须不阻断但可追踪，权限矩阵必须继续弱化 `post.create` 兼容桥，MySQLStore / 老库升级必须与 MemoryStore 口径一致，ManifestValidator / dry-run / manifest + 配置型安装和升级预览 / 执行提供安全的外部生态预备能力。当前 MySQLStore / 老库升级专项已完成关键链路验证，剩余风险主要是生产大库备份、回滚、耗时、外部插件真实 DDL migration、外部服务 Webhook、升级流程和版本兼容矩阵设计。

## MySQLStore 与老库升级边界

新装库结构：

- 权威 schema：`db/mysql/001_schema.sql` 与 `internal/store/schema.go`。
- 插件平台表：`plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`。
- 审计表：`admin_logs` 已包含 `old_value`、`new_value`、`metadata_json`。
- 内容 / 板块兼容字段：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types`。

老库升级顺序建议：

1. 先备份数据库，确认可恢复。
2. 在预发或临时测试库执行 `db/mysql/001_schema.sql`，确认新装 schema 可重复执行。
3. 按编号执行 `db/mysql/migrations/004_community_plugins.sql` 到 `010_hook_executions.sql`。
4. 启动应用，让 MySQLStore 启动迁移辅助与 `seedPlugins` 兜底补齐内置插件记录和默认子站插件关系。
5. 执行 MySQLStore 专项测试或等价手工验收，确认插件启停、配置校验、迁移记录、Hook 记录和审计查询可用。

当前已验证：

- `004` 会确保 `plugins` 表存在，避免按编号执行时 `community_plugins` 回填依赖缺表。
- `005` 使用 MySQL 8 `INFORMATION_SCHEMA + PREPARE` 幂等补齐 `topics.plugin_code`、`categories.plugin_code` 和 `categories.allowed_content_types`。
- `004`-`010` 插件迁移脚本在测试库连续执行两轮通过。
- 可选集成测试 `DEVHUB_MYSQL_TESTS=1 ... go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v` 已覆盖 MySQLStore 与 MemoryStore 关键插件平台行为一致性。

仍需保留的风险说明：

- 当前迁移体系不引入复杂外部 migration runner；生产升级仍需手工备份、预发演练和回滚预案。
- 当前内置 plugin migration 仍是 up/no-op 记录型 runner；外部插件阶段如需真实 DDL runner，必须重新设计事务边界、失败恢复、备份和回滚。

## v1.3.4 收口与 v1.3.5 边界

v1.3.4 的完成口径是“插件异常治理与平台基础能力收口”，不是插件市场或动态运行时实现。本阶段已经收口：

- failed migration 注入、启用阻断、retry 恢复和迁移审计。
- HookBus blocking / non-blocking 失败注入、执行记录、审计和后台 Hooks Tab 失败摘要。
- `ContentTypeDefinition.create_permission` 驱动的创建权限矩阵，`post.create` 继续降级为 `core.topic.create` 兼容桥。
- MySQLStore / 老库升级专项验证，覆盖插件平台结构和核心行为一致性。
- 插件健康状态和审计定位能力，包含 `hook_warning` / `hook_error` 与插件审计多维筛选。
- Manifest 校验、dry-run、manifest + 配置型安装、最小升级执行、归档 / 恢复、批量归档 / 恢复和健康总览。

v1.3.5 的边界是治理体验收口，不新增危险运行时能力。当前工作区已完成以下主体能力，剩余工作是验收、skip 处置、PluginContent 小范围对齐和版本切分：

- `/admin-next/plugins` 信息架构、筛选 / 批量操作 / 列表 / 详情层级优化。
- 完整安装向导和完整升级向导。
- 批量归档 / 恢复影响预览、成功 / 失败表格和审计跳转。
- 状态治理页异常处理入口。
- PluginContent 历史治理体验对齐。
- 最小后台 E2E 回归。

当前实现进展：

- `/admin-next/plugins` 已重排为页面头部主操作、列表 / 状态治理双视图、核心统计卡、健康摘要、筛选面板、批量操作面板和精简表格。
- 插件表格操作列已收口为“详情 / 配置 / 更多”，危险操作仍保留二次确认和后端权限校验。
- Manifest 校验、dry-run、安装、升级预览和执行升级已进入右侧抽屉分步流程，结构化展示 errors、warnings、依赖、冲突、迁移计划、安装影响、版本兼容和 diff。
- 批量归档 / 恢复已展示操作前影响预览和操作后 succeeded / failed 明细，并提供审计跳转。
- 状态治理视图已聚合迁移待处理、迁移失败、Hook 异常、配置无效、依赖缺失和已归档插件；它是异常处理入口，不是完整监控系统。

以下能力仍不属于当前实现范围：插件市场、远程安装、在线更新、Go 动态加载、第三方插件沙箱、hard uninstall 和 migration down。

## 阶段 B：插件治理体验增强

阶段 B 开始把 P1 中的部分体验能力落到后台插件治理中心，但仍不改变插件平台安全边界：

- 后台插件治理相关页面接入 `vue-i18n`，默认语言为 `zh-CN`。用户可见标签、按钮、状态、筛选项和提示文案应集中到 `web/admin-app/src/i18n` 管理；`plugin_code`、`content_type`、`hook_name`、权限码和 JSON key 等技术值继续保留原始值。
- 当前已补齐插件中心、插件详情抽屉、子站插件配置抽屉、配置编辑器、通用 PluginContent 和审计列表中的主要插件治理文案；其中 `config_schema`、`config_json`、`resolved_config` 等作为用户可见标签时显示为“配置模型 / 子站配置 / 最终生效配置”，作为 JSON key 或接口字段时仍保持原值，便于调试。
- 插件配置编辑器支持“表单模式 + JSON 高级模式”。当前统一组件为 `PluginConfigEditor`（`web/admin-app/src/components/plugin/PluginConfigEditor.vue`），表单模式根据 `config_schema.properties` 做基础渲染，支持 string、number、integer、boolean、array、object 和 enum；复杂配置仍可使用 JSON 高级模式。
- 配置编辑器展示配置差异和最终生效配置。差异预览只用于管理员确认，保存仍以后端 `config_schema` 校验为准；敏感字段会在预览中脱敏。
- 通用 PluginContent 页支持多选、批量隐藏 / 恢复、审核通过 / 拒绝、置顶 / 取消置顶、加精 / 取消加精，并提供审计入口。审计入口会跳转到通用治理审计页并预填 `plugin_code`、`content_type`、`action`、`target_type` 和插件编码 metadata，便于定位本次插件内容批量操作。权限、插件状态、内容归属和审计仍由后端批量治理接口强制校验，前端隐藏或按钮禁用不能替代权限控制。
- Hook 排障：插件详情抽屉 Hooks Tab 会展示 Hook 聚合统计和最近执行记录，执行记录支持筛选查询与详情抽屉；查询接口严格使用 admin token + `plugin.read`，测试/开发环境才允许通过注入接口模拟失败，不暴露为生产功能。
- 代码结构治理：插件治理后端 handler 与 service 已按功能拆分，路由与返回结构不变；入口路由仍在 `internal/transport/httpapi/router.go`，插件相关 handler 分散到 `internal/transport/httpapi/plugin_*_handler.go`；插件生命周期相关 service 方法从 `internal/service/service.go` 迁出到 `internal/service/plugin_lifecycle_service.go`，其余插件能力继续维持按 `plugin_*.go` 分文件组织。

当前限制：

- 当前 `en-US` 只是占位语言包；若后续需要完整多语言生态，需要补齐翻译和语言切换入口。
- 自动表单是基础版本，不包含完整 JSON Schema、字段分组或更复杂嵌套字段矩阵；v1.5.0 已补齐配置版本历史与回滚 dry-run 预览，但真实回滚与敏感字段加密仍在后续版本推进。
- PluginContent 已接入批量隐藏 / 恢复、审核、置顶和加精；更细粒度权限矩阵、跨页面审计高亮和更完整审计筛选 E2E 仍待后续补测。

## 阶段 C/D/E/F：SDK 模板、生命周期、软卸载和外部生态设计

本阶段继续沿插件平台主线推进，但仍不实现插件市场、插件包上传安装、远程安装、Go 动态加载、第三方沙箱或第三方代码执行。

阶段 C 已新增插件 SDK / 模板规范：

- [插件 SDK 文档](PLUGIN_SDK.md)
- [插件生成模板](PLUGIN_TEMPLATE.md)
- [可校验 manifest 示例](examples/plugin-manifest-example.json)
- [manifest 示例](plugins/manifest.example.json)
- [插件目录模板](plugins/plugin-template.md)
- [config_schema 开发指南](plugins/config-schema-guide.md)
- [Hook 开发指南](plugins/hook-guide.md)
- [migration 开发指南](plugins/migration-guide.md)
- [权限开发指南](plugins/permission-guide.md)
- [菜单与路由指南](plugins/menu-route-guide.md)
- [外部插件生态评估与预备设计](plugins/external-plugin-ecosystem.md)

阶段 C 还补齐了 manifest 校验和安装预备能力：

- `PluginManifestValidator` 可以校验 manifest JSON 的基础字段、内容类型、权限、菜单、路由、Hook、配置模型、迁移、依赖和资产路径。
- `POST /api/v1/admin/plugins/manifest/validate`、`POST /api/v1/admin/plugins/dry-run`、`POST /api/v1/admin/plugins/install`、`POST /api/v1/admin/plugins/:code/upgrade/dry-run` 和 `POST /api/v1/admin/plugins/:code/upgrade` 提供 manifest + 配置型插件的安全预备 / 执行闭环。
- 当前 install 只写入插件记录、默认配置和迁移记录，不执行第三方本地代码，也不加载前端资源。

阶段 D 已为内置插件补齐安装生命周期展示模型：

- 内置插件仍通过 Go registry 注册，启动 / store 初始化时同步到 `plugins` 表。
- `plugins.status` 支持 `archived` 和 `migration_failed`，并派生 `install_status`、`runtime_status`、`health_status`、`lifecycle_status`、`status_reason`、`installed_at`、`archived_at`、`last_health_check_at` 给后台展示。
- 启用前会检查插件存在、未归档、配置有效、依赖可用、无 failed migration；当前内置 pending/no-op migration 仍作为治理提示，不阻断启用。

阶段 E 已实现插件软卸载 / 归档 / 恢复最小闭环：

- `POST /api/v1/admin/plugins/:code/archive` 将插件置为 `archived`。
- `POST /api/v1/admin/plugins/:code/restore` 将归档插件恢复为 `disabled`，不会自动启用。
- `POST /api/v1/admin/plugins/bulk-archive` / `bulk-restore` 支持批量软卸载 / 恢复，但不删除任何历史数据。
- v1.7.0-P0-07：新增 `plugin_uninstall_tasks` 软卸载任务记录与治理 API（impact / list / detail / retry / delete），用于审计与失败可重试；实际软卸载语义仍以 `plugins.status=archived` 为源。
- v1.7.0-P0-08：新增 `plugin_upgrade_tasks` 升级任务记录与升级 API（upgrade-impact / upgrade-from-package / list / detail / retry / delete），升级输入依赖 `compat-check(can_install=true)`，升级后默认不自动启用、不执行 migration，需要重新 enable-precheck + enable。
- 归档后禁止新建该插件内容、禁止子站启用、隐藏入口；历史内容、配置、迁移记录、审计记录和 SEO 均保留。
- 归档插件仍允许后台进入通用 `PluginContent` 历史内容治理页；页面必须提示“插件已归档，只能治理历史内容，不能新建”，当前已覆盖批量隐藏 / 恢复、审核、置顶和加精。
- 归档 / 恢复写入 `plugin.archived`、`plugin.restored`、`plugin.archive.failed`、`plugin.restore.failed` 审计。
- 2026-05-12 已补浏览器回归：前台发布页不展示归档插件 content_type、强传归档 content_type 被拒绝、子站不能启用归档插件、历史 `/topics/:id` SEO 不丢；后台覆盖归档 badge、影响范围、恢复后默认 disabled 提示和 PluginContent 归档态历史治理。

内置插件 manifest 对照表：

| 插件 | content_types | 权限 | 菜单 | 路由 | Hook | config_schema | migrations | 生命周期 |
|---|---|---|---|---|---|---|---|---|
| `qa` | `question` | `qa.question.create` 等 | 前台 / 后台 / 版主 | 声明级 | 创建前后等 | 已声明 | `qa_questions` / `qa_answers` | installed / enabled / archived 等 |
| `docs` | `document` | `docs.document.create` 等 | 前台 / 后台 / 版主 | 声明级 | 创建前后等 | 已声明 | `docs_spaces` / `docs_documents` | installed / enabled / archived 等 |
| `wiki` | `wiki_page` | `wiki.page.create` 等 | 前台 / 后台 / 版主 | 声明级 | 创建 / 更新等 | 已声明 | `wiki_spaces` / `wiki_pages` / `wiki_page_versions` | installed / enabled / archived 等 |
| `projects` | `project` | `projects.project.create` 等 | 前台 / 后台 | 声明级 | 平台 Hook | 已声明 | 平台记录 | 平台治理已接入，业务闭环待完善 |
| `jobs` | `job` | `jobs.job.create` 等 | 前台 / 后台 | 声明级 | 平台 Hook | 已声明 | 平台记录 | 平台治理已接入，业务闭环待完善 |
| `ai_works` | `ai_work` | `ai_works.work.create` 等 | 前台 / 后台 | 声明级 | 平台 Hook | 已声明 | 平台记录 | 平台治理已接入，业务闭环待完善 |

阶段 F 已新增 [外部插件生态评估与预备设计](plugins/external-plugin-ecosystem.md)。推荐路线是 manifest + 配置型插件，再评估外部服务型插件；Go 动态插件暂不推荐。当前仍不执行第三方代码，外部服务型 Hook / Webhook、升级流程、兼容矩阵和批量治理仅进入设计或预备阶段。

## v1.4.0-P1-07 依赖检查与版本兼容矩阵

- `dependencies` 当前支持对象数组：`code`、`version`、`required`、`reason`，并兼容旧字符串数组；旧字符串依赖按 required 处理。
- 统一依赖检查覆盖存在性、安装 / 启用状态、归档、迁移失败、配置无效、版本约束、自依赖和循环依赖；required 不满足阻断 validate / dry-run / install / upgrade dry-run / upgrade / enable，optional 缺失只 warning。
- 版本约束只支持数字 `x.y.z`、精确版本、`>=`、`>`、`<=`、`<` 和空格组合范围；不支持 npm 风格 `^`、`~`、`||` 或预发布标签。
- Core 兼容矩阵由后端统一计算，当前 Core 来自项目 `VERSION`；`min_core_version` 高于当前 Core 或 `compatible_core_version` 不满足时阻断。
- 后台安装向导、升级向导和插件详情 Dependencies 区域展示依赖总览、逐项状态、阻断原因、Core 兼容状态和升级依赖 diff；不做自动安装依赖、远程下载、市场推荐或依赖图大屏。

## v1.6.0-P0-04：远程插件索引只读镜像

远程插件索引模块是插件市场前置能力，只读读取静态 `index.json` 元数据，不下载、不安装、不执行任何远程包内容。

模块边界：

- `plugin_remote_indexes`：保存索引源配置、最近拉取状态、索引 hash 和缓存元数据。
- Service：负责 URL 安全校验、SSRF 防御、GET 拉取、JSON schema 校验、trusted publisher 联动、Core 兼容性和已安装状态计算。
- Handler：提供后台 API，保持 admin token + `plugin.read` / `plugin.manage` 权限边界。
- Admin UI：`/admin-next/plugins/remote-indexes` 只读展示索引源、远程插件列表和详情。

安全边界：远程索引不会触发 package download、zip 解压、安装、升级、动态加载、脚本沙箱或第三方代码执行。远程 publisher 不会自动进入本地 trusted publishers。

## v1.6.0-P0-05：插件包版本仓库与升级差异对比

插件包版本仓库是聚合层，不新增运行时能力。它从已安装插件、本地插件仓库、上传包记录和远程只读索引中聚合同一 `plugin_code` 的多个版本，标记 installed/local/uploaded/remote 来源、风险、签名、publisher trust 和 Core 兼容性。

升级差异对比仍复用声明型 manifest 和 package dry-run，不执行第三方代码、外部 SQL 或动态前端资产。`diff_sections` 按 manifest 能力边界分组，供后台和审批详情展示；远程索引版本保持只读，不能直接升级。

## v1.6.0-P1-09 插件治理后台信息架构

后台“系统插件”在不改变后端 API 和安全边界的前提下，按功能分为六个治理分组：

- 插件运营：概览、插件列表、内容治理、配置中心。
- 插件包治理：本地插件仓库、zip 上传包、安装入口、导出入口、版本仓库、升级差异。
- 安全与可信：checksum / 风险报告、签名验签、可信发布者、敏感配置加密、密钥轮换。
- 流程与恢复：安装 / 升级审批、操作历史、失败恢复、回滚预览。
- 运行治理：依赖兼容、Hook 排障、前台入口、搜索索引、事件通知。
- 远程与开发者：远程插件索引、SDK 文档、Manifest 模板、插件包规范。

UI 只负责展示后端返回的状态、风险和建议，不在前端重新推导安全结论。公共展示组件包括状态 tag、风险 tag、签名 / checksum 摘要、安全边界提示和结构化错误提示；敏感字段、密文、私钥和系统绝对路径仍不得展示给前端。

## v1.7.0-P0-01 远程插件包下载架构

远程插件包下载只负责把远程包安全落到 staging，不进入安装阶段。

- Handler：`POST /api/v1/admin/plugins/packages/download`、`GET /api/v1/admin/plugins/packages/staging`、`GET /api/v1/admin/plugins/packages/staging/:id`、`DELETE /api/v1/admin/plugins/packages/staging/:id`，均要求 admin token，写操作需要 `plugin.write`，读操作需要 `plugin.read`。
- Service：统一执行 URL 校验、DNS/IP SSRF 防护、重定向复检、大小限制、临时文件写入、sha256 计算与 staging rename。
- Store：`plugin_package_downloads` 持久化下载记录；MemoryStore / MySQLStore 均支持，失败也保留错误状态和错误信息。
- Storage：文件保存到 `storage/plugins/staging/downloads/`，不使用远程文件名，不允许写出 staging 目录。
- Audit：记录 `plugin.package.download.requested/success/failed/rejected`、`plugin.package.checksum.failed`、`plugin.package.staging.deleted`。

安全边界：远程下载不会安装插件、不会启用插件、不会解压执行包内容、不会运行脚本、不会加载 Go plugin、不会执行 SQL、不会动态加载前端资产。

## v1.7.0-P0-03 插件依赖 / 兼容性检查架构

P0-03 在 staging 包完成解压安全检查与 manifest 预校验之后执行。后端新增 `plugin_package_prechecks` 作为预检输入来源，新增 `plugin_package_compat_checks` 保存每次依赖 / 兼容性检查结果；MemoryStore 和 MySQLStore 均持久化记录，不只存在内存。

Service 层负责所有安全结论：读取 precheck passed 记录、解析 manifest、读取当前 `VERSION`、检查 Core 版本约束、依赖存在 / 启用 / 版本、plugin_code / content_type / permission / menu / route 冲突、HookBus 兼容、简化 config_schema、migration 声明和 `can_install`。前端只展示后端返回的 status、blockers、warnings 和 summary，不重新计算风险。

Handler 提供 `POST /api/v1/admin/plugins/packages/prechecks/:id/compat-check`、compat-check 列表 / 详情 / 删除接口，并写入 `plugin.package.compat_check.*` 审计。该链路不会安装、启用、注册权限 / 菜单 / 路由 / Hook、不会执行 migration 或任何插件代码。

## v1.7.0-P0-05 插件启用前安全检查架构（enable-precheck）

P0-05 在“插件已安装但未启用”的状态下执行启用前最后检查，产出 `can_enable` 结论供后续启用流程使用。本轮只检查，不真正启用或注册到运行时。

- Handler：
  - `POST /api/v1/admin/plugins/:code/enable-precheck`
  - `GET /api/v1/admin/plugins/enable-prechecks`
  - `GET /api/v1/admin/plugins/enable-prechecks/:id`
  - `DELETE /api/v1/admin/plugins/enable-prechecks/:id`

## v1.7.0-P0-06 插件启用与运行时注册架构（enable）

启用阶段用于将已安装插件从 `disabled` 切换为 `enabled`，并让插件的声明能力进入平台治理链路（内容类型校验、权限矩阵、菜单/路由/Hook 声明可见性等）。该阶段不执行插件包内代码，不运行脚本，不加载 Go plugin，不自动执行 migration。

输入约束：

- 必须基于 enable-precheck 记录执行：`enable-precheck status=passed|warning` 且 `can_enable=true`。
- 启用时会做 TOCTOU 快速再校验（配置/依赖/迁移/content_type 冲突等），防止安装后状态变化导致绕过。
- 本轮策略：存在 pending migration 直接阻断启用。

核心链路：

1. 创建 `plugin_enable_tasks` 记录，状态 `enabling`。
2. 校验 enable-precheck / readiness / 冲突。
3. 更新插件状态为 `enabled`（DB-as-source，前后台展示和内容创建校验链路据此生效）。
4. 写入注册摘要与 effective_config 快照（不泄露敏感值）。
5. 写入审计日志：`plugin.enable.*` / `plugin.runtime.registered`。
6. 触发 HookBus：`AfterPluginEnabled`（non-blocking）。

相关接口：

- `POST /api/v1/admin/plugins/enable-prechecks/:id/enable`
- `GET /api/v1/admin/plugins/enable-tasks`
- `GET /api/v1/admin/plugins/enable-tasks/:id`
- `POST /api/v1/admin/plugins/enable-tasks/:id/retry`
- `DELETE /api/v1/admin/plugins/enable-tasks/:id`

安全边界：启用流程只注册声明能力并切换状态，不会执行插件包内代码/脚本，不会加载 Go plugin，不会自动执行 migration。
