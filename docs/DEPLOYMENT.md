# DevHub 部署与启动文档

[返回文档大纲](README.md)

更新时间：2026-05-09

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

默认端口为 `8090`，默认数据仓库为 `CMS_STORE=memory`。修改 Go 后端代码后需要重启服务；只改 Go 且前后台产物无需重建时，可以使用 `./dev.sh restart --no-build` 或 `./dev.sh --local-go restart --no-build`。

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
