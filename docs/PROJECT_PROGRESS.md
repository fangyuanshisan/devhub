# DevHub 项目进度

[返回文档入口](README.md)

更新时间：2026-05-10

本文档只记录当前仓库真实状态、当前风险和下一步任务。历史版本能力已并入当前分支，详情见对应 Release Notes；旧版本已解决问题不再占用当前主体。

## 当前版本结论

当前版本为 `v1.3.1`，主题是“插件化关键入口封口与权限校验补强版”。DevHub 当前定位为多子站通用开源社区程序，默认演示为开发者社区。

Core 保留用户、认证、子站、板块、通用内容、评论、标签、搜索、通知、SEO、权限、审计、插件注册和分发能力。问答、文档、Wiki、项目、招聘、AI 作品已按内置系统插件建模：`qa -> question`、`docs -> document`、`wiki -> wiki_page`、`projects -> project`、`jobs -> job`、`ai_works -> ai_work`。

当前实现仍保留历史表名以保证兼容：`topics` 是当前通用内容表，`categories` 是当前通用板块表。

## 当前已完成

- 插件注册：`internal/plugins/registry.go` 和 `internal/plugins/qa|docs|wiki|projects|jobs|aiworks` 提供内置插件定义、内容类型映射、菜单、权限和路由描述。
- 插件声明规范：当前已统一到 manifest 风格声明，包含插件本体、内容类型定义、权限定义、菜单定义、路由定义、`config_schema`、依赖、最小 Core 版本和 Hook 声明。
- 全局插件状态：`plugins` 表和 MemoryStore / MySQLStore 均支持 `installed`、`enabled`、`disabled`，并提供全局插件列表、启用和禁用 API。
- 子站插件状态：`community_plugins` 表和 MemoryStore / MySQLStore 均支持按子站启用 / 禁用、配置和排序插件。
- 两层状态判断：插件在某个子站可用需要同时满足 `plugins.status=enabled` 和 `community_plugins.status=enabled`；`core` 作为兼容内置能力在 Service 层特殊视为可用。
- 内容模型兼容：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 已进入 schema 与 Store。
- 发布校验：`POST /api/v1/topics` 已走统一 `ValidateTopicPluginAccess`，会归一 `doc -> document`、`wiki -> wiki_page`，并校验插件状态、子站插件状态、板块插件绑定和允许内容类型。
- 板块管理校验：MemoryStore / MySQLStore 在创建或编辑子站板块时校验 `plugin_code` 与 `content_type` 匹配，并拒绝绑定全局或子站未启用的插件。
- 插件 API：全局插件 API、子站插件 API、前台子站插件展示 API 和版主插件菜单 API 已在 `router.go` 注册。
- 插件业务闭环：
  - `qa`：发布 `question` 时写入 `qa_questions`；回答写入 `qa_answers`；采纳后回写已解决状态和最佳答案。
  - `docs`：发布 `document` 时写入 `docs_documents`，并支持基础文档树读取。
  - `wiki`：发布 `wiki_page` 时写入 `wiki_pages` 和初始 `wiki_page_versions`；编辑时新增版本记录。
- `project` / `job` / `ai_work` 已完成插件归属迁移：`projects -> project`、`jobs -> job`、`ai_works -> ai_work`，发布校验、权限码、菜单声明和历史 `plugin_code` 迁移口径已接入；专属扩展表和完整业务闭环尚未完成。
- 权限上下文：`CreateTopicRequest.ActorPermissions` / `ActorContext` 均由服务端从 token、后台身份和版主 scope 计算，客户端请求体不能覆盖。
- 配置合并：`plugins.config_json` 与 `community_plugins.config_json` 已落库并可写，返回 `resolved_config.default/global/community/effective` 合并视图。
- HookBus：Service 层已有最小内部 HookBus，当前调用点覆盖 `BeforeCreateContent`、`AfterCreateContent` 和 `AfterCreateComment`，不做第三方动态执行。
- 前台入口：子站插件公开接口会隐藏 `config_json` / `resolved_config` 等后台配置；子站板块导航会按子站插件状态过滤。
- 后台入口：`/admin-next/plugins` 作为系统插件管理入口；插件业务页通过系统插件列表进入，默认不散落在后台左侧导航。
- 最小体验闭环：
  - 后台全局插件管理已支持查看插件列表、基础声明与全局启用/禁用。
  - 后台子站插件配置已支持启用/禁用、`config_json` 编辑和数字排序保存。
  - 前台子站页和发布页会按当前子站已启用插件收口入口与内容类型。
  - 版主工作台已补最小插件治理入口区，并按当前子站插件状态与权限过滤。
- 审计：全局插件状态、子站插件状态、子站插件配置和排序已接入 `admin_logs`；当前 old/new 以 `target` 文本摘要形式记录。
- Wiki schema：当前只保留插件化 `wiki_spaces`、`wiki_pages`、`wiki_page_versions` 语义，旧 `wiki_revisions` 预留冲突已清理。
- SEO 保护：`/topics/:id` 仍由 Go 动态输出 SEO HTML，插件禁用不影响历史内容详情访问。
- 技术债收口：`Service.CreatePost` 已封口，不再作为业务写入口；`/api/v1/posts` 写接口继续废弃；后台 `admin/posts` 创建入口在兼容 `post.create` 基础权限之外，叠加真实内容类型对应的插件 create 权限。
- 后台编辑边界：后台内容编辑已禁止修改子站、板块、`content_type` 和 `plugin_code` 归属字段；如后续需要迁移归属，必须走单独迁移专项和完整插件校验。

## 当前部分完成

- 子站插件管理 UI：后台已具备最小配置闭环，包括启用 / 禁用、`config_json` 编辑、JSON 校验和数字排序；多浏览器矩阵、批量操作和更强可视化仍待后续专项验收。
- 插件权限：后台菜单和版主菜单已按权限过滤；发布链路已按内容类型做最小权限码校验：
  - `question -> qa.question.create`
  - `document -> docs.document.create`
  - `wiki_page -> wiki.page.create`
  - `project -> projects.project.create`
  - `job -> jobs.job.create`
  - `ai_work -> ai_works.work.create`
  - Core 兼容类型 `article`、`news` 当前仍为粗粒度 `core.topic.create`（兼容旧 `post.create`，后续可按内容类型细化）。
- 权限兼容桥：`post.create` 仍作为 `core.topic.create` 的过渡兼容权限存在；它不是长期主权限，后续需要随 Core 内容定义或 article/news 插件化一起收口。
- Docs / Wiki 业务体验：当前已具备基础空间、文档树读取、版本列表等最小闭环；拖拽排序、完整回滚 UI、协作锁和专用编辑体验仍待后续。
- Projects / Jobs / AI Works 业务体验：当前完成插件归属、发布校验、权限码和菜单声明；专属扩展表、专属管理页和完整业务流程仍待后续。
- 插件路由：当前是注册描述 + Core 分发，不是真正动态运行时加载器。
- Hook 机制：当前已有最小内部 HookBus，但只覆盖创建内容与创建评论相关调用点；Update/Delete/Search/Notification/SEO Hook 调度仍待完善。
- 配置校验：当前已完成默认配置、全局配置、子站配置三层合并和 JSON 格式校验；`config_schema` 强校验仍待后续。
- 验收覆盖：已做文档与路由核对；完整 Docker 启动、真实 token API、浏览器页面和 SEO curl 矩阵仍需按测试文档继续补测。

## 当前未完成

- 子站插件配置 UI 的完整浏览器验收矩阵，包括多子站、禁用提示、保存失败提示和排序持久化回归。
- 更细粒度的权限体系：例如 Core 兼容类型 `article` / `news` 的细分权限码、按子站/板块维度配置权限矩阵、以及更明确的错误码与权限配置 API（当前仍为最小校验闭环）。
- HookBus 的完整调用点：`BeforeUpdateContent`、`AfterUpdateContent`、`BeforeDeleteContent`、`AfterDeleteContent`、`OnSearchIndex`、`OnNotificationBuild`、`OnSEOBuild`。
- `config_schema` 强校验和更友好的配置表单渲染。
- `admin_logs` 的 old/new diff 结构化字段；当前插件治理 diff 仍主要以 `target` 文本摘要记录。
- `qa` 取消采纳最佳答案。
- Docs 文档树专用编辑 UI。
- Docs 文档树拖拽、批量排序和更完整的空间管理体验。
- Wiki 版本回滚接口与协作编辑交互。
- Projects / Jobs / AI Works 的专属扩展表和完整业务闭环。
- 插件市场、插件包上传、远程插件安装、在线更新和 Go 动态插件加载均不属于当前阶段。

## 当前风险

- 历史数据可能存在 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 或 `community_plugins` 缺失 / 不一致，生产升级前需要迁移演练和抽样校验。
- 子站插件禁用后已有内容应继续可读；后续改发布、列表或 SEO 时要避免把禁用插件误当作历史内容 404 条件。
- API 和最小 UI 已注册不等于完整产品闭环；子站插件配置、排序和插件入口仍需继续做真实浏览器矩阵验收。
- `/sitemap.xml` 当前仍是单文件动态输出，内容规模扩大后需要 sitemap index / 分片。
- 用户提出的 `docs/BACKUP_ROLLBACK.md` 与仓库真实文件名不一致；当前真实文件是 `docs/BACKUP_AND_ROLLBACK.md`。

## 下一步任务

1. 用真实 admin/user token 补测全局插件、子站插件、版主菜单和跨子站发布矩阵。
2. 完成子站插件配置 UI 浏览器验收：`config_json`、排序、禁用影响提示、失败提示和保存后刷新。
3. 细化插件权限矩阵，尤其是 Core 兼容类型 `article` / `news` 的权限码策略。
4. 补跑 `/topics/:id` SEO curl 检查，确认插件禁用后历史内容源码不退化。
5. 为 Docs / Wiki 规划最小可用专用管理体验，但不引入复杂编辑器或协作系统。
6. 如需要后台迁移内容子站、板块或类型，设计单独迁移 API，并逐条校验插件状态、子站插件状态、板块绑定、allowed_content_types 和权限码。

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
- [ ] `Service.CreatePost` 不能绕过插件发布校验。
- [ ] `POST/PUT/DELETE /api/v1/posts*` 写接口返回 `410 Gone` 或明确废弃。
- [ ] `POST /api/v1/admin/posts` 创建 `question/document/wiki_page` 时分别需要 `qa.question.create`、`docs.document.create`、`wiki.page.create`。
- [ ] 后台编辑内容不能修改子站、板块、`content_type` 或 `plugin_code`。
