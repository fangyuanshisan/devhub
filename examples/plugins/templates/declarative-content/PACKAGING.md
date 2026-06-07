# Packaging

最小打包流程：

```bash
cd examples/plugins/templates/declarative-content
zip -qr ../../official-links-template.zip .
```

验收流程：

1. 上传 zip。
2. 查看预检结果。
3. 修复阻断项；warning 需要管理员确认。
4. promote 到本地仓库。
5. 执行 install dry-run。
6. 确认安装。
7. 启用插件。
8. 在插件详情中验证配置、content_type、权限和菜单。

注意：

- 不要把真实 token、用户数据、运行时代码或外部 SQL 放进包里。
- `migrations/` 是唯一迁移入口。
- 根目录 SQL 不会执行。
- dry-run 不执行 SQL、不写插件状态、不调用外部服务。
