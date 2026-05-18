<template>
  <el-collapse class="technical-collapse">
    <el-collapse-item v-for="block in blocks" :key="block.name" :name="block.name" :title="block.title">
      <pre class="json-box">{{ formatJSON(block.value) }}</pre>
    </el-collapse-item>
  </el-collapse>
</template>

<script setup>
defineProps({
  blocks: { type: Array, default: () => [] },
});

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return String(value ?? '');
  }
}
</script>

<style scoped>
.technical-collapse :deep(.el-collapse-item__header) {
  padding: 0 12px;
  font-weight: 600;
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
  word-break: break-word;
}
</style>
