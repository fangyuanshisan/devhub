# DevHub Webhook / HTTP 插件服务协议（设计）

[返回文档入口](README.md)

更新时间：2026-05-16

本文档为 **Webhook / HTTP 插件服务协议设计文档**，用于连接 DevHub Core 与外部插件服务。

实现状态补充：

- 协议主体仍以“设计”为主，但 v1.7.5~v1.7.8 已分阶段落地部分治理能力（见下方条目）；这不代表完整第三方运行模型已完成。
- v1.7.5 已实现 non_blocking delivery 的治理能力增强（delivery 记录、重试调度、circuit breaker、最小后台治理入口与审计），但不代表完整协议已落地。
- v1.7.6 已实现 DevHub 发送端 HMAC-SHA256 签名与 Webhook Secret 管理/轮换（仅治理与签名，不执行第三方代码）。
- v1.7.7 已实现“插件服务回调 Core API”的最小通道：callback token（Bearer）+ scope 白名单 + community scope 校验 + callback request 记录与审计（不等于完整插件 SDK/运行时）。
- v1.7.8 已补齐 Webhook 后台治理的 Events 视图（Admin API + UI Tab），并提供仓库内官方 mock receiver（`cmd/webhook-mock-receiver`）用于端到端验签/失败注入/重试熔断验证（不执行第三方代码）。

重要边界：

- DevHub **不加载**第三方插件代码。
- DevHub **不执行**插件包内脚本。
- DevHub **不加载** Go plugin / `.so` / `.dll` / `.node` / wasm。
- 第三方插件后端以 **独立 HTTP 服务**运行。
- DevHub Core 通过受控 HTTP 协议调用插件服务；插件服务通过受控 Core API 回调 DevHub。
- 插件不能直接访问 DevHub 数据库，不能绕过 Core 权限、审计和生命周期治理。
- 本文的协议与对象均为设计，不代表已经实现投递队列、重试、secret 管理或 token 系统。

本协议与运行模型关系：

- 运行模型总体说明见 `docs/PLUGIN_RUNTIME_MODEL.md`。
- 本文聚焦“HTTP 插件服务 + Webhook 投递”的协议细节：签名、幂等、重试、审计、限流与隔离。

## 1. 定位与原则

Webhook / HTTP 插件服务是 DevHub 第三方插件运行模型的推荐方向之一。推荐表述：

> Webhook / HTTP 插件服务协议用于连接 DevHub Core 与外部插件服务。DevHub Core 负责用户、权限、内容、配置、审计、生命周期和安全边界；外部插件服务只通过声明的 Hook、Action 和受控 API 与 Core 交互。

核心原则：

1. Core 主导：Core 决定何时投递、投递哪些事件、是否 blocking、超时与重试策略。
2. 插件受控：插件服务只能接收 manifest 声明过的 hook/action，并只能回调授权的 Core API scopes。
3. 可审计：每次投递、失败、重试、熔断、回调都必须有可追踪的审计与投递记录。
4. 可隔离：插件服务故障不能拖垮 Core；non_blocking 默认不影响主流程；blocking 必须短超时且谨慎使用。

## 2. 协议对象（设计）

### 2.1 Plugin Service

表示一个外部插件服务（未来可落库/配置，但本轮不实现）。

字段建议：

- `plugin_code`
- `service_url`
- `health_check_url`
- `manifest_url`
- `status`：`enabled|disabled|circuit_open`
- `timeout_ms`
- `retry_policy`
- `auth_type`：`hmac_sha256|token`（设计，默认推荐 `hmac_sha256`）
- `signing_secret_ref`（引用，不存明文）
- `allowed_events`
- `allowed_actions`
- `api_scopes`
- `created_at`
- `updated_at`

### 2.2 Webhook Event

表示 DevHub 发给插件服务的事件（一个业务事件，多次投递共享同一 `event_id`）。

字段建议：

- `event_id`（全局唯一，建议 `evt_...`）
- `event_name` / `hook_name`
- `event_type`（例如 `hook|action|lifecycle`）
- `plugin_code`
- `mode`：`blocking|non_blocking`
- `community_id`
- `actor_type` / `actor_id`
- `resource_type` / `resource_id`
- `occurred_at`
- `payload`
- `request_id`（Core 侧 request trace）
- `idempotency_key`（供插件服务去重）

### 2.3 Webhook Delivery

表示一次投递记录（同一 event 多次重试会产生多个 delivery 记录）。

字段建议：

- `delivery_id`（唯一，建议 `del_...`）
- `event_id`
- `plugin_code`
- `target_url`
- `status`：`created|success|failed|retry_scheduled|retry_exhausted|skipped|circuit_open`
- `attempt` / `max_attempts`
- `request_headers`（可脱敏存储）
- `request_body_sha256`
- `response_status`
- `response_body_excerpt`（长度限制 + 脱敏）
- `error_message`
- `started_at` / `finished_at` / `duration_ms`
- `next_retry_at`

### 2.4 Plugin Action

表示 Core 主动请求插件执行某个动作（例如“刷新索引/生成摘要”等）。

字段建议：

- `action_name`
- `plugin_code`
- `actor_type` / `actor_id`
- `input`
- `timeout_ms`
- `idempotency_key`

### 2.5 Plugin Callback

表示插件服务回调 DevHub Core API 的身份模型。

实现状态：

- v1.7.7 已落地 **最小** callback token 与 scopes（仅 `config.read`、`audit.write`）以及 `/api/v1/plugin-callback/*` 的最小回调 API。
- 仍未实现：插件代表用户操作（actor 代理）、更完整的 scopes、插件回调通用 Core API、SDK/模板自动注入等。

字段建议：

- `plugin_code`
- `install_id`
- `publisher_id`
- `scope`
- `actor_type` / `actor_id`
- `request_id`
- `api_path`
- `method`
- `payload`

## 3. 事件类型（第一阶段，设计）

说明：以下事件类型仅为设计，不代表已实现对外投递。

### 3.1 内容相关

- `content.before_create`：blocking 可选；允许修改 payload：否；默认超时 2s；失败策略：blocking 失败阻断；需审计：是
- `content.after_create`：non_blocking 推荐；允许修改 payload：否；默认超时 5s；失败策略：重试 + 最终审计；需审计：是
- `content.before_update`：blocking 可选；允许修改 payload：否；默认超时 2s；失败策略：阻断；需审计：是
- `content.after_update`：non_blocking 推荐；允许修改 payload：否；默认超时 5s；失败策略：重试；需审计：是
- `content.before_delete`：blocking 可选；允许修改 payload：否；默认超时 2s；失败策略：阻断；需审计：是
- `content.after_delete`：non_blocking 推荐；允许修改 payload：否；默认超时 5s；失败策略：重试；需审计：是
- `content.before_moderate`：blocking 可选；允许修改 payload：否；默认超时 2s；失败策略：阻断；需审计：是
- `content.after_moderate`：non_blocking 推荐；允许修改 payload：否；默认超时 5s；失败策略：重试；需审计：是

### 3.2 评论相关

- `comment.after_create`：non_blocking；默认超时 5s；失败策略：重试；需审计：是
- `comment.after_delete`：non_blocking；默认超时 5s；失败策略：重试；需审计：是

### 3.3 标签相关

- `tag.after_bind`：non_blocking；默认超时 5s；失败策略：重试；需审计：是
- `tag.after_unbind`：non_blocking；默认超时 5s；失败策略：重试；需审计：是

### 3.4 搜索相关

- `search.index_requested`：non_blocking；默认超时 5s；失败策略：重试；需审计：是

### 3.5 通知相关

- `notification.before_send`：blocking 不推荐，默认 non_blocking；允许修改 payload：否；默认超时 5s；失败策略：不阻断（除非显式 blocking）；需审计：是
- `notification.after_send`：non_blocking；默认超时 5s；失败策略：重试；需审计：是

### 3.6 SEO 相关

- `seo.before_build`：blocking 不推荐；默认 non_blocking；默认超时 5s；失败策略：不阻断；需审计：是
- `seo.after_build`：non_blocking；默认超时 5s；失败策略：重试；需审计：是

### 3.7 插件生命周期相关

- `plugin.after_installed`：non_blocking；默认超时 5s；失败策略：重试；需审计：是
- `plugin.before_enabled`：blocking 可选（慎用）；默认超时 2s；失败策略：阻断启用；需审计：是
- `plugin.after_enabled`：non_blocking；默认超时 5s；失败策略：重试；需审计：是
- `plugin.before_disabled`：blocking 可选；默认超时 2s；失败策略：阻断禁用或按策略；需审计：是
- `plugin.after_disabled`：non_blocking；默认超时 5s；失败策略：重试；需审计：是
- `plugin.before_uninstalled`：blocking 可选；默认超时 2s；失败策略：阻断或按策略；需审计：是
- `plugin.after_uninstalled`：non_blocking；默认超时 5s；失败策略：重试；需审计：是
- `plugin.before_upgraded`：blocking 可选；默认超时 2s；失败策略：阻断升级；需审计：是
- `plugin.after_upgraded`：non_blocking；默认超时 5s；失败策略：重试；需审计：是

## 4. blocking / non_blocking 策略

### 4.1 blocking Hook

规则：

1. Core 等待插件响应后才继续主流程。
2. 插件响应失败或超时，主流程默认阻断（由 hook 策略决定）。
3. 必须设置短超时（建议默认 2s，最大 5s）。
4. 必须写审计，并保存 delivery 记录。

适合场景（设计建议）：

- `content.before_create`、`content.before_update`、`content.before_moderate`
- `config_validate`（action 形式）
- `permission_check`（action 形式）

风险提示：

- blocking Hook 不能滥用；外部插件服务不可用会直接影响用户操作。
- 默认推荐第三方插件优先使用 non_blocking。

### 4.2 non_blocking Hook

规则：

1. Core 不等待插件最终成功完成（可异步投递）。
2. 插件失败不影响主流程，但必须记录失败与审计。
3. 可按策略重试，直到达到最大次数。

适合场景：

- `content.after_*`
- 通知、分析、搜索索引、SEO after build

## 5. HTTP 请求与响应格式（设计）

### 5.1 Hook 投递请求

请求：

`POST {service_url}/hooks/{hook_name}`

Headers：

- `X-DevHub-Event-ID`
- `X-DevHub-Delivery-ID`
- `X-DevHub-Plugin-Code`
- `X-DevHub-Timestamp`
- `X-DevHub-Signature`
- `X-DevHub-Signature-Alg`：`HMAC-SHA256`
- `X-DevHub-Idempotency-Key`
- `X-DevHub-Request-ID`
- `Content-Type: application/json`

Body（示例）：

```json
{
  "schema_version": "1",
  "event_id": "evt_xxx",
  "delivery_id": "del_xxx",
  "hook_name": "content.after_create",
  "mode": "non_blocking",
  "plugin_code": "demo",
  "occurred_at": "2026-05-16T00:00:00Z",
  "actor": { "type": "user", "id": 123 },
  "community": { "id": 1, "slug": "php" },
  "resource": { "type": "topic", "id": 1001 },
  "payload": {},
  "metadata": { "request_id": "req_xxx" }
}
```

响应（通用）：

```json
{ "ok": true, "message": "accepted", "result": {} }
```

失败：

```json
{ "ok": false, "error_code": "PLUGIN_REJECTED", "message": "reason", "details": {} }
```

blocking 决策响应（示例）：

```json
{ "ok": true, "decision": "allow" }
```

或：

```json
{ "ok": false, "decision": "deny", "message": "blocked reason" }
```

### 5.2 Action 调用请求

请求：

`POST {service_url}/actions/{action_name}`

Body 建议包含：

- `schema_version`
- `action_name`
- `plugin_code`
- `actor` / `community`（可选）
- `input`
- `idempotency_key`
- `request_id`

响应：

- `ok=true` 返回 `result`
- `ok=false` 返回 `error_code/message/details`

## 6. 鉴权、签名与防重放（设计）

### 6.1 HMAC-SHA256 签名

推荐签名算法：`HMAC-SHA256`。

签名串（建议）：

`timestamp + "." + method + "." + path + "." + body_sha256`

其中：

- `timestamp`：`X-DevHub-Timestamp`
- `method`：HTTP method
- `path`：请求 path（不含 host）
- `body_sha256`：请求 body 的 sha256（hex）

Headers：

- `X-DevHub-Timestamp`
- `X-DevHub-Signature`
- `X-DevHub-Signature-Alg: HMAC-SHA256`

要求：

1. 插件服务必须验证签名，失败返回 401。
2. Core 使用插件安装时分配的 secret 进行签名；secret 不明文展示。
3. secret 支持轮换（设计：`secret_ref` + version）。
4. timestamp 防重放：默认允许时间偏差 5 分钟。
5. `event_id` / `delivery_id` / `idempotency_key` 必须进入审计字段，便于追踪。

### 6.2 防重放

插件服务侧要求：

- 校验 timestamp 在允许窗口内。
- 记录已处理的 `idempotency_key`（或 `event_id + idempotency_key`）并拒绝重复处理。

Core 侧要求：

- 同一事件重试保持相同 `event_id` 与 `idempotency_key`。
- delivery_id 每次尝试可变化，但都关联到同一 `event_id`。

## 7. 幂等与重试（设计）

幂等规则：

1. 每个业务事件有 `event_id`。
2. 每次投递有 `delivery_id`。
3. 同一事件多次重试必须保持同一 `event_id`。
4. `idempotency_key` 用于插件服务侧去重，插件服务需持久化已处理 key。

重试建议：

### 7.1 non_blocking

- attempt 1：立即
- attempt 2：1 分钟后
- attempt 3：5 分钟后
- attempt 4：15 分钟后
- attempt 5：1 小时后
- 超过次数：标记 `retry_exhausted` 并写审计

HTTP 状态码策略（建议）：

- `2xx`：成功，不重试
- `4xx`：默认不重试（鉴权/签名失败、非法 payload 等）
- `5xx` / timeout：可重试
- `429`：按 `Retry-After`（若存在）重试，否则按退避策略重试

### 7.2 blocking

- 默认不做后台重试：用户请求内最多一次调用。
- 超时/失败按策略直接返回错误或阻断主流程。
- 可设计“人工重试”入口，但不得自动重复用户操作。

## 8. 超时、限流与熔断（设计）

超时建议：

- blocking hook：默认 2s，最大 5s
- non_blocking hook：默认 5s，最大 15s
- health check：默认 1s

限流建议：

- 每插件每分钟最大事件数（全局）
- 每插件并发投递数
- 每社区每插件事件限制
- 重试队列容量上限（避免雪崩）

熔断建议：

1. 连续失败达到阈值后，进入 `circuit_open`，暂停投递（写 `skipped/circuit_open`）。
2. 暂停期间后台显示插件服务不健康，管理员可手动恢复或等待自动半开探测（设计）。
3. 熔断不影响 Core 主进程。
4. blocking hook 熔断策略需要显式定义：默认建议“拒绝（fail closed）”或“跳过（fail open）”必须按 hook 类型配置。

## 9. 插件回调 Core API（设计）

原则：

1. 插件不能直接访问数据库，只能调用受控 API。
2. 插件调用必须携带插件身份并受 `api_scopes` 约束。
3. 插件调用必须写审计。
4. 插件 token 可吊销；token 不等同管理员 token。
5. 插件代表用户操作时必须携带 actor，并受用户权限约束；插件作为系统操作时受 scope 限制。

建议身份字段（审计维度）：

- `plugin_code`
- `install_id`
- `publisher_id`
- `token_id`
- `scope`
- `actor_type` / `actor_id`
- `community_id`
- `request_id`

注意：本轮只做设计，不实现 token 系统或 callback API。

## 10. Manifest 字段（设计）

以下字段为设计字段，不代表已实现：

```json
{
  "runtime": {
    "type": "http_service",
    "service_url": "https://plugin.example.com",
    "health_check": "/health",
    "timeout_ms": 3000
  },
  "webhooks": [
    {
      "hook": "content.after_create",
      "url": "/hooks/content.after_create",
      "mode": "non_blocking",
      "retry": true,
      "timeout_ms": 5000
    }
  ],
  "api_scopes": ["content.read", "notification.send"],
  "auth": { "type": "hmac", "secret_ref": "plugin-secret" }
}
```

与 SDK/模板关系（设计）：

- SDK/模板未来可以生成上述字段，但必须明确“不会生成 secret 明文、不会默认启用 blocking hook”。
- Core 侧实现后，仍需管理员审核并授权 `api_scopes` 与 webhook 订阅范围。

## 11. 后台治理能力（设计）

本轮只设计，不实现 UI。建议未来治理页/Tab：

1. 插件服务状态（health + circuit）
2. Webhook 事件订阅（allowed_events/allowed_actions）
3. 投递记录（delivery list/detail）
4. 最近失败与错误聚合
5. 重试队列
6. 熔断状态与恢复入口
7. 手动重试（仅 non_blocking）
8. secret 轮换（ref/version）
9. API scopes 授权视图
10. 审计日志跳转

## 12. 审计动作（设计）

建议审计 action：

- `plugin.webhook.delivery.created`
- `plugin.webhook.delivery.success`
- `plugin.webhook.delivery.failed`
- `plugin.webhook.delivery.retry_scheduled`
- `plugin.webhook.delivery.retry_exhausted`
- `plugin.webhook.delivery.skipped`
- `plugin.webhook.circuit.opened`
- `plugin.webhook.circuit.closed`
- `plugin.webhook.signature.failed`
- `plugin.webhook.callback.accepted`
- `plugin.webhook.callback.rejected`
- `plugin.webhook.secret.rotated`

metadata 建议包含：

- `plugin_code`
- `hook_name`
- `event_id`
- `delivery_id`
- `mode`
- `target_url`
- `response_status`
- `duration_ms`
- `error_message`（脱敏）
- `retry_count`
- `actor_type` / `actor_id`
- `community_id`
- `request_id`

## 13. 后续实现任务拆分（建议）

仅作为后续任务建议（不在本轮实现）：

1. 插件服务注册/配置与 health 探测。
2. HMAC secret 管理与轮换（不出现在 UI/API 明文）。
3. delivery 持久化与重试队列（non_blocking）。
4. circuit breaker 与治理入口。
5. 插件受控 API token 与 scopes 授权模型。
6. HookBus 远程投递适配（blocking/non_blocking）。
7. 官方示例插件（公告/友情链接）验证 end-to-end。

## 14. 本轮检查记录

- 已新增/更新 Webhook 协议文档：是。
- 已明确 HTTP 插件服务是推荐第三方运行方式：是。
- 已明确 DevHub 不加载第三方代码：是。
- 已明确签名/鉴权、防重放、重试、幂等、超时、限流、熔断：是（设计）。
- 已明确插件回调 Core API 受控模型：是（设计）。
- 本轮为 Webhook / HTTP 插件服务协议设计任务，只修改文档，未修改代码，未执行测试、构建或 E2E。
