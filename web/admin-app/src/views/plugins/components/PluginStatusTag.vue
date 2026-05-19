<template>
  <el-tag :type="tagType" effect="plain" :data-testid="testid">{{ label }}</el-tag>
</template>

<script setup>
import { computed } from 'vue';
import { pluginStatusText, pluginTagType } from '@/modules/plugins/statusText';

const props = defineProps({
  value: { type: String, default: '' },
  testid: { type: String, default: '' },
});

const tagType = computed(() => {
  return pluginTagType(props.value);
});

const label = computed(() => pluginStatusText(props.value));
</script>
