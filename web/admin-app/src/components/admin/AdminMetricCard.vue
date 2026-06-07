<template>
  <component :is="clickable ? 'button' : 'section'" class="admin-metric-card" :class="{ 'is-clickable': clickable }" type="button" :data-testid="testid" @click="handleClick">
    <div class="admin-metric-card__top">
      <div class="admin-metric-card__label">{{ label }}</div>
      <AdminStatusTag v-if="status" :value="status" />
    </div>
    <div class="admin-metric-card__value">{{ value }}</div>
    <div v-if="help" class="admin-metric-card__help">{{ help }}</div>
    <div v-if="trend" class="admin-metric-card__trend">{{ trend }}</div>
  </component>
</template>

<script setup>
import { computed } from 'vue';
import AdminStatusTag from './AdminStatusTag.vue';

const props = defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], default: '-' },
  status: { type: String, default: '' },
  help: { type: String, default: '' },
  trend: { type: String, default: '' },
  testid: { type: String, default: '' },
  clickable: { type: Boolean, default: false },
});
const emit = defineEmits(['click']);
const clickable = computed(() => props.clickable);

function handleClick(event) {
  if (clickable.value) emit('click', event);
}
</script>
