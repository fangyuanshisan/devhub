import { t } from './index';

export function textOrDash(value) {
  return value == null || value === '' ? '-' : String(value);
}

export function pluginStatusLabel(status) {
  const key = String(status || 'unknown');
  return t(`plugin.${camelKey(key)}`) || key;
}

export function pluginHealthLabel(status) {
  const key = String(status || 'unknown');
  return t(`plugin.healthText.${key}`) || key;
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
  return t(`plugin.content.status.${key}`) || key || '-';
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
  return /(token|password|secret|key)/i.test(String(key || ''));
}

export function topLevelChangedKeys(oldConfig = {}, newConfig = {}) {
  const keys = new Set([...Object.keys(oldConfig || {}), ...Object.keys(newConfig || {})]);
  return Array.from(keys).filter((key) => JSON.stringify(oldConfig?.[key]) !== JSON.stringify(newConfig?.[key]));
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
