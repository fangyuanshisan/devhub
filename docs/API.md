# DevHub API 文档

[返回文档入口](README.md)

更新时间：2026-05-14（v1.5.0 插件包治理收口验收；历史 v1.4.0 能力保持可追溯）

本文档只记录当前仓库真实可用 API。接口路径以 `internal/transport/httpapi/router.go` 为准；未实现能力集中放在“规划 / 未完成”小节，不写入当前真实 API 主体。

## 通用规则

- API 前缀：`/api/v1`。
- 认证方式：`Authorization: Bearer <access_token>`。
- 前台用户 token：`token_type=user`，用于发帖、评论、关注、举报、用户中心和 `/api/v1/moderator/*`。
- 后台管理员 token：`token_type=admin`，用于 `/api/v1/admin/*`。
- 错误响应：兼容旧结构 `{"error":"错误信息"}`，插件治理相关接口新增结构化错误字段（不会移除 `error` 字段）：

```json
{
  "error": "插件依赖未满足，无法启用",
  "code": "plugin_dependency_disabled",
  "message": "插件依赖未满足，无法启用",
  "details": {
    "plugin_code": "docs",
    "dependency_code": "qa",
    "current_status": "disabled",
    "required_version": ">=1.0.0"
  },
  "suggestion": "请先启用 qa 插件后重试"
}
```

### 插件治理统一错误码（v1.4.0-P1-10）

说明：

- 插件治理接口在返回 `{"error": "..."} ` 的同时，会尽量补充 `code/message/details/suggestion`，供后台统一 UI 展示与可操作诊断。
- `message` 面向用户可读；`details` 只放诊断信息（不得包含 token/secret/password 等敏感值）；`suggestion` 提供可执行的修复建议。

错误码清单：

| code | 含义（示例 message） | 常见 details 字段 | 典型修复建议 |
| --- | --- | --- | --- |
| `plugin_not_found` | 插件不存在 | `plugin_code` | 检查插件编码，或先执行 manifest 安装 |
| `plugin_not_installed` | 插件尚未安装 | `plugin_code` | 先在安装向导中安装插件 |
| `plugin_archived` | 插件已归档 | `plugin_code` | 先恢复插件后重试 |
| `plugin_disabled` | 插件未启用 | `plugin_code` | 先启用插件后重试 |
| `plugin_config_invalid` | 插件配置无效 | `plugin_code`,`reason` | 修复配置后重试 |
| `plugin_migration_failed` | 插件迁移失败 | `plugin_code`,`migration_name` | 到“迁移”Tab 重试或处理失败原因 |
| `plugin_dependency_missing` | 依赖缺失 | `plugin_code`,`dependency_code`,`required`,`required_version`,`current_status`,`dependency_chain` | 安装并启用依赖插件 |
| `plugin_dependency_disabled` | 依赖未启用 | 同上 | 先启用依赖插件 |
| `plugin_dependency_archived` | 依赖已归档 | 同上 | 先恢复依赖插件 |
| `plugin_dependency_version_mismatch` | 依赖版本不满足 | 同上 | 升级/降级依赖插件到兼容版本 |
| `plugin_dependency_cycle` | 循环依赖/自依赖 | `plugin_code`,`dependency_chain` | 调整依赖关系避免环 |
| `plugin_core_version_incompatible` | Core 版本不兼容 | `plugin_code`,`core_version`,`min_core_version`,`compatible_core_version`,`messages` | 升级 Core 或选择兼容版本插件 |
| `plugin_permission_denied` | 权限不足（插件治理） | `permission_code` | 为当前账号/角色补齐权限后重试 |
| `plugin_config_permission_denied` | 缺少插件配置权限 | `permission_code` | 为当前账号/角色补齐配置相关权限 |
| `plugin_content_permission_denied` | 缺少内容治理权限 | `permission_code` | 为当前账号/角色补齐治理权限 |
| `plugin_hook_blocked` | blocking Hook 阻断 | `plugin_code`,`hook_name`,`blocking` | 查看 Hooks Tab 最近失败记录 |
| `plugin_hook_failed` | non-blocking Hook 失败 | 同上 | 查看 Hooks Tab 最近失败记录 |
| `plugin_config_schema_invalid` | config_schema 校验失败 | `plugin_code`,`path`,`reason` | 按字段路径修复配置后重试 |
| `plugin_manifest_invalid` | manifest 校验失败 | `errors` | 按 errors 修复 manifest 后重试 |
| `plugin_package_path_invalid` | 插件包路径不合法 | `path`,`allowed_roots` | 使用允许目录内的相对路径 |
| `plugin_package_not_found` | 插件包目录不存在 | `path` | 检查路径或先创建插件包目录 |
| `plugin_package_template_invalid` | 插件包模板初始化参数无效 | `reason` | 修复 code/content_type/name 等字段后重试 |
| `plugin_package_template_exists` | 初始化目标目录已存在 | `path` | 换用新 code；当前后台初始化不暴露 force 覆盖 |
| `plugin_package_upload_invalid_type` | 上传包类型不允许 | `filename` | 仅上传 `.zip`；不支持 tar/gz/rar/7z |
| `plugin_package_upload_too_large` | 上传 zip 超过大小限制 | `max_bytes` | 将 zip 控制在 20MB 以内 |
| `plugin_package_zip_invalid` | zip 文件无法解析 | - | 重新打包为合法 zip |
| `plugin_package_zip_entry_path_invalid` | zip entry 路径不合法 | `entry` | 使用普通包内相对路径 |
| `plugin_package_zip_slip_detected` | 检测到 zip slip / 路径穿越 | `entry` | 移除 `../`、绝对路径或 Windows 盘符路径 |
| `plugin_package_zip_bomb_detected` | 检测到 zip bomb 风险 | `entry` | 移除重复 entry 或异常压缩结构 |
| `plugin_package_zip_too_many_files` | zip 解压文件数量超过限制 | `max_files` | 文件数控制在 300 个以内 |
| `plugin_package_zip_file_too_large` | zip 内单文件超过限制 | `entry`,`max_bytes` | 单文件控制在 5MB 以内 |
| `plugin_package_zip_total_size_exceeded` | zip 解压后总大小超过限制 | `max_bytes` | 解压总大小控制在 50MB 以内 |
| `plugin_package_zip_nested_archive_forbidden` | zip 内嵌压缩包被禁止 | `entry` | 移除内嵌 zip/tar/gz/rar/7z 等压缩包 |
| `plugin_package_zip_symlink_forbidden` | zip 包含 symlink | `entry` | 移除软链接后重新打包 |
| `plugin_package_zip_multiple_manifests` | zip 中发现多个 manifest.json | `manifests` | 本轮每次只上传一个插件包 |
| `plugin_package_zip_manifest_missing` | zip 中找不到可用 manifest.json | `manifest` | 在 zip 根目录或单一顶层目录放置 manifest.json |
| `plugin_package_upload_extract_failed` | zip 解压失败 | `entry` | 重新打包后上传 |
| `plugin_package_upload_blocked` | 上传包扫描后被阻断 | `upload_id` | 根据风险报告修复后重新上传 |
| `plugin_package_upload_not_found` | 上传记录不存在 | `upload_id` | 刷新上传记录或重新上传 |
| `plugin_package_upload_invalid_status` | 上传包当前状态不允许该动作 | `upload_id`,`status` | 按可执行动作列表处理 |
| `plugin_package_upload_lifecycle_invalid` | 生命周期流转不合法 | `upload_id`,`status` | 刷新详情后按当前状态重新操作 |
| `plugin_package_upload_rescan_failed` | 上传包重新扫描失败 | `upload_id` | 检查 staging 文件是否仍存在 |
| `plugin_package_upload_cancel_failed` | 上传包取消失败 | `upload_id` | 刷新详情后重试 |
| `plugin_package_upload_delete_failed` | 上传包删除失败 | `upload_id` | 检查文件权限后重试 |
| `plugin_package_upload_cleanup_failed` | 上传包清理失败 | - | 检查 storage 权限与磁盘状态 |
| `plugin_package_upload_expired` | 上传包已过期 | `upload_id`,`expires_at` | 重新上传插件包 |
| `plugin_package_upload_already_deleted` | 上传包已删除 | `upload_id` | 查看审计记录或重新上传 |
| `plugin_package_upload_approval_required` | 上传包导入需要审批 | `upload_id`,`risk_level` | 先提交导入审批 |
| `plugin_package_upload_approval_blocked` | 当前上传包不能提交导入审批 | `upload_id`,`status` | 修复阻断项或重新上传 |
| `plugin_package_upload_promote_requires_approval` | promote 前需要导入审批 | `upload_id` | 审批通过后再 promote |
| `plugin_package_upload_promote_failed` | 上传包 promote 失败 | `upload_id` | 查看详情中的错误码和建议 |
| `plugin_package_upload_promote_target_exists` | 上传包 promote 目标已存在 | `path` | 更换 code 或清理目标目录 |
| `plugin_package_upload_action_not_allowed` | 当前动作不可用 | `action`,`status` | 查看详情 actions 中的 reason |
| `plugin_package_upload_status_conflict` | 状态冲突 | `status` | 先完成或取消正在进行的审批 |
| `plugin_package_promote_blocked` | 上传包禁止转入本地仓库 | `upload_id`,`status` | 修复 blocked 风险、checksum 或 manifest 错误 |
| `plugin_package_promote_target_exists` | 本地仓库目标目录已存在 | `path` | 删除/更名已有目录后重试；默认不覆盖 |
| `plugin_package_promote_failed` | 转入本地仓库失败 | `path` | 检查目录权限和磁盘空间 |
| `plugin_package_manifest_missing` | 缺少 manifest.json | `path` | 在插件包根目录补充 manifest.json |
| `plugin_package_manifest_invalid` | manifest.json 非法 | `path`,`reason` | 修复 manifest 后重试 |
| `plugin_package_dangerous_file` | 检测到危险文件 | `path` | 移除 `.sh/.sql/.js/.ts` 等危险文件 |
| `plugin_package_file_too_large` | 单文件超过大小限制 | `path`,`size` | 缩小单文件大小 |
| `plugin_package_too_large` | 插件包总大小超过限制 | `path`,`total_size` | 缩小包体积 |
| `plugin_package_too_many_files` | 文件数量超过限制 | `path`,`total_files` | 减少文件数量 |
| `plugin_package_unknown_files` | 发现未知文件 | `path`,`unknown_files` | 检查未知文件是否应加入 allow 列表 |
| `plugin_package_dry_run_blocked` | 本地插件包 dry-run 被阻断 | `path` | 根据 blocking 原因修复后重试 |
| `plugin_package_checksum_missing` | checksums.json 缺失（warning） | `path` | 建议补充 checksums.json（sha256） |
| `plugin_package_checksum_invalid` | checksums.json 非法 | `path`,`reason` | 修复 checksums.json 后重试 |
| `plugin_package_checksum_unsupported_algorithm` | checksum algorithm 不支持 | `algorithm` | 当前仅支持 sha256 |
| `plugin_package_checksum_duplicate_path` | checksums.json path 重复 | `path` | 移除重复项后重试 |
| `plugin_package_checksum_file_missing` | checksums.json 声明文件不存在 | `path` | 补齐文件或修复 files 列表 |
| `plugin_package_checksum_mismatch` | checksum 不匹配 | `mismatched` | 重新生成 checksums.json 或修复文件内容 |
| `plugin_package_file_not_covered` | 存在未被 checksum 覆盖文件（warning） | `extra` | 建议补齐 checksums 覆盖范围 |
| `plugin_package_symlink_forbidden` | 插件包包含软链接（禁止） | `path` | 移除软链接文件 |
| `plugin_package_size_limit_exceeded` | 文件大小超限（禁止） | `path` | 缩小文件体积或移除大文件 |
| `plugin_package_file_count_exceeded` | 文件数量超限（禁止） | `total_files` | 减少文件数量 |
| `plugin_package_risk_blocked` | 风险评估阻断 dry-run | `items` | 根据风险项修复后重试 |
| `plugin_package_signature_invalid` | signature.json 非法或签名字段不合法 | `path`,`reason` | 修复 signature.json 后重试 |
| `plugin_package_signature_unsupported_algorithm` | 不支持的签名算法 | `algorithm` | 当前仅支持 ed25519 |
| `plugin_package_signature_verification_failed` | 签名验签失败 | `publisher_id`,`public_key_id` | 检查 checksums.json/签名是否匹配，必要时重新生成签名 |
| `plugin_package_signature_payload_invalid` | signature payload / payload_algorithm 不支持 | `payload`,`payload_algorithm` | 当前仅支持 `payload=checksums.json` 与 `payload_algorithm=sha256` |
| `plugin_package_signature_signed_file_missing` | signature.json 声明的签名文件不存在 | `path` | 补齐文件或修复 signed_files 列表 |
| `plugin_package_signature_path_invalid` | signature.json signed_files 路径不合法 | `path` | 使用包内相对路径且禁止 `..` |
| `plugin_package_signature_manifest_unsigned` | signature.json 未签名 manifest/checksums | `missing` | 在 signed_files 中补齐 manifest.json 与 checksums.json |
| `plugin_package_signature_publisher_unknown` | 签名发布者未建立本地可信关系 | `publisher_id`,`public_key_id` | 在可信发布者页面维护公钥或仅在测试环境使用 |
| `plugin_package_signature_publisher_blocked` | 签名发布者被本地策略 blocked | `publisher_id`,`public_key_id` | 更换发布者或恢复可信发布者状态 |
| `plugin_package_signature_publisher_revoked` | 签名发布者被本地策略 revoked | `publisher_id`,`public_key_id` | 更换发布者签名或移除插件包 |
| `plugin_package_publisher_invalid` | publisher.json 非法 | `path`,`reason` | 修复 publisher.json 后重试 |
| `plugin_package_trusted_publishers_unavailable` | trusted_publishers 配置不可用（warning） | `path`,`reason` | 检查 `storage/plugins/trusted_publishers.json` 访问权限与 JSON 格式 |
| `plugin_trusted_publisher_not_found` | 可信发布者不存在 | `id` | 刷新列表后重试 |
| `plugin_trusted_publisher_duplicate` | publisher_id + public_key_id 重复 | `publisher_id`,`public_key_id` | 使用新的 key id 或编辑已有记录 |
| `plugin_trusted_publisher_invalid_key` | 可信发布者公钥不合法 | `public_key_algorithm`,`public_key` | 当前仅支持 32 字节 Ed25519 公钥 base64 |
| `plugin_trusted_publisher_invalid_status` | 可信发布者状态不合法 | `status` | 使用 trusted / blocked / revoked |
| `plugin_trusted_publisher_permission_denied` | 缺少可信发布者管理权限 | `permission_code` | 需要 `plugin.manage` |
| `plugin_package_repository_not_found` | 插件包仓库目录不存在 | `root` | 创建仓库目录或检查路径 |
| `plugin_package_repository_forbidden` | 插件包仓库路径不允许 | `root`,`allowed_roots` | 使用白名单目录下的仓库路径 |
| `plugin_package_scan_failed` | 插件包仓库扫描失败 | `root`,`reason` | 检查目录权限或文件状态 |
| `plugin_package_invalid` | 插件包无效 | `path` | 补齐 manifest/checksum 后重试 |
| `plugin_package_detail_not_found` | 插件包详情不存在 | `path` | 检查路径或先扫描仓库 |
| `plugin_package_install_blocked` | 插件包安装被阻断 | `path`,`risk_level`,`blocked_code` | 修复阻断原因后重试 |
| `plugin_package_install_failed` | 插件包安装失败 | `plugin_code` | 查看后台日志后重试 |
| `plugin_package_already_installed` | 同编码插件已安装 | `plugin_code` | 走 upgrade 流程升级插件 |
| `plugin_package_dependency_missing` | required 依赖未满足 | `dependencies` | 先安装并启用 required 依赖插件 |
| `plugin_package_core_incompatible` | Core 版本不兼容 | `core_version`,`min_core_version`,`compatible_core_version` | 升级 Core 或选择兼容版本插件包 |
| `plugin_config_version_not_found` | 配置版本不存在 | `version_id` | 刷新版本列表后重试 |
| `plugin_config_version_invalid_scope` | 配置版本 scope 不匹配 | `scope`,`actual_scope` | 从正确 scope 打开版本 |
| `plugin_config_version_schema_invalid` | 目标版本配置未通过 schema（用于 rollback dry-run 的 `blocked_code`） | `schema_validation.errors` | 修复配置或选择其他版本 |
| `plugin_config_encryption_key_missing` | 缺少插件配置加密密钥（敏感字段无法保存） | - | 配置 `DEVHUB_PLUGIN_CONFIG_KEYS` 或 `DEVHUB_PLUGIN_CONFIG_KEY_ID/DEVHUB_PLUGIN_CONFIG_KEY` 后重试 |
| `plugin_config_encryption_key_invalid` | 插件配置加密密钥不合法 | - | 检查密钥长度/格式（推荐 base64 32 bytes）与 key_id 配置 |
| `plugin_config_encrypt_failed` | 敏感字段加密失败 | `plugin_code` | 检查密钥配置后重试 |
| `plugin_config_decrypt_failed` | 敏感字段解密失败 | `plugin_code` | 检查密钥是否变更；避免轮换导致旧密文不可读 |
| `plugin_config_key_missing` | 插件配置密钥缺失 | - | 配置 `DEVHUB_PLUGIN_CONFIG_KEYS` 或 split 环境变量后重试 |
| `plugin_config_key_invalid` | 插件配置密钥配置不合法 | - | 检查 JSON/split 配置格式与 key 长度 |
| `plugin_config_key_current_missing` | 缺少 current key | - | 确保 current key_id 存在且 keys 中包含该 key_id |
| `plugin_config_key_not_found` | 指定 key_id 不存在 | `key_id` | 补齐 old keys 或修复密文 key_id |
| `plugin_config_cipher_invalid` | 密文格式不合法或发现敏感明文 | `field_path` | 修复密文格式或重新写入敏感字段 |
| `plugin_config_cipher_version_unsupported` | 不支持的密文版本 | `cipher_version` | 升级到支持该版本的 Core 或重新加密 |
| `plugin_config_cipher_key_missing` | 密文缺少 key_id 或缺少解密 key | `key_id` | 补齐旧 key 后重试 |
| `plugin_config_cipher_decrypt_failed` | 密文解密失败 | `key_id` | 检查 key 是否正确；避免误删旧 key |
| `plugin_config_rotation_dry_run_blocked` | 密钥轮换 dry-run 被阻断 | `decrypt_failed`,`missing_key` | 先修复阻断项（补齐 old keys / 修复密文）后重试 |
| `plugin_config_rotation_reencrypt_failed` | 密钥轮换 re-encrypt 失败 | `plugin_code` | 查看日志并修复后重试 |
| `plugin_config_rotation_confirm_key_mismatch` | confirm_current_key_id 与当前 key 不匹配 | `current_key_id` | 刷新页面后重新确认 |
| `plugin_config_rotation_history_unsupported` | 暂不支持配置历史轮换 | - | 本版本默认只轮换当前配置；历史轮换后续补齐 |
| `plugin_config_rotation_permission_denied` | 缺少密钥轮换权限 | `permission_code` | 需要 `plugin.manage` |
| `plugin_approval_not_found` | 审批记录不存在 | `id` | 刷新审批列表后重试 |
| `plugin_approval_invalid_action` | 不支持的审批动作 | `action` | action 仅支持 install / upgrade |
| `plugin_approval_invalid_status` | 状态不允许该操作 | `status` | 检查状态流转后重试 |
| `plugin_approval_create_blocked` | 创建审批被阻断 | `blocked_code`,`blocked_reasons` | 先修复 dry-run 阻断项再提交审批 |
| `plugin_approval_permission_denied` | 无权限执行该审批动作 | `requested_by` | 使用申请人或具备更高权限账号处理 |
| `plugin_approval_reject_reason_required` | 拒绝原因必填 | - | 提交 comment 后重试 |
| `plugin_approval_execute_blocked` | 执行前重新校验阻断 | `blocked_code`,`blocked_reasons` | 修复后重新提交审批 |
| `plugin_approval_execute_failed` | 审批执行失败 | `error_code` | 查看后台日志，修复后重试 |
| `plugin_approval_already_executed` | 审批已执行 | `id` | 无需重复执行 |
| `plugin_approval_already_canceled` | 审批已撤销 | `id` | 重新提交审批 |
| `plugin_export_not_supported` | 插件不支持导出 | `plugin_code`,`reason` | 仅导出声明型插件；运行时代码/用户数据不在导出范围 |
| `plugin_export_dry_run_failed` | 导出 dry-run 失败 | `plugin_code`,`reason` | 根据预览错误修复后重试 |
| `plugin_export_failed` | 导出写入失败 | `plugin_code`,`output_dir`,`reason` | 检查导出目录权限和磁盘状态 |
| `plugin_export_path_invalid` | 导出目录不合法 | `output_dir` | 使用 `storage/plugins/exports/` 下的相对目录 |
| `plugin_export_output_exists` | 导出目录已存在 | `output_dir` | 换一个目录或使用 `force=true` |
| `plugin_export_manifest_invalid` | 导出 manifest 校验失败 | `plugin_code`,`errors` | 修复插件 manifest 声明后重试 |
| `plugin_export_sensitive_value_detected` | 检测到敏感配置泄露风险 | `plugin_code`,`paths` | 仅导出脱敏示例配置，不导出真实配置 |
| `plugin_export_user_data_forbidden` | 检测到用户数据导出风险 | `plugin_code` | 移除用户内容/通知/审计等业务数据 |
| `plugin_export_checksum_failed` | checksums.json 生成失败 | `output_dir`,`reason` | 检查文件写入结果后重试 |
| `plugin_export_package_dry_run_failed` | 导出后包自检失败 | `output_dir`,`status` | 根据 package dry-run warning/error 修复导出内容 |
| `plugin_export_permission_denied` | 缺少插件导出权限 | `permission_code` | 查看预览需 `plugin.read`，正式导出需 `plugin.write` |
- 权限错误统一返回 `403`，典型格式为 `{"error":"无权限"}` 或 `{"error":"缺少权限 <permission_code>，不能创建该类型内容"}`；插件内容创建必须以 `ContentTypeDefinition.create_permission` 为准，`post.create` 只作为 `core.topic.create` 的历史兼容桥。
- 分页参数：`page`、`page_size`，默认按接口实现处理，建议 `page_size <= 50`。

## 插件 API

说明：后台全局插件管理页、子站插件配置抽屉和版主插件菜单均继续使用本节现有接口。2026-05-11 插件系统专项验收未新增 API，也未改变返回字段语义；验收重点是确认现有插件启停、impact、config、Hook、audit、migration 和通用插件内容治理接口可支撑后台治理中心与 E2E 回归。2026-05-11 MySQLStore / 老库升级专项同样未新增生产 API，验证的是 MySQLStore 下这些接口背后的状态、迁移、Hook、审计与配置校验链路和 MemoryStore 口径一致。

### 前台导航与发布入口解析（v1.4.0-P1-11）

说明：

- 这组接口用于前台统一“导航入口 / 发布入口”的可见性治理，尽量由后端复用插件状态、子站插件状态、依赖、配置、迁移与权限判断，避免前端重复拼接规则。
- 这些接口只影响“入口是否显示/是否可点击/原因提示”，不绕过后端真实写操作校验；用户直接访问创建 API 时仍会被后端强校验拦截。

`GET /api/v1/navigation`

- 认证：可选（guest 可访问；登录后会按权限/登录态过滤可见性）。
- 用途：总站级导航入口（不包含子站板块绑定判断）。
- 返回：`items` 列表，字段包括 `location/title/route/plugin_code/content_type/required_permission/visible/reason/reason_code/details`。

`GET /api/v1/communities/:slug/navigation`

- 认证：可选（登录态影响 `require_login/permission` 相关入口）。
- 用途：子站级导航入口（包含子站插件启用、依赖/配置/迁移、板块绑定等判断）。
- 返回：同上，额外返回 `community_slug`。

`GET /api/v1/communities/:slug/create-options`

- 认证：可选（guest 仍可访问，但会返回 `visible=false` + `plugin_login_required` 的原因提示）。
- 用途：前台“发布内容 / 创建内容类型”候选列表；前台发布页应只展示 `visible=true` 的内容类型，并在不可见时展示可操作原因。
- 返回：`items` 列表，字段包括 `content_type/plugin_code/title/route/required_permission/visible/reason/reason_code/details`。

`GET /api/v1/admin/plugins/:code/menus/preview`

- 认证：后台 admin token（不允许 user token / moderator token）。
- 权限：`plugin.read`。
- 用途：后台插件详情页“前台入口预览”，用于诊断某插件声明的前台菜单在指定子站/分类上下文下为什么不可见。
- 查询参数：
  - `community_slug`：可选
  - `category_id`：可选
- 返回：`items` 列表，字段包括 `location/title/route/content_type/required_permission/visible/reason/reason_code/details`。

### 全局插件 API

`GET /api/v1/plugins`

- 认证：不需要。
- 返回：只返回全局 `enabled` 插件。
- 用途：前台判断系统可用插件能力。
- 说明：当前返回的是内置系统插件统一 manifest 视图，包括内容类型、权限、菜单、路由声明与 `config_schema` 预留字段。
- 安全处理：公共接口不返回 `config_json` 和 `resolved_config`。

响应示例：

```json
{
  "items": [
    {
      "code": "qa",
      "plugin_code": "qa",
      "name": "问答",
      "version": "1.0.0",
      "status": "enabled",
      "content_types": ["question"]
    }
  ]
}
```

`GET /api/v1/admin/plugins`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 返回：全部注册插件，包括插件状态、`config_schema`、`config_json`、`resolved_config` 和轻量 `health` 摘要。
- 状态口径：`plugins.status` 当前接受 `discovered`、`installed`、`migrated`、`configured`、`enabled`、`disabled`、`running`、`archived`、`config_invalid`、`migration_pending`、`migration_failed`、`dependency_missing`；但发布可用性仍只以 `enabled` 为通过状态，其余状态均不会放行新建内容。
- 生命周期字段：返回对象会派生 `install_status`、`lifecycle_status`、`status_reason`、`installed_at`、`archived_at`、`last_health_check_at`，用于后台展示插件安装生命周期。当前这些字段由 `plugins.status`、时间戳、迁移和健康状态派生；manifest + 配置型安装已可写入插件记录，但插件包 zip、远程安装和动态加载仍未实现。
- `health` 字段：后台治理摘要，由全局状态、配置校验、迁移记录、依赖状态和 Hook 失败统计计算；不是 Prometheus / Grafana 级监控。

`health` 示例：

```json
{
  "status": "hook_warning",
  "status_reason": "存在 Hook 失败记录",
  "config_status": "valid",
  "migration_status": "ok",
  "hook_status": "hook_warning",
  "dependency_status": "ok",
  "recent_error": "qa 插件仅允许创建 question",
  "suggested_action": "查看 Hooks Tab 的最近失败记录",
  "pending_migrations_count": 0,
  "failed_migrations_count": 0,
  "hook_failure_count": 1,
  "last_hook_error": "qa 插件仅允许创建 question"
}
```

健康状态当前支持：

- `healthy`：配置、迁移、依赖和 Hook 摘要均正常。
- `disabled`：全局插件已禁用，只影响新发布和入口展示，不影响历史内容。
- `archived`：插件已归档 / 软卸载，禁止新发布和子站启用；历史内容、配置、迁移记录、审计记录和 SEO 保留。
- `config_invalid`：插件配置未通过 `config_schema` 校验或被显式标记为配置无效。
- `migration_pending`：存在待处理迁移记录；当前内置 no-op pending 不阻断启用，但会提示治理风险。
- `error`：存在 failed migration 等阻断性异常。
- `dependency_missing`：依赖插件缺失或未启用。
- `hook_warning`：存在 Hook 失败记录，但尚未达到错误阈值。
- `hook_error`：Hook 失败次数达到当前轻量阈值（当前为失败次数 `>= 3`）。

说明：`health.status_reason` 会返回当前主要状态原因，供后台运行状态 Tab 展示。该健康摘要是轻量治理提示，不是完整监控系统。

`GET /api/v1/admin/plugins/health`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：插件治理中心健康总览。
- 返回：`items` 为全部插件的 `PluginHealth` 摘要，`summary` 为按状态聚合的计数。
- 计数口径：当前包含 `healthy`、`warning`、`error`、`disabled`、`migration_pending`、`config_invalid`、`dependency_missing`、`hook_warning`、`hook_error`、`archived`。这是轻量治理摘要，不是监控告警系统。

响应示例：

```json
{
  "items": [
    {
      "plugin_code": "qa",
      "status": "healthy",
      "status_reason": "无需处理",
      "config_status": "valid",
      "migration_status": "ok",
      "hook_status": "ok",
      "dependency_status": "ok"
    }
  ],
  "summary": {
    "healthy": 6,
    "archived": 0,
    "hook_error": 0
  }
}
```

`GET /api/v1/admin/plugins/:code/health`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：查询单个插件的健康摘要。
- 返回：单个 `PluginHealth` 对象，字段同 `GET /api/v1/admin/plugins` 中的 `health`。

`GET /api/v1/admin/plugins/:code/readiness`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：Readiness Check / 操作阻断原因诊断接口。用于后台在“为什么不能启用/升级/配置保存”等场景下提供可操作提示；该接口只做诊断，不替代真实操作校验。
- 查询参数：
  - `action`：可选，默认 `enable`。当前实现优先覆盖 `enable`（其余 action 后续补齐）。
- 返回：

```json
{
  "plugin_code": "docs",
  "action": "enable",
  "status": "blocked",
  "checks": [
    {
      "key": "dependency.qa",
      "title": "依赖插件 qa",
      "status": "blocked",
      "reason": "依赖插件 qa 当前状态为 disabled",
      "suggestion": "请先启用依赖插件后重试",
      "code": "plugin_dependency_disabled",
      "dependency_code": "qa"
    }
  ]
}
```

说明：

- `status`：`pass / warning / blocked`。
- `checks[].status`：同上；warning 表示可操作但存在风险提示（例如可选依赖缺失、待处理迁移等）。
- `checks[].code`：统一插件治理错误码，可用于后台统一 UI 提示与高亮。

`POST /api/v1/admin/plugins/:code/enable`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 路径参数：`code` 为插件 code，例如 `qa`、`docs`、`wiki`。
- 启用前检查：Service 层会校验插件存在、全局配置符合 `config_schema`、依赖插件已启用、没有 `failed` 迁移记录。
- 归档限制：`archived` 插件不能直接启用，必须先恢复为 disabled / installed 状态，再由管理员手动启用。
- 迁移策略：当前内置 migration 是 up/no-op 记录型迁移，`pending` migration 会通过健康状态和迁移 Tab 提示，但不阻断启用；`failed` migration 会阻断启用。
- 返回：更新后的插件对象。
- 审计：写入插件状态变更审计日志。

`POST /api/v1/admin/plugins/:code/disable`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 影响：禁用全局插件后，所有子站都不能继续发布该插件内容；已有内容详情不应受影响。
- 返回：更新后的插件对象。
- 审计：写入插件状态变更审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

常见错误：

- `400 {"error":"插件不存在"}`
- `400 {"error":"插件状态不合法"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

`POST /api/v1/admin/plugins/:code/archive`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：软卸载 / 归档插件。
- 行为：将全局插件状态置为 `archived`；禁止该插件新建内容、前台入口、后台管理入口和子站启用；保留历史内容、配置、迁移记录和审计记录。
- 影响分析：后台归档确认会复用 `GET /api/v1/admin/plugins/:code/impact` 展示历史内容、启用子站、绑定板块、待迁移和 Hook 异常计数。
- 审计：成功写入 `plugin.archived`；失败写入 `plugin.archive.failed`。

`POST /api/v1/admin/plugins/:code/restore`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：恢复已归档插件。
- 行为：恢复前校验配置、依赖和 failed migration；成功后状态变为 `disabled`，不会自动启用。管理员需要再次执行 `enable`。
- 常见错误：存在 failed migration 时返回 `400 {"error":"插件存在失败迁移 ... 请先重试或处理迁移错误"}`。
- 审计：成功写入 `plugin.restored`；失败写入 `plugin.restore.failed`。

`POST /api/v1/admin/plugins/bulk-archive`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：批量软卸载 / 归档插件。该接口不删除内容、配置、迁移记录或审计记录。
- 请求：

```json
{
  "plugin_codes": ["qa", "docs"]
}
```

- 返回：逐项结果，单个插件失败不会吞掉其他插件结果。

```json
{
  "succeeded": [
    {
      "plugin_code": "qa",
      "status": "archived"
    }
  ],
  "failed": [
    {
      "plugin_code": "docs",
      "error": "插件已归档"
    }
  ]
}
```

- 审计：成功项写入 `plugin.archived`，失败项写入 `plugin.archive.failed`。
- 边界：当前仅支持软卸载 / 归档，不支持硬卸载和删除数据。

`POST /api/v1/admin/plugins/bulk-restore`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：批量恢复已归档插件。
- 请求：

```json
{
  "plugin_codes": ["qa", "docs"]
}
```

- 返回：逐项结果；恢复成功后插件进入 `disabled`，不会自动 enabled。
- 审计：成功项写入 `plugin.restored`，失败项写入 `plugin.restore.failed`。

`PUT /api/v1/admin/plugins/:code/config`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 请求：

```json
{
  "config_json": {
    "example": true
  }
}
```

- 返回：更新后的插件对象，包含 `config_json` 和 `resolved_config.default/global/community/effective`。
- 校验：会按插件 `config_schema` 执行后端强校验（简化 JSON Schema），至少覆盖 `type`、`required`、`enum`、`object`、`boolean`、`string`、`number`、`integer`、`min/max`、`default` 与未知字段策略。
- 审计：写入插件全局配置审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段；`metadata_json.changed_keys` 记录本次变更的顶层配置键。
- 清空：提交 `{"config_json": null}` 或空配置会清空全局覆盖配置，并同样写入配置审计。

### 插件配置版本历史（v1.5.0-P1-05）

> 说明：版本历史与 diff 会对敏感字段脱敏；本轮提供“回滚 dry-run 预览”，不提供真实回滚写入。

`GET /api/v1/admin/plugins/:code/config/versions`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 查询参数：`page`、`page_size`。
- 返回：`items/pagination`，每项包含 `version_no/scope/community_id/changed_keys/source/operator_name/created_at/config_hash`。

`GET /api/v1/admin/plugins/:code/config/versions/:version_id`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 返回：版本详情（`config_json` 已脱敏）+ `diff`（已脱敏）。

`POST /api/v1/admin/plugins/:code/config/versions/:version_id/rollback/dry-run`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：回滚预览（dry-run）。不会写入配置、不会生成新版本、不会改变插件状态。
- 行为：读取目标版本配置，与当前配置做 diff，并使用**当前** `config_schema` 重新校验；校验失败返回 `status=blocked`，并返回 `blocked_code=plugin_config_version_schema_invalid` 与 `suggestion`。

`GET /api/v1/admin/communities/:id/plugins/:code/config/versions`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 返回：子站 scope 的版本历史（`scope=community`，`community_id=:id`）。

`GET /api/v1/admin/communities/:id/plugins/:code/config/versions/:version_id`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 返回：子站配置版本详情（`config_json/diff` 已脱敏）。

`POST /api/v1/admin/communities/:id/plugins/:code/config/versions/:version_id/rollback/dry-run`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：子站配置回滚预览（dry-run），不写入配置。

### 插件配置密钥轮换（v1.6.0-P1-07）

> 说明：本节仅提供密钥状态查看、轮换预检（dry-run）与受控 re-encrypt。不会返回密钥明文，不会返回密文，也不会返回敏感明文。默认不处理配置历史版本（`include_config_versions=false`）。

`GET /api/v1/admin/plugins/config-keys/status`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 返回：`current_key_id/loaded_key_ids/legacy_v1_supported/status/warnings`；不返回 key material。

`POST /api/v1/admin/plugins/config-keys/rotation/dry-run`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：轮换预检，不写入配置、不生成新版本、不改变插件状态。
- 请求：

```json
{
  "scope": "all",
  "plugin_code": "",
  "community_id": 0,
  "include_config_versions": false
}
```

- 返回：`status=ok|warning|blocked` + `summary/items`；`decrypt_failed/missing_key` 会导致 `blocked`。

`POST /api/v1/admin/plugins/config-keys/rotation/re-encrypt`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：将可解密的旧密文重新加密为 `enc:v2` 并使用 current key 写回（只修改敏感字段密文，不返回明文/密文）。
- 行为：服务端强制先执行 dry-run；dry-run blocked 时拒绝执行；`confirm_current_key_id` 必须匹配当前 key。
- 请求：

```json
{
  "scope": "plugin",
  "plugin_code": "wechat_connect",
  "community_id": 0,
  "include_config_versions": false,
  "confirm_current_key_id": "key-2026-01"
}
```

> 注意：re-encrypt 会改变存储密文，因此会产生新的配置版本记录（source=key_rotation），便于审计追踪。

`POST /api/v1/admin/plugins/manifest/validate`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：校验 manifest JSON 是否符合插件平台契约。
- 请求体：manifest JSON，或 `{"manifest": {...}}` 包装体。
- 返回：`valid`、`errors`、`warnings`、`impact_summary`、`normalized_manifest`、`checksum`、`dependency_summary`、`dependencies`、`compatibility`、`content_type_conflicts`、`permission_conflicts`、`migration_plan`、`install_preview`。
- `dependencies` 为逐项检查结果，包含 `code`、`version`、`required`、`reason`、`status`、`satisfied`、`current_version`、`current_status`、`message`、`chain`。
- `dependency_summary` 聚合 `total`、`required`、`optional`、`satisfied`、`warnings`、`blocking`、`missing`、`disabled`、`archived`、`version_issues`、`cycles`。
- `compatibility` 包含 `core_version`、`min_core_version`、`compatible_core_version`、`status` 和 `messages`。`min_core_version` 高于当前 Core 会返回 `plugin_core_version_incompatible` 错误。

`POST /api/v1/admin/plugins/dry-run`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：对 manifest 做安装前 dry-run，校验冲突、依赖和影响分析，不写入插件记录。
- 返回：与 manifest validate 一致的结果视图；required 依赖不满足时 `valid=false`，optional 依赖不满足时仅进入 warning。

`POST /api/v1/admin/plugins/packages/dry-run`

- 认证：后台 admin token（不允许 user token / moderator token）。
- 权限：`plugin.write`。
- 用途：本地插件包 dry-run 导入预览。只做安全读取 + 文件扫描 + manifest 校验 + 安装预览，不安装插件、不执行插件代码、不执行外部 SQL、不动态加载前端资产。
- 请求：

```json
{
  "path": "examples/plugins/demo_notice"
}
```

- 路径限制：只允许读取项目根目录下白名单目录：
  - `examples/plugins/`
  - `plugins-local/`
  - `storage/plugins/packages/`
  - `storage/plugins/exports/`
  - `storage/plugins/staging/`
  - `storage/plugins/quarantine/`
  - `.devhub/plugins/`
- 返回：包含 `package`、`file_scan`、`checksum`、`signature`、`risk_report`、`manifest_validation`、`install_dry_run`、`status`、`blocked_code`、`blocked_reasons`、`warnings`、`errors`。
- `status`：`ok|warning|blocked`。
- `blocked_code`：当 `status=blocked` 时返回阻断原因代码（例如 `plugin_package_dangerous_file` / `plugin_package_manifest_invalid`）。
  - `blocked_reasons`：可选，返回所有阻断原因 code（用于 UI 逐项展示）。
  - `checksum.status`：`ok|warning|failed|missing`；缺失 checksums.json 为 `missing`（warning），不匹配为 `failed`（blocked）。
  - `risk_report.level`：`low|medium|high|blocked`，由后端根据扫描/校验/依赖/兼容结果评估，前端不得伪造。
  - `signature`：签名与可信来源摘要（可选能力，但字段会返回）。
    - `trust_status`：`trusted|unknown|blocked|revoked|unsigned`
    - `verification_status`：`verified|failed|missing|unsupported|publisher_unknown`
    - `payload`：当前仅支持 `checksums.json`。
    - `payload_algorithm`：当前仅支持 `sha256`，真实验签消息为 `sha256(raw bytes of checksums.json)`。
    - `fingerprint`：可信发布者公钥指纹，格式为 `sha256:<base64url>`。
    - 说明：`publisher.json` 不会被自动信任；可信状态只来自后台可信发布者记录，`storage/plugins/trusted_publishers.json` 仅作为空存储时的本地 seed / fallback。

`POST /api/v1/admin/plugins/packages/upload`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 请求：`multipart/form-data`，字段 `file` 必须是 `.zip`。
- 用途：上传 zip 插件包，保存到 `storage/plugins/uploads/`，解压到唯一 `storage/plugins/staging/{upload_id}/` 沙箱，随后复用本地插件包 scanner / checksum / signature / risk_report / dry-run。上传后不会安装插件、不会执行代码、不会执行 SQL、不会动态加载前端资产。
- 限制：
  - 上传 zip 最大 20MB。
  - 解压后总大小最大 50MB。
  - 解压后文件数最大 300。
  - 单个解压文件最大 5MB。
  - 目录深度最大 8。
  - 禁止 zip 内嵌 zip/tar/gz/rar/7z 等压缩包。
  - 禁止 `../`、绝对路径、Windows 盘符、空路径、过长文件名、URL 编码绕过、symlink、hardlink 和特殊设备文件。
  - 当前实现不做压缩比阈值检查；通过总大小、单文件、文件数、目录深度、重复 entry 和嵌套压缩包限制防御 zip bomb。
- 插件包根目录识别：
  - zip 根目录直接存在 `manifest.json` 时，以根目录为包目录。
  - zip 只有单个顶层目录且该目录内存在 `manifest.json` 时，以该顶层目录为包目录。
  - 找不到 `manifest.json` 或发现多个 `manifest.json` 时 blocked。
- 返回：`upload_id/filename/status/staging_path/package_path/zip_scan/file_scan/checksum/signature/manifest_validation/install_dry_run/risk_report/can_promote/can_submit_approval/warnings/errors`。路径只返回项目内相对路径。
- blocked 策略：zip 结构类错误会删除半包并返回结构化错误；解压成功但 package dry-run blocked 的包会移动到 `storage/plugins/quarantine/{upload_id}/` 供查看风险，不可 promote。

`GET /api/v1/admin/plugins/packages/uploads/:upload_id`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：查看上传包详情；blocked/quarantine 包仍可查看详情。
- 返回：上传基础信息、当前状态、zip scan、package scan、checksum、signature、risk_report、manifest validate、dry-run、导入审批信息、安装审批信息和 `actions` 可执行动作列表。路径只返回项目相对路径，不返回系统绝对路径。

`GET /api/v1/admin/plugins/packages/uploads`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 查询参数：`status`、`risk_level`、`keyword`、`uploaded_by`、`package_code`、`publisher_id`、`trust_status`、`page`、`page_size`。
- 返回：`items`、`pagination`、`summary`。`items` 含 `upload_id/original_filename/package_code/package_name/package_version/status/risk_level/checksum_status/signature_status/trust_status/uploaded_by/uploaded_at/expires_at/promoted_path/approval_id/install_approval_id/error_code/risk_summary`。

`POST /api/v1/admin/plugins/packages/uploads/:upload_id/rescan`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 行为：仅 `uploaded/scanned/staged/blocked/failed` 可重新扫描；重新执行 package scanner / checksum / signature / manifest validate / install dry-run 并刷新 risk_report 与状态。不会安装插件、不会执行代码/SQL、不会动态加载资产。

`POST /api/v1/admin/plugins/packages/uploads/:upload_id/approval`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 行为：为 `staged` 上传包提交导入 / promote 审批，复用 `plugin_approval_requests`，`action=package_promote`，审批快照包含 `upload_id`、risk_report、signature、checksum、dry-run。

`POST /api/v1/admin/plugins/packages/uploads/:upload_id/approve`

- 认证：后台 admin token。
- 权限：`plugin.approve`。
- 行为：审批通过导入申请，状态从 `approval_pending` 变为 `approved`。审批通过不绕过 promote 前重新扫描。

`POST /api/v1/admin/plugins/packages/uploads/:upload_id/reject`

- 认证：后台 admin token。
- 权限：`plugin.approve`。
- 行为：拒绝导入申请，状态变为 `approval_rejected`；后续需要重新上传或按策略重新提交审批。

`POST /api/v1/admin/plugins/packages/uploads/:upload_id/promote`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：将已上传且通过校验的 staging 包复制到 `storage/plugins/packages/{manifest.code}/`，进入本地插件仓库；不会安装插件。
- 请求：

```json
{ "force": false }
```

- 行为：
  - 仅 `staged/approved` 且非 blocked、checksum 非 failed、manifest valid 的 staging 上传包可 promote。
  - promote 前重新 dry-run；复制后再次 dry-run 复检。
  - 目标目录已存在默认拒绝，返回 `plugin_package_promote_target_exists`。
  - blocked、quarantine、dangerous file、checksum mismatch、manifest invalid 均禁止 promote。
  - promote 写入 `admin_logs`，但不写插件表、不写 migration 表、不执行代码/SQL、不动态加载资产。

`POST /api/v1/admin/plugins/packages/uploads/:upload_id/cancel`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 行为：取消未 promote / 未安装 / 未删除 / 未过期的上传包，状态变为 `canceled`，写入 `admin_logs`。

`DELETE /api/v1/admin/plugins/packages/uploads/:upload_id`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 行为：删除上传包记录对应的 upload / staging 文件并将记录标为 `deleted`；若已 `promoted`，不会删除 `storage/plugins/packages/` 中的本地仓库包；若已安装，不会卸载插件。`approval_pending/install_approval_pending` 不能直接删除。

`POST /api/v1/admin/plugins/packages/uploads/cleanup`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 行为：手动清理 `expired/deleted/canceled/failed` 上传包的 upload / staging 文件，并可将超过 `expires_at` 的非终态记录标为 `expired`。不会删除本地插件仓库、不会删除已安装插件。本轮不做定时任务。

`POST /api/v1/admin/plugins/packages/templates/preview`

- 认证：后台 admin token（不允许 user token / moderator token）。
- 权限：`plugin.write`。
- 用途：预览后台“初始化插件包”将要生成的声明型插件模板；不写入文件、不创建目录、不执行 dry-run。
- 请求：

```json
{
  "code": "demo_notice",
  "name": "示例公告插件",
  "content_type": "notice",
  "content_name": "公告",
  "description": "用于公告内容类型的声明型插件。",
  "author": "DevHub Team",
  "with_config": true,
  "with_hooks": true,
  "with_migration": true
}
```

- 输出目录固定为 `storage/plugins/packages/{code}`；不接受任意 path，不暴露 `force` 覆盖。
- 返回：

```json
{
  "template": {
    "code": "demo_notice",
    "name": "示例公告插件",
    "content_type": "notice",
    "content_name": "公告",
    "output_dir": "storage/plugins/packages/demo_notice",
    "package_path": "storage/plugins/packages/demo_notice",
    "files": ["manifest.json", "README.md", "config.example.json", "docs/registry-example.md"]
  },
  "status": "ok",
  "warnings": []
}
```

`POST /api/v1/admin/plugins/packages/templates`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：在 `storage/plugins/packages/{code}` 下初始化声明型插件包模板，然后服务端自动执行 package dry-run。
- 行为：
  - 复用 CLI `plugin:new` 的同一套模板生成服务，不复制两套生成逻辑。
  - 后台初始化默认不生成 `registry.example.go`，避免 `.go` 文件触发插件包 dangerous file 阻断；registry 接入说明写入 `docs/registry-example.md`。
  - 目标目录已存在时返回 `plugin_package_template_exists`，当前后台不支持覆盖。
  - 初始化成功后写入 `admin_logs`，并返回自动 dry-run 的 `status/risk_report/manifest_validation/warnings/errors`。
- 返回：

```json
{
  "message": "插件包模板已初始化，并已完成 package dry-run",
  "template": {
    "package_path": "storage/plugins/packages/demo_notice",
    "files": ["manifest.json", "README.md", "config.example.json", "docs/registry-example.md"]
  },
  "dry_run": {
    "status": "warning",
    "risk_report": { "level": "medium" },
    "manifest_validation": { "valid": true }
  },
  "status": "warning",
  "warnings": []
}
```

`GET /api/v1/admin/plugins/packages`

- 认证：后台 admin token（不允许 user token / moderator token）。
- 权限：`plugin.read`。
- 用途：扫描本地插件仓库目录并返回 discovered packages 列表（只读扫描/校验/展示，不安装、不执行、不动态加载）。
- 查询参数：
  - `root`：可选，默认 `storage/plugins/packages`
  - `status`：可选，`all|ok|warning|blocked|invalid`
  - `keyword`：可选，按 `code/name/path` 模糊搜索
  - `risk_level`：可选，`low|medium|high|blocked`
  - `checksum_status`：可选，`ok|warning|failed|missing`
  - `manifest_valid`：可选，`true|false`
  - `page` / `page_size`
- 返回：`items/pagination/summary`，其中 `items` 每项包含 `path/code/name/version/status/risk_level/risk_summary/checksum_status/manifest_valid/total_files/total_size/updated_at/warnings/errors`，以及可选签名摘要 `signature`（同 dry-run）。

`GET /api/v1/admin/plugins/packages/detail`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：查看单个插件包详情（复用 dry-run 结果视图；blocked/invalid 也可查看原因）。
- 查询参数：
  - `path`：必填，例如 `storage/plugins/packages/demo_notice`
- 返回：与 `POST /api/v1/admin/plugins/packages/dry-run` 相同的结构（包含 `checksum/signature/risk_report/manifest_validation/install_dry_run` 等）。

`GET /api/v1/admin/plugins/trusted-publishers`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：查看本地可信发布者列表；该列表是插件包可信状态的唯一来源。
- 查询参数：`status`、`keyword`、`page`、`page_size`。
- 返回：`items/pagination/summary`。`items` 包含 `publisher_id/name/homepage/email/public_key_id/public_key_algorithm/public_key/fingerprint/status/notes/created_by/updated_by/created_at/updated_at/revoked_at/blocked_at`。

`GET /api/v1/admin/plugins/trusted-publishers/:id`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：查看单个可信发布者详情。

`POST /api/v1/admin/plugins/trusted-publishers`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：新增可信发布者公钥；不会保存私钥，也不会从远程同步。
- 请求：

```json
{
  "publisher_id": "devhub-official",
  "name": "DevHub Official",
  "homepage": "https://example.com",
  "email": "security@example.com",
  "public_key_id": "devhub-official-2026",
  "public_key_algorithm": "ed25519",
  "public_key": "base64-ed25519-public-key",
  "notes": "本地可信发布者"
}
```

`PUT /api/v1/admin/plugins/trusted-publishers/:id`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：更新可信发布者元信息或公钥；`publisher_id + public_key_id` 必须唯一。

`POST /api/v1/admin/plugins/trusted-publishers/:id/block`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：将发布者标记为 blocked；后续匹配该公钥的插件包 dry-run / promote / install / approval 执行会 blocked。

`POST /api/v1/admin/plugins/trusted-publishers/:id/revoke`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：将发布者标记为 revoked；后续匹配该公钥的插件包会 blocked。

`POST /api/v1/admin/plugins/trusted-publishers/:id/restore`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：恢复为 trusted。

`DELETE /api/v1/admin/plugins/trusted-publishers/:id`

- 认证：后台 admin token。
- 权限：`plugin.manage`。
- 用途：删除本地可信发布者记录；删除后对应插件包会变为 unknown publisher，而不会自动信任包内 publisher.json。

`POST /api/v1/admin/plugins/packages/install`

- 认证：后台 admin token。
- 权限：`plugin.approve`（仅审批人可直接执行；普通管理员需先提交审批申请）。
- 用途：从**本地插件包**安装声明型插件（最小闭环）。服务端会强制复跑 package dry-run（scan/checksum/risk/manifest validate/install preview），通过后复用现有 manifest install 写入插件记录。
- 请求：

```json
{
  "path": "storage/plugins/packages/demo_notice",
  "confirm_risk_level": "low"
}
```

- 行为与边界：
  - 只写入 manifest 声明、默认配置、迁移 pending 记录与审计；不执行第三方代码、不执行外部 raw SQL、不动态加载前端资产。
  - 安装成功后插件状态固定为 `disabled`（不自动启用）。
  - 若同 `code` 插件已安装，返回 `plugin_package_already_installed`（提示走 upgrade）。
- 返回：`plugin`（含 `source_type=local_package`）、`package`、`checksum`、`risk_level`、`install_result`、`warnings`。

`POST /api/v1/admin/plugins/:code/export/dry-run`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：预览“已安装声明型插件”导出为本地插件包目录；不写文件、不生成目录、不导出真实配置。
- 请求：

```json
{
  "include_docs": true,
  "include_migrations": true,
  "include_publisher": false,
  "include_signature_stub": false
}
```

- 返回字段：
  - `plugin_code` / `version` / `status`：插件与预览状态（`ok|warning|blocked`）。
  - `export_preview.files`：预计生成的相对文件列表，至少包含 `manifest.json`、`README.md`、`config.example.json`、`checksums.json`。
  - `export_preview.output_dir`：预计输出目录，默认在 `storage/plugins/exports/{plugin_code}-{version}-{timestamp}`。
  - `contains_sensitive_values=false` / `contains_user_data=false` / `contains_runtime_code=false` / `contains_external_sql=false`：导出安全边界检查。
  - `warnings` / `errors`：可读提示；敏感字段不会进入 details。

`POST /api/v1/admin/plugins/:code/export`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：正式导出已安装声明型插件为受控本地插件包目录。
- 请求字段：同 dry-run，额外支持 `output_dir`（必须位于 `storage/plugins/exports/`）和 `force`（已存在目录时是否清理重写）。
- 行为：
  - 执行前强制调用导出 dry-run。
  - 写入 `manifest.json`、`README.md`、`config.example.json`、`checksums.json`，可选写入 `docs/usage.md`、`migrations/exported_migrations.json`、`publisher.json`、`signature.json` 结构草案。
  - `config.example.json` 只来自 `default_config` 或 `config_schema` 示例生成，敏感字段统一写为 `REPLACE_ME`，不会导出当前全局/子站真实配置，也不会导出 `enc:v1:` 密文。
  - 生成 `checksums.json` 后自动执行 package dry-run 自检；自检失败会以 warning 返回，不会写入插件安装数据。
  - 成功/失败均写入 `admin_logs`，details 不包含敏感配置、用户数据或密文。
- 返回字段：
  - `plugin_code` / `version` / `output_dir`
  - `files`
  - `checksum_status=generated`
  - `package_dry_run_status=ok|warning|failed`
  - `warnings`

边界：

- 不生成 zip、不远程上传、不发布到插件市场。
- 不导出用户数据、帖子/评论/通知/审计原始明细、Hook 执行历史、搜索索引、运行时代码、外部 SQL、私钥或任何真实 token/secret/password/credential。

`POST /api/v1/admin/plugins/approvals`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：创建插件治理审批申请（本轮覆盖：本地插件包 install、插件 upgrade）。
- 请求示例：

```json
{
  "action": "install",
  "package_path": "storage/plugins/packages/demo_notice",
  "reason": "安装公告插件用于测试"
}
```

- 返回：审批记录（`id/status/action/plugin_code/package_path/risk_report_json/dry_run_json` 等）。

`GET /api/v1/admin/plugins/approvals`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：审批列表（分页/筛选）。
- 查询参数：
  - `status`：可选，`pending|approved|rejected|canceled|executed|failed`
  - `action`：可选，`install|upgrade`
  - `plugin_code`：可选
  - `requested_by` / `reviewed_by`：可选
  - `page` / `page_size`

`GET /api/v1/admin/plugins/approvals/:id`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：审批详情（包含 dry-run/risk 快照；不包含敏感配置明文或密文）。

`POST /api/v1/admin/plugins/approvals/:id/approve`

- 认证：后台 admin token。
- 权限：`plugin.approve`。
- 用途：审批通过。
- 请求：`{"comment":"...可选..."}`。

`POST /api/v1/admin/plugins/approvals/:id/reject`

- 认证：后台 admin token。
- 权限：`plugin.approve`。
- 用途：审批拒绝（comment 必填）。
- 请求：`{"comment":"...必填..."}`。

`POST /api/v1/admin/plugins/approvals/:id/cancel`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：撤销 pending 申请（默认仅申请人可撤销；审批人可强制撤销）。

`POST /api/v1/admin/plugins/approvals/:id/execute`

- 认证：后台 admin token。
- 权限：`plugin.approve`。
- 用途：执行已通过审批的申请。服务端会强制**重新 dry-run 校验**（scan/checksum/risk/依赖/兼容），校验通过后才会真正安装/升级。

`POST /api/v1/admin/plugins/install`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：安装 manifest + 配置型插件。
- 行为：写入插件记录、默认配置、迁移待处理记录和审计，但不执行第三方代码。
- 初始状态：`install_status=installed`、`runtime_status=disabled`。
- 阻断：required dependency 缺失、disabled、archived、migration_failed、config_invalid、version_mismatch、自依赖、循环依赖，或 Core 版本不兼容时拒绝安装；optional dependency 缺失允许安装但返回 warning。


### 插件 dependencies 与 Core 兼容规则

`dependencies` 当前支持两种输入：旧兼容字符串数组（如 `["qa"]`，视为 required）和对象数组：

```json
{
  "code": "qa",
  "version": ">=1.0.0 <2.0.0",
  "required": true,
  "reason": "需要问答内容类型作为数据来源"
}
```

字段：`code` 为依赖插件编码；`version` 为版本约束；`required` 默认为 `true`；`reason` 为可选说明。当前版本约束只支持数字 `x.y.z`、精确版本、`>=`、`>`、`<=`、`<` 和空格连接的范围组合（如 `>=1.2.0 <2.0.0`），不支持 `^`、`~`、`||`、预发布标签或 npm 完整语法。

依赖状态：`satisfied`、`missing`、`disabled`、`archived`、`migration_failed`、`config_invalid`、`version_mismatch`、`circular_dependency`、`self_dependency`、`optional_missing`。required 不满足会阻断 validate / dry-run / install / upgrade dry-run / upgrade / enable；optional 缺失只 warning。自依赖和循环依赖当前一律阻断，包括 optional 循环。

Core 兼容：`min_core_version` 缺失为 warning；`min_core_version` 高于当前 Core 为 `plugin_core_version_incompatible` 并阻断；`compatible_core_version` 存在但当前 Core 不满足时按 incompatible 阻断。当前 Core 版本来自项目 `VERSION`，不是前端写死。

启用插件时会重新执行 required dependency 检查；依赖插件处于 disabled、archived、migration_failed、config_invalid、dependency_missing 时不能视为可用。错误码 / message 包含 `plugin_dependency_missing`、`plugin_dependency_disabled`、`plugin_dependency_archived`、`plugin_dependency_version_mismatch`、`plugin_dependency_cycle`、`plugin_core_version_incompatible` 等前缀或同义状态。

`POST /api/v1/admin/plugins/:code/upgrade/dry-run`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：对现有插件做升级预览，返回版本兼容矩阵、变更字段和 diff，不写入插件记录。
- 请求体：manifest JSON，要求 `code` 与路径中的 `:code` 一致。
- 返回：
  - `current_version`
  - `new_version`
  - `current_core_version`
  - `compatible_core_version`
  - `compatibility_status`
  - `changed_keys`
  - `diff.current`
  - `diff.new`
  - `validation`
  - `dependency_diff`：包含新增依赖、删除依赖、版本约束变化、required 变化和 changed_dependencies。
- 阻断：新 manifest 的 required dependency 或 Core 兼容检查不满足时，预览返回 `validation.valid=false` / blocked 信息，不写入数据。

`POST /api/v1/admin/plugins/:code/upgrade`

- 认证：后台 admin token。
- 权限：`plugin.approve`（仅审批人可直接执行；普通管理员需先提交升级审批申请）。
- 用途：执行 manifest + 配置型插件的安全升级。
- 请求体：manifest JSON，要求 `code` 与路径中的 `:code` 一致，且 `version` 必须高于当前版本。
- 阻断：required dependency 或 Core 兼容检查不满足时拒绝升级；不允许降级或同版本重复升级。
- 行为：
  - 校验 manifest。
  - 校验版本兼容性。
  - 更新插件 manifest / version / checksum / 声明元信息。
  - 为新增 migration 生成 `pending` 记录。
  - 保留历史配置、迁移记录和审计记录。
  - 不执行第三方代码，不执行外部 raw SQL。
- 返回：升级后的 `plugin` 对象，以及升级时使用的兼容矩阵 / diff / validation 摘要。

常见错误：

- `400 {"error":"插件不存在"}`
- `400 {"error":"config_json 必须是合法 JSON"}`
- `400 {"error":"$ 缺少必填字段 ..."}`
- `400 {"error":"$.field 必须是 boolean/string/number/integer/object"}`
- `400 {"error":"$.field 值不在允许范围 enum 内"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

`GET /api/v1/admin/plugins/:code/impact`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：插件治理中心的“禁用前影响分析”入口，用于展示禁用该插件对系统范围的影响计数。
- 返回：影响范围统计（计数型字段，尽量轻量且可缓存）。

响应示例：

```json
{
  "plugin_code": "qa",
  "existing_contents_count": 120,
  "enabled_communities_count": 3,
  "disabled_communities_count": 1,
  "categories_count": 5,
  "topics_count": 120,
  "recent_contents_count": 8,
  "pending_topics_count": 4,
  "pending_contents_count": 4,
  "menus_count": 3,
  "frontend_menus_count": 1,
  "moderator_menus_count": 1,
  "admin_menus_count": 1,
  "configs_count": 4,
  "pending_migrations_count": 0,
  "recent_hook_errors_count": 0
}
```

当前统计边界：

- `existing_contents_count` 与 `topics_count` 同步保留，前者是新治理口径，后者用于兼容旧 UI。
- `recent_contents_count` 统计近 7 天内容。
- `pending_contents_count` 与 `pending_topics_count` 同步保留。
- `recent_hook_errors_count` 来自 `hook_executions` 最近 7 天失败记录；该字段是轻量健康提示，不等同于完整 Hook 健康状态或重试系统。

`GET /api/v1/admin/plugins/:code/hooks`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：插件详情 Hooks Tab 展示运行时统计与最近执行记录。
- 返回：`items` 为按 Hook 聚合的统计，`recent_executions` 为最近 20 条执行记录。

响应示例：

```json
{
  "items": [
    {
      "hook_name": "BeforeCreateContent",
      "plugin_code": "qa",
      "mode": "blocking",
      "blocking": true,
      "execution_count": 12,
      "failure_count": 1,
      "avg_duration_ms": 1.3,
      "last_executed_at": "2026-05-11 15:30:01",
      "last_failed_at": "2026-05-11 15:20:10",
      "last_error": "qa 插件仅允许创建 question"
    }
  ],
  "recent_executions": [
    {
      "id": 1,
      "hook_name": "AfterCreateContent",
      "plugin_code": "qa",
      "mode": "non_blocking",
      "content_type": "question",
      "content_id": 123,
      "community_id": 1,
      "category_id": 101,
      "actor_type": "user",
      "actor_id": 2,
      "started_at": "2026-05-11 15:30:01",
      "finished_at": "2026-05-11 15:30:01",
      "duration_ms": 0,
      "success": true,
      "blocking": false,
      "metadata_json": "{\"handler_index\":0}"
    }
  ]
}
```

常见错误：

- `404 {"error":"插件不存在"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

`GET /api/v1/admin/plugins/:code/hooks/executions`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：Hooks 排障页查询某插件的 `hook_executions` 执行记录列表（支持基础筛选 + 分页）。该接口不提供重试、清空或告警能力。
- 查询参数（可选）：
  - `hook_name`
  - `mode`
  - `success`：`true/false/1/0`
  - `blocking`：`true/false/1/0`
  - `content_type`
  - `content_id`
  - `community_id`
  - `actor_type`
  - `actor_id`
  - `request_id`
  - `start_time` / `end_time`：`YYYY-MM-DD HH:mm:ss`
  - `page` / `page_size`（`page_size<=100`）
- 返回：

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

`POST /api/v1/admin/plugins/:code/hooks/:name/e2e-fail`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：仅用于 E2E / API 测试注入 Hook 失败，验证 blocking / non-blocking Hook 异常治理闭环。
- 启用条件：仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 的测试 / 开发环境可用；生产环境不应依赖该接口。
- 请求：

```json
{
  "mode": "blocking",
  "error_message": "E2E blocking hook failure"
}
```

- 清除注入：

```json
{
  "clear": true
}
```

- 行为：
  - `BeforeCreateContent` 等 blocking Hook 注入失败后，内容创建会被后端阻断，不写入脏数据，并写入 `hook_executions`。
  - `AfterCreateContent` 等 non-blocking Hook 注入失败后，主流程继续，内容仍可创建，并写入 `hook_executions`。
  - blocking 失败写入 `plugin.hook.blocked` 审计；non-blocking 失败写入 `plugin.hook.failed` 审计。
  - 注入 / 清除操作本身写入 `plugin.hook.test_injection` 审计。
- 说明：这是测试 helper，不是普通生产治理接口；真实 Hook 失败由内置 HookBus handler 返回错误产生。

常见错误：

- `404 {"error":"测试 Hook 注入接口未启用"}`
- `404 {"error":"插件不存在"}`
- `400 {"error":"hook mode 不合法"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

`GET /api/v1/admin/plugins/:code/audit-logs`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：插件详情“审计”Tab 的专用查询入口。
- 查询参数：
  - `community_id`：可选，按子站过滤。
  - `action`：可选，按动作关键字模糊过滤。
  - `type`：可选，默认 `all`。
  - `actor_type`：可选。
  - `actor`：可选，按 `actor_type` 模糊过滤。
  - `actor_user_id`：可选，按 `actor_id` 精确过滤。
  - `target_type`：可选，按 target 文本中的 target type 片段过滤。
  - `target_id`：可选，按 target 文本中的 target id 片段过滤。
  - `plugin_code`：可选，默认使用路径中的 `code`，同时匹配 target / metadata / old_value / new_value。
  - `metadata`：可选，按 `metadata_json` 关键字过滤。
  - `request_id`：可选，按 `metadata_json` 里的 request id 或 request id 字符串过滤。
  - `start_time`、`end_time`：可选，按 `created_at` 字符串时间范围过滤。
  - `target`：可选，默认使用插件 code 模糊匹配。
  - `page`、`page_size`：分页参数。
- 返回：`domain.PageResponse`，items 为 `AdminLog`，包含 `old_value`、`new_value`、`metadata_json`。
- 覆盖范围：插件启停、插件配置、子站插件配置、子站插件排序、Hook 失败审计，以及带 `plugin_code` 的插件内容治理操作。
- 阶段 B / v1.4 后台联动：通用 PluginContent 页的“查看审计日志”入口会跳转到 `/admin-next/audit-logs`，并通过 query 预填 `plugin_code`、`content_type`、`action=批量治理主题`、`target_type=topics`、`metadata=<plugin_code>`；通用审计页会读取这些 query 并带入 `/api/v1/admin/audit-logs` 查询。

响应示例：

```json
{
  "items": [
    {
      "id": 100,
      "type": "system",
      "actor_type": "admin_user",
      "actor_id": 1,
      "action": "更新插件状态",
      "target": "plugins#qa",
      "site": "community:1",
      "old_value": "{\"status\":\"enabled\"}",
      "new_value": "{\"status\":\"disabled\"}",
      "metadata_json": "{\"scope\":\"global\",\"plugin_code\":\"qa\",\"operation\":\"plugin_status\",\"request_id\":\"req-xxx\"}",
      "created_at": "2026-05-11 16:00:00"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "has_more": false
}
```

常见错误：

- `404 {"error":"插件不存在"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

`GET /api/v1/admin/plugins/:code/migrations`

- 认证：后台 admin token。
- 权限：`plugin.read`。
- 用途：插件详情“迁移”Tab 查询内置插件 migration 声明与执行记录。
- 返回：`items` 为迁移列表，`summary` 为 success / pending / failed 计数。
- 当前策略：只支持内置插件 up migration；已成功迁移不会重复执行；down / rollback 仅保留 `rollback_supported` 标识，不执行真实回滚。

响应示例：

```json
{
  "items": [
    {
      "plugin_code": "qa",
      "migration_version": "1.0.0",
      "migration_name": "qa_questions",
      "direction": "up",
      "checksum": "builtin:qa:qa_questions:v1",
      "status": "success",
      "finished_at": "2026-05-11 18:00:00",
      "duration_ms": 0,
      "rollback_supported": false,
      "declared": true
    }
  ],
  "summary": {
    "plugin_code": "qa",
    "total": 2,
    "pending": 0,
    "failed": 0,
    "success": 2
  }
}
```

`POST /api/v1/admin/plugins/:code/migrations/run`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：执行该插件所有声明的待处理 migration。
- 行为：内置 qa/docs/wiki migration 当前是幂等 no-op 校验记录；schema 表结构由 `001_schema.sql` 和启动迁移保证，runner 负责写入 running / success / failed 状态与审计。
- 审计：写入 `plugin.migration.run` 与 `plugin.migration.success`；失败写入 `plugin.migration.failed`。

`POST /api/v1/admin/plugins/:code/migrations/:name/retry`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：执行或重试单条 migration。
- 行为：如果该 migration 已是 `success`，接口返回现有成功记录，不重复破坏数据。
- 审计：写入 `plugin.migration.retry` 与 `plugin.migration.success`；失败写入 `plugin.migration.failed`。

`POST /api/v1/admin/plugins/:code/migrations/:name/e2e-fail`

- 认证：后台 admin token。
- 权限：`plugin.write`。
- 用途：仅用于 E2E / API 测试构造 failed migration，验证失败迁移阻断全局启用和子站启用。
- 启用条件：仅在 `DEVHUB_E2E_TESTING=1` 或 `CMS_STORE=memory` 的测试 / 开发环境可用；生产 MySQL 环境不应依赖该接口。
- 请求：

```json
{
  "error_message": "E2E forced migration failure"
}
```

- 行为：写入或覆盖该内置 migration 的 `failed` 记录；随后 `POST /api/v1/admin/plugins/:code/enable` 与 `POST /api/v1/admin/communities/:id/plugins/:code/enable` 都会被 Service readiness 拦截。
- 审计：写入 `plugin.migration.failed`，`metadata_json.operation=plugin_migration_test_injection`。
- 说明：这是测试 helper，不是普通生产治理接口；真实生产失败由 migration runner 写入。

常见错误：

- `404 {"error":"插件不存在"}`
- `400 {"error":"迁移不存在"}`
- `400 {"error":"当前仅支持 up migration"}`
- `404 {"error":"测试迁移注入接口未启用"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`

### 前台子站插件展示 API

`GET /api/v1/communities/:slug/plugins`

- 认证：不需要。
- 路径参数：`slug` 为子站 slug。
- 返回：当前子站可用插件，只包含“全局 enabled + 子站 enabled”的插件。
- 安全处理：不返回后台配置字段 `config_json`，也不暴露 `global_status` / `community_status`。
- 用途：前台子站首页、发布页和导航收口。

常见错误：

- `404 {"error":"子站不存在"}`
- `400 {"error":"..."}`：读取子站插件状态失败。

### 后台子站插件 API

`GET /api/v1/admin/communities/:id/plugins`

- 认证：后台 admin token。
- 权限：`site.read`，并经过子站管理范围校验。
- 返回：某个子站的插件列表，包含全局状态叠加后的子站状态、内容类型、菜单、权限、`config_json` 和 `resolved_config`。

`POST /api/v1/admin/communities/:id/plugins/:code/enable`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 规则：全局 disabled 插件不能被子站启用。
- 返回：更新后的插件对象。
- 审计：写入子站插件状态变更审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

`POST /api/v1/admin/communities/:id/plugins/:code/disable`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 影响：只影响该子站的新发布、导航、菜单和管理入口，不影响历史内容访问。
- 返回：更新后的插件对象。
- 审计：写入子站插件状态变更审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

`PUT /api/v1/admin/communities/:id/plugins/:code/config`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 请求：

```json
{
  "config_json": {
    "example": true
  }
}
```

- 返回：更新后的插件对象，包含子站覆盖后的 `resolved_config.default/global/community/effective`。
- 校验：会按插件 `config_schema` 执行后端强校验（简化 JSON Schema），至少覆盖 `type`、`required`、`enum`、`object`、`boolean`、`string`、`number`、`integer`、`min/max`、`default` 与未知字段策略。
- 审计：写入子站插件配置审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段；`metadata_json.changed_keys` 记录本次变更的顶层配置键。
- 清空：提交 `{"config_json": null}` 或空配置会清空子站覆盖配置，并同样写入配置审计。

`PUT /api/v1/admin/communities/:id/plugins/sort`

- 认证：后台 admin token。
- 权限：`site.write`，并经过子站管理范围校验。
- 请求：

```json
{
  "codes": ["qa", "docs", "wiki"]
}
```

- 返回：

```json
{
  "updated": 3
}
```

- 审计：写入子站插件排序审计日志。
  当前同时写入 `admin_logs.target` 文本摘要和 `old_value` / `new_value` / `metadata_json` 结构化字段。

常见错误：

- `400 {"error":"插件不存在"}`
- `400 {"error":"插件全局未启用，不能在子站启用"}`
- `400 {"error":"插件状态不合法"}`
- `401 {"error":"未登录"}`
- `403 {"error":"无权限"}`
- `404 {"error":"子站不存在"}`

`GET /api/v1/admin/communities/:id/plugins/:code/impact`

- 认证：后台 admin token。
- 权限：`site.read`，并经过子站管理范围校验。
- 用途：子站插件治理的“禁用前影响分析”入口，用于展示在某个子站范围内禁用该插件的影响计数。
- 返回：同全局 impact，但计数会尽量收敛到该子站范围（例如该子站板块数、该子站内容数）。

### 插件菜单 API

`GET /api/v1/admin/plugin-menus`

- 认证：后台 admin token。
- 返回：全局 enabled 插件的 `admin` 菜单，并按当前后台用户权限过滤。

`GET /api/v1/moderator/plugin-menus`

- 认证：前台 user token。
- 授权：当前用户必须是启用状态子站版主。
- 查询参数：可选 `community_slug` 或 `community_id`。不传时按当前用户可治理子站范围汇总去重。
- 返回：同时满足全局 enabled、子站 enabled、当前用户有权限的 `moderator` 插件菜单。

常见错误：

- `401 {"error":"未登录"}`
- `403 {"error":"当前用户不是启用状态子站版主"}`
- `403 {"error":"无权管理该子站内容"}`

## 发布 API 与插件校验

`POST /api/v1/topics`

- 认证：前台 user token。
- 请求字段：`community_id` / `community_slug`、`category_id`、`content_type`、`title`、`summary`、`content`、`tags` 等。
- 写入：保存归一后的 `topics.content_type` 和 `topics.plugin_code`。

插件校验流程：

1. 解析 community。
2. 解析 category，并校验 category 属于 community。
3. 归一 `content_type`：`doc -> document`，`wiki -> wiki_page`。
4. 根据内容类型推断插件：`question -> qa`，`document -> docs`，`wiki_page -> wiki`，`project -> projects`，`job -> jobs`，`ai_work -> ai_works`，其他 Core 兼容类型 -> `core`。
5. 校验全局插件是否 enabled。
6. 校验当前子站插件是否 enabled。
7. 校验 `category.plugin_code` 是否匹配。
8. 校验 `content_type` 是否在 `category.allowed_content_types` 内。
9. 校验当前用户是否具备内容类型对应的发布权限码。

常见错误：

- `403 {"error":"缺少权限 qa.question.create，不能创建该类型内容"}`
- `403 {"error":"缺少权限 docs.document.create，不能创建该类型内容"}`
- `403 {"error":"缺少权限 wiki.page.create，不能创建该类型内容"}`
- `400 {"error":"内容类型不能为空"}`
- `400 {"error":"内容类型不合法"}`
- `400 {"error":"插件全局未启用"}`
- `400 {"error":"当前子站未启用该插件"}`
- `400 {"error":"板块不存在"}`
- `400 {"error":"当前板块未绑定对应插件"}`
- `400 {"error":"内容类型与板块不匹配"}`

当前限制：

- 发布链路已校验插件状态、子站插件状态和板块约束。
- 发布链路已增加插件权限码校验：
  - `question -> qa.question.create`
  - `document -> docs.document.create`
  - `wiki_page -> wiki.page.create`
  - `project -> projects.project.create`
  - `job -> jobs.job.create`
  - `ai_work -> ai_works.work.create`
  - `article`、`news` 当前仍使用粗粒度 `core.topic.create`（兼容旧 `post.create`）。
- 当前内容类型权限码来自统一 `ContentTypeDefinition` 声明，而不是散落在各个 handler 中。
- `projects/jobs/ai_works` 当前已完成插件归属、发布校验、权限码和菜单声明；专属扩展表与完整业务流程仍是后续任务。

`POST /api/v1/admin/posts`

- 认证：后台 admin token。
- 基础权限：`post.create`，作为历史后台内容创建兼容权限。
- 动态权限：接口会先将请求转换为 Topic 创建请求，归一 `content_type` 并推断 `plugin_code`，再叠加校验真实内容类型对应的插件 create 权限。
- 写入链路：内部调用 `Service.CreateTopic`，继续执行全局插件状态、子站插件状态、板块绑定、`allowed_content_types` 和发布权限码校验。

动态 create 权限：

- `question -> qa.question.create`
- `document -> docs.document.create`
- `wiki_page -> wiki.page.create`
- `project -> projects.project.create`
- `job -> jobs.job.create`
- `ai_work -> ai_works.work.create`
- `article/news -> core.topic.create`，并兼容旧 `post.create`

常见错误：

- `403 {"error":"缺少权限 qa.question.create，不能创建该类型内容"}`
- `403 {"error":"缺少权限 docs.document.create，不能创建该类型内容"}`
- `403 {"error":"缺少权限 wiki.page.create，不能创建该类型内容"}`
- `400 {"error":"插件全局未启用"}`
- `400 {"error":"当前子站未启用该插件"}`
- `400 {"error":"内容类型与板块不匹配"}`

说明：`admin/posts` 是后台兼容入口，不是绕过插件体系的独立写入口。`post.create` 只是第一层兼容基础权限，真实内容类型权限由插件权限码决定。

`GET /api/v1/admin/posts`

- 认证：后台 admin token。
- 用途：后台内容治理列表和通用 `PluginContent` 插件内容治理页。
- 主要筛选：
  - `site`：子站 slug；`portal` 表示全局视角。
  - `board`：板块；`all` 表示不限板块。
  - `q`：标题 / 摘要 / 正文关键词。
  - `status`：`all`、`publish`、`offline`、`pinned`、`recommended` 等。
  - `content_type`：归一后的内容类型，如 `question`、`document`、`wiki_page`。
  - `plugin_code`：插件编码，如 `qa`、`docs`、`wiki`。
- 插件内容页必须同时传 `plugin_code` 和 `content_type`，后端按 AND 精确过滤，避免前端用 OR 兜底混入其他插件内容。
- 返回行会带上 `plugin_code` 和 `content_type`；历史数据缺失时由后端按板块 / 内容类型做防御性归一。

`PUT /api/v1/admin/posts/:id`

- 认证：后台 admin token。
- 权限：`post.update`，并经过当前后台用户的子站治理范围校验。
- 当前策略：禁止修改内容归属和内容类型。
- 允许更新：标题、摘要、正文、标签、状态、置顶、精华等非归属字段。
- 禁止更新：`site/community_id`、`board/category_id`、`content_type`、`plugin_code`。

常见错误：

- `400 {"error":"后台编辑不允许修改内容归属子站，请通过迁移专项处理"}`
- `400 {"error":"后台编辑不允许修改内容板块或内容类型，请通过迁移专项处理"}`

说明：如果后续需要后台迁移内容子站、板块或插件类型，应新增迁移专项接口，并逐条校验插件全局状态、子站插件状态、板块绑定、`allowed_content_types` 和对应插件权限码。

## 插件声明约定

当前内置插件统一按 manifest 风格组织：

- `PluginManifest`
- `ContentTypeDefinition`
- `PermissionDefinition`
- `MenuDefinition`
- `RouteDefinition`
- `HookDefinition`

当前说明：

- 这是当前内置系统插件规范；完整插件系统是最高优先级长期主线。插件市场、插件包、远程安装、在线更新和动态加载进入 P2 / P3 路线，但不是当前真实 API。
- `HookDefinition` 当前是扩展点声明；`Service` 已有内部 HookBus，当前调用点覆盖 `BeforeCreateContent`、`AfterCreateContent`、`BeforeUpdateContent`、`AfterUpdateContent`、`BeforeDeleteContent`、`AfterDeleteContent`、`AfterCreateComment`、`OnSearchIndex`、`OnNotificationBuild` 和 `OnSEOBuild`。
- HookBus 执行会写入 `hook_executions`；blocking hook 失败会返回错误并写入 `plugin.hook.blocked` 审计，non-blocking hook 失败不会阻断主流程但会写入 `plugin.hook.failed` 审计。
- Search / Notification / SEO 当前只是最小事件派发，尚未实现复杂索引、通知模板或结构化 SEO 插件处理器。
- 配置优先级按“默认配置 -> `plugins.config_json` -> `community_plugins.config_json`”合并；API 用 `resolved_config.default/global/community/effective` 表达合并视图。
- 当前配置已完成 JSON 合法性校验和简化 `config_schema` 基础校验；后台基础自动表单、配置 diff UI 和 effective config 预览已接入插件治理体验。更完整 JSON Schema、深层嵌套、字段分组、配置版本回滚是后续插件平台任务。

## 插件平台基线对账

本节记录当前真实 API 与平台能力边界，避免把目标路线图误读为当前已完成接口。

已完成：

- 全局插件：`GET /api/v1/plugins`、`GET /api/v1/admin/plugins`、全局启用 / 禁用、全局配置、全局 impact。
- 子站插件：前台子站插件展示、后台子站插件列表、启用 / 禁用、配置、排序、子站 impact。
- 插件菜单：后台插件菜单和版主插件菜单会按插件状态、子站状态和权限过滤。
- 插件配置：全局与子站配置 API 保存时会做 JSON 合法性校验和简化 `config_schema` 校验，返回 `resolved_config`。
- 插件审计：插件启停、配置和排序操作写入 `old_value`、`new_value`、`metadata_json`。
- 插件健康：`GET /api/v1/admin/plugins` 返回轻量 `health` 摘要；插件详情可通过“运行状态”Tab 查看配置、迁移、Hook、依赖和最近错误。
- 插件健康 API：`GET /api/v1/admin/plugins/health` 返回健康总览，`GET /api/v1/admin/plugins/:code/health` 返回单插件健康摘要。
- 插件迁移：`plugin_migrations` 记录表、内置 migration 声明、迁移查询 API、执行 / 重试 API、迁移审计和后台迁移 Tab 已完成第一阶段闭环。
- Manifest 校验与 dry-run：`POST /api/v1/admin/plugins/manifest/validate` 和 `POST /api/v1/admin/plugins/dry-run` 可校验 manifest、冲突、依赖和安装影响，dry-run 不写入插件记录。
- Manifest + 配置型插件安装：`POST /api/v1/admin/plugins/install` 可安装只包含声明、配置、权限、菜单、Hook 元信息和迁移计划的插件，初始状态为 installed + disabled，不执行第三方代码。
- 升级预览：
### 插件 dependencies 与 Core 兼容规则

`dependencies` 当前支持两种输入：旧兼容字符串数组（如 `["qa"]`，视为 required）和对象数组：

```json
{
  "code": "qa",
  "version": ">=1.0.0 <2.0.0",
  "required": true,
  "reason": "需要问答内容类型作为数据来源"
}
```

字段：`code` 为依赖插件编码；`version` 为版本约束；`required` 默认为 `true`；`reason` 为可选说明。当前版本约束只支持数字 `x.y.z`、精确版本、`>=`、`>`、`<=`、`<` 和空格连接的范围组合（如 `>=1.2.0 <2.0.0`），不支持 `^`、`~`、`||`、预发布标签或 npm 完整语法。

依赖状态：`satisfied`、`missing`、`disabled`、`archived`、`migration_failed`、`config_invalid`、`version_mismatch`、`circular_dependency`、`self_dependency`、`optional_missing`。required 不满足会阻断 validate / dry-run / install / upgrade dry-run / upgrade / enable；optional 缺失只 warning。自依赖和循环依赖当前一律阻断，包括 optional 循环。

Core 兼容：`min_core_version` 缺失为 warning；`min_core_version` 高于当前 Core 为 `plugin_core_version_incompatible` 并阻断；`compatible_core_version` 存在但当前 Core 不满足时按 incompatible 阻断。当前 Core 版本来自项目 `VERSION`，不是前端写死。

启用插件时会重新执行 required dependency 检查；依赖插件处于 disabled、archived、migration_failed、config_invalid、dependency_missing 时不能视为可用。错误码 / message 包含 `plugin_dependency_missing`、`plugin_dependency_disabled`、`plugin_dependency_archived`、`plugin_dependency_version_mismatch`、`plugin_dependency_cycle`、`plugin_core_version_incompatible` 等前缀或同义状态。

`POST /api/v1/admin/plugins/:code/upgrade/dry-run` 可返回现有插件与新 manifest 的版本兼容矩阵、变更字段和 diff，不执行真实升级；`POST /api/v1/admin/plugins/:code/upgrade` 已提供最小执行闭环，会更新插件 manifest / version / checksum 并保留历史数据。
- 软卸载 / 归档 / 恢复：单个归档 / 恢复和批量归档 / 恢复 API 已提供最小闭环；归档不删除历史内容、配置、迁移记录或审计记录。
- SDK / 模板：`docs/PLUGIN_SDK.md` 与 `docs/PLUGIN_TEMPLATE.md` 已说明 manifest 字段、生成器和安全边界；`go run ./cmd/devhub plugin:new ...` 只生成声明、配置、文档和示例，不新增 API，也不执行第三方代码。

部分完成：

- 插件权限矩阵：发布和菜单使用 manifest 权限码；完整权限分配 API、按 community / category 的权限配置 UI 和更细错误码仍未完成。
- HookBus：已有内部调度、调用点、`hook_executions` 执行记录、Hook 统计 API、失败审计和轻量健康摘要；尚未实现重试 API、告警系统或第三方动态 Hook。
- 插件迁移：当前 runner 只支持内置插件 up/no-op 执行记录、失败记录和失败重试；不支持 migration down、真实 rollback、迁移前备份或外部插件迁移包。
- 插件后台治理：已有列表、详情、配置、impact、审计、运行状态、迁移 Tab 和通用内容页；告警、重试策略和影响对象明细仍未完成。
- 插件安装：当前已支持 manifest + 配置型插件的 dry-run、安装记录和升级执行最小闭环，但不支持插件包文件上传、远程市场、前端资产动态加载或第三方本地代码执行。

预留 / 后续：

- `discovered`、`migrated`、`configured`、`running`、`config_invalid`、`migration_pending`、`dependency_missing` 已是 `plugins.status` 可接受值；完整外部插件安装器状态机、自动迁移、状态告警和完整分步向导仍需继续演进。
- 外部服务型 Webhook、插件包 zip dry-run、插件包签名、远程安装、市场、动态加载和沙箱均不是当前真实能力。

下一阶段 API / 验收目标：

- `v1.3.4` 的优先级是插件异常治理与平台基础能力收口，不新增插件市场、远程安装或动态加载 API。
- 插件迁移方向：补 failed migration 注入、启用阻断、retry 恢复和审计定位的 API / E2E；现有 `GET /api/v1/admin/plugins/:code/migrations`、`POST /api/v1/admin/plugins/:code/migrations/run`、`POST /api/v1/admin/plugins/:code/migrations/:name/retry` 需要覆盖失败与恢复场景。
- HookBus 方向：补 blocking / non-blocking Hook 失败注入、`hook_executions` 可见性、后台 Hooks Tab 断言和审计定位；不引入第三方 Hook、远程 Hook 或 Webhook。
- 权限矩阵：内容创建、后台创建、版主菜单均按 `ContentTypeDefinition.create_permission` 与插件权限码判断；API 测试已覆盖 `post.create` 只能桥接 `core.topic.create`，不能替代 `qa/docs/wiki/projects/jobs/ai_works` 的 create 权限。后续仍需补更细 RBAC 分配 UI 与插件内容治理操作矩阵。
- MySQLStore 方向：老库升级和 MySQLStore 下插件迁移、Hook 执行、审计、全局 / 子站启停、配置 schema 校验已完成专项验证；后续继续补生产大库备份 / 回滚演练和历史 SEO 的 MySQL 端完整浏览器矩阵。

## 当前真实 API 索引

认证：

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
POST /api/v1/admin/login
POST /api/v1/admin/refresh
POST /api/v1/admin/logout
GET  /api/v1/admin/me
```

子站与板块：

```http
GET /api/v1/communities
GET /api/v1/communities/:slug
GET /api/v1/communities/:slug/home
GET /api/v1/communities/:slug/overview
GET /api/v1/communities/:slug/stats
GET /api/v1/communities/:slug/categories
GET /api/v1/communities/:slug/tags
GET /api/v1/communities/:slug/moderators
```

Topic 与搜索：

```http
GET    /api/v1/topics
GET    /api/v1/topics/:id
GET    /api/v1/topics/:id/qa
GET    /api/v1/topics/:id/docs
GET    /api/v1/topics/:id/wiki/versions
POST   /api/v1/topics
PUT    /api/v1/topics/:id
DELETE /api/v1/topics/:id
GET    /api/v1/search/topics
```

评论与问答：

```http
GET  /api/v1/topics/:id/comments
POST /api/v1/topics/:id/comments
POST /api/v1/topics/:id/comments/:commentId/replies
POST /api/v1/topics/:id/comments/:commentId/accept
POST /api/v1/topics/:id/solve
```

插件扩展只读接口：

- `GET /api/v1/topics/:id/qa`
  - 仅适用于 `question`
  - 返回 `qa_questions` 扩展状态和 `qa_answers` 列表
- `GET /api/v1/topics/:id/docs`
  - 仅适用于 `document`
  - 返回 `docs_documents` 扩展行和当前文档空间的基础文档树
- `GET /api/v1/topics/:id/wiki/versions`
  - 仅适用于 `wiki_page`
  - 返回 `wiki_pages` 扩展行和 `wiki_page_versions` 列表

标签：

```http
GET    /api/v1/tags
GET    /api/v1/tags/hot
GET    /api/v1/tags/suggestions
GET    /api/v1/tags/suggest
GET    /api/v1/tags/by-slug/:tag
GET    /api/v1/tags/:tag
GET    /api/v1/tags/:tag/topics
GET    /api/v1/communities/:slug/tags/:tag
GET    /api/v1/communities/:slug/tags/:tag/topics
GET    /api/v1/admin/tags
GET    /api/v1/admin/tags/:id
GET    /api/v1/admin/tags/:id/topics
GET    /api/v1/admin/tags/:id/aliases
POST   /api/v1/admin/tags
PUT    /api/v1/admin/tags/:id
POST   /api/v1/admin/tags/:id/aliases
DELETE /api/v1/admin/tags/:id/aliases/:aliasId
POST   /api/v1/admin/tags/:id/enable
POST   /api/v1/admin/tags/:id/disable
POST   /api/v1/admin/tags/:id/merge
POST   /api/v1/admin/tags/:id/recalculate
POST   /api/v1/admin/tags/recalculate-all
```

互动与用户中心：

```http
POST /api/v1/topics/:id/like
POST /api/v1/topics/:id/favorite
GET  /api/v1/topics/:id/interaction
POST /api/v1/actions/toggle
POST /api/v1/reactions/toggle
POST /api/v1/favorites/toggle
POST /api/v1/follows/toggle
GET  /api/v1/me/favorites
GET  /api/v1/me/follows
GET  /api/v1/me/activities
GET  /api/v1/me/notifications
POST /api/v1/me/notifications/read-all
POST /api/v1/me/notifications/:id/read
```

举报、治理和审计：

```http
POST /api/v1/reports
GET  /api/v1/admin/reports
GET  /api/v1/admin/reports/:id
POST /api/v1/admin/reports/:id/handle
POST /api/v1/admin/reports/batch-handle
GET  /api/v1/admin/comments
POST /api/v1/admin/comments/batch
POST /api/v1/admin/topics/batch
GET  /api/v1/admin/audit-logs
GET  /api/v1/admin/plugins
GET  /api/v1/admin/plugins/health
GET  /api/v1/admin/plugins/:code/health
POST /api/v1/admin/plugins/manifest/validate
POST /api/v1/admin/plugins/dry-run
POST /api/v1/admin/plugins/install
POST /api/v1/admin/plugins/:code/upgrade/dry-run
POST /api/v1/admin/plugins/:code/upgrade
POST /api/v1/admin/plugins/:code/enable
POST /api/v1/admin/plugins/:code/disable
POST /api/v1/admin/plugins/:code/archive
POST /api/v1/admin/plugins/:code/restore
POST /api/v1/admin/plugins/bulk-archive
POST /api/v1/admin/plugins/bulk-restore
PUT  /api/v1/admin/plugins/:code/config
GET  /api/v1/admin/plugins/:code/impact
GET  /api/v1/admin/plugins/:code/hooks
GET  /api/v1/admin/plugins/:code/audit-logs
GET  /api/v1/admin/plugins/:code/migrations
POST /api/v1/admin/plugins/:code/migrations/run
POST /api/v1/admin/plugins/:code/migrations/:name/retry
POST /api/v1/admin/plugins/:code/migrations/:name/e2e-fail
GET  /api/v1/admin/communities/:id/plugins
POST /api/v1/admin/communities/:id/plugins/:code/enable
POST /api/v1/admin/communities/:id/plugins/:code/disable
PUT  /api/v1/admin/communities/:id/plugins/:code/config
PUT  /api/v1/admin/communities/:id/plugins/sort
GET  /api/v1/admin/communities/:id/plugins/:code/impact
GET  /api/v1/admin/plugin-menus
GET  /api/v1/moderator/plugin-menus
```

`GET /api/v1/admin/audit-logs` 返回的插件治理日志会包含结构化字段：

- `old_value`：操作前 JSON diff，例如旧状态或旧配置。
- `new_value`：操作后 JSON diff，例如新状态或新配置。
- `metadata_json`：操作上下文，例如 `plugin_code`、`community_id`、`operation`。

当前通用审计接口支持的筛选参数包括：

- `type`
- `actor_type`
- `actor`
- `actor_user_id`
- `action`
- `target`
- `target_type`
- `target_id`
- `plugin_code`
- `community_id`
- `metadata`
- `request_id`
- `start_time`
- `end_time`
- `page`
- `page_size`

说明：非插件历史日志可能仍只有 `target` 文本摘要。

`POST /api/v1/admin/topics/batch`

- 认证：后台 admin token，或明确允许的子站版主身份。
- 权限：`topic.moderate`，并由后端继续校验当前操作者是否可治理每条内容所属子站。
- 用途：后台通用内容治理与 PluginContent 通用插件内容页的批量治理入口。
- 请求：

```json
{
  "ids": [1, 2],
  "action": "hide",
  "note": "PluginContent hide qa"
}
```

- 当前支持动作：`feature`、`unfeature`、`pin`、`unpin`、`hide`、`restore`、`approve`、`reject`、`lock-comments`、`unlock-comments`、`delete`。
- PluginContent UI 当前接入 `hide` / `restore`、`approve` / `reject`、`pin` / `unpin`、`feature` / `unfeature`。
- 审计：对插件内容会写入带 `plugin_code` / `content_type` / `community_id` / `category_id` / `content_id` / `operation` 的插件内容治理审计；成功和失败单项都会保留结构化插件审计，同时保留批量治理摘要审计。
- 常见错误：单条内容无权治理时，该条返回 `error`，不会伪造成成功。

版主工作台：

```http
GET  /api/v1/moderator/communities
GET  /api/v1/moderator/plugin-menus
GET  /api/v1/moderator/dashboard
GET  /api/v1/moderator/reports
POST /api/v1/moderator/reports/:id/handle
GET  /api/v1/moderator/topics
POST /api/v1/moderator/topics/:id/feature
POST /api/v1/moderator/topics/:id/unfeature
POST /api/v1/moderator/topics/:id/pin
POST /api/v1/moderator/topics/:id/unpin
POST /api/v1/moderator/topics/:id/hide
POST /api/v1/moderator/topics/:id/restore
POST /api/v1/moderator/topics/:id/lock-comments
POST /api/v1/moderator/topics/:id/unlock-comments
GET  /api/v1/moderator/comments
POST /api/v1/moderator/comments/:id/hide
POST /api/v1/moderator/comments/:id/restore
GET  /api/v1/moderator/audit-logs
```

兼容 API：

```http
GET    /api/v1/sites
GET    /api/v1/sites/:site
GET    /api/v1/sites/:site/overview
GET    /api/v1/boards
GET    /api/v1/posts
GET    /api/v1/posts/:id
```

说明：

- `POST/PUT/DELETE /api/v1/posts*` 写接口已废弃，当前返回 `410 Gone`，请使用 `/api/v1/topics`。
- `Service.CreatePost` 已在业务层封口，不再直接调用仓储裸写入；`repo.CreatePost` 仅作为 legacy / seed / migration 或兼容层内部能力保留，不作为业务写入口。

SEO 端点：

```http
GET /topics/:id/
GET /c/:slug/
GET /tags/:tag/
GET /c/:slug/tags/:tag/
GET /sitemap.xml
GET /robots.txt
```

## MySQLStore / 老库升级验收说明

本轮没有新增专用生产 API；MySQLStore 专项通过可选集成测试和 SQL 升级脚本验证现有插件 API 的真实后端能力。

已验证范围：

- 新装 schema 包含 `plugins`、`community_plugins`、`plugin_migrations`、`hook_executions`、结构化 `admin_logs`。
- 老库升级字段包含 `topics.plugin_code`、`categories.plugin_code`、`categories.allowed_content_types`。
- MySQLStore 下全局禁用插件会阻断对应 `content_type` 创建。
- MySQLStore 下子站禁用插件只阻断该子站创建，其他子站不受影响。
- failed migration 会阻断全局启用和子站启用，retry 成功后可恢复。
- Hook 执行记录可查询，插件治理审计可查询。
- 全局 / 子站插件配置保存都会执行后端 `config_schema` 校验。

建议验证命令：

```bash
docker compose -f docker-compose.dev.yml up -d mysql
docker compose -f docker-compose.dev.yml exec -T mysql mysql -uroot -pDevhub_root_123456 -e "DROP DATABASE IF EXISTS devhub_test; CREATE DATABASE devhub_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL PRIVILEGES ON devhub_test.* TO 'devhub'@'%'; FLUSH PRIVILEGES;"
DEVHUB_MYSQL_TESTS=1 DB_HOST=127.0.0.1 DB_PORT=3307 DB_USER=devhub DB_PASSWORD=Devhub_123456 DB_NAME=devhub_test go test ./internal/service -run TestMySQLStorePluginPlatformConsistency -count=1 -v
```

## 规划 / 未完成

以下内容不是当前真实可用 API：

- P0：插件内容治理操作矩阵、完整 RBAC 分配 UI、community / category 级权限配置和更细错误码。
- P0/P1：`config_schema` 结构化错误响应、更完整 JSON Schema 能力、深层 diff 和配置版本 API。
- P0/P1：HookBus 告警、失败重试、更多业务处理器、插件搜索 / 通知 / SEO 扩展 API。
- P1：插件生成模板、插件依赖检查 UI、插件版本兼容矩阵 UI 和更完整开发者体验；插件 SDK / 开发规范文档已建立在 `docs/plugins/`。
- P2：本地插件包 zip、插件包安装向导、插件升级向导增强、外部服务型 Webhook、插件 migration runner、插件包签名校验和插件市场雏形。
- P3：远程插件市场、在线更新、动态加载能力评估、插件沙箱和插件权限隔离。
- 插件权限配置 API、按子站 / 板块维护细粒度权限矩阵，以及 Core 兼容类型 `article/news` 的长期权限收口。
- Docs 文档树专用编辑 API 的完整形态。
- Wiki 版本历史、版本对比和回滚 API 的完整形态。
- 取消最佳答案 / 取消已解决状态接口。
- 标签趋势统计和标签运营分析 API。

## v1.6.0-P0-04 远程插件索引只读镜像 API

远程插件索引只读取 `index.json` 元数据；不会下载 `package_url`、不会安装远程插件、不会执行代码、不会动态加载前端资产，也不会自动信任远程 publisher。

### 索引源

```http
GET    /api/v1/admin/plugins/remote-indexes
POST   /api/v1/admin/plugins/remote-indexes
PUT    /api/v1/admin/plugins/remote-indexes/:id
POST   /api/v1/admin/plugins/remote-indexes/:id/enable
POST   /api/v1/admin/plugins/remote-indexes/:id/disable
DELETE /api/v1/admin/plugins/remote-indexes/:id
POST   /api/v1/admin/plugins/remote-indexes/:id/fetch
```

查看需要 `plugin.read`；新增、更新、启用、禁用、删除和拉取需要 `plugin.manage`。`index_url` 仅支持 http / https，生产建议 https；`file://`、localhost、127.0.0.1 和内网地址会被 SSRF 防护拦截。

索引源字段：`id`、`source_id`、`name`、`index_url`、`homepage`、`description`、`status`、`trust_policy`、`last_fetch_status`、`last_fetch_at`、`last_error_code`、`last_error_message`、`last_index_hash`。

### 远程插件列表 / 详情

```http
GET /api/v1/admin/plugins/remote-indexes/:id/plugins
GET /api/v1/admin/plugins/remote-indexes/:id/plugins/:code
```

列表字段包括：`code`、`name`、`latest_version`、`publisher_id`、`public_key_id`、`license`、`min_core_version`、`compatible_core_version`、`package_sha256`、`signature_url`、`installed`、`local_version`、`version_status`、`core_compatibility`、`publisher_trust_status`、`risk_level`、`risk_summary`。

详情会展示所有版本的 `package_url`、`package_sha256`、`manifest_sha256`、`signature_url`、publisher、trust status、Core 兼容性和只读限制提示。`package_sha256` 只是远程元数据声明，本轮不会下载包内容进行校验。

新增错误码：`plugin_remote_index_not_found`、`plugin_remote_index_invalid_url`、`plugin_remote_index_forbidden_url`、`plugin_remote_index_fetch_failed`、`plugin_remote_index_fetch_timeout`、`plugin_remote_index_response_too_large`、`plugin_remote_index_invalid_json`、`plugin_remote_index_schema_invalid`、`plugin_remote_index_disabled`、`plugin_remote_index_plugin_not_found`、`plugin_remote_index_package_url_invalid`、`plugin_remote_index_publisher_unknown`、`plugin_remote_index_publisher_blocked`、`plugin_remote_index_core_incompatible`、`plugin_remote_index_readonly`。

## v1.6.0-P0-05 插件包版本仓库与升级差异 API

版本仓库是只读聚合视图，不会下载远程包、不会自动升级、不会执行第三方代码 / SQL / 前端资产。

### `GET /api/v1/admin/plugins/versions`

权限：`plugin.read`。

查询参数：

- `plugin_code`：可选，按插件 code 过滤。
- `source`：可选，`installed` / `local_package` / `uploaded_package` / `remote_index` / `exported_package`。
- `status`：可选，按版本状态过滤。
- `keyword`：可选，按 code / name / version / path / publisher 搜索。
- `page` / `page_size`：分页。

返回：

```json
{
  "items": [
    {
      "plugin_code": "demo_notice",
      "plugin_name": "Demo Notice",
      "installed_version": "1.0.0",
      "latest_local_version": "1.1.0",
      "latest_remote_version": "1.2.0",
      "update_available": true,
      "sources": ["installed", "local_package", "remote_index"],
      "versions": []
    }
  ],
  "pagination": {"page": 1, "page_size": 20, "total": 1},
  "summary": {"total": 1, "installed": 1, "local_packages": 1, "uploaded_packages": 0, "remote_index": 1, "update_available": 1, "readonly": 1}
}
```

### `GET /api/v1/admin/plugins/:code/versions`

权限：`plugin.read`。返回单个插件的版本列表，包含 installed、本地仓库包、上传包、远程索引版本。远程索引版本带 `readonly=true`，只能展示，不能直接升级。

### `GET /api/v1/admin/plugins/:code/versions/:version`

权限：`plugin.read`。查询参数可带 `source`、`package_path`、`remote_index_id` 精确定位同版本多来源记录。

### `POST /api/v1/admin/plugins/:code/versions/:version/upgrade-diff`

权限：`plugin.write` 或更高。请求体：

```json
{
  "source": "local_package",
  "package_path": "storage/plugins/packages/demo_notice-1.1.0"
}
```

返回结构化升级差异：

```json
{
  "plugin_code": "demo_notice",
  "current_version": "1.0.0",
  "target_version": "1.1.0",
  "source": "local_package",
  "status": "ok|warning|blocked",
  "summary": {"added": 3, "removed": 1, "changed": 5, "high_risk": 2, "blocked": 0},
  "diff_sections": [
    {
      "section": "permissions",
      "title": "权限变化",
      "risk_level": "high",
      "items": [
        {"section": "permissions", "path": "permissions.demo_notice.manage", "type": "added", "risk_level": "high", "message": "新增高危插件权限"}
      ]
    }
  ],
  "risk_report": {"level": "high", "score": 70, "summary": "升级差异包含高风险变更"},
  "compatibility": {"core_version": "v1.6.0", "status": "compatible"},
  "dependencies": {"total": 0, "blocking": 0}
}
```

差异范围包括基础信息、content_types、content_type_definitions、permissions、menus、routes、config_schema、default_config、dependencies、hooks、migrations 与 package metadata。敏感路径（password / secret / token / key / credential / app_secret / aes_key 等）返回 `[REDACTED]`。

高风险规则：删除 content_type / permission、 新增高危权限、新增 required dependency、依赖约束收紧、config_schema 删除字段 / type 变化 / required 新增、Hook 从 non-blocking 改为 blocking、新增 migration、Core 约束提高或收窄、签名从 trusted verified 退化为 unknown / unsigned。

阻断规则：目标版本小于等于当前版本、Core 不兼容、required dependency 缺失、checksum mismatch、签名验签失败、publisher blocked / revoked、manifest invalid、dangerous file、package risk blocked、远程索引版本直接升级。

新增错误码：`plugin_version_not_found`、`plugin_version_source_invalid`、`plugin_version_compare_failed`、`plugin_version_downgrade_forbidden`、`plugin_version_same_version`、`plugin_version_package_missing`、`plugin_version_remote_readonly`、`plugin_upgrade_diff_failed`、`plugin_upgrade_diff_blocked`、`plugin_upgrade_diff_high_risk`、`plugin_upgrade_diff_sensitive_redacted`、`plugin_upgrade_target_core_incompatible`、`plugin_upgrade_target_dependency_missing`、`plugin_upgrade_target_signature_invalid`、`plugin_upgrade_target_publisher_blocked`。
