<template>
  <el-tag class="admin-risk-tag" :type="tagType" effect="plain" :data-testid="testid">{{ label }}</el-tag>
</template>

<script setup>
import { computed } from 'vue';
import { pluginRiskText } from '@/modules/plugins/statusText';

const props = defineProps({
  level: { type: String, default: 'unknown' },
  testid: { type: String, default: '' },
});

const normalized = computed(() => String(props.level || 'unknown').toLowerCase());
const label = computed(() => {
  const value = normalized.value;
  if (value === 'safe') return '安全';
  if (value === 'unknown') return '未知风险';
  return pluginRiskText(value, '未知风险');
});
const tagType = computed(() => {
  const value = normalized.value;
  if (value === 'safe' || value === 'low') return 'success';
  if (value === 'warning' || value === 'medium' || value === 'high') return 'warning';
  if (value === 'blocked') return 'danger';
  return 'info';
});
</script>
