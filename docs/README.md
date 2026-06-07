# DevHub 文档入口

DevHub 当前文档围绕稳定版 `v1.9.0` 官方插件生态稳定版维护。历史版本 Release Notes 和历史任务原文只作追溯依据，不作为当前 Codex 必读主列表。

补充：当前 `VERSION` 为 `v1.9.0`。发布归档与发布后 targeted smoke 已完成，P0=0、P1=0、P2=0，版本可以冻结；建议先提交当前文档归档，再创建 v1.9.0 annotated tag。v1.9.0 冻结后进入 v1.9.1 维护候选与 v1.10.0 规划阶段，不再向 v1.9.0 追加新功能。文档仍以“真实实现 + 设计边界分离”为原则维护；不把未开放的第三方代码执行、远程动态加载、任意远程 iframe、插件市场、远程在线安装、package scripts、自动安装、自动启用或 blocking Hook 写成已完成。当前前端插件挂载只支持官方 allowlist，仍不是开放式第三方前端运行时。

当前项目目标已统一为 **Core + 插件 的开源服务底座**：Core 提供稳定基础能力，插件承载业务扩展能力；默认社区能力是 Core 基础能力之一，不再作为项目唯一定位。

## 当前有效文档

1. [Codex / AI Agent 固定规则](AGENT_RULES.md)
   - 后续 Agent 任务必须遵守的项目名称、端口、入口、SEO、环境、文档同步和验收边界规则。

2. [项目总览](../README.md)
   - 项目定位、技术栈、启动入口、当前能力和目录结构。

3. [项目进度](PROJECT_PROGRESS.md)
   - 当前版本结论、已完成、部分完成、未完成、风险、下一步和当前验收清单。

4. [API 文档](API.md)
   - 当前真实 API、认证要求、插件 API、发布校验和集中规划 / 未完成接口。

5. [SEO 文档](SEO.md)
   - `/topics/:id`、子站页、标签页、sitemap 和 robots 的 SEO 保护要求。

6. [插件架构说明](PLUGIN_ARCHITECTURE.md)
   - Core 与 Plugins 边界、两层插件状态、qa / docs / wiki 内置插件和当前限制。

7. [完整插件系统路线图](PLUGIN_SYSTEM_ROADMAP.md)
   - 长期最高优先级目标，定义插件生命周期、治理、运行时、审计、迁移、后台和 E2E 要求；当前稳定版本为 `v1.9.0` 官方插件生态稳定版，下一阶段为 v1.9.1 维护候选 / v1.10.0 规划。

8. [插件运行模型设计](PLUGIN_RUNTIME_MODEL.md)
   - 定义 Core 内置插件、外部 HTTP 服务插件、iframe / sandbox 前端插件三种运行模式，以及前端挂载、受控 API、HookBus、隔离边界、manifest 运行字段和官方示例插件验证方向。

8. [插件前端挂载模型（设计）](PLUGIN_FRONTEND_MOUNT_MODEL.md)
   - 定义插件前端 slots、iframe/sandbox 隔离策略与 postMessage 通信协议（设计口径）；明确插件前端不能绕过 Core 权限、审计与插件生命周期状态。

9. [Webhook / HTTP 插件服务协议（设计）](PLUGIN_WEBHOOK_PROTOCOL.md)
   - 定义 Core 调用外部插件服务的协议：事件类型、blocking/non_blocking、请求格式、签名鉴权、防重放、幂等与重试、超时/限流/熔断、回调 Core API 的受控模型、审计与后台治理规划。

10. [Webhook 协议实现拆解（v1.7.3）](PLUGIN_WEBHOOK_IMPLEMENTATION_PLAN.md)
   - 将协议设计拆成可落地阶段：non_blocking delivery、delivery 记录、重试队列、熔断、签名与鉴权、后台治理入口；并明确 blocking Hook 后置。

11. [Webhook 插件使用方法](PLUGIN_WEBHOOK_USAGE.md)
   - 面向管理员的实操指南：external_service non-blocking 投递、Webhook Secret、Callback Token、健康检查、投递记录和排障入口。

补充（v1.7.8 已落地）：

- 官方 mock receiver：`cmd/webhook-mock-receiver/main.go`（用于端到端验签/失败注入/重试熔断验证；不执行第三方代码）

11.1 [external_service Webhook 技术债记录](PLUGIN_WEBHOOK_TECH_DEBT.md)
   - 记录本地联调中暴露的 Docker loopback、HTTP allowlist、全局配置与运行配置混淆、升级入口和健康摘要误导等技术债。

12. [插件 SDK 文档](PLUGIN_SDK.md)
   - 插件声明规范、生命周期、manifest 字段、内容类型、权限、菜单、配置、Hook、migration 和安全边界。

13. [声明型插件开发者指南](PLUGIN_DEVELOPER_GUIDE.md)
   - 面向插件开发者的当前可用能力说明、不能做的能力边界、插件包结构、manifest 编写指南、纯声明型内容插件模板、external_service Webhook 模板、安全规范、升级规范和验收清单；建议从这里开始，再复制两个官方模板。

14. [插件生成模板](PLUGIN_TEMPLATE.md)
   - `go run ./cmd/devhub plugin:new` 脚手架用法、生成目录、校验规则和模板边界。

15. [本地插件包规范（草案）](PLUGIN_PACKAGE.md)
   - 插件包目录结构、允许/危险文件规则、大小限制，以及本地插件包 dry-run 导入预览接口与后台入口。
   - v1.7 补充：远程包 staging 下载、compat-check 与启用前安全检查（enable-precheck）均只做安全治理与结论输出，不会安装/启用/注册/执行。
   - v1.8.4-S4 补充：`scripts/plugin-package-build.sh` 可生成官方包 / 模板 zip，`scripts/plugin-package-check.sh` 可对目录、zip 或内置 `official_announcement` 做本地校验，输出 passed / warning / blocked。

16. [备份与回滚文档](BACKUP_AND_ROLLBACK.md)
   - 生产备份、恢复、紧急回滚和 v1.8.4-S5 插件安装 / 升级演练清单；明确加密 key、Webhook Secret / Callback Token 元数据、本地插件仓库、`failure_stage`、PluginRegistry reload、MySQLStore 与 MemoryStore 边界。

17. [测试文档](TESTING.md)
   - 已实现必测项、后续补测项、必要历史回归和 SEO 回归命令。

18. [后台 UI 基础设计规范](ADMIN_UI_GUIDELINES.md)
   - 后台治理型控制台 design tokens、通用组件、插件中心 / SecretCenter / Webhook / 当前生效配置视觉收口规则。

19. [v1.9.0 Release Notes](releases/v1.9.0.md)
   - 当前稳定版本口径：官方插件生态稳定版；v1.9.0 已完成发布归档与发布后 targeted smoke，P0=0 / P1=0 / P2=0。

20. [v1.8.4 Release Notes](releases/v1.8.4.md)
   - 当前版本口径：官方插件生态与生产可用性增强；v1.8.4-S1 将 `official_links` 推进为生产可用官方友情链接插件包。

21. [v1.8.3 Release Notes](releases/v1.8.3.md)
   - 后台插件治理稳定性与中文体验、PluginRegistry reload、声明型插件可用闭环、真实插件包验收、external_service non-blocking Webhook 和真实声明型插件“安装到使用”闭环。

20. [v1.7.2 Release Notes](releases/v1.7.2.md)
   - 插件运行模型设计：Core 内置插件、外部 HTTP 服务插件、iframe / sandbox 前端插件、受控 API、HookBus、隔离边界和 manifest 运行字段设计。作为历史设计背景保留。

21. [v1.8.0 Release Notes](releases/v1.8.0.md)
   - 官方插件前端挂载模型与 iframe/sandbox 容器设计：slots、postMessage 协议与权限/状态 gating（文档设计，不修改代码）。

22. [v1.8.1 Release Notes](releases/v1.8.1.md)
   - 官方公告插件前端挂载最小实现：内置官方插件 + Host + iframe + postMessage 最小闭环（不执行第三方代码，不允许远程 iframe URL）。

23. [v1.8.2 Release Notes](releases/v1.8.2.md)
   - iframe / sandbox 通用容器与 postMessage Host helper：前台首页、`/c/:slug` 与后台插件详情复用统一挂载机制（仅 allowlist 官方内置插件；仍不允许远程 iframe URL）。

24. [v1.7.3 Release Notes](releases/v1.7.3.md)
   - Webhook / HTTP 插件服务协议实现拆解：以 non_blocking delivery 为第一优先级，拆解 delivery 记录、重试队列、熔断、签名鉴权与后台治理入口；并准备官方公告插件端到端验证方案。本轮只改文档，不新增真实投递实现。

25. [v1.7.1 Release Notes](releases/v1.7.1.md)
   - 插件包 detached signature（devhub-signature.json）验签与可信发布者增强：Ed25519 真实验签、验签记录、与 compat-check/install/upgrade 联动、默认阻断 unsigned。

26. [v1.7.0 Release Notes](releases/v1.7.0.md)
   - 远程插件包治理与安装安全增强：远程包安全下载到 staging、解压安全检查与 manifest 预校验、compat-check、安装事务/回滚、enable-precheck、enable、软卸载与升级任务闭环（不执行第三方代码、不自动更新、不做市场）。

27. [v1.6.0 Release Notes](releases/v1.6.0.md)
   - 插件包上传与分发前置能力：zip 上传安全沙箱、上传包生命周期、真实签名验签、可信发布者、远程索引、版本仓库、失败恢复预览、配置密钥轮换和后台 UI 收口。

28. [v1.5.0 Release Notes](releases/v1.5.0.md)
   - 插件包治理收口：本地插件包规范、dry-run、checksum / 风险报告、仓库扫描、安装闭环、配置版本历史、敏感配置加密、审批流、导出与签名/可信来源草案。

29. [v1.4.0 Release Notes](releases/v1.4.0.md)
   - 插件内容治理增强：精确过滤、批量治理、审计闭环和当前验收记录。

30. [v1.3.5 Release Notes](releases/v1.3.5.md)
    - 插件治理中心信息架构、完整安装 / 升级向导、批量归档 / 恢复影响预览、状态治理页和最小 E2E 回归。

31. [部署启动文档](DEPLOYMENT.md)
   - 本地启动、构建行为、8090 端口排查、Go 模块网络和二进制排障启动。

32. [Docker 一键线上部署](DOCKER_DEPLOYMENT.md)
   - Ubuntu 服务器 Docker Compose 部署 DevHub + MySQL，包含 `.env.docker`、构建启动、日志、状态和数据卷说明。

## 历史版本归档

- [v1.0.0 Release Notes](releases/v1.0.0.md)
- [v1.1.0 Release Notes](releases/v1.1.0.md)
- [v1.1.1 Release Notes](releases/v1.1.1.md)
- [v1.1.3 Release Notes](releases/v1.1.3.md)
- [v1.1.4 Release Notes](releases/v1.1.4.md)
- [v1.1.5 Release Notes](releases/v1.1.5.md)
- [v1.2.0 Release Notes](releases/v1.2.0.md)
- [v1.2.1 Release Notes](releases/v1.2.1.md)
- [v1.3.0 Release Notes](releases/v1.3.0.md)
- [v1.3.1 Release Notes](releases/v1.3.1.md)
- [v1.3.2 Release Notes](releases/v1.3.2.md)
- [v1.3.3 Release Notes](releases/v1.3.3.md)
- [历史产品需求原文](archive/2026-05-09-product-requirements.md)

## 辅助文档

- [Web 目录说明](../web/README.md)：前台 / 后台源码目录说明。
- [后台前端说明](../web/admin-app/README.md)：Vue 后台开发说明。
- [插件 manifest 示例](examples/plugin-manifest-example.json)：可通过当前 ManifestValidator 的声明型插件示例。
- [official_links 官方友情链接插件包](../examples/plugins/official_links)：v1.8.4-S1 起生产化的官方声明型内容插件。
- [official_webhook_notify 官方 Webhook 样板插件](../examples/plugins/official_webhook_notify)：v1.8.4-S3 起生产化的官方 external_service Webhook 通知样板。
- [纯声明型内容插件模板](../examples/plugins/templates/declarative-content)：基于 `official_links` 场景的可复制插件包模板。
- [external_service Webhook 插件模板](../examples/plugins/templates/external-service-webhook)：基于 `official_webhook_notify` 场景的可复制插件包模板。
- [主题与 UI 架构](THEME_AND_UI_ARCHITECTURE.md)：前台主题 token 和 UI 架构说明。
- [UI 样式指南](UI_STYLE_GUIDE.md)：前台样式规范。
- [历史文档归档](archive/README.md)：不再作为当前验收依据的历史规划说明。

## 维护规则

- 每轮遗留项统一沉淀到 [项目进度](PROJECT_PROGRESS.md) 的“当前版本结论 / 已完成 / 部分完成 / 未完成 / 下一阶段目标”结构。
- 版本范围内的限制写入对应 Release Notes；测试缺口写入 [测试文档](TESTING.md)。
- [API 文档](API.md) 以真实可用接口为主；未实现接口只能集中放在“规划 / 未完成”小节。
- [项目总览](../README.md) 只保留项目定位、入口和当前能力概览，不承载详细未完成问题。
- 根目录不再保留临时任务大文档；历史任务原文统一归档到 [历史文档归档](archive/README.md)。
- 新增文档前，优先判断能否合并进现有文档，避免文档继续膨胀。
