<template>
  <section class="plugin-page" data-testid="plugin-config-hub-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">插件运营</div>
        <h2>配置中心</h2>
        <p class="muted">聚合全局插件配置入口；子站配置请前往“子站管理 / 插件配置”。</p>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" class="mb" title="说明：此页只聚合入口与诊断，不重复实现子站完整管理能力。" />
    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-table v-loading="loading" :data="rows" border stripe :empty-text="'暂无插件'" data-testid="plugin-config-table">
      <el-table-column prop="name" label="插件" min-width="220">
        <template #default="{ row }">
          <div class="plugin-title compact-title">
            <strong>{{ row.name }}</strong>
            <span class="mono">{{ row.code }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="配置模型" width="120">
        <template #default="{ row }">
          <el-tag :type="hasSchema(row) ? 'success' : 'info'" effect="plain">{{ hasSchema(row) ? '有' : '无' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="配置状态" width="140">
        <template #default="{ row }">
          <el-tag :type="row.health?.config_status === 'invalid' || row.status === 'config_invalid' ? 'danger' : 'success'" effect="plain">
            {{ configStatusLabel(row) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPlugin(row, 'config')">打开配置</el-button>
          <el-button link type="primary" @click="router.push('/communities')">子站配置</el-button>
        </template>
      </el-table-column>
    </el-table>

    <PluginDetailDrawer v-model="drawerVisible" :plugin="drawerPlugin" :plugins="items" :initial-tab="drawerTab" @refresh="load" @open-plugin="openByCode" />
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import PluginDetailDrawer from '@/components/plugin/PluginDetailDrawer.vue';
import { usePluginData } from './usePluginData';
import { genericStatusLabel } from '@/i18n/formatters';

const router = useRouter();
const drawerVisible = ref(false);
const drawerPlugin = ref(null);
const drawerTab = ref('config');
const { items, loading, error, load } = usePluginData();

onMounted(load);

const rows = computed(() => (items.value || []).slice().sort((a, b) => String(a.code).localeCompare(String(b.code))));

function hasSchema(row) {
  return row && row.config_schema && Object.keys(row.config_schema || {}).length > 0;
}

function configStatusLabel(row) {
  return genericStatusLabel(row?.health?.config_status || (row?.status === 'config_invalid' ? 'invalid' : 'valid'));
}

function openPlugin(row, tab) {
  drawerPlugin.value = row;
  drawerTab.value = tab || 'overview';
  drawerVisible.value = true;
}

function openByCode(code, tab = 'config') {
  const target = (items.value || []).find((p) => (p.code || p.plugin_code) === code);
  if (!target) return;
  openPlugin(target, tab);
}
</script>
