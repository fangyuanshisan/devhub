# Demo Notice

这是 DevHub “本地插件包目录规范 + dry-run 导入预览”示例插件包。

本目录用于演示：

- 插件包目录结构（manifest / README / docs / migrations / assets 等）。
- 文件扫描与危险文件阻断（dry-run 只做读取与校验，不执行任何文件）。
- 复用现有 manifest validate / dry-run 逻辑生成预览信息。

安全边界（当前版本）：

- 不安装插件。
- 不执行插件代码。
- 不执行外部 SQL。
- 不动态加载前端资产。
- 不做 zip 上传与远程市场。

可在后台 `系统插件 -> 安装升级` 页面使用“本地插件包 dry-run”输入：

`examples/plugins/demo_notice_install`

