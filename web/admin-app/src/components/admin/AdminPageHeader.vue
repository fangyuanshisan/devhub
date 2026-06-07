<template>
  <header class="admin-page-header" :data-testid="testid">
    <div class="admin-page-header__main">
      <div v-if="breadcrumbs.length" class="admin-page-header__breadcrumbs">
        <span v-for="(item, index) in breadcrumbs" :key="`${item}-${index}`">{{ item }}</span>
      </div>
      <div class="admin-page-header__title-row">
        <h1>{{ title }}</h1>
        <AdminStatusTag v-if="status" :value="status" />
      </div>
      <p v-if="description" class="admin-page-header__description">{{ description }}</p>
      <p v-if="helpText" class="admin-page-header__help">{{ helpText }}</p>
    </div>
    <div v-if="$slots.actions || primaryAction || secondaryActions.length" class="admin-page-header__actions">
      <slot name="actions">
        <el-button v-for="action in secondaryActions" :key="action.key || action.label" :type="action.type || 'default'" :plain="action.plain !== false" :disabled="action.disabled" :loading="action.loading" :data-testid="action.testid" @click="$emit('action', action)">
          {{ action.label }}
        </el-button>
        <el-button v-if="primaryAction" type="primary" :disabled="primaryAction.disabled" :loading="primaryAction.loading" :data-testid="primaryAction.testid" @click="$emit('action', primaryAction)">
          {{ primaryAction.label }}
        </el-button>
      </slot>
    </div>
  </header>
</template>

<script setup>
import AdminStatusTag from './AdminStatusTag.vue';

defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' },
  breadcrumbs: { type: Array, default: () => [] },
  primaryAction: { type: Object, default: null },
  secondaryActions: { type: Array, default: () => [] },
  status: { type: String, default: '' },
  helpText: { type: String, default: '' },
  testid: { type: String, default: '' },
});
defineEmits(['action']);
</script>
