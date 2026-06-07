# DevHub Plugin Package Fixtures

本目录用于 v1.8.3-S13 / S15 真实插件包链路验收。

生成命令：

```bash
./scripts/build-plugin-package-fixtures.sh
```

可选传入后缀，避免 E2E 共享环境中插件编码冲突：

```bash
./scripts/build-plugin-package-fixtures.sh --suffix "$(date +%s)"
```

生成结果位于 `scripts/fixtures/plugin-packages/dist/`：

- `devhub-fixture-valid-plugin*.zip`：可通过 upload -> precheck -> promote -> install dry-run -> install 的声明型插件包。
- `devhub-fixture-blocked-plugin*.zip`：包含 `scripts/install.sh`，必须被阻断，后端 promote 必须拒绝。
- `devhub-fixture-deprecated-schema-plugin*.zip`：包含根目录 `001_schema.sql` 和 `migrations/001_init.sql`，用于验证 deprecated warning；根目录 SQL 不作为标准迁移入口。
- `devhub-fixture-links-plugin*.zip`：S15 真实声明型友情链接插件，声明菜单、权限、`friend_link*` content_type、配置 Schema 和 `migrations/001_init.sql`，用于验证“安装到使用”的完整业务闭环。

这些 fixture 只用于测试，不是正式插件包；不包含真实 token、Secret、远程代码或动态加载入口。
