<template>
  <AdminRiskTag :level="level" :testid="testid" />
</template>

<script setup>
import AdminRiskTag from '@/components/admin/AdminRiskTag.vue';

defineProps({
  level: { type: String, default: '' },
  testid: { type: String, default: '' },
});
</script>
