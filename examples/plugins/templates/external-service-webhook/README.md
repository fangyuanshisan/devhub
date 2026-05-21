# External Service Webhook Plugin Template

这是 DevHub 官方 `external_service` Webhook 插件模板，基于 `official_webhook_notify` 示例整理。它展示如何声明 external_service 配置、non-blocking Hook、健康检查、失败策略和手动 retry 排障入口。

## 适用场景

- 内容发布通知。
- 外部审核系统。
- 消息推送。
- CRM / 工单系统集成。
- 低耦合异步扩展。

## DevHub 做什么

- 读取 manifest 中的 Hook 声明。
- 在业务事件发生后异步 `POST {endpoint_url}{hook.path}`。
- 记录 `hook_executions(service_type=external_service)`。
- 按配置执行 health check。
- 对失败类执行记录提供手动 retry。

## DevHub 不做什么

- 不运行插件包里的代码。
- 不部署外部 receiver 服务。
- 不开放 blocking Hook。
- 不开放远程 iframe。
- 不把 external_service 当成完整第三方运行模型。

## 外部服务要求

- 自行部署 receiver。
- 校验签名或 Bearer token。
- 使用 `execution_id`、`event_id` 或 `idempotency_key` 做幂等。
- 对健康检查路径返回 2xx。
- 不把 token / secret 写入 manifest、README 或 `config.example.json`。

## 后台配置

安装插件后，到插件详情或 Webhook 治理中配置：

- `enabled`
- `endpoint_url`
- `health_check_path`
- `timeout_ms`
- `failure_policy`
- `auth_type`
- `token`
- `warning_threshold`
- `error_threshold`

`token` 只允许写入，不回显明文；已有 token 在后台显示为“已配置密钥 / 可替换”。

## 验收流程

1. 复制模板并修改 `code`、`name`、Hook path 和权限码。
2. 实现并部署外部 receiver。
3. 打包 zip 后走 upload -> precheck -> promote -> install dry-run -> install。
4. 在后台配置 endpoint 和 token。
5. 执行 health check。
6. 触发内容创建事件。
7. 在 Webhook 治理查看 `hook_executions`。
8. 让 receiver 返回 500 或停掉 receiver，确认失败记录可手动 retry。

模板还包含 `checksums.json`，用于提供文件完整性摘要。

可配合 `cmd/webhook-mock-receiver` 做本地验收。
