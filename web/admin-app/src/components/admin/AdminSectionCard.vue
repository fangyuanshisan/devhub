<template>
  <section class="admin-section-card" :class="{ 'is-empty': empty }" :data-testid="testid">
    <div v-if="title || description || status || $slots.actions" class="admin-section-card__header">
      <div>
        <h2 v-if="title" class="admin-section-card__title">{{ title }}</h2>
        <p v-if="description" class="admin-section-card__description">{{ description }}</p>
      </div>
      <div class="admin-section-card__actions">
        <AdminStatusTag v-if="status" :value="status" />
        <slot name="actions" />
      </div>
    </div>
    <div v-loading="loading" class="admin-section-card__body">
      <AdminEmptyState v-if="empty" :title="emptyTitle" :description="emptyDescription" />
      <slot v-else />
    </div>
  </section>
</template>

<script setup>
import AdminEmptyState from './AdminEmptyState.vue';
import AdminStatusTag from './AdminStatusTag.vue';

defineProps({
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  status: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  empty: { type: Boolean, default: false },
  emptyTitle: { type: String, default: '暂无数据' },
  emptyDescription: { type: String, default: '当前没有可展示的记录。' },
  testid: { type: String, default: '' },
});
</script>
