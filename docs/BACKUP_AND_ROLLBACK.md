# DevHub 备份与回滚文档

[返回文档大纲](README.md)

更新时间：2026-05-10

本文档用于 DevHub v1.x 的上线前归档、备份和紧急回滚。当前版本为 v1.2.1；生产执行前请先在预发环境演练，避免直接覆盖线上数据。

## 需要备份的内容

- MySQL 数据库：业务数据、用户、权限、Topic、评论、通知、治理记录和审计日志。
- 部署配置：`.env`、systemd 配置、反向代理配置、Docker Compose 配置等，不要提交密钥到仓库。
- 上传文件目录：如果部署环境启用了上传目录，需要与数据库同批次备份。
- 静态构建产物：`web/frontend`、`web/admin-vue`，或生产环境实际发布目录。
- 当前二进制：例如 `.devhub/devhub` 或生产环境中的 `devhub` 可执行文件。
- 文档归档：`README.md`、`docs/`、`CHANGELOG.md`、`VERSION`。
- release 包：打包后的源码、构建产物、二进制和 schema 快照。

## MySQL 备份

示例：

```bash
backup_dir=backups/$(date +%Y%m%d_%H%M%S)
mkdir -p "$backup_dir"
mysqldump -u USER -p DB_NAME > "$backup_dir/devhub.sql"
```

开发库示例：

```bash
mysqldump -h127.0.0.1 -P3307 -u devhub -p devhub > backup_$(date +%Y%m%d).sql
```

建议同时记录当前 Git commit、二进制校验值和 schema 文件版本：

```bash
git rev-parse HEAD > "$backup_dir/git_commit.txt"
sha256sum .devhub/devhub > "$backup_dir/devhub.sha256"
cp db/mysql/001_schema.sql "$backup_dir/001_schema.sql"
```

## MySQL 恢复

示例：

```bash
mysql -u USER -p DB_NAME < backup_YYYYMMDD.sql
```

开发库示例：

```bash
mysql -h127.0.0.1 -P3307 -u devhub -p devhub < backup_YYYYMMDD.sql
```

恢复前建议：

- 停止 DevHub 服务，避免写入中数据不一致。
- 先备份当前异常状态，保留排查线索。
- 确认备份文件对应的代码版本和 schema 版本。
- 对生产库执行恢复前，先在临时库验证 SQL 可导入。

## 构建产物备份

前台和后台产物目录：

```text
web/frontend
web/admin-vue
```

示例：

```bash
backup_dir=backups/$(date +%Y%m%d_%H%M%S)
mkdir -p "$backup_dir"
tar -czf "$backup_dir/frontend.tgz" web/frontend
tar -czf "$backup_dir/admin-vue.tgz" web/admin-vue
```

如果生产环境使用独立静态目录，请备份实际发布目录，而不是只备份仓库内目录。

## 二进制回滚

发布新版本前保留上一版二进制：

```bash
cp /opt/devhub/devhub /opt/devhub/releases/devhub.$(date +%Y%m%d_%H%M%S)
```

回滚时：

```bash
systemctl stop devhub
cp /opt/devhub/releases/devhub.<old-version> /opt/devhub/devhub
chmod +x /opt/devhub/devhub
systemctl start devhub
```

本地排障可用：

```bash
lsof -i :8090
kill <pid>
PORT=8090 CMS_STORE=memory ./.devhub/devhub
```

## Git 版本回滚

查看版本：

```bash
git tag
git show v1.1.0
git show v1.0.0
```

回到 v1.1.0：

```bash
git checkout v1.1.0
```

如果需要从 tag 创建修复分支：

```bash
git checkout -b hotfix/v1.1.0 v1.1.0
```

生产发布不建议直接在服务器上做复杂 Git 操作；更稳妥的方式是在 CI 或构建机产出 release 包，再部署到服务器。

## 数据库回滚注意事项

- schema 变更后回滚必须谨慎，旧代码可能无法识别新字段，新代码也可能依赖新字段。
- 回滚数据库前必须先备份当前异常状态。
- 生产环境不建议直接覆盖数据库，除非已经确认业务允许丢弃回滚点之后的数据。
- 如果只是代码问题，优先回滚二进制和静态产物，数据库保持当前状态。
- 涉及治理、通知、审计日志的回滚要确认是否会重复触发业务动作。

## 紧急回滚流程

1. 宣布变更冻结，停止自动发布。
2. 停止 DevHub 服务。
3. 备份当前异常状态，包括数据库、日志、配置、二进制和静态产物。
4. 恢复上一版二进制和静态构建产物。
5. 恢复配置文件，确认 `PORT=8090`、`CMS_STORE` 和数据库连接正确。
6. 如确需数据库回滚，先在临时库验证备份，再导入生产库。
7. 启动服务。
8. 检查健康接口、首页、后台、Topic 详情、sitemap 和 robots。
9. 检查错误日志和关键业务接口。
10. 记录事故原因、回滚版本、执行人和恢复时间。

## 回滚后检查

```bash
curl -I http://127.0.0.1:8090/
curl http://127.0.0.1:8090/api/v1/health
curl http://127.0.0.1:8090/api/v1/topics
curl http://127.0.0.1:8090/admin-next
curl http://127.0.0.1:8090/topics/1
curl http://127.0.0.1:8090/sitemap.xml
curl http://127.0.0.1:8090/robots.txt
```

`/topics/:id` 仍需保持 Go 动态 SEO HTML，不能回滚到纯 CSR 空壳。
