<template>
  <section class="page-card" data-testid="admin-plugins-page">
    <el-alert
      :title="t('plugin.description')"
      type="info"
      show-icon
      :closable="false"
      class="intro-alert"
    />

    <div class="stats-grid" data-testid="plugin-stats">
      <div class="stat-card">
        <div class="stat-k">{{ t('plugin.stats.total') }}</div>
        <div class="stat-v">{{ stats.total }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">{{ t('plugin.stats.enabled') }}</div>
        <div class="stat-v">{{ stats.enabled }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">{{ t('plugin.stats.disabled') }}</div>
        <div class="stat-v">{{ stats.disabled }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">{{ t('plugin.stats.system') }}</div>
        <div class="stat-v">{{ stats.system }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">{{ t('plugin.stats.hasSchema') }}</div>
        <div class="stat-v">{{ stats.hasSchema }}</div>
      </div>
    </div>

    <div class="toolbar">
      <div>
        <h2>{{ t('plugin.title') }}</h2>
        <p>{{ t('plugin.coreNote') }}</p>
      </div>
      <div class="tool-actions">
        <el-input v-model="filters.q" data-testid="plugin-search" :placeholder="t('plugin.filters.searchPlaceholder')" clearable style="width: 220px" />
        <el-select v-model="filters.status" :placeholder="t('plugin.filters.status')" clearable style="width: 140px">
          <el-option :label="t('common.all')" value="all" />
          <el-option :label="pluginStatusLabel('enabled')" value="enabled" />
          <el-option :label="pluginStatusLabel('disabled')" value="disabled" />
        </el-select>
        <el-select v-model="filters.contentType" :placeholder="t('plugin.contentType')" clearable filterable style="width: 180px">
          <el-option v-for="ct in allContentTypes" :key="ct" :label="ct" :value="ct" />
        </el-select>
        <el-select v-model="filters.system" :placeholder="t('plugin.system')" clearable style="width: 140px">
          <el-option :label="t('common.all')" value="all" />
          <el-option :label="t('plugin.filters.onlySystem')" value="yes" />
          <el-option :label="t('plugin.filters.nonSystem')" value="no" />
        </el-select>
        <el-select v-model="filters.hasSchema" :placeholder="t('plugin.config.schema')" clearable style="width: 160px">
          <el-option :label="t('common.all')" value="all" />
          <el-option :label="t('plugin.filters.hasSchema')" value="yes" />
          <el-option :label="t('plugin.filters.noSchema')" value="no" />
        </el-select>
        <el-button @click="load">{{ t('common.refresh') }}</el-button>
      </div>
    </div>
    <el-table v-loading="loading" :data="filteredItems" border stripe :empty-text="`暂无${t('plugin.pluginColumn')}`">
      <el-table-column :label="t('plugin.pluginColumn')" min-width="210">
        <template #default="{ row }">
          <div class="plugin-title">
            <strong>{{ row.name }}</strong>
            <span>{{ row.code }}</span>
          </div>
          <el-tag v-if="row.is_system" size="small" type="primary">{{ t('plugin.system') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" :label="t('plugin.version')" width="100" />
      <el-table-column :label="t('plugin.status')" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" effect="light">{{ pluginStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.health')" min-width="170">
        <template #default="{ row }">
          <div class="health-cell">
            <el-tag :type="healthType(row.health?.status)" effect="light">{{ pluginHealthLabel(row.health?.status) }}</el-tag>
            <span class="muted">{{ row.health?.suggested_action || t('plugin.noneSuggestion') }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.contentType')" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="type in row.content_types || []" :key="type" class="mr-6" effect="plain">{{ type }}</el-tag>
          <span v-if="!(row.content_types || []).length" class="muted">{{ t('common.none') }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.capabilitySummary')" min-width="220">
        <template #default="{ row }">
          <div class="metric-line">
            <el-tag type="info" effect="plain">{{ t('plugin.capability.permissions') }} {{ (row.permissions || []).length }}</el-tag>
            <el-tag type="info" effect="plain">{{ t('plugin.capability.menus') }} {{ (row.menus || []).length }}</el-tag>
            <el-tag :type="hasConfigSchema(row) ? 'success' : 'info'" effect="plain">
              {{ t('plugin.capability.schema') }} {{ hasConfigSchema(row) ? t('common.yes') : t('common.no') }}
            </el-tag>
            <el-tag :type="(row.hooks || []).length ? 'success' : 'info'" effect="plain">{{ t('plugin.capability.hooks') }} {{ (row.hooks || []).length }}</el-tag>
            <el-tag :type="statusMetricType(row.health?.config_status)" effect="plain">{{ t('plugin.config.title') }} {{ pluginHealthLabel(row.health?.config_status) }}</el-tag>
            <el-tag :type="statusMetricType(row.health?.migration_status)" effect="plain">{{ t('plugin.migration.title') }} {{ pluginHealthLabel(row.health?.migration_status) }}</el-tag>
            <el-tag :type="statusMetricType(row.health?.hook_status)" effect="plain">{{ t('plugin.capability.hook') }} {{ pluginHealthLabel(row.health?.hook_status) }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.recentError')" min-width="220">
        <template #default="{ row }">
          <span class="muted">{{ row.health?.recent_error || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="description" :label="t('plugin.descriptionLabel')" min-width="260" />
      <el-table-column :label="t('plugin.action')" fixed="right" width="300">
        <template #default="{ row }">
          <el-button link type="primary" :data-testid="`plugin-detail-${row.code}`" @click="openManifest(row)">{{ t('common.detail') }}</el-button>
          <el-button link type="info" @click="openManifest(row, 'permissions')">{{ t('plugin.capability.permissions') }}</el-button>
          <el-button link type="info" @click="openManifest(row, 'menus')">{{ t('plugin.capability.menus') }}</el-button>
          <el-button link type="primary" @click="openManifest(row, 'config')">{{ t('plugin.config.title') }}</el-button>
          <el-button v-if="canOpen(row)" link type="primary" :data-testid="`plugin-manage-${row.code}`" @click="openPlugin(row)">{{ t('common.manage') }}</el-button>
          <el-button v-if="row.status !== 'enabled' && row.status !== 'archived'" link type="success" :data-testid="`plugin-enable-${row.code}`" @click="setStatus(row, 'enabled')">{{ t('common.enable') }}</el-button>
          <el-button v-if="row.status === 'enabled'" link type="warning" :data-testid="`plugin-disable-${row.code}`" @click="setStatus(row, 'disabled')">{{ t('common.disable') }}</el-button>
          <el-button v-if="row.status !== 'archived'" link type="danger" :data-testid="`plugin-archive-${row.code}`" @click="archive(row)">{{ t('common.archive') }}</el-button>
          <el-button v-if="row.status === 'archived'" link type="success" :data-testid="`plugin-restore-${row.code}`" @click="restore(row)">{{ t('common.restore') }}</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>

  <PluginDetailDrawer v-model="manifestDialog" :plugin="manifestTarget" :initial-tab="manifestInitialTab" @refresh="load" />
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import { archivePlugin, disablePlugin, enablePlugin, pluginImpact, plugins, restorePlugin } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';
import PluginDetailDrawer from '@/components/plugin/PluginDetailDrawer.vue';
import { t } from '@/i18n';
import { pluginHealthLabel, pluginStatusLabel } from '@/i18n/formatters';

const auth = useAuthStore();
const router = useRouter();
const items = ref([]);
const filters = reactive({
  q: '',
  status: 'all',
  contentType: '',
  system: 'all',
  hasSchema: 'all',
});
const loading = ref(false);
const manifestDialog = ref(false);
const manifestTarget = ref(null);
const manifestInitialTab = ref('overview');

const allContentTypes = computed(() => {
  const seen = new Set();
  for (const p of items.value || []) {
    for (const ct of p.content_types || []) seen.add(ct);
  }
  return Array.from(seen).sort();
});

const filteredItems = computed(() => {
  const q = (filters.q || '').trim().toLowerCase();
  return (items.value || []).filter((p) => {
    if (filters.status && filters.status !== 'all' && p.status !== filters.status) return false;
    if (filters.contentType && filters.contentType !== '' && !(p.content_types || []).includes(filters.contentType)) return false;
    if (filters.system === 'yes' && !p.is_system) return false;
    if (filters.system === 'no' && p.is_system) return false;
    const hasSchema = hasConfigSchema(p);
    if (filters.hasSchema === 'yes' && !hasSchema) return false;
    if (filters.hasSchema === 'no' && hasSchema) return false;
    if (!q) return true;
    return (p.code || '').toLowerCase().includes(q) || (p.name || '').toLowerCase().includes(q);
  });
});

const stats = computed(() => {
  const list = items.value || [];
  const total = list.length;
  const enabled = list.filter((p) => p.status === 'enabled').length;
  const disabled = list.filter((p) => p.status === 'disabled').length;
  const system = list.filter((p) => p.is_system).length;
  const hasSchema = list.filter((p) => hasConfigSchema(p)).length;
  return { total, enabled, disabled, system, hasSchema };
});

async function load() {
  loading.value = true;
  try {
    const data = await plugins();
    items.value = data.items || [];
  } finally {
    loading.value = false;
  }
}

async function setStatus(row, status) {
  if (status === 'disabled') {
    const lines = await impactLines(row);
    await ElMessageBox.confirm(
      `${t('plugin.disableConfirmPrefix')}\n\n${lines.join('\n')}`,
      t('plugin.disableConfirmTitle'),
      { type: 'warning', confirmButtonText: t('plugin.confirmDisable'), cancelButtonText: t('common.cancel') },
    );
  } else {
    await ElMessageBox.confirm(t('plugin.enableConfirmText'), t('plugin.enableConfirmTitle'), {
      type: 'info',
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
  await ElMessageBox.confirm(
    `${t('plugin.archiveConfirmPrefix')}\n\n${lines.join('\n')}`,
    t('plugin.archiveConfirmTitle'),
    { type: 'warning', confirmButtonText: t('plugin.confirmArchive'), cancelButtonText: t('common.cancel') },
  );
  await archivePlugin(row.code);
  ElMessage.success(t('plugin.archivedDone'));
  await load();
}

async function restore(row) {
  await ElMessageBox.confirm(t('plugin.restoreConfirmText'), t('plugin.restoreConfirmTitle'), {
    type: 'info',
    confirmButtonText: t('plugin.confirmRestore'),
    cancelButtonText: t('common.cancel'),
  });
  await restorePlugin(row.code);
  ElMessage.success(t('plugin.restoredDone'));
  await load();
}

async function impactLines(row) {
  let impact = null;
  try {
    impact = await pluginImpact(row.code);
  } catch {
    impact = null;
  }
  const lines = [];
  if (impact) {
    lines.push(`${t('plugin.impact.enabledCommunities')}：${impact.enabled_communities_count ?? 0}`);
    lines.push(`${t('plugin.impact.disabledCommunities')}：${impact.disabled_communities_count ?? 0}`);
    lines.push(`${t('plugin.impact.blockedCategories')}：${impact.categories_count ?? 0}`);
    lines.push(`${t('plugin.impact.existingContents')}：${impact.existing_contents_count ?? impact.topics_count ?? 0}（${t('plugin.impact.historySeoSafe')}）`);
    if (typeof impact.recent_contents_count === 'number') lines.push(`${t('plugin.impact.recentContents')}：${impact.recent_contents_count}`);
    if (typeof impact.pending_contents_count === 'number') lines.push(`${t('plugin.impact.pendingContents')}：${impact.pending_contents_count}`);
    if (typeof impact.configs_count === 'number') lines.push(`${t('plugin.impact.configOverrides')}：${impact.configs_count}`);
    if (typeof impact.pending_migrations_count === 'number') lines.push(`${t('plugin.impact.pendingMigrations')}：${impact.pending_migrations_count}`);
    lines.push(
      `${t('plugin.impact.menus')}：${impact.menus_count}（${t('plugin.impact.frontendMenus')} ${impact.frontend_menus_count} / ${t('plugin.impact.moderatorMenus')} ${impact.moderator_menus_count} / ${t('plugin.impact.adminMenus')} ${impact.admin_menus_count}）`,
    );
    if (typeof impact.recent_hook_errors_count === 'number') lines.push(`${t('plugin.impact.recentHookErrors')}：${impact.recent_hook_errors_count}`);
  } else {
    lines.push(t('plugin.impact.unavailable'));
  }
  return lines;
}

function canOpen(row) {
  const target = adminMenu(row)?.path;
  return Boolean(target && (row.status === 'enabled' || row.status === 'archived') && hasPermission(row));
}

function hasPermission(row) {
  const required = adminMenu(row)?.permission || row.permissions?.[0]?.code || '';
  return !required || auth.can(required);
}

function adminMenu(row) {
  return (row.menus || []).find((item) => (item.area || item.location) === 'admin');
}

function openPlugin(row) {
  const target = adminMenu(row)?.path;
  if (!target) return;
  router.push(target);
}

function openManifest(row, tab = 'overview') {
  manifestTarget.value = row;
  manifestInitialTab.value = tab || 'overview';
  manifestDialog.value = true;
}

function statusType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'danger';
  if (status === 'archived') return 'info';
  return 'info';
}

function healthType(status) {
  if (status === 'healthy') return 'success';
  if (status === 'disabled' || status === 'archived') return 'info';
  if (status === 'warning' || status === 'migration_pending' || status === 'hook_warning') return 'warning';
  if (status === 'hook_error') return 'danger';
  if (status === 'error' || status === 'config_invalid' || status === 'dependency_missing') return 'danger';
  return 'info';
}

function statusMetricType(status) {
  if (status === 'ok' || status === 'valid') return 'success';
  if (status === 'warning' || status === 'pending' || status === 'hook_warning') return 'warning';
  if (status === 'failed' || status === 'invalid' || status === 'missing' || status === 'hook_error') return 'danger';
  return 'info';
}

function hasConfigSchema(row) {
  const schema = row?.config_schema;
  if (!schema) return false;
  if (Array.isArray(schema)) return schema.length > 0;
  if (typeof schema === 'object') return Object.keys(schema).length > 0;
  return true;
}

onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: flex-end; gap: 16px; }
.tool-actions { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; justify-content: flex-end; }
.toolbar h2 { margin: 0 0 6px; }
.toolbar p { margin: 0; color: #64748b; }
.intro-alert { border-radius: 12px; }
.stats-grid { display: grid; grid-template-columns: repeat(5, minmax(150px, 1fr)); gap: 12px; }
.stat-card { min-height: 76px; border: 1px solid #e2e8f0; border-radius: 14px; background: #fff; padding: 10px 14px; }
.stat-k { color: #64748b; font-size: 12px; }
.stat-v { color: #0f172a; font-size: 22px; font-weight: 700; margin-top: 4px; }
.mr-6 { margin-right: 6px; }
.mb { margin-bottom: 12px; }
.muted { color: #64748b; }
.plugin-title { display: grid; gap: 2px; margin-bottom: 6px; }
.plugin-title strong { color: #0f172a; }
.plugin-title span { color: #64748b; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.metric-line { display: flex; flex-wrap: wrap; gap: 6px; }
.health-cell { display: grid; gap: 6px; }
.page-card :deep(.el-table__cell) { padding: 8px 0; }
.page-card :deep(.el-table .cell) { line-height: 1.35; }
.json-box { margin: 0; padding: 14px; border-radius: 12px; background: #0f172a; color: #dbeafe; max-height: 360px; overflow: auto; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; line-height: 1.55; white-space: pre-wrap; }
.json-box.compact { max-height: 180px; }
</style>
