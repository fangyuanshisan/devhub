<template>
  <el-drawer v-model="visible" class="admin-detail-drawer" :title="title" :size="size" :data-testid="testid">
    <div class="action-panel in-drawer">
      <div class="action-panel-header">
        <div>
          <h3>{{ title }}</h3>
          <p v-if="subtitle" class="muted">{{ subtitle }}</p>
        </div>
        <div class="action-panel-tools">
          <AdminStatusTag v-if="status" :value="status" />
          <slot name="actions" />
        </div>
      </div>
      <el-tabs v-if="tabs.length" v-model="activeTab">
        <el-tab-pane v-for="item in tabs" :key="item.name" :label="item.label" :name="item.name">
          <slot :name="item.name" />
        </el-tab-pane>
      </el-tabs>
      <slot v-else />
      <AdminTechnicalDetails v-if="technicalDetails.length" class="mt" :blocks="technicalDetails" />
    </div>
  </el-drawer>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import AdminStatusTag from './AdminStatusTag.vue';
import AdminTechnicalDetails from './AdminTechnicalDetails.vue';

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '详情' },
  subtitle: { type: String, default: '' },
  status: { type: String, default: '' },
  tabs: { type: Array, default: () => [] },
  initialTab: { type: String, default: '' },
  technicalDetails: { type: Array, default: () => [] },
  size: { type: String, default: 'var(--dh-admin-drawer-width)' },
  testid: { type: String, default: '' },
});
const emit = defineEmits(['update:modelValue', 'update:tab']);

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});
const activeTab = ref(props.initialTab || props.tabs[0]?.name || '');

watch(() => props.initialTab, (value) => {
  if (value) activeTab.value = value;
});
watch(activeTab, (value) => emit('update:tab', value));
</script>
