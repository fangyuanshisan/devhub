<template>
  <section class="plugin-page" data-testid="plugin-navigation-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">运行排障</div>
        <h2>前台入口</h2>
        <p class="muted">聚合前台菜单声明与可见性诊断入口；详细预览与原因请打开单插件 Menus Tab。</p>
      </div>
    </div>

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-table v-loading="loading" :data="rows" border stripe :empty-text="'暂无菜单声明'" data-testid="plugin-navigation-table">
      <el-table-column prop="name" label="插件" min-width="220">
        <template #default="{ row }">
          <div class="plugin-title compact-title">
            <strong>{{ row.name }}</strong>
            <span class="mono">{{ row.code }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="前台菜单数" width="120">
        <template #default="{ row }">
          <span class="mono">{{ frontendMenusCount(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="说明" min-width="260">
        <template #default="{ row }">
          <span class="muted">{{ row.health?.status_reason || row.status_reason || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPlugin(row, 'menus')">菜单预览</el-button>
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
const drawerTab = ref('menus');

const { items, loading, error, load } = usePluginData();
onMounted(load);

const rows = computed(() => (items.value || []).filter((p) => Array.isArray(p.menus) && p.menus.length));

function frontendMenusCount(row) {
  const menus = row?.menus || [];
  return menus.filter((m) => String(m.area || '').toLowerCase() === 'frontend').length;
}

function openPlugin(row, tab) {
  drawerPlugin.value = row;
  drawerTab.value = tab || 'overview';
  drawerVisible.value = true;
}

function openByCode(code, tab = 'menus') {
  const target = (items.value || []).find((p) => (p.code || p.plugin_code) === code);
  if (!target) return;
  openPlugin(target, tab);
}
</script>

