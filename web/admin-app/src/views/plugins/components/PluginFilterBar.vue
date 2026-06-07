<template>
  <div class="plugin-filter-bar" :data-testid="testid">
    <div class="plugin-filter-bar-title">
      <div>
        <strong>{{ title }}</strong>
        <span v-if="tip" class="muted">{{ tip }}</span>
      </div>
      <slot name="actions" />
    </div>
    <div class="plugin-filter-bar-controls">
      <slot />
    </div>
    <div v-if="$slots.advanced" class="plugin-filter-bar-advanced">
      <slot name="advanced" />
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, default: '' },
  tip: { type: String, default: '' },
  testid: { type: String, default: '' },
});
</script>

<style scoped>
.plugin-filter-bar {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  margin: 12px 0;
  border-radius: 8px;
  background: #fff;
  border: 1px solid #e9edf3;
}

.plugin-filter-bar-title {
  width: 100%;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.plugin-filter-bar-title strong {
  display: block;
  line-height: 20px;
}

.plugin-filter-bar-controls {
  display: grid;
  width: 100%;
  grid-template-columns: minmax(220px, 1.5fr) repeat(4, minmax(136px, 1fr)) repeat(2, minmax(82px, 0.55fr));
  align-items: center;
  gap: 8px;
}

.plugin-filter-bar-advanced {
  display: grid;
  width: 100%;
  grid-template-columns: repeat(auto-fit, minmax(160px, 220px));
  align-items: center;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px dashed #e5eaf2;
}

.plugin-filter-bar-controls :deep(.el-input),
.plugin-filter-bar-controls :deep(.el-select),
.plugin-filter-bar-controls :deep(.el-button),
.plugin-filter-bar-advanced :deep(.el-input),
.plugin-filter-bar-advanced :deep(.el-select),
.plugin-filter-bar-advanced :deep(.el-button) {
  width: 100%;
}

@media (max-width: 1180px) {
  .plugin-filter-bar-controls {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .plugin-filter-bar-controls {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
