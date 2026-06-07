# Hook 声明

`plugin_a7b0cc04` 声明 `AfterCreateContent` external_service Hook：

- 触发时机：`feishu_link` 内容创建成功后。
- 运行模式：`non_blocking`，不阻断主发布流程。
- 投递路径：`POST /hooks/content.after_create`。
- 重试：`retry_enabled=true`，最多 3 次。
- 执行记录：写入 `hook_executions`，可在后台 Webhook / external_service 执行记录中查看和手动重试。

注意：平台不加载第三方动态代码；Webhook 由 DevHub external_service 运行时按后台保存的 `endpoint_url` 发起 HTTP 投递。
