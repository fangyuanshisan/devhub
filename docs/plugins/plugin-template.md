# 插件目录模板

本模板用于规范 DevHub 后续插件开发目录。当前阶段系统不会扫描或加载 `plugins/example/` 目录，内置系统插件仍通过 Go 代码中的 registry 注册。

```text
plugins/example/
  manifest.json
  README.md
  migrations/
    001_example_items.sql
  hooks/
    README.md
  config/
    defaults.json
  assets/
    icon.svg
  admin/
    README.md
  frontend/
    README.md
  docs/
    usage.md
```

## 文件说明

- `manifest.json`：插件声明入口，对齐 [manifest.example.json](manifest.example.json)。
- `README.md`：插件能力、安装约束、权限、迁移和回滚风险说明。
- `migrations/`：插件迁移声明和 SQL 草案。当前 DevHub 只对内置插件做迁移记录，外部插件迁移执行为后续能力。
- `hooks/`：Hook 处理器说明。当前不加载第三方代码，Hook 运行时仅服务内置插件。
- `config/defaults.json`：插件默认配置。最终合并规则为默认配置 < 全局配置 < 子站配置。
- `assets/`：图标、截图等静态说明资源。当前不会自动发布到前台。
- `admin/`、`frontend/`：后续插件 UI 资源预留目录。当前不支持动态前端加载。
- `docs/`：插件使用和运维文档。

## 当前边界

- 已支持：内置插件 manifest/registry、状态、配置、权限、菜单、Hook、迁移记录和审计治理。
- 未支持：外部插件包上传、远程安装、在线更新、Go 动态插件、第三方脚本沙箱和动态前端加载。
- 推荐路线：先用该模板沉淀 manifest + 配置型插件规范，再评估外部服务型插件。
