# DevHub

DevHub 是一个 **Core + 插件** 的开源服务底座。Core 提供账号、权限、内容基础模型、分类、标签、评论、通知、后台治理、API、SEO、安全边界和插件生命周期等稳定基础能力；插件系统承载业务扩展、前后台页面扩展、Hook、配置、内容类型、第三方集成和垂直功能扩展。

DevHub 提供默认社区能力，但长期目标不是单一社区程序，而是面向社区、内容站、知识库、内部工具平台和垂直业务系统等场景的可扩展服务底座。插件系统是 DevHub 长期架构的核心组成部分，不是附属功能。

当前版本口径：`v1.8.3`，主题为“后台插件治理稳定性、声明型插件业务闭环与 external_service non-blocking Webhook 收口”。当前仓库 `VERSION` 为最终版本来源。

当前只维护两个入口：

```text
前台：http://127.0.0.1:8090/
后台：http://127.0.0.1:8090/admin-next
```

旧的原生 HTML / CSS / JS 前台和后台已经移除。`/admin`、`/admin/:site` 只保留为兼容重定向。

## 文档导航

所有文档的大纲和跳转入口见：[docs/README.md](docs/README.md)。

常用文档：

- [AGENT_RULES.md](docs/AGENT_RULES.md)
- [PROJECT_PROGRESS.md](docs/PROJECT_PROGRESS.md)
- [API.md](docs/API.md)
- [TESTING.md](docs/TESTING.md)
- [DEPLOYMENT.md](docs/DEPLOYMENT.md)
- [BACKUP_AND_ROLLBACK.md](docs/BACKUP_AND_ROLLBACK.md)
- [SEO.md](docs/SEO.md)
- [PLUGIN_ARCHITECTURE.md](docs/PLUGIN_ARCHITECTURE.md)
- [PLUGIN_RUNTIME_MODEL.md](docs/PLUGIN_RUNTIME_MODEL.md)
- [docs/releases/v1.8.3.md](docs/releases/v1.8.3.md)
- [docs/releases/v1.8.2.md](docs/releases/v1.8.2.md)
- [docs/releases/v1.7.2.md](docs/releases/v1.7.2.md)
- [docs/releases/v1.7.1.md](docs/releases/v1.7.1.md)
- [docs/releases/v1.7.0.md](docs/releases/v1.7.0.md)
- [docs/releases/v1.6.0.md](docs/releases/v1.6.0.md)
- [docs/releases/v1.5.0.md](docs/releases/v1.5.0.md)
- [docs/releases/v1.4.0.md](docs/releases/v1.4.0.md)
- [docs/releases/v1.3.5.md](docs/releases/v1.3.5.md)
- [CHANGELOG.md](CHANGELOG.md)

## 当前能力

- Core 通用能力：用户账号、前台 / 后台认证、角色权限、子站、板块、通用内容、分类、标签、评论、搜索、通知、关注、收藏、举报、后台管理基础框架、API、审计、SEO、安全边界和插件生命周期 / 插件包治理。
- 内置系统插件：`qa` 提供 `question`，`docs` 提供 `document`，`wiki` 提供 `wiki_page`；`projects`、`jobs`、`ai_works` 已接管 `project`、`job`、`ai_work` 的插件归属。
- 插件扩展能力：当前支持 manifest 声明、内容类型、权限、菜单、Hook、配置 schema、external_service 受控 health check 预备、插件包治理、远程索引只读镜像、staging 下载、安全预检、compat-check、安装 / 启用 / 软卸载 / 升级治理和签名验签；第三方插件运行模型仍在设计中，仍不执行第三方代码或开放动态加载。
- 插件后台治理体验：当前插件列表、详情抽屉、Webhook 治理、插件包治理、上传记录、远程插件包、审批中心、操作历史、配置版本历史、可信发布者和官方公告插件预览已按功能域与页内 Tab 做中文化整理；不改变底层生命周期、Webhook 协议或 Secret / Token 安全模型。
- 插件状态：支持全局插件状态 `plugins.status` 和子站插件状态 `community_plugins.status`；禁用插件只影响新发布、导航、菜单和管理入口，不影响历史内容详情 SEO。
- 兼容内容类型：`article`、`news` 等仍作为 Core 兼容内容类型存在；`project`、`job`、`ai_work` 已完成插件归属迁移，但专属扩展表和完整业务闭环仍留到后续版本。
- 内容与互动：Topic 列表、详情、发布、编辑、删除、浏览数、点赞、收藏、关注、评论、问答采纳、用户中心和通知中心。
- 标签治理：支持标签 SEO 页、标签聚合、标签关注、发布页标签建议、后台标签 CRUD、启用 / 禁用 / 合并、标签别名、统计重算、审计和 sitemap / canonical 治理。
- 治理与后台：支持举报、版主子站范围治理、内容治理、评论治理、子站管理、板块管理、系统插件管理和审计日志。
- 存储模式：MemoryStore 与 MySQLStore。

## 当前定位

DevHub 当前定位为 **Core + 插件 的开源服务底座**。Core 保持稳定、克制和通用，不承载过多垂直业务；插件扩展业务能力，但不能绕过 Core 的权限、安全、审计和生命周期治理。

当前 `v1.8.3` 重点是后台插件治理稳定性、中文体验、声明型插件从安装到使用的真实闭环、插件包 upload -> promote -> install 验收、PluginRegistry 运行态刷新，以及 external_service non-blocking Webhook 投递闭环。完整第三方代码执行、远程动态加载、远程 iframe、插件市场和 blocking Hook 仍未开放，详见 [插件系统路线图](docs/PLUGIN_SYSTEM_ROADMAP.md) 与 [v1.8.3 Release Notes](docs/releases/v1.8.3.md)。

当前范围和限制以 [v1.8.3 Release Notes](docs/releases/v1.8.3.md) 为准，长期滚动状态见 [项目进度](docs/PROJECT_PROGRESS.md)。

历史版本说明见 `docs/README.md` 的“历史版本归档”。

## Roadmap

- P0：持续收口真实插件包与声明型插件“安装到使用”闭环，确保 upload、precheck、promote、本地仓库、install dry-run、install、enable、community enable、菜单、content_type、权限、配置和阻断规则可验收。
- P1：继续完善后台插件治理中文体验、状态提示、审计、运行记录和真实 fixture 回归，避免只靠 mock 数据判断插件能力。
- P2：收口 external_service non-blocking Webhook、health warning / error、retry、skipped 和 token 不泄露边界。
- P3：继续沉淀官方声明型插件样例，例如公告、友情链接、SEO 扩展或统计代码插件。
- P4：在插件包治理、运行模型和官方插件验证稳定后，再推进远程分发、插件市场和第三方生态。

## 目录结构

```text
.
├── main.go                         # Go 服务入口
├── dev.sh                          # 一键启动 / 停止脚本
├── internal/
│   ├── domain/                     # 业务模型、请求和响应结构
│   ├── store/                      # MemoryStore / MySQLStore / schema
│   ├── service/                    # 应用服务层
│   └── transport/httpapi/          # Gin 路由和 HTTP 处理器
├── db/mysql/                       # MySQL schema 与迁移草案
├── web/
│   ├── frontend-app/               # Astro + Vue Islands 前台源码
│   ├── frontend/                   # 前台构建产物，由 Go 托管
│   ├── admin-app/                  # Vue 3 + Element Plus 后台源码
│   └── admin-vue/                  # 后台构建产物，由 Go 托管
├── docs/
│   ├── README.md                   # 文档大纲
│   ├── PROJECT_PROGRESS.md         # 当前进度、迭代结果和下一步
│   ├── API.md                      # 当前真实 API 路径、参数、响应和限制
│   ├── TESTING.md                  # 手工验收和回归测试清单
│   ├── DEPLOYMENT.md               # 启动、部署和 8090 排障
│   ├── BACKUP_AND_ROLLBACK.md      # 备份、恢复和紧急回滚
│   ├── SEO.md                      # 百度 SEO 保护要求
│   ├── releases/                   # 版本归档说明
│   └── archive/                    # 历史规划归档说明
├── VERSION                         # 当前归档版本
├── CHANGELOG.md                    # 版本变化记录
└── docs/archive/                   # 历史任务原文和旧规划归档
```

## 一键启动与停止

推荐始终使用仓库根目录的一键脚本：

```bash
./dev.sh
```

脚本默认固定端口为 `8090`，默认使用内存仓库，适合快速预览。它会自动构建 Astro 前台和 Vue 后台；Node 构建优先使用本机 `npm`，没有本机 `npm` 时使用 Docker Node。Go 后端默认统一使用 Docker Go 镜像，当前用户无 Docker 权限时，脚本会自动尝试 `sudo docker`。

常用命令：

```bash
./dev.sh                         # 一键启动，等同于 ./dev.sh start
./dev.sh stop                    # 一键停止
./dev.sh restart --no-build      # 一键重启，跳过前后台构建
./dev.sh --local-go restart --no-build
                                 # 临时使用宿主机 Go 启动后端
./dev.sh status                  # 查看 8090 当前状态
./dev.sh start --mysql           # 启动 MySQL 并使用 MySQLStore
DOCKER="sudo docker" ./dev.sh    # 也可以显式指定 sudo docker
```

端口策略：

- 默认端口固定为 `8090`，不会每次自动新开端口。
- 如果 `8090` 已经是 DevHub 服务，脚本会复用当前服务并跳过启动新的 Go 进程。
- 如果改了 Go 后端代码，使用 `./dev.sh restart --no-build` 让脚本重启服务。
- 如果想明确停止服务，使用 `./dev.sh stop`。
- 如果端口被其他程序占用，脚本会打印 `lsof`、`kill`、`docker ps`、`docker stop` 的排查命令。

Go 依赖下载默认使用国内代理：

```text
GOPROXY=https://goproxy.cn,direct
GOSUMDB=sum.golang.google.cn
```

需要换代理时：

```bash
GOPROXY=https://goproxy.io,direct ./dev.sh
```

## 数据模式

`dev.sh` 默认使用内存模式：

```bash
CMS_STORE=memory ./dev.sh
```

使用 MySQL 模式：

```bash
./dev.sh --mysql
```

MySQL 开发库配置：

```text
Host: 127.0.0.1
Port: 3307
User: devhub
Password: Devhub_123456
Database: devhub
```

应用直接 `go run .` 时默认监听 `8090`；如果未设置 `CMS_STORE=memory`，会默认连接 MySQL。推荐本地开发仍通过 `./dev.sh` 启动，避免端口、构建产物和数据模式不一致。

## 页面入口

```text
/                       前台总站
/c/php/                 PHP 子站 canonical 首页，Go 动态输出 SEO HTML
/site/php/              PHP 子站兼容入口，301 到 /c/php/
/search/                搜索页
/topics/new/            发布 Topic
/c/:site/topics/new/    子站发布 Topic
/topics/:id/            Topic 详情，Go 动态输出 SEO HTML
/posts/:id/             兼容入口，301 跳转到 /topics/:id/
/tags/:tag/             标签聚合页，Go 动态输出 SEO HTML
/c/:site/tags/:tag/     子站标签聚合页，Go 动态输出 SEO HTML
/me/favorites           我的收藏
/me/follows             我的关注
/me/activities          我的动态
/notifications          通知中心
/me/notifications       通知中心别名
/moderator              独立版主工作台
/moderator/reports      版主举报处理
/moderator/topics       版主内容治理
/moderator/comments     版主评论治理
/moderator/audit-logs   版主审计日志
/admin-next             当前后台
/admin-next/content     内容管理
/admin-next/comments    评论管理
/admin-next/reports     举报管理
/admin-next/communities 子站管理
/admin-next/tags        标签管理
/admin-next/moderators  版主管理
/admin-next/audit-logs  治理审计日志
/admin-next/...         后台前端路由
```

兼容入口：

```text
/admin                  重定向到 /admin-next
/admin/:site            重定向到 /admin-next?site=:site
```

## API 概览

健康检查：

```bash
curl http://127.0.0.1:8090/api/v1/health
```

兼容 API：

```text
GET    /api/v1/sites
GET    /api/v1/boards
GET    /api/v1/posts
GET    /api/v1/posts/:id
POST   /api/v1/posts
PUT    /api/v1/posts/:id
DELETE /api/v1/posts/:id
GET    /api/v1/search
GET    /api/v1/hot
GET    /api/v1/tags
GET    /api/v1/posts/:id/comments
```

通用社区 API：

```text
GET    /api/v1/communities
GET    /api/v1/communities/:slug
GET    /api/v1/communities/:slug/home
GET    /api/v1/communities/:slug/stats
GET    /api/v1/communities/:slug/categories
GET    /api/v1/communities/:slug/tags
GET    /api/v1/communities/:slug/moderators
GET    /api/v1/topics
GET    /api/v1/topics/:id
POST   /api/v1/topics
PUT    /api/v1/topics/:id
DELETE /api/v1/topics/:id
GET    /api/v1/topics/:id/comments
POST   /api/v1/topics/:id/comments
POST   /api/v1/topics/:id/comments/:commentId/replies
POST   /api/v1/topics/:id/comments/:commentId/accept
GET    /api/v1/search/topics
POST   /api/v1/reports
POST   /api/v1/topics/:id/like
POST   /api/v1/topics/:id/favorite
POST   /api/v1/reactions/toggle
POST   /api/v1/favorites/toggle
POST   /api/v1/follows/toggle
GET    /api/v1/activities
GET    /api/v1/me/favorites
GET    /api/v1/me/follows
GET    /api/v1/me/activities
GET    /api/v1/me/notifications
POST   /api/v1/me/notifications/:id/read
POST   /api/v1/me/notifications/read-all
GET    /api/v1/notifications
POST   /api/v1/notifications/:id/read
POST   /api/v1/notifications/read-all
GET    /api/v1/tags
GET    /api/v1/tags/hot
GET    /api/v1/tags/suggestions
GET    /api/v1/tags/suggest
GET    /api/v1/tags/by-slug/:tag
GET    /api/v1/tags/:tag
GET    /api/v1/tags/:tag/topics
GET    /api/v1/communities/:slug/tags/:tag
GET    /api/v1/communities/:slug/tags/:tag/topics
```

说明：真实 API 路径、响应字段、认证要求和当前限制以 [docs/API.md](docs/API.md) 为准。`GET /api/v1/search/topics?sort=unsolved` 当前只返回未解决问答，`sort=featured` 当前只返回精华内容。

认证与后台 API：

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
POST /api/v1/admin/login
POST /api/v1/admin/refresh
POST /api/v1/admin/logout
GET  /api/v1/admin/me
GET  /api/v1/admin/overview
GET  /api/v1/admin/communities
POST /api/v1/admin/communities
GET  /api/v1/admin/communities/:id
PUT  /api/v1/admin/communities/:id
POST /api/v1/admin/communities/:id/enable
POST /api/v1/admin/communities/:id/disable
POST /api/v1/admin/communities/reorder
GET  /api/v1/admin/communities/:id/categories
POST /api/v1/admin/communities/:id/categories
PUT  /api/v1/admin/categories/:id
POST /api/v1/admin/categories/:id/enable
POST /api/v1/admin/categories/:id/disable
POST /api/v1/admin/categories/reorder
GET  /api/v1/admin/posts
POST /api/v1/admin/posts
PUT  /api/v1/admin/posts/:id
DELETE /api/v1/admin/posts/:id
GET  /api/v1/admin/comments
GET  /api/v1/admin/reports
GET  /api/v1/admin/reports/:id
POST /api/v1/admin/reports/:id/handle
POST /api/v1/admin/reports/batch-handle
GET  /api/v1/admin/moderators
POST /api/v1/admin/moderators
PUT  /api/v1/admin/moderators/:id
DELETE /api/v1/admin/moderators/:id
POST /api/v1/admin/topics/:id/feature
POST /api/v1/admin/topics/:id/pin
POST /api/v1/admin/topics/:id/hide
POST /api/v1/admin/topics/:id/restore
POST /api/v1/admin/topics/:id/lock-comments
POST /api/v1/admin/topics/:id/unlock-comments
POST /api/v1/admin/topics/batch
POST /api/v1/admin/comments/:id/hide
POST /api/v1/admin/comments/:id/restore
POST /api/v1/admin/comments/batch
GET  /api/v1/admin/audit-logs
GET  /api/v1/admin/sites
GET  /api/v1/admin/users
GET  /api/v1/admin/roles
GET  /api/v1/admin/permissions
GET  /api/v1/admin/tags
POST /api/v1/admin/tags
GET  /api/v1/admin/tags/:id
PUT  /api/v1/admin/tags/:id
GET  /api/v1/admin/tags/:id/topics
POST /api/v1/admin/tags/:id/enable
POST /api/v1/admin/tags/:id/disable
GET  /api/v1/admin/settings
```

版主工作台 API：

```text
GET  /api/v1/moderator/communities
GET  /api/v1/moderator/dashboard
GET  /api/v1/moderator/reports
POST /api/v1/moderator/reports/:id/handle
GET  /api/v1/moderator/topics
POST /api/v1/moderator/topics/:id/feature
POST /api/v1/moderator/topics/:id/unfeature
POST /api/v1/moderator/topics/:id/pin
POST /api/v1/moderator/topics/:id/unpin
POST /api/v1/moderator/topics/:id/hide
POST /api/v1/moderator/topics/:id/restore
POST /api/v1/moderator/topics/:id/lock-comments
POST /api/v1/moderator/topics/:id/unlock-comments
GET  /api/v1/moderator/comments
POST /api/v1/moderator/comments/:id/hide
POST /api/v1/moderator/comments/:id/restore
GET  /api/v1/moderator/audit-logs
```

说明：`/api/v1/moderator/*` 使用前台 user token，并强制校验 `community_moderators` 子站授权；普通用户和跨子站访问返回 403。

说明：`/admin-next/communities`、`/admin-next/moderators`、`/admin-next/content`、`/admin-next/comments`、`/admin-next/reports` 和 `/admin-next/audit-logs` 已接入当前真实后台 API。子站配置、子站板块、批量 Topic / Comment 治理和批量举报处理都会写入 `admin_logs`；审计接口返回的 `actor_user_id`、`target_type`、`target_id`、`community_id` 为基于当前文本日志的派生字段。

后台开发种子账号：

```text
admin / admin123       super_admin，全站
operator / admin123    site_admin，仅 PHP
auditor / admin123     moderator，仅 Go
```

## 前台说明

前台源码在 `web/frontend-app`。Astro 输出静态 HTML，Vue Islands 负责登录、点赞、评论等局部交互。

前台构建默认使用 fallback 数据，保证没有后端 API 时也能构建成功。需要用运行中的 API 生成静态页面时：

```bash
cd web/frontend-app
FRONTEND_API_BASE=http://127.0.0.1:8090/api/v1 \
FRONTEND_SITE_URL=http://127.0.0.1:8090 \
npm run build
```

## 后台说明

后台源码在 `web/admin-app`，构建产物在 `web/admin-vue`。当前后台是唯一维护的后台前端，旧原生后台已经移除。

后台基于 `/admin-next` 的 Vue Router 运行，Go 的 `NoRoute` 会把 `/admin-next/...` 交给后台前端处理。

## 开发检查

Codex / Agent 默认不要在完成任务时自动跑测试；测试由开发者手动执行，避免拖慢日常改动。需要一键验收时运行：

```bash
./scripts/test-all.sh
```

常用检查命令：

```bash
bash -n dev.sh
GOCACHE=/tmp/go-build go test ./...
go build -o .devhub/devhub .
cd web/frontend-app && npm run build
cd web/admin-app && npm run build
```

当前项目进度和下一步见：[docs/PROJECT_PROGRESS.md](docs/PROJECT_PROGRESS.md)。

## Git 版本管理

当前仓库使用 `main` 作为默认分支。建议每轮迭代按功能创建短分支，例如：

```bash
git checkout -b feature/comments-qa
```

版本库纳入源码、schema、迁移、文档和 `package-lock.json`；以下内容不入库：

- `node_modules/`
- `web/frontend/`
- `web/admin-vue/`
- `.devhub/`
- `.agents/`、`.codex/`、`.claude/`
- 本地 `.env*`、日志、PID、构建缓存

提交前建议执行：

```bash
go test ./...
go build -o .devhub/devhub .
cd web/frontend-app && npm run build
cd web/admin-app && npm run build
git status
```

本地没有 `npm` 时，可使用 `dev.sh` 或 Docker Node 构建；构建产物由脚本生成，不需要提交。

v1.3.4 归档建议命令：

```bash
git status
git diff
git add .
git commit -m "chore: release DevHub v1.3.4"
git tag v1.3.4
git push origin main
git push origin v1.3.4
```

打 tag 前必须先确认工作区没有未审阅差异，且测试矩阵通过。

## 项目状态

README 只保留项目定位、入口和当前能力概览；当前仍未完成项、风险和下一步统一维护在 [docs/PROJECT_PROGRESS.md](docs/PROJECT_PROGRESS.md)，版本范围限制见对应 release notes。
