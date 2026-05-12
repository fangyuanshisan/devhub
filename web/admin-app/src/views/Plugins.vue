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

    <div class="health-grid" data-testid="plugin-health-summary">
      <div v-for="card in healthCards" :key="card.key" class="stat-card health-card">
        <div class="stat-k">{{ card.label }}</div>
        <div class="stat-v">{{ card.value }}</div>
        <div class="stat-sub">{{ card.tip }}</div>
      </div>
    </div>

    <div class="toolbar">
      <div>
        <h2>{{ t('plugin.title') }}</h2>
        <p>{{ t('plugin.coreNote') }}</p>
      </div>
      <div class="tool-actions">
        <el-button type="primary" plain data-testid="plugin-manifest-validate" @click="openManifestDialog('validate')">{{ t('plugin.ops.validateManifest') }}</el-button>
        <el-button type="primary" plain data-testid="plugin-manifest-dry-run" @click="openManifestDialog('dry-run')">{{ t('plugin.ops.dryRun') }}</el-button>
        <el-button type="success" plain data-testid="plugin-manifest-install" @click="openManifestDialog('install')">{{ t('plugin.ops.install') }}</el-button>
        <el-button type="warning" plain data-testid="plugin-bulk-archive" :disabled="!selectedRows.length" @click="bulkAction('archive')">{{ t('plugin.ops.bulkArchive') }}</el-button>
        <el-button type="info" plain data-testid="plugin-bulk-restore" :disabled="!selectedRows.length" @click="bulkAction('restore')">{{ t('plugin.ops.bulkRestore') }}</el-button>
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
    <el-table v-loading="loading" :data="filteredItems" border stripe :empty-text="`暂无${t('plugin.pluginColumn')}`" @selection-change="onSelectionChange">
      <el-table-column type="selection" width="55" />
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
          <el-button link type="success" :data-testid="`plugin-upgrade-preview-${row.code}`" @click="openManifestDialog('upgrade-dry-run', row)">{{ t('plugin.ops.upgradePreview') }}</el-button>
          <el-button link type="warning" :data-testid="`plugin-upgrade-${row.code}`" @click="openManifestDialog('upgrade', row)">{{ t('plugin.ops.upgrade') }}</el-button>
          <el-button v-if="canOpen(row)" link type="primary" :data-testid="`plugin-manage-${row.code}`" @click="openPlugin(row)">{{ t('common.manage') }}</el-button>
          <el-button v-if="row.status !== 'enabled' && row.status !== 'archived'" link type="success" :data-testid="`plugin-enable-${row.code}`" @click="setStatus(row, 'enabled')">{{ t('common.enable') }}</el-button>
          <el-button v-if="row.status === 'enabled'" link type="warning" :data-testid="`plugin-disable-${row.code}`" @click="setStatus(row, 'disabled')">{{ t('common.disable') }}</el-button>
          <el-button v-if="row.status !== 'archived'" link type="danger" :data-testid="`plugin-archive-${row.code}`" @click="archive(row)">{{ t('common.archive') }}</el-button>
          <el-button v-if="row.status === 'archived'" link type="success" :data-testid="`plugin-restore-${row.code}`" @click="restore(row)">{{ t('common.restore') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <section v-if="manifestDialogVisible" class="action-panel" data-testid="plugin-manifest-panel">
      <div class="action-panel-header">
        <div>
          <h3>{{ manifestDialogTitle }}</h3>
          <p>{{ manifestDialogTip }}</p>
        </div>
        <div class="action-panel-tools">
          <el-button data-testid="plugin-manifest-cancel" @click="manifestDialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button :loading="manifestLoading" type="primary" data-testid="plugin-manifest-submit" @click="submitManifestAction">{{ manifestDialogActionLabel }}</el-button>
        </div>
      </div>
      <el-input
        v-model="manifestText"
        type="textarea"
        :rows="16"
        data-testid="plugin-manifest-input"
        :placeholder="t('plugin.ops.manifestPlaceholder')"
      />
    </section>

    <section v-if="resultDialogVisible" class="action-panel" data-testid="plugin-result-panel">
      <div class="action-panel-header">
        <div>
          <h3>{{ resultDialogTitle }}</h3>
          <p>{{ resultDialogMessage }}</p>
        </div>
        <div class="action-panel-tools">
          <el-button
            v-if="resultDialogKind === 'bulk'"
            data-testid="plugin-result-audit"
            type="success"
            plain
            @click="openAuditLogs"
          >{{ t('plugin.ops.viewAuditLogs') }}</el-button>
          <el-button data-testid="plugin-result-close" type="primary" @click="resultDialogVisible = false">{{ t('common.close') }}</el-button>
        </div>
      </div>
      <template v-if="resultDialogKind === 'manifest'">
        <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
          <el-descriptions-item :label="t('plugin.ops.resultValid')">{{ manifestResult.valid ? t('common.yes') : t('common.no') }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.sourceType')">{{ manifestResult.install_preview?.source_type || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.checksum')">
            <span class="mono">{{ manifestResult.checksum || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.initialStatus')">{{ pluginStatusLabel(manifestResult.install_preview?.initial_status) }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.dependencies')">{{ (manifestResult.dependencies || []).join(', ') || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.impactSummary')">
            {{ t('plugin.ops.contentTypesCount') }} {{ manifestResult.impact_summary?.content_types_count ?? 0 }}，
            {{ t('plugin.ops.permissionsCount') }} {{ manifestResult.impact_summary?.permissions_count ?? 0 }}，
            {{ t('plugin.ops.menusCount') }} {{ manifestResult.impact_summary?.menus_count ?? 0 }}，
            {{ t('plugin.ops.routesCount') }} {{ manifestResult.impact_summary?.routes_count ?? 0 }}
          </el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-if="resultErrors.length"
          :title="t('plugin.ops.errors')"
          type="error"
          show-icon
          :closable="false"
          class="mb"
        >
          <ul class="result-list">
            <li v-for="(item, idx) in resultErrors" :key="`err-${idx}`">{{ item }}</li>
          </ul>
        </el-alert>
        <el-alert
          v-if="resultWarnings.length"
          :title="t('plugin.ops.warnings')"
          type="warning"
          show-icon
          :closable="false"
          class="mb"
        >
          <ul class="result-list">
            <li v-for="(item, idx) in resultWarnings" :key="`warn-${idx}`">{{ item }}</li>
          </ul>
        </el-alert>
        <div class="result-grid">
          <div class="result-box">
            <h4>{{ t('plugin.ops.contentTypeConflicts') }}</h4>
            <div class="tag-wrap">
              <el-tag v-for="item in manifestResult.content_type_conflicts || []" :key="item" type="danger" effect="plain">{{ item }}</el-tag>
              <span v-if="!(manifestResult.content_type_conflicts || []).length" class="muted">-</span>
            </div>
          </div>
          <div class="result-box">
            <h4>{{ t('plugin.ops.permissionConflicts') }}</h4>
            <div class="tag-wrap">
              <el-tag v-for="item in manifestResult.permission_conflicts || []" :key="item" type="warning" effect="plain">{{ item }}</el-tag>
              <span v-if="!(manifestResult.permission_conflicts || []).length" class="muted">-</span>
            </div>
          </div>
          <div class="result-box">
            <h4>{{ t('plugin.ops.migrationPlan') }}</h4>
            <div class="tag-wrap">
              <el-tag v-for="item in manifestResult.migration_plan || []" :key="item.migration_name" effect="plain">{{ item.migration_name }} / {{ item.migration_version }}</el-tag>
              <span v-if="!(manifestResult.migration_plan || []).length" class="muted">-</span>
            </div>
          </div>
          <div class="result-box">
            <h4>{{ t('plugin.ops.installPreview') }}</h4>
            <pre class="json-box compact">{{ formatJSON(manifestResult.install_preview || {}) }}</pre>
          </div>
        </div>
      </template>
      <template v-else-if="resultDialogKind === 'upgrade'">
        <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
          <el-descriptions-item :label="t('plugin.ops.currentVersion')">{{ resultUpgrade.current_version || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.newVersion')">{{ resultUpgrade.new_version || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.currentCoreVersion')">{{ resultUpgrade.current_core_version || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.compatibilityStatus')">{{ compatibilityLabel(resultUpgrade.compatibility_status) }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.changedKeys')" :span="2">{{ (resultUpgrade.changed_keys || []).join(', ') || '-' }}</el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-if="upgradeWarnings.length"
          :title="t('plugin.ops.warnings')"
          type="warning"
          show-icon
          :closable="false"
          class="mb"
        >
          <ul class="result-list">
            <li v-for="(item, idx) in upgradeWarnings" :key="`uwarn-${idx}`">{{ item }}</li>
          </ul>
        </el-alert>
        <el-alert
          v-if="upgradeErrors.length"
          :title="t('plugin.ops.errors')"
          type="error"
          show-icon
          :closable="false"
          class="mb"
        >
          <ul class="result-list">
            <li v-for="(item, idx) in upgradeErrors" :key="`uerr-${idx}`">{{ item }}</li>
          </ul>
        </el-alert>
        <div class="result-grid">
          <div class="result-box">
            <h4>{{ t('plugin.ops.diffCurrent') }}</h4>
            <pre class="json-box compact">{{ formatJSON(resultUpgrade.diff?.current || {}) }}</pre>
          </div>
          <div class="result-box">
            <h4>{{ t('plugin.ops.diffNew') }}</h4>
            <pre class="json-box compact">{{ formatJSON(resultUpgrade.diff?.new || {}) }}</pre>
          </div>
        </div>
      </template>
      <template v-else>
        <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
          <el-descriptions-item :label="t('plugin.ops.succeededCount')">{{ (resultDialogPayload.succeeded || []).length }}</el-descriptions-item>
          <el-descriptions-item :label="t('plugin.ops.failedCount')">{{ (resultDialogPayload.failed || []).length }}</el-descriptions-item>
        </el-descriptions>
        <div class="result-grid">
          <div class="result-box">
            <h4>{{ t('plugin.ops.succeeded') }}</h4>
            <el-table :data="resultDialogPayload.succeeded || []" border stripe :empty-text="t('common.none')">
              <el-table-column prop="plugin_code" :label="t('plugin.code')" min-width="180" />
              <el-table-column prop="status" :label="t('plugin.status')" width="120" />
            </el-table>
          </div>
          <div class="result-box">
            <h4>{{ t('plugin.ops.failed') }}</h4>
            <el-table :data="resultDialogPayload.failed || []" border stripe :empty-text="t('common.none')">
              <el-table-column prop="plugin_code" :label="t('plugin.code')" min-width="180" />
              <el-table-column prop="error" :label="t('plugin.ops.error')" min-width="260" />
            </el-table>
          </div>
        </div>
      </template>
    </section>
  </section>

  <PluginDetailDrawer v-model="manifestDialog" :plugin="manifestTarget" :initial-tab="manifestInitialTab" @refresh="load" />
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import {
  archivePlugin,
  bulkArchivePlugins,
  bulkRestorePlugins,
  disablePlugin,
  dryRunPluginManifest,
  dryRunPluginUpgrade,
  enablePlugin,
  installPluginManifest,
  pluginHealthSummary,
  pluginImpact,
  plugins,
  restorePlugin,
  validatePluginManifest,
  upgradePlugin,
} from '@/api/admin';
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
const selectedRows = ref([]);
const healthSummary = ref({});
const manifestDialogVisible = ref(false);
const manifestDialogAction = ref('validate');
const manifestDialogTarget = ref(null);
const manifestText = ref('');
const manifestLoading = ref(false);
const resultDialogVisible = ref(false);
const resultDialogKind = ref('manifest');
const resultDialogTitle = ref('');
const resultDialogMessage = ref('');
const resultDialogPayload = ref({});
const resultAuditQuery = ref(null);
const manifestResult = computed(() => resultDialogPayload.value?.validation || resultDialogPayload.value || {});
const resultErrors = computed(() => Array.isArray(manifestResult.value?.errors) ? manifestResult.value.errors : []);
const resultWarnings = computed(() => Array.isArray(manifestResult.value?.warnings) ? manifestResult.value.warnings : []);

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

const healthCards = computed(() => {
  const summary = healthSummary.value || {};
  return [
    { key: 'healthy', label: pluginHealthLabel('healthy'), value: summary.healthy || 0, tip: t('plugin.healthCard.tip') },
    { key: 'warning', label: pluginHealthLabel('warning'), value: summary.warning || 0, tip: t('plugin.healthCard.tip') },
    { key: 'error', label: pluginHealthLabel('error'), value: summary.error || 0, tip: t('plugin.healthCard.tip') },
    { key: 'disabled', label: pluginHealthLabel('disabled'), value: summary.disabled || 0, tip: t('plugin.healthCard.tip') },
    { key: 'archived', label: pluginHealthLabel('archived'), value: summary.archived || 0, tip: t('plugin.healthCard.tip') },
    { key: 'migration_pending', label: pluginHealthLabel('migration_pending'), value: summary.migration_pending || 0, tip: t('plugin.healthCard.tip') },
    { key: 'config_invalid', label: pluginHealthLabel('config_invalid'), value: summary.config_invalid || 0, tip: t('plugin.healthCard.tip') },
    { key: 'dependency_missing', label: pluginHealthLabel('dependency_missing'), value: summary.dependency_missing || 0, tip: t('plugin.healthCard.tip') },
    { key: 'hook_error', label: pluginHealthLabel('hook_error'), value: summary.hook_error || 0, tip: t('plugin.healthCard.tip') },
  ];
});

const manifestDialogTitle = computed(() => {
  if (manifestDialogAction.value === 'dry-run') return t('plugin.ops.dryRun');
  if (manifestDialogAction.value === 'upgrade-dry-run') return t('plugin.ops.upgradePreview');
  if (manifestDialogAction.value === 'upgrade') return t('plugin.ops.upgrade');
  if (manifestDialogAction.value === 'install') return t('plugin.ops.install');
  return t('plugin.ops.validateManifest');
});

const manifestDialogTip = computed(() => {
  if (manifestDialogAction.value === 'dry-run') return t('plugin.ops.dryRunTip');
  if (manifestDialogAction.value === 'upgrade-dry-run') return t('plugin.ops.upgradeTip');
  if (manifestDialogAction.value === 'upgrade') return t('plugin.ops.upgradeConfirmTip');
  if (manifestDialogAction.value === 'install') return t('plugin.ops.installTip');
  return t('plugin.ops.validateTip');
});

const manifestDialogActionLabel = computed(() => {
  if (manifestDialogAction.value === 'dry-run') return t('plugin.ops.dryRun');
  if (manifestDialogAction.value === 'upgrade-dry-run') return t('plugin.ops.upgradePreview');
  if (manifestDialogAction.value === 'upgrade') return t('plugin.ops.upgrade');
  if (manifestDialogAction.value === 'install') return t('plugin.ops.install');
  return t('plugin.ops.validateManifest');
});

const resultUpgrade = computed(() => resultDialogPayload.value || {});
const upgradeErrors = computed(() => Array.isArray(resultUpgrade.value.validation?.errors) ? resultUpgrade.value.validation.errors : []);
const upgradeWarnings = computed(() => Array.isArray(resultUpgrade.value.validation?.warnings) ? resultUpgrade.value.validation.warnings : []);

async function load() {
  loading.value = true;
  try {
    const [list, health] = await Promise.all([
      plugins(),
      pluginHealthSummary().catch(() => null),
    ]);
    items.value = list.items || [];
    healthSummary.value = health?.summary || {};
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

function onSelectionChange(rows) {
  selectedRows.value = Array.isArray(rows) ? rows : [];
}

function openManifestDialog(action, row = null) {
  manifestDialogAction.value = action;
  manifestDialogTarget.value = row;
  manifestText.value = manifestTemplate(row);
  resultDialogKind.value = 'manifest';
  resultDialogVisible.value = false;
  manifestDialogVisible.value = true;
}

async function submitManifestAction() {
  let payload;
  try {
    payload = JSON.parse(manifestText.value || '{}');
  } catch {
    ElMessage.error(t('plugin.ops.manifestInvalidJson'));
    return;
  }
  manifestLoading.value = true;
  try {
    let response;
    if (manifestDialogAction.value === 'install' || manifestDialogAction.value === 'upgrade') {
      await ElMessageBox.confirm(
        manifestDialogAction.value === 'install' ? t('plugin.ops.installTip') : t('plugin.ops.upgradeConfirmTip'),
        manifestDialogAction.value === 'install' ? t('plugin.ops.install') : t('plugin.ops.upgrade'),
        {
          type: 'warning',
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
        },
      );
    }
    if (manifestDialogAction.value === 'validate') response = await validatePluginManifest(payload);
    else if (manifestDialogAction.value === 'dry-run') response = await dryRunPluginManifest(payload);
    else if (manifestDialogAction.value === 'upgrade-dry-run') {
      const targetCode = manifestDialogTarget.value?.code || payload?.code;
      response = await dryRunPluginUpgrade(targetCode, payload);
      resultDialogKind.value = 'upgrade';
    }
    else if (manifestDialogAction.value === 'upgrade') {
      const targetCode = manifestDialogTarget.value?.code || payload?.code;
      response = await upgradePlugin(targetCode, payload);
      resultDialogKind.value = 'upgrade';
    } else response = await installPluginManifest(payload);
    if (manifestDialogAction.value !== 'upgrade-dry-run' && manifestDialogAction.value !== 'upgrade') resultDialogKind.value = 'manifest';
    manifestDialogVisible.value = false;
    resultDialogTitle.value = manifestResultTitle(manifestDialogAction.value);
    resultDialogMessage.value = manifestResultMessage(manifestDialogAction.value);
    resultDialogPayload.value = response;
    resultDialogVisible.value = true;
    await load();
    if (manifestDialogAction.value === 'install') {
      ElMessage.success(t('plugin.ops.installDone'));
    } else if (manifestDialogAction.value === 'upgrade') {
      ElMessage.success(t('plugin.ops.upgradeExecuted'));
    } else {
      ElMessage.success(t('plugin.ops.validateDone'));
    }
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('common.saveFailed')));
  } finally {
    manifestLoading.value = false;
  }
}

async function bulkAction(action) {
  if (!selectedRows.value.length) {
    ElMessage.warning(t('plugin.ops.noSelection'));
    return;
  }
  const label = action === 'archive' ? t('plugin.ops.bulkArchive') : t('plugin.ops.bulkRestore');
  const impactLines = await bulkImpactLines(selectedRows.value.map((row) => row.code).filter(Boolean));
  await ElMessageBox.confirm(
    `${label} ${selectedRows.value.length}${t('common.selectedItems')}？\n\n${impactLines.join('\n')}`,
    label,
    { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') },
  );
  const codes = selectedRows.value.map((row) => row.code).filter(Boolean);
  const response = action === 'archive'
    ? await bulkArchivePlugins({ codes })
    : await bulkRestorePlugins({ codes });
  resultDialogKind.value = 'bulk';
  resultDialogTitle.value = label;
  resultDialogMessage.value = t('plugin.ops.bulkDone');
  resultDialogPayload.value = response;
  resultAuditQuery.value = {
    action: action === 'archive' ? 'plugin.bulk_archived' : 'plugin.bulk_restored',
    target: 'plugins#bulk',
    metadata: codes.join(','),
  };
  resultDialogVisible.value = true;
  selectedRows.value = [];
  await load();
}

function manifestResultTitle(action) {
  if (action === 'dry-run') return t('plugin.ops.dryRunResult');
  if (action === 'upgrade-dry-run') return t('plugin.ops.upgradeResult');
  if (action === 'upgrade') return t('plugin.ops.upgradeExecuteResult');
  if (action === 'install') return t('plugin.ops.installResult');
  return t('plugin.ops.validateResult');
}

function manifestResultMessage(action) {
  if (action === 'dry-run') return t('plugin.ops.dryRunDone');
  if (action === 'upgrade-dry-run') return t('plugin.ops.upgradeDone');
  if (action === 'upgrade') return t('plugin.ops.upgradeExecuted');
  if (action === 'install') return t('plugin.ops.installDone');
  return t('plugin.ops.validateDone');
}

async function bulkImpactLines(codes) {
  if (!codes.length) return [t('plugin.ops.noSelection')];
  const items = await Promise.all(
    codes.map(async (code) => {
      try {
        return await pluginImpact(code);
      } catch {
        return null;
      }
    }),
  );
  const summary = items.reduce(
    (acc, item) => {
      if (!item) return acc;
      acc.contents += Number(item.existing_contents_count ?? item.topics_count ?? 0);
      acc.communities += Number(item.enabled_communities_count ?? 0);
      acc.categories += Number(item.categories_count ?? 0);
      acc.migrations += Number(item.pending_migrations_count ?? 0);
      acc.hooks += Number(item.recent_hook_errors_count ?? 0);
      return acc;
    },
    { contents: 0, communities: 0, categories: 0, migrations: 0, hooks: 0 },
  );
  return [
    `${t('plugin.ops.selectedPlugins')}：${codes.join(', ')}`,
    `${t('plugin.impact.existingContents')}：${summary.contents}`,
    `${t('plugin.impact.enabledCommunities')}：${summary.communities}`,
    `${t('plugin.impact.blockedCategories')}：${summary.categories}`,
    `${t('plugin.impact.pendingMigrations')}：${summary.migrations}`,
    `${t('plugin.impact.recentHookErrors')}：${summary.hooks}`,
  ];
}

function openAuditLogs() {
  const query = resultAuditQuery.value || {
    action: 'plugin.upgraded',
    target: 'plugins',
    metadata: '',
  };
  router.push({ path: '/admin-next/audit-logs', query });
}

function manifestTemplate(row = null) {
  if (row) return manifestFromPlugin(row);
  const ts = Date.now();
  return JSON.stringify(
    {
      code: `e2e_plugin_${ts}`,
      name: 'E2E 演示插件',
      version: '1.0.0',
      description: 'E2E 使用的 manifest 示例',
      compatible_core_version: '>=1.3.4',
      is_system: false,
      content_types: [
        {
          type: `e2e_type_${ts}`,
          name: 'E2E 内容',
          plugin_code: `e2e_plugin_${ts}`,
          create_permission: `e2e.plugin.create.${ts}`,
        },
      ],
      permissions: [
        {
          code: `e2e.plugin.create.${ts}`,
          name: '创建 E2E 内容',
          scope: 'community',
        },
      ],
      menus: [],
      routes: [],
      hooks: [],
      config_schema: { type: 'object', properties: {} },
      migrations: [],
    },
    null,
    2,
  );
}

function manifestFromPlugin(row) {
  if (!row) return manifestTemplate();
  const current = safeJSON(row.manifest_json || {});
  const contentTypes = Array.isArray(row.content_type_definitions) && row.content_type_definitions.length
    ? row.content_type_definitions
    : (current.content_types || []).map((type) => ({ type, name: type, plugin_code: row.code, create_permission: `${row.code}.${type}.create` }));
  return JSON.stringify(
    {
      code: row.code,
      name: row.name,
      version: row.version || '1.0.0',
      description: row.description || '',
      compatible_core_version: row.compatible_core_version || '>=1.3.4',
      is_system: Boolean(row.is_system),
      content_types: contentTypes,
      permissions: row.permissions || current.permissions || [],
      menus: row.menus || current.menus || [],
      routes: row.routes || current.routes || [],
      hooks: row.hooks || current.hooks || [],
      config_schema: row.config_schema || current.config_schema || { type: 'object', properties: {} },
      migrations: row.migrations || current.migrations || [],
    },
    null,
    2,
  );
}

function safeJSON(value) {
  if (!value) return {};
  if (typeof value === 'string') {
    try {
      return JSON.parse(value);
    } catch {
      return {};
    }
  }
  return typeof value === 'object' ? value : {};
}

function compatibilityLabel(status) {
  if (status === 'compatible') return t('plugin.ops.compatible');
  if (status === 'incompatible') return t('plugin.ops.incompatible');
  return t('plugin.ops.unknownCompatibility');
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

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
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
.health-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px; }
.stat-card { min-height: 76px; border: 1px solid #e2e8f0; border-radius: 14px; background: #fff; padding: 10px 14px; }
.stat-k { color: #64748b; font-size: 12px; }
.stat-v { color: #0f172a; font-size: 22px; font-weight: 700; margin-top: 4px; }
.stat-sub { margin-top: 4px; color: #94a3b8; font-size: 12px; line-height: 1.4; }
.mr-6 { margin-right: 6px; }
.mb { margin-bottom: 12px; }
.muted { color: #64748b; }
.plugin-title { display: grid; gap: 2px; margin-bottom: 6px; }
.plugin-title strong { color: #0f172a; }
.plugin-title span { color: #64748b; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.metric-line { display: flex; flex-wrap: wrap; gap: 6px; }
.health-cell { display: grid; gap: 6px; }
.action-panel {
  border: 1px solid #dbe4f0;
  border-radius: 16px;
  background: #fff;
  padding: 14px;
}
.action-panel-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 12px;
}
.action-panel-header h3 {
  margin: 0;
  font-size: 16px;
  color: #0f172a;
}
.action-panel-header p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 13px;
}
.action-panel-tools {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.page-card :deep(.el-table__cell) { padding: 8px 0; }
.page-card :deep(.el-table .cell) { line-height: 1.35; }
.json-box { margin: 0; padding: 14px; border-radius: 12px; background: #0f172a; color: #dbeafe; max-height: 360px; overflow: auto; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; line-height: 1.55; white-space: pre-wrap; }
.json-box.compact { max-height: 180px; }
</style>
