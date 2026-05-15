<template>
  <el-tag :type="tagType" effect="plain" :data-testid="testid">{{ level || '-' }}</el-tag>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  level: { type: String, default: '' },
  testid: { type: String, default: '' },
});

const tagType = computed(() => {
  const level = String(props.level || '').toLowerCase();
  if (level === 'low') return 'success';
  if (level === 'medium' || level === 'high') return 'warning';
  if (level === 'blocked') return 'danger';
  return 'info';
});
</script>
