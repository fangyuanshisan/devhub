<template>
  <el-select-v2
    :model-value="modelValue"
    :options="options"
    :loading="loading"
    :placeholder="placeholder"
    :size="size"
    :height="300"
    :style="selectStyle"
    clearable
    filterable
    default-first-option
    @update:model-value="emit('update:modelValue', $event || '')"
  />
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  modelValue: { type: String, default: '' },
  options: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  placeholder: { type: String, default: '选择插件' },
  size: { type: String, default: 'small' },
  width: { type: String, default: '100%' },
});

const emit = defineEmits(['update:modelValue']);

const selectStyle = computed(() => ({ width: props.width }));
</script>
