<template>
  <section class="plugin-page" data-testid="plugin-permissions-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">安装与治理</div>
        <h2>权限矩阵</h2>
        <p class="muted">聚合插件权限入口与缺失风险定位；详细矩阵与引用关系请打开单插件 Permissions Tab。</p>
      </div>
    </div>

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-table v-loading="loading" :data="rows" border stripe :empty-text="'暂无权限声明'" data-testid="plugin-permissions-table">
      <el-table-column prop="name" label="插件" min-width="220">
        <template #default="{ row }">
          <div class="plugin-title compact-title">
            <strong>{{ row.name }}</strong>
            <span class="mono">{{ row.code }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="权限数" width="120">
        <template #default="{ row }">
          <span class="mono">{{ (row.permissions || []).length }}</span>
        </template>
      </el-table-column>
      <el-table-column label="说明" min-width="260">
        <template #default="{ row }">
          <span class="muted">{{ row.health?.status_reason || row.status_reason || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPlugin(row, 'permissions')">查看矩阵</el-button>
          <el-button link type="primary" @click="openPlugin(row, 'readiness')">启用检查</el-button>
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
const drawerTab = ref('permissions');

const { items, loading, error, load } = usePluginData();
onMounted(load);

const rows = computed(() => (items.value || []).filter((p) => Array.isArray(p.permissions) && p.permissions.length));

function openPlugin(row, tab) {
  drawerPlugin.value = row;
  drawerTab.value = tab || 'overview';
  drawerVisible.value = true;
}

function openByCode(code, tab = 'permissions') {
  const target = (items.value || []).find((p) => (p.code || p.plugin_code) === code);
  if (!target) return;
  openPlugin(target, tab);
}
</script>
