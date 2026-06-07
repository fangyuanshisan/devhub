<template>
  <div class="preview" data-testid="plugin-config-preview">
    <el-card shadow="never" class="preview-card">
      <template #header>
        <div class="head">
          <div class="title">配置模型</div>
          <div class="sub">默认 / 当前 / 最终生效 对比</div>
        </div>
      </template>

      <div class="section">
        <div class="sec-title">配置差异</div>
        <div class="chips">
          <el-tag v-if="changed.length" type="warning" effect="plain" data-testid="config-changed-keys">
            变更字段：{{ changed.join(', ') }}
          </el-tag>
          <el-tag v-else type="success" effect="plain">没有配置变更</el-tag>
        </div>
      </div>

      <div class="section">
        <div class="sec-title">默认配置</div>
        <pre class="json-box compact">{{ formatJSON(maskedDefault) }}</pre>
      </div>

      <div class="section">
        <div class="sec-title">当前配置</div>
        <pre class="json-box compact">{{ formatJSON(maskedCurrent) }}</pre>
      </div>

      <div class="section">
        <div class="sec-title">最终生效配置</div>
        <pre class="json-box compact">{{ formatJSON(maskedEffective) }}</pre>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { changedPaths, maskSensitiveConfig, safeJSON } from '@/i18n/formatters';

const props = defineProps({
  schema: { type: Object, default: null },
  defaultConfig: { type: Object, default: () => ({}) },
  originalConfig: { type: Object, default: () => ({}) },
  currentConfig: { type: Object, default: () => ({}) },
  effectiveConfig: { type: Object, default: () => ({}) },
});

const maskedDefault = computed(() => maskSensitiveConfig(safeJSON(props.defaultConfig)));
const maskedCurrent = computed(() => maskSensitiveConfig(safeJSON(props.currentConfig)));
const maskedEffective = computed(() => maskSensitiveConfig(safeJSON(props.effectiveConfig)));

const changed = computed(() => {
  const paths = changedPaths(safeJSON(props.originalConfig), safeJSON(props.currentConfig), { maxDepth: 4 });
  // Display without "$." prefix.
  return paths.map((p) => p.replace(/^\$\./, ''));
});

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
}
</script>

<style scoped>
.preview-card { border-radius: 12px; }
.head { display: grid; gap: 2px; }
.title { font-weight: 700; color: #0f172a; }
.sub { color: #94a3b8; font-size: 12px; }
.section { margin-top: 10px; }
.sec-title { font-weight: 700; color: #0f172a; font-size: 12px; margin-bottom: 6px; }
.chips { display: flex; flex-wrap: wrap; gap: 6px; }
.json-box {
  margin: 0;
  padding: 12px;
  border-radius: 12px;
  background: #0f172a;
  color: #dbeafe;
  max-height: 220px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
}
.json-box.compact { max-height: 180px; }
</style>

