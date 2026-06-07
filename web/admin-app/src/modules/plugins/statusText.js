const STATUS_TEXT = {
  all: '全部',
  unknown: '未知',
  enabled: '已启用',
  disabled: '已禁用',
  archived: '已归档',
  installed: '已安装',
  discovered: '已发现',
  migrated: '已迁移',
  configured: '已配置',
  running: '运行中',
  config_invalid: '配置异常',
  migration_pending: '待迁移',
  migration_failed: '迁移失败',
  dependency_missing: '依赖缺失',
  healthy: '正常',
  warning: '警告',
  error: '异常',
  blocked: '已阻断',
  staged: '待确认',
  uploaded: '已上传',
  scanned: '已扫描',
  promoted: '已转入本地仓库',
  approval_pending: '待审批',
  install_approval_pending: '待安装审批',
  approval_rejected: '审批已拒绝',
  approved: '已批准',
  rejected: '已拒绝',
  canceled: '已取消',
  deleted: '已删除',
  failed: '失败',
  expired: '已过期',
  pending: '待处理',
  created: '已创建',
  success: '成功',
  ok: '正常',
  valid: '有效',
  invalid: '无效',
  skipped: '已跳过',
  revoked: '已吊销',
  active: '启用中',
  inactive: '未启用',
  closed: '正常',
  open: '已熔断',
  half_open: '半开探测',
  circuit_open: '熔断中',
  retry_scheduled: '等待重试',
  retry_exhausted: '重试耗尽',
  sending: '发送中',
  delivering: '投递中',
  delivered: '已投递',
  recovered: '已恢复',
  previous: '上一版本',
  applied: '已应用',
  passed: '通过',
  preview: '预检查',
  downloaded: '已下载',
  upgraded: '已升级',
  installed_pending_enable: '已安装，待启用',
  not_installed: '未安装',
  accepted: '已接受',
  trusted: '可信',
  official: '官方',
  enterprise: '企业私有',
  local_dev: '本地开发',
  unsigned: '未签名',
  verified: '验签通过',
  unverified: '未验签',
  unsupported: '不支持',
  missing: '缺失',
  mismatch: '不匹配',
  publisher_unknown: '发布者未受信任',
  metadata_only: '仅元数据',
  compatible: '兼容',
  incompatible: '不兼容',
  optional_missing: '可选依赖缺失',
  version_mismatch: '版本不匹配',
  needs_reencrypt: '需要重新加密',
  already_current: '已是当前密钥',
  legacy_v1: '旧版密文',
  decrypt_failed: '解密失败',
  missing_key: '缺少密钥',
  pass: '通过',
};

const RISK_TEXT = {
  low: '低风险',
  medium: '中风险',
  high: '高风险',
  blocked: '已阻断',
  risk_blocked: '插件包风险过高，已阻断',
};

const REASON_TEXT = {
  risk_blocked: '插件包风险过高，已阻断',
  manifest_invalid: '插件清单校验失败',
  plugin_package_manifest_invalid: '插件清单校验失败',
  checksum_failed: '文件完整性校验失败',
  plugin_package_checksum_failed: '文件完整性校验失败',
  dangerous_file: '插件包包含危险文件',
  plugin_package_dangerous_file: '插件包包含危险文件',
  dependency_missing: '插件依赖缺失',
  signature_invalid: '插件签名无效',
  plugin_package_signature_invalid: '插件签名无效',
  publisher_unknown: '发布者未受信任',
  publisher_blocked: '发布者已被阻断',
  publisher_revoked: '发布者已被撤销信任',
  core_incompatible: '当前 DevHub 版本不兼容',
  content_type_conflict: '内容类型冲突',
  permission_conflict: '权限声明冲突',
  package_not_found: '插件包不存在',
  plugin_package_not_found: '插件包不存在',
  upload_not_found: '上传记录不存在',
  plugin_package_upload_not_found: '上传记录不存在',
  promote_failed: '转入本地仓库失败',
  plugin_package_promote_failed: '转入本地仓库失败',
  install_failed: '安装失败',
  plugin_package_install_failed: '安装失败',
  plugin_package_promote_blocked: '当前上传包已被阻断，不能转入本地仓库',
  plugin_package_promote_target_exists: '本地仓库目标目录已存在',
  plugin_package_install_source_invalid: '该包尚未转入本地仓库，不能安装',
  plugin_package_install_dry_run_required: '请先执行安装预检查',
  plugin_package_install_dry_run_expired: '安装预检查已过期，请重新执行',
  plugin_package_install_dry_run_mismatch: '安装预检查计划与当前插件包不一致',
  plugin_package_install_dry_run_invalid: '安装预检查计划不可用，请重新执行',
  plugin_package_install_blocked: '插件包存在阻断项，不能安装',
  plugin_package_already_installed: '该插件已安装，无需重复安装。若要修改 Webhook 地址，请进入 external_service 配置；若要更新插件包，请使用版本仓库中的升级差异与审批流程。',
  plugin_package_core_incompatible: '当前 DevHub 版本不兼容',
  plugin_package_dependency_missing: '插件依赖缺失',
  plugin_package_zip_manifest_missing: '插件包缺少 manifest.json',
  plugin_package_zip_multiple_manifests: '插件包包含多个 manifest.json',
  plugin_package_migrations_invalid: '迁移文件不符合规范',
  plugin_package_upload_invalid_type: '只允许上传 .zip 插件包',
  plugin_package_upload_too_large: '上传 zip 超过大小限制',
  plugin_package_upload_failed: '上传插件包失败',
  plugin_package_upload_blocked: '上传包预检已阻断',
  plugin_package_upload_invalid_status: '当前上传包状态不允许执行该操作',
  plugin_package_upload_action_not_allowed: '当前状态不允许执行该操作',
  plugin_package_upload_approval_blocked: '当前状态不能提交导入审批',
  plugin_package_upload_lifecycle_invalid: '上传包审批流程状态异常',
  plugin_package_upload_already_deleted: '上传包已删除',
  plugin_package_upload_status_conflict: '当前上传包状态冲突',
  plugin_package_upload_delete_not_allowed: '当前上传包状态不允许删除',
  plugin_package_cleanup_confirm_required: '请先执行清理 dry-run',
  plugin_package_cleanup_confirm_invalid: '清理确认 token 无效',
  plugin_package_delete_path_forbidden: '删除路径不在允许目录内',
  plugin_package_repository_delete_not_allowed: '本地仓库包不允许删除',
  plugin_package_repository_installed: '已安装包不能直接删除',
  plugin_package_repository_not_promoted: '仅本地仓库未安装包可删除',
  plugin_package_repository_path_forbidden: '本地仓库路径不允许删除',
  plugin_package_precheck_not_found: '插件包预检记录不存在',
  plugin_package_signature_precheck_not_passed: '只有预检通过的插件包才能验签',
  plugin_package_signature_source_missing: '预检记录缺少插件包来源，无法验签',
  plugin_package_signature_source_invalid: '插件包来源不符合验签要求',
  plugin_package_signature_failed: '插件包验签失败',
  plugin_package_signature_invalid_request: '验签请求参数无效',
  plugin_package_signature_not_found: '验签记录不存在',
  plugin_package_download_invalid_request: '插件包下载请求无效',
  plugin_package_download_checksum_mismatch: '远程插件包校验和不一致',
  plugin_package_download_checksum_invalid: '远程插件包校验和格式不合法',
  plugin_package_download_url_invalid: '插件包下载地址不合法',
  plugin_package_download_url_forbidden: '插件包下载地址被安全策略拒绝',
  plugin_package_download_type_unsupported: '插件包格式不受支持',
  plugin_package_download_redirect_blocked: '远程插件包重定向次数超过限制',
  plugin_package_download_too_large: '远程插件包超过大小限制',
  plugin_package_download_failed: '远程插件包下载失败',
  plugin_package_staging_not_found: '暂存插件包不存在',
  plugin_package_staging_delete_failed: '暂存插件包删除失败',
  plugin_remote_index_disabled: '远程索引已禁用',
  plugin_remote_index_invalid_json: '远程索引 JSON 格式无效',
  plugin_remote_index_response_too_large: '远程索引响应超过大小限制',
  plugin_config_key_missing: '插件配置密钥缺失',
  token_missing: '缺少 Token',
  token_invalid: 'Token 不合法',
  token_disabled: 'Token 已禁用',
  token_revoked: 'Token 已吊销',
  token_expired: 'Token 已过期',
  scope_denied: '权限范围不足',
  community_scope_denied: '子站权限范围不匹配',
  plugin_disabled: '插件未启用',
  hook_warning: 'Hook 有警告',
  hook_error: 'Hook 异常',
  service_unreachable: '外部服务不可达',
  timeout: '请求超时',
  unauthorized: '认证失败',
  config_missing: '配置缺失',
  config_schema_error: '配置格式不符合要求',
  migration_required: '需要执行迁移',
  migration_error: '迁移异常',
  dependency_error: '依赖异常',
};

const ACTION_TEXT = {
  enable: '启用',
  disable: '禁用',
  archive: '归档',
  restore: '恢复',
  promote: '转入本地仓库',
  install: '安装',
  dry_run: '预检查',
  'dry-run': '预检查',
  validate: '校验',
  rescan: '重新扫描',
  retry: '重试',
  cancel: '取消',
  delete: '删除',
  detail: '查看详情',
  view_details: '查看详情',
  view_audit: '查看审计',
  view_impact: '查看影响',
  health_check: '健康检查',
};

const SUGGESTION_TEXT = {
  checksum_failed: '请重新生成 checksums.json 后再上传。',
  manifest_invalid: '请修正 manifest.json 后重新上传或重新预检查。',
  dangerous_file: '请移除可执行文件、脚本或路径穿越文件后重新打包。',
  publisher_unknown: '请先在发布者与信任中导入并启用该发布者。',
  publisher_revoked: '请更换可信发布者或重新签名。',
  core_incompatible: '请升级 DevHub 或选择兼容当前版本的插件包。',
  plugin_package_install_dry_run_required: '请先对本地仓库包执行安装预检查。',
  plugin_package_install_dry_run_expired: '请重新执行安装预检查并确认最新计划。',
};

export function pluginText(code, fallback = '-') {
  const key = normalizeKey(code);
  return STATUS_TEXT[key] || RISK_TEXT[key] || REASON_TEXT[key] || ACTION_TEXT[key] || fallbackText(code, fallback);
}

export function pluginStatusText(status, fallback = '-') {
  const key = normalizeKey(status);
  return STATUS_TEXT[key] || fallbackText(status, fallback);
}

export function pluginRiskText(level, fallback = '-') {
  const key = normalizeKey(level);
  return RISK_TEXT[key] || pluginStatusText(key, fallback);
}

export function pluginReasonText(reason, fallback = '-') {
  const key = normalizeKey(reason);
  return REASON_TEXT[key] || pluginStatusText(key, fallback);
}

export function pluginActionText(action, fallback = '-') {
  const key = normalizeKey(action);
  return ACTION_TEXT[key] || fallbackText(action, fallback);
}

export function pluginSuggestionText(code, fallback = '') {
  const key = normalizeKey(code);
  return SUGGESTION_TEXT[key] || fallback;
}

export function pluginMessageText(input, fallback = '操作失败，请查看错误详情') {
  if (!input) return fallback;
  if (typeof input === 'string') return pluginReasonText(input, input);
  const code = input.code || input.error_code || input.reason_code || input.blocked_code;
  const message = input.message || input.error || input.reason || input.error_message;
  const suggestion = input.suggestion || input.details?.suggestion || pluginSuggestionText(code);
  const main = message && message !== code ? message : pluginReasonText(code, message || fallback);
  return [main, suggestion ? `建议：${suggestion}` : ''].filter(Boolean).join(' ');
}

export function pluginTagType(status) {
  const key = normalizeKey(status);
  if (['enabled', 'ok', 'valid', 'healthy', 'success', 'staged', 'approved', 'promoted', 'applied', 'verified', 'trusted', 'closed', 'accepted', 'compatible', 'active', 'passed'].includes(key)) return 'success';
  if (['warning', 'pending', 'running', 'approval_pending', 'install_approval_pending', 'retry_scheduled', 'half_open', 'previous', 'medium', 'high', 'needs_reencrypt', 'publisher_unknown', 'unsigned', 'missing', 'scanned', 'uploaded'].includes(key)) return 'warning';
  if (['disabled', 'archived', 'canceled', 'deleted', 'skipped', 'expired', 'unknown', 'discovered', 'inactive'].includes(key)) return 'info';
  if (['blocked', 'failed', 'error', 'invalid', 'config_invalid', 'migration_failed', 'dependency_missing', 'hook_error', 'revoked', 'mismatch', 'retry_exhausted', 'circuit_open', 'open', 'rejected', 'incompatible', 'publisher_blocked', 'publisher_revoked'].includes(key)) return 'danger';
  return 'info';
}

function normalizeKey(value) {
  return String(value || '').trim().toLowerCase().replace(/[\s.-]+/g, '_');
}

function fallbackText(value, fallback) {
  const raw = String(value || '').trim();
  return raw || fallback;
}
