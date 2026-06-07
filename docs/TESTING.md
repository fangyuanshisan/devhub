# DevHub 测试文档

[返回文档入口](README.md)

更新时间：2026-05-30

本文档记录当前仓库滚动测试目标、已执行验收记录和后续补测项。历史版本测试只保留必要回归，不再展开旧版本完整矩阵。

## 手动测试入口

Codex / Agent 默认不要在完成任务时自动跑测试、E2E 或完整验收；测试由开发者按需手动执行。完整手动入口：

```bash
./scripts/test-all.sh
```

常用变体：

```bash
./scripts/test-all.sh --quick
./scripts/test-all.sh --go-only
./scripts/test-all.sh --frontend-only --quiet
```

## v1.9.0 发布后文档阶段

当前稳定版本为 `v1.9.0` 官方插件生态稳定版。发布归档与发布后 targeted smoke 已完成，结论为 P0=0，P1=0，P2=0，版本可以冻结；建议先提交当前文档归档，再创建 `v1.9.0` annotated tag。v1.9.0 冻结后不再追加新功能。

下一阶段测试口径：`v1.9.1` 作为维护候选，优先承载小修、文档、测试脚本固化和兼容性补丁；`v1.10.0` 作为规划阶段，后续若进入实现需重新定义对应的手工 / 自动验收矩阵。本文不把 `v1.9.1` 写成已发布，也不把 `v1.10.0` 写成已实现。

本轮文档阶段更新只要求轻量一致性检查；Go 测试、Go build、前后台 quick、MySQL full smoke 和 `scripts/test-all.sh` 不因纯文档更新自动执行，必要时由开发者按需手动运行。

本轮实际检查：`git diff --check` 通过；stale phrase grep 无输出；状态词 grep 的基础 `grep` 形式无输出，已用等价 `rg` 正则确认 release-facing 文档包含 `v1.9.0 官方插件生态稳定版`、`P0=0`、`P1=0`、`P2=0`、`v1.9.1` 和 `v1.10.0`。

## v1.9.0-S11 发布归档与发布后 smoke 测试记录

本轮目标：完成 v1.9.0 发布归档确认、tag 建议、发布后轻量 smoke、最终文档口径一致性和下一版本规划。S11 不新增功能、不改变业务逻辑 / DB schema / 插件语义 / SecretCenter / external_service / Webhook Secret / Callback Token 安全模型。

执行结论：P0=0，P1=0，P2=0。`v1.9.0` 可以冻结；建议先提交当前 release archive，再创建 tag `v1.9.0`。

已执行命令与结果：

- `git status --short`：初始为空。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260530-131149/`，仅有既有 Vite chunk size warning。
- `bash scripts/check-admin-plugin-ia.sh`：通过，截图目录 `.devhub/screenshots/plugin-ia`。
- `git diff --check`：通过。
- 发布一致性 grep：覆盖 `v1.9.0`、`v1.8.4`、`blocking Hook`、`插件市场`、`远程在线安装`、`DEVHUB_PLUGIN_CONFIG_KEYS`、`encrypted_value`、`token_ref`、`SecretCenter`、`MySQLStore`、`P0`、`P1`、`P2`，未发现与 v1.9.0 稳定版边界冲突的 release-facing 文案。

发布归档确认：

- `VERSION = v1.9.0`。
- `docs/releases/v1.9.0.md` 包含稳定版定位、最终能力清单、S1-S10 acceptance、S11 smoke、安全边界、`P0/P1/P2=0`、freeze / release recommendation。
- `CHANGELOG.md` 包含 `v1.9.0` 工作线记录。
- `README.md` 当前版本为 `v1.9.0` 稳定版，不暗示 dynamic runtime / marketplace / remote install / blocking Hook。
- `docs/PROJECT_PROGRESS.md` 记录可冻结 / 可发布、`P0/P1/P2=0` 和下一版本建议。
- 本文记录 S10 acceptance、S11 release check commands、MySQLStore acceptance、admin quick、敏感扫描、未执行检查与原因。

MySQLStore / admin smoke 说明：S11 已执行 admin quick 与 admin IA smoke。MySQLStore full smoke 未重复执行；S9 已完成 MySQLStore 生产模式专项、Webhook full flow、后台 IA、DB / 文件敏感扫描，且 S11 未修改 Store、migration、DB schema、external_service、SecretCenter 或 Webhook 投递逻辑。

未执行检查及原因：

- Go `gofmt` / `go test ./...` / `go build ./...`：未执行，S11 未修改 Go。
- `bash -n scripts/check-admin-plugin-ia.sh` / `bash -n scripts/run-feishu-webhook-flow.sh`：未执行，S11 未修改脚本；S10 已执行并通过。
- `./scripts/check-frontend.sh --frontend-only --quick`：未执行，S11 未修改前台源码、SEO 路由或 sites/posts 兼容 API。
- MySQLStore full smoke / receiver full flow：未执行，S9 已覆盖且 S11 未改变相关逻辑。
- `scripts/test-all.sh`：未执行，S11 按发布归档 / post-release smoke 范围执行 targeted checks。

安全边界确认：未执行第三方插件代码；未开放 Go plugin、JS sandbox、WASM、Lua、远程 iframe、remote component、任意远程 JS、插件市场、远程在线安装、blocking Hook、package scripts、自动安装或自动启用；未回显 token / secret / Authorization / root key / `encrypted_value` / `secret_hash`；未改变 Webhook Secret / Callback Token 明文一次性返回安全模型、SecretCenter 加密逻辑、`migrations/` 唯一入口、dry-run no-SQL、upload / promote / install / upgrade 语义、`/topics/:id` / `/c/:slug` SEO 或 sites/posts 兼容 API。

## v1.9.0-S10 发布候选总验收与冻结判断测试记录

本轮目标：对 v1.9.0 release candidate 做最终验收和冻结判断，复核 S1-S9 验收记录、版本口径、文档一致性、Go / 后台 quick 检查、敏感信息边界和 P0/P1/P2 状态。S10 不新增功能。

执行结论：P0=0，P1=0，P2=0。建议冻结 `v1.9.0`，建议人工确认当前 diff 后创建 tag `v1.9.0`。

已执行命令与结果：

- `git diff --check`：通过。
- `docker run --rm -v "$PWD:/workspace" -w /workspace golang:1.22-bookworm gofmt -w internal/service/webhook_secrets.go`：通过。
- `bash -n scripts/check-admin-plugin-ia.sh`：通过。
- `bash -n scripts/run-feishu-webhook-flow.sh`：通过。
- `docker run --rm -v "$PWD:/workspace" -w /workspace golang:1.22-bookworm go test ./...`：首次使用默认 `proxy.golang.org` 下载模块时出现 EOF，未进入编译。
- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -e GOSUMDB=sum.golang.google.cn -v "$PWD:/workspace" -w /workspace golang:1.22-bookworm go test ./...`：通过。
- `docker run --rm -v "$PWD:/workspace" -w /workspace golang:1.22-bookworm go build -buildvcs=false ./...`：首次使用默认 `proxy.golang.org` 下载模块时出现 EOF，未进入编译。
- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -e GOSUMDB=sum.golang.google.cn -v "$PWD:/workspace" -w /workspace golang:1.22-bookworm go build -buildvcs=false ./...`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260530-123002/`，仅有既有 Vite chunk size warning。
- 发布一致性 grep：已复核 `v1.9.0`、`v1.8.4`、`blocking Hook`、`插件市场`、`远程在线安装`、`DEVHUB_PLUGIN_CONFIG_KEYS`、`encrypted_value`、`token_ref`、`SecretCenter`、`MySQLStore` 在发布相关文档中的口径。

S9 未实跑项的 S10 复核：

- upgrade warning confirm / 真实执行 / blocked confirm：S10 `go test ./...` 覆盖 `TestPluginUpgradeDryRunStructuredPlansAndConfirm`、`TestPluginUpgradeDryRunBlocksSameVersionEvenWithConfirm`、`TestUpgradePluginManifest` 和 `TestPluginUpgradeFromCompatRequiresConfirmForWarning`。
- cleanup installed / enabled 跳过：S10 `go test ./...` 覆盖 `TestCleanupPluginPackageRepository_TestPackagesPreviewExecuteAndSkipsInstalled`。
- Webhook Secret 缺 root key readiness：既有 v1.8.4-S14 无 root key MySQL 临时后端已验收返回 `webhook_secret_encryption_key_missing` 与启动 env 建议；S10 未改变该逻辑，只改 Webhook Secret API 响应脱敏。

未执行检查及原因：

- `./scripts/check-frontend.sh --frontend-only --quick`：未执行，S9/S10 未修改前台源码、SEO 路由或 sites/posts 兼容 API。
- MySQLStore full smoke / receiver full flow：未重跑，S9 已完成 MySQLStore 生产模式专项和敏感扫描；S10 未再修改 Store、migration、DB schema、external_service、SecretCenter 或 Webhook 投递逻辑。
- `bash scripts/check-admin-plugin-ia.sh`：S10 未重跑浏览器矩阵，S9 已执行并归档 `.devhub/screenshots/plugin-ia`；S10 未修改前端或 IA 脚本。
- `scripts/test-all.sh`：未执行；S10 按发布冻结矩阵执行 targeted checks，完整一键手工入口仍是 `./scripts/test-all.sh`。

安全边界确认：未执行第三方插件代码；未开放 Go plugin、JS sandbox、WASM、Lua、远程 iframe、remote component、任意远程 JS、插件市场、远程在线安装或 blocking Hook；未回显 token / secret / Authorization / root key / `encrypted_value` / `secret_hash`；未改变 Webhook Secret / Callback Token 明文一次性返回安全模型、SecretCenter 加密逻辑、`migrations/` 唯一入口、dry-run no-SQL、upload / promote / install / upgrade 语义、`/topics/:id` / `/c/:slug` SEO 或 sites/posts 兼容 API。

## v1.9.0-S9 MySQLStore 生产模式总验收记录

本轮目标：以 MySQLStore 为主，对 v1.9.0 官方插件生态稳定版做生产模式总验收，重点确认插件安装记录、配置记录、SecretCenter、`token_ref`、`hook_executions`、`admin_logs`、HTTP Allowlist、cleanup、模板生成、upgrade dry-run、Webhook Secret、Callback Token、后台点击矩阵和敏感信息扫描。

执行结论：P0=0，P1=0，P2=3。MySQLStore 核心链路通过；未新增功能、未改变安全模型、未执行第三方插件代码。验收中发现 Webhook Secret 创建响应会在 `secret_record` 中回显 `secret_hash`，已在 service 层修复为所有治理 API 响应统一清除 `secret_hash/secret_ciphertext`，DB 内部哈希仍用于验签。

已执行命令与结果：

- `./dev.sh start --mysql`：通过。MySQL ready，前后台 build 通过；首次发现 8090 已有 DevHub，脚本复用既有服务。`curl /api/v1/health` 后续确认当前服务为 `store=mysql`。
- `SUDO=/bin/false DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18084 ./dev.sh start --mysql --no-build`：通过。用于 Webhook full flow，避免 `restart` 在 root-owned 旧进程清理时触发交互 sudo。
- `curl -s http://127.0.0.1:8090/api/v1/health`：通过，返回 `store=mysql`、`database=devhub`。
- `git diff --check`：通过。
- `docker run --rm -v "$PWD:/workspace" -w /workspace golang:1.22-bookworm gofmt -w internal/service/webhook_secrets.go`：通过。
- `SUDO=/bin/false DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18084 ./dev.sh start --mysql`：修复后通过，重新完成前后台 build 并启动 MySQLStore 后端。
- `bash -n scripts/check-admin-plugin-ia.sh`：通过。
- `bash -n scripts/run-feishu-webhook-flow.sh`：通过。
- `DEVHUB_PLUGIN_CODE=official_webhook_notify DEVHUB_WEBHOOK_FLOW=full DEVHUB_WEBHOOK_AUTH_TYPE=bearer ... ./scripts/run-feishu-webhook-flow.sh`：通过。success、500 fail、timeout、manual retry 均完成，receiver 收到 11 次投递。
- `bash scripts/check-admin-plugin-ia.sh`：通过，报告 `.devhub/screenshots/plugin-ia/report.json`，截图目录 `.devhub/screenshots/plugin-ia`。
- HTTP Allowlist API smoke：通过。默认、环境变量、后台新增 / 删除、最终生效列表可查；wildcard、`0.0.0.0`、CIDR、path、query、fragment、userinfo、非 HTTP scheme 均返回 400。
- 模板 preview / generate / export API smoke：通过。preview 不落盘；generate 写 `storage/plugins/drafts/{code}`；export zip 10 个 entry，包含 `migrations/001_init.sql`，forbidden files 为 0。
- cleanup preview / cleanup API smoke：通过。受控模板 draft promote 到本地仓库后，preview 返回 confirm_token；cleanup 删除未安装候选并跳过当前安装包；写入 `admin_logs`。
- upgrade dry-run API smoke：通过 same-version blocked、version bump safe、cross-major warning 和 blocked confirm 绕过拒绝；结构化 plans 可见。warning 真 confirm、confirm_token 过期和真实 upgrade 未执行，列 P2。
- Webhook Secret create / rotate / disable / revoke API smoke：通过。明文只在创建 / 轮换响应中一次性返回，列表 / 详情 / 状态变更 / 非明文字段无 Authorization / `encrypted_value` / `secret_hash` / `secret_ciphertext` / token_hash。
- Callback Token create + callback API smoke：通过。创建响应一次性返回明文；`GET /api/v1/plugin-callback/config` 与 `POST /api/v1/plugin-callback/audit-events` 均返回 200；非明文字段无 token_hash / 明文。
- MySQL 敏感扫描：通过，详见下方。
- `.devhub/checks`、`.devhub/screenshots/plugin-ia`、模板 drafts/export zip 敏感扫描：无命中。

MySQLStore 验收项：

| # | 验收项 | 结果 | 证据 / 说明 |
|---|---|---|---|
| 1 | MySQLStore 启动成功 | pass | `/api/v1/health store=mysql` |
| 2 | 数据库迁移成功 | pass | 核心表存在；`plugin_migrations` 有 success 记录；未新增迁移 |
| 3 | 插件安装记录落库 | pass | `plugins` 中官方插件均存在 |
| 4 | 插件配置记录落库 | pass | `plugins.config_json` / `community_plugins` 仍为配置持久化口径 |
| 5 | 插件启用 / 禁用 / 归档状态落库 | pass | MySQL 查询确认 enabled；历史状态治理由现有记录 / 脚本覆盖 |
| 6 | community enable / disable 状态落库 | pass | Webhook full flow 子站 enable 成功；IA fixture 幂等确认 |
| 7 | SecretCenter create 成功 | pass | external_service token 写入 SecretCenter；Webhook Secret create |
| 8 | SecretCenter update 成功 | pass | Webhook Secret rotate / external_service token rotate 更新元数据 |
| 9 | SecretCenter disable 成功 | pass | S9 专用 `secret://s9/acceptance*` preview + disable |
| 10 | SecretCenter revoke 成功 | pass | S9 专用 `secret://s9/acceptance*` preview + strong confirm revoke |
| 11 | `secret_ref` 持久化正确 | pass | `secret_refs` 存在，effective-config 可查 |
| 12 | `token_ref` 持久化正确 | pass | `secret://external_service/official_webhook_notify/token` |
| 13 | external_service token 保存为 token_ref | pass | `plugin_external_services.token_ref` 存在 |
| 14 | token 明文不落库 | pass | 本轮测试 token 在相关表计数 0 |
| 15 | health check 成功 | pass | Webhook full flow `healthy` |
| 16 | delivery success 成功 | pass | `hook_executions id=184 status=success` |
| 17 | delivery fail 记录失败 | pass | 500 -> `retry_exhausted/http_status_failed` |
| 18 | delivery timeout 记录 network_timeout | pass | timeout -> `retry_exhausted/network_timeout` |
| 19 | manual retry 产生新记录 | pass | source=215 retry=217 success |
| 20 | `hook_executions` 落库 | pass | 记录数 217 |
| 21 | `admin_logs` 落库 | pass | 记录数 727+，Webhook / cleanup / allowlist / callback 均写入 |
| 22 | HTTP Allowlist 新增成功 | pass | API add 200 |
| 23 | HTTP Allowlist 删除成功 | pass | API delete 200 |
| 24 | endpoint safety 读取最终 allowlist | pass | effective list 含 env `http://host.docker.internal:18084`；删除后台 origin 后不再生效 |
| 25 | cleanup preview 不删除数据 | pass | preview 返回 confirm_token |
| 26 | cleanup 删除候选包记录 | pass | 未安装候选被删除 |
| 27 | cleanup 删除 storage 文件 | pass | `test_s9clean*` repo 删除后不存在 |
| 28 | installed 包不能 cleanup | pass | `plugin_a7b0cc04` skipped |
| 29 | enabled 包不能 cleanup | pass | 当前包 skipped；未做破坏性删除尝试 |
| 30 | template preview 不落盘 | pass | preview 仅返回 template |
| 31 | template generate 写 drafts | pass | `storage/plugins/drafts/s9tpl*` |
| 32 | template export zip 可用 | pass | `.devhub/s9-template-export.zip` 检查通过后已清理 |
| 33 | export zip 不含 `001_schema.sql` | pass | forbidden entries 0 |
| 34 | export zip 不含可执行代码 | pass | forbidden entries 0 |
| 35 | upgrade dry-run safe 可用 | pass | version bump / permission diff risk=`safe` |
| 36 | upgrade dry-run warning 可用 | pass | `official_links` cross-major `2.0.0` risk=`warning` 且返回 confirm_token |
| 37 | upgrade dry-run blocked 可用 | pass | same-version risk=`blocked` |
| 38 | warning confirm 可用 | not run | P2，S9 未执行真实 warning confirm |
| 39 | blocked 不能 confirm 绕过 | pass | same-version blocked + `confirmed=true` upgrade 返回 400 |
| 40 | Webhook Secret readiness 拦截可用 | not run | 当前进程已配置 key；缺 key readiness 由既有测试覆盖，S9 P2 |
| 41 | Webhook Secret create 可用 | pass | create 200；响应 record 不含 `secret_hash/secret_ciphertext` |
| 42 | Webhook Secret rotate 可用 | pass | rotate 200；响应 record 不含 `secret_hash/secret_ciphertext` |
| 43 | Callback Token create 可用 | pass | create 200 |
| 44 | Callback Token callback API 可用 | pass | config / audit 200 |
| 45 | 后台点击矩阵通过 | pass | `scripts/check-admin-plugin-ia.sh` 通过 |
| 46 | 1024 截图归档 | pass | `.devhub/screenshots/plugin-ia/1024-*.png` |
| 47 | 1366 截图归档 | pass | `.devhub/screenshots/plugin-ia/1366-*.png` |
| 48 | `secret_refs` 无 token 明文 | pass | S9 token count 0 |
| 49 | `plugin_external_services` 无 token 明文 | pass | S9 token count 0；cipher/hash 为空 |
| 50 | `admin_logs` 无 Bearer Header | pass | Binary `Bearer ` count 0 |
| 51 | `admin_logs` 无 Authorization | pass | count 0 |
| 52 | `admin_logs` 无 encrypted_value | pass | count 0 |
| 53 | `hook_executions` 无 Bearer Header | pass | count 0 |
| 54 | `hook_executions` 无 Authorization | pass | count 0 |
| 55 | `hook_executions` 无 encrypted_value | pass | count 0 |
| 56 | `system_settings` 不含敏感明文 | pass | sensitive count 0；allowlist origin 可明文 |
| 57 | `.devhub/checks` 不含敏感明文 | pass | `rg -a` 无命中 |
| 58 | `.devhub/screenshots` 不含敏感明文 | pass | `rg -a` 无命中 |
| 59 | 不开放 blocking Hook | pass | 边界未变 |
| 60 | 不执行第三方代码 | pass | 只执行 Core / API / receiver 测试脚本 |
| 61 | 不改变 Webhook Secret / Callback Token 安全模型 | pass | 明文一次性返回，非明文字段不泄露；Webhook Secret hash/ciphertext 仅留 DB 内部 |
| 62 | P0 / P1 / P2 分级已记录 | pass | P0=0，P1=0，P2=3 |

敏感扫描结果：

- MySQL 关键计数：`admin_logs_sensitive=0`、`hook_executions_sensitive=0`、`callback_plain_prefix_db=0`、`external_service_token_ref=1`、`system_settings_sensitive=0`；`admin_logs` / `hook_executions` 对 `secret_hash`、`token_hash`、`encrypted_value`、`Authorization`、`cbsk_` 和本轮测试 token 扫描均为 0。
- `secret_refs_enc_v_internal=8`：仅为 SecretCenter DB 内部密文存储，不在 API / 页面 / 日志 / 截图中暴露。
- `webhook_secret_hash_internal=12`：仅为 Webhook Secret DB 内部验签哈希；S9 修复后治理 API create / rotate / detail / list / disable / revoke 响应均不回显 `secret_hash` 或 `secret_ciphertext`。
- `admin_logs` 小写 `bearer ` 命中 2 条，内容为历史 allowlist usage 文案 `S21 restore bearer token_ref`，不是 `Bearer <token>` Header；二进制区分 `Bearer ` 扫描为 0。
- 文件扫描命令：`rg -a -n "s9-prod-acceptance-token-20260530|Authorization|encrypted_value|token_hash|secret_hash|cbsk_" .devhub/screenshots/plugin-ia .devhub/checks`，结果无命中；本轮生成 export zip / drafts / promoted test package 已清理，`storage/plugins/drafts` / `storage/plugins/packages` 无 `s9tpl*` / `s9zip*` 临时目录。

未执行检查及原因：

- Go `go test ./...` / `go build ./...`：未执行；S9 只做一处 Webhook Secret 响应脱敏修复，已执行 Docker `gofmt`、`./dev.sh start --mysql` 构建启动和专项 API smoke。
- `./scripts/check-frontend.sh --admin-only --quick` / `--frontend-only --quick`：未执行，原因是 S9 未修改前后台源码；`dev.sh start --mysql` 已完成一次前后台 build。
- `scripts/test-all.sh`：未执行，S9 已按 MySQL 生产模式总验收重点执行专项脚本、API smoke、DB / 文件扫描和文档同步。

安全边界确认：未执行第三方插件代码；未开放 Go plugin、JS 沙箱、远程 iframe、remote component、任意远程 JS、插件市场、远程在线安装或 blocking Hook；未回显 token / secret / Authorization / root key / `encrypted_value` / `secret_hash`；未改变 Webhook Secret / Callback Token 明文一次性返回安全模型；未改变 SecretCenter 加密逻辑；未破坏 `migrations/` 唯一入口；未破坏 dry-run 不执行 SQL；未改变 upload / promote / install / upgrade 语义；未破坏 `/topics/:id` 或 `/c/:slug` SEO；未删除 sites/posts 兼容 API。

## v1.9.0-S8 后台插件治理点击矩阵稳定化测试记录

本轮目标：稳定后台插件治理 5 个域和系统设置 / SecretCenter / HTTP Allowlist / 当前生效配置入口的浏览器点击矩阵，归档 1024 / 1366 截图，验证旧路由、Tab 状态、页面空态、技术详情折叠和敏感信息不回显。

覆盖结论：

- 插件列表 / 总览：入口、状态筛选、刷新、详情抽屉、配置、external_service 运行记录、SecretCenter `token_ref`、effective config 跳转和审计入口均可点击；无白屏、无 `undefined` / `null`、无英文 `No Data`。
- 安装与升级 / 插件包治理：流程工作台、上传包、本地包与预检、manifest 校验、package dry-run、本地仓库刷新、详情抽屉、install dry-run、版本与升级、清理入口和“升级差异”文案保持可用；1024 宽度操作列可用。
- Webhook 与回调：总览、投递记录、异常处理、高级治理、Webhook 密钥、回调 Token、回调请求、external_service 执行、manual retry 入口和 external_service 配置直达可用；endpoint / path / timeout / failure policy 可见，token / Authorization 不回显。
- 安全与发布者：可信发布者列表、公钥 / `key_id`、可信级别、影响分析、配置密钥入口可用；技术字段可查但不主导页面，操作列在 1024 宽度下可用。
- 运行与审计：操作历史、Hook 排障、搜索索引、审计日志入口可用；`admin_logs` / `hook_executions` 不白屏，metadata 仍在技术详情或详情抽屉中折叠展示。
- 系统设置：总览、当前生效配置、SecretCenter、外部服务策略 / HTTP Allowlist、配置审计可用；root key 只读，SecretCenter 详情 / copy ref / 来源配置 / 审计入口可用，HTTP Allowlist 新增弹窗对 `0.0.0.0` 给出中文拒绝。
- 旧路由：插件列表、插件包治理、Webhook、可信发布者、admin logs、uploads、本地仓库、模板/安装、SecretCenter/system settings、effective config 等旧入口均落到当前 IA Tab；刷新后 Tab 保留，上传旧路由 query/hash 保留。
- 截图归档：`.devhub/screenshots/plugin-ia/1024-plugin-overview.png`、`1366-plugin-overview.png`、`1024-plugin-packages.png`、`1366-plugin-packages.png`、`1024-webhook-governance.png`、`1366-webhook-governance.png`、`1024-trusted-publishers.png`、`1366-trusted-publishers.png`、`1024-admin-logs.png`、`1366-admin-logs.png`、`1024-system-settings.png`、`1366-system-settings.png`、`1024-secret-center.png`、`1366-secret-center.png`、`1024-effective-config.png`、`1366-effective-config.png`。
- 安全边界：本轮不执行第三方插件代码，不开放 Go plugin、JS sandbox、远程 iframe、remote component、插件市场、远程在线安装或 blocking Hook；不改变 Webhook Secret / Callback Token 安全模型、SecretCenter 加密逻辑、`migrations/` 唯一入口或 dry-run no-SQL 规则。

已执行命令与结果：

- `bash -n scripts/check-admin-plugin-ia.sh`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，最终日志目录 `.devhub/checks/20260530-103504/`，仅有既有 Vite chunk size warning。
- `bash scripts/check-admin-plugin-ia.sh`：通过，报告 `.devhub/screenshots/plugin-ia/report.json`，截图目录 `.devhub/screenshots/plugin-ia`。
- `git diff --check`：通过。
- 敏感扫描：`rg -a -n "s8-browser-matrix-token-value|s6-browser-matrix-token-value|Bearer|Authorization|encrypted_value|enc:v|token_hash|DEVHUB_PLUGIN_CONFIG_KEYS|root key|secret plaintext|password plaintext|token plaintext" .devhub/screenshots/plugin-ia .devhub/checks` 无命中。浏览器页内断言也未发现 token / secret / Authorization / `encrypted_value` / `enc:v` / token_hash 回显。

待最终收口 / 本轮未执行：

- Go `gofmt` / `go test ./...` / `go build ./...`：未执行，原因是本轮未修改 Go。
- `./scripts/check-frontend.sh --frontend-only --quick`：未执行，原因是本轮未修改前台应用。
- MySQLStore smoke / integration：未执行，原因是本轮未改 Store、migration、DB schema 或 MySQL-only 行为。
- `scripts/test-all.sh`：未执行，S8 已按浏览器点击矩阵范围执行 admin quick、IA 脚本、bash 语法、diff 和敏感扫描。

## v1.9.0-S7 upgrade dry-run UI 复核补漏测试记录

本轮目标：只复核并补强后台 upgrade dry-run 结果 UI 对既有结构化响应的消费，不改后端结构、不改 dry-run / upgrade / approval 语义。

覆盖结论：

- dry-run 入口：`/admin-next/plugins/install` 的 manifest `upgrade-dry-run` 与 `upgrade` 向导继续使用 `POST /api/v1/admin/plugins/:code/upgrade/dry-run`。
- `blocked_items` / `warnings`：已在结果顶部前置；blocked 文案说明不能继续且不能被 confirm 绕过，warning 文案说明必须显式确认。
- warning confirm：既有 `upgradeConfirmChecked` 规则保留，warning 未勾选时不能进入执行。
- blocked non-bypass：既有 `upgradePreviewBlocked` 规则保留，blocked 时下一步按钮禁用；UI 也新增顶部阻断提示。
- `diff_sections`：按 section 以 tabs 展示风险、类型、路径、说明、before/after，原始 JSON 保留在技术详情折叠区。
- `impact_summary`：继续显示站点、内容类型、权限、配置、Hook 计数，以及 migration / external_service / frontend_mount / menu-route 影响标记，不输出 `undefined` / `null`。
- `rollback_boundary`：继续展示“不提供完整自动回滚”、migration down / manifest / config / assets / external_service 回滚边界和人工处理项，不夸大回滚能力。
- `migration_plan` / `config_plan` / `permission_plan` / `content_type_plan` / `hook_plan` / `frontend_mount_plan`：新增分区计划卡片；migration 区明确 dry-run 不执行 SQL、不执行 migration down。
- `dependency_diff`：新增结构化计数和表格，覆盖新增 / 删除 / 版本约束变化 / required 变化。
- `failure_stage` / `failure_reason`：失败区域展示阶段、原因、request id（如果响应提供）、下一步建议，并提示查看 `admin_logs` / `hook_executions`。
- secret diff masking：UI 只展示后端返回的脱敏值；敏感 path 仍由后端返回 `[REDACTED]`，本轮未新增任何明文 token / secret / Authorization / root key / `encrypted_value` 展示入口。
- frontend_mount diff：新增表格展示 `mount_point`、`component_key`、`render_mode`、allowlist 结果和阻断说明；危险 remote / inline / 非官方 render mode 仍由既有后端 allowlist 校验阻断。

已执行命令与结果：

- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260530-100834/`，仅有既有 Vite chunk size warning。
- `git diff --check`：通过。
- 敏感关键词扫描：`rg -n "Bearer|Authorization|encrypted_value|enc:v|token_hash|DEVHUB_PLUGIN_CONFIG_KEYS|root key|secret plaintext|password plaintext|token plaintext" web/admin-app/src/views/plugins/PluginInstallUpgrade.vue .devhub/checks/20260530-100834` 无命中；本轮 UI 文件和本轮 admin quick 日志未出现敏感明文或密文字段。
- 结构化字段扫描：`diff_sections`、`impact_summary`、`rollback_boundary`、`frontend_mount_plan`、`dependency_diff`、`failure_stage`、`failure_reason`、`admin_logs`、`hook_executions`、`blocked_items`、`warnings` 已命中本轮 UI 与文档记录。

待最终收口 / 本轮未执行：

- Go `gofmt` / `go test ./...` / `go build ./...`：未执行，原因是本轮未修改 Go。
- `./scripts/check-frontend.sh --frontend-only --quick`：未执行，原因是本轮未修改前台应用。
- MySQLStore smoke / integration：未执行，原因是本轮未改真实升级数据流、Store 或 Go dry-run / upgrade 逻辑；如后续需要生产模式验收，应覆盖 `admin_logs`、upgrade dry-run safe / warning / blocked 和敏感扫描。
- 浏览器点击脚本 / 截图归档：未执行，原因是本轮未启动浏览器；可用 `bash scripts/check-admin-plugin-ia.sh` 做人工矩阵补测。
- dry-run response samples / upgrade plan tables / `admin_logs` / `hook_executions` DB 扫描：未执行，原因是本轮未启动服务或创建真实升级 dry-run 数据，只修改 UI 消费层；相关字段已在源码和文档中扫描确认。

安全边界：本轮不执行第三方插件代码，不开放 Go plugin、JS sandbox、远程 iframe、remote component、任意远程 JS、插件市场、远程在线安装、blocking Hook、package scripts、自动安装或自动启用；不改变 upload / precheck / promote / install / upgrade / approval 语义；`migrations/` 仍是唯一迁移入口，dry-run 不执行 SQL；blocked 不能通过 confirm 继续。

## v1.9.0-S6 插件模板生成稳定化测试记录

本轮目标：验证后台 preview / generate / export zip 与 CLI `plugin:new` 共用 `PluginTemplateGenerator`，输出只包含声明型模板文件、`docs/*.md` 和 `migrations/001_init.sql`，不生成可执行代码、package scripts、根目录 `001_schema.sql` 或 blocking Hook 模板。

新增 / 强化覆盖点：

- `internal/plugins/scaffold`：覆盖 content / external_service / admin_tool / frontend_mount 四类模板；检查 `docs/usage.md`、`docs/content-types.md`、`docs/permissions.md`、按需 `docs/hooks.md` / `docs/migrations.md`、`docs/registry-example.md` 和 `migrations/001_init.sql`；拒绝 `registry.example.go`、根目录 `001_schema.sql`、`package.json`、`index.js`、`main.go` 和 blocking Hook。
- `internal/service`：覆盖 preview 不落盘、generate 写入 `storage/plugins/drafts/{code}`、export zip 不包含 forbidden markers，并打开 zip entry 检查 required entries 与 zip-slip 路径。
- CLI：实跑 content / external_service / admin_tool / frontend_mount 四类生成；`/tmp` 输出用于文件树和敏感扫描；`.devhub/plugins/*` 输出用于 package check。
- 安全扫描：生成目录扫描 `.go`、JS/MJS/CJS、WASM、`package.json`、根目录 `001_schema.sql`、`scripts/`、`remote_entry`、`script_url`、`inline_html`、`iframe_url`、`remote_component`、`eval`、`Authorization`、`Bearer`、`encrypted_value`、`enc:v`、`token_hash`、`DEVHUB_PLUGIN_CONFIG_KEYS`、`root key` 和 `secret plaintext`。

已执行命令与结果：

- `/usr/local/go/bin/gofmt -w cmd/devhub/main.go internal/plugins/scaffold/scaffold.go internal/plugins/scaffold/scaffold_test.go internal/service/plugin_package_template_service.go internal/service/plugin_package_template_test.go`：通过。
- `git diff --check`：通过。
- `/usr/local/go/bin/go test ./internal/plugins/scaffold -count=1`：通过。
- `/usr/local/go/bin/go test ./internal/service -run 'TestPluginPackageTemplate' -count=1`：通过。
- `/usr/local/go/bin/go test ./internal/transport/httpapi -run 'TestAdminPluginPackageTemplateHTTPPreviewGenerateExport' -count=1`：通过。
- CLI content / external_service / admin_tool / frontend_mount 生成：通过，输出分别包含声明文件与 `docs/*.md`；content 额外包含 `migrations/001_init.sql`，external_service 只声明 non-blocking external_service Hook，admin_tool / frontend_mount 默认不生成 Hook / migration。
- `./scripts/plugin-package-check.sh .devhub/plugins/s6_cli_content`、`s6_cli_external`、`s6_cli_admin`、`s6_cli_mount`：均为 warning，无 blocker；warning 仅为缺 `checksums.json`、`publisher.json`、`signature.json`；manifest_valid=true；external_service 校验 passed；frontend_mount allowlist passed；content migration plan 为 `migrations/001_init.sql will_execute=false`。
- `/usr/local/go/bin/go build ./...`：通过。
- `/usr/local/go/bin/go test ./...`：失败，失败点为既有环境权限问题：`internal/service TestPluginUpgradeFromCompatRequiresConfirmForWarning` 无法在 root-owned `storage/plugins/staging/test` 下创建 `up_demo_confirm`，报 `permission denied`；模板相关定向测试均已通过。

未执行 / 说明：

- 未执行 `./scripts/check-frontend.sh --admin-only --quick`、`--frontend-only --quick` 或 Playwright：本轮未修改前端代码，后台模板 UI 字段行为沿用既有实现；本轮以后端生成器 / CLI / docs 收口为主。
- 未执行 MySQLStore：本轮不改 storage schema、repo conflict 查询、admin_logs 结构或 install/promote 语义；MySQLStore 路径与 S4 官方插件回归不同，不涉及 SecretCenter / external_service token 持久化。
- package check 曾对 `/tmp/devhub-s6-cli-*` 输出返回 blocked，原因是 package checker 正确拒绝允许目录外路径；随后改在 `.devhub/plugins/*` 重新生成并完成 package check。

## v1.9.0-S4 官方插件稳定化回归测试记录

本轮目标：对 `official_links`、`official_announcement`、`official_webhook_notify` 做稳定化回归，只确认既有官方插件模型稳定可用，不扩大第三方运行时、Hook、SecretCenter、Webhook Secret、Callback Token、SEO 或迁移边界。

新增 / 强化覆盖点：

- `official_webhook_notify`：官方样板包版本升至 `1.0.1`，manifest 补齐 `external_service.subscribed_hooks=["AfterCreateContent"]`，并同步 checksums；external-service-webhook 模板补齐同一订阅声明。
- Webhook 回归脚本：`scripts/run-feishu-webhook-flow.sh` 保持默认 feishu_link 行为，新增 `DEVHUB_PUBLISH_CONTENT_TYPE`、`DEVHUB_PUBLISH_PLUGIN_CODE`、`DEVHUB_CATEGORY_PLUGIN_CODE`、`DEVHUB_CATEGORY_ID` 覆盖项，用于以 core `article` 触发 `official_webhook_notify`。
- `official_links`：覆盖 package check、声明型 upload / precheck / promote / install dry-run / install / enable / community enable、配置读写、菜单、权限矩阵、创建 friend_link、community disabled / global disabled / archived 阻断和历史内容读取。
- `official_announcement`：覆盖内置 package check、Host context gating、admin preview / audit 脱敏、首页 iframe 挂载、同源 iframe 路由 `/plugins/official-announcement/iframe` 和 `sandbox=allow-scripts`。
- `official_webhook_notify`：覆盖 MySQLStore + fresh receiver 下 health check、success、500 -> `retry_exhausted`、timeout -> `network_timeout/retry_exhausted`、manual retry -> success、`hook_executions`、`admin_logs`、`token_ref` 与 SecretCenter 元数据。
- 浏览器 IA：覆盖 5 个插件治理域、三官方插件详情、`official_links` 管理页、Webhook 治理、运行与审计、旧路由和 1024 / 1366 截图；截图目录 `.devhub/screenshots/plugin-ia`。
- 敏感扫描：API / DB / 浏览器 / receiver 输出不允许出现测试 token 明文、Authorization Header、`Bearer ` 明文、`encrypted_value`、token_hash 或 `DEVHUB_PLUGIN_CONFIG_KEYS` 持久化。

已执行命令：

```bash
bash -n scripts/run-feishu-webhook-flow.sh
./scripts/plugin-package-check.sh examples/plugins/official_links
./scripts/plugin-package-check.sh examples/plugins/official_webhook_notify
./scripts/plugin-package-check.sh --builtin official_announcement
./scripts/plugin-package-check.sh examples/plugins/templates/external-service-webhook
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm gofmt -w ...
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./internal/service -run 'TestDryRunPluginPackage_Official|TestConfigSchemaPatternsAreJSCompatible|TestExternalService|TestSecretCenter|TestSystemEffectiveConfig' -count=1
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./internal/transport/httpapi -run 'TestOfficialAnnouncement|TestAdminPostCreateSupportsDeclarativePluginContentType' -count=1
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./...
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go build -buildvcs=false ./...
./scripts/check-frontend.sh --admin-only --quick
./scripts/check-frontend.sh --frontend-only --quick
docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-declarative-install-use.spec.js
docker compose run --rm frontend-e2e npx playwright test tests/e2e/official-announcement.spec.js
bash scripts/check-admin-plugin-ia.sh
DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18083 ./dev.sh restart --mysql --no-build
DEVHUB_PLUGIN_CODE=official_webhook_notify ... DEVHUB_WEBHOOK_FLOW=full DEVHUB_WEBHOOK_AUTH_TYPE=bearer ./scripts/run-feishu-webhook-flow.sh
docker exec sns-mysql-1 mysql ... # MySQLStore 状态、hook_executions/admin_logs/token_ref 与敏感扫描
git diff --check
```

当前结果：

- Package check：`official_links` / `official_webhook_notify` / external-service-webhook 模板为 warning（缺 publisher / signature，无 blocker）；内置 `official_announcement` passed。
- Go / build：Docker `gofmt`、定向 Go 测试、Docker `go test ./...`、Docker `go build -buildvcs=false ./...` 均通过。
- Frontend quick：后台 quick 通过，日志目录 `.devhub/checks/20260530-000340/`，仅有既有 Vite chunk size warning；前台 quick 通过，日志目录 `.devhub/checks/20260530-000340/`。
- Playwright：`plugin-declarative-install-use.spec.js` 首次因 root-owned fixture dist 目录 EACCES 未进入业务断言，修正目录属主后复跑通过，1 passed；`official-announcement.spec.js` 通过，1 passed。
- Browser IA：`bash scripts/check-admin-plugin-ia.sh` 通过，报告 `.devhub/screenshots/plugin-ia/report.json`，`sensitiveScan=[]`，截图目录 `.devhub/screenshots/plugin-ia`。
- MySQLStore receiver：`official_webhook_notify@1.0.1` full flow 通过，success / 500 / timeout / manual retry 均有执行记录，receiver 只记录 `authorization_present=True`，不输出 Authorization 明文。
- MySQL 敏感扫描：本轮测试 token 明文在 `secret_refs`、`plugin_external_services`、`admin_logs`、`hook_executions`、`plugins`、`system_settings` 中均为 0；`plugin_external_services.token_ciphertext/token_hash` 为空；`Authorization` Header / `Bearer ` 明文计数为 0。`secret_refs.encrypted_value` 中的 `enc:v*` 仅作为 DB 内部密文存储，API / 浏览器 / 日志不暴露。
- `git diff --check`：通过。

未执行项：

- 未执行 `scripts/test-all.sh`：S4 已按官方插件稳定化范围执行 Go 全量、Go build、前后台 quick、两个相关 E2E、浏览器 IA、MySQLStore receiver 和敏感扫描。
- 未执行无关 E2E 全量矩阵：本轮未修改非插件治理、非官方插件、非 external_service 的用户流程。

安全边界：本轮不新增 API，不改数据库 schema，不持久化 `DEVHUB_PLUGIN_CONFIG_KEYS`，不展示 plaintext token / secret / Authorization / root key / `encrypted_value` / token_hash，不改变 SecretCenter 加密逻辑、Webhook Secret / Callback Token 模型、external_service non-blocking delivery、upload / precheck / promote / install / upgrade 语义，不开放第三方代码执行、Go plugin、JS/WASM/Lua sandbox、远程 iframe、remote component、任意远程 JS、插件市场、远程在线安装、blocking Hook、package scripts、自动安装或自动启用。SEO 路由 `/topics/:id` 与 `/c/:slug` 未改动。

## v1.9.0-S3 当前生效配置与排障视图稳定化测试记录

本轮目标：验证系统设置 -> 当前生效配置可稳定展示基础运行信息、external_service 当前生效配置、SecretCenter 状态、HTTP Allowlist 来源、Webhook / Callback 安全摘要、脱敏 `token_ref` / `secret_ref` 元数据、quick links 和可复制脱敏诊断文本。

新增 / 强化覆盖点：

- effective-config API：断言 `root_key_status` 不携带 root key env 示例，`secret_center_status`、`webhook_callback_security`、`http_allowlist_source`、external_service 行级 `auth_type`、健康字段、endpoint origin、allowlist match/source/message 和 token source metadata 均存在且脱敏。
- diagnostic_text：断言复制用 JSON 不包含 token 明文、secret 明文、Authorization、`DEVHUB_PLUGIN_CONFIG_KEYS` env 示例、`encrypted_value` 或 token_hash；带敏感 query 的 endpoint 在 diagnostic_text 中被脱敏。
- HTTP Allowlist：覆盖后台 allowlist source 为 `admin_setting`，external_service endpoint origin 命中 allowlist；未命中 / Docker loopback / health failure 仍通过 next_steps 给出下一步。
- SecretCenter / root key：缺 root key 时只返回启动注入建议，不回显或复制 root key；Secret disabled/revoked、missing token_ref 等状态继续产生 next_steps。
- Admin UI：新增 `system-effective-config.spec.js`，覆盖当前生效配置页关键区块渲染、external_service / HTTP Allowlist / SecretCenter / Webhook Callback 区块可见、复制诊断文本脱敏、页面不出现 `undefined` 或 `encrypted_value`。
- 布局：System.vue 对长 `endpoint_url`、`endpoint_origin`、`token_ref`、技术 JSON 和按钮区增加换行 / 横向滚动 / 手动复制兜底；1024 / 1366 宽度检查由后台 quick 与最小 System E2E 覆盖，若需截图矩阵可再手动跑浏览器截图。

已执行命令：

```bash
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm gofmt -w internal/domain/secret_ref.go internal/domain/system_effective_config.go internal/service/mysql_integration_test.go internal/service/secret_center.go internal/service/secret_center_test.go internal/service/system_effective_config.go
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./internal/service -run 'TestSystemEffectiveConfig|TestSecretCenter' -count=1
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./...
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go build -buildvcs=false ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
./scripts/check-frontend.sh --frontend-only --quick
docker compose up -d --force-recreate devhub
docker compose run --rm admin-e2e npx playwright test tests/e2e/system-effective-config.spec.js
# 另以 admin-e2e Playwright 对 1024x768 / 1366x768 执行轻量布局检查。
```

当前结果：

- Docker `gofmt`：通过。
- SecretCenter / effective-config 定向 Go 测试：通过。
- Docker `go test ./...`：通过。
- Docker `go build -buildvcs=false ./...`：通过。
- `git diff --check`：通过。
- 后台 quick：通过，日志目录 `.devhub/checks/20260529-231354/`，仅有既有 Vite chunk size warning。
- 前台 quick：通过，日志目录 `.devhub/checks/20260529-231554/`。
- 最小 System E2E：首次在旧 `sns-devhub-1` 容器未重启时失败，失败点为 stale backend notes 仍含旧 `encrypted_value` 文案；`docker compose up -d --force-recreate devhub` 后复跑 `system-effective-config.spec.js` 通过，1 passed。
- 1024 / 1366 布局：通过。admin-e2e Playwright 在 1024x768 与 1366x768 下打开系统设置 -> 当前生效配置，`system-effective-runtime`、`system-effective-secretcenter`、`system-effective-webhook-callback` 可见，页面无 `undefined` / `encrypted_value`，`bodyOverflowPx=0`。

未执行项：

- 未执行完整 admin-e2e 全量矩阵：S3 只改系统设置 / effective-config / SecretCenter 排障视图，已新增并执行最小相关 System E2E。
- 未执行 `scripts/test-all.sh`：本轮按影响面执行 Go 全量、Go build、前后台 quick、最小相关 E2E 和 diff 检查。
- 未执行 MySQL DB 级敏感扫描：S3 不改存储 schema 或写入语义；后续发布前总验收可复用 S1/S2 的 MySQL 敏感扫描脚本补跑。

安全边界：本轮不展示 plaintext secret/token/Authorization、Webhook Secret 明文、Callback Token 明文、`encrypted_value`、token_hash 或 root key；root key 仍只启动注入，后台不保存、不生成、不修改；不改变 SecretCenter 加密逻辑、Webhook Secret / Callback Token 安全模型、external_service non-blocking delivery、upload / promote / install 语义，也不开放 blocking Hook、第三方代码执行、插件市场或远程在线安装。

## v1.9.0-S2 SecretCenter 操作闭环收口测试记录

本轮目标：验证 SecretCenter 详情、使用关系、来源跳转、禁用 / 吊销影响预览、轮换回源、Secret 审计入口和 effective-config troubleshooting 补强不会泄露敏感信息，也不改变 SecretCenter / external_service / Webhook Secret / Callback Token 安全模型。

已新增 / 强化的覆盖点：

- Secret detail API：断言 `record/source/usages` 含 `usage_type/source_type/source_id/source_code`，且不含明文 token、Webhook Secret、Callback Token、`encrypted_value` 字段或 `token_hash`。
- usage computation：覆盖 external_service token、Webhook Secret、Callback Token 和 plugin config sensitive field 的来源字段。
- impact preview：禁用 preview 仍返回 external_service 影响、可能失败的 health check / delivery 统计。
- audit recording：吊销 Webhook Secret 经 SecretCenter 入口后写 `secret_center.secret.revoked`，metadata 包含 `source_type/source_code` 等脱敏筛选线索。
- effective-config：断言 external_service 行含 `config_source`，disabled secret 会产生 `next_steps`；`diagnostic_text` 不包含明文或 `encrypted_value` 字段。
- root-key-missing diagnostics：缺 root key 时 effective-config 返回脱敏 `next_steps`，继续说明 root key 只能启动注入，不保存到后台 / DB。

已执行命令：

```bash
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm gofmt -w internal/domain/secret_ref.go internal/domain/system_effective_config.go internal/service/secret_center.go internal/service/system_effective_config.go internal/service/secret_center_test.go
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./internal/service -run 'TestSecretCenter|TestSystemEffectiveConfig' -count=1
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./...
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go build -buildvcs=false ./...
./scripts/check-frontend.sh --admin-only --quick
docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-config-encryption.spec.js
git diff --check
```

当前结果：

- Docker `gofmt`：通过；宿主机无 `gofmt`。
- SecretCenter / effective-config 定向 Go 测试：通过。
- Docker `go test ./...`：通过。
- Docker `go build -buildvcs=false ./...`：通过。
- 后台 quick：通过，最终日志目录 `.devhub/checks/20260529-220558/`，仅有既有 Vite chunk size warning。
- 最小敏感配置 E2E：`plugin-config-encryption.spec.js` 通过，1 passed。
- `git diff --check`：通过。

未执行项：

- 未执行前台 quick：本轮未修改前台应用。
- 未执行完整 admin-e2e 全量矩阵：当前仓库没有专门的 System / SecretCenter E2E 文件；已执行最接近的敏感配置专项，不跑无关全量。
- 未执行 `scripts/test-all.sh`：本轮已按影响面执行 Go 全量、Go build、后台 quick 和最小相关 E2E。

安全边界：本轮不展示 plaintext secret/token/Authorization、Webhook Secret 明文、Callback Token 明文、`encrypted_value` 或可逆密文；不在 SecretCenter 做无上下文轮换；不修改 root key 保存方式；不改变 external_service non-blocking delivery、Webhook Secret / Callback Token 模型、upload / promote / install 语义，也不开放 blocking Hook、第三方代码执行、插件市场或远程在线安装。

## v1.9.0-S1 feishu_link receiver 全链路回归

本轮目标：补齐 v1.8.4 RC 遗留 P2，使用 fresh standalone receiver 验证 `feishu_link` / `plugin_a7b0cc04` external_service Webhook 的 success / fail / timeout / retry 全链路，并覆盖 `hook_executions`、`admin_logs`、`token_ref` 与敏感信息扫描。

脚本增强：

- `scripts/run-feishu-webhook-flow.sh` 新增 `DEVHUB_WEBHOOK_FLOW=full`。
- `full` 会依次执行：
  - receiver `ok`：创建内容后 `AfterCreateContent` 投递成功，`status=success`、`response_status=200`。
  - receiver `500`：投递重试后进入 `status=retry_exhausted`、`error_code=http_status_failed`。
  - receiver `timeout`：投递重试后进入 `status=retry_exhausted`、`error_code=network_timeout`。
  - manual retry：先制造 `retry_exhausted`，再切回 receiver `ok`，调用 `POST /api/v1/admin/plugins/:code/hooks/executions/:execution_id/retry`，新执行记录成功。
- 脚本会检查 external_service 保存响应、health check 响应、`hook_executions`、manual retry 响应和 `admin_logs` 响应，不允许出现测试 token 明文、Authorization 字段、`encrypted_value` 或 `token_hash`。

本轮执行命令：

```bash
bash -n scripts/run-feishu-webhook-flow.sh
DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18082 ./dev.sh restart --mysql --no-build
DEVHUB_WEBHOOK_PORT=18082 DEVHUB_WEBHOOK_ENDPOINT=http://host.docker.internal:18082 DEVHUB_WEBHOOK_FLOW=full DEVHUB_WEBHOOK_AUTH_TYPE=bearer DEVHUB_WEBHOOK_TOKEN=<测试 token> ./scripts/run-feishu-webhook-flow.sh
docker exec sns-mysql-1 mysql ... # 扫描 secret_refs / plugin_external_services / admin_logs / hook_executions
```

结果：

- `bash -n scripts/run-feishu-webhook-flow.sh`：通过。
- `./dev.sh restart --mysql --no-build`：通过；MySQLStore 后端以 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18082` 启动。
- Full receiver regression：通过；成功、500 失败、timeout、manual retry 均完成。receiver 共收到 11 次投递，manual retry 成功；receiver 侧仅记录 `authorization_present=True`，不输出 Authorization 明文。
- API 响应敏感扫描：通过；external_service 保存 / health check / `hook_executions` / manual retry / `admin_logs` 响应未出现测试 token 明文、Authorization 字段、`encrypted_value` 或 `token_hash`。
- MySQL 敏感扫描：通过；`secret_refs_plain_token=0`、`plugin_external_services_plain_token=0`、`admin_logs_plain_token=0`、`hook_executions_plain_token=0`。
- MySQL external_service 存储状态：`plugin_a7b0cc04` 为 `auth_type=bearer`、`token_ref=secret://external_service/plugin_a7b0cc04/token`、`token_ciphertext=empty`、`token_hash=empty`。

跳过项：

- Go test / Go build 未执行：本轮未修改 Go 代码。
- admin/frontend quick、Playwright、浏览器点击矩阵未执行：本轮未修改前后台 UI。
- package governance 链路未执行：本轮只覆盖 Webhook receiver 回归。
- `scripts/test-all.sh` 未执行：本轮已按 S1 范围执行 MySQL + Webhook 专项。

边界确认：不新增 API，不改变数据库 schema，不改变 SecretCenter / external_service / Webhook Secret / Callback Token 安全模型，不开放 blocking Hook，不执行第三方代码，不开放远程 iframe、插件市场或远程在线安装。

## v1.8.4-S23 发布收口测试归档

本轮目标：归档 `v1.8.4` RC acceptance，冻结 P0/P1/P2、执行检查、跳过检查和安全边界。结论：`v1.8.4` RC acceptance pass，可冻结；P0 无，P1 无，P2 为 standalone fresh `feishu receiver` 全链路 `success / fail / timeout / retry` 未重跑，非阻塞。

S23 归档结果：

- 总体 RC acceptance：pass，依据 S21 生产候选总验收矩阵。
- Admin IA browser/script：pass，S21 已执行 `bash scripts/check-admin-plugin-ia.sh`，截图目录 `.devhub/screenshots/plugin-ia`；脚本对当前 IA 标题、按钮、选择器和运行域可见 Tab 的断言已对齐。
- MySQLStore 定向集成：pass，S21 已执行 `DEVHUB_MYSQL_TESTS=1 ... go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v`，5 个子用例均通过。
- 敏感信息扫描：pass，S21 抽查 API 响应、quick/E2E 日志、MySQL `admin_logs` / `hook_executions`，未发现 Bearer / Authorization / `enc:v*` / `encrypted_value` / 测试 token / secret plaintext / token_hash；`admin_logs` 仅有 root key 环境变量示例文案，不是实际 key 值。
- Runtime hygiene：pass，HTTP allowlist 已恢复为 `[]`；`plugin_a7b0cc04` 已恢复为 `http://host.docker.internal:18081` + bearer `token_ref`；不暴露 token / secret / Authorization。
- P2：未新起独立 fresh `feishu receiver` 重新跑 `success / fail / timeout / retry`；该项是补充回归，建议放入 `v1.8.5` 或 `v1.8.4` post-release smoke，不阻塞 release candidate。

S23 本轮执行命令：

```bash
git status --short
rg -n "v1.8.4|P0|P1|P2|SecretCenter|token_ref|external_service|HTTP Allowlist|blocking Hook|远程在线安装|插件市场|DEVHUB_PLUGIN_CONFIG_KEYS|encrypted_value" docs/releases/v1.8.4.md docs/PROJECT_PROGRESS.md docs/TESTING.md CHANGELOG.md README.md
git diff --check
```

S23 本轮结果：

- `git status --short`：显示 S23 文档改动以及既有 S21 `internal/service/mysql_integration_test.go`、`scripts/check-admin-plugin-ia.sh` 改动；未回滚。
- 发布一致性 grep：命中 S23 / S21 发布文档、项目进度、测试归档、CHANGELOG 和 README 中的 `v1.8.4`、P0/P1/P2、SecretCenter、`token_ref`、external_service、HTTP Allowlist 与安全边界关键词。
- `git diff --check`：通过。

S23 跳过项：

- Go test / Go build / frontend quick / Playwright / MySQL receiver smoke 未重跑：本轮只改发布归档文档，不改代码或运行语义；S21 已完成 RC 验收。
- `scripts/test-all.sh` 未执行：S23 不需要额外全量入口。
- `bash -n scripts/check-admin-plugin-ia.sh` 未执行：本轮未修改脚本。

S23 安全边界：不执行第三方插件代码；不开放 Go plugin、JS sandbox、远程 iframe、remote component、任意远程 JS、插件市场、远程在线安装、blocking Hook、自动安装、自动启用或 package scripts 执行；`migrations/` 仍是唯一迁移入口，dry-run 不执行 SQL；Webhook Secret / Callback Token 模型不变；`DEVHUB_PLUGIN_CONFIG_KEYS` 不存入 admin/backend DB，root key 不写入 DB；token / secret / Authorization / `encrypted_value` 不回显；`admin_logs` / `hook_executions` 未发现敏感泄露。

## v1.8.4-S21 生产候选总验收

本轮目标：确认 `v1.8.4` 是否具备发布候选条件。结论：未发现 P0 / P1；仅有 P2 级补充记录，建议进入 S23 发布收口。

验收矩阵：

| 模块 | 场景 | 结果 | 证据 | 问题等级 | 备注 |
| -- | -- | -- | -- | ---- | -- |
| 插件包链路 | upload / precheck / promote / install dry-run / install / enable / duplicate install | pass | `go test ./...`；`plugin-governance.spec.js` 13 passed | none | upload/promote/install/enable 语义保持分离 |
| 插件包 cleanup | preview、confirm、受控路径、installed skipped、审计 | pass | `go test ./...` 中 package repository cleanup 覆盖 | none | 不影响已安装插件 |
| 后台模板生成 | preview/generate/export 与安全模板内容 | pass | `go test ./...` 中 package template service/http 覆盖 | none | 不生成根目录 `001_schema.sql` |
| external_service Webhook | health check、non-blocking delivery、retry/skipped、执行记录 | pass | `go test ./...`；`/admin/plugins/hooks/executions` API 抽查 | none | 未新起 receiver，沿用 S15 实跑和当前历史记录 |
| SecretCenter | 列表、详情、来源、使用关系、禁用/吊销预览、强确认 | pass | `go test ./...`；`/secret-center/secrets` API 抽查 | none | 无明文或 `encrypted_value` |
| HTTP Allowlist | 默认 localhost、非 localhost HTTP 拒绝、后台来源增删校验 | pass | `go test ./...`；`/system/external-service/http-allowlist` API 抽查 | none | 不允许 wildcard / 0.0.0.0 / CIDR |
| 当前生效配置 | 版本、store、SecretCenter、external_service、token_ref、allowlist、诊断文本 | pass | `/system/effective-config` API 抽查 | none | 诊断文本脱敏 |
| upgrade dry-run UI | safe/warning/blocked、confirm、blocked 不绕过、技术详情 | pass | `plugin-governance.spec.js` upgrade smoke；Go upgrade dry-run tests | none | 后端语义未变 |
| 后台 UI 回归 | 核心插件治理页、System、技术详情默认折叠、1024 / 1366 截图矩阵 | pass | `./scripts/check-frontend.sh --admin-only --quick`；`plugin-governance.spec.js`；`bash scripts/check-admin-plugin-ia.sh` | none | 截图目录 `.devhub/screenshots/plugin-ia` |
| 敏感信息扫描 | API / hook_executions / admin_logs / 测试日志 | pass | `/tmp/devhub-s21-*.json`、`.devhub/checks/20260529-190252`、MySQL `admin_logs` / `hook_executions` SQL 扫描 | none | Bearer/Authorization/`enc:v*`/`encrypted_value`/测试 token/secret plaintext/token_hash 均为 0；`admin_logs` 中仅有 root key 环境变量示例文案 |
| Memory / MySQL | 核心路径无明显分歧 | pass | Compose memory E2E 13 passed；`DEVHUB_MYSQL_TESTS=1 ... TestMySQLStorePluginPlatformConsistency` 通过 | none | 未重跑完整 MySQL receiver smoke |
| 文档 | 记录结果、矩阵、分级、命令、跳过项、安全边界 | pass | 本文、项目进度、release notes、CHANGELOG | none | 无需更新 API 主体 |

已执行命令：

```bash
git status --short
git diff --check
bash -n scripts/check-frontend.sh
./scripts/check-frontend.sh --admin-only --quick
docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js
go version
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go test ./...
docker run --rm -v "$PWD:/app" -w /app golang:1.22-bookworm go build -buildvcs=false ./...
bash scripts/check-admin-plugin-ia.sh
docker exec sns-mysql-1 mysql -uroot -pDevhub_root_123456 -e "DROP DATABASE IF EXISTS devhub_test; CREATE DATABASE devhub_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL PRIVILEGES ON devhub_test.* TO 'devhub'@'%'; FLUSH PRIVILEGES;"
docker run --rm --network host -v "$PWD:/app" -v /home/liuwei/go/pkg/mod:/go/pkg/mod -v /home/liuwei/.cache/go-build:/root/.cache/go-build -w /app -e GOPROXY=off -e GOSUMDB=off -e DEVHUB_MYSQL_TESTS=1 -e DB_HOST=127.0.0.1 -e DB_PORT=3307 -e DB_USER=devhub -e DB_PASSWORD=Devhub_123456 -e DB_NAME=devhub_test golang:1.22-bookworm go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v
```

结果：

- `git diff --check`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-190252/`，仅有既有 Vite chunk size warning。
- `docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js`：通过，13 passed。
- 宿主机 `go version`：失败，宿主机无 `go` 命令；已改用 Docker Go。
- Docker `go test ./...`：通过。
- Docker `go build -buildvcs=false ./...`：通过。
- `bash scripts/check-admin-plugin-ia.sh`：通过，截图目录 `.devhub/screenshots/plugin-ia`；执行前修正脚本对当前 IA 标题、按钮、选择器和运行域可见 Tab 的断言。
- MySQLStore 定向集成测试：首次使用 `GOPROXY=https://goproxy.cn,direct` 仍因 `github.com/go-sql-driver/mysql@v1.8.1` TLS handshake timeout 未进入测试本体；改挂宿主 Go module cache 并设置 `GOPROXY=off GOSUMDB=off` 后通过，5 个子用例均通过。

未执行检查及原因：

- 未新起独立 feishu receiver 跑全新 success / fail / timeout / retry 联调：S15 已有 MySQL + receiver 实跑，本轮用 Go 全量测试与当前 MySQL API/hook_executions 抽查复核；记录为 P2。
- 未执行 `scripts/test-all.sh`：本轮已执行指定最低检查、Go 全量、后台 quick 和指定 E2E。

敏感信息扫描：

- 抽查文件：`/tmp/devhub-s21-effective-config.json`、`/tmp/devhub-s21-allowlist.json`、`/tmp/devhub-s21-secret-center.json`、`/tmp/devhub-s21-hook-executions.json`、`/tmp/devhub-s21-admin-logs.json`、`.devhub/checks/20260529-190252/`、`web/admin-app/test-results`。
- 未发现 Bearer token 明文、Authorization Header、`enc:v*` 密文、`encrypted_value` 字段、测试 token、Webhook Secret / Callback Token 明文或 token_hash。
- MySQL `admin_logs` 中有 2 条 Webhook Secret 创建记录包含 `DEVHUB_PLUGIN_CONFIG_KEYS='[{"id":"local-v1","key":"base64-xxx","primary":true}]'` 这类 root key 环境变量示例文案；未记录实际 root key 值。
- 允许项：`token_ref`、`secret_ref`、`key_id`、masked value、hash/digest、类型标签（如 Webhook Secret / Callback Token）和说明性文案。

问题分级：

- P0：无。
- P1：无。
- P2：未新起独立 feishu receiver 跑全新 success / fail / timeout / retry 联调。

发布建议：`v1.8.4` 生产候选总验收通过，建议进入 S23 发布收口。安全边界保持：未改变 API 语义、插件生命周期、external_service 投递逻辑、SecretCenter 安全模型、Webhook Secret / Callback Token 安全模型、upload / promote / install 语义和 upgrade dry-run 语义；未开放 blocking Hook；未执行第三方代码；未回显 token / secret / Authorization / `encrypted_value` 明文。

## v1.8.4-S20 后台 UI 基础设计系统与插件中心视觉收口验收项

本轮目标：建立后台 design tokens 与通用组件，并优先让插件列表、插件详情、SecretCenter、Webhook 治理、当前生效配置和插件包治理核心入口接入统一视觉。

验收项：

1. 后台存在统一 design tokens：背景、面板、边框、文字、主色、成功、警告、危险、信息、圆角、阴影、页面宽度、间距、表格密度、标签高度和抽屉宽度。
2. 后台存在通用组件：PageHeader、SectionCard、MetricCard、StatusTag、RiskTag、ActionBar、EmptyState、DetailDrawer、TechnicalDetails、InlineHint。
3. 插件列表页标题、说明、主操作按钮统一。
4. 插件列表页关键指标以统一指标卡展示。
5. 插件详情抽屉首屏展示结论：当前状态、当前健康、版本、风险和下一步建议。
6. 插件详情技术详情继续默认折叠。
7. SecretCenter 列表不展示明文 secret、token、Authorization 或 `encrypted_value`。
8. SecretCenter 详情抽屉显示状态标签、使用关系和来源入口，并将元数据技术详情默认折叠。
9. Webhook 治理页总览区将投递记录、待重试、熔断和 external_service 历史失败分开。
10. Webhook retry 按钮仍只在既有可重试记录上展示。
11. 当前生效配置页展示系统版本、store mode、root key、SecretCenter 和 HTTP Allowlist 指标。
12. 当前生效配置 diagnostic_text 可复制，并以技术详情默认折叠展示。
13. 插件包治理入口、本地包与预检入口、版本仓库入口使用统一页面标题区。
14. 旧 `data-testid` 不删除；新增组件只补充兼容选择器。
15. 不改变 API 语义、插件生命周期、external_service 投递逻辑或 SecretCenter 安全模型。

本轮执行结果（2026-05-29）：

1. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-165356/`；仅有 Vite 既有 chunk size warning。
2. `docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js`：连续两次失败，首个用例 beforeEach 中 `GET http://devhub:8090/api/v1/admin/plugins` 报 `getaddrinfo EAI_AGAIN devhub`；未进入页面断言，12 个后续用例未运行。该失败判断为 Compose / E2E 容器内服务名解析环境问题，不是本轮 UI selector 断言失败。
3. `git diff --check`：通过。
4. Go test / Go build：未执行，原因是本轮未修改 Go 后端语义。
5. 后续建议：在 Compose DNS 稳定后重跑 `docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js`，并按需补 1024 / 1366 宽度截图。

S20 后置复核（2026-05-29）：

1. 根因：`devhub` Compose 服务存在，`admin-e2e` 与 `devhub` 同在默认 network，baseURL 为 `http://devhub:8090`；但 `devhub` 启动依赖预构建 `./.devhub/devhub`，当前工作区文件不存在，容器退出后未加入网络，导致 `admin-e2e` 内解析 `devhub` 报 `EAI_AGAIN`。
2. 修复：`devhub` 改为 Go 镜像容器内构建临时 `/tmp/devhub` 并启动 memory store，新增 `/api/v1/health` healthcheck；`admin-e2e` / `frontend-e2e` 改为等待 `service_healthy`；后台 Playwright global setup 增加 DNS / HTTP ready retry；`scripts/check-frontend.sh` 在 E2E 失败时输出 Compose ps、devhub 日志、容器内 DNS 与 health 探测。
3. 诊断：`admin-e2e` 容器内 `getent hosts devhub` 可解析，`wget http://devhub:8090/api/v1/health` 返回 200。
4. 插件治理 E2E 复跑：`docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js` 不再因 DNS / 服务不可达失败；前 3 个用例通过，第 4 个用例进入页面断言后超时于插件详情 Tab `前端挂载` 定位，属于 UI 断言 / 页面结构差异，后续需按 S20 UI 回归处理。
5. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-173455/`；仅有既有 Vite chunk size warning。
6. `git diff --check`：通过。
7. `bash -n scripts/check-frontend.sh`：通过；Docker `node --check web/admin-app/tests/e2e/support/global-setup.js`：通过。宿主机未执行 `node --check`，原因是当前环境没有 `node` 命令。
8. Go test / Go build：未执行，原因是本轮仅修改 Compose / E2E / 前端测试运行配置和文档，未修改 Go 后端代码；Compose devhub 启动时已为 E2E 临时执行 `go build -buildvcs=false -o /tmp/devhub .`。
9. 安全边界：本轮未改变 API 语义、插件生命周期、external_service 投递逻辑、SecretCenter 安全模型，未开放 blocking Hook，未执行第三方代码，未回显 token / secret / Authorization / `encrypted_value`。

S20 插件详情断言适配复核（2026-05-29）：

1. 根因：插件详情抽屉 S20 后将旧 `前端挂载` 独立 Tab 收敛到 `能力` 摘要和 `技术详情`，旧 `Webhook` / `安全凭据` / `审计日志` 等低频 Tab 也收敛到 `能力`、`运行记录` 或对应治理域；测试仍按旧 Tab 文案、英文 schema `required` 文案和全局文本定位断言。
2. 修复：`plugin-governance.spec.js` 改为点击当前可见主 Tab（概览 / 能力 / 运行记录 / 配置 / 技术详情），在 `能力` 面板断言前端挂载说明；schema 错误断言兼容中文 `必填字段`；多匹配文案收窄到当前抽屉 / panel；审计断言改为走当前 `运行与审计` 治理域入口，并保留 API 级审计过滤验证。
3. 稳定选择器：插件详情抽屉现有面板补充 `plugin-capabilities-panel`、`plugin-capabilities-summary`、`plugin-runtime-panel`，未改变页面结构。
4. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-183019/`；仅有既有 Vite chunk size warning。
5. `docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js`：通过，`13 passed`。
6. `git diff --check`：通过。
7. 本轮未修改 shell 脚本，未执行 `bash -n`；未修改 Go 后端，未执行 Go test / Go build。
8. 安全边界：本轮未改变 API 语义、插件生命周期、external_service 投递逻辑、SecretCenter 安全模型，未开放 blocking Hook，未执行第三方代码，未回显 token / secret / Authorization / `encrypted_value`。

## v1.8.4-S19 SecretCenter 操作闭环与当前生效配置排障入口验收项

本轮目标：SecretCenter 详情、使用关系、来源跳转、禁用 / 吊销影响预览、强确认、轮换来源入口和当前生效配置排障入口形成闭环，同时保持不回显敏感明文。

验收项：

1. 系统设置 -> SecretCenter 列表可展示 external_service、Webhook Secret、Callback Token 和测试数据分类。
2. 查看详情抽屉显示 ref、namespace、name、type、status、key_id、usage、associated object、created/updated/last used/last rotated。
3. 详情抽屉显示 available / masked value / 来源类型 / 来源插件 / 来源配置项 / 来源治理页。
4. 空值显示“暂无”，不出现 undefined / null。
5. external_service 使用关系来自 `plugin_external_services.token_ref`，展示 plugin_code、endpoint_url、enabled、current_health、last_success_at、last_failure_at。
6. Webhook Secret 使用关系来自 `plugin_webhook_secrets.secret_ref`，可跳到 Webhook 密钥治理。
7. Callback Token 使用关系来自 `plugin_callback_tokens.token_ref`，可跳到 Callback Token 治理。
8. `plugin_config` namespace 识别为插件配置敏感字段，可跳到插件配置页。
9. 未知来源显示“未知来源”，页面不白屏。
10. 测试 / fixture Secret 标记为测试，并隐藏或禁用危险操作。
11. 禁用前调用 preview，展示影响插件、external_service/webhook/callback 数、可能失败的健康检查 / 投递和受影响业务。
12. 吊销前调用 preview，并要求输入完整 ref 强确认。
13. `active` 可禁用；`disabled` / `revoked` 禁用被拒绝。
14. `active` / `disabled` 可吊销；`revoked` 重复吊销被拒绝。
15. 危险操作后端权限必须拒绝普通后台身份；允许 `super_admin` / `secret.manage` / `system.write` / `plugin.manage`。
16. 禁用 / 吊销后列表和详情刷新。
17. external_service token_ref 被禁用 / 吊销后，health check / 投递返回清晰 token disabled / revoked 原因。
18. 轮换入口不在 SecretCenter 收集明文：external_service 跳插件 external_service 配置，Webhook 跳 Webhook 密钥治理，Callback 跳回调 Token 治理，plugin config 跳插件配置页。
19. 当前生效配置 external_service 行提供去配置、健康检查、查看运行记录、查看 Secret、查看审计。
20. token_ref 行提供查看 Secret、复制 ref、查看审计或来源入口。
21. allowlist 行提供去外部服务策略、复制 origin、查看审计。
22. 脱敏诊断文本可复制，并有中文说明“内容已脱敏”。
23. admin_logs 可通过 ref、plugin_code、origin 或 action 过滤定位。
24. API 响应、admin_logs、hook_executions 不包含 token / secret / Authorization / root key / encrypted_value。
25. 不改变 SecretCenter 加密逻辑。
26. 不改变 Webhook Secret / Callback Token 安全模型。
27. 不开放第三方代码执行、远程插件系统或 blocking Hook。

本轮执行结果（2026-05-29）：

1. 宿主机 `gofmt`：未执行成功，原因是当前环境没有 `gofmt` 命令。
2. Docker `gofmt`：通过。
3. Docker `go test ./...`：首次失败，原因是吊销强确认新增后旧测试未传 `confirm_ref`，且 fixture token_ref 按测试数据规则被阻止危险操作；已修正测试与 fixture 处理后重跑通过。
4. Docker `go build ./...`：通过。
5. `./scripts/check-frontend.sh --admin-only --quick`：通过，最终日志目录 `.devhub/checks/20260529-162605/`，仅有 Vite 既有 chunk size warning。
6. `git diff --check`：通过。
7. 本轮新增 / 修正后端覆盖：Secret 详情 API 脱敏、external_service / Webhook Secret / Callback Token 使用关系、禁用 / 吊销影响预览、吊销强确认失败审计、来源凭据经 SecretCenter 吊销时的 SecretCenter 审计。
8. MySQLStore smoke 与敏感信息 DB/log 扫描：未执行；后续建议按 S15 smoke 复核 SecretCenter / Webhook / Callback / external_service 脱敏链路。
9. 1024 / 1366 浏览器截图和 admin-e2e：未执行；本轮用后台 quick build 覆盖编译层面，后续可按需用 admin-e2e 补截图。

## v1.8.4-S17 系统配置可读视图与 Secret 脱敏展示验收项

本轮目标：系统设置新增“当前生效配置”脱敏可读视图，让管理员能看懂当前真正生效的 external_service、HTTP Allowlist 和 SecretCenter 配置；非敏感配置明文展示，敏感配置只展示 ref、状态、key_id 与脱敏后缀。

验收项：

1. 当前生效配置视图可打开。
2. `endpoint_url` 明文展示。
3. `health_check_path` 明文展示。
4. `timeout_ms` 明文展示。
5. `failure_policy` 明文展示。
6. HTTP Allowlist origin 明文展示，并区分系统默认 / 环境变量 / 后台配置 / 最终生效列表。
7. token 不明文展示。
8. secret 不明文展示。
9. Authorization 不明文展示。
10. root key 不明文展示。
11. `encrypted_value` 不展示。
12. `token_ref` 可见。
13. `token_status` 可见。
14. `token_key_id` 可见。
15. `token_masked` 可见，如当前 keyring 可安全生成脱敏后缀。
16. SecretCenter 列表显示用途。
17. SecretCenter 列表显示最近使用。
18. SecretCenter 列表显示最近轮换。
19. SecretCenter 列表显示 key_id。
20. SecretCenter 列表不显示明文或密文。
21. external_service `token_ref` 缺失时提示清晰。
22. `token_ref` 不存在时提示清晰。
23. `token_ref` disabled / revoked 时提示清晰。
24. 复制脱敏诊断信息可用。
25. 复制内容不包含 token 明文。
26. 复制内容不包含 secret 明文。
27. 复制内容不包含 Authorization。
28. 复制内容不包含 `DEVHUB_PLUGIN_CONFIG_KEYS`。
29. 复制内容不包含 `encrypted_value`。
30. 页面不出现 undefined / null。
31. 不改变 SecretCenter 加密逻辑。
32. 不改变 Webhook Secret / Callback Token 安全模型。

本轮执行结果（2026-05-29）：

1. 宿主机 `gofmt`：未执行成功，原因是当前环境没有 `gofmt` 命令。
2. Docker `gofmt`：通过。
3. Docker `go test ./...`：通过。
4. Docker `go build ./...`：通过。
5. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-143634/`，仅有 Vite 既有 chunk size warning。
6. 1024 / 1366 浏览器宽度：未执行 Playwright 截图检查；本仓库规则将浏览器 E2E 作为手工入口，本轮用后台 quick build 覆盖编译层面，后续可按需用 admin-e2e 补截图。
7. MySQL 明文扫描：未重新执行；本轮没有改变 SecretCenter 加密或落库逻辑，S14/S15 已有 MySQL 明文扫描记录。后续如需复核，可按 S15 smoke 执行。

## v1.8.4-S11 Webhook 联调体验与 external_service 配置治理验收项

建议手工命令：

```bash
gofmt -w internal/domain/api_error.go internal/domain/plugin_external_service.go internal/service/plugin_external_service.go internal/service/plugin_external_service_delivery.go internal/transport/httpapi/router.go
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build
```

联调建议流程：

1. 启动 feishu_link 独立测试服务：`http://127.0.0.1:18081`。
2. Docker 后端下先配置 `endpoint_url=http://127.0.0.1:18081`，验证连接失败提示 Docker loopback。
3. 改为 `endpoint_url=http://172.17.0.1:18081`，不加 allowlist 时验证 `endpoint_safety_rejected` 提示。
4. 使用 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build` 重启。
5. 保存 external_service 运行配置，`health_check_path=/health`，执行 health check。
6. 触发 `AfterCreateContent`，确认测试服务 `/requests` 出现 `POST /hooks/content.after_create` 且 `authorization_present=true`；`GET /health` 成功不要求 `/requests` 有记录，因为该服务只记录 `POST /hooks/*`。

验收项：

1. `endpoint_url=http://127.0.0.1:18081` 在 Docker 后端连接失败时提示 Docker loopback 问题。
2. Docker loopback 提示包含 `host.docker.internal`。
3. Docker loopback 提示包含 Docker host gateway。
4. Docker loopback 提示包含宿主机局域网 IP。
5. `endpoint_url=http://172.17.0.1:18081` 被安全策略拒绝时提示 HTTP allowlist。
6. 安全拒绝提示包含 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST` 示例。
7. `dev.sh` 透传 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST`。
8. 插件详情概览有“配置 external_service”入口。
9. Webhook 治理页有“配置 external_service”入口。
10. 插件详情显示当前实际运行 external_service endpoint。
11. 插件详情说明全局配置不会自动覆盖 external_service 运行配置。
12. 全局配置页出现 `endpoint_url`、`health_check_path`、`token`、`timeout_ms`、`failure_policy` 字段时提示用户去 external_service 配置；普通全局 `enabled` 不应误触发该提示。
13. external_service 配置保存后 health check 使用新 endpoint。
14. external_service 投递使用运行配置 endpoint。
15. 当前健康、最近成功、最近失败、历史失败计数分开展示。
16. 当前 healthy 时旧失败不再展示为当前异常。
17. 历史失败仍可在运行记录中查看。
18. 历史失败带时间戳。
19. `plugin_package_already_installed` 提示包含配置插件入口。
20. `plugin_package_already_installed` 提示包含配置 external_service 入口。
21. `plugin_package_already_installed` 提示包含查看版本 / 升级差异入口。
22. 本地仓库同 code 已安装包显示“升级差异”而不是单纯“可对比”。
23. 插件包详情抽屉显示下一步建议。
24. `18081` feishu_link 测试服务端口说明已记录。
25. `18090` mock receiver 端口如存在已区别说明。
26. token 不回显。
27. Authorization 不回显。
28. secret 不回显。
29. 不开放 blocking Hook。
30. 不执行第三方代码。
31. 不改变 Webhook Secret / Callback Token 安全模型。

本轮执行结果（2026-05-28）：

1. 宿主机 `gofmt`：未执行成功，原因是当前环境没有 `gofmt` 命令。
2. Docker `gofmt`：通过，使用 `golang:1.22-bookworm` 格式化 Go 变更文件。
3. Docker `go test ./...`：通过。
4. Docker `go build ./...`：在 `golang:1.22-bookworm` 容器内因缺少 git 导致 VCS stamping 失败；改用 `go build -buildvcs=false ./...` 通过。
5. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260528-130051/`。
6. `git diff --check`：通过。
7. 手工联调：未执行；未启动 feishu_link receiver，未在本地 Docker 后端下实际验证 `127.0.0.1` connection refused、`172.17.0.1` allowlist、AfterCreateContent 投递、`/requests`、fail / timeout / retry_scheduled 和 manual retry。请按本节联调建议流程补验。

## v1.8.4-S12 Webhook 联调体验回归与后台点击验证（S11 回归验收）

说明：S12 只做 S11 回归与必要小修（不加新能力、不开放 blocking Hook、不执行第三方代码、不改变 Webhook Secret / Callback Token 安全模型）。

本轮手工联调执行要点（2026-05-28）：

1. 以 Docker 后端运行 DevHub，feishu_link receiver 在宿主机 `18081` 端口运行。
2. 配置 `endpoint_url=http://127.0.0.1:18081` 后执行 health check，返回 `error_code=network_connection_refused`，suggestion 明确：容器内 `127.0.0.1` 指向容器自身，并提示 `host.docker.internal` / Docker host gateway / 宿主机局域网 IP。
3. 配置 `endpoint_url=http://172.17.0.1:18081` 且未设置 allowlist 时保存被拒绝：`code=endpoint_safety_rejected`，suggestion 明确策略并展示推荐命令 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build`。
4. 设置 allowlist 后重启并保存 external_service 运行配置成功，health check `healthy`。
5. 创建 `feishu_link` 内容触发 `AfterCreateContent`：feishu_link receiver `/requests` 出现 `POST /hooks/content.after_create`，且 `authorization_present=true`；DevHub 响应与执行记录不回显 token / Authorization 明文。
6. 插件健康摘要验证：当前状态为 `healthy` 时，历史失败时间戳仍可在记录中查看，但不再覆盖顶部“当前健康”。
7. 重复安装验证：对已安装的 `plugin_a7b0cc04` 本地包重复 install 返回 `plugin_package_already_installed`，并提供“配置插件 / 配置 external_service / 查看版本与升级差异”等下一步引导字段。
8. 联调脚本：使用 `scripts/run-feishu-webhook-flow.sh` 一键跑通“配置 external_service -> health check -> 创建内容 -> 外部投递成功”的闭环；如宿主机 `18081` 已被 feishu_link receiver 占用，可设置 `DEVHUB_START_RECEIVER=0` 复用已有 receiver。
9. 回归小修：联调过程中发现后台 `POST /api/v1/admin/posts` 在设置非默认 status 时可能 panic（`SetTopicStatus` 返回 nil 覆盖 topic 指针）；S12 已修复并在脚本中回归通过。

本轮检查命令（2026-05-28）：

```bash
docker run --rm -v "$PWD:/app" -w /app golang:1.22-alpine gofmt -w internal/transport/httpapi/router.go
docker run --rm -v "$PWD:/app" -w /app golang:1.22-alpine go test ./...
docker run --rm -v "$PWD:/app" -w /app golang:1.22-alpine go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
```

本轮执行结果（2026-05-28）：

1. Docker `gofmt`：通过；宿主机缺少 `gofmt`。
2. Docker `go test ./...`：通过。
3. Docker `go build ./...`：通过。
4. `git diff --check`：通过。
5. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260528-153818/`。

## v1.8.4-S13 系统配置中心与敏感配置治理收口验收项

建议命令：

```bash
gofmt -w internal/domain/plugin_config_keys.go internal/domain/system_sensitive_config.go internal/service/system_sensitive_config.go internal/service/webhook_secrets.go internal/transport/httpapi/router.go internal/transport/httpapi/system_sensitive_config_handler.go
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
```

验收项：

1. 系统设置页出现“敏感配置状态（只读）”卡片，展示启动 keyring 状态（ok/blocked）、current_key_id 和 env 示例。
2. 系统设置页只读展示 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST` 的解析结果（allowlist_origins），不允许在 UI 中修改。
3. 启动密钥缺失时：
   - key 状态为 blocked；
   - 创建 Webhook Secret 会被前端 readiness 拦截或后端拒绝；
   - 后端错误码仍为 `webhook_secret_encryption_key_missing`，并在 suggestion 中包含 JSON 形式与 split 形式两种 env 示例，且明确“后台不会保存或生成 root key、修改需重启”。
4. 启动密钥已配置时：key 状态为 ok，Webhook Secret 可创建/轮换，明文只在创建/轮换成功响应里展示一次。
5. 后台表格中 `secret_ref/token_ref` 统一脱敏显示（可复制引用），不在 UI 中到处铺开完整引用字符串。
6. 不把 Webhook Secret 明文、Callback Token 明文、external_service token 明文、Authorization header 或 root key 写入 API 响应（除创建/轮换“仅展示一次”响应外）、日志或 `admin_logs`。

## v1.8.4-S14 Core SecretCenter 引用层与 external_service token_ref 落地验收项

建议命令：

```bash
# 宿主机若缺少 gofmt，可用 Docker gofmt（示例见下方执行结果）
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
./dev.sh restart --mysql --no-build
```

验收项（最小闭环 + 安全边界）：

1. SecretCenter 支持 `secret_ref=secret://{namespace}/{name}` 引用规范，且有明确校验规则（namespace/name 小写字母/数字/_/-，name 可用 `/` 分段）。
2. invalid secret_ref（空值、非 `secret://`、包含 `..`、包含反斜杠、非法字符、空 namespace/name）会被拒绝。
3. SecretCenter 的 create/update 在 root key 缺失时必须失败，并返回两种 env 示例（JSON/split），明确“root key 只来自启动环境变量、后台不会保存或生成、修改需重启”。
4. SecretCenter 不提供 Admin API 明文读取接口；列表/详情不返回明文 value，也不返回 `encrypted_value`。
5. SecretCenter 支持最小生命周期：create、update（轮换）、disable、revoke；disabled/revoked 的 secret_ref 运行时 resolve 必须拒绝。
6. SecretCenter resolve 仅供后端运行时内部使用，并更新 `last_used_at/usage_count`（best-effort）。
7. MySQL 存在 `secret_refs` 存储表（或等价存储），且不保存 plaintext token/secret。
8. MemoryStore 与 MySQLStore 口径一致：同一组 SecretCenter API 在两种 store 下返回一致字段语义（差异仅限持久化能力）。
9. external_service 保存 bearer token 时写入 SecretCenter，并生成/更新 `token_ref=secret://external_service/{plugin_code}/token`。
10. external_service 运行配置存储不返回 plaintext token；配置查询返回 `token_ref` 与 `token_secret` 元数据（status/key_id/last_used_at/rotated_at）。
11. external_service health check 在 `token_ref` missing/disabled/revoked 时返回清晰错误码与 suggestion，不回显 token 明文。
12. external_service delivery 在运行时 resolve token_ref 获取 Authorization header（仅内部使用），不会把 token/Authorization 写入 `hook_executions`。
13. `hook_executions` 不包含 token/secret/Authorization 明文，也不包含敏感 payload 明文（仅允许 excerpt/sha256 等安全摘要）。
14. `admin_logs` 不包含 token/secret/root key 明文，也不包含 `encrypted_value` 全量值。
15. 系统设置只读敏感配置状态接口 `GET /api/v1/admin/system/sensitive-config/status` 返回 SecretCenter 统计（secret_refs 总数与 namespace_counts），不返回任何明文 secret/token/key。
16. 后台系统设置页可只读展示 SecretCenter 统计与 external_service namespace count。
17. 后台插件详情 external_service 区域展示 token_ref 与 token_secret 元数据，不回显明文。
18. Webhook Secret 与 Callback Token 现有安全模型不变：明文只在创建/轮换成功响应里展示一次，不进入列表/审计/执行记录。
19. 不把 `DEVHUB_PLUGIN_CONFIG_KEYS` 保存到后端数据库，不提供后端创建/保存 root key 的能力。
20. 不把 root key 明文写入 DB、API 响应、`admin_logs` 或 `hook_executions`。
21. 不引入 KMS/Vault 集成，不引入 Docker/K8s secret 自动化，不引入复杂多租户 secret 模型。
22. 不开放 blocking Hook、不执行第三方代码、不开放 Go plugin、JS 沙箱、远程 iframe、插件市场或远程在线安装。
23. external_service legacy token 存量兼容：历史记录可继续运行（不会回显明文），新写入统一走 SecretCenter token_ref。
24. 新增 SecretCenter API 与 schema 变更必须有审计：secret create/update/disable/revoke 与 external_service token_ref 绑定变更必须写 `admin_logs`（metadata 不含明文）。
25. MySQL dev smoke：`./dev.sh start|restart --mysql` 后能登录后台并打开系统设置页，API 不 500。
26. MySQL 中 `secret_refs.encrypted_value` 不应出现明文 token（应为 `enc:v2:*` 形式）。
27. MySQL 中 `plugin_external_services.token_ciphertext/token_hash` 在新写入路径下应为空，仅保存 `token_ref`。
28. MySQL 中 `admin_logs` / `hook_executions` 搜索不到 token 明文（例如 `%Bearer %` 或测试 token 字符串）。
29. disabled/revoked token_ref 在 health check/delivery 中会被拒绝，并留下明确错误码，便于排障。
30. 文档同步：API、测试、release notes、项目进度、Webhook 使用文档必须更新，且不得把未来规划写成已完成能力。

本轮执行结果（2026-05-28）：

1. Docker `gofmt`：通过（宿主机无 gofmt）。
2. Docker `go test ./...`：通过；新增覆盖 SecretCenter root key 缺失、external_service bearer token_ref 保存 / health check / delivery resolve、revoke 阻断和 admin_logs / hook_executions 脱敏。
3. Docker `go build ./...`：在容器内因缺少 git 的 VCS stamping 失败；改用 `go build -buildvcs=false ./...` 通过。
4. `git diff --check`：通过。
5. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260528-192821/`。
6. `./dev.sh start --mysql`：通过完成 MySQL 与前后台构建；首次因 8090 已有 Go 容器而复用服务，随后停止旧容器并用 `./dev.sh start --mysql --no-build` 启动当前代码的 MySQL 后端完成 smoke。临时 Go 容器已停止；MySQL dev 模式下验证：
   - `GET /api/v1/admin/system/sensitive-config/status` 返回 `secret_center` 统计，且不包含任何明文 secret/token/key。
   - `PUT /api/v1/admin/plugins/qa/external-service` 写入 bearer token 后返回 `token_ref=secret://external_service/qa/token` 与 `token_secret` 元数据，不回显明文。
   - MySQL `secret_refs.encrypted_value` 为 `enc:v2:*`，不保存 token 明文；`plugin_external_services.token_ciphertext/token_hash` 为空。
   - `admin_logs`、`hook_executions` 中搜索不到测试 token 明文，也搜索不到 `Bearer ` / `Authorization`。
   - `POST /api/v1/admin/secret-center/secrets/disable|revoke` 后 external_service health check 返回 `plugin_external_service_token_disabled|plugin_external_service_token_revoked` 且原因清晰。
   - 另起无 root key 的 8092 临时 MySQL 后端后，Webhook Secret 创建返回 `webhook_secret_encryption_key_missing`，suggestion 包含 JSON/split 两种 env 示例并明确“后台不会保存或生成 root key、修改需重启”。

## v1.8.4-S15 SecretCenter + Webhook + external_service MySQL 实跑验收

说明：S15 是 MySQLStore 下的实跑验收与收口修复，不新增插件运行能力，不开放 blocking Hook，不执行第三方代码，不改变 Webhook Secret / Callback Token 安全模型。

环境前置：

```bash
DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18081 ./dev.sh restart --mysql --no-build
NO_PROXY=127.0.0.1,localhost no_proxy=127.0.0.1,localhost <smoke-command>
```

验收项：

1. SecretCenter 直接 create/update 成功，MySQL `secret_refs.encrypted_value` 为 `enc:v2:*`，不包含测试明文。
2. external_service 保存 bearer token 后只持久化 `token_ref=secret://external_service/{plugin_code}/token`，`token_ciphertext/token_hash` 在新路径下为空。
3. external_service health check 通过真实 receiver 返回 healthy。
4. 创建 `feishu_link` 内容触发 `AfterCreateContent` external_service 投递，`hook_executions` 成功且 response_status 为 200。
5. Webhook Secret create/rotate 成功；明文仅在创建/轮换响应中返回一次，列表/详情不回显；MySQL `plugin_webhook_secrets.secret_ciphertext` 为 `enc:v2:*`。
6. Callback Token create 成功；列表/详情不回显明文；插件服务回调 `GET /api/v1/plugin-callback/config` 与 `POST /api/v1/plugin-callback/audit-events` 成功并写入 accepted callback request。
7. MySQL 扫描 `secret_refs`、`plugin_external_services`、`plugin_webhook_secrets`、`plugin_callback_tokens`、`plugin_callback_requests`、`admin_logs`、`hook_executions`，不得出现本轮测试敏感值、`Bearer ` 或 `Authorization`。
8. 文档记录实跑结果和发现的 MySQLStore 问题；不把测试 token、Webhook Secret、Callback Token、Authorization header 或 root key 写进文档。

本轮执行结果（2026-05-28）：

- `./dev.sh restart --mysql --no-build`：通过；后端以 `CMS_STORE=mysql` 运行，`DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18081` 生效。
- S15 MySQL smoke：通过。
  - SecretCenter direct create/update：通过，`secret_refs` 密文落库。
  - external_service bearer token_ref + health check：通过，receiver health 为 healthy。
  - external_service actual webhook delivery：通过，创建 topic 后 `hook_executions` 成功且 response_status=200。
  - Webhook Secret create/rotate：通过，API 列表/详情脱敏，MySQL 密文落库。
  - Callback Token create + config/audit callbacks：通过，2 条 callback request 为 accepted。
  - DB/API plaintext leakage scan：通过；扫描 6 个本轮测试敏感值，最近 `secret_refs=6`、`admin_logs=62`、`hook_executions=13`、`plugin_callback_requests=2` 均未出现测试明文、`Bearer ` 或 `Authorization`。
- 实跑中发现并修复：
  - MySQLStore `AppendPluginWebhookSecret` 的 `INSERT plugin_webhook_secrets` values 比列数多 1 个占位，导致 Webhook Secret 创建失败。
  - MySQLStore `AppendPluginCallbackToken` 的 `INSERT plugin_callback_tokens` values 比列数多 1 个占位，导致 Callback Token 创建失败。
- 未运行 `go test`、Playwright 或 `scripts/test-all.sh`；本仓库规则要求这些作为手工入口，后续一键复核可手动执行 `scripts/test-all.sh`。

2026-05-29 对照复核清单（基于 2026-05-28 S14/S15 已执行记录整理，未重新跑 MySQL smoke）：

1. `./dev.sh start --mysql`：已在 S14 记录通过；S15 使用 `./dev.sh restart --mysql --no-build` 跑通 MySQLStore 实跑。
2. 未配置 `DEVHUB_PLUGIN_CONFIG_KEYS` 时 Webhook Secret 创建被拦截：S14 无 root key 临时后端验证通过，返回 `webhook_secret_encryption_key_missing` 与两种 env 示例。
3. 配置 `DEVHUB_PLUGIN_CONFIG_KEYS` 后 SecretCenter 状态正常：S14/S15 系统敏感配置状态接口与 SecretCenter create/update 验证通过。
4. external_service token 保存为 `token_ref`：S14/S15 验证保存 bearer token 后返回并持久化 `secret://external_service/{plugin_code}/token`。
5. MySQL 里不存 token 明文：S15 验证 `secret_refs.encrypted_value` 为 `enc:v2:*`，`plugin_external_services.token_ciphertext/token_hash` 在新路径下为空。
6. `admin_logs` 不含 token / secret / Authorization 明文：S14/S15 扫描通过。
7. `hook_executions` 不含 token / secret / Authorization 明文：S14/S15 扫描通过。
8. `feishu_link` 投递能通过 `token_ref` 获取 token：S15 创建 `feishu_link` 内容触发 `AfterCreateContent` 投递成功，`hook_executions.response_status=200`。
9. disabled / revoked `token_ref` 能阻断投递：S14 验证 health check 返回 `plugin_external_service_token_disabled` / `plugin_external_service_token_revoked`，并留下清晰原因。
10. Webhook Secret / Callback Token 现有语义未破坏：S15 验证 Webhook Secret create/rotate 明文仅返回一次、Callback Token create + config/audit callbacks 成功，列表/详情/审计不回显明文。

## v1.8.4-S15 系统敏感配置状态卡片 UI 优化验收项

说明：本轮只优化后台系统设置页“敏感配置与运行安全状态”的展示体验，不改变后端 API、SecretCenter 语义、external_service token_ref 语义、Webhook Secret / Callback Token 安全模型，也不开放后台编辑 root key 或 HTTP Allowlist。

验收项：

1. 敏感配置状态区域可打开，标题为“敏感配置与运行安全状态”。
2. 插件配置加密卡片显示正常。
3. SecretCenter 引用层卡片显示正常。
4. external_service HTTP Allowlist 卡片显示正常。
5. 状态标签显示为“正常 / 未配置 / 异常 / 未就绪 / 已配置”等中文文案。
6. `current_key_id` 显示为“当前密钥 ID”。
7. `key_count` 显示为“可用密钥数”。
8. `secret_refs` 显示为“Secret 引用数”。
9. 环境变量示例不使用 textarea。
10. 环境变量示例使用代码块或只读 code panel。
11. 复制按钮可用。
12. 复制成功有中文提示。
13. 示例值标注不能直接用于生产。
14. root key 不回显。
15. token 不回显。
16. secret 不回显。
17. Authorization 不回显。
18. HTTP allowlist 明确为只读。
19. 后台不支持编辑 `DEVHUB_PLUGIN_CONFIG_KEYS`。
20. 后台不支持编辑 HTTP allowlist。
21. 1024 宽度下布局正常。
22. 1366 宽度下布局正常。
23. 页面不出现 undefined / null。
24. 本轮未改变 SecretCenter 后端语义。
25. 本轮未改变 Webhook Secret / Callback Token 安全模型。

本轮执行结果（2026-05-28）：

- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260528-211200/`。
- admin-e2e 轻量浏览器检查：通过，截图目录 `.devhub/screenshots/s15-sensitive-status/`。
  - 1024 宽度：3 张状态卡片可见，textarea 数量为 0，未发现 undefined/null/原始字段/裸 ok，未发现 root key/token/secret/Authorization。
  - 1366 宽度：3 张状态卡片可见，textarea 数量为 0，未发现 undefined/null/原始字段/裸 ok，未发现 root key/token/secret/Authorization。
- `git diff --check`：通过。
- 未运行 `go test ./...` / `go build ./...`：本轮未修改 Go。
- 未运行 `./scripts/check-frontend.sh --frontend-only --quick`：本轮未修改前台用户侧应用。
- 2026-05-29 复核：第三张卡片标题与验收口径统一为 `external_service HTTP Allowlist`；仍只改后台 UI 文案和文档。`git diff --check` 通过；`./scripts/check-frontend.sh --admin-only --quick` 通过，日志目录 `.devhub/checks/20260529-105653/`，仅有 Vite 既有 chunk size warning。未运行 Go 检查：本轮未修改 Go；未运行 `./scripts/check-frontend.sh --frontend-only --quick`：本轮未修改前台用户侧应用。

## v1.8.4-S16 系统设置页安全状态与 HTTP Allowlist 受控配置验收项

建议命令：

```bash
gofmt -w internal/domain/plugin_external_service.go internal/domain/system_settings.go internal/service/external_service_http_allowlist.go internal/service/plugin_external_service.go internal/service/plugin_external_service_delivery.go internal/service/system_sensitive_config.go internal/service/plugin_external_service_test.go internal/store/memory.go internal/store/mysql.go internal/store/schema.go internal/transport/httpapi/router.go internal/transport/httpapi/system_sensitive_config_handler.go
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
```

MySQL 手工建议：

```bash
./dev.sh start --mysql
# 登录后台后新增 http://172.17.0.1:18081，并检查 system_settings / admin_logs / external_service endpoint 校验。
```

验收项：

1. root key 卡片只读。
2. root key 不能后台编辑、保存或生成。
3. 后台不能编辑 `DEVHUB_PLUGIN_CONFIG_KEYS`。
4. SecretCenter 卡片有查看 Secret 和查看审计入口。
5. Secret 管理页未完整实现时显示中文占位，不白屏。
6. external_service HTTP Allowlist 卡片可打开管理区。
7. 系统默认 localhost / 127.0.0.1 / ::1 allowlist 可见。
8. 环境变量 allowlist 来源可见。
9. 后台 allowlist 来源可见。
10. 最终生效 allowlist 可见。
11. 后台 allowlist 可新增 exact origin。
12. 后台 allowlist 可删除。
13. 系统默认和环境变量来源不可删除。
14. wildcard 被拒绝。
15. `0.0.0.0` / `0.0.0.0/0` 被拒绝。
16. CIDR 被拒绝。
17. 带 path 的 origin 被拒绝。
18. 带 query / fragment 的 origin 被拒绝。
19. 非 `http://` scheme 被拒绝。
20. 非 localhost HTTP 未在 allowlist 时被拒绝。
21. 非 localhost HTTP 在后台 allowlist 时允许。
22. 删除后台 allowlist 后再次拒绝。
23. 新增 allowlist 写 `admin_logs`。
24. 删除 allowlist 写 `admin_logs`。
25. 普通管理员无权修改，只读可见。
26. API / 页面不回显 token / secret / Authorization。
27. 不改变 SecretCenter 后端语义。
28. 不改变 Webhook Secret / Callback Token 安全模型。
29. 不开放 blocking Hook。
30. 不执行第三方代码。

本轮执行结果（2026-05-29）：

- 宿主机 `gofmt`：未执行成功，原因是当前环境没有 `gofmt` 命令。
- Docker `gofmt`：通过。
- 定向 `go test ./internal/service -run "TestExternalServiceHTTPAllowlist|TestExternalServiceAdminHTTPAllowlist|TestExternalServiceHTTPAllowlistOriginValidation" -count=1`：通过。
- Docker `go test ./...`：通过。
- Docker `go build ./...`：通过。
- `git diff --check`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-121151/`，仅有 Vite 既有 chunk size warning。
- `./scripts/check-frontend.sh --frontend-only --quick`：误由文档检索命令中的反引号触发并通过，日志目录 `.devhub/checks/20260529-120820/`；该项不是本轮要求的后台 quick 检查，不替代 `--admin-only --quick`。
- MySQL smoke：`./dev.sh start --mysql --no-build` 首次因 8090 已有 DevHub 服务复用旧进程；随后执行 `./dev.sh restart --mysql --no-build` 启动当前代码。验证 `GET/POST/DELETE /api/v1/admin/system/external-service/http-allowlist` 可用；新增 `http://172.17.0.1:18081` 后 `PUT /api/v1/admin/plugins/qa/external-service` 允许该 HTTP endpoint；删除后同 endpoint 再次返回 `endpoint_safety_rejected`。MySQL `system_settings(namespace=external_service,key=http_allowlist)` 存在，删除后 `JSON_LENGTH(value_json)=0`；`admin_logs` 存在新增 / 删除审计记录，origin 为 `http://172.17.0.1:18081`。Smoke 后已执行 `./dev.sh stop` 停止当前 Go 服务。

### S16 IA 补充验收项

1. 系统设置页存在二级导航：总览 / 安全与密钥 / 外部服务策略 / SecretCenter / 配置审计。
2. 总览页只展示三张摘要卡片，标题严格为“启动级加密密钥 / SecretCenter 引用层 / external_service HTTP 策略”。
3. 环境变量示例只出现在“安全与密钥”。
4. root key 在“安全与密钥”中仍只读，不能后台编辑。
5. HTTP Allowlist 管理只出现在“外部服务策略”。
6. 后台 allowlist 新增 / 删除入口仍可用。
7. SecretCenter 入口和 Secret 元数据列表在“SecretCenter”。
8. SecretCenter 列表不显示明文 secret 或 encrypted_value。
9. admin_logs / 配置变更记录筛选和列表在“配置审计”，该页不夹带基础设置表单。
10. 页面不回显 token / secret / Authorization。
11. 本轮不改变后端 allowlist、SecretCenter、Webhook Secret 或 Callback Token 安全语义。

S16 IA 补充执行结果（2026-05-29）：

- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-133146/`，仅有 Vite 既有 chunk size warning。
- 本次精确命名与归位复核后再次执行 `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-134509/`，仅有 Vite 既有 chunk size warning。
- `git diff --check`：通过。
- 未运行 Go 测试 / Go 构建：本轮只调整后台页面 IA、文案和入口组织，未修改后端 Go 语义。

### S17 系统设置二级导航与安全配置 IA 收口验收项

1. 系统设置页仍为 5 个 Tab：总览 / 安全与密钥 / 外部服务策略 / SecretCenter / 配置审计。
2. 总览只显示三张状态摘要卡片，不显示环境变量代码块。
3. 安全与密钥只展示 root key 状态、root key 环境变量示例和“后台不能修改 root key”提示。
4. external_service HTTP allowlist 的环境变量示例在“外部服务策略”，不出现在“安全与密钥”。
5. 外部服务策略集中展示 allowlist 状态、来源、后台受控新增 / 删除入口。
6. SecretCenter 专门展示 secret_ref / token_ref 元数据，不显示明文 secret 或 encrypted_value。
7. 配置审计专门展示 admin_logs / 配置变更记录。
8. 页面不回显 token / secret / Authorization。
9. 本轮不改变 SecretCenter、Webhook、external_service 的后端安全语义。

S17 执行结果（2026-05-29）：

- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-135228/`，仅有 Vite 既有 chunk size warning。
- `git diff --check`：通过。
- 未运行 Go 测试 / Go 构建：本轮只调整后台页面 IA、文案和入口组织，未修改后端 Go 语义。

### S18 SecretCenter 页面文案与治理体验优化验收项

1. SecretCenter 页面可打开。
2. 页面标题为“敏感配置引用”或同等清晰文案。
3. 页面说明 secret_ref / token_ref 是引用，不是明文。
4. 页面说明不会显示 token / secret 明文。
5. Secret 引用字段改为“敏感配置引用”。
6. 命名空间字段改为“所属业务”。
7. key_id 字段改为“加密密钥 ID”。
8. active 状态显示为“正常”。
9. disabled 状态显示为“已禁用”。
10. revoked 状态显示为“已吊销”。
11. 类型筛选可用。
12. 状态筛选可用。
13. 关键词搜索可用。
14. external_service Secret 标记为外部服务。
15. webhook Secret 标记为 Webhook Secret。
16. callback Secret 标记为 Callback Token。
17. s15smoke / e2e / fixture / test / demo 数据标记为测试数据。
18. 复制引用按钮只复制 ref。
19. 详情抽屉不显示明文。
20. 审计入口可用或有中文占位。
21. 未实现的轮换操作不做假成功：external_service / Webhook / Callback 引导到来源治理页，其他 Secret 显示中文占位。
22. 页面不显示 token 明文。
23. 页面不显示 secret 明文。
24. 页面不显示 Authorization。
25. 页面不显示 encrypted_value 全文。
26. 页面不显示 root key。
27. 页面不允许编辑 DEVHUB_PLUGIN_CONFIG_KEYS。
28. 页面不出现 undefined / null。
29. 本轮未改变 SecretCenter 后端加密语义。
30. 本轮未改变 Webhook Secret / Callback Token 安全模型。

建议浏览器检查：

1. 打开系统设置 -> SecretCenter。
2. 页面标题显示为“敏感配置引用”。
3. 说明卡片清楚说明“不显示明文”。
4. 表格字段中文化。
5. 类型筛选、状态筛选和搜索可用。
6. 测试数据显示“测试”标签。
7. 复制引用按钮可用，复制的是 ref。
8. 查看详情抽屉可用，不显示 token / secret / Authorization / encrypted_value。
9. 查看审计入口可用。
10. external_service 类型可跳转到 external_service 配置；Webhook / Callback 类型可跳转到对应治理页。
11. 未实现的轮换入口显示中文说明，不做假成功。

S18 执行结果（2026-05-29）：

- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260529-140945/`，仅有 Vite 既有 chunk size warning。
- `git diff --check`：通过。
- 未运行 Go 测试 / Go 构建：本轮只调整后台页面文案、信息结构、筛选和基础治理入口，未修改 Go。
- 未运行 `./scripts/check-frontend.sh --frontend-only --quick`：本轮未修改前台用户侧应用。

## v1.8.4-S5 生产备份 / 回滚 / 升级演练增强验收项

说明：S5 不新增完整自动 rollback、migration down 或运行时能力，主要补齐生产备份、安装 / 升级前检查、失败阶段、PluginRegistry reload 失败和人工恢复路径的文档口径。当前以 `docs/BACKUP_AND_ROLLBACK.md`、`docs/DEPLOYMENT.md`、`docs/PLUGIN_PACKAGE.md`、`docs/PLUGIN_DEVELOPER_GUIDE.md` 和 release notes 为准。

建议命令：

```bash
git diff --check
rg -n "backup|备份|rollback|回滚|failure_stage|failure_reason|dry-run|migration down|PluginRegistry reload|加密 key|MemoryStore|MySQLStore|blocked|warning|confirm" docs/BACKUP_AND_ROLLBACK.md docs/DEPLOYMENT.md docs/PLUGIN_PACKAGE.md docs/PLUGIN_DEVELOPER_GUIDE.md docs/PLUGIN_SYSTEM_ROADMAP.md docs/TESTING.md docs/releases/v1.8.4.md CHANGELOG.md README.md
```

生产演练建议步骤：

```bash
./scripts/plugin-package-build.sh examples/plugins/official_links
./scripts/plugin-package-check.sh examples/plugins/official_links
./scripts/plugin-package-build.sh examples/plugins/official_webhook_notify
./scripts/plugin-package-check.sh examples/plugins/official_webhook_notify
./dev.sh start --mysql
./scripts/check-frontend.sh --admin-only --quick
```

手动验收流程：

1. package build。
2. package check。
3. upload / precheck。
4. promote 到本地插件仓库。
5. install dry-run。
6. install。
7. upgrade dry-run safe。
8. upgrade dry-run warning 并用 `confirm=true` 确认。
9. upgrade dry-run blocked，确认不能通过 `confirm=true` 绕过。
10. dry-run plan 过期拒绝。
11. checksum 不一致拒绝。
12. migration plan 不一致拒绝。
13. PluginRegistry reload 成功路径；失败路径只在预发用可恢复数据演练。
14. MySQL 模式启动并验收插件安装、配置、`admin_logs`、`hook_executions` 和 upgrade dry-run。

验收项：

1. 生产备份清单已记录。
2. 插件配置加密 key 备份说明已记录。
3. 插件包本地仓库备份说明已记录。
4. `admin_logs` 保留策略已记录。
5. `hook_executions` 保留策略已记录。
6. MemoryStore 生产限制已记录。
7. MySQLStore 生产建议已记录。
8. 插件安装前检查清单已记录。
9. 插件升级前检查清单已记录。
10. upgrade dry-run `safe` 演练已记录。
11. upgrade dry-run `warning` `confirm=true` 演练已记录。
12. upgrade dry-run `blocked` 演练已记录。
13. dry-run plan 过期拒绝说明已记录。
14. checksum 不一致拒绝说明已记录。
15. migration plan 不一致拒绝说明已记录。
16. `failure_stage` 说明已记录。
17. `failure_reason` 说明已记录。
18. `registry_reload` 失败处理说明已记录。
19. rollback 支持边界已记录。
20. migration down 支持边界已记录。
21. 配置恢复路径已记录。
22. 插件包本地仓库恢复路径已记录。
23. `disabled` / `archived` 不是完整回滚的说明已记录。
24. `migrations/` 仍是唯一入口。
25. dry-run 不执行 SQL。
26. 不执行第三方代码。
27. 不开放插件市场。
28. 不开放远程在线安装。
29. 不开放 blocking Hook。
30. Webhook Secret 元数据备份说明已记录，且不导出 Secret 明文。
31. Callback Token 元数据备份说明已记录，且不导出 token 明文。
32. external_service 配置备份与 token_ref / 加密 key 恢复说明已记录。
33. safe / warning / blocked 风险等级处理说明已记录。
34. warning 必须 `confirm=true`，blocked 不能 confirm 绕过。
35. MySQLStore 生产模式验收项已记录。

本轮执行结果（2026-05-22）：

1. `git diff --check`：通过。
2. 关键词 grep：通过，备份 / rollback / failure_stage / dry-run / migration down / PluginRegistry reload / 加密 key / MemoryStore / MySQLStore 关键字均已在预期文档中出现。
3. 本轮仅更新文档，不改运行代码，不执行 Go / 前后台构建。
4. 2026-05-22 复核：`go test ./internal/service -run "TestPluginUpgradeDryRun|TestPluginUpgradeDryRunStructuredPlansAndConfirm|TestPluginUpgradeDryRunBlocksSameVersionEvenWithConfirm|TestPluginUpgradeDryRunDependencyDiffAndUpgradeBlocksNewRequiredDependency" -count=1`、`go test ./...`、`go build ./...` 与 `git diff --check` 均通过。

本轮复核结果（2026-05-27）：

1. `git diff --check`：通过。
2. 关键词 grep：通过，`backup|备份|rollback|回滚|failure_stage|failure_reason|dry-run|migration down|PluginRegistry reload|加密 key|MemoryStore|MySQLStore|blocked|warning|confirm` 均已在预期文档中出现。
3. 本轮仅更新文档，不改 Go、后台或前台源码，未执行 Go 测试、Go build、后台 quick、前台 quick 或 MySQL smoke；如需完整运行验收，按本节建议命令手动执行。

## v1.8.4-S6 浏览器点击矩阵补齐验收项

说明：S6 主要补齐后台插件治理浏览器点击矩阵，覆盖 5 个治理域、3 个官方插件、旧路由、Tab 保持、敏感信息脱敏和 1024 / 1366 截图回归；当前以 `scripts/check-admin-plugin-ia.sh` 和 `./scripts/check-frontend.sh --admin-only --quick` 为准。

建议命令：

```bash
./scripts/check-admin-plugin-ia.sh
./scripts/check-frontend.sh --admin-only --quick
```

验收项：

1. 插件总览页面可打开，且能搜索 `official_links`、`official_announcement`、`official_webhook_notify`。
2. 插件包治理、Webhook 治理、发布者与信任、运行记录 / 审计页面均可打开。
3. `official_links` 管理入口可打开，页面不白屏，中文空状态和敏感字段脱敏可见。
4. `official_announcement` 配置 / 前端挂载 / 预览入口可打开，preview iframe 仍为内置路由，token / secret 不回显。
5. `official_webhook_notify` 配置 / health check / 外部服务执行 / manual retry 入口可打开，token / Authorization / secret 不回显。
6. `official_webhook_notify` 在当前 dev 环境因缺少 external_service token 加密 key，浏览器矩阵自动回退到 `auth_type=none` 继续完成页面点击与脱敏检查，不把回退写成生产通过。
7. 旧路由能跳转到正确治理域并保持 Tab，刷新后 Tab 状态保持。
8. 浏览器前进 / 后退基于实际侧边栏点击保持正确。
9. 1366 / 1024 两档截图已保存到 `.devhub/screenshots/plugin-ia`，报告文件为 `.devhub/screenshots/plugin-ia/report.json`。
10. 不出现 `Cannot read properties`、`undefined`、`null`、`No Data` 等白屏或异常占位。
11. `./scripts/check-admin-plugin-ia.sh` 通过，`./scripts/check-frontend.sh --admin-only --quick` 通过。

## v1.8.4-S8 插件包本地仓库测试数据批量清理验收项

说明：S8 新增插件包治理 -> 本地仓库的测试包 / 未安装包批量清理能力。清理前必须 preview；执行会删除未安装本地仓库包的 promoted upload 记录和 `storage/plugins/packages/` 目录；installed / enabled / active task 包必须跳过。

建议命令：

```bash
go test ./internal/service -run 'TestCleanupPluginPackageRepository_TestPackagesPreviewExecuteAndSkipsInstalled|TestDeletePluginPackageRepositoryPackage' -count=1
go test ./internal/transport/httpapi -run TestAdminPluginPackageRepositoryCleanupRoutesExist -count=1
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
```

验收项：

1. cleanup preview 不删除数据。
2. cleanup preview 返回可清理测试包列表。
3. cleanup preview 返回 skipped installed 包。
4. cleanup preview 返回 skipped enabled 包。
5. `e2e_upload_*` 测试包可被识别。
6. `fixture_*` 测试包可被识别。
7. blocked 包可被清理。
8. invalid 包可被清理。
9. promoted 未安装包可被清理。
10. installed 包不能被清理。
11. enabled 插件当前包不能被清理。
12. 有 active task 的包不能被清理。
13. cleanup 删除数据库本地包记录。
14. cleanup 删除 storage 对应包文件。
15. storage 文件不存在时不整体失败。
16. 路径穿越被拒绝。
17. 清理动作写 admin_logs。
18. 后台清理按钮可用。
19. 清理确认弹窗显示数量和明细。
20. 清理后列表刷新。
21. 空结果显示中文空状态。
22. MemoryStore 清理行为正确。
23. MySQLStore 清理行为正确，如已执行。
24. 清理不影响 installed / enabled 插件。
25. 清理不影响 PluginRegistry 运行态。
26. 不执行第三方代码。
27. 不开放动态加载。
28. 不改变 Webhook Secret / Callback Token 安全模型。

本轮执行结果（2026-05-27）：

- Docker `gofmt`：通过；宿主机 `gofmt` 不存在，改用 Docker `/usr/local/go/bin/gofmt`。
- `go test ./internal/service -run TestCleanupPluginPackageRepository_TestPackagesPreviewExecuteAndSkipsInstalled -count=1`：通过。首次未限定唯一测试前缀时按真实规则识别并清理了工作区未跟踪的历史 e2e / fixture 本地仓库目录；确认未删除 Git 跟踪文件，随后改为 `s8cleanup_` 前缀隔离后通过。
- `go test ./internal/transport/httpapi -run TestAdminPluginPackageRepositoryCleanupRoutesExist -count=1`：通过，覆盖 `/cleanup/preview`、`/cleanup` 和 `/repository/cleanup` 路由。
- `go test ./...`：失败，失败项为 `internal/plugins/scaffold` 的既有 `TestGenerateRefusesExistingDirectoryWithoutForce`（期望 existing directory 在未 force 时失败）。本轮 cleanup 定向 service / httpapi 测试已通过。
- `go build ./...`：通过。
- `git diff --check`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260527-183536/`。
- MySQLStore 手工 smoke：未执行；如需验证 MySQLStore 下 cleanup preview / cleanup / admin_logs / storage 删除一致性，按 `./dev.sh start --mysql` 后构造 e2e / fixture 包记录执行。

## v1.8.4-S9 后台初始化插件包表单优化验收项

说明：S9 优化后台“创建插件模板 / 初始化插件包”表单和模板生成输入逻辑。后台、CLI 和 export 继续共用 `PluginTemplateGenerator`；模板只生成声明文件和 `migrations/001_init.sql`，不安装、不启用、不执行 SQL、不执行第三方代码。

建议命令：

```bash
gofmt -w internal/domain/plugin_package_template.go internal/plugins/package_scanner.go internal/plugins/scaffold/scaffold.go internal/service/plugin_package_template_service.go internal/service/plugin_package_template_test.go internal/transport/httpapi/plugin_package_handler.go internal/transport/httpapi/plugin_package_template_http_test.go internal/transport/httpapi/router.go
go test ./internal/plugins/scaffold -count=1
go test ./internal/service -run 'TestPluginPackageTemplate' -count=1
go test ./internal/transport/httpapi -run TestAdminPluginPackageTemplateHTTPPreviewGenerateExport -count=1
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
```

验收项：

1. 插件名称填写后自动生成 `plugin_code`。
2. `plugin_code` 可在高级设置中手动修改。
3. 手动修改后不会被插件名称再次覆盖。
4. `plugin_code` 支持小写字母、数字、下划线和中划线。
5. `plugin_code` 非法时被拒绝。
6. `plugin_code` 冲突时被拒绝。
7. 保留 code 被拒绝。
8. 插件类型可选择内容型插件。
9. 插件类型可选择外部服务型插件。
10. 插件类型可选择后台工具型插件。
11. 插件类型可选择前端挂载型插件。
12. 内容型插件显示“内容数据类型”。
13. 非内容型插件隐藏“内容数据类型”。
14. 内容显示名称根据插件名称自动生成。
15. 发布者 / 作者可选择 DevHub Team。
16. 发布者 / 作者可选择当前组织。
17. 发布者 / 作者可选择当前用户。
18. 发布者 / 作者可选择可信发布者，如存在。
19. 发布者 / 作者可选择自定义高级选项。
20. 不同插件类型动态显示字段。
21. 隐藏字段不会错误提交。
22. preview 不落盘。
23. preview 返回 `plugin_code`。
24. preview 返回 `content_type`，如适用。
25. preview 返回权限列表。
26. preview 返回菜单列表。
27. preview 返回 migrations 列表。
28. preview 返回文件树。
29. preview 返回 conflicts。
30. generate 写入 `storage/plugins/drafts/{plugin_code}/`。
31. generate 拒绝路径穿越。
32. export zip 可用。
33. 生成文件包含 `migrations/001_init.sql`。
34. 生成文件不包含 `001_schema.sql`。
35. 生成模板不包含 `.go`。
36. 生成模板不包含可执行 JS。
37. 生成模板不包含 `.wasm`。
38. 生成模板不包含二进制。
39. 生成模板不包含 package scripts。
40. 生成模板不执行 SQL。
41. 生成模板不执行第三方代码。
42. 生成模板不自动 install。
43. 生成模板不自动 enable。
44. CLI 模板生成能力不退化。
45. 后台和 CLI 共用 `PluginTemplateGenerator`。

说明：S9 的模板生成主路径主要走 Service + filesystem；MemoryStore / MySQLStore 差异仅体现在同一 Repository 接口读取已安装插件做冲突校验。若需要 MySQL 专项 smoke，可按 `./dev.sh start --mysql` 后用后台 preview / generate / export API 复测。

本轮执行结果（2026-05-27）：

- 宿主机 `gofmt` / `go`：不可用，均返回 `command not found`；本轮改用 Docker `golang:1.22-bookworm` 执行 Go 格式化、测试和构建。
- Docker `gofmt`：通过。
- `go test ./internal/plugins/scaffold -count=1`：通过，覆盖 CLI 共用生成器的基础模板生成、已存在目录策略和不生成根目录 `001_schema.sql`。
- `go test ./internal/service -run "TestPluginPackageTemplate" -count=1`：通过，覆盖 preview 不落盘、generate 写入 drafts、`migrations/001_init.sql`、前端挂载类型、保留 code conflict 和 export zip。
- `go test ./internal/transport/httpapi -run TestAdminPluginPackageTemplateHTTPPreviewGenerateExport -count=1`：通过，覆盖后台 preview / generate / export API。
- `go test ./...`：通过。
- `go build ./...`：通过。
- `git diff --check`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260527-191859/`。
- MySQLStore 专项 smoke：未执行。模板生成主路径不依赖 MySQL 专属实现；如需复核 MySQL 下已安装插件 / content_type 冲突查询，可按 `./dev.sh start --mysql` 后执行后台 preview / generate / export API smoke。

本轮补充结果（2026-05-27，中划线 code 规则复核）：

- 修正：后台前端、后端 `PluginTemplateGenerator` 和 manifest validator 统一允许 `plugin_code` 使用小写字母、数字、下划线和中划线；内容数据类型仍保持小写字母、数字、下划线规则，code 含中划线时默认 content_type 会转为下划线。
- 新增 `TestGenerateAllowsHyphenatedCode`：覆盖 CLI / 后台共用生成器可生成 `demo-links`，且只包含 `migrations/001_init.sql`，不生成根目录 `001_schema.sql`。
- 新增 `TestPluginPackageTemplateRejectsPathTraversalCode`：覆盖 `../evil` 这类路径穿越 code 会被拒绝。
- 宿主机 `gofmt` / `go`：仍不可用，返回 `command not found`；本轮继续使用 Docker `golang:1.22-bookworm`。
- Docker `gofmt`：通过。
- `go test ./internal/plugins/scaffold -count=1`：通过。
- `go test ./internal/service -run TestPluginPackageTemplate -count=1`：通过，覆盖 preview / generate / export、前端挂载类型、保留 code conflict 和路径穿越拒绝。
- `go test ./internal/transport/httpapi -run TestAdminPluginPackageTemplateHTTPPreviewGenerateExport -count=1`：通过，覆盖后台 preview / generate / export API。
- CLI 模板生成测试：`go run ./cmd/devhub plugin:new --code cli-hyphen-test --name "CLI Hyphen Test" --content_type cli_hyphen_item --content_name "CLI 内容" --output /tmp/devhub-cli-template-test` 通过，生成 `migrations/001_init.sql`。
- `go test ./...`：通过。
- `go build ./...`：通过。
- `git diff --check`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，最终日志目录 `.devhub/checks/20260527-215346/`；Vite 仍有既有 chunk size warning，不影响本次 quick 结果。
- MySQLStore 专项 smoke：未执行。模板生成主路径不依赖 MySQL 专属实现；如需复核 MySQL 下已安装插件 / content_type 冲突查询，可按 `./dev.sh start --mysql` 后执行后台 preview / generate / export API smoke。

后台配置编辑器补充结果（2026-05-27）：

- 修正 JSON 高级模式可能把实际配置对象作为字符串回传，导致 schema 校验出现 `$: must be object` 的问题。
- 现在 JSON 高级模式会先把字符串配置解析为对象，解析成功后再按 `config_schema` 校验和保存；解析失败或根节点不是对象时显示中文错误。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260527-222015/`；Vite 仍有既有 chunk size warning，不影响本次 quick 结果。

后台配置模型错误提示补充结果（2026-05-27）：

- 复核 `official_webhook_notify` 全局配置 schema：只允许 `enabled` 和 `receiver_note`；`endpoint_url`、`health_check_path`、`token` 等 external_service 字段应在插件详情的外部服务配置表单中维护。
- Ajv 校验错误已中文化：未声明字段显示“当前配置模型不允许字段 xxx”，根节点类型、缺少必填、枚举、格式和正则错误也会给出中文提示。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260527-222521/`；Vite 仍有既有 chunk size warning，不影响本次 quick 结果。

## v1.8.4-S20 upgrade dry-run 结构化差异与影响范围复核

说明：当前仓库已存在结构化 upgrade dry-run 与 upgrade confirm 规则；本轮重新复核并确认其仍覆盖版本计划、包摘要、变更摘要、影响范围、migrations / config / permissions / content_types / hooks / external_service / frontend_mounts / dependencies 分区计划、风险项、阻断项、确认要求和回滚边界。

已执行检查：

1. `docker run --rm -v "$PWD":/workspace -w /workspace golang:1.22-bookworm /bin/sh -lc 'export PATH=/usr/local/go/bin:$PATH; git config --global --add safe.directory /workspace; go test ./...'`：通过。
2. `docker run --rm -v "$PWD":/workspace -w /workspace golang:1.22-bookworm /bin/sh -lc 'export PATH=/usr/local/go/bin:$PATH; git config --global --add safe.directory /workspace; go build ./...'`：通过。
3. `git diff --check`：通过。
4. `POST /api/v1/admin/plugins/:code/upgrade/dry-run` 的现有测试覆盖仍通过，warning 需要 `confirm=true`，blocked 不能绕过 confirm。

复核结论：

- `from_version -> to_version`、`manifest`、`permissions`、`content_types`、`hooks`、`external_service`、`config_schema`、`menus`、`frontend_mounts`、`migrations`、`dependencies`、`impact_summary`、`risk_level`、`confirm_required`、`confirm_token`、`rollback_boundary` 仍在当前实现中。
- dry-run 仍不执行 SQL、不执行代码、不改变安装状态、不改变 PluginRegistry。
- 文档与实现保持一致，不需要额外改动即可满足当前 upgrade dry-run 收口要求。

2026-05-22 再复核补充：

1. `go test ./internal/service -run "TestPluginUpgradeDryRun|TestPluginUpgradeDryRunStructuredPlansAndConfirm|TestPluginUpgradeDryRunBlocksSameVersionEvenWithConfirm|TestPluginUpgradeDryRunDependencyDiffAndUpgradeBlocksNewRequiredDependency" -count=1`：通过。
2. `go test ./...`：通过。
3. `go build ./...`：通过。
4. `git diff --check`：通过。
5. 现有 upgrade dry-run 仍可返回结构化版本计划、包摘要、变更摘要、影响范围、迁移 / 配置 / 权限 / content_types / Hook / 前端挂载 / dependencies 分区、风险项、阻断项、确认要求和回滚边界；warning 仍需显式确认，blocked 仍不可绕过。

## v1.8.4-S1 official_links 官方友情链接插件生产化验收项

说明：S1 把 `official_links` 从声明型样板推进为生产可用官方友情链接插件。当前实现复用声明型 content_type、Core `topics` 内容模型、通用 `PluginContent` 后台治理和搜索页前台展示，不新增专属业务运行时，也不执行插件包代码。

建议命令：

```bash
go test ./internal/transport/httpapi -run TestAdminPostCreateSupportsDeclarativePluginContentType -count=1
go test ./internal/service -run TestDryRunPluginPackage_OfficialLinksExampleOK -count=1
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
./scripts/check-frontend.sh --frontend-only --quick
```

验收项：

1. official_links 插件包路径存在：`examples/plugins/official_links/`。
2. official_links 插件包包含 `manifest.json`、`README.md`、`config.example.json`、`checksums.json` 和 `migrations/001_init.sql`。
3. official_links 插件包可生成 / 打包，不包含 package scripts、可执行二进制、远程 iframe、远程 JS、inline_html 或 remote_component。
4. official_links 可 upload。
5. official_links 可 precheck。
6. official_links 可 promote。
7. official_links 可进入 local repository。
8. official_links 可 install dry-run。
9. official_links install dry-run 不执行 SQL。
10. official_links install 只基于 `migrations/`；当前 `migrations/001_init.sql` 是 no-op 计划文件。
11. official_links install 不执行 package scripts。
12. official_links install 成功后 PluginRegistry reload。
13. official_links 可 enable。
14. official_links 可 community enable。
15. official_links 后台菜单 enabled 后可见，入口为 `/admin-next/official-links`。
16. official_links 菜单 disabled / archived / soft_uninstalled 后隐藏。
17. official_links 权限矩阵可见：`official_links.menu.view`、`official_links.link.create`、`official_links.link.manage`、`official_links.config.manage`。
18. official_links 配置可读，配置项包括 `enabled`、`title`、`display_position`、`max_links`、`show_logo`、`open_in_new_tab`。
19. official_links 配置可保存，保存按 `config_schema` 校验。
20. official_links 配置保存写审计。
21. official_links 可通过通用 `PluginContent` / `admin/posts` 新增友情链接，写入 `topics.plugin_code=official_links`、`topics.content_type=friend_link`。
22. official_links 可编辑友情链接。
23. official_links 可禁用或删除友情链接。
24. official_links 可通过通用批量治理完成排序 / 置顶 / 推荐等排序类治理。
25. official_links 子站 enabled 后前台可通过 `/search/?scope=community&community_slug=:slug&content_type=friend_link` 展示。
26. official_links 子站 disabled 后前台插件入口不展示。
27. official_links 全局 disabled 后不能新建友情链接。
28. official_links archived 后阻断治理操作。
29. official_links soft_uninstalled 后不能新建 / 编辑。
30. official_links 历史友情链接仍可通过通用 Topic 详情读取。
31. official_links 历史审计仍可通过后台审计筛选查看。
32. official_links 权限不足时 API 拒绝或菜单隐藏，并显示中文提示。
33. official_links 不执行第三方代码。
34. official_links 不开放远程 iframe。
35. official_links 不开放 remote component。
36. official_links 不开放 blocking Hook。
37. official_links 不破坏 `/topics/:id` SEO、`/c/:slug` SEO 或 sites/posts 兼容 API。

本轮执行结果（2026-05-21）：

1. `gofmt -l`：通过，无输出。
2. `go test ./internal/transport/httpapi -run TestAdminPostCreateSupportsDeclarativePluginContentType -count=1`：通过，覆盖声明型 content_type 创建、搜索筛选和 disabled 阻断。
3. `go test ./internal/service -run TestDryRunPluginPackage_OfficialLinksExampleOK -count=1`：通过。
4. `go test ./...`：通过。
5. `go build ./...`：通过。
6. `./scripts/build-plugin-package-fixtures.sh --suffix official-links-s1`：通过，输出目录 `scripts/fixtures/plugin-packages/dist`。
7. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260521-170708/`。
8. `./scripts/check-frontend.sh --frontend-only --quick`：通过，日志目录 `.devhub/checks/20260521-170739/`。
9. `./dev.sh start --mysql`：通过，完整经过 MySQL ready、临时 API、前台构建、后台构建和最终 Go 服务启动。
10. MySQL official_links smoke：通过。真实 Admin API 覆盖 zip upload、precheck、promote、local repository、install dry-run、`confirm_risk_level=medium` 安装、enable、community enable、创建 `friend_link` 板块、创建友情链接、子站 disabled 阻断、全局 disabled 阻断和搜索展示；修复后 `/api/v1/search/topics?scope=community&community_slug=php&content_type=friend_link` 仅返回 `official_links/friend_link`，未混入 `article`。
11. `git diff --check`：通过。

环境说明：第一次 Go 定向测试因模块代理下载 `mimetype` 出现 EOF，未进入测试本体；使用 `GOPROXY=https://goproxy.cn,direct` 重试后通过。

## v1.8.4-S3 official_webhook_notify 官方 Webhook 通知样板生产化验收项

说明：S3 把 `official_webhook_notify` 作为官方 external_service Webhook 通知样板进行生产化复查，重点覆盖安装链路、health check、异步投递、失败重试、审计和脱敏边界。当前实现继续保持 external_service non-blocking 子集，不开放第三方运行模型。

建议命令：

```bash
go test ./internal/service -run "TestDryRunPluginPackage_OfficialWebhookNotifyExampleOK|TestExternalServiceHealthCheckWarningAndRecovery|TestExternalServiceValidationAndDisabledPluginSkipped|TestExternalServiceManualRetrySuccessAndForbiddenStates|TestExternalServiceManualRetryRejectsSkippedAndDisabledPlugin" -count=1
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
./scripts/check-frontend.sh --frontend-only --quick
```

验收项：

1. official_webhook_notify 插件包路径存在：`examples/plugins/official_webhook_notify/`。
2. official_webhook_notify 插件包包含 `manifest.json`、`README.md`、`config.example.json`、`checksums.json` 和 `migrations/README.md`。
3. official_webhook_notify 插件包可生成 / 打包，不包含 package scripts、可执行二进制、远程 iframe、远程 JS、inline_html 或 remote_component。
4. official_webhook_notify 可 upload。
5. official_webhook_notify 可 precheck。
6. official_webhook_notify 可 promote。
7. official_webhook_notify 可进入 local repository。
8. official_webhook_notify 可 install dry-run。
9. official_webhook_notify install dry-run 不执行 SQL。
10. official_webhook_notify install 只基于 `migrations/`；当前 `migrations/` 仅保留说明文件。
11. official_webhook_notify install 不执行 package scripts。
12. official_webhook_notify install 成功后 PluginRegistry reload。
13. official_webhook_notify 可 enable。
14. official_webhook_notify 可 community enable。
15. 后台可配置 `endpoint_url`。
16. 后台可配置 `token`，且 token 不回显。
17. `Authorization` 不回显。
18. `health_check_path` 可配置。
19. `timeout_ms` 可配置。
20. `failure_policy` 可配置。
21. 手动 health check 成功显示 healthy。
22. 手动 health check 失败显示 warning / error。
23. health check 失败原因可见，且长度受限。
24. `AfterCreateContent` 可触发 external_service 投递。
25. 主业务流程不等待远端响应。
26. mock receiver 能收到请求。
27. `hook_executions` 有 success 记录。
28. `hook_executions` 有 failed / timeout 记录。
29. `failed` / `timeout` / `retry_exhausted` 可手动 retry。
30. `success` 不允许 retry。
31. `skipped` 不允许 retry。
32. `internal` 不允许 retry。
33. `disabled` 插件不允许 retry。
34. `archived` 插件不允许 retry。
35. `community disabled` 不继续投递。
36. `external_service disabled` 不继续投递。
37. `disabled` / `archived` / `soft_uninstalled` 不继续投递。
38. `admin_logs` 有配置、health check、retry、投递状态变化审计。
39. token 不进响应明文、不进日志明文、不进审计明文、不进 `hook_executions` 明文。
40. Secret 不回显。
41. Callback Token 不回显。
42. 不执行第三方代码。
43. 不开放远程 iframe。
44. 不开放 remote component。
45. 不开放 blocking Hook。

## v1.8.4-S4 插件包打包 / 校验 CLI 优化验收项

说明：S4 为官方插件包和模板补齐本地 build / check CLI，让开发者先在仓库内生成 zip、再做 dry-run 校验，然后再进入 upload -> precheck -> promote -> install dry-run -> install 链路。CLI 只复用现有 package dry-run / manifest 校验逻辑，不执行 SQL、不执行插件代码、不执行 package scripts、不访问远程市场。

建议命令：

```bash
./scripts/plugin-package-build.sh --all-official --out .devhub/plugin-packages/s4
./scripts/plugin-package-check.sh examples/plugins/official_links
./scripts/plugin-package-check.sh .devhub/plugin-packages/s4/official_links-1.0.0.zip
./scripts/plugin-package-check.sh examples/plugins/official_webhook_notify
./scripts/plugin-package-check.sh examples/plugins/templates/declarative-content
./scripts/plugin-package-check.sh examples/plugins/templates/external-service-webhook
./scripts/plugin-package-check.sh --builtin official_announcement
```

验收项：

1. 插件包打包命令可执行。
2. official_links 包可生成。
3. official_webhook_notify 包可生成。
4. declarative-content 模板包可生成。
5. external-service-webhook 模板包可生成。
6. 插件包校验命令可执行。
7. manifest 缺失时 blocked。
8. manifest JSON 非法时 blocked。
9. plugin_code 非法时 blocked。
10. config_schema 编译失败时 blocked。
11. JSON Schema 正则不兼容时 blocked。
12. migrations/ 文件命名异常时 blocked 或 warning，按当前策略。
13. 根目录 `001_schema.sql` 只 warning。
14. package scripts blocked。
15. 可执行二进制 blocked。
16. `iframe_url` blocked。
17. `script_url` blocked。
18. `remote_entry` blocked。
19. `inline_html` blocked。
20. `remote_component` blocked。
21. `unknown mount_point` blocked。
22. `unknown component_key` blocked。
23. external_service endpoint 非法时 blocked。
24. external_service timeout_ms 越界时 blocked。
25. external_service failure_policy 非法时 blocked。
26. 校验不执行 SQL。
27. 校验不执行插件代码。
28. 校验不动态加载资产。
29. 校验不访问远程插件市场。
30. 校验输出 blockers / warnings / 建议修复动作。
31. blocked 返回非零退出码。
32. warnings 处理策略：返回 0，但建议修复后再上传生产环境。

本轮执行结果（2026-05-21）：

1. `./scripts/plugin-package-build.sh --all-official --out .devhub/plugin-packages/s4`：通过，生成 official_links / official_webhook_notify / declarative-content / external-service-webhook 的 zip 产物，并自动调用本地校验。
2. `./scripts/plugin-package-check.sh examples/plugins/official_links`：通过，状态 `warning`（缺少 publisher.json / signature.json）。
3. `./scripts/plugin-package-check.sh .devhub/plugin-packages/s4/official_links-1.0.0.zip`：通过，状态 `warning`。
4. `./scripts/plugin-package-check.sh examples/plugins/official_webhook_notify`：通过，状态 `warning`。
5. `./scripts/plugin-package-check.sh examples/plugins/templates/declarative-content`：通过，状态 `warning`。
6. `./scripts/plugin-package-check.sh examples/plugins/templates/external-service-webhook`：通过，状态 `warning`，并确认 `health_check_path` 配置 schema 与 external_service 声明可编译。
7. `./scripts/plugin-package-check.sh --builtin official_announcement`：通过，状态 `passed`，确认内置插件可校验其 config_schema 和官方 allowlist 前端挂载。
8. `./scripts/build-plugin-package-fixtures.sh --suffix s4`：通过，生成 blocked / deprecated / valid fixture 及官方链接 fixture。
9. `./scripts/plugin-package-check.sh scripts/fixtures/plugin-packages/dist/devhub-fixture-blocked-plugin-s4.zip`：通过触发 `blocked`，退出码为 `1`，输出 blockers / warnings / 建议修复动作。
10. `go test ./cmd/plugin-package-cli -run TestNonExistent -count=0`：通过，用于确认 CLI 可编译。
11. `go test ./internal/service -run TestConfigSchemaPatternsAreJSCompatible -count=1`：通过，确认公告插件与官方 Webhook 模板的配置 schema 可正常编译。
12. `go test ./...`：通过。
13. `go build ./...`：通过。
14. `git diff --check`：通过。

## v1.8.3 插件系统发布前总验收记录

执行日期：2026-05-21。

本轮验收覆盖 official_links 声明型插件完整链路、official_webhook_notify external_service Webhook 链路、插件升级 dry-run / confirm / blocked / warning / failure_stage、声明型插件开发者指南与两个官方模板、frontend_mount 官方 allowlist、MemoryStore / MySQLStore 双模式、后台插件治理页面稳定性和文档一致性。

执行结果：

1. `gofmt -l`：通过，无未格式化 Go 文件。
2. `go test ./...`：通过，使用 Docker `golang:1.22-bookworm`。
3. `go build ./...`：通过，使用 Docker `golang:1.22-bookworm`。
4. 插件专项定向测试：通过，覆盖 package dry-run、official_webhook_notify、external_service health / disabled gating、manual retry 相关链路、upgrade、frontend_mount allowlist 和公共插件 API 敏感字段剥离。
5. `./scripts/build-plugin-package-fixtures.sh --suffix total-acceptance`：通过，可生成声明型插件包 fixture。
6. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260521-140147/`。
7. `./scripts/check-frontend.sh --frontend-only --quick`：通过，日志目录 `.devhub/checks/20260521-140207/`。
8. `./scripts/check-admin-plugin-ia.sh`：通过，截图目录 `.devhub/screenshots/plugin-ia`。
9. `./dev.sh start --mysql`：通过；完整经过 MySQL ready、临时 API、前台构建、后台构建和最终 Go 服务启动。
10. MySQL smoke：通过；`/api/v1/health` 返回 `store=mysql`，后台登录、插件列表、审计列表可访问，公共插件列表未暴露 `config_json`、`resolved_config`、`frontend_mounts`。
11. `git diff --check`：通过。

未完成 / 环境限制：

- 额外尝试执行 MySQLStore 可选集成单测 `DEVHUB_MYSQL_TESTS=1 ... go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v` 两次，均在测试 setup 前因容器下载 `github.com/go-sql-driver/mysql@v1.8.1` 发生 `TLS handshake timeout` 失败，未进入测试用例本体；该项不记录为通过。MySQL 模式应用启动和 smoke 已按上方记录通过。

验收中修复：

- MySQLStore 新装 schema 中 `plugin_webhook_secrets` 与 `webhook_circuit_breakers` 的 `(plugin_code, target_url)` 相关索引在 utf8mb4 下超出 MySQL 3072 byte 限制；已将两个被索引的 `target_url` 字段从 `VARCHAR(1000)` 收敛为 `VARCHAR(512)`，并同步 `db/mysql/001_schema.sql` 与 `internal/store/schema.go`。Webhook Secret 创建同步增加 `target_url` 512 字符上限校验；未参与索引的 `webhook_deliveries.target_url` 保持原长度。

验收结论：

1. official_links 链路由真实声明型 fixture、package dry-run / install / enable / community enable / content_type / 权限矩阵 / 配置 / disabled / archived 阻断用例覆盖；dry-run 不执行 SQL，install 只基于 `migrations/`。
2. official_webhook_notify 链路由官方样板包 dry-run、external_service 配置、health check、hook_executions、manual retry 和 token 不回显规则覆盖；投递失败不阻塞主业务，blocking Hook 未开放。
3. upgrade 链路继续要求 warning 显式确认，blocked 不可绕过；dry-run 不执行 SQL、不改变安装状态、不刷新 registry，失败阶段和下一步建议可见。
4. frontend_mount 只支持官方 allowlist；未知挂载点 / 组件、远程 iframe / JS、inline HTML、remote component、eval 和可执行 JS 入口资产会被阻断，运行时过滤 disabled / archived / 子站 disabled 插件并过滤敏感 props。
5. MemoryStore 与 MySQLStore 在本轮关键插件能力 smoke 和自动化测试范围内保持一致；MySQLStore 已验证 schema 初始化、插件列表、审计和公共插件 API 敏感字段剥离。
6. 后台治理页面 quick build 与 S18 插件 IA 浏览器回归均通过；旧路由 / 入口 / 按钮矩阵继续由 `.devhub/screenshots/plugin-ia` 记录。
7. 安全边界保持：不执行第三方代码、不开放 Go plugin / JS 沙箱 / 远程 iframe / remote component / 插件市场 / 远程在线安装 / blocking Hook；token / secret / Authorization 不回显，不进日志或审计明文。

发布建议：`v1.8.3` 插件系统可进入冻结候选；冻结前仍建议按发布流程人工复核真实浏览器点击矩阵、生产 MySQL 大库备份 / 回滚演练和部署环境密钥配置。

## 后台插件中心中文状态和异常提示统一验收项

说明：本轮只覆盖后台插件模块中文状态、按钮、错误提示、风险提示和异常原因收口，不引入完整 i18n 框架，不改变英文枚举真实值或 API `code` 字段。

建议命令：

```bash
go test ./...
go build -o .devhub/devhub .
bash -n dev.sh
./scripts/check-frontend.sh --admin-only --quick
# 插件治理相关 E2E：按当前环境执行 web/admin-app 中 plugin*.spec.js 或 check-frontend 的 admin E2E 子集。
```

验收项：

1. 插件列表、详情抽屉、配置中心、内容治理、权限矩阵、运行记录、Webhook 治理中不裸露 `enabled/disabled/archived` 等核心英文状态。
2. 插件包 upload / scan / promote / install / dry-run 的风险、阻断码和错误原因显示中文。
3. `risk_blocked/manifest_invalid/checksum_failed/dangerous_file/signature_invalid/publisher_unknown/core_incompatible` 等插件包原因可显示中文。
4. 健康状态、readiness、Hook、外部服务和依赖原因显示中文。
5. 审批、归档、禁用、恢复、安装、promote、重新扫描、健康检查等操作按钮和确认提示使用中文并说明影响。
6. 后端插件接口保留 `code`，同时继续返回或可映射中文 `message/suggestion`。
7. 前端 HTTP 错误提示优先显示中文 message；仅有 code 时使用插件映射兜底，并保留“错误码”用于排障。
8. 同一状态在列表、详情、toast、审计、健康页中展示口径一致。
9. 技术字段名如 `plugin_code`、`publisher_id`、`key_id`、`token_ref`、`secret_ref`、`manifest` 可保留，但用户操作文案必须中文化。
10. 不展示 token 明文、Secret 明文、Authorization Header、完整 HMAC signature 或敏感配置。
11. 不执行第三方代码，不开放动态加载，不开放远程 iframe，不做 blocking Hook。
12. 现有 E2E 不因文案调整退化；测试断言不依赖脆弱英文枚举文案。

## v1.8.3：后台插件治理页面稳定性、信息架构与中文体验验收项

说明：v1.8.3 只优化后台页面稳定性、结构与中文体验，不修改插件生命周期、Webhook 协议、Secret / Token 安全模型、Host + iframe + postMessage 协议。本轮按用户要求，后台改动完成后执行后台 quick 检查。

建议命令：

```bash
./scripts/check-frontend.sh --admin-only --quick
```

1. 插件管理页面不白屏。
2. Webhook 治理页不白屏。
3. 插件详情抽屉打开不报运行时异常。
4. 插件详情切换不同插件不报错。
5. 空数据场景显示中文空状态。
6. 左侧插件导航收敛为 5 个治理域，或页面内等效分组。
7. 插件列表页信息分组清晰。
8. 插件详情页使用 Tab 分组。
9. Webhook 治理能力不再全部堆在插件详情首页。
10. 插件包治理有独立分组或 Tab。
11. Secret / Callback Token 有独立分组或 Tab。
12. official_announcement 配置和预览入口清晰。
13. 主要按钮使用中文。
14. 状态 badge 使用中文。
15. 空状态使用中文说明。
16. 错误提示使用中文。
17. 危险操作有中文确认提示。
18. Secret 明文不显示在表格。
19. Callback Token 明文不显示在表格。
20. 远程 iframe URL 不开放。
21. 第三方代码执行仍未开放。
22. 插件市场仍未开放。
23. blocking Hook 仍未开放。
24. 后台构建通过。
25. S18 的按钮行为与旧路由回归已作为常规守护项，`/plugins/*` 旧入口不得白屏、跳错或丢 Tab。
26. 插件详情、Webhook、插件包治理的低频 Tab 或旧路由直达后，页面标题、面包屑和左侧高亮仍需正确。

### v1.8.3-S16 后台插件治理复杂度继续压缩验收项

建议命令：

```bash
./scripts/check-admin-plugin-ia.sh
./scripts/check-frontend.sh --admin-only --quick
```

1. 左侧仍只保留 5 个插件治理域。
2. 插件总览默认视图不过载，只突出概览、插件列表、待处理事项和快捷操作。
3. 插件包治理呈现为流程型工作台，而不是多个平级 Tab 堆叠。
4. 插件包治理能清楚提示执行预检、转入本地仓库、执行安装 dry-run、开始安装和查看阻断原因。
5. Webhook 治理默认聚焦总览、投递记录、异常处理、重试和熔断。
6. Webhook 密钥、回调 Token、回调请求和原始 JSON 进入高级治理或技术详情。
7. 发布者与信任默认弱化 `key_id` / `publisher_id` 等技术字段。
8. 运行记录 / 审计只作为排障和追踪入口，不承担日常治理入口。
9. 插件详情抽屉默认 Tab 数减少，默认只展示当前插件摘要。
10. 插件详情里的 Webhook、权限、content_type、external_service 只展示摘要和跳转入口。
11. 完整治理操作跳转到对应治理域。
12. 技术详情默认折叠。
13. JSON 原始数据默认折叠。
14. 每个异常状态都有下一步操作提示。
15. 1366 宽度下 5 个治理域布局自然。
16. 1024 宽度下 5 个治理域可用。
17. 详情抽屉 1024 宽度下可用。
18. `official_announcement` 详情仍清晰，配置、挂载、预览入口可识别。
19. 真实声明型 fixture 插件详情可用，如当前环境已有 fixture。
20. Secret 明文不显示。
21. Callback Token 明文不显示。
22. external_service token 明文不显示。
23. token_hash、Authorization Header、完整 HMAC signature 不显示。
24. 远程 iframe URL 仍未开放。
25. 第三方代码执行仍未开放。
26. blocking Hook 仍未开放。
27. 后台 quick 检查通过。

### v1.8.3-S18 插件后台按钮行为与旧路由跳转回归验收项

说明：S18 只修按钮行为、旧路由兼容、页面跳转和 Tab 定位；不改变 API、插件底层逻辑、Webhook 协议、Secret / Token 安全模型，不开放远程 iframe、第三方代码执行或 blocking Hook。

长期守护口径：插件后台不再只按页面名记入口，而是按任务划分去向。管理员先决定“要安装 / 启用 / 排障 / 看审计 / 管密钥 / 看原始声明”，再进入对应治理域、页内 Tab 或详情抽屉。

建议命令：

```bash
./scripts/check-admin-plugin-ia.sh
./scripts/check-frontend.sh --admin-only --quick
```

本轮执行结果：

- `./scripts/check-admin-plugin-ia.sh`：通过，覆盖 5 个治理域、主要按钮点击、详情抽屉配置 / 能力 / 运行记录 / 技术详情、旧路由到新治理域 / Tab、标题 / 面包屑 / 页面不白屏；截图目录 `.devhub/screenshots/plugin-ia`。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260520-202752/`。
- S18 追加复查：技术详情最小分组、启用检查可见入口、迁移重试 import、导出入口和原始声明折叠已纳入本节常规验收；`./scripts/check-admin-plugin-ia.sh` 单独重跑通过，截图目录 `.devhub/screenshots/plugin-ia`。
- S20 / upgrade dry-run 复查：结构化差异、影响范围、风险等级、warning 确认和 blocked 阻断规则仍按当前实现生效；`go test ./internal/service -run 'TestBuildPluginManifestDiffFrontendMounts|TestFrontendMountsForRuntimeAllowlistAndStatus|TestValidatePluginManifestJSONFrontendMountAllowlist|TestUpgrade' -count=1`、`go test ./...`、`go build ./...`、`git diff --check`、`./scripts/check-frontend.sh --admin-only --quick` 均通过，前台 quick 额外生成站点和 topic 相关静态页面，未改 `/topics/:id` SEO 动态 HTML，也未删除 sites/posts 兼容 API；日志目录 `.devhub/checks/20260521-122006/`。

验收项：

1. 插件总览主要按钮可点击。
2. 插件列表查看详情按钮可打开正确插件。
3. 插件详情配置 / 能力 / 运行 / 技术详情 Tab 可切换。
4. 插件包治理主要按钮可点击。
5. blocked 包 promote 按钮禁用且后端拒绝。
6. 本地仓库包安装 dry-run 入口可用。
7. Webhook 治理主要按钮可点击。
8. Webhook 投递记录入口跳转正确。
9. Webhook 密钥入口跳转正确。
10. 回调 Token 入口跳转正确。
11. 发布者与信任主要按钮可点击。
12. 运行记录 / 审计主要按钮可点击。
13. 旧路由能跳转到正确治理域。
14. 旧路由能定位到正确 Tab。
15. 刷新后 Tab 状态保持。
16. 浏览器前进 / 后退 Tab 状态正常。
17. 左侧导航高亮正确。
18. 页面标题不为空、不重复。
19. 面包屑不为空、不重复。
20. 未实现功能显示中文占位，不白屏。
21. 按钮点击不出现 Cannot read properties。
22. 按钮点击不出现 undefined / null。
23. 远程 iframe URL 仍未开放。
24. 第三方代码执行仍未开放。
25. blocking Hook 仍未开放。
26. admin quick 检查通过。
27. 插件详情的技术详情里能直接看到迁移明细和刷新 / 重试入口。
28. 迁移明细默认不再依赖隐藏旧 Tab 才能查看。
29. 插件详情的技术详情按“启用检查 / 迁移明细 / 导出本地插件包 / 原始声明 JSON”最小分组展示。
30. 启用检查入口不再依赖隐藏旧 readiness Tab，可刷新检查并运行启用预检；真正启用仍走后端校验。
31. 导出本地插件包入口在技术详情中可见，导出范围仍仅包含 manifest、README、config.example.json、checksums，不包含敏感配置、用户数据、运行时代码或外部 SQL。
32. 原始声明 / JSON 默认折叠，config_schema、resolved_config、content_type、权限、前端挂载、Webhook / Hook 和运行健康原始摘要继续脱敏展示。
33. 插件后台入口应能用“任务 -> 入口”描述清楚，不再只靠页面名猜位置。
34. 日常状态查看和单插件处理应落到插件总览 / 插件列表 / 详情抽屉。
35. 安装、预检、promote、dry-run、安装计划应落到插件包治理。
36. Webhook 投递、异常、密钥、回调 Token、回调请求应落到 Webhook 治理。
37. 发布者、公钥、可信性、影响分析应落到发布者与信任。
38. 操作历史、审计、Hook 排障、最近错误应落到运行记录 / 审计。
39. 原始声明、配置模型、迁移明细、导出包结构应落到插件详情技术详情。
40. 插件列表主按钮与行内动作应统一为任务语言，例如“看详情 / 处理配置 / 更多任务 / 去审计 / 去启用 / 去停用 / 做启用检查 / 去软卸载”。
41. 插件列表顶部安装相关按钮应统一为任务语言，例如“校验清单 / 查看预检 / 去安装”。

### v1.8.3-S22 前端插件挂载继续收口验收项

说明：S22 继续收口前端插件挂载，只开放官方 allowlist 挂载点和官方组件 key；不开放任意远程 iframe、远程 JS、远程 CSS、inline HTML 或第三方前端运行时。不改变 API、插件逻辑、Webhook 协议或 Secret / Token 安全模型。

建议命令：

```bash
go test ./internal/plugins ./internal/service -run 'TestValidatePluginManifestJSONFrontendMountAllowlist|TestFrontendMountsForRuntimeAllowlistAndStatus|TestBuildPluginManifestDiffFrontendMounts' -count=1
go test ./...
go build ./...
git diff --check
./scripts/check-frontend.sh --admin-only --quick
./scripts/check-frontend.sh --frontend-only --quick
```

本轮执行结果：`gofmt` 已执行；定向前端挂载单测、`go test ./...`、`go build ./...`、`git diff --check`、后台 quick 和前台 quick 均通过。前后台 quick 日志目录为 `.devhub/checks/20260521-012440/`。

验收项：

1. manifest 可声明 `frontend_mounts`，但只允许官方 allowlist 内的挂载点和组件 key。
2. 未知挂载点会 blocked。
3. 未知组件 key 会 blocked。
4. `iframe_url` / `script_url` / `remote_entry` / `remote_url` / `inline_html` 会 blocked。
5. secret / token 不进入前端 props。
6. 已启用插件的合法挂载可在运行时返回。
7. disabled 插件不返回挂载。
8. archived 插件不返回挂载。
9. 子站未启用不返回挂载。
10. 历史未知组件 key 会被跳过，不白屏。
11. 后台插件详情能看到前端挂载状态、挂载点、组件 key 和说明。
12. 预检失败能展示中文错误码和说明。
13. upgrade dry-run 能展示 frontend_mounts 差异。
14. 不会渲染插件声明的远程 iframe；官方内置同源 Host / helper 仍只能服务 allowlist 组件。
15. 不会执行插件传入 HTML / JS。
16. 页面标题、面包屑和左侧高亮保持正确。
17. admin quick 检查通过。
18. frontend quick 检查通过，前台 static build 仍生成站点和 topic 相关页面。
19. 运行时只返回已安装、已启用 / running、未归档、未软卸载的插件挂载。
20. 子站页面只返回当前子站 enabled 的插件挂载。
21. 历史未知 component_key 会被跳过并返回 warning，不导致页面白屏。
22. secret / token / authorization / password / credential 类 props 不会传给前端组件。
23. 官方 helper 只创建内置同源 iframe，不读取插件声明的远程 iframe、远程脚本或 raw HTML。

### v1.8.3-S20 插件升级 dry-run / confirm / 失败边界验收项

建议命令：

```bash
./scripts/check-frontend.sh --admin-only --quick
# 如需后端回归，再执行 go test ./... 和 go build ./...（按当前环境使用 Docker / dev.sh）
```

本轮执行结果：

- `gofmt`：已通过 Docker `golang:1.22-bookworm` 执行。
- `go test ./...`：通过。
- `go test ./internal/transport/httpapi -count=1`：通过（审计事件补充后复查）。
- `go test ./internal/transport/httpapi -run TestAdminPluginPackageRepositoryCleanupRoutesExist -count=1`：通过（显式 repository cleanup 路由补充后复查）。
- `go build ./...`：通过。
- `git diff --check`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260521-000711/`。

1. 插件升级入口不白屏。
2. upgrade dry-run 可展示结构化差异，不再只靠原始 JSON。
3. upgrade dry-run 能展示版本计划、变更摘要、影响范围和回滚边界。
4. warning 升级需要勾选确认，按钮不能直接放行。
5. blocked 升级禁用确认按钮或直接拦截，不允许 confirm 绕过。
6. 升级结果能展示 `from_version -> to_version`。
7. 升级失败能展示 `failure_stage` / `failure_reason` / 审计建议。
8. 原始 JSON 保留在技术详情折叠区。
9. secret / token / Authorization / Webhook Secret 不回显。
10. 旧路由、当前治理域入口和插件详情入口都能进入升级功能。
11. 页面标题、面包屑、左侧高亮和 Tab 定位保持正确。
12. 未实现的回滚能力有中文边界说明，不白屏。
13. 升级 API 不会执行第三方代码、不开放远程 iframe、不开放 blocking Hook。
14. upgrade confirm 只接受显式确认，不接受无确认的 warning 升级。
15. blocked upgrade 即使 `confirm=true` 也会拒绝。
16. upgrade dry-run 不执行 SQL，不执行 migration down，不改变插件记录。
17. admin quick 检查通过。
18. 2026-05-21 复查：warning 仍需显式确认，blocked 仍不可绕过；dry-run plan 过期或与当前包 checksum / migration plan 不一致时继续拒绝升级；升级失败仍返回 `failure_stage` / `failure_reason`，后台结果页保留下一步建议。

### v1.8.3-S21 声明型插件开发者指南与官方模板验收项

说明：S21 只固化开发者指南和两个官方插件模板，不开放第三方代码执行、Go plugin、JS 沙箱、远程 iframe、插件市场、远程在线安装或 blocking Hook。

建议命令：

```bash
go test ./internal/service -run TestDryRunPluginPackage_OfficialTemplatesOK -count=1
go test ./...
go build ./...
git diff --check
```

验收项：

1. [声明型插件开发者指南](PLUGIN_DEVELOPER_GUIDE.md) 存在，并说明当前能做什么、不能做什么。
2. 纯声明型内容插件模板存在：`examples/plugins/templates/declarative-content/`。
3. external_service Webhook 插件模板存在：`examples/plugins/templates/external-service-webhook/`。
4. 两个模板都包含 `manifest.json`、`README.md`、`config.example.json`、`migrations/` 和打包说明。
5. Webhook 模板包含 `receiver.example.md`。
6. 纯声明型模板可声明 content_type、permissions、menus、config_schema 和 migrations/。
7. Webhook 模板可声明 `service_type=external_service`、`mode=non_blocking`、Hook path、method、timeout、retry 和 failure_policy。
8. 两个模板的 manifest 能通过当前 package dry-run。
9. 两个模板不包含真实 secret / token。
10. 两个模板不包含可执行代码。
11. 两个模板不包含远程 iframe URL。
12. 两个模板不声明 blocking Hook。
13. 两个模板不包含根目录 SQL；`migrations/` 仍是唯一迁移入口。
14. 模板说明文件 `PACKAGING.md`、`package.example.md`、`receiver.example.md` 被包扫描器当作允许文档，不触发危险文件阻断。
15. 文档中的命令、路径和接口名称与当前代码一致。
16. 本轮不改变 API、不改变 Webhook 协议、不改变 Secret / Token 安全模型。
17. 两个模板是官方推荐起点：先复制模板，再改 manifest、README、`config.example.json` 和 `migrations/`，不塞 package scripts、远程 iframe 或可执行资产。
18. `frontend_mounts` 相关预检 / install dry-run 会阻断未知 mount_point、未知 component_key、unsupported render_mode、`iframe_url`、`script_url`、`remote_entry`、`external_js`、`inline_html`、`remote_component`、`eval` 和未白名单的可执行 JS 资产。

插件后台路由兼容表：

| 旧入口 | 新治理域 / Tab |
| --- | --- |
| `/plugins/list` | `/plugins/overview?tab=list` |
| `/plugins/content` | `/plugins/overview?tab=content` |
| `/plugins/config` | `/plugins/overview?tab=config` |
| `/plugins/navigation` | `/plugins/overview?tab=navigation` |
| `/plugins/permissions` | `/plugins/overview?tab=permissions` |
| `/plugins/developer` | `/plugins/overview?tab=developer` |
| `/plugins/install`、`/plugins/packages/local`、`/plugins/packages/install`、`/plugins/packages/export`、`/plugins/manifest` | `/plugins/packages?tab=install` |
| `/plugins/packages/uploads`、`/plugins/package-uploads` | `/plugins/packages?tab=uploads` |
| `/plugins/packages/remote`、`/plugins/remote-packages` | `/plugins/packages?tab=remote-packages` |
| `/plugins/remote-indexes` | `/plugins/packages?tab=remote-indexes` |
| `/plugins/versions`、`/plugins/upgrade-diff` | `/plugins/packages?tab=versions` |
| `/plugins/dependencies` | `/plugins/packages?tab=dependencies` |
| `/plugins/approvals` | `/plugins/packages?tab=approvals` |
| `/plugins/events`、`/plugins/webhook-events` | `/plugins/webhooks?tab=events` |
| `/plugins/webhook-deliveries` | `/plugins/webhooks?tab=deliveries` |
| `/plugins/webhook-retry`、`/plugins/webhook-circuits` | `/plugins/webhooks?tab=exceptions` |
| `/plugins/webhook-secrets` | `/plugins/webhooks?tab=secrets` |
| `/plugins/callback-tokens` | `/plugins/webhooks?tab=callback_tokens` |
| `/plugins/callback-requests` | `/plugins/webhooks?tab=callback_requests` |
| `/plugins/trusted-publishers` | `/plugins/publishers?tab=list` |
| `/plugins/config-keys`、`/plugins/security` | `/plugins/publishers?tab=config-keys` |
| `/plugins/operations` | `/plugins/runtime?tab=operations` |
| `/plugins/audit` | `/plugins/runtime?tab=audit` |
| `/plugins/hooks`、`/plugins/diagnostics` | `/plugins/runtime?tab=hooks` |
| `/plugins/search-index` | `/plugins/runtime?tab=search-index` |

低频入口收纳关系：

- 插件详情 Webhook、Webhook 密钥、回调 Token、前端挂载入口收纳到“能力”摘要，并提供 Webhook 治理跳转。
- 插件详情审计和 Hook 排障入口收纳到“运行记录”，并提供运行记录 / 审计治理跳转。
- 插件详情 readiness 不再只依赖隐藏旧 Tab：启用检查已恢复到“技术详情”的可见分组；dependencies、content_type、permissions、routes 收纳到“原始声明 / JSON”，migrations 收纳到“迁移明细”，并显示中文说明。
- Webhook 治理事件、密钥、回调 Token、回调请求属于高级治理或旧路由直达 Tab；投递记录和异常处理仍是主入口。
- 插件包治理远程索引、远程包、依赖兼容、审批属于低频 Tab；旧路由直达时该 Tab 会可见。

### v1.8.3-S9 PluginRegistry reload 运行态刷新验收项

1. 插件安装成功后 registry reload。
2. 插件安装失败或回滚后 registry 不污染。
3. 插件升级成功后 registry reload。
4. 插件升级失败或回滚后 registry 保持旧状态。
5. 插件启用后 IsPluginEnabled 立即生效。
6. 插件停用后 IsPluginEnabled 立即失效。
7. 插件软卸载 / 归档后前端挂载停止。
8. 插件软卸载 / 归档后 Webhook 投递停止。
9. 子站插件启用后 IsPluginEnabledForCommunity 立即生效。
10. 子站插件停用后 IsPluginEnabledForCommunity 立即失效。
11. 插件全局配置保存后 ResolvedConfig 更新。
12. 插件子站配置保存后 ResolvedConfig 更新。
13. 配置保存失败不刷新脏状态。
14. reload 失败保留旧 registry 快照。
15. reload 失败有日志 / 审计。
16. disabled 插件 callback token 不可用。
17. disabled 插件不前端挂载。
18. disabled 插件不 Webhook 投递。
19. official_announcement 配置修改后 Host context 能读到新配置。
20. `/c/:slug` community disabled 后公告不显示。
21. reload 不执行第三方代码。
22. reload 不开放动态加载。
23. reload 不改变 Webhook 协议。
24. reload 不改变 Secret / Token 安全模型。

### v1.8.3-S10 声明型插件可用闭环验收项

后端新增覆盖：`internal/service/plugin_manifest_test.go` 的 `TestDeclarativePluginManifestCapabilitiesClosedLoop`。

1. 插件声明菜单能被 Registry / 运行态插件列表识别。
2. 插件 enabled 后菜单可见。
3. 插件 disabled 后菜单隐藏或不可用。
4. 插件 archived / soft_uninstalled 后菜单隐藏或不可用。
5. 插件声明 content_type 能被 Registry / Service 识别。
6. 插件 enabled 后 content_type 可用于创建内容。
7. 插件 disabled 后 content_type 不能新建内容。
8. 插件 archived 后 content_type 不能新建内容。
9. 板块不允许的 content_type 不能发布。
10. 权限不足不能发布插件 content_type。
11. 插件声明权限能进入权限矩阵。
12. 插件权限按 plugin_code 分组展示。
13. 权限不足时菜单隐藏或 API 拒绝。
14. 子站 enabled 后插件能力可用。
15. 子站 disabled 后插件能力不可用。
16. 全局 disabled 时子站 enabled 也不可用。
17. disabled 插件不前端挂载。
18. disabled 插件不 Webhook 投递。
19. disabled 插件 callback token 不可用。
20. archived 插件不前端挂载。
21. archived 插件不 Webhook 投递。
22. archived 插件 callback token 不可用。
23. 历史内容仍可查看。
24. 历史审计仍可查看。
25. Registry reload 后声明能力状态立即生效。
26. 不执行第三方代码。
27. 不开放动态加载。
28. 不改变 Webhook 协议。
29. 不改变 Secret / Token 安全模型。

### v1.8.3-S11 external_service Webhook 运行时预备验收项

后端新增覆盖：`internal/service/plugin_external_service_test.go` 的 `TestExternalServiceHealthCheckWarningAndRecovery`、`TestExternalServiceValidationAndDisabledPluginSkipped`。

2026-05-21 复查：后台表单入口位于插件详情抽屉“运行记录 / 外部服务配置”，Webhook 治理页“外部服务执行”用于查看投递记录和失败重试；已执行 `go test ./internal/service -run "TestExternalServiceHealthCheckWarningAndRecovery|TestExternalServiceValidationAndDisabledPluginSkipped" -count=1` 通过。

1. external_service endpoint 可以配置。
2. external_service timeout_ms 可以配置。
3. external_service failure_policy 可以配置。
4. external_service auth_type=none 可用。
5. external_service auth_type=bearer 可配置但 token 不回显。
6. 非法 endpoint 被拒绝。
7. `javascript:` / `data:` / `file:` / `ftp:` endpoint 被拒绝。
8. health check 成功后状态 healthy。
9. health check 失败后 failure_count 增加。
10. 达到 warning_threshold 后状态 warning。
11. 达到 error_threshold 后状态 error。
12. 成功恢复后状态 healthy。
13. hook_execution 成功记录可查看。
14. hook_execution 失败记录可查看。
15. hook_execution 不记录 token 明文。
16. hook_execution 不记录 Authorization Header。
17. disabled 插件 external_service skipped。
18. archived 插件 external_service skipped。
19. 子站 disabled 时 external_service 不执行子站 Hook（本轮 health check 层不绕过插件/子站 gating；完整子站 Hook 触发仍待后续远程 Hook 投递实现）。
20. 后台插件详情显示 health warning / error。
21. external_service 不执行第三方代码。
22. external_service 不开放动态加载。
23. external_service 不改变 Webhook 协议。
24. external_service 不改变 Secret / Callback Token 安全模型。
25. blocking Hook 仍未开放。

### v1.8.3-S14 external_service non-blocking Webhook 投递闭环验收项

后端新增覆盖：`internal/service/plugin_external_service_test.go` 的 `TestExternalServiceNonBlockingHookDeliverySuccess`、`TestExternalServiceNonBlockingHookRetryWarningAndSkipped`、`TestExternalServiceManifestRejectsBlockingHook`。

1. external_service Hook 可以声明 `non_blocking`。
2. external_service Hook 不支持 `blocking`，manifest 校验会拒绝。
3. 触发 Hook 时主业务流程不等待远端响应。
4. Hook 触发后创建 `hook_execution`。
5. `hook_execution pending -> running -> success` 可记录。
6. 2xx 返回 `success`。
7. timeout 返回 `timeout`。
8. 5xx 进入 `retry_scheduled`。
9. 429 按策略进入 `retry_scheduled`。
10. 4xx 默认不重试。
11. 401 / 403 默认不重试。
12. 超过 `max_attempts` 后 `retry_exhausted`。
13. `failure_policy=warn` 达阈值后 `health=warning`。
14. `failure_policy=error` 达阈值后 `health=error`。
15. 成功后 health 可恢复 `healthy`。
16. 插件 disabled 时 `execution=skipped`。
17. 插件 archived 时 `execution=skipped`。
18. external_service disabled 时 `execution=skipped`。
19. 子站 disabled 时 `execution=skipped`。
20. `auth_type=none` 可投递。
21. `auth_type=bearer` 可投递但 token 不记录。
22. `hook_execution` 不记录 token 明文。
23. `hook_execution` 不记录 Authorization Header。
24. 后台能通过 Hook 执行记录查看 `external_service` 投递结果。
25. 后台能查看 health warning / error。
26. 投递失败不导致后台白屏。
27. 不执行第三方代码。
28. 不开放动态加载。
29. blocking Hook 仍未开放。
30. 不改变 Secret / Callback Token 安全模型。

### v1.8.3-S19 external_service Webhook 可交付闭环验收项

官方样板包：`examples/plugins/official_webhook_notify`。

后端新增覆盖：`internal/service/plugin_external_service_test.go` 的 `TestExternalServiceManualRetrySuccessAndForbiddenStates`、`TestExternalServiceManualRetryRejectsSkippedAndDisabledPlugin`；`internal/service/plugin_package_dryrun_test.go` 的 `TestDryRunPluginPackage_OfficialWebhookNotifyExampleOK` 固化官方样板包必须通过现有 package dry-run 规则。

1. 官方样板包包含 `manifest.json`、`README.md`、`config.example.json`、`checksums.json` 和 `migrations/`。
2. 样板包 manifest 声明 `AfterCreateContent`、`service_type=external_service`、`mode=non_blocking`、`method=POST`、`retry_enabled=true`、`max_attempts=3`、`failure_policy=warn`。
3. 样板包不包含可执行代码、危险文件、真实 secret、用户数据、远程 iframe URL、根目录 SQL 或外部 SQL。
4. 样板包可用于 upload -> precheck -> promote -> install dry-run -> install。
5. 样板包可配合 `cmd/webhook-mock-receiver` 验证健康检查和内容创建后的 `/hooks/content.after_create` 投递。
6. external_service 配置表单可在后台插件详情加载。
7. external_service 配置表单可保存 endpoint、health_check_path、timeout_ms、failure_policy、auth_type、warning_threshold、error_threshold。
8. `auth_type=none` 时 token 输入不要求填写。
9. `auth_type=bearer` 时 token 只写入，不回显明文；已有 token 显示“已配置密钥 / 可替换”。
10. external_service 健康检查按钮可触发并刷新健康摘要。
11. 失败类 external_service hook_execution 显示“重试”按钮。
12. success 记录不显示重试按钮，后端也拒绝重试。
13. skipped 记录不显示重试按钮，后端也拒绝重试。
14. internal/builtin Hook 不显示重试按钮，后端也拒绝重试。
15. 跨插件 code 手动重试被后端拒绝。
16. disabled / archived / soft_uninstalled 插件不实际调用 endpoint。
17. 手动重试复用 external_service 投递逻辑。
18. 手动重试创建新的 hook_execution，metadata 标记 `manual_retry=true` 和来源执行记录。
19. 手动重试写入 `admin_logs`，action 为 `plugin.webhook.manual_retry`。
20. 手动重试成功会更新 external_service health。
21. 手动重试失败按现有 failure_policy / failure_count / warning_threshold / error_threshold 处理。
22. Webhook 治理“外部服务执行”Tab 可按插件编码查看 external_service 执行记录。
23. Webhook 治理失败记录可点击“重试”并刷新列表。
24. Authorization Header、Bearer token、Webhook Secret、Callback Token 不进入 API 响应。
25. Authorization Header、Bearer token、Webhook Secret、Callback Token 不进入 `hook_executions` 明文。
26. Authorization Header、Bearer token、Webhook Secret、Callback Token 不进入 admin_logs 明文。
27. 不改变 Webhook Secret / Callback Token 安全模型。
28. 不开放远程 iframe URL。
29. 不执行第三方插件代码。
30. 不开放 blocking Hook。
31. 不实现插件市场、远程自动安装或在线更新。
32. `gofmt`、`go test ./...`、`go build ./...`、`git diff --check`、`./scripts/check-frontend.sh --admin-only --quick` 需执行；如本地缺少 Go / Node / Docker 环境，验收记录必须写明未执行原因，不能写成通过。

### v1.8.4-S4 配置模型正则兼容修复验收项

说明：S4 统一修复 `official_announcement` 的 `link_url` 配置规则和 `official_webhook_notify_template` 的 `health_check_path` 配置规则，使用 JS / JSON Schema 兼容的路径正则，并补充配置模型编译回归测试。

建议命令：

```bash
go test ./internal/service -run TestConfigSchemaPatternsAreJSCompatible -count=1
go test ./...
go build ./...
git diff --check
```

验收项：

1. `official_announcement` 的 `link_url` pattern 使用 JS / JSON Schema 兼容写法。
2. `official_webhook_notify_template` 的 `health_check_path` pattern 使用 JS / JSON Schema 兼容写法。
3. 配置模型可正常编译并接受合法路径。
4. 非法路径（例如带空白或远程 URL）仍被拒绝。
5. `scripts/check-frontend.sh` 不再包含旧空白字符类写法。

本轮执行结果（2026-05-21）：

1. 旧字符类残留 grep：无结果。
2. `go test ./internal/service -run TestConfigSchemaPatternsAreJSCompatible -count=1`：通过。
3. `go test ./...`：通过。
4. `go build ./...`：通过；Docker 容器内需先设置 Git `safe.directory` 以允许 VCS stamping。
5. `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260521-190416/`。
6. `git diff --check`：通过。

### v1.8.3-S15 真实声明型插件“安装到使用”闭环验收项

真实 fixture：`scripts/build-plugin-package-fixtures.sh` 生成 `devhub-fixture-links-plugin*.zip`，插件编码为 `official_links*`，content_type 为 `friend_link*`。

建议命令：

```bash
./scripts/build-plugin-package-fixtures.sh --suffix smoke
go test ./internal/transport/httpapi -run TestAdminPostCreateSupportsDeclarativePluginContentType -count=1
go build -o .devhub/devhub .
docker compose up -d --force-recreate devhub
docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-declarative-install-use.spec.js
```

验收项：

1. 真实声明型插件包可生成。
2. 真实声明型插件包可上传。
3. precheck 可识别 manifest。
4. precheck 可识别 `migrations/`。
5. precheck 可识别 menus。
6. precheck 可识别 content_types。
7. precheck 可识别 permissions。
8. precheck 可识别 config_schema。
9. promote 后进入本地仓库。
10. install 前必须重新 dry-run。
11. dry-run 不执行 SQL。
12. install 只基于 `migrations/`。
13. install 不执行 package scripts。
14. install 成功后 PluginRegistry reload。
15. 插件可全局启用。
16. 插件可子站启用。
17. 插件菜单 enabled 后可见。
18. 插件菜单 disabled 后隐藏或不可用。
19. content_type enabled 后可用。
20. content_type disabled 后不可新建。
21. 板块不允许 content_type 时发布失败。
22. 权限不足时发布失败。
23. 权限矩阵显示插件权限。
24. 配置可读取。
25. 配置可保存。
26. 配置保存后 ResolvedConfig 更新。
27. 前端 / 后台挂载声明可见。
28. 子站 disabled 后插件能力不可用。
29. 全局 disabled 后插件能力不可用。
30. archived 后插件新能力停止。
31. archived 后历史内容可查看。
32. 历史审计可查看。
33. 不执行第三方代码。
34. 不开放动态加载。
35. 不开放远程 iframe。
36. blocking Hook 仍未开放。

### v1.8.3-S1 第一批收敛验收项

1. Webhook 治理页不白屏。
2. events / deliveries 等接口返回 null 时显示空状态。
3. 插件详情抽屉打开不报 maturityLabel 异常。
4. 插件列表不因 null plugin 报 code 读取异常。
5. 插件详情切换不同插件不报错。
6. 左侧插件导航收敛为 5 个治理域，或等效页面分组。
7. 插件列表首屏信息密度降低。
8. 1024 宽度下插件列表可用性改善。
9. 插件包治理能力有统一分组。
10. 上传包详情 JSON 默认折叠到技术详情。
11. 可信发布者页面主信息更清晰。
12. official_announcement 可在插件列表快速识别。
13. official_announcement 详情保留专属说明和预览入口。
14. 主要按钮使用中文。
15. 状态 badge 使用中文。
16. 空状态使用中文。
17. 错误提示使用中文。
18. Secret 明文不显示在表格。
19. Callback Token 明文不显示在表格。
20. 远程 iframe URL 不开放。
21. 第三方代码执行仍未开放。
22. blocking Hook 仍未开放。
23. 后台构建通过。
24. `/admin-next/plugins/operations` 接口响应为空对象或已解包业务对象时，不因 `undefined.items` 白屏。

### v1.8.3-S8 插件包 migrations 规范收口验收项

1. 插件包 migrations/ 被识别为唯一标准迁移目录。
2. migrations/ 下 SQL 按文件名排序。
3. 根目录 001_schema.sql 不再执行。
4. 根目录 001_schema.sql 出现时给出 deprecated warning。
5. 插件模板不再生成根目录 001_schema.sql。
6. 示例插件包使用 migrations/ 或明确无迁移。
7. dry-run 不执行 SQL。
8. dry-run 不修改数据库。
9. dry-run 不写插件安装状态。
10. dry-run 输出 migration plan。
11. dry-run 输出 deprecated warning。
12. install 执行只基于 migrations/。
13. upgrade 执行只基于 migrations/。
14. 根目录其他 .sql 文件不执行。
15. 不执行插件包脚本。
16. 不执行第三方代码。
17. 不改变 Webhook 协议。
18. 不改变 Secret / Token 安全模型。

### v1.8.3-S2 二级导航收敛与三级 Tab 重组验收项

1. 左侧插件导航只保留 5 个治理域。
2. 插件总览域包含总览 / 插件列表 / 配置中心 / 前端挂载 / 内容治理 / 权限矩阵 / 开发者工具 Tab。
3. 插件包治理域包含远程索引 / 远程包下载 / 暂存上传包 / 本地包与预检 / 依赖兼容性 / 版本升级 / 审批 Tab。
4. Webhook 治理域包含总览 / 事件通知 / 投递记录 / 重试队列 / 熔断状态 / Webhook 密钥 / 回调 Token / 回调请求 Tab。
5. 发布者与信任域包含发布者列表 / 公钥 / 可信级别 / 影响分析 / 密钥轮换 Tab。
6. 运行记录 / 审计域包含操作历史 / 审计日志 / Hook 排障 / 搜索索引 / 最近错误 Tab。
7. 旧插件路由不白屏。
8. 旧插件路由能进入对应治理域和 Tab。
9. 刷新页面后 Tab 状态保持。
10. 低频入口不再直接堆在左侧导航。
11. 同一功能不在多个治理域重复展示完整表格。
12. Webhook 页面空数据不白屏。
13. 插件详情抽屉不重复展示全局治理表格。
14. 导航和 Tab 文案中文化。
15. Secret 明文不显示。
16. Callback Token 明文不显示。
17. 远程 iframe URL 仍不开放。
18. 第三方代码执行仍未开放。
19. blocking Hook 仍未开放。
20. 后台构建通过。

### v1.8.3-S4 总览页布局与筛选体验验收项

建议命令：

```bash
./scripts/check-admin-plugin-ia.sh
./scripts/check-frontend.sh --admin-only --quick
```

说明：`check-admin-plugin-ia.sh` 是轻量插件 IA 回归脚本，用于浏览器打开 5 个治理域、截图、验证旧路由到 Tab 跳转、检查标题 / 面包屑 / 中文空状态和 1024 宽度溢出；不替代完整后台 E2E。

1. 插件总览页 1366 宽度布局自然。
2. 插件总览页 1024 宽度布局可用。
3. 插件总览页不再出现异常左侧大块空白。
4. 插件总览页内容不再明显偏右。
5. 插件列表筛选区为紧凑横向筛选栏。
6. 搜索、状态、健康、内容类型、能力、重置、刷新默认可见。
7. 高级筛选默认折叠。
8. 健康摘要不再占据过多首屏空间。
9. 快捷操作位置自然。
10. 插件列表主表格在首屏更突出。
11. 插件详情抽屉不重复展示全局治理表格。
12. official_announcement 配置入口清晰。
13. official_announcement 前端挂载说明清晰。
14. official_announcement 预览入口清晰。
15. 旧路由与 Tab 检查有脚本或文档命令。
16. 空状态完成第二轮中文化。
17. 弹窗完成第二轮中文化。
18. 按钮完成第二轮中文化。
19. Tooltip 完成第二轮中文化。
20. 分页完成第二轮中文化。
21. No Data 不裸露。
22. 20/page 不裸露。
23. Secret 明文不显示在表格。
24. Callback Token 明文不显示在表格。
25. 远程 iframe URL 仍未开放。
26. 第三方代码执行仍未开放。
27. blocking Hook 仍未开放。
28. admin quick 检查通过。

### v1.8.3-S5 插件详情抽屉视觉 polish 与信息减负验收项

1. 插件详情抽屉打开不白屏。
2. 插件详情抽屉切换不同插件不报错。
3. 长 Tab 可读性改善。
4. 概览 Tab 不再平铺技术详情。
5. 配置 Tab 只展示高频配置和配置入口。
6. 原始配置 JSON 默认进入技术详情。
7. 前端挂载 Tab 展示 iframe / sandbox / postMessage 摘要。
8. Webhook Tab 只展示当前插件摘要和跳转入口。
9. Webhook 密钥与回调 Token 收敛到安全凭据。
10. 技术详情默认折叠。
11. 上传包 / Manifest / config_schema 等 JSON 有中文标题。
12. 空技术详情显示“暂无技术详情”。
13. official_announcement 配置入口清晰。
14. official_announcement 前端挂载说明清晰。
15. official_announcement 预览入口清晰。
16. Secret 明文不显示。
17. Callback Token 明文不显示。
18. token_hash 不显示。
19. Authorization Header 不显示。
20. 完整 HMAC signature 不显示。
21. 远程 iframe URL 仍未开放。
22. 第三方代码执行仍未开放。
23. blocking Hook 仍未开放。
24. admin quick 检查通过。

### v1.8.3-S6 插件详情抽屉性能拆包与视觉回归验收项

1. PluginConfigVersionsDialog 按需加载。
2. 插件详情抽屉低频 Tab 按需加载。
3. 技术详情 Tab 按需加载或默认折叠。
4. 1024 宽度下详情抽屉可用。
5. 1024 宽度下 Tab 横向滚动自然。
6. 1024 宽度下技术详情 JSON 不撑破抽屉。
7. 1366 宽度下详情抽屉布局自然。
8. official_announcement 配置入口清晰。
9. official_announcement 前端挂载说明清晰。
10. official_announcement 预览入口清晰。
11. 配置版本弹窗不溢出。
12. 配置版本弹窗空状态中文。
13. 技术详情默认折叠。
14. 技术详情空状态显示“暂无技术详情”。
15. Secret 明文不显示。
16. Callback Token 明文不显示。
17. token_hash 不显示。
18. Authorization Header 不显示。
19. 完整 HMAC signature 不显示。
20. 远程 iframe URL 仍未开放。
21. 第三方代码执行仍未开放。
22. blocking Hook 仍未开放。
23. admin quick 检查通过。

### v1.8.3-S7 official_announcement 固定 fixture 与浏览器回归验收项

建议命令：

```bash
./scripts/check-admin-plugin-ia.sh
./scripts/check-frontend.sh --admin-only --quick
```

说明：`check-admin-plugin-ia.sh` 会在浏览器回归前通过现有后台 API 幂等准备 `official_announcement` fixture；该脚本用于插件 IA / UI 回归，不替代完整后台 E2E。

1. official_announcement fixture 可准备。
2. fixture 多次执行不重复污染数据：配置一致时跳过重复写入。
3. official_announcement 在插件列表中稳定可见。
4. 插件列表显示“官方公告插件”。
5. 插件列表显示“官方内置插件”标识。
6. official_announcement 详情抽屉可打开。
7. 配置 Tab 可截图。
8. 前端挂载 Tab 可截图。
9. 预览区域可截图。
10. iframe 容器可见。
11. fixture 下配置 `enabled=true`。
12. fixture 下 `message` 非空，固定为“欢迎使用 DevHub 官方公告插件”。
13. PHP 或首个可用子站启用状态可覆盖。
14. 浏览器回归脚本能强制覆盖 official_announcement，不再因为数据缺失条件跳过。
15. 页面不显示 callback token 明文。
16. 页面不显示 webhook secret 明文。
17. iframe DOM 不包含 token_hash、authorization、secret/token query。
18. iframe 路由仍为内置 `/plugins/official-announcement/iframe`。
19. iframe sandbox 仍为 `allow-scripts`。
20. 不开放远程 iframe URL。
21. 不执行第三方代码。
22. blocking Hook 仍未开放。
23. admin quick 检查通过。

本轮结果记录：

- `./scripts/check-frontend.sh --admin-only --quick`：通过（后台 build 成功），日志目录 `.devhub/checks/20260518-194453/`。
- `./scripts/check-frontend.sh --admin-only --quick`：v1.8.3-S1 复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260518-214606/`。
- `./scripts/check-frontend.sh --admin-only --quick`：操作历史页响应解包修复后复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260518-215550/`。
- `./scripts/check-frontend.sh --admin-only --quick`：v1.8.3-S2 复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260518-221921/`。
- `./scripts/check-frontend.sh --admin-only --quick`：插件后台筛选条 UI 修正后复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260518-223539/`。
- `./scripts/check-frontend.sh --admin-only --quick`：插件后台二级导航纠偏后复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260518-225027/`。
- `scripts/check-admin-plugin-ia.sh`：v1.8.3-S4 浏览器 IA 回归通过，截图目录 `/workspace/.devhub/screenshots/v1.8.3-s4`，覆盖 5 个治理域、旧路由到 Tab 跳转、1024 宽度和中文空状态。
- `./scripts/check-frontend.sh --admin-only --quick`：v1.8.3-S4 复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260519-002055/`。
- `./scripts/check-frontend.sh --admin-only --quick`：v1.8.3-S5 复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260519-003946/`。
- `scripts/check-admin-plugin-ia.sh`：v1.8.3-S6 浏览器回归通过，截图目录 `.devhub/screenshots/v1.8.3-s6`，覆盖普通插件详情 1024、技术详情、配置 Tab、配置版本弹窗、5 个治理域 1366；当前测试数据未返回 `official_announcement`，脚本记录跳过原因。
- `./scripts/check-frontend.sh --admin-only --quick`：v1.8.3-S6 复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260519-010359/`。
- `scripts/check-admin-plugin-ia.sh`：v1.8.3-S7 浏览器回归通过，截图目录 `.devhub/screenshots/plugin-ia`，报告文件 `.devhub/screenshots/plugin-ia/report.json`；fixture 固定 `official_announcement` 全局与 PHP 子站启用，强制覆盖官方公告插件列表、详情概览、公告配置、前端挂载、公告预览、iframe 内置路由和 sandbox。
- `./scripts/check-frontend.sh --admin-only --quick`：v1.8.3-S7 复跑通过（后台 build 成功），日志目录 `.devhub/checks/20260519-014314/`。
- `gofmt internal/transport/httpapi/plugin_mount_host_helper.go`、`go test ./...`、`go build ./...`：v1.8.3-S7 通过；本轮因修复后台预览 Host helper 鉴权而补跑 Go 检查。
- `git diff --check`：通过。
- `bash -n scripts/check-frontend.sh`：通过。

补充说明：首次 quick 检查因 Docker named volume 中 `web/admin-app/node_modules` 属主为 root，导致 Vite 无法写入 `node_modules/.vite-temp` 而失败；已在 `scripts/check-frontend.sh` 中增加运行前 workspace 权限初始化，复跑通过。

本轮还需人工关注：上传记录、远程插件包、审批中心、操作历史、配置版本历史/回滚预览页面主要字段为中文；用户可见 `dry-run` 操作显示为“预检”，JSON key / 接口字段仍可保留原值用于排障；上传包转入本地仓库、删除上传包、操作清理、可信发布者禁用/吊销/恢复等危险动作有中文确认提示。

## 测试目标定位

DevHub 当前目标是 **Core + 插件 的开源服务底座**，测试体系不再只围绕默认社区页面。测试需要覆盖 Core 基础能力、后台管理能力、插件生命周期、插件包治理、插件权限治理、插件配置治理、插件治理 UI、前台默认社区能力和 SEO 输出。

本轮“统一 DevHub 项目目标为 Core + 插件服务底座”为纯文档整理与项目目标统一任务，未修改代码，未执行测试、构建或 E2E。

v1.7.2 插件运行模型设计任务同样为文档设计任务：主要修改文档，未修改代码，未执行测试、构建或 E2E。后续实现阶段需要新增针对 HTTP 插件服务、iframe / sandbox 挂载、受控 API token、HookBus 远程调用、权限隔离和审计的专项测试矩阵。

## Webhook / HTTP 插件服务协议测试规划（设计）

说明：本节是 v1.7.2 Webhook / HTTP 插件服务协议的测试规划草案，本轮不执行测试。

建议覆盖：

1. 签名正确投递成功（2xx）。
2. 签名错误投递失败（401/403），并写入审计 `plugin.webhook.signature.failed`。
3. timestamp 过期失败（防重放窗口外）。
4. body 被篡改（body_sha256 不一致）失败。
5. blocking hook 返回 `decision=deny` 时阻断主流程。
6. non_blocking hook 失败不阻断主流程，但写入 delivery failed + 审计。
7. 5xx / timeout 触发重试并记录 `retry_scheduled`。
8. 4xx 默认不重试（除非明确允许 429）。
9. 429 按 `Retry-After` 重试；无 `Retry-After` 时按退避策略重试。
10. 超过最大重试次数标记 `retry_exhausted` 并写审计。
11. `idempotency_key` 重复投递插件服务不重复处理（幂等）。
12. 插件服务超时记录失败，包含 duration_ms 与错误码。
13. 熔断开启后暂停投递并记录 `circuit_open/skipped`。
14. 插件回调 Core API scope 不足时被拒绝并写审计（设计）。
15. 插件回调写入审计 `plugin.webhook.callback.accepted/rejected`（设计）。

补充（v1.7.3 文档拆解阶段）：

- 实现阶段拆解见 `docs/PLUGIN_WEBHOOK_IMPLEMENTATION_PLAN.md`，第一优先级为 non_blocking delivery（不阻断主流程），blocking Hook 明确后置。
- 官方示例插件（公告插件）端到端验证方案见 `docs/plugins/official-announcement-plugin.md`，用于后续验证 delivery 记录 / 重试 / 熔断 / 审计 / 后台治理入口（仍不执行第三方代码）。

## v1.7.5：Webhook 重试队列与熔断机制（实现）

说明：本节为 v1.7.5 实现轮验收清单。本轮仅治理 non_blocking delivery 的 **重试队列** 与 **熔断机制**，仍不执行第三方插件代码，仍不实现 blocking Hook。

本轮检查命令说明（按最低检查要求）：

- 已执行：`gofmt`、`go test ./...`、`go build`。
- 后台构建：在当前执行环境中，`node` 缺失导致 `cd web/admin-app && npm run build` 无法执行（`vite` 不可用，且出现 Windows/UNC CMD 提示）。按 `docs/AGENT_RULES.md`，建议使用 Docker/compose 或 `scripts/check-frontend.sh --admin-only` 在具备 node 的容器环境执行并归档结果；未执行不能写成通过。

1. 5xx delivery 失败后进入 `retry_scheduled`。
2. timeout delivery 失败后进入 `retry_scheduled`。
3. 429 delivery 根据 `Retry-After` 或默认策略进入 `retry_scheduled`。
4. 4xx delivery 默认不重试。
5. `retry_scheduled` 有 `next_retry_at`。
6. 到期 delivery 可以被 `retry-due` 扫描重试。
7. 超过 `max_attempts` 后进入 `retry_exhausted`。
8. `retry_exhausted` 可以手动重试（记录 `manual_retry`）。
9. `success` delivery 默认不能重复重试。
10. 重试保持 `event_id` 稳定（不重新生成新 event_id）。
11. 重试保持 `idempotency_key` 稳定（如当前实现未落库，需在后续任务补齐并保持稳定）。
12. 每次 attempt 有记录（`attempt/max_attempts`）。
13. 连续失败达到阈值后 circuit `open`。
14. circuit `open` 后 delivery 被标记为 `circuit_open` / `skipped`。
15. circuit `open` 后不继续打爆插件服务。
16. 到达 `next_probe_at` 后允许 `half_open` 探测。
17. `half_open` 成功后 circuit `closed`。
18. `half_open` 失败后继续 `open`。
19. 管理员可以手动恢复熔断。
20. 熔断状态可以在后台查看。
21. delivery 列表可以在后台查看。
22. 手动重试可以在后台触发。
23. 重试和熔断操作写入审计。
24. 普通用户不能访问 Webhook 治理 API。
25. non_blocking delivery 失败不影响 Core 主流程。
26. disabled 插件不投递。
27. soft_uninstalled 插件不投递。
28. 本轮不执行第三方插件代码。
29. 本轮不实现 blocking Hook。
30. 不影响 `/topics/:id` SEO。
31. 不影响 `/c/:slug` SEO。

## v1.7.6：Webhook 签名鉴权与 Secret 轮换（实现）

说明：本节为 v1.7.6 实现轮的轻量验收清单。本轮补齐 DevHub → 插件服务的 HMAC-SHA256 发送端签名、Secret 管理与轮换窗口，并保持 non_blocking delivery 的安全边界。

1. 可以创建 Webhook Secret。
2. 创建 Secret 时明文只返回一次。
3. Secret 列表不返回明文。
4. Secret 详情不返回明文。
5. 可以禁用 Secret。
6. disabled Secret 不能用于签名。
7. 可以吊销 Secret。
8. revoked Secret 不能用于签名。
9. 可以轮换 Secret。
10. 轮换后新 Secret 为 active。
11. 轮换后旧 Secret 为 previous。
12. grace period 内接收端示例可接受 previous Secret（文档/示例）。
13. grace period 后 previous Secret 过期（expired）。
14. 没有 active Secret 时 delivery 签名失败。
15. 发送 Webhook 时包含签名 Header。
16. body_sha256 与实际 body 一致。
17. body 被篡改时接收端验签失败（文档/示例）。
18. timestamp 过期时接收端验签失败（文档/示例）。
19. path 不一致时验签失败（文档/示例）。
20. method 不一致时验签失败（文档/示例）。
21. HMAC 签名不一致时验签失败（文档/示例）。
22. 远程 401 默认不重试。
23. 远程 403 默认不重试。
24. 远程 5xx 仍按重试队列处理。
25. 远程 timeout 仍按重试队列处理。
26. 签名生成失败写入审计（或写入 delivery signature_status，并可通过审计追踪）。
27. Secret 创建 / 轮换 / 禁用 / 吊销写入审计。
28. 审计不包含 Secret 明文。
29. 后台可以查看 Secret 状态。
30. 后台可以轮换 Secret。
31. 官方公告插件文档包含接收端验签说明。
32. 普通用户不能访问 Webhook Secret API。
33. 本轮不执行第三方代码。
34. 本轮不实现 blocking Hook。
35. 不影响 `/topics/:id` SEO。
36. 不影响 `/c/:slug` SEO。

## v1.7.6-S1：Webhook 签名鉴权与 Secret 轮换（专项验收 / 补测 / 修缺口）

说明：本节为 v1.7.6-S1 专项验收清单，用于确认 v1.7.6 的签名鉴权与 Secret 轮换真实可用，且无敏感信息泄露。本轮仍只覆盖 non_blocking Webhook。

1. Webhook 请求包含签名 Header（Event/Delivery/Plugin/Timestamp/Signature/Alg/Idempotency/RequestID/BodySHA256/SecretRef）。
2. `X-DevHub-Signature-Alg=HMAC-SHA256` 且 `X-DevHub-Signature` 为 `v1=<hex>` 格式。
3. signing string 规则与文档一致：`timestamp.method.path.body_sha256`。
4. `body_sha256` 与实际 body 一致，并参与签名（body 改变后验签失败）。
5. `timestamp` 参与签名（timestamp 改变后验签失败）。
6. `path` 与 `method` 参与签名（任一不一致验签失败）。
7. Secret 创建接口明文只返回一次（创建成功响应包含 `secret`，列表/详情不返回）。
8. Secret 列表/详情不返回明文，也不返回 `secret_ciphertext`。
9. Secret 明文不进入审计、不进入日志、不进入 delivery 记录（包括 request_headers_json）。
10. disabled Secret 不能用于签名（不会发送网络请求，delivery 标记 failed 且 `signature_status=secret_disabled`）。
11. revoked Secret 不能用于签名（同上，`signature_status=secret_revoked`）。
12. expired Secret 不能用于签名（同上，`signature_status=secret_expired`）。
13. 轮换后新 Secret 为 active，旧 Secret 为 previous 并设置 grace period。
14. grace period 到期后 previous Secret 变为 expired（可通过 sweep 或治理动作触发）。
15. 无 active Secret 时 delivery 签名失败（`signature_status=secret_missing`）。
16. 远端 401/403 默认不重试（不进入 `retry_scheduled`，`next_retry_at` 为空）。
17. 远端 5xx / timeout 仍按重试队列策略调度（不属于本节新增，但需保持不退化）。
18. request_headers_json 中 signature 被脱敏存储（`v1=[REDACTED]`），不存完整签名。

## v1.7.7：Webhook 插件回调 Core API（Callback Token + Scopes）（实现）

说明：本节为 v1.7.7 实现轮的轻量验收清单。本轮为外部 HTTP 插件服务增加受控 Core API 回调通道（callback token + scope + community scope），仍不执行第三方插件代码，仍不实现 blocking Hook。

1. 可以创建 callback token（`POST /api/v1/admin/plugins/callback-tokens`）。
2. 创建 token 时明文只返回一次（响应包含 `token`，列表/详情不返回）。
3. token 列表不返回明文（`GET /api/v1/admin/plugins/callback-tokens`）。
4. token 详情不返回明文（`GET /api/v1/admin/plugins/callback-tokens/:id`）。
5. 可以禁用 token（disabled token 调用 callback API 返回 401）。
6. 可以恢复 disabled token。
7. 可以吊销 token（revoked token 调用 callback API 返回 401）。
8. expired token 调用 callback API 返回 401（如实现包含 expires_at 校验）。
9. 可以轮换 token（轮换后新 token 明文只返回一次；旧 token 按策略失效）。
10. 缺少 token 返回 401（TOKEN_MISSING）。
11. token 无效返回 401（TOKEN_INVALID）。
12. scope 不足返回 403（SCOPE_DENIED）。
13. community scope 不匹配返回 403（COMMUNITY_SCOPE_DENIED）。
14. 插件 global disabled / soft_uninstalled 后 token 不能调用 callback API（PLUGIN_DISABLED / plugin_disabled）。
15. `config.read` 只能读取本插件配置（`GET /api/v1/plugin-callback/config?community_id=...`）。
16. `audit.write` 可以写入插件审计事件（`POST /api/v1/plugin-callback/audit-events`）。
17. `audit.write` 不能伪造 admin/Core action（action 必须以 `plugin_code.` 前缀开头）。
18. callback request 有记录（`plugin_callback_requests` 或等效记录），且不保存 token 明文。
19. Token 创建/轮换/禁用/吊销写入审计（admin_logs）。
20. callback accepted/rejected 写入审计（admin_logs）。
21. 后台 `Webhook 治理` 页可查看 Callback Tokens 与 Callback Requests（页内 Tab，不进入左侧菜单）。
22. 普通用户不能访问 Token 管理 API（`/api/v1/admin/plugins/callback-tokens*`）。
23. 本轮不执行第三方代码。
24. 本轮不实现 blocking Hook。
25. 不影响 `/topics/:id` SEO。
26. 不影响 `/c/:slug` SEO。

## v1.7.8：Webhook 后台治理与官方公告插件端到端验证（端到端验收清单）

说明：本节用于 v1.7.8 端到端验证。验证目标是打通 non_blocking Webhook 的治理闭环与官方公告插件（official_announcement）的验证步骤。仍不执行第三方不可信代码，仍不实现 blocking Hook。

场景 1：正常投递成功

1. 创建/启用 `official_announcement`（作为治理与协议验证样例；不执行第三方代码）。
2. 创建 Webhook Secret（active）。
3. 创建 callback token（scope：`config.read` + `audit.write`，并设置 `community_scope`）。
4. 触发 `content.after_create`（或用现有治理入口触发一条 non_blocking delivery）。
5. 生成 `webhook_event` 记录。
6. 生成 `webhook_delivery` 记录。
7. official mock receiver 验签通过并返回 2xx。
8. delivery status=`success`。
9. 后台 `Webhook 治理` 页可查看 Events/Deliveries。
10. 审计存在（delivery created/success + callback accepted）。

场景 2：签名失败（receiver 401）

1. receiver 端使用错误 secret_ref/plaintext secret（导致验签失败）。
2. 触发 event。
3. receiver 返回 401。
4. delivery status=`failed`。
5. 不进入无限重试（401 默认不重试）。
6. 审计存在（delivery failed / receiver 401）。

场景 3：5xx 重试

1. receiver 返回 500。
2. delivery status=`retry_scheduled`。
3. `next_retry_at` 有值。
4. 执行 `retry-due` 后重新投递。
5. receiver 恢复后 delivery status=`success`。

场景 4：429 Retry-After

1. receiver 返回 429 且带 `Retry-After`。
2. delivery 进入 `retry_scheduled`。
3. `next_retry_at` 按 `Retry-After` 或默认策略设置。
4. 后续重试成功。

场景 5：熔断与恢复

1. receiver 连续失败达到阈值。
2. circuit breaker status=`open`。
3. 新 delivery 标记 `circuit_open`/`skipped`。
4. 管理员手动恢复熔断（status=`closed`）。
5. 恢复后投递成功。

场景 6：Secret 轮换

1. active Secret 投递成功。
2. 轮换 Secret：新 secret_ref 为 active，旧为 previous。
3. 新 delivery 使用新 secret_ref。
4. grace window 内接收端允许 previous secret_ref。
5. grace window 后 previous 过期（expired），不再允许验签。

场景 7：Callback config.read

1. token scope 包含 `config.read`。
2. receiver 使用 callback token 调用 `GET /api/v1/plugin-callback/config?community_id=...`。
3. 返回本插件 effective_config。
4. 不能读取其他插件配置。
5. callback request 有记录。
6. 审计存在（config.read + accepted/rejected）。

场景 8：Callback audit.write

1. token scope 包含 `audit.write`。
2. receiver 调用 `POST /api/v1/plugin-callback/audit-events`。
3. action 必须以 `official_announcement.` 前缀开头，防止伪造 admin/Core 审计。
4. callback request 有记录。

场景 9：scope denied

1. 创建不包含 `audit.write` 的 token。
2. 调用 audit-events。
3. 返回 403（SCOPE_DENIED）。
4. callback request status=`rejected`。
5. 审计存在（scope denied）。

场景 10：plugin disabled

1. 禁用 `official_announcement`。
2. 不再投递 Webhook（disabled/soft_uninstalled 插件不投递）。
3. callback token 调用 callback API 返回 403（PLUGIN_DISABLED）。
4. 历史 delivery / audit 仍可查看。

## v1.7.9：Webhook non_blocking 链路总验收（收口清单）

说明：本节为 v1.7.9 总验收清单。目标是确认 non_blocking Webhook 链路“真实闭环可用”，并且权限、审计、敏感信息保护与 MySQL/MemoryStore 行为一致。

1. non_blocking event 可以生成并持久化（`webhook_events`）。
2. delivery 可以生成并持久化（`webhook_deliveries`）。
3. delivery 成功可记录为 `success`。
4. 5xx 进入 `retry_scheduled`。
5. timeout 进入 `retry_scheduled`。
6. 429 按 `Retry-After` 或默认策略进入 `retry_scheduled`。
7. 4xx 默认不重试（含 400/404/422）。
8. 401 / 403 默认不重试。
9. 超过 `max_attempts` 进入 `retry_exhausted`。
10. 手动重试可用（failed / retry_exhausted）。
11. 连续失败达到阈值后 circuit breaker 进入 `open`。
12. `open` 后暂停投递（delivery 标记 `circuit_open`/`skipped`）。
13. `half_open` 探测成功后回到 `closed`。
14. 管理员手动恢复熔断可用。
15. Webhook 签名 Header 完整（Event/Delivery/Plugin/Timestamp/Signature/Alg/Idempotency/RequestID/BodySHA256/SecretRef）。
16. `body_sha256` 参与签名（body 变化导致验签失败）。
17. `timestamp/method/path` 参与签名（任一不一致验签失败）。
18. Secret 创建明文只返回一次；列表/详情不返回明文/密文。
19. Secret 轮换可用（active/previous grace window 生效；previous 可过期）。
20. Secret 明文不进入日志/审计/delivery（含 request_headers_json）。
21. callback token 创建明文只返回一次；列表/详情不返回明文。
22. callback token scope 校验生效（scope 不足返回 403）。
23. callback token community scope 校验生效（不匹配返回 403）。
24. 插件 disabled / soft_uninstalled 后 callback token 不可用（403 PLUGIN_DISABLED）。
25. `config.read` 只能读取本插件配置。
26. `audit.write` 不能伪造 admin/Core action（必须以 `plugin_code.` 前缀开头）。
27. callback request 有记录且不保存 token 明文/Authorization header。
28. 后台治理 UI 可查看 Events/Deliveries/Circuits/Secrets/Callback Tokens/Callback Requests。
29. 后台治理 UI 重要操作有确认（重试/熔断/secret&token 吊销等）。
30. 官方公告插件验证路径覆盖：成功投递、签名失败、5xx 重试、429、熔断恢复、Secret 轮换、callback config.read/audit.write、scope denied、plugin disabled。
31. 本轮仍不执行第三方代码。
32. 本轮仍不实现 blocking Hook。
33. 不影响 `/topics/:id` SEO。
34. 不影响 `/c/:slug` SEO。

## v1.8.0：插件前端挂载模型（文档设计验收清单）

说明：v1.8.0 为“官方插件前端挂载模型与 iframe / sandbox 容器设计版”，本轮以文档设计为主，未修改代码，未执行测试、构建或 E2E。以下清单用于后续实现阶段（例如 v1.8.1）对照验收，不代表本轮已实现。

1. 文档明确 slots 列表：`admin.sidebar.menu`、`admin.plugin.detail.tab`、`admin.dashboard.card`、`frontend.header.nav`、`frontend.home.section`、`frontend.topic.sidebar`、`frontend.topic.after_content`、`frontend.user.menu`、`moderator.sidebar.menu`。
2. 文档明确插件前端扩展默认使用 iframe 容器。
3. 文档明确 iframe 必须启用 sandbox，并给出基线策略与禁止项。
4. 文档明确插件 iframe 与 DevHub Host 的 postMessage 通信模型（envelope、握手、上下文、受控请求/响应）。
5. 文档明确 postMessage 必须校验 origin、plugin_code、mount_slot、request_id。
6. 文档明确插件前端不能直接读取 DevHub token/cookie，不得绕过 Core API。
7. 文档明确插件页面必须受 plugin global enabled/disabled/soft_uninstalled 状态控制。
8. 文档明确插件页面必须受 community plugin enabled/disabled 状态控制（涉及 community 的 slots）。
9. 文档明确插件页面必须受用户权限控制（admin/moderator/user）。
10. 文档明确 `/topics/:id` 与 `/c/:slug` SEO 红线：插件前端扩展不得破坏 SEO 动态 HTML 兜底。
11. 文档明确官方公告插件 `official_announcement` 的前后台挂载验证方案（设计）：后台配置 Tab + 前台首页区块。
12. 文档明确哪些能力为设计中/未完成：真实挂载实现、Host 消息通道落地、配置写入通道（例如 `config.write`）等。

## v1.8.1：官方公告插件前端挂载最小实现（轻量验收清单）

说明：本节用于 v1.8.1 的最小闭环验收。官方公告插件为内置官方插件，iframe 页面为仓库内置页面；不支持任意远程 iframe URL，不执行第三方不可信代码。

1. 前台首页在插件 enabled 且配置 `enabled=true`、`message` 非空时显示公告 Host + iframe。
1. 子站页 `/c/:slug` 在插件全局 enabled 且子站插件 enabled 且配置 `enabled=true`、`message` 非空时显示公告 Host + iframe。
2. 插件 disabled 时前台不显示公告。
3. 插件 soft_uninstalled（archived）时前台不显示公告。
4. iframe `sandbox="allow-scripts"` 生效（不默认开启 `allow-same-origin`）。
5. iframe 路由 `GET /plugins/official-announcement/iframe` 不被 StaticFile/NoRoute fallback 吃掉。
6. iframe 启动后会发送 `devhub.plugin.ready`，Host 返回 `devhub.plugin.context`。
7. iframe 可通过 postMessage 请求 `config.read`，Host 返回官方公告插件公开配置（不包含敏感值）。
8. iframe 可通过 postMessage 请求 `audit.write`，Host 写入 `official_announcement.*` 审计事件。
9. Host API `GET /api/v1/plugins/official-announcement/context` 不返回 callback token / webhook secret / Authorization header。
10. Host API `POST /api/v1/plugins/official-announcement/audit-events` 不允许写入非 `official_announcement.*` 的 action。
11. 后台插件详情页对 `official_announcement` 显示“公告预览”Tab（Host + iframe）。
12. 权限不足用户看不到或无法使用后台预览（后端同样拒绝 `area=admin`）。
13. iframe 加载失败不影响首页主内容与 SEO。
14. 本轮不执行第三方不可信代码、不做远程 JS 动态加载、不允许远程 iframe URL。
15. 不影响 `/topics/:id` SEO 与 `/c/:slug` SEO。

## v1.8.2：iframe / sandbox 通用容器与 postMessage Host helper（轻量验收清单）

说明：本节用于 v1.8.2 的复用性验收。目标是把 v1.8.1 的 Host + iframe + postMessage 机制抽取为共享 helper，并让前台首页、`/c/:slug` 与后台插件详情复用同一套机制（第一阶段仅 allowlist 官方内置插件，不允许远程 iframe URL）。

1. 前台首页通过 `/plugins/assets/devhub-plugin-mount-host.js` 挂载 `official_announcement`（不再复制大段挂载脚本）。
2. `/c/:slug` 通过 `/plugins/assets/devhub-plugin-mount-host.js` 挂载 `official_announcement`，且携带 `data-community-slug` 参与 gating。
3. 后台插件详情页（公告预览 Tab）通过 `PluginIframeMount` 组件复用同一 helper。
4. iframe sandbox 仍为 `allow-scripts`（不默认开启 `allow-same-origin` 等高权限）。
5. iframe src 仍为内置路由 `/plugins/official-announcement/iframe`（不允许远程 URL）。
6. postMessage `ready -> context -> config.read -> config.result -> audit.write` 仍可跑通。
7. postMessage 未知 type 被拒绝（type 白名单生效）。
8. postMessage plugin_code / mount_id 不匹配被拒绝。
9. context 不包含 token / callback token / webhook secret。
10. config.read 仍只返回 `official_announcement` 的公开安全配置（不读取其他插件配置）。
11. audit.write 仍只允许 `official_announcement.*` action，不能伪造 admin。
12. plugin disabled / soft_uninstalled 时不挂载。
13. community disabled（子站插件 disabled）时不挂载。
14. iframe 加载失败不影响主页面内容与 SEO。
15. `/c/:slug` SEO 不被破坏（title/canonical/JSON-LD/h1/主体内容仍可读）。
16. `/topics/:id` SEO 不被破坏（title/canonical/JSON-LD/h1/主体内容仍可读）。
17. 本轮不执行第三方代码、不做 JS 注入、不动态加载远程 JS。

## v1.5.0 收口验收（2026-05-14）

本节记录 v1.5.0 插件包治理收口后的最终验收结果，作为当前仓库“已真实跑过的检查”口径来源。

已执行并通过：

- `gofmt -w $(git ls-files '*.go')`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `./scripts/check-frontend.sh --admin-only`（后台 build + Playwright：通过，包含插件包 dry-run、仓库、安装、审批、签名、导出等用例）
- `./scripts/check-frontend.sh --frontend-only`（前台 build + Playwright：通过）
- SEO curl 回归（在本地 8090 服务可用情况下执行）：
  - `curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|canonical|<h1|<article|application/ld\\+json'`
  - `curl -s http://127.0.0.1:8090/c/php/ | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'`

结论：

- Playwright 不再保留 `test.skip` / `test.only`；收口验收未发现长期跳过项。

## v1.5.0 收口补充：后台初始化插件包（2026-05-14）

本节记录后台“系统插件 -> 安装升级 -> 初始化插件包”能力的本轮验证结果。

已执行并通过：

- `gofmt -w cmd/devhub/main.go internal/domain/plugin_package_template.go internal/plugins/scaffold/scaffold.go internal/plugins/scaffold/scaffold_test.go internal/service/plugin_package_template_service.go internal/service/plugin_package_template_test.go internal/transport/httpapi/plugin_package_handler.go internal/transport/httpapi/router.go`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `git diff --check`
- `./scripts/check-frontend.sh --admin-only`（后台 build 通过；后台 Playwright `49 passed`）

覆盖结论：

- 后端新增模板 preview/create 服务测试：预览不写文件；正式初始化不生成 `registry.example.go`，生成 `docs/registry-example.md`，并自动 package dry-run。
- 后台构建和现有后台 E2E 全量通过；本轮未改前台页面、前台 SEO 或前台导航，因此未执行 `--frontend-only` 与 SEO curl。

## v1.6.0-P0-01 插件包 zip 上传安全沙箱（2026-05-14）

本节记录“管理员上传 zip 插件包、受控沙箱解压、复用 scanner/checksum/signature/risk_report/dry-run、promote 到本地仓库”的本轮验证结果。

已执行并通过：

- `gofmt -w internal/domain/plugin_package_upload.go internal/plugins/package_zip.go internal/plugins/package_zip_test.go internal/plugins/package_scanner.go internal/service/plugin_package_upload.go internal/service/plugin_package_upload_test.go internal/transport/httpapi/plugin_package_handler.go internal/transport/httpapi/plugin_package_upload_http_test.go internal/transport/httpapi/router.go`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-package-upload.spec.js`（通过，`4 passed`）
- `./scripts/check-frontend.sh --admin-only`（后台 build + 完整后台 Playwright 通过，`53 passed`）

已覆盖：

- 合法 zip 上传成功，根目录 `manifest.json` 与单顶层目录 `manifest.json` 均可识别。
- 无 manifest、多 manifest、非 zip、上传大小超限、`../`、绝对路径、Windows 盘符、symlink、嵌套压缩包、文件数超限、单文件超限、解压总大小超限均被拒绝。
- dangerous `.sh` 与 checksum mismatch 上传后进入 blocked/quarantine，不能 promote。
- 上传后不写插件表、不写 migration 表、不执行代码。
- promote ok 包进入 `storage/plugins/packages/{code}`，目标存在被拒绝；blocked 包 promote 被拒绝。
- HTTP 覆盖 user token 被拒绝、上传失败写 `admin_logs`、详情不返回系统绝对路径。
- 后台 E2E 覆盖上传区域、安全边界提示、zip scan、package scan、risk_report、manifest validate、dry-run、非法类型、zip slip、危险文件、blocked code/suggestion、promote 后仓库列表可见；上传区域不提供直接安装按钮。

本轮只改后端插件包 API 与后台插件安装升级页，不涉及前台内容、导航、搜索或 SEO，因此不执行 `--frontend-only` 与 SEO curl。

## v1.6.0-P0-02 插件包上传生命周期治理（2026-05-14）

本节记录“上传包从 staging 临时文件升级为生命周期对象，并支持列表、审批、promote、取消、删除、cleanup”的本轮验证结果。

已执行并通过：

- `gofmt`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-package-upload-lifecycle.spec.js`（通过，`1 passed`）
- `./scripts/check-frontend.sh --admin-only`

已覆盖：

- 上传后创建 `plugin_package_uploads` 记录，合法 zip 进入 `staged`，危险 zip 进入 `blocked`。
- 上传包详情返回扫描快照、risk_report、manifest validate、dry-run、可执行动作和不可用原因。
- `staged` 上传包可重新扫描、提交导入审批、审批通过后 promote。
- `approval_pending` 不能直接 promote；审批通过后状态为 `approved`。
- promote 前重新 dry-run，成功后状态为 `promoted`，本地插件仓库出现包目录。
- blocked 包不能 promote。
- 上传包删除白名单收口为 `uploaded/staged/blocked/failed/expired`；promoted / installed / 审批中记录不能直接删除。
- cleanup 必须先 dry-run，返回 `confirm_token/will_delete_count/will_free_bytes/items`；确认执行才删除文件和上传记录。
- 本地仓库包删除允许 `storage/plugins/packages/` 下所有未安装包；已安装包拒绝并提示先归档 / 软卸载插件。
- 本地仓库 cleanup 默认预览 `blocked/invalid`，只删除未安装包，不删除已安装插件。
- 后台 E2E 覆盖上传包管理页、上传、列表、详情、rescan、导入审批、审批通过、promote、cancel、delete、blocked 禁止 promote，以及页面不显示上传后直接安装、远程市场入口或动态加载入口。

调试记录：

- 首次单独运行 `plugin-package-upload-lifecycle.spec.js` 时失败于 Docker 内 `devhub` 服务名解析，原因是尚未执行本轮要求的 `go build -o .devhub/devhub .`，compose 中 devhub 服务找不到二进制；构建二进制并重启 devhub 后通过。
- 本轮只改后端插件包 API 与后台插件上传包管理页，不涉及前台内容、导航、搜索或 SEO，因此不执行 `--frontend-only` 与 SEO curl。

## v1.4.0-P1-13 收口补充（2026-05-13）

本节记录插件后台信息架构与按功能分页优化后的后台验收结果。

已执行并通过：

- 历史导航专项已通过；当前由 `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-governance-pages.spec.js` 承接并扩展覆盖。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-dependencies.spec.js`：通过，`2 passed`。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-governance.spec.js`：通过，`13 passed`。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 Playwright `35 passed`。

说明：

- 本轮只改后台插件板块信息架构与管理页路由，没有触碰前台导航或 SEO 共享逻辑，因此未额外执行 `--frontend-only` 与 SEO curl。
- 旧路由 `/admin-next/plugins`、`/admin-next/plugins/governance`、`/admin-next/qa`、`/admin-next/docs`、`/admin-next/wiki`、`/admin-next/projects`、`/admin-next/jobs`、`/admin-next/ai-works` 均保持兼容。

## v1.5.0-P0-01 插件包 dry-run（2026-05-13）

本节记录“本地插件包规范草案 + dry-run 导入预览”的最小验收命令口径。

必须执行：

- `gofmt -w $(git ls-files '*.go')`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-package-dryrun.spec.js`
- `./scripts/check-frontend.sh --admin-only`

说明：

- 本轮只涉及后端与后台插件板块安装升级页，不涉及前台与 SEO，因此不强制执行 `--frontend-only` 与 SEO curl（如后续改动触及前台再补）。

## v1.5.0-P0-02 插件包安全校验与风险报告（2026-05-13）

本节记录“checksums.json（sha256）校验 + 危险规则强化 + risk_report 风险报告”的最小验收命令与结果。

已执行（通过）：

- `gofmt -w internal/domain/plugin_package.go internal/service/plugin_package_service.go internal/plugins/*.go internal/service/plugin_package_dryrun_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-package-security.spec.js`：通过，`4 passed`。
- `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 ./scripts/check-frontend.sh --admin-only`：通过，后台 Playwright `42 passed`。

说明：

- 本机 `8090` 端口被占用时，可用 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:<port>` 覆盖 Playwright baseURL；本轮在本机启动 `PORT=8091 CMS_STORE=memory DEVHUB_E2E_TESTING=1 ./.devhub/devhub` 后执行上述 E2E。
- 本轮只改后端与后台 dry-run 页面，不涉及前台与 SEO，因此未执行 `--frontend-only` 与 SEO curl。

## v1.5.0-P0-03 本地插件仓库目录与扫描（2026-05-13）

已执行（通过）：

- `gofmt -w $(git ls-files '*.go')`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-package-repository.spec.js`：通过。
- `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 ./scripts/check-frontend.sh --admin-only`：通过。

说明：

- 若本机 `8090` 端口被占用，可通过 `DEVHUB_E2E_ORIGIN` 指定 Playwright baseURL；仓库扫描与详情仅依赖后端 API，不涉及前台与 SEO。

## v1.5.0-P0-04 本地插件包安装闭环（2026-05-13）

已执行（通过）：

- `gofmt -w internal/domain/plugin_package_install.go internal/service/plugin_package_install.go internal/service/plugin_package_install_test.go internal/service/service.go internal/transport/httpapi/router.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `PORT=8091 CMS_STORE=memory DEVHUB_E2E_TESTING=1 ./.devhub/devhub`：启动用于 Playwright 的本机后端（通过 `DEVHUB_E2E_ORIGIN` 指向）。
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-package-install.spec.js`：通过。
- `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 ./scripts/check-frontend.sh --admin-only`：通过，后台 Playwright `44 passed`。

说明：

- 插件包安装仍只写入声明/配置/迁移 pending/审计；不执行第三方代码/SQL，不动态加载前端资产。
- 本轮只涉及后端与后台插件安装升级页，不涉及前台与 SEO，因此未执行 `--frontend-only` 与 SEO curl。

## v1.5.0-P1-05 插件配置版本历史与回滚预览（dry-run）（2026-05-13）

已执行（通过）：

- `gofmt -w internal/domain/plugin_config_versions.go internal/plugins/config_diff.go internal/plugins/config_diff_test.go internal/service/plugin_config_versions.go internal/service/plugin_config_versions_test.go internal/transport/httpapi/router_auth_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `PORT=8091 CMS_STORE=memory DEVHUB_E2E_TESTING=1 ./.devhub/devhub`：启动用于 Playwright 的本机后端；随后执行：
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-config-versions.spec.js`：通过，`1 passed`。

说明：

- 本轮新增配置版本历史与回滚预览（dry-run），只涉及后端与后台插件配置页，不涉及前台与 SEO，因此未执行 `--frontend-only` 与 SEO curl。
- 回滚仅提供 dry-run 预览，不提供真实回滚写入。

## v1.5.0-P1-06 插件敏感配置加密存储（2026-05-13）

已执行（通过）：

- `gofmt -w $(git ls-files '*.go')`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `PORT=8091 CMS_STORE=memory DEVHUB_E2E_TESTING=1 DEVHUB_PLUGIN_CONFIG_KEY=<test_key> ./.devhub/devhub`：启动用于 Playwright 的本机后端；随后执行：
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-config-encryption.spec.js`：通过，`1 passed`。

说明：

- 本轮只涉及后端与后台插件配置链路，不涉及前台与 SEO，因此未执行 `--frontend-only` 与 SEO curl。

## v1.5.0-P1-07 插件安装 / 升级审批流（2026-05-14）

已执行（通过）：

- `gofmt -w internal/domain/plugin_approvals.go internal/service/plugin_approvals.go internal/service/plugin_approval_sanitize.go internal/store/auth.go internal/store/memory.go internal/store/mysql.go internal/store/schema.go internal/transport/httpapi/router.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- 后台 E2E（通过，本轮新增 spec）：
  - `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://127.0.0.1:8091 admin-e2e bash -lc 'cd /workspace && PORT=8091 CMS_STORE=memory DEVHUB_E2E_TESTING=1 ./.devhub/devhub & sleep 0.8; cd /workspace/web/admin-app && npm run test:e2e -- tests/e2e/plugin-approvals.spec.js'`：通过，`1 passed`。

未通过 / 未执行（原因）：

- `./scripts/check-frontend.sh --admin-only`：失败（本环境无法从 `admin-e2e` 容器访问 `host.docker.internal:8090`，且脚本不会自动启动后端；失败日志：`.devhub/checks/20260514-001823/-web-admin-app-E2E.log`）。本轮已用“容器内启动后端 + 指定 `DEVHUB_E2E_ORIGIN`”方式验证新增 `plugin-approvals` E2E。

## v1.5.0-P2-10 已安装插件导出为本地插件包（2026-05-14）

本轮新增声明型插件包导出能力，测试重点是：dry-run 不写文件、正式导出生成 `manifest.json` / `README.md` / 脱敏 `config.example.json` / `checksums.json`，导出后 package dry-run 自检通过，后台导出面板展示安全边界且不提供 zip/远程发布入口。

已执行（阶段性通过，最终完整矩阵以本轮结束记录为准）：

- `gofmt -w internal/domain/plugin_export.go internal/service/plugin_export_service.go internal/service/plugin_package_service.go internal/plugins/package_scanner.go internal/transport/httpapi/plugin_package_handler.go internal/transport/httpapi/router.go`：通过。
- `go test ./internal/service ./internal/transport/httpapi ./internal/plugins`：通过。
- `go test ./internal/service -run 'TestPluginPackageExport' -count=1`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。

本轮最终还需执行：

- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-package-export.spec.js`
- `./scripts/check-frontend.sh --admin-only`

说明：

- 本轮只改后端插件导出 API 与后台插件详情抽屉，不涉及前台页面、前台导航或 SEO 共享逻辑，因此不强制执行 `--frontend-only` 与 SEO curl。
- 导出结果会写入 `storage/plugins/exports/`；该目录仅作为受控本地输出目录，不代表 zip 下载、远程市场发布或插件市场审核。

## v1.5.0 最终收口验证（2026-05-14）

已执行并通过：

- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `./scripts/check-frontend.sh --admin-only`
- `./scripts/check-frontend.sh --frontend-only`

结论：

- 当前版本口径已统一为 `v1.5.0`。
- 插件包治理收口链路（dry-run、checksum、风险报告、仓库扫描、安装、配置历史、敏感配置加密、审批、签名/可信来源草案、导出）与后台 / 前台 / SEO 回归已收口。
- 历史 `v1.4.0` / `v1.3.x` 仅保留追溯，不再作为当前必测项标题。

## 已实现必测

基础检查：

- `go test ./...`
- `go build` 或 `go build -buildvcs=false ./...`
- `cd web/frontend-app && npm run build`
- `cd web/admin-app && npm run build`
- `bash -n dev.sh`

插件 API：

- `plugins.status` 支持扩展状态模型：`discovered`、`installed`、`migrated`、`configured`、`enabled`、`disabled`、`running`、`archived`、`config_invalid`、`migration_pending`、`migration_failed`、`dependency_missing`。
- 发布可用性只认全局 `enabled`；除 `enabled` 外的全局状态均不能新建该插件内容。
- `GET /api/v1/plugins` 只返回全局 enabled 插件。
- `GET /api/v1/plugins` 返回统一 manifest 风格的插件声明结构，包括内容类型、权限、菜单、路由和 `config_schema` 预留字段。
- `GET /api/v1/plugins` 和 `GET /api/v1/communities/:slug/plugins` 不暴露 `config_json` / `resolved_config`。
- `GET /api/v1/admin/plugins` 返回 `qa`、`docs`、`wiki` 的全局状态。
- `GET /api/v1/admin/plugins` 返回轻量 `health` 摘要，至少包含 overall/config/migration/hook/dependency 状态、最近错误和建议操作。
- `GET /api/v1/admin/plugins/:code/audit-logs` 可按插件 code 查询插件启停、配置、Hook 失败和插件内容治理审计。
- `GET /api/v1/admin/plugins/:code/migrations` 可列出内置插件 migration 声明、执行状态、失败原因和 summary。
- `POST /api/v1/admin/plugins/:code/migrations/run` 可执行该插件全部待处理 migration，并写入 `plugin.migration.run` / `plugin.migration.success` / `plugin.migration.failed` 审计。
- `POST /api/v1/admin/plugins/:code/migrations/:name/retry` 可执行或重试单条 migration，并写入 `plugin.migration.retry` / `plugin.migration.success` / `plugin.migration.failed` 审计。
- `POST /api/v1/admin/plugins/:code/migrations/:name/e2e-fail` 仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用于 E2E/API 测试注入 failed migration；注入后全局启用和子站启用必须被后端阻断。
- 已经 `success` 的 migration 再次执行不会重复破坏数据，应返回现有成功记录或保持 success。
- `PUT /api/v1/admin/plugins/:code/config` 可以保存合法 JSON，非法 JSON 应返回 400，并写入审计日志。
- `PUT /api/v1/admin/plugins/:code/config` 缺少 required、enum 非法、type 错误、integer 非整数、数字越界时应返回 400。
- 插件启停、全局配置、子站启停、子站配置和子站排序审计日志应包含 `old_value`、`new_value`、`metadata_json` 结构化字段。
- 配置审计 `metadata_json.changed_keys` 应记录本次变更的顶层配置键。
- `POST /api/v1/admin/plugins/:code/disable` 可以禁用全局插件，并写入审计日志。
- `POST /api/v1/admin/plugins/:code/enable` 可以启用全局插件，并写入审计日志。
- `POST /api/v1/admin/plugins/:code/enable` 启用前会校验配置、依赖和失败迁移；`failed` migration 应阻断启用，当前内置 no-op `pending` migration 不阻断但应通过 health / 迁移 Tab 提示。
- `GET /api/v1/communities/:slug/plugins` 只返回该子站全局 enabled 且子站 enabled 的插件。
- `GET /api/v1/admin/communities/:id/plugins` 返回某个子站的插件状态列表。
- `POST /api/v1/admin/communities/:id/plugins/:code/disable` 可以禁用某个子站插件。
- `POST /api/v1/admin/communities/:id/plugins/:code/enable` 可以启用某个子站插件。
- `PUT /api/v1/admin/communities/:id/plugins/:code/config` 可以保存合法 JSON，非法 JSON 应返回 400，并写入审计日志。
- `PUT /api/v1/admin/communities/:id/plugins/:code/config` 同样必须执行后端 `config_schema` 强校验，不能只依赖前端 Ajv。
- `PUT /api/v1/admin/communities/:id/plugins/sort` 可以调整排序，并写入审计日志。
- 全局 disabled / `config_invalid` / `migration_pending` 插件不能被子站启用。
- 子站启用插件时也应复用 Service 层 readiness 检查；当插件存在 `failed` migration 时，子站启用应失败。
- `GET /api/v1/moderator/plugin-menus` 只返回全局 enabled、子站 enabled 且当前用户有权限的插件菜单。
- `GET /api/v1/admin/plugins/:code/impact` 返回历史内容数、启用/禁用子站数、绑定板块数、近 7 天内容数、审核中内容数、菜单声明数、配置覆盖数和待执行迁移数。
- `GET /api/v1/admin/communities/:id/plugins/:code/impact` 返回同类字段，但内容、板块和子站状态计数应收敛到该子站范围。
- 插件 manifest 结构一致性测试覆盖 `code/name/version/is_system/content_types/permissions/menus/routes/config_schema/min_core_version`。
- `doc -> document`、`wiki -> wiki_page` 和 `content_type -> plugin_code` 映射测试应通过。
- 插件配置合并测试覆盖 schema 默认值、全局配置、子站配置和 `effective` 生效值。
- 插件生成模板：`go run ./cmd/devhub plugin:new ...` 应生成 `manifest.json`、`README.md`、`config.example.json`、`docs/*.md` 和可选 `migrations/001_init.sql`；不生成 `registry.example.go`、根目录 `001_schema.sql`、Go/JS/WASM/binary 文件、package scripts 或 blocking Hook 模板。生成的 manifest 必须通过 `PluginManifestValidator`，示例配置必须通过当前简化 `config_schema` 校验。

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
- 全局 `config_invalid` / `migration_pending` 状态下，后端强传对应 `content_type` 也应拒绝创建。
- 板块不能绑定当前子站未启用的插件。
- `category.plugin_code` 与 `content_type` 对应插件不匹配时拒绝发布。
- `content_type` 不在 `category.allowed_content_types` 内时拒绝发布。
- 采纳回答后，`GET /api/v1/topics/:id/qa` 中的 `is_resolved`、`accepted_answer_id` 和回答接受状态会更新。
- 编辑 `wiki_page` 后，`GET /api/v1/topics/:id/wiki/versions` 返回的版本数应增加。

## 插件系统专项验收与 E2E 回归归档（2026-05-11）

本轮执行了插件系统专项验收，目标是确认当前插件平台已经具备可治理、可配置、可启停、可审计、可观测、可迁移的基础形态，并把 E2E 回归结果归档为长期状态来源。

已自动化并通过：

- 后端基础：`go test ./...`、`go build -o .devhub/devhub .`、`bash -n dev.sh`。
- 后台构建：`docker compose run --rm admin-e2e npm run build` 通过；Vite chunk size warning 不作为失败。
- 前台 E2E：`./scripts/check-frontend.sh --frontend-only --e2e-only` 通过，14 passed。
- 后台 E2E：`./scripts/check-frontend.sh --admin-only --e2e-only` 首次发现测试污染和旧文案断言，修复后通过，15 passed。
- SEO curl：`/topics/1/` 可见 title、canonical、Article JSON-LD、article、h1；`/c/php/` 可见 title、description、canonical、h1 和结构化数据。

本轮 E2E 已覆盖：

- 插件启停：后台插件治理 E2E 覆盖全局禁用确认、影响分析提示、全局 disabled 时子站不能启用，并在测试结束恢复 `qa` enabled。
- content_type 强校验：前台插件联动 E2E 覆盖禁用插件后发布入口不可用、后端强传禁用 content_type 被拒绝、非法 content_type 被拒绝。
- 历史内容与 SEO：前台 E2E 和 curl 均确认禁用插件不影响历史 Topic 动态详情基础 SEO。
- 权限边界：前台 E2E 覆盖普通用户不能访问版主工作台、版主只能访问授权页面，并包含跨子站 API 越权拦截。
- 插件配置：后台 E2E 覆盖插件详情配置 Tab、JSON Editor / Ajv 错误提示、子站插件配置非法 schema 值拦截。
- 后台插件中心：后台 E2E 覆盖插件列表筛选、详情 Tabs、影响分析、审计入口、子站插件配置、通用 PluginContent 页。
- 后台细操作：后台 E2E 覆盖 reports 处理与审计、moderators 创建/更新/禁用 API 反映到列表、PluginContent 内容治理和审计。

本轮发现并修复：

- 后台插件治理 E2E 中，全局禁用 `qa` 的测试失败时没有恢复插件状态，导致并行运行的插件内容页测试被污染。已将该 describe 设为 serial，并在 `finally` 中恢复 `qa` enabled。
- 影响分析弹窗已经展示“当前启用子站 / 将阻止发布的板块 / 已有历史内容”等真实字段，旧断言仍查找“将影响子站”。已同步测试断言。

仍需后续补测：

- 插件迁移：失败注入、重试成功、重复执行不破坏数据、审计定位的完整 E2E。
- Hook 治理：blocking Hook 人为失败并阻断主流程、non-blocking Hook 失败不阻断但记录后台可见的完整 E2E。
- 插件审计：按 plugin_code / community_id / action 的更多组合筛选，以及 old_value / new_value / metadata_json 的浏览器断言。
- 插件健康状态：migration_pending、config_invalid、hook warning/error 等多状态 UI 回归。

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
- `/admin-next/plugins` 全局禁用确认会展示 impact 计数；Hook 错误统计来自 `hook_executions` 最近 7 天失败记录，不应伪造成完整健康评分。
- `/admin-next/communities` 的插件配置抽屉支持启用 / 禁用、全局 / 子站双状态、全局禁用原因、`config_json` 编辑、schema 参考、JSON Editor、Ajv 基础校验和数字排序保存。
- `/admin-next/communities` 子站禁用确认会展示该子站范围内的 impact 计数。
- `/moderator` 可按当前子站显示插件治理入口；全局 disabled 或子站 disabled 插件不显示。

HookBus 最小检查：

- 创建 Topic 时触发 `BeforeCreateContent` 和 `AfterCreateContent`。
- `BeforeCreateContent` blocking hook 返回错误时阻断创建，并写入 `hook_executions` 与 `plugin.hook.blocked` 审计。
- `AfterCreateContent` non-blocking hook 返回错误时不阻断主流程，并写入 `hook_executions` 与 `plugin.hook.failed` 审计。
- `POST /api/v1/admin/plugins/:code/hooks/:name/e2e-fail` 仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用于 E2E/API 测试注入 Hook 失败；自动化覆盖 blocking 失败阻断创建、non-blocking 失败不阻断创建、执行记录查询、审计定位和后台 Hooks Tab 失败摘要。
- 插件详情 Hooks Tab 能展示执行次数、失败次数、失败率、平均耗时、最近执行、最近失败和最近错误。
- 插件详情“运行状态”Tab 能展示 overall、config_status、migration_status、hook_status、dependency_status、pending/failed migrations、hook failures、recent_error 和 suggested_action。
- 插件详情“迁移”Tab 能展示 migration 列表、状态、最近执行时间、失败原因、rollback_supported 标识，并提供执行/重试入口。
- 插件详情“审计”Tab 使用插件专用审计接口，不再只依赖通用 audit-log target 前缀。
- 插件内容治理操作（隐藏/恢复、置顶/取消置顶、加精/取消加精、评论治理等已有操作）应写入带 `plugin_code` 的结构化审计 metadata；Core 内容可继续使用普通审计。
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

插件前台入口 / 菜单可见性（v1.4.0-P1-11）：

- 后端：`go test ./...`（至少覆盖 `internal/transport/httpapi/plugin_navigation_test.go` 中的 navigation/create-options/menus preview 鉴权）
- 后台 E2E：`docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-navigation-admin.spec.js`
- 前台 E2E：`docker compose run --rm frontend-e2e npm run test:e2e -- tests/e2e/plugin-navigation.spec.js`
- SEO 回归：

```bash
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1|<article|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/c/php/ | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'
```

## 后台 E2E Docker Runner

## 前后台前端统一检查脚本

推荐使用 `scripts/check-frontend.sh` 作为前台与后台构建 / E2E 的统一入口。脚本默认使用 Docker Compose runner，不依赖宿主机 Node/npm，并把日志写入 `.devhub/checks/{时间戳}/`。

常用命令：

```bash
./scripts/check-frontend.sh
./scripts/check-frontend.sh --target admin
./scripts/check-frontend.sh --target frontend
./scripts/check-frontend.sh --target both
./scripts/check-frontend.sh --quick
./scripts/check-frontend.sh --strict
./scripts/check-frontend.sh --quiet
```

目标选择：

- `--target admin`：只检查后台 `web/admin-app`。
- `--target frontend`：只检查前台 `web/frontend-app`。
- `--target both`：同时检查前台与后台。
- `--admin-only` 等价于 `--target admin`，`--frontend-only` 等价于 `--target frontend`。
- 在交互式终端直接运行且没有传 target 时，脚本会询问检查范围；在非交互环境默认检查 `both`，避免 CI 卡住。

模式说明：

- 默认：build + E2E。
- `--quick` / `--build-only`：只跑 build。
- `--e2e-only`：只跑 E2E。
- `--strict`：如果 `package.json` 存在 `lint` / `typecheck` 脚本，则额外执行；不存在则显示 `SKIP`，不视为失败。
- `--rebuild`：执行前先构建对应 e2e 镜像。
- `--remove-orphans`：传给 `docker compose run`，用于清理 orphan container 提示。
- 默认实时显示完整日志，同时落盘；`--quiet` 可只显示摘要和失败日志尾部。
- `--tail-lines N`：失败时展示最后 N 行日志。

结果含义：

- `PASS`：必要检查通过。
- `FAIL`：必要检查失败，脚本最终 `exit 1`。
- `SKIP`：compose 服务或 package script 不存在，按可选项跳过；例如未来某 runner 暂未创建时不会误报失败。

`admin-e2e` 与 `frontend-e2e` 关系：

- `admin-e2e` 用于后台构建和后台 Playwright E2E。
- `frontend-e2e` 用于前台构建和前台 Playwright E2E。
- 两者都由 `scripts/check-frontend.sh` 调度；也可以继续直接使用下方单独 Docker Compose 命令。

后台 Playwright E2E 使用项目内固定 Docker 镜像，不依赖宿主机 Node/npm：

```bash
docker compose build admin-e2e
docker compose run --rm admin-e2e npm run build
docker compose run --rm admin-e2e
```

v1.4.0-P1-10 增量用例：

```bash
docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-readiness-errors.spec.js
```

说明：

- 该用例依赖 `GET /api/v1/admin/plugins/:code/readiness`；如果当前环境仍在运行旧后端二进制且未包含 readiness 路由，会在用例内探测并 `test.skip`（避免把“后端未升级”误判为前端回归）。

本轮 Hook 排障页专项最小回归：

```bash
docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-hooks.spec.js
```

本轮插件 SDK / 模板专项最小回归：

```bash
go test ./internal/plugins/scaffold
go run ./cmd/devhub plugin:new --code smoke_links --name "Smoke Links" --content_type smoke_link --content_name "Smoke Link" --output .devhub/tmp/plugin-template-smoke --with_config --with_hooks --with_migration --force
```

2026-05-12 执行结果：上述 `go test ./internal/plugins/scaffold` 与 smoke test 均已通过；同时 `gofmt`、`go test ./...`、`go build -o .devhub/devhub .`、`git diff --check`、`bash -n dev.sh` 和 `bash -n scripts/check-frontend.sh` 通过。本轮未修改后台页面或前台 SEO，未执行后台 Docker build / E2E 与 SEO curl。

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
- `/admin-next/reports` 举报管理打开、状态/对象筛选、测试举报处理和审计日志联动。
- `/admin-next/moderators` 版主管理打开、列表展示、通过 API 验证新增/更新/停用版主并在 UI 列表中反映。
- `/admin-next/plugins` 打开与搜索筛选。
- 插件详情 Tabs、`config_schema` / `resolved_config` 展示和 schema 错误提示。
- 全局禁用确认、impact 提示和全局 disabled 后子站启用限制。
- 子站插件配置抽屉、JSON Editor 与 Ajv 错误提示。
- 通用 `PluginContent` 页入口、子站筛选和状态筛选。
- `/admin-next/qa`、`/admin-next/docs`、`/admin-next/wiki` 插件内容页打开、筛选和通用内容表展示。
- 插件内容治理代表链路：通过后台 API 对已有插件内容执行隐藏 / 恢复，并在审计日志中验证记录。
- 前台 seed 用户 `liuwei / 方圆十三 / a123456` 会在 MemoryStore 和 MySQLStore 初始化时自动写入，可直接用于前台登录、发布和手工冒烟。
- 前台种子用户 `liuwei / 方圆十三 / a123456` 会在 MemoryStore 和 MySQLStore 初始化时自动写入，可直接用于手工登录、前台发布和前台用户 E2E。

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
- `frontend_e2e_node_modules` 命名卷挂载到 `/workspace/web/frontend-app/node_modules`；`web/frontend-app/docker/e2e-entrypoint.sh` 只在卷缺少依赖时从镜像层复制 `/app/node_modules/.` 的内容，避免出现 `node_modules/node_modules` 嵌套。
- “不依赖宿主机 Node/npm”指依赖安装、构建和测试均在 Docker 镜像或容器内完成；项目仍然通过 `package.json` / `package-lock.json` 声明 npm 依赖。
- 默认测试目标是 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8090`，运行 E2E 前需要先启动 DevHub 后端服务。
- E2E 前建议先执行 `docker compose run --rm frontend-e2e npm run build`，让 Go 服务读取最新 `web/frontend` 静态产物。

当前已自动化覆盖：

- `/` 总站首页打开，基础导航可见，游客看不到总后台入口。
- `/c/php/` 和 `/c/go/` 子站首页打开，并检查 canonical 基础 SEO 元素。
- `/search/` 关键词搜索提交，结果区域可见。
- `/topics/1/` 动态 Topic 详情打开，包含 h1、article 和 JSON-LD。
- `/topics/new/` 未登录发布拦截，提交后提示登录。
- 登录用户 Topic 详情互动：点赞、收藏、关注主题和发表评论。
- 用户中心联动：`/notifications`、`/me/activities`、`/me/favorites`、`/me/follows` 登录态访问。
- 登录发布流程：必填校验、发布 `article` 成功并打开新详情页。
- 插件状态联动：临时启用 QA 板块后验证 `question` 可选；全局禁用 `qa` 后发布页隐藏入口，API 强传 `question` 被拒绝，非法 `content_type` 被拒绝，历史 `/topics/:id` SEO 仍可访问；测试结束恢复插件和板块状态。
- 插件迁移异常联动：后台 E2E 通过测试 helper 注入 `qa_questions` failed migration，验证全局启用和子站启用被后端阻断，迁移 Tab 能看到失败原因，retry 后迁移变为 success，插件恢复启用，审计可查到 `plugin.migration.failed/retry/success`。
- 版主工作台边界：普通前台用户访问 `/moderator*` 无权限；PHP 版主可访问授权页面；API 强传 Go 子站 `community_id` 被 403 拦截。
- `/tags/go/` 与 `/c/php/tags/laravel/` 标签页打开，并检查 canonical 基础 SEO 元素。

## 已实现但后续补测

- 插件全局 `config_json` 和子站 `config_json` 接口已有自动化覆盖，仍需要真实 admin token 做联调补测。
- 插件声明里的 `config_schema`、`dependencies`、`min_core_version` 和 `hooks` 已有结构测试；`config_schema` 已接入简化后端校验，仍需要继续补测前后台展示、错误提示、边界值和完整 JSON Schema 不支持场景。
- 子站插件排序接口已有自动化覆盖，仍需要真实 admin token 和浏览器做联调补测。
- 后台插件治理中心和核心后台页面已有 Docker 化 Playwright E2E runner；仍需继续补多浏览器、更多真实 UI 操作、视觉细节和 MySQLStore 场景。
- 后台版主管理新增 / 编辑 / 停用当前通过真实 API + UI 列表验证组合覆盖；Element Plus 表单纯 UI 操作仍可作为后续更细浏览器用例。
- 后台插件内容页当前覆盖通用页打开、筛选和 API 治理代表链路；更完整的插件内容详情抽屉、审核按钮 UI 闭环仍待后续。
- 前台发布页按全局插件状态和当前板块状态已有浏览器覆盖；仍需继续补“子站 A 禁用 / 子站 B 启用”的跨子站矩阵。
- 子站导航按插件状态显示仍需继续做多子站浏览器验收。
- 版主菜单按子站插件状态和权限过滤已有 PHP 版主基础覆盖；仍需继续补多版主账号、多子站插件启停组合和 UI 操作。
- MySQL 老库插件平台结构与核心行为已通过专项验证；仍需在接入真实生产数据前做预发库备份 / 回滚演练和历史内容 SEO 的 MySQL 端浏览器矩阵。

## 待实现后补测

- 插件权限矩阵后续补测：缺少对应 create 权限时前台发布 / 后台创建 / 版主菜单均被拒绝或隐藏的更多角色组合。
- 更细粒度的权限体系补测：例如 Core 兼容类型 `article` / `news` 的细分权限码、按子站/板块维度配置权限矩阵与更明确的错误码（当前发布链路已实现最小权限码校验）。
- Projects / Jobs / AI Works 的专属扩展表、专属管理页和完整业务流程。
- P0：HookBus 的完整业务处理器、关键 Hook 事务回滚、非关键 Hook 统一错误日志和重试策略。
- P0：`config_schema` 更多边界值、错误提示和完整浏览器矩阵。
- P1：`config_schema` 配置表单增强，包括深层嵌套、字段分组、复杂数组和完整 JSON Schema 不支持场景。
- Docs 文档树专用编辑 UI、拖拽排序和批量排序。
- Wiki 版本回滚和协作编辑交互。
- QA 取消采纳最佳答案。

## 完整插件系统平台验收矩阵

基线说明：

- 当前插件平台是内置系统插件平台，不是第三方插件市场或动态插件加载环境。
- v1.3.2 起 `discovered`、`migrated`、`configured`、`running`、`config_invalid`、`migration_pending`、`dependency_missing` 已进入全局状态枚举；但完整自动状态机仍未完成，当前发布判定仍只认 `enabled`。
- `plugin_migrations` 当前支持内置 up/no-op runner、失败记录和失败重试；migration down、真实 rollback、迁移前备份和外部插件迁移包仍需后续补测。
- HookBus 当前可派发并执行内置 handler；`hook_executions` 已记录执行结果、失败统计、最近错误和平均耗时，后台 Hooks Tab 已展示这些运行时字段。插件列表/详情已展示轻量健康摘要；独立健康 API、告警和重试策略仍需后续补测。

P0 已实现或必测：

- Manifest 字段一致性：`code/name/version/is_system/content_types/permissions/menus/routes/config_schema/min_core_version/hooks`。
- `content_type -> plugin_code` 映射：`question/docs/wiki_page/project/job/ai_work` 均映射到对应插件，`doc/wiki` 能归一。
- 插件全局 enabled / disabled：全局 disabled 后不能新发布对应内容，历史详情仍可访问。
- 子站 enabled / disabled：子站 disabled 后仅该子站不能新发布对应内容，其他子站不受影响。
- 板块绑定：`categories.plugin_code` 与 `allowed_content_types` 不匹配时拒绝发布。
- 权限码校验：发布时使用 ContentTypeDefinition 中的 create_permission。
- ActorContext 来源可信：由服务端 token / admin / moderator scope 计算，客户端请求体不能伪造。
- 菜单过滤：前台、后台、版主菜单按全局状态、子站状态、权限和 scope 过滤。
- `config_json` 合法性和简化 `config_schema` 校验：非法 JSON、缺少 required、enum 非法、类型错误、integer 非整数和数字越界应保存失败。
- `resolved_config.effective`：schema 默认值 < 全局配置 < 子站配置。
- `admin_logs` old/new/metadata：插件启停、配置、排序写入 `old_value`、`new_value`、`metadata_json`，配置变更包含 `changed_keys`。
- disabled 插件历史内容访问：`/topics/:id` 不返回 404。
- `/topics/:id` SEO 动态 HTML：title、description、h1、article、标签、JSON-LD 不丢失。
- migration 新装 / 老库升级：`001_schema.sql`、`internal/store/schema.go` 和 migrations 字段口径一致。
- v1.3.4 MySQLStore 专项：可选集成测试已覆盖 `plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、`admin_logs` 新装结构，`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 升级字段，全局 / 子站插件启停强拦截、failed migration 阻断与 retry、Hook 记录、插件审计查询和 config_schema 校验。
- v1.3.4 插件迁移异常链路：failed migration 阻断全局启用和子站启用，retry 成功后恢复启用，已 success 的 migration 不重复破坏数据，审计可定位，后台迁移 Tab 可见失败原因。
- v1.3.4 插件权限矩阵链路：`ContentTypeDefinition.create_permission` 已覆盖 question/document/wiki_page/project/job/ai_work/core；`post.create` 只桥接 `core.topic.create`，不能创建插件内容；普通前台 token 不能调用插件治理 API；版主插件菜单继续受全局状态、子站状态、community scope 和插件权限过滤。

P0 待实现 / 待补测：

- HookBus 更多事件矩阵：v1.3.4 已自动化覆盖创建链路的 blocking / non-blocking 失败注入、`hook_executions`、审计和后台 Hooks Tab；更新 / 删除 / Search / Notification / SEO 仍需补更多异常矩阵。
- 权限矩阵后续补测：后台插件内容治理的细粒度操作权限、完整 RBAC 分配 UI、community / category 级权限配置仍需后续覆盖。
- MySQLStore 后续矩阵：插件平台结构和核心 Service 行为已自动化验证；仍需补真实 MySQL 服务下 `/topics/:id`、`/c/:slug` SEO 浏览器 / curl 矩阵和生产大库升级演练。
- `config_schema` 浏览器矩阵：后端强校验已接入，仍需补更多真实浏览器错误提示、深层 diff 和完整 JSON Schema 不支持场景。
- HookBus 业务处理器：Create / Update / Delete / Search / Notification / SEO 不仅能派发事件，还要继续补更多插件处理器、告警和重试策略验收。
- 插件 migration runner：当前已支持内置插件 up/no-op 查询、执行、失败记录、重试和后台迁移 Tab；真实 rollback、down migration、迁移前备份仍待后续。
- 完整真实 token 验收矩阵：全局禁用、子站禁用、跨子站发布、版主菜单、历史 SEO。

P1 / P2 / P3 后续验收：

- P1：schema 自动表单增强、插件 SDK 文档、插件生成模板、依赖检查、版本兼容检查、插件搜索 / 通知 / SEO 扩展。
- P2：本地插件包 zip、插件包安装向导、插件升级向导增强、外部服务型 Webhook、插件 migration runner、插件包签名校验、插件市场雏形。

## v1.3.4 插件异常治理测试矩阵收口

本节用于归档 v1.3.4 的真实覆盖状态。分类含义：

- 已自动化：已有 Go 测试或 Playwright E2E，并在本阶段执行过或纳入最低检查。
- 部分自动化：已有 API / 单元测试或单一路径 E2E，但未覆盖完整角色 / Store / 浏览器矩阵。
- 手工验证：已有明确命令或页面步骤，但本阶段未完整自动化。
- 未覆盖：尚无稳定测试保护，需后续补测。
- 跳过项：本阶段明确不做或缺少环境。

已自动化：

- 插件迁移失败注入与启用阻断：API 测试和后台 E2E 覆盖 failed migration、全局启用阻断、子站启用阻断、retry 恢复、success 不重复破坏和审计定位。
- HookBus 异常链路：API 测试和后台 E2E 覆盖 `BeforeCreateContent` blocking 失败阻断、`AfterCreateContent` non-blocking 失败不阻断、`hook_executions` 查询、审计定位和 Hooks Tab 失败摘要。
- 插件权限矩阵：Go/API 测试覆盖 `question/document/wiki_page/project/job/ai_work/article/news` 的 create permission 映射、`post.create` 只能桥接 `core.topic.create`、普通前台 token 不能访问插件治理 API。
- MySQLStore 专项：可选集成测试覆盖新装结构、升级字段、全局/子站启停强拦截、failed migration 阻断与 retry、Hook 记录、插件治理审计查询和 `config_schema` 校验。
- 插件健康状态：Go 测试覆盖 `healthy`、`config_invalid`、failed migration `error`、Hook 多次失败 `hook_error`。
- 审计筛选：Go 测试覆盖 `plugin_code`、`action`、`community_id`、`metadata`、`request_id` 组合筛选；后台 E2E 覆盖插件详情审计 Tab 的 action/community/metadata 查询。

部分自动化：

- `hook_warning` / `hook_error` 前端 UI：后台 E2E 覆盖 Hook 失败后运行状态可见；更多插件、多 Hook 名称、多时间范围组合仍待补。
- 插件审计筛选 UI：已覆盖插件详情抽屉中的核心筛选；通用 `/admin-next/audit-logs` 的 target_type/target_id/request_id/time range 浏览器矩阵仍待补。
- `config_invalid` 状态：后端拒绝非法配置并有 Go 测试覆盖显式状态；常规 UI/API 不会持久化非法配置，因此“配置无效持久状态”的浏览器场景仍属于治理预留。
- MySQLStore 浏览器矩阵：Service 层 MySQL 集成测试已覆盖核心行为；真实 MySQL 服务下 `/topics/:id`、`/c/:slug` SEO 浏览器 / curl 矩阵仍待补。

手工验证：

- 生产大库升级前备份、预发执行 `004`-`010`、启动应用、抽样验证历史内容和回滚预案。
- 插件禁用后前台导航、发布页和版主菜单在多子站、多账号组合下的视觉矩阵。
- 后台插件详情的迁移、Hooks、审计、运行状态 Tabs 在慢网和大日志量下的可读性。

未覆盖：

- HookBus Update / Delete / Search / Notification / SEO 的异常注入矩阵。
- Hook 重试、告警、自动恢复和外部监控。
- 插件内容治理批量操作权限矩阵和完整 RBAC 分配 UI。
- 深层 config diff、配置版本、配置回滚和自动表单复杂字段矩阵。
- 插件包 zip、完整安装 / 升级向导、外部服务型 Webhook、市场和动态加载。

跳过项及原因：

- 插件市场、插件上传、远程安装、Go 动态加载和第三方沙箱：不属于 v1.3.4 范围。
- migration down、硬回滚、迁移前自动备份：当前 runner 是内置 up/no-op 记录型 runner，后续 P2/P3 再设计。
- Projects / Jobs / AI Works 专属业务闭环：当前只验收平台治理归属，不补专属扩展表或专属页面。

## MySQLStore / 老库升级专项命令

用途：验证插件平台在真实 MySQLStore 和老库显式 SQL 升级路径下与 MemoryStore 口径一致。测试必须使用一次性测试库，库名建议包含 `test`。

准备 MySQL 测试库：

```bash
docker compose -f docker-compose.dev.yml up -d mysql
docker compose -f docker-compose.dev.yml exec -T mysql mysql -uroot -pDevhub_root_123456 -e "DROP DATABASE IF EXISTS devhub_test; CREATE DATABASE devhub_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL PRIVILEGES ON devhub_test.* TO 'devhub'@'%'; FLUSH PRIVILEGES;"
```

运行 MySQLStore 插件平台一致性测试：

```bash
DEVHUB_MYSQL_TESTS=1 DB_HOST=127.0.0.1 DB_PORT=3307 DB_USER=devhub DB_PASSWORD=Devhub_123456 DB_NAME=devhub_test go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v
```

验证显式插件迁移 SQL 可重复执行：

```bash
docker compose -f docker-compose.dev.yml exec -T mysql mysql -uroot -pDevhub_root_123456 -e "DROP DATABASE IF EXISTS devhub_upgrade_test; CREATE DATABASE devhub_upgrade_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL PRIVILEGES ON devhub_upgrade_test.* TO 'devhub'@'%'; FLUSH PRIVILEGES;"
docker compose -f docker-compose.dev.yml exec -T mysql mysql -udevhub -pDevhub_123456 devhub_upgrade_test < db/mysql/001_schema.sql
for round in 1 2; do
  for f in db/mysql/migrations/004_community_plugins.sql db/mysql/migrations/005_core_plugins.sql db/mysql/migrations/006_plugin_config_json.sql db/mysql/migrations/007_admin_logs_structured_plugin_audit.sql db/mysql/migrations/008_plugin_migrations.sql db/mysql/migrations/009_plugin_status_model.sql db/mysql/migrations/010_hook_executions.sql; do
    docker compose -f docker-compose.dev.yml exec -T mysql mysql -udevhub -pDevhub_123456 devhub_upgrade_test < "$f"
  done
done
```

已验证范围：

- 新装结构：`plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、`admin_logs`。
- 升级字段：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types`。
- 行为一致性：全局禁用强拦截、子站禁用仅影响当前子站、failed migration 阻断启用、retry 恢复、Hook 记录查询、插件治理审计查询、配置 schema 校验。

注意：

- 该测试默认跳过，必须显式设置 `DEVHUB_MYSQL_TESTS=1`；测试代码还会拒绝在库名不含 `test` 的数据库上运行。
- 生产升级仍需备份、预发演练和回滚预案；当前内置 plugin migration 仍是 up/no-op 记录型 runner，不支持 migration down 或硬回滚。
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
   - 点击“清空为空对象”把配置重置为空对象；
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
9. JSON 辅助操作：点击“清空为空对象”应将配置置为空对象；然后点击保存应写入后端。
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
   - 插件名、plugin_code、插件状态、健康状态和内容类型数量；
   - disabled / archived 时提示“不能新建，但历史内容仍可治理”。
3. 子站筛选：
   - 选择一个子站后，列表应切换到该子站内容（通过 `admin/posts?site=<slug>&plugin_code=<code>&content_type=<type>`）。
4. 状态筛选：
   - 切换 publish/offline/pending，列表筛选应生效。
5. 列表中展示后端返回的 `plugin_code` / `content_type` 字段，且不同插件内容不会因前端 OR 过滤混入。
6. 批量治理：
   - 批量隐藏 / 恢复不退化；
   - 批量审核通过 / 拒绝、置顶 / 取消置顶、加精 / 取消加精可触发既有后台 topic 治理能力；
   - 操作后显示成功 / 失败明细，并可跳转审计日志。

## 阶段 B 插件治理体验手工验收

本节用于验收“阶段 B：引入 i18n，统一优化插件治理体验”。

已自动化或构建覆盖：

- 后台构建会覆盖 i18n 挂载、插件详情抽屉、配置编辑器和 PluginContent 编译可用性。
- 既有后台 E2E 继续覆盖 `/admin-next/plugins`、插件详情 Tabs、JSON/Ajv 校验、禁用确认、子站插件配置和 PluginContent 入口。
- 2026-05-11 复查截图暴露的漏网英文后，已补齐插件详情概览 / 运行状态 / 内容类型 / Hook / 迁移 / 路由 / 审计 Tab，以及子站插件配置弹窗中的 `config_schema`、`config_json`、`resolved_config` 标签中文化；后台 E2E 同步断言当前中文文案并通过 `18 passed`。
- 本轮命令结果：`go test ./...`、`go build -o .devhub/devhub .`、`bash -n dev.sh`、`bash -n scripts/check-frontend.sh`、`./scripts/check-frontend.sh --quick`、`./scripts/check-frontend.sh --admin-only --e2e-only` 均通过；`--quick` 最新日志目录 `.devhub/checks/20260511-225223/`，后台 E2E 最新日志目录 `.devhub/checks/20260511-225516/`。

需要手工确认：

1. 后台插件中心主要用户可见英文应改为中文：状态、健康状态、配置模型、Hook、迁移、审计、按钮和筛选项均应显示中文标签。
2. `plugin_code`、`content_type`、`hook_name`、权限码和 JSON key 仍保持原始技术值，不翻译后提交给后端。
3. 插件详情 -> 配置 Tab：
   - 能在“表单模式”和“JSON 高级模式”之间切换；
   - 表单模式能渲染 string / number / integer / boolean / array / object / enum 基础字段；
   - JSON 高级模式仍可编辑原始 JSON；
   - schema 校验错误可见，保存按钮应受错误状态限制；
   - 配置差异能展示原配置、新配置和变更字段；
   - 敏感字段在差异预览中脱敏；
   - 最终生效配置可见。
4. 子站插件配置弹窗：
   - 继续展示全局状态和子站状态；
   - 配置编辑器同样支持表单 / JSON 模式、差异和最终生效配置预览；
   - 子站配置覆盖全局配置的提示仍可见。
   - `配置模型`、`子站配置`、`最终生效配置` 等标签应显示中文；JSON key 和插件技术值可以保留英文。
5. PluginContent：
   - 支持子站、状态、关键词和内容类型筛选；
   - 列表展示插件编码、内容类型、子站、状态、更新时间和评论数；
   - 多选后可批量隐藏 / 恢复、审核通过 / 拒绝、置顶 / 取消置顶、加精 / 取消加精；
   - 批量操作前有确认并显示影响数量；
   - 操作后刷新列表，展示成功 / 失败明细，并提供“查看审计日志”入口；
   - “查看审计日志”会进入 `/admin-next/audit-logs`，并预填 `plugin_code`、`content_type`、`action`、`target_type` 和 metadata 筛选；
   - 详情抽屉可展示当前内容的基础治理信息。

仍需后续补测：

- 后台全站英文文案扫描与非插件页面中文化。
- PluginContent 更完整的跨页面审计高亮和完整权限矩阵。
- 复杂 `config_schema` 嵌套对象、数组对象和字段分组渲染。

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

## 阶段 C/D/E/F 插件生命周期与软卸载验收

本节用于阶段 C/D/E/F 的最小验收，不扩展成大规模 E2E。

已自动化：

- Go 单测覆盖内置插件归档后 `plugins.status=archived`、`lifecycle_status=archived`。
- Go 单测覆盖归档插件不能通过 `ValidateTopicPluginAccess` 新建内容。
- Go 单测覆盖归档插件不能被子站启用。
- Go 单测覆盖恢复插件后默认进入 `disabled`，不会自动启用。
- Go 单测覆盖 failed migration 会阻断恢复。
- API 测试覆盖 `POST /api/v1/admin/plugins/:code/archive`、`restore`、归档后创建拦截、子站启用拦截、恢复后再启用和审计定位。
- 前台 Playwright E2E 覆盖归档 `qa` 后发布页不再展示 `question`，对应问答板块不再可选。
- 前台 Playwright E2E 覆盖强传归档插件 `content_type=question` 被后端拒绝。
- 前台 Playwright E2E 覆盖子站启用归档插件被后端拒绝。
- 前台 Playwright E2E 覆盖归档插件后历史 `/topics/2/` 仍可访问，`h1`、`article` 和动态 SEO 基础元素存在。
- 后台 Playwright E2E 覆盖归档确认弹窗影响范围、归档 badge、详情归档时间和恢复后默认 `disabled` 提示。
- 后台 Playwright E2E 覆盖归档插件历史内容仍可进入 PluginContent 查看，并显示只能治理历史内容、不能新建的提示。
- 后台 Playwright E2E 覆盖归档态 PluginContent 历史内容批量隐藏 / 恢复仍按后台权限可用。

手工 / 轻量冒烟：

- 后台插件列表应显示生命周期字段和状态原因。
- 插件详情审计 Tab 可查 `plugin.archived` / `plugin.restored`。

未覆盖 / 后续：

- 生产 MySQL 大库归档 / 恢复耗时专项。
- 归档后所有插件、所有子站导航入口的完整浏览器矩阵；当前自动化以 `qa/question` 为代表路径。
- PluginContent 归档态自动化覆盖批量隐藏 / 恢复，并以置顶 / 取消置顶覆盖新增批量治理最小链路；完整只读策略和权限矩阵仍待补。
- 外部插件包安装、上传、远程安装、Go 动态加载、第三方沙箱和硬卸载均未实现，不能写成已验收。
- P2：后台构建存在 Vite chunk size warning，主要来自 `PluginConfigEditor` 等大 chunk；后续可考虑按需加载或手动拆包。

2026-05-12 归档态专项执行结果：

- `./scripts/check-frontend.sh --frontend-only`：通过，前台 build 通过，前台 E2E `16 passed`；覆盖归档插件入口隐藏、强传拦截、子站启用阻断和历史 Topic SEO 回归。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 E2E `20 passed`；覆盖归档插件治理中心细节、PluginContent 历史治理提示和归档态批量隐藏 / 恢复。

## v1.7.0-P0-07 软卸载任务（plugin_uninstall_tasks）轻量验收清单

1. 远程/包安装插件可执行 `POST /api/v1/admin/plugins/:code/soft-uninstall`。
2. 内置系统插件返回 `plugin_soft_uninstall_system_forbidden`。
3. 存在 enabled 的 required 依赖插件时返回 `plugin_soft_uninstall_dependency_blocked`。
4. 仅 optional 依赖存在时允许软卸载，但返回 warnings。
5. `GET /api/v1/admin/plugins/:code/uninstall-impact` 能返回 impact 摘要与 dependents 列表。
6. 软卸载成功后插件全局状态变为 `archived`。
7. 软卸载成功后不能新建该插件 content_type（后端强校验阻断）。
8. 软卸载成功后菜单/入口隐藏（按 status!=enabled gating）。
9. 软卸载成功后 HookBus 不再执行该插件 Hook（按 plugin status gating）。
10. 历史内容详情仍可访问；`/topics/:id` 和 `/c/:slug` SEO 不受影响。
11. `GET /api/v1/admin/plugins/uninstall-tasks` 可查询任务列表。
12. `GET /api/v1/admin/plugins/uninstall-tasks/:id` 可查询任务详情。
13. 失败任务可 `POST /api/v1/admin/plugins/uninstall-tasks/:id/retry` 重试。
14. `DELETE /api/v1/admin/plugins/uninstall-tasks/:id` 标记 deleted，不删除插件文件/配置/数据。
15. 审计日志包含：`plugin.soft_uninstall.requested/started/success/failed`、`plugin.runtime.unregistered`。

## v1.7.0-P0-08 升级任务（plugin_upgrade_tasks）轻量验收清单

1. 同 `plugin_code` 的更高版本包（`compat-check.can_install=true`）可以进入升级影响分析：`GET /api/v1/admin/plugins/:code/upgrade-impact?target_compat_check_id=...`。
2. 不同 `plugin_code` 的包返回 `plugin_upgrade_target_code_mismatch`。
3. 低版本或相同版本返回 `plugin_version_same_version`（不支持降级/重复升级）。
4. `precheck.status!=passed` 的包不能进入升级。
5. staging download `status!=downloaded` 的包不能进入升级（`plugin_upgrade_target_download_invalid`）。
6. `checksum_missing` 的包默认不能进入升级（`plugin_upgrade_target_checksum_missing`）。
7. sha256 不一致不能进入升级（`plugin_upgrade_target_checksum_invalid`）。
8. 升级执行：`POST /api/v1/admin/plugins/:code/upgrade-from-package` 会创建 `plugin_upgrade_tasks` 记录并写入审计 `plugin.upgrade.requested/started/success/failed`，warning 升级需要显式 `confirm=true`。
9. 升级前会重新对目标包执行 package dry-run，危险文件/blocked 风险/ checksum mismatch 会阻断升级。
10. 升级成功后插件默认不自动启用；需要重新 enable-precheck + enable。
11. 目标版本声明 migrations 时，升级后插件会进入 `migration_pending`（但不会自动执行 migration）。
12. `GET /api/v1/admin/plugins/upgrade-tasks` / `GET /api/v1/admin/plugins/upgrade-tasks/:id` 可查询任务列表与详情。
13. `POST /api/v1/admin/plugins/upgrade-tasks/:id/retry` 仅允许重试 `failed` 的任务。
14. `DELETE /api/v1/admin/plugins/upgrade-tasks/:id` 仅标记 `deleted`，不影响插件当前状态与历史内容。
15. 升级过程不执行插件代码/脚本、不加载 Go plugin、不执行 migration、不影响历史内容详情与 SEO（/topics/:id、/c/:slug）。
16. blocked 升级即使勾选确认也不能继续。

## v1.7.0-P0-09 插件运行时治理验收（P0-01 ~ P0-08 链路）验收清单

本节用于 v1.7.0 P0 阶段的运行时治理验收：确认远程/包插件从 staging 下载、预检、兼容性检查、安装、启用前检查、启用注册、软卸载、升级的关键链路真实可用；仍不执行第三方代码，不做动态加载，不做市场。

下载治理（P0-01）：

1. 仅允许 https 下载；http/file/内网/localhost/重定向到内网均拒绝（SSRF 防护）。
2. 下载后写入 staging 并记录 status/file_size/final_url/content_type/sha256。
3. sha256 正确为 downloaded；不一致为 checksum_failed；缺失为 checksum_missing（默认不能进入后续安装/升级）。
4. 下载失败清理临时文件，并记录 error_message。
5. staging 列表/详情/删除 API 可用，且删除会删除文件或标记 deleted。

预检与 manifest 校验（P0-02）：

1. 合法包可 precheck；缺 manifest/非法 JSON/危险文件/路径穿越/symlink/hardlink/超限等均被阻断。
2. 核心越权声明（core/admin/system 权限、敏感路由覆盖、外链菜单、未知 hook）被阻断。
3. 预检结果保存并可查询，预检过程不执行任何包内代码/脚本。

兼容性检查（P0-03）：

1. 仅 precheck passed 可执行 compat-check。
2. can_install 由后端计算：errors=>false，warnings=>true(status=warning)。
3. core 版本不兼容、required 依赖缺失/版本不满足、plugin_code/content_type/permission/route/menu 冲突均阻断。

启用链路（P0-05/06）：

1. enable-precheck 基于已安装插件执行，can_enable 由后端计算。
2. enable 必须依赖 enable-precheck(can_enable=true)；migration pending/failed、配置无效、依赖缺失、冲突均阻断。
3. 启用成功后 status=enabled，运行时能力可被平台识别；不执行插件代码/脚本，不加载 Go plugin，不自动执行 migration，不自动启用所有社区。

软卸载（P0-07）：

1. 远程插件可软卸载；内置插件禁止。
2. required 依赖阻断软卸载；optional 依赖仅 warning。
3. 软卸载后禁止新建能力，但保留历史内容访问与 SEO（/topics/:id、/c/:slug）。

升级（P0-08）：

1. 升级输入必须来自同 plugin_code 更高版本 compat-check(can_install=true) 的包；checksum_missing 默认禁止。
2. 升级前重新 dry-run；升级后默认 disabled/migration_pending，不自动启用；需要重新 enable-precheck + enable。
3. 升级不执行插件代码/脚本、不执行 migration；失败不破坏旧版本治理状态与历史内容访问。

审计与权限边界：

1. 关键操作（download/precheck/compat-check/install/enable-precheck/enable/soft-uninstall/upgrade）均写入 admin_logs。
2. 普通前台用户不能访问 /api/v1/admin/*；版主不能跨社区越权治理。

## v1.7.0-P0-10 远程插件包治理总验收与发布归档（P0 收口）清单

本节用于 v1.7.0 P0 总收口：对 P0-01 ~ P0-09 的“远程插件包治理主链路 + 横向治理”做最终验收记录与发布归档口径统一。

执行过的检查命令（2026-05-16）：

1. 后端：`gofmt`（对改动 Go 文件）、`go test ./...`（通过）、`go build`（通过；如 `.devhub/` 不可写则用 `go build -o .tmp/bin/devhub .`）。
2. 后台：admin-only：`./scripts/check-frontend.sh --admin-only`（通过，admin build + Playwright 62 passed）。
3. 前台：frontend-only：`./scripts/check-frontend.sh --frontend-only`（通过，frontend build + Playwright 17 passed）。
4. SEO 抽查：
   - compose devhub 未暴露宿主机端口，且 `curl` 镜像拉取在当前环境失败；因此使用 `frontend-e2e` 容器内 Node HTTP 请求验证 `/topics/1/` 与 `/c/php/`：
     - `/topics/1/`：存在 `<title>`、canonical、`application/ld+json`；
     - `/c/php/`：存在 `<title>`、canonical。

发布归档口径（必须满足）：

1. 文档口径以：`VERSION`、`README.md`、`CHANGELOG.md`、`docs/releases/v1.7.0.md`、`docs/API.md`、`docs/PROJECT_PROGRESS.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/PLUGIN_PACKAGE.md`（如存在）为准，不得互相矛盾。
2. v1.7.0 P0 仅包含“治理链路”与“安全边界”，不包含插件市场、远程自动更新、动态加载、第三方代码执行、sandbox、硬卸载、migration down、自动依赖安装。

## Manifest 校验、dry-run 与配置型安装验收

本节用于记录 manifest + 配置型插件安装预备能力，不代表插件市场、插件包上传、远程安装或动态加载已完成。

已自动化：

- Go 单测覆盖合法 manifest 校验通过，并返回 normalized manifest、checksum、impact summary 和 install preview。
- Go 单测覆盖 code 冲突失败。
- Go 单测覆盖 content_type 冲突失败。
- Go 单测覆盖非法 Hook 名称失败。
- Go 单测覆盖非法 `config_schema` 失败。
- Go 单测覆盖依赖缺失时返回 warning / dependency 信息，不直接执行第三方代码。
- Go 单测覆盖 `InstallPluginManifest` 安装 manifest + 配置型插件后初始为 installed + disabled。
- Go 单测覆盖安装后的 manifest-only content type 可通过统一内容类型映射读取 create permission。
- Go 单测覆盖批量归档 / 恢复接口返回逐项 succeeded / failed 结果。

部分自动化：

- API 文档已记录 `POST /api/v1/admin/plugins/manifest/validate`、`POST /api/v1/admin/plugins/dry-run`、`POST /api/v1/admin/plugins/install`、`GET /api/v1/admin/plugins/health`、`POST /api/v1/admin/plugins/bulk-archive` 和 `bulk-restore`；浏览器 UI 已补最小 E2E，安装向导已进入抽屉分步流程，后续只继续增强插件包上传、依赖明细和更细影响对象列表。
- Health summary 当前由 Go/API 行为和后台已有运行状态视图覆盖；独立 `/plugins/health` 页面级展示仍待后续。

未覆盖 / 后续：

- 外部服务型 Webhook 真实 HTTP 调用、签名、超时和失败策略。
- 插件升级 dry-run / upgrade 已有抽屉分步展示；后续补独立版本兼容矩阵页面和更细升级影响对象列表。
- 插件包 zip 读取、签名校验、文件安全扫描和插件市场页面。
- 版本兼容矩阵 UI 和依赖版本范围完整自动化。
- 外部插件迁移真实 DDL runner、migration down、硬回滚和迁移前备份。

## 升级 dry-run 与版本兼容矩阵验收

本节用于 P2 升级预备能力；当前已完成最小升级执行闭环和后台抽屉式升级向导，但回滚、migration down、插件包升级和更细版本变更审计仍待后续。

已自动化：

- Go 单测覆盖 `POST /api/v1/admin/plugins/:code/upgrade/dry-run` 的兼容矩阵返回，包含 current/new 版本、Core 兼容范围、变更字段和 diff。
- Go 单测覆盖升级 dry-run 对现有插件 code 的校验与版本不递增时的警告信息。
- Go 单测覆盖 `POST /api/v1/admin/plugins/:code/upgrade` 的真实执行闭环，包含版本更新、manifest checksum 变更和状态保持。
- 后台 E2E 覆盖插件列表中的“升级预览”入口、结果面板和版本兼容摘要。

部分自动化：

- 版本兼容矩阵目前基于 manifest validate + upgrade dry-run 两层信息拼装，升级执行结果仍以结构化结果面板展示，完整升级 UI 向导、回滚和版本变更审计仍待后续。

未覆盖 / 后续：

- 版本兼容矩阵独立页面。
- 升级影响分析的更细粒度对象列表。

## `/admin-next/plugins` 治理入口收口验收

本节记录本轮后台插件治理中心最小 UI 收口和对应 E2E 结果。

已自动化：

- `/admin-next/plugins` 顶部健康总览卡片可见，至少展示健康、警告、异常、已禁用、已归档、迁移待处理、配置无效、依赖缺失和 Hook 异常聚合。
- `/admin-next/plugins` 顶部支持 `校验 Manifest`、`Dry-run 预览`、`安装插件`、`批量归档`、`批量恢复` 入口。
- `/admin-next/plugins` manifest 面板可打开并提交合法 manifest JSON，返回结构化结果摘要。
- `/admin-next/plugins` dry-run 面板可打开并返回结构化安装预览。
- `/admin-next/plugins` install 面板可打开并返回结构化安装结果。
- `/admin-next/plugins` upgrade dry-run 可展示当前版本 / 新版本 / 兼容状态 / 变更字段；真实 upgrade 由 Go 单测覆盖最小执行闭环。
- `/admin-next/plugins` 批量归档 / 恢复可对选中插件返回结果摘要。
- `/admin-next/plugins` 的 manifest 输入、校验结果、dry-run、安装确认、升级预览和升级执行已收口为右侧抽屉分步流程，不再作为页内长面板挤压插件列表；原 `plugin-manifest-panel`、`plugin-manifest-input` 等 E2E 锚点保留。
- `/admin-next/plugins` 批量归档 / 恢复已支持操作前 impact 预览和操作后 succeeded / failed 明细展示，并提供审计跳转入口。
- `/admin-next/plugins` 状态治理视图已作为异常处理入口，按迁移待处理、迁移失败、Hook 异常、配置无效、依赖缺失和已归档插件聚合。
- 插件详情抽屉可展示运行状态说明、归档态提示、状态原因和建议操作。
- 现有插件治理 E2E、PluginContent、迁移、Hook、审计和归档态浏览器链路未退化。

已跳过的旧回归：

- `opens plugin detail tabs and shows schema validation errors`
- `archives plugin and shows archived state with restore entry`

说明：本轮优先把测试步骤收敛到升级执行、安装向导和批量治理的最小闭环，旧详情 / 归档回归已由其他插件治理测试间接覆盖，保留会增加执行时间和状态污染风险。

部分自动化：

- 安装 / 升级向导的后端行为已有 Go / API 保护；本轮后台 E2E 已覆盖 manifest 输入、校验 / dry-run、确认安装、升级预览、确认升级和结果面板的最小浏览器链路。
- 状态治理视图当前聚合插件列表已有健康摘要和状态字段；更多告警、自动恢复和外部服务 Webhook 状态仍是后续能力。

未覆盖 / 后续：

- 插件包 zip 上传、外部服务型 Webhook、版本兼容矩阵独立页面和更细批量治理策略。

## v1.3.5 测试收口

已覆盖：

- `/admin-next/plugins` 信息架构调整后仍可打开，健康总览、筛选、批量操作、详情抽屉入口不退化。
- 安装向导三步流：Manifest 输入、校验 / dry-run 结构化预览、确认安装。
- 升级向导三步流：Manifest 输入、兼容矩阵 / diff / migration plan 预览、确认执行升级。
- 批量归档 / 恢复：影响预览、二次确认、`succeeded` / `failed` 表格、审计跳转。
- 状态治理页：异常插件、迁移待处理、Hook 异常、配置无效、依赖缺失和归档插件入口可读。
- PluginContent：归档 / 禁用插件历史内容仍可查看，批量隐藏 / 恢复不退化。

历史执行记录：

- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 E2E `28 passed / 1 skipped`；日志目录 `.devhub/checks/20260512-231335/`。（该 skipped 已在 2026-05-13 的 v1.4.0 收口验收中恢复并通过）

精简原则：

- 只新增覆盖新向导和关键入口的最小后台 E2E。
- 不在 v1.3.5 UI 收口阶段扩展插件市场、远程安装、动态加载或外部服务型 Webhook 的浏览器矩阵。

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

- E2E 默认通过 `docker-compose.yml` 内部启动一个 `devhub` 服务（memory store），并将 `DEVHUB_E2E_ORIGIN` 设为 `http://devhub:8090`；因此通常不再要求你先在宿主机启动后端。
- 如需改为访问宿主机已启动的 DevHub（例如用 MySQLStore/自定义端口），可覆盖 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:<port>`（或按实际可达地址设置）。
- E2E 前需要先用 `admin-e2e` 构建后台静态产物；不要和 E2E 并行写 `web/admin-vue`，否则可能读到短暂不完整的静态资源。

## v1.4.0-P1-07 依赖检查与版本兼容矩阵验收

已覆盖能力：

- 后端单测覆盖无依赖、required 依赖满足 / 缺失 / disabled / archived / migration_failed / config_invalid、optional 缺失 warning、版本满足 / version_mismatch、自依赖、两节点 / 三节点循环、Core 兼容、upgrade dependency diff、upgrade required 新依赖阻断和 enable 时重新检查依赖。
- 后台 Playwright 新增 `web/admin-app/tests/e2e/plugin-dependencies.spec.js`，覆盖安装向导依赖矩阵、required 缺失、optional 缺失、Core 不兼容、升级向导新增 required 依赖阻断、插件详情 Dependencies 区域和依赖插件定位入口。
- 后台安装 / 升级向导展示 `dependency_summary`、逐项 `dependencies`、Core `compatibility` 和 `dependency_diff`；后端仍负责最终阻断，不依赖前端判断。

本轮执行记录：

- `gofmt -w internal/plugins/scaffold/scaffold.go internal/plugins/manifest_validator.go internal/plugins/version_compat.go internal/plugins/version_compat_test.go internal/service/service.go internal/service/plugin_dependencies_test.go internal/domain/models.go internal/transport/httpapi/router.go`：通过。
- `go test ./internal/plugins ./internal/plugins/scaffold ./internal/service ./internal/transport/httpapi`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；Vite 仍有既有大 chunk warning。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-dependencies.spec.js`：首次因后端服务仍为旧进程、UI 未拿到新依赖结构失败；使用当前代码重启 DevHub 后复跑通过，`2 passed`。
- `./scripts/check-frontend.sh --admin-only`：最终通过，后台 build 通过，后台 E2E `30 passed / 1 skipped`；日志目录 `.devhub/checks/20260513-003728/`。（该 skipped 已在 2026-05-13 的 v1.4.0 收口验收中恢复并通过）

仍遗留：

- 不支持自动安装依赖、远程下载依赖、插件市场推荐或依赖图大屏。
- 不支持 npm 风格 `^`、`~`、`||`、预发布标签等复杂版本约束。
- optional 循环依赖当前按 error 阻断，后续如要放宽需同步后端、UI 和文档。

## 2026-05-14：v1.6.0-P0-03 真实签名验签与可信发布者

新增 / 调整覆盖：

- 后端：`internal/service/plugin_trusted_publishers_test.go` 覆盖可信发布者新增、重复拒绝、非法算法 / 非法公钥拒绝、block / restore，以及 block 后已签名包变为 blocked。
- 后端：`internal/plugins/plugin_package_signature_test.go` 覆盖 trusted+verified、unknown publisher 技术验签、unsupported algorithm blocked、revoked publisher blocked。
- 后台 E2E：`web/admin-app/tests/e2e/plugin-trusted-publishers.spec.js` 覆盖可信发布者页面、新增、编辑、block、restore、详情与安全边界文案。
- 后台 E2E：`web/admin-app/tests/e2e/plugin-package-signature-verify.spec.js` 覆盖本地仓库 signed package verified、unknown publisher、invalid signature blocked 和详情页签名信息。

必测命令：

- `gofmt`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-trusted-publishers.spec.js`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-package-signature-verify.spec.js`
- `./scripts/check-frontend.sh --admin-only`

当前结果：上述必测命令已执行并通过；`./scripts/check-frontend.sh --admin-only` 通过，后台 build 通过，后台 E2E `56 passed`，日志目录 `.devhub/checks/20260514-213651/`。本轮未涉及前台内容、导航、搜索或 SEO，未执行 frontend-only / SEO curl。

## 2026-05-14：v1.6.0-P0-04 远程插件索引只读镜像测试

新增覆盖：

- 后端：`internal/service/plugin_remote_index_test.go` 覆盖索引源创建、URL SSRF 防御、禁用源不可拉取、远程 index.json 拉取、无 package_url 下载、invalid JSON、unknown publisher 风险和响应过大。
- 后台 E2E：`web/admin-app/tests/e2e/plugin-remote-indexes.spec.js` 覆盖只读页面、索引源新增、拉取、远程插件列表、详情、trusted / unknown / incompatible 展示，以及不显示下载、安装、自动更新或动态加载入口。

必测命令：

- `gofmt`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-remote-indexes.spec.js`
- `./scripts/check-frontend.sh --admin-only`

本轮未修改前台内容、搜索或 SEO；不需要 frontend-only / SEO curl。

P0-04 最终结果：上述必测命令已执行并通过；`./scripts/check-frontend.sh --admin-only` 通过，后台完整 E2E `57 passed`，日志目录 `.devhub/checks/20260514-225300/`。

## v1.6.0-P0-05 插件包版本仓库与升级差异对比（2026-05-14）

新增覆盖：

- 后端 `internal/service/plugin_versions_test.go`：版本仓库聚合 installed / uploaded / remote，版本比较 same / downgrade / remote readonly 阻断，manifest diff 高风险与敏感字段脱敏。
- 后台 E2E `web/admin-app/tests/e2e/plugin-versions-upgrade-diff.spec.js`：版本仓库页面、installed/local/remote 来源展示、remote readonly、升级差异摘要、权限 / 配置 schema / 依赖 diff、高风险高亮、敏感字段脱敏和提交升级审批入口。

本轮执行记录：

- `gofmt`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；Vite 仍有既有大 chunk warning。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-versions-upgrade-diff.spec.js`：通过，`1 passed`。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台完整 E2E `58 passed`；日志目录 `.devhub/checks/20260515-005940/`。

未执行：本轮未修改前台内容、搜索或 SEO，未执行 `./scripts/check-frontend.sh --frontend-only` 与 SEO curl。

## v1.6.0-P1-07 插件配置密钥轮换策略（2026-05-15）

新增覆盖：

- 后端：
  - `internal/plugins/config_keyring_test.go`：密钥环环境变量解析与校验（legacy/split/JSON 形式）。
  - `internal/service/plugin_config_key_rotation_test.go`：rotation dry-run + re-encrypt（v1 -> v2，current key_id）。
- 后台 E2E：`web/admin-app/tests/e2e/plugin-config-key-rotation.spec.js` 覆盖密钥状态页面、rotation dry-run 与 re-encrypt（不展示 key material / 不展示密文）。

执行记录：

- `gofmt`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-config-key-rotation.spec.js`：通过，`1 passed`。
- `./scripts/check-frontend.sh --admin-only`：通过（包含该新增 E2E），日志目录以本轮摘要为准。

已知限制：

- `include_config_versions=true` 暂不支持，返回 `plugin_config_rotation_history_unsupported`（历史轮换后续补齐）。

## v1.6.0-P1-09 插件治理 UI 分页与 E2E 基建（2026-05-15）

新增 / 调整覆盖：

- 后台 E2E：`web/admin-app/tests/e2e/plugin-governance-pages.spec.js` 覆盖 `/admin-next/plugins`、插件治理按“功能域分层导航（模块/功能域/页面）”、旧路由兼容、本地仓库 / zip 上传包 / 可信发布者 / 远程索引 / 版本仓库 / 操作历史 / 密钥轮换等入口，以及统一安全边界文案。
- E2E helper：`web/admin-app/tests/e2e/helpers/pluginHelpers.js` 提供打开插件页、断言安全边界、断言错误码、打开包治理 / 安全 / 远程 / 操作页等基础能力。
- 旧导航专项由 `plugin-governance-pages.spec.js` 承接；本轮未新增 `test.only`、`page.pause` 或长期 skipped 用例。
- 配置密钥轮换 E2E 增强了密钥缺失 / blocked UI 的稳定断言，避免环境未配置 key 时只出现短暂 toast。

Fixture 结论：

- 本轮不搬迁既有后端与后台 E2E 插件包 fixture，继续沿用现有 `internal/testdata/plugin-packages/` 与 `web/admin-app/tests/fixtures/` 口径。
- 危险包、zip slip、checksum mismatch、签名失败等 fixture 仍只用于测试目录，不进入正式 examples；后续建议补统一生成脚本以减少重复 fixture。

执行记录：

- `gofmt`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-governance-pages.spec.js`：通过，`4 passed`。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-config-key-rotation.spec.js tests/e2e/plugin-governance-pages.spec.js`：通过，`5 passed`。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台完整 E2E `62 passed`，日志目录 `.devhub/checks/20260515-130417/`。

未执行：本轮未修改前台内容、前台导航、搜索或 SEO 共享逻辑，未执行 `./scripts/check-frontend.sh --frontend-only` 与 `/topics/1/`、`/c/php/` SEO curl。

## 2026-05-16 插件模块导航层级重构（按功能域）

验收重点：

- `/admin-next` 后台导航采用三级结构：一级模块（最左侧）/ 二级功能域（第二列分组）/ 三级具体页面（按钮）。
- 页内状态筛选使用 Tab，并同步 URL query（例如插件列表 `?status=enabled&health=error`），不改变左侧菜单选中态。
- 面包屑展示模块/功能域/页面；当页面无更深层级时不重复显示末级标题。
- 旧路由兼容：`/admin-next/plugins`、`/admin-next/plugins/governance`、`/admin-next/plugins/manifest`、`/admin-next/plugins/diagnostics` 不 404。

推荐最小回归：

- 后台 E2E：`docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-governance-pages.spec.js`

本轮实际执行：

- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-governance-pages.spec.js`：通过，`4 passed`。

未执行：

- 前台 `web/frontend-app` build 与前台 E2E：本轮未修改前台页面、SEO 或共享导航逻辑，未做前台构建回归（如需全量回归请按 `docs/TESTING.md` 的完整矩阵执行）。

## v1.6.0-P1-10 插件包上传与分发前置能力总验收（2026-05-15）

E2E 覆盖复查：

- v1.4 插件治理：`plugin-governance.spec.js`、`plugin-content.spec.js`、`plugin-hooks.spec.js`、`plugin-dependencies.spec.js`、`plugin-readiness-errors.spec.js`、`plugin-navigation-admin.spec.js` 均存在。
- v1.5 插件包治理：`plugin-package-dryrun.spec.js`、`plugin-package-security.spec.js`、`plugin-package-repository.spec.js`、`plugin-package-install.spec.js`、`plugin-config-versions.spec.js`、`plugin-config-encryption.spec.js`、`plugin-approvals.spec.js`、`plugin-package-signature.spec.js`、`plugin-package-export.spec.js` 均存在。
- v1.6 插件上传与分发前置：`plugin-package-upload.spec.js`、`plugin-package-upload-lifecycle.spec.js`、`plugin-trusted-publishers.spec.js`、`plugin-package-signature-verify.spec.js`、`plugin-remote-indexes.spec.js`、`plugin-versions-upgrade-diff.spec.js`、`plugin-operation-recovery.spec.js`、`plugin-config-key-rotation.spec.js`、`plugin-governance-pages.spec.js` 均存在。
- `plugin-package-export-zip.spec.js` 当前不存在；当前实现只有本地目录包导出和 `plugin-package-export.spec.js` 覆盖，不提供 zip 下载导出，已登记为 v1.7 技术债。

skipped / flaky / TODO 结论：

- 本轮检查未发现 `test.only` 或 `page.pause`。
- `plugin-readiness-errors.spec.js` 文档保留的环境探测 skip 只用于旧后端二进制未包含 readiness 路由时避免误判；完整后台 E2E 在当前构建下通过时不视为长期 skipped。
- 未新增 flaky 标记；若后续 Docker 网络或旧二进制导致 readiness 探测 skip，应优先重建 `.devhub/devhub` 并重启 compose。

最终执行记录：

- `gofmt -w $(git ls-files '*.go')`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台完整 E2E `62 passed`，日志目录 `.devhub/checks/20260515-133558/`。
- `./scripts/check-frontend.sh --frontend-only`：通过，前台 build 通过，前台 E2E `17 passed`，日志目录 `.devhub/checks/20260515-133815/`。
- SEO curl `/topics/1/`：通过，命中 title / canonical / Article JSON-LD / article / h1。
- SEO curl `/c/php/`：通过，命中 title / description / canonical / h1 / topics / tag-cloud。

## v1.7.0-P0-01 远程插件包下载到 staging 轻量验收清单

后端新增覆盖：`internal/service/plugin_package_download_test.go`。

验收项：

1. https 插件包 URL 可以下载到 staging。
2. http URL 被拒绝。
3. localhost URL 被拒绝。
4. 127.0.0.1 URL 被拒绝。
5. 内网 IP URL 被拒绝。
6. 重定向到内网地址应被拒绝（服务端 redirect hook 与 DialContext 均复检；专项 E2E 后续补）。
7. 超过大小限制的下载被中止（服务端限制已实现；专项 fixture 后续补）。
8. sha256 正确时状态为 `downloaded`。
9. sha256 错误时状态为 `checksum_failed`。
10. sha256 错误时不能进入后续安装。
11. 下载失败会记录 `error_message`。
12. 下载失败会清理临时文件。
13. staging 列表可以查询。
14. staging 包可以删除。
15. 删除 staging 包会删除文件并标记 `deleted`。
16. 审计记录包含 download success / failed / rejected / checksum failed / staging deleted。
17. 没有 sha256 的包被标记为 `checksum_missing`。
18. 不执行插件包内代码。
19. 不安装插件。
20. 不影响当前内置插件启停。
21. 不影响 `/topics/:id` SEO。
22. 不影响 `/c/:slug` SEO。

本轮最低检查记录：

- `gofmt`：通过。
- `docker run --rm -v "$PWD":/workspace -w /workspace golang:1.23-bookworm gofmt -w ...`：通过。
- `docker run --rm -v "$PWD":/workspace -w /workspace golang:1.23-bookworm go test ./...`：通过。
- `docker run --rm -v "$PWD":/workspace -w /workspace golang:1.23-bookworm go build -buildvcs=false -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm devhub go test ./...`：未执行成功，原因是当前 `devhub` 运行镜像不包含 `go` 可执行文件；已改用一次性 `golang:1.23-bookworm` Docker 容器执行 Go 检查。

## v1.7.0-P0-03 插件依赖 / 兼容性检查轻量验收清单

后端新增覆盖：`internal/service/plugin_package_compat_test.go`。

验收项：

1. precheck passed 的包可以执行 compat-check。
2. precheck failed 的包不能执行 compat-check。
3. compatible_core_version 兼容时通过。
4. compatible_core_version 不兼容时 status=incompatible。
5. required 依赖缺失时失败。
6. optional 依赖缺失时只产生 warning。
7. 依赖版本不满足时失败。
8. plugin_code 与内置插件冲突时失败。
9. plugin_code 与已安装插件冲突时失败。
10. content_type 冲突时失败。
11. permission 声明 core.* 被拒绝。
12. permission 声明 admin.* 被拒绝。
13. menu 外链被拒绝。
14. route 覆盖核心敏感 API 被拒绝。
15. route 覆盖 /topics/:id 被拒绝。
16. 未知 Hook 被拒绝。
17. default_config 不符合 config_schema 时失败。
18. migration direction=down 不执行，只产生错误。
19. 有 errors 时 can_install=false。
20. 只有 warnings 时 can_install=true 且 status=warning。
21. can_install 由后端返回。
22. 审计记录包含 compat_check success / failed / incompatible。
23. 不安装插件。
24. 不启用插件。
25. 不注册权限 / 菜单 / 路由 / Hook。
26. 不执行 migration。
27. 不影响当前内置插件启停。
28. 不影响 `/topics/:id` SEO。
29. 不影响 `/c/:slug` SEO。

本轮最低检查记录：

- `gofmt`：通过。
- `docker run --rm -v "$PWD":/workspace -v /tmp/devhub-go-mod-cache:/go/pkg/mod -v /tmp/devhub-go-build-cache:/root/.cache/go-build -w /workspace golang:1.23-bookworm gofmt -w ...`：通过。

## v1.7.0-P0-05 插件启用前安全检查（enable-precheck）轻量验收清单

后端新增覆盖：`internal/service/plugin_enable_precheck_test.go`。

验收项：

1. installed 插件可以执行 enable-precheck。
2. 未安装成功的插件不能执行 enable-precheck。
3. archived 插件不能执行 enable-precheck。
4. dependency_missing 插件不能执行 enable-precheck。
5. 文件被篡改（manifest 与安装快照不一致）时 enable-precheck 失败。
6. Core 版本不兼容时 can_enable=false。
7. required 依赖缺失或版本不满足时 can_enable=false。
8. optional 依赖缺失只产生 warning。
9. 配置无效时 can_enable=false。
10. migration pending / failed 时 can_enable=false。
11. content_type 冲突时 can_enable=false。
12. permission 越权或冲突时 can_enable=false。
13. route 覆盖核心敏感路径时 can_enable=false。
14. 未知 Hook 时 can_enable=false。
15. 只有 warnings 时 can_enable=true 且 status=warning。
16. can_enable 由后端返回。
17. enable-precheck 写入审计（requested/success/failed/blocked/deleted）。
18. enable-precheck 不启用插件，不注册权限/菜单/路由/Hook，不执行 migration，不执行插件代码。

## v1.7.0-P0-06 插件启用与运行时注册（enable）轻量验收清单

1. enable-precheck `can_enable=true` 且 `status=passed|warning` 的插件可以启用（`POST /api/v1/admin/plugins/enable-prechecks/:id/enable`）。
2. enable-precheck `can_enable=false` 或 `status=failed|blocked|deleted` 的插件不能启用。
3. enable-precheck 过期时不能启用（默认 TTL 600 秒；`DEVHUB_PLUGIN_ENABLE_PRECHECK_TTL_SECONDS`）。
4. migration pending / failed 时不能启用。
5. 配置无效时不能启用。
6. 依赖缺失或版本不满足时不能启用。
7. content_type 冲突时不能启用。
8. 启用成功后插件状态为 `enabled`。
9. 启用成功后创建链路能识别插件声明的 `content_type_definitions`（仍受子站启用、板块绑定、权限校验约束）。
10. 启用成功后前台/后台菜单按插件状态与权限规则显示（前端隐藏不作为安全边界）。
11. 启用成功后会触发 `AfterPluginEnabled`（non-blocking，如 HookBus 支持）。
12. 启用过程写入审计：`plugin.enable.requested/started/success/failed/retry`、`plugin.runtime.registered`。
13. 启用失败不留下半启用状态：插件状态不应为 enabled。
14. 启用阶段不执行插件包内代码、不运行 package scripts、不加载 Go plugin、不自动执行 migration、不自动为所有子站启用。
19. 不影响当前内置插件启停。
20. 不影响 /topics/:id SEO。
21. 不影响 /c/:slug SEO。
- `docker run --rm -v "$PWD":/workspace -v /tmp/devhub-go-mod-cache:/go/pkg/mod -v /tmp/devhub-go-build-cache:/root/.cache/go-build -w /workspace golang:1.23-bookworm go test ./...`：通过。

## v1.7.1 插件包签名验签（detached signature）轻量验收清单

1. 可以新增 / 更新 trusted publisher（仅公钥，不保存私钥）。
2. 可以 block / revoke / restore trusted publisher；`blocked/revoked` 的 key 不能通过验签。
3. `expires_at` 过期的 key 不能通过验签，验签结果为 `key_expired`。
4. Ed25519 签名包（`devhub-signature.json`）可以验签通过（`status=verified`）。
5. signature_url 下载仅允许 https + .json，且走 SSRF 防护、重定向复检与大小限制（默认 64KB）。
6. 不支持的算法被拒绝（`algorithm_unsupported`）。
7. 公钥不存在时验签结果为 `untrusted_publisher`。
8. package_sha256 / manifest_sha256 / plugin_code / version 不一致时验签失败（`hash_mismatch` / `payload_mismatch`）。
9. payload 被篡改验签失败（`failed`）。
10. 未提供签名文件时验签结果为 `unsigned`（不阻断 staging/precheck，但默认阻断 install/upgrade）。
11. 默认策略下：compat-check 缺少验签记录或非 `verified` 时 `can_install=false`（阻断 install/upgrade）。
12. install/upgrade 执行前会再次要求 `verified`（后端强校验，不能靠前端伪造）。
13. 验签结果会写入 `plugin_package_signatures`（或等效记录）。
14. 验签操作与 publisher 变更写入审计（admin_logs）。
15. 整个验签链路不执行插件代码、不运行 package scripts、不加载 Go plugin，不影响 /topics/:id 与 /c/:slug SEO。

### v1.8.3-S12 插件包 upload -> promote -> install 验收闭环

后端新增覆盖：`internal/service/plugin_package_install_test.go` 的本地仓库安装 / dry-run 计划校验回归；`internal/service/plugin_package_upload_test.go` 的 promote 后进入本地仓库与安装 dry-run 回归。

1. 上传包可进入暂存区。
2. 上传包预检可生成 passed / warning / blocked。
3. blocked 上传包不能 promote。
4. failed 上传包不能 promote。
5. promote 接口后端强校验 blocked 状态。
6. promote 成功后生成本地仓库记录。
7. promote 成功后记录 source_upload_id。
8. promote 不执行 SQL。
9. promote 不安装插件。
10. promote 不启用插件。
11. promote 不执行第三方代码。
12. upload 暂存包不能直接 install。
13. 本地仓库包 install 前必须重新 dry-run。
14. 无 dry-run plan 时 install 被拒绝。
15. dry-run 过期时 install 被拒绝。
16. dry-run checksum 与当前包不一致时 install 被拒绝。
17. install dry-run 不执行 SQL。
18. install dry-run 输出 migration plan。
19. install 只执行 migrations/ 计划。
20. install 不执行根目录 001_schema.sql。
21. install 不执行 package scripts。
22. install 成功后触发 PluginRegistry reload。
23. install 失败不污染 registry。
24. 相关操作写审计。

### v1.8.3-S21 插件包清理 / 删除入口验收

后端新增覆盖：`internal/service/plugin_package_upload_test.go` 验证 promoted 上传记录不能被直接删除、上传 cleanup 必须 dry-run + confirm；`internal/service/plugin_package_repository_test.go` 验证 promoted 未安装本地仓库包可删除、已安装包拒绝删除并保留文件。

最低检查要求：

- `go test ./internal/service -run 'TestPluginPackageUploadLifecycle_DeletePromotedBlockedAndCleanup|TestDeletePluginPackageRepositoryPackage' -count=1`
- `go test ./...`
- `go build ./...`
- 后台 build：`./scripts/check-frontend.sh --admin-only --quick`
- 最小 E2E / service：以上 service 测试覆盖本轮 cleanup / delete 安全规则；如需浏览器回归，追加 `web/admin-app/tests/e2e/plugin-package-upload-lifecycle.spec.js` 与插件包治理页 smoke。

本轮执行结果：

- Docker `gofmt`：通过。
- `go test ./internal/service -run 'TestPluginPackageUploadLifecycle_DeletePromotedBlockedAndCleanup|TestDeletePluginPackageRepositoryPackage' -count=1`：通过。
- `go test ./internal/service -run 'TestDeletePluginPackageRepositoryPackage' -count=1`：通过（blocked/invalid 非 promoted 测试包删除规则补充后复查）。
- `go test ./...`：通过。
- `go build ./...`：通过；首次 Docker 构建因 Git safe.directory / VCS stamping 失败，设置 `/workspace` 为 safe directory 后复跑通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260527-123700/`；显式 repository cleanup 别名补充后复跑通过，日志目录 `.devhub/checks/20260527-124613/`。
- blocked/invalid 非 promoted 测试包删除规则补充后再次复跑后台 build 通过，日志目录 `.devhub/checks/20260527-131421/`。
- 本地仓库所有未安装包可删除规则补充后，`go test ./internal/service -run 'TestDeletePluginPackageRepositoryPackage' -count=1` 通过，后台 build 通过，日志目录 `.devhub/checks/20260527-132144/`。
- `git diff --check`：通过。
25. 后台 blocked 状态显示中文阻断原因。
26. 后台未 dry-run 时禁用 install 按钮。
27. 不执行第三方代码。
28. 不开放动态加载。
29. 不改变 Webhook 协议。
30. 不改变 Secret / Token 安全模型。

本轮执行结果：

- `gofmt`：已执行。
- `go test ./...`：通过。
- `go build ./...`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260519-095725/`。

### v1.8.3-S13 真实插件包验收 S12 链路

fixture 生成脚本：

```bash
./scripts/build-plugin-package-fixtures.sh --suffix smoke
```

生成物：

- `devhub-fixture-valid-plugin{suffix}.zip`：可通过链路的真实 zip，包含 `manifest.json`、`checksums.json`、`README.md`、`config.example.json`、`migrations/001_init.sql`，不包含 package scripts。
- `devhub-fixture-blocked-plugin{suffix}.zip`：包含危险 `scripts/install.sh`，用于验证 blocked / failed 包不可 promote，脚本不会被执行。
- `devhub-fixture-deprecated-schema-plugin{suffix}.zip`：包含根目录 `001_schema.sql` 和 `migrations/001_init.sql`，用于验证 deprecated warning 与根目录 schema 不执行。

真实链路 E2E：

```bash
docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-package-real-fixtures.spec.js
```

验收项：

1. valid fixture 插件包可生成。
2. blocked fixture 插件包可生成。
3. deprecated schema fixture 插件包可生成。
4. valid 包可上传到暂存区。
5. blocked 包可上传到暂存区。
6. blocked 包 precheck 后状态为 blocked / failed。
7. blocked 包 promote 被后端拒绝。
8. blocked 包不会进入本地仓库。
9. valid 包 precheck 后状态为 passed / warning。
10. valid 包 promote 成功。
11. promote 后本地仓库出现对应包。
12. promote 不执行 SQL。
13. promote 不安装插件。
14. promote 不启用插件。
15. promote 不执行第三方代码。
16. upload 暂存包不能直接 install。
17. 本地仓库包未 install dry-run 时不能 install。
18. install dry-run 输出 migration plan。
19. install dry-run 不执行 SQL。
20. install dry-run 绑定 package_id / checksum / plugin_code / version。
21. dry-run 过期或不匹配时 install 被拒绝。
22. install 只执行 migrations/。
23. install 不执行根目录 001_schema.sql。
24. install 不执行 package scripts。
25. install 成功后 PluginRegistry reload。
26. install 失败不污染 registry。
27. 相关操作写审计。
28. 后台 blocked 状态显示中文阻断原因。
29. 后台未 dry-run 时禁用 install 按钮。
30. 不执行第三方代码。
31. 不开放动态加载。
32. 不改变 Webhook 协议。
33. 不改变 Secret / Token 安全模型。

当前说明：S13 E2E 使用真实 Admin API 上传 zip、执行 promote / dry-run / install，并打开后台“暂存上传包”和“本地包与预检”页面做可见性 smoke；该检查不替代全量后台 E2E。

本轮执行结果：

- `bash -n scripts/build-plugin-package-fixtures.sh`：通过。
- `bash -n dev.sh`：通过。
- `./scripts/build-plugin-package-fixtures.sh --suffix check`：通过，生成 valid / blocked / deprecated schema 三类真实 zip。
- `docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-package-real-fixtures.spec.js`：通过，1 个真实链路 E2E 通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `./scripts/check-frontend.sh --admin-only --quick`：通过，日志目录 `.devhub/checks/20260519-120918/`。

### 本轮执行命令与结果（v1.7.1）

- `gofmt -w …`：已执行（仅对变更 Go 文件）。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
