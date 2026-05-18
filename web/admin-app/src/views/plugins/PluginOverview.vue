<template>
  <section class="plugin-page" data-testid="plugin-overview-page">
    <div class="plugin-page-header overview-workbench-header">
      <div>
        <div class="eyebrow">插件运营</div>
        <h2>插件概览</h2>
        <p class="muted">插件后台按功能分页的全局视角入口；更多单插件细节请打开插件详情抽屉。</p>
      </div>
      <div class="primary-actions compact-actions">
        <el-button type="primary" plain @click="go('/plugins/list')">插件列表</el-button>
        <el-button type="primary" plain @click="go('/plugins/install')">安装升级</el-button>
        <el-dropdown trigger="click" @command="goQuick">
          <el-button plain>更多治理</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="/plugins/config">配置中心</el-dropdown-item>
              <el-dropdown-item command="/plugins/navigation">前端挂载</el-dropdown-item>
              <el-dropdown-item command="/plugins/content">内容治理</el-dropdown-item>
              <el-dropdown-item command="/plugins/dependencies">依赖兼容</el-dropdown-item>
              <el-dropdown-item command="/plugins/hooks">Hook 排障</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <el-alert
      v-if="error"
      type="error"
      show-icon
      :closable="false"
      :title="error"
      class="mb"
    />

    <div class="stats-grid compact overview-stats" data-testid="plugin-overview-stats">
      <button class="stat-card stat-button" type="button" @click="go('/plugins/list')">
        <div class="stat-k">插件总数</div>
        <div class="stat-v">{{ stats.total }}</div>
      </button>
      <button class="stat-card stat-button" type="button" @click="go('/plugins/list', { status: 'enabled' })">
        <div class="stat-k">已启用</div>
        <div class="stat-v">{{ stats.enabled }}</div>
      </button>
      <button class="stat-card stat-button" type="button" @click="go('/plugins/list', { status: 'disabled' })">
        <div class="stat-k">已禁用</div>
        <div class="stat-v">{{ stats.disabled }}</div>
      </button>
      <button class="stat-card stat-button" type="button" @click="go('/plugins/list', { status: 'archived' })">
        <div class="stat-k">已归档</div>
        <div class="stat-v">{{ stats.archived }}</div>
      </button>
      <button class="stat-card stat-card-danger stat-button" type="button" @click="go('/plugins/list', { health: 'error' })">
        <div class="stat-k">异常插件</div>
        <div class="stat-v">{{ stats.abnormal }}</div>
      </button>
    </div>

    <el-collapse v-model="panels" class="health-collapse overview-collapse">
      <el-collapse-item name="recent">
        <template #title>
          <div class="collapse-title">
            <strong>最近异常</strong>
            <span class="muted">只展示需要处理的插件</span>
          </div>
        </template>
        <el-table v-loading="loading" :data="recentAbnormal" border stripe :empty-text="'暂无异常'" data-testid="plugin-overview-recent-abnormal">
          <el-table-column prop="name" label="插件" min-width="180">
            <template #default="{ row }">
              <div class="plugin-title compact-title">
                <strong>{{ row.name }}</strong>
                <span class="mono">{{ row.code }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="140">
            <template #default="{ row }">
              <el-tag :type="healthType(row.health?.status)" effect="plain">{{ pluginHealthLabel(row.health?.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="原因" min-width="260">
            <template #default="{ row }">
              <span class="muted">{{ row.health?.recent_error || row.health?.status_reason || row.status_reason || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button link type="primary" @click="openPlugin(row, 'runtime')">查看详情</el-button>
              <el-button link type="primary" @click="openPlugin(row, suggestedTab(row))">去处理</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>

      <el-collapse-item name="health">
        <template #title>
          <div class="collapse-title">
            <strong>健康摘要</strong>
            <span class="muted">默认收起，需要排障时展开</span>
          </div>
        </template>
        <div class="health-grid" data-testid="plugin-overview-health">
          <button v-for="card in healthCards" :key="card.key" class="stat-card stat-button health-card" type="button" @click="go('/plugins/list', { health: card.key })">
            <div class="stat-k">{{ card.label }}</div>
            <div class="stat-v">{{ card.value }}</div>
            <div class="stat-sub">{{ card.tip }}</div>
          </button>
        </div>
      </el-collapse-item>
    </el-collapse>

    <PluginDetailDrawer v-model="drawerVisible" :plugin="drawerPlugin" :plugins="items" :initial-tab="drawerTab" @refresh="load" @open-plugin="openByCode" />
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import PluginDetailDrawer from '@/components/plugin/PluginDetailDrawer.vue';
import { pluginHealthLabel } from '@/i18n/formatters';
import { usePluginData } from './usePluginData';

const router = useRouter();
const panels = ref(['recent']);
const drawerVisible = ref(false);
const drawerPlugin = ref(null);
const drawerTab = ref('overview');

const { items, healthSummary, loading, error, load } = usePluginData();

onMounted(load);

const stats = computed(() => {
  const list = items.value || [];
  const enabled = list.filter((p) => p.status === 'enabled').length;
  const disabled = list.filter((p) => p.status === 'disabled').length;
  const archived = list.filter((p) => p.status === 'archived').length;
  const abnormal = list.filter((p) => p.health?.status === 'error' || p.status === 'migration_failed' || p.status === 'dependency_missing' || p.status === 'config_invalid').length;
  return { total: list.length, enabled, disabled, archived, abnormal };
});

const recentAbnormal = computed(() => {
  const list = items.value || [];
  const abnormal = list.filter((p) => p.health?.status && p.health.status !== 'healthy');
  return abnormal.slice(0, 8);
});

function go(path, query = null) {
  if (query) router.push({ path, query });
  else router.push(path);
}

function goQuick(path) {
  if (!path) return;
  go(path);
}

function openPlugin(row, tab) {
  drawerPlugin.value = row;
  drawerTab.value = tab || 'overview';
  drawerVisible.value = true;
}

function openByCode(code, tab = 'overview') {
  const target = (items.value || []).find((p) => (p.code || p.plugin_code) === code);
  if (!target) return;
  openPlugin(target, tab);
}

function healthType(status) {
  if (!status || status === 'healthy') return 'success';
  if (status === 'hook_warning') return 'warning';
  if (status === 'migration_pending') return 'warning';
  if (status === 'dependency_missing') return 'warning';
  if (status === 'config_invalid') return 'danger';
  if (status === 'hook_error') return 'danger';
  if (status === 'error') return 'danger';
  if (status === 'disabled') return 'info';
  if (status === 'archived') return 'info';
  return 'info';
}

function suggestedTab(row) {
  const st = row?.health?.status;
  if (st === 'dependency_missing') return 'dependencies';
  if (st === 'config_invalid') return 'config';
  if (st === 'migration_pending' || st === 'error' || row?.status === 'migration_failed') return 'migrations';
  if (st === 'hook_warning' || st === 'hook_error') return 'hooks';
  return 'overview';
}

const healthCards = computed(() => {
  const summary = healthSummary.value || {};
  return [
    { key: 'healthy', label: '健康', value: summary.healthy || 0, tip: '无需处理' },
    { key: 'migration_pending', label: '迁移待处理', value: summary.migration_pending || 0, tip: '进入迁移 Tab 处理' },
    { key: 'config_invalid', label: '配置无效', value: summary.config_invalid || 0, tip: '修复配置后重试' },
    { key: 'dependency_missing', label: '依赖缺失', value: summary.dependency_missing || 0, tip: '先满足依赖' },
    { key: 'hook_warning', label: 'Hook 警告', value: summary.hook_warning || 0, tip: '查看最近失败记录' },
    { key: 'hook_error', label: 'Hook 异常', value: summary.hook_error || 0, tip: '优先处理阻断项' },
    { key: 'archived', label: '已归档', value: summary.archived || 0, tip: '入口关闭，历史保留' },
  ];
});
</script>
