<template>
  <span class="masked-ref" :title="title">
    <span class="text">{{ masked }}</span>
    <el-button v-if="value" link type="primary" size="small" class="copy" @click="copy">复制</el-button>
  </span>
</template>

<script setup>
import { computed } from 'vue';
import { ElMessage } from 'element-plus';
import { maskRef } from '@/i18n/formatters';

const props = defineProps({
  value: { type: String, default: '' },
  title: { type: String, default: '引用已脱敏显示，可复制原始值。' },
});

const masked = computed(() => maskRef(props.value));

async function copy() {
  try {
    await navigator.clipboard.writeText(String(props.value || ''));
    ElMessage.success('已复制');
  } catch (e) {
    ElMessage.error(e?.message || '复制失败');
  }
}
</script>

<style scoped>
.masked-ref { display: inline-flex; align-items: center; gap: 6px; }
.text { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
.copy { padding: 0; }
</style>

