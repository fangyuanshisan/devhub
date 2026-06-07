<template>
  <el-tag class="admin-status-tag" :type="tagType" effect="plain" :data-testid="testid">{{ label }}</el-tag>
</template>

<script setup>
import { computed } from 'vue';
import { pluginStatusText, pluginTagType } from '@/modules/plugins/statusText';

const props = defineProps({
  value: { type: String, default: 'unknown' },
  testid: { type: String, default: '' },
});

const tagType = computed(() => pluginTagType(props.value));
const label = computed(() => pluginStatusText(props.value, '未知'));
</script>
