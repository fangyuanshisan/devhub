import { isSensitiveKey } from '@/i18n/formatters';

export function normalizeSchema(schema) {
  if (!schema || typeof schema !== 'object') return null;
  return schema;
}

export function normalizeType(def) {
  const field = def && typeof def === 'object' ? def : {};
  const raw = Array.isArray(field.type) ? field.type.find((item) => item !== 'null') : field.type;
  if (field.enum) return raw || 'string';
  if (raw === 'integer' || raw === 'number' || raw === 'boolean' || raw === 'array' || raw === 'object' || raw === 'string') return raw;
  return 'string';
}

export function getEnumOptions(def) {
  const field = def && typeof def === 'object' ? def : {};
  if (Array.isArray(field.enum)) {
    const names = pickEnumNames(field);
    return field.enum.map((value, idx) => ({ value, label: names?.[idx] ?? String(value) }));
  }
  // Support oneOf/anyOf title mapping.
  const variants = Array.isArray(field.oneOf) ? field.oneOf : Array.isArray(field.anyOf) ? field.anyOf : null;
  if (variants && variants.length) {
    const options = [];
    for (const item of variants) {
      if (!item || typeof item !== 'object') continue;
      const value = item.const ?? (Array.isArray(item.enum) ? item.enum[0] : undefined);
      if (value === undefined) continue;
      options.push({ value, label: item.title || String(value) });
    }
    return options.length ? options : null;
  }
  return null;
}

function pickEnumNames(field) {
  if (!field || typeof field !== 'object') return null;
  if (Array.isArray(field.enumNames)) return field.enumNames;
  if (Array.isArray(field['x-enumNames'])) return field['x-enumNames'];
  if (field['x-enumNames'] && typeof field['x-enumNames'] === 'object') return null;
  return null;
}

export function isSensitiveField(key, def) {
  const field = def && typeof def === 'object' ? def : {};
  const format = String(field.format || '').toLowerCase();
  if (field['x-sensitive'] === true) return true;
  if (field.writeOnly === true) return true;
  if (format === 'password') return true;
  if (isSensitiveKey(key)) return true;
  return false;
}

export function getFieldGroup(def) {
  const field = def && typeof def === 'object' ? def : {};
  return field['x-group'] || field.group || field['ui:group'] || '';
}

export function getWidget(def) {
  const field = def && typeof def === 'object' ? def : {};
  return field['ui:widget'] || field['x-widget'] || field.widget || '';
}

export function isMultiline(def) {
  const field = def && typeof def === 'object' ? def : {};
  const widget = String(getWidget(field) || '').toLowerCase();
  if (widget === 'textarea' || widget === 'multiline') return true;
  if (field.multiline === true) return true;
  return false;
}

export function asNumberOrNull(value) {
  if (value === '' || value == null) return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
}

