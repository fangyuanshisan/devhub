# 官方示例插件：公告插件（端到端验证方案）

[返回文档入口](../README.md)

更新时间：2026-05-19

本文定义官方示例插件“公告插件（official_announcement）”的端到端验证方案，并记录 v1.8.1 已落地的“官方插件前端挂载最小闭环”。官方公告插件用于验证 DevHub 插件系统能力，不用于执行第三方不可信代码。

重要边界：

- 官方示例插件用于“协议验证与端到端流程验证”，不用于执行不可信第三方代码。
- 不做 Go 动态加载，不运行外部 package scripts，不加载 `.so/.dll/.node/wasm`。
- 外部插件服务以“官方提供的模拟服务/测试桩”形态存在，用于验签、回包、失败注入与熔断验证。

## 1. 插件定位

插件编码：`official_announcement`

目标：通过公告类业务（简单、可见、可配置）验证：

- 菜单/页面入口（v1.8.1 已落地最小 Host + iframe 挂载）。
- 插件配置（声明 `config_schema` 与默认配置）。
- Hook non_blocking 投递（content.after_create 等）。
- delivery 记录、重试队列、熔断、审计、后台治理入口。

v1.8.1 已实现的前端挂载闭环：

- `official_announcement` 作为内置官方插件加入 `internal/plugins`。
- iframe 页面为仓库内置页面：`GET /plugins/official-announcement/iframe`。
- Host API 为浏览器安全 API：
  - `GET /api/v1/plugins/official-announcement/context`
  - `POST /api/v1/plugins/official-announcement/audit-events`
- 前台首页挂载：满足配置 `enabled=true` 且 `message` 非空时展示公告（Host + iframe + postMessage）。
- 后台插件详情页挂载：仅当插件为 `official_announcement` 时显示“公告预览”Tab（Host + iframe）。

重要边界（v1.8.1）：

- 不允许远程 iframe URL；iframe 页面必须来自 DevHub 仓库内置路由。
- Host API 不返回 callback token / webhook secret，不向浏览器暴露任何后端凭证。
- 仍不执行第三方不可信代码，不做 JS 注入，不做远程动态加载。

v1.8.1-S2 补充（子站页 `/c/:slug` 挂载验收口径）：

- 子站页挂载通过 `GET /api/v1/plugins/official-announcement/context` 携带 `community_slug`，后端强校验子站插件启用状态。
- Host 写审计 `POST /api/v1/plugins/official-announcement/audit-events` 支持携带 `community_slug`，子站插件 disabled 时拒绝写入，避免绕过 gating。

v1.8.2 补充（通用容器/Helper 抽取）：

- 前台首页、子站页 `/c/:slug`、后台插件详情页不再各自复制一套挂载脚本，而是统一复用后端内置 helper：
  - `GET /plugins/assets/devhub-plugin-mount-host.js`
- helper 仅对官方内置插件 allowlist 生效（第一阶段仅 `official_announcement`），不支持任意远程 iframe URL；iframe 仍固定为内置路由：`GET /plugins/official-announcement/iframe`。

v1.8.3 补充（后台详情稳定性与中文体验）：

- 插件详情抽屉中 `official_announcement` 的概览、配置、前端挂载和公告预览说明已更明确：
  - 官方内置插件，用于验证前端挂载模型。
  - 不执行第三方代码。
  - iframe 只允许内置页面，不允许远程 iframe URL。
  - 配置入口用于公告开关、公告内容、链接文字、链接地址和是否允许关闭。
  - 前端挂载入口展示 iframe 路由、sandbox 策略、postMessage 状态。
  - 预览入口区分前台预览 / 后台预览语义，不暴露 callback token / webhook secret。
  - 安全提示明确“远程 iframe URL：否”“第三方代码执行：否”。
- 本轮只调整后台页面结构、稳定性保护和中文说明，不改变 Host + iframe + postMessage 协议。
- v1.8.3-S1 补充：插件列表中 `official_announcement` 直接显示“官方公告插件 / 官方内置”标识，能力摘要突出“前端挂载 / 配置 / iframe 预览”，便于后台管理员从列表快速定位并进入详情；不改变官方公告插件的 Host API、iframe 路由、postMessage 协议或安全边界。
- v1.8.3-S5 补充：插件详情抽屉将 `official_announcement` 的“公告配置 / 前端挂载 / 公告预览”作为高频入口展示；配置 Tab 直接显示公告开关、公告内容、链接文字、链接地址、是否允许关闭等摘要，前端挂载 Tab 说明首页 / 子站页 / 后台预览、iframe 路由、sandbox 与 postMessage 状态，公告预览继续复用官方内置 Host。原始配置和调试 JSON 进入“技术详情”并脱敏；不改变 Host API、iframe 路由、postMessage 协议或安全边界。

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

v1.8.0 补充（前端挂载模型设计）：

- 官方公告插件可作为第一个“前后台挂载”验证插件（仍不执行第三方不可信代码）：
  - 后台：插件详情页 `admin.plugin.detail.tab` 增加“公告配置”Tab（iframe 展示配置页，读取配置使用受控通道）
  - 前台：首页 `frontend.home.section` 挂载公告展示区（iframe 只读展示）
- iframe/sandbox 与 postMessage 通信模型（设计口径）见：`docs/PLUGIN_FRONTEND_MOUNT_MODEL.md`。

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

v1.7.7 补充（插件服务回调 Core API：Callback Token + Scopes）：

本仓库在 v1.7.7 增加了“插件服务回调 Core API”的最小受控通道：

- Callback Token：外部插件服务使用 `Authorization: Bearer cbsk_...` 调用 `/api/v1/plugin-callback/*`。
- Webhook Secret：仅用于 DevHub -> 插件服务 Webhook 签名，两者不是同一能力。

最小 scopes（当前实现）：

- `config.read`：读取本插件在指定 `community_id` 下的 effective config（已按 `config_schema` 脱敏）。
- `audit.write`：写入以 `plugin_code.` 前缀开头的插件审计 action（禁止伪造 Core/admin 审计）。

Callback Token 使用示例（curl）：

- 读取配置（需要 scope：`config.read`，且 `community_id` 必须在 token 的 community_scope 内）：

```bash
curl -sS -H "Authorization: Bearer cbsk_xxx" \
  -H "X-DevHub-Plugin-Code: official_announcement" \
  -H "X-DevHub-Token-Ref: cbtk_xxx" \
  -H "X-DevHub-Request-ID: cbreq_demo_1" \
  "http://localhost:8090/api/v1/plugin-callback/config?community_id=1"
```

- 写入插件审计（需要 scope：`audit.write`，action 必须以 `official_announcement.` 前缀开头）：

```bash
curl -sS -X POST -H "Authorization: Bearer cbsk_xxx" \
  -H "X-DevHub-Plugin-Code: official_announcement" \
  -H "X-DevHub-Token-Ref: cbtk_xxx" \
  -H "X-DevHub-Request-ID: cbreq_demo_2" \
  -H "Content-Type: application/json" \
  -d '{"action":"official_announcement.received_event","resource_type":"webhook_event","resource_id":"evt_xxx","metadata":{"note":"demo"}}' \
  "http://localhost:8090/api/v1/plugin-callback/audit-events"
```

验证点（建议补测）：

1. Token 缺失 / 无效 / disabled / revoked / expired → 401。
2. scope 不足 → 403。
3. community_scope 不匹配 → 403。
4. 插件 global disabled / soft_uninstalled → 403（callback 通道禁止绕过插件状态）。
5. callback request 记录可在后台 `Webhook 治理` 页的 `Callback Requests` 中查看（不保存 token 明文）。

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
