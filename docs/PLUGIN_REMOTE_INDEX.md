# DevHub 远程插件索引只读镜像规范（v1.6.0-P0-04）

本规范定义 DevHub 读取远程插件索引元数据的只读能力。远程索引在 Core + 插件服务底座中的定位是 **插件分发信息源**：当前阶段只读展示远程插件包元数据，为后续安全下载到 staging、包校验、trusted publisher 校验和插件生态做准备。

远程索引不是插件市场，也不是远程安装能力；系统不会因为读取索引而下载 `package_url`，不会安装远程插件，不会执行远程包内容，不会动态加载前端资产，也不会自动信任远程 `publisher`。远程索引不等于远程安装。

## index.json 结构

远程索引是一个静态 JSON 文件，推荐命名为 `index.json`：

```json
{
  "schema_version": "1",
  "generated_at": "2026-01-01T00:00:00Z",
  "source": {
    "source_id": "devhub-official-index",
    "name": "DevHub Official Plugin Index",
    "homepage": "https://example.com/devhub/plugins",
    "description": "只读远程插件索引示例"
  },
  "plugins": [
    {
      "code": "demo_notice",
      "name": "示例公告插件",
      "description": "声明型公告插件示例",
      "latest_version": "1.0.0",
      "versions": [
        {
          "version": "1.0.0",
          "min_core_version": "v1.6.0",
          "compatible_core_version": ">=1.6.0 <1.7.0",
          "package_url": "https://example.com/packages/demo_notice-1.0.0.zip",
          "package_sha256": "...",
          "manifest_sha256": "...",
          "signature_url": "https://example.com/packages/demo_notice-1.0.0.devhub-signature.json",
          "publisher_id": "devhub-official",
          "public_key_id": "devhub-official-2026",
          "license": "MIT",
          "tags": ["notice", "content"]
        }
      ]
    }
  ]
}
```

最低字段：`schema_version`、`generated_at`、`source.source_id`、`source.name`、`plugins[].code`、`plugins[].name`、`plugins[].description`、`plugins[].latest_version`、`versions[].version`、`versions[].min_core_version`、`versions[].compatible_core_version`、`versions[].package_url`、`versions[].package_sha256`、`versions[].publisher_id`、`versions[].public_key_id`、`versions[].license`。

## 索引源配置与安全限制

后台可维护只读索引源：`source_id`、`name`、`index_url`、`homepage`、`description`、`status`、`trust_policy`、最近拉取状态和索引 hash。

安全规则：

- `index_url` 只支持 `http` / `https`；生产建议 `https`。
- 禁止 `file://`、空 host、本机地址、`localhost`、`127.0.0.1`、内网 IP 和 link-local 地址；测试环境可设置 `DEVHUB_ALLOW_LOCAL_REMOTE_INDEX=1`。
- GET 请求有超时和响应大小限制，当前响应上限为 2MB。
- 只接受 JSON；非 JSON Content-Type 只给 warning，但 JSON 内容必须可解析。
- 拉取只请求 `index_url`，不会请求 `package_url`、`signature_url` 或任何远程资产（远程包下载与 detached signature 下载需要显式调用 staging 下载/验签 API）。
- 拉取失败只影响该索引源状态，不影响本地插件功能。

## API

- `GET /api/v1/admin/plugins/remote-indexes`：索引源列表，需 `plugin.read`。
- `POST /api/v1/admin/plugins/remote-indexes`：新增索引源，需 `plugin.manage`。
- `PUT /api/v1/admin/plugins/remote-indexes/:id`：更新索引源，需 `plugin.manage`。
- `POST /api/v1/admin/plugins/remote-indexes/:id/enable`：启用索引源，需 `plugin.manage`。
- `POST /api/v1/admin/plugins/remote-indexes/:id/disable`：禁用索引源，需 `plugin.manage`。
- `DELETE /api/v1/admin/plugins/remote-indexes/:id`：删除索引源，需 `plugin.manage`。
- `POST /api/v1/admin/plugins/remote-indexes/:id/fetch`：拉取 `index.json`，需 `plugin.manage`。
- `GET /api/v1/admin/plugins/remote-indexes/:id/plugins`：查看远程插件列表，需 `plugin.read`。
- `GET /api/v1/admin/plugins/remote-indexes/:id/plugins/:code`：查看远程插件详情，需 `plugin.read`。

## trusted publisher 联动

远程索引中的 `publisher_id` + `public_key_id` 会和本地可信发布者记录比对：

- 本地为 `trusted`：展示 trusted。
- 本地不存在：展示 unknown，并进入风险提示。
- 本地为 `blocked` / `revoked`：展示 blocked / revoked，并标记 blocked 风险。

远程索引不能自动创建、更新或信任 publisher；远程源声明 trusted 不会被采信。

## Core 兼容性和本地安装状态

列表和详情会展示：

- 当前 Core 与 `min_core_version` / `compatible_core_version` 的兼容状态。
- 同 code 插件是否已安装、本地版本、是否有远程新版本。
- `update_available`、`local_newer`、`installed` 或 `not_installed` 状态。

本轮仅展示这些状态，不触发远程下载、升级 dry-run 或安装。

## 风险规则

- `index_url` 非 HTTPS：warning。
- unknown publisher：warning。
- blocked / revoked publisher：blocked。
- Core 不兼容：blocked。
- `package_sha256` 缺失：blocked。
- `package_url` 非法协议：blocked；HTTP URL 给 warning。
- `signature_url` 缺失：warning；HTTP URL 给 warning。`signature_url` 只是 detached signature 元数据地址，本文件拉取阶段不会请求它。
- JSON schema invalid、响应过大、拉取超时、SSRF 风险 URL：blocked / error。

`package_sha256` 只是远程元数据声明，本轮不会下载包内容，因此也不会把它当成已校验结果。

## 后台页面

后台路由：`/admin-next/plugins/remote-indexes`。

页面能力：索引源列表、新增 / 编辑 / 启用 / 禁用 / 删除、拉取、远程插件列表、远程插件详情、publisher trust、Core 兼容性、本地安装状态和风险提示。页面明确显示只读边界，不提供下载、安装、自动更新、远程市场或动态加载入口。

示例文件：`docs/examples/plugin-remote-index.example.json`。

## 与版本仓库的关系（v1.6.0-P0-05）

远程索引中的版本会出现在插件版本仓库中，来源标记为 `remote_index`。这些记录只包含远程元数据：`package_url`、`package_sha256`、publisher、Core 兼容性和风险提示。DevHub 不下载 `package_url`，不执行远程包 dry-run，不把远程版本当成本地可升级包。

如需对远程版本执行升级差异对比，必须先通过后续受控下载 / 当前 zip 上传沙箱 / promote 流程把插件包纳入 `storage/plugins/packages/`。直接对 `remote_index` 调用 upgrade-diff 会返回 `plugin_version_remote_readonly`。

## 后台入口分组（v1.6.0-P1-09）

远程插件索引页面归入后台“系统插件 / 远程与开发者”分组。页面仍保持只读定位：只展示远程 index.json 元数据、publisher trust、Core 兼容性和本地安装状态，不提供下载、安装、自动更新、动态加载或远程市场入口。

## v1.7.0-P0-01：远程包下载到 staging

v1.7.0-P0-01 开始补齐远程安装链路的第一步：服务端可以把远程索引中的 `package_url` 或管理员手工指定的包 URL 安全下载到 `storage/plugins/staging/downloads/`。

边界仍然严格：

- 只下载到 staging，不安装、不启用、不解压执行、不运行包内脚本、不加载 Go plugin、不执行 SQL、不动态加载前端资产。
- 下载前校验 URL：仅 HTTPS，拒绝 localhost、回环地址、内网地址、link-local 地址、非法协议和重定向到禁止地址。
- 下载过程限制重定向次数、连接 / 读取超时和默认 20MB 文件大小。
- 下载后计算 sha256；远程索引提供 `package_sha256` 时必须匹配。没有 sha256 的下载记录标记为 `checksum_missing`，仅保留 staging 文件，不能进入自动安装。
- 当前只允许 `.zip`、`.tar.gz`、`.tgz`。未知格式和可执行 / 脚本类包不会进入 staging。
- `package_sha256` 在远程索引中仍只是元数据声明；只有经过 `/api/v1/admin/plugins/packages/download` 下载并比对后，才成为本地记录中的实际校验结果。

API 见 `docs/API.md`：`POST /api/v1/admin/plugins/packages/download`、`GET /api/v1/admin/plugins/packages/staging`、`GET /api/v1/admin/plugins/packages/staging/:id`、`DELETE /api/v1/admin/plugins/packages/staging/:id`。

## v1.7.0-P0-03 与兼容性检查的关系

远程索引版本仍然只是只读元数据，不能直接进入 compat-check。远程包必须先通过安全下载到 staging，再完成解压安全检查与 manifest 预校验，生成 `plugin_package_prechecks.status=passed` 记录后，才能调用 `POST /api/v1/admin/plugins/packages/prechecks/:id/compat-check`。

compat-check 会使用预检 manifest 检查 Core 版本、依赖、plugin_code、content_type、权限、菜单、路由、Hook、config_schema 和 migration 兼容性，并返回 `can_install`。本轮仍不下载依赖插件、不安装远程包、不启用插件、不执行 migration、不执行第三方代码。

## v1.7.0-P0-05 与启用前检查（enable-precheck）的关系

启用前检查的输入来自“已安装但未启用”的插件，并且强制要求存在最近的 `precheck passed + compat-check can_install=true` 链路，避免绕过远程包治理链路直接启用。enable-precheck 只做复检与结论输出，不会真正启用插件或注册运行时能力。
