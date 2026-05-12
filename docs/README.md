# DevHub 文档入口

DevHub 当前文档只围绕 `v1.3.4` 真实状态和 `v1.3.5` 下一阶段草案维护。历史版本 Release Notes 和历史任务原文只作追溯依据，不作为当前 Codex 必读主列表。

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
   - 长期最高优先级目标，定义插件生命周期、治理、运行时、审计、迁移、后台和 E2E 要求；当前已完成 `v1.3.4` 插件异常治理与平台基础能力收口，下一阶段草案为 `v1.3.5` 插件治理体验与安装升级向导收口。

8. [测试文档](TESTING.md)
   - 已实现必测项、后续补测项、必要历史回归和 SEO 回归命令。

9. [v1.3.4 Release Notes](releases/v1.3.4.md)
   - 插件异常治理、Manifest 校验 / dry-run / 安装、归档 / 恢复、最小升级执行、failed migration 阻断、Hook 失败注入、权限矩阵、MySQLStore 专项和当前限制。

10. [v1.3.5 Draft](releases/v1.3.5.md)
    - 下一阶段迭代集合：插件治理中心信息架构、完整安装 / 升级向导、批量归档 / 恢复影响预览、状态治理页和最小 E2E 回归。

10. [部署启动文档](DEPLOYMENT.md)
   - 本地启动、构建行为、8090 端口排查、Go 模块网络和二进制排障启动。

11. [备份与回滚文档](BACKUP_AND_ROLLBACK.md)
    - v1.x 上线前后需要备份的内容、MySQL 备份恢复、二进制回滚、Git 回滚和紧急回滚流程。

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
