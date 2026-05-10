# DevHub 部署与启动文档

[返回文档大纲](README.md)

更新时间：2026-05-10

本文档记录 DevHub 当前真实启动方式、端口约定和本地排障流程。项目名称保持 DevHub，正式本地入口保持：

```text
前台：http://127.0.0.1:8090/
后台：http://127.0.0.1:8090/admin-next
```

## 推荐启动方式

优先使用仓库根目录的脚本：

```bash
./dev.sh
./dev.sh --restart
./dev.sh restart --no-build
./dev.sh --local-go --restart
./dev.sh status
./dev.sh stop
```

默认端口为 `8090`，默认数据仓库为 `CMS_STORE=memory`。`main.go` 直接运行也默认监听 `8090`，仍可通过 `PORT=8080 go run .` 覆盖。修改 Go 后端代码后需要重启服务；只改 Go 且前后台产物无需重建时，可以使用 `./dev.sh restart --no-build` 或 `./dev.sh --local-go restart --no-build`。

### 开发模式

内存模式：

```bash
CMS_STORE=memory ./dev.sh --restart
```

MySQL 模式：

```bash
./dev.sh --mysql --restart
```

停止和状态检查：

```bash
./dev.sh --stop
./dev.sh --status
```

## 构建行为

- 前台 Astro 和后台 Vue 构建优先使用本机 `npm`。
- 如果本机没有 `npm`，`dev.sh` 会尝试使用 Docker Node 构建前后台产物。
- Go 后端可通过脚本启动，也可以在排障时先 `go build` 产出二进制再启动。
- 不要把临时代理、私钥或 token 写入代码和文档；需要代理时通过环境变量传入。

常用环境变量：

```text
PORT=8090
CMS_STORE=memory|mysql
DB_HOST=127.0.0.1
DB_PORT=3307
DB_USER=devhub
DB_PASSWORD=Devhub_123456
DB_NAME=devhub
```

## 前端构建

前台源码位于 `web/frontend-app`，构建产物输出到 `web/frontend`：

```bash
cd web/frontend-app
FRONTEND_SITE_URL=http://127.0.0.1:8090 npm run build
```

后台源码位于 `web/admin-app`，构建产物输出到 `web/admin-vue`：

```bash
cd web/admin-app
npm run build
```

`dev.sh` 会在启动前按需构建前台和后台；本机没有 `npm` 时，脚本会尝试使用 Docker Node 构建。CI 使用 `npm ci`，依赖 `web/frontend-app/package-lock.json` 和 `web/admin-app/package-lock.json`。

## Docker 和 MySQL 开发环境

`docker-compose.dev.yml` 提供 MySQL 8 开发库：

```text
Host: 127.0.0.1
Port: 3307
User: devhub
Password: Devhub_123456
Database: devhub
Volume: devhub_mysql_data
```

启动 MySQLStore：

```bash
./dev.sh --mysql --restart
```

脚本会启动 MySQL 容器，等待数据库可用，然后以 `CMS_STORE=mysql` 启动 Go 服务。手动初始化时使用：

```bash
mysql -h127.0.0.1 -P3307 -u devhub -p devhub < db/mysql/001_schema.sql
```

如果已经运行过旧版本，请先确认 `db/mysql/migrations/` 中的迁移是否已经在目标库执行。

## v1.1.0 Schema 升级说明

v1.1.0 增强了 `communities` 和 `categories` 表。新库可直接使用 `db/mysql/001_schema.sql` 初始化；旧库升级前建议先在预发环境验证。

新增或确认的 `communities` 字段：

```text
logo, cover_image, slogan, theme_color,
seo_title, seo_description, seo_keywords,
sort_order, status,
follower_count, topic_count, comment_count, hot_score,
announcement_title, announcement_content, announcement_url
```

新增或确认的 `categories` 字段：

```text
visible, nav_visible, postable, seo_title, seo_description, status
```

当前 Go 启动时会通过内置迁移辅助尽量补齐缺失列，并补齐 PHP、Go、Java、AI、Frontend 子站 seed 数据。生产库仍建议：

1. 先备份 MySQL。
2. 在预发库执行启动和接口回归。
3. 确认 `/api/v1/communities/php`、`/api/v1/communities/php/stats`、`/c/php/` 和 `/sitemap.xml` 正常。

## v1.1.1 身份边界升级说明

v1.1.1 增加或确认以下字段，用于区分前台用户、后台人员和子站版主审计来源：

```text
refresh_tokens.token_type
admin_logs.actor_type
admin_logs.actor_id
```

新库可直接使用 `db/mysql/001_schema.sql` 初始化；旧库启动时由内置迁移辅助尽量补齐。生产库建议先在预发环境确认：

- 前台登录返回 `token_type=user`。
- 后台登录返回 `token_type=admin`。
- 前台 refresh token 和后台 refresh token 互不刷新。
- 审计日志能返回 `actor_type` 和 `actor_id`。

说明：`refresh_tokens.user_id` 当前按 `token_type=user/admin` 分别指向 `users.id` 或 `admin_users.id`，v1.1.1 会尝试移除旧的单一 `users` 外键。生产库如果历史外键名称不同，需要 DBA 手动确认并处理。

## v1.1.3 版主工作台升级说明

v1.1.3 新增前台版主工作台静态页面和 `/api/v1/moderator/*` API，不新增 MySQL schema。升级时需要重新构建前台 Astro 产物，确保以下文件存在：

```text
web/frontend/moderator/index.html
web/frontend/moderator/reports/index.html
web/frontend/moderator/topics/index.html
web/frontend/moderator/comments/index.html
web/frontend/moderator/audit-logs/index.html
```

后端仍使用 `community_moderators` 判断版主范围，使用 `admin_logs` 写入 `actor_type=moderator` 的审计记录。生产升级后建议验证普通用户访问 `/api/v1/moderator/*` 返回 403，PHP 版主只能访问 PHP 子站数据。
4. 再切换生产服务。

## v1.2.0 标签系统升级说明

v1.2.0 增强了 `tags` 表和标签相关页面 / API。新库可直接使用 `db/mysql/001_schema.sql` 初始化；旧库启动时由内置迁移辅助尽量补齐缺失字段。

## v1.2.1 标签治理升级说明

v1.2.1 在 v1.2.0 基础上继续增强标签治理能力：

- `tags` 新增或确认 `merged_to_id`、`hot_score`，并增加 `status=merged` 语义。
- 新增 `tag_aliases` 表，用于标签别名解析与治理。

新库可直接使用 `db/mysql/001_schema.sql` 初始化；旧库升级时由内置迁移辅助尽量补齐缺失字段。建议升级前在预发环境验证 alias 冲突、merge 迁移与去重、统计重算与 sitemap 过滤。

新增或确认的 `tags` 字段：

```text
status VARCHAR(32) DEFAULT 'enable'
follower_count INT UNSIGNED DEFAULT 0
seo_title VARCHAR(255) DEFAULT ''
seo_description VARCHAR(500) DEFAULT ''
seo_keywords VARCHAR(500) DEFAULT ''
```

旧状态值会尽量转换：

- `1` / `enabled` / 空值 -> `enable`
- `0` / `disabled` -> `disable`

升级后需要重新构建前台和后台：

- 前台构建会移除旧 Astro 预生成标签详情页，`/tags/:tag/` 由 Go 动态输出。
- 后台构建会生成 `/admin-next/tags` 标签管理页面。

生产升级后建议验证：

```bash
curl -s http://127.0.0.1:8090/api/v1/tags
curl -s http://127.0.0.1:8090/api/v1/tags/suggestions?community_slug=php
curl -s http://127.0.0.1:8090/tags/laravel/
curl -s http://127.0.0.1:8090/sitemap.xml | rg '/tags/'
curl -s http://127.0.0.1:8090/topics/1/ | rg '<title>|description|<h1'
```

## 生产部署建议

- 使用 `go build -o devhub .` 产出二进制，配合 systemd、supervisor 或容器编排守护进程。
- 通过 `PORT=8090`、`CMS_STORE=mysql`、`DB_*` 环境变量传入配置，不要把生产密钥写入仓库。
- 使用 Nginx、Caddy 或云负载均衡作为反向代理，开启 HTTPS。
- 由 Go 托管 `web/frontend`、`web/admin-vue` 和 `/_astro` 等静态资源；反向代理不要屏蔽这些路径。
- 日志输出到文件或标准输出，并配置日志轮转。
- 定期备份 MySQL、配置、上传目录、二进制和静态构建产物。
- 上线后检查 `/topics/:id`、`/sitemap.xml`、`/robots.txt`，确保百度 SEO 动态详情页未退化。
- 回滚流程见 `docs/BACKUP_AND_ROLLBACK.md`。

## Go 模块网络检查

如果 `go run` 或脚本启动阶段出现模块解析卡住，先检查：

```bash
go env GOPROXY
go env GOPRIVATE
go env GONOSUMDB
go env GOSUMDB
```

第六轮收尾验收时，本机观测值为：

```text
GOPROXY=https://goproxy.cn,direct
GOPRIVATE=
GONOSUMDB=
GOSUMDB=sum.golang.org
```

仓库当前 `go.mod` / `go.sum` 未发现 Gitee 私有依赖。若看到残留的 `git-upload-pack` 子进程，通常是上一次 Go 或 Git 网络操作遗留；先清理残留进程，再优先使用已缓存依赖执行 `go build`。

## 端口和残留进程排查

```bash
lsof -i :8090 || true
ps -ef | grep -E 'devhub|go run|git-upload-pack|tmp/go-build' | grep -v grep || true
```

确认是残留进程后再清理：

```bash
kill <pid>
```

如果是 Docker 容器占用端口，使用 `docker ps` 查到容器后再 `docker stop <container>`。

## 二进制排障启动

当 `go run` 因模块解析或网络问题卡住，或脚本重启后遇到端口释放竞态时，可以临时使用二进制启动：

```bash
mkdir -p .devhub
go build -o .devhub/devhub .
PORT=8090 CMS_STORE=memory ./.devhub/devhub
```

需要后台常驻时：

```bash
repo_dir=$(pwd)
setsid -f bash -c "cd '$repo_dir' && PORT=8090 CMS_STORE=memory exec ./.devhub/devhub >> .devhub/server.log 2>&1"
```

启动后检查：

```bash
lsof -i :8090 || true
curl -I http://127.0.0.1:8090/
curl http://127.0.0.1:8090/api/v1/health
curl http://127.0.0.1:8090/api/v1/topics
```

停止二进制进程：

```bash
lsof -i :8090
kill <pid>
```

## 第六轮收尾启动结论

2026-05-09 收尾验收中，`./dev.sh --local-go --restart` 已完成前台和后台构建，但最终本地 Go 服务启动阶段没有稳定留下 8090 进程；随后 `go run` 链路出现 Gitee `git-upload-pack` 残留子进程。清理残留进程后，使用 `go build -o .devhub/devhub .` 产出二进制，并通过 `setsid` 后台启动，8090 已稳定响应。

## 第八轮构建补充

第八轮补丁修改了 Go 后端、`web/admin-app` 和文档。后台 Vite dev proxy 已统一指向 `http://127.0.0.1:8090`。本机无 `npm` 时，前后台构建可直接使用 Docker Node：

```bash
docker run --rm -e NPM_CONFIG_REGISTRY=https://registry.npmmirror.com -e FRONTEND_SITE_URL=http://127.0.0.1:8090 -v "$PWD/web/frontend-app:/app" -v "$PWD/web/frontend:/frontend" -w /app node:20-alpine sh -lc 'if [ ! -d node_modules ]; then npm install --registry=https://registry.npmmirror.com; fi; npm run build'
docker run --rm -e NPM_CONFIG_REGISTRY=https://registry.npmmirror.com -v "$PWD/web/admin-app:/app" -v "$PWD/web/admin-vue:/admin-vue" -w /app node:20-alpine sh -lc 'if [ ! -d node_modules ]; then npm install --registry=https://registry.npmmirror.com; fi; npm run build'
```

构建产物仍输出到 `web/admin-vue`，由 Go 在 `/admin-next` 挂载。修改 Go 后端后仍需执行 `./dev.sh --restart`，或先 `go build -o .devhub/devhub .` 再按二进制排障启动。

## v1.1.0 上线前检查

```bash
go test ./...
go build -o .devhub/devhub .
./dev.sh --restart
curl -I http://127.0.0.1:8090/
curl http://127.0.0.1:8090/api/v1/health
curl http://127.0.0.1:8090/api/v1/communities/php
curl http://127.0.0.1:8090/api/v1/communities/php/stats
curl http://127.0.0.1:8090/admin-next
curl http://127.0.0.1:8090/admin-next/communities
curl http://127.0.0.1:8090/c/php/
curl -I http://127.0.0.1:8090/site/php
curl http://127.0.0.1:8090/topics/1
curl http://127.0.0.1:8090/sitemap.xml
curl http://127.0.0.1:8090/robots.txt
```

本地 `go run` 若因模块解析或网络问题卡住，优先使用本文档的二进制排障启动，不要把临时代理、token 或私钥写进代码。
