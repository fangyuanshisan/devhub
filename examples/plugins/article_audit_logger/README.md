# Article Audit Logger

这是一个独立运行的 external_service Webhook 插件示例。DevHub 不执行插件包代码；插件包只声明订阅关系，真正的接收端在 `cmd/article-audit-logger` 中单独启动。

## 当前能力

- 订阅 `AfterCreateContent`。
- DevHub 在用户发布内容后异步 `POST {endpoint_url}/hooks/content.after_create`。
- 独立接收端默认只记录 `content_type=article` 的日志。
- 投递记录进入 `hook_executions(service_type=external_service)`，可在后台 Webhook 治理页查看和重试。
- 接收端可通过 `ARTICLE_AUDIT_LOGGER_CONTENT_TYPE` 扩展到其他内容类型，为后续多内容审计系统预留入口。

## 本地启动

```bash
ARTICLE_AUDIT_LOGGER_ADDR=:18110 \
go run ./cmd/article-audit-logger
```

如需简单鉴权：

```bash
ARTICLE_AUDIT_LOGGER_TOKEN=dev-secret \
go run ./cmd/article-audit-logger
```

同时在后台 external_service 配置中把 `auth_type` 设置为 `bearer`，token 填同一个值。

## 插件配置

1. 在后台本地插件包 dry-run 中输入 `examples/plugins/article_audit_logger`，确认后安装。
2. 启用插件，并在需要审计的子站中启用插件。
3. 在插件详情的 external_service 配置中填写：
   - `endpoint_url`: `http://127.0.0.1:18110`
   - `health_check_path`: `/health`
   - `timeout_ms`: `3000`
   - `failure_policy`: `warn`
   - `auth_type`: `none`，或使用 `bearer` 配合 `ARTICLE_AUDIT_LOGGER_TOKEN`
4. 用户发布文章后，接收端会输出类似日志：

```text
article_published event_id=evt_xxx execution_id=extsvc_exec_xxx topic_id=123 community_id=1 actor_type=user actor_id=7 title="..." status=1 source_plugin_code=core occurred_at=...
```

## 安全边界

- 插件包不包含运行时代码，不执行外部 SQL。
- 接收端是独立进程，由部署方自行运维。
- Webhook 为 non-blocking，接收端失败不会阻断用户发布文章。
- 后续接多内容审计系统时，可以替换接收端内部逻辑，保留 manifest 中的订阅和投递治理能力。
