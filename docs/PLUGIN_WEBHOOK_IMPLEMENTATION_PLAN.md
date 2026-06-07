# Webhook / HTTP 插件服务协议实现拆解（v1.7.3 计划）

[返回文档入口](README.md)

更新时间：2026-05-16

本文将 `docs/PLUGIN_WEBHOOK_PROTOCOL.md` 中的 Webhook / HTTP 插件服务协议设计拆解为可落地、可验收、可分阶段执行的实现任务清单。优先级以 non_blocking delivery 为第一优先级，不在第一阶段引入 blocking Hook。本轮为 Webhook 协议实现拆解与官方示例插件验证准备任务，只修改文档，未修改代码，未执行测试、构建或 E2E。

## 阶段 0：协议口径确认（文档对账）

目标：确认协议设计已覆盖关键对象与安全边界，并产出缺口清单与实现优先级。

对账项：

1. Plugin Service
2. Webhook Event
3. Webhook Delivery
4. Plugin Action
5. Plugin Callback
6. 签名鉴权
7. 幂等
8. 重试
9. 超时
10. 限流
11. 熔断
12. 审计
13. 后台治理
14. 测试规划

输出：

- 文档缺口清单（如有）
- 实现优先级清单
- 明确不实现内容清单

结论（历史 v1.7.2 文档状态，当前运行口径以 v1.8.3 为准）：

- 协议设计已覆盖上述要点，详见 `docs/PLUGIN_WEBHOOK_PROTOCOL.md`。
- 第一优先级：non_blocking delivery + delivery 记录 + 重试队列 + 熔断 + 审计 + 后台治理入口。
- blocking Hook 明确后置（需要事务/回滚/用户错误提示/降级策略更严谨设计）。

## 阶段 1：non_blocking delivery 最小闭环（P0）

目标：实现非阻塞 Webhook 投递的最小可用闭环。

范围：

1. 定义 Webhook Event 模型。
2. 定义 Webhook Delivery 模型。
3. 支持 non_blocking Hook 投递（只投递，不阻断主流程）。
4. 投递结果可查询。
5. 投递写审计。
6. disabled / soft_uninstalled 插件不投递。
7. 不实现 blocking Hook。
8. 不实现插件回调 Core API。
9. 不执行插件代码。

建议数据结构（仅设计，后续落库）：

`webhook_events`：

- `id`
- `event_id`
- `plugin_code`
- `hook_name`
- `mode`（固定 `non_blocking`）
- `event_type`
- `resource_type` / `resource_id`
- `community_id`
- `actor_type` / `actor_id`
- `payload_json`
- `metadata_json`
- `status`（`pending|delivering|delivered|failed|skipped`）
- `occurred_at`
- `created_at`

`webhook_deliveries`：

- `id`
- `delivery_id`
- `event_id`
- `plugin_code`
- `target_url`
- `status`（`pending|sending|success|failed|retry_scheduled|retry_exhausted|skipped`）
- `attempt` / `max_attempts`
- `request_body_sha256`
- `response_status`
- `response_body_excerpt`
- `error_message`
- `duration_ms`
- `next_retry_at`
- `started_at` / `finished_at`
- `created_at` / `updated_at`

验收标准：

1. non_blocking Hook 可以生成 event。
2. event 可以生成 delivery。
3. delivery 可以记录 success/failed。
4. failed 不阻断主流程。
5. delivery 可查询。
6. 审计有记录。
7. disabled 插件不投递。
8. soft_uninstalled 插件不投递。

## 阶段 2：重试队列（P0+）

目标：实现 non_blocking delivery 的失败重试机制。

范围：

1. delivery 失败后进入 `retry_scheduled`。
2. 支持 `next_retry_at`。
3. 支持 `max_attempts`。
4. 支持指数退避或固定退避。
5. 4xx 默认不重试。
6. 5xx / timeout 可重试。
7. 429 按 `Retry-After` 设计。
8. 超过最大次数后 `retry_exhausted`。
9. 重试记录可查询。
10. 重试写审计。

默认重试策略（建议）：

- attempt 1：立即
- attempt 2：1 分钟后
- attempt 3：5 分钟后
- attempt 4：15 分钟后
- attempt 5：1 小时后

实现建议：

- 可先做“扫描式重试”（定时扫描 `next_retry_at <= now()` 的 delivery），不引入复杂队列系统。
- 如项目已有任务机制，优先复用；否则先提供手动触发重试入口（后台/接口）作为过渡。

验收标准：

1. 5xx 失败进入 `retry_scheduled`。
2. timeout 进入 `retry_scheduled`。
3. 4xx 不重试。
4. 超过 max_attempts 后 `retry_exhausted`。
5. 每次重试 attempt 递增并记录。
6. 重试不重复生成 event_id。
7. idempotency_key 保持稳定（同 event）。
8. 重试审计完整。

实现状态（v1.7.5）：

- 已落地：delivery 状态扩展 `retry_scheduled/retry_exhausted`、`next_retry_at` 调度、attempt/max_attempts、以及 429 `Retry-After` 优先策略。
- 已提供：DB 扫描式 `retry-due` 批量重试接口与单条手动重试接口（不引入外部队列系统）。

## 阶段 3：熔断机制（P1）

目标：防止异常插件服务持续拖累投递系统。

范围：

1. 统计插件服务连续失败次数。
2. 达到阈值后 `circuit_open`。
3. `circuit_open` 时暂停投递。
4. 暂停期间 delivery 标记 `skipped/circuit_open`。
5. 管理员可查看熔断状态与原因。
6. 管理员可手动恢复（close）。
7. 支持 `half_open` 探测（可选）。
8. 熔断写审计。

建议数据结构（仅设计）：

`webhook_circuit_breakers`：

- `id`
- `plugin_code`
- `target_url`
- `status`（`closed|open|half_open`）
- `failure_count`
- `success_count`
- `opened_at` / `closed_at`
- `next_probe_at`
- `last_error_message`
- `created_at` / `updated_at`

默认策略（建议）：

1. 连续失败 5 次打开熔断。
2. 熔断 10 分钟后允许一次 half_open 探测。
3. 探测成功关闭熔断；失败继续 open。
4. 管理员可手动关闭熔断。

验收标准：

1. 连续失败达到阈值后 circuit_open。
2. circuit_open 后不继续投递。
3. delivery 标记 skipped/circuit_open。
4. 管理员可查看熔断原因。
5. 管理员可手动恢复。
6. 熔断操作有审计。
7. 熔断不影响 Core 主流程。

实现状态（v1.7.5）：

- 已落地：`plugin_code + target_url` 维度的 circuit breaker（`closed/open/half_open`），默认连续失败阈值 5，`open` 后 10 分钟允许一次 probe（half_open）。
- 已提供：后台查询熔断状态、手动 close/open 熔断；熔断只暂停投递，不禁用插件本身。

## 阶段 4：Webhook 签名与鉴权（P1）

目标：保护 DevHub → 插件服务的请求，减少伪造/篡改/重放风险。

范围：

1. 使用 HMAC-SHA256。
2. 请求 headers 包含：
   - `X-DevHub-Event-ID`
   - `X-DevHub-Delivery-ID`
   - `X-DevHub-Plugin-Code`
   - `X-DevHub-Timestamp`
   - `X-DevHub-Signature`
   - `X-DevHub-Signature-Alg`
   - `X-DevHub-Idempotency-Key`
   - `X-DevHub-Request-ID`
3. 签名内容：`timestamp + "." + method + "." + path + "." + body_sha256`
4. 支持 `secret_ref`（引用，不明文展示）
5. 支持 secret 轮换设计（版本化 ref）
6. timestamp 防重放（5 分钟窗口）
7. 签名失败记录 delivery failed（并写审计）

阶段说明：

- 第一批可只实现 Core 发送端签名。
- 接收端验签可在官方示例插件服务桩中实现，用于验证签名设计（仍不执行第三方代码）。

验收标准：

1. 请求生成签名 header。
2. body 改变后签名不一致。
3. timestamp 参与签名。
4. delivery 记录 request_body_sha256。
5. secret 不明文展示。
6. 签名相关审计完整。

实现状态（v1.7.6）：

- 已落地：DevHub 发送端 HMAC-SHA256 签名（`timestamp.method.path.body_sha256`），delivery 记录持久化 `signature_alg/secret_ref/body_sha256/signature_status/signed_at/signature_error`（不记录 secret 明文）。
- 已落地：Webhook Secret 管理（create/rotate/disable/enable/revoke）与 active/previous grace window；Secret 明文只在创建/轮换成功响应中返回一次。
- 已提供：后台治理入口（Webhook 治理页内新增 Secrets Tab）与对应 Admin API（`/api/v1/admin/plugins/webhooks/secrets*`）。

## 阶段 5：后台治理入口（P1）

目标：让管理员可观测 Webhook 运行状态并进行必要操作。

最小治理入口（建议）：

1. Webhook 事件列表（event list）
2. delivery 列表
3. delivery 详情
4. 最近失败
5. 重试状态
6. 熔断状态
7. 手动重试失败 delivery（仅 non_blocking）
8. 手动恢复熔断
9. 支持按 plugin_code/status/hook_name 筛选

验收标准：

1. 可查看 event。
2. 可查看 delivery。
3. 可见失败原因。
4. 可见 retry_count / attempt。
5. 可见 next_retry_at。
6. 可见 circuit_open。
7. 可手动重试失败 delivery。
8. 可手动恢复 circuit breaker。
9. UI 不白屏。

## 阶段 6：blocking Hook（明确后置）

目标：将 blocking Hook 放到后续阶段，不进入 non_blocking delivery 第一批实现范围。

原因：

- blocking Hook 会阻断发布/审核/更新等主流程，需要更严谨的超时、回滚、错误与降级策略。

后续需要补齐的设计/实现项：

1. blocking Hook 超时策略（短超时 + 明确 fail-open/fail-closed）
2. blocking Hook 错误返回规范（用户可读）
3. blocking Hook 与事务边界（避免半状态）
4. blocking Hook 与审计/追踪
5. blocking Hook 与降级策略（熔断时行为）
6. 是否允许第三方插件使用 blocking（默认不开放或需白名单）

### v1.7.9：blocking Hook 设计评估补充（不实现）

结论：blocking Hook 可以进入“设计收口阶段”，但**不建议直接进入实现阶段**。外部 HTTP blocking Hook 会直接影响主流程可用性与延迟，应当以“默认关闭 + 白名单 + 短超时 + 强审计 + 明确降级”为前提。

必须回答（设计要点）：

1. 哪些 Hook 可以考虑 blocking：仅极少数 `before_*`（内容创建/更新/审核前的风控/合规校验）。
2. 哪些 Hook 禁止 blocking：所有 `after_*`（通知、SEO、索引、统计等）必须 non_blocking。
3. 默认超时：建议 2s（最大 5s，可配置）。
4. 超时策略：按 Hook 分类明确 fail-open / fail-closed（风控类默认 fail-closed，体验增强类 fail-open）。
5. 用户提示：blocking deny/timeout 必须返回用户可读错误，并写审计（不得泄露 secret/token）。
6. 事务边界：blocking Hook 不应在 DB 事务内长时间等待外部 HTTP；建议在 DB 写入前完成 blocking 校验。
7. 是否允许外部 HTTP 插件使用：默认不开放；如开放需更强信任模型。
8. 是否仅允许官方/可信发布者：建议仅 `official/trusted`，并需要管理员显式启用。
9. 是否需要更高 trust_level：是（至少 trusted）。
10. 是否需要熔断降级：需要（避免外部服务抖动拖垮主流程）。
11. 审计：blocking Hook 必须单独审计（包括 decision/latency/error）。
12. 显式开关：必须（hook 级别/插件级别/社区级别至少一种，默认关闭）。
13. 与重试：blocking Hook **禁止后台自动重试**（避免重复用户动作）。
14. 性能影响：需要预算（P95/P99）与限流策略；必须可观测。

最小实现前置条件（建议 v1.8.0 前完成）：

- Hook 白名单/黑名单 + 默认关闭开关
- 统一超时与降级策略
- 事务边界明确（不会产生半事务状态）
- 审计/可观测性（可定位是谁阻断、为何阻断、耗时）
- 管理后台治理入口（开关/最近阻断/最近超时）

## 官方示例插件（端到端验证）关联点

官方公告插件端到端验证方案见 `docs/plugins/official-announcement-plugin.md`。

验证重点（按阶段）：

- 阶段 1：生成 event + delivery，失败不阻断。
- 阶段 2：5xx/timeout 触发重试与 attempt 递增。
- 阶段 3：连续失败触发熔断并暂停投递。
- 阶段 4：签名 header 生成与服务端验签。
- 阶段 5：后台可观测、可筛选、可手动重试/恢复熔断。
