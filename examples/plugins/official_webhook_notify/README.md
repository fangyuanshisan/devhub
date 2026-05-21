# Official Webhook Notify

这是 DevHub 官方 `external_service` Webhook 示例插件。它用于验证声明型插件包的 upload -> promote -> install 链路，以及 external_service non-blocking Webhook 投递、后台配置、健康检查和失败重试。

## 当前能力

- 插件包只包含 `manifest.json`、`README.md`、`config.example.json`、`checksums.json` 和 `migrations/`。
- manifest 声明 `AfterCreateContent` Hook，`service_type=external_service`，`mode=non_blocking`。
- DevHub 在内容创建后按声明异步 `POST {endpoint_url}/hooks/content.after_create`。
- DevHub 记录 `hook_executions(service_type=external_service)`，可在后台详情抽屉或 Webhook 治理中排障。
- 失败类执行记录可以由管理员手动重试。

## 安全边界

- DevHub 不执行插件包代码。
- 插件服务由外部自行部署和运维。
- DevHub 只负责按 manifest 声明异步投递 HTTP 请求。
- 本包不包含运行时代码、package scripts、远程 iframe URL、真实 secret、用户数据或外部 SQL。
- `migrations/` 是唯一迁移入口；本示例没有真实迁移，因此只保留说明文件。
- external_service token、Webhook Secret、Callback Token 和 Authorization Header 不会在列表、执行记录、日志或审计中明文展示。

## 本地验收

1. 打包本目录为 zip，或在后台本地插件包 dry-run 中输入 `examples/plugins/official_webhook_notify`。
2. 走插件包治理：upload -> precheck -> promote -> install dry-run -> install。
3. 启动 mock receiver：

```bash
go run ./cmd/webhook-mock-receiver
```

4. 在后台插件详情的“运行记录”里配置 external_service：
   - `endpoint_url`: `http://127.0.0.1:19090`
   - `health_check_path`: `/health`
   - `timeout_ms`: `3000`
   - `failure_policy`: `warn`
   - `auth_type`: `none`
5. 点击“执行健康检查”，状态应为正常。
6. 创建任意内容触发 `AfterCreateContent`，mock receiver 应收到 `/hooks/content.after_create` 请求。
7. 停掉 mock receiver 或让它返回 500，确认失败记录进入 `retry_scheduled` / `retry_exhausted`，再在后台点击“重试”。

## 请求形状

DevHub 投递的 payload 使用 `schema_version=1`，包含 `plugin_code`、`hook_name`、`event_id`、`execution_id`、`resource_type`、`resource_id`、`community_id`、`actor`、`occurred_at` 和最小事件数据。接收端必须自行做好幂等处理。
