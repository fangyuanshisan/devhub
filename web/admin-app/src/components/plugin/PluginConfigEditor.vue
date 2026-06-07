<template>
  <div class="config-editor" data-testid="plugin-json-editor">
    <div data-testid="plugin-config-editor" style="display:none"></div>
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <slot name="title" />
        <el-tag v-if="schemaErrors.length" data-testid="schema-error-badge" type="danger" effect="light">{{ t('plugin.config.schemaFailed') }}</el-tag>
        <el-tag v-else data-testid="schema-success-badge" type="success" effect="light">{{ t('plugin.config.schemaPassed') }}</el-tag>
      </div>
      <div class="toolbar-right">
        <el-radio-group v-model="mode" size="small" data-testid="config-mode-toggle">
          <el-radio-button label="form">{{ t('plugin.config.formMode') }}</el-radio-button>
          <el-radio-button label="json">{{ t('plugin.config.jsonEditor') }}</el-radio-button>
        </el-radio-group>
        <el-button size="small" @click="format">{{ t('common.format') }}</el-button>
        <el-button size="small" data-testid="json-editor-clear-object" @click="clearObject">{{ t('common.clearObject') }}</el-button>
        <el-button size="small" @click="copy">{{ t('common.copy') }}</el-button>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" :title="t('plugin.config.tip')" />

    <div class="layout-grid" :class="{ 'no-preview': !showPreview }">
      <div class="left">
        <div v-if="mode === 'form'" class="form-mode" data-testid="plugin-config-form-mode">
          <el-empty v-if="!groups.length" :description="t('plugin.config.noFormSchema')" />
          <div v-else class="group-stack">
            <el-card v-for="group in groups" :key="group.key" shadow="never" class="group-card" :data-testid="`config-group-${group.key}`">
              <template #header>
                <div class="group-head">
                <div class="group-title">{{ group.title }}</div>
                <div class="group-sub">{{ group.subtitle }}</div>
              </div>
            </template>
              <el-form label-width="170px">
                <el-form-item v-for="field in group.fields" :key="field.path" :required="field.required">
                  <template #label>
                    <span>{{ field.title }}</span>
                    <span class="mono key-hint">{{ field.debugKey }}</span>
                    <el-tag v-if="field.required" class="label-tag" size="small" type="danger" effect="plain">{{ t('plugin.config.required') }}</el-tag>
                  </template>

                  <div class="field-control">
                    <el-select
                      v-if="field.enumOptions"
                      v-model="field.model.value"
                      clearable
                      filterable
                      :data-testid="`config-field-${field.path}`"
                    >
                      <el-option v-for="item in field.enumOptions" :key="String(item.value)" :label="String(item.label)" :value="item.value" />
                    </el-select>

                    <el-switch
                      v-else-if="field.type === 'boolean'"
                      v-model="field.model.value"
                      :data-testid="`config-field-${field.path}`"
                    />

                    <el-input-number
                      v-else-if="field.type === 'number' || field.type === 'integer'"
                      v-model="field.model.value"
                      :step="field.type === 'integer' ? 1 : 0.1"
                      :precision="field.type === 'integer' ? 0 : undefined"
                      :min="field.minimum"
                      :max="field.maximum"
                      controls-position="right"
                      :data-testid="`config-field-${field.path}`"
                    />

                    <el-input
                      v-else-if="field.type === 'array' && field.arrayKind === 'primitive'"
                      :model-value="arrayToText(field.model.value)"
                      type="textarea"
                      :rows="2"
                      :placeholder="t('plugin.config.arrayPlaceholder')"
                      :data-testid="`config-field-${field.path}`"
                      @update:model-value="(v) => (field.model.value = textToArray(v, field.arrayItemType))"
                    />

                    <el-alert
                      v-else-if="field.type === 'array' && field.arrayKind === 'object'"
                      type="warning"
                      show-icon
                      :closable="false"
                      class="inline-alert"
                      :title="t('plugin.config.arrayObjectTip')"
                    />

                    <el-input
                      v-else-if="field.type === 'object'"
                      :model-value="objectToText(field.model.value)"
                      type="textarea"
                      :rows="4"
                      :placeholder="field.placeholder"
                      :data-testid="`config-field-${field.path}`"
                      @update:model-value="(v) => setObjectField(field.model, v)"
                    />

                    <div v-else-if="field.sensitive" class="sensitive-wrap">
                      <el-input
                        v-model="field.model.value"
                        :type="field.reveal.value ? 'text' : 'password'"
                        :show-password="false"
                        :placeholder="field.placeholder"
                        :data-testid="`config-field-${field.path}`"
                      />
                      <el-button size="small" @click="field.reveal.value = !field.reveal.value" :data-testid="`config-sensitive-toggle-${field.path}`">
                        {{ field.reveal.value ? t('common.hide') : t('common.show') }}
                      </el-button>
                    </div>

                    <el-input
                      v-else
                      v-model="field.model.value"
                      :type="field.multiline ? 'textarea' : inputType(field.format)"
                      :rows="field.multiline ? 3 : undefined"
                      :placeholder="field.placeholder"
                      :data-testid="`config-field-${field.path}`"
                    />
                  </div>

                  <div class="field-help">
                    <span v-if="field.description">{{ field.description }}</span>
                    <span v-if="field.defaultValue !== undefined">{{ t('plugin.config.default') }}：{{ formatInline(field.defaultValue) }}</span>
                    <span v-if="field.minimum !== undefined">{{ t('plugin.config.minimum') }}：{{ field.minimum }}</span>
                    <span v-if="field.maximum !== undefined">{{ t('plugin.config.maximum') }}：{{ field.maximum }}</span>
                    <span v-if="field.stringLimits">{{ field.stringLimits }}</span>
                    <span v-if="field.booleanHint">{{ field.booleanHint }}</span>
                  </div>
                </el-form-item>
              </el-form>
            </el-card>
          </div>
        </div>

        <Suspense v-else>
          <template #default>
            <AsyncJsonEditorVue
              v-model="localValue"
              mode="tree"
              :main-menu-bar="true"
              :navigation-bar="true"
              :status-bar="true"
              data-testid="plugin-config-json-mode"
            />
          </template>
          <template #fallback>
            <div class="lazy-state">JSON 编辑器加载中...</div>
          </template>
        </Suspense>

        <div v-if="schemaErrors.length" class="error-box" data-testid="schema-error-box">
          <div class="error-title">{{ t('plugin.config.schemaErrors') }}</div>
          <ul class="error-list">
            <li v-for="(e, idx) in schemaErrors" :key="idx">{{ e }}</li>
          </ul>
        </div>
      </div>

      <div class="right" v-if="showPreview">
        <PluginConfigPreview
          :schema="schema"
          :default-config="defaultConfig"
          :original-config="originalConfig"
          :current-config="safeLocal"
          :effective-config="effectiveConfig"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import Ajv from 'ajv';
import { t } from '@/i18n';
import { safeJSON } from '@/i18n/formatters';
import PluginConfigPreview from './PluginConfigPreview.vue';
import { asNumberOrNull, getEnumOptions, getFieldGroup, isMultiline, isSensitiveField, normalizeSchema, normalizeType } from './configSchema';

const props = defineProps({
  modelValue: { type: Object, required: true },
  schema: { type: Object, default: null },
  defaultConfig: { type: Object, default: () => ({}) },
  originalConfig: { type: Object, default: () => ({}) },
  effectiveConfig: { type: Object, default: () => ({}) },
  showPreview: { type: Boolean, default: true },
});
const emit = defineEmits(['update:modelValue', 'schema-errors', 'mode-change']);

const localValue = ref(props.modelValue ?? {});
const mode = ref('form');
const AsyncJsonEditorVue = defineAsyncComponent({
  loader: () => import('json-editor-vue'),
  loadingComponent: {
    render: () => h('div', { class: 'lazy-state' }, 'JSON 编辑器加载中...'),
  },
  errorComponent: {
    render: () => h('div', { class: 'lazy-state error' }, 'JSON 编辑器加载失败，请稍后重试'),
  },
});

watch(
  () => props.modelValue,
  (v) => {
    localValue.value = v ?? {};
  },
);
watch(
  () => localValue.value,
  (v) => {
    const normalized = parseConfigObject(v);
    if (normalized.ok) {
      if (typeof v === 'string') localValue.value = normalized.value;
      emit('update:modelValue', normalized.value);
    }
  },
  { deep: true },
);
watch(mode, (v) => emit('mode-change', v));

const ajv = computed(() => {
  try {
    return new Ajv({ allErrors: true, strict: false, allowUnionTypes: true });
  } catch {
    return null;
  }
});

const schemaErrors = computed(() => {
  const schema = normalizeSchema(props.schema);
  if (!schema) return [];
  if (!ajv.value) return ['Ajv 初始化失败'];
  const normalized = parseConfigObject(localValue.value);
  if (!normalized.ok) {
    const out = [normalized.error];
    emit('schema-errors', out);
    return out;
  }
  try {
    const validate = ajv.value.compile(schema);
    const ok = validate(normalized.value);
    if (ok) {
      emit('schema-errors', []);
      return [];
    }
    const errs = validate.errors || [];
    const out = errs.map(formatSchemaError);
    emit('schema-errors', out);
    return out;
  } catch (e) {
    const out = [`${t('plugin.config.compileFailed')}：${String(e?.message || e)}`];
    emit('schema-errors', out);
    return out;
  }
});

const safeLocal = computed(() => {
  const normalized = parseConfigObject(localValue.value);
  return safeJSON(normalized.ok ? normalized.value : {});
});

function parseConfigObject(value) {
  let next = value ?? {};
  if (typeof next === 'string') {
    const text = next.trim();
    if (!text) return { ok: true, value: {} };
    try {
      next = JSON.parse(text);
    } catch (e) {
      return { ok: false, error: `JSON 解析失败：${String(e?.message || e)}` };
    }
  }
  if (!next || typeof next !== 'object' || Array.isArray(next)) {
    return { ok: false, error: 'JSON 配置必须是对象，例如 {"enabled": true}' };
  }
  return { ok: true, value: next };
}

function formatSchemaError(error) {
  const path = error?.instancePath || '$';
  if (error?.keyword === 'additionalProperties') {
    const field = error?.params?.additionalProperty || '';
    return field ? `${path}: 当前配置模型不允许字段 ${field}` : `${path}: 当前配置包含未声明字段`;
  }
  if (error?.keyword === 'type') {
    const expected = error?.params?.type;
    if (path === '$' && expected === 'object') return 'JSON 配置必须是对象，例如 {"enabled": true}';
    if (expected) return `${path}: 类型不匹配，应为 ${expected}`;
  }
  if (error?.keyword === 'required') {
    const field = error?.params?.missingProperty || '';
    return field ? `${path}: 缺少必填字段 ${field}` : `${path}: 缺少必填字段`;
  }
  if (error?.keyword === 'enum') return `${path}: 不在允许选项中`;
  if (error?.keyword === 'format') return `${path}: 格式不合法`;
  if (error?.keyword === 'pattern') return `${path}: 格式不符合要求`;
  return `${path}: ${error?.message || '配置不符合模型'}`;
}

const groups = computed(() => {
  const schema = normalizeSchema(props.schema) || {};
  const properties = schema.properties && typeof schema.properties === 'object' ? schema.properties : {};
  const required = new Set(Array.isArray(schema.required) ? schema.required : []);
  const map = new Map();

  for (const [key, def] of Object.entries(properties)) {
    const fieldDef = def && typeof def === 'object' ? def : {};
    const type = normalizeType(fieldDef);
    const rawGroup = String(getFieldGroup(fieldDef) || '').trim();
    const groupKey = rawGroup || 'basic';
    const groupTitle = rawGroup || '基础配置';
    const group = map.get(groupKey) || { key: groupKey, title: groupTitle, subtitle: '', fields: [] };

    // Default fill: only for missing keys, never override.
    if (safeLocal.value?.[key] === undefined && fieldDef.default !== undefined) {
      localValue.value[key] = fieldDef.default;
    }

    const field = buildField({ key, def: fieldDef, type, required: required.has(key) });
    group.fields.push(field);
    map.set(groupKey, group);
  }

  if (!map.has('basic')) map.set('basic', { key: 'basic', title: '基础配置', subtitle: '', fields: [] });
  const arr = Array.from(map.values());
  // Keep basic first, others by key.
  return arr.sort((a, b) => {
    if (a.key === 'basic') return -1;
    if (b.key === 'basic') return 1;
    return String(a.key).localeCompare(String(b.key));
  });
});

function buildField({ key, def, type, required }) {
  const enumOptions = getEnumOptions(def);
  const format = String(def.format || '');
  const multiline = isMultiline(def);
  const sensitive = isSensitiveField(key, def);
  const reveal = ref(false);

  const minimum = def.minimum ?? def.min;
  const maximum = def.maximum ?? def.max;
  const minLength = def.minLength;
  const maxLength = def.maxLength;
  const stringLimits =
    typeof minLength === 'number' || typeof maxLength === 'number'
      ? `${typeof minLength === 'number' ? `minLength=${minLength}` : ''}${typeof minLength === 'number' && typeof maxLength === 'number' ? ', ' : ''}${typeof maxLength === 'number' ? `maxLength=${maxLength}` : ''}`
      : '';

  const booleanHint = (() => {
    if (type !== 'boolean') return '';
    if (def['x-labels'] && typeof def['x-labels'] === 'object') {
      const tLabel = def['x-labels'].true || def['x-labels'].yes || '';
      const fLabel = def['x-labels'].false || def['x-labels'].no || '';
      if (tLabel || fLabel) return `${tLabel || 'true'} / ${fLabel || 'false'}`;
    }
    return '';
  })();

  const arrayInfo = getArrayInfo(def);

  const model = computed({
    get: () => safeLocal.value?.[key],
    set: (v) => {
      if (type === 'integer') {
        const n = asNumberOrNull(v);
        localValue.value[key] = n == null ? null : Math.trunc(n);
        return;
      }
      if (type === 'number') {
        const n = asNumberOrNull(v);
        localValue.value[key] = n == null ? null : n;
        return;
      }
      localValue.value[key] = v;
    },
  });

  const placeholder = sensitive ? t('plugin.config.sensitivePlaceholder') : '';
  const title = def.title || key;
  const debugKey = `(${key})`;
  return {
    path: key,
    debugKey,
    title,
    type,
    format,
    multiline,
    required,
    description: def.description || '',
    defaultValue: def.default,
    minimum,
    maximum,
    stringLimits,
    booleanHint,
    placeholder,
    enumOptions,
    sensitive,
    reveal,
    model,
    arrayKind: arrayInfo.kind,
    arrayItemType: arrayInfo.itemType,
  };
}

function getArrayInfo(def) {
  const items = def?.items && typeof def.items === 'object' ? def.items : null;
  const kind = items && (items.type === 'object' || items.properties) ? 'object' : 'primitive';
  const itemType = items ? normalizeType(items) : 'string';
  return { kind, itemType };
}

function inputType(format) {
  const f = String(format || '').toLowerCase();
  if (f === 'email') return 'email';
  if (f === 'uri' || f === 'url') return 'url';
  return 'text';
}

function arrayToText(value) {
  if (Array.isArray(value)) return value.join(', ');
  if (value == null) return '';
  return String(value);
}

function textToArray(value, itemType) {
  const parts = String(value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
  if (itemType === 'integer') return parts.map((x) => parseInt(x, 10)).filter((n) => Number.isFinite(n));
  if (itemType === 'number') return parts.map((x) => Number(x)).filter((n) => Number.isFinite(n));
  return parts;
}

function objectToText(value) {
  try {
    return JSON.stringify(value || {}, null, 2);
  } catch {
    return '{}';
  }
}

function setObjectField(model, value) {
  try {
    model.value = JSON.parse(value || '{}');
  } catch {
    model.value = value;
  }
}

async function copy() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(localValue.value ?? {}, null, 2));
    ElMessage.success(t('common.copied'));
  } catch {
    ElMessage.warning(t('common.copyUnsupported'));
  }
}

function format() {
  try {
    localValue.value = JSON.parse(JSON.stringify(localValue.value ?? {}));
    ElMessage.success(t('common.formatDone'));
  } catch {
    ElMessage.error(t('common.formatFailed'));
  }
}

function clearObject() {
  localValue.value = {};
  ElMessage.success(t('common.clearDone'));
}

function formatInline(value) {
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}
</script>

<style scoped>
.config-editor { display: grid; gap: 12px; }
.editor-toolbar { display: flex; justify-content: space-between; gap: 12px; align-items: center; }
.toolbar-left { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; justify-content: flex-end; }
.layout-grid { display: grid; grid-template-columns: minmax(0, 1fr) 420px; gap: 12px; align-items: start; }
.layout-grid.no-preview { grid-template-columns: minmax(0, 1fr); }
.form-mode { border: 1px solid #e2e8f0; border-radius: 12px; padding: 14px; background: #fff; }
.group-stack { display: grid; gap: 10px; }
.group-card { border-radius: 12px; }
.group-head { display: grid; gap: 2px; }
.group-title { font-weight: 700; color: #0f172a; }
.group-sub { color: #94a3b8; font-size: 12px; }
.field-help { margin-top: 6px; display: flex; gap: 10px; flex-wrap: wrap; color: #64748b; font-size: 12px; line-height: 1.5; }
.label-tag { margin-left: 6px; }
.key-hint { margin-left: 6px; color: #94a3b8; font-size: 11px; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.error-box { border: 1px solid #fecaca; background: #fff1f2; border-radius: 12px; padding: 12px; }
.error-title { font-weight: 700; color: #991b1b; margin-bottom: 8px; }
.error-list { margin: 0; padding-left: 18px; color: #7f1d1d; font-size: 12px; line-height: 1.6; }
.inline-alert { margin: 6px 0; }
.sensitive-wrap { display: flex; gap: 8px; align-items: center; }
.field-control { width: 100%; }
.lazy-state {
  padding: 14px;
  border: 1px solid #dbeafe;
  border-radius: 10px;
  background: #f8fbff;
  color: #475569;
  font-size: 13px;
}
.lazy-state.error {
  border-color: #fecaca;
  background: #fff7f7;
  color: #b91c1c;
}
@media (max-width: 1100px) {
  .layout-grid { grid-template-columns: 1fr; }
}
</style>
