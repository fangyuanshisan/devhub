import { t } from './index';

export function textOrDash(value) {
  return value == null || value === '' ? '-' : String(value);
}

export function pluginStatusLabel(status) {
  const key = String(status || 'unknown');
  return STATUS_LABELS[key] || t(`plugin.${camelKey(key)}`) || key;
}

export function pluginHealthLabel(status) {
  const key = String(status || 'unknown');
  return STATUS_LABELS[key] || t(`plugin.healthText.${key}`) || key;
}

export function capabilityLabel(key) {
  return t(`plugin.capability.${camelKey(key)}`) || key;
}

export function fieldLabel(key) {
  return t(`field.${key}`) || key;
}

export function auditActionLabel(action) {
  return t(`plugin.audit.action.${action}`) || action || '-';
}

export function contentStatusLabel(status) {
  const key = String(status || '');
  return STATUS_LABELS[key] || t(`plugin.content.status.${key}`) || key || '-';
}

export function maturityLabel(plugin) {
  if (!plugin) return '-';
  if (plugin.code === 'qa' || plugin.code === 'docs' || plugin.code === 'wiki') return '平台治理已接入';
  return '业务闭环待完善';
}

export function statusTagType(status) {
  if (status === 'enabled' || status === 'success' || status === 'ok' || status === 'valid' || status === 'healthy') return 'success';
  if (status === 'disabled' || status === 'archived') return 'info';
  if (status === 'warning' || status === 'pending' || status === 'running' || status === 'migration_pending' || status === 'hook_warning') return 'warning';
  if (status === 'failed' || status === 'invalid' || status === 'error' || status === 'migration_failed' || status === 'config_invalid' || status === 'dependency_missing' || status === 'hook_error') return 'danger';
  return 'info';
}

export function migrationStatusLabel(status) {
  return pluginHealthLabel(status || 'pending');
}

export function genericStatusLabel(status) {
  const key = String(status || '').toLowerCase();
  return STATUS_LABELS[key] || status || '-';
}

export function packageRiskLabel(level) {
  const key = String(level || '').toLowerCase();
  return RISK_LABELS[key] || level || '-';
}

export function trustLevelLabel(level) {
  const key = String(level || '').toLowerCase();
  return TRUST_LABELS[key] || genericStatusLabel(level);
}

export function maskSensitiveConfig(value) {
  if (Array.isArray(value)) return value.map((item) => maskSensitiveConfig(item));
  if (!value || typeof value !== 'object') return value;
  return Object.fromEntries(
    Object.entries(value).map(([key, val]) => {
      if (isSensitiveKey(key)) return [key, val == null || val === '' ? val : '******'];
      return [key, maskSensitiveConfig(val)];
    }),
  );
}

export function isSensitiveKey(key) {
  return /(token|password|secret|key|credential)/i.test(String(key || ''));
}

export function topLevelChangedKeys(oldConfig = {}, newConfig = {}) {
  const keys = new Set([...Object.keys(oldConfig || {}), ...Object.keys(newConfig || {})]);
  return Array.from(keys).filter((key) => JSON.stringify(oldConfig?.[key]) !== JSON.stringify(newConfig?.[key]));
}

export function changedPaths(oldConfig = {}, newConfig = {}, options = {}) {
  const maxDepth = Number.isFinite(options?.maxDepth) ? options.maxDepth : 4;
  const out = [];
  walkDiff('$', oldConfig ?? {}, newConfig ?? {}, 0, maxDepth, out);
  return out.filter((p) => p !== '$');
}

function walkDiff(path, oldVal, newVal, depth, maxDepth, out) {
  if (JSON.stringify(oldVal) === JSON.stringify(newVal)) return;
  if (depth >= maxDepth) {
    out.push(path);
    return;
  }
  const oldIsObj = oldVal && typeof oldVal === 'object';
  const newIsObj = newVal && typeof newVal === 'object';
  const oldIsArr = Array.isArray(oldVal);
  const newIsArr = Array.isArray(newVal);
  if (oldIsArr || newIsArr) {
    out.push(path);
    return;
  }
  if (!oldIsObj || !newIsObj) {
    out.push(path);
    return;
  }
  const keys = new Set([...Object.keys(oldVal || {}), ...Object.keys(newVal || {})]);
  if (keys.size === 0) {
    out.push(path);
    return;
  }
  for (const key of Array.from(keys).sort()) {
    walkDiff(`${path}.${key}`, oldVal?.[key], newVal?.[key], depth + 1, maxDepth, out);
  }
}

export function safeJSON(value) {
  if (!value) return {};
  if (typeof value === 'string') {
    try {
      return JSON.parse(value);
    } catch {
      return {};
    }
  }
  return typeof value === 'object' ? value : {};
}

function camelKey(key) {
  return String(key || '').replace(/_([a-z])/g, (_, ch) => ch.toUpperCase());
}

const STATUS_LABELS = {
  unknown: '未知',
  all: '全部',
  enabled: '已启用',
  disabled: '已停用',
  installed: '已安装',
  soft_uninstalled: '已软卸载',
  archived: '已归档',
  migration_pending: '待迁移',
  migration_failed: '迁移失败',
  rollback_failed: '回滚失败',
  config_invalid: '配置无效',
  dependency_missing: '依赖缺失',
  hook_warning: 'Hook 警告',
  hook_error: 'Hook 异常',
  running: '运行中',
  pending: '待处理',
  created: '已创建',
  delivering: '投递中',
  delivered: '已投递',
  sending: '发送中',
  success: '成功',
  failed: '失败',
  warning: '有警告',
  ok: '正常',
  valid: '有效',
  invalid: '无效',
  blocked: '已阻断',
  skipped: '已跳过',
  retry_scheduled: '等待重试',
  retry_exhausted: '重试耗尽',
  circuit_open: '熔断中',
  closed: '正常',
  open: '已熔断',
  half_open: '半开探测',
  active: '启用中',
  previous: '上一版本',
  revoked: '已吊销',
  expired: '已过期',
  trusted: '可信',
  official: '官方',
  enterprise: '企业私有',
  local_dev: '本地开发',
  uploaded: '已上传',
  scanned: '已扫描',
  staged: '已暂存',
  downloaded: '已下载',
  approved: '已批准',
  promoted: '已转入仓库',
  upgraded: '已升级',
  preview: '预检',
  approval_pending: '等待审批',
  approval_rejected: '审批拒绝',
  install_approval_pending: '等待安装审批',
  installed_pending_enable: '已安装待启用',
  canceled: '已取消',
  deleted: '已删除',
  applied: '已应用',
  verified: '验签通过',
  missing: '缺失',
  mismatch: '不匹配',
  unsigned: '未签名',
  unverified: '未验签',
  unsupported: '不支持',
  publisher_unknown: '发布者未知',
  accepted: '已接受',
  rejected: '已拒绝',
  executed: '已执行',
  passed: '通过',
  recovered: '已恢复',
  incompatible: '不兼容',
  compatible: '兼容',
  optional_missing: '可选依赖缺失',
  version_mismatch: '版本不匹配',
};

const RISK_LABELS = {
  low: '低风险',
  medium: '中风险',
  high: '高风险',
  blocked: '已阻断',
};

const TRUST_LABELS = {
  official: '官方',
  trusted: '可信',
  enterprise: '企业私有',
  local_dev: '本地开发',
  blocked: '已禁用',
  revoked: '已吊销',
  expired: '已过期',
};
