# DevHub Plugin Template

[返回文档入口](README.md)

更新时间：2026-05-14

本页说明两种声明型插件模板初始化方式：

1. CLI：`go run ./cmd/devhub plugin:new ...`
2. 后台：`系统插件 -> 安装升级 -> 初始化插件包`

两者复用同一套后端模板生成逻辑，适合准备 manifest、配置示例、权限说明、Hook 声明和 migration 声明；都不会生成可动态执行的第三方运行时能力。

## CLI 命令

```bash
go run ./cmd/devhub plugin:new \
  --code demo_links \
  --name "Demo Links" \
  --content_type demo_link \
  --content_name "演示链接" \
  --description "用于展示 DevHub 声明型插件模板的示例。" \
  --author "DevHub Team" \
  --output examples/plugins \
  --with_config \
  --with_hooks \
  --with_migration
```

可选参数：

- `--code`：必填，插件编码。
- `--name`：必填，插件名称。
- `--content_type`：内容类型；不传时按 code 推导。
- `--content_name`：内容类型名称。
- `--description`：插件说明。
- `--author`：作者。
- `--with_config`：生成 `config_schema` 示例，默认开启。
- `--with_hooks`：生成 Hook 声明示例，默认开启。
- `--with_migration`：生成 migration 声明示例，默认开启。
- `--output`：输出父目录，默认 `examples/plugins`。
- `--force`：允许覆盖已有目录。

## 生成目录

```text
examples/plugins/demo_links/
  manifest.json
  README.md
  config.example.json
  content-type.md
  permissions.md
  hooks.md
  migrations.md
  registry.example.go
```

说明：CLI 默认仍生成 `registry.example.go`，用于解释“内置系统插件需随 DevHub 源码显式注册并编译发布”。该文件不会被系统动态扫描或加载；如果要把 CLI 生成目录直接放入本地插件仓库 dry-run，请删除 `.go` 文件或改用后台初始化入口。

## 后台初始化插件包

入口：

```text
/admin-next/plugins/install
系统插件 -> 安装升级 -> 初始化插件包
```

表单字段：

- `code`
- `name`
- `content_type`
- `content_name`
- `description`
- `author`
- `with_config`
- `with_hooks`
- `with_migration`

后台初始化固定写入：

```text
storage/plugins/packages/{code}/
  manifest.json
  README.md
  config.example.json
  content-type.md
  permissions.md
  hooks.md
  migrations.md
  docs/
    registry-example.md
```

后台初始化和 CLI 的差异：

- 后台不允许任意输出路径，固定为 `storage/plugins/packages/{code}`。
- 后台不暴露 `force`，目录已存在时拒绝覆盖。
- 后台默认不生成 `registry.example.go`，避免 `.go` 文件被插件包扫描器识别为 dangerous file。
- registry 接入说明写入 `docs/registry-example.md`。
- 初始化成功后服务端自动执行 package dry-run，返回路径、风险等级、manifest 校验结果、warnings/errors。
- dry-run 未 blocked 时，后台可继续提交“安装审批”；直接安装仍按当前权限要求由 `plugin.approve` 执行。

说明：模板目录本身也符合“本地插件包目录规范”的最小形态，可用于后台“本地插件包 dry-run”预览（见 `docs/PLUGIN_PACKAGE.md`）。已安装的声明型插件也可以从后台导出为同类目录（manifest、README、脱敏 `config.example.json`、`checksums.json`），但导出不会包含真实敏感配置、用户数据、运行时代码或外部 SQL。

## 校验规则

- `code` 不能为空。
- `code` 只能使用小写字母、数字、下划线，并以小写字母开头。
- `name` 不能为空。
- `content_type` 不能为空。
- `content_type` 使用同样的编码规则。
- 输出目录已存在时默认拒绝覆盖。
- `--force` 会删除并重新生成该插件目录。
- 生成的 JSON 会格式化。
- 生成后的 `manifest.json` 会调用项目现有 `PluginManifestValidator` 校验，包括 dependencies 与 Core 兼容字段。
- 生成的 `config.example.json` 会调用当前简化 `config_schema` 校验。


## dependencies 模板口径

模板默认生成空 `dependencies`。开发者可手动增加：

```json
{ "code": "qa", "version": ">=1.0.0", "required": true, "reason": "依赖问答内容类型" }
```

当前只支持数字 `x.y.z` 约束、精确版本、比较符和空格组合范围；不支持自动安装依赖、插件市场推荐、远程下载依赖或动态加载依赖代码。

## 模板边界

模板适合：

- 准备 manifest + 配置型插件声明。
- 与后台 manifest validate / dry-run / install / upgrade dry-run 流程配合。
- 沉淀内容类型、权限、菜单、Hook 和 migration 设计。

模板不支持：

- 动态执行 Go / JS / WASM / Lua 代码。
- 上传插件包。
- 插件市场。
- 远程安装或在线更新。
- 外部 raw SQL。
- migration down。
- 远程 Webhook 执行。
- 前端资产动态加载。

CLI 输出中的 `registry.example.go` 只是内置系统插件接入示例。它不会被系统扫描或动态加载；后台初始化版不会生成该 `.go` 文件，而是生成 `docs/registry-example.md` 说明。
