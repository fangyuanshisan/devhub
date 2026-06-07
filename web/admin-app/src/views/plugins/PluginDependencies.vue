<template>
  <section class="plugin-page" data-testid="plugin-dependencies-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">安装与治理</div>
        <h2>依赖兼容</h2>
        <p class="muted">依赖缺失、版本不满足、循环依赖与 Core 兼容状态的聚合视角；细节请打开单插件详情的 Dependencies Tab。</p>
      </div>
      <div class="primary-actions">
        <el-button type="primary" plain @click="router.push('/plugins/install')">去安装升级</el-button>
      </div>
    </div>

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-table v-loading="loading" :data="rows" border stripe :empty-text="'暂无依赖异常'" data-testid="plugin-dependencies-table">
      <el-table-column prop="name" label="插件" min-width="220">
        <template #default="{ row }">
          <div class="plugin-title compact-title">
            <strong>{{ row.name }}</strong>
            <span class="mono">{{ row.code }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="依赖状态" width="160">
        <template #default="{ row }">
          <el-tag :type="dependencyTagType(row)" effect="plain">{{ dependencyStatusLabel(row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="原因" min-width="260">
        <template #default="{ row }">
          <span class="muted">{{ row.health?.recent_error || row.health?.status_reason || row.status_reason || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPlugin(row, 'dependencies')">查看依赖</el-button>
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
import { pluginStatusText } from '@/modules/plugins/statusText';

const router = useRouter();
const drawerVisible = ref(false);
const drawerPlugin = ref(null);
const drawerTab = ref('dependencies');

const { items, loading, error, load } = usePluginData();
onMounted(load);

const rows = computed(() => {
  const list = items.value || [];
  const abnormal = list.filter((p) => p.status === 'dependency_missing' || p.health?.status === 'dependency_missing' || p.health?.dependency_status === 'missing');
  return abnormal.slice().sort((a, b) => String(a.code).localeCompare(String(b.code)));
});

function dependencyTagType(row) {
  const st = row?.health?.dependency_status || row?.health?.status || row?.status;
  if (st === 'missing' || st === 'dependency_missing') return 'danger';
  return 'warning';
}

function dependencyStatusLabel(row) {
  return pluginStatusText(row?.health?.dependency_status || row?.health?.status || row?.status);
}

function openPlugin(row, tab) {
  drawerPlugin.value = row;
  drawerTab.value = tab || 'overview';
  drawerVisible.value = true;
}

function openByCode(code, tab = 'dependencies') {
  const target = (items.value || []).find((p) => (p.code || p.plugin_code) === code);
  if (!target) return;
  openPlugin(target, tab);
}
</script>
