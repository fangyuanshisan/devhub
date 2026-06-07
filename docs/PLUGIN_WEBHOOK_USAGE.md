# DevHub Webhook 插件使用方法

[返回文档入口](README.md)

更新时间：2026-05-29

这份文档面向插件管理员和维护者，说明当前仓库里已经能用的 Webhook 子集怎么配、怎么看、怎么排障。当前可用的是 `external_service` 的 **non-blocking delivery 子集**、Webhook Secret 管理、Callback Token 管理和受控回调 API；**完整第三方运行模型仍未实现**。

## 先分清三种能力

| 能力 | 方向 | 作用 |
| --- | --- | --- |
| `external_service` | DevHub -> 外部服务 | 让插件声明 Hook 后，由 DevHub 异步投递 HTTP 请求。当前只支持 non-blocking。 |
| Webhook Secret | DevHub -> 插件服务 | 用于 DevHub 向插件服务发送 Webhook 时做签名。 |
| Callback Token | 插件服务 -> DevHub Core | 用于插件服务调用受控 Core API，例如读取配置或写入插件审计。 |

边界先说清：

- 不执行第三方代码。
- 不开放 blocking Hook。
- 不开放远程 iframe URL。
- 不把 Secret / Token 明文暴露给列表、详情、日志或审计。
- `external_service` 只负责 HTTP 投递，不等于完整第三方运行时。

## 适合什么场景

- 内容创建、更新、审核后通知外部系统。
- 把 DevHub 事件同步到独立插件服务。
- 做异步统计、通知、索引、排障回传。
- 让插件服务在受控范围内读取配置或写审计。

不适合：

- 阻断主流程的同步校验。
- 远程执行插件包里的代码。
- 直接访问 DevHub 数据库。

## 推荐使用顺序

1. 在插件 manifest 里声明 Hook。
2. 在后台或 Admin API 里配置 `external_service`。
3. 先做健康检查。
4. 触发一次真实业务事件。
5. 在 Webhook 治理里看投递记录、异常、熔断和事件。
6. 如果插件服务还要回调 Core，再创建 Callback Token。

## 1. 声明 Hook

manifest 里的 `hooks[]` 需要声明成 `service_type=external_service`，并且只能是 `mode=non_blocking`。

示例：

```json
{
  "name": "AfterCreateContent",
  "mode": "non_blocking",
  "service_type": "external_service",
  "path": "/hooks/content.after_create",
  "method": "POST",
  "timeout_ms": 3000,
  "retry_enabled": true,
  "max_attempts": 3,
  "failure_policy": "warn",
  "enabled": true
}
```

注意：

- `path` 必须是相对路径，不能写成完整域名。
- 第一版只支持 `POST`。
- `max_attempts` 最高 5 次。
- `failure_policy` 常用 `warn`，不要默认写成 blocking 思路。

## 官方示例插件

本仓库新增官方样板插件包：

```text
examples/plugins/official_webhook_notify/
  manifest.json
  README.md
  config.example.json
  checksums.json
  migrations/README.md
```

用途：

- 作为 external_service Webhook 插件的可复制样板。
- 验证 upload -> precheck -> promote -> install dry-run -> install。
- 配合 `cmd/webhook-mock-receiver` 验证健康检查、`AfterCreateContent` 投递、失败记录和手动重试。

该包不包含可执行代码、危险文件、真实 secret、用户数据、远程 iframe URL 或外部 SQL。`migrations/` 只保留说明文件；如后续有真实迁移，也只能放在 `migrations/` 下。

## 官方 Webhook 模板

如果要开发自己的 external_service Webhook 插件，建议从模板复制：

```text
examples/plugins/templates/external-service-webhook/
  manifest.json
  README.md
  config.example.json
  checksums.json
  receiver.example.md
  PACKAGING.md
  migrations/000_noop.json
```

该模板基于 `official_webhook_notify`，但它仍只是插件包模板，不是内置运行时：DevHub 不运行 receiver，不执行插件包代码，只按 manifest 中的 `service_type=external_service` / `mode=non_blocking` Hook 声明投递 HTTP 请求。

完整开发步骤、安全规范和验收清单见 [声明型插件开发者指南](PLUGIN_DEVELOPER_GUIDE.md)。

本地打包和校验：

```bash
./scripts/plugin-package-build.sh examples/plugins/official_webhook_notify
./scripts/plugin-package-check.sh examples/plugins/official_webhook_notify
./scripts/plugin-package-build.sh examples/plugins/templates/external-service-webhook
./scripts/plugin-package-check.sh examples/plugins/templates/external-service-webhook
```

这些命令只生成 / 校验本地包，不执行 receiver 代码、不调用 external_service、不访问远程市场，也不会安装或启用插件。校验输出 `blocked` 时必须先修复 blocker；`warning` 通常表示未签名或缺少 publisher，可在测试环境继续 upload / precheck，但生产发布前建议补齐。

## 2. 配置 external_service

当前后台插件详情概览卡片和“运行记录”区域都提供“配置 external_service”入口，Webhook 治理页也提供同类直达入口；也可以使用 Admin API。

接口：

- `GET /api/v1/admin/plugins/:code/external-service`
- `PUT /api/v1/admin/plugins/:code/external-service`
- `POST /api/v1/admin/plugins/:code/external-service/health-check`

常用字段：

- `endpoint_url`
- `health_check_path`
- `timeout_ms`
- `failure_policy`
- `auth_type`
- `token`
- `enabled`

建议：

- 线上只用 `https://`。
- 本地开发可用 `localhost` / `127.0.0.1` / `::1`。
- 非 localhost 的 HTTP endpoint 默认会被安全策略拒绝。生产建议使用 HTTPS；本地开发如果要访问 Docker host gateway，例如 `http://172.17.0.1:18081`，需要显式 allowlist。allowlist 有两种来源：启动环境变量来源和后台受控配置来源；系统默认允许 `localhost` / `127.0.0.1` / `::1`。

```bash
DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build
```

- `dev.sh` 会把 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST` 透传给本地 Go 或 Docker Go 进程。
- 也可以在“系统设置 -> 敏感配置与运行安全状态 -> external_service HTTP Allowlist”中新增后台配置来源的 exact origin。后台新增 / 删除需要系统设置写权限、风险确认、格式校验和审计；只允许 `http://host[:port]`，不允许 path、query、fragment、wildcard、`0.0.0.0` 或 CIDR。
- 保存配置、健康检查和实际投递都会按当前运行时环境重新执行同一套 HTTP 安全策略；如果重启后移除了 allowlist，已保存的非 localhost HTTP endpoint 也会被 `endpoint_safety_rejected` 拦截。
- 如果 DevHub 后端运行在 Docker 容器中，`127.0.0.1` 指向容器自身；宿主机上的 receiver 应使用 `host.docker.internal`、Docker host gateway 地址或宿主机局域网 IP。
- `token` 只在写入时提交，不要期望列表里能看到明文。
- `auth_type=none` 时不用填写 token。
- `auth_type=bearer` 时可以写入新 token；已有 token 只显示“已配置密钥 / 可替换”，不会回显明文。
- v1.8.4-S14 起 bearer token 会写入 Core SecretCenter，并以 `token_ref`（引用）形式绑定到 external_service 运行配置：`secret://external_service/{plugin_code}/token`。接口/审计/执行记录不回显 token 明文；external_service 查询与保存成功响应都会返回 `token_ref` 与 `token_secret` 元数据（status/key_id/last_used_at/rotated_at），不会返回明文、密文、token_hash 或 Authorization。
- v1.8.4-S18 起，系统设置 -> SecretCenter 的“敏感配置引用”页面会把这些 `secret://...` 引用按外部服务、Webhook Secret、Callback Token、测试数据等类型展示；该页面只显示 ref 和元数据，不显示 token / secret 明文。修改 external_service token 仍应回到对应插件的 external_service 配置中替换。
- v1.8.4-S17 补充，系统设置 -> 当前生效配置会展示当前实际运行的 external_service 配置：`endpoint_url`、`health_check_path`、`enabled`、`timeout_ms`、`failure_policy`、当前健康、最近成功和最近失败可以明文查看；token 只显示 `token_ref`、状态、`key_id`、脱敏值和提示，不显示 token 明文或 Authorization。页面提供“复制脱敏诊断信息”，复制内容可用于提交问题或排障，不包含 token / secret / Authorization / root key / `DEVHUB_PLUGIN_CONFIG_KEYS` / `encrypted_value`。
- v1.8.4-S19 起，SecretCenter 详情会显示 external_service、Webhook Secret、Callback Token 的真实来源和使用关系；禁用 / 吊销前必须先看影响预览，吊销需要输入完整 ref 强确认。轮换入口不会在 SecretCenter 收集明文：external_service token 跳回插件 external_service 配置，Webhook Secret 跳到 Webhook 密钥治理，Callback Token 跳到回调 Token 治理。当前生效配置页提供“去配置 / 健康检查 / 查看运行记录 / 查看 Secret / 查看审计”等排障入口。
- v1.9.0-S2 起，SecretCenter 详情和审计进一步补齐 `usage_type/source_type/source_id/source_code` 等脱敏来源字段，可按 `secret_ref`、namespace/key、来源类型、来源 ID 或 plugin_code 排查；当前生效配置页会展示 config source、token source 和 next_steps，用于提示缺 root key、缺 `token_ref`、HTTP allowlist 拒绝、Secret disabled/revoked 等下一步处理。
- v1.9.0-S3 起，当前生效配置页进一步显示 `auth_type`、endpoint origin、HTTP Allowlist 来源与匹配状态、`last_health_status/last_checked_at/last_error_at/last_error_summary`、`token_ref` namespace/name/usage/source 元数据、SecretCenter/root key 状态和 Webhook / Callback 安全摘要。非 localhost HTTP 未命中 allowlist 时会提示改 HTTPS 或新增 exact origin；Docker `127.0.0.1` caveat、健康检查失败、missing token_ref、secret disabled/revoked 也会进入 next_steps。复制的 `diagnostic_text` 仍只包含脱敏诊断信息，clipboard 失败时可从手动复制文本框取出。
- 插件全局 `config_json` 和 external_service 运行配置不是同一份配置。Webhook 健康检查和投递只读取 external_service 运行配置；在全局配置里写 `endpoint_url`、`health_check_path`、`token`、`timeout_ms` 或 `failure_policy` 不会自动影响投递。

## 3. 先跑健康检查

在插件详情的运行区域，`external_service` 模块里可以直接点“健康检查”。

健康检查会：

- 对 `{endpoint_url}{health_check_path}` 发 `GET`。
- 记录到 `hook_executions(service_type=external_service)`。
- 2xx 视为 healthy。
- 3xx 视为 warning。
- 4xx / 5xx / 超时 / 连接错误按失败策略处理。

如果插件当前是 `disabled`、`archived` 或 `soft_uninstalled`，不会真的打外部 endpoint，只会留下 skipped 记录。

健康摘要现在按“当前健康、当前 endpoint、最近成功、最近失败、24h / 7d 历史失败计数”分开展示。当前状态已经恢复 healthy 时，旧失败仍保留在运行记录中，但不会再覆盖顶部当前健康。

## 4. 看投递和异常

进入后台 `Webhook 治理`：

- `投递记录` 看每次 Hook 是否成功。
- `异常处理` 看重试队列、熔断状态和待处理异常。
- `外部服务执行` 看 `hook_executions(service_type=external_service)`，失败类记录可手动重试。
- `事件` 看事件链路。
- `Webhook 密钥` 看签名凭据的治理。
- `回调 Token` 看插件服务回调 Core 的凭据治理。
- `回调请求` 看插件服务实际回调记录。

常见判断：

- `retry_scheduled` / `retry_exhausted`：多半是超时、5xx 或 429。
- `skipped`：通常是插件状态、子站状态、endpoint 缺失或 token 缺失。
- `circuit_open`：先修外部服务，再去处理熔断。

## 4.1 手动重试失败投递

入口：

- 插件详情 -> 运行记录 -> 外部服务执行记录。
- Webhook 治理 -> 外部服务执行。

规则：

- 只对 `service_type=external_service` 的失败类记录显示“重试”。
- `success`、`skipped`、internal/builtin Hook、健康检查记录不显示重试入口，也会被后端拒绝。
- 点击前会提示：将重新向外部服务投递该事件，请确认外部服务具备幂等处理能力。
- 重试会创建新的执行记录，并在 metadata 标记 `manual_retry=true`、来源执行记录和 operator。
- 插件 disabled / archived / soft_uninstalled 或 external_service disabled 时不会实际调用 endpoint，会返回中文错误。

接口：

```text
POST /api/v1/admin/plugins/:code/hooks/executions/:execution_id/retry
```

权限：后台 admin token + `plugin.manage`。

返回里会包含本次 `retry_execution_id` / `retry_record_id`，可回到执行列表定位结果。

## 5. 配 Webhook Secret

如果你的插件服务要接收 DevHub 发出的 Webhook，就在 `Webhook 治理 -> Webhook 密钥` 管理密钥。

原则：

- 密钥明文只在创建或轮换时展示一次。
- 列表、详情、审计和执行记录都不回显明文。
- 轮换后要让接收端接受新的 `secret_ref`，再逐步结束旧密钥窗口。
- 创建/轮换需要启动加密密钥（`DEVHUB_PLUGIN_CONFIG_KEYS` 或 split 形式）已配置；该 root key 只能来自启动环境变量或外部 Secret 系统注入，后台不会保存、生成或修改，修改后需要重启生效。可在“系统设置 -> 敏感配置与运行安全状态”或 `/api/v1/admin/plugins/config-keys/status` 查看只读状态与示例环境变量。后台不能编辑 `DEVHUB_PLUGIN_CONFIG_KEYS`；系统设置页只允许受控管理 external_service HTTP allowlist 的后台配置来源，不会显示真实 token / secret / Authorization。

## 6. 配 Callback Token

如果插件服务要回调 DevHub Core API，就在 `Webhook 治理 -> 回调 Token` 创建令牌。

当前最小 scope：

- `config.read`
- `audit.write`

接口：

- `GET /api/v1/plugin-callback/config`
- `POST /api/v1/plugin-callback/audit-events`

说明：

- Token 不等于管理员权限。
- 只允许访问授权 scope 和授权 `community_scope`。
- 明文只在创建 / 轮换成功时返回一次。

## 7. 一个最短的实际流程

1. 使用 `examples/plugins/official_webhook_notify` 做本地插件包 dry-run，或打包 zip 后上传。
2. 走插件包治理：upload -> precheck -> promote -> install dry-run -> install。
3. 启动 receiver。feishu_link 独立本地联调服务推荐端口是 `18081`；仓库内官方 `cmd/webhook-mock-receiver` 是另一个工具，默认端口为 `18090`。
4. 在插件详情 -> 运行记录里配置 `endpoint_url=http://127.0.0.1:18081`、`auth_type=none`。如果 DevHub 后端在 Docker 容器中而 receiver 在宿主机，改用 `http://172.17.0.1:18081`、`host.docker.internal` 或宿主机局域网 IP，并按上面的 allowlist 命令重启。
5. 点一次健康检查，确认状态是 healthy。
6. 触发一次业务事件，例如创建内容。
7. 去 `Webhook 治理 -> 外部服务执行` 看 `AfterCreateContent` 投递记录。
8. 停掉 mock receiver 或让它返回 500，确认失败记录可手动重试。
9. 如果插件服务还要读配置或写审计，再创建 Callback Token。

注意：feishu_link 本地测试服务只记录 `POST /hooks/*`，`GET /health` 成功不代表 `/requests` 一定有记录。触发 `AfterCreateContent` 后再看 `/requests`，应出现 `POST /hooks/content.after_create`，并且只展示 `authorization_present=true`，不会保存 Authorization 明文。

回归记录（v1.8.4-S12，2026-05-28）：已在 Docker 后端场景验证 `http://127.0.0.1:*` connection refused 会给出 Docker loopback 提示；`http://172.17.0.1:*` 未 allowlist 会被 `endpoint_safety_rejected` 拒绝并展示推荐命令；设置 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081` 重启后保存配置成功、health check `healthy`、AfterCreateContent 投递可在 receiver `/requests` 看到记录，且不回显 Authorization 明文。联调过程中修复后台 `POST /api/v1/admin/posts` 在设置非默认 status 的边界场景可能 panic 的问题。

回归记录（v1.9.0-S1，2026-05-29）：`scripts/run-feishu-webhook-flow.sh` 新增 `DEVHUB_WEBHOOK_FLOW=full`，用于 fresh standalone receiver 一次性覆盖 success、500 failure、timeout 和 manual retry。示例：

```bash
DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://host.docker.internal:18082 ./dev.sh restart --mysql --no-build
DEVHUB_WEBHOOK_PORT=18082 \
DEVHUB_WEBHOOK_ENDPOINT=http://host.docker.internal:18082 \
DEVHUB_WEBHOOK_FLOW=full \
DEVHUB_WEBHOOK_AUTH_TYPE=bearer \
DEVHUB_WEBHOOK_TOKEN=<测试 token> \
./scripts/run-feishu-webhook-flow.sh
```

该脚本会验证 bearer token 保存后只展示 `token_ref=secret://external_service/{plugin_code}/token` / `token_secret` 元数据，并扫描 external_service 保存响应、health check、`hook_executions`、manual retry 和 `admin_logs` 响应，确认不回显 token / Authorization / `encrypted_value` / token_hash。MySQLStore 实跑还应配套 DB 扫描 `secret_refs`、`plugin_external_services`、`admin_logs` 和 `hook_executions`。

## 8. 排障顺序

先看三件事：

1. 插件状态是不是 `enabled`。
2. `external_service` 配置是不是已经存在。
3. `Webhook 治理` 里有没有投递记录或失败原因。

常见诊断：

- `endpoint_safety_rejected`：当前 endpoint 是非 localhost HTTP，默认安全策略拒绝；使用 HTTPS，或为本地开发配置 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST`。
- `network_connection_refused` + `http://127.0.0.1:*`：如果 DevHub 后端在 Docker 容器中，`127.0.0.1` 指向容器自身。改用 `host.docker.internal`、Docker host gateway 地址或宿主机局域网 IP。
- `network_timeout`：检查 receiver 是否启动、端口是否可达、Docker 网络是否连通。
- `token_missing` / `token_invalid`：重新写入 external_service Bearer Token 或检查插件配置加密 key；响应和日志不会回显 token 明文。

再看这些常见问题：

- 没有投递：大多是插件未启用、子站未启用、endpoint 缺失或 `enabled=false`。
- 一直重试：大多是超时、5xx 或 429。
- 一直 401/403：多半是认证方式或 token 有问题。
- 看不到凭据：这是正常的，凭据不回显明文。

## 9. 管理员在后台该去哪里

- 看整体状态：`插件总览`
- 看这个插件：`插件列表` -> `查看详情`
- 看外部服务：`插件详情` -> `运行`
- 看投递链路：`Webhook 治理`
- 看操作和审计：`运行记录 / 审计`

## 10. 这套能力还没做什么

- 没有完整第三方运行时。
- 没有 blocking Hook。
- 没有远程 iframe URL。
- 没有第三方代码执行。
- 没有把 Secret / Token 明文开放给前端表格。
