# DevHub 本地插件包规范（草案）

[返回文档入口](README.md)

更新时间：2026-05-14

本文件是 **v1.5.0-P0-01 ~ P0-04** 的阶段性成果：为 DevHub 定义“本地插件包 / 本地插件仓库”规范，并提供 **dry-run 导入预览** 与 **最小安装闭环** 能力。

注意：

- P0-01 ~ P0-03：只做 **安全读取 + 校验 + 预览**（dry-run），不安装插件。
- P0-04：补齐“从本地插件包安装声明型插件”的最小闭环，但仍然 **不执行插件代码、不执行外部 SQL、不动态加载前端资产**。

## 目标与边界

目标：

- 定义本地插件包目录结构与字段口径。
- 后台支持从允许目录读取插件包，并返回 dry-run 预览信息（文件扫描 / manifest 校验 / 安装预览）。

明确不做：

- 不做 zip 上传。
- 不做远程市场 / 远程下载 / 在线更新。
- P0-01 ~ P0-03 不做正式安装本地插件包（只做 dry-run）。
- 不做动态加载 Go 代码，不执行 JS / WASM / Lua，不提供脚本沙箱。
- 不执行任何第三方本地代码。
- 不执行外部 raw SQL。
- 不动态加载前端资产。
- 不做插件签名与可信发布者体系（后续版本再做）。

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
│  └─ 001_init.json
├─ assets/
│  └─ preview.png
└─ checksums.json
```

规则：

1. `manifest.json` 是唯一主声明文件（必须）。
2. `README.md` 是插件说明（建议）。
3. `config.example.json` 是示例配置（建议），不得包含真实 secret。
4. `migrations/` 当前只允许声明型迁移描述文件；dry-run 不执行任何 SQL。
5. `assets/` 当前只允许预览素材；dry-run 不会动态加载到前台。
6. `checksums.json` 当前支持 **sha256 完整性校验**（不做签名校验）。
7. 目录名建议与 `manifest.code` 一致；不一致会 warning。
8. 未知文件默认 warning；危险文件会 blocked。

> 注：v1.5.0 当前只支持“声明型插件”安装（写入 manifest/配置/迁移记录/审计），不支持携带代码或可执行资产的插件包。

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

- 后台安装升级页当前只提供“提交安装审批”和“审批中心”入口，不提供本地插件包一键正式安装按钮；正式安装由审批执行 API 完成。
- 不提供 zip 上传按钮。
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
- `migrations/001_init.json`（声明示例，不执行）

## 插件包签名与可信来源（草案，v1.5.0-P2-09）

本章节为后续“插件市场/可信分发”打基础：在不引入远程市场、不自动下载、不做 PKI 证书链/公钥服务器的前提下，定义插件包签名元数据与本地可信来源配置，并将签名状态纳入 dry-run / 风险报告。

边界（本轮明确不做）：

- 不从远程拉取 publisher 信息，不自动信任包内 `publisher.json`。
- 不做证书链验证、不做可信发布者平台、不做密钥轮换。
- 不执行任何第三方代码/SQL，不动态加载前端资产。

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

- 当前仅支持 `ed25519`；不支持算法会 **blocked**。
- `signed_files` 必须是相对路径，禁止 `..`、禁止绝对路径；声明不存在文件会 **blocked**。
- `signed_files` 必须至少包含 `manifest.json` 与 `checksums.json`；否则 **blocked**。
- 为避免“签名覆盖未校验文件”，除 `manifest.json` / `checksums.json` 外，`signed_files` 中的文件必须出现在 `checksums.json` 覆盖列表中；否则 **blocked**。

签名对象（当前实现）：

- `message = sha256(raw_bytes_of_checksums.json)`（注意：是 checksums.json 的原始 bytes，不做 canonical JSON）
- `signature = ed25519.Sign(public_key, message)`

### 本地可信来源 trusted_publishers

可信来源只来自本地配置文件（不会从远程同步）：

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

- 只有 `publisher_id + public_key_id` 与本地配置匹配，且本地配置 `status=trusted`，才会标记 `trust_status=trusted`。
- 若本地配置标记为 `blocked` / `revoked`：dry-run 直接 **blocked**（即使包内自称可信或验签通过）。
- 若 publisher 不在本地配置中：`trust_status=unknown`（不阻断，但会产生高风险提示）。

### 签名状态与风险报告联动

dry-run / 仓库扫描 / 详情接口会返回 `signature` 字段摘要，并纳入 `risk_report`：

- `trusted + verified`：不新增风险项（仍以 checksum/危险文件等为准）。
- `unsigned / signature missing`：warning（风险至少 medium）。
- `unknown publisher + verified`：风险至少 high（提示仅技术验签通过，未建立可信来源）。
- `unsupported algorithm / verification failed / publisher blocked|revoked`：blocked。

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

- 敏感配置明文或 `enc:v1:` 密文。
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
- 覆盖 `manifest.json`、`README.md`、`config.example.json`、`docs/**/*.md`、`migrations/**/*.json`、`publisher.json`、`signature.json`、`package.json`（如存在）。
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
