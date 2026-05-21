# DevHub 插件前端挂载模型（设计）

[返回文档入口](README.md)

更新时间：2026-05-18

本文档用于定义 **DevHub 插件前端挂载模型** 与 **iframe / sandbox 容器隔离策略**，并给出 `postMessage` 通信协议（设计口径）。

重要说明：

- 本文为 **v1.8.0 文档设计**，不代表当前代码已实现“第三方插件前端挂载”。  
- DevHub 当前仍坚持运行时安全边界：**不执行第三方不可信代码**、不做远程动态加载、不做 Go plugin 动态加载。  
- 插件前端挂载的主要方向仍是 **iframe + sandbox + postMessage**（设计口径），但当前实现只开放官方 allowlist 挂载，不开放任意远程 iframe URL、不开放任意远程 JS / CSS 入口、不开放第三方前端运行时；插件页面不能绕过 Core 的权限、审计与插件生命周期状态。
- v1.8.3-S22 进一步把 manifest / 预检 / 运行时收口到官方 allowlist：只接受官方挂载点和官方组件 key，预检 / install dry-run 会阻断未知挂载点、未知组件 key、`iframe_url`、`script_url`、`remote_entry`、`external_js`、`inline_html`、`remote_component`、`eval` 和不支持的 `render_mode`。`official_announcement` 的配置与预览入口仍在详情内展示，但不代表第三方前端挂载已开放。该调整只改变后台信息展示和安全收口，不改变 Host + iframe + postMessage 协议。
- 运行时只返回已安装、已启用 / running、未归档、未软卸载，且当前子站已启用的插件挂载；未知历史组件会被跳过并返回 warning，不会白屏。传给前端组件的 props 会过滤 secret / token / authorization / password / credential 类字段，官方 helper 仍只创建内置同源 iframe，不读取插件声明的远程 iframe 或脚本入口。

## 1. 目标与非目标

### 1.1 目标

1. 定义插件前端可扩展的挂载点（slot）集合，并明确每个 slot 的上下文与权限边界。
2. 定义 iframe 容器隔离策略（sandbox、跨域策略、能力最小化）。
3. 定义插件 iframe 与 DevHub Host 的 `postMessage` 协议（握手、上下文、请求/响应、错误码）。
4. 明确插件挂载受以下状态共同约束：
   - 插件全局 enabled / disabled / soft_uninstalled
   - 子站（community）插件 enabled / disabled
   - 当前用户权限（admin / moderator / user）与 scope
5. 以 `official_announcement` 作为第一个端到端验证插件，给出前后台挂载方案（仍不执行第三方代码）。

### 1.2 非目标（本轮不做）

- 不实现 blocking Hook。
- 不做“第三方 JS 直接注入 Core 页面 DOM”的扩展模式。
- 不做动态加载远程前端资产（remote JS/CSS 注入）。
- 不把插件 iframe 直接接入 Core 的 user/admin token。
- 不设计插件市场、远程自动更新、第三方代码执行。

## 2. 挂载点（slots）设计

挂载点命名统一使用：`{area}.{page}.{position}`。

约定字段（设计）：

- `slot`：挂载点名称（固定枚举）。
- `title`：展示标题（菜单/Tab/卡片标题）。
- `path`：DevHub 内部路由路径（用户点击后进入的“宿主页面”，不是 iframe 的真实 URL）。
- `iframe.src`：iframe 的入口 URL（建议为独立插件域名/路径）。
- `mode`：`iframe`（当前只定义该模式）。

### 2.1 管理端 slots

1. `admin.sidebar.menu`
   - 目的：让插件在后台左侧菜单出现入口（进入宿主页面，再加载 iframe）。
   - 权限：必须通过后台权限校验（admin 或被允许的角色）。
   - 备注：菜单只负责“页面入口”，不承担状态筛选（状态应在页内 Tab / query）。

2. `admin.plugin.detail.tab`
   - 目的：插件详情页内扩展 Tab（例如“公告配置”“Webhook 状态”）。
   - 权限：同插件详情页（admin）。
   - 备注：适合做插件配置页、插件治理页（只读/读写由权限与 scope 决定）。

3. `admin.dashboard.card`
   - 目的：后台仪表盘卡片（例如“插件健康摘要”）。
   - 权限：同 dashboard（admin）。
   - 备注：默认只读；尽量避免在 card 中放高危操作。

4. `moderator.sidebar.menu`
   - 目的：版主工作台侧边栏入口（如果项目存在 `moderator` 工作台）。
   - 权限：必须同时满足：
     - 当前用户是该 community 的 moderator
     - 插件在该 community 范围内 enabled
   - 备注：版主不是全局 admin，必须做 community scope 校验。

### 2.2 前台 slots

1. `frontend.header.nav`
   - 目的：前台顶部导航扩展（如公告入口、外链）。
   - 权限：默认对所有访问者可见；若插件声明需要登录，则只对已登录用户可见。

2. `frontend.home.section`
   - 目的：首页区块扩展（如公告 banner、推荐模块）。
   - 权限：默认对所有访问者可见；需要按 community/站点配置决定显示与否。

3. `frontend.topic.sidebar`
   - 目的：内容详情页侧边栏扩展。
   - 约束：不得破坏 `/topics/:id` 的 SEO 动态 HTML 兜底；插件扩展应是“附加区块”，可以延迟加载。

4. `frontend.topic.after_content`
   - 目的：内容详情正文后扩展区块。
   - 约束：同上，不得影响 SEO 主内容输出与可访问性兜底。

5. `frontend.user.menu`
   - 目的：用户中心菜单扩展。
   - 权限：仅对已登录用户可见；仍需插件 enabled 与 community enabled。

## 3. iframe / sandbox 隔离策略（设计）

### 3.1 总体原则

1. 插件页面默认在 **iframe** 中运行。
2. iframe 必须启用 `sandbox`，默认不允许插件直接拿到与 DevHub 同等的浏览器能力。
3. 插件页面默认与 DevHub **不同源**（推荐），避免同源脚本访问带来的边界模糊。
4. 插件页面不能读取 DevHub 的 `Authorization token`、cookie 或任何敏感配置；只能通过 Host 的受控接口能力获取必要信息。

### 3.2 sandbox 属性建议（基线）

默认 sandbox（建议基线，按场景增量开启）：

- `allow-scripts`：允许插件页面执行脚本（否则无法进行任何 UI 逻辑）。
- `allow-forms`：允许提交表单（用于“配置页”场景；若仅展示可禁用）。

默认不启用（除非明确需要且有安全评估）：

- `allow-popups` / `allow-popups-to-escape-sandbox`
- `allow-top-navigation` / `allow-top-navigation-by-user-activation`
- `allow-downloads`
- `allow-modals`
- `allow-pointer-lock`

关于 `allow-same-origin`：

- 若插件 iframe 与宿主页面 **同源**，不建议同时开启 `allow-scripts` + `allow-same-origin`；这会显著削弱 sandbox 的隔离价值。
- 推荐插件 iframe 使用 **独立插件域名**（与宿主不同源），并在确需持久化存储时才评估是否允许 `allow-same-origin`（此时仍不等于与宿主同源）。

### 3.3 iframe src 策略（推荐）

推荐把插件前端入口放在独立域名（示例）：

- DevHub Host：`https://devhub.example.com`
- 插件前端：`https://plugin-ui.devhub.example.com/{plugin_code}/...`

这样可以：

1. 清晰区分 cookie / storage 边界。
2. 避免插件 JS 与 DevHub 页面处于同源环境。
3. 让 `postMessage` 成为唯一受控通道。

## 4. postMessage 通信协议（设计）

本节定义插件 iframe 与 DevHub Host 的通信方式。协议目标是：**插件需要能力必须显式申请，Host 决策并执行**。

### 4.1 安全校验要求

Host 必须校验：

1. `origin`：消息来源必须在允许列表（按 plugin_code/环境配置）。
2. `plugin_code`：消息声明的 plugin_code 必须与当前挂载的 plugin_code 一致。
3. `mount_slot`：请求必须带 slot，并与当前 iframe 的 slot 一致。
4. `request_id`：每个请求必须唯一；Host 侧需要限流与去重（防刷/重放）。

插件侧必须校验：

1. `event.origin` 必须等于 DevHub Host origin。
2. 只接受 `target` 为 `devhub.plugin_host` 的消息。

### 4.2 消息结构（统一 envelope）

```json
{
  "v": 1,
  "target": "devhub.plugin_host | devhub.plugin_iframe",
  "type": "handshake.init | handshake.ack | context.get | context.result | core.request | core.result | ui.resize | error",
  "plugin_code": "official_announcement",
  "mount_slot": "frontend.home.section",
  "request_id": "req_01H...",
  "payload": {}
}
```

### 4.3 握手与上下文

1) 插件 → Host：`handshake.init`

payload（示例）：

```json
{
  "plugin_ui_version": "0.1.0",
  "capabilities": ["ui.resize", "core.request"],
  "expected_host_origin": "https://devhub.example.com"
}
```

2) Host → 插件：`handshake.ack`

Host 必须在 ack 中返回 “受控上下文”，但不得包含敏感 token：

```json
{
  "host_origin": "https://devhub.example.com",
  "server_time": "2026-05-18T00:00:00Z",
  "plugin_state": {
    "global_status": "enabled",
    "community_status": "enabled",
    "soft_uninstalled": false
  },
  "actor": {
    "type": "user|admin_user|moderator|guest",
    "id": 123
  },
  "community": {
    "id": 1,
    "slug": "php"
  },
  "permissions": {
    "can_view": true,
    "can_configure": false
  },
  "allowed_core_requests": [
    { "name": "config.get", "scopes": ["config.read"] },
    { "name": "audit.write", "scopes": ["audit.write"] }
  ]
}
```

说明：

- `allowed_core_requests` 是 Host 侧的 allowlist，最终仍需服务端校验，前端不可作为安全边界。
- 插件如果需要执行写操作（例如保存配置），必须经过 Host 的权限判断，并且最终由服务端做鉴权与审计。

### 4.4 受控 Core 请求（core.request）

插件 iframe 不直接调用 DevHub 任何带身份的 API（不持有 token），而是通过 `core.request` 请求 Host 代为执行。

请求示例：

```json
{
  "name": "config.get",
  "params": { "community_id": 1 }
}
```

Host 执行规则（设计）：

1. 先校验：插件状态（global/community/soft_uninstalled）与用户权限。
2. 再校验：该 slot 是否允许该请求类型（例如前台 slot 不允许写配置）。
3. 再校验：请求参数是否在允许范围内（community_id 必须与上下文一致）。
4. Host 使用自身已存在的受控后端通道执行：
   - **读取配置**：复用 “受控 API” 的能力边界（实现上可选择复用 plugin-callback 或新增 plugin-frontend 命名空间；v1.8.0 只定义设计，不声明已实现）。
   - **写审计**：复用 audit 写入约束（action 必须以 `plugin_code.` 前缀开头）。
5. Host 返回 `core.result`，并在服务端写入审计/记录。

响应示例：

```json
{
  "ok": true,
  "result": { "effective_config": { "enabled": true, "message": "Hi" } }
}
```

错误响应示例：

```json
{
  "ok": false,
  "error_code": "SCOPE_DENIED|PLUGIN_DISABLED|COMMUNITY_SCOPE_DENIED|PERMISSION_DENIED",
  "message": "reason"
}
```

### 4.5 UI resize（可选）

插件 iframe 可以发 `ui.resize`，Host 根据 slot 决定是否允许自适应高度：

- `admin.dashboard.card`：允许自适应高度（有上限）
- `frontend.home.section`：允许自适应高度（有上限）
- `admin.plugin.detail.tab`：允许

## 5. 插件状态与权限控制（必须是服务端最终裁决）

### 5.1 插件生命周期状态 gates

挂载入口必须同时满足：

1. 插件全局状态 `enabled`。
2. 插件在当前 community 范围内 `enabled`（若 slot 属于 community 场景）。
3. 插件不处于 `soft_uninstalled` 状态。

不满足时：

- 前台 slot：不渲染插件入口（或渲染一个“该功能已停用”的占位，但不加载 iframe）。
- 后台 slot：入口可保留，但进入后显示“插件已禁用/已归档”状态页（不加载 iframe），便于治理与恢复。

### 5.2 用户权限 gates

1. 管理端挂载点（admin.*）必须通过后台权限校验；权限不足用户不能进入插件配置页。
2. 版主挂载点（moderator.*）必须校验 moderator 的 community scope。
3. 前台挂载点（frontend.*）默认只读，写操作必须进一步校验（例如“点击统计”通过审计写入，但不得提升用户权限）。

## 6. 官方公告插件（official_announcement）前端挂载验证方案（设计）

目标：让 `official_announcement` 成为第一个“前后台均可验证”的官方插件，验证前端挂载模型但仍不执行第三方不可信代码。

### 6.1 后台：插件详情页 Tab（配置页）

slot：`admin.plugin.detail.tab`

- Tab 名称：`公告配置`
- iframe 页面：展示当前公告配置（enabled/message/display_mode）
- 读取配置：通过受控 `core.request(name=config.get)` 获取 `effective_config`
- 保存配置：**后续阶段**再设计（需要更严格权限与写入通道；v1.8.0 不实现）

权限要求：

- 仅 admin 可访问（或项目现有权限体系允许的角色）

### 6.2 前台：首页区块（展示页）

slot：`frontend.home.section`

- iframe 页面：只读展示公告
- 读取配置：`core.request(config.get)` 获取 `effective_config`
- 点击/曝光记录：`core.request(audit.write)` 写入 `official_announcement.banner_viewed` / `...clicked`（action 必须以 `official_announcement.` 前缀）

状态要求：

- 插件 disabled → 前台不展示
- community 未启用插件 → 前台不展示

## 7. 已完成 vs 设计中

已完成（截至 v1.7.9）：

- Webhook non_blocking 链路治理（events/deliveries/retry/circuit/signature/secrets）。
- 插件服务回调 Core API 的最小通道（callback token + scopes + callback requests）。

设计中 / 未完成（本文覆盖的范围）：

- 插件前端 slots 的真实渲染与宿主页面实现。
- iframe 容器与统一 postMessage Host（消息通道、allowlist、权限联动）。
- 插件前端对“配置写入”的受控通道（例如 `config.write`）。
- 前端挂载的可观测性与性能预算（LCP/CLS、错误上报等）。
