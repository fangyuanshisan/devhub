# 官方示例插件：公告插件（端到端验证方案）

[返回文档入口](../README.md)

更新时间：2026-05-16

本文定义官方示例插件“公告插件（official_announcement）”的端到端验证方案，用于后续验证 Webhook / HTTP 插件服务协议的实现（non_blocking delivery 优先）。本轮为 Webhook 协议实现拆解与官方示例插件验证准备任务，只修改文档，未修改代码，未执行测试、构建或 E2E。

重要边界：

- 官方示例插件用于“协议验证与端到端流程验证”，不用于执行不可信第三方代码。
- 不做 Go 动态加载，不运行外部 package scripts，不加载 `.so/.dll/.node/wasm`。
- 外部插件服务以“官方提供的模拟服务/测试桩”形态存在，用于验签、回包、失败注入与熔断验证。

## 1. 插件定位

插件编码：`official_announcement`

目标：通过公告类业务（简单、可见、可配置）验证：

- 菜单/页面入口（未来 iframe 挂载，当前只写设计）。
- 插件配置（声明 `config_schema` 与默认配置）。
- Hook non_blocking 投递（content.after_create 等）。
- delivery 记录、重试队列、熔断、审计、后台治理入口。

## 2. 公告插件能力设计（不涉及第三方代码执行）

业务能力（设计）：

1. 后台配置公告内容（文本/Markdown）。
2. 前台首页或站点顶部展示公告（slot：`frontend.home.section` 或 `frontend.header.nav`）。
3. 支持启用/禁用公告（配置字段）。
4. 支持按社区显示公告（community config）。
5. 支持公告点击统计（后续阶段，可用事件/审计替代）。

协议验证能力（用于 E2E 验证）：

1. 接收 `content.after_create` non_blocking 事件。
2. 接收 `plugin.after_enabled` / `plugin.after_disabled` 生命周期事件（non_blocking）。
3. 支持失败注入（返回 500/timeout）以验证重试队列。
4. 支持连续失败注入以验证熔断。
5. 支持验签（HMAC-SHA256）验证签名与防篡改。

## 3. Manifest 设计（示例，仅设计字段）

说明：以下 manifest 字段仅用于设计验证，不代表当前系统已支持外部 http_service 运行时或 webhooks 自动投递。

```json
{
  "code": "official_announcement",
  "name": "官方公告插件",
  "version": "0.1.0",
  "description": "用于验证 Webhook / HTTP 插件服务协议的官方示例插件（不执行第三方代码）。",
  "compatible_core_version": ">=1.7.0 <2.0.0",
  "runtime": {
    "type": "http_service",
    "service_url": "https://official-announcement.devhub.local",
    "health_check": "/health",
    "timeout_ms": 5000
  },
  "webhooks": [
    { "hook": "content.after_create", "url": "/hooks/content.after_create", "mode": "non_blocking", "retry": true, "timeout_ms": 5000 },
    { "hook": "plugin.after_enabled", "url": "/hooks/plugin.after_enabled", "mode": "non_blocking", "retry": true, "timeout_ms": 5000 },
    { "hook": "plugin.after_disabled", "url": "/hooks/plugin.after_disabled", "mode": "non_blocking", "retry": true, "timeout_ms": 5000 }
  ],
  "api_scopes": ["config.read", "config.write", "audit.write"],
  "auth": { "type": "hmac", "secret_ref": "official-announcement-secret" },
  "config_schema": {
    "type": "object",
    "properties": {
      "enabled": { "type": "boolean", "default": false },
      "message": { "type": "string", "default": "" },
      "display_mode": { "type": "string", "enum": ["banner", "home_section"], "default": "banner" }
    },
    "required": ["enabled"],
    "additionalProperties": false
  }
}
```

挂载点（设计）：

- `frontend.home.section`
- `frontend.header.nav`（可选）
- `admin.sidebar.menu`（配置入口）
- `admin.plugin.detail.tab`（投递/健康状态概览）

## 4. 外部插件服务（官方模拟服务/测试桩）设计

服务端点（设计）：

- `GET /health`：返回服务健康与版本。
- `GET /manifest`：返回摘要（用于对账）。
- `POST /hooks/{hook_name}`：接收事件投递，返回 2xx/4xx/5xx/超时以注入测试场景。

失败注入（设计）：

- `?fail=500`：返回 500。
- `?fail=timeout`：延迟超过 timeout。
- `?fail=429`：返回 429 + Retry-After。

验签（设计）：

- 验证 `X-DevHub-Signature-Alg=HMAC-SHA256`。
- 验证签名串：`timestamp.method.path.body_sha256`。
- 验证 timestamp 在 5 分钟窗口内。
- 验证 `idempotency_key` 去重。

v1.7.6 补充（实现现状 + 接收端验签示例）：

DevHub v1.7.6 已在发送端落地 HMAC-SHA256 签名，并在 delivery 记录中写入：

- `signature_alg`：`HMAC-SHA256`
- `secret_ref`：例如 `whsec_xxx`
- `body_sha256`：请求 body 的 sha256 hex
- `signature_status`：`signed|secret_missing|secret_disabled|secret_revoked|secret_expired|sign_failed`

接收端验签示例（伪代码，Go 风格，仅用于官方测试桩/示例服务，不代表 DevHub 会执行第三方代码）：

```go
func verify(req *http.Request, body []byte, secrets map[string]string) error {
  ts := req.Header.Get("X-DevHub-Timestamp")
  sig := req.Header.Get("X-DevHub-Signature") // v1=<hex>
  alg := req.Header.Get("X-DevHub-Signature-Alg")
  bodySHA := req.Header.Get("X-DevHub-Body-SHA256")
  secretRef := req.Header.Get("X-DevHub-Secret-Ref")
  if alg != "HMAC-SHA256" { return errors.New("alg unsupported") }
  if !within5Min(ts) { return errors.New("timestamp expired") }
  if sha256Hex(body) != bodySHA { return errors.New("body hash mismatch") }

  secret := secrets[secretRef] // secret_ref -> plaintext secret
  if secret == "" { return errors.New("secret missing") }

  signing := ts + "." + strings.ToUpper(req.Method) + "." + req.URL.Path + "." + bodySHA
  expected := "v1=" + hmacSHA256Hex([]byte(secret), signing)
  if !constantTimeEqual(expected, sig) { return errors.New("signature mismatch") }
  return nil
}
```

Secret 轮换窗口说明（建议）：

- DevHub 发送端只使用 `active` secret 签名；
- 插件接收端在 grace period 内可同时接受 `active` 和 `previous` 的 secret_ref；
- grace period 后 `previous` 进入 `expired`，接收端应停止接受旧 secret_ref。

## 5. 端到端验证目标（按实现阶段）

### 阶段 1：non_blocking delivery 最小闭环

验证点：

1. `content.after_create` 触发 event + delivery 记录。
2. 插件服务返回 500 不阻断内容创建主流程。
3. disabled/soft_uninstalled 插件不投递。
4. 审计写入 delivery created/failed。

### 阶段 2：重试队列

验证点：

1. 5xx/timeout 进入 retry_scheduled。
2. attempt 递增且 event_id 不变。
3. 超过 max_attempts 进入 retry_exhausted 并写审计。

### 阶段 3：熔断

验证点：

1. 连续失败达到阈值进入 circuit_open。
2. circuit_open 后暂停投递并标记 skipped/circuit_open。
3. 管理员可手动恢复。
4. 手动恢复后可再次触发投递成功（circuit closed → 正常投递）。
5. 429 场景：插件服务返回 `Retry-After` 后，delivery 进入 `retry_scheduled`，并在到期后可被 `retry-due` 扫描重试。

### 阶段 4：签名与鉴权

验证点：

1. DevHub 生成签名 headers。
2. body 被篡改签名不一致。
3. timestamp 过期被拒绝（401）。

### 阶段 5：后台治理入口

验证点：

1. 可查看 event 列表与 delivery 列表。
2. 可查看失败原因、attempt、next_retry_at。
3. 可手动重试 non_blocking delivery。
4. 可查看与恢复熔断。

## 7. v1.7.5 关联说明（实现现状）

说明：v1.7.5 已实现 delivery 重试与熔断治理的核心链路（非阻塞），用于支撑本示例插件后续端到端验证：

- 5xx/timeout/429（含 `Retry-After`）触发 `retry_scheduled` 并写入 `next_retry_at`。
- `retry-due` 扫描式重试（DB-as-queue）不阻断主流程。
- 连续失败达到阈值触发 `open` 熔断；到达 `next_probe_at` 允许 `half_open` 探测；管理员可手动 close/open。

本示例插件仍不执行第三方不可信代码；官方示例仅用于协议与治理验证。

## 6. 备选示例插件：友情链接

友情链接插件（`official_links`）更适合验证“配置 + 前台 slot 挂载”，对 Hook 依赖较弱。建议作为第二个示例插件，等 Webhook non_blocking 闭环稳定后再补。
