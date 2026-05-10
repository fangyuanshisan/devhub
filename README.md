# DevHub

DevHub 是一个多子站技术社区 CMS。当前项目使用 Go + Gin 提供后端 API 与静态资源托管，前台使用 Astro + Vue Islands，后台使用 Vue 3 + Element Plus。

当前版本：`v1.3.0`，版本主题为“Core + Plugins 架构拆分版”。

当前只维护两个入口：

```text
前台：http://127.0.0.1:8090/
后台：http://127.0.0.1:8090/admin-next
```

旧的原生 HTML / CSS / JS 前台和后台已经移除。`/admin`、`/admin/:site` 只保留为兼容重定向。

## 文档导航

所有文档的大纲和跳转入口见：[docs/README.md](docs/README.md)。

常用文档：

- [Codex / AI Agent 固定规则](docs/AGENT_RULES.md)
- [项目进度](docs/PROJECT_PROGRESS.md)
- [API 文档](docs/API.md)
- [测试文档](docs/TESTING.md)
- [部署启动文档](docs/DEPLOYMENT.md)
- [备份与回滚](docs/BACKUP_AND_ROLLBACK.md)
- [SEO 文档](docs/SEO.md)
- [插件架构说明](docs/PLUGIN_ARCHITECTURE.md)
- [v1.3.0 Release Notes](docs/releases/v1.3.0.md)
- [v1.2.1 Release Notes](docs/releases/v1.2.1.md)
- [v1.2.0 Release Notes](docs/releases/v1.2.0.md)
- [v1.1.5 Release Notes](docs/releases/v1.1.5.md)
- [v1.1.4 Release Notes](docs/releases/v1.1.4.md)
- [v1.1.3 Release Notes](docs/releases/v1.1.3.md)
- [v1.1.1 Release Notes](docs/releases/v1.1.1.md)
- [v1.1.0 Release Notes](docs/releases/v1.1.0.md)
- [v1.0.0 Release Notes](docs/releases/v1.0.0.md)
- [变更日志](CHANGELOG.md)
- [需求原文](更新.md)

## 当前能力

- 多子站：总站、PHP、Go、Java、AI、Frontend；v1.1.0 起子站具备独立首页、SEO、配置、板块、版主、统计、关注和公告。
- 多板块：社区、问答中心、开源项目、AI 作品、招聘内推、Wiki、文档；后台支持按子站管理板块、启用 / 禁用、排序和导航展示。
- 插件架构：v1.3.0 起 Core 保留通用能力，问答、文档、Wiki 由 `qa`、`docs`、`wiki` 三个内置系统插件注册内容类型、菜单、权限和路由描述。
- 内容：列表、详情、发布、编辑、删除、浏览数、点赞、收藏、关注、标签、热门排序。
- 标签：v1.2.1 起支持标签详情 SEO 页、标签下内容聚合、标签关注、发布页标签建议、后台标签 CRUD、启用 / 禁用 / 合并、标签别名、标签统计重算、治理审计和 sitemap / canonical 治理。
- 通用 Topic：已支持 `article`、`question`、`project`、`ai_work`、`job`、`wiki_page`、`document`、`news` 等内容类型，并兼容旧 `wiki` / `doc` 参数。
- 搜索：支持全站、子站、板块、关键词、标签筛选，并支持 `sort=unsolved` 未解决问答筛选。
- 评论：支持 Topic 评论列表、加载更多、发表评论、回复评论、问答采纳和最佳答案展示；评论点赞仍仅保留旧兼容接口。
- 互动：MemoryStore / MySQLStore 均支持 Topic 点赞、收藏、关注、我的收藏、我的关注、我的动态、通知中心、评论动态和评论通知。
- 治理：支持举报 Topic / Comment，后台举报处理，版主子站范围管理，精华、置顶、隐藏、恢复、评论锁定、评论隐藏。
- 用户与权限：前台 `users`、后台 `admin_users`、子站版主 `community_moderators` 边界已整理；前台 / 后台 token 分离，后台 RBAC 与版主子站范围治理可用。
- 版主工作台：`/moderator` 提供独立轻量工作台，版主可处理自己子站的举报、主题、评论和审计日志。
- 后台：控制台、内容管理、举报管理、评论审核、子站管理、子站板块管理、版主管理、用户权限、运营工具、数据统计、系统设置。
- 存储：支持内存模式和 MySQL 模式。

## v1.3.0 定位

DevHub v1.3.0 是“Core + Plugins 架构拆分版”。本版本新增 `plugins` 表、插件注册定义、插件状态 API 和后台插件入口；`topics` 作为兼容实现中的 Core 内容表新增 `plugin_code`，`categories` 作为 Core 板块表新增 `plugin_code` / `allowed_content_types`；问答、文档、Wiki 分别迁移为 `qa`、`docs`、`wiki` 内置系统插件。

本版本不做插件市场、插件压缩包上传安装、远程更新或复杂动态模块加载。

## v1.2.1 定位

DevHub v1.2.1 是“标签合并、别名与统计重算版”。本版本在 v1.2.0 的标签页、标签 SEO、标签关注、后台标签 CRUD 和 sitemap 收录基础上，补齐标签别名、标签合并、标签统计重算、alias / merged SEO 处理、后台标签治理审计和 admin-next 标签治理入口。

本版本暂不做标签趋势统计、标签运营分析看板、大规模异步统计任务和 AI 推荐标签。

## v1.2.0 定位

DevHub v1.2.0 是“标签系统增强版”。本版本新增 `/tags/:tag/` 全站标签 SEO 页、`/c/:communitySlug/tags/:tag/` 子站标签 SEO 页、标签详情 API、标签内容聚合、标签关注、发布页标签建议、后台标签 CRUD、标签启用 / 禁用、标签 SEO 字段、标签关联内容查看和 sitemap 标签收录。

本版本不做标签趋势统计、标签运营分析、大规模异步统计任务和 AI 推荐标签；这些能力留到后续版本。

## v1.1.5 定位

DevHub v1.1.5 是“前台 UI 美化专项”。本补丁已并入当前 v1.3.0 工作分支，只优化前台全局视觉、导航、首页、子站页、Topic 列表、Topic 详情、搜索页、发布页、“我的”页面、版主入口和移动端响应式样式；不修改 API、Store、数据库、路由、鉴权、业务逻辑或 Go 动态 SEO 结构。

## v1.1.4 定位

DevHub v1.1.4 是“前台登录态与权限入口修复版”。本补丁已并入当前 v1.3.0 工作分支，修复前台登录状态恢复、子站关注和“我的”类页面误判未登录、普通会员误见总后台入口、版主工作台入口、发布 `question` 板块匹配，以及后台子站入口重复问题。

## v1.1.3 定位

DevHub v1.1.3 是“独立版主工作台 MVP”。本版本新增 `/moderator`、`/moderator/reports`、`/moderator/topics`、`/moderator/comments`、`/moderator/audit-logs`，让子站版主使用前台 `users` 登录态和 `community_moderators` 授权关系治理自己负责的子站。

版主工作台复用现有举报、Topic、Comment 治理能力和 `admin_logs`，但通过 `/api/v1/moderator/*` 做专用权限入口。普通用户不能访问，跨子站治理返回 403，复杂 RBAC 和版主任期 / 绩效统计留到后续。

## v1.1.1 定位

DevHub v1.1.1 是“前后台身份边界整理版”。本版本明确 `users`、`admin_users`、`community_moderators` 三类身份：前台用户负责社区行为，后台人员负责后台管理，子站版主通过前台用户身份获得指定子站的治理权限。

前台登录态和后台登录态已经分离。前台推荐使用 `devhub_user_token` / `devhub_user_refresh_token`，后台使用 `devhub_admin_token` / `devhub_admin_refresh_token`；普通前台 token 不能访问后台特权接口，后台 admin token 也不会被当作前台用户身份。

## v1.1.0 定位

DevHub v1.1.0 是“子站模块增强版”。本版本把子站从“内容筛选维度”升级为“独立社区空间”：每个启用子站都有 `/c/:slug/` 首页、Go 动态 SEO HTML、独立配置、独立板块、版主展示、统计、关注按钮和公告区域。

DevHub v1.0.0 仍是第一个可运行大版本归档；v1.1.0 在不改变前台 `/`、后台 `/admin-next`、默认端口 `8090` 和 `/topics/:id` SEO 动态详情页的前提下增强子站模块。

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
└── 更新.md                         # 本轮产品需求原文
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

说明：第五轮互动、第六轮评论 / 采纳、第七轮举报 / 治理接口的真实路径、响应字段和部分完成项以 [docs/API.md](docs/API.md) 为准。`GET /api/v1/search/topics?sort=unsolved` 当前只返回未解决问答，`sort=featured` 当前只返回精华内容。

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

v1.3.0 归档建议命令：

```bash
git status
git diff
git add .
git commit -m "chore: release DevHub v1.3.0"
git tag v1.3.0
git push origin main
git push origin v1.3.0
```

打 tag 前必须先确认工作区没有未审阅差异，且测试矩阵通过。

## 已知限制

- 子站自定义导航仍使用“启用板块生成默认导航”的方式，深度自定义导航留到后续版本。
- 版主工作台是 MVP，不包含复杂 RBAC、权限点矩阵、版主任期或绩效统计。
- MySQL refresh token 仍通过 `token_type` 区分前台用户和后台人员，并已移除单一 `users` 外键；后续生产化 migration 可进一步拆分字段命名。
- 后台人员参与前台社区互动时仍应拥有独立 `users` 身份，admin-user 绑定关系留到后续。
- 标签详情 SEO 页和标签后台管理已在 v1.2.0 完成；标签合并 / 别名和统计重算已在 v1.2.1 完成，趋势统计仍留到后续版本。
- 评论点赞未纳入 v1.1.0 主线。
- 问答支持采纳和更换最佳答案，暂不支持取消已解决状态。
- 标签关注已在 v1.2.0 接入标签页；用户关注前台入口仍可继续增强，完整关注流留到后续版本。
- `/sitemap.xml` 目前动态输出但未做大规模分片。
- 生产部署仍需按实际环境配置进程守护、反向代理、HTTPS、日志轮转和定时备份。

## Roadmap

- v1.3.0：推荐、关注流和内容发现。
- v1.4.0：用户成长、声望和个人主页。
- v1.5.0：后台运营、治理和数据统计增强。
- v1.6.0：生产化、migration、性能和 CI/CD。
