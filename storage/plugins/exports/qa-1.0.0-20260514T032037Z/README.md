# 问答插件

- code: `qa`
- version: `1.0.0`
- description: 提供 question 内容类型、问题发布、回答、采纳和已解决状态。
- exported_at: `2026-05-14T03:20:37Z`
- devhub_core_version: `v1.4.0`
- source_status: `enabled`

## 内容类型

- `question`

## 权限摘要

- `qa.question.create` 发布问题
- `qa.question.audit` 审核问题
- `qa.answer.create` 提交回答
- `qa.answer.accept` 采纳回答

## 配置说明

`config.example.json` 仅为示例配置，不是当前环境配置备份；敏感字段使用 `REPLACE_ME` 占位。

## 依赖说明

- 无 required 依赖

## 安装方式

1. 将目录复制到 DevHub 本地插件仓库，例如 `storage/plugins/packages/`。
2. 在后台执行本地插件包 dry-run，检查 checksum、risk_report、依赖和 Core 兼容状态。
3. 通过校验后再按审批/安装流程安装，安装后默认 disabled，需要手动配置并启用。

## 安全边界

- 本导出包不包含第三方运行时代码。
- 本导出包不包含外部 SQL。
- 本导出包不包含敏感配置明文或密文。
- 本导出包不包含用户数据、通知数据、审计日志或 Hook 执行历史。
