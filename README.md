# DevHub

DevHub 是一个多子站技术社区 CMS。当前项目使用 Go + Gin 提供后端 API 与静态资源托管，前台使用 Astro + Vue Islands，后台使用 Vue 3 + Element Plus。

当前只维护两个入口：

```text
前台：http://127.0.0.1:8090/
后台：http://127.0.0.1:8090/admin-next
```

旧的原生 HTML / CSS / JS 前台和后台已经移除。`/admin`、`/admin/:site` 只保留为兼容重定向。

## 文档导航

所有文档的大纲和跳转入口见：[docs/README.md](docs/README.md)。

常用文档：

- [项目进度](docs/PROJECT_PROGRESS.md)
- [API 文档](docs/API.md)
- [测试文档](docs/TESTING.md)
- [部署启动文档](docs/DEPLOYMENT.md)
- [SEO 文档](docs/SEO.md)
- [需求原文](更新.md)

## 当前能力

- 多子站：总站、PHP、Go、Java、AI、Frontend。
- 多板块：社区、问答中心、开源项目、AI 作品、招聘内推、Wiki、文档。
- 内容：列表、详情、发布、编辑、删除、浏览数、点赞、收藏、关注、标签、热门排序。
- 通用 Topic：已支持 `article`、`question`、`project`、`ai_work`、`job`、`wiki`、`doc`、`news` 等内容类型。
- 搜索：支持全站、子站、板块、关键词、标签筛选，并支持 `sort=unsolved` 未解决问答筛选。
- 评论：支持 Topic 评论列表、加载更多、发表评论、回复评论、问答采纳和最佳答案展示；评论点赞仍仅保留旧兼容接口。
- 互动：MemoryStore / MySQLStore 均支持 Topic 点赞、收藏、关注、我的收藏、我的关注、我的动态、通知中心、评论动态和评论通知。
- 治理：支持举报 Topic / Comment，后台举报处理，版主子站范围管理，精华、置顶、隐藏、恢复、评论锁定、评论隐藏。
- 用户与权限：前台登录、注册、会话恢复、refresh token 刷新、退出登录、JWT 会话、后台 RBAC 与站点范围上下文。
- 后台：控制台、内容管理、举报管理、评论审核、站点管理、用户权限、运营工具、数据统计、系统设置。
- 存储：支持内存模式和 MySQL 模式。

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
│   ├── SEO.md                      # 百度 SEO 保护要求
│   └── archive/                    # 历史规划归档说明
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

应用直接 `go run .` 时，如果未设置 `CMS_STORE=memory`，会默认连接 MySQL。推荐本地开发仍通过 `./dev.sh` 启动，避免端口、构建产物和数据模式不一致。

## 页面入口

```text
/                       前台总站
/site/php/              PHP 子站
/c/php/                 PHP 子站别名
/search/                搜索页
/topics/new/            发布 Topic
/c/:site/topics/new/    子站发布 Topic
/topics/:id/            Topic 详情，Go 动态输出 SEO HTML
/posts/:id/             兼容入口，301 跳转到 /topics/:id/
/tags/:tag/             标签聚合页
/me/favorites           我的收藏
/me/follows             我的关注
/me/activities          我的动态
/notifications          通知中心
/me/notifications       通知中心别名
/admin-next             当前后台
/admin-next/reports     举报管理
/admin-next/moderators  版主管理
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
GET    /api/v1/communities/:slug/categories
GET    /api/v1/communities/:slug/tags
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
GET  /api/v1/admin/overview
GET  /api/v1/admin/posts
POST /api/v1/admin/posts
PUT  /api/v1/admin/posts/:id
DELETE /api/v1/admin/posts/:id
GET  /api/v1/admin/comments
GET  /api/v1/admin/reports
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
GET  /api/v1/admin/users
GET  /api/v1/admin/settings
```

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
cd web/frontend-app && npm run build
cd web/admin-app && npm run build
git status
```

本地没有 `npm` 时，可使用 `dev.sh` 或 Docker Node 构建；构建产物由脚本生成，不需要提交。
