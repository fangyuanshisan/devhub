# DevHub Codex / AI Agent Rules

[返回文档大纲](README.md)

更新时间：2026-05-10

本文档是后续 DevHub 1.x 开发任务的固定协作规则。Codex / AI Agent 接到任务后，应先阅读并遵守本文档，再阅读 `docs/README.md`、`docs/PROJECT_PROGRESS.md`、`docs/API.md`、`docs/PLUGIN_ARCHITECTURE.md`、`docs/SEO.md`、`docs/TESTING.md` 和当前版本 release 文档。旧版本 Release Notes 只作追溯依据，不作为当前必读主线。

## 固定项目约束

- 项目名称必须保持 `DevHub`，不要改成 LearnKu、LearnKu Clone 或其他名称。
- 默认端口保持 `8090`。
- 前台入口保持 `/`。
- 后台入口保持 `/admin-next`。
- 不要删除 `sites/posts` 兼容 API。
- 不要恢复旧原生 HTML / CSS / JS 页面。
- 不要把 `/topics/:id` 改成纯 CSR 空壳。
- 不要重新引入 `getStaticPaths` 预生成 Topic 详情页。
- `/topics/:id` 必须保持百度 SEO 友好的 Go 动态 HTML。
- 后续如果增加列表页、子站页、标签页 SEO，也不能破坏 `/topics/:id` 的动态 SEO 兜底。

## 身份边界规则

- 前台用户 `users` 与后台人员 `admin_users` 必须分开。
- 前台用户用于注册、登录、发帖、评论、点赞、收藏、关注、举报、我的动态和我的通知。
- 后台人员用于 `/admin-next` 登录、后台管理、系统配置、后台用户管理和全局审计。
- 子站版主不是全局后台管理员；版主本质仍是 `users`，通过 `community_moderators` 获得指定子站治理权限。
- 版主工作台使用 `users` 登录态和 `community_moderators` 授权关系，不是 `admin_users` 后台。
- 前台会员导航不得默认暴露 `/admin-next` 总后台入口；子站版主只显示 `/moderator` 版主工作台入口。
- 前台“我的”类页面和关注操作必须携带前台 user token，不能使用后台 admin token 或无 token 请求。
- 发布页必须让 `content_type` 与当前子站 `categories.content_type` 保持一致，不能为绕过错误取消后端校验。
- v1.3.0 起问答、文档、Wiki 属于内置系统插件；发布页还必须遵守 `categories.plugin_code`、`categories.allowed_content_types` 和插件 enabled 状态。
- 后台子站管理主入口统一为 `/admin-next/communities`；`/admin-next/sites` 只能作为隐藏兼容入口或重定向。
- 版主工作台必须由后端校验 `community_id` 范围，不能只靠前端隐藏按钮。
- 版主工作台不能越权管理其他子站，不能管理后台人员、版主分配、全局子站配置或系统设置。
- 复杂 RBAC、权限点矩阵、版主任期和绩效统计不属于当前阶段。
- `/api/v1/admin/*` 必须使用后台身份，或明确允许的子站版主身份，并继续校验 `community_id` / 子站范围。
- 普通前台 token 不能访问后台接口。
- 后台 admin token 不能被当作前台 `users` 身份使用。
- 审计日志必须尽量记录 `actor_type` 和 `actor_id`，区分 `admin_user`、`moderator` 和 `system`。
- demo user 只能作为开发阶段兜底或 seed 数据，不应作为生产权限规则。

## 环境与启动规则

- 开发、构建、测试、数据库环境以 Docker / docker compose / `dev.sh` 为准。
- 不依赖宿主机 `npm`、`node`、`go`、`mysql` 一定存在。
- 前台 / 后台前端检查优先使用 `scripts/check-frontend.sh` 和项目内 Docker Compose E2E runner；后台 Playwright E2E 使用 `admin-e2e` 服务，前台 Playwright E2E 使用 `frontend-e2e` 服务，避免依赖宿主机 Node/npm 或每轮临时拉取浏览器镜像。
- 推荐开发启动：`./dev.sh --restart`。
- 修改 Go 后端代码后必须重启服务，优先使用 `./dev.sh --restart`。
- 如果 `dev.sh` 或 `go run` 因网络、Go 模块解析、端口释放竞态等问题不稳定，可使用二进制兜底：

```bash
go build -o .devhub/devhub .
PORT=8090 CMS_STORE=memory ./.devhub/devhub
```

- 如果 Docker 镜像、容器、权限或网络缺失，需要明确提醒用户处理，不要把临时代理、token、私钥写入代码或文档。

## 文档同步规则

- 代码、接口、页面、数据结构、SEO、部署、测试发生变化后必须同步文档。
- 不允许只改代码不改文档。
- 文档必须反映真实实现，不要把未来规划写成已完成能力。
- 如果某能力只有后端完成、前端未接入，文档必须明确标注。
- 如果某页面只是占位，不要写成已完成。
- API 路径以 `internal/transport/httpapi/router.go` 的真实路由为准。
- 数据结构以 `internal/domain`、`internal/store`、`db/mysql/001_schema.sql` 和启动迁移辅助的真实实现为准。

## 任务结果记录规则

每轮任务完成后，Codex 必须同步更新项目文档中的执行结果。聊天回复只能作为摘要，项目文档才是长期状态来源。

每轮至少需要更新或确认以下内容：

1. 本轮任务标题
2. 修改范围
3. 已完成事项
4. 未完成事项
5. 新发现风险
6. 已执行检查命令和结果
7. 失败项或跳过项及原因
8. 下一轮建议
9. 是否影响 API / 数据库 / 权限 / SEO / 插件系统 / 前后台 UI

要求：

- 不能只在聊天回复里说完成，文档不记录。
- 未执行的测试不能写成已通过。
- 未完成能力不能写成已完成。
- 预留、部分完成、待验收能力必须明确标注。
- 项目长期状态以文档为准，不以对话摘要为准。

## 开发与验收边界

- 日常开发阶段只做最低必要检查，避免每个小任务都执行完整测试矩阵。
- 最低检查通常包括：
  - 相关文件路径存在。
  - Markdown 链接和格式基本正确。
  - Go 代码改动后执行 `gofmt`、`go test ./...` 或至少说明无法执行原因。
  - 前后台改动后按真实环境选择 `dev.sh` 或 Docker Node 构建，无法执行时说明原因。
- 完整测试矩阵单独作为验收任务执行，详见 `docs/TESTING.md`。
- 不要在普通文档或小补丁任务中扩展大型业务功能。

## Git 与工作区规则

- 可能存在用户或其他 Agent 的未提交改动，不要回滚不属于当前任务的变更。
- 不要执行 `git reset --hard`、`git checkout --` 等破坏性命令，除非用户明确要求。
- 不要自动打 git tag，除非用户明确要求。
- 版本归档前应提醒用户先检查 `git status` 和 `git diff`。

## 当前 1.x 版本方向

- 当前版本线：`v1.3.x`，主题是“Core + Plugins 架构拆分与插件平台收口”。
- 当前最高优先级长期主线：完成完整插件系统。Core 只保留通用社区底座，业务能力逐步通过插件声明、状态、权限、菜单、配置、Hook、migration、API、SEO、通知、搜索和测试矩阵扩展；下一阶段完整目标以 `docs/PLUGIN_SYSTEM_ROADMAP.md` 为准。
- 当前 P0 目标：把内置系统插件底座做稳，包括 Manifest、Registry、ActorContext、权限码、全局插件状态、子站插件状态、板块绑定、发布校验、菜单过滤、config_json、config_schema 基础校验、HookBus 调用点、结构化审计、migration 边界和测试矩阵。
- 历史版本能力已并入当前分支；需要追溯时再阅读对应 Release Notes。
- 下一步以 `docs/PROJECT_PROGRESS.md` 的“当前未完成 / 风险 / 下一步”为准，不要从旧 Roadmap 推断当前任务。
