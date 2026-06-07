# Receiver Example

Webhook receiver 由插件开发者自行部署，DevHub 不会运行插件包中的服务代码。

最小接口：

- `GET /health`：返回 2xx 表示健康。
- `POST /hooks/content.after_create`：接收内容创建后的事件。

接收方建议：

1. 校验 `Authorization: Bearer <token>` 或 DevHub Webhook 签名。
2. 使用 `X-DevHub-Execution-ID`、`X-DevHub-Event-ID`、`X-DevHub-Idempotency-Key` 做幂等。
3. 返回 2xx 表示已处理。
4. 对 429 / 5xx / timeout 做可重试设计。
5. 不记录 Authorization Header、Webhook Secret、Callback Token 或敏感 payload 明文。

本地可使用：

```bash
go run ./cmd/webhook-mock-receiver
```

`cmd/webhook-mock-receiver` 默认监听 `18090`，它是仓库内官方 mock receiver。feishu_link 独立本地联调服务推荐端口是 `18081`，两者是不同工具。

然后在后台配置：

```text
endpoint_url=http://127.0.0.1:18090
health_check_path=/health
auth_type=none
```
