# Declarative Content Plugin Template

这是 DevHub 官方纯声明型内容插件模板，基于 `official_links` / 友情链接能力整理，用于展示“不需要外部服务、不执行代码”的插件如何声明内容类型、权限、菜单和配置。

## 适用场景

- 友情链接。
- 简单内容块。
- 站点配置扩展。
- 后台菜单扩展。
- 不需要外部服务的声明型插件。

## 包内有什么

- `manifest.json`：插件能力声明。
- `config.example.json`：示例配置，不包含真实 secret。
- `checksums.json`：文件完整性摘要，便于预检。
- `migrations/001_init.sql`：迁移入口示例；dry-run 只生成计划。
- `README.md`：模板说明。
- `PACKAGING.md`：从模板到插件包的最小流程。

## 安全边界

- 插件包不包含可执行代码。
- 插件不能执行根目录 SQL；`migrations/` 是唯一迁移入口。
- 插件不能动态加载未知资产。
- 插件不能注册远程 iframe。
- 插件不能执行第三方后端逻辑。
- 插件 disabled / archived 后会阻断新能力，历史内容仍按 Core 边界保留。

## 使用步骤

1. 复制本目录并改名为你的插件编码。
2. 修改 `manifest.json` 的 `code`、`name`、`version`、权限码、content_type 和菜单路径。
3. 修改 `config.example.json`，不要写真实密钥。
4. 需要迁移时只在 `migrations/` 下追加文件；不需要迁移时保留说明文件。
5. 打包 zip 后走后台插件包治理：upload -> precheck -> promote -> install dry-run -> install。
6. 安装后在后台启用插件，并按子站启用、配置和验证菜单 / content_type。

DevHub 只读取可信声明元数据，并通过 Core 的权限、状态、审计和生命周期治理这些能力。
