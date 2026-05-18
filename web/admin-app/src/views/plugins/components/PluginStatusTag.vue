<template>
  <el-tag :type="tagType" effect="plain" :data-testid="testid">{{ label }}</el-tag>
</template>

<script setup>
import { computed } from 'vue';
import { genericStatusLabel } from '@/i18n/formatters';

const props = defineProps({
  value: { type: String, default: '' },
  testid: { type: String, default: '' },
});

const tagType = computed(() => {
  const status = String(props.value || '').toLowerCase();
  if (['enabled', 'ok', 'staged', 'approved', 'promoted', 'applied', 'verified', 'trusted', 'success', 'closed', 'accepted', 'compatible', 'active'].includes(status)) return 'success';
  if (['disabled', 'warning', 'approval_pending', 'needs_reencrypt', 'medium', 'pending', 'running', 'retry_scheduled', 'half_open', 'previous'].includes(status)) return 'warning';
  if (['archived', 'uploaded', 'scanned', 'canceled', 'deleted', 'skipped', 'expired'].includes(status)) return 'info';
  if (['blocked', 'failed', 'config_invalid', 'migration_failed', 'revoked', 'mismatch', 'retry_exhausted', 'circuit_open', 'open', 'rejected', 'incompatible'].includes(status)) return 'danger';
  return 'info';
});

const label = computed(() => genericStatusLabel(props.value));
</script>
