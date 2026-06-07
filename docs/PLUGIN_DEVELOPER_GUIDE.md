# DevHub 声明型插件开发者指南

[返回文档入口](README.md)

更新时间：2026-05-29

本文面向想要编写 DevHub 插件包的开发者。当前 DevHub 支持的是 **声明型插件能力** 和 `external_service` **non-blocking Webhook 投递子集**；完整第三方运行模型仍未实现。

建议的起步顺序是：先复制模板，再改 `manifest.json`，再补 `migrations/`，最后走 upload -> precheck -> promote -> install dry-run -> install。

## 1. DevHub 插件系统能做什么

当前已经可复制使用的能力：

- 声明插件基本信息：`code`、`name`、`version`、`description`、`author`。
- 声明权限、内容类型、后台 / 前台菜单和配置 schema。
- 声明 Hook 元数据。
- 声明 `external_service` Webhook：DevHub 根据 Hook 声明异步投递 HTTP 请求。
- 声明官方 allowlist 前端挂载：只能选官方允许的 `mount_point` 与 `component_key`，不开放任意远程 iframe 或第三方前端运行时。
- 通过插件包 upload -> precheck -> promote -> install dry-run -> install。
- 通过后台启用、禁用、配置、健康检查、查看执行记录、查看审计和手动 retry。

v1.9.0-S1 稳定性补充：官方插件生态稳定版优先稳定现有声明型能力和 external_service non-blocking Webhook，不扩大第三方运行时边界。`feishu_link` receiver 已用 fresh standalone receiver 复核 success / fail / timeout / manual retry；开发者可以用 `scripts/run-feishu-webhook-flow.sh` 的 `DEVHUB_WEBHOOK_FLOW=full` 作为本地 external_service 回归参考。

v1.9.0-S2 SecretCenter 补充：插件开发者应继续把真实 token / secret 留在运行配置写入流程中，插件包、manifest、README 和示例配置只写占位说明。SecretCenter 详情会展示 `usage_type/source_type/source_id/source_code` 等脱敏来源字段，轮换入口会回到 external_service / Webhook Secret / Callback Token / 插件配置来源页；不要设计依赖 SecretCenter 直接回显或编辑明文的流程。

## 2. DevHub 插件系统不能做什么

当前明确不开放：

- 不执行第三方代码。
- 不支持 Go plugin。
- 不支持 JS 沙箱。
- 不支持远程 iframe。
- 不支持插件市场。
- 不支持远程在线安装。
- 不支持 blocking Hook。
- 不允许插件包内 SQL 直接执行。
- 不允许动态加载未知资产。

v1.9.0 继续保持这些边界：不开放插件市场、远程在线安装、自动安装、自动启用、package scripts、Go plugin、JS/WASM/Lua sandbox、远程 iframe、remote component 或 blocking Hook。

HookBus 只派发事件、记录执行、触发 `external_service` HTTP 投递；不会加载或执行插件包里的处理器代码。

## 3. 插件包标准结构

推荐结构：

```text
plugin-code/
  manifest.json
  README.md
  config.example.json
  migrations/
  docs/                 可选
  assets/               可选，仅限当前包规则允许的预览素材
  checksums.json        建议
  PACKAGING.md          建议
```

规则：

- `manifest.json` 是声明能力的核心。
- `migrations/` 是唯一迁移入口。
- 根目录 SQL 不作为执行入口。
- dry-run 只生成计划，不执行 SQL。
- 不允许危险文件、package scripts、真实 secret、用户数据、运行时代码或远程 iframe URL。
- `config.example.json` 只能放示例值，不能放真实 token。
- 官方模板默认会带 `checksums.json`，让预检更接近真实交付包。
- 如果插件使用 `external_service` token，运行配置应通过后台或 Admin API 写入 SecretCenter，之后只通过 `token_ref=secret://external_service/{plugin_code}/token` 引用；系统设置的“当前生效配置”只展示 endpoint 等非敏感字段和 token_ref / 状态 / key_id / 脱敏值 / config source / next_steps，不会回显 token 明文。SecretCenter 的轮换入口只会跳回来源配置，插件开发者不应设计依赖“在 SecretCenter 直接编辑明文”的流程。

## 4. manifest.json 编写指南

最小字段：

```json
{
  "code": "my_plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "description": "声明型插件示例",
  "author": "Your Team",
  "compatible_core_version": ">=1.8.3",
  "min_core_version": "1.8.3",
  "is_system": false
}
```

权限示例：

```json
{
  "permissions": [
    {
      "plugin_code": "my_plugin",
      "code": "my_plugin.item.create",
      "name": "创建内容",
      "scope": "community"
    }
  ]
}
```

内容类型示例：

```json
{
  "content_types": ["my_item"],
  "content_type_definitions": [
    {
      "type": "my_item",
      "name": "我的内容",
      "plugin_code": "my_plugin",
      "create_permission": "my_plugin.item.create",
      "edit_permission": "my_plugin.item.manage",
      "delete_permission": "my_plugin.item.manage",
      "audit_permission": "my_plugin.item.manage",
      "default_status": "draft",
      "seo_type": "Article"
    }
  ]
}
```

菜单示例：

```json
{
  "menus": [
    {
      "plugin_code": "my_plugin",
      "code": "my_plugin.admin",
      "title": "我的插件",
      "path": "/admin-next/plugins/overview?tab=content&plugin_code=my_plugin",
      "location": "admin",
      "area": "admin",
      "permission": "my_plugin.item.manage"
    }
  ]
}
```

配置 schema 示例：

```json
{
  "config_schema": {
    "type": "object",
    "additionalProperties": false,
    "required": ["enabled"],
    "properties": {
      "enabled": {
        "type": "boolean",
        "title": "启用插件",
        "default": true
      }
    }
  },
  "default_config": {
    "enabled": true
  }
}
```

Hook 示例：

```json
{
  "hooks": [
    {
      "plugin_code": "my_plugin",
      "name": "AfterCreateContent",
      "mode": "non_blocking",
      "failure_policy": "log",
      "timeout_ms": 3000
    }
  ]
}
```

`external_service` Webhook 示例：

```json
{
  "hooks": [
    {
      "plugin_code": "my_webhook",
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
  ]
}
```

依赖示例：

```json
{
  "dependencies": [
    {
      "code": "qa",
      "version": ">=1.0.0",
      "required": false,
      "reason": "可选复用问答入口"
    }
  ]
}
```

迁移声明示例：

```json
{
  "migrations": [
    {
      "plugin_code": "my_plugin",
      "migration_version": "1.0.0",
      "migration_name": "my_plugin_init",
      "direction": "up",
      "checksum": "sha256:template",
      "rollback_supported": false
    }
  ]
}
```

## 5. 模板一：纯声明型内容插件

模板路径：

```text
examples/plugins/templates/declarative-content/
```

该模板基于 `official_links` / 友情链接场景，适用于：

- 友情链接。
- 简单内容块。
- 站点配置扩展。
- 后台菜单扩展。
- 不需要外部服务的插件。

如果要参考生产化官方实现，请优先阅读 `examples/plugins/official_links/`。该包从 v1.8.4-S1 起作为官方友情链接插件维护，使用 `friend_link` content_type、通用 `PluginContent` 后台治理和搜索页前台展示，不包含运行时代码或插件私有业务表。

开发流程：

1. 复制模板。
2. 修改 `manifest.json` 的插件编码、名称、版本、权限码和 content_type。
3. 修改 README。
4. 准备 `migrations/`。
5. 使用 `./scripts/plugin-package-build.sh <插件目录>` 打包 zip。
6. 使用 `./scripts/plugin-package-check.sh <插件目录或 zip>` 本地校验，确认没有 blocker。
7. 上传预检。
8. promote。
9. install dry-run。
10. install。
11. enable。
12. 后台验证菜单、content_type、配置和权限。

## 6. 模板二：external_service Webhook 插件

模板路径：

```text
examples/plugins/templates/external-service-webhook/
```

该模板基于 `official_webhook_notify`，适用于：

- 内容发布通知。
- 外部审核系统。
- 消息推送。
- CRM / 工单系统集成。
- 低耦合异步扩展。

开发流程：

1. 复制模板。
2. 修改 Hook 声明。
3. 修改 external_service 配置 schema。
4. 实现并自行部署外部 receiver 服务。
5. 使用 `./scripts/plugin-package-build.sh <插件目录>` 打包插件。
6. 使用 `./scripts/plugin-package-check.sh <插件目录或 zip>` 本地校验，确认没有 blocker。
7. 上传预检。
8. 安装插件。
9. 后台配置 endpoint 和 token。
10. 执行 health check。
11. 触发业务事件。
12. 查看 `hook_executions`。
13. 失败后手动 retry。

## 7. 前端挂载

DevHub 当前只支持官方 allowlist 内的前端挂载，不支持任意远程 iframe 或第三方前端运行时。

开发者只能在 manifest 里声明：

- 官方允许的 `mount_point`。
- 官方允许的 `component_key`。
- `props_schema` 或 `config_ref`。

不允许声明：

- `iframe_url`。
- `script_url` / `remote_entry` / `remote_url`。
- `inline_html`。
- `remote_component`。
- 任意第三方前端 JS / CSS 执行入口。

前端挂载的典型使用方式：

1. 选择官方挂载点。
2. 选择官方组件 key。
3. 用 `props_schema` 描述组件需要的公开参数。
4. 把敏感配置放进后台配置，不写进前端 props。
5. 通过 upload -> precheck 先看阻断项。
6. 通过后台详情或升级 dry-run 查看挂载差异。

开发者验收时应确认：

- 允许列表内的挂载点可以通过。
- 未知挂载点会被 blocked。
- 未知组件 key 会被 blocked。
- 远程 iframe / 远程脚本 / 内联 HTML 会被 blocked。
- secret / token 不会进入前端 props。

## 8. 安全规范

- manifest 不写真实 token。
- `config.example.json` 不写真实 token。
- README 不写真实密钥。
- secret 字段只允许后台写入。
- token 不回显。
- Authorization Header 不落日志。
- Callback Token / Webhook Secret 不进审计明文。
- external_service token、Webhook Secret、Callback Token 等运行时敏感配置依赖启动加密密钥（`DEVHUB_PLUGIN_CONFIG_KEYS` 或 split 形式）；该 root key 只能来自启动环境变量或外部 Secret 系统注入，DevHub 后台不会保存或生成，修改后需重启生效。
- external_service bearer token 会由 Core SecretCenter 保存为 `token_ref=secret://external_service/{plugin_code}/token`；插件包、manifest、README 和示例配置只能写占位说明，不能写真实 token，也不能假设后台或 API 会回显明文。
- Webhook receiver 必须校验签名或 token。
- Webhook receiver 必须支持幂等。

## 9. 版本升级规范

- `version` 必须递增。
- `code` 不可变。
- publisher 变化要谨慎。
- 删除 permission / content_type / config 字段属于危险变更。
- migration 只追加，不依赖自动 down。
- 升级前应通过 upgrade dry-run 查看差异。
- `blocked` 不能强行升级。
- `warning` 需要管理员确认。
- 升级失败时，后台任务详情和执行结果会显示失败阶段、失败原因和下一步建议，方便管理员判断停在哪一步。

当前不提供完整自动回滚；涉及数据库迁移的失败需要管理员根据备份和迁移计划人工处理。

## 10. 常见错误

- `manifest.code` 与包目录名不一致。
- 缺少 `migrations/` 目录。
- 根目录放 `001_schema.sql`。
- 写入真实 token。
- 声明 unsupported `service_type`。
- 声明 blocking Hook。
- `endpoint_url` 非法。
- 删除 content_type 导致升级风险。
- config required 字段变化导致 warning。
- 插件 disabled / archived 状态下 Webhook 不投递。

## 11. 最小验收清单

纯声明型内容插件：

- zip 可预检。
- 可 promote。
- 可 install dry-run。
- 可 install。
- 可 enable。
- 后台菜单 / content_type / config 可见。
- disable 后能力不可用或按当前设计隐藏。
- uninstall / soft_uninstall 行为符合文档。

external_service Webhook 插件：

- zip 可预检。
- 可 promote。
- 可 install dry-run。
- 可 install。
- 可配置 endpoint。
- health check 可执行。
- 触发业务事件后 mock receiver 收到请求。
- `hook_executions` 有记录。
- 失败可 retry。
- token / secret 不回显。
- 不要把 `endpoint_url`、`health_check_path`、`token`、`timeout_ms`、`failure_policy` 写进全局 `config_json` 后期待它影响投递；Webhook 健康检查和投递只读取 external_service 运行配置。
- Docker 本地联调时，如果 DevHub 后端在容器中，`127.0.0.1` 指向容器自身；宿主机 receiver 请使用 `host.docker.internal`、Docker host gateway 地址或宿主机局域网 IP。
- 非 localhost HTTP endpoint 默认会被拒绝，本地开发可显式配置 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build`；生产建议使用 HTTPS。

## 12. 当前官方模板

| 模板 | 路径 | 作用 |
| --- | --- | --- |
| 官方友情链接插件 | `examples/plugins/official_links/` | v1.8.4-S1 生产化官方声明型插件，展示 `friend_link` content_type、权限、菜单、配置、no-op migrations/ 和通用 PluginContent 管理。 |
| 纯声明型内容插件 | `examples/plugins/templates/declarative-content/` | 从 `official_links` 场景出发，展示 content_type、权限、菜单、配置和 migrations/。 |
| external_service Webhook 插件 | `examples/plugins/templates/external-service-webhook/` | 从 `official_webhook_notify` 场景出发，展示 external_service non-blocking Hook、配置、health check 和 retry。 |

两个模板都不包含第三方运行时代码、真实 secret、远程 iframe、blocking Hook 或插件市场能力。

后台创建模板（v1.8.4-S9）：

- 管理员也可以在后台插件包治理页使用“创建插件模板 / 初始化插件包”。插件名称会自动生成 `plugin_code`，高级设置可手动修改。
- 插件类型可选内容型、外部服务型、后台工具型、前端挂载型；非内容型插件默认不显示或提交内容数据类型。
- “内容类型”在后台表单中已改为“内容数据类型”，“内容类型名”改为“内容显示名称”；发布者 / 作者支持 DevHub Team、当前组织、当前用户、可信发布者和自定义。
- preview 会展示 code、content_type（如适用）、权限、菜单、Hook、external_service / frontend_mount、`migrations/001_init.sql`、文件树和 conflicts；preview 不落盘。
- generate 固定写入 `storage/plugins/drafts/{code}/`，export zip 和 CLI `plugin:new` 使用同一套 `PluginTemplateGenerator`。模板只生成声明文件、`docs/*.md` 和 `migrations/001_init.sql`，不生成根目录 `001_schema.sql`、`registry.example.go`、Go/JS/WASM/binary 文件、package scripts 或 blocking Hook 模板，不安装、不启用、不执行 SQL、不执行代码。

## 13. 生产备份与升级演练

开发插件包时，除了关注 manifest / migrations / config_schema，还要提前想到生产安装和升级怎么做备份、怎么演练、失败后怎么恢复：

- 上线前先备份主数据库、插件安装记录、插件配置记录、插件配置加密 key、本地插件包仓库、上传暂存区、`admin_logs`、`hook_executions`、Webhook Secret 元数据、Callback Token 元数据、external_service 配置、站点 / 子站插件启用状态和 migrations 执行记录。
- 不要把真实 token / secret 写进 manifest、README、`config.example.json`、日志、审计或测试截图；生产备份也不应把明文导出到不安全位置。加密 key 丢失后，加密配置无法恢复。
- 安装前先跑 `scripts/plugin-package-check.sh`，确认 package check 通过、precheck 无 blockers、`migrations/` 结构合法、根目录 `001_schema.sql` 不执行、install dry-run 已从本地仓库包执行且 plan 未过期。
- 升级前确认当前版本、目标版本、upgrade dry-run、`risk_level`、manifest / permissions / content_types / config_schema / frontend_mount / external_service / migrations / dependencies diff，以及影响站点 / 子站、权限和菜单数量。
- `safe` 仍需备份；`warning` 升级必须显式 `confirm=true`；`blocked` 不能通过 confirm、审批或前端按钮绕过。
- dry-run plan 过期、checksum 不一致或 migration plan 不一致必须拒绝，不能复用旧计划。
- `failure_stage` 和 `failure_reason` 不是装饰字段，而是人工恢复入口：先看失败停在哪一步，再决定是重试、恢复配置、恢复本地仓库，还是回滚数据库备份。
- `PluginRegistry reload` 失败不代表自动回滚成功；通常应保留旧快照，查看 `plugin.registry.reload.failed` 审计，修好 manifest / 配置 / 依赖 / 解密问题后再重新 reload。
- `MemoryStore` 只适合本地开发 / 临时测试；生产环境建议以 MySQLStore 为准，并在 MySQL 模式验收插件安装、配置、`admin_logs`、`hook_executions` 和 upgrade dry-run。
- 当前不承诺完整自动 rollback 或 migration down；migration 已执行后的自动回退依赖未来版本规划。开发者不要把“disabled / archived”理解成完整回滚，它们只是停止新能力或进入治理入口。

详细清单和演练步骤见 [备份与回滚文档](BACKUP_AND_ROLLBACK.md)；开发者提交插件包前，也应按 [测试文档](TESTING.md) 中的 v1.8.4-S5 验收项做自检。
