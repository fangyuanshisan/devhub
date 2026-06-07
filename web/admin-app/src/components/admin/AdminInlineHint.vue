<template>
  <div class="admin-inline-hint" :class="`admin-inline-hint--${type}`" :data-testid="testid">
    <el-icon v-if="icon"><component :is="icon" /></el-icon>
    <div>
      <strong v-if="title">{{ title }}</strong>
      <div><slot>{{ description }}</slot></div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { CircleCheck, InfoFilled, WarningFilled } from '@element-plus/icons-vue';

const props = defineProps({
  type: { type: String, default: 'info' },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  testid: { type: String, default: '' },
});

const icon = computed(() => {
  if (props.type === 'success') return CircleCheck;
  if (props.type === 'warning' || props.type === 'danger') return WarningFilled;
  return InfoFilled;
});
</script>
