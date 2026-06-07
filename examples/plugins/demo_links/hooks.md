# Hook 声明

HookDefinition 当前是声明和治理边界，不是第三方动态 Hook 运行时。

- blocking Hook 失败会阻断主流程，并写入 hook_executions 与 plugin.hook.blocked 审计。
- non-blocking Hook 失败不阻断主流程，但会写入 hook_executions 与 plugin.hook.failed 审计。
- 当前不支持第三方动态 Hook 处理器。
- 当前不支持远程 Webhook 执行。
