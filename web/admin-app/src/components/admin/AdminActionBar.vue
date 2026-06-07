<template>
  <div class="admin-action-bar" :data-testid="testid">
    <el-tooltip v-for="action in actions" :key="action.key || action.label" :disabled="!action.disabledReason" :content="action.disabledReason">
      <span>
        <el-button :type="action.type || typeFor(action)" :plain="action.plain !== false" :disabled="action.disabled || !!action.disabledReason" :loading="action.loading" :data-testid="action.testid" @click="$emit('action', action)">
          {{ action.label }}
        </el-button>
      </span>
    </el-tooltip>
  </div>
</template>

<script setup>
defineProps({
  actions: { type: Array, default: () => [] },
  testid: { type: String, default: '' },
});
defineEmits(['action']);

function typeFor(action) {
  if (action.danger) return 'danger';
  if (action.primary) return 'primary';
  return 'default';
}
</script>
