# DevHub 插件架构说明

[返回文档入口](README.md)

更新时间：2026-05-10

## 版本定位

`v1.3.0` 是 Core + Plugins 架构拆分版。问答、文档、Wiki 不再作为 Core 硬编码业务类型描述，而是由 `qa`、`docs`、`wiki` 三个内置系统插件注册内容类型、菜单、权限和路由描述。

## Core 边界

Core 只保留通用社区能力：

- 用户、前台认证、后台认证、子站版主授权。
- 子站、板块、通用内容、评论、标签、搜索、通知、收藏、关注、举报、审计日志。
- 权限、后台 RBAC、版主 scope 校验。
- `/topics/:id`、子站页、标签页、sitemap 和 robots 的 SEO 兜底。
- 插件注册、插件状态、插件菜单、插件权限描述和基础分发。

当前兼容命名：

- `topics` 是当前通用内容表，对应目标架构中的 `contents`。
- `categories` 是当前通用板块表，对应目标架构中的 `channels`。

## 内置系统插件

- `qa`：问答插件，提供 `question` 内容类型，并承载问题、回答、采纳和已解决状态。
- `docs`：文档插件，提供 `document` 内容类型，并预留文档空间、文档树和文档详情能力。
- `wiki`：Wiki 插件，提供 `wiki_page` 内容类型，并预留页面版本、回滚和协作编辑能力。

`project`、`job`、`ai_work` 当前仍是 Core 兼容内容类型或后续插件候选，尚未完整拆成独立插件。

## 两层插件状态

插件状态分两层：

- `plugins.status`：全局插件状态，表示插件是否在系统层面可用。
- `community_plugins.status`：子站插件状态，表示某个子站是否启用该插件。

状态值：

- `installed`：已安装 / 已注册，但未启用。
- `enabled`：已启用。
- `disabled`：已禁用。

插件在某个子站可用必须同时满足：

- 插件已注册。
- `plugins.status = enabled`。
- `community_plugins.status = enabled`。
- 当前板块 `categories.plugin_code` 匹配插件。
- 当前 `content_type` 在 `categories.allowed_content_types` 内。
- 当前用户具备对应权限。

特殊说明：

- `core` 是兼容内置能力，Service 层视为始终可用，不要求写入 `plugins` 或 `community_plugins`。
- 禁用全局插件会影响所有子站的新发布、导航、菜单和管理入口。
- 禁用子站插件只影响该子站的新发布、导航、菜单和后台管理入口。
- 禁用插件不影响历史内容访问，尤其不能破坏 `/topics/:id` SEO 动态 HTML。

## 发布校验流程

发布 Topic 时应走统一插件校验：

1. 解析 community。
2. 解析 category，并校验 category 属于 community。
3. 归一 `content_type`：`doc -> document`，`wiki -> wiki_page`。
4. 根据 `content_type` 推断 `plugin_code`：`question -> qa`，`document -> docs`，`wiki_page -> wiki`，其他兼容类型 -> `core`。
5. 校验全局插件状态。
6. 校验当前子站插件状态。
7. 校验 `category.plugin_code` 是否匹配。
8. 校验 `content_type` 是否在 `category.allowed_content_types` 内。
9. 校验当前用户权限。
10. 写入归一后的 `topics.content_type` 和 `topics.plugin_code`。

当前真实状态：

- 步骤 1-8 已在 `ValidateTopicPluginAccess` 和 Store 层板块校验中落地。
- 步骤 9 的插件权限码细粒度发布拦截仍是部分完成 / 待实现。

## 数据结构

插件相关表：

- `plugins`
- `community_plugins`
- `qa_questions`
- `qa_answers`
- `docs_spaces`
- `docs_documents`
- `wiki_spaces`
- `wiki_pages`
- `wiki_page_versions`

增强字段：

- `topics.plugin_code`
- `categories.plugin_code`
- `categories.allowed_content_types`

## API 与菜单

已注册的插件 API 包括：

- `GET /api/v1/plugins`
- `GET /api/v1/communities/:slug/plugins`
- `GET /api/v1/admin/plugins`
- `POST /api/v1/admin/plugins/:code/enable`
- `POST /api/v1/admin/plugins/:code/disable`
- `GET /api/v1/admin/plugin-menus`
- `GET /api/v1/admin/communities/:id/plugins`
- `POST /api/v1/admin/communities/:id/plugins/:code/enable`
- `POST /api/v1/admin/communities/:id/plugins/:code/disable`
- `PUT /api/v1/admin/communities/:id/plugins/:code/config`
- `PUT /api/v1/admin/communities/:id/plugins/sort`
- `GET /api/v1/moderator/plugin-menus`

菜单策略：

- 后台左侧导航只保留“系统插件”入口。
- 插件业务管理页通过系统插件列表进入，避免 qa / docs / wiki 直接散落到左侧导航。
- 版主插件菜单必须同时满足全局 enabled、子站 enabled、当前用户是该子站版主、当前用户具备菜单权限。

## 当前限制

- 当前阶段不做插件市场。
- 当前阶段不做插件包上传。
- 当前阶段不做远程插件下载或在线更新。
- 当前阶段不做 Go 动态插件加载。
- 插件路由当前是注册描述 + Core 分发，不是真正动态运行时路由加载。
- Docs / Wiki 的专用编辑体验仍是部分完成。
- 子站插件配置和排序已有 API，但后台 UI 仍需继续完善和验收。
