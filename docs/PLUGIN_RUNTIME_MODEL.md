# DevHub 插件运行模型（设计）

[返回文档入口](README.md)

更新时间：2026-05-17

本文件承接 `docs/PLUGIN_ARCHITECTURE.md` 中的运行模型总述，集中描述第三方插件如何运行、如何挂载前端、如何通过 HTTP 服务接入后端、如何通过受控 API 与 Core 交互，以及哪些能力当前仅为设计、哪些仍待实现。

本文件只做设计，不代表已实现第三方插件运行、HTTP 服务协议、iframe 沙箱或受控 API token 系统。

## 1. 运行模型分层

### 1.1 Core 内置插件

- 随 DevHub 主仓库发布，运行在 DevHub 进程内。
- 直接使用 Core 内部接口与 HookBus。
- 受 Core 权限、审计、配置、生命周期与 SEO 边界约束。
- 适合作为官方基础能力（如 QA、Docs、Wiki、Projects、Jobs、AI Works）。

### 1.2 外部 HTTP 服务插件

- 插件后端作为独立 HTTP 服务运行。
- DevHub Core 不加载第三方代码，不加载 Go plugin。
- DevHub 通过 Webhook / Action / Callback 协议与插件交互。
- 插件服务只能调用授权的 Core API scopes。
- 插件服务故障不能拖垮 Core 主流程。
- 这是中期最推荐的第三方插件运行方式。

### 1.3 前端 iframe / sandbox 插件

- 插件前端页面通过 iframe 或 sandbox 容器挂载。
- DevHub 不直接执行不可信 JS。
- 插件与 DevHub 通过 postMessage 或受控 SDK 通信。
- 插件只能访问授权上下文，不能绕过权限访问后台 API。
- 适合后台配置页、数据面板、低代码页面和第三方工具页面。

## 2. 前端挂载模型

插件可扩展的位置（设计字段）：

- `admin.sidebar.menu`
- `admin.plugin.detail.tab`
- `admin.dashboard.card`
- `frontend.header.nav`
- `frontend.home.section`
- `frontend.topic.sidebar`
- `frontend.topic.after_content`
- `frontend.user.menu`
- `moderator.sidebar.menu`

规则：

1. 挂载点必须在 manifest 中显式声明。
2. 挂载点必须经过权限校验和插件状态校验。
3. 挂载点必须受 community plugin 状态控制。
4. 直接注入 JS 不是默认方向。
5. 当前仅为设计字段，不代表 Core 已实现对应 slot。

## 3. 后端运行模型

第三方插件后端建议采用 HTTP 插件服务协议：

- DevHub Core 通过 HTTP 调用外部插件服务。
- 插件服务通过受控 API 回调 Core。
- HookBus 投递支持 blocking / non_blocking 策略；第三方插件运行模型中，默认优先 non_blocking。
- 插件服务必须支持健康检查、签名鉴权、幂等和超时控制。

建议接口（设计，不代表已实现）：

- `GET /health`
- `GET /manifest`
- `POST /hooks/:hookName`
- `POST /actions/:actionName`
- `POST /config/validate`
- `POST /install/precheck`
- `POST /enable/precheck`

## 4. 受控 Core API 模型

插件调用 Core API 必须满足：

1. 不能直接访问数据库。
2. 不能绕过 Core 权限与审计。
3. 必须携带插件身份：`plugin_code`、`install_id`、`publisher_id`、`scope`、`actor_type`、`actor_id`、`community_id`。
4. API token 必须可吊销且具备作用域。
5. 插件 token 不等于管理员 token，不应继承用户全部权限。

建议 scope（设计）：

- `content.read`
- `content.write`
- `comment.read`
- `comment.write`
- `tag.read`
- `notification.send`
- `search.index`
- `seo.extend`
- `config.read`
- `config.write`
- `audit.write`

## 5. Manifest 运行模型字段

`manifest` 中与运行模型相关的设计字段：

```json
{
  "runtime": {
    "type": "internal|http_service|iframe",
    "entry": "string",
    "health_check": "/health",
    "timeout_ms": 3000,
    "permissions": [],
    "scopes": []
  },
  "frontend": {
    "mounts": [
      {
        "slot": "admin.sidebar.menu",
        "title": "Demo",
        "path": "/admin-next/plugins/demo",
        "mode": "iframe",
        "url": "/plugins/demo/admin"
      }
    ]
  },
  "backend": {
    "service_url": "https://plugin-service.example.com",
    "auth": "signed_request",
    "hooks": []
  },
  "api_scopes": [
    "content.read",
    "content.write"
  ]
}
```

说明：

- 这些字段当前只是设计字段，未落地为真实运行时。
- `runtime.type=http_service` 是第三方插件的推荐运行方式之一。
- `backend.service_url` 后续必须经过可信发布者、SSRF、签名、鉴权和重试边界校验。
- `frontend.mounts` 只表示挂载声明，不等于 Core 已实现该 slot。

## 6. 已完成能力与未完成能力

### 已完成（当前仓库已存在）

- Core + 插件服务底座目标统一。
- 插件包治理主链路。
- 远程插件包下载 / 预检 / 兼容性 / 安装事务 / 启用前检查 / 启用 / 软卸载 / 升级。
- Ed25519 签名验签与可信发布者管理。
- 内置插件治理。
- HookBus 基础能力。
- 远程索引只读镜像与版本仓库。

### 设计中 / 未完成

- 第三方插件前端挂载。
- iframe / sandbox。
- HTTP 插件服务协议真实实现。
- 插件 SDK。
- 插件受控 API token。
- 插件运行时资源隔离。
- 插件市场。
- 远程自动更新。
- 动态代码加载。

## 7. 路线图引用

本文件与以下设计文档配套：

- `docs/PLUGIN_ARCHITECTURE.md`
- `docs/PLUGIN_SYSTEM_ROADMAP.md`
- `docs/PLUGIN_WEBHOOK_PROTOCOL.md`
- `docs/PLUGIN_WEBHOOK_IMPLEMENTATION_PLAN.md`
