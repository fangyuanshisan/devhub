# DevHub 项目进度

[返回文档入口](README.md)

更新时间：2026-05-10

本文档只记录当前仓库真实状态、当前风险和下一步任务。历史版本能力已并入当前分支，详情见对应 Release Notes；旧版本已解决问题不再占用当前主体。

## 当前版本结论

当前版本为 `v1.3.0`，主题是“Core + Plugins 架构拆分版”。DevHub 当前定位为多子站通用开源社区程序，默认演示为开发者社区。

Core 保留用户、认证、子站、板块、通用内容、评论、标签、搜索、通知、SEO、权限、审计、插件注册和分发能力。问答、文档、Wiki 已按内置系统插件建模：`qa -> question`、`docs -> document`、`wiki -> wiki_page`。

当前实现仍保留历史表名以保证兼容：`topics` 是当前通用内容表，`categories` 是当前通用板块表。`project`、`job`、`ai_work` 等仍是 Core 兼容内容类型或后续插件候选，不要写成已经完整插件化。

## 当前已完成

- 插件注册：`internal/plugins/registry.go` 和 `internal/plugins/qa|docs|wiki` 提供内置插件定义、内容类型映射、菜单、权限和路由描述。
- 全局插件状态：`plugins` 表和 MemoryStore / MySQLStore 均支持 `installed`、`enabled`、`disabled`，并提供全局插件列表、启用和禁用 API。
- 子站插件状态：`community_plugins` 表和 MemoryStore / MySQLStore 均支持按子站启用 / 禁用、配置和排序插件。
- 两层状态判断：插件在某个子站可用需要同时满足 `plugins.status=enabled` 和 `community_plugins.status=enabled`；`core` 作为兼容内置能力在 Service 层特殊视为可用。
- 内容模型兼容：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 已进入 schema 与 Store。
- 发布校验：`POST /api/v1/topics` 已走统一 `ValidateTopicPluginAccess`，会归一 `doc -> document`、`wiki -> wiki_page`，并校验插件状态、子站插件状态、板块插件绑定和允许内容类型。
- 板块管理校验：MemoryStore / MySQLStore 在创建或编辑子站板块时校验 `plugin_code` 与 `content_type` 匹配，并拒绝绑定全局或子站未启用的插件。
- 插件 API：全局插件 API、子站插件 API、前台子站插件展示 API 和版主插件菜单 API 已在 `router.go` 注册。
- 前台入口：子站插件公开接口会隐藏 `config_json` 等后台配置；子站板块导航会按子站插件状态过滤。
- 后台入口：`/admin-next/plugins` 作为系统插件管理入口；插件业务页通过系统插件列表进入，默认不散落在后台左侧导航。
- Wiki schema：当前只保留插件化 `wiki_spaces`、`wiki_pages`、`wiki_page_versions` 语义，旧 `wiki_revisions` 预留冲突已清理。
- SEO 保护：`/topics/:id` 仍由 Go 动态输出 SEO HTML，插件禁用不影响历史内容详情访问。

## 当前部分完成

- 子站插件管理 UI：后台已有基础查看和启用 / 禁用入口；`config_json` 可视化编辑与排序控件仍需补齐或专项验收。
- 插件权限：后台菜单和版主菜单已按权限过滤；发布链路尚未按 `qa.question.create`、`docs.document.create`、`wiki.page.create` 等插件权限码做细粒度用户权限拦截。
- 插件权限：发布链路已按内容类型做权限码拦截（`question/docs/wiki_page`），Core 兼容类型当前仍为粗粒度 `post.create`。
- Docs / Wiki 业务体验：已具备表结构、注册定义和通用内容发布 / 展示能力；文档树、版本历史、回滚和协作编辑 UI 仍是后续专项。
- 插件路由：当前是注册描述 + Core 分发，不是真正动态运行时加载器。
- 验收覆盖：已做文档与路由核对；完整 Docker 启动、真实 token API、浏览器页面和 SEO curl 矩阵仍需按测试文档继续补测。

## 当前未完成

- 子站插件 `config_json` 可视化编辑和排序 UI。
- 发布链路的插件权限码细粒度校验。
- Docs 文档树专用编辑 UI。
- Wiki 版本历史、版本回滚和协作编辑交互。
- 插件市场、插件包上传、远程插件安装、在线更新和 Go 动态插件加载均不属于当前阶段。

## 当前风险

- 历史数据可能存在 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 或 `community_plugins` 缺失 / 不一致，生产升级前需要迁移演练和抽样校验。
- 子站插件禁用后已有内容应继续可读；后续改发布、列表或 SEO 时要避免把禁用插件误当作历史内容 404 条件。
- API 已注册不等于完整产品闭环；子站插件配置和排序目前更偏 API 能力，后台 UI 仍需继续完善。
- `/sitemap.xml` 当前仍是单文件动态输出，内容规模扩大后需要 sitemap index / 分片。
- 用户提出的 `docs/BACKUP_ROLLBACK.md` 与仓库真实文件名不一致；当前真实文件是 `docs/BACKUP_AND_ROLLBACK.md`。

## 下一步任务

1. 补齐后台子站插件配置 UI：`config_json` 编辑、排序控件、禁用影响提示和失败提示。
2. 将插件权限码接入发布校验，避免前台或后台绕过 `qa.*`、`docs.*`、`wiki.*` 权限。
3. 用真实 admin/user token 补测全局插件、子站插件、版主菜单和跨子站发布矩阵。
4. 补跑 `/topics/:id` SEO curl 检查，确认插件禁用后历史内容源码不退化。
5. 为 Docs / Wiki 规划最小可用专用管理体验，但不引入复杂编辑器或协作系统。

## 当前验收清单

- [ ] `go test ./...`
- [ ] `go build` 或 `go build -buildvcs=false ./...`
- [ ] `cd web/frontend-app && npm run build`
- [ ] `cd web/admin-app && npm run build`
- [ ] `GET /api/v1/plugins` 只返回全局 enabled 插件。
- [ ] `GET /api/v1/communities/:slug/plugins` 只返回全局 enabled 且子站 enabled 插件。
- [ ] 管理员可以查看、启用和禁用全局插件。
- [ ] 管理员可以查看、启用和禁用某个子站插件。
- [ ] 子站禁用 `qa` 后，该子站不能发布 `question`；其他启用 `qa` 的子站不受影响。
- [ ] 子站禁用 `docs` 后，该子站不能发布 `document`。
- [ ] 子站禁用 `wiki` 后，该子站不能发布 `wiki_page`。
- [ ] 板块不能绑定当前子站未启用的插件。
- [ ] 前台发布页只展示当前子站可发布的内容类型。
- [ ] 版主插件菜单只返回全局 enabled、子站 enabled 且当前用户有权限的插件菜单。
- [ ] 禁用插件后，已有 `/topics/:id` 详情页仍可访问并保留 SEO HTML。
- [ ] `/sitemap.xml` 和 `/robots.txt` 正常返回。
