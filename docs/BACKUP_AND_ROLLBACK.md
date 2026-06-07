# DevHub 备份与回滚文档

[返回文档大纲](README.md)

更新时间：2026-05-22

本文档用于 DevHub v1.x 的上线前归档、备份和紧急回滚。当前版本以仓库根目录 `VERSION` 和 release 文档为准；本文不固化具体版本号。生产执行前请先在预发环境演练，避免直接覆盖线上数据。

## 需要备份的内容

- MySQL 数据库：业务数据、用户、权限、Topic、评论、通知、治理记录和审计日志。
- 部署配置：`.env`、systemd 配置、反向代理配置、Docker Compose 配置等，不要提交密钥到仓库。
- 上传文件目录：如果部署环境启用了上传目录，需要与数据库同批次备份。
- 静态构建产物：`web/frontend`、`web/admin-vue`，或生产环境实际发布目录。
- 当前二进制：例如 `.devhub/devhub` 或生产环境中的 `devhub` 可执行文件。
- 文档归档：`README.md`、`docs/`、`CHANGELOG.md`、`VERSION`。
- release 包：打包后的源码、构建产物、二进制和 schema 快照。

## 生产备份清单（v1.8.4-S5）

生产环境在执行插件安装、升级、启停、归档、恢复前，管理员必须先确认备份可用，且已经在临时库或预发环境验证过恢复路径。最小备份范围如下：

| 备份对象 | 需要覆盖的内容 | 恢复关注点 |
| --- | --- | --- |
| 主数据库 | 所有业务表、插件治理表、审计表和迁移记录 | 这是插件安装 / 升级 / 配置恢复的主恢复点。 |
| 插件安装记录 | `plugins`、声明 manifest、版本、状态、checksum、source metadata | 恢复后应与本地插件包仓库中的包版本一致。 |
| 插件配置记录 | 全局配置、子站配置、配置版本历史、脱敏 diff | 敏感字段恢复依赖同一组加密 key。 |
| 插件配置加密 key | `DEVHUB_PLUGIN_CONFIG_KEYS`、`DEVHUB_PLUGIN_CONFIG_KEY_ID`、`DEVHUB_PLUGIN_CONFIG_KEY`、old keys 等部署侧密钥材料 | 加密 key 丢失后，历史加密配置无法解密恢复。 |
| 插件包本地仓库 | `storage/plugins/packages/` | 用于恢复 promoted 后的本地包，避免 DB 指向不存在的包路径。 |
| 插件包上传暂存区 | `storage/plugins/uploads/`、`storage/plugins/staging/`、`storage/plugins/quarantine/`，如当前部署启用 | 仅用于排障、重扫和重新 promote；过期清理前先确认不再需要事故证据。 |
| `admin_logs` | 插件安装、升级、配置、health check、manual retry、registry reload 等审计 | 用于确认谁执行了什么、失败停在哪一步。 |
| `hook_executions` | 内置 Hook、external_service health check / 投递 / retry 执行记录 | 用于判断 webhook / external_service 失败是否需要重试或告警。 |
| Webhook Secret 元数据 | Secret ref、状态、轮换窗口、签名元信息 | 不备份明文 Secret；明文只在创建 / 轮换时展示一次。 |
| Callback Token 元数据 | token hash、scope、community scope、状态、回调请求记录 | 不备份明文 token；明文只在创建 / 轮换时展示一次。 |
| Core SecretCenter 元数据 | `secret_refs` 的 ref、namespace、name、key_id、encrypted_value、状态、last_used_at、rotated_at | 不备份明文 secret/token；恢复依赖同一组启动 root key。 |
| external_service 配置 | endpoint、health_check_path、timeout、failure_policy、auth_type、token_ref、enabled、健康状态 | Bearer token 明文不应导出；恢复依赖加密 key 和 token_ref 记录。 |
| 站点 / 子站插件启用状态 | `community_plugins`、排序、子站配置、禁用状态 | 恢复后要重新检查全局 enabled 与子站 enabled 的组合。 |
| migrations 执行记录 | `plugin_migrations`、执行状态、失败原因、声明来源 | 用于判断哪些 migration 已执行、哪些仍 pending / failed。 |

安全注意事项：

- 加密 key 丢失后，历史加密配置无法恢复；恢复数据库但缺少 key 只会得到不可解密密文。
- token / secret 明文本身不应被日志、审计、备份摘要或工单系统保存。
- 备份时不能把 token / secret 明文导出到不安全位置；只保留必要的密文、hash、secret_ref、token_ref 和状态元数据。
- MemoryStore 只适合本地开发、临时演示或无持久化测试，不适合作为生产持久化方案。
- MySQLStore 是生产建议模式；生产验收、备份、恢复和回滚演练都应以 MySQLStore 为准。

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

## 插件安装前检查清单

执行 upload、precheck、promote、install dry-run 或 install 前，至少确认：

1. 插件包来源可信，签名 / publisher / checksum 风险已按当前策略处理。
2. `scripts/plugin-package-check.sh` 已通过；`blocked` 必须修复后重跑，`warning` 必须说明原因。
3. upload / precheck 无 blockers，blocked 上传包不能 promote。
4. `migrations/` 结构合法，迁移文件按文件名排序且只来自 `migrations/`。
5. 根目录 `001_schema.sql` 不执行，只作为 deprecated warning；根目录其他 SQL 仍应阻断。
6. install dry-run 已从本地插件仓库包重新执行，不能复用 upload / staging 旧计划。
7. dry-run plan 未过期。
8. dry-run plan checksum、manifest checksum、plugin code、version、migration plan hash 与当前包一致。
9. dry-run 不执行 SQL、不执行代码、不改变安装状态、不刷新 PluginRegistry。
10. 插件不包含 package scripts、可执行二进制、Go / JS / PHP / shell 等危险文件。
11. 插件不包含远程 iframe、remote JS、remote component、inline HTML 或未允许的前端执行入口。
12. 当前数据库、配置、加密 key、本地插件仓库、上传暂存区、审计和执行记录已备份。
13. `PluginRegistry reload` 失败时已有处理路径：保留旧快照、查看审计、修复元数据或配置后重试 reload。

## 插件升级前检查清单

执行 upgrade dry-run 或 upgrade 前，至少确认：

1. 当前版本和目标版本明确，目标版本高于当前版本。
2. upgrade dry-run 已对当前目标包 / manifest 执行。
3. `risk_level` 已确认，只能按 `safe`、`warning`、`blocked` 三类处理。
4. `safe` 可按正常审批 / 执行流程继续，但仍需保留备份。
5. `warning` 必须由管理员显式 `confirm=true`，并记录确认原因。
6. `blocked` 不能通过 `confirm=true`、审批或前端按钮绕过。
7. manifest diff 已检查。
8. permissions diff 已检查。
9. content_types / content_type_definitions diff 已检查。
10. config_schema / default_config diff 已检查，必填项变化和敏感字段变化已评估。
11. frontend_mount diff 已检查；未知挂载点、远程 iframe、remote JS、inline HTML 必须 blocked。
12. external_service diff 已检查，包括 endpoint、auth_type、timeout、failure_policy 和 Hook 声明变化。
13. migrations diff 已检查，确认是否新增、删除或重命名 migration。
14. dependencies diff 已检查，required 依赖缺失 / disabled / archived / 版本不满足必须阻断。
15. 影响站点 / 子站数量已确认。
16. 影响权限、菜单、路由和治理入口已确认。
17. 主数据库、插件配置、加密 key、本地包仓库和审计记录备份已完成。
18. dry-run plan 未过期，且 plan 与当前包 checksum / manifest / migration plan 一致。

## upgrade dry-run 风险等级处理

| 风险等级 | 含义 | 处理方式 |
| --- | --- | --- |
| `safe` | 未发现需要确认的高风险差异或阻断项 | 可继续审批 / 执行；生产仍需先完成备份。 |
| `warning` | 存在可接受但需要管理员确认的风险，例如配置 required 变化、权限扩大、content_type 影响或未签名包 | 必须显式 `confirm=true`，并在变更记录中说明接受原因。 |
| `blocked` | 存在不可绕过风险，例如同版本 / 降级、Core 不兼容、required 依赖不满足、manifest invalid、危险 frontend_mount、blocking Hook、checksum mismatch 或 migration plan 不一致 | 不允许安装 / 升级；修复包或环境后重新 build / check / dry-run。 |

## failure_stage 处理建议

| failure_stage | 失败含义 | 是否已修改数据库 | 是否可重新执行 | 是否需要人工恢复 / 备份 | 排查入口 |
| --- | --- | --- | --- | --- | --- |
| `precheck` | 包来源、权限、manifest、checksum、危险文件或依赖预检失败 | 否 | 修复包后可重新 upload / precheck | 通常不需要恢复备份 | package detail、risk_report、`admin_logs` |
| `dry_run` | dry-run 计划生成或计划校验失败 | 否 | 修复 checksum、manifest 或 migration plan 后可重跑 | 通常不需要恢复备份 | dry-run 响应、blocked_reasons、`admin_logs` |
| `confirm` | warning 未确认、确认 token / plan 不匹配、plan 过期，或尝试绕过 blocked | 否 | warning 可重新 dry-run 后确认；blocked 不可重试绕过 | 通常不需要恢复备份 | upgrade dry-run、审批详情、`admin_logs` |
| `migration` | migration 执行或迁移记录写入失败 | 是，可能已有 DDL / 记录变更 | 谨慎；先确认哪些 migration 已执行 | 可能需要恢复数据库备份；当前无自动 migration down | `plugin_migrations`、DB 日志、`admin_logs` |
| `config_migration` | 配置迁移、默认配置写入、配置 schema 适配或敏感配置处理失败 | 可能已写部分配置或配置版本 | 修复配置 / key 后可重试保存或重新升级 | 可能需要从配置备份恢复；key 缺失时无法解密历史密文 | 配置版本、key status、`admin_logs` |
| `registry_reload` | PluginRegistry 运行态快照刷新失败 | 通常不新增业务数据；安装 / 升级记录可能已写入 | 修复元数据 / 配置 / 依赖后可重试 reload | 通常先不恢复 DB；保留旧快照并重试 | `plugin.registry.reload.failed` 审计、服务日志 |
| `enable` | 全局启用或子站启用失败 | 通常仅状态写入前失败，具体以审计为准 | 修复依赖、配置、migration 状态后可重新 enable | 通常不需要恢复备份 | readiness / enable-precheck、`admin_logs` |
| `webhook` | Webhook Secret、签名、投递记录、重试或熔断治理失败 | 不影响主业务写入；可能新增执行 / 审计记录 | 失败类 non-blocking 记录可按规则 retry | 通常不需要恢复数据库备份 | Webhook 治理、`hook_executions`、`admin_logs` |
| `external_service` | external_service endpoint、health check、token_ref、投递或 manual retry 失败 | 不影响主业务写入；可能新增执行 / 健康记录 | 修复 endpoint / token / 状态后可 health check 或 retry | 通常不需要恢复数据库备份；token 恢复依赖加密 key | external_service 配置、`hook_executions`、health 摘要 |
| `unknown` | 未归类异常或调用链未返回结构化阶段 | 不确定 | 先暂停重复执行，确认状态后再决定 | 先备份当前异常状态；必要时恢复数据库备份 | 服务日志、`admin_logs`、`hook_executions`、DB 状态 |

通用处置顺序：

1. 先记录 `failure_stage`、`failure_reason`、request_id、插件 code、当前版本、目标版本和操作者。
2. 查看 `admin_logs`，确认最后一个成功动作和失败动作。
3. 查看 `hook_executions`，确认是否涉及 Hook、Webhook 或 external_service。
4. 如果阶段是 `registry_reload`，先确认旧快照是否仍在服务中可用，再修复元数据后重试 PluginRegistry reload。
5. 如果阶段是 `migration`，不要假设可自动回退；先确认已执行 migration，再决定恢复数据库备份或手工修复。
6. 如果阶段不明确，先备份当前异常状态，再做恢复。

## 回滚边界

当前支持边界：

- 当前不提供完整自动 rollback。
- 当前不提供 migration down 复杂体系。
- 当前不会自动回滚第三方数据或外部服务状态。
- install / upgrade 失败后优先依赖 dry-run、失败阶段、审计、备份和人工恢复路径，而不是自动回滚。
- `PluginRegistry reload` 成功时才替换运行态快照；reload 失败应保留旧快照并写审计。
- 插件配置可通过配置备份值重新保存恢复；敏感字段恢复依赖同一组加密 key。
- 插件包本地仓库可通过备份目录恢复，或重新 promote 原 zip / 原包。
- `disabled` / `archived` 不是完整回滚，只是停止新能力或进入治理入口；历史内容、配置、迁移和审计仍保留。

install 失败处理：

- 发生在 `precheck`、`dry_run`、`confirm` 阶段时，通常未写入安装状态，修复后重跑。
- 发生在写入安装记录或 migration 记录后，需要按 `failure_stage` 对照数据库状态；不要只看前端 toast 判断是否已修改。
- 发生在 `registry_reload` 阶段时，优先修复 registry 输入并重试 reload；不要把运行态旧快照误判为数据库已回滚。

upgrade 失败处理：

- warning 未确认、blocked、plan 过期、checksum 不一致、migration plan 不一致时，拒绝升级且不应修改安装状态。
- 如果 manifest / version / migration pending 记录已经写入但 registry reload 失败，应以数据库记录、审计和旧运行态快照为准制定恢复步骤。
- 如果 migration 已执行，自动回退依赖未来版本规划；当前应优先从数据库备份恢复，或由 DBA 基于已执行 migration 做人工修复。

PluginRegistry reload 失败处理：

1. 不继续执行依赖新运行态的启用、配置或投递动作。
2. 查看 `plugin.registry.reload.failed` 审计和服务日志。
3. 检查 manifest、config_schema、依赖、状态、子站配置和敏感配置解密是否可读。
4. 修复后重试 reload；旧快照继续服务时，也要把数据库状态和后台提示对齐。
5. 若 reload 失败源自不可恢复的数据写入，再考虑恢复数据库备份。

配置恢复：

- 优先从插件配置记录和配置版本历史恢复；敏感字段只应通过后台重新写入或从受控密文 + key 恢复。
- 加密 key 丢失时，已有 `enc:v1` / `enc:v2` 密文无法解密，不能通过审计或日志找回明文。
- 不要从 `admin_logs`、`hook_executions` 或截图中寻找 token / secret 明文；这些位置不应保存明文。

插件包本地仓库恢复：

- 恢复 `storage/plugins/packages/` 后，确认目录、`manifest.json`、`checksums.json`、版本号和 DB 中 `source_type/source_path/checksum` 仍一致。
- 如果本地仓库丢失但原上传包仍在，可重新 promote 原包，再执行 install / upgrade dry-run。
- 如果上传暂存区也丢失，需要重新上传原始 zip，并重新 precheck / promote / dry-run；不能凭旧 dry-run 结果直接安装或升级。

## MySQL / MemoryStore 生产建议

- MySQLStore 是生产推荐模式。
- MemoryStore 只适合本地开发、临时演示或无持久化测试。
- 生产环境优先确保 MySQL 备份、配置加密 key 和插件包本地仓库可恢复。
- 对于需要回滚演练的插件升级，先在预发或临时库验证 dry-run 和恢复路径，再碰生产库。

MySQL 生产验收建议：

1. 使用 `./dev.sh start --mysql` 或生产等价启动参数进入 MySQLStore。
2. 验证 `/api/v1/health` 返回 MySQL 存储模式。
3. 验证插件 upload -> precheck -> promote -> install dry-run -> install。
4. 验证插件配置保存、敏感字段脱敏、配置 key status 和配置恢复路径。
5. 验证 `admin_logs` 可查询安装 / 升级 / reload / health check / retry 审计。
6. 验证 `hook_executions` 可查询 Hook、health check、external_service 投递和 retry 记录。
7. 验证 upgrade dry-run 的 safe / warning / blocked、plan 过期拒绝、checksum 不一致拒绝和 migration plan 不一致拒绝。
8. 验证 PluginRegistry reload 成功路径；如要演练失败，必须在预发环境通过可恢复方式制造失败。

## 升级演练流程

生产变更前建议在预发环境按以下顺序演练；如果没有自动脚本，按后台或 Admin API 手动完成同等步骤：

```bash
./scripts/plugin-package-build.sh examples/plugins/official_links
./scripts/plugin-package-check.sh examples/plugins/official_links
./scripts/plugin-package-build.sh examples/plugins/official_webhook_notify
./scripts/plugin-package-check.sh examples/plugins/official_webhook_notify
./dev.sh start --mysql
git diff --check
./scripts/check-frontend.sh --admin-only --quick
```

手动流程：

1. build 插件包。
2. check 插件包，确认输出不是 `blocked`。
3. upload zip。
4. precheck，确认无 blockers。
5. promote 到本地插件仓库。
6. 对本地仓库包执行 install dry-run。
7. install，确认安装后默认 disabled，并触发 PluginRegistry reload。
8. 对目标版本执行 upgrade dry-run，记录 `risk_level`、`confirm_required`、`confirm_token`、`rollback_boundary`。
9. 演练 `safe` 升级：确认不需要额外确认但仍需要备份。
10. 演练 `warning` 升级：使用 `confirm=true`，确认审计记录 `confirm_used=true`。
11. 演练 `blocked` 升级：确认即使传 `confirm=true` 也被拒绝。
12. 演练 dry-run plan 过期：确认必须重新 dry-run。
13. 演练 checksum 不一致：修改或替换包后确认旧 plan 被拒绝。
14. 演练 migration plan 不一致：变更 `migrations/` 后确认旧 plan 被拒绝。
15. 演练 PluginRegistry reload 成功；失败路径只在预发用可恢复数据演练。
16. 检查 `admin_logs` 和 `hook_executions`。
17. 检查后台插件治理页面技术 JSON 默认折叠、敏感字段脱敏、失败阶段 / 原因和下一步建议可见。

## 审计与执行记录保留建议

- `admin_logs`、`hook_executions`、Webhook Secret 元数据和 Callback Token 元数据应作为生产审计的一部分长期保留，至少保留到对应变更完成并完成事故复盘。
- 不要在未确认排障结束前清理这些记录，也不要因为“看起来像噪音”就先删。
- 如果需要做归档或脱敏导出，只能导出脱敏后的审计摘要，不能把 token / secret 明文导出到不安全位置。
- 回滚并不会自动清除这些记录；失败阶段、重试记录和审计记录仍应保留，用于后续追踪和复盘。

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
