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
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  margin: 12px 0;
  border-radius: 10px;
  background: #fff;
  border: 1px solid #e9edf3;
}

.plugin-filter-bar-title {
  min-width: 220px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.plugin-filter-bar-title strong {
  display: block;
  line-height: 20px;
}

.plugin-filter-bar-controls {
  display: flex;
  flex: 1;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}
</style>

