# DevHub Docker 一键线上部署

[返回文档大纲](README.md)

该流程适合在 Ubuntu 服务器上用 Docker Compose 一键启动 DevHub + MySQL。服务器需要已安装 Docker Engine 和 Docker Compose plugin。

## 首次部署

```bash
cd /opt/devhub
cp .env.docker.example .env.docker
vi .env.docker
```

至少修改：

```text
DEVHUB_PUBLIC_URL=https://你的域名
DEVHUB_MYSQL_PASSWORD=生产数据库密码
DEVHUB_MYSQL_ROOT_PASSWORD=生产 root 密码
```

启动：

```bash
./scripts/docker-deploy.sh up
```

入口：

```text
前台：http://服务器IP:8090/
后台：http://服务器IP:8090/admin-next
健康检查：http://服务器IP:8090/api/v1/health
```

如果配置了 Nginx / Caddy / 云负载均衡，请反代到 `127.0.0.1:8090`，并开启 HTTPS。

## 常用命令

```bash
./scripts/docker-deploy.sh status
./scripts/docker-deploy.sh logs
./scripts/docker-deploy.sh restart
./scripts/docker-deploy.sh down
```

## 数据卷

Compose 会创建以下持久化卷：

```text
devhub_mysql_data      MySQL 数据
devhub_storage         DevHub 上传与运行存储
devhub_plugins_local   本地插件仓库
```

停止容器不会删除数据卷。生产回滚和清理前请先备份这些卷。
