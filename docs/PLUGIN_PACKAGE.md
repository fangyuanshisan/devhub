# DevHub 本地插件包规范（草案）

[返回文档入口](README.md)

更新时间：2026-05-14

本文件是 **Core + 插件服务底座** 中插件生态的包格式与治理规范。插件包不是简单上传文件，而是未来插件生态的基础单元：它承载 manifest、config_schema、权限声明、Hook 声明、前端入口声明、后端服务声明、签名信息、发布者信息和安装前审查材料。

本文件也是 **v1.5.0-P0-01 ~ P0-04** 与 **v1.6.0-P0-01 ~ P1-10** 的阶段性成果：为 DevHub 定义“本地插件包 / 本地插件仓库”规范，并提供 **dry-run 导入预览**、**最小安装闭环**、**后台初始化插件包**、**zip 上传安全沙箱**、**上传包生命周期治理**、**Ed25519 真实签名验签 / 可信发布者管理**、**远程索引只读镜像**、**版本仓库 / 升级差异** 与 **操作恢复 / 密钥轮换边界** 能力。

注意：

- v1.5.0-P0-01 ~ P0-03：只做 **安全读取 + 校验 + 预览**（dry-run），不安装插件。
- P0-04：补齐“从本地插件包安装声明型插件”的最小闭环，但仍然 **不执行插件代码、不执行外部 SQL、不动态加载前端资产**。
- v1.6.0-P0-01：后台新增“上传插件包 zip”，只进入 `uploads/staging/quarantine` 安全沙箱，随后复用 package scanner / checksum / signature / risk_report / dry-run；上传后不自动安装。
- v1.6.0-P0-02：上传包升级为可追踪生命周期对象，支持列表、详情、重新扫描、导入审批、promote、取消、删除和手动 cleanup。
- v1.6.0-P0-03：签名从草案升级为 Ed25519 真实验签，后台可维护可信发布者公钥；promote / install / approval 执行前会重新验签。

## 目标与边界

目标：

- 定义本地插件包目录结构与字段口径。
- 后台支持从允许目录读取插件包，并返回 dry-run 预览信息（文件扫描 / manifest 校验 / 安装预览）。
- 后台支持初始化声明型插件包模板，并把模板写入受控本地插件仓库目录。
- 后台支持上传 zip 插件包，并在受控沙箱目录中安全解压、扫描、预览和风险报告。

明确不做：

- 不做远程市场 / 远程下载 / 在线更新。
- 不做动态加载 Go 代码，不执行 JS / WASM / Lua，不提供脚本沙箱。
- 不执行任何第三方本地代码。
- 不执行外部 raw SQL。
- 不动态加载前端资产。
- 不做上传后自动安装。
- 不做完整 PKI / CA 证书链 / 远程可信源同步；当前可信来源只来自后台本地可信发布者记录。
- 不做完整插件运行模型；第三方插件前端挂载、后端隔离运行、HTTP 插件服务协议、iframe 沙箱和受控 API 调用仍处于规划中。

## 运行模型字段设计（v1.7.2）

以下字段用于未来插件运行模型设计，当前只是 manifest 规划字段，不代表已实现：

```json
{
  "runtime": {
    "type": "internal|http_service|iframe",
    "entry": "string",
    "health_check": "/health",
    "timeout_ms": 3000,
    "permissions": [],
    "scopes": []
  },
  "frontend": {
    "mounts": [
      {
        "slot": "admin.sidebar.menu",
        "title": "Demo",
        "path": "/admin-next/plugins/demo",
        "mode": "iframe",
        "url": "/plugins/demo/admin"
      }
    ]
  },
  "backend": {
    "service_url": "https://plugin-service.example.com",
    "auth": "signed_request",
    "hooks": []
  },
  "api_scopes": [
    "content.read",
    "content.write"
  ]
}
```

设计边界：

- `runtime.type=internal` 仅用于 Core 内置插件或官方随仓库编译插件。
- `runtime.type=http_service` 表示后端能力由独立 HTTP 插件服务承载。
- `runtime.type=iframe` 表示插件主要提供前端隔离页面。
- `frontend.mounts` 只声明挂载位置，不代表当前 Core 已实现该 slot。
- `backend.service_url` 必须经过可信来源、SSRF、防重放和签名校验设计；v1.8.3-S11 当前只落地 external_service 的受控 endpoint 配置与 health check 预备能力，不等于远程 Hook 投递或第三方代码执行已开放。
- `api_scopes` 是插件申请的最大 Core API 权限范围，最终授权仍由管理员和 Core 策略决定。

v1.8.3-S11 实现边界：

- external_service 配置可在后台保存 endpoint、health_check_path、timeout_ms、failure_policy、auth_type 与 token_ref。
- health check 只做 `GET {endpoint_url}{health_check_path}` 探活，并记录 `hook_executions(service_type=external_service)`；不会执行插件包脚本、不会加载远程代码、不会开放 blocking Hook。
- external_service token 与 Webhook Secret、Callback Token 分离管理，列表 / 详情 / 执行记录 / 审计不回显明文。

完整设计见 [插件运行模型设计](PLUGIN_RUNTIME_MODEL.md)。

Webhook / HTTP 插件服务协议设计见 [Webhook / HTTP 插件服务协议（设计）](PLUGIN_WEBHOOK_PROTOCOL.md)。

实现阶段拆解（v1.7.3 文档阶段）见 [Webhook 协议实现拆解（v1.7.3）](PLUGIN_WEBHOOK_IMPLEMENTATION_PLAN.md) 与官方示例插件验证方案（公告插件）：`docs/plugins/official-announcement-plugin.md`。两者均为任务拆解/验证方案，不代表已实现真实投递、重试队列、熔断或外部插件服务运行。

前端挂载模型（v1.8.0 文档阶段）见：`docs/PLUGIN_FRONTEND_MOUNT_MODEL.md`（slots、iframe/sandbox、postMessage 设计口径）。注意：该文档仅为设计，不代表已支持第三方插件前端挂载或动态加载前端资产。

## 插件包目录结构

推荐结构（示例）：

```text
devhub-plugin-demo/
├─ manifest.json
├─ README.md
├─ config.example.json
├─ docs/
│  └─ usage.md
├─ migrations/
│  └─ 001_init.sql
├─ assets/
│  └─ preview.png
└─ checksums.json
```

规则：

1. `manifest.json` 是唯一主声明文件（必须）。
2. `README.md` 是插件说明（建议）。
3. `config.example.json` 是示例配置（建议），不得包含真实 secret。
4. `migrations/` 是插件包数据库迁移的唯一标准入口；SQL 文件按文件名排序生成计划，dry-run 不执行任何 SQL。
5. `assets/` 当前只允许预览素材；dry-run 不会动态加载到前台。
6. `checksums.json` 当前支持 **sha256 完整性校验**；如存在 `signature.json`，会对 `checksums.json` 摘要执行 Ed25519 真实验签。
7. 目录名建议与 `manifest.code` 一致；不一致会 warning。
8. 未知文件默认 warning；危险文件会 blocked。

> 注：v1.5.0 当前支持本地插件仓库扫描、插件包详情、dry-run、安装确认与审批执行链路；安装仍走 `plugin.approve` 执行权限，不允许携带代码或可执行资产的插件包。

### v1.8.3-S8 migrations/ 规范收口

DevHub 插件包数据库迁移规范统一为 `migrations/`：

- 新包只能把迁移 SQL 放在 `migrations/` 下，推荐 `migrations/001_init.sql`、`migrations/002_add_xxx.sql`。
- `migrations/` 下的 `.sql` 文件按文件名排序，dry-run 只返回 `migration_plan`，且计划项 `will_execute=false`。
- 根目录 `001_schema.sql` 已废弃；预检发现时只给 deprecated warning，提示迁移到 `migrations/001_init.sql`，不会执行，也不会进入标准 migration plan。
- 根目录其他 `.sql` 文件仍视为危险文件并阻断，避免插件包根目录任意 SQL 被误认作迁移入口。
- dry-run 是“计划预览”，不是“试执行”：不会创建表、修改表、删除表、写 migration 记录、写插件安装状态、执行插件脚本或调用远程服务产生副作用。
- install / upgrade 真实流程仍会服务端复跑 dry-run，并且只基于 `migrations/` 计划和 manifest 声明做治理记录；不会执行根目录 SQL 或第三方 SQL。
- 后续版本可考虑把根目录 `001_schema.sql` 从 warning 升级为 error；本轮先保留兼容 warning。
- v1.8.3-S12 进一步收口 upload -> promote -> install：promote 只把包转入 `storage/plugins/packages/` 本地仓库；install 只能从本地仓库包发起，且必须携带当前 install dry-run 计划凭证 `dry_run_id`。upload/staging 阶段的 dry-run 只用于预检，不可直接替代 install dry-run。

### 后台中文状态与异常提示

后台插件包治理页面展示给管理员的状态和异常原因统一走插件模块中文映射：

- 状态示例：`uploaded` 显示为“已上传”，`scanned` 显示为“已扫描”，`promoted` 显示为“已转入本地仓库”，`approval_pending` 显示为“待审批”，`blocked` 显示为“已阻断”。
- 风险 / 阻断示例：`manifest_invalid` 显示为“插件清单校验失败”，`checksum_failed` 显示为“文件完整性校验失败”，`signature_invalid` 显示为“插件签名无效”，`publisher_unknown` 显示为“发布者未受信任”，`core_incompatible` 显示为“当前 DevHub 版本不兼容”。
- 后端 API 仍保留英文 `code` 和现有枚举值；前端优先展示中文 `message`，仅有 code 时用 `web/admin-app/src/modules/plugins/statusText.js` 映射兜底并保留错误码，便于排障和自动化测试。
- 技术字段名如 `plugin_code`、`checksum`、`manifest`、`publisher_id` 可保留原名；面向管理员的按钮、空状态、确认提示和可操作异常建议必须中文化。

## 后台初始化插件包（v1.5.0 收口补充）

入口：

```text
系统插件 -> 安装升级 -> 初始化插件包
```

API：

- `POST /api/v1/admin/plugins/packages/templates/preview`
- `POST /api/v1/admin/plugins/packages/templates`

权限：

- 需要 admin token。
- 需要 `plugin.write`。

生成规则：

- 输出目录固定为 `storage/plugins/packages/{code}`。
- `code` / `content_type` 必须满足小写字母、数字、下划线并以小写字母开头的现有编码规则。
- 不允许任意 path。
- 不暴露 `force`；目标目录已存在时返回 `plugin_package_template_exists`。
- 初始化成功后写入 `admin_logs`。
- 初始化成功后自动执行 package dry-run，并返回 dry-run 状态、风险等级、manifest 校验结果、warnings/errors。

后台初始化生成文件：

```text
storage/plugins/packages/{code}/
├─ manifest.json
├─ README.md
├─ config.example.json
├─ content-type.md
├─ permissions.md
├─ hooks.md
├─ migrations.md
└─ docs/
   └─ registry-example.md
```

与 CLI `plugin:new` 的差异：

- CLI 默认仍生成 `registry.example.go`，用于源码内置插件接入示例。
- 后台初始化版默认不生成 `registry.example.go`，避免 `.go` 文件触发插件包扫描器的 dangerous file blocked。
- 后台初始化版把 registry 接入说明放在 `docs/registry-example.md`。
- dry-run 未 blocked 时，后台允许继续提交“安装审批”；直接安装仍需 `plugin.approve`。

边界：不做远程市场 / 远程下载；不执行插件代码；不执行 SQL；不生成前端动态页面；不改变插件运行时边界。

## Zip 上传安全沙箱（v1.6.0-P0-01）

入口：

```text
系统插件 -> 安装升级 -> 上传插件包 zip
```

API：

- `POST /api/v1/admin/plugins/packages/upload`
- `GET /api/v1/admin/plugins/packages/uploads/:upload_id`
- `POST /api/v1/admin/plugins/packages/uploads/:upload_id/promote`

权限：

- 上传：admin token + `plugin.write`
- 详情：admin token + `plugin.read`
- promote：admin token + `plugin.write`

目录：

```text
storage/plugins/uploads/       # 原始 zip，按 upload_id 命名
storage/plugins/staging/       # 成功解压后的安全沙箱
storage/plugins/quarantine/    # 解压成功但 dry-run blocked 的隔离目录
storage/plugins/packages/      # 本地插件仓库；promote 后才进入
```

上传流程：

1. 仅接受 `.zip`，拒绝 `.tar`、`.gz`、`.rar`、`.7z` 等格式。
2. 为每次上传生成 `pkg_upload_*`。
3. zip 先保存到 `uploads/`，再解压到 `staging/{upload_id}/`。
4. 解压前逐个检查 entry，任何结构类阻断都会删除 staging 半包。
5. 解压后识别单个插件包根目录。
6. 复用 v1.5 的 scanner / checksum / signature / risk_report / dry-run。
7. `ok|warning` 保持在 staging，可查看详情或 promote；`blocked` 移入 quarantine，仅可查看风险详情，不可 promote。
8. 上传和 promote 均写入 `admin_logs`。

zip slip 防御：

- 禁止 `../manifest.json`、`dir/../../evil`。
- 禁止 `/etc/passwd` 等绝对路径。
- 禁止 `C:\Windows\...` 等 Windows 盘符路径。
- 禁止空文件名、过长文件名、`\0`。
- 禁止 URL 编码绕过（如 `%2e`、`%2f`、`%5c`）。
- 解压前用 `path.Clean` 校验，解压时再次确认输出路径仍位于 `staging/{upload_id}/` 内。
- 解压后 dry-run 继续按本地插件包路径规则扫描，不返回系统绝对路径给前端。

zip bomb / 文件结构限制：

- 上传 zip 最大：20MB。
- 解压后总大小最大：50MB。
- 单个解压文件最大：5MB。
- 解压文件数最大：300。
- 最大目录深度：8。
- 重复 entry / 重复输出路径：blocked。
- zip 内嵌 zip/tar/gz/tgz/rar/7z/bz2/xz：blocked。
- symlink：blocked。
- hardlink / 特殊设备文件：blocked（zip entry 阶段禁止非常规文件；解压后 scanner 继续检测 hardlink）。
- 当前未做压缩比阈值检查；通过上传大小、解压总大小、单文件大小、文件数、深度、重复 entry 与嵌套压缩包限制进行防御。

插件包根目录识别：

1. zip 根目录直接有 `manifest.json`：以 zip 根目录为插件包目录。
2. zip 只有一个顶层目录且目录内有 `manifest.json`：以该顶层目录为插件包目录。
3. 找不到 `manifest.json`：blocked，`plugin_package_zip_manifest_missing`。
4. 多个 `manifest.json`：blocked，`plugin_package_zip_multiple_manifests`；本轮不支持批量上传。
5. `manifest.json` 不在根目录或单一顶层目录：blocked。

promote 规则：

- 只允许生命周期状态为 `staged/approved` 且非 blocked 的上传包 promote。
- promote 前重新 dry-run；复制到 `storage/plugins/packages/{manifest.code}/` 后再次 dry-run。
- 目标目录已存在默认拒绝，返回 `plugin_package_promote_target_exists`；当前 API 支持 `force` 字段，但后台默认不覆盖。
- checksum mismatch、dangerous file、manifest invalid、quarantine 包、路径穿越包均不能 promote。
- promote 只把包转入本地仓库，不安装插件、不提交审批、不启用插件、不执行代码/SQL、不动态加载前端资产。
- promote 成功后仍需重新执行 install dry-run；upload/staging 的旧 dry-run 结果不能直接用于 install。
- promote 成功后，本地仓库列表会展示 `source_upload_id/promoted_at`，用于追溯来源上传包；执行人以 promote 审计日志为准。
- install 只接受本地仓库包路径，且服务端会校验 `dry_run_id` 是否与当前 path / plugin_code / version / manifest checksum / checksum status / migration plan hash 一致并未过期；无有效 `dry_run_id`、过期或不一致时拒绝安装。

## 上传包生命周期治理（v1.6.0-P0-02）

上传 zip 不再只是临时文件，而是 `plugin_package_uploads` 记录。MemoryStore 与 MySQLStore 保持同一字段语义：

- 基础字段：`upload_id`、`original_filename`、`uploaded_by`、`uploaded_by_name`、`uploaded_at`、`status`、`expires_at`、`deleted_at`。
- 包信息：`package_code`、`package_name`、`package_version`、`upload_path`、`staging_path`、`package_path`、`promoted_path`。
- 扫描摘要：`compressed_size`、`uncompressed_size`、`file_count`、`checksum_status`、`signature_status`、`publisher_id`、`trust_status`、`risk_level`。
- 快照 JSON：`zip_scan_json`、`file_scan_json`、`risk_report_json`、`manifest_validation_json`、`install_dry_run_json`、`metadata_json`。
- 审批关联：`approval_id`（导入 / promote 审批）、`install_approval_id`（安装审批预留）。
- 错误字段：`error_code`、`error_message`。

状态集合：

```text
uploaded -> scanned -> staged
uploaded -> scanned -> blocked
staged -> approval_pending -> approved -> promoted
staged -> approval_pending -> approval_rejected
staged -> promoted
promoted -> install_approval_pending -> installed
staged/approval_rejected/blocked/failed -> canceled
staged/blocked/approval_rejected/canceled/failed/expired/promoted -> deleted
任意非终态超过 expires_at -> expired（本轮由手动 cleanup 触发）
处理中异常 -> failed
```

后端强校验规则：

- `blocked/deleted/expired/canceled` 不能 promote，也不能提交安装审批。
- `staged/approved` 可以 promote；`approval_pending` 必须先完成审批。
- `promoted` 后才进入本地插件包安装 / 安装审批流程；promote 不安装插件。
- `installed` 后不能再次按同 code 走安装，应进入升级流程。
- `deleted` 不支持恢复；需要重新上传。
- 状态变更写入 `admin_logs`，对外 API 只返回相对路径。

导入审批：

- 上传包导入审批复用 `plugin_approval_requests`，新增 action：`package_promote` / `package_import`。
- 当前后台支持从上传包详情提交 `package_promote` 审批，并保存 upload_id、risk_report、signature、checksum、dry-run 快照。
- 审批通过只把上传包置为 `approved`；promote 前仍重新扫描、重新 dry-run，不允许用审批绕过 blocked/checksum/manifest 风险。
- blocked 包不能提交导入审批；warning/high-risk 包可进入审批或由管理员按策略 promote。

生命周期 API：

- `GET /api/v1/admin/plugins/packages/uploads`：上传包列表，支持 status/risk/keyword/uploader/package/publisher/trust 分页筛选，返回 summary。
- `GET /api/v1/admin/plugins/packages/uploads/:upload_id`：详情，返回 record、扫描快照、审批信息和 actions。
- `POST /api/v1/admin/plugins/packages/uploads/:upload_id/rescan`：重新执行扫描与 dry-run，刷新 risk_report 和状态。
- `POST /api/v1/admin/plugins/packages/uploads/:upload_id/approval`：提交导入审批。
- `POST /api/v1/admin/plugins/packages/uploads/:upload_id/approve` / `reject`：审批通过 / 拒绝。
- `POST /api/v1/admin/plugins/packages/uploads/:upload_id/promote`：复制到 `storage/plugins/packages/{code}/`，不安装。
- `POST /api/v1/admin/plugins/packages/uploads/:upload_id/cancel`：取消上传包。
- `DELETE /api/v1/admin/plugins/packages/uploads/:upload_id`：删除上传 / staging 文件并保留审计记录；promoted 包不会删除本地仓库包。
- `POST /api/v1/admin/plugins/packages/uploads/cleanup`：手动清理 `expired/deleted/canceled/failed` 的 upload / staging 文件，不删除本地仓库或已安装插件。

后台入口：

```text
系统插件 -> 上传包管理
/admin-next/plugins/packages/uploads
```

页面能力：上传 zip、筛选列表、详情抽屉、zip scan、package scan、checksum、signature、risk_report、manifest validate、dry-run、可执行动作与不可用原因、rescan、提交导入审批、审批通过 / 拒绝、promote、cancel、delete、cleanup。页面明确展示：上传包只是 staging，promote 不等于安装，不执行插件代码、不执行 SQL、不加载前端资产。

## 允许文件 / 未知文件 / 危险文件

允许文件（allow）：

- `manifest.json`
- `README.md`
- `LICENSE`
- `config.example.json`
- `checksums.json`
- `docs/**/*.md`
- `migrations/**/*.json`
- `assets/**/*.png|jpg|jpeg|webp`
- `examples/**/*.md|json`

未知文件（unknown）：

- 不在 allow 列表中的文件（不会自动阻断，但会 warning）。
- 当前 `assets/**/*.svg` 作为 unknown 处理（后续若确定安全策略可提升为 allow）。

危险文件（dangerous，dry-run 必须 blocked）：

- 任意路径包含 `../` 或路径穿越。
- 绝对路径且不在允许根目录内。
- 软链接、硬链接（如果可检测）。
- 任意隐藏目录（以 `.` 开头），除非明确允许。
- 任意具可执行权限位的文件（best-effort）。
- `node_modules/`、`.git/`、`vendor/`、`dist/`、`build/` 目录内容。
- 扩展名：`.sh`、`.bash`、`.zsh`、`.ps1`、`.bat`、`.cmd`、`.exe`、`.dll`、`.so`、`.dylib`、`.php`、`.go`、`.js`、`.mjs`、`.ts`、`.tsx`、`.jsx`、`.wasm`、`.lua`、`.py`、`.rb`、`.jar`、`.class`、`.sql`、`.env`。

说明：本轮 **不会执行任何文件**，危险文件阻断的目标是避免“导入阶段”成为隐蔽的执行入口。

## 大小限制（默认常量）

默认限制：

- 单文件最大：2MB
- 插件包总大小最大：10MB
- 文件数量最大：100
- `manifest.json` 最大：256KB（超限必 blocked）
- `README.md` 最大：512KB（超限会 warning）
- `config.example.json` 最大：256KB（超限会 warning）

## Dry-run API

`POST /api/v1/admin/plugins/packages/dry-run`

认证与权限：

- 需要 admin token（不允许 user token / moderator token）。
- 需要 `plugin.write` 权限。

请求体：

```json
{ "path": "examples/plugins/demo_notice" }
```

响应（简化示意）：

```json
{
  "package": { "path": "examples/plugins/demo_notice", "code": "demo_notice", "manifest_found": true },
  "file_scan": { "total_files": 5, "dangerous_files": [], "unknown_files": [] },
  "checksum": { "algorithm": "sha256", "status": "ok|warning|failed|missing" },
  "manifest_validation": { "valid": true, "errors": [], "warnings": [] },
  "install_dry_run": { "valid": true, "impact_summary": {}, "install_preview": {} },
  "migration_plan": [
    { "path": "migrations/001_init.sql", "name": "001_init", "source": "migrations", "will_execute": false }
  ],
  "risk_report": { "level": "low|medium|high|blocked", "score": 0, "summary": "", "items": [] },
  "status": "ok|warning|blocked",
  "blocked_code": "plugin_package_dangerous_file|plugin_package_manifest_invalid|plugin_package_dry_run_blocked",
  "blocked_reasons": [],
  "warnings": [],
  "errors": []
}
```

路径安全规则：

- 只允许读取项目根目录下的白名单目录：
  - `examples/plugins/`
  - `plugins-local/`
  - `storage/plugins/packages/`
  - `storage/plugins/exports/`
  - `storage/plugins/staging/`
  - `storage/plugins/quarantine/`
  - `.devhub/plugins/`
- 会对 path 做 clean/normalize，拒绝 `../` 路径穿越。

## 本地插件包安装（v1.5.0-P0-04）

`POST /api/v1/admin/plugins/packages/install`

认证与权限：

- 需要 admin token。
- 直接执行安装需要 `plugin.approve`（审批人权限）；普通管理员应先通过 `POST /api/v1/admin/plugins/approvals` 提交安装审批。

用途：

- 从白名单仓库目录中选择一个**校验通过**的插件包，执行“声明型插件”安装闭环。
- 服务端会在安装前强制复跑 dry-run（scan/checksum/risk/manifest validate/install preview），不信任前端缓存结果。
- 安装后的 manifest 声明能力会进入运行态：`content_types` / `content_type_definitions`、`permissions`、`menus`、`config_schema` 等由 Core 读取并参与子站启用、发布校验、权限矩阵和后台治理展示。
- 插件仍需全局启用和子站启用后才能使用声明能力；全局 disabled、子站 disabled、archived / soft_uninstalled 会阻断新内容、菜单和新能力，历史内容 / 审计保留可查。

边界：

- 不执行第三方本地代码、不执行外部 raw SQL。
- 不动态加载前端资产、不做 Go 动态加载、不做脚本沙箱。
- 安装成功后插件默认 `disabled`，不自动启用、不自动对子站启用、不自动暴露前台入口。

安装阻断（简化）：

- `risk_report.level=blocked` / 危险文件 / checksum 不匹配 / manifest 校验失败：直接阻断。
- required 依赖不满足、Core 版本不兼容：直接阻断。
- 同编码插件已安装：阻断并提示走 upgrade。

## 本地插件仓库目录（v1.5.0-P0-03）

本地插件仓库用于集中存放多个插件包，并提供“扫描 / 列表 / 详情 / dry-run”治理能力。

推荐仓库目录（默认）：

- `storage/plugins/packages/`

仍支持扫描的白名单根目录（不允许扫描系统任意路径）：

- `examples/plugins/`（示例插件包目录）
- `plugins-local/`（本地私有包目录）
- `storage/plugins/packages/`（推荐仓库目录）
- `.devhub/plugins/`（开发/测试目录）

扫描规则（简化）：

1. 只扫描仓库目录下 **一级子目录**；每个一级子目录视为一个插件包。
2. 插件包目录中必须有 `manifest.json`；缺失时在列表中标记为 `invalid`。
3. 插件包扫描/校验仍复用 `ScanPluginPackage + checksums + risk_report + dry-run` 逻辑。
4. 扫描过程不执行任何文件，不写入插件/配置/migration 表，不动态加载 assets。

### checksums.json（sha256）

`checksums.json` 用于校验插件包文件完整性（不做签名、仅做 sha256 摘要比对）。

建议结构：

```json
{
  "algorithm": "sha256",
  "files": [
    { "path": "manifest.json", "sha256": "..." }
  ]
}
```

规则（简化）：

1. 仅支持 `sha256`；`algorithm` 缺失默认 `sha256`。
2. `checksums.json` 缺失：warning（不阻断）。
3. 非法 JSON / 不支持算法 / 重复 path / 声明文件不存在：blocked。
4. checksum 不匹配：blocked。
5. 包内存在未被 checksums.json 覆盖的文件：当前先 warning（后续版本可能升级为 blocked）。
6. `path` 必须是相对路径，禁止 `..` 与绝对路径；必须指向插件包内文件。

### 风险报告（risk_report）

Dry-run 会返回 `risk_report`：

- `level`：`low|medium|high|blocked`
- `score`：0~100（当前为固定区间分值）
- `items`：风险项明细（`code/level/path/message/suggestion`）

规则（简化）：

- `blocked`：危险文件、路径穿越、超限、checksum 不匹配/非法、manifest 非法等。
- `high`：checksum 覆盖不完整、optional 依赖缺失、Core 兼容 warning 等。
- `medium`：缺少 checksums.json、存在未知文件等。
- `low`：无错误无告警。

## 后台页面

入口：`系统插件 -> 安装升级`

能力：

- 输入本地相对路径并执行扫描 / dry-run。
- 展示文件扫描统计（allowed / unknown / dangerous）。
- 展示 manifest 校验结果与安装预览（复用现有 manifest validate / dry-run 逻辑）。
- 显示“不会安装 / 不会执行代码 / 不会执行 SQL / 不会动态加载前端资产”的边界提示。

限制：

- 后台安装升级页当前提供本地插件仓库扫描、详情、dry-run 和安装审批入口；直接安装需要 `plugin.approve`，普通管理员走审批提交。
- zip 上传区域只提供上传扫描、详情和 promote，不提供上传后直接安装按钮。
- 不提供远程市场入口。

## 示例插件包

示例路径：

- `examples/plugins/demo_notice/`
 - `examples/plugins/demo_signed_notice/`（签名/可信来源示例，见下文）

包含：

- `manifest.json`
- `README.md`
- `config.example.json`
- `docs/usage.md`
- `migrations/001_init.sql`（只进入预检/计划，不执行；新 SQL 迁移统一放在 `migrations/`）

## 插件包签名与可信来源（v1.6.0-P0-03）

本章节为后续“插件市场/可信分发”打基础：在不引入远程市场、不自动下载、不做 PKI 证书链/公钥服务器的前提下，实现 `signature.json` Ed25519 真实验签，并由后台可信发布者记录决定 `trust_status`。

边界（本轮明确不做）：

- 不从远程拉取 publisher 信息，不自动信任包内 `publisher.json`。
- 不做证书链验证、不做密钥轮换、不做私钥管理后台或在线签名服务。
- 不执行任何第三方代码/SQL，不动态加载前端资产。

## 插件包 detached signature（v1.7.1）

v1.7.0 引入“远程包下载到 staging + 预检 + compat-check + 安装/启用治理”，但仍无法解决“来源可信/链路未被篡改”。v1.7.1 增加 detached signature（与插件包分离的签名元数据文件）：

- 文件名：`devhub-signature.json`
- 来源：可来自下载请求中的 `signature_url`，或包内同名文件；若两者同时存在且内容不一致，则验签失败（阻断）。
- 算法：Ed25519（base64 public key + base64 signature）。
- payload：对 `signature_payload` 的 canonical JSON bytes 进行签名与验签（payload 必须是稳定结构体，禁止 map）。

最小 payload 绑定字段：

- `plugin_code` / `version` / `compatible_core_version`（如存在）必须与 `manifest.json` 一致。
- `package_sha256` 必须等于 staging download 的 `sha256_actual`。
- `manifest_sha256` 必须等于预检目录中 `manifest.json` 的实际 sha256。
- `publisher_id + key_id` 必须匹配本地可信发布者记录（`plugin_trusted_publishers`），且状态为 `trusted`、未过期（`expires_at`）。

默认策略：

- unsigned 包允许停留在 staging / precheck，但 **默认禁止** 进入 install / upgrade 链路（compat-check 会阻断 `can_install`）。
- 仅在 dev 场景可通过 `DEVHUB_PLUGIN_REQUIRE_SIGNED_PACKAGES=0` 放开 unsigned（仍会产生审计与风险提示；生产不建议开启）。

### 目录结构新增文件

插件包可选包含：

```text
publisher.json
signature.json
```

说明：

- 缺少 `publisher.json` / `signature.json` **不阻断** dry-run，但会产生风险 warning。
- `publisher.json` / `signature.json` 非法 JSON 会 **blocked**。

### publisher.json（发布者声明）

示例结构：

```json
{
  "publisher_id": "devhub-official",
  "name": "DevHub Official",
  "homepage": "https://example.com",
  "email": "security@example.com",
  "public_key_id": "devhub-official-2026",
  "public_key_algorithm": "ed25519",
  "public_key": "base64-public-key",
  "trust_level": "unknown"
}
```

规则（简化）：

- `publisher.json` 仅作为**包内声明**，不代表可信；可信状态由本地 `trusted_publishers` 决定。
- 禁止存放私钥；后端不会返回任何私钥字段。

### signature.json（签名声明）

示例结构：

```json
{
  "version": "1",
  "algorithm": "ed25519",
  "signed_at": "2026-01-01T00:00:00Z",
  "publisher_id": "devhub-official",
  "public_key_id": "devhub-official-2026",
  "payload_algorithm": "sha256",
  "payload": "checksums.json",
  "signed_files": [
    "manifest.json",
    "checksums.json",
    "README.md",
    "config.example.json"
  ],
  "signature": "base64-signature"
}
```

规则（简化）：

- 当前仅支持 `algorithm=ed25519`；不支持算法会 **blocked**。
- 当前仅支持 `payload=checksums.json` 与 `payload_algorithm=sha256`。
- `signed_files` 必须是相对路径，禁止 `..`、禁止绝对路径；声明不存在文件会 **blocked**。
- `signed_files` 必须至少包含 `manifest.json` 与 `checksums.json`；否则 **blocked**。
- 为避免“签名覆盖未校验文件”，除 `manifest.json` / `checksums.json` 外，`signed_files` 中的文件必须出现在 `checksums.json` 覆盖列表中；否则 **blocked**。

签名对象（当前实现）：

- `message = sha256(raw_bytes_of_checksums.json)`（注意：是 checksums.json 的原始 bytes，不做 canonical JSON）
- `signature = ed25519.Sign(private_key, message)`
- 验签使用后台可信发布者公钥；如果 publisher 未建立本地可信关系，可使用包内 `publisher.json` 公钥做技术验签，但 `trust_status` 仍为 `unknown`。

### 本地可信发布者 trusted publishers

可信来源只来自后台可信发布者记录（不会从远程同步）。历史本地配置文件可作为空存储时的 seed / fallback：

- `storage/plugins/trusted_publishers.json`

示例结构：

```json
{
  "publishers": [
    {
      "publisher_id": "devhub-official",
      "name": "DevHub Official",
      "public_key_id": "devhub-official-2026",
      "public_key_algorithm": "ed25519",
      "public_key": "base64-public-key",
      "status": "trusted",
      "notes": "DevHub 官方示例发布者"
    }
  ]
}
```

规则（简化）：

- 后台持久化字段：`publisher_id`、`name`、`homepage`、`email`、`public_key_id`、`public_key_algorithm`、`public_key`、`fingerprint`、`status`、`notes`、`created_by`、`updated_by`、时间戳与 metadata。
- 只有 `publisher_id + public_key_id` 与后台记录匹配，且 `status=trusted`，才会标记 `trust_status=trusted`。
- 若后台记录标记为 `blocked` / `revoked`：dry-run / promote / install / approval 执行直接 **blocked**（即使包内自称可信或验签通过）。
- 若 publisher 不在后台可信发布者中：`trust_status=unknown`（不阻断，但会产生 high 风险提示）。
- 后台不会保存私钥；`public_key` 可展示并计算 `fingerprint=sha256:<base64url>`。

可信发布者 API：

- `GET /api/v1/admin/plugins/trusted-publishers`：列表，需 `plugin.read`。
- `GET /api/v1/admin/plugins/trusted-publishers/:id`：详情，需 `plugin.read`。
- `POST /api/v1/admin/plugins/trusted-publishers`：新增，需 `plugin.manage`。
- `PUT /api/v1/admin/plugins/trusted-publishers/:id`：更新，需 `plugin.manage`。
- `POST /api/v1/admin/plugins/trusted-publishers/:id/block`：block，需 `plugin.manage`。
- `POST /api/v1/admin/plugins/trusted-publishers/:id/revoke`：revoke，需 `plugin.manage`。
- `POST /api/v1/admin/plugins/trusted-publishers/:id/restore`：恢复 trusted，需 `plugin.manage`。
- `DELETE /api/v1/admin/plugins/trusted-publishers/:id`：删除本地记录，需 `plugin.manage`。

### 签名状态与风险报告联动

dry-run / 仓库扫描 / 详情接口会返回 `signature` 字段摘要，并纳入 `risk_report`：

- `trusted + verified`：不新增风险项（仍以 checksum/危险文件等为准）。
- `unsigned / signature missing`：warning（风险至少 medium）。
- `unknown publisher + verified`：风险至少 high（提示仅技术验签通过，未建立可信来源）。
- `publisher_unknown`：high（缺少可信或包内公钥，无法完成真实验签）。
- `unsupported algorithm / verification failed / publisher blocked|revoked / signature invalid`：blocked。

流程联动：

- package dry-run、仓库扫描、zip upload / upload detail 均返回真实签名验签结果。
- promote 前重新验签；`failed / blocked / revoked` 阻断 promote。
- install 前重新验签；`failed / blocked / revoked` 阻断安装，`unsigned / unknown` 至少 warning。
- approval 创建保存签名快照；approval execute 前重新验签，签名状态变化会进入风险与错误提示。

## 已安装插件导出为本地插件包（v1.5.0-P2-10）

目标：把已安装的**声明型插件**导出为可被本地插件包 dry-run / 仓库扫描识别的目录，便于备份、迁移、二次分发和后续仓库治理。导出不等于发布到市场，也不会导出运行时代码或真实配置。

### 导出目录结构

默认输出到受控目录：

- `storage/plugins/exports/{plugin_code}-{version}-{timestamp}/`

导出包结构：

```text
exported-plugin/
├─ manifest.json
├─ README.md
├─ config.example.json
├─ checksums.json
├─ publisher.json              # 可选，仅草案结构
├─ signature.json              # 可选，仅草案结构
├─ docs/
│  └─ usage.md                 # 可选
└─ migrations/
   └─ exported_migrations.json # 可选，声明型摘要，不含外部 SQL
```

`storage/plugins/exports/` 已加入本地插件包允许读取目录，因此导出后可以直接用 package dry-run 自检；如需作为正式仓库包，可复制到 `storage/plugins/packages/`。

### 导出范围

会导出：

- `manifest.json`：来自已安装插件当前 manifest / registry / 数据库声明。
- `README.md`：自动生成，包含插件名称、code、版本、描述、内容类型、权限、配置、依赖、安装方式、安全边界、导出时间和 DevHub Core 版本。
- `config.example.json`：示例配置，不是当前环境配置备份。
- `checksums.json`：sha256，覆盖核心导出文件。
- 可选 `docs/usage.md`、`migrations/exported_migrations.json`、`publisher.json`、`signature.json` 草案。

不会导出：

 - 敏感配置明文或 `enc:v1:` / `enc:v2:` 密文。
- 当前全局配置真实值、子站覆盖配置真实值。
- 用户数据、帖子、评论、通知、审计原始明细、Hook 执行历史、搜索索引。
- 插件运行时代码、动态前端资产、外部 SQL、私钥、token、secret、password、credential。
- zip 包、远程市场发布信息或上传任务。

### manifest.json 规则

- 保留声明型字段：`code/name/version/description/author/license/min_core_version/compatible_core_version/content_types/content_type_definitions/permissions/menus/routes/config_schema/dependencies/hooks/events/notification_templates/search/migrations`。
- 不导出运行时状态字段作为安装默认值，例如 `enabled/disabled/archived`、`status_reason`、安装来源等。
- 导出时会把包内安装默认状态保持为普通 manifest 声明；插件安装后的 enabled/disabled 仍由目标环境安装流程决定。
- 导出的 manifest 必须通过当前 manifest validate；同 code 已安装只影响目标环境安装，不影响导出包本身生成。

### config.example.json 脱敏规则

- 优先使用 `default_config`；缺失时按 `config_schema` 生成示例。
- 敏感字段统一写为 `REPLACE_ME`。
- 敏感字段识别复用插件配置规则：`x-sensitive`、`writeOnly`、`format=password` 以及字段名命中 `password/passwd/secret/token/api_key/key/credential/app_secret/aes_key/private_key/client_secret` 等。
- boolean / number / enum 使用默认值或安全示例；object 生成对象结构；array 为空数组。
- 不读取当前生产配置真实值，不读取子站配置真实值，不输出密文。

### checksums.json 生成规则

- 固定使用 `sha256`。
- 覆盖 `manifest.json`、`README.md`、`config.example.json`、`docs/**/*.md`、`migrations/**/*.sql`、`migrations/exported_migrations.json`、`publisher.json`、`signature.json`、`package.json`（如存在）。
- `checksums.json` 不包含自身。
- 路径使用包内相对路径，按字典序稳定排序。
- 导出后 package dry-run 应能通过 checksum 校验；若自检 warning，会在导出结果中展示。

### API

- `POST /api/v1/admin/plugins/:code/export/dry-run`
  - 权限：`plugin.read`
  - 不写文件，只返回预计文件列表、输出目录和安全边界检查。
- `POST /api/v1/admin/plugins/:code/export`
  - 权限：`plugin.write`
  - 正式写入 `storage/plugins/exports/`，生成 checksums，并自动执行 package dry-run 自检。

### 后台入口

插件详情抽屉 Overview 中提供“导出本地插件包”面板：

- 支持选择 `include_docs`、`include_migrations`、`include_publisher`、`include_signature_stub`。
- 先执行 dry-run，展示将导出的文件、安全检查、输出目录、warning/error。
- dry-run 未 blocked 后才允许正式导出。
- 导出成功展示输出目录、checksums 状态、package dry-run 自检状态，并提供复制路径。
- 页面明确不提供 zip 下载、远程发布或插件市场发布入口。

## 远程插件索引只读镜像（v1.6.0-P0-04）

远程插件索引是未来插件市场前的只读元数据镜像能力。系统只拉取一个静态 `index.json`，展示远程插件包 code、name、版本、`package_url`、`package_sha256`、`signature_url`、publisher、license、Core 兼容性和风险提示。

当前边界：不下载远程包、不安装远程插件、不自动更新、不执行远程代码、不执行 SQL、不动态加载前端资产、不自动信任远程 publisher。`package_sha256` 只是远程源声明，不代表 DevHub 已校验远程包内容。

索引规范、API、安全限制、SSRF 防御、风险规则和示例见 [PLUGIN_REMOTE_INDEX.md](PLUGIN_REMOTE_INDEX.md)。示例索引文件为 `docs/examples/plugin-remote-index.example.json`。

## 插件包版本仓库与升级差异对比（v1.6.0-P0-05）

版本仓库是 DevHub 对同一插件多版本来源的统一治理视图，不是远程市场，也不会下载或自动安装插件包。当前聚合来源：

- `installed`：当前已安装插件版本。
- `local_package`：`storage/plugins/packages/` 本地插件仓库中的插件包版本。
- `uploaded_package`：zip 上传 staging / promoted 上传包记录。
- `remote_index`：远程只读索引中的版本元数据，只读展示，不能直接升级。
- `exported_package`：已导出的本地插件包（如后续存在导出记录，可纳入聚合）。

版本比较规则：支持 `x.y.z` 和可选 `v` 前缀；目标版本必须高于当前版本。相同版本返回 `plugin_version_same_version`，低版本返回 `plugin_version_downgrade_forbidden`。复杂 semver 预发布标签不是本轮目标。

升级差异对比通过 `POST /api/v1/admin/plugins/:code/versions/:version/upgrade-diff` 执行。该接口会读取目标本地包 manifest，并复用 package dry-run、checksum、signature、risk_report、manifest validate 和 Core / dependency 校验；不会写入插件表，不会执行代码、SQL 或前端资产。

`diff_sections` 按功能分组返回：基础信息、内容类型、内容类型定义、权限、菜单、路由、配置 schema、默认配置、依赖、Hook、迁移声明。每个 diff item 都包含 `section`、`path`、`type`、`risk_level`、`before`、`after` 和 `message`，敏感字段统一脱敏为 `[REDACTED]`。

高风险变更包括删除 content_type / permission、 新增高危权限、新增 required dependency、依赖约束收紧、config_schema 删除字段 / type 变化 / required 新增、Hook 改为 blocking、新增 migration。阻断项包括目标版本不升、Core 不兼容、required dependency 缺失、checksum mismatch、签名失败、publisher blocked / revoked、manifest invalid、dangerous file 和 remote_index 直接升级。

后台 `/admin-next/plugins/versions` 展示版本仓库、单插件版本列表和升级差异抽屉；远程索引版本会标记只读，必须先通过受控上传 / promote 进入本地仓库后才能参与真实升级流程。

## 后台插件包治理分组（v1.6.0-P1-09）

后台“系统插件”已把插件包相关能力集中到“插件包治理”分组：本地插件仓库、zip 上传包、插件包安装、插件包导出、版本仓库和升级差异。安全相关能力集中到“安全与可信”：风险报告、签名验签、可信发布者、敏感配置加密和密钥轮换。

页面文案继续强调：zip 上传只是进入 staging，promote 只是转入本地插件仓库，安装仍需服务端重新 dry-run / checksum / signature / risk_report / 权限和审批校验；系统不会执行第三方代码、外部 SQL 或动态前端资产。


## v1.6.0 收口限制：zip 下载导出

当前已支持将已安装声明型插件导出为 `storage/plugins/exports/` 下的本地插件包目录，并生成 `manifest.json`、`README.md`、脱敏 `config.example.json`、`checksums.json` 和可选 publisher / signature 结构草案。当前不提供 zip 下载包、不提供在线签名打包服务，也不保存或导出私钥；zip export 下载能力登记为 v1.7 后续任务。

## v1.7.0-P0-01：远程插件包安全下载到 staging

本轮新增的是远程插件安装链路的 P0 第一步：把远程插件包安全下载到 staging，仍不安装插件。

目录与记录：

- staging 目录：`storage/plugins/staging/downloads/`。
- 表：`plugin_package_downloads`，记录 `plugin_code/version/source_url/final_url/status/file_size/sha256_expected/sha256_actual/content_type/staging_path/error_message/downloaded_at`。
- staging 文件名由 `plugin_code`、`version`、`sha256` 前缀和时间戳生成，不使用远程文件名，避免路径穿越和文件名注入。
- 下载失败会清理临时 `.part` 文件；成功后保留 staging 文件等待后续检查 / 解压任务处理。

下载安全规则：

- 仅允许 `https://`。
- 禁止 `file://`、`http://`、`ftp://`、`gopher://` 等协议。
- 禁止 localhost、`127.0.0.0/8`、`0.0.0.0`、`::1`、`10.0.0.0/8`、`172.16.0.0/12`、`192.168.0.0/16`、`169.254.0.0/16`、`fc00::/7`、`fe80::/10` 等本机 / 内网 / link-local 地址。
- 最多 3 次重定向；每次重定向都会重新执行 URL、DNS 和 IP 安全校验。
- 默认最大下载 20MB，可通过 `DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_MAX_BYTES` 调整。
- 当前允许 `.zip`、`.tar.gz`、`.tgz`；不支持可执行文件、脚本、共享库或未知二进制格式。

sha256 规则：

- 如果远程索引或请求提供 `sha256`，下载后必须匹配，否则状态为 `checksum_failed`，清理文件并禁止后续安装。
- 如果未提供 `sha256`，状态为 `checksum_missing`；文件保留用于人工排查，但不能进入自动安装链路。
- sha256 以实际下载字节计算，不信任 `Content-Length`。

边界：

- 不安装、不启用、不解压安装、不运行 package scripts、不执行第三方代码、不执行外部 SQL、不加载 Go plugin、不动态加载前端资产。
- 本轮不做插件市场、远程更新、依赖安装、完整 sandbox 或签名新能力；如果已有签名机制，后续阶段可在解压 / dry-run 中复用。

## v1.7.0-P0-03：插件依赖 / 兼容性检查

本阶段是远程插件安装链路的安装前闸门。输入必须来自上一阶段预检通过的 `plugin_package_prechecks` 记录，且 `status=passed`；预检失败、manifest 无效、下载记录已失败 / rejected / checksum_failed / deleted 或 staging 文件丢失时，不能执行兼容性检查。

数据模型：

- `plugin_package_prechecks`：保存预检来源、manifest_json、package/staging 路径、checksum_status 和错误信息。本轮只补齐 compat-check 需要的最小持久化模型，不把完整解压预检 UI 写成已完成。
- `plugin_package_compat_checks`：保存 compat-check 历史记录，包括 Core 兼容、依赖、冲突、权限、菜单、路由、Hook、config_schema、migration、warnings、errors、`can_install` 和 summary。

检查规则：

- Core：`compatible_core_version` 必须存在，当前 Core 来自 `VERSION`。支持 `1.7.0`、`>=1.7.0`、`>=1.7.0 <2.0.0`、`^1.7.0`、`~1.7.0`；不支持复杂预发布标签、`||` 或 npm 全量 semver。
- 依赖：`dependencies` 支持字符串数组、对象数组和 `{ "plugins": [...] }`。required 依赖缺失或版本不满足会阻断；optional 依赖缺失只 warning；不会自动下载、安装或启用依赖。
- `plugin_code` / `content_type`：不能抢占内置插件、已安装插件或现有内容类型；同一 manifest 内重复声明会失败。
- 权限：权限码必须以 `plugin_code.` 开头，禁止声明 `core.*`、`admin.*`、`system.*` 或其他插件前缀；scope 仅允许 `global/community/channel/content`。
- 菜单 / 路由：必须是站内路径，禁止外链、`javascript:`、`data:` 和路径穿越；禁止覆盖 `/admin-next`、`/login`、`/register`、`/topics`、`/c`、`/api/v1/admin/plugins/*`、`/api/v1/admin/users/*`、`/api/v1/admin/system/*` 等核心入口。
- Hook：Hook 名称必须属于当前 HookBus；mode 只能 `blocking/non_blocking`；failure_policy 只能 `block/log/ignore`，blocking Hook 禁止 `ignore`。
- config_schema：只验证 DevHub 当前简化 JSON Schema 能力；`default_config` 如存在必须满足 schema。
- migrations：manifest 声明只允许 `direction=up`；插件包文件迁移只认 `migrations/`，本轮只记录 pending / 生成计划，不执行 migration，不支持 migration down。

结论规则：

- 有 errors 时 `can_install=false`。
- 只有 warnings 时 `can_install=true` 且 `status=warning`。
- Core 不兼容、required 依赖缺失、依赖版本不满足、plugin_code/content_type/permission/route 冲突、未知 Hook、default_config 不符合 schema、migration 不兼容均会阻断。
- 没有 sha256 或 `checksum_missing` 的包默认不能进入安装链路。

API：

- `POST /api/v1/admin/plugins/packages/prechecks/:id/compat-check`
- `GET /api/v1/admin/plugins/packages/compat-checks`
- `GET /api/v1/admin/plugins/packages/compat-checks/:id`
- `DELETE /api/v1/admin/plugins/packages/compat-checks/:id`

后台“上传包管理”页新增最小兼容性检查入口：管理员输入 precheck ID 后执行检查，列表展示 status、`can_install`、blockers 和 warnings。页面不提供安装 / 启用按钮。

边界：本轮不会安装插件、不会启用插件、不会注册权限 / 菜单 / 路由 / Hook、不会执行 migration、不会运行包内脚本、不会加载 Go plugin、不会安装 npm / composer / go 依赖，也不会修改现有 qa/docs/wiki/projects/jobs/ai_works 业务逻辑。

## v1.7.0-P0-05：插件启用前安全检查（enable-precheck）

启用前安全检查是启用流程的最后一道闸门：对已安装但未启用插件执行文件完整性复检、manifest 再校验、依赖/配置/迁移状态复检与冲突复检，输出 `can_enable` 结论。

输入来源与约束：

- 仅允许对“已安装且未启用”的插件执行；不会修改插件状态。
- 为避免跳过远程包链路，启用前检查强制要求存在最近的 `plugin_package_prechecks(status=passed)` 且对应 `plugin_package_compat_checks(can_install=true)`。
- 对 `source_type=local_package` 的插件，会基于预检记录中的 `package_path` 复检文件完整性；其他来源（builtin/manifest）不做包文件复检。

复检范围（只检查，不执行）：

1. 文件完整性：重新执行 package scan、危险文件检测、checksums 校验；读取 `manifest.json` 并与安装时 `manifest_checksum` 比对，不一致将阻断。
2. manifest 再校验：复用 `manifest validate` 规则，包含 Core 兼容范围、敏感路径保护、Hook 白名单与 config_schema/migrations 合法性。
3. 依赖再检查：required 依赖缺失/版本不满足阻断；optional 缺失仅 warning；不会自动安装/启用依赖。
4. 配置有效性：当前 `config_json` 必须满足 `config_schema`，无效则阻断。
5. 迁移状态：pending/failed 迁移阻断；本轮不执行 migration。
6. 冲突复检：permissions/menus/routes/hooks/content_types 与当前已启用插件/核心声明冲突将阻断（或按规则 warning）。

边界：

- 本轮只做检查，不启用插件、不注册运行时、不开放前台/后台入口、不允许创建内容、不执行插件代码/脚本、不执行 migration。
