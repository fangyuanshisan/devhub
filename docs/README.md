# DevHub 文档大纲

DevHub 文档已收敛为少量当前有效入口。日常优先看下面这些文档。

## 当前有效文档

1. [Codex / AI Agent 固定规则](AGENT_RULES.md)
   - 后续 Agent 任务必须遵守的项目名称、端口、入口、SEO、环境、文档同步和验收边界规则。

2. [项目总览](../README.md)
   - 项目定位、目录结构、启动方式、入口、API 概览。

3. [项目进度](PROJECT_PROGRESS.md)
   - 当前做到哪里、最近几轮结果、已完成、风险、下一步、验收清单。

4. [API 文档](API.md)
   - 当前真实 API 路径、请求参数、响应结构、错误返回、认证要求、评论问答、举报治理、版主管理、批量治理和审计日志能力。

5. [SEO 文档](SEO.md)
   - 百度 SEO 兜底策略，尤其是 `/topics/:id` 动态 HTML 的保护要求。

6. [v1.1.3 Release Notes](releases/v1.1.3.md)
   - 独立版主工作台 MVP 的页面、API、权限边界、审计日志和已知限制。

7. [v1.1.1 Release Notes](releases/v1.1.1.md)
   - 前后台身份边界整理版的身份模型、token 边界、middleware、审计 actor 和已知限制。

8. [v1.1.0 Release Notes](releases/v1.1.0.md)
   - 子站模块增强版的版本定位、数据结构变化、API、SEO、后台、测试清单和已知限制。

9. [测试文档](TESTING.md)
   - 页面、接口、互动、SEO、memory / mysql 模式的手工验收清单。

10. [部署启动文档](DEPLOYMENT.md)
   - 本地启动、构建行为、8090 端口排查、Go 模块网络和二进制排障启动。

11. [备份与回滚文档](BACKUP_AND_ROLLBACK.md)
   - v1.x 上线前后需要备份的内容、MySQL 备份恢复、二进制回滚、Git 回滚和紧急回滚流程。

12. [v1.0.0 Release Notes](releases/v1.0.0.md)
   - 首个可运行大版本的版本定位、启动部署、测试清单、已知限制和下一版本规划。

13. [需求原文](../更新.md)
   - 产品需求原始文档，用于核对目标和范围。

## 辅助文档

- [Web 目录说明](../web/README.md)：前台 / 后台源码目录说明。
- [后台前端说明](../web/admin-app/README.md)：Vue 后台开发说明。
- [历史文档归档](archive/README.md)：不再作为当前验收依据的历史规划说明。

## 维护规则

- 当前状态变化后，优先更新 [项目进度](PROJECT_PROGRESS.md)。
- API 变化后，同步更新 [API 文档](API.md) 和 [项目总览](../README.md) 的 API 概览。
- 页面入口、启动方式变化后，同步更新 [项目总览](../README.md)、[部署启动文档](DEPLOYMENT.md) 和 [测试文档](TESTING.md)。
- SEO 相关变化后，同步更新 [SEO 文档](SEO.md)。
- 版本归档变化后，同步更新对应 [Release Notes](releases/v1.1.3.md)、[变更日志](../CHANGELOG.md) 和根目录 [VERSION](../VERSION)。
- Agent 协作规则变化后，同步更新 [AGENT_RULES.md](AGENT_RULES.md)，不要散落在临时对话里。
- 新增文档前，优先判断能否合并进现有文档，避免文档继续膨胀。
