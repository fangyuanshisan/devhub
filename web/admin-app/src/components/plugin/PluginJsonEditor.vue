<template>
  <div class="editor-shell" data-testid="plugin-json-editor">
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <slot name="title" />
        <el-tag v-if="schemaErrors.length" data-testid="schema-error-badge" type="danger" effect="light">schema 校验失败</el-tag>
        <el-tag v-else data-testid="schema-success-badge" type="success" effect="light">schema 通过</el-tag>
      </div>
      <div class="toolbar-right">
        <el-button size="small" @click="format">格式化</el-button>
        <el-button size="small" data-testid="json-clear-object" @click="clearObject">清空为 {}</el-button>
        <el-button size="small" @click="copy">复制</el-button>
      </div>
    </div>

    <JsonEditorVue v-model="localValue" mode="tree" :main-menu-bar="true" :navigation-bar="true" :status-bar="true" />

    <div v-if="schemaErrors.length" class="error-box" data-testid="schema-error-box">
      <div class="error-title">Schema 校验错误</div>
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

const props = defineProps({
  modelValue: { type: Object, required: true },
  schema: { type: Object, default: null },
});
const emit = defineEmits(['update:modelValue', 'schema-errors']);

const localValue = ref(props.modelValue ?? {});

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
    const out = [`schema 编译失败：${String(e?.message || e)}`];
    emit('schema-errors', out);
    return out;
  }
});

async function copy() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(localValue.value ?? {}, null, 2));
    ElMessage.success('已复制');
  } catch {
    ElMessage.warning('当前浏览器不支持自动复制');
  }
}

function format() {
  // JsonEditorVue 内部会格式化展示；这里保持值不变即可。
  // 但为了“可预期”，我们通过 stringify/parse 规整 key 顺序与缩进。
  try {
    localValue.value = JSON.parse(JSON.stringify(localValue.value ?? {}));
    ElMessage.success('已格式化');
  } catch {
    ElMessage.error('格式化失败');
  }
}

function clearObject() {
  localValue.value = {};
  ElMessage.success('已清空');
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
</style>
