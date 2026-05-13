# DevHub 本地插件包规范（草案）

[返回文档入口](README.md)

更新时间：2026-05-13

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

- 不提供“正式安装本地插件包”按钮（后续版本再做）。
- 不提供 zip 上传按钮。
- 不提供远程市场入口。

## 示例插件包

示例路径：

- `examples/plugins/demo_notice/`

包含：

- `manifest.json`
- `README.md`
- `config.example.json`
- `docs/usage.md`
- `migrations/001_init.json`（声明示例，不执行）
