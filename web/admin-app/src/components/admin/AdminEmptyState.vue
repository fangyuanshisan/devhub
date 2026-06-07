<template>
  <section class="admin-empty-state" :data-testid="testid">
    <div class="admin-empty-state__title">{{ title }}</div>
    <div v-if="description" class="admin-empty-state__description">{{ description }}</div>
    <div v-if="$slots.action" class="admin-empty-state__actions mt">
      <slot name="action" />
    </div>
  </section>
</template>

<script setup>
defineProps({
  title: { type: String, default: '暂无数据' },
  description: { type: String, default: '' },
  testid: { type: String, default: '' },
});
</script>
