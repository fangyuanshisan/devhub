# DevHub 测试文档

[返回文档入口](README.md)

更新时间：2026-05-13

本文档只记录当前 v1.4.0 必测项和后续补测项。历史版本测试只保留必要回归，不再展开旧版本完整矩阵。

## v1.4.0 收口验收（2026-05-13）

本节记录 v1.4.0 插件平台增强（依赖治理、错误码/Readiness、前台入口治理）合并后的收口验收结果，作为当前仓库“已真实跑过的检查”口径来源。

已执行并通过：

- `gofmt -w $(git ls-files '*.go')`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `./scripts/check-frontend.sh --admin-only`（后台 build + Playwright：`35 passed`，包含新增 `plugin-pages-navigation.spec.js`）
- `./scripts/check-frontend.sh --frontend-only`（前台 build + Playwright：`17 passed`）
- SEO curl 回归（在本地 8090 服务可用情况下执行）：
  - `curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|canonical|<h1|<article|application/ld\\+json'`
  - `curl -s http://127.0.0.1:8090/c/php/ | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'`

结论：

- Playwright 不再保留 `test.skip` / `test.only`；收口验收未发现长期跳过项。

## v1.4.0-P1-13 收口补充（2026-05-13）

本节记录插件后台信息架构与按功能分页优化后的后台验收结果。

已执行并通过：

- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-pages-navigation.spec.js`：通过，`2 passed`。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-dependencies.spec.js`：通过，`2 passed`。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-governance.spec.js`：通过，`13 passed`。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 Playwright `35 passed`。

说明：

- 本轮只改后台插件板块信息架构与管理页路由，没有触碰前台导航或 SEO 共享逻辑，因此未额外执行 `--frontend-only` 与 SEO curl。
- 旧路由 `/admin-next/plugins`、`/admin-next/plugins/governance`、`/admin-next/qa`、`/admin-next/docs`、`/admin-next/wiki`、`/admin-next/projects`、`/admin-next/jobs`、`/admin-next/ai-works` 均保持兼容。

## v1.5.0-P0-01 插件包 dry-run（2026-05-13）

本节记录“本地插件包规范草案 + dry-run 导入预览”的最小验收命令口径。

必须执行：

- `gofmt -w $(git ls-files '*.go')`
- `go test ./...`
- `go build -o .devhub/devhub .`
- `git diff --check`
- `bash -n dev.sh`
- `bash -n scripts/check-frontend.sh`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-package-dryrun.spec.js`
- `./scripts/check-frontend.sh --admin-only`

说明：

- 本轮只涉及后端与后台插件板块安装升级页，不涉及前台与 SEO，因此不强制执行 `--frontend-only` 与 SEO curl（如后续改动触及前台再补）。

## v1.5.0-P0-02 插件包安全校验与风险报告（2026-05-13）

本节记录“checksums.json（sha256）校验 + 危险规则强化 + risk_report 风险报告”的最小验收命令与结果。

已执行（通过）：

- `gofmt -w internal/domain/plugin_package.go internal/service/plugin_package_service.go internal/plugins/*.go internal/service/plugin_package_dryrun_test.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-package-security.spec.js`：通过，`4 passed`。
- `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 ./scripts/check-frontend.sh --admin-only`：通过，后台 Playwright `42 passed`。

说明：

- 本机 `8090` 端口被占用时，可用 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:<port>` 覆盖 Playwright baseURL；本轮在本机启动 `PORT=8091 CMS_STORE=memory DEVHUB_E2E_TESTING=1 ./.devhub/devhub` 后执行上述 E2E。
- 本轮只改后端与后台 dry-run 页面，不涉及前台与 SEO，因此未执行 `--frontend-only` 与 SEO curl。

## v1.5.0-P0-03 本地插件仓库目录与扫描（2026-05-13）

已执行（通过）：

- `gofmt -w $(git ls-files '*.go')`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-package-repository.spec.js`：通过。
- `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 ./scripts/check-frontend.sh --admin-only`：通过。

说明：

- 若本机 `8090` 端口被占用，可通过 `DEVHUB_E2E_ORIGIN` 指定 Playwright baseURL；仓库扫描与详情仅依赖后端 API，不涉及前台与 SEO。

## v1.5.0-P0-04 本地插件包安装闭环（2026-05-13）

已执行（通过）：

- `gofmt -w internal/domain/plugin_package_install.go internal/service/plugin_package_install.go internal/service/plugin_package_install_test.go internal/service/service.go internal/transport/httpapi/router.go`：通过。
- `go test ./...`：通过。
- `go build -o .devhub/devhub .`：通过。
- `git diff --check`：通过。
- `bash -n dev.sh`：通过。
- `bash -n scripts/check-frontend.sh`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过。
- `PORT=8091 CMS_STORE=memory DEVHUB_E2E_TESTING=1 ./.devhub/devhub`：启动用于 Playwright 的本机后端（通过 `DEVHUB_E2E_ORIGIN` 指向）。
- `docker compose run --rm -e DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 admin-e2e npm run test:e2e -- tests/e2e/plugin-package-install.spec.js`：通过。
- `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8091 ./scripts/check-frontend.sh --admin-only`：通过，后台 Playwright `44 passed`。

说明：

- 插件包安装仍只写入声明/配置/迁移 pending/审计；不执行第三方代码/SQL，不动态加载前端资产。
- 本轮只涉及后端与后台插件安装升级页，不涉及前台与 SEO，因此未执行 `--frontend-only` 与 SEO curl。

## 已实现必测

基础检查：

- `go test ./...`
- `go build` 或 `go build -buildvcs=false ./...`
- `cd web/frontend-app && npm run build`
- `cd web/admin-app && npm run build`
- `bash -n dev.sh`

插件 API：

- `plugins.status` 支持扩展状态模型：`discovered`、`installed`、`migrated`、`configured`、`enabled`、`disabled`、`running`、`archived`、`config_invalid`、`migration_pending`、`migration_failed`、`dependency_missing`。
- 发布可用性只认全局 `enabled`；除 `enabled` 外的全局状态均不能新建该插件内容。
- `GET /api/v1/plugins` 只返回全局 enabled 插件。
- `GET /api/v1/plugins` 返回统一 manifest 风格的插件声明结构，包括内容类型、权限、菜单、路由和 `config_schema` 预留字段。
- `GET /api/v1/plugins` 和 `GET /api/v1/communities/:slug/plugins` 不暴露 `config_json` / `resolved_config`。
- `GET /api/v1/admin/plugins` 返回 `qa`、`docs`、`wiki` 的全局状态。
- `GET /api/v1/admin/plugins` 返回轻量 `health` 摘要，至少包含 overall/config/migration/hook/dependency 状态、最近错误和建议操作。
- `GET /api/v1/admin/plugins/:code/audit-logs` 可按插件 code 查询插件启停、配置、Hook 失败和插件内容治理审计。
- `GET /api/v1/admin/plugins/:code/migrations` 可列出内置插件 migration 声明、执行状态、失败原因和 summary。
- `POST /api/v1/admin/plugins/:code/migrations/run` 可执行该插件全部待处理 migration，并写入 `plugin.migration.run` / `plugin.migration.success` / `plugin.migration.failed` 审计。
- `POST /api/v1/admin/plugins/:code/migrations/:name/retry` 可执行或重试单条 migration，并写入 `plugin.migration.retry` / `plugin.migration.success` / `plugin.migration.failed` 审计。
- `POST /api/v1/admin/plugins/:code/migrations/:name/e2e-fail` 仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用于 E2E/API 测试注入 failed migration；注入后全局启用和子站启用必须被后端阻断。
- 已经 `success` 的 migration 再次执行不会重复破坏数据，应返回现有成功记录或保持 success。
- `PUT /api/v1/admin/plugins/:code/config` 可以保存合法 JSON，非法 JSON 应返回 400，并写入审计日志。
- `PUT /api/v1/admin/plugins/:code/config` 缺少 required、enum 非法、type 错误、integer 非整数、数字越界时应返回 400。
- 插件启停、全局配置、子站启停、子站配置和子站排序审计日志应包含 `old_value`、`new_value`、`metadata_json` 结构化字段。
- 配置审计 `metadata_json.changed_keys` 应记录本次变更的顶层配置键。
- `POST /api/v1/admin/plugins/:code/disable` 可以禁用全局插件，并写入审计日志。
- `POST /api/v1/admin/plugins/:code/enable` 可以启用全局插件，并写入审计日志。
- `POST /api/v1/admin/plugins/:code/enable` 启用前会校验配置、依赖和失败迁移；`failed` migration 应阻断启用，当前内置 no-op `pending` migration 不阻断但应通过 health / 迁移 Tab 提示。
- `GET /api/v1/communities/:slug/plugins` 只返回该子站全局 enabled 且子站 enabled 的插件。
- `GET /api/v1/admin/communities/:id/plugins` 返回某个子站的插件状态列表。
- `POST /api/v1/admin/communities/:id/plugins/:code/disable` 可以禁用某个子站插件。
- `POST /api/v1/admin/communities/:id/plugins/:code/enable` 可以启用某个子站插件。
- `PUT /api/v1/admin/communities/:id/plugins/:code/config` 可以保存合法 JSON，非法 JSON 应返回 400，并写入审计日志。
- `PUT /api/v1/admin/communities/:id/plugins/:code/config` 同样必须执行后端 `config_schema` 强校验，不能只依赖前端 Ajv。
- `PUT /api/v1/admin/communities/:id/plugins/sort` 可以调整排序，并写入审计日志。
- 全局 disabled / `config_invalid` / `migration_pending` 插件不能被子站启用。
- 子站启用插件时也应复用 Service 层 readiness 检查；当插件存在 `failed` migration 时，子站启用应失败。
- `GET /api/v1/moderator/plugin-menus` 只返回全局 enabled、子站 enabled 且当前用户有权限的插件菜单。
- `GET /api/v1/admin/plugins/:code/impact` 返回历史内容数、启用/禁用子站数、绑定板块数、近 7 天内容数、审核中内容数、菜单声明数、配置覆盖数和待执行迁移数。
- `GET /api/v1/admin/communities/:id/plugins/:code/impact` 返回同类字段，但内容、板块和子站状态计数应收敛到该子站范围。
- 插件 manifest 结构一致性测试覆盖 `code/name/version/is_system/content_types/permissions/menus/routes/config_schema/min_core_version`。
- `doc -> document`、`wiki -> wiki_page` 和 `content_type -> plugin_code` 映射测试应通过。
- 插件配置合并测试覆盖 schema 默认值、全局配置、子站配置和 `effective` 生效值。
- 插件生成模板：`go run ./cmd/devhub plugin:new ...` 应生成 `manifest.json`、`README.md`、`config.example.json`、`content-type.md`、`permissions.md`、`hooks.md`、`migrations.md` 和 `registry.example.go`；生成的 manifest 必须通过 `PluginManifestValidator`，示例配置必须通过当前简化 `config_schema` 校验。

发布与板块：

- `Service.CreatePost` 不再能绕过插件校验；业务创建必须走 `Service.CreateTopic`。
- `POST /api/v1/posts`、`PUT /api/v1/posts/:id`、`DELETE /api/v1/posts/:id` 写接口返回 `410 Gone` 或明确废弃。
- `POST /api/v1/topics` 写入归一后的 `content_type` 和 `plugin_code`。
- `doc` 参数保存为 `document`，`wiki` 参数保存为 `wiki_page`。
- 发布 `question` 后可通过 `GET /api/v1/topics/:id/qa` 看到 `qa_questions` 扩展状态。
- 发布 `document` 后可通过 `GET /api/v1/topics/:id/docs` 看到 `docs_documents` 扩展行与基础文档树。
- 发布 `wiki_page` 后可通过 `GET /api/v1/topics/:id/wiki/versions` 看到 `wiki_pages` 扩展行与版本列表。
- 发布权限码映射来自统一 `ContentTypeDefinition` 声明。
- 发布权限校验：
  - `question` 需要 `qa.question.create`
  - `document` 需要 `docs.document.create`
  - `wiki_page` 需要 `wiki.page.create`
  - `project` 需要 `projects.project.create`
  - `job` 需要 `jobs.job.create`
  - `ai_work` 需要 `ai_works.work.create`
- 子站禁用 `qa` 后，该子站不能发布 `question`。
- 其他仍启用 `qa` 的子站可以继续发布 `question`。
- 子站禁用 `docs` 后，该子站不能发布 `document`。
- 子站禁用 `wiki` 后，该子站不能发布 `wiki_page`。
- 全局 `config_invalid` / `migration_pending` 状态下，后端强传对应 `content_type` 也应拒绝创建。
- 板块不能绑定当前子站未启用的插件。
- `category.plugin_code` 与 `content_type` 对应插件不匹配时拒绝发布。
- `content_type` 不在 `category.allowed_content_types` 内时拒绝发布。
- 采纳回答后，`GET /api/v1/topics/:id/qa` 中的 `is_resolved`、`accepted_answer_id` 和回答接受状态会更新。
- 编辑 `wiki_page` 后，`GET /api/v1/topics/:id/wiki/versions` 返回的版本数应增加。

## 插件系统专项验收与 E2E 回归归档（2026-05-11）

本轮执行了插件系统专项验收，目标是确认当前插件平台已经具备可治理、可配置、可启停、可审计、可观测、可迁移的基础形态，并把 E2E 回归结果归档为长期状态来源。

已自动化并通过：

- 后端基础：`go test ./...`、`go build -o .devhub/devhub .`、`bash -n dev.sh`。
- 后台构建：`docker compose run --rm admin-e2e npm run build` 通过；Vite chunk size warning 不作为失败。
- 前台 E2E：`./scripts/check-frontend.sh --frontend-only --e2e-only` 通过，14 passed。
- 后台 E2E：`./scripts/check-frontend.sh --admin-only --e2e-only` 首次发现测试污染和旧文案断言，修复后通过，15 passed。
- SEO curl：`/topics/1/` 可见 title、canonical、Article JSON-LD、article、h1；`/c/php/` 可见 title、description、canonical、h1 和结构化数据。

本轮 E2E 已覆盖：

- 插件启停：后台插件治理 E2E 覆盖全局禁用确认、影响分析提示、全局 disabled 时子站不能启用，并在测试结束恢复 `qa` enabled。
- content_type 强校验：前台插件联动 E2E 覆盖禁用插件后发布入口不可用、后端强传禁用 content_type 被拒绝、非法 content_type 被拒绝。
- 历史内容与 SEO：前台 E2E 和 curl 均确认禁用插件不影响历史 Topic 动态详情基础 SEO。
- 权限边界：前台 E2E 覆盖普通用户不能访问版主工作台、版主只能访问授权页面，并包含跨子站 API 越权拦截。
- 插件配置：后台 E2E 覆盖插件详情配置 Tab、JSON Editor / Ajv 错误提示、子站插件配置非法 schema 值拦截。
- 后台插件中心：后台 E2E 覆盖插件列表筛选、详情 Tabs、影响分析、审计入口、子站插件配置、通用 PluginContent 页。
- 后台细操作：后台 E2E 覆盖 reports 处理与审计、moderators 创建/更新/禁用 API 反映到列表、PluginContent 内容治理和审计。

本轮发现并修复：

- 后台插件治理 E2E 中，全局禁用 `qa` 的测试失败时没有恢复插件状态，导致并行运行的插件内容页测试被污染。已将该 describe 设为 serial，并在 `finally` 中恢复 `qa` enabled。
- 影响分析弹窗已经展示“当前启用子站 / 将阻止发布的板块 / 已有历史内容”等真实字段，旧断言仍查找“将影响子站”。已同步测试断言。

仍需后续补测：

- 插件迁移：失败注入、重试成功、重复执行不破坏数据、审计定位的完整 E2E。
- Hook 治理：blocking Hook 人为失败并阻断主流程、non-blocking Hook 失败不阻断但记录后台可见的完整 E2E。
- 插件审计：按 plugin_code / community_id / action 的更多组合筛选，以及 old_value / new_value / metadata_json 的浏览器断言。
- 插件健康状态：migration_pending、config_invalid、hook warning/error 等多状态 UI 回归。

后台兼容入口：

- `POST /api/v1/admin/posts` 创建 `question` 时需要 `qa.question.create`。
- `POST /api/v1/admin/posts` 创建 `document` 时需要 `docs.document.create`。
- `POST /api/v1/admin/posts` 创建 `wiki_page` 时需要 `wiki.page.create`。
- 缺少对应插件权限时返回 `403`。
- 禁用 `qa` 后，`POST /api/v1/admin/posts` 不能创建 `question`。
- 禁用 `docs` 后，`POST /api/v1/admin/posts` 不能创建 `document`。
- 禁用 `wiki` 后，`POST /api/v1/admin/posts` 不能创建 `wiki_page`。
- 后台编辑内容不能绕过 `allowed_content_types`。
- 后台编辑内容不能绕过插件 enabled 状态。
- v1.3.1 起禁止后台普通编辑修改内容归属：修改 `site`、`board`、`content_type` 或 `plugin_code` 应失败。
- 已有内容列表、详情、评论、标签、收藏和关注不受影响。

页面与 SEO：

- `/` 可访问。
- `/search/` 可访问。
- `/topics/new/` 可访问。
- `/c/php/` 可访问。
- `/c/php/` 子站页插件导航会按当前子站 enabled 插件显示 / 隐藏问答、文档、Wiki 入口。
- `/topics/:id/` 由 Go 动态输出 SEO HTML。
- `/sitemap.xml` 正常返回。
- `/robots.txt` 正常返回。
- 插件 disabled 后，已有内容详情页仍可访问。

后台与版主体验：

- `/admin-next/plugins` 可查看插件 name/code/version/status/content_types/permissions/menus/schema 摘要。
- `/admin-next/plugins` 支持打开插件详情抽屉，按基础信息、内容类型、权限、菜单、配置、路由和 Hooks 分区展示。
- `/admin-next/plugins` 支持查看 `config_schema` 和 `resolved_config`，全局插件配置使用 JSON Editor + Ajv 做客户端基础校验，后端保存时继续按简化 `config_schema` 二次校验。
- `/admin-next/plugins` 全局启用 / 禁用有确认提示，并明确 disabled 不影响历史内容详情和 SEO。
- `/admin-next/plugins` 全局禁用确认会展示 impact 计数；Hook 错误统计来自 `hook_executions` 最近 7 天失败记录，不应伪造成完整健康评分。
- `/admin-next/communities` 的插件配置抽屉支持启用 / 禁用、全局 / 子站双状态、全局禁用原因、`config_json` 编辑、schema 参考、JSON Editor、Ajv 基础校验和数字排序保存。
- `/admin-next/communities` 子站禁用确认会展示该子站范围内的 impact 计数。
- `/moderator` 可按当前子站显示插件治理入口；全局 disabled 或子站 disabled 插件不显示。

HookBus 最小检查：

- 创建 Topic 时触发 `BeforeCreateContent` 和 `AfterCreateContent`。
- `BeforeCreateContent` blocking hook 返回错误时阻断创建，并写入 `hook_executions` 与 `plugin.hook.blocked` 审计。
- `AfterCreateContent` non-blocking hook 返回错误时不阻断主流程，并写入 `hook_executions` 与 `plugin.hook.failed` 审计。
- `POST /api/v1/admin/plugins/:code/hooks/:name/e2e-fail` 仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 可用于 E2E/API 测试注入 Hook 失败；自动化覆盖 blocking 失败阻断创建、non-blocking 失败不阻断创建、执行记录查询、审计定位和后台 Hooks Tab 失败摘要。
- 插件详情 Hooks Tab 能展示执行次数、失败次数、失败率、平均耗时、最近执行、最近失败和最近错误。
- 插件详情“运行状态”Tab 能展示 overall、config_status、migration_status、hook_status、dependency_status、pending/failed migrations、hook failures、recent_error 和 suggested_action。
- 插件详情“迁移”Tab 能展示 migration 列表、状态、最近执行时间、失败原因、rollback_supported 标识，并提供执行/重试入口。
- 插件详情“审计”Tab 使用插件专用审计接口，不再只依赖通用 audit-log target 前缀。
- 插件内容治理操作（隐藏/恢复、置顶/取消置顶、加精/取消加精、评论治理等已有操作）应写入带 `plugin_code` 的结构化审计 metadata；Core 内容可继续使用普通审计。
- 更新 Topic、置顶、精选、隐藏 / 恢复、评论锁时触发 `BeforeUpdateContent` 和 `AfterUpdateContent`。
- 删除 Topic 时触发 `BeforeDeleteContent` 和 `AfterDeleteContent`。
- 创建评论时触发 `AfterCreateComment`。
- 搜索、通知、Topic SEO 读取分别触发 `OnSearchIndex`、`OnNotificationBuild`、`OnSEOBuild`；当前只要求事件可派发，不要求存在复杂插件处理器。

SEO 源码检查：

```bash
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1|<article|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/topics/'
curl -s http://127.0.0.1:8090/robots.txt
```

插件前台入口 / 菜单可见性（v1.4.0-P1-11）：

- 后端：`go test ./...`（至少覆盖 `internal/transport/httpapi/plugin_navigation_test.go` 中的 navigation/create-options/menus preview 鉴权）
- 后台 E2E：`docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-navigation-admin.spec.js`
- 前台 E2E：`docker compose run --rm frontend-e2e npm run test:e2e -- tests/e2e/plugin-navigation.spec.js`
- SEO 回归：

```bash
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1|<article|article-tags|application/ld\\+json'
curl -s http://127.0.0.1:8090/c/php/ | rg '<title>|description|canonical|<h1|/topics/|tag-cloud'
```

## 后台 E2E Docker Runner

## 前后台前端统一检查脚本

推荐使用 `scripts/check-frontend.sh` 作为前台与后台构建 / E2E 的统一入口。脚本默认使用 Docker Compose runner，不依赖宿主机 Node/npm，并把日志写入 `.devhub/checks/{时间戳}/`。

常用命令：

```bash
./scripts/check-frontend.sh
./scripts/check-frontend.sh --target admin
./scripts/check-frontend.sh --target frontend
./scripts/check-frontend.sh --target both
./scripts/check-frontend.sh --quick
./scripts/check-frontend.sh --strict
./scripts/check-frontend.sh --quiet
```

目标选择：

- `--target admin`：只检查后台 `web/admin-app`。
- `--target frontend`：只检查前台 `web/frontend-app`。
- `--target both`：同时检查前台与后台。
- `--admin-only` 等价于 `--target admin`，`--frontend-only` 等价于 `--target frontend`。
- 在交互式终端直接运行且没有传 target 时，脚本会询问检查范围；在非交互环境默认检查 `both`，避免 CI 卡住。

模式说明：

- 默认：build + E2E。
- `--quick` / `--build-only`：只跑 build。
- `--e2e-only`：只跑 E2E。
- `--strict`：如果 `package.json` 存在 `lint` / `typecheck` 脚本，则额外执行；不存在则显示 `SKIP`，不视为失败。
- `--rebuild`：执行前先构建对应 e2e 镜像。
- `--remove-orphans`：传给 `docker compose run`，用于清理 orphan container 提示。
- 默认实时显示完整日志，同时落盘；`--quiet` 可只显示摘要和失败日志尾部。
- `--tail-lines N`：失败时展示最后 N 行日志。

结果含义：

- `PASS`：必要检查通过。
- `FAIL`：必要检查失败，脚本最终 `exit 1`。
- `SKIP`：compose 服务或 package script 不存在，按可选项跳过；例如未来某 runner 暂未创建时不会误报失败。

`admin-e2e` 与 `frontend-e2e` 关系：

- `admin-e2e` 用于后台构建和后台 Playwright E2E。
- `frontend-e2e` 用于前台构建和前台 Playwright E2E。
- 两者都由 `scripts/check-frontend.sh` 调度；也可以继续直接使用下方单独 Docker Compose 命令。

后台 Playwright E2E 使用项目内固定 Docker 镜像，不依赖宿主机 Node/npm：

```bash
docker compose build admin-e2e
docker compose run --rm admin-e2e npm run build
docker compose run --rm admin-e2e
```

v1.4.0-P1-10 增量用例：

```bash
docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-readiness-errors.spec.js
```

说明：

- 该用例依赖 `GET /api/v1/admin/plugins/:code/readiness`；如果当前环境仍在运行旧后端二进制且未包含 readiness 路由，会在用例内探测并 `test.skip`（避免把“后端未升级”误判为前端回归）。

本轮 Hook 排障页专项最小回归：

```bash
docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-hooks.spec.js
```

本轮插件 SDK / 模板专项最小回归：

```bash
go test ./internal/plugins/scaffold
go run ./cmd/devhub plugin:new --code smoke_links --name "Smoke Links" --content_type smoke_link --content_name "Smoke Link" --output .devhub/tmp/plugin-template-smoke --with_config --with_hooks --with_migration --force
```

2026-05-12 执行结果：上述 `go test ./internal/plugins/scaffold` 与 smoke test 均已通过；同时 `gofmt`、`go test ./...`、`go build -o .devhub/devhub .`、`git diff --check`、`bash -n dev.sh` 和 `bash -n scripts/check-frontend.sh` 通过。本轮未修改后台页面或前台 SEO，未执行后台 Docker build / E2E 与 SEO curl。

说明：

- `admin-e2e` 使用 `web/admin-app/Dockerfile.e2e`，基础镜像固定为 `mcr.microsoft.com/playwright:v1.59.1-noble`。
- 镜像构建阶段执行 `npm ci`；运行阶段复用镜像依赖，并通过 Docker volume 避免把 `node_modules` 写入仓库。
- 默认测试目标是 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8090`，运行 E2E 前需要先启动 DevHub 后端服务。
- E2E 前建议先执行 `docker compose run --rm admin-e2e npm run build`，让 Go 服务读取最新 `web/admin-vue` 静态产物；不要和 E2E 并行写 `web/admin-vue`。
- 首次构建会拉取较大的 Playwright 基础镜像，后续重复执行会复用本地 `sns-admin-e2e` 镜像。

当前已自动化覆盖：

- 后台登录与保护页面会话边界。
- `/admin-next/content` 内容管理打开与标题搜索。
- `/admin-next/comments` 评论管理打开与筛选。
- `/admin-next/communities` 子站管理打开与搜索。
- `/admin-next/tags` 标签管理打开与搜索。
- `/admin-next/audit-logs` 治理审计打开与动作筛选。
- `/admin-next/reports` 举报管理打开、状态/对象筛选、测试举报处理和审计日志联动。
- `/admin-next/moderators` 版主管理打开、列表展示、通过 API 验证新增/更新/停用版主并在 UI 列表中反映。
- `/admin-next/plugins` 打开与搜索筛选。
- 插件详情 Tabs、`config_schema` / `resolved_config` 展示和 schema 错误提示。
- 全局禁用确认、impact 提示和全局 disabled 后子站启用限制。
- 子站插件配置抽屉、JSON Editor 与 Ajv 错误提示。
- 通用 `PluginContent` 页入口、子站筛选和状态筛选。
- `/admin-next/qa`、`/admin-next/docs`、`/admin-next/wiki` 插件内容页打开、筛选和通用内容表展示。
- 插件内容治理代表链路：通过后台 API 对已有插件内容执行隐藏 / 恢复，并在审计日志中验证记录。
- 前台 seed 用户 `liuwei / 方圆十三 / a123456` 会在 MemoryStore 和 MySQLStore 初始化时自动写入，可直接用于前台登录、发布和手工冒烟。
- 前台种子用户 `liuwei / 方圆十三 / a123456` 会在 MemoryStore 和 MySQLStore 初始化时自动写入，可直接用于手工登录、前台发布和前台用户 E2E。

## 前台 E2E Docker Runner

前台 Playwright E2E 使用项目内固定 Docker 镜像，不依赖宿主机 Node/npm：

```bash
docker compose build frontend-e2e
docker compose run --rm frontend-e2e npm run build
docker compose run --rm frontend-e2e
```

说明：

- `frontend-e2e` 使用 `web/frontend-app/Dockerfile.e2e`，基础镜像固定为 `mcr.microsoft.com/playwright:v1.59.1-noble`。
- 镜像构建阶段执行 `npm ci`；为适应慢网环境，Dockerfile 为 `npm ci` 增加了 fetch retry / timeout 参数。
- 运行阶段通过 Docker volume 复用依赖，不把 `node_modules` 写入仓库。
- `frontend_e2e_node_modules` 命名卷挂载到 `/workspace/web/frontend-app/node_modules`；`web/frontend-app/docker/e2e-entrypoint.sh` 只在卷缺少依赖时从镜像层复制 `/app/node_modules/.` 的内容，避免出现 `node_modules/node_modules` 嵌套。
- “不依赖宿主机 Node/npm”指依赖安装、构建和测试均在 Docker 镜像或容器内完成；项目仍然通过 `package.json` / `package-lock.json` 声明 npm 依赖。
- 默认测试目标是 `DEVHUB_E2E_ORIGIN=http://host.docker.internal:8090`，运行 E2E 前需要先启动 DevHub 后端服务。
- E2E 前建议先执行 `docker compose run --rm frontend-e2e npm run build`，让 Go 服务读取最新 `web/frontend` 静态产物。

当前已自动化覆盖：

- `/` 总站首页打开，基础导航可见，游客看不到总后台入口。
- `/c/php/` 和 `/c/go/` 子站首页打开，并检查 canonical 基础 SEO 元素。
- `/search/` 关键词搜索提交，结果区域可见。
- `/topics/1/` 动态 Topic 详情打开，包含 h1、article 和 JSON-LD。
- `/topics/new/` 未登录发布拦截，提交后提示登录。
- 登录用户 Topic 详情互动：点赞、收藏、关注主题和发表评论。
- 用户中心联动：`/notifications`、`/me/activities`、`/me/favorites`、`/me/follows` 登录态访问。
- 登录发布流程：必填校验、发布 `article` 成功并打开新详情页。
- 插件状态联动：临时启用 QA 板块后验证 `question` 可选；全局禁用 `qa` 后发布页隐藏入口，API 强传 `question` 被拒绝，非法 `content_type` 被拒绝，历史 `/topics/:id` SEO 仍可访问；测试结束恢复插件和板块状态。
- 插件迁移异常联动：后台 E2E 通过测试 helper 注入 `qa_questions` failed migration，验证全局启用和子站启用被后端阻断，迁移 Tab 能看到失败原因，retry 后迁移变为 success，插件恢复启用，审计可查到 `plugin.migration.failed/retry/success`。
- 版主工作台边界：普通前台用户访问 `/moderator*` 无权限；PHP 版主可访问授权页面；API 强传 Go 子站 `community_id` 被 403 拦截。
- `/tags/go/` 与 `/c/php/tags/laravel/` 标签页打开，并检查 canonical 基础 SEO 元素。

## 已实现但后续补测

- 插件全局 `config_json` 和子站 `config_json` 接口已有自动化覆盖，仍需要真实 admin token 做联调补测。
- 插件声明里的 `config_schema`、`dependencies`、`min_core_version` 和 `hooks` 已有结构测试；`config_schema` 已接入简化后端校验，仍需要继续补测前后台展示、错误提示、边界值和完整 JSON Schema 不支持场景。
- 子站插件排序接口已有自动化覆盖，仍需要真实 admin token 和浏览器做联调补测。
- 后台插件治理中心和核心后台页面已有 Docker 化 Playwright E2E runner；仍需继续补多浏览器、更多真实 UI 操作、视觉细节和 MySQLStore 场景。
- 后台版主管理新增 / 编辑 / 停用当前通过真实 API + UI 列表验证组合覆盖；Element Plus 表单纯 UI 操作仍可作为后续更细浏览器用例。
- 后台插件内容页当前覆盖通用页打开、筛选和 API 治理代表链路；更完整的插件内容详情抽屉、审核按钮 UI 闭环仍待后续。
- 前台发布页按全局插件状态和当前板块状态已有浏览器覆盖；仍需继续补“子站 A 禁用 / 子站 B 启用”的跨子站矩阵。
- 子站导航按插件状态显示仍需继续做多子站浏览器验收。
- 版主菜单按子站插件状态和权限过滤已有 PHP 版主基础覆盖；仍需继续补多版主账号、多子站插件启停组合和 UI 操作。
- MySQL 老库插件平台结构与核心行为已通过专项验证；仍需在接入真实生产数据前做预发库备份 / 回滚演练和历史内容 SEO 的 MySQL 端浏览器矩阵。

## 待实现后补测

- 插件权限矩阵后续补测：缺少对应 create 权限时前台发布 / 后台创建 / 版主菜单均被拒绝或隐藏的更多角色组合。
- 更细粒度的权限体系补测：例如 Core 兼容类型 `article` / `news` 的细分权限码、按子站/板块维度配置权限矩阵与更明确的错误码（当前发布链路已实现最小权限码校验）。
- Projects / Jobs / AI Works 的专属扩展表、专属管理页和完整业务流程。
- P0：HookBus 的完整业务处理器、关键 Hook 事务回滚、非关键 Hook 统一错误日志和重试策略。
- P0：`config_schema` 更多边界值、错误提示和完整浏览器矩阵。
- P1：`config_schema` 配置表单增强，包括深层嵌套、字段分组、复杂数组和完整 JSON Schema 不支持场景。
- Docs 文档树专用编辑 UI、拖拽排序和批量排序。
- Wiki 版本回滚和协作编辑交互。
- QA 取消采纳最佳答案。

## 完整插件系统平台验收矩阵

基线说明：

- 当前插件平台是内置系统插件平台，不是第三方插件市场或动态插件加载环境。
- v1.3.2 起 `discovered`、`migrated`、`configured`、`running`、`config_invalid`、`migration_pending`、`dependency_missing` 已进入全局状态枚举；但完整自动状态机仍未完成，当前发布判定仍只认 `enabled`。
- `plugin_migrations` 当前支持内置 up/no-op runner、失败记录和失败重试；migration down、真实 rollback、迁移前备份和外部插件迁移包仍需后续补测。
- HookBus 当前可派发并执行内置 handler；`hook_executions` 已记录执行结果、失败统计、最近错误和平均耗时，后台 Hooks Tab 已展示这些运行时字段。插件列表/详情已展示轻量健康摘要；独立健康 API、告警和重试策略仍需后续补测。

P0 已实现或必测：

- Manifest 字段一致性：`code/name/version/is_system/content_types/permissions/menus/routes/config_schema/min_core_version/hooks`。
- `content_type -> plugin_code` 映射：`question/docs/wiki_page/project/job/ai_work` 均映射到对应插件，`doc/wiki` 能归一。
- 插件全局 enabled / disabled：全局 disabled 后不能新发布对应内容，历史详情仍可访问。
- 子站 enabled / disabled：子站 disabled 后仅该子站不能新发布对应内容，其他子站不受影响。
- 板块绑定：`categories.plugin_code` 与 `allowed_content_types` 不匹配时拒绝发布。
- 权限码校验：发布时使用 ContentTypeDefinition 中的 create_permission。
- ActorContext 来源可信：由服务端 token / admin / moderator scope 计算，客户端请求体不能伪造。
- 菜单过滤：前台、后台、版主菜单按全局状态、子站状态、权限和 scope 过滤。
- `config_json` 合法性和简化 `config_schema` 校验：非法 JSON、缺少 required、enum 非法、类型错误、integer 非整数和数字越界应保存失败。
- `resolved_config.effective`：schema 默认值 < 全局配置 < 子站配置。
- `admin_logs` old/new/metadata：插件启停、配置、排序写入 `old_value`、`new_value`、`metadata_json`，配置变更包含 `changed_keys`。
- disabled 插件历史内容访问：`/topics/:id` 不返回 404。
- `/topics/:id` SEO 动态 HTML：title、description、h1、article、标签、JSON-LD 不丢失。
- migration 新装 / 老库升级：`001_schema.sql`、`internal/store/schema.go` 和 migrations 字段口径一致。
- v1.3.4 MySQLStore 专项：可选集成测试已覆盖 `plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、`admin_logs` 新装结构，`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types` 升级字段，全局 / 子站插件启停强拦截、failed migration 阻断与 retry、Hook 记录、插件审计查询和 config_schema 校验。
- v1.3.4 插件迁移异常链路：failed migration 阻断全局启用和子站启用，retry 成功后恢复启用，已 success 的 migration 不重复破坏数据，审计可定位，后台迁移 Tab 可见失败原因。
- v1.3.4 插件权限矩阵链路：`ContentTypeDefinition.create_permission` 已覆盖 question/document/wiki_page/project/job/ai_work/core；`post.create` 只桥接 `core.topic.create`，不能创建插件内容；普通前台 token 不能调用插件治理 API；版主插件菜单继续受全局状态、子站状态、community scope 和插件权限过滤。

P0 待实现 / 待补测：

- HookBus 更多事件矩阵：v1.3.4 已自动化覆盖创建链路的 blocking / non-blocking 失败注入、`hook_executions`、审计和后台 Hooks Tab；更新 / 删除 / Search / Notification / SEO 仍需补更多异常矩阵。
- 权限矩阵后续补测：后台插件内容治理的细粒度操作权限、完整 RBAC 分配 UI、community / category 级权限配置仍需后续覆盖。
- MySQLStore 后续矩阵：插件平台结构和核心 Service 行为已自动化验证；仍需补真实 MySQL 服务下 `/topics/:id`、`/c/:slug` SEO 浏览器 / curl 矩阵和生产大库升级演练。
- `config_schema` 浏览器矩阵：后端强校验已接入，仍需补更多真实浏览器错误提示、深层 diff 和完整 JSON Schema 不支持场景。
- HookBus 业务处理器：Create / Update / Delete / Search / Notification / SEO 不仅能派发事件，还要继续补更多插件处理器、告警和重试策略验收。
- 插件 migration runner：当前已支持内置插件 up/no-op 查询、执行、失败记录、重试和后台迁移 Tab；真实 rollback、down migration、迁移前备份仍待后续。
- 完整真实 token 验收矩阵：全局禁用、子站禁用、跨子站发布、版主菜单、历史 SEO。

P1 / P2 / P3 后续验收：

- P1：schema 自动表单增强、插件 SDK 文档、插件生成模板、依赖检查、版本兼容检查、插件搜索 / 通知 / SEO 扩展。
- P2：本地插件包 zip、插件包安装向导、插件升级向导增强、外部服务型 Webhook、插件 migration runner、插件包签名校验、插件市场雏形。

## v1.3.4 插件异常治理测试矩阵收口

本节用于归档 v1.3.4 的真实覆盖状态。分类含义：

- 已自动化：已有 Go 测试或 Playwright E2E，并在本阶段执行过或纳入最低检查。
- 部分自动化：已有 API / 单元测试或单一路径 E2E，但未覆盖完整角色 / Store / 浏览器矩阵。
- 手工验证：已有明确命令或页面步骤，但本阶段未完整自动化。
- 未覆盖：尚无稳定测试保护，需后续补测。
- 跳过项：本阶段明确不做或缺少环境。

已自动化：

- 插件迁移失败注入与启用阻断：API 测试和后台 E2E 覆盖 failed migration、全局启用阻断、子站启用阻断、retry 恢复、success 不重复破坏和审计定位。
- HookBus 异常链路：API 测试和后台 E2E 覆盖 `BeforeCreateContent` blocking 失败阻断、`AfterCreateContent` non-blocking 失败不阻断、`hook_executions` 查询、审计定位和 Hooks Tab 失败摘要。
- 插件权限矩阵：Go/API 测试覆盖 `question/document/wiki_page/project/job/ai_work/article/news` 的 create permission 映射、`post.create` 只能桥接 `core.topic.create`、普通前台 token 不能访问插件治理 API。
- MySQLStore 专项：可选集成测试覆盖新装结构、升级字段、全局/子站启停强拦截、failed migration 阻断与 retry、Hook 记录、插件治理审计查询和 `config_schema` 校验。
- 插件健康状态：Go 测试覆盖 `healthy`、`config_invalid`、failed migration `error`、Hook 多次失败 `hook_error`。
- 审计筛选：Go 测试覆盖 `plugin_code`、`action`、`community_id`、`metadata`、`request_id` 组合筛选；后台 E2E 覆盖插件详情审计 Tab 的 action/community/metadata 查询。

部分自动化：

- `hook_warning` / `hook_error` 前端 UI：后台 E2E 覆盖 Hook 失败后运行状态可见；更多插件、多 Hook 名称、多时间范围组合仍待补。
- 插件审计筛选 UI：已覆盖插件详情抽屉中的核心筛选；通用 `/admin-next/audit-logs` 的 target_type/target_id/request_id/time range 浏览器矩阵仍待补。
- `config_invalid` 状态：后端拒绝非法配置并有 Go 测试覆盖显式状态；常规 UI/API 不会持久化非法配置，因此“配置无效持久状态”的浏览器场景仍属于治理预留。
- MySQLStore 浏览器矩阵：Service 层 MySQL 集成测试已覆盖核心行为；真实 MySQL 服务下 `/topics/:id`、`/c/:slug` SEO 浏览器 / curl 矩阵仍待补。

手工验证：

- 生产大库升级前备份、预发执行 `004`-`010`、启动应用、抽样验证历史内容和回滚预案。
- 插件禁用后前台导航、发布页和版主菜单在多子站、多账号组合下的视觉矩阵。
- 后台插件详情的迁移、Hooks、审计、运行状态 Tabs 在慢网和大日志量下的可读性。

未覆盖：

- HookBus Update / Delete / Search / Notification / SEO 的异常注入矩阵。
- Hook 重试、告警、自动恢复和外部监控。
- 插件内容治理批量操作权限矩阵和完整 RBAC 分配 UI。
- 深层 config diff、配置版本、配置回滚和自动表单复杂字段矩阵。
- 插件包 zip、完整安装 / 升级向导、外部服务型 Webhook、市场和动态加载。

跳过项及原因：

- 插件市场、插件上传、远程安装、Go 动态加载和第三方沙箱：不属于 v1.3.4 范围。
- migration down、硬回滚、迁移前自动备份：当前 runner 是内置 up/no-op 记录型 runner，后续 P2/P3 再设计。
- Projects / Jobs / AI Works 专属业务闭环：当前只验收平台治理归属，不补专属扩展表或专属页面。

## MySQLStore / 老库升级专项命令

用途：验证插件平台在真实 MySQLStore 和老库显式 SQL 升级路径下与 MemoryStore 口径一致。测试必须使用一次性测试库，库名建议包含 `test`。

准备 MySQL 测试库：

```bash
docker compose -f docker-compose.dev.yml up -d mysql
docker compose -f docker-compose.dev.yml exec -T mysql mysql -uroot -pDevhub_root_123456 -e "DROP DATABASE IF EXISTS devhub_test; CREATE DATABASE devhub_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL PRIVILEGES ON devhub_test.* TO 'devhub'@'%'; FLUSH PRIVILEGES;"
```

运行 MySQLStore 插件平台一致性测试：

```bash
DEVHUB_MYSQL_TESTS=1 DB_HOST=127.0.0.1 DB_PORT=3307 DB_USER=devhub DB_PASSWORD=Devhub_123456 DB_NAME=devhub_test go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v
```

验证显式插件迁移 SQL 可重复执行：

```bash
docker compose -f docker-compose.dev.yml exec -T mysql mysql -uroot -pDevhub_root_123456 -e "DROP DATABASE IF EXISTS devhub_upgrade_test; CREATE DATABASE devhub_upgrade_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL PRIVILEGES ON devhub_upgrade_test.* TO 'devhub'@'%'; FLUSH PRIVILEGES;"
docker compose -f docker-compose.dev.yml exec -T mysql mysql -udevhub -pDevhub_123456 devhub_upgrade_test < db/mysql/001_schema.sql
for round in 1 2; do
  for f in db/mysql/migrations/004_community_plugins.sql db/mysql/migrations/005_core_plugins.sql db/mysql/migrations/006_plugin_config_json.sql db/mysql/migrations/007_admin_logs_structured_plugin_audit.sql db/mysql/migrations/008_plugin_migrations.sql db/mysql/migrations/009_plugin_status_model.sql db/mysql/migrations/010_hook_executions.sql; do
    docker compose -f docker-compose.dev.yml exec -T mysql mysql -udevhub -pDevhub_123456 devhub_upgrade_test < "$f"
  done
done
```

已验证范围：

- 新装结构：`plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、`admin_logs`。
- 升级字段：`topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types`。
- 行为一致性：全局禁用强拦截、子站禁用仅影响当前子站、failed migration 阻断启用、retry 恢复、Hook 记录查询、插件治理审计查询、配置 schema 校验。

注意：

- 该测试默认跳过，必须显式设置 `DEVHUB_MYSQL_TESTS=1`；测试代码还会拒绝在库名不含 `test` 的数据库上运行。
- 生产升级仍需备份、预发演练和回滚预案；当前内置 plugin migration 仍是 up/no-op 记录型 runner，不支持 migration down 或硬回滚。
- P3：远程插件市场、在线更新、动态加载能力评估、插件沙箱、插件权限隔离。

## 必要历史回归

- 普通前台会员不能看到总后台入口。
- 前台 user token 与后台 admin token 不能混用。
- `/api/v1/moderator/*` 只接受前台 user token 和有效子站版主授权。
- 标签 alias / merged URL 不进入 sitemap，并能跳转或 canonical 到主标签。
- disabled / merged 标签不进入 sitemap。
- `sites/posts` 兼容 API 继续可用。
- 隐藏 Topic 不进入 sitemap，隐藏详情页带 `noindex,follow`。

## 后台插件管理 UI 手工验收矩阵

以下步骤用于补充 Docker 化 Playwright 最小 E2E 尚未覆盖的浏览器矩阵；执行后需要把结果回填到当前任务验收记录：

1. 使用后台管理员登录 `/admin-next`，进入“系统插件”，确认顶部说明卡片说明禁用不影响历史内容。
2. 在 `/admin-next/plugins` 顶部确认统计卡片展示：全部、enabled、disabled、system、有 schema 的插件数量（与列表筛选一致）。
3. 在 `/admin-next/plugins` 使用筛选工具栏分别验证：
   - 按 code/name 搜索
   - 按 status 筛选 enabled/disabled
   - 按 content_type 筛选（列表只保留包含该 content_type 的插件）
   - 按 system 筛选
   - 按 schema 筛选
2. 确认全局插件列表展示插件名称、code、版本、system 标识、全局状态、内容类型、权限数量、菜单数量和 schema 状态。
3. 点击“详情”，确认插件详情抽屉可打开，并能查看基础信息、内容类型、权限、菜单、配置、路由和 Hooks tabs。
4. 在详情抽屉的“配置”tab 中确认 `config_schema` 与 `resolved_config` 使用 JSON 代码块展示，复制按钮可用或有明确失败提示。
5. 点击全局“配置”，输入合法 JSON，例如 `{"mode":"test"}`，保存后刷新列表，确认全局配置持久化。
6. 在全局“配置”中输入非法 JSON，例如 `{bad`，确认前端阻止保存并提示 `config_json 不是合法 JSON`。
7. 全局禁用任一测试插件，确认二次确认文案说明所有子站不能新发布、入口隐藏、历史内容和 SEO 不受影响；再启用恢复。
8. 进入“子站管理”，打开任一子站的“插件配置”抽屉，确认列表展示插件名称、code、版本、全局状态、子站状态、内容类型和 `sort_order`。
9. 禁用子站插件，确认二次确认文案说明只影响当前子站的新发布、导航、发布入口和版主菜单，不影响历史内容和 SEO。
10. 启用子站插件，确认全局 enabled 时可恢复子站 enabled；若先在全局插件页禁用该插件，再回到子站启用应被禁用并显示“该插件已被全局禁用，不能在子站启用”。
11. 打开子站“配置”，确认能看到 schema 参考；输入合法 JSON 并保存，刷新后确认 `config_json` 持久化。
12. 再次打开子站“配置”，输入非法 JSON，确认前端阻止保存并提示 `config_json 不是合法 JSON`。
13. 使用上移 / 下移或数字排序调整 `sort_order`，点击“保存排序”，刷新后确认顺序保持。
14. 到对应子站前台和发布页确认禁用插件入口不展示；直接强传对应 `content_type` 发布应由接口拒绝。

## 插件详情抽屉 Tabs + JSON Editor 手工验收

以下步骤用于验收“插件详情抽屉 Tabs + JSON 配置编辑器”增强。基础路径已由 `admin-e2e` 覆盖，复制、保存持久化和视觉细节仍需手工补测：

1. 后台管理员登录 `/admin-next/plugins`，点击任一插件“详情”，确认抽屉打开，并包含 Tabs：
   - 概览 / 内容类型 / 权限 / 菜单 / 配置 / Hooks / 路由。
2. 点击列表“权限/菜单/配置”快捷按钮，确认能直接打开抽屉并切换到对应 Tab。
3. 在“权限”Tab：
   - 输入权限码关键字搜索，表格过滤生效；
   - 点击“复制”能复制权限码，或在不支持剪贴板时有明确提示。
4. 在“菜单”Tab：确认显示 area/title/path/permission/sort_order，并提示“菜单展示仍受全局/子站状态和权限过滤影响”。
5. 在“配置”Tab：确认同时展示：
   - `config_schema`（只读 JSON）
   - `config_json`（JSON Editor 可编辑）
   - `resolved_config`（只读 JSON）
6. 在 JSON Editor 中修改 `config_json` 为合法 JSON（例如 `{}` 或 `{ "example": true }`），点击保存：
   - 若 `config_schema` 存在：应先通过 Ajv 校验；
   - 保存成功后关闭抽屉再打开，配置值应保持；
   - 后端应仍做二次校验并写审计日志（见 admin_logs）。
7. 模拟 schema 校验失败（例如 schema 要求必填字段但配置缺失）：
   - 页面应显示 schema 错误列表；
   - “保存”按钮应被禁用或保存失败有明确提示。
8. 验证辅助操作：
   - 点击“格式化”保持 JSON 值可读；
   - 点击“清空为空对象”把配置重置为空对象；
   - 点击“复制”复制当前 JSON。
9. 在 “Hooks”Tab：确认不伪造 handler 状态；“平台调用点”列仅在可确认接入时显示“存在”，其余显示“未知/未覆盖”。

## 子站插件配置抽屉升级 手工验收

以下步骤用于验收“子站插件配置抽屉升级”（`/admin-next/communities` -> 子站 -> “插件”）：

1. 后台管理员登录 `/admin-next/communities`，打开任一子站的“插件配置”抽屉，确认顶部展示：
   - 子站名称、slug；
   - 子站 enabled 插件数量、子站 disabled 插件数量、全局 disabled 插件数量。
2. 验证筛选能力：
   - 切换：全部 / 子站已启用 / 子站未启用 / 全局已禁用；
   - 按 name/code 搜索；
   - 按 content_type 筛选（只保留包含该 content_type 的插件）。
3. 列表字段检查：确认展示插件名称、code、全局状态、子站状态、content_types、sort_order、配置覆盖状态与操作按钮。
4. 全局 disabled 限制：对全局 disabled 的插件，“启用”按钮应禁用；并在说明中明确“该插件已被全局禁用，不能在子站启用”。
5. 子站 disabled 提示：对子站 disabled 的插件，说明中应提示“当前子站未启用，不能新发布对应内容”。
6. 配置编辑：点击某插件“配置”，确认弹窗内：
   - `config_schema` 只读展示；
   - `config_json` 使用 JSON Editor；
   - `resolved_config` 只读展示（如接口返回）。
7. Ajv 校验：当 schema 校验失败时应显示错误，并禁止保存；修复错误后可保存。
8. 保存配置后刷新：保存成功后，关闭弹窗并刷新列表；再次打开该插件配置，配置值应保持。
9. JSON 辅助操作：点击“清空为空对象”应将配置置为空对象；然后点击保存应写入后端。
10. 排序能力：通过 input-number 或上移/下移调整顺序，点击“保存排序”，刷新后顺序保持。

## 插件治理中心：影响分析与审计入口 手工验收

1. 在 `/admin-next/plugins` 全局禁用任一插件（测试用），确认禁用确认弹窗中：
   - 若 impact 接口可用，会显示“影响子站/影响板块/已有内容/审核中内容/菜单声明”等计数；
   - 若 impact 接口不可用，会显示“影响范围统计待后端接口支持或当前环境暂不可用”，不出现假数字。
2. 在 `/admin-next/communities` -> 子站 -> “插件配置”中禁用某个子站插件，确认禁用确认弹窗中：
   - 若 impact 接口可用，会显示该子站范围内的板块数/已有内容数等计数；
   - 不影响历史内容与 SEO 的提示仍存在。
3. 在 `/admin-next/plugins` 打开任一插件详情抽屉，进入“审计”Tab：
   - 默认按 `plugins#<plugin_code>` 前缀筛选；
   - 可按“动作关键字”进一步筛选；
   - 可输入 `community_id` 限定子站范围；
   - 能查看 old_value/new_value/metadata_json（如为空表示该条日志未结构化，不伪造）。

## PluginContent（插件内容页）手工验收

1. 在 `/admin-next/plugins` 点击某插件的“管理”按钮进入插件内容页（`/qa`、`/docs` 等）。
2. 页面顶部显示：
   - 插件名、plugin_code、插件状态、健康状态和内容类型数量；
   - disabled / archived 时提示“不能新建，但历史内容仍可治理”。
3. 子站筛选：
   - 选择一个子站后，列表应切换到该子站内容（通过 `admin/posts?site=<slug>&plugin_code=<code>&content_type=<type>`）。
4. 状态筛选：
   - 切换 publish/offline/pending，列表筛选应生效。
5. 列表中展示后端返回的 `plugin_code` / `content_type` 字段，且不同插件内容不会因前端 OR 过滤混入。
6. 批量治理：
   - 批量隐藏 / 恢复不退化；
   - 批量审核通过 / 拒绝、置顶 / 取消置顶、加精 / 取消加精可触发既有后台 topic 治理能力；
   - 操作后显示成功 / 失败明细，并可跳转审计日志。

## 阶段 B 插件治理体验手工验收

本节用于验收“阶段 B：引入 i18n，统一优化插件治理体验”。

已自动化或构建覆盖：

- 后台构建会覆盖 i18n 挂载、插件详情抽屉、配置编辑器和 PluginContent 编译可用性。
- 既有后台 E2E 继续覆盖 `/admin-next/plugins`、插件详情 Tabs、JSON/Ajv 校验、禁用确认、子站插件配置和 PluginContent 入口。
- 2026-05-11 复查截图暴露的漏网英文后，已补齐插件详情概览 / 运行状态 / 内容类型 / Hook / 迁移 / 路由 / 审计 Tab，以及子站插件配置弹窗中的 `config_schema`、`config_json`、`resolved_config` 标签中文化；后台 E2E 同步断言当前中文文案并通过 `18 passed`。
- 本轮命令结果：`go test ./...`、`go build -o .devhub/devhub .`、`bash -n dev.sh`、`bash -n scripts/check-frontend.sh`、`./scripts/check-frontend.sh --quick`、`./scripts/check-frontend.sh --admin-only --e2e-only` 均通过；`--quick` 最新日志目录 `.devhub/checks/20260511-225223/`，后台 E2E 最新日志目录 `.devhub/checks/20260511-225516/`。

需要手工确认：

1. 后台插件中心主要用户可见英文应改为中文：状态、健康状态、配置模型、Hook、迁移、审计、按钮和筛选项均应显示中文标签。
2. `plugin_code`、`content_type`、`hook_name`、权限码和 JSON key 仍保持原始技术值，不翻译后提交给后端。
3. 插件详情 -> 配置 Tab：
   - 能在“表单模式”和“JSON 高级模式”之间切换；
   - 表单模式能渲染 string / number / integer / boolean / array / object / enum 基础字段；
   - JSON 高级模式仍可编辑原始 JSON；
   - schema 校验错误可见，保存按钮应受错误状态限制；
   - 配置差异能展示原配置、新配置和变更字段；
   - 敏感字段在差异预览中脱敏；
   - 最终生效配置可见。
4. 子站插件配置弹窗：
   - 继续展示全局状态和子站状态；
   - 配置编辑器同样支持表单 / JSON 模式、差异和最终生效配置预览；
   - 子站配置覆盖全局配置的提示仍可见。
   - `配置模型`、`子站配置`、`最终生效配置` 等标签应显示中文；JSON key 和插件技术值可以保留英文。
5. PluginContent：
   - 支持子站、状态、关键词和内容类型筛选；
   - 列表展示插件编码、内容类型、子站、状态、更新时间和评论数；
   - 多选后可批量隐藏 / 恢复、审核通过 / 拒绝、置顶 / 取消置顶、加精 / 取消加精；
   - 批量操作前有确认并显示影响数量；
   - 操作后刷新列表，展示成功 / 失败明细，并提供“查看审计日志”入口；
   - “查看审计日志”会进入 `/admin-next/audit-logs`，并预填 `plugin_code`、`content_type`、`action`、`target_type` 和 metadata 筛选；
   - 详情抽屉可展示当前内容的基础治理信息。

仍需后续补测：

- 后台全站英文文案扫描与非插件页面中文化。
- PluginContent 更完整的跨页面审计高亮和完整权限矩阵。
- 复杂 `config_schema` 嵌套对象、数组对象和字段分组渲染。

## 2026-05-10 插件治理中心专项验收记录

本轮执行了命令行、API 和 SEO 层面的集中验收。当时尚未接入自动化浏览器 runner，因此插件治理中心 UI 的点击矩阵按上方手工步骤保留；后续已新增 Docker 化 Playwright 最小 E2E runner。

已执行并通过：

- `go test ./...`
- `go build -o .devhub/devhub .`
- Docker Node 后台构建：`docker run --rm -v "$PWD/web/admin-app":/app -w /app node:20-alpine sh -lc "npm ci && npm run build"`
- 临时服务：`PORT=18090 CMS_STORE=memory ./.devhub/devhub`
- `/admin-next`、`/admin-next/plugins`、`/admin-next/qa` 均返回 200。
- `GET /api/v1/admin/plugins` 可返回插件声明、状态、权限、菜单、`config_schema` 和 `resolved_config`。
- `GET /api/v1/admin/plugins/qa/impact` 返回全局 impact 轻量计数。
- `GET /api/v1/admin/communities/1/plugins/qa/impact` 返回子站 impact 轻量计数。
- `PUT /api/v1/admin/plugins/qa/config` 合法配置返回 200，缺少 required 字段返回 400。
- `PUT /api/v1/admin/communities/1/plugins/qa/config` 类型错误返回 400。
- 全局禁用 `qa` 后，子站启用 `qa` 返回 400，提示插件全局未启用；验收后已恢复 `qa` enabled。
- `GET /api/v1/admin/audit-logs?target=plugins%23qa` 返回插件启停审计，并包含 `old_value`、`new_value`、`metadata_json`。
- 首页源码未暴露 `/admin-next` 总后台入口；版主入口为登录后按权限显示的隐藏入口。
- `GET /api/v1/moderator/plugin-menus?community_slug=php` 使用前台 user token 返回当前子站可见插件菜单。
- `/topics/1/` 返回 SEO HTML，包含 title、description、h1、article、标签链接和 Article JSON-LD。
- 全局禁用 `qa` 后访问已有 question `/topics/2/` 仍返回 SEO HTML；验收后已恢复 `qa` enabled。
- `/c/php/` 返回子站 SEO HTML，包含 title、description、canonical、h1、topic 链接和热门标签。
- `/sitemap.xml` 和 `/robots.txt` 正常返回。

失败 / 跳过项：

- `cd web/admin-app && npm run build` 在宿主机失败，原因是当前环境没有 `npm`；已按项目规则用 Docker Node 完成构建。
- 未执行 `cd web/frontend-app && npm run build`：本轮未修改前台代码。
- 未执行真实浏览器点击矩阵：当时仓库没有 Playwright/Cypress 等自动化 runner；后续已新增 `admin-e2e` 最小 E2E，完整多账号/多子站矩阵仍需扩展。

发现的问题：

- P0：未发现。
- P1：未发现会阻塞插件治理中心基础可用性的后端/API/SEO 问题。

## 阶段 C/D/E/F 插件生命周期与软卸载验收

本节用于阶段 C/D/E/F 的最小验收，不扩展成大规模 E2E。

已自动化：

- Go 单测覆盖内置插件归档后 `plugins.status=archived`、`lifecycle_status=archived`。
- Go 单测覆盖归档插件不能通过 `ValidateTopicPluginAccess` 新建内容。
- Go 单测覆盖归档插件不能被子站启用。
- Go 单测覆盖恢复插件后默认进入 `disabled`，不会自动启用。
- Go 单测覆盖 failed migration 会阻断恢复。
- API 测试覆盖 `POST /api/v1/admin/plugins/:code/archive`、`restore`、归档后创建拦截、子站启用拦截、恢复后再启用和审计定位。
- 前台 Playwright E2E 覆盖归档 `qa` 后发布页不再展示 `question`，对应问答板块不再可选。
- 前台 Playwright E2E 覆盖强传归档插件 `content_type=question` 被后端拒绝。
- 前台 Playwright E2E 覆盖子站启用归档插件被后端拒绝。
- 前台 Playwright E2E 覆盖归档插件后历史 `/topics/2/` 仍可访问，`h1`、`article` 和动态 SEO 基础元素存在。
- 后台 Playwright E2E 覆盖归档确认弹窗影响范围、归档 badge、详情归档时间和恢复后默认 `disabled` 提示。
- 后台 Playwright E2E 覆盖归档插件历史内容仍可进入 PluginContent 查看，并显示只能治理历史内容、不能新建的提示。
- 后台 Playwright E2E 覆盖归档态 PluginContent 历史内容批量隐藏 / 恢复仍按后台权限可用。

手工 / 轻量冒烟：

- 后台插件列表应显示生命周期字段和状态原因。
- 插件详情审计 Tab 可查 `plugin.archived` / `plugin.restored`。

未覆盖 / 后续：

- 生产 MySQL 大库归档 / 恢复耗时专项。
- 归档后所有插件、所有子站导航入口的完整浏览器矩阵；当前自动化以 `qa/question` 为代表路径。
- PluginContent 归档态自动化覆盖批量隐藏 / 恢复，并以置顶 / 取消置顶覆盖新增批量治理最小链路；完整只读策略和权限矩阵仍待补。
- 外部插件包安装、上传、远程安装、Go 动态加载、第三方沙箱和硬卸载均未实现，不能写成已验收。
- P2：后台构建存在 Vite chunk size warning，主要来自 `PluginConfigEditor` 等大 chunk；后续可考虑按需加载或手动拆包。

2026-05-12 归档态专项执行结果：

- `./scripts/check-frontend.sh --frontend-only`：通过，前台 build 通过，前台 E2E `16 passed`；覆盖归档插件入口隐藏、强传拦截、子站启用阻断和历史 Topic SEO 回归。
- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 E2E `20 passed`；覆盖归档插件治理中心细节、PluginContent 历史治理提示和归档态批量隐藏 / 恢复。

## Manifest 校验、dry-run 与配置型安装验收

本节用于记录 manifest + 配置型插件安装预备能力，不代表插件市场、插件包上传、远程安装或动态加载已完成。

已自动化：

- Go 单测覆盖合法 manifest 校验通过，并返回 normalized manifest、checksum、impact summary 和 install preview。
- Go 单测覆盖 code 冲突失败。
- Go 单测覆盖 content_type 冲突失败。
- Go 单测覆盖非法 Hook 名称失败。
- Go 单测覆盖非法 `config_schema` 失败。
- Go 单测覆盖依赖缺失时返回 warning / dependency 信息，不直接执行第三方代码。
- Go 单测覆盖 `InstallPluginManifest` 安装 manifest + 配置型插件后初始为 installed + disabled。
- Go 单测覆盖安装后的 manifest-only content type 可通过统一内容类型映射读取 create permission。
- Go 单测覆盖批量归档 / 恢复接口返回逐项 succeeded / failed 结果。

部分自动化：

- API 文档已记录 `POST /api/v1/admin/plugins/manifest/validate`、`POST /api/v1/admin/plugins/dry-run`、`POST /api/v1/admin/plugins/install`、`GET /api/v1/admin/plugins/health`、`POST /api/v1/admin/plugins/bulk-archive` 和 `bulk-restore`；浏览器 UI 已补最小 E2E，安装向导已进入抽屉分步流程，后续只继续增强插件包上传、依赖明细和更细影响对象列表。
- Health summary 当前由 Go/API 行为和后台已有运行状态视图覆盖；独立 `/plugins/health` 页面级展示仍待后续。

未覆盖 / 后续：

- 外部服务型 Webhook 真实 HTTP 调用、签名、超时和失败策略。
- 插件升级 dry-run / upgrade 已有抽屉分步展示；后续补独立版本兼容矩阵页面和更细升级影响对象列表。
- 插件包 zip 读取、签名校验、文件安全扫描和插件市场页面。
- 版本兼容矩阵 UI 和依赖版本范围完整自动化。
- 外部插件迁移真实 DDL runner、migration down、硬回滚和迁移前备份。

## 升级 dry-run 与版本兼容矩阵验收

本节用于 P2 升级预备能力；当前已完成最小升级执行闭环和后台抽屉式升级向导，但回滚、migration down、插件包升级和更细版本变更审计仍待后续。

已自动化：

- Go 单测覆盖 `POST /api/v1/admin/plugins/:code/upgrade/dry-run` 的兼容矩阵返回，包含 current/new 版本、Core 兼容范围、变更字段和 diff。
- Go 单测覆盖升级 dry-run 对现有插件 code 的校验与版本不递增时的警告信息。
- Go 单测覆盖 `POST /api/v1/admin/plugins/:code/upgrade` 的真实执行闭环，包含版本更新、manifest checksum 变更和状态保持。
- 后台 E2E 覆盖插件列表中的“升级预览”入口、结果面板和版本兼容摘要。

部分自动化：

- 版本兼容矩阵目前基于 manifest validate + upgrade dry-run 两层信息拼装，升级执行结果仍以结构化结果面板展示，完整升级 UI 向导、回滚和版本变更审计仍待后续。

未覆盖 / 后续：

- 版本兼容矩阵独立页面。
- 升级影响分析的更细粒度对象列表。

## `/admin-next/plugins` 治理入口收口验收

本节记录本轮后台插件治理中心最小 UI 收口和对应 E2E 结果。

已自动化：

- `/admin-next/plugins` 顶部健康总览卡片可见，至少展示健康、警告、异常、已禁用、已归档、迁移待处理、配置无效、依赖缺失和 Hook 异常聚合。
- `/admin-next/plugins` 顶部支持 `校验 Manifest`、`Dry-run 预览`、`安装插件`、`批量归档`、`批量恢复` 入口。
- `/admin-next/plugins` manifest 面板可打开并提交合法 manifest JSON，返回结构化结果摘要。
- `/admin-next/plugins` dry-run 面板可打开并返回结构化安装预览。
- `/admin-next/plugins` install 面板可打开并返回结构化安装结果。
- `/admin-next/plugins` upgrade dry-run 可展示当前版本 / 新版本 / 兼容状态 / 变更字段；真实 upgrade 由 Go 单测覆盖最小执行闭环。
- `/admin-next/plugins` 批量归档 / 恢复可对选中插件返回结果摘要。
- `/admin-next/plugins` 的 manifest 输入、校验结果、dry-run、安装确认、升级预览和升级执行已收口为右侧抽屉分步流程，不再作为页内长面板挤压插件列表；原 `plugin-manifest-panel`、`plugin-manifest-input` 等 E2E 锚点保留。
- `/admin-next/plugins` 批量归档 / 恢复已支持操作前 impact 预览和操作后 succeeded / failed 明细展示，并提供审计跳转入口。
- `/admin-next/plugins` 状态治理视图已作为异常处理入口，按迁移待处理、迁移失败、Hook 异常、配置无效、依赖缺失和已归档插件聚合。
- 插件详情抽屉可展示运行状态说明、归档态提示、状态原因和建议操作。
- 现有插件治理 E2E、PluginContent、迁移、Hook、审计和归档态浏览器链路未退化。

已跳过的旧回归：

- `opens plugin detail tabs and shows schema validation errors`
- `archives plugin and shows archived state with restore entry`

说明：本轮优先把测试步骤收敛到升级执行、安装向导和批量治理的最小闭环，旧详情 / 归档回归已由其他插件治理测试间接覆盖，保留会增加执行时间和状态污染风险。

部分自动化：

- 安装 / 升级向导的后端行为已有 Go / API 保护；本轮后台 E2E 已覆盖 manifest 输入、校验 / dry-run、确认安装、升级预览、确认升级和结果面板的最小浏览器链路。
- 状态治理视图当前聚合插件列表已有健康摘要和状态字段；更多告警、自动恢复和外部服务 Webhook 状态仍是后续能力。

未覆盖 / 后续：

- 插件包 zip 上传、外部服务型 Webhook、版本兼容矩阵独立页面和更细批量治理策略。

## v1.3.5 测试收口

已覆盖：

- `/admin-next/plugins` 信息架构调整后仍可打开，健康总览、筛选、批量操作、详情抽屉入口不退化。
- 安装向导三步流：Manifest 输入、校验 / dry-run 结构化预览、确认安装。
- 升级向导三步流：Manifest 输入、兼容矩阵 / diff / migration plan 预览、确认执行升级。
- 批量归档 / 恢复：影响预览、二次确认、`succeeded` / `failed` 表格、审计跳转。
- 状态治理页：异常插件、迁移待处理、Hook 异常、配置无效、依赖缺失和归档插件入口可读。
- PluginContent：归档 / 禁用插件历史内容仍可查看，批量隐藏 / 恢复不退化。

历史执行记录：

- `./scripts/check-frontend.sh --admin-only`：通过，后台 build 通过，后台 E2E `28 passed / 1 skipped`；日志目录 `.devhub/checks/20260512-231335/`。（该 skipped 已在 2026-05-13 的 v1.4.0 收口验收中恢复并通过）

精简原则：

- 只新增覆盖新向导和关键入口的最小后台 E2E。
- 不在 v1.3.5 UI 收口阶段扩展插件市场、远程安装、动态加载或外部服务型 Webhook 的浏览器矩阵。

## 2026-05-11：固定 DevHub 后台 E2E Docker 镜像

本轮目标是为后台 E2E 固定项目内 Docker 镜像，避免重复临时拉取大型 Playwright 镜像，并保证本地与 CI 可以复用同一条命令链路。

已执行并通过：

- `docker compose build admin-e2e`
- `docker compose run --rm admin-e2e npm run build`
- `docker compose run --rm admin-e2e`

验收结果：

- `admin-e2e` 首次构建会拉取 `mcr.microsoft.com/playwright:v1.59.1-noble`，后续复用本地 `sns-admin-e2e` 镜像。
- 后台构建在容器内完成，不依赖宿主机 `npm` / `node`。
- Playwright E2E 当前 5 条用例全部通过，覆盖插件治理中心核心路径。
- `node_modules`、`web/admin-vue`、`playwright-report` 和 `test-results` 均为忽略产物，不提交仓库。

注意：

- E2E 依赖已启动的 DevHub 服务，默认访问 `http://host.docker.internal:8090`。
- E2E 前需要先用 `admin-e2e` 构建后台静态产物；不要和 E2E 并行写 `web/admin-vue`，否则可能读到短暂不完整的静态资源。

## v1.4.0-P1-07 依赖检查与版本兼容矩阵验收

已覆盖能力：

- 后端单测覆盖无依赖、required 依赖满足 / 缺失 / disabled / archived / migration_failed / config_invalid、optional 缺失 warning、版本满足 / version_mismatch、自依赖、两节点 / 三节点循环、Core 兼容、upgrade dependency diff、upgrade required 新依赖阻断和 enable 时重新检查依赖。
- 后台 Playwright 新增 `web/admin-app/tests/e2e/plugin-dependencies.spec.js`，覆盖安装向导依赖矩阵、required 缺失、optional 缺失、Core 不兼容、升级向导新增 required 依赖阻断、插件详情 Dependencies 区域和依赖插件定位入口。
- 后台安装 / 升级向导展示 `dependency_summary`、逐项 `dependencies`、Core `compatibility` 和 `dependency_diff`；后端仍负责最终阻断，不依赖前端判断。

本轮执行记录：

- `gofmt -w internal/plugins/scaffold/scaffold.go internal/plugins/manifest_validator.go internal/plugins/version_compat.go internal/plugins/version_compat_test.go internal/service/service.go internal/service/plugin_dependencies_test.go internal/domain/models.go internal/transport/httpapi/router.go`：通过。
- `go test ./internal/plugins ./internal/plugins/scaffold ./internal/service ./internal/transport/httpapi`：通过。
- `docker compose run --rm admin-e2e npm run build`：通过；Vite 仍有既有大 chunk warning。
- `docker compose run --rm admin-e2e npm run test:e2e -- tests/e2e/plugin-dependencies.spec.js`：首次因后端服务仍为旧进程、UI 未拿到新依赖结构失败；使用当前代码重启 DevHub 后复跑通过，`2 passed`。
- `./scripts/check-frontend.sh --admin-only`：最终通过，后台 build 通过，后台 E2E `30 passed / 1 skipped`；日志目录 `.devhub/checks/20260513-003728/`。（该 skipped 已在 2026-05-13 的 v1.4.0 收口验收中恢复并通过）

仍遗留：

- 不支持自动安装依赖、远程下载依赖、插件市场推荐或依赖图大屏。
- 不支持 npm 风格 `^`、`~`、`||`、预发布标签等复杂版本约束。
- optional 循环依赖当前按 error 阻断，后续如要放宽需同步后端、UI 和文档。
