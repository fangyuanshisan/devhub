# DevHub 项目进度

[返回文档大纲](README.md)

更新时间：2026-05-10

本文档记录当前仓库的真实实现状态。需求来源以根目录 `更新.md` 为准；旧规划文档仅作背景参考。

后续 Codex / AI Agent 任务应先阅读 `docs/AGENT_RULES.md`，再以本文档判断当前进度和版本边界。本文档只记录真实状态；未来规划不得写成已完成能力。

## 当前结论

DevHub 当前是 “Go API + Astro 前台 + Vue 后台” 的多子站社区 CMS。默认入口和端口保持不变：

```text
前台：http://127.0.0.1:8090/
后台：http://127.0.0.1:8090/admin-next
```

首页、子站页、搜索页和用户中心使用 Astro 静态壳 + 运行时 API；Topic 详情页 `/topics/:id` 仍由 Go 动态输出 SEO HTML。第五轮互动闭环已完成基础实现：点赞、收藏、关注、我的收藏、我的关注、我的动态、通知列表和已读逻辑在 MemoryStore 与 MySQLStore 中均可用。第六轮已完成评论列表、发表评论、回复评论、问答采纳、未解决筛选、评论动态和评论通知的基础闭环。第七轮已完成举报、版主范围治理、精华、置顶、隐藏、评论锁定和后台最小治理入口。第八轮补丁后，admin-next 后台内容 CRUD、版主管理 CRUD、批量治理、批量举报处理和治理审计日志均有真实页面入口和 API 封装。

当前版本为 `v1.3.0`，版本主题是“Core + Plugins 架构拆分版”。本轮把问答、文档、Wiki 从核心内容类型中拆分为内置系统插件：Core 保留用户、子站、板块、通用内容表、评论、标签、权限、通知、搜索、审计和插件注册；`qa`、`docs`、`wiki` 插件分别注册自己的内容类型、菜单、权限和路由描述。为兼容历史实现，物理表名仍保留 `topics` / `categories`，它们分别作为 Core `contents` / `channels` 的当前落地表，并新增 `plugin_code`、`allowed_content_types` 等字段。

v1.2.1 标签治理增强已并入当前分支，版本主题是“标签合并、别名与统计重算版”。该版本在 v1.2.0 基础上新增标签别名、标签合并、标签统计重算、merged / disabled 标签的 SEO 与 sitemap 治理、后台标签治理审计和 `/admin-next/tags` 管理增强。

v1.1.4 补丁已并入当前分支，主题是“前台登录态与权限入口修复版”。本补丁统一前台 user token 的导航恢复、子站关注、“我的”类页面请求和版主入口展示；普通前台会员不再显示 `/admin-next` 总后台入口；发布页按 `categories.content_type` 自动匹配板块；后台子站管理主入口统一为 `/admin-next/communities`，`/admin-next/sites` 保留隐藏兼容重定向。

v1.1.5 补丁已并入当前分支，主题是“前台 UI 美化专项”。本补丁只优化前台样式和响应式体验，统一全局颜色、字体层级、间距、卡片、按钮、标签、空状态、导航、首页、子站页、Topic 列表、Topic 详情、搜索页、发布页、“我的”页面和版主工作台入口视觉表现；不修改 API、Store、数据库、路由、鉴权、业务逻辑或 Go 动态 SEO 结构。

2026-05-10 已补充主题与 UI 架构文档，并开始把前台样式 token、站点静态配置、内容类型映射和页面重复文案从页面文件中收拢。当前蓝白科技风被明确记录为默认主题 `default-blue`，不是唯一主题；前台结构已为后续多主题、多站点、多频道配置预留扩展空间。

## 近期迭代摘要

- 第一轮：同步通用社区 schema，补齐 PHP、Go、Java、AI、Frontend 五个子站 seed 数据。
- 第二轮：前台逐步迁移到 `communities/topics/search` API，并保留 `sites/posts` 兼容 API。
- 第三轮：完成发布入口、发布页、`POST /api/v1/topics`，MemoryStore / MySQLStore 均支持创建 Topic。
- 第四轮：增强 `/api/v1/search/topics`，支持关键词、范围、子站、分类、标签、排序、分页组合筛选；新增 `/api/v1/tags/hot`。
- SEO 整改：`/topics/:id` 改由 Go 动态输出 SEO 友好的 HTML；`/sitemap.xml` 和 `/robots.txt` 改为 Go 动态输出。
- 前台登录完善：首页登录/注册接入 `/api/v1/auth/*`，导航支持会话恢复、refresh token 刷新和退出登录。
- 第五轮互动联动：补齐点赞、收藏、关注、我的收藏、我的关注、我的动态、通知中心和详情页互动状态；新增用户中心页面入口。
- 第六轮评论问答：补齐 `GET/POST /api/v1/topics/:id/comments`、回复、采纳最佳答案、`sort=unsolved` 未解决筛选、`commented/accepted_answer` 动态和 `topic_commented/comment_replied/answer_accepted` 通知。
- 第七轮社区治理：补齐 topic/comment 举报、举报后台处理、版主子站权限、精华、置顶、隐藏、恢复、评论锁定、评论隐藏和 sitemap 隐藏过滤。
- 第八轮后台治理：补齐 admin-next 内容 CRUD 写入 `topics`、版主管理 CRUD、topic/comment/report 批量治理、举报重复 pending 限制、`/admin-next/moderators` 和 `/admin-next/audit-logs`。
- 第九轮 v1.0.0 归档：补齐 `VERSION`、`CHANGELOG.md`、`docs/releases/v1.0.0.md`、`docs/BACKUP_AND_ROLLBACK.md`、GitHub Actions CI、测试矩阵、部署归档和文档对账。
- v1.1.0 子站模块增强：增强 `communities/categories` 模型，新增 `/c/:slug` Go 动态 SEO 子站页、子站统计、公开版主、子站后台配置、子站板块管理、子站公告、子站关注计数和 sitemap 子站收录。
- v1.1.1 身份边界整理：前台登录发放 `token_type=user`，后台登录发放 `token_type=admin`；前台推荐 localStorage key 为 `devhub_user_token` / `devhub_user_refresh_token`，后台继续使用 sessionStorage `devhub_admin_token` / `devhub_admin_refresh_token`；`/api/v1/admin/*` 默认校验后台身份，子站版主 user token 只获得自己子站的治理类权限；`admin_logs` 增加并读写 `actor_type` / `actor_id`。
- v1.1.3 独立版主工作台 MVP：新增 `/moderator`、`/moderator/reports`、`/moderator/topics`、`/moderator/comments`、`/moderator/audit-logs`；新增 `/api/v1/moderator/*` 专用 API，复用现有 reports/topics/comments 治理能力和 `admin_logs`，但强制使用前台 user token 与 `community_moderators` 子站 scope。
- v1.1.4 前台登录态与权限入口修复：修复部分前台页面不显示登录状态、子站关注和“我的”类接口未携带 user token、普通会员看到总后台入口、版主入口不按身份展示、发布 question 默认板块不匹配，以及后台重复子站入口问题。
- v1.1.5 前台 UI 美化专项：统一前台视觉 token、导航、卡片、表单、标签、空状态、Topic 列表 / 详情、搜索、发布、“我的”页面和移动端样式；未改接口、业务逻辑、数据库、路由或 SEO 主体结构。
- v1.2.0 标签系统增强：新增 `/tags/:tag/` 全站标签 SEO 页和 `/c/:slug/tags/:tag/` 子站标签 SEO 页，公开标签详情 / 聚合 / 建议 API，发布页标签建议，标签关注，`/admin-next/tags` 后台标签 CRUD、启用 / 禁用、SEO 字段和关联内容查看，以及 sitemap 全站 / 子站标签收录。
- v1.2.1 标签治理增强：新增标签 alias 管理、标签 merge、标签统计重算、merged 状态和 `merged_to_id`、`tag_aliases` 表、alias / merged 301 解析、后台标签治理审计，以及 admin-next 标签别名 / 合并 / 重算入口。
- v1.3.0 Core + Plugins 架构拆分：新增 `plugins` 表与插件注册服务；`topics` 增加 `plugin_code`，`categories` 增加 `plugin_code` / `allowed_content_types`；新增 `qa`、`docs`、`wiki` 内置插件目录、插件权限、插件菜单和插件运行状态 API；新增 `qa_questions`、`qa_answers`、`docs_spaces`、`docs_documents`、`wiki_spaces`、`wiki_pages`、`wiki_page_versions`；发布内容时校验板块绑定插件、插件启用状态和允许内容类型。
- 主题与 UI 架构整理：新增 `docs/THEME_AND_UI_ARCHITECTURE.md` 和 `docs/UI_STYLE_GUIDE.md`；前台新增 `tokens.css` 作为主题 token 入口；首页、搜索页、子站页、发布页改为更多复用共享站点配置与内容类型映射，减少页面级硬编码。

## v1.3.0 Core + Plugins 架构拆分进度

已完成：

- 内置系统插件定义：`qa` / `docs` / `wiki`（内容类型、菜单、权限、路由描述）。
- 插件运行时状态：`plugins` 表 + `installed/enabled/disabled`，并提供启用/禁用 API 与后台入口。
- 子站插件运行时状态：`community_plugins` 表 + 子站启用/禁用/配置/排序 API；插件需同时满足“全局 enabled + 子站 enabled”才可用于发布、导航和菜单。
- Core 表字段增强：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types`。
- 发布校验：`POST /api/v1/topics` 已统一走 `ValidateTopicPluginAccess`，对 `plugin_code`、全局插件状态、子站插件状态、板块绑定和 `allowed_content_types` 做一致性校验，并兼容 `doc/wiki -> document/wiki_page`。
- 前台发布页：内容类型选择按“启用插件 + 板块 allowed_content_types”收口。
- 子站插件：发布、板块绑定、子站导航、公开插件列表和版主插件菜单均受子站插件状态影响。
- Wiki schema：仅保留插件化 `wiki_spaces`、`wiki_pages`、`wiki_page_versions`，已清理旧预留 `wiki_revisions` 冲突。

部分完成：

- 子站插件管理 UI：后台已有子站插件启用/禁用入口；`config_json` 编辑和排序控件仍需继续补齐。
- 插件权限：后台菜单和版主菜单已按权限过滤；发布时按 `qa.question.create`、`docs.document.create`、`wiki.page.create` 等权限码做细粒度拦截仍需继续接入。
- docs/wiki 专用编辑体验：已具备表结构、注册定义和通用内容发布/展示能力；专用文档树、版本历史和回滚 UI 仍待专项增强。

未完成：

- 子站插件 `config_json` 可视化编辑与排序 UI。
- 发布链路的插件权限码细粒度校验。
- Docs 文档树专用编辑 UI、Wiki 版本历史 / 回滚交互。

风险：

- 历史数据中可能存在 `categories.plugin_code/allowed_content_types` 为空或不一致的情况，依赖兼容分支推断为 `core`；建议生产升级时先执行迁移 SQL 并抽样校验板块配置。
- 全量 Docker 启动、真实 admin/user token API、浏览器页面和 SEO curl 验收仍需按 `docs/TESTING.md` 继续补测。

下一步：

- 补齐子站插件配置 UI 的 `config_json` 编辑和排序操作。
- 将发布权限码检查并入统一插件校验，避免后台或前台绕过插件权限。
- 按 `docs/TESTING.md` 补跑 v1.3.0 子站插件禁用、跨子站发布、迁移和 SEO 验收。

## v1.1.5 前台 UI 美化专项范围

已完成能力：

- 全局样式：统一前台 CSS 变量，覆盖主色、辅助色、成功 / 警告 / 危险色、背景、卡片、边框、阴影、圆角、页面宽度和响应式断点。
- 顶部导航：优化 Logo、子站切换、搜索框、登录 / 注册、用户菜单、通知、发布按钮和版主工作台入口的视觉层级；普通会员不显示总后台入口的 v1.1.4 规则保持不变。
- 首页：优化总站聚合 header、频道卡片、精选内容、子站推荐、最新内容、热门内容、热门标签和右侧信息栏样式。
- 子站页：优化 `/c/:slug` 运行时前台页面的子站 header、主题色 accent、关注 / 发帖按钮、板块导航、排序 tab、Topic 列表、热门标签、公告和版主侧栏视觉表现。
- Topic 列表：统一 Topic 卡片、内容类型徽章、子站 / 板块标识、标签、置顶 / 精华 / 已解决 / 未解决等状态徽章和统计信息样式。
- Topic 详情：优化标题、面包屑、作者与元信息、正文排版、代码块、引用、表格、标签、互动按钮和评论区样式；未改变 Go 动态 SEO 输出结构。
- 搜索页：优化搜索表单、筛选区、当前条件摘要、结果列表和分页样式。
- 发布页：优化表单、输入框、下拉、标签选择、错误 / 成功提示、提交按钮和说明侧栏样式；未改变发布逻辑和 `content_type` 校验。
- 我的页面：优化 `/me/favorites`、`/me/follows`、`/me/activities`、`/notifications` 的卡片、空状态、未读提示和快捷入口样式。
- 移动端：补充 `1024px`、`760px`、`480px` 响应式规则，改善单列布局、导航换行、搜索表单、标签换行、Topic 卡片和发布页可用性。

已知限制：

- v1.1.5 不是完整设计系统重构，未新增组件库、主题后台或复杂动效。
- 本轮只做 CSS / 视觉层级优化，不做 Playwright 全量视觉回归；完整桌面 / 移动端截图验收应作为后续验收任务。
- 深色主题做基础适配，后续仍可单独做主题系统专项。

未改变内容：

- 未修改 API、Store、数据库 schema、鉴权、关注、发布、评论、版主权限或 admin-next 业务。
- 未修改 `/topics/:id`、`/c/:slug`、`/tags/:tag` 的 Go 动态 SEO 主体结构。
- `sites/posts` 兼容 API 继续保留。

## v1.1.4 前台登录态与权限入口修复范围

已完成能力：

- 前台 Header 统一使用 `devhub_user_token` / `devhub_user_refresh_token`，兼容旧 token key，刷新页面后通过 `/api/v1/auth/me` 恢复用户状态。
- `/api/v1/auth/me` 返回 `is_moderator` 和 `moderated_communities`，用于前台决定是否显示版主工作台入口。
- 普通前台会员菜单不再显示 `/admin-next` 总后台入口；子站版主显示 `/moderator` 入口。
- Go 动态 SEO 页 `/c/:slug` 和 `/topics/:id` 运行时 Header 增加前台 user token 恢复与版主入口展示，不改变 SEO 主体 HTML。
- 子站关注、我的收藏、我的关注、我的动态和通知页面请求携带前台 user token，未登录时展示友好登录提示。
- 发布页根据 `?type=` / `?content_type=` 和当前子站板块 `content_type` 自动选择匹配板块；后端校验使用 `category.content_type` 优先、旧 `type` 兜底。
- 后台菜单只保留 `/admin-next/communities` 一个子站管理入口；`/admin-next/sites` 保留隐藏兼容并重定向到 `/admin-next/communities`。

已知限制：

- v1.1.4 不做复杂 RBAC、独立权限矩阵、OAuth / SSO 或完整认证系统重构。
- 总后台入口仍需要用户主动访问 `/admin-next` 并通过后台 admin 登录态进入，不由前台会员菜单暴露。

## v1.2.1 标签治理增强范围

已完成能力：

- 标签别名：后台支持 `GET/POST/DELETE /api/v1/admin/tags/:id/aliases`，同一 `site_key` 范围内 alias_slug 去重，并禁止与现有 `tags.slug` 冲突。
- 标签解析：`GET /api/v1/tags/:tag`、`GET /api/v1/tags/by-slug/:tag`、`GET /api/v1/communities/:slug/tags/:tag`、`GET /api/v1/tags/suggestions` 与 `GET /api/v1/search/topics?tag=` 均支持 alias 归一。
- 标签合并：后台支持 `POST /api/v1/admin/tags/:id/merge`，source tag 合并后变为 `status=merged`，并写入 `merged_to_id`。
- 迁移与去重：MySQLStore 合并时迁移 `topic_tags`、迁移并去重 `follows`、迁移 `tag_aliases`；MemoryStore 同步更新帖子标签和关注记录。
- 统计重算：后台支持 `POST /api/v1/admin/tags/:id/recalculate` 与 `POST /api/v1/admin/tags/recalculate-all`，重算 `topic_count`、`follower_count`、`hot_score`。
- SEO / sitemap：alias URL 与 merged source URL 301 到 canonical 主标签 URL；disabled / merged / alias URL 不进入 sitemap。
- 后台标签管理：`/admin-next/tags` 新增状态筛选、别名查看 / 添加 / 删除、标签合并、单个重算、全部重算和 merged 状态展示。
- 审计：标签别名、新增 / 删除别名、标签合并、单个 / 全量重算均写入治理 audit logs。

已知限制：

- 当前真实实现继续沿用 `tags.site_key` 作为子站范围字段，未把 `community_id` 直接写入 `tags` 表。
- 标签热度当前沿用简化公式：`hot_score = topic_count * 10 + follower_count * 20`。
- 标签趋势统计、运营分析和大规模异步统计任务仍留待后续版本。

## v1.2.0 标签系统增强范围

已完成能力：

- 标签页：`/tags/:tag/` 和 `/c/:slug/tags/:tag/` 由 Go 动态输出 SEO HTML，包含 title、description、canonical、h1、说明、内容链接、子站链接、相关标签和关注按钮。
- 标签详情 API：`GET /api/v1/tags/:tag`、`GET /api/v1/tags/by-slug/:tag`、`GET /api/v1/communities/:slug/tags/:tag` 按 slug、名称或 ID 获取启用标签。
- 标签内容聚合：`GET /api/v1/tags/:tag/topics` 和 `GET /api/v1/communities/:slug/tags/:tag/topics` 支持 `content_type`、`sort`、分页，返回真实 Topic 列表；`sort=unsolved` 可聚合未解决问答。
- 标签关注：复用 `POST /api/v1/follows/toggle`，`target_type=tag`，MemoryStore / MySQLStore 都会维护 `follower_count`。
- 发布页标签建议：`GET /api/v1/tags/suggestions` 和别名 `GET /api/v1/tags/suggest` 按当前子站返回启用标签，发布页最多选择 5 个标签。
- 后台标签管理：`/admin-next/tags` 支持列表、筛选、新增、编辑、启用 / 禁用、SEO 字段、前台跳转和关联内容查看。
- 标签 SEO 字段：`seo_title`、`seo_description`、`seo_keywords` 已进入 domain、MemoryStore、MySQLStore 和 schema。
- sitemap：`/sitemap.xml` 追加启用全站标签 canonical `/tags/:slug/` 和启用子站标签 canonical `/c/:slug/tags/:tag/`，禁用标签不收录。
- 链接调整：Topic、子站、首页、搜索和收藏列表中的标签链接指向真实标签页；子站上下文优先 `/c/:slug/tags/:tag/`，总站上下文使用 `/tags/:tag/`。
- SEO 保护：`/topics/:id` 和 `/c/:slug` 仍由 Go 动态输出 SEO HTML，未改成 CSR。

已知限制：

- v1.2.0 不做标签趋势统计；标签合并和标签别名已在 v1.2.1 完成。
- 标签趋势图和更复杂的标签运营分析留到后续版本。
- sitemap 仍是单文件动态输出，内容规模扩大后需要分片。

## v1.1.3 独立版主工作台 MVP 范围

已完成能力：

- 前台页面：`/moderator` 工作台首页、举报处理、内容治理、评论治理和审计日志页面已生成静态壳，运行时调用 `/api/v1/moderator/*`。
- 专用 API：`GET /api/v1/moderator/communities`、`GET /api/v1/moderator/dashboard`、`GET/POST reports`、`GET/POST topics`、`GET/POST comments`、`GET audit-logs` 已落地。
- 权限边界：版主 API 只接受前台 `users` token；普通用户返回 403；后台 admin token 不作为版主工作台身份。
- 子站 scope：所有列表和单项治理都会按 `community_moderators.user_id + community_id + status=1` 校验，跨子站访问返回 403。
- 举报治理：版主可查看和处理自己子站举报，处理后沿用现有隐藏 topic/comment 联动规则。
- Topic 治理：版主可对自己子站 topic 执行精华 / 取消精华、置顶 / 取消置顶、隐藏 / 恢复、锁定 / 解锁评论。
- Comment 治理：版主可隐藏 / 恢复自己子站评论；最佳答案不能隐藏的既有保护仍保留。
- 审计：版主工作台治理操作写入 `admin_logs`，`actor_type=moderator`、`actor_id=users.id`，并带子站 scope；版主审计页只读取负责子站日志。
- 前端入口：前台用户菜单新增“版主工作台”入口；完整后台 `/admin-next` 仍保留给后台人员。

已知限制：

- v1.1.3 不做复杂 RBAC、权限点矩阵、版主任期、绩效统计或独立版主登录体系。
- 版主工作台是轻量运行时页面，不替代 admin-next 的完整后台能力。
- super_admin 仍建议使用 admin-next 和 admin API 进行全站治理。
- 完整 Playwright 浏览器 Console 验收留到后续专门验收任务。

## v1.1.1 前后台身份边界整理范围

已完成能力：

- 身份模型：`users` 负责社区行为，`admin_users` 负责后台管理，`community_moderators` 把部分 `users` 授权为指定子站版主。
- Token 边界：JWT 增加 `token_type` 和 `aud`；前台 user token 使用 `token_type=user,aud=devhub_frontend`，后台 admin token 使用 `token_type=admin,aud=devhub_admin`。
- 登录入口：`POST /api/v1/auth/login` 调用前台 `users` 登录；`POST /api/v1/admin/login` 调用后台 `admin_users` 登录。
- Middleware：前台写操作使用 user auth；后台接口使用 admin auth，并允许已启用的子站版主 user token 进入有限治理接口。
- 版主 scope：版主只能按 `community_moderators.user_id + community_id + status=1` 管理自己负责子站；不能管理后台人员、全局配置、版主分配或其他子站。
- 审计 actor：`admin_logs` 记录并返回 `actor_type` / `actor_id`，用于区分后台人员、子站版主和系统操作。
- 前端存储：前台推荐 `devhub_user_token` / `devhub_user_refresh_token`，兼容读取旧 `devhub_access_token` / `devhub_refresh_token`；后台继续独立使用 `devhub_admin_token` / `devhub_admin_refresh_token`。

已知限制：

- MemoryStore 的 demo 用户数据仍复用轻量内存结构，文档中按开发兜底处理，不作为生产身份规则。
- MySQL 现有 `refresh_tokens.user_id` 仍复用于 user/admin refresh token，依靠 `token_type` 区分，v1.1.1 已移除单一 `users` 外键；后续生产化 migration 可进一步拆分字段名。
- 后台人员如果需要参与前台社区互动，仍应拥有独立 `users` 身份；本轮未做 admin-user 绑定关系。
- 子站版主仍可被管理员纳入 `/api/v1/admin/*` 的受限治理 scope；v1.1.3 已新增独立 `/moderator` 工作台作为版主优先入口。

## v1.1.0 子站模块增强范围

已完成能力：

- 子站增强字段：logo、cover image、slogan、theme color、SEO 字段、sort_order、status、follower_count、topic_count、comment_count、hot_score 和公告字段。
- 子站首页：`/c/:slug/` 展示子站名称、slogan、简介、统计、关注按钮、发帖入口、板块导航、置顶、精华、最新、热门、未解决问答、热门标签、版主和公告。
- 兼容入口：`/site/:slug` 301 跳转到 `/c/:slug/`。
- 子站 SEO：`/c/:slug/` 由 Go 动态输出 title、description、canonical、h1、内容链接、标签链接和板块链接。
- 子站 API：`GET /api/v1/communities/:slug/stats`、`GET /api/v1/communities/:slug/moderators` 已可用。
- 子站后台：`/admin-next/communities` 提供子站新增、编辑、启用 / 禁用、排序、SEO、公告、前台跳转和版主跳转。
- 子站板块管理：后台抽屉支持子站板块查看、新增、编辑、启用 / 禁用、排序、导航展示、发帖开关和 SEO 字段。
- 子站关注：复用 `POST /api/v1/follows/toggle`，`target_type=community` 更新 follower_count，写入 `activities.action=followed`，并在我的关注中展示。
- sitemap：包含启用子站 `/c/php/`、`/c/go/`、`/c/java/`、`/c/ai/`、`/c/frontend/`，不包含禁用或归档子站。
- MemoryStore / MySQLStore：均支持子站增强字段、统计、公开版主、后台子站 CRUD、后台板块 CRUD、子站关注计数和 sitemap 过滤。
- 审计：新增 / 编辑 / 启用 / 禁用 / 排序子站，以及新增 / 编辑 / 启用 / 禁用 / 排序子站板块，均写入 `admin_logs`。

已知限制：

- v1.1.0 使用启用板块生成默认子站导航，深度自定义导航配置留到后续版本。
- 标签趋势统计仍留到后续版本；标签后台管理已在 v1.2.0 完成，标签合并 / 别名 / 统计重算已在 v1.2.1 完成。
- 完整关注流留到 v1.3.0；本版本完成关注状态、follower_count、动态和我的关注展示。
- 评论点赞、取消已解决状态、推荐算法、声望积分和复杂运营分析不属于本版本。
- sitemap 仍是单文件动态输出，内容规模扩大后需要分片。

## v1.0.0 归档范围

已完成能力：

- 多子站、板块、Topic 模型、发布、列表、详情、编辑、删除。
- 搜索、筛选、标签基础能力、热门标签和子站标签合集。
- 点赞、收藏、关注、动态、通知、用户中心页面。
- 评论、回复、问答采纳、最佳答案展示、未解决筛选。
- 举报、版主范围权限、精华、置顶、隐藏、恢复、评论锁定、评论隐藏。
- admin-next 内容 CRUD、评论管理、举报管理、版主管理、批量治理、治理审计日志。
- `/topics/:id` Go 动态百度 SEO HTML、动态 sitemap、robots。
- MemoryStore / MySQLStore 双模式。
- 测试矩阵、部署文档、备份回滚文档、CI 和 v1.0.0 release notes。

已知限制：

- 标签详情页、标签 SEO 聚合页和标签后台管理已在 v1.2.0 完成；标签合并 / 别名 / 统计重算已在 v1.2.1 完成，趋势统计仍待后续。
- 评论点赞未实现到运行时评论区。
- 问答支持采纳和更换最佳答案，暂不支持取消已解决状态。
- 标签关注已在 v1.2.0 接入标签页；用户关注前台作者入口仍可继续增强。
- `/sitemap.xml` 当前未做大规模分片。
- migration 仍以基础 SQL 和少量迁移脚本为主，生产升级需要预发演练。
- 生产部署仍需根据真实环境配置守护进程、反向代理、HTTPS、日志和定时备份。

后续建议规划：

- v1.2.0：标签专项增强、标签 SEO 聚合页和标签后台管理已完成。
- v1.2.1：标签治理增强、标签合并 / 别名 / 统计重算已完成。
- v1.3.0：推荐、关注流和内容发现。
- v1.4.0：用户成长、声望和个人主页。
- v1.5.0：后台运营、治理和数据统计增强。
- v1.6.0：生产化、migration、性能和 CI/CD。

## 第五轮互动联动状态

按 2026-05-09 当前代码，状态如下：

- 点赞 / 取消点赞：已完成。真实接口为 `POST /api/v1/topics/:id/like`，也保留 `POST /api/v1/reactions/toggle`。第一次调用写入 `reactions`，再次调用删除；返回 `liked`、`like_count`、`hot_score`。
- 收藏 / 取消收藏：已完成。真实接口为 `POST /api/v1/topics/:id/favorite`，兼容 `POST /api/v1/favorites/toggle`。第一次调用写入 `favorites`，再次调用删除；返回 `favorited`、`favorite_count`、`hot_score`。
- 关注 / 取消关注：已完成。真实接口为 `POST /api/v1/follows/toggle`，支持 `user`、`community`、`tag`、`topic`；再次调用取消关注。
- 我的收藏页面：已完成。前台路由 `/me/favorites`，接口 `GET /api/v1/me/favorites`。
- 我的关注页面：已完成。前台路由 `/me/follows`，接口 `GET /api/v1/me/follows`。
- 我的动态页面：已完成。前台路由 `/me/activities`，接口 `GET /api/v1/me/activities`；兼容 `GET /api/v1/activities`。
- 通知页面：已完成。前台路由 `/notifications`，`/me/notifications` 为同一静态壳别名；接口 `GET /api/v1/me/notifications`。
- 通知单条已读：已完成。真实接口为 `POST /api/v1/me/notifications/:id/read`；兼容旧 `POST /api/v1/notifications/:id/read`。
- 通知全部已读：已完成。真实接口为 `POST /api/v1/me/notifications/read-all`；兼容旧 `POST /api/v1/notifications/read-all`。
- `hot_score` 更新逻辑：已完成基础公式。Topic 互动后使用 `view_count + comment_count*5 + like_count*3 + favorite_count*4` 重新计算；MemoryStore 运行时计算，MySQLStore 写回 `topics.hot_score`。
- MemoryStore 支持情况：已完成本轮基础能力。支持点赞、取消点赞、收藏、取消收藏、关注、取消关注、我的收藏、我的关注、我的动态、通知列表、单条已读、全部已读；内存模式不持久化，当前进程内状态正确。
- MySQLStore 支持情况：已完成本轮基础能力。支持 `reactions`、`favorites`、`follows`、`activities`、`notifications` 读写；toggle 避免重复计数；取消操作使用 `GREATEST(...,0)` 避免负数。
- 是否写入 `activities`：已完成。发布写入 `created_topic`；点赞写入 `liked`；收藏写入 `favorited`；关注写入 `followed`。
- 是否写入 `notifications`：已完成基础规则。点赞、收藏会给 Topic 作者生成通知；关注用户会给被关注用户生成通知；评论 Topic 生成 `topic_commented`，回复评论生成 `comment_replied`，回答被采纳生成 `answer_accepted`；自己操作自己的内容不通知自己。
- 是否保持 `/topics/:id` 百度 SEO HTML 不被破坏：已保持。详情页仍由 Go 动态输出 title、meta description、h1、正文、标签链接、发布时间和 Article JSON-LD；点赞、收藏、关注、评论作为运行时增强。
- 第七轮已补齐：举报、版主范围治理、精华、置顶、隐藏和评论锁定。
- 尚未完成的问题：评论点赞本轮未实现；采纳支持更换最佳答案但暂不支持取消已解决状态；复杂权限继续沿用当前轻量 RBAC 和 demo user 方案。

## 第六轮评论系统和问答采纳状态

按 2026-05-09 当前代码，状态如下：

- 评论列表：已完成。真实接口为 `GET /api/v1/topics/:id/comments`，支持 `page`、`page_size`、`sort=best/latest/oldest`；详情页评论区支持加载更多。
- 发表评论：已完成。真实接口为 `POST /api/v1/topics/:id/comments`，请求字段为 `content`，长度限制为 2 到 5000 个字符。
- 回复评论：已完成。真实接口为 `POST /api/v1/topics/:id/comments/:commentId/replies`，会校验父评论必须属于当前 Topic。
- 评论统计联动：已完成。评论和回复都会更新 `comment_count`、`last_active_at`、`updated_at`，并按统一公式刷新 `hot_score`。
- 评论动态：已完成。评论 Topic 写入 `activities.action=commented,target_type=topic`；回复写入 `activities.action=commented,target_type=comment`，`target_id` 指向被回复评论。
- 评论通知：已完成。非作者评论 Topic 生成 `topic_commented`；非本人回复评论生成 `comment_replied`；自己评论 / 回复自己不通知。
- 问答状态：已完成。`topics.content_type=question` 使用 `is_solved` 和 `best_comment_id` 表示已解决状态。
- 采纳最佳答案：已完成。真实接口为 `POST /api/v1/topics/:id/comments/:commentId/accept`，兼容 `POST /api/v1/topics/:id/solve`。
- 更换最佳答案：已完成。新采纳评论 `is_best=true`，同 Topic 其他评论 `is_best=false`；暂不支持取消已解决状态。
- 采纳动态和通知：已完成。采纳写入 `accepted_answer` 动态；使用实际操作者作为 actor；非本人时生成 `answer_accepted` 通知。
- 未解决筛选：已完成。`GET /api/v1/search/topics?sort=unsolved` 和 `GET /api/v1/topics?sort=unsolved` 只返回 `content_type=question AND is_solved=0`。
- MemoryStore 支持情况：已完成第六轮基础能力，当前进程内支持评论、回复、采纳、未解决筛选、动态和通知。
- MySQLStore 支持情况：已完成第六轮基础能力，支持 `comments` 表读写、统计更新、动态、通知、采纳状态和未解决筛选；评论创建、统计更新、活动和通知尽量在事务内完成，采纳最佳答案同样使用事务更新 Topic、评论、动态和通知。
- SEO 是否保持不被破坏：已保持。`/topics/:id` 仍由 Go 动态输出 SEO HTML；评论区和最佳答案为运行时增强，当前不强制进入初始 SEO HTML。
- 尚未完成事项：评论点赞未纳入本轮；取消已解决状态未实现；评论 anchor 跳转当前可后续优化。

## 第七轮举报和社区治理状态

按 2026-05-09 当前代码，状态如下：

- 用户举报：已完成。真实接口为 `POST /api/v1/reports`，支持举报 `topic`、`comment`、`user`、`wiki`；本轮前台在 Topic 详情页和评论区提供举报入口。
- 举报管理：已完成最小后台。真实接口为 `GET /api/v1/admin/reports`、`GET /api/v1/admin/reports/:id`、`POST /api/v1/admin/reports/:id/handle`；`accepted` 会联动隐藏 topic 或 comment，`rejected` 只更新举报状态。
- 版主权限：已完成最小可用判断。`currentUserID` 优先使用登录用户，未登录兜底 demo `user_id=1`；后台管理员 user 1 / `super_admin` 可管理全部，`community_moderators` 中启用用户只能管理对应子站。seed 中 user 2 是 PHP 版主，user 3 是 Go 版主。
- Topic 精华：已完成。`POST /api/v1/admin/topics/:id/feature` toggle `topics.is_featured`，前台列表和详情数据展示“精华”，`sort=featured` 只返回精华内容。
- Topic 置顶：已完成。`POST /api/v1/admin/topics/:id/pin` toggle `topics.is_pinned`，普通列表默认置顶优先展示，前台显示“置顶”。
- Topic 隐藏 / 恢复：已完成。`POST /api/v1/admin/topics/:id/hide` 设置 `status=0`，`restore` 设置 `status=1`；普通列表、搜索和 sitemap 过滤隐藏 topic；后台内容管理仍可看到隐藏内容。
- 评论锁定：已完成。`POST /api/v1/admin/topics/:id/lock-comments` 和 `unlock-comments` 更新 `comment_locked`；前台详情页显示“评论已锁定”，后端创建评论和回复时强制拦截。
- Comment 隐藏 / 恢复：已完成。`POST /api/v1/admin/comments/:id/hide` 设置隐藏，`restore` 恢复正常；普通评论列表过滤隐藏评论。隐藏最佳答案当前被禁止，避免破坏问答采纳闭环。
- 举报处理联动：已完成。接受 topic 举报会隐藏 topic；接受 comment 举报会隐藏评论；如果目标评论是最佳答案，处理会返回错误并保持原状态。
- 后台入口：已完成最小入口。`/admin-next/reports` 提供举报列表、筛选和接受 / 驳回操作；内容管理提供精华、置顶、隐藏、锁定评论操作；评论管理提供隐藏 / 恢复评论操作。
- MemoryStore 支持情况：已完成创建举报、查询举报、处理举报、版主判断、精华、置顶、隐藏 / 恢复 topic、锁定 / 解锁评论、隐藏 / 恢复 comment、普通列表过滤隐藏内容和评论。
- MySQLStore 支持情况：已完成 `reports`、`community_moderators`、`topics.comment_locked` 等 schema 和迁移补齐；支持举报分页、处理、权限判断、topic/comment 治理字段更新、普通列表和搜索过滤隐藏内容，兼容 MySQL 8。
- SEO 保护：已保持。正常 `/topics/:id` 仍由 Go 动态输出 title、description、h1、正文、标签和发布时间；隐藏 topic 返回带 `noindex,follow` 的“内容已隐藏”动态 HTML，不输出原正文；`/sitemap.xml` 不包含隐藏 topic。
- 尚未完成事项：第七轮隐藏最佳答案采用禁止策略；复杂登录 / 权限系统留到后续。

## 第八轮 admin-next 后台 CRUD、版主管理、批量治理和治理审计状态

按 2026-05-09 当前代码，状态如下：

- 后台内容 CRUD：已完成。`GET/POST/PUT/DELETE /api/v1/admin/posts` 路径保持兼容命名，但后台真实读写 `topics`；公开 `sites/posts` 兼容 API 未删除。
- 版主管理 CRUD：已完成。真实接口为 `GET/POST/PUT/DELETE /api/v1/admin/moderators`；删除采用停用策略，将 `community_moderators.status` 置为 `0`。
- 版主权限：已保持轻量实现。`super_admin` 可管理全部；启用版主可按 `community_id` 范围查看和处理自己子站的举报与内容。版主分配目前要求管理员操作。
- 批量 Topic 治理：已完成。`POST /api/v1/admin/topics/batch` 支持 `feature/unfeature/pin/unpin/hide/restore/lock-comments/unlock-comments/delete`。
- 批量 Comment 治理：已完成。`POST /api/v1/admin/comments/batch` 支持 `hide/restore/delete`；隐藏最佳答案仍然禁止。
- 批量举报处理：已完成。`POST /api/v1/admin/reports/batch-handle` 支持批量 `accepted/rejected`，逐条返回成功和失败原因。
- 举报频率限制：已完成最小限制。同一用户对同一对象已有 `pending` 举报时，新举报返回 `同一对象已有待处理举报，请勿重复提交`；已处理后可再次举报。
- 治理审计日志：已完成。新增 `GET /api/v1/admin/audit-logs`，支持 `site/type/action/target/target_type/actor/actor_user_id/community_id/page/page_size`；新增/更新版主、批量治理、举报处理、内容 CRUD 都会写入 `admin_logs`。当前表结构保存 `actor/action/target/site` 文本字段，接口返回的 `actor_user_id`、`target_type`、`target_id`、`community_id` 为派生字段。
- admin-next 页面：已完成。内容管理支持新增、编辑、删除和批量治理；评论管理支持勾选批量处理；举报管理支持批量接受 / 驳回；`/admin-next/moderators` 提供版主管理；`/admin-next/audit-logs` 提供独立治理审计入口；系统设置页仍保留基础日志区。
- MemoryStore 支持情况：已完成第八轮能力。支持版主 CRUD、后台内容 CRUD、批量治理、举报 pending 去重、审计日志筛选分页。
- MySQLStore 支持情况：已完成第八轮能力。支持 `community_moderators` CRUD、`topics` 后台 CRUD、`reports` pending 去重索引、批量治理依赖的单项更新和 `admin_logs` 筛选分页，兼容 MySQL 8。
- SEO 保护：已保持。第八轮未改变 `/topics/:id` Go 动态 SEO HTML；隐藏内容仍不进入普通列表、搜索和 sitemap。
- 第八轮补丁收口：已完成 `admin.js` 版主管理 / 批量治理 / 审计日志封装对账；`Content.vue`、`Comments.vue`、`Reports.vue` 已接入真实批量接口并支持备注；后台 Vite dev proxy、`main.go` 默认端口和文档口径均统一为 `8090`。
- 尚未完成事项：复杂登录系统、细粒度版主分配审批、评论点赞、取消已解决状态不在本轮主线。标签关注已在 v1.2.0 标签页接入；用户关注前台入口仍作为后续增强。

## 第六轮收尾验收记录

2026-05-09 已完成第六轮收尾验收：

- 启动方式：清理残留进程后，使用 `go build -o .devhub/devhub .` 产出二进制，并以 `PORT=8090 CMS_STORE=memory` 通过 `setsid` 后台启动。
- 8090 状态：`curl -I /`、`GET /api/v1/health`、`GET /api/v1/topics` 均返回 200，当前 8090 有稳定 DevHub 进程监听。
- Gitee 卡住问题：仓库依赖未发现 Gitee 私有模块；已清理残留 `git-upload-pack` 子进程，最终避开 `go run`，使用二进制启动完成验收。
- Go 测试：本轮已执行 `go test ./...` 并通过。
- 前端构建：`./dev.sh --local-go --restart` 已完成 Astro 前台和 Vue 后台构建；本机无 `npm` 时由 Docker Node 构建。
- 评论验收：`GET /api/v1/topics/1/comments` 返回正常；`POST /api/v1/topics/1/comments` 新增评论 ID `9`；`POST /api/v1/topics/1/comments/9/replies` 新增回复 ID `10`；`comment_count`、`last_active_at`、`hot_score` 均更新。
- 采纳验收：topic 1 是 article，采纳返回 400；topic 2 是 question，新增回答 ID `11` 后采纳成功，`is_solved=true`，`best_comment_id=11`，评论 `is_best=true`。
- 未解决筛选：采纳 topic 2 后，`GET /api/v1/search/topics?sort=unsolved` 不再返回 topic 2，只返回未解决 question。
- 动态和通知：`GET /api/v1/me/activities` 返回 `commented` 与 `accepted_answer`；`GET /api/v1/me/notifications` 返回正常。memory 模式当前认证统一为 demo/user 1，本次自操作未新增自通知，符合“不通知自己”规则；非本人通知路径已在 Store 实现中保留。
- SEO 回归：`/topics/1` 源码保留 `<title>`、`meta description`、`<h1>`、`<article>`、正文和标签链接；`/sitemap.xml`、`/robots.txt` 均返回 200。
- 页面回归：`/`、`/search?sort=unsolved`、`/topics/1`、`/me/activities`、`/notifications`、`/admin-next`、`/topics/new`、`/c/php` 均返回 200。

## 第九轮 v1.0.0 归档验收记录

2026-05-09 已完成 v1.0.0 归档验收：

- 第八轮补丁收口：已确认 `main.go` 和 `dev.sh` 默认端口统一为 `8090`，`admin.js` 已包含版主管理、批量治理、批量举报和审计日志封装，`/admin-next/moderators` 与 `/admin-next/audit-logs` 已有真实路由和页面。
- 版本文件：新增 `VERSION=v1.0.0`、`CHANGELOG.md` 和 `docs/releases/v1.0.0.md`。
- CI：新增 `.github/workflows/ci.yml`，覆盖 Go test、Go build、前台构建、后台构建、schema 和核心文档存在性检查。
- 文档：补齐 v1.0.0 测试矩阵、部署说明、备份回滚、SEO 归档、API 模块归档和项目进度归档。
- Go 测试：`go test ./...` 通过。
- Go 构建：`go build -o .devhub/devhub .` 通过。
- 前台构建：Docker Node 执行 Astro build 通过。
- 后台构建：Docker Node 执行 Vite build 通过，仅有 chunk size warning。
- 启动方式：已执行 `./dev.sh --local-go restart --no-build`；最终使用 `PORT=8090 CMS_STORE=memory ./.devhub/devhub` 二进制后台常驻，避免 `go run` 临时进程链路。
- URL 验收：`/`、`/admin-next`、`/api/v1/communities`、`/api/v1/topics`、`/api/v1/search/topics?keyword=go`、`/topics/1`、`/sitemap.xml`、`/robots.txt` 均返回 200。
- 后台验收：`/admin-next/content`、`/admin-next/comments`、`/admin-next/reports`、`/admin-next/moderators`、`/admin-next/audit-logs`、`/admin-next/sites`、`/admin-next/users`、`/admin-next/system` 均返回 200；后台 reports、moderators、audit-logs API 均返回 200。
- 批量治理验收：Topic 批量 `feature/unfeature`、Comment 批量 `hide/restore`、Report 批量 `rejected` 均返回 `updated=1, failed=0`。
- SEO 回归：`/topics/1` 源码保留 `<title>`、`meta description`、`<h1>`、`<article>`、正文、标签链接、发布时间和 Article JSON-LD；隐藏 Topic 不进入 sitemap，恢复后重新进入。

## v1.1.0 子站增强验收记录

2026-05-09 已完成 v1.1.0 子站增强实现：

- 版本文件：`VERSION` 更新为 `v1.1.0`，`CHANGELOG.md` 新增 v1.1.0，新增 `docs/releases/v1.1.0.md`。
- 数据模型：`communities` 增强 logo、cover、slogan、theme、SEO、统计和公告字段；`categories` 增强导航展示、发帖开关、SEO 和状态字段。
- 前台：`/c/:slug/` 由 Go 动态输出子站首页和 SEO HTML；`/site/:slug` 301 到 `/c/:slug/`。
- API：新增 `GET /api/v1/communities/:slug/stats`、`GET /api/v1/communities/:slug/moderators` 和后台 communities/categories CRUD。
- 后台：新增 `/admin-next/communities`，`/admin-next/sites` 保持兼容并指向同一页面；页面支持子站配置和板块管理。
- sitemap：启用子站进入 `/sitemap.xml`；禁用 / 归档子站不进入 sitemap。
- SEO 保护：`/topics/:id` Go 动态 SEO HTML 未重构，继续保留 title、description、h1、正文和标签链接。
- 测试状态：`bash -n dev.sh`、`go test ./...`、`go build -o .devhub/devhub .`、前台 Docker Node 构建、后台 Docker Node 构建、`./dev.sh --restart`、二进制后台常驻启动和关键 URL / SEO / sitemap / 后台入口回归均已完成；实测结果记录在 `docs/TESTING.md`。

## 已完成

### 项目结构

- 技术栈：Go + Gin、Astro + Vue Islands、Vue 3 + Element Plus、MySQL 8。
- 后端按 `domain / store / service / transport/httpapi` 拆分。
- 保留 `MemoryStore` 和 `MySQLStore` 两种数据仓库。
- 前台入口 `/`，后台入口 `/admin-next`；`/admin` 和 `/admin/:site` 只做兼容重定向。

### 启动脚本

- `dev.sh` 默认端口固定为 `8090`。
- 支持内存模式和 MySQL 模式。
- 无本机 `npm` 时可使用 Docker Node 构建前后台。
- 修改 Go 代码后应使用 `./dev.sh --restart` 或 `./dev.sh restart --no-build` 重启服务。

### 数据模型

当前 Go 内置 schema 和 `db/mysql/001_schema.sql` 包含：

- 兼容模型：`sites`、`boards`、`posts`、`comments`、`tags`、`notifications`。
- 用户和权限模型：`users`、`roles`、`permissions`、`role_permissions`、`user_roles`、`refresh_tokens`。
- 通用社区模型：`communities`、`categories`、`topics`、`topic_tags`、`reactions`、`favorites`、`follows`、`activities`、`reports`、`community_moderators`、`wiki_pages`、`wiki_revisions`。
- 已扩展字段：`comments.topic_id/reply_to_user_id/user_id/content_html/is_best/updated_at/deleted_at`，`topics.is_solved/best_comment_id/comment_count/last_active_at`，`notifications.actor_user_id/type/target_type/target_id/topic_id/comment_id/read_at`，`activities.topic_id/metadata`，`reactions/favorites/follows.updated_at`。

### 后端 API

第五轮新增或增强：

- `POST /api/v1/topics/:id/like`
- `POST /api/v1/topics/:id/favorite`
- `GET /api/v1/topics/:id/interaction`
- `POST /api/v1/follows/toggle`
- `GET /api/v1/me/favorites`
- `GET /api/v1/me/follows`
- `GET /api/v1/me/activities`
- `GET /api/v1/me/notifications`
- `POST /api/v1/me/notifications/:id/read`
- `POST /api/v1/me/notifications/read-all`
- `GET /api/v1/topics/:id/comments`
- `POST /api/v1/topics/:id/comments`
- `POST /api/v1/topics/:id/comments/:commentId/replies`
- `POST /api/v1/topics/:id/comments/:commentId/accept`
- `GET /api/v1/search/topics?sort=unsolved`
- `POST /api/v1/reports`
- `GET /api/v1/admin/reports`
- `GET /api/v1/admin/reports/:id`
- `POST /api/v1/admin/reports/:id/handle`
- `POST /api/v1/admin/topics/:id/feature`
- `POST /api/v1/admin/topics/:id/pin`
- `POST /api/v1/admin/topics/:id/hide`
- `POST /api/v1/admin/topics/:id/restore`
- `POST /api/v1/admin/topics/:id/lock-comments`
- `POST /api/v1/admin/topics/:id/unlock-comments`
- `POST /api/v1/admin/comments/:id/hide`
- `POST /api/v1/admin/comments/:id/restore`

保留兼容接口：

- `POST /api/v1/reactions/toggle`
- `POST /api/v1/favorites/toggle`
- `GET /api/v1/activities`
- `GET /api/v1/notifications`
- `GET /api/v1/notifications/unread-count`
- `POST /api/v1/notifications/:id/read`
- `POST /api/v1/notifications/read-all`
- `sites/posts` 兼容 API 未删除。

### 前台

- 首页、子站页、搜索页、发布页、标签页可用。
- `/topics/:id` 由 Go 动态输出 SEO HTML；交互按钮通过运行时 API 增强。
- 新增 `/me/favorites`、`/me/follows`、`/me/activities`、`/notifications` 页面。
- `/me/notifications` 作为通知中心静态壳别名由 Go 路由 fallback 到同一产物。
- 导航用户菜单新增我的动态、我的收藏、我的关注、通知中心入口。
- 子站页新增关注子站按钮。
- 详情页点赞、收藏、关注主题按钮初始化状态正确，点击后更新状态和计数。
- 详情页评论区支持运行时加载评论、加载更多、发表评论、回复评论；question 类型支持采纳按钮并展示“最佳答案”，非作者默认不显示采纳入口。
- 详情页支持举报 Topic 和举报评论；评论锁定时显示锁定提示并禁用普通提交入口。
- 列表页和搜索页展示置顶、精华、问答状态，隐藏内容不进入普通列表。

### 后台

- `/admin-next` 和 `/admin-next/...` 深层路由可用。
- `/admin`、`/admin/:site` 保持兼容重定向。
- `/admin-next/reports` 提供举报管理入口，支持举报筛选、列表、单条接受 / 驳回和批量接受 / 驳回。
- `/admin-next/content` 提供内容 CRUD、精华、置顶、隐藏 / 恢复、锁定 / 解锁评论和批量治理入口。
- `/admin-next/comments` 提供评论隐藏 / 恢复 / 删除和批量治理入口。
- `/admin-next/moderators` 提供版主管理 CRUD，删除为停用。
- `/admin-next/audit-logs` 提供治理审计列表和筛选。

## 当前仍未完成 / 风险 / 下一步

已完成：

- v1.3.0 插件全局状态、子站状态、发布校验、板块绑定、前台发布页和版主插件菜单已进入可运行闭环。
- v1.2.1 标签别名、标签合并、统计重算、sitemap / SEO 治理和后台治理审计已完成。
- v1.2.0 标签 SEO 页、标签关注、发布页标签建议和后台标签基础管理已完成。

部分完成：

- 评论点赞：数据字段和旧评论点赞接口保留，详情页运行时评论区尚未接入完整评论点赞体验。
- 问答采纳：支持采纳和更换最佳答案，暂不支持取消已解决状态。
- 用户关注：后端支持 `target_type=user` 并触发 `user_followed` 通知，前台作者信息区域的关注入口仍需继续增强。
- 子站插件管理：API 和基础启用/禁用入口可用，`config_json` 可视化编辑与排序 UI 仍需补齐。
- Docs / Wiki 插件：基础注册、表结构和通用内容发布可用，专用文档树、版本历史和回滚交互仍需后续专项。

未完成：

- 发布链路的插件权限码细粒度校验，例如 `qa.question.create`、`docs.document.create`、`wiki.page.create`。
- 标签趋势统计、标签运营分析和大规模异步统计任务。
- 复杂 RBAC、版主任期 / 绩效统计、推荐、关注流、admin-user 与 frontend-user 绑定关系和生产化 migration。

风险：

- 如果后端代码已修改但 `8090` 上已有服务正在运行，普通 `./dev.sh` 可能复用旧服务；应使用 `./dev.sh --restart`。
- Docker Go / Docker Node 构建依赖本机 Docker 权限和网络。
- 历史数据中 `categories.plugin_code/allowed_content_types` 或 `community_plugins` 可能需要生产升级前抽样校验。
- `/sitemap.xml` 当前最多输出 5000 条 Topic；内容量上来后需要拆分 sitemap index。

下一步：

1. 补齐子站插件配置 UI 的 `config_json` 编辑和排序操作。
2. 将发布权限码检查并入统一插件校验，避免后台或前台绕过插件权限。
3. 按 `docs/TESTING.md` 补跑 v1.3.0 子站插件禁用、跨子站发布、迁移和 SEO 验收。

## 验收清单

- [x] 项目名称保持 DevHub。
- [x] 默认端口固定为 `8090`。
- [x] `main.go` 直接运行默认端口已统一为 `8090`，`PORT` 仍可覆盖。
- [x] 前台 `/` 可打开。
- [x] 后台 `/admin-next` 可打开。
- [x] `/topics/:id` 由 Go 动态输出 SEO HTML。
- [x] 新发布 Topic 后无需重新 build 即可访问 `/topics/:id`。
- [x] `sites/posts` 兼容 API 未删除。
- [x] `POST /api/v1/topics/:id/like` 返回 `liked/like_count/hot_score`。
- [x] `POST /api/v1/topics/:id/favorite` 返回 `favorited/favorite_count/hot_score`。
- [x] `POST /api/v1/follows/toggle` 返回 `followed/target_type/target_id`。
- [x] `/me/favorites` 页面和 `GET /api/v1/me/favorites` 可用。
- [x] `/me/follows` 页面和 `GET /api/v1/me/follows` 可用。
- [x] `/me/activities` 页面和 `GET /api/v1/me/activities` 可用。
- [x] `/notifications` 页面和 `GET /api/v1/me/notifications` 可用。
- [x] 通知单条已读、全部已读可用。
- [x] 点赞、收藏、关注写入动态。
- [x] 点赞、收藏、用户关注触发基础通知。
- [x] MemoryStore 支持本轮互动闭环。
- [x] MySQLStore 支持本轮互动闭环。
- [x] SEO 文档已记录百度 SEO 保护要求。
- [x] `GET /api/v1/topics/:id/comments` 已记录第六轮评论列表。
- [x] `POST /api/v1/topics/:id/comments` 已记录发表评论。
- [x] `POST /api/v1/topics/:id/comments/:commentId/replies` 已记录回复评论。
- [x] `POST /api/v1/topics/:id/comments/:commentId/accept` 已记录采纳最佳答案。
- [x] 未解决筛选 `sort=unsolved` 可用。
- [x] 评论和采纳写入动态与通知。
- [x] `/topics/:id` 评论功能未破坏百度 SEO HTML。
- [x] `POST /api/v1/reports` 可创建举报。
- [x] `/admin-next/reports` 可查看和处理举报。
- [x] `/admin-next/moderators` 可查看、新增、编辑和停用版主。
- [x] `/admin-next/content` 可执行批量 Topic 治理。
- [x] `/admin-next/comments` 可执行批量 Comment 治理。
- [x] `/admin-next/audit-logs` 可查看治理审计日志。
- [x] 管理员可处理全部举报，版主仅可处理自己子站举报。
- [x] Topic 精华、置顶、隐藏、恢复、评论锁定、解锁可用。
- [x] Comment 隐藏、恢复可用，隐藏最佳答案会被拒绝。
- [x] 隐藏 Topic 不进入普通列表、搜索和 sitemap。
- [x] 隐藏 Topic 详情页返回 noindex 的“内容已隐藏”动态 HTML。
- [x] `VERSION` 已记录 `v1.2.0`。
- [x] `CHANGELOG.md` 已记录 v1.1.0 主要变化和限制。
- [x] `CHANGELOG.md` 已记录 v1.1.1 身份边界整理。
- [x] `CHANGELOG.md` 已记录 v1.1.3 独立版主工作台 MVP。
- [x] `CHANGELOG.md` 已记录 v1.2.0 标签系统增强。
- [x] `docs/releases/v1.1.0.md` 已记录版本定位、数据结构、API、SEO、后台、测试、限制和 tag 建议。
- [x] `docs/releases/v1.1.1.md` 已记录前后台身份边界整理。
- [x] `docs/releases/v1.1.3.md` 已记录独立版主工作台 MVP。
- [x] `docs/releases/v1.2.0.md` 已记录标签系统增强。
- [x] `docs/BACKUP_AND_ROLLBACK.md` 已记录备份、恢复和紧急回滚流程。
- [x] `.github/workflows/ci.yml` 已补充 Go / 前台 / 后台基础 CI。
