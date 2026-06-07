# official_links

`official_links` 是 DevHub 官方友情链接插件。它从 v1.8.4-S1 起作为生产可用官方插件包维护，用于验证和承载声明型 content_type、权限、菜单、配置、子站启用、禁用阻断和历史内容可读能力。

## 包结构

- `manifest.json`：声明插件能力。
- `config.example.json`：示例配置，不包含真实 secret。
- `checksums.json`：包文件完整性摘要。
- `migrations/001_init.sql`：唯一迁移入口，当前为 no-op 计划文件，说明插件复用 Core 内容表。
- `migrations/README.md`：说明当前插件复用 Core 内容表，不需要插件私有业务表。

包内不包含运行时代码、package scripts、可执行二进制、远程 iframe、远程 JS、inline HTML 或 remote component。

## 能力声明

- content_type：`friend_link`
- 后台菜单：`友情链接管理`，路径 `/admin-next/official-links`
- 前台入口：`/search/?content_type=friend_link`
- 权限：
  - `official_links.menu.view`
  - `official_links.link.create`
  - `official_links.link.manage`
  - `official_links.config.manage`

## 数据模型

本插件不创建私有业务表，友情链接作为 Core 内容保存：

| 字段 | Core 存储 | 说明 |
| --- | --- | --- |
| id | `topics.id` | 友情链接 ID |
| community_id | `topics.community_id` | 子站 |
| title | `topics.title` | 链接标题 |
| url | `topics.content` | 链接地址 |
| description | `topics.summary` | 描述 |
| logo_url | 暂不单独存储 | 后续可由附件 / 扩展字段补充 |
| sort_order | `pinned/recommended/created_at` | 当前用置顶 / 推荐和时间排序 |
| status | `topics.status` | publish / hidden / offline / pending |
| created_by / updated_by | Core 审计与内容记录 | 操作者通过审计追踪 |
| created_at / updated_at | `topics.created_at/updated_at` | 时间 |

## 使用流程

1. 在插件包治理中上传包。
2. 执行 precheck。
3. promote 到本地仓库。
4. 执行 install dry-run。
5. install。
6. 启用插件，安装成功后 PluginRegistry 会刷新运行态快照。
7. 在目标子站启用插件，并给子站板块允许 `friend_link`。
8. 在后台 `/admin-next/official-links` 管理友情链接。
9. 前台通过 `/search/?content_type=friend_link` 或子站插件菜单查看。

## 安全边界

- DevHub 不执行插件包代码。
- `migrations/` 是唯一迁移入口；当前 `001_init.sql` 为 no-op 计划文件，dry-run 记录但不执行。
- dry-run 不执行 SQL。
- install 不执行 package scripts。
- disabled / archived / soft_uninstalled 后阻断新建和治理能力，历史内容仍可读。
- 不开放远程 iframe、远程 JS、remote component 或 blocking Hook。
