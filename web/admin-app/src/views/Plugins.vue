<template>
  <section class="page-card" data-testid="admin-plugins-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">{{ t('plugin.pageEyebrow') }}</div>
        <h2>{{ t('plugin.title') }}</h2>
        <p>{{ t('plugin.description') }}</p>
      </div>
      <div class="primary-actions">
        <el-button type="primary" plain data-testid="plugin-manifest-validate" @click="openManifestDialog('validate')">{{ t('plugin.ops.validateManifest') }}</el-button>
        <el-button type="primary" plain data-testid="plugin-manifest-dry-run" @click="openManifestDialog('dry-run')">{{ t('plugin.ops.dryRun') }}</el-button>
        <el-button type="success" plain data-testid="plugin-manifest-install" @click="openManifestDialog('install')">{{ t('plugin.ops.install') }}</el-button>
      </div>
    </div>

    <el-tabs v-model="activeView" class="plugin-main-tabs">
      <el-tab-pane :label="t('plugin.list')" name="list" />
      <el-tab-pane :label="t('plugin.governance.title')" name="governance" />
    </el-tabs>

    <template v-if="activeView === 'list'">
      <div class="stats-grid" data-testid="plugin-stats">
        <button class="stat-card stat-button" type="button" @click="filters.status = 'all'">
          <div class="stat-k">{{ t('plugin.stats.total') }}</div>
          <div class="stat-v">{{ stats.total }}</div>
        </button>
        <button class="stat-card stat-button" type="button" @click="filters.status = 'enabled'">
          <div class="stat-k">{{ t('plugin.stats.enabled') }}</div>
          <div class="stat-v">{{ stats.enabled }}</div>
        </button>
        <button class="stat-card stat-button" type="button" @click="filters.status = 'disabled'">
          <div class="stat-k">{{ t('plugin.stats.disabled') }}</div>
          <div class="stat-v">{{ stats.disabled }}</div>
        </button>
        <button class="stat-card stat-button" type="button" @click="filters.status = 'archived'">
          <div class="stat-k">{{ t('plugin.stats.archived') }}</div>
          <div class="stat-v">{{ stats.archived }}</div>
        </button>
        <button class="stat-card stat-card-danger stat-button" type="button" @click="filters.health = 'error'">
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
            <button v-for="card in healthCards" :key="card.key" class="stat-card stat-button health-card" type="button" @click="filters.health = card.key">
              <div class="stat-k">{{ card.label }}</div>
              <div class="stat-v">{{ card.value }}</div>
              <div class="stat-sub">{{ card.tip }}</div>
            </button>
          </div>
        </el-collapse-item>
      </el-collapse>

      <div class="filter-panel" data-testid="plugin-filter-panel">
        <div>
          <strong>{{ t('plugin.filters.title') }}</strong>
          <span>{{ t('plugin.filters.tip') }}</span>
        </div>
        <div class="filter-actions">
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
          <el-button @click="resetFilters">{{ t('common.reset') }}</el-button>
          <el-button @click="load">{{ t('common.refresh') }}</el-button>
        </div>
      </div>

      <div class="bulk-panel" data-testid="plugin-bulk-panel">
        <span>{{ t('common.selected') }} {{ selectedRows.length }} {{ t('plugin.ops.selectedPluginsUnit') }}</span>
        <div class="bulk-actions">
          <el-button type="warning" plain data-testid="plugin-bulk-archive" :disabled="!selectedRows.length" @click="bulkAction('archive')">{{ t('plugin.ops.bulkArchive') }}</el-button>
          <el-button type="info" plain data-testid="plugin-bulk-restore" :disabled="!selectedRows.length" @click="bulkAction('restore')">{{ t('plugin.ops.bulkRestore') }}</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="filteredItems" border stripe :empty-text="`暂无${t('plugin.pluginColumn')}`" @selection-change="onSelectionChange">
        <el-table-column type="selection" width="48" />
        <el-table-column :label="t('plugin.pluginColumn')" min-width="220">
          <template #default="{ row }">
            <div class="plugin-title">
              <strong>{{ row.name }}</strong>
              <span>{{ row.code }}</span>
            </div>
            <div class="tag-wrap">
              <el-tag v-if="row.is_system" size="small" type="primary">{{ t('plugin.system') }}</el-tag>
              <el-tag v-if="row.install_status" size="small" type="info" effect="plain">{{ pluginStatusLabel(row.install_status) }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="version" :label="t('plugin.version')" width="96" />
        <el-table-column :label="t('plugin.status')" min-width="150">
          <template #default="{ row }">
            <div class="status-stack">
              <el-tag :type="statusType(row.status)" effect="light">{{ pluginStatusLabel(row.status) }}</el-tag>
              <span class="muted">{{ pluginStatusLabel(row.lifecycle_status || row.runtime_status || row.status) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.health')" min-width="210">
          <template #default="{ row }">
            <div class="health-cell">
              <el-tag :type="healthType(row.health?.status)" effect="light">{{ pluginHealthLabel(row.health?.status) }}</el-tag>
              <span class="muted">{{ row.health?.status_reason || row.status_reason || row.health?.suggested_action || t('plugin.noneSuggestion') }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.contentType')" min-width="160">
          <template #default="{ row }">
            <div class="tag-wrap">
              <el-tag v-for="type in (row.content_types || []).slice(0, 3)" :key="type" effect="plain">{{ type }}</el-tag>
              <el-tag v-if="(row.content_types || []).length > 3" type="info" effect="plain">+{{ (row.content_types || []).length - 3 }}</el-tag>
              <span v-if="!(row.content_types || []).length" class="muted">{{ t('common.none') }}</span>
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
              <el-button link type="primary" :data-testid="`plugin-detail-${row.code}`" @click="openManifest(row)">{{ t('common.detail') }}</el-button>
              <el-button link type="primary" @click="openManifest(row, 'config')">{{ t('plugin.config.title') }}</el-button>
              <el-dropdown trigger="click" @command="(command) => handlePluginCommand(row, command)">
                <el-button link type="info">{{ t('plugin.ops.more') }}</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="permissions">{{ t('plugin.capability.permissions') }}</el-dropdown-item>
                    <el-dropdown-item command="menus">{{ t('plugin.capability.menus') }}</el-dropdown-item>
                    <el-dropdown-item command="hooks">{{ t('plugin.tabs.hooks') }}</el-dropdown-item>
                    <el-dropdown-item command="migrations">{{ t('plugin.tabs.migrations') }}</el-dropdown-item>
                    <el-dropdown-item command="runtime">{{ t('plugin.tabs.runtime') }}</el-dropdown-item>
                    <el-dropdown-item command="audit">{{ t('plugin.tabs.audit') }}</el-dropdown-item>
                    <el-dropdown-item v-if="canOpen(row)" command="manage" :data-testid="`plugin-manage-${row.code}`">{{ t('common.manage') }}</el-dropdown-item>
                    <el-dropdown-item command="upgrade-preview" :data-testid="`plugin-upgrade-preview-${row.code}`">{{ t('plugin.ops.upgradePreview') }}</el-dropdown-item>
                    <el-dropdown-item command="upgrade" :data-testid="`plugin-upgrade-${row.code}`">{{ t('plugin.ops.upgrade') }}</el-dropdown-item>
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
    </template>

    <template v-else>
      <div class="governance-grid" data-testid="plugin-governance-panel">
        <article v-for="group in governanceGroups" :key="group.key" class="governance-card">
          <div class="governance-card-head">
            <div>
              <h3>{{ group.title }}</h3>
              <p>{{ group.suggestion }}</p>
            </div>
            <el-tag :type="group.type">{{ group.items.length }}</el-tag>
          </div>
          <el-table :data="group.items" border stripe :empty-text="t('common.none')">
            <el-table-column :label="t('plugin.pluginColumn')" min-width="180">
              <template #default="{ row }">
                <div class="plugin-title compact-title">
                  <strong>{{ row.name }}</strong>
                  <span>{{ row.code }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.health')" min-width="180">
              <template #default="{ row }">
                <el-tag :type="healthType(row.health?.status)" effect="light">{{ pluginHealthLabel(row.health?.status) }}</el-tag>
                <div class="muted governance-reason">{{ row.health?.recent_error || row.health?.status_reason || row.status_reason || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.action')" width="120">
              <template #default="{ row }">
                <el-button link type="primary" @click="openGovernanceItem(row, group.tab)">{{ t('common.view') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </article>
      </div>
    </template>

    <el-drawer v-model="manifestDialogVisible" append-to-body destroy-on-close size="820px" :with-header="false" class="plugin-action-drawer">
      <section class="action-panel in-drawer" data-testid="plugin-manifest-panel">
        <div class="action-panel-header">
          <div>
            <h3>{{ manifestDialogTitle }}</h3>
            <p>{{ manifestDialogTip }}</p>
          </div>
          <div class="action-panel-tools">
            <el-button :data-testid="wizardStep > 0 ? 'plugin-result-close' : 'plugin-manifest-cancel'" @click="manifestDialogVisible = false">{{ wizardStep > 0 ? t('common.close') : t('common.cancel') }}</el-button>
            <el-button v-if="wizardStep > 0 && !isWizardResultStep" :disabled="manifestLoading" @click="wizardBack">{{ t('common.back') }}</el-button>
            <el-button v-if="canShortcutConfirm" :loading="manifestLoading" type="warning" plain @click="confirmWizardAction">
              {{ manifestDialogAction === 'install' ? t('plugin.wizard.confirmInstall') : t('plugin.wizard.confirmUpgrade') }}
            </el-button>
            <el-button :loading="manifestLoading" :disabled="!wizardCanProceed" type="primary" data-testid="plugin-manifest-submit" @click="submitManifestAction">{{ manifestDialogActionLabel }}</el-button>
          </div>
        </div>
        <el-steps :active="wizardStep" finish-status="success" align-center class="wizard-steps">
          <el-step v-for="step in wizardSteps" :key="step" :title="step" />
        </el-steps>
        <el-input
          v-if="wizardStep === 0"
          v-model="manifestText"
          type="textarea"
          :rows="24"
          data-testid="plugin-manifest-input"
          :placeholder="t('plugin.ops.manifestPlaceholder')"
        />
        <div v-else data-testid="plugin-result-panel">
          <el-alert v-if="resultErrors.length" :title="t('plugin.ops.errors')" type="error" show-icon :closable="false" class="mb">
            <ul class="result-list">
              <li v-for="(item, idx) in resultErrors" :key="`err-${idx}`">{{ item }}</li>
            </ul>
          </el-alert>
          <el-alert v-if="resultWarnings.length" :title="t('plugin.ops.warnings')" type="warning" show-icon :closable="false" class="mb">
            <ul class="result-list">
              <li v-for="(item, idx) in resultWarnings" :key="`warn-${idx}`">{{ item }}</li>
            </ul>
          </el-alert>

          <template v-if="manifestDialogAction === 'upgrade' || manifestDialogAction === 'upgrade-dry-run'">
            <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
              <el-descriptions-item :label="t('plugin.ops.currentVersion')">{{ resultUpgrade.current_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.newVersion')">{{ resultUpgrade.new_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.currentCoreVersion')">{{ resultUpgrade.current_core_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.compatibilityStatus')">{{ compatibilityLabel(resultUpgrade.compatibility_status) }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.changedKeys')" :span="2">{{ (resultUpgrade.changed_keys || []).join(', ') || '-' }}</el-descriptions-item>
            </el-descriptions>
            <el-alert v-if="isWizardConfirmStep" :title="t('plugin.wizard.confirmUpgradeTip')" type="warning" show-icon :closable="false" class="mb" />
            <div class="result-grid">
              <div class="result-box">
                <h4>{{ t('plugin.ops.diffCurrent') }}</h4>
                <pre class="json-box compact">{{ formatJSON(resultUpgrade.diff?.current || {}) }}</pre>
              </div>
              <div class="result-box">
                <h4>{{ t('plugin.ops.diffNew') }}</h4>
                <pre class="json-box compact">{{ formatJSON(resultUpgrade.diff?.new || wizardResult || {}) }}</pre>
              </div>
            </div>
          </template>

          <template v-else>
            <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
              <el-descriptions-item :label="t('plugin.ops.resultValid')">{{ manifestResult.valid ? t('common.yes') : t('common.no') }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.sourceType')">{{ manifestResult.install_preview?.source_type || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.checksum')"><span class="mono">{{ manifestResult.checksum || '-' }}</span></el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.initialStatus')">{{ pluginStatusLabel(manifestResult.install_preview?.initial_status) }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.dependencies')">{{ (manifestResult.dependencies || []).join(', ') || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.impactSummary')">
                {{ t('plugin.ops.contentTypesCount') }} {{ manifestResult.impact_summary?.content_types_count ?? 0 }}，
                {{ t('plugin.ops.permissionsCount') }} {{ manifestResult.impact_summary?.permissions_count ?? 0 }}，
                {{ t('plugin.ops.menusCount') }} {{ manifestResult.impact_summary?.menus_count ?? 0 }}，
                {{ t('plugin.ops.routesCount') }} {{ manifestResult.impact_summary?.routes_count ?? 0 }}
              </el-descriptions-item>
            </el-descriptions>
            <el-alert v-if="isWizardConfirmStep" :title="t('plugin.wizard.confirmInstallTip')" type="warning" show-icon :closable="false" class="mb" />
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
                  <el-tag v-for="item in manifestResult.migration_plan || []" :key="item.migration_name || item.name" effect="plain">{{ item.migration_name || item.name }} / {{ item.migration_version || item.version }}</el-tag>
                  <span v-if="!(manifestResult.migration_plan || []).length" class="muted">-</span>
                </div>
              </div>
              <div class="result-box">
                <h4>{{ t('plugin.ops.installPreview') }}</h4>
                <pre class="json-box compact">{{ formatJSON(isWizardResultStep ? wizardResult : manifestResult.install_preview || {}) }}</pre>
              </div>
            </div>
          </template>
        </div>
      </section>
    </el-drawer>

    <el-drawer v-model="bulkDialogVisible" append-to-body destroy-on-close size="820px" :with-header="false" class="plugin-action-drawer">
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
  health: 'all',
  contentType: '',
  system: 'all',
  hasSchema: 'all',
});
const loading = ref(false);
const activeView = ref('list');
const healthPanels = ref(['health']);
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
const wizardStep = ref(0);
const wizardValidation = ref({});
const wizardPreview = ref({});
const wizardResult = ref({});
const resultAuditQuery = ref(null);
const bulkDialogVisible = ref(false);
const bulkMode = ref('archive');
const bulkStep = ref('preview');
const bulkLoading = ref(false);
const bulkPreviewRows = ref([]);
const bulkResult = ref({ succeeded: [], failed: [] });
const bulkCodes = ref([]);
const wizardDisplayPayload = computed(() => {
  if (isWizardResultStep.value) return wizardResult.value || {};
  if (isWizardPreviewStep.value) return wizardPreview.value || {};
  return wizardValidation.value || {};
});
const manifestResult = computed(() => wizardDisplayPayload.value?.validation || wizardDisplayPayload.value || {});
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
    if (filters.health && filters.health !== 'all' && (p.health?.status || '') !== filters.health) return false;
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
  const archived = list.filter((p) => p.status === 'archived').length;
  const system = list.filter((p) => p.is_system).length;
  const hasSchema = list.filter((p) => hasConfigSchema(p)).length;
  const abnormal = list.filter((p) => isAbnormalPlugin(p)).length;
  return { total, enabled, disabled, archived, system, hasSchema, abnormal };
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
  if (isWizardResultStep.value) return t('common.close');
  if (manifestDialogAction.value === 'validate' && wizardStep.value === 1) return t('common.close');
  if (manifestDialogAction.value === 'dry-run' && wizardStep.value === 2) return t('common.close');
  if (manifestDialogAction.value === 'upgrade-dry-run' && wizardStep.value === 1) return t('common.close');
  if (isWizardConfirmStep.value) return manifestDialogAction.value === 'install' ? t('plugin.wizard.confirmInstall') : t('plugin.wizard.confirmUpgrade');
  if (manifestDialogAction.value === 'install' && wizardStep.value === 2) return t('common.next');
  if (manifestDialogAction.value === 'upgrade' && wizardStep.value === 1) return t('common.next');
  if ((manifestDialogAction.value === 'dry-run' || manifestDialogAction.value === 'install') && wizardStep.value === 1) return t('plugin.ops.dryRun');
  if (manifestDialogAction.value === 'dry-run') return t('plugin.ops.dryRun');
  if (manifestDialogAction.value === 'upgrade-dry-run') return t('plugin.ops.upgradePreview');
  if (manifestDialogAction.value === 'upgrade') return t('plugin.ops.upgrade');
  if (manifestDialogAction.value === 'install') return t('plugin.ops.install');
  return t('plugin.ops.validateManifest');
});

const wizardSteps = computed(() => {
  if (manifestDialogAction.value === 'validate') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult')];
  if (manifestDialogAction.value === 'dry-run') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult'), t('plugin.wizard.dryRunPreview')];
  if (manifestDialogAction.value === 'install') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult'), t('plugin.wizard.dryRunPreview'), t('plugin.wizard.confirmInstall'), t('plugin.wizard.installResult')];
  if (manifestDialogAction.value === 'upgrade-dry-run') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.compatibilityMatrix')];
  return [t('plugin.wizard.manifestInput'), t('plugin.wizard.compatibilityMatrix'), t('plugin.wizard.confirmUpgrade'), t('plugin.wizard.upgradeResult')];
});

const isWizardConfirmStep = computed(() =>
  (manifestDialogAction.value === 'install' && wizardStep.value === 3)
  || (manifestDialogAction.value === 'upgrade' && wizardStep.value === 2),
);

const isWizardResultStep = computed(() =>
  (manifestDialogAction.value === 'install' && wizardStep.value === 4)
  || (manifestDialogAction.value === 'upgrade' && wizardStep.value === 3),
);

const isWizardPreviewStep = computed(() =>
  (manifestDialogAction.value === 'dry-run' && wizardStep.value === 2)
  || (manifestDialogAction.value === 'install' && wizardStep.value === 2)
  || (manifestDialogAction.value === 'upgrade-dry-run' && wizardStep.value === 1)
  || (manifestDialogAction.value === 'upgrade' && wizardStep.value === 1),
);

const canShortcutConfirm = computed(() =>
  (manifestDialogAction.value === 'install' && (wizardStep.value === 1 || wizardStep.value === 2) && !resultErrors.value.length)
  || (manifestDialogAction.value === 'upgrade' && wizardStep.value === 1 && !upgradeErrors.value.length),
);

const wizardCanProceed = computed(() => {
  if (manifestLoading.value) return false;
  if ((manifestDialogAction.value === 'dry-run' || manifestDialogAction.value === 'install') && wizardStep.value === 1 && resultErrors.value.length) return false;
  if (manifestDialogAction.value === 'install' && wizardStep.value === 2 && resultErrors.value.length) return false;
  if (manifestDialogAction.value === 'upgrade' && wizardStep.value === 1 && upgradeErrors.value.length) return false;
  return true;
});

const resultUpgrade = computed(() => wizardDisplayPayload.value || {});
const upgradeErrors = computed(() => Array.isArray(resultUpgrade.value.validation?.errors) ? resultUpgrade.value.validation.errors : []);
const upgradeWarnings = computed(() => Array.isArray(resultUpgrade.value.validation?.warnings) ? resultUpgrade.value.validation.warnings : []);

const governanceGroups = computed(() => {
  const list = items.value || [];
  return [
    {
      key: 'migration_pending',
      title: t('plugin.governance.migrationPending'),
      tab: 'migrations',
      type: 'warning',
      items: list.filter((p) => p.health?.status === 'migration_pending' || p.health?.migration_status === 'pending'),
      suggestion: t('plugin.governance.migrationPendingTip'),
    },
    {
      key: 'migration_failed',
      title: t('plugin.governance.migrationFailed'),
      tab: 'migrations',
      type: 'danger',
      items: list.filter((p) => p.status === 'migration_failed' || p.health?.migration_status === 'failed' || p.health?.status === 'error'),
      suggestion: t('plugin.governance.migrationFailedTip'),
    },
    {
      key: 'hook_error',
      title: t('plugin.governance.hookError'),
      tab: 'hooks',
      type: 'danger',
      items: list.filter((p) => ['hook_warning', 'hook_error'].includes(p.health?.status) || ['hook_warning', 'hook_error'].includes(p.health?.hook_status)),
      suggestion: t('plugin.governance.hookErrorTip'),
    },
    {
      key: 'config_invalid',
      title: t('plugin.governance.configInvalid'),
      tab: 'config',
      type: 'danger',
      items: list.filter((p) => p.status === 'config_invalid' || p.health?.status === 'config_invalid' || p.health?.config_status === 'invalid'),
      suggestion: t('plugin.governance.configInvalidTip'),
    },
    {
      key: 'dependency_missing',
      title: t('plugin.governance.dependencyMissing'),
      tab: 'runtime',
      type: 'warning',
      items: list.filter((p) => p.status === 'dependency_missing' || p.health?.status === 'dependency_missing' || p.health?.dependency_status === 'missing'),
      suggestion: t('plugin.governance.dependencyMissingTip'),
    },
    {
      key: 'archived',
      title: t('plugin.governance.archived'),
      tab: 'overview',
      type: 'info',
      items: list.filter((p) => p.status === 'archived'),
      suggestion: t('plugin.governance.archivedTip'),
    },
  ];
});

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

function resetFilters() {
  filters.q = '';
  filters.status = 'all';
  filters.health = 'all';
  filters.contentType = '';
  filters.system = 'all';
  filters.hasSchema = 'all';
}

function openManifestDialog(action, row = null) {
  manifestDialogAction.value = action;
  manifestDialogTarget.value = row;
  manifestText.value = manifestTemplate(row);
  wizardStep.value = 0;
  wizardValidation.value = {};
  wizardPreview.value = {};
  wizardResult.value = {};
  manifestDialogVisible.value = true;
}

async function submitManifestAction() {
  if (isWizardResultStep.value) {
    manifestDialogVisible.value = false;
    return;
  }
  let payload;
  try {
    payload = JSON.parse(manifestText.value || '{}');
  } catch {
    ElMessage.error(t('plugin.ops.manifestInvalidJson'));
    return;
  }
  manifestLoading.value = true;
  try {
    if (manifestDialogAction.value === 'validate') {
      if (wizardStep.value === 0) {
        wizardValidation.value = await validatePluginManifest(payload);
        wizardStep.value = 1;
        ElMessage.success(t('plugin.ops.validateDone'));
      } else {
        manifestDialogVisible.value = false;
      }
      return;
    }

    if (manifestDialogAction.value === 'dry-run' || manifestDialogAction.value === 'install') {
      if (wizardStep.value === 0) {
        wizardValidation.value = await validatePluginManifest(payload);
        wizardStep.value = 1;
        ElMessage.success(t('plugin.ops.validateDone'));
        return;
      }
      if (wizardStep.value === 1) {
        if (resultErrors.value.length) return;
        wizardPreview.value = await dryRunPluginManifest(payload);
        wizardStep.value = 2;
        ElMessage.success(t('plugin.ops.dryRunDone'));
        return;
      }
      if (manifestDialogAction.value === 'dry-run') {
        manifestDialogVisible.value = false;
        return;
      }
      if (wizardStep.value === 2) {
        wizardStep.value = 3;
        return;
      }
      if (wizardStep.value === 3) {
        wizardResult.value = await installPluginManifest(payload);
        wizardStep.value = 4;
        ElMessage.success(t('plugin.ops.installDone'));
        await load();
        return;
      }
    }

    if (manifestDialogAction.value === 'upgrade-dry-run' || manifestDialogAction.value === 'upgrade') {
      const targetCode = manifestDialogTarget.value?.code || payload?.code;
      if (wizardStep.value === 0) {
        wizardPreview.value = await dryRunPluginUpgrade(targetCode, payload);
        wizardStep.value = 1;
        ElMessage.success(t('plugin.ops.upgradeDone'));
        return;
      }
      if (manifestDialogAction.value === 'upgrade-dry-run') {
        manifestDialogVisible.value = false;
        return;
      }
      if (wizardStep.value === 1) {
        wizardStep.value = 2;
        return;
      }
      if (wizardStep.value === 2) {
        wizardResult.value = await upgradePlugin(targetCode, payload);
        wizardStep.value = 3;
        ElMessage.success(t('plugin.ops.upgradeExecuted'));
        await load();
      }
    }
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('common.saveFailed')));
  } finally {
    manifestLoading.value = false;
  }
}

function wizardBack() {
  if (wizardStep.value <= 0 || isWizardResultStep.value || manifestLoading.value) return;
  wizardStep.value -= 1;
}

async function confirmWizardAction() {
  let payload;
  try {
    payload = JSON.parse(manifestText.value || '{}');
  } catch {
    ElMessage.error(t('plugin.ops.manifestInvalidJson'));
    return;
  }
  manifestLoading.value = true;
  try {
    if (manifestDialogAction.value === 'install') {
      if (wizardStep.value === 1) {
        wizardPreview.value = await dryRunPluginManifest(payload);
      }
      wizardResult.value = await installPluginManifest(payload);
      wizardStep.value = 4;
      ElMessage.success(t('plugin.ops.installDone'));
      await load();
      return;
    }
    if (manifestDialogAction.value === 'upgrade') {
      const targetCode = manifestDialogTarget.value?.code || payload?.code;
      wizardResult.value = await upgradePlugin(targetCode, payload);
      wizardStep.value = 3;
      ElMessage.success(t('plugin.ops.upgradeExecuted'));
      await load();
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
  bulkMode.value = action;
  bulkStep.value = 'preview';
  bulkResult.value = { succeeded: [], failed: [] };
  bulkCodes.value = selectedRows.value.map((row) => row.code).filter(Boolean);
  bulkPreviewRows.value = [];
  bulkDialogVisible.value = true;
  bulkLoading.value = true;
  try {
    bulkPreviewRows.value = await Promise.all(
      selectedRows.value.map(async (row) => {
        let impact = null;
        try {
          impact = await pluginImpact(row.code);
        } catch {
          impact = null;
        }
        return {
          code: row.code,
          name: row.name,
          status: row.status,
          health: row.health?.status || '',
          contents: impact?.existing_contents_count ?? impact?.topics_count ?? 0,
          communities: impact?.enabled_communities_count ?? 0,
          categories: impact?.categories_count ?? 0,
          migrations: impact?.pending_migrations_count ?? 0,
          hookErrors: impact?.recent_hook_errors_count ?? 0,
        };
      }),
    );
  } finally {
    bulkLoading.value = false;
  }
}

async function confirmBulkAction() {
  if (!bulkCodes.value.length) return;
  bulkLoading.value = true;
  try {
    const response = bulkMode.value === 'archive'
      ? await bulkArchivePlugins({ codes: bulkCodes.value })
      : await bulkRestorePlugins({ codes: bulkCodes.value });
    bulkResult.value = response || { succeeded: [], failed: [] };
    bulkStep.value = 'result';
    resultAuditQuery.value = {
      action: bulkMode.value === 'archive' ? 'plugin.bulk_archived' : 'plugin.bulk_restored',
      target: 'plugins#bulk',
      metadata: bulkCodes.value.join(','),
    };
    selectedRows.value = [];
    ElMessage.success(t('plugin.ops.bulkDone'));
    await load();
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('common.saveFailed')));
  } finally {
    bulkLoading.value = false;
  }
}

function openAuditLogs() {
  const query = resultAuditQuery.value || {
    action: 'plugin.upgraded',
    target: 'plugins',
    metadata: '',
  };
  router.push({ path: '/admin-next/audit-logs', query });
}

function handlePluginCommand(row, command) {
  if (command === 'permissions' || command === 'menus' || command === 'audit' || command === 'hooks' || command === 'migrations' || command === 'runtime') {
    openManifest(row, command);
    return;
  }
  if (command === 'manage') {
    openPlugin(row);
    return;
  }
  if (command === 'upgrade-preview') {
    openManifestDialog('upgrade-dry-run', row);
    return;
  }
  if (command === 'upgrade') {
    openManifestDialog('upgrade', row);
    return;
  }
  if (command === 'enable') {
    setStatus(row, 'enabled');
    return;
  }
  if (command === 'disable') {
    setStatus(row, 'disabled');
    return;
  }
  if (command === 'archive') {
    archive(row);
    return;
  }
  if (command === 'restore') restore(row);
}

function openGovernanceItem(row, tab = 'runtime') {
  openManifest(row, tab);
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

function isAbnormalPlugin(row) {
  const status = row?.health?.status || row?.status;
  return ['error', 'config_invalid', 'migration_failed', 'dependency_missing', 'hook_error'].includes(status);
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
.plugin-page-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; padding: 14px 16px; border: 1px solid #dbe4f0; border-radius: 12px; background: #fff; }
.plugin-page-header h2 { margin: 2px 0 6px; color: #0f172a; }
.plugin-page-header p { margin: 0; color: #64748b; line-height: 1.5; }
.eyebrow { color: #2563eb; font-size: 12px; font-weight: 600; }
.primary-actions { display: flex; gap: 10px; flex-wrap: wrap; justify-content: flex-end; }
.plugin-main-tabs { margin-bottom: -4px; }
.intro-alert { border-radius: 12px; }
.stats-grid { display: grid; grid-template-columns: repeat(5, minmax(150px, 1fr)); gap: 12px; }
.health-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px; }
.stat-card { min-height: 76px; border: 1px solid #e2e8f0; border-radius: 14px; background: #fff; padding: 10px 14px; }
.stat-button { text-align: left; cursor: pointer; appearance: none; font: inherit; }
.stat-button:hover { border-color: #93c5fd; background: #f8fbff; }
.stat-card-danger { border-color: #fecaca; background: #fff7f7; }
.stat-k { color: #64748b; font-size: 12px; }
.stat-v { color: #0f172a; font-size: 22px; font-weight: 700; margin-top: 4px; }
.stat-sub { margin-top: 4px; color: #94a3b8; font-size: 12px; line-height: 1.4; }
.health-collapse { border: 1px solid #dbe4f0; border-radius: 12px; background: #fff; padding: 0 12px; }
.filter-panel { display: flex; justify-content: space-between; align-items: center; gap: 14px; padding: 12px; border: 1px solid #dbe4f0; border-radius: 12px; background: #fff; }
.filter-panel span { display: block; margin-top: 3px; color: #94a3b8; font-size: 12px; }
.filter-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
.filter-actions .el-input { width: 220px; }
.filter-actions .el-select { width: 150px; }
.bulk-panel { display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 10px 12px; border: 1px dashed #cbd5e1; border-radius: 12px; background: #f8fafc; color: #64748b; }
.bulk-actions { display: flex; gap: 8px; }
.mr-6 { margin-right: 6px; }
.mb { margin-bottom: 12px; }
.muted { color: #64748b; }
.plugin-title { display: grid; gap: 2px; margin-bottom: 6px; }
.plugin-title strong { color: #0f172a; }
.plugin-title span { color: #64748b; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.compact-title { margin-bottom: 0; }
.metric-line { display: flex; flex-wrap: wrap; gap: 6px; }
.health-cell { display: grid; gap: 6px; }
.status-stack { display: grid; gap: 6px; justify-items: start; }
.tag-wrap { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
.row-actions { display: flex; align-items: center; gap: 6px; }
.governance-grid { display: grid; gap: 12px; }
.governance-card { border: 1px solid #dbe4f0; border-radius: 12px; background: #fff; padding: 12px; }
.governance-card-head { display: flex; justify-content: space-between; gap: 12px; align-items: flex-start; margin-bottom: 10px; }
.governance-card h3 { margin: 0 0 4px; color: #0f172a; font-size: 16px; }
.governance-card p { margin: 0; color: #64748b; font-size: 13px; }
.governance-reason { margin-top: 5px; font-size: 12px; }
.action-panel {
  border: 1px solid #dbe4f0;
  border-radius: 16px;
  background: #fff;
  padding: 14px;
}
.action-panel.in-drawer {
  border: 0;
  border-radius: 0;
  padding: 0;
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
.wizard-steps { margin-bottom: 16px; }
.result-list { margin: 0; padding-left: 18px; }
.result-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.result-box { border: 1px solid #e2e8f0; border-radius: 12px; padding: 12px; background: #fff; min-width: 0; }
.result-box h4 { margin: 0 0 10px; color: #0f172a; font-size: 14px; }
.page-card :deep(.el-table__cell) { padding: 8px 0; }
.page-card :deep(.el-table .cell) { line-height: 1.35; }
.json-box { margin: 0; padding: 14px; border-radius: 12px; background: #0f172a; color: #dbeafe; max-height: 360px; overflow: auto; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; line-height: 1.55; white-space: pre-wrap; }
.json-box.compact { max-height: 180px; }
:global(.plugin-action-drawer .el-drawer__body) {
  padding: 18px;
  overflow: auto;
}
.action-panel.in-drawer :deep(.el-textarea__inner) {
  min-height: calc(100vh - 150px) !important;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  line-height: 1.55;
}
</style>
