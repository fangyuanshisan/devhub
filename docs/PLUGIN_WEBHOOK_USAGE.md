# DevHub Webhook 插件使用方法

[返回文档入口](README.md)

更新时间：2026-05-20

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

## 2. 配置 external_service

当前后台详情页可以查看外部服务状态并执行健康检查；真正的配置保存入口以 Admin API 为主。

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
- `token` 只在写入时提交，不要期望列表里能看到明文。

## 3. 先跑健康检查

在插件详情的运行区域，`external_service` 模块里可以直接点“健康检查”。

健康检查会：

- 对 `{endpoint_url}{health_check_path}` 发 `GET`。
- 记录到 `hook_executions(service_type=external_service)`。
- 2xx 视为 healthy。
- 3xx 视为 warning。
- 4xx / 5xx / 超时 / 连接错误按失败策略处理。

如果插件当前是 `disabled`、`archived` 或 `soft_uninstalled`，不会真的打外部 endpoint，只会留下 skipped 记录。

## 4. 看投递和异常

进入后台 `Webhook 治理`：

- `投递记录` 看每次 Hook 是否成功。
- `异常处理` 看重试队列、熔断状态和待处理异常。
- `事件` 看事件链路。
- `Webhook 密钥` 看签名凭据的治理。
- `回调 Token` 看插件服务回调 Core 的凭据治理。
- `回调请求` 看插件服务实际回调记录。

常见判断：

- `retry_scheduled` / `retry_exhausted`：多半是超时、5xx 或 429。
- `skipped`：通常是插件状态、子站状态、endpoint 缺失或 token 缺失。
- `circuit_open`：先修外部服务，再去处理熔断。

## 5. 配 Webhook Secret

如果你的插件服务要接收 DevHub 发出的 Webhook，就在 `Webhook 治理 -> Webhook 密钥` 管理密钥。

原则：

- 密钥明文只在创建或轮换时展示一次。
- 列表、详情、审计和执行记录都不回显明文。
- 轮换后要让接收端接受新的 `secret_ref`，再逐步结束旧密钥窗口。

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

1. 在 manifest 里声明 `external_service` Hook。
2. 保存插件后，在后台配置 `endpoint_url` 和 `auth_type`。
3. 点一次健康检查，确认状态是 healthy。
4. 触发一次业务事件，例如创建内容。
5. 去 `Webhook 治理 -> 投递记录` 看记录。
6. 如果插件服务还要读配置或写审计，再创建 Callback Token。

## 8. 排障顺序

先看三件事：

1. 插件状态是不是 `enabled`。
2. `external_service` 配置是不是已经存在。
3. `Webhook 治理` 里有没有投递记录或失败原因。

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

