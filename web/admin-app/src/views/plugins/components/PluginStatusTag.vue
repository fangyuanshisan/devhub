<template>
  <AdminStatusTag :value="value" :testid="testid" />
</template>

<script setup>
import AdminStatusTag from '@/components/admin/AdminStatusTag.vue';

defineProps({
  value: { type: String, default: '' },
  testid: { type: String, default: '' },
});
</script>
