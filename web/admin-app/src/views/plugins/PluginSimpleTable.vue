<template>
  <el-table :data="rows" border stripe :empty-text="emptyText" :data-testid="testid">
    <el-table-column prop="name" label="插件" min-width="200">
      <template #default="{ row }">
        <div class="plugin-title compact-title">
          <strong>{{ row.name }}</strong>
          <span class="mono">{{ row.code }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="status" label="状态" width="120">
      <template #default="{ row }">
        <el-tag :type="statusTagType(row.status)" effect="plain">{{ row.status }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="reason" label="说明" min-width="260">
      <template #default="{ row }">
        <span class="muted">{{ row.reason || row.health?.recent_error || row.health?.status_reason || '-' }}</span>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="160">
      <template #default="{ row }">
        <el-button link type="primary" @click="$emit('open', row)">打开详情</el-button>
        <el-button v-if="tab" link type="primary" @click="$emit('open', row, tab)">指定 Tab</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup>
defineProps({
  rows: { type: Array, default: () => [] },
  emptyText: { type: String, default: '暂无数据' },
  testid: { type: String, default: '' },
  tab: { type: String, default: '' },
});
defineEmits(['open']);

function statusTagType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'warning';
  if (status === 'archived') return 'info';
  if (status === 'config_invalid') return 'danger';
  if (status === 'dependency_missing') return 'warning';
  if (status === 'migration_failed') return 'danger';
  return 'info';
}
</script>

