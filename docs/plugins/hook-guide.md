# Hook 开发指南

Hook 是 DevHub 插件运行时的扩展点。当前 HookBus 只服务内置系统插件，不执行第三方动态代码，不支持远程 Hook 或 Webhook。

## Hook 类型

blocking Hook：

- `BeforeCreateContent`
- `BeforeUpdateContent`
- `BeforeDeleteContent`
- `BeforeModerateContent`

blocking Hook 返回错误时必须阻断主流程，并写入 `hook_executions` 与 `plugin.hook.blocked` 审计。

non-blocking Hook：

- `AfterCreateContent`
- `AfterUpdateContent`
- `AfterCreateComment`
- `OnSearchIndex`
- `OnNotificationBuild`
- `OnSEOBuild`

non-blocking Hook 失败不阻断主流程，但必须记录失败并写入 `plugin.hook.failed` 审计。

## HookContext

Hook 上下文至少包含：

- `hook_name`
- `plugin_code`
- `content_type`
- `content_id`
- `community_id`
- `category_id`
- `actor_type`
- `actor_id`
- `user_id`
- `admin_user_id`
- `request_id`
- `metadata`

## 执行记录

`hook_executions` 记录：

- hook 名称、插件编码、模式、是否 blocking。
- 内容、子站、操作者和 request_id。
- 开始 / 结束时间、耗时、成功状态、错误信息和 metadata。

## 当前边界

- Search / Notification / SEO Hook 当前可作为最小事件派发和记录点，不代表已完成完整搜索、通知、SEO 业务处理器。
- 外部插件 Hook、远程 Hook、Webhook、脚本沙箱均为后续生态设计，不在当前运行时范围。
