# DevHub 插件系统长期完善路线图

[返回文档入口](README.md)

状态：Core + 插件服务底座长期路线与当前需求分层文档。本文定义 DevHub 插件系统的目标流程、治理能力、运行时能力、审计能力和测试要求，并按当前代码事实区分已完成、收尾、下一阶段和预留能力；插件运行模型详细设计见 [插件运行模型设计](PLUGIN_RUNTIME_MODEL.md)。

更新时间：2026-05-19

## 当前实现快照与需求分层

DevHub 的长期目标是 **Core + 插件 的开源服务底座**。Core 稳定化、插件扩展点、插件治理、插件运行模型和插件生态是后续开发主线；默认社区能力是 Core 基础能力之一，不是项目唯一边界。

`VERSION` 当前为 `v1.8.3`；当前实现已从早期插件包治理 / 签名验签 / 运行模型设计推进到后台插件治理稳定性、声明型插件真实业务闭环、插件包真实 zip 验收、PluginRegistry reload、external_service non-blocking Webhook 和中文状态提示收口。远程包、声明型插件和 external_service 仍不执行第三方代码、不开放动态加载、不开放远程 iframe、不做 blocking Hook。

当前实现快照：

- 已完成：v1.5 本地插件包规范、dry-run、checksum / risk_report、本地仓库扫描、本地包安装、配置版本历史、敏感配置加密、审批流、签名草案、已安装声明型插件目录导出。
- 已完成：v1.6 zip 上传安全沙箱、上传包生命周期列表 / 详情 / rescan / approval / promote / cancel / delete / cleanup、真实签名验签、可信发布者 CRUD / block / revoke / restore、远程索引只读镜像、版本仓库、升级差异、操作历史 / recover dry-run / cleanup、配置密钥轮换 dry-run / re-encrypt、后台插件治理按功能域分层导航与 E2E helper（导航入口见 `web/admin-app/src/router/adminNav.js`）。
- 已完成：v1.7-P0-01 远程插件包下载到 `storage/plugins/staging/downloads/`，包含 HTTPS-only、SSRF 防护、重定向复检、大小限制、sha256 校验、失败清理和审计记录。
- 已完成：v1.7-P0-05 插件启用前安全检查（enable-precheck）：基于已安装插件二次复检文件/manifest/依赖/配置/迁移/冲突，产出 `can_enable` 结论并持久化记录；本轮只检查不真正启用或注册运行时。
- 部分完成：操作快照和恢复以 dry-run / cleanup / 人工恢复建议为主，不提供全量业务数据自动回滚；配置历史默认不做批量 re-encrypt；插件包导出当前是目录包导出，不提供 zip 下载与正式签名打包。
- 仍未完成 / 后置：远程包解压预校验、远程安装、在线更新、自动安装依赖、远程可信源同步、完整 PKI / CA、动态加载、脚本沙箱、第三方代码执行、外部 raw SQL 自动执行、hard uninstall、migration down、完整远程插件市场交易 / 评分 / 评论系统。

## Core + 插件服务底座阶段路线

- 阶段一：插件包治理与生命周期收口。目标是完成插件包结构、manifest、checksum/signature、staging、预检、compat-check、安装、启用、软卸载、升级、审计和失败恢复。
- 阶段二：远程插件包 staging 下载与安全校验。目标是远程索引中的包可以安全下载到本地 staging，但不自动安装、不运行、不动态加载。
- 阶段三：插件运行模型设计。目标是明确 Core 内置插件、外部 HTTP 服务插件、iframe / sandbox 前端插件三种运行模式，以及第三方插件前端、后端、Hook、权限、配置、API 调用和安全隔离模型。
- 阶段四：前端插件挂载模型。目标是定义后台菜单、前台 slots、配置页面、iframe / sandbox / postMessage 等隔离边界。
- 阶段五：HTTP 插件服务协议。目标是定义后端插件服务以独立 HTTP 服务方式运行，DevHub 通过 HookBus 和受控 API 与插件通信。
- 阶段六：真实官方插件验证。目标是实现公告、友情链接、SEO 扩展、统计代码等官方示例插件，用真实插件反推 Core 扩展点是否可用。
- 阶段七：插件市场 / 远程分发能力。目标是在插件包治理、运行模型和官方插件验证完成后，再推进远程分发、插件市场和第三方插件生态。

下一阶段建议优先级：先把 `docs/PLUGIN_RUNTIME_MODEL.md` 中的前端挂载、HTTP 插件服务协议、受控 API token 和隔离边界拆成实现任务，再推进官方示例插件验证和远程分发能力。动态 Go 加载、脚本沙箱、第三方代码执行、完整插件市场交易 / 评分 / 评论系统继续后置。

补充：v1.8.0 文档阶段已新增插件前端挂载模型设计文档：

- `docs/PLUGIN_FRONTEND_MOUNT_MODEL.md`：定义 slots、iframe/sandbox 基线策略与 postMessage 协议（设计口径，未实现第三方插件前端挂载）。

v1.8.1 实现补充：

- 已落地内置官方公告插件 `official_announcement` 的前端挂载最小闭环（Host + iframe + postMessage），用于验证前后台挂载模型（仍不执行第三方不可信代码，不支持任意远程 iframe URL）。

v1.8.2 实现补充：

- 抽取官方插件前端挂载共享 helper：`GET /plugins/assets/devhub-plugin-mount-host.js`，用于前台首页、`/c/:slug` 与后台插件详情复用统一挂载逻辑，降低复制风险并保持 SEO 红线不变。
- 新增 Host 浏览器安全 API：`/api/v1/plugins/official-announcement/context`、`/api/v1/plugins/official-announcement/audit-events` 与内置 iframe 路由：`/plugins/official-announcement/iframe`。

v1.8.3 后台治理稳定性与体验补充：

- v1.8.3-S1 第一批收敛已完成：上游过滤 `null` 插件项，保护插件详情抽屉空插件状态；插件列表首屏减负并突出 `official_announcement`；上传包详情将原始 JSON 折叠到“技术详情”；可信发布者主列表降低信息密度。该批次只调整后台 UI / IA / 中文体验，不改变插件生命周期、Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S2 已完成后台插件二级导航收敛与三级 Tab 重组：左侧只保留插件总览、插件包治理、Webhook 治理、发布者与信任、运行记录 / 审计 5 个治理域，原 20+ 入口沉到页内 Tab；旧路由继续重定向到新治理域和 `?tab=`，减少低频能力在左侧堆叠。
- v1.8.3-S4 已完成插件总览页布局与筛选体验收口：总览页修正左侧异常留白和内容偏右，健康摘要默认折叠，快捷操作压缩；插件列表筛选区改为紧凑横向筛选栏，高级筛选默认折叠；新增轻量脚本 `scripts/check-admin-plugin-ia.sh` 用于回归 5 个治理域、旧路由到 Tab、1024 宽度和中文空状态。
- v1.8.3-S5 已完成插件详情抽屉视觉 polish 与信息减负：详情抽屉可见 Tab 收敛为概览、配置、前端挂载、Webhook、安全凭据、运行记录、审计日志、技术详情；Webhook 密钥和回调 Token 合并到安全凭据；原始配置、声明 JSON、低频技术字段收纳到默认折叠的技术详情并脱敏展示。该批次只改后台 UI 和文档，不改变 API、插件生命周期、Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S6 已完成插件详情抽屉性能拆包与视觉回归：配置版本弹窗、技术详情和 JSON 编辑器按需加载，低频 Tab 使用懒渲染；1024 宽度下普通插件详情、配置 Tab、技术详情和配置版本弹窗已截图回归。剩余大 chunk 主要来自后台主入口、内容页和按需 `json-editor-vue`。
- v1.8.3-S7 已完成 `official_announcement` 浏览器回归固定 fixture：`scripts/check-admin-plugin-ia.sh` 通过后台登录和现有插件配置 / 启用 API 幂等准备官方公告插件，确保插件列表、详情概览、公告配置、前端挂载和公告预览从条件覆盖变为强制覆盖；同时修复 admin area Host context / audit 请求鉴权，token 只用于 Host 请求，不进入 iframe。fixture 不写入真实 token / Secret，不改变 Host + iframe + postMessage 协议或安全边界。
- v1.8.3-S8 已完成插件包 migrations 规范收口：插件包 dry-run / 预检 / 安装 / 升级只读取 `migrations/` 下的 `.sql` 文件并生成只读 migration plan；根目录 `001_schema.sql` 已降级为 deprecated warning，新模板 / 示例统一生成 `migrations/001_init.sql`，dry-run 不执行 SQL、不修改数据库。
- v1.8.3-S9 已完成 PluginRegistry reload 运行态刷新收口：Service 层新增统一刷新入口和运行态快照，安装、升级、启停、软卸载 / 归档、恢复、全局配置变更、子站启停、子站配置变更以及配置轮换后统一刷新；reload 成功后原子替换快照，失败时保留旧快照并写审计，不执行第三方代码、不开放动态加载、不改变 Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S10 已完成声明型插件可用闭环：manifest 安装插件的菜单、content_type、权限和配置声明进入运行态；子站启停、发布校验、权限矩阵和归档阻断均可识别声明型插件。disabled / archived / 子站 disabled 会停止新能力，历史内容和审计继续可查；仍不执行第三方代码、不开放动态加载、不改变 Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S11 已完成 external_service Webhook 运行时预备：新增外部服务配置、endpoint / timeout / failure_policy / auth_type 校验、Bearer token 加密保存与 token_ref、受控 health check、`hook_executions(service_type=external_service)` 记录、warning / error 健康状态联动以及后台详情抽屉展示。disabled / archived 插件不会调用 endpoint，只记录 skipped；本轮仍不执行第三方代码、不开放动态加载、不做 blocking Hook。
- v1.8.3-S14 已完成 external_service non-blocking Webhook 投递闭环：`service_type=external_service` 的 Hook 声明只能使用 `mode=non_blocking`，触发后异步投递并记录 `hook_executions`、retry、skipped 与 health warning/error；插件 disabled / archived / 子站 disabled 或 external_service disabled 时不调用 endpoint。blocking Hook、第三方代码执行、动态加载和远程 iframe 仍未开放。
- v1.8.3-S12 继续收口插件包 upload -> promote -> install：blocked / failed 上传包不可 promote；promote 只转入本地仓库；install 只能从本地仓库包执行，并且必须携带当前 install dry-run 计划凭证 `dry_run_id`。install 前服务端会再次 dry-run 并校验 path / plugin_code / version / manifest checksum / checksum status / migration plan hash 是否与当前包一致，upload/staging 阶段旧结果不可直接复用。
- v1.8.3-S13 已用真实 zip 插件包验收 S12 链路：新增 fixture 生成脚本和 valid / blocked / deprecated schema 三类包，后台 E2E 使用真实 Admin API 覆盖 upload、precheck、blocked promote 拒绝、promote 入本地仓库、无 install dry-run 拒绝、dry-run plan、安装、审计和后台页面状态 smoke；promote 不安装 / 不启用 / 不执行 SQL，install 只基于 `migrations/`，根目录 `001_schema.sql` 不执行。
- v1.8.3-S15 已用真实声明型友情链接插件验收“安装到使用”闭环：fixture 生成 `official_links*` 包，声明 `friend_link*` content_type、菜单、权限、配置 Schema 和 `migrations/001_init.sql`。后台 E2E 通过真实 Admin API 覆盖 upload、precheck、promote、本地仓库 install dry-run、install、PluginRegistry reload、全局启用、子站启用、菜单可见、权限矩阵、配置读写、content_type 创建、板块不允许 / 子站禁用 / 全局禁用 / 归档阻断和历史内容可读。仍不执行第三方代码、不开放动态加载、不开放远程 iframe、不做 blocking Hook。
- v1.8.3-S16 已完成后台插件治理复杂度继续压缩：5 个治理域保持不变，默认只展示高频任务；插件总览低频入口沉到“高级治理”，插件包治理改为流程型工作台，Webhook 治理聚焦投递、失败、重试和熔断，发布者与信任和运行记录 / 审计弱化技术字段，插件详情抽屉只保留当前插件摘要、能力、运行和技术详情入口。该批次只调整后台 UI / IA / 回归脚本和文档，不改变 API、插件生命周期、Webhook 协议或 Secret / Token 安全模型。
- 后台插件中心中文状态和异常提示已完成收口：状态、风险、阻断原因、Hook / 健康原因、操作名和建议文案集中在 `web/admin-app/src/modules/plugins/statusText.js`，插件列表、详情、插件包治理、上传记录、promote/install、远程索引、版本升级、配置密钥等页面复用同一口径。英文枚举和 API code 继续作为稳定机器字段，不改变业务逻辑或 API 兼容性。
- 插件后台治理页面按“一级插件模块 / 5 个治理域 / 详情三级 Tab”继续收口，5 个治理域为：插件总览、插件包治理、Webhook 治理、可信发布者、运行记录 / 审计。
- Webhook 治理页补齐空数据 / 缺字段安全默认值，避免 Events、Deliveries、Retry、Circuit Breakers、Secrets、Callback Tokens、Callback Requests 因 `null` 响应导致白屏。
- 插件详情抽屉补齐运行时依赖并保护空插件状态，避免打开详情或页面初始化时出现运行时异常。
- 插件详情抽屉将概览、配置、前端挂载、Webhook、Webhook 密钥、回调 Token、运行记录、审计日志分组展示，避免所有治理能力堆在首页。
- Webhook 治理页拆分为事件、投递记录、重试队列、熔断状态、Webhook 密钥、回调 Token、回调请求。
- 插件包治理页继续使用本地仓库、初始化插件包、上传 zip 页内 Tab，并明确远程索引、暂存区、预检、启用前检查、升级任务之间的安全边界。
- 上传记录、远程插件包、审批中心、操作历史、配置版本历史/回滚预览、可信发布者等页面补齐中文字段、中文状态、中文确认弹窗和“预检”口径，技术字段值仍按原值展示以便排障。
- 本轮只改后台信息架构与中文体验，不改插件生命周期、Webhook 协议、Secret / Token 安全模型，不开放远程 iframe、第三方代码执行、插件市场或 blocking Hook。

v1.7.3 任务拆解补充（文档阶段）：

- 已新增 Webhook / HTTP 插件服务协议的实现阶段拆解：`docs/PLUGIN_WEBHOOK_IMPLEMENTATION_PLAN.md`（non_blocking delivery 优先，blocking Hook 明确后置）。
- 已新增官方公告插件端到端验证方案：`docs/plugins/official-announcement-plugin.md`（用于验证 delivery 记录 / 重试 / 熔断 / 审计与后台治理入口；不执行第三方代码）。

下一阶段建议（实现阶段）：

- v1.7.4：Webhook non_blocking delivery 最小实现（event/delivery 持久化 + 最小投递 + 最小审计 + 最小后台查看；不做 blocking、不做复杂重试、不做熔断）。
- v1.7.5：Webhook 重试队列与熔断机制（non_blocking）。
- v1.7.6：Webhook 签名鉴权与 Secret 轮换（发送端签名 + Secret 管理/轮换 + 最小后台治理；仍仅 non_blocking）。
- v1.7.7：Webhook 插件服务回调 Core API（callback token + scopes + community scope + callback request 记录与审计）。
- v1.7.8：Webhook 后台治理与官方公告插件端到端验证（补齐 Events 视图；提供官方 mock receiver；形成端到端验证闭环）。

## v1.7.2 插件运行模型设计结论

- 推荐运行模式：Core 内置插件、外部 HTTP 服务插件、前端 iframe / sandbox 插件。
- 推荐第三方后端方式：独立 HTTP 插件服务，DevHub 通过签名请求调用插件服务，插件通过受控 Core API 回写或读取系统能力。
- 推荐第三方前端方式：iframe / sandbox 容器挂载，配合 `postMessage` 或受控 SDK 通信；默认不直接注入第三方 JS。
- 受控 API：插件 API token 需要绑定 `plugin_code`、`publisher_id`、`install_id`、`community_id`、`actor` 和 scopes；不能等同管理员 token。
- 隔离边界：插件不能直接访问数据库、读取系统配置、获取用户 token / 密钥、覆盖核心路由 / 权限、绕过 enabled / community 状态或调用未授权 API。
- 官方示例插件验证：优先公告插件或友情链接插件，再扩展 SEO 扩展、统计代码和 Webhook 通知插件。
- 本轮为插件运行模型设计任务，主要修改文档，未修改代码，未执行测试、构建或 E2E。

Webhook / HTTP 插件服务协议（设计）：

- 设计文档：`docs/PLUGIN_WEBHOOK_PROTOCOL.md`。
- 后续实现建议：先落地 non_blocking 投递（delivery 记录 + 重试队列 + 熔断），再评估可控的 blocking hook（短超时 + 明确失败策略）。

v1.7.5 实现补充：

- 已实现 non_blocking delivery 的治理能力增强：delivery 记录状态机、`retry_scheduled/retry_exhausted`、`next_retry_at` 调度、DB 扫描式 `retry-due`、以及 `plugin_code + target_url` 维度的熔断（`closed/open/half_open`）。
- 已提供后台最小治理入口：插件 → 运行时治理 → Webhook 治理（Deliveries/Circuit Breakers）。
- 仍未实现：blocking Hook、插件回调 Core API token。

v1.7.6 实现补充：

- 已实现 DevHub → 插件服务的 HMAC-SHA256 发送端签名（包含 `timestamp/method/path/body_sha256` signing string）。
- 已实现 Webhook Secret 管理与轮换窗口（active/previous + grace period），Secret 明文只在创建/轮换响应中展示一次。
- 已提供后台最小治理入口：插件 → 运行时治理 → Webhook 治理（新增 Secrets Tab），并提供对应 Admin API（`/api/v1/admin/plugins/webhooks/secrets*`）。

v1.8.3-S11 实现补充：

- 已实现 external_service 运行时预备配置与探活：`/api/v1/admin/plugins/:code/external-service` 管理 endpoint / timeout / failure_policy / auth_type / token_ref，`/health-check` 执行受控 HTTP GET 探活。
- external_service token 与 Webhook Secret、Callback Token 明确区分：external_service token 用于 DevHub 调用外部服务；Webhook Secret 用于 DevHub → 插件服务签名；Callback Token 用于插件服务 → DevHub Core 回调。三者均不在列表、详情、执行记录、日志或审计中回显明文。
- health check 结果统一进入 `hook_executions(service_type=external_service)`，并联动插件健康摘要；仍不代表远程 Hook 投递、远程代码执行或 blocking Hook 已开放。

## 历史收尾：v1.3.5 插件治理体验与安装升级向导

本节保留为 v1.3.5 草案与阶段目标追溯；当前版本线以 `v1.8.3` 的验收记录与 `docs/PROJECT_PROGRESS.md` 为准。

P0 目标状态：

1. `/admin-next/plugins` 信息架构重排：已完成按功能分页与二级导航分组（概览/列表/内容治理/安装升级/配置中心/依赖兼容/Hook 排障/事件通知/搜索索引/前台入口/权限矩阵/审计日志/开发者工具），旧路由保持兼容。
2. 完整安装向导：已完成抽屉分步流程并补最小 E2E。
3. 完整升级向导：已完成抽屉分步流程并补最小 E2E。
4. 批量归档 / 恢复体验：已完成影响预览、成功 / 失败明细表和审计跳转。
5. 状态治理页：已在 `/admin-next/plugins` 内形成异常处理视图。
6. PluginContent 体验对齐：已具备归档态提示、筛选、详情、多选、批量隐藏 / 恢复和审计入口；仍需小范围视觉和状态提示对齐。
7. 最小 E2E：已覆盖新入口和核心安全确认；历史遗留的 skipped 用例已在 v1.4.0 收口验收中恢复/替代覆盖并通过，`docs/TESTING.md` 以最新验收为准。

v1.3.5 剩余收尾需求：

1. 跑并记录最小验收命令，确认 Go、后台 build、插件治理 E2E 不退化。
2. 决定是否正式切版到 `v1.3.5`，同步 `VERSION`、README 当前版本、CHANGELOG 和 Release Notes。
3. 处理后台 E2E 中保留 skip 的两条旧长链路：删除、恢复启用，或写明替代覆盖。
4. 对 PluginContent 做小范围对齐：头部状态、禁用 / 归档提示、审计入口和筛选布局，不扩大为专属插件业务页。

## 下一阶段：v1.4 插件平台增强

P1 建议优先级：

- 前台插件入口与菜单可见性治理（navigation/create-options/menus preview + E2E + 文档）。
- 插件内容治理操作权限矩阵。
- RBAC 分配 UI 草案。
- `config_schema` 深层对象、数组和字段分组的表单体验增强。
- Hook 排障页：执行记录、失败详情、blocking / non-blocking、最近错误和手动重试入口。
- 插件模板后续增强：插件依赖解析、版本兼容检查深化、签名与包格式设计。
- Docs / Wiki 专用体验信息架构。
- Core 兼容内容类型 `article` / `news` 权限细化。

继续禁止：

- 插件市场、远程安装、在线更新、Go 动态插件、脚本沙箱、硬卸载、migration down、删除历史内容或删除审计记录。

## 下一阶段：v1.5 插件分发能力（建议范围）

定位：在不引入“远程市场/在线更新/动态执行环境”的前提下，先把“插件包规范 + 本地导入 dry-run + 安全边界 + 审计闭环”做扎实。

建议优先级（v1.5.0）：

- 插件包规范草案（目录结构、manifest 入口、版本、依赖、assets 限制、checksum/签名预留字段）。
- 插件包导入 dry-run（只做校验/影响分析与报告，不做远程下载、不自动安装依赖、不执行第三方代码）。
- 插件签名 / checksum 草案（仅定义与校验接口，先不引入完整证书体系）。
- 本地插件仓库目录规范（为未来市场做准备，但不做市场 UI）。
- 插件安装/升级审批流（轻量：审批记录 + 审计 + 执行前重新 dry-run 校验，不做复杂工作流）。
- 插件配置版本历史（diff + 回滚 dry-run 预览已落地；真实回滚与灰度不在当前范围）。
- 敏感配置加密存储（服务端加密、审计脱敏；前端仅展示脱敏值）。
- 搜索索引异步重建队列（将重建从同步请求拆出，保留可观测与错误码）。
- 插件治理 handler/service 小步拆分（降低 `router.go` / `service.go` 继续膨胀的风险）。

当前已落地（v1.5.0-P0-01 / P0-02 / P0-03 / P0-04 / P1-05 / P1-07 / P1-08 / P2-09 / P2-10）：

- 本地插件包规范草案：`docs/PLUGIN_PACKAGE.md`。
- 本地插件包 dry-run API：`POST /api/v1/admin/plugins/packages/dry-run`（白名单目录、安全扫描与预览）。
- 后台安装升级页新增“本地插件包 dry-run”区域与示例插件包 `examples/plugins/demo_notice/`。
- `checksums.json`（sha256）校验与 `risk_report` 风险报告（low/medium/high/blocked）。
- 本地插件仓库扫描：发现包列表/详情/dry-run（仓库目录建议 `storage/plugins/packages/`）。
- 本地插件包安装闭环：`POST /api/v1/admin/plugins/packages/install`（安装前强制复跑 dry-run；安装后默认 disabled）。
- 配置版本历史与回滚 dry-run 预览：保存全局/子站插件配置会写入版本记录，后台提供版本列表/详情（diff 脱敏）与回滚预览（不写入）。
- 插件安装/升级审批流：`/admin-next/plugins/approvals` + `POST /api/v1/admin/plugins/approvals`（审批通过后执行；执行前重新校验）。
- 插件治理 service/handler 拆分：插件 manifest/package/config/approval 等 handler 拆文件，生命周期 service 迁出，保持 API 兼容。
- 插件包签名与可信来源治理：`publisher.json`、`signature.json`、Ed25519 真实验签、后台可信发布者管理和签名风险联动。
- 已安装声明型插件导出：`POST /api/v1/admin/plugins/:code/export/dry-run` 与 `POST /api/v1/admin/plugins/:code/export`，输出 `storage/plugins/exports/`，生成 manifest/README/脱敏 config.example/checksums 并自动 package dry-run 自检。
- zip 上传安全沙箱与生命周期治理：`POST /api/v1/admin/plugins/packages/upload` 上传 `.zip` 到 `storage/plugins/uploads/`，安全解压到 `storage/plugins/staging/{upload_id}/` 或 blocked quarantine，复用 scanner/checksum/signature/risk_report/dry-run；`plugin_package_uploads` 记录生命周期；`GET /uploads` 列表，`GET /uploads/:upload_id` 详情，`POST /rescan` 重扫，`POST /approval` 提交导入审批，`POST /approve|reject` 审批，`POST /promote` 转入 `storage/plugins/packages/{code}/`，`POST /cancel`、`DELETE`、`POST /cleanup` 治理 staging 文件；所有动作均不安装插件。

仍明确不做 / 后置：

- Go 动态加载。
- JS/WASM/Lua 脚本沙箱。
- 远程插件市场与在线更新。
- 自动下载安装依赖。
- 第三方代码执行。

## 1. 文档目标

本文档用于规划 DevHub 插件系统的长期完善方向。

本阶段不讨论具体插件业务实现，例如 QA、Docs、Wiki、Projects、Jobs、AI Works 的具体字段和业务逻辑，而是站在“完整插件平台流程”的角度，定义 DevHub 插件系统应具备的生命周期、治理能力、后台能力、运行时能力、审计能力和测试要求。

核心目标是将当前插件系统从“内置插件治理雏形”升级为“可安装、可配置、可启停、可迁移、可审计、可观测、可扩展”的完整插件平台。

## 历史阶段记录：v1.3.4 插件迁移与 Hook 失败注入验收闭环

本节保留为 v1.3.4 任务来源追溯。当前验收与下一步目标以上方“当前实现快照与下一阶段”为准。

### 目标

在 v1.3.3 插件平台治理收口之后，本阶段优先验证“异常场景是否可靠”。本阶段不新增插件、不做插件市场、不做远程安装、不做动态加载，也不补具体业务插件功能；重点补齐插件迁移失败、Hook 失败、权限拒绝、状态恢复和 MySQL 升级路径的自动化保护。

推荐版本主题：

- `v1.3.4`：插件异常治理与验收闭环版。

### 优先级

本阶段属于 P0 插件运行治理闭环。

优先级排序：

1. 插件迁移失败注入与启用阻断。
2. HookBus blocking / non-blocking 失败注入。
3. 插件权限矩阵继续收口。
4. MySQLStore / 老库升级专项（已完成插件平台结构与核心行为验证；生产大库备份 / 回滚演练仍待后续）。
5. P1 体验增强准备。

### 范围一：插件迁移失败注入与启用阻断

目标：

- 确认 `plugin_migrations.failed` 会真实阻断插件全局启用和子站启用。
- 确认失败迁移可以 retry，并在修复后恢复启用。
- 确认后台展示、API 返回、审计记录和 E2E 都能覆盖失败与恢复链路。

需求：

1. 构造或提供测试专用 failed migration 状态。
2. 插件全局启用时，如果存在 failed migration，应返回明确错误。
3. 子站启用插件时，如果存在 failed migration，应返回明确错误。
4. 后台插件详情迁移 Tab 应展示失败迁移、错误信息和重试入口。
5. 重试成功后，迁移状态应变为 success。
6. 重试成功后，插件可重新启用。
7. `plugin.migration.run`、`plugin.migration.failed`、`plugin.migration.retry`、`plugin.migration.success` 必须写入审计。
8. E2E 或 API 测试必须覆盖失败、阻断、重试、恢复和审计定位。

验收：

- failed migration 阻断全局启用。
- failed migration 阻断子站启用。
- retry 成功后可以恢复启用。
- 已 success 的 migration 不重复破坏数据。
- 后台能看到失败原因。
- 审计日志能定位迁移操作。

### 范围二：HookBus blocking / non-blocking 失败注入

目标：

- 确认 blocking Hook 失败会阻断主流程。
- 确认 non-blocking Hook 失败不会阻断主流程，但必须可追踪。
- 确认后台 Hooks Tab 能展示最近失败、失败次数和错误信息。

需求：

1. 提供测试专用 Hook 失败注入能力，避免污染生产逻辑。
2. `BeforeCreateContent` 或等价 blocking Hook 返回错误时，内容创建必须失败。
3. blocking Hook 失败应写入 `hook_executions`。
4. blocking Hook 失败应写入 `plugin.hook.blocked` 审计。
5. `AfterCreateContent` 或等价 non-blocking Hook 返回错误时，内容创建仍应成功。
6. non-blocking Hook 失败应写入 `hook_executions`。
7. non-blocking Hook 失败应写入 `plugin.hook.failed` 审计。
8. 后台 Hooks Tab 应展示执行次数、失败次数、最近执行、最近失败、平均耗时和最近错误。
9. Search / Notification / SEO Hook 当前可以继续作为最小事件派发，但状态必须真实标注，不得伪造完整业务处理器。

验收：

- blocking Hook 失败时主流程被阻断。
- non-blocking Hook 失败时主流程继续。
- 两类失败都可在 Hook 执行记录中查询。
- 两类失败都有审计或错误日志。
- 后台 Hooks Tab 能看到失败摘要。

### 范围三：插件权限矩阵继续收口

目标：

- 继续弱化 `post.create` 的历史兼容地位。
- 确认内容创建、后台创建、版主菜单和插件内容治理都按插件权限码判断。

需求：

1. 内容创建权限继续来自 `ContentTypeDefinition.create_permission`。
2. `question` 必须使用 `qa.question.create`。
3. `document` 必须使用 `docs.document.create`。
4. `wiki_page` 必须使用 `wiki.page.create`。
5. `project` 必须使用 `projects.project.create`。
6. `job` 必须使用 `jobs.job.create`。
7. `ai_work` 必须使用 `ai_works.work.create`。
8. `article` / `news` 暂继续使用 `core.topic.create`，`post.create` 仅作为兼容桥。
9. 后台 `admin/posts` 创建入口必须叠加真实插件 create 权限。
10. 版主插件菜单必须同时受全局插件状态、子站插件状态、community scope 和权限码约束。
11. 普通用户 token 不能调用插件治理 API。

验收：

- 缺少对应插件 create 权限时不能发布。
- 缺少对应插件权限时后台创建失败。
- 非授权版主不能看到或操作其他子站插件菜单。
- `post.create` 不再被描述为长期主权限。

### 范围四：MySQLStore / 老库升级专项

目标：

- 确认插件平台在 MySQLStore 和老库升级场景下与 MemoryStore 行为一致。

需求：

1. 校验 `plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、`admin_logs` 结构在新装和升级场景一致。
2. 校验 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 在老库升级后可用。
3. 校验 migration 文件编号无冲突。
4. 校验 MySQLStore 下插件启停、配置校验、迁移记录、Hook 执行记录和审计写入可用。
5. 文档补充老库升级顺序、备份、回滚和风险。

验收：

- MySQLStore 下全局禁用插件后不能新建对应内容。
- MySQLStore 下子站禁用插件后仅该子站不能新建对应内容。
- MySQLStore 下 failed migration 阻断启用。
- MySQLStore 下 Hook 执行记录可查询。
- MySQLStore 下插件治理审计可查询。

### 范围五：P1 体验增强准备

本阶段只做规划和接口边界确认，不做大范围 UI 或业务实现。

后续 P1 可进入：

- `config_schema` 自动表单增强。
- 插件 SDK / 模板。
- 插件内容治理页更多批量操作。
- Docs / Wiki 专用体验。
- 插件搜索 / 通知 / SEO 扩展。

### 不做内容

本阶段不做：

- 插件市场。
- 插件上传安装。
- 远程插件安装或在线更新。
- Go 动态插件加载。
- 第三方插件沙箱。
- 新增业务插件。
- Projects / Jobs / AI Works 专属业务闭环。
- Docs / Wiki 复杂编辑体验。
- 大规模 UI 重构。
- 删除 `sites/posts` 兼容读取。
- 重命名 `topics -> contents` 或 `categories -> channels`。

### 文档要求

执行本阶段任务时必须同步更新：

- `docs/PROJECT_PROGRESS.md`
- `docs/PLUGIN_ARCHITECTURE.md`
- `docs/API.md`
- `docs/TESTING.md`
- 对应 Release Notes
- `CHANGELOG.md`

文档必须继续区分：

- 已完成。
- 部分完成。
- 预留。
- 后续规划。
- 未执行测试。
- 跳过项及原因。

### 最低检查要求

如果修改 Go 后端：

- `gofmt`
- `go test ./...`
- `go build -o .devhub/devhub .`

如果修改后台：

- `docker compose run --rm admin-e2e npm run build`

如果修改前台：

- `docker compose run --rm frontend-e2e npm run build`

如果新增或修改 E2E：

- `./scripts/check-frontend.sh --quick`
- `./scripts/check-frontend.sh --admin-only`
- `./scripts/check-frontend.sh --frontend-only`，或说明无法执行原因。

## 2. 总体目标

DevHub 插件系统的最终形态应覆盖完整插件生命周期：

注册 -> 安装 -> 迁移 -> 配置 -> 授权 -> 启用 -> 运行 -> 监控 -> 审计 -> 升级 -> 禁用 -> 卸载

系统需要保证：

1. 插件来源可识别。
2. 插件安装可追踪。
3. 插件配置可校验。
4. 插件权限可管理。
5. 插件启停可控。
6. 插件 Hook 可观测。
7. 插件迁移可执行。
8. 插件内容可治理。
9. 插件操作可审计。
10. 插件异常可定位。
11. 插件升级可演进。
12. 插件禁用后历史内容仍然安全可访问。

## 3. 插件系统设计原则

### 3.1 不绑定具体插件业务

插件系统应优先完善平台流程，而不是继续堆叠具体插件业务。

本阶段不重点实现：

- QA 具体问答业务增强。
- Docs 具体文档业务增强。
- Wiki 具体协作业务增强。
- Projects 专属字段。
- Jobs 专属字段。
- AI Works 专属字段。
- 第三方插件市场。

本阶段重点实现：

- 插件生命周期。
- 插件状态管理。
- 插件配置管理。
- 插件权限管理。
- 插件 Hook 治理。
- 插件迁移治理。
- 插件审计。
- 插件影响分析。
- 插件内容通用治理。
- 插件 E2E 回归验证。

### 3.2 前端隐藏不是权限控制

任何插件入口、菜单、按钮的隐藏都只是用户体验优化。

真正的安全边界必须在后端完成，包括：

- `content_type` 合法性校验。
- 插件全局启停校验。
- 插件社区启停校验。
- 用户权限校验。
- 版主权限边界校验。
- API 越权访问拦截。
- Hook 阻断能力。

### 3.3 插件禁用不应破坏历史内容

插件禁用后，应禁止新建相关内容，但默认不应破坏已有历史内容。

推荐默认策略：

- 禁止新建该插件内容。
- 前台发布入口隐藏或禁用。
- 后端强传 `content_type` 必须失败。
- 历史内容详情页仍可访问。
- SEO 基础信息保留。
- 历史内容是否允许评论、点赞、编辑、搜索，由配置决定。

## 4. 插件生命周期需求

### 4.1 插件发现 / 注册

当前系统已有内置插件注册表。后续应抽象为统一插件注册机制。

插件来源分为：

- 内置插件：随 DevHub 系统发布，通过代码 registry 注册，默认由系统维护。
- 外部插件：通过 `manifest.json` 声明，可上传、导入或扫描，可安装、升级、禁用、卸载。

插件注册信息至少包含：

- `plugin_code`
- `name`
- `version`
- `description`
- `author`
- `homepage`
- `license`
- `compatible_core_version`
- `content_types`
- `permissions`
- `menus`
- `routes`
- `hooks`
- `config_schema`
- `migrations`
- `assets`
- `dependencies`
- `status`

### 4.2 插件安装

插件不应因为出现在 registry 中就默认等于可运行。

插件状态应拆分为：

- `discovered`：系统识别到插件。
- `installed`：插件已安装。
- `migrated`：插件迁移已完成。
- `configured`：插件配置已通过校验。
- `enabled`：插件已启用。
- `running`：插件处于可运行状态。

安装流程应包括：

1. 读取 manifest。
2. 校验 `plugin_code` 唯一性。
3. 校验插件版本兼容性。
4. 校验依赖插件是否存在并满足版本要求。
5. 写入 `plugin_installations`。
6. 初始化默认配置。
7. 注册权限。
8. 注册菜单。
9. 注册路由。
10. 生成待执行 migration。
11. 写入审计日志。

### 4.3 插件迁移

插件可能需要自己的数据表、字段、索引和初始化数据，因此必须具备迁移系统。

迁移流程：

1. 插件安装或升级。
2. 系统检查插件 migrations。
3. 后台展示待执行迁移。
4. 管理员执行迁移。
5. 系统记录执行结果。
6. 执行失败时允许查看错误和重试。
7. 后续预留 rollback 能力。

迁移记录至少包含：

- `plugin_code`
- `migration_version`
- `migration_name`
- `direction`
- `status`
- `started_at`
- `finished_at`
- `duration_ms`
- `error_message`
- `checksum`
- `executor`
- `rollback_supported`

第一阶段要求：

- 支持 up migration。
- 支持迁移执行记录。
- 支持失败记录。
- 支持失败重试。
- 支持迁移审计。

第二阶段再考虑：

- migration down。
- 回滚。
- 迁移前备份。
- 迁移影响分析。

## 5. 插件启用 / 禁用需求

插件启停应区分：

1. 全局启用 / 禁用。
2. 社区级启用 / 禁用。
3. 用户权限级可用性。
4. 菜单显示 / 隐藏。
5. 内容创建允许 / 禁止。
6. 历史内容访问策略。
7. 搜索索引策略。
8. SEO 策略。

### 5.1 全局禁用规则

全局禁用插件后：

- 所有社区不能新建该插件内容。
- 后台显示插件禁用状态。
- 前台发布入口隐藏或不可用。
- 强传 `content_type` 必须被后端拦截。
- 历史内容详情页仍可访问。
- SEO 基础信息不丢。
- 搜索是否展示由插件配置决定。

### 5.2 社区禁用规则

社区禁用插件后：

- 当前社区不能新建该插件内容。
- 当前社区发布入口隐藏或不可用。
- 当前社区强传 `content_type` 必须被后端拦截。
- 其他社区不受影响。
- 当前社区历史内容仍可访问。

### 5.3 禁用前影响分析

插件禁用前必须展示影响范围，而不是只显示静态提醒。

影响范围至少包括：

- 已有内容数量。
- 绑定社区数量。
- 启用社区数量。
- 禁用社区数量。
- 绑定分类数量。
- 最近发布数量。
- 待审核数量。
- 相关菜单数量。
- 相关配置数量。
- 可能影响的前台入口。
- 是否存在待执行迁移。
- 是否存在近期 Hook 错误。

后台禁用弹窗应展示真实影响范围，不得伪造统计数据。

## 6. 插件配置管理需求

插件配置不应长期停留在 JSON 文本框阶段。

完整配置流程：

1. 插件通过 `config_schema` 声明配置结构。
2. 后台根据 schema 展示配置表单或 JSON 编辑器。
3. 前端进行基础格式校验。
4. 后端进行强校验。
5. 保存配置版本。
6. 生成配置 diff。
7. 写入审计日志。
8. 运行时读取 effective config。

### 6.1 配置层级

插件配置应支持多层合并：

系统默认配置 -> 插件默认配置 -> 全局插件配置 -> 社区插件配置 -> 最终运行配置 `effective_config`

后台应能展示：

- 默认配置。
- 全局配置。
- 社区覆盖配置。
- 最终生效配置。
- 配置来源。
- 配置差异。

### 6.2 配置能力分阶段

第一阶段：

- JSON 编辑。
- 后端 schema 校验。
- 配置保存。
- 配置 diff。
- 审计记录。

第二阶段：

- 自动表单增强。
- 字段标题。
- 字段说明。
- 默认值展示。
- 字段分组。
- 敏感字段隐藏。

第三阶段：

- 配置版本回滚。
- 配置模板。
- 配置变更预览。
- 配置灰度生效。

## 7. 插件权限管理需求

插件权限不应只靠硬编码。

插件应通过 manifest 声明自己的权限，例如：

- `plugin.{code}.create`
- `plugin.{code}.update`
- `plugin.{code}.delete`
- `plugin.{code}.moderate`
- `plugin.{code}.manage`
- `plugin.{code}.configure`

平台级插件权限包括：

- 插件安装权限。
- 插件卸载权限。
- 插件启用权限。
- 插件禁用权限。
- 插件配置权限。
- 插件迁移权限。
- 插件审计查看权限。
- 插件 Hook 查看权限。
- 插件内容管理权限。

权限流程：

1. 插件声明 permissions。
2. 插件安装时注册权限。
3. 角色系统可分配权限。
4. 前端菜单根据权限显示。
5. 后端 API 强制校验权限。
6. 操作写入审计日志。

必须保证：

- 菜单隐藏不等于权限控制。
- 后端 API 必须校验权限。
- E2E 必须覆盖越权访问。
- 普通用户 token 不能访问后台插件治理接口。
- 非授权版主不能管理其他社区插件内容。

## 8. 插件菜单与路由需求

插件系统应统一管理：

- 前台菜单。
- 后台菜单。
- 社区菜单。
- 用户中心菜单。
- 版主工作台菜单。

插件 manifest 可声明 admin menus：

- `path`
- `title`
- `icon`
- `permission`
- `order`
- `plugin_code`

插件 manifest 可声明 frontend menus：

- `path`
- `title`
- `community_scope`
- `permission`
- `order`
- `plugin_code`

菜单渲染流程：

1. 插件启用。
2. 菜单注册。
3. 根据全局插件状态过滤。
4. 根据社区插件状态过滤。
5. 根据用户权限过滤。
6. 前端渲染菜单。

必须避免：

- 菜单显示但路由不可访问。
- 菜单隐藏但直接访问路由仍可操作。
- 插件禁用但入口仍然存在。
- 社区禁用但当前社区还能发布内容。

E2E 必须覆盖：

- 插件启用时入口显示。
- 插件禁用时入口隐藏或不可用。
- 直接访问路由。
- 直接调用 API。

## 9. 插件 Hook 运行时需求

Hook 是插件扩展系统能力的核心。

完整 Hook 流程：

1. 业务事件发生。
2. 平台构造 HookContext。
3. 查找当前启用插件。
4. 过滤声明了对应 Hook 的插件。
5. 校验插件全局状态和社区状态。
6. 按优先级执行 Hook。
7. blocking Hook 可以阻断主流程。
8. non-blocking Hook 失败不阻断主流程。
9. 记录 Hook 执行日志。
10. 返回结果或继续主流程。

### 9.1 Hook 类型

blocking Hook：

- `BeforeCreateContent`
- `BeforeUpdateContent`
- `BeforeDeleteContent`
- `BeforeModerateContent`

non-blocking Hook：

- `AfterCreateContent`
- `AfterUpdateContent`
- `AfterCreateComment`
- `OnSearchIndex`
- `OnNotificationBuild`
- `OnSEOBuild`

### 9.2 Hook 治理能力

必须记录：

- `hook_name`
- `plugin_code`
- `mode`
- `content_id`
- `actor_id`
- `community_id`
- `request_id`
- `started_at`
- `finished_at`
- `duration_ms`
- `success`
- `error_message`
- `blocking`
- `metadata`

后台插件详情应展示：

- Hook 声明。
- Hook 是否启用。
- 最近执行时间。
- 最近失败时间。
- 最近错误。
- 执行次数。
- 失败次数。
- 平均耗时。
- 失败率。
- blocking / non-blocking 类型。

## 10. 插件内容类型流程需求

插件通过 `content_type` 接入 DevHub 内容系统。

完整流程：

1. 插件声明 `content_types`。
2. 平台注册 `content_type`。
3. 发布页根据启用状态展示可发布类型。
4. 用户选择 `content_type`。
5. 后端根据 `content_type` 找到所属 `plugin_code`。
6. 校验 `content_type` 是否存在。
7. 校验插件是否全局启用。
8. 校验当前社区是否启用。
9. 校验用户是否有创建权限。
10. 执行 `BeforeCreateContent` Hook。
11. 创建内容。
12. 执行 `AfterCreateContent` Hook。

必须覆盖：

- 合法 `content_type` 创建成功。
- 非法 `content_type` 创建失败。
- 插件禁用后 `content_type` 创建失败。
- 社区禁用后 `content_type` 创建失败。
- 历史内容详情不受插件禁用影响。
- 前端入口隐藏不能替代后端强校验。

## 11. 插件前台流程需求

前台应根据以下因素动态渲染插件能力：

- 插件是否全局启用。
- 当前社区是否启用。
- 当前用户是否有权限。
- 当前 `content_type` 是否允许发布。
- 插件配置是否影响展示。
- 插件菜单是否允许显示。

### 11.1 发布页流程

发布页流程：

1. 进入发布页。
2. 加载当前社区启用插件。
3. 加载可发布 `content_type`。
4. 用户选择内容类型。
5. 前端展示对应字段。
6. 用户提交。
7. 后端再次校验插件状态、社区状态和权限。
8. 创建成功或返回错误。

### 11.2 历史内容详情流程

历史内容详情流程：

1. 用户访问历史内容详情页。
2. 系统根据 `content_type` 找到插件。
3. 即使插件禁用，也允许历史内容展示。
4. SEO 基础元素正常。
5. 评论、点赞、收藏、搜索展示等能力由配置决定。

需要明确产品规则：

- 插件禁用后历史内容能否评论？
- 插件禁用后历史内容能否点赞？
- 插件禁用后历史内容能否被搜索？
- 插件禁用后历史内容能否编辑？
- 插件禁用后后台是否可继续管理历史内容？

推荐默认规则：

- 历史详情可访问。
- SEO 保留。
- 新建禁止。
- 评论和点赞默认保留。
- 编辑是否允许由配置控制。
- 搜索展示由配置控制。

## 12. 插件后台治理需求

后台插件中心不应只是插件列表，而应支持完整治理。

后台插件中心应包含：

1. 插件列表。
2. 插件详情。
3. 插件配置。
4. 插件影响范围。
5. 插件内容管理。
6. 插件 Hook 运行状态。
7. 插件审计。
8. 插件迁移。
9. 插件运行健康状态。

插件详情页建议 Tab：

- 基础信息。
- 内容类型。
- 权限。
- 菜单。
- 路由。
- 配置。
- 影响范围。
- Hook。
- 迁移。
- 审计。
- 运行状态。

### 12.1 插件内容管理页

建议提供统一插件内容管理页：

- `/admin-next/plugin-content/:plugin_code`

也可以继续兼容：

- `/admin-next/qa`
- `/admin-next/docs`
- `/admin-next/wiki`
- `/admin-next/projects`
- `/admin-next/jobs`
- `/admin-next/ai-works`

通用插件内容页应支持：

- 插件内容列表。
- `plugin_code` 展示。
- `content_type` 展示。
- 子站筛选。
- 状态筛选。
- 关键词搜索。
- 查看详情。
- 隐藏 / 恢复。
- 审核通过 / 拒绝。
- 置顶 / 取消置顶。
- 加精 / 取消加精。
- 批量操作。
- 操作后审计日志跳转。

## 13. 插件审计需求

所有插件治理动作必须写入审计日志。

必须审计：

- 插件安装。
- 插件卸载。
- 插件启用。
- 插件禁用。
- 插件配置修改。
- 社区插件启用。
- 社区插件禁用。
- 社区插件配置修改。
- 插件迁移执行。
- 插件迁移失败。
- 插件 Hook 失败。
- 插件内容管理操作。
- 插件权限修改。
- 插件菜单修改。
- 插件升级。
- 插件降级。
- 插件恢复。

审计字段建议包含：

- `actor_id`
- `actor_role`
- `action`
- `target_type`
- `target_id`
- `plugin_code`
- `community_id`
- `old_value`
- `new_value`
- `metadata_json`
- `request_id`
- `created_at`

后台插件详情页应提供“审计”Tab，默认按 `plugin_code` 过滤相关审计记录。

## 14. 插件监控与健康状态需求

插件不应只有 `enabled` / `disabled` 两种状态。

建议增加健康状态：

- `healthy`
- `warning`
- `error`
- `disabled`
- `migration_pending`
- `config_invalid`
- `dependency_missing`

状态来源：

- 配置是否有效。
- 迁移是否完成。
- 依赖是否满足。
- Hook 是否连续失败。
- 最近运行是否报错。
- 插件是否兼容当前 DevHub core 版本。

后台应展示：

- 插件运行状态。
- 配置状态。
- 迁移状态。
- Hook 状态。
- 最近错误。
- 建议操作。

## 15. 插件升级需求

当前已支持 manifest + 配置型插件的最小升级预览和执行闭环。后续要把它扩展为完整升级向导、升级影响分析、版本变更审计和插件包升级流程。

升级流程：

1. 检测新版本。
2. 检查核心版本兼容性。
3. 检查依赖变化。
4. 展示变更说明。
5. 展示影响范围。
6. 备份当前配置。
7. 执行新迁移。
8. 刷新 manifest。
9. 校验权限、菜单、路由变化。
10. 更新插件版本。
11. 写入审计日志。

升级可能影响：

- 新增权限。
- 删除权限。
- 新增配置项。
- 删除配置项。
- 迁移数据。
- Hook 行为变化。
- 前台入口变化。
- 后台菜单变化。

升级前必须做 impact preview。

## 16. 插件卸载需求

卸载比禁用危险，应谨慎设计。

推荐先支持软卸载，不急于实现硬卸载。

### 16.1 软卸载

软卸载行为：

- 禁用插件。
- 保留配置。
- 保留内容。
- 保留审计。
- 不删除数据表。
- 不删除迁移记录。

### 16.2 硬卸载

硬卸载属于危险操作，应满足：

- 二次确认。
- 输入 `plugin_code` 确认。
- 有权限校验。
- 有数据备份提示。
- 有 migration down 支持。
- 有完整审计。

硬卸载前必须明确：

- 是否删除配置。
- 是否删除内容。
- 是否删除插件表。
- 是否删除迁移记录。
- 是否保留审计。
- 是否保留历史内容。

## 17. E2E 测试需求

插件系统必须有完整 E2E 回归保护。

### 17.1 插件启停

- 插件启用后入口显示。
- 插件禁用后入口隐藏或不可用。
- 全局禁用后所有社区不可新建。
- 社区禁用后仅当前社区不可新建。
- 禁用后恢复原状态。

### 17.2 content_type 强校验

- 合法 `content_type` 创建成功。
- 非法 `content_type` 创建失败。
- 插件禁用后强传 `content_type` 创建失败。
- 社区禁用后强传 `content_type` 创建失败。

### 17.3 历史内容

- 插件禁用后历史内容详情仍可访问。
- SEO title / h1 / article 基础元素仍存在。
- 历史内容不会因为插件禁用而 404。

### 17.4 权限边界

- 普通用户不能访问插件后台。
- 普通用户不能调用插件治理 API。
- 非授权版主不能操作其他社区插件内容。
- 后台不同角色权限隔离。

### 17.5 插件配置

- 合法配置保存成功。
- 非法 JSON 保存失败。
- 不符合 schema 的配置保存失败。
- 配置修改写入审计。

### 17.6 Hook 治理

- blocking Hook 失败时主流程被阻断。
- non-blocking Hook 失败时主流程继续。
- Hook 失败被记录。
- 后台能看到 Hook 错误。

## 18. 分阶段实施优先级

### P0：插件运行治理闭环

优先完成：

- 插件状态模型。
- 插件影响范围。
- 插件配置校验口径统一。
- 插件启停强拦截。
- 插件审计。
- Hook 执行记录。
- E2E 覆盖 `content_type` 强校验。

目标：确保插件可控、可查、可追踪、可恢复。

### P1：插件内容管理闭环

优先完成：

- 统一插件内容管理页。
- 子站筛选。
- 状态筛选。
- 关键词搜索。
- 查看详情。
- 隐藏 / 恢复。
- 审核操作。
- 审计联动。
- 禁用后历史内容策略。

目标：让后台可以真正治理插件内容。

### P2：插件迁移闭环

优先完成：

- 迁移列表。
- 待执行迁移。
- 执行迁移。
- 失败记录。
- 重试。
- 日志。
- 审计。

目标：为插件安装和升级打基础。

### P3：插件配置体验

优先完成：

- schema 自动表单增强。
- 配置 diff。
- effective config。
- 敏感字段。
- 配置版本。
- 配置回滚。

目标：让插件配置从 JSON 文本框升级为可用配置中心。

### P4：插件生态

后续考虑：

- manifest 导入。
- 上传包生命周期治理。
- 插件依赖解析。
- 插件版本升级。
- 插件软卸载 / 硬卸载。
- 插件市场。

目标：支持真正外部插件生态。

## 19. 推荐目标架构

### Plugin Registry

职责：

- 识别插件。
- 读取 manifest。
- 管理插件元信息。
- 管理插件版本。

### Plugin Installer

职责：

- 安装插件。
- 升级插件。
- 禁用插件。
- 卸载插件。
- 检查依赖。
- 检查兼容性。

### Plugin Config Manager

职责：

- `config_schema` 校验。
- 配置合并。
- effective config 生成。
- 配置 diff。
- 配置审计。

### Plugin Permission Manager

职责：

- 插件权限注册。
- 角色绑定。
- 后端强校验。
- 权限审计。

### Plugin Runtime

职责：

- 插件状态判断。
- `content_type` 校验。
- 菜单过滤。
- 路由过滤。
- Hook 派发。
- 运行时配置读取。

### Plugin Migration Runner

职责：

- 执行插件迁移。
- 记录迁移状态。
- 失败重试。
- 回滚预留。
- 迁移审计。

### Plugin Audit

职责：

- 插件治理操作留痕。
- 插件内容操作留痕。
- Hook 错误留痕。
- 配置变更留痕。
- 迁移执行留痕。

### Plugin Admin UI

职责：

- 插件列表。
- 插件详情。
- 插件配置。
- 插件影响范围。
- 插件迁移。
- 插件 Hook。
- 插件审计。
- 插件内容管理。
- 插件健康状态。

### Plugin E2E

职责：

- 验证插件启停。
- 验证配置校验。
- 验证后端强拦截。
- 验证历史内容访问。
- 验证权限边界。
- 验证审计记录。
- 验证 Hook 治理。

## 20. 验收标准

插件系统完善后，应满足：

1. 插件有完整状态生命周期。
2. 插件可以安装、启用、禁用、配置和审计。
3. 插件配置有后端 schema 校验。
4. 插件启停影响范围可查看。
5. 插件禁用后前台入口不可用。
6. 插件禁用后强传 `content_type` 被后端拦截。
7. 插件禁用后历史内容仍可访问。
8. 插件操作写入审计。
9. Hook 执行有记录、耗时和错误信息。
10. blocking Hook 可以阻断主流程。
11. non-blocking Hook 失败不影响主流程但可追踪。
12. 插件内容有统一后台治理页面。
13. 插件迁移可执行、失败可查看、可重试。
14. 插件权限可以注册、分配和后端强校验。
15. 插件后台页面能展示影响范围、审计、Hook、迁移和健康状态。
16. E2E 覆盖插件启停、配置、强拦截、历史内容、权限边界和审计。
17. 文档清楚区分已完成能力和后续能力。

## 21. 总结

DevHub 插件系统后续不应继续围绕某个具体插件功能打补丁，而应围绕完整插件生命周期建设平台能力。

核心方向是：

- 安装可追踪。
- 配置可校验。
- 启停可控制。
- 权限可管理。
- Hook 可观测。
- 迁移可执行。
- 内容可治理。
- 操作可审计。
- 异常可定位。
- 升级可演进。

当这套插件平台流程补齐后，QA、Docs、Wiki、Projects、Jobs、AI Works 以及未来第三方插件，都可以统一落到同一套治理体系上。

## v1.4.0-P1-07 完成记录：依赖检查与版本兼容矩阵

已完成：结构化 `dependencies`、轻量版本约束、required / optional 阻断规则、Core 兼容矩阵、循环依赖检查、validate / dry-run / install / upgrade / enable 复用检查、后台安装 / 升级向导展示和插件详情 Dependencies 区域。

后续保留：自动安装依赖、插件市场推荐、远程依赖下载、依赖图大屏、插件签名、动态加载、脚本沙箱、复杂 semver / npm constraint、migration down 和 hard uninstall。

## v1.6.0-P0-04 远程插件索引只读镜像

已落地：

- 静态 `index.json` 规范与示例。
- 后台远程索引源配置、拉取和只读展示。
- SSRF 防御、响应大小限制、拉取超时和 JSON schema 校验。
- 远程插件列表 / 详情，展示 publisher trust、Core 兼容性、本地安装状态和风险提示。

仍后置：

- 远程插件包下载。
- 远程安装 / 在线更新。
- 远程可信源自动同步。
- 插件市场交易 / 评论 / 排行。
- 动态加载或第三方代码执行。

下一步建议：v1.6.0-P0-05 插件包版本仓库与升级差异对比增强。

## v1.6.0-P0-05 插件包版本仓库与升级差异对比

已新增版本仓库视图与升级 diff 能力：聚合 installed / local_package / uploaded_package / remote_index 版本，支持 `x.y.z` / `vX.Y.Z` 比较，并在升级前展示 manifest、权限、菜单、路由、配置 schema、依赖、Hook、迁移和风险差异。

仍不支持自动升级、远程下载安装、远程市场、动态加载、脚本沙箱、第三方代码执行、外部 SQL 或自动回滚。下一步建议进入 `v1.6.0-P0-06`：插件包安装 / 升级回滚保护与失败恢复。

## v1.6.0-P1-09 插件治理 UI 分页与测试基建整理

已完成后台插件治理二级导航收束、旧路由兼容、公共展示组件抽取、前端插件 API facade、导航 E2E helper 与 `plugin-governance-pages.spec.js`。本轮没有新增插件底层能力，也没有改变安装、升级、审批、签名、远程索引、密钥轮换或权限语义。

遗留技术债：插件包 fixture 仍分散在后端 testdata 与后台 E2E fixtures 中，部分大型插件页面仍可继续拆分为更小的业务组件；建议在 v1.6 总验收前继续整理 fixture 生成脚本和插件专项 E2E 分类。

下一步建议：`v1.6.0-P1-10`：v1.6 插件包上传与分发前置能力总验收。
