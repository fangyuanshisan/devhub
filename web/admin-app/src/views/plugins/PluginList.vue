<template>
  <section class="plugin-page" data-testid="admin-plugins-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">{{ t('plugin.pageEyebrow') }}</div>
        <h2>{{ t('plugin.title') }}</h2>
        <p>{{ t('plugin.description') }}</p>
      </div>
      <div class="primary-actions">
        <el-button type="primary" plain data-testid="plugin-manifest-validate" @click="goInstall">{{ t('plugin.ops.validateManifest') }}</el-button>
        <el-button type="primary" plain data-testid="plugin-manifest-dry-run" @click="goInstall">{{ t('plugin.ops.dryRun') }}</el-button>
        <el-button type="success" plain data-testid="plugin-manifest-install" @click="goInstall">{{ t('plugin.ops.install') }}</el-button>
      </div>
    </div>

    <el-tabs :model-value="statusTab" class="page-tabs" data-testid="plugin-status-tabs" @tab-change="applyStatusTab">
      <el-tab-pane name="all" label="全部插件" />
      <el-tab-pane name="enabled" label="已启用" />
      <el-tab-pane name="disabled" label="已禁用" />
      <el-tab-pane name="archived" label="回收站" />
      <el-tab-pane name="abnormal" label="异常插件" />
    </el-tabs>

    <div class="stats-grid" data-testid="plugin-stats">
      <button class="stat-card stat-button" type="button" @click="applyStatusTab('all')">
        <div class="stat-k">{{ t('plugin.stats.total') }}</div>
        <div class="stat-v">{{ stats.total }}</div>
      </button>
      <button class="stat-card stat-button" type="button" @click="applyStatusTab('enabled')">
        <div class="stat-k">{{ t('plugin.stats.enabled') }}</div>
        <div class="stat-v">{{ stats.enabled }}</div>
      </button>
      <button class="stat-card stat-button" type="button" @click="applyStatusTab('disabled')">
        <div class="stat-k">{{ t('plugin.stats.disabled') }}</div>
        <div class="stat-v">{{ stats.disabled }}</div>
      </button>
      <button class="stat-card stat-button" type="button" @click="applyStatusTab('archived')">
        <div class="stat-k">{{ t('plugin.stats.archived') }}</div>
        <div class="stat-v">{{ stats.archived }}</div>
      </button>
      <button class="stat-card stat-card-danger stat-button" type="button" @click="applyStatusTab('abnormal')">
        <div class="stat-k">{{ t('plugin.stats.abnormal') }}</div>
        <div class="stat-v">{{ stats.abnormal }}</div>
      </button>
    </div>

    <el-collapse v-model="healthPanels" class="health-collapse">
      <el-collapse-item name="health">
        <template #title>
          <strong>{{ t('plugin.healthOverview') }}</strong>
        </template>
        <div class="health-grid" data-testid="plugin-health-summary">
          <button v-for="card in healthCards" :key="card.key" class="stat-card stat-button health-card" type="button" @click="applyHealth(card.key)">
            <div class="stat-k">{{ card.label }}</div>
            <div class="stat-v">{{ card.value }}</div>
            <div class="stat-sub">{{ card.tip }}</div>
          </button>
        </div>
      </el-collapse-item>
    </el-collapse>

    <PluginErrorAlert v-if="error" class="mb" :message="error" data-testid="plugin-list-error" />

    <PluginFilterBar :title="t('plugin.filters.title')" :tip="t('plugin.filters.tip')" testid="plugin-filter-panel">
      <el-input v-model="filters.q" data-testid="plugin-search" :placeholder="t('plugin.filters.searchPlaceholder')" clearable />
      <el-select v-model="filters.status" :placeholder="t('plugin.filters.status')" clearable>
        <el-option :label="t('common.all')" value="all" />
        <el-option :label="pluginStatusLabel('enabled')" value="enabled" />
        <el-option :label="pluginStatusLabel('disabled')" value="disabled" />
        <el-option :label="pluginStatusLabel('archived')" value="archived" />
      </el-select>
      <el-select v-model="filters.health" :placeholder="t('plugin.filters.health')" clearable>
        <el-option :label="t('common.all')" value="all" />
        <el-option v-for="card in healthCards" :key="card.key" :label="card.label" :value="card.key" />
      </el-select>
      <el-select v-model="filters.contentType" :placeholder="t('plugin.contentType')" clearable filterable>
        <el-option v-for="ct in allContentTypes" :key="ct" :label="ct" :value="ct" />
      </el-select>
      <el-select v-model="filters.system" :placeholder="t('plugin.system')" clearable>
        <el-option :label="t('common.all')" value="all" />
        <el-option :label="t('plugin.filters.onlySystem')" value="yes" />
        <el-option :label="t('plugin.filters.nonSystem')" value="no" />
      </el-select>
      <el-select v-model="filters.hasSchema" :placeholder="t('plugin.config.schema')" clearable>
        <el-option :label="t('common.all')" value="all" />
        <el-option :label="t('plugin.filters.hasSchema')" value="yes" />
        <el-option :label="t('plugin.filters.noSchema')" value="no" />
      </el-select>
      <el-button data-testid="plugin-filter-reset" @click="resetFilters">{{ t('common.reset') }}</el-button>
      <el-button type="primary" data-testid="plugin-filter-refresh" @click="load">{{ t('common.refresh') }}</el-button>
    </PluginFilterBar>

    <PluginEmptyState v-if="!loading && !error && !filteredItems.length" testid="plugin-empty-state" description="暂无符合条件的插件" />

    <div class="batch-panel" data-testid="plugin-batch-panel">
      <span class="muted">{{ t('common.selected') }} {{ selectedRows.length }} {{ t('common.selectedItems') }}</span>
      <div class="batch-actions">
        <el-button type="warning" plain :disabled="!selectedRows.length" data-testid="plugin-bulk-archive" @click="openBulkDialog('archive')">{{ t('plugin.ops.bulkArchive') }}</el-button>
        <el-button type="success" plain :disabled="!selectedRows.length" data-testid="plugin-bulk-restore" @click="openBulkDialog('restore')">{{ t('plugin.ops.bulkRestore') }}</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="filteredItems" border stripe data-testid="plugin-table" @selection-change="onSelectionChange">
      <el-table-column type="selection" width="48" />
      <el-table-column :label="t('plugin.pluginColumn')" min-width="220">
        <template #default="{ row }">
          <div class="plugin-title">
            <strong>{{ row.name }}</strong>
            <span class="mono">{{ row.code }}</span>
          </div>
          <div class="tag-wrap">
            <el-tag v-if="row.is_system" type="info" effect="plain">{{ t('plugin.filters.onlySystem') }}</el-tag>
            <el-tag :type="statusTagType(row.status)" effect="plain">{{ pluginStatusLabel(row.status) }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="version" :label="t('plugin.version')" width="110" />
      <el-table-column :label="t('plugin.health')" min-width="200">
        <template #default="{ row }">
          <el-tag :type="healthType(row.health?.status)" effect="plain">{{ pluginHealthLabel(row.health?.status) }}</el-tag>
          <div class="muted">{{ row.health?.recent_error || row.health?.status_reason || row.status_reason || '-' }}</div>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.contentTypes')" min-width="220">
        <template #default="{ row }">
          <div class="tag-wrap">
            <el-tag v-for="ct in (row.content_types || [])" :key="ct" type="info" effect="plain">{{ ct }}</el-tag>
            <span v-if="!(row.content_types || []).length" class="muted">-</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.capabilitySummary')" min-width="230">
        <template #default="{ row }">
          <div class="metric-line">
            <el-tag type="info" effect="plain">{{ t('plugin.capability.permissions') }} {{ (row.permissions || []).length }}</el-tag>
            <el-tag type="info" effect="plain">{{ t('plugin.capability.menus') }} {{ (row.menus || []).length }}</el-tag>
            <el-tag :type="hasConfigSchema(row) ? 'success' : 'info'" effect="plain">{{ t('plugin.capability.schema') }} {{ hasConfigSchema(row) ? t('common.yes') : t('common.no') }}</el-tag>
            <el-tag :type="(row.hooks || []).length ? 'success' : 'info'" effect="plain">{{ t('plugin.capability.hooks') }} {{ (row.hooks || []).length }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.recentError')" min-width="190">
        <template #default="{ row }">
          <span class="muted">{{ row.health?.recent_error || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.action')" fixed="right" width="190">
        <template #default="{ row }">
          <div class="row-actions">
            <el-button link type="primary" :data-testid="`plugin-detail-${row.code}`" @click="openPlugin(row)">{{ t('common.detail') }}</el-button>
            <el-button link type="primary" @click="openPlugin(row, 'config')">{{ t('plugin.config.title') }}</el-button>
            <el-dropdown trigger="click" @command="(command) => handlePluginCommand(row, command)">
              <el-button link type="info">{{ t('plugin.ops.more') }}</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="permissions">{{ t('plugin.capability.permissions') }}</el-dropdown-item>
                  <el-dropdown-item command="menus">{{ t('plugin.capability.menus') }}</el-dropdown-item>
                  <el-dropdown-item command="dependencies">{{ t('plugin.tabs.dependencies') }}</el-dropdown-item>
                  <el-dropdown-item command="hooks">{{ t('plugin.tabs.hooks') }}</el-dropdown-item>
                  <el-dropdown-item command="migrations">{{ t('plugin.tabs.migrations') }}</el-dropdown-item>
                  <el-dropdown-item command="runtime">{{ t('plugin.tabs.runtime') }}</el-dropdown-item>
                  <el-dropdown-item command="audit">{{ t('plugin.tabs.audit') }}</el-dropdown-item>
                  <el-dropdown-item v-if="row.status !== 'enabled' && row.status !== 'archived'" command="enable" :data-testid="`plugin-enable-${row.code}`">{{ t('common.enable') }}</el-dropdown-item>
                  <el-dropdown-item v-if="row.status === 'enabled'" command="disable" :data-testid="`plugin-disable-${row.code}`">{{ t('common.disable') }}</el-dropdown-item>
                  <el-dropdown-item v-if="row.status !== 'archived'" command="archive" :data-testid="`plugin-archive-${row.code}`" divided>{{ t('common.archive') }}</el-dropdown-item>
                  <el-dropdown-item v-if="row.status === 'archived'" command="restore" :data-testid="`plugin-restore-${row.code}`" divided>{{ t('common.restore') }}</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="bulkDialogVisible" append-to-body destroy-on-close size="780px" :with-header="false" class="plugin-action-drawer">
      <section class="action-panel in-drawer" data-testid="plugin-bulk-result-panel">
        <div class="action-panel-header">
          <div>
            <h3>{{ bulkMode === 'archive' ? t('plugin.ops.bulkArchive') : t('plugin.ops.bulkRestore') }}</h3>
            <p>{{ bulkStep === 'preview' ? t('plugin.ops.bulkPreviewTip') : t('plugin.ops.bulkResultTip') }}</p>
          </div>
          <div class="action-panel-tools">
            <el-button data-testid="plugin-result-close" @click="bulkDialogVisible = false">{{ t('common.close') }}</el-button>
            <el-button v-if="bulkStep === 'result'" type="success" plain data-testid="plugin-result-audit" @click="openAuditLogs">{{ t('plugin.ops.viewAuditLogs') }}</el-button>
            <el-button v-if="bulkStep === 'preview'" :loading="bulkLoading" type="warning" data-testid="plugin-bulk-confirm" @click="confirmBulkAction">{{ bulkMode === 'archive' ? t('plugin.ops.confirmBulkArchive') : t('plugin.ops.confirmBulkRestore') }}</el-button>
          </div>
        </div>
        <template v-if="bulkStep === 'preview'">
          <el-alert :title="t('plugin.ops.bulkConfirmTip')" type="warning" show-icon :closable="false" class="mb" />
          <el-table v-loading="bulkLoading" :data="bulkPreviewRows" border stripe>
            <el-table-column prop="name" :label="t('plugin.name')" min-width="160" />
            <el-table-column prop="code" :label="t('plugin.code')" min-width="150" />
            <el-table-column :label="t('plugin.status')" width="110">
              <template #default="{ row }">{{ pluginStatusLabel(row.status) }}</template>
            </el-table-column>
            <el-table-column prop="contents" :label="t('plugin.impact.existingContents')" width="120" />
            <el-table-column prop="communities" :label="t('plugin.impact.enabledCommunities')" width="130" />
            <el-table-column prop="categories" :label="t('plugin.impact.blockedCategories')" width="140" />
            <el-table-column prop="migrations" :label="t('plugin.impact.pendingMigrations')" width="120" />
            <el-table-column prop="hookErrors" :label="t('plugin.impact.recentHookErrors')" width="130" />
          </el-table>
        </template>
        <template v-else>
          <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
            <el-descriptions-item :label="t('plugin.ops.succeededCount')">{{ (bulkResult.succeeded || []).length }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.ops.failedCount')">{{ (bulkResult.failed || []).length }}</el-descriptions-item>
          </el-descriptions>
          <div class="result-grid">
            <div class="result-box">
              <h4>{{ t('plugin.ops.succeeded') }}</h4>
              <el-table :data="bulkResult.succeeded || []" border stripe :empty-text="t('common.none')">
                <el-table-column prop="plugin_code" :label="t('plugin.code')" min-width="160" />
                <el-table-column prop="status" :label="t('plugin.status')" width="120" />
              </el-table>
            </div>
            <div class="result-box">
              <h4>{{ t('plugin.ops.failed') }}</h4>
              <el-table :data="bulkResult.failed || []" border stripe :empty-text="t('common.none')">
                <el-table-column prop="plugin_code" :label="t('plugin.code')" min-width="160" />
                <el-table-column prop="error" :label="t('plugin.ops.error')" min-width="240" />
              </el-table>
            </div>
          </div>
        </template>
      </section>
    </el-drawer>

    <PluginDetailDrawer v-model="drawerVisible" :plugin="drawerPlugin" :plugins="items" :initial-tab="drawerTab" @refresh="load" @open-plugin="openByCode" />
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import { archivePlugin, bulkArchivePlugins, bulkRestorePlugins, disablePlugin, enablePlugin, pluginImpact, plugins, restorePlugin } from '@/api/admin';
import PluginDetailDrawer from '@/components/plugin/PluginDetailDrawer.vue';
import { t } from '@/i18n';
import { pluginHealthLabel, pluginStatusLabel } from '@/i18n/formatters';
import { usePluginData } from './usePluginData';
import PluginEmptyState from './components/PluginEmptyState.vue';
import PluginErrorAlert from './components/PluginErrorAlert.vue';
import PluginFilterBar from './components/PluginFilterBar.vue';
import { confirmDanger, confirmInfo } from './components/useDangerConfirm';

const router = useRouter();
const route = useRoute();

const { items, healthSummary, loading, error, load } = usePluginData();

const filters = reactive({
  q: '',
  status: 'all',
  health: 'all',
  contentType: '',
  system: 'all',
  hasSchema: 'all',
});

const selectedRows = ref([]);
const healthPanels = ref(['health']);
const drawerVisible = ref(false);
const drawerPlugin = ref(null);
const drawerTab = ref('overview');

const bulkDialogVisible = ref(false);
const bulkMode = ref('archive');
const bulkStep = ref('preview');
const bulkLoading = ref(false);
const bulkPreviewRows = ref([]);
const bulkResult = ref({ succeeded: [], failed: [] });
const bulkCodes = ref([]);
const resultAuditQuery = ref(null);

onMounted(load);

watch(
  () => route.query,
  (q) => {
    if (q?.status) filters.status = String(q.status);
    if (q?.health) filters.health = String(q.health);
  },
  { immediate: true },
);

const stats = computed(() => {
  const list = items.value || [];
  const enabled = list.filter((p) => p.status === 'enabled').length;
  const disabled = list.filter((p) => p.status === 'disabled').length;
  const archived = list.filter((p) => p.status === 'archived').length;
  const abnormal = list.filter((p) => p.health?.status === 'error' || p.health?.status === 'hook_error').length;
  return { total: list.length, enabled, disabled, archived, abnormal };
});

const allContentTypes = computed(() => {
  const set = new Set();
  (items.value || []).forEach((p) => (p.content_types || []).forEach((ct) => set.add(ct)));
  return Array.from(set);
});

const healthCards = computed(() => {
  const summary = healthSummary.value || {};
  const tip = t('plugin.healthCard.tip');
  return [
    { key: 'healthy', label: pluginHealthLabel('healthy'), value: summary.healthy || 0, tip },
    { key: 'warning', label: pluginHealthLabel('warning'), value: summary.warning || 0, tip },
    { key: 'error', label: pluginHealthLabel('error'), value: summary.error || 0, tip },
    { key: 'disabled', label: pluginHealthLabel('disabled'), value: summary.disabled || 0, tip },
    { key: 'archived', label: pluginHealthLabel('archived'), value: summary.archived || 0, tip },
    { key: 'migration_pending', label: pluginHealthLabel('migration_pending'), value: summary.migration_pending || 0, tip },
    { key: 'config_invalid', label: pluginHealthLabel('config_invalid'), value: summary.config_invalid || 0, tip },
    { key: 'dependency_missing', label: pluginHealthLabel('dependency_missing'), value: summary.dependency_missing || 0, tip },
    { key: 'hook_error', label: pluginHealthLabel('hook_error'), value: summary.hook_error || 0, tip },
  ];
});

const statusTab = computed(() => {
  if (filters.health === 'error') return 'abnormal';
  return filters.status || 'all';
});

const filteredItems = computed(() => {
  const q = String(filters.q || '').trim().toLowerCase();
  return (items.value || []).filter((p) => {
    if (filters.status && filters.status !== 'all' && p.status !== filters.status) return false;
    if (filters.health && filters.health !== 'all') {
      if (filters.health === 'error') {
        if (!(p.health?.status === 'error' || p.health?.status === 'hook_error')) return false;
      } else if (p.health?.status !== filters.health) return false;
    }
    if (filters.contentType) {
      const types = p.content_types || [];
      if (!types.includes(filters.contentType)) return false;
    }
    if (filters.system && filters.system !== 'all') {
      if (filters.system === 'yes' && !p.is_system) return false;
      if (filters.system === 'no' && p.is_system) return false;
    }
    if (filters.hasSchema && filters.hasSchema !== 'all') {
      const has = hasConfigSchema(p);
      if (filters.hasSchema === 'yes' && !has) return false;
      if (filters.hasSchema === 'no' && has) return false;
    }
    if (q) {
      const hay = `${p.name || ''} ${p.code || ''}`.toLowerCase();
      if (!hay.includes(q)) return false;
    }
    return true;
  });
});

function applyStatusTab(value) {
  if (!value) return;
  if (value === 'abnormal') {
    filters.status = 'all';
    filters.health = 'error';
  } else {
    filters.status = value;
    if (filters.health === 'error') filters.health = 'all';
  }
  const next = { ...route.query };
  if (filters.status && filters.status !== 'all') next.status = filters.status;
  else delete next.status;
  if (filters.health && filters.health !== 'all') next.health = filters.health;
  else delete next.health;
  router.replace({ query: next });
}

function applyHealth(value) {
  if (!value) return;
  filters.health = value;
  const next = { ...route.query };
  if (filters.status && filters.status !== 'all') next.status = filters.status;
  else delete next.status;
  if (filters.health && filters.health !== 'all') next.health = filters.health;
  else delete next.health;
  router.replace({ query: next });
}

function hasConfigSchema(row) {
  return row && row.config_schema && Object.keys(row.config_schema || {}).length > 0;
}

function statusTagType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'warning';
  if (status === 'archived') return 'info';
  return 'info';
}

function healthType(status) {
  if (!status || status === 'healthy') return 'success';
  if (status === 'hook_warning') return 'warning';
  if (status === 'hook_error') return 'danger';
  if (status === 'dependency_missing') return 'warning';
  if (status === 'config_invalid') return 'danger';
  if (status === 'migration_pending') return 'warning';
  if (status === 'error') return 'danger';
  return 'info';
}

function resetFilters() {
  filters.q = '';
  filters.status = 'all';
  filters.health = 'all';
  filters.contentType = '';
  filters.system = 'all';
  filters.hasSchema = 'all';
}

function onSelectionChange(rows) {
  selectedRows.value = Array.isArray(rows) ? rows : [];
}

function openPlugin(row, tab = 'overview') {
  drawerPlugin.value = row;
  drawerTab.value = tab;
  drawerVisible.value = true;
}

function openByCode(code, tab = 'overview') {
  const target = (items.value || []).find((p) => (p.code || p.plugin_code) === code);
  if (!target) return;
  openPlugin(target, tab);
}

function goInstall() {
  router.push('/plugins/install');
}

async function impactLines(row) {
  const res = await pluginImpact(row.code).catch(() => null);
  const impact = res || {};
  const lines = [
    `${t('plugin.impact.existingContents')}: ${impact.existing_contents ?? '-'}`,
    `${t('plugin.impact.enabledCommunities')}: ${impact.enabled_communities ?? '-'}`,
    `${t('plugin.impact.blockedCategories')}: ${impact.blocked_categories ?? '-'}`,
    `${t('plugin.impact.pendingMigrations')}: ${impact.pending_migrations ?? '-'}`,
    `${t('plugin.impact.recentHookErrors')}: ${impact.recent_hook_errors ?? '-'}`,
  ];
  return lines;
}

async function setStatus(row, status) {
  if (status === 'disabled') {
    const lines = await impactLines(row);
    await confirmDanger(
      `${t('plugin.disableConfirmPrefix')}\n\n${lines.join('\n')}`,
      t('plugin.disableConfirmTitle'),
      { type: 'warning', confirmButtonText: t('plugin.confirmDisable'), cancelButtonText: t('common.cancel') },
    );
  } else {
    await confirmInfo(t('plugin.enableConfirmText'), t('plugin.enableConfirmTitle'), {
      confirmButtonText: t('plugin.confirmEnable'),
      cancelButtonText: t('common.cancel'),
    });
  }
  if (status === 'enabled') await enablePlugin(row.code);
  else await disablePlugin(row.code);
  ElMessage.success(t('plugin.statusUpdated'));
  await load();
}

async function archive(row) {
  const lines = await impactLines(row);
  await confirmDanger(
    `${t('plugin.archiveConfirmPrefix')}\n\n${lines.join('\n')}`,
    t('plugin.archiveConfirmTitle'),
    { type: 'warning', confirmButtonText: t('plugin.confirmArchive'), cancelButtonText: t('common.cancel') },
  );
  await archivePlugin(row.code);
  ElMessage.success(t('plugin.archivedDone'));
  await load();
}

async function restore(row) {
  await confirmInfo(t('plugin.restoreConfirmText'), t('plugin.restoreConfirmTitle'), {
    confirmButtonText: t('plugin.confirmRestore'),
    cancelButtonText: t('common.cancel'),
  });
  await restorePlugin(row.code);
  ElMessage.success(t('plugin.restoredDone'));
  await load();
}

async function handlePluginCommand(row, command) {
  if (command === 'enable') return setStatus(row, 'enabled');
  if (command === 'disable') return setStatus(row, 'disabled');
  if (command === 'archive') return archive(row);
  if (command === 'restore') return restore(row);
  return openPlugin(row, command);
}

function openBulkDialog(mode) {
  bulkMode.value = mode;
  bulkStep.value = 'preview';
  bulkDialogVisible.value = true;
  bulkCodes.value = (selectedRows.value || []).map((r) => r.code);
  bulkPreviewRows.value = [];
  bulkResult.value = { succeeded: [], failed: [] };
  resultAuditQuery.value = null;
  confirmBulkPreview();
}

async function confirmBulkPreview() {
  bulkLoading.value = true;
  try {
    if (bulkMode.value === 'archive') {
      const res = await bulkArchivePlugins({ codes: bulkCodes.value, preview: true });
      bulkPreviewRows.value = res.items || [];
    } else {
      const res = await bulkRestorePlugins({ codes: bulkCodes.value, preview: true });
      bulkPreviewRows.value = res.items || [];
    }
  } finally {
    bulkLoading.value = false;
  }
}

async function confirmBulkAction() {
  bulkLoading.value = true;
  try {
    let res;
    if (bulkMode.value === 'archive') res = await bulkArchivePlugins({ codes: bulkCodes.value });
    else res = await bulkRestorePlugins({ codes: bulkCodes.value });
    bulkResult.value = res || { succeeded: [], failed: [] };
    resultAuditQuery.value = res?.audit_query || null;
    bulkStep.value = 'result';
    await load();
  } finally {
    bulkLoading.value = false;
  }
}

function openAuditLogs() {
  router.push({ path: '/audit-logs', query: resultAuditQuery.value || {} });
}
</script>
