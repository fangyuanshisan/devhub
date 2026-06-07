# Packaging

最小打包流程：

```bash
cd examples/plugins/templates/external-service-webhook
zip -qr ../../official-webhook-notify-template.zip .
```

验收流程：

1. 上传 zip。
2. 查看预检结果。
3. promote 到本地仓库。
4. 执行 install dry-run。
5. 确认安装。
6. 后台配置 external_service endpoint。
7. 执行 health check。
8. 触发业务事件并查看 Webhook 治理记录。
9. 失败后手动 retry。

注意：

- 包内不包含 receiver 运行时代码。
- 包内不写真实 token / secret。
- `migrations/` 仍是唯一迁移入口。
- `service_type=external_service` 只能使用 `mode=non_blocking`。
- DevHub 只投递 HTTP 请求，不执行第三方代码。
