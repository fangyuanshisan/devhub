<template>
  <el-tag :type="tagType" effect="plain" :data-testid="testid">{{ value || '-' }}</el-tag>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  value: { type: String, default: '' },
  testid: { type: String, default: '' },
});

const tagType = computed(() => {
  const status = String(props.value || '').toLowerCase();
  if (['enabled', 'ok', 'staged', 'approved', 'promoted', 'applied', 'verified', 'trusted'].includes(status)) return 'success';
  if (['disabled', 'warning', 'approval_pending', 'needs_reencrypt', 'medium'].includes(status)) return 'warning';
  if (['archived', 'uploaded', 'scanned', 'canceled', 'deleted'].includes(status)) return 'info';
  if (['blocked', 'failed', 'config_invalid', 'migration_failed', 'revoked', 'mismatch'].includes(status)) return 'danger';
  return 'info';
});
</script>
