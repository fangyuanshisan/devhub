<template>
  <section class="plugin-page" data-testid="plugin-install-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">安装与治理</div>
        <h2>安装升级</h2>
        <p class="muted">Manifest 校验、dry-run、安装与升级入口。不会执行第三方代码，不支持远程下载与动态加载。</p>
      </div>
      <div class="primary-actions">
        <el-button type="primary" plain data-testid="plugin-manifest-validate" @click="openManifestDialog('validate')">{{ t('plugin.ops.validateManifest') }}</el-button>
        <el-button type="primary" plain data-testid="plugin-manifest-dry-run" @click="openManifestDialog('dry-run')">{{ t('plugin.ops.dryRun') }}</el-button>
        <el-button type="success" plain data-testid="plugin-manifest-install" @click="openManifestDialog('install')">{{ t('plugin.ops.install') }}</el-button>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" class="mb" title="限制：不自动安装依赖、不远程下载、不动态加载、不执行第三方代码。" />

    <div class="filter-panel mb">
      <div>
        <strong>升级目标插件</strong>
        <span class="muted">升级预览/升级需要选择一个已安装插件。</span>
      </div>
      <div class="filter-actions">
        <div data-testid="plugin-upgrade-target-select">
          <el-select v-model="targetCode" filterable clearable placeholder="选择插件" style="min-width: 260px">
            <el-option v-for="p in items" :key="p.code" :label="`${p.name} (${p.code})`" :value="p.code" />
          </el-select>
        </div>
        <el-button type="warning" plain :disabled="!targetCode" data-testid="plugin-upgrade-preview-selected" @click="openManifestDialog('upgrade-dry-run', findPlugin(targetCode))">{{ t('plugin.ops.upgradePreview') }}</el-button>
        <el-button type="danger" plain :disabled="!targetCode" data-testid="plugin-upgrade-selected" @click="openManifestDialog('upgrade', findPlugin(targetCode))">{{ t('plugin.ops.upgrade') }}</el-button>
      </div>
    </div>

    <el-empty v-if="!items.length && !loading" description="暂无插件数据" />
    <el-skeleton v-if="loading" :rows="6" animated />

    <!-- ===== copied/kept from legacy Plugins.vue to avoid E2E regression ===== -->
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
              <el-descriptions-item :label="t('plugin.ops.minCoreVersion')">{{ resultCompatibility(resultUpgrade).min_core_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.compatibleCoreVersion')">{{ resultCompatibility(resultUpgrade).compatible_core_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.changedKeys')" :span="2">{{ (resultUpgrade.changed_keys || []).join(', ') || '-' }}</el-descriptions-item>
            </el-descriptions>
            <div class="result-grid mb">
              <div class="result-box full-width" data-testid="plugin-upgrade-dependency-matrix">
                <h4>{{ t('plugin.dependencies.matrix') }}</h4>
                <div class="tag-wrap mb">
                  <el-tag :type="dependencySummaryType(resultUpgrade.validation?.dependency_summary)" effect="plain">
                    {{ dependencySummaryText(resultUpgrade.validation?.dependency_summary) }}
                  </el-tag>
                  <el-tag :type="compatibilityStatusType(resultCompatibility(resultUpgrade).status)" effect="plain">
                    {{ compatibilityLabel(resultCompatibility(resultUpgrade).status) }}
                  </el-tag>
                </div>
                <el-table :data="dependencyRows(resultUpgrade)" border stripe :empty-text="t('common.none')">
                  <el-table-column prop="code" :label="t('plugin.code')" min-width="140" />
                  <el-table-column prop="version" :label="t('plugin.dependencies.requiredVersion')" min-width="140" />
                  <el-table-column :label="t('plugin.dependencies.required')" width="100">
                    <template #default="{ row }">{{ row.required ? t('plugin.dependencies.requiredDep') : t('plugin.dependencies.optionalDep') }}</template>
                  </el-table-column>
                  <el-table-column :label="t('plugin.status')" width="150">
                    <template #default="{ row }">
                      <el-tag :type="dependencyStatusType(row.status, row.satisfied)" effect="plain">{{ dependencyStatusLabel(row.status) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="current_version" :label="t('plugin.dependencies.currentVersion')" width="130" />
                  <el-table-column prop="message" :label="t('plugin.dependencies.message')" min-width="220" />
                </el-table>
              </div>
              <div class="result-box full-width" data-testid="plugin-upgrade-dependency-diff">
                <h4>{{ t('plugin.dependencies.diff') }}</h4>
                <pre class="json-box compact">{{ formatJSON(resultUpgrade.dependency_diff || {}) }}</pre>
              </div>
            </div>
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
              <el-descriptions-item :label="t('plugin.ops.dependencies')">{{ dependencySummaryText(manifestResult.dependency_summary) }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.compatibilityStatus')">
                <el-tag :type="compatibilityStatusType(manifestResult.compatibility?.status)" effect="plain">{{ compatibilityLabel(manifestResult.compatibility?.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.impactSummary')">
                {{ t('plugin.ops.contentTypesCount') }} {{ manifestResult.impact_summary?.content_types_count ?? 0 }}，
                {{ t('plugin.ops.permissionsCount') }} {{ manifestResult.impact_summary?.permissions_count ?? 0 }}，
                {{ t('plugin.ops.menusCount') }} {{ manifestResult.impact_summary?.menus_count ?? 0 }}，
                {{ t('plugin.ops.routesCount') }} {{ manifestResult.impact_summary?.routes_count ?? 0 }}
              </el-descriptions-item>
            </el-descriptions>
            <div class="result-grid mb">
              <div class="result-box full-width" data-testid="plugin-dependency-summary">
                <h4>{{ t('plugin.dependencies.matrix') }}</h4>
                <div class="tag-wrap mb">
                  <el-tag :type="dependencySummaryType(manifestResult.dependency_summary)" effect="plain">
                    {{ dependencySummaryText(manifestResult.dependency_summary) }}
                  </el-tag>
                  <span class="muted">{{ (manifestResult.compatibility?.messages || []).join('；') || '-' }}</span>
                </div>
                <el-table :data="dependencyRows(manifestResult)" border stripe :empty-text="t('common.none')">
                  <el-table-column prop="code" :label="t('plugin.code')" min-width="140" />
                  <el-table-column prop="version" :label="t('plugin.dependencies.requiredVersion')" min-width="140" />
                  <el-table-column :label="t('plugin.dependencies.required')" width="100">
                    <template #default="{ row }">{{ row.required ? t('plugin.dependencies.requiredDep') : t('plugin.dependencies.optionalDep') }}</template>
                  </el-table-column>
                  <el-table-column :label="t('plugin.status')" width="150">
                    <template #default="{ row }">
                      <el-tag :type="dependencyStatusType(row.status, row.satisfied)" effect="plain">{{ dependencyStatusLabel(row.status) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="current_version" :label="t('plugin.dependencies.currentVersion')" width="130" />
                  <el-table-column prop="message" :label="t('plugin.dependencies.message')" min-width="220" />
                </el-table>
              </div>
            </div>
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
    <!-- ===== end legacy wizard ===== -->
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  dryRunPluginManifest,
  dryRunPluginUpgrade,
  installPluginManifest,
  plugins,
  upgradePlugin,
  validatePluginManifest,
} from '@/api/admin';
import { t } from '@/i18n';
import { pluginStatusLabel } from '@/i18n/formatters';

const items = ref([]);
const loading = ref(false);
const targetCode = ref('');

onMounted(load);

async function load() {
  loading.value = true;
  try {
    const list = await plugins();
    items.value = list.items || [];
  } finally {
    loading.value = false;
  }
}

function findPlugin(code) {
  return (items.value || []).find((p) => (p.code || p.plugin_code) === code) || null;
}

// === legacy wizard state/methods (copied with minimal adjustments) ===
const manifestDialogVisible = ref(false);
const manifestDialogAction = ref('validate');
const manifestDialogTarget = ref(null);
const manifestText = ref('');
const manifestLoading = ref(false);
const wizardStep = ref(0);
const wizardValidation = ref({});
const wizardPreview = ref({});
const wizardResult = ref({});
const wizardDisplayPayload = computed(() => {
  if (isWizardResultStep.value) return wizardResult.value || {};
  if (isWizardPreviewStep.value) return wizardPreview.value || {};
  return wizardValidation.value || {};
});
const manifestResult = computed(() => wizardDisplayPayload.value?.validation || wizardDisplayPayload.value || {});
const resultErrors = computed(() => Array.isArray(manifestResult.value?.errors) ? manifestResult.value.errors : []);
const resultWarnings = computed(() => Array.isArray(manifestResult.value?.warnings) ? manifestResult.value.warnings : []);

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

const wizardSteps = computed(() => {
  if (manifestDialogAction.value === 'validate') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult')];
  if (manifestDialogAction.value === 'dry-run') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult'), t('plugin.wizard.dryRunPreview')];
  if (manifestDialogAction.value === 'install') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult'), t('plugin.wizard.dryRunPreview'), t('plugin.wizard.confirmInstall'), t('plugin.wizard.installResult')];
  if (manifestDialogAction.value === 'upgrade-dry-run') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.compatibilityMatrix')];
  if (manifestDialogAction.value === 'upgrade') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.compatibilityMatrix'), t('plugin.wizard.confirmUpgrade'), t('plugin.wizard.upgradeResult')];
  return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult')];
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

const canShortcutConfirm = computed(() => false);

const manifestDialogActionLabel = computed(() => {
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

const wizardCanProceed = computed(() => {
  if (manifestLoading.value) return false;
  if ((manifestDialogAction.value === 'dry-run' || manifestDialogAction.value === 'install') && wizardStep.value === 1 && resultErrors.value.length) return false;
  if (manifestDialogAction.value === 'install' && wizardStep.value === 2 && resultErrors.value.length) return false;
  return true;
});

function openManifestDialog(action, row = null) {
  manifestDialogAction.value = action;
  manifestDialogTarget.value = row;
  wizardStep.value = 0;
  wizardValidation.value = {};
  wizardPreview.value = {};
  wizardResult.value = {};
  manifestDialogVisible.value = true;
}

function wizardBack() {
  wizardStep.value = Math.max(0, wizardStep.value - 1);
}

async function submitManifestAction() {
  manifestLoading.value = true;
  try {
    let manifest;
    try {
      manifest = JSON.parse(manifestText.value || '{}');
    } catch {
      ElMessage.error(t('plugin.ops.manifestInvalidJson'));
      return;
    }

    if (manifestDialogAction.value === 'validate') {
      wizardValidation.value = await validatePluginManifest({ manifest });
      wizardStep.value = 1;
      return;
    }
    if (manifestDialogAction.value === 'dry-run') {
      if (wizardStep.value === 0) {
        wizardValidation.value = await validatePluginManifest({ manifest });
        wizardStep.value = 1;
        return;
      }
      if (wizardStep.value === 1) {
        wizardPreview.value = await dryRunPluginManifest({ manifest });
        wizardStep.value = 2;
        return;
      }
      manifestDialogVisible.value = false;
      return;
    }
    if (manifestDialogAction.value === 'install') {
      if (wizardStep.value === 0) {
        wizardValidation.value = await validatePluginManifest({ manifest });
        wizardStep.value = 1;
        return;
      }
      if (wizardStep.value === 1) {
        wizardPreview.value = await dryRunPluginManifest({ manifest });
        wizardStep.value = 2;
        return;
      }
      if (wizardStep.value === 2) {
        wizardStep.value = 3;
        return;
      }
      if (wizardStep.value === 3) {
        wizardResult.value = await installPluginManifest({ manifest });
        wizardStep.value = 4;
        await load();
        return;
      }
      manifestDialogVisible.value = false;
      return;
    }
    if (manifestDialogAction.value === 'upgrade-dry-run') {
      const code = manifestDialogTarget.value?.code || targetCode.value;
      wizardPreview.value = await dryRunPluginUpgrade(code, { manifest });
      wizardStep.value = 1;
      return;
    }
    if (manifestDialogAction.value === 'upgrade') {
      const code = manifestDialogTarget.value?.code || targetCode.value;
      if (wizardStep.value === 0) {
        wizardPreview.value = await dryRunPluginUpgrade(code, { manifest });
        wizardStep.value = 1;
        return;
      }
      if (wizardStep.value === 1) {
        wizardStep.value = 2;
        return;
      }
      if (wizardStep.value === 2) {
        await ElMessageBox.confirm(t('plugin.ops.upgradeConfirmTip'), t('plugin.ops.upgrade'), { type: 'warning' });
        wizardResult.value = await upgradePlugin(code, { manifest });
        wizardStep.value = 3;
        await load();
        return;
      }
      manifestDialogVisible.value = false;
    }
  } finally {
    manifestLoading.value = false;
  }
}

function confirmWizardAction() {}

// helpers copied from Plugins.vue
function formatJSON(obj) {
  try {
    return JSON.stringify(obj || {}, null, 2);
  } catch {
    return String(obj || '');
  }
}

function compatibilityLabel(status) {
  if (status === 'compatible') return 'compatible';
  if (status === 'warning') return 'warning';
  if (status === 'incompatible') return 'incompatible';
  return status || '-';
}

function resultCompatibility(result) {
  return result?.compatibility || {};
}

function compatibilityStatusType(status) {
  if (status === 'compatible') return 'success';
  if (status === 'warning') return 'warning';
  if (status === 'incompatible') return 'danger';
  return 'info';
}

function dependencyRows(result) {
  return result?.validation?.dependencies || result?.dependencies || [];
}

function dependencySummaryType(summary) {
  if (summary === 'blocked') return 'danger';
  if (summary === 'warning') return 'warning';
  if (summary === 'pass') return 'success';
  return 'info';
}

function dependencySummaryText(summary) {
  if (summary === 'blocked') return 'blocked';
  if (summary === 'warning') return 'warning';
  if (summary === 'pass') return 'pass';
  return summary || '-';
}

function dependencyStatusType(status, satisfied) {
  if (satisfied) return 'success';
  if (status === 'optional_missing') return 'warning';
  if (status === 'version_mismatch') return 'danger';
  if (status === 'missing') return 'danger';
  return 'info';
}

function dependencyStatusLabel(status) {
  return status || '-';
}
</script>
