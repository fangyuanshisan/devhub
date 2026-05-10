# DevHub 插件架构说明

[返回文档大纲](README.md)

更新时间：2026-05-10

## 版本定位

`v1.3.0` 是 Core + Plugins 架构拆分版。Core 不再把问答、文档、Wiki 作为硬编码核心类型处理，而是通过内置系统插件注册内容类型、菜单、权限和路由描述。

## Core 边界

Core 保留通用能力：

- 用户、前台认证、后台认证、子站版主授权。
- 子站、板块、通用内容、评论、标签、搜索、通知、收藏、关注、举报、审计日志。
- 后台 RBAC、版主 scope 校验、SEO 动态页和 sitemap / robots。
- 插件注册、插件状态、插件菜单、插件权限描述。

当前物理表兼容历史命名：

- `topics` 是 Core 内容表，对应需求中的 `contents`。
- `categories` 是 Core 板块表，对应需求中的 `channels`。

## 插件边界

内置插件目录：

- `internal/plugins/qa`：注册 `qa` 插件和 `question` 内容类型。
- `internal/plugins/docs`：注册 `docs` 插件和 `document` 内容类型。
- `internal/plugins/wiki`：注册 `wiki` 插件和 `wiki_page` 内容类型。

插件状态：

- `installed`：已安装但未启用。
- `enabled`：已启用，可发布对应内容。
- `disabled`：已禁用，对应板块不能继续发布新内容。

## 发布校验

`POST /api/v1/topics` 会执行以下校验：

- 当前 `category` 必须属于目标 `community`。
- `category.plugin_code` 必须匹配 `content_type` 对应插件。
- 插件类型必须是 `enabled`。
- `content_type` 必须在 `category.allowed_content_types` 中。
- 历史 `doc` / `wiki` 会归一为 `document` / `wiki_page`。

## 数据表

新增：

- `plugins`
- `qa_questions`
- `qa_answers`
- `docs_spaces`
- `docs_documents`
- `wiki_spaces`
- `wiki_pages`
- `wiki_page_versions`

增强：

- `topics.plugin_code`
- `categories.plugin_code`
- `categories.allowed_content_types`

## 当前限制

- 暂未做插件市场、压缩包上传安装、远程更新。
- 插件路由当前以注册描述和 Core 分发为主，没有引入独立运行时加载器。
- 文档树、Wiki 协作编辑和版本回滚已具备表结构与基础注册，完整专用编辑 UI 留到后续迭代。
