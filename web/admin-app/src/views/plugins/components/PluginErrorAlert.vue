<template>
  <el-alert
    v-if="message"
    type="error"
    show-icon
    :closable="false"
    :title="message"
    data-testid="plugin-error-alert"
  />
</template>

<script setup>
defineProps({
  message: { type: String, default: '' },
});
</script>
