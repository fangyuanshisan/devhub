# 插件权限指南

插件权限必须由 manifest / registry 声明，并由后端强校验。前端隐藏按钮或菜单只是体验优化，不能作为权限边界。

## 命名规范

推荐格式：

- `{plugin}.{resource}.create`
- `{plugin}.{resource}.edit`
- `{plugin}.{resource}.delete`
- `{plugin}.{resource}.audit`
- `{plugin}.manage`
- `{plugin}.configure`

示例：

- `qa.question.create`
- `docs.document.create`
- `wiki.page.create`
- `projects.project.create`
- `jobs.job.create`
- `ai_works.work.create`

## Core 兼容权限

- `article` / `news` 当前继续使用 `core.topic.create`。
- `post.create` 仅作为历史兼容桥，不是长期主权限。
- 后续应逐步把内容创建入口全部收敛到 content type 对应 create 权限。

## 后端强校验

后端必须校验：

- `plugin_code`
- `content_type`
- `community_id`
- `category_id`
- 插件全局状态
- 子站插件状态
- 当前 actor 权限

普通用户 token 不能访问后台插件治理 API。非授权版主不能管理其他子站插件内容。

## 平台管理权限

当前平台治理动作优先复用已有后台插件管理权限：

- `plugin.read`：查看插件列表、健康、impact、审计和迁移信息。
- `plugin.write`：启用 / 禁用、配置、归档 / 恢复、dry-run、manifest validate、install、批量归档 / 恢复。
- `plugin.approve`：审批/执行高风险治理动作（例如本地插件包安装、插件升级审批执行）。

后续若要拆分更细权限，可再预留：

- `plugin.archive`
- `plugin.restore`
- `plugin.install`
- `plugin.upgrade`
- `plugin.dry_run`

## 作用域

当前权限声明支持：

- `global`
- `community`
- `category`
- `own`

版主权限必须受 community scope 限制。
