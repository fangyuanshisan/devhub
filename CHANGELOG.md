# Changelog

## Unreleased

### Documentation

- Move release-facing docs from v1.9.0 release wrap-up to post-archive planning: current stable version remains `v1.9.0` 官方插件生态稳定版, archival and targeted smoke are complete with P0=0 / P1=0 / P2=0, tagging is recommended after committing docs, v1.9.0 receives no new features, `v1.9.1` is a maintenance candidate, and `v1.10.0` is a planning stage for larger governance / architecture work. This is documentation-only and does not change runtime code, API, schema, Store, Webhook, SecretCenter, external_service, SEO, frontend, admin implementation, or plugin security boundaries.

## v1.9.0 (2026-05-30)

### Changed

- v1.9.0-S11：完成发布归档与发布后 smoke。确认 `VERSION=v1.9.0`，Release Notes / CHANGELOG / README / PROJECT_PROGRESS / TESTING 已同步官方插件生态稳定版口径、S1-S10 acceptance、`P0/P1/P2=0`、freeze / tag recommendation、S11 post-release smoke 和下一版本建议；修正插件架构文档中旧 `v1.8.3` 当前版本摘要。S11 执行 `git status --short`、`./scripts/check-frontend.sh --admin-only --quick`、`bash scripts/check-admin-plugin-ia.sh`、`git diff --check` 和发布一致性 grep；后台 quick 与 IA smoke 均通过，MySQLStore full smoke 复用 S9 证据，因为 S11 未修改 Store、migration、DB schema、external_service、SecretCenter 或 Webhook 投递逻辑。本轮不新增功能、不改 API/schema/权限/SEO/插件运行时，不开放第三方代码执行、Go plugin、JS sandbox、WASM、Lua、远程 iframe、remote component、插件市场、远程在线安装、blocking Hook、package scripts、自动安装或自动启用。
- v1.9.0-S10：完成发布候选总验收与冻结判断。复核 v1.9.0 最终能力清单、S1-S9 记录、MySQLStore 生产模式结果、官方插件生态稳定版、插件包治理、模板生成、upgrade dry-run UI、后台 IA、SecretCenter / `token_ref` / external_service / HTTP Allowlist、敏感扫描归档、README / VERSION / Release Notes / PROJECT_PROGRESS / TESTING / CHANGELOG 一致性和 P0/P1/P2 状态。S10 执行 `git diff --check`、Docker `gofmt`、`bash -n scripts/check-admin-plugin-ia.sh`、`bash -n scripts/run-feishu-webhook-flow.sh`、Docker `go test ./...`、Docker `go build -buildvcs=false ./...` 和 `./scripts/check-frontend.sh --admin-only --quick`；Go 首次默认代理下载 EOF，按项目代理重试后通过。最终 P0=0、P1=0、P2=0，建议冻结并人工复核后创建 tag `v1.9.0`。本轮不新增功能、不改 API/schema/权限/SEO/插件运行时，不开放第三方代码执行、Go plugin、JS sandbox、远程 iframe、remote component、插件市场、远程在线安装、blocking Hook、package scripts、自动安装或自动启用。
- v1.9.0-S9：完成 MySQLStore 生产模式总验收。执行 `./dev.sh start --mysql`，并以 MySQLStore + `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18084` 复核官方插件生态稳定版核心链路：插件安装 / 配置 / 状态持久化抽查、SecretCenter `token_ref`、external_service health / success / fail / timeout / manual retry、`hook_executions`、`admin_logs`、HTTP Allowlist 增删与拒绝矩阵、插件包 cleanup preview / cleanup、模板 preview / generate / export、upgrade dry-run safe / warning / blocked、Webhook Secret create / rotate / disable / revoke、Callback Token create / callback API、后台 IA 点击矩阵和敏感信息扫描。验收中修复 Webhook Secret 治理 API `secret_record` 回显 `secret_hash` 的脱敏缺口，响应统一清除 `secret_hash/secret_ciphertext`，DB 内部哈希仍保留用于验签。S9 不新增 API、不改 schema、不改变 upload / precheck / promote / install / upgrade 语义、不改变 SecretCenter 加密或 Webhook Secret / Callback Token 明文一次性返回安全模型；仍不执行第三方代码、不开放 Go plugin、JS sandbox、远程 iframe、remote component、插件市场、远程在线安装或 blocking Hook。P0/P1 为 0；P2 为 upgrade warning 真 confirm / confirm_token 过期 / upgrade 真执行未实跑、当前进程未实跑缺 key readiness、cleanup installed/enabled 禁删未做破坏性删除尝试。
- v1.9.0-S8：稳定化后台插件治理点击矩阵。`scripts/check-admin-plugin-ia.sh` 继续覆盖插件列表、安装与升级、Webhook 与回调、安全与发布者、运行与审计 5 个治理域，并新增系统设置 / SecretCenter / HTTP Allowlist / 当前生效配置截图和旧路由断言；截图归档为 `.devhub/screenshots/plugin-ia/1024-*.png` 与 `1366-*.png`。系统设置 Tab 现在同步 URL query，并新增 `/system/effective-config`、`/system/secret-center`、`/system/http-allowlist` 等兼容入口。配置密钥页面去除直接展示 `enc:v*` 与 root key env 名的提示，改为中文说明。本轮只做后台 IA / 脚本 / 文案稳定化，不新增 API、不改 schema、不改变插件包、Webhook、SecretCenter 或升级语义；仍不执行第三方代码、不开放 blocking Hook、远程运行时、插件市场或远程在线安装。
- v1.9.0-S7：补强后台 upgrade dry-run 结构化结果展示。`/admin-next/plugins/install` 的 upgrade dry-run / upgrade 向导继续消费既有结构化响应，但顶部现在前置 `blocked_items` / `warnings`，summary 补齐插件编码、Core 兼容、风险、确认、回滚边界和影响能力；`diff_sections` 按 section 展示 before/after；migration / config / permission / content_type / hook / frontend_mount 分区计划、`dependency_diff` 和 `frontend_mount_plan` 均改为可读表格 / 卡片；失败区展示 `failure_stage` / `failure_reason` / request id / next step，并提示查 `admin_logs` / `hook_executions`。本轮只做 UI review/fixup，不新增 API、不改 schema、不改变 upload / precheck / promote / install / upgrade / approval 语义；仍不执行第三方代码、不开放 blocking Hook 或远程运行时，secret diff 继续只显示脱敏值。
- v1.9.0-S6：稳定化插件模板生成链路。后台 preview / generate / export zip 和 CLI `plugin:new` 统一复用 `internal/plugins/scaffold.PluginTemplateGenerator`；CLI 补齐 `plugin_type`、前端挂载和 external_service 参数，并按插件类型默认 Hook / migration 开关。模板输出收口为声明型文件、`docs/*.md` 和 `migrations/001_init.sql`，不再生成 `registry.example.go`、根目录 `001_schema.sql`、Go/JS/WASM/binary 文件、package scripts 或 blocking Hook 模板；content / external_service 只保留 non-blocking Hook 声明。本轮不新增 API、不改 schema、不改变 upload / precheck / promote / install / upgrade 语义，不开放第三方运行时、远程 iframe、remote component、任意远程 JS、插件市场、远程在线安装、自动安装或自动启用。
- v1.9.0-S4：完成官方插件稳定化回归，范围限定 `official_links`、`official_announcement`、`official_webhook_notify`。修正 `official_webhook_notify` 官方样板包为 `1.0.1`，补齐 `external_service.subscribed_hooks=["AfterCreateContent"]` 并同步 checksums，使其安装后能订阅既有 Core 内容创建事件并执行 non-blocking external_service 投递；external-service-webhook 模板同步补齐订阅声明。`scripts/run-feishu-webhook-flow.sh` 增加发布内容类型、发布 plugin_code、板块 plugin_code 与固定 category id 覆盖项，默认 feishu_link 流程不变，可用 core `article` 触发 `official_webhook_notify` 全链路。回归覆盖 package check、声明型安装到使用、official_announcement Host + iframe、浏览器 IA 矩阵、MySQLStore fresh receiver success / 500 / timeout / manual retry、hook_executions、admin_logs、token_ref 和敏感扫描；`official_links` / `official_webhook_notify` 仅有缺 publisher/signature warning，无 blocker。本轮不新增 API、不改 schema、不改变 SecretCenter / Webhook Secret / Callback Token 安全模型，不开放第三方运行时、远程 iframe、任意远程 JS、插件市场、远程在线安装、blocking Hook、package scripts、自动安装或自动启用。
- v1.9.0-S3：完成当前生效配置与排障视图稳定化。`GET /api/v1/admin/system/effective-config` 在既有路由上新增 `root_key_status`（不含 root key 或 env 示例）、`secret_center_status`、`http_allowlist_source`、`webhook_callback_security`、`quick_links`，external_service 行新增 `auth_type`、`last_health_status`、`last_checked_at`、`last_error_at`、`last_error_summary`、`endpoint_origin`、`http_allowlist_source/http_allowlist_matched/http_allowlist_message` 和 `token_namespace/token_name/token_usage_type/token_source_type/token_source_id/token_source_code` 等脱敏排障字段。后台当前生效配置页分区展示基础运行信息、SecretCenter、Webhook / Callback 安全摘要、HTTP Allowlist、external_service 和脱敏诊断文本；复制失败时提供手动复制 textarea。`diagnostic_text` 过滤 token / secret / Authorization / root key / `DEVHUB_PLUGIN_CONFIG_KEYS` env 示例 / `encrypted_value` / token_hash；quick links 覆盖去配置、健康检查、运行记录、Secret、HTTP Allowlist、root key 状态、Webhook/Callback 和审计。本轮不改 schema、不改变 SecretCenter 加密、Webhook Secret / Callback Token 模型或 external_service non-blocking delivery，不开放 blocking Hook、第三方代码执行、插件市场或远程在线安装。
- v1.9.0-S2：完成 SecretCenter 操作闭环收口补强。SecretCenter 详情 / 列表新增脱敏派生字段 `usage_type`、`source_type`、`source_id`、`source_code`；后台详情抽屉展示 namespace、name/key、status、type、usage_type、key_id、source_type/source_id/source_code、created/updated by、timestamps、description 和使用关系，未知来源不提供误导性跳转，测试 / fixture / seed / demo 数据不开放危险操作。使用关系覆盖 external_service token、Webhook Secret、Callback Token、plugin config sensitive field、test / fixture / seed / demo 和 other / unknown；禁用 / 吊销继续先 preview，吊销要求完整 ref 强确认，轮换只跳回来源配置。SecretCenter 审计 metadata 增加 namespace/name、usage_type、source_type/source_id/source_code、plugin_code、config_entry，便于按 `secret_ref` 与来源筛选；effective-config 增加顶层和 external_service 行级 `next_steps`，后台展示 config source、token source、下一步建议和复制脱敏诊断入口。本轮不返回 token / secret / Authorization / root key / `encrypted_value` / token_hash，不改 SecretCenter 加密、Webhook Secret / Callback Token 模型、external_service non-blocking delivery 或插件安装语义，不开放 blocking Hook、第三方代码执行、插件市场或远程在线安装。
- v1.9.0-S1：启动“官方插件生态稳定版”工作线并补齐 fresh `feishu_link` receiver 全链路回归。`scripts/run-feishu-webhook-flow.sh` 新增 `DEVHUB_WEBHOOK_FLOW=full`，覆盖 success、500 failure -> `retry_exhausted`、timeout -> `network_timeout` / `retry_exhausted` 和 manual retry -> success；MySQLStore 下以新 receiver 端口 `18082` 实跑通过。Bearer token 写入 external_service 后仅保存 / 返回 `token_ref=secret://external_service/plugin_a7b0cc04/token` 与 `token_secret` 元数据，`plugin_external_services.token_ciphertext/token_hash` 为空；API 与 MySQL 扫描确认 `secret_refs`、`plugin_external_services`、`admin_logs`、`hook_executions` 未出现测试 token 明文、`Bearer `、`Authorization`、`encrypted_value` 或 `token_hash`。本轮不新增 API、不改 schema、不改变 SecretCenter / external_service / Webhook Secret / Callback Token 安全模型，不开放插件市场、远程在线安装、blocking Hook、第三方代码执行、package scripts、自动安装或自动启用。

### Documentation

- v1.8.4-S23：完成发布收口与归档。Release Notes、PROJECT_PROGRESS、TESTING、CHANGELOG 和 README 已同步为 `v1.8.4` release candidate / 可冻结口径，主题保持“官方插件生态与生产可用性增强”。最终能力按模块收口为：官方插件生态（`official_links`、`official_webhook_notify`、声明型模板）、插件包治理（upload / precheck / promote / install dry-run / install / upgrade dry-run / cleanup / 模板生成）、external_service Webhook（non-blocking 投递、health check、retry/skipped/manual retry）、SecretCenter / `token_ref`、HTTP Allowlist、系统设置 / 排障视图、admin plugin IA 和生产接受度 / 敏感扫描。P0：无；P1：无；P2：未新起 standalone fresh `feishu receiver` 全链路 `success / fail / timeout / retry`，该项为补充回归，建议进入 `v1.8.5` 或 `v1.8.4` post-release smoke，不阻塞 RC。安全边界继续冻结：不开放插件市场、远程在线安装、blocking Hook、Go plugin、JS sandbox、远程 iframe、remote component、任意远程 JS、自动安装、自动启用或 package scripts；`migrations/` 仍是唯一迁移入口，dry-run 不执行 SQL；Webhook Secret / Callback Token 模型不变；`DEVHUB_PLUGIN_CONFIG_KEYS` / root key / token / secret / Authorization / `encrypted_value` 不入库不回显。
- v1.8.4-S21：完成生产候选总验收。已记录初始工作区状态，执行 `git diff --check`、`bash -n scripts/check-frontend.sh`、`./scripts/check-frontend.sh --admin-only --quick`、`docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js`、Docker `go test ./...` 和 Docker `go build -buildvcs=false ./...`，均通过；宿主机无 `go`，已说明并改用 Docker Go。后续补跑 `scripts/check-admin-plugin-ia.sh` 通过并生成 `.devhub/screenshots/plugin-ia`，补跑 `DEVHUB_MYSQL_TESTS=1` MySQLStore 定向集成测试通过；同时修正点击矩阵脚本对当前 IA 标题 / 按钮 / 选择器的断言和 MySQL 集成测试对“迁移失败”同义错误文案的断言，不改变业务语义。验收覆盖插件包链路、cleanup、模板生成、external_service Webhook、SecretCenter、HTTP Allowlist、当前生效配置、upgrade dry-run UI、后台 UI、敏感信息扫描、Memory/MySQL 核心路径和文档状态。未发现 P0/P1，仅记录 P2：未新起独立 feishu receiver 全新 success / fail / timeout / retry 全链路。结论：`v1.8.4` 生产候选总验收通过，建议进入 S23 发布收口；安全边界保持不变，未改变 API / 生命周期 / SecretCenter / Webhook / Callback Token / external_service / upload-promote-install / upgrade dry-run 语义，未开放 blocking Hook，未执行第三方代码，未回显敏感明文。
- v1.8.4-S20 后置补充：适配 `plugin-governance.spec.js` 中插件详情抽屉 Tab 断言。S20 后旧 `前端挂载` Tab 已合并到 `能力` 摘要和 `技术详情`，旧 Webhook / 安全凭据 / 审计相关低频 Tab 也收敛到能力、运行记录或治理域；E2E 改为验证当前主 Tab、能力面板中的前端挂载说明、中文化 schema 错误、当前运行记录面板和运行与审计治理入口。插件详情抽屉仅补充稳定 `data-testid`，未恢复旧 UI 结构，不改变 API、插件生命周期、external_service、SecretCenter、权限语义或敏感信息脱敏策略。`plugin-governance.spec.js` 已复跑通过，13 passed。
- v1.8.4-S20 后置：修复 admin-e2e Compose DNS 稳定性。根因是 `devhub` 服务依赖本地预构建 `./.devhub/devhub`，当前工作区缺少该二进制导致容器退出并从 network 移除，`admin-e2e` 内表现为 `getaddrinfo EAI_AGAIN devhub`。现改为 Compose `devhub` 容器内构建临时 `/tmp/devhub` 后启动 memory store，新增 `/api/v1/health` healthcheck，`admin-e2e` / `frontend-e2e` 等待 `service_healthy`；后台 Playwright global setup 增加 DNS / HTTP ready retry，`scripts/check-frontend.sh` 增加失败诊断。插件治理 E2E 已复跑并进入页面断言阶段，不再因 DNS / 服务不可达失败；当前剩余失败为插件详情 `前端挂载` Tab 定位超时，后续按 UI 断言差异处理。本轮不改变 API、插件生命周期、external_service、SecretCenter 或安全边界。
- v1.8.4-S20：建立后台 UI 基础设计系统并轻量收口插件中心高频页面。新增 `admin-tokens.css`、`admin-layout.css`、`admin-components.css` 和 `web/admin-app/src/components/admin/` 通用组件（PageHeader、SectionCard、MetricCard、StatusTag、RiskTag、ActionBar、EmptyState、DetailDrawer、TechnicalDetails、InlineHint）；插件列表、插件详情抽屉、Webhook 治理、系统设置 / 当前生效配置、SecretCenter、插件包治理入口、本地包与预检入口、版本仓库入口已接入统一页头、指标卡、状态标签或默认折叠技术详情。旧 `data-testid` 保持兼容，敏感信息继续不回显。本轮不改变 API、插件生命周期、external_service 投递逻辑或 SecretCenter 安全模型，不引入新大型 UI 框架，不开放 blocking Hook 或第三方代码执行。
- v1.8.4-S19：SecretCenter 操作闭环与当前生效配置排障入口收口。新增 Secret 详情 / 使用关系 / 禁用预览 / 吊销预览接口，禁用与吊销执行前重新校验状态，吊销要求完整 ref 强确认，危险操作权限提升到 `super_admin` / `secret.manage` / `system.write` / `plugin.manage`。SecretCenter 列表和详情可基于真实存储关系展示 external_service、Webhook Secret、Callback Token、plugin config sensitive field 与测试数据来源；轮换入口只跳转来源治理，不做假成功。成功 / 失败动作均写 SecretCenter 审计（disabled / revoked / disable.failed / revoke.failed）。当前生效配置页补充去配置、健康检查、运行记录、查看 Secret、复制 ref/origin 和审计入口，诊断文本保持脱敏。不回显 token/secret/Authorization/root key/encrypted_value，不改变加密模型、Webhook Secret / Callback Token 安全模型，不执行第三方代码、不开放 blocking Hook。
- v1.8.4-S17 补充：新增系统设置“当前生效配置”脱敏可读视图和聚合 API `GET /api/v1/admin/system/effective-config`，并新增单插件 `GET /api/v1/admin/plugins/:code/external-service/effective-config`。external_service 的 endpoint_url、health_check_path、enabled、timeout_ms、failure_policy、健康时间线和 HTTP Allowlist 来源 / 最终生效列表明文展示；token/secret/Authorization 仍不回显，只展示 token_ref、token_status、token_key_id、token_masked 与状态提示。SecretCenter 元数据补充名称、类型、用途、关联对象、脱敏值和可用状态；后台支持复制后端生成的脱敏诊断 JSON，不包含 token/secret/Authorization/root key/DEVHUB_PLUGIN_CONFIG_KEYS/encrypted_value/数据库 DSN。未改变 SecretCenter 加密逻辑、Webhook Secret / Callback Token 安全模型，未执行第三方代码、不开放 blocking Hook。
- v1.8.4-S18：优化系统设置页 SecretCenter Tab 为“敏感配置引用”治理页。页面说明 `secret://...` 是 secret_ref / token_ref 引用地址，不是 token / secret 明文；新增说明卡片、类型筛选、状态筛选、关键词搜索、测试数据标签、中文字段、复制引用、详情抽屉、审计入口和来源配置跳转；external_service / webhook / callback / s15smoke/e2e/fixture/test/demo 数据可区分。轮换入口只引导到来源治理或中文占位，不做假成功；禁用 / 吊销复用已有 SecretCenter metadata API。页面不回显 token/secret/Authorization，不显示 encrypted_value 全文或 root key，不允许编辑 `DEVHUB_PLUGIN_CONFIG_KEYS`，不改变 SecretCenter 加密逻辑、secret_ref/token_ref 解析、Webhook Secret / Callback Token 安全模型，不执行第三方代码、不开放 blocking Hook。
- v1.8.4-S16：优化系统设置页“敏感配置与运行安全状态”和 external_service HTTP Allowlist 受控配置。启动级加密密钥继续只读，root key 仍只能来自环境变量或外部 Secret 系统，后台不保存、不生成、不修改 `DEVHUB_PLUGIN_CONFIG_KEYS`；SecretCenter 卡片增加查看 Secret / 查看审计入口，未完整实现的 Secret 管理页显示中文占位不白屏。external_service HTTP Allowlist 拆分系统默认、环境变量来源、后台配置来源和最终生效列表，新增 `GET/POST/DELETE /api/v1/admin/system/external-service/http-allowlist`，后台来源存入 `system_settings(namespace=external_service,key=http_allowlist)`。后台新增 / 删除要求权限、风险确认 / 删除确认、格式校验和审计；只允许 exact `http://host[:port]`，拒绝 wildcard、`0.0.0.0`、CIDR、path/query/fragment 和非 HTTP scheme。endpoint 保存、health check 和 non-blocking delivery 已读取最终 allowlist；仍不默认放开所有 HTTP endpoint，不回显 token/secret/Authorization，不改变 SecretCenter 加密逻辑或 Webhook Secret / Callback Token 安全模型，不执行第三方代码、不开放 blocking Hook。
- v1.8.4-S17：系统设置二级导航与安全配置 IA 收口。5 个 Tab 保持为“总览 / 安全与密钥 / 外部服务策略 / SecretCenter / 配置审计”；总览只显示三张状态摘要卡片；安全与密钥只展示 root key 状态、root key 环境变量示例和后台不能修改 root key 的提示；external_service HTTP allowlist 的环境变量示例移动到外部服务策略，并与来源列表、受控新增 / 删除入口集中展示；SecretCenter 专门展示 secret_ref / token_ref 元数据；配置审计专门展示 admin_logs / 配置变更记录。本轮只改后台页面 IA 与文案，不改变 SecretCenter、Webhook、external_service 后端安全语义，不回显 token/secret/Authorization。
- v1.8.4-S16 IA 补充：系统设置页新增“总览 / 安全与密钥 / 外部服务策略 / SecretCenter / 配置审计”二级导航。总览只展示“启动级加密密钥 / SecretCenter 引用层 / external_service HTTP 策略”三张摘要卡片；环境变量示例移动到安全与密钥；HTTP Allowlist 管理移动到外部服务策略；SecretCenter 入口和元数据列表移动到 SecretCenter；admin_logs / 配置变更记录筛选和列表移动到配置审计，且该页不夹带基础设置表单。本轮只改后台页面 IA、文案和入口组织，不改变后端安全语义。
- v1.8.4-S11：收口 Webhook 本地联调体验与 external_service 配置治理。后端在 endpoint 安全拒绝、连接拒绝、超时、HTTP 状态失败、token 缺失 / 无效、插件 / 子站 / 服务停用等场景补充安全诊断、suggestion 和 diagnostics，且 diagnostics endpoint 会脱敏 userinfo 和敏感 query key；非 localhost HTTP 继续默认拒绝，提示 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build`。后台插件详情和 Webhook 治理页提供“配置 external_service”直达入口，详情页把当前健康、最近成功、最近失败、24h / 7d 历史失败计数和当前 endpoint 分开展示；全局 `config_json` 中 `endpoint_url`、`health_check_path`、`token`、`timeout_ms`、`failure_policy` 会提示不会自动覆盖运行配置，普通全局 `enabled` 不误报。重复安装提示改为引导配置 / external_service / 版本仓库 / 升级差异 / 审批流程；版本仓库 CTA 使用“升级差异”。文档明确 Docker 容器内 `127.0.0.1` 语义、18081 feishu_link receiver 与 18090 mock receiver 的区别；仍不开放 blocking Hook、不执行第三方代码、不改变 Webhook Secret / Callback Token 安全模型。
- v1.8.4-S12：S11 回归验收与后台点击验证。已在 Docker 后端场景验证 `127.0.0.1` loopback 诊断、`172.17.0.1` HTTP allowlist 拦截与推荐命令提示、allowlist 生效后 health check/投递可用、external_service 配置入口与全局配置提示清晰、健康摘要拆分不再被旧失败误导、重复安装错误引导配置或升级流程；联调过程中修复后台 `POST /api/v1/admin/posts` 在设置非默认 status 的边界场景可能 panic 的问题。不新增能力、不开放 blocking Hook、不执行第三方代码、不改变 Webhook Secret / Callback Token 安全模型。为保证 bearer token 在本地 Docker Go 运行时可用，`dev.sh` 会读取仓库根目录 `.env` 并透传 `DEVHUB_PLUGIN_CONFIG_KEY*`。
- v1.8.4-S13：系统配置中心与敏感配置治理收口。新增只读系统敏感配置状态接口（启动加密 keyring 状态 + external_service HTTP allowlist）；系统设置页展示只读状态与 env 示例；Webhook Secret 创建前增加 readiness 检查；缺少启动密钥时错误提示补齐两种 env 示例并明确“后台不会保存或生成 root key、修改需重启”；后台统一对 `secret_ref/token_ref` 做脱敏展示（可复制引用，不铺开明文）。
- v1.8.4-S14：落地 Core SecretCenter 最小引用层，并将 external_service bearer token 统一收口为 `token_ref=secret://external_service/{plugin_code}/token`。新增 `secret_refs` 存储表与 SecretCenter 元数据 API（不提供明文读取）；external_service health check/投递运行时 resolve token_ref（仅内部使用），API/审计/执行记录不回显 token/Authorization；系统敏感配置只读状态增加 SecretCenter 统计；后台 external_service 配置区域展示 token_ref 与 token_secret 元数据，保存响应也返回 token_secret 元数据但不回显明文；支持 disable/revoke 后失败原因清晰。S14 收口复核补强了 MemoryStore secret_refs 索引、token_ref 脱敏展示、定向测试和 MySQL smoke。
- v1.8.4-S15：完成 SecretCenter + Webhook + external_service 在 MySQLStore 下的实跑验收。以 Docker Go + MySQL + feishu_link receiver 验证 SecretCenter 直接写入/轮换、external_service bearer token_ref 写入/health check/真实投递、Webhook Secret 创建/轮换、Callback Token 创建并调用 config/audit 回调、MySQL 密文落库和 admin_logs/hook_executions/plugin_callback_requests 脱敏；修复 MySQLStore `plugin_webhook_secrets` 与 `plugin_callback_tokens` insert 占位数量不匹配导致 Webhook Secret / Callback Token 创建失败的问题。仍不回显 token/secret/Authorization，不改变 Webhook Secret / Callback Token 安全模型。
- v1.8.4-S15：优化后台系统设置“敏感配置与运行安全状态”只读展示。原调试表格改为“插件配置加密 / SecretCenter 引用层 / external_service HTTP Allowlist”三张状态卡片，状态文案中文化，环境变量示例改为只读代码块并提供复制按钮；后台仍不保存或生成 root key，不编辑 `DEVHUB_PLUGIN_CONFIG_KEYS` 或 HTTP Allowlist，不显示真实 root key/token/secret/Authorization。
- v1.8.4-S9：优化后台初始化插件包 / 创建插件模板表单。插件名称自动生成 code，高级设置可手动修改，code 支持小写字母、数字、下划线和中划线；新增内容型、外部服务型、后台工具型、前端挂载型插件类型并动态显示字段；“内容类型 / 内容类型名”改为“内容数据类型 / 内容显示名称”且只在内容型插件展示；“作者”改为“发布者 / 作者”下拉。preview 返回 code、content_type、权限、菜单、Hook、external_service、frontend_mount、migrations、文件树和 conflicts；generate 写入 `storage/plugins/drafts/{code}`，export zip 使用同一套 `PluginTemplateGenerator`。模板只生成声明文件和 `migrations/001_init.sql`，不生成根目录 `001_schema.sql`，不安装、不启用、不执行 SQL、不执行第三方代码。
- v1.8.4-S8：新增插件包本地仓库测试数据批量清理。后台本地仓库支持“清理测试包 / 清理未安装包 / 清理 blocked / invalid”，新增 `/api/v1/admin/plugins/packages/cleanup/preview` 预览和 `/cleanup` 执行；识别 `e2e_`、`fixture_`、`test_`、`demo_` 与 `e2e_upload_*` 测试包，支持清理未安装的 blocked / invalid / warning / promoted 包。执行前后端重新校验，删除 `storage/plugins/packages/` 目录和对应 promoted upload 记录，installed / enabled / active task 包强制 skipped；清理写 `plugin.package.cleanup.*` 审计，不删除配置、安装记录、历史内容、admin_logs 或 hook_executions，不执行第三方代码。
- 插件包清理 / 删除入口补齐：上传记录页支持删除 `uploaded/staged/blocked/failed/expired`，批量 cleanup 支持 `dry_run`、`confirm_token`、`statuses`、`older_than_days` 并返回 `will_delete_count/will_free_bytes/items`；本地仓库页支持删除 `storage/plugins/packages/` 下所有未安装包，并提供 `/plugins/packages/repository/cleanup` 显式清理 blocked/invalid 别名接口，已安装包禁删并提示先归档 / 软卸载插件。新增审计事件 `plugin_package_upload.deleted`、`plugin_package_upload.cleanup`、`plugin_package_quarantine.deleted`、`plugin_package_repository.deleted`、`plugin_package_repository.cleanup`；删除路径仍限制在 uploads / staging / quarantine / packages，不执行插件代码或 SQL。
- v1.8.4-S6：补齐浏览器点击矩阵。`scripts/check-admin-plugin-ia.sh` 现在覆盖 5 个治理域、3 个官方插件、37 条旧路由和 1024 / 1366 两档截图；`official_links`、`official_announcement`、`official_webhook_notify` 的详情、配置、挂载、预览、health check、manual retry、审计和脱敏检查都已纳入回归。当前 dev 环境缺少 external_service token 加密 key，官方 Webhook 矩阵在保存配置时回退到 `auth_type=none` 继续完成页面点击，不代表生产默认配置；仍不开放第三方代码执行、远程 iframe、remote component 或 blocking Hook。
- v1.8.4-S5：补齐生产备份 / 回滚 / 升级演练文档。`docs/BACKUP_AND_ROLLBACK.md`、`docs/DEPLOYMENT.md`、`docs/PLUGIN_PACKAGE.md`、`docs/PLUGIN_DEVELOPER_GUIDE.md`、`docs/PLUGIN_SYSTEM_ROADMAP.md`、`docs/TESTING.md`、`docs/API.md`、`docs/releases/v1.8.4.md`、`README.md` 和 `docs/PROJECT_PROGRESS.md` 统一记录生产备份清单、安装 / 升级前检查、upgrade dry-run `safe / warning / blocked`、warning 必须 `confirm=true`、blocked 不能绕过、`failure_stage` / `failure_reason` 处理建议、`PluginRegistry reload` 失败路径、配置与加密 key 恢复、本地仓库恢复、Webhook Secret / Callback Token / external_service 元数据备份、MySQLStore 生产建议和 MemoryStore 限制；仍不承诺完整自动 rollback 或 migration down，不改变插件运行能力。
- v1.8.4-S4：新增插件包打包 / 校验 CLI。`scripts/plugin-package-build.sh` 可在仓库内生成官方包 / 模板 zip，`scripts/plugin-package-check.sh` 可对目录、zip 或内置 `official_announcement` 做本地校验，统一输出 passed / warning / blocked；CLI 只复用现有 package dry-run / manifest 校验逻辑，不执行 SQL、不执行插件包代码、不执行 package scripts，也不访问远程市场。
- v1.8.4-S4：修复配置模型正则兼容性。`official_announcement` 的 `link_url` 和 `official_webhook_notify_template` 的 `health_check_path` 统一改为 JS / JSON Schema 兼容路径正则，`scripts/check-frontend.sh` 的包脚本探测也去掉了仓库内残留的旧空白字符类写法；新增 `TestConfigSchemaPatternsAreJSCompatible` 验证配置模型可正常编译并接受合法路径。仍不改 external_service 业务逻辑、不改 Webhook 协议、不改插件安全模型。
- v1.8.4-S3：完成 `official_webhook_notify` 官方 external_service Webhook 通知样板生产化复查。官方样板包 `examples/plugins/official_webhook_notify/` 继续覆盖 install dry-run、健康检查、异步投递、hook_executions、manual retry、disabled / archived / community disabled 阻断和安全脱敏边界；定向 service tests 已通过。`official_webhook_notify` 仍只代表 external_service non-blocking Webhook 子集，不开放第三方代码执行、远程 iframe、remote component 或 blocking Hook。
- v1.8.4-S1：完成 `official_links` 官方友情链接插件生产化。新增官方包 `examples/plugins/official_links/`，声明 `friend_link` content_type、权限、后台 / 前台菜单、配置 schema、默认配置和 `migrations/001_init.sql` no-op 计划文件；后台新增 `/admin-next/official-links` 复用通用 `PluginContent` 管理友情链接，前台搜索 / 子站导航补充 `friend_link` / `official_links` 入口，配置 `enabled=false` 时隐藏展示入口；修复搜索接口只认可 Core 固定 content_type 导致 `friend_link` 筛选失效的问题；新增 official_links package dry-run 和声明型 content_type 搜索回归测试。`VERSION` 切到 `v1.8.4`；仍不执行第三方代码，不开放 Go plugin、JS 沙箱、远程 iframe、remote component、插件市场、远程在线安装或 blocking Hook。
- v1.8.3-S23：完成插件系统发布前总验收与稳定性收口。覆盖 official_links、official_webhook_notify、upgrade dry-run/confirm/failure_stage、开发者指南与模板、frontend_mount allowlist、MemoryStore/MySQLStore、后台治理和文档一致性；修复 MySQLStore Webhook 相关表 `target_url` 联合索引过长问题，并为 Webhook Secret 创建补充 `target_url` 512 字符上限校验，`./dev.sh start --mysql`、插件 IA 回归与 Go/前后台/fixture/专项检查均通过。仍不开放插件市场、远程在线安装、第三方代码执行、Go plugin、JS 沙箱、远程 iframe、remote component、远程 JS 或 blocking Hook。
- v1.8.3-S22：前端插件挂载继续收口到官方 allowlist。manifest 现在可以声明 `frontend_mounts`，但只允许官方挂载点和官方组件 key；预检 / install dry-run / upgrade dry-run 会阻断未知挂载点、未知组件 key、unsupported render_mode、`iframe_url`、`script_url`、`remote_entry`、`external_js`、`inline_html`、`remote_component`、`eval` 和未白名单的可执行 JS 资产，运行时只渲染已安装、已启用 / running、未归档、未软卸载、当前子站可用的 allowlist 挂载，且会过滤 secret / token / authorization / password / credential 类 props。后台插件详情前端挂载表格、官方公告插件样板和升级 diff 都已同步，不改变 API、Webhook 协议、Secret / Token 安全模型，也不开放任意远程 iframe、远程 JS 或第三方前端运行时。
- v1.8.3-S22 运行时复查：前端挂载运行时继续只返回已安装、enabled / running、未 archived、未 soft_uninstalled 且当前子站 enabled 的 allowlist 挂载；unknown component 跳过并返回 warning，secret / token / authorization / password / credential 类 props 不传给前端组件，官方 helper 仍只创建内置同源 iframe。
- v1.8.3-S21：补齐声明型插件开发者指南，并把 `official_links` 与 Webhook demo 固化为两个官方模板。新增 `docs/PLUGIN_DEVELOPER_GUIDE.md`、`examples/plugins/templates/declarative-content/` 和 `examples/plugins/templates/external-service-webhook/`；模板覆盖 content_type / 权限 / 菜单 / config_schema / migrations/ 与 external_service non-blocking Hook / health check / retry 验收，不包含运行时代码、真实 secret、远程 iframe 或 blocking Hook。
- v1.8.3-S21 复查：开发者指南继续对齐 allowlist 前端挂载和升级失败阶段 / 下一步建议的说明，帮助插件开发者更快避坑；仍不执行第三方代码，不开放远程 iframe 或 blocking Hook。
- v1.8.3-S21 再复查：开发者指南现在被当作声明型插件的默认起步入口，明确 `migrations/` 唯一入口、dry-run 不执行 SQL、`external_service` 仅是 non-blocking 子集；相关链接也同步到文档入口与路线图。
- v1.8.3-S20：完成插件升级体验收口。`upgrade dry-run` 现在返回结构化版本计划、变更摘要、影响范围、风险项和回滚边界；`upgrade` 与基于 compat-check 的升级都要求 warning 显式确认，blocked 不可绕过；后台升级向导改为结构化展示，原始 JSON 收进技术详情折叠区。仍不执行第三方代码，不开放远程 iframe、不开放 blocking Hook，不改变 Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S20 复查：warning 仍需显式确认，blocked 仍不可绕过；dry-run plan 过期或与当前包 checksum / migration plan 不一致时继续拒绝升级；升级失败继续返回 `failure_stage` / `failure_reason`，后台结果页保留下一步建议。
- v1.8.3-S19：完成 external_service Webhook 可交付闭环。新增官方样板包 `examples/plugins/official_webhook_notify`、手动重试 API、后台 external_service 配置表单和 Webhook 治理“外部服务执行”入口；补充样板包 dry-run 守护测试和 Core 版本读取修正，保证样板包按当前 `VERSION` 通过现有预检；补充 API、使用方法、包规范、架构边界、测试验收、项目进度和 v1.8.3 release notes。仍不执行第三方代码、不开放 blocking Hook、不开放远程 iframe、不改变 Webhook Secret / Callback Token 安全模型。
- external_service 后台配置表单复查：插件详情抽屉“运行记录 / 外部服务配置”继续覆盖 endpoint、token 写入、health_check_path、timeout_ms、failure_policy、enabled 和健康状态展示；Webhook 治理“外部服务执行”继续用于投递记录和失败重试。token 不回显，不进入日志 / 审计 / hook_executions 明文；插件 disabled / archived 时不调用 endpoint。
- 新增 Webhook 插件使用方法文档 `docs/PLUGIN_WEBHOOK_USAGE.md`，把 external_service non-blocking 投递、Webhook Secret、Callback Token、健康检查、投递记录、异常处理和排障入口整理成管理员可执行流程；仅补文档，不改 API、Webhook 协议或安全模型。
- 当前版本口径统一到 `v1.8.3`：`VERSION` 从 `v1.7.1` 更新为 `v1.8.3`，README、docs/README、AGENT_RULES、PROJECT_PROGRESS、PLUGIN_ARCHITECTURE、PLUGIN_SYSTEM_ROADMAP、PLUGIN_PACKAGE、PLUGIN_WEBHOOK_IMPLEMENTATION_PLAN 和 v1.8.3 release notes 已同步当前主线说明；历史 v1.7.x 文档保留为追溯背景。
- v1.8.0：新增插件前端挂载模型设计文档（slots + iframe/sandbox + postMessage 协议与权限/状态 gating），明确官方公告插件作为首个前后台挂载验证方向（文档设计，未修改代码、未执行测试）。

### Added

- v1.8.3：优化后台插件治理页面稳定性、IA 和中文体验。修复 Webhook 治理页空数据/缺字段白屏风险与插件详情抽屉运行时异常；左侧插件导航收敛为插件总览、插件包治理、Webhook 治理、可信发布者、运行记录/审计 5 个治理域；插件详情抽屉拆出前端挂载、Webhook、Webhook 密钥、回调 Token 等三级 Tab；Webhook 治理页拆成事件、投递记录、重试队列、熔断状态、Webhook 密钥、回调 Token、回调请求；插件包治理、上传记录、远程插件包、审批中心、操作历史、配置版本历史、可信发布者和官方公告插件预览补充中文状态、按钮、空状态与安全边界说明。本轮不改变插件生命周期、Webhook 协议、Secret / Token 安全模型，不开放远程 iframe、第三方代码执行、插件市场或 blocking Hook。
- v1.8.3-S1：完成插件后台稳定性修复与 IA 第一批收敛。过滤 null 插件项并保护详情抽屉空插件状态；插件列表首屏减负并快速标识 official_announcement；上传包详情 JSON 默认折叠到“技术详情”；可信发布者列表突出主信息并弱化 publisher_id/key_id 技术字段；后台 quick build 通过。
- v1.8.3-S2：完成插件后台二级导航收敛与三级 Tab 重组。左侧插件导航只保留 5 个治理域，原 20+ 入口沉到页内 Tab；旧路由兼容跳转到新治理域和 `?tab=`；Webhook 治理新增总览 Tab；后台 quick build 通过。
- v1.8.3-S4：修正插件后台总览页左侧异常留白和内容偏右；插件列表筛选区改为紧凑横向筛选栏，高级筛选默认折叠；健康摘要和批量操作减少首屏占用；official_announcement 详情入口更直观；新增 `scripts/check-admin-plugin-ia.sh` 用于轻量回归 5 个治理域和旧路由 Tab。
- v1.8.3-S5：优化插件详情抽屉视觉层级与信息密度。可见 Tab 收敛为概览、配置、前端挂载、Webhook、安全凭据、运行记录、审计日志、技术详情；Webhook 密钥和回调 Token 合并为安全凭据，原始配置和声明 JSON 默认进入技术详情并脱敏，official_announcement 配置/挂载/预览入口更直观；后台 quick build 通过。
- v1.8.3-S6：完成插件详情抽屉性能拆包与 1024 宽度视觉回归。配置版本弹窗、技术详情和 JSON 编辑器改为按需加载，详情抽屉低频 Tab 懒渲染；`PluginConfigEditor` chunk 明显降低，普通插件详情、技术详情、配置 Tab 和配置版本弹窗已截图回归；后台 quick build 通过。
- v1.8.3-S7：为 `official_announcement` 浏览器回归补固定 fixture。`scripts/check-admin-plugin-ia.sh` 通过现有后台 API 幂等准备官方公告插件全局 / 子站启用状态和公告配置，并强制覆盖插件列表、详情概览、公告配置、前端挂载、公告预览截图；同时修复后台预览 Host helper 请求 context / audit 未带 admin Authorization 导致 iframe 不创建的问题。token 只用于 Host 请求，不暴露给 iframe；不开放远程 iframe，不改变插件逻辑或 Webhook 协议。
- v1.8.3-S8：收口插件包 migrations 规范。插件包 dry-run / 预检 / 安装 / 升级只读取 `migrations/` 下的 `.sql` 文件并生成只读 migration plan；根目录 `001_schema.sql` 降级为 deprecated warning，不再作为标准迁移入口或执行目标；新模板 / 示例统一生成 `migrations/001_init.sql`，dry-run 不执行 SQL、不修改数据库、不写安装状态。
- v1.8.3-S9：收口 PluginRegistry reload 运行态刷新。安装、升级、启停、软卸载 / 归档、恢复、全局配置、子站插件状态和子站配置变化后刷新运行态快照；reload 失败保留旧快照并写审计，不执行第三方代码、不开放动态加载。
- v1.8.3-S10：完成声明型插件可用闭环。manifest 插件声明的菜单、content_type、权限和配置声明可进入运行态、子站启用、发布校验和权限矩阵；disabled / archived / 子站 disabled 会阻断新能力但保留历史内容可读；不改变 Webhook 协议、Secret / Token 安全模型或动态加载边界。
- v1.8.3-S11：完成 external_service Webhook 运行时预备。新增外部服务配置模型、Admin API、endpoint / timeout / failure_policy / auth_type 校验、Bearer token 加密引用、受控 health check、`hook_executions(service_type=external_service)` 记录和健康 warning/error 联动；后台插件详情展示外部服务健康摘要。插件 disabled / archived 时只记录 skipped，不调用 endpoint；不执行第三方代码、不开放动态加载、不做 blocking Hook。
- v1.8.3-S14：完成 external_service non-blocking Webhook 投递闭环。声明型插件可用 `service_type=external_service` 声明 non-blocking Hook，触发后异步 `POST {endpoint_url}{hook.path}` 并写入 `hook_executions`；支持 timeout、auth_type=none/bearer、retry、failure_policy、health warning/error 和 disabled / archived / 子站 disabled 阻断。投递不阻塞主业务流程，不记录 token 明文或 Authorization Header；仍不执行第三方代码、不开放动态加载、不实现 blocking Hook。
- v1.8.3-S12：继续收口插件包 upload -> promote -> install。blocked / failed 上传包不可 promote；promote 只转入本地仓库，不安装、不启用、不执行脚本；本地仓库列表展示 `source_upload_id/promoted_at` 以追溯来源上传包；install 只能从本地仓库包执行，且必须携带当前 install dry-run 计划凭证 `dry_run_id`，服务端会重新 dry-run 并核对 path / plugin_code / version / manifest checksum / checksum status / migration plan hash；upload/staging 阶段旧 dry-run 结果不能直接复用。dry-run 仍不执行 SQL，install 只基于 `migrations/` 计划执行，根目录 `001_schema.sql` 不执行。
- v1.8.3-S13：新增真实插件包 fixture 验收 S12 链路。脚本可生成 valid / blocked / deprecated schema 三类 zip，后台 E2E 用真实 Admin API 验证上传、预检、blocked promote 拒绝、promote 入本地仓库、安装前重新 dry-run、安装、审计和插件包治理页面 smoke；不执行第三方代码，不开放动态加载，不改变 Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S15：新增真实声明型友情链接插件 fixture 验收“安装到使用”完整闭环。`devhub-fixture-links-plugin*.zip` 通过真实 Admin API 覆盖 upload、precheck、promote、本地仓库 install dry-run、install、PluginRegistry reload、全局启用、子站启用、菜单、`friend_link*` content_type、权限矩阵、配置读写、禁用 / 归档阻断和历史内容可读；后台 `admin/posts` 可在受控 admin 场景显式携带声明型 `category_id/content_type/plugin_code` 并继续走 Service 层校验，避免 manifest 内容类型创建时回退到 `core`。不执行第三方代码、不开放动态加载、不开放远程 iframe、不实现 blocking Hook。
- v1.8.3-S16：继续压缩后台插件治理复杂度。5 个治理域保持不变，但默认界面进一步轻量化：插件总览低频入口沉到“高级治理”，插件包治理改为 upload/precheck/promote/install dry-run/install 的流程型工作台，Webhook 治理聚焦投递记录和异常处理，发布者与信任、运行记录 / 审计弱化技术字段，插件详情抽屉压缩为当前插件摘要、能力、运行和技术详情入口。`scripts/check-admin-plugin-ia.sh` 已同步新抽屉结构并完成浏览器截图回归；后台 quick build 通过。不改变 API、插件逻辑、Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S18：修复插件后台按钮行为与旧路由跳转回归。插件详情抽屉治理按钮不再跳到错误 `/admin-next` 路径；被压缩隐藏的 Webhook、密钥、Token、前端挂载、审计、Hook、readiness、dependencies、content_type、permissions、migrations、routes 等旧入口会落到现有 Tab 并显示中文合并说明；补齐 `/plugins/package-uploads`、`/plugins/remote-packages`、`/plugins/webhook-*`、`/plugins/callback-*` 等旧路由到新治理域 / Tab 的映射。浏览器 IA 回归和后台 quick build 通过；不改变 API、插件逻辑、Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S18 追加：复查隐藏旧技术入口，插件详情“技术详情”按启用检查、迁移明细、导出本地插件包、原始声明 JSON 做最小分组；启用检查和迁移明细不再依赖隐藏旧 Tab，迁移重试按钮补齐前端 API import，导出入口保持可见，原始声明继续默认折叠和脱敏。不改变 API、插件逻辑、Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S18 再追加：把插件总览、详情抽屉和长期文档统一改成任务导向口径。管理员现在先按“安装 / 启用 / 排障 / 看审计 / 管密钥 / 看原始声明”判断入口，再进入对应治理域，不再只靠页面名猜位置。不改变 API、插件逻辑、Webhook 协议或 Secret / Token 安全模型。
- v1.8.3-S18 再再追加：插件列表主按钮和行内动作也统一成任务语言，主按钮改为“校验清单 / 查看预检 / 去安装”，行内动作改为“看详情 / 处理配置 / 更多任务 / 去审计 / 去启用 / 去停用 / 做启用检查 / 去软卸载 / 恢复入口”。不改变 API、插件逻辑、Webhook 协议或 Secret / Token 安全模型。
- 后台插件中心中文状态和异常提示统一：新增 `web/admin-app/src/modules/plugins/statusText.js` 作为插件模块状态 / 风险 / 阻断原因 / Hook 健康原因 / 操作名 / 建议文案集中映射；插件列表、详情、上传包、promote/install、远程索引、版本升级、配置密钥等页面复用同一中文口径；前端错误提示优先展示中文 message，仅有 code 时映射中文并保留错误码。未改变 API code、状态枚举、插件生命周期或安全边界。
- 优化插件后台公共筛选条布局：标题说明置顶、筛选控件网格排布、按钮与条件同行，改善宽屏下控件过长和按钮位置松散的问题。
- 调整插件模块二级导航呈现：保留左侧二级导航栏，在“插件管理”分组下直接展示 5 个治理域，页内 Tab 继续作为三级导航。
- 修复 `/admin-next/plugins/operations` 操作历史页读取 `undefined.items` 的运行时错误，兼容前端 HTTP 拦截器已解包的业务响应和空列表响应。
- 修复 `scripts/check-frontend.sh` 的 Docker named volume 权限初始化：当 `node_modules`、`test-results`、`playwright-report` 曾由 root 容器创建时，检查脚本会先修正属主，再执行后台/前台构建与 E2E，避免 Vite `.vite-temp` 或 Playwright 报告写入失败。
- v1.8.1：新增内置官方插件 `official_announcement`，落地前台首页与后台插件详情页“公告预览”Tab 的 Host + iframe（`sandbox=allow-scripts`）最小挂载闭环；新增 Host 浏览器安全 API（context + audit-events）与内置 iframe 路由（不允许远程 URL，不暴露 callback token / webhook secret，不执行第三方不可信代码）。
- v1.8.1-S2：对子站页 `/c/:slug` 的公告挂载做 SEO 回归与收口：保持 `<title/canonical/JSON-LD/h1>` 与主体内容不变，补齐 context/audit 的 `community_slug` gating 防绕过。
- v1.8.2：新增官方插件前端挂载共享 helper `GET /plugins/assets/devhub-plugin-mount-host.js`，用于前台首页、`/c/:slug` 与后台插件详情复用统一 Host + iframe + postMessage 挂载机制（仅 allowlist 官方内置插件；仍不允许远程 iframe URL）。

### Documentation

- Added v1.7.2 Webhook / HTTP plugin service protocol design (signing, replay protection, idempotency, retry, timeouts/rate limits/circuit breaker, audit, and governance planning).
- Added v1.7.3 implementation plan for the Webhook / HTTP plugin service protocol (non-blocking delivery first, with delivery records, retry queue, circuit breaker, audit, and minimal governance UI; blocking hooks explicitly deferred).
- Added an official example plugin plan (official announcement plugin) for end-to-end verification of non-blocking Webhook delivery without executing third-party code.
- Added v1.7.2 plugin runtime model design, defining Core built-in plugins, external HTTP service plugins, and iframe/sandbox frontend plugins.
- Documented frontend plugin mount slots, controlled Core API scopes, HookBus participation, runtime isolation boundaries, and manifest runtime design fields as planning-only capabilities.
- Adjusted DevHub's project goal to a Core + plugin open-source service foundation.
- Unified README, project progress, plugin architecture, plugin roadmap, API, testing, SEO, plugin package, remote index, SDK, template, release notes, and agent rules around the Core/plugin boundary.
- Clarified that Core provides stable base capabilities while plugins carry business extensions, frontend/backend extension points, Hook/config/content-type integration, and ecosystem growth.
- Clarified that default community capabilities are part of Core, not the sole project positioning.
- Recorded this as a documentation-only task: no code changes, no tests, no builds, no E2E.

### Added

- Webhook governance (v1.7.5): added webhook deliveries retry scheduling (`retry_scheduled`/`retry_exhausted`, `next_retry_at`, `attempt/max_attempts`) and circuit breaker (`closed/open/half_open`) with minimal admin APIs and admin UI page. Non-blocking only; no third-party plugin code execution.
- Webhook signing & secret rotation (v1.7.6): added HMAC-SHA256 signed webhook delivery (sender-side), persisted signature metadata on deliveries (no plaintext secret), plus admin-managed webhook secrets (create/rotate/disable/enable/revoke) with one-time secret display and a minimal admin UI tab.
- Webhook controlled Core API callback (v1.7.7): added callback tokens (bearer, one-time plaintext display, stored as `token_hash` only), minimal scope whitelist (`config.read`, `audit.write`), community scope checks, plugin callback endpoints (`/api/v1/plugin-callback/config`, `/api/v1/plugin-callback/audit-events`), callback request logs, and minimal admin UI tabs (Callback Tokens/Callback Requests).
- Webhook governance (v1.7.8): added webhook events list API and an Events tab in the admin Webhook governance page; added an official webhook mock receiver binary for end-to-end verification (no third-party code execution).
- Webhook governance (v1.7.9): completed non-blocking chain acceptance checklist and documented blocking-hook risk assessment (design only; still no blocking execution).

## v1.7.1 (2026-05-16)

v1.7.1 is the signed plugin package verification and trusted publisher enhancement track. It adds detached signature verification (`devhub-signature.json`) over remote/staging packages and gates staging→compat-check→install/upgrade by verified signatures by default.

- Added detached signature model `devhub-signature.json` (canonical JSON payload) and Ed25519 verification; signature payload binds `plugin_code`, `version`, `package_sha256`, `manifest_sha256`, and publisher key identifiers.
- Extended remote package download request/record with optional `signature_url` (HTTPS + `.json` only, SSRF-protected, 64KB max) for detached signature retrieval during verification.
- Added persistent `plugin_package_signatures` records plus admin APIs: `POST /api/v1/admin/plugins/packages/prechecks/:id/verify-signature` and `/api/v1/admin/plugins/packages/signatures` list/detail/delete.
- Enforced signature gating in compat-check/install/upgrade by default (`DEVHUB_PLUGIN_REQUIRE_SIGNED_PACKAGES`), blocking unsigned/unverified/untrusted/revoked/expired signatures from entering install/upgrade flows.
- Enhanced trusted publishers with optional `expires_at` (expired keys are blocked during verification).
- Added minimal admin UI entry for signature verification and signature record list in the package governance page.

## v1.7.0 (2026-05-16)

v1.7.0 is the remote plugin package governance and installation-safety enhancement track. P0-01 adds safe remote package download into staging only; it does not install, enable, extract for execution, run plugin code, execute SQL, or dynamically load frontend assets.

- Added `POST /api/v1/admin/plugins/packages/download` for HTTPS-only remote package download into `storage/plugins/staging/downloads/`, with SSRF protections, redirect revalidation, size limits, supported extension checks, temporary-file cleanup, sha256 calculation, and checksum mismatch blocking.
- Added staging list/detail/delete APIs under `/api/v1/admin/plugins/packages/staging` and a lightweight admin UI entry in the plugin upload package management page.
- Added persistent `plugin_package_downloads` records for status, source/final URL, file size, sha256 expected/actual, content type, staging path, errors, and audit traceability.
- Added plugin package dependency / compatibility checks over `plugin_package_prechecks.status=passed` records, with persistent `plugin_package_compat_checks`, Core version constraints (`^` / `~` included), dependency validation, plugin/content/permission/menu/route/Hook/config_schema/migration blockers, backend-computed `can_install`, audit logs, and a minimal admin UI. This still does not install, enable, register, migrate, execute, or dynamically load plugin code.
- Added plugin enable-precheck (`POST /api/v1/admin/plugins/:code/enable-precheck`) with persistent `plugin_enable_prechecks` results and minimal admin UI entry; runs file/manifest/dependency/config/migration/conflict checks only and does not enable/register/execute.
- Added plugin enable based on enable-precheck (`POST /api/v1/admin/plugins/enable-prechecks/:id/enable`) with persistent `plugin_enable_tasks` records, audit logs, TOCTOU re-checks, pending-migration blocking, and a minimal admin UI entry. This registers only manifest-declared governance capabilities (content types/permissions/menus/routes/hooks/config snapshot) and still does not execute plugin code, scripts, or migrations.
- Added compat-check driven plugin upgrade tasks: `GET /api/v1/admin/plugins/:code/upgrade-impact?target_compat_check_id=...`, `POST /api/v1/admin/plugins/:code/upgrade-from-package`, plus `plugin_upgrade_tasks` list/detail/retry/delete APIs. Upgrades re-run package dry-run server-side, require sha256-verified staging downloads, do not auto-enable, and do not execute migrations or third-party code.

## v1.6.0 (2026-05-15)

Current `VERSION` is `v1.6.0`. v1.6.0 closes the plugin package upload and distribution-prep track: zip upload sandbox, upload lifecycle, real Ed25519 signature verification, trusted publishers, read-only remote indexes, version repository, upgrade diff, operation recovery previews, config key rotation, and admin plugin UI grouping are in place. See `docs/releases/v1.6.0.md`.


- Added admin zip plugin package upload into a controlled sandbox: `.zip` only, 20MB upload limit, 50MB extracted limit, 5MB per-file limit, 300-file limit, depth limit, nested archive blocking, zip slip checks, symlink/special-file blocking, and package-root detection.
- Added upload detail and promote APIs. Promote copies validated staging packages into `storage/plugins/packages/{code}` but does not install, enable, execute code, run SQL, or dynamically load frontend assets.
- Added the admin upload panel under `/admin-next/plugins/install`, showing zip scan, package scan, checksum, signature, risk report, manifest validation, dry-run, blocked reasons, suggestions, and promote action.
- Added backend and Playwright coverage for valid zip upload, invalid type, zip slip, dangerous files, checksum/risk reuse, promote, auth rejection, and audit logging.
- Added `plugin_package_uploads` lifecycle records for uploaded zip packages, with list/detail/rescan/import-approval/approve/reject/promote/cancel/delete/cleanup APIs.
- Added `/admin-next/plugins/packages/uploads` upload package management page with lifecycle filters, scan snapshots, action reasons, import approval, promote, cancellation, deletion, and cleanup controls.
- Kept the v1.6 boundary strict: promote is not install; uploaded packages still do not execute plugin code, run SQL, dynamically load frontend assets, or connect to a remote marketplace.
- Added Ed25519 real signature verification for plugin packages and an admin trusted publishers management page/API; signed package dry-run, repository scan, upload detail, promote, install, and approval execution now use verified/trusted/unknown/blocked/revoked signature states in `risk_report`.
- Added plugin config encryption keyring + rotation tooling: `enc:v2:<key_id>:...` ciphertext format, multi-key decryption for legacy `enc:v1`, admin key status endpoint and rotation dry-run/re-encrypt APIs, plus `/admin-next/plugins/config-keys` UI (no key material exposure, no KMS/Vault, no scheduled rotation).
- Added read-only remote plugin indexes, plugin package version repository, structured upgrade diffs, and operation recovery previews; these remain metadata / governance flows only and do not download, install, execute code, run SQL, or dynamically load assets.
- Closed v1.6.0 acceptance with Go tests/build, admin build + full admin E2E (`62 passed`), frontend build + frontend E2E (`17 passed`), SEO curl checks for `/topics/1/` and `/c/php/`, and documented that zip export download/signing packaging is still a v1.7 follow-up.
- Refined the admin plugin governance UI information architecture into six function groups, added compatibility redirects for older plugin routes, extracted lightweight shared plugin display components, added grouped plugin API exports, and replaced the old navigation E2E with `plugin-governance-pages.spec.js`.

## v1.5.0 (2026-05-14)

Current `VERSION` is `v1.5.0`. v1.5.0 closes the plugin package governance track: local package dry-run, checksums and risk reports, local repository scanning, local package install, config version history, sensitive config encryption, approval flow, signing/trusted-source draft, installed-plugin export, and the matching docs/E2E surface are now in place. See `docs/releases/v1.5.0.md`.

- Added the local plugin package governance track: package dry-run, checksum/risk reporting, repository scanning, install closure, config history, sensitive config encryption, approval flow, signing/trusted-source draft, and local plugin export, all with admin-facing UI and tests.
- Kept the package governance boundary strict: no zip uploads, no remote market, no remote install/download, no dynamic loading, no third-party code execution, no external SQL execution, and no user-data export.

## Unreleased

- Plugins: add `plugin_uninstall_tasks` and admin APIs for soft uninstall task tracking (`/admin/plugins/:code/soft-uninstall`, uninstall impact, list/detail/retry/delete). Soft uninstall archives plugin without deleting content/config/files. (v1.7.0-P0-07)

- Added a local plugin package spec draft (`docs/PLUGIN_PACKAGE.md`) and an admin dry-run API for scanning and previewing local plugin packages (`POST /api/v1/admin/plugins/packages/dry-run`); this is a safe read/validate/preview flow only (no install, no code/SQL execution, no dynamic frontend asset loading).
- Added an admin UI section under `/admin-next/plugins/install` for running local plugin package dry-run previews, plus a safe example package at `examples/plugins/demo_notice/`.
- Extended local plugin package dry-run with sha256 `checksums.json` verification, strengthened dangerous-file rules (symlink/executable/hidden dirs), and a backend-generated `risk_report` (low/medium/high/blocked) rendered in the admin UI.
- Added a local plugin package repository scanner (`GET /api/v1/admin/plugins/packages`) with filters/pagination plus a detail endpoint (`GET /api/v1/admin/plugins/packages/detail`), and surfaced the repository view under the admin install/upgrade page.
- Added a minimal local plugin package install flow (`POST /api/v1/admin/plugins/packages/install`) that re-runs dry-run server-side and installs manifest-only plugins as `disabled` with `source_type=local_package` (no code/SQL execution, no dynamic frontend asset loading).
- Added a draft plugin package signature + trusted publisher model: optional `publisher.json`/`signature.json`, local-only `storage/plugins/trusted_publishers.json`, Ed25519 verification over `sha256(raw checksums.json bytes)`, risk-report integration, and admin UI display in the package repository/dry-run/detail/install confirmation flows.
- Added plugin config version history for global/community plugin configs, including redacted diff views, version details, audit linkage via `config_version_id`, and a rollback **dry-run** preview API/UI (no actual rollback write).
- Added backend encryption for sensitive plugin config fields (AES-256-GCM with `enc:v1:` prefix) while keeping API/audit/history outputs redacted; supports placeholder retention (`[ENCRYPTED]`/`******`) and empty-string clearing.
- Added a minimal plugin approval flow for high-risk operations (install/upgrade): approval requests + admin approval center (`/admin-next/plugins/approvals`) + execution with mandatory re-dry-run checks; direct local package install and plugin upgrade execution now require `plugin.approve`.
- Refactored plugin governance backend structure by splitting plugin-related HTTP handlers and service methods into focused modules (no API/behavior changes); updated E2E compose defaults so Playwright runs against an internal `devhub` service (`DEVHUB_E2E_ORIGIN=http://devhub:8090` by default).
- Added installed declarative plugin export to local plugin package directories (`POST /api/v1/admin/plugins/:code/export/dry-run` and `/export`), generating `manifest.json`, README, redacted `config.example.json`, `checksums.json`, optional docs/migration/signature stubs, and an admin detail-drawer export panel; exports never include sensitive config, user data, runtime code, SQL, zip downloads, or remote publishing.

## v1.4.0 (2026-05-12)

Current `VERSION` is `v1.4.0`. The plugin-content governance enhancement work is now validated with Go tests/build plus Docker-based admin build and Playwright; see `docs/releases/v1.4.0.md`.

- Closed v1.4.0 acceptance: full Go + Docker build + admin/frontend Playwright + SEO curl checks pass (admin E2E `35 passed`, frontend E2E `17 passed`), and no long-term `test.skip/test.only` remains in Playwright suites.
- Began `v1.4.0` plugin-content governance enhancement work: `GET /api/v1/admin/posts` now supports precise `plugin_code + content_type` filtering and returns plugin ownership fields for admin post rows.
- Reworked the admin plugin area into function pages under `/admin-next/plugins/overview|list|content|install|config|dependencies|hooks|events|search-index|navigation|permissions|audit|developer`, while keeping legacy plugin routes compatible.
- Upgraded `PluginContent` into a fuller governance page with plugin name/code/status/health/type-count header, disabled/archived history notices, aligned filter/batch layout, result details, audit jump query metadata, and recent-governance entry from the detail drawer.
- Expanded PluginContent batch governance from hide/restore to approve/reject, pin/unpin, and feature/unfeature while keeping structured plugin audit metadata.
- Added minimal PluginContent E2E coverage for archived-history governance, precise plugin filtering, hide/restore regression, and a batch pin/unpin chain.
- Reorganized the documentation system so `docs/README.md` is the canonical entry point, `docs/PROJECT_PROGRESS.md` holds current state, `docs/PLUGIN_SYSTEM_ROADMAP.md` holds long-term and next-version plugin goals, and historical task material is archived under `docs/archive/`.
- Added `docs/releases/v1.3.5.md` as the next-stage draft for plugin-governance UI and install / upgrade wizard closure.
- Archived the old root `更新.md` product task document into `docs/archive/2026-05-09-product-requirements.md`; it is no longer a current acceptance source.
- Began Stage B plugin-governance experience work in the admin UI with `vue-i18n` and a default zh-CN dictionary for plugin-center wording, status labels, config panels, audit labels, and PluginContent actions.
- Completed another plugin-governance i18n cleanup pass for the plugin detail drawer, community plugin config drawer, PluginConfigEditor hints, PluginContent content statuses, and audit action labels; technical values such as `plugin_code`, `content_type`, Hook names, and JSON keys remain visible as raw values where useful.
- Upgraded plugin config editing from JSON-only to a basic schema-driven form mode plus JSON advanced mode, with effective-config preview and config-diff display.
- Added a plugin Hooks troubleshooting view in the admin plugin detail drawer, including hook stats, recent `hook_executions`, a filterable executions drawer with pagination, an execution-detail drawer, and audit-log jump links.
- Added plugin SDK/template documentation plus `go run ./cmd/devhub plugin:new` for generating manifest-only plugin skeletons that validate through the existing ManifestValidator and config schema checks.
- Enhanced the generic PluginContent governance page with content-type filtering, detail drawer, multi-select, batch hide, batch restore, and audit-log entry points while reusing the existing audited backend batch topic API.
- Connected PluginContent audit-log entry points to the generic audit log page with prefilled action, target type, and plugin metadata filters.
- Aligned Stage B documentation wording so basic schema-driven forms, effective config, config diff, PluginContent batch hide/restore, and audit-log jumps are treated as landed baseline capabilities, while deep schema support and advanced batch governance remain future work.
- Added plugin SDK/template documentation under `docs/plugins/`, including manifest, config schema, Hook, migration, permission, menu/route, and external ecosystem guides.
- Added built-in plugin lifecycle response fields and soft-uninstall archive/restore APIs; archived plugins block new content and community enablement while preserving history, config, migrations, audit logs, and SEO.
- Extended plugin status storage for `archived` and `migration_failed`, with matching MySQL schema/migration updates.
- Added frontend and admin E2E coverage for archived-plugin entry linkage: archived plugin content types disappear from publish pages, direct archived `content_type` submissions fail, community enablement is blocked, historical topic SEO remains accessible, and PluginContent can still govern archived historical content.
- Added a manifest validator plus admin validate/dry-run/install APIs for safe manifest + configuration-style plugins; installation records metadata and pending migrations but does not execute third-party code.
- Added plugin health summary APIs and bulk archive/restore APIs for the governance center.
- Extended dynamic content-type permission lookup so manifest-installed plugins can participate in the unified create-permission chain.
- Added a P2 upgrade dry-run preview API and UI branch so admins can inspect current/new version compatibility, changed keys, and diff without performing a real upgrade.
- Added the real plugin upgrade execution API and UI entry, preserving config/migration/audit history while updating version and manifest metadata.
- Documented that external-service webhooks, upgrade flows, plugin package upload/signing, remote marketplace, dynamic loading, sandboxing, hard uninstall, and migration down remain future work.
- Extended the admin plugin governance center UI with health summary cards, manifest validate / dry-run / install entry panels, bulk archive / restore actions, and a clearer runtime/archive-status banner in the plugin detail drawer.
- Moved plugin manifest validate / dry-run / install and upgrade/bulk result panels from inline page sections into drawers, keeping the same actions and test anchors while avoiding the plugin list being pushed into a nested long-scroll layout.
- Reworked `/admin-next/plugins` into a clearer governance layout with a page action header, list/status-governance views, compact summary cards, a filter panel, a bulk-action panel, and a reduced table action column.
- Turned manifest validate / dry-run / install and upgrade preview / execution into drawer-based step flows with structured validation, impact, compatibility, confirmation, and result panels.
- Added bulk archive / restore impact previews, succeeded / failed result tables, and audit-log jump actions in the plugin governance UI.
- Updated the admin plugin-governance E2E suite for the new action grouping and step-flow UI; latest admin check passes with `21 passed / 2 skipped`.
- Re-aligned plugin documentation with current code facts: `v1.3.5` is now treated as an implemented-but-unreleased governance closure, with remaining requirements reorganized into release cleanup, `v1.4` platform enhancement, and later plugin distribution work.
- Began v1.4.0-P1-10 unified plugin governance error codes: plugin endpoints now return `code/message/details/suggestion` while keeping legacy `error`, and permission denials on plugin routes expose `details.permission_code` for actionable fixes.
- Added admin `GET /api/v1/admin/plugins/:code/readiness` and a plugin-detail “操作诊断” tab to explain why enable/upgrade/config actions are blocked, including dependency/core/config_schema/migration checks.
- Enhanced the plugin-detail permissions tab with missing/high-risk highlights, filters, and reference lookup across menus/routes/content types.
- Began v1.4.0-P1-11 frontend entry/menu visibility governance: added `/api/v1/navigation`, `/api/v1/communities/:slug/navigation`, `/api/v1/communities/:slug/create-options`, and admin `/api/v1/admin/plugins/:code/menus/preview` for unified navigation/create gating and actionable “why hidden” diagnostics.

## v1.3.4

DevHub v1.3.4 is the plugin failure-governance and platform-foundation closure release.

- Added E2E/API-only failed plugin migration injection guarded by `DEVHUB_E2E_TESTING=1` or `CMS_STORE=memory`.
- Verified failed plugin migrations block both global plugin enablement and per-community plugin enablement until retry succeeds.
- Added audit coverage for migration failure injection, retry, and success recovery.
- Added admin Playwright coverage for the migration tab failure reason, retry action, restore flow, and plugin audit lookup.
- Added E2E/API-only HookBus failure injection guarded by `DEVHUB_E2E_TESTING=1` or `CMS_STORE=memory`.
- Verified blocking Hook failures block content creation without dirty writes and record `hook_executions` plus `plugin.hook.blocked` audit.
- Verified non-blocking Hook failures keep content creation successful while recording `hook_executions` plus `plugin.hook.failed` audit.
- Added admin Playwright coverage for Hooks tab failure summaries and plugin audit lookup.
- Tightened the plugin permission matrix around `ContentTypeDefinition.create_permission`; `post.create` now remains documented and tested only as a `core.topic.create` compatibility bridge, not as a plugin-content create permission.
- Added API tests proving `post.create` cannot create plugin-owned content, plugin create permissions can create their own content types, and frontend user tokens cannot call plugin governance APIs.
- Ran a dedicated MySQLStore / legacy-database upgrade pass for plugin platform schema, plugin migrations, hook executions, audit logs, global/community plugin state, failed migration readiness, and config schema validation.
- Hardened MySQL plugin upgrade migrations: `004_community_plugins.sql` now tolerates numeric-order execution before `005`, and `005_core_plugins.sql` now adds plugin fields idempotently.
- Added lightweight plugin health status reasons and Hook-derived `hook_warning` / `hook_error` summaries for the admin plugin governance center.
- Expanded plugin audit filtering by plugin code, community, action, actor, target, metadata, request id, and time range.
- Archived the v1.3.4 testing matrix into automated, partially automated, manual, uncovered, and skipped categories, and scoped P1 to plugin experience work rather than new plugin-market capabilities.
- Keep plugin marketplace, package upload/install, remote install/update, Go dynamic loading, and third-party sandboxing out of the current implementation scope.

## v1.3.3

DevHub v1.3.3 is the plugin platform governance closure release.

### Changed

- Added Service-level plugin enable readiness checks for both global and per-community enable actions.
- Plugin enable now checks plugin existence, global config schema validity, enabled dependencies, and failed plugin migrations before allowing `enabled`.
- Kept built-in pending up/no-op migrations non-blocking for enable, while surfacing them through plugin health and the migration tab; failed migrations block enable until retried or resolved.
- Clarified v1.3.3 documentation boundaries for lifecycle states, config schema validation, HookBus observability, migration no-op runner, plugin permissions, `post.create` compatibility, and admin plugin governance center coverage.
- Added `docs/releases/v1.3.3.md` and updated README, docs index, API, architecture, testing, project progress, changelog, and VERSION to the v1.3.3 release line.

### Known Limitations

- Plugin lifecycle states are accepted by schema/Store but still do not form a full automatic state machine.
- Plugin migrations remain built-in up/no-op records; migration down, true rollback, pre-migration backup, and external plugin migration packages remain follow-up work.
- HookBus remains for built-in plugins only; third-party dynamic hooks, webhooks, remote execution, sandboxing, and plugin marketplace capabilities are not implemented.

## v1.3.2

DevHub v1.3.2 is the plugin platform governance enhancement release.

### Changed

- Calibrated plugin-platform documentation to distinguish completed capabilities, partial capabilities, reserved concepts, and future roadmap items before continuing new plugin work.
- Moved HookBus into the plugin platform layer (`internal/plugins`) and registered minimal built-in hook handlers for system plugins.
- Added `hook_executions` runtime records for built-in HookBus execution, plus `/api/v1/admin/plugins/:code/hooks` for Hook statistics and recent executions.
- Recorded blocking Hook failures as `plugin.hook.blocked` and non-blocking Hook failures as `plugin.hook.failed` audit entries.
- Added lightweight plugin health summaries to admin plugin responses and the admin plugin governance UI.
- Added `/api/v1/admin/plugins/:code/audit-logs` for plugin-scoped audit queries.
- Added structured `plugin_code` audit metadata for plugin content governance actions such as hide/restore, pin/feature, comment governance, and batch topic moderation.
- Enforced `config_schema` validation when saving plugin `config_json` (both global `plugins.config_json` and per-community `community_plugins.config_json`).
- Added schema default values to `resolved_config.effective` and recorded plugin config audit diffs via `metadata_json.changed_keys`.
- Added `plugin_migrations` table (schema + migration) for tracking plugin migration execution state.
- Added built-in plugin migration declarations for qa/docs/wiki, plus admin APIs and UI for listing, running, and retrying first-stage up/no-op migrations.
- Recorded plugin migration run/retry/success/failure actions in structured audit logs.
- Enhanced `/admin-next/plugins` towards a plugin governance center baseline UI (stats cards, filter toolbar, clearer status/capability badges).
- Upgraded the admin plugin detail drawer into a tabbed governance view and replaced the global-config textarea with a JSON editor powered by `json-editor-vue` + `Ajv` client-side schema validation.
- Upgraded the admin community plugin drawer with filtering, clearer status/override indicators, and a JSON editor for `community_plugins.config_json` powered by `json-editor-vue` + `Ajv` schema validation.
- Added lightweight plugin impact analysis endpoints and surfaced impact hints in disable confirmations; added an audit tab to the admin plugin detail drawer (backed by `admin/audit-logs`) and improved the generic PluginContent page with community/status filters.
- Extended the plugin status model beyond `enabled` / `disabled` to support governance states such as `config_invalid` and `migration_pending`, while keeping content creation strictly gated on global `enabled` plus community `enabled`.
- Expanded plugin impact analysis counts to include existing contents, enabled/disabled communities, recent contents, pending contents, config overrides, and pending migrations; disable confirmations now surface the richer impact context without implying historical content or SEO deletion.
- Archived a plugin-governance acceptance pass covering Go tests/build, Docker Node admin build, impact APIs, audit logs, config schema failures, global/community plugin state limits, moderator menus, and `/topics/:id` SEO regression.
- Added a fixed Docker-based admin Playwright E2E runner (`admin-e2e`) using `mcr.microsoft.com/playwright:v1.59.1-noble`, with containerized admin build and a minimal plugin-governance browser test suite.
- Added a fixed Docker-based frontend Playwright E2E runner (`frontend-e2e`) with containerized frontend build and a first-stage public navigation / SEO smoke suite.
- Expanded admin Playwright E2E coverage from the plugin-governance center to login, content, comments, communities, tags, and audit-log smoke paths.
- Archived a plugin-system acceptance pass with Go tests/build, admin Docker build, frontend 14-test E2E pass, admin 15-test E2E pass, and `/topics/:id` / `/c/:slug` SEO curl regressions.
- Fixed admin plugin-governance E2E state isolation by restoring globally disabled plugins in `finally` and aligning impact-dialog assertions with the current real impact fields.

### Baseline Notes

- Current plugin runtime state is still based on `plugins.status`, `community_plugins.status`, and `plugin_migrations.status`; the expanded statuses are accepted by schema/Store, but only global `enabled` is publish-enabled until a full lifecycle state machine is implemented.
- `plugin_migrations` now supports built-in up/no-op migration listing, execution records, failed-state retry, audit, and an admin migration tab; migration down, real rollback, pre-migration backup, and external plugin migration packages remain follow-up work.
- HookBus dispatch exists for built-in plugins and now persists execution records/statistics; retry policy, alerting, and external monitoring remain follow-up work.
- Plugin health is a lightweight governance summary, not a Prometheus/Grafana-style monitoring system.
- Plugin config diff is currently top-level `changed_keys`; deep-path diff, version history, rollback, and gray release remain follow-up work.

## Next

- (empty)

## v1.3.1

DevHub v1.3.1 is the plugin-entry hardening and permission-boundary release.

### Changed

- Reframed the complete plugin system as the highest-priority long-term roadmap, split into P0 platform closure, P1 platform enhancements, P2 plugin distribution, and P3 advanced runtime capabilities.
- Sealed `Service.CreatePost` as a legacy/deprecated business entry so normal writes must go through `Service.CreateTopic` and plugin publishing validation.
- Kept `/api/v1/posts` write endpoints deprecated with `410 Gone`; read compatibility remains.
- Hardened `POST /api/v1/admin/posts` with dynamic plugin create permission checks on top of the legacy base `post.create` gate.
- Forbid normal admin content editing from changing site/community, board/category, `content_type`, or `plugin_code`; ownership/type migration must be handled by a future migration-specific workflow.
- Disabled site and board selectors in the admin content edit form to match the backend ownership-change policy.
- Added plugin create permissions to demo site-admin role seeds for MemoryStore and MySQLStore.
- Standardized plugin platform contracts with manifest/content-type/permission/menu/route/hook structure tests.
- Added trusted server-side `ActorContext` injection for topic creation instead of trusting request-body permissions.
- Added writable global `plugins.config_json`, admin config API, config merge view, and public API config scrubbing.
- Expanded the minimal internal HookBus call points to content create/update/delete, comment creation, search, notification, and SEO events.
- Added structured plugin audit fields (`old_value`, `new_value`, `metadata_json`) and writes for plugin status/config/sort governance actions.
- Made the admin plugin page consume manifest-declared admin menu paths instead of hardcoded plugin route maps.
- Improved the admin global plugin page with an explanatory card, status badges, capability summaries, a tabbed plugin-detail drawer, JSON config/schema display, and clearer enable/disable confirmations.
- Improved the admin community plugin drawer with global/community status badges, enablement summaries, disabled-reason hints, schema reference display, JSON formatting/validation, and reliable sort-order updates.
- Added tests for plugin mappings, config JSON validation, public config hiding, plugin audit logs, and moderator plugin-menu scope filtering.

### Known Limitations

- `post.create` remains a compatibility bridge for `core.topic.create`; it is not the long-term primary permission.
- HookBus is still minimal: search/notification/SEO currently dispatch events but do not yet have full plugin business handlers, retry, or unified error logging.
- `plugins.config_json` and `community_plugins.config_json` validate JSON syntax only; `config_schema` enforcement remains follow-up work.
- The improved admin plugin UI still needs a real browser acceptance matrix; there is no automated browser test runner in the repo yet.
- Plugin impact analysis currently provides lightweight count-only endpoints; affected-object detail lists (e.g. impacted category IDs) are still follow-up work.
- Non-plugin historical audit logs may still only have `admin_logs.target` text summaries.
- `project`, `job`, and `ai_work` are plugin-owned but still lack dedicated extension tables and full business workflows.
- Plugin packages, marketplace, remote install/update, and dynamic loading are not implemented in v1.3.1; they are staged as P2/P3 plugin-platform roadmap items rather than permanent exclusions.

## v1.3.0

DevHub v1.3.0 is the Core + Plugins architecture split release.

Current status: v1.3.0 is code-level integrated for built-in `qa`, `docs`, `wiki`, `projects`, `jobs`, and `ai_works` system plugins, global plugin state, per-community plugin state, and publishing validation. Some browser acceptance, plugin-specific product workflows, and fine-grained permission matrices remain follow-up work and are listed below.

### Added

- Built-in plugin registry with `qa`, `docs`, `wiki`, `projects`, `jobs`, and `ai_works` system plugins.
- MySQL `plugins` table with `installed`, `enabled`, and `disabled` states.
- Per-community plugin enablement via the `community_plugins` table.
- `topics.plugin_code` plus `categories.plugin_code` and `categories.allowed_content_types`.
- Plugin-owned tables: `qa_questions`, `qa_answers`, `docs_spaces`, `docs_documents`, `wiki_spaces`, `wiki_pages`, and `wiki_page_versions`.
- Admin plugin APIs and lightweight admin-next plugin management / plugin content entries.
- Public community plugin API, admin community plugin APIs, and moderator plugin menu API.

### Changed

- `question`, `document`, `wiki_page`, `project`, `job`, and `ai_work` are now owned by `qa`, `docs`, `wiki`, `projects`, `jobs`, and `ai_works` plugins rather than hardcoded as Core-only types.
- Topic publishing validates category plugin binding, global plugin status, per-community plugin status, and allowed content types.
- Legacy `doc` / `wiki` request values are normalized to `document` / `wiki_page` for compatibility.
- `project`, `job`, and `ai_work` have plugin ownership, publish validation, permissions, and menus; plugin-specific extension tables and full product workflows remain follow-up work.

### Known Limitations

- Plugin marketplace, package upload, remote update, and dynamic loading are not v1.3.0 implementation scope; they are staged in the longer plugin-platform roadmap.
- Plugin route loading is currently registry metadata plus Core dispatch; dynamic route/runtime loading is a later-stage platform capability.
- Dedicated Docs tree editing UI and Wiki collaboration / rollback UI remain follow-up work.
- Community plugin `config_json` and sort have a minimal admin UI, but still need full browser matrix acceptance and stronger product polish.
- Publishing currently enforces minimal permission-code checks for plugin-owned types. Core-compatible `article` and `news` still use a coarse permission (`core.topic.create`, compatible with legacy `post.create`) and do not yet support fine-grained per-type permission matrices.

## v1.2.1

DevHub v1.2.1 is the tag governance enhancement release.

### Added

- Tag aliases with admin CRUD APIs, alias-based suggestion matching, alias URL resolution, and audit-log writes.
- Tag merge APIs and admin-next merge UI, including topic-tag migration, follow deduplication, merged status, and merged-target tracking.
- Tag statistic recalculation for single-tag and all-tag operations in MemoryStore and MySQLStore.
- MySQL schema support for `tags.merged_to_id`, `tags.hot_score`, and the `tag_aliases` table.

### Changed

- Public tag resolution now normalizes direct slug, alias slug, and merged source tags to the canonical target tag.
- Merged and disabled tags no longer enter sitemap, and alias URLs are not emitted as sitemap entries.
- Tag SEO pages prefer 301 redirects to canonical target URLs for alias and merged-source access.
- admin-next tag management now exposes alias management, merged status, merge target selection, and statistic recalculation.

### Known Limitations

- Tag trend analytics, operator dashboards, and large-scale async recalculation jobs are still out of scope.
- AI-assisted tag recommendations remain planned for a later release.

## v1.1.5

DevHub v1.1.5 is the frontend UI polish release.

### Changed

- Unified frontend visual tokens for colors, typography, spacing, borders, shadows, radius, and responsive page width.
- Polished the frontend header, site switcher, search box, logged-in user menu, publish button, and moderator workspace entry.
- Improved homepage, community pages, topic cards, topic detail typography, search page, publish page, user-center pages, notification cards, empty states, and lightweight moderator workspace visuals.
- Added responsive refinements for desktop, tablet, and mobile layouts.

### Unchanged

- No API, Store, database schema, route, auth, follow, publish, comment, moderator-permission, or admin-next business logic was changed.
- `/topics/:id`, `/c/:slug`, and `/tags/:tag` remain Go-rendered SEO pages and were not converted to CSR shells.

## v1.1.4

DevHub v1.1.4 is the frontend login-state and permission-entry fix release.

### Fixed

- Frontend user login state is restored consistently across header, Go-rendered community pages, and Go-rendered topic pages.
- Community follow, favorites, follows, activities, and notifications pages now send the frontend user token and no longer misreport logged-in users as unauthenticated.
- Normal frontend users no longer see the full admin-next backend entry in the frontend user menu.
- Community moderators see the `/moderator` workspace entry based on `is_moderator`.
- Publishing `question` content now matches the question category instead of defaulting to an article category.
- The admin-next menu now exposes only one community management entry.

### Changed

- `GET /api/v1/auth/me` now includes `is_moderator` and `moderated_communities`.
- `/admin-next/sites` remains as a hidden compatibility route and redirects to `/admin-next/communities`.
- Topic creation validation now prefers `categories.content_type` and falls back to legacy `categories.type`.

## v1.2.0

DevHub v1.2.0 is the tag system enhancement release.

### Added

- Go-rendered Baidu-friendly tag aggregation SEO pages at `/tags/:tag/`.
- Go-rendered community tag aggregation SEO pages at `/c/:communitySlug/tags/:tag/`.
- Public tag detail, tag-topic aggregation, and tag suggestion APIs.
- Public community tag detail and community tag-topic aggregation APIs.
- Tag follow UX on the tag SEO page using existing `POST /api/v1/follows/toggle`.
- Publish-page tag suggestions scoped to the selected community.
- admin-next tag management at `/admin-next/tags`, including CRUD, enable/disable, SEO fields, and related-topic viewing.
- MySQL schema and startup migration support for tag `follower_count`, SEO fields, and `enable/disable` status.
- Dynamic sitemap entries for enabled global tags and enabled community tag pages.

### Changed

- Topic and community tag links now point to canonical tag pages instead of only search filters; community context links use `/c/:communitySlug/tags/:tag/`.
- Tags are first-class manageable records in MemoryStore and MySQLStore, while still preserving existing topic tag behavior.
- `/topics/:id` and `/c/:slug` SEO output remains Go-rendered and unchanged in responsibility.

### Known Limitations

- Tag merge, tag aliasing, and tag trend statistics are still not part of v1.2.0.
- Tag custom redirect/canonical migration after future merges remains planned.
- Sitemap is still a single dynamic file and is not yet sharded.

## v1.1.3

DevHub v1.1.3 is the independent moderator workspace MVP release.

### Added

- Independent frontend moderator workspace at `/moderator`, `/moderator/reports`, `/moderator/topics`, `/moderator/comments`, and `/moderator/audit-logs`.
- Dedicated `/api/v1/moderator/*` APIs that use frontend `users` tokens and `community_moderators` scope checks.
- Moderator dashboard, managed community list, scoped reports, scoped topics, scoped comments, and scoped audit-log views.
- Moderator actions for handling reports, feature/unfeature, pin/unpin, hide/restore topics, lock/unlock comments, and hide/restore comments.
- Moderator audit-log writes with `actor_type=moderator`, `actor_id=users.id`, and community scope.

### Changed

- The frontend user menu now links to the independent moderator workspace.
- Moderator governance no longer needs to enter the full admin-next UI for the MVP workflow.
- MemoryStore frontend registration now creates a real user with a bcrypt password hash, so newly registered accounts can log in with their own password.

### Known Limitations

- Complex RBAC, permission matrix editing, moderator tenure, and performance statistics remain out of scope.
- The moderator workspace is a lightweight runtime API page and not a full replacement for admin-next.
- Super admins should continue to use admin APIs and admin-next for full-system governance.

## v1.1.1

DevHub v1.1.1 is the frontend/admin identity boundary cleanup release.

### Added

- Scoped JWT claims for frontend user tokens and backend admin tokens.
- Separate frontend user login and backend admin login flows.
- Moderator-scoped admin API access for enabled community moderators.
- `actor_type` and `actor_id` audit-log fields for admin, moderator, and system actions.
- Identity-boundary test cases and v1.1.1 release documentation.

### Changed

- `/api/v1/admin/*` now uses backend admin identity by default, with explicit scoped moderator allowance.
- Frontend token storage now prefers `devhub_user_token` and `devhub_user_refresh_token`, while keeping compatibility with old keys.
- Audit log UI now displays and filters actor identity type.

### Known Limitations

- MemoryStore still uses demo seed users for local development.
- MySQL refresh tokens still use `user_id` with `token_type` to distinguish `users.id` from `admin_users.id`.
- Admin-user to frontend-user binding is left for a later productionization pass.

## v1.1.0

DevHub v1.1.0 is the sub-site module enhancement release. It upgrades communities from a simple content filter into independent community spaces with their own profile, SEO, boards, moderators, stats, follow state, and announcements.

### Added

- Enhanced community profile fields: logo, cover image, slogan, theme color, SEO title/description/keywords, counters, hot score, and announcement fields.
- Go-rendered Baidu-friendly community SEO pages for `/c/:slug`, including title, description, canonical, h1, board links, topic links, tag links, stats, moderators, and follow action.
- `/site/:slug` compatibility redirect to canonical `/c/:slug/`.
- Community stats API, public community moderator API, enhanced community tags/categories responses, and community follow counter updates.
- admin-next community management page at `/admin-next/communities`, with create/edit, enable/disable, sort order, SEO fields, announcement fields, frontend links, moderator links, and board management.
- Admin community and category CRUD APIs, reorder APIs, status APIs, and audit-log writes for community/board changes.
- MemoryStore and MySQLStore support for enhanced communities, enhanced categories, community stats, community sitemap filtering, public moderators, and community follow counts.
- Sitemap entries for enabled communities such as `/c/php/`, `/c/go/`, `/c/java/`, `/c/ai/`, and `/c/frontend/`.
- v1.1.0 release documentation and test matrix.

### Changed

- Community pages are now treated as first-class spaces instead of only content-list aliases.
- `/c/:slug` is the canonical community URL; `/site/:slug` remains compatible by redirecting.
- MySQL schema and startup migration helpers now include v1.1.0 community/category fields.
- README, API, SEO, deployment, testing, and project progress docs are updated for the v1.1.0 scope.

### Known Limitations

- v1.1.0 uses enabled categories as the default community navigation; deeper custom navigation is left for a later release.
- Advanced tag features such as aliases, merging, and tag admin were planned for v1.2.0 and are now delivered in v1.2.0–v1.2.1; tag trend statistics remains out of scope.
- A complete followed-community feed remains planned for a later release; this release completes follow state, follower count, activities, and "my follows" visibility.
- Comment likes, canceling solved status, recommendation algorithms, reputation, and complex analytics are outside this release.
- Sitemap output is still single-file dynamic output and is not yet sharded for very large installations.

## v1.0.0

DevHub v1.0.0 is the first runnable archive release of the project.

### Added

- Multi-community DevHub structure for Portal, PHP, Go, Java, AI, and Frontend communities.
- Topic publishing, listing, detail, editing, and compatibility `sites/posts` APIs.
- Search and filtering by keyword, community, content type, tag, status, featured, and unsolved questions.
- Basic tag capability, hot tags, and community tag aggregation.
- Topic likes, favorites, follows, user activities, and notifications.
- Comments, replies, question solved state, best-answer acceptance, and unsolved filtering.
- Topic and comment reports, moderator-scoped governance, featured, pinned, hidden, restored, and comment-lock moderation.
- admin-next backend for content CRUD, comments, reports, moderator CRUD, batch governance, and audit logs.
- Baidu-friendly dynamic SEO HTML for `/topics/:id`, dynamic sitemap, and robots.txt.
- MemoryStore and MySQLStore modes with MySQL 8 schema coverage.
- Testing, deployment, SEO, backup, rollback, and release archive documentation.
- Basic GitHub Actions CI for Go tests/builds, frontend build, admin build, SQL checks, and docs checks.

### Changed

- Default port is unified to `8090` across `main.go`, `dev.sh`, README, docs, and admin dev proxy.
- `/topics/:id` keeps Go-rendered SEO HTML and uses the current Astro CSS asset from the build output.
- admin-next frontend API wrappers are aligned with backend moderator, batch governance, report, and audit-log routes.
- Documentation has been consolidated around the v1.0.0 runnable release scope.

### Known Limitations

- Advanced tag features were planned for v1.2.0; tag detail SEO pages and tag admin landed in v1.2.0, while aliases/merging/recalculation landed in v1.2.1. Tag trend statistics remains out of scope.
- Runtime comment likes are not part of v1.0.0.
- Accepted questions support changing the best answer, but do not yet support canceling solved status.
- Tag-follow and user-follow backend support exists, while richer frontend entry points remain future work.
- Sitemap output is dynamic but not yet sharded for very large content volumes.
- Production deployment still needs environment-specific process supervision, reverse proxy, HTTPS, logging, and backup scheduling.

## v1.4.0-P1-07

### Added

- Plugin dependency checks now support structured `dependencies` with `code`, `version`, `required`, and `reason`, while keeping legacy string dependencies compatible.
- Manifest validate, dry-run, install, upgrade dry-run, upgrade, and enable now share dependency and Core compatibility blocking rules.
- Admin plugin install / upgrade flows and detail drawer now show dependency matrix, Core compatibility, blocking reasons, and dependency diff.
- Added backend dependency/version tests and `web/admin-app/tests/e2e/plugin-dependencies.spec.js`.

### Notes

- Version constraints are intentionally lightweight: exact `x.y.z`, comparison operators, and whitespace-combined ranges only.
- DevHub still does not support automatic dependency installation, plugin marketplace, remote install, dynamic loading, script sandbox, plugin signing, migration down, or hard uninstall.

## v1.6.0-P0-04

### Added

- Added readonly remote plugin index sources, fetch, plugin list/detail APIs, SSRF-safe URL validation, trust/compatibility/risk metadata, and admin UI at `/admin-next/plugins/remote-indexes`.
- Added `docs/PLUGIN_REMOTE_INDEX.md` and `docs/examples/plugin-remote-index.example.json`.

### Notes

- Remote indexes are metadata-only: DevHub does not download remote packages, install remote plugins, execute code, dynamically load assets, or auto-trust remote publishers in this phase.

## v1.6.0-P0-05

### Added

- Added plugin package version repository APIs and admin page to aggregate installed, local package, uploaded package, and remote index versions.
- Added structured upgrade diff previews with grouped `diff_sections`, high-risk/blocking rules, sensitive value redaction, and upgrade approval handoff.

### Notes

- Remote index versions remain readonly metadata and cannot be upgraded directly; DevHub still does not auto-upgrade, download remote packages, execute plugin code, execute SQL, or dynamically load assets.
