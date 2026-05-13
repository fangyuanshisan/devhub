<template>
  <section class="plugin-page" data-testid="plugin-hooks-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">运行排障</div>
        <h2>Hook 排障</h2>
        <p class="muted">聚合 Hook 警告/异常插件与最近错误摘要；具体执行记录请进入单插件 Hooks Tab。</p>
      </div>
    </div>

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-table v-loading="loading" :data="rows" border stripe :empty-text="'暂无 Hook 异常'" data-testid="plugin-hooks-table">
      <el-table-column prop="name" label="插件" min-width="220">
        <template #default="{ row }">
          <div class="plugin-title compact-title">
            <strong>{{ row.name }}</strong>
            <span class="mono">{{ row.code }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="Hook 状态" width="140">
        <template #default="{ row }">
          <el-tag :type="hookTagType(row)" effect="plain">{{ row.health?.hook_status || row.health?.status || '-' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最近错误" min-width="260">
        <template #default="{ row }">
          <span class="muted">{{ row.health?.last_hook_error || row.health?.recent_error || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPlugin(row, 'hooks')">进入 Hooks</el-button>
        </template>
      </el-table-column>
    </el-table>

    <PluginDetailDrawer v-model="drawerVisible" :plugin="drawerPlugin" :plugins="items" :initial-tab="drawerTab" @refresh="load" @open-plugin="openByCode" />
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import PluginDetailDrawer from '@/components/plugin/PluginDetailDrawer.vue';
import { usePluginData } from './usePluginData';

const drawerVisible = ref(false);
const drawerPlugin = ref(null);
const drawerTab = ref('hooks');

const { items, loading, error, load } = usePluginData();
onMounted(load);

const rows = computed(() => {
  const list = items.value || [];
  return list.filter((p) => ['hook_warning', 'hook_error'].includes(p.health?.status) || ['hook_warning', 'hook_error'].includes(p.health?.hook_status));
});

function hookTagType(row) {
  const st = row?.health?.hook_status || row?.health?.status;
  if (st === 'hook_error') return 'danger';
  if (st === 'hook_warning') return 'warning';
  return 'info';
}

function openPlugin(row, tab) {
  drawerPlugin.value = row;
  drawerTab.value = tab || 'overview';
  drawerVisible.value = true;
}

function openByCode(code, tab = 'hooks') {
  const target = (items.value || []).find((p) => (p.code || p.plugin_code) === code);
  if (!target) return;
  openPlugin(target, tab);
}
</script>

