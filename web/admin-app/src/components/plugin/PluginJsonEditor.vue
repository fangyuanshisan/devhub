<template>
  <div class="editor-shell" data-testid="plugin-json-editor">
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <slot name="title" />
        <el-tag v-if="schemaErrors.length" data-testid="schema-error-badge" type="danger" effect="light">{{ t('plugin.config.schemaFailed') }}</el-tag>
        <el-tag v-else data-testid="schema-success-badge" type="success" effect="light">{{ t('plugin.config.schemaPassed') }}</el-tag>
      </div>
      <div class="toolbar-right">
        <el-radio-group v-model="mode" size="small">
          <el-radio-button label="form">{{ t('plugin.config.formMode') }}</el-radio-button>
          <el-radio-button label="json">{{ t('plugin.config.jsonEditor') }}</el-radio-button>
        </el-radio-group>
        <el-button size="small" @click="format">{{ t('common.format') }}</el-button>
        <el-button size="small" data-testid="json-editor-clear-object" @click="clearObject">{{ t('common.clearObject') }}</el-button>
        <el-button size="small" @click="copy">{{ t('common.copy') }}</el-button>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" :title="t('plugin.config.tip')" />

    <div v-if="mode === 'form'" class="form-mode" data-testid="plugin-config-form-mode">
      <el-empty v-if="!schemaFields.length" :description="t('plugin.config.noFormSchema')" />
      <el-form v-else label-width="170px">
        <el-form-item v-for="field in schemaFields" :key="field.key" :required="field.required">
          <template #label>
            <span>{{ field.title }}</span>
            <el-tag v-if="field.required" class="label-tag" size="small" type="danger" effect="plain">{{ t('plugin.config.required') }}</el-tag>
          </template>

          <el-select v-if="field.enumValues" v-model="localValue[field.key]" clearable filterable>
            <el-option v-for="item in field.enumValues" :key="String(item)" :label="String(item)" :value="item" />
          </el-select>
          <el-switch v-else-if="field.type === 'boolean'" v-model="localValue[field.key]" />
          <el-input-number
            v-else-if="field.type === 'number' || field.type === 'integer'"
            v-model="localValue[field.key]"
            :step="field.type === 'integer' ? 1 : 0.1"
            :precision="field.type === 'integer' ? 0 : undefined"
            :min="field.minimum"
            :max="field.maximum"
            controls-position="right"
          />
          <el-input
            v-else-if="field.type === 'array'"
            :model-value="arrayToText(localValue[field.key])"
            type="textarea"
            :rows="2"
            :placeholder="t('plugin.config.arrayPlaceholder')"
            @update:model-value="(v) => setArrayField(field.key, v)"
          />
          <el-input
            v-else-if="field.type === 'object'"
            :model-value="objectToText(localValue[field.key])"
            type="textarea"
            :rows="4"
            placeholder='{"enabled": true}'
            @update:model-value="(v) => setObjectField(field.key, v)"
          />
          <el-input v-else v-model="localValue[field.key]" :type="field.sensitive ? 'password' : 'text'" show-password />

          <div class="field-help">
            <span v-if="field.description">{{ field.description }}</span>
            <span v-if="field.defaultValue !== undefined">{{ t('plugin.config.default') }}：{{ formatInline(field.defaultValue) }}</span>
            <span v-if="field.enumValues">{{ t('plugin.config.enum') }}：{{ field.enumValues.join(' / ') }}</span>
            <span v-if="field.minimum !== undefined">{{ t('plugin.config.minimum') }}：{{ field.minimum }}</span>
            <span v-if="field.maximum !== undefined">{{ t('plugin.config.maximum') }}：{{ field.maximum }}</span>
          </div>
        </el-form-item>
      </el-form>
    </div>

    <JsonEditorVue v-else v-model="localValue" mode="tree" :main-menu-bar="true" :navigation-bar="true" :status-bar="true" />

    <div v-if="showDiff" class="config-preview">
      <el-collapse>
        <el-collapse-item :title="t('plugin.config.configDiff')" name="diff">
          <div class="diff-summary">
            <el-tag v-if="changedKeys.length" type="warning" effect="plain">{{ t('plugin.config.changedKeys') }}：{{ changedKeys.join(', ') }}</el-tag>
            <el-tag v-else type="success" effect="plain">{{ t('plugin.config.noChanges') }}</el-tag>
          </div>
          <div class="preview-grid">
            <div>
              <h4>{{ t('plugin.config.oldValue') }}</h4>
              <pre class="json-box compact">{{ formatJSON(maskedOriginal) }}</pre>
            </div>
            <div>
              <h4>{{ t('plugin.config.newValue') }}</h4>
              <pre class="json-box compact">{{ formatJSON(maskedLocal) }}</pre>
            </div>
          </div>
        </el-collapse-item>
        <el-collapse-item v-if="resolvedConfig && Object.keys(resolvedConfig).length" :title="t('plugin.config.effectiveConfig')" name="effective">
          <pre class="json-box compact">{{ formatJSON(maskSensitiveConfig(resolvedConfig)) }}</pre>
        </el-collapse-item>
      </el-collapse>
    </div>

    <div v-if="schemaErrors.length" class="error-box" data-testid="schema-error-box">
      <div class="error-title">{{ t('plugin.config.schemaErrors') }}</div>
      <ul class="error-list">
        <li v-for="(e, idx) in schemaErrors" :key="idx">{{ e }}</li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import JsonEditorVue from 'json-editor-vue';
import Ajv from 'ajv';
import { t } from '@/i18n';
import { isSensitiveKey, maskSensitiveConfig, safeJSON, topLevelChangedKeys } from '@/i18n/formatters';

const props = defineProps({
  modelValue: { type: Object, required: true },
  schema: { type: Object, default: null },
  originalValue: { type: Object, default: () => ({}) },
  resolvedConfig: { type: Object, default: null },
  showDiff: { type: Boolean, default: true },
});
const emit = defineEmits(['update:modelValue', 'schema-errors']);

const localValue = ref(props.modelValue ?? {});
const mode = ref('form');

watch(
  () => props.modelValue,
  (v) => {
    localValue.value = v ?? {};
  },
);

watch(
  () => localValue.value,
  (v) => {
    emit('update:modelValue', v ?? {});
  },
  { deep: true },
);

const ajv = computed(() => {
  try {
    return new Ajv({ allErrors: true, strict: false, allowUnionTypes: true });
  } catch {
    return null;
  }
});

const schemaErrors = computed(() => {
  if (!props.schema || typeof props.schema !== 'object') return [];
  if (!ajv.value) return ['Ajv 初始化失败'];
  try {
    const validate = ajv.value.compile(props.schema);
    const ok = validate(localValue.value);
    if (ok) {
      emit('schema-errors', []);
      return [];
    }
    const errs = validate.errors || [];
    const out = errs.map((e) => `${e.instancePath || '$'}: ${e.message || 'invalid'}`);
    emit('schema-errors', out);
    return out;
  } catch (e) {
    const out = [`${t('plugin.config.compileFailed')}：${String(e?.message || e)}`];
    emit('schema-errors', out);
    return out;
  }
});

const schemaFields = computed(() => {
  const schema = props.schema && typeof props.schema === 'object' ? props.schema : {};
  const properties = schema.properties && typeof schema.properties === 'object' ? schema.properties : {};
  const required = new Set(Array.isArray(schema.required) ? schema.required : []);
  return Object.entries(properties).map(([key, def]) => {
    const field = def && typeof def === 'object' ? def : {};
    const type = normalizeType(field);
    if (localValue.value?.[key] === undefined && field.default !== undefined) {
      localValue.value[key] = field.default;
    }
    return {
      key,
      title: field.title || key,
      type,
      required: required.has(key),
      description: field.description || '',
      enumValues: Array.isArray(field.enum) ? field.enum : null,
      defaultValue: field.default,
      minimum: field.minimum ?? field.min,
      maximum: field.maximum ?? field.max,
      sensitive: isSensitiveKey(key) || field.sensitive === true,
    };
  });
});

const changedKeys = computed(() => topLevelChangedKeys(safeJSON(props.originalValue), safeJSON(localValue.value)));
const maskedOriginal = computed(() => maskSensitiveConfig(safeJSON(props.originalValue)));
const maskedLocal = computed(() => maskSensitiveConfig(safeJSON(localValue.value)));

async function copy() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(localValue.value ?? {}, null, 2));
    ElMessage.success(t('common.copied'));
  } catch {
    ElMessage.warning(t('common.copyUnsupported'));
  }
}

function format() {
  // JsonEditorVue 内部会格式化展示；这里保持值不变即可。
  // 但为了“可预期”，我们通过 stringify/parse 规整 key 顺序与缩进。
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

function normalizeType(field) {
  const raw = Array.isArray(field.type) ? field.type.find((item) => item !== 'null') : field.type;
  if (field.enum) return raw || 'string';
  if (raw === 'integer' || raw === 'number' || raw === 'boolean' || raw === 'array' || raw === 'object') return raw;
  return 'string';
}

function arrayToText(value) {
  if (Array.isArray(value)) return value.join(', ');
  if (value == null) return '';
  return String(value);
}

function setArrayField(key, value) {
  localValue.value[key] = String(value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function objectToText(value) {
  return formatJSON(value || {});
}

function setObjectField(key, value) {
  try {
    localValue.value[key] = JSON.parse(value || '{}');
  } catch {
    localValue.value[key] = value;
  }
}

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
}

function formatInline(value) {
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}
</script>

<style scoped>
.editor-shell {
  display: grid;
  gap: 12px;
}
.editor-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}
.toolbar-left {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
}
.error-box {
  border: 1px solid #fecaca;
  background: #fff1f2;
  border-radius: 12px;
  padding: 12px;
}
.error-title {
  font-weight: 700;
  color: #991b1b;
  margin-bottom: 8px;
}
.error-list {
  margin: 0;
  padding-left: 18px;
  color: #7f1d1d;
  font-size: 12px;
  line-height: 1.6;
}
.form-mode {
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 14px;
  background: #fff;
}
.label-tag {
  margin-left: 6px;
}
.field-help {
  margin-top: 6px;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  color: #64748b;
  font-size: 12px;
  line-height: 1.5;
}
.config-preview {
  display: grid;
  gap: 10px;
}
.diff-summary {
  margin-bottom: 10px;
}
.preview-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}
.preview-grid h4 {
  margin: 0 0 8px;
  color: #0f172a;
}
.json-box {
  margin: 0;
  padding: 14px;
  border-radius: 12px;
  background: #0f172a;
  color: #dbeafe;
  max-height: 360px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
}
.json-box.compact {
  max-height: 220px;
}
@media (max-width: 900px) {
  .preview-grid {
    grid-template-columns: 1fr;
  }
}
</style>
