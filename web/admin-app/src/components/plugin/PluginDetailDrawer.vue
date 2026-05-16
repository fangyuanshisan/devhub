<template>
  <el-drawer v-model="visible" :title="title" size="920px" data-testid="plugin-detail-drawer" class="plugin-detail-drawer">
    <template v-if="plugin">
      <div class="drawer-content">
        <div class="hero">
        <div class="hero-left">
          <div class="hero-title">
            <h3>{{ plugin.name }}</h3>
            <el-tag :type="statusType(plugin.status)">{{ pluginStatusLabel(plugin.status) }}</el-tag>
            <el-tag :type="healthType(plugin.health?.status)">{{ pluginHealthLabel(plugin.health?.status) }}</el-tag>
            <el-tag v-if="plugin.is_system" type="primary">{{ t('plugin.system') }}</el-tag>
          </div>
          <p class="hero-desc">{{ plugin.description || t('plugin.noDescription') }}</p>
          <div class="hero-metrics">
            <el-tag type="info" effect="plain">{{ t('plugin.contentTypes') }} {{ (plugin.content_types || []).length }}</el-tag>
            <el-tag type="info" effect="plain">{{ t('plugin.capability.permissions') }} {{ (plugin.permissions || []).length }}</el-tag>
            <el-tag type="info" effect="plain">{{ t('plugin.capability.menus') }} {{ (plugin.menus || []).length }}</el-tag>
            <el-tag :type="(plugin.hooks || []).length ? 'success' : 'info'" effect="plain">{{ t('plugin.capability.hooks') }} {{ (plugin.hooks || []).length }}</el-tag>
          </div>
        </div>
        <div class="hero-right">
          <div class="code-pill">{{ plugin.code }}</div>
          <div class="meta-line">{{ t('plugin.version') }}: {{ plugin.version }}</div>
        </div>
        </div>

        <el-alert
          v-if="plugin.status_reason || plugin.health?.suggested_action || plugin.status === 'archived'"
          :title="t('plugin.runtime.bannerTitle')"
          :type="healthType(plugin.health?.status || plugin.status)"
          show-icon
          :closable="false"
          class="mt"
        >
          <template #default>
            <div class="banner-lines">
              <div><strong>{{ t('plugin.runtime.statusReason') }}：</strong>{{ plugin.status_reason || plugin.health?.status_reason || '-' }}</div>
              <div><strong>{{ t('plugin.runtime.suggestedAction') }}：</strong>{{ plugin.health?.suggested_action || '-' }}</div>
              <div v-if="plugin.status === 'archived'">{{ t('plugin.runtime.archivedTip') }}</div>
            </div>
          </template>
        </el-alert>

        <el-tabs v-model="tab" class="tabs" data-testid="plugin-detail-tabs">
        <el-tab-pane :label="t('plugin.tabs.overview')" name="overview">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('field.name')">{{ plugin.name }}</el-descriptions-item>
            <el-descriptions-item :label="t('field.plugin_code')">{{ plugin.code }}</el-descriptions-item>
            <el-descriptions-item :label="t('field.version')">{{ plugin.version }}</el-descriptions-item>
            <el-descriptions-item :label="t('field.status')">
              <el-tag :type="statusType(plugin.status)">{{ pluginStatusLabel(plugin.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('field.health')">
              <el-tag :type="healthType(plugin.health?.status)">{{ pluginHealthLabel(plugin.health?.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('plugin.isSystem')">{{ plugin.is_system ? t('common.yes') : t('common.no') }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.maturity')">{{ maturityLabel(plugin) }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.suggestedAction')">{{ plugin.health?.suggested_action || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.lifecycle.installStatus')">{{ pluginStatusLabel(plugin.install_status) }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.lifecycle.lifecycleStatus')">{{ pluginStatusLabel(plugin.lifecycle_status) }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.lifecycle.installedAt')">{{ plugin.installed_at || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.lifecycle.archivedAt')">{{ plugin.archived_at || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.lifecycle.statusReason')" :span="2">{{ plugin.status_reason || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.contentTypes')" :span="2">{{ (plugin.content_types || []).join(', ') || '-' }}</el-descriptions-item>
          </el-descriptions>
          <el-alert
            class="mt"
            type="info"
            show-icon
            :closable="false"
            :title="t('plugin.maturityNote')"
          />
          <section class="export-panel mt" data-testid="plugin-export-panel">
            <div>
              <h4>导出本地插件包</h4>
              <p>导出声明型插件包：manifest、README、config.example.json、checksums，不包含敏感配置、用户数据、运行时代码或外部 SQL。</p>
            </div>
            <el-button type="primary" plain data-testid="plugin-export-open" @click="openExportDialog">导出本地插件包</el-button>
          </section>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.runtime')" name="runtime">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            class="mb"
            :title="t('plugin.runtimeNote')"
          />
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('plugin.runtime.overallStatus')">
              <el-tag :type="healthType(plugin.health?.status)">{{ pluginHealthLabel(plugin.health?.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.suggestedAction')">{{ plugin.health?.suggested_action || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.configStatus')">
              <el-tag :type="metricType(plugin.health?.config_status)">{{ pluginHealthLabel(plugin.health?.config_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.migrationStatus')">
              <el-tag :type="metricType(plugin.health?.migration_status)">{{ pluginHealthLabel(plugin.health?.migration_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.hookStatus')">
              <el-tag :type="metricType(plugin.health?.hook_status)">{{ pluginHealthLabel(plugin.health?.hook_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.dependencyStatus')">
              <el-tag :type="metricType(plugin.health?.dependency_status)">{{ pluginHealthLabel(plugin.health?.dependency_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.pendingMigrations')">{{ plugin.health?.pending_migrations_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.failedMigrations')">{{ plugin.health?.failed_migrations_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.hookFailures')">{{ plugin.health?.hook_failure_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item :label="t('field.updated_at')">{{ plugin.health?.updated_at || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.recentError')" :span="2">{{ plugin.health?.recent_error || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('plugin.runtime.statusReason')" :span="2">{{ plugin.health?.status_reason || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.readiness')" name="readiness">
          <div class="sub-toolbar">
            <el-tag :type="readinessTagType(readinessResult?.status)" effect="plain">
              {{ t('plugin.readiness.overall') }}：{{ readinessStatusLabel(readinessResult?.status) }}
            </el-tag>
            <el-button size="small" :loading="readinessLoading" data-testid="plugin-readiness-refresh" @click="loadReadiness">{{ t('common.refresh') }}</el-button>
            <el-button
              v-if="canRunEnablePrecheck"
              size="small"
              type="primary"
              plain
              :loading="enablePrecheckLoading"
              data-testid="plugin-enable-precheck-run"
              @click="runEnablePrecheck"
            >
              {{ t('plugin.ops.enablePrecheck') }}
            </el-button>
          </div>
          <el-alert type="info" show-icon :closable="false" class="mb" :title="t('plugin.ops.enablePrecheckTip')" />
          <el-alert
            v-if="enablePrecheckResult"
            :title="t('plugin.ops.enablePrecheckResult')"
            :type="enablePrecheckResult.can_enable ? 'success' : 'error'"
            show-icon
            :closable="false"
            class="mb"
          >
            <template #default>
              <div class="banner-lines">
                <div><strong>status：</strong>{{ enablePrecheckResult.status }}</div>
                <div><strong>can_enable：</strong>{{ enablePrecheckResult.can_enable }}</div>
                <div v-if="(enablePrecheckResult.errors || []).length"><strong>errors：</strong>{{ (enablePrecheckResult.errors || []).join('；') }}</div>
                <div v-if="(enablePrecheckResult.warnings || []).length"><strong>warnings：</strong>{{ (enablePrecheckResult.warnings || []).join('；') }}</div>
              <div v-if="enablePrecheckResult.id && enablePrecheckResult.can_enable">
                  <el-button type="primary" size="small" :loading="enableTaskLoading" data-testid="plugin-enable-from-precheck" @click="enableFromPrecheck">
                    {{ t('plugin.ops.enableFromPrecheck') }}
                  </el-button>
                </div>
              </div>
            </template>
          </el-alert>

          <div class="mt" data-testid="plugin-upgrade-panel">
            <el-alert type="info" show-icon :closable="false" class="mb" :title="t('plugin.ops.packageUpgradeTip')" />
            <div class="sub-toolbar">
              <el-select
                v-model="selectedUpgradeCompatCheckID"
                size="small"
                style="max-width: 340px"
                data-testid="plugin-upgrade-compat-select"
                :placeholder="t('plugin.ops.packageUpgradeSelectPlaceholder')"
              >
                <el-option v-for="it in upgradeCompatChecks" :key="`upgrade-compat-${it.id}`" :value="it.id" :label="upgradeCompatLabel(it)" />
              </el-select>
              <el-button size="small" :loading="upgradeLoading" data-testid="plugin-upgrade-load-impact" @click="loadUpgradeImpact">
                {{ t('plugin.ops.packageUpgradeLoadImpact') }}
              </el-button>
              <el-button
                v-if="canRunPackageUpgrade"
                size="small"
                type="warning"
                plain
                :disabled="!selectedUpgradeCompatCheckID"
                :loading="upgradeLoading"
                data-testid="plugin-upgrade-run"
                @click="runPackageUpgrade"
              >
                {{ t('plugin.ops.packageUpgrade') }}
              </el-button>
              <el-button size="small" :loading="upgradeTasksLoading" data-testid="plugin-upgrade-tasks-refresh" @click="loadUpgradeTasks">
                {{ t('plugin.ops.packageUpgradeTasks') }}
              </el-button>
            </div>

            <el-alert
              v-if="upgradeImpact"
              :title="t('plugin.ops.packageUpgradeImpactTitle')"
              :type="upgradeImpact.can_upgrade ? 'success' : 'error'"
              show-icon
              :closable="false"
              class="mb"
            >
              <template #default>
                <div class="banner-lines">
                  <div><strong>can_upgrade：</strong>{{ upgradeImpact.can_upgrade }}</div>
                  <div><strong>old_version：</strong>{{ upgradeImpact.old_version }}</div>
                  <div><strong>new_version：</strong>{{ upgradeImpact.new_version }}</div>
                  <div v-if="(upgradeImpact.errors || []).length"><strong>errors：</strong>{{ (upgradeImpact.errors || []).join('；') }}</div>
                  <div v-if="(upgradeImpact.warnings || []).length"><strong>warnings：</strong>{{ (upgradeImpact.warnings || []).join('；') }}</div>
                  <div v-if="upgradeImpact.manifest_diff_summary">
                    <strong>diff_summary：</strong>
                    added={{ upgradeImpact.manifest_diff_summary.added ?? 0 }},
                    removed={{ upgradeImpact.manifest_diff_summary.removed ?? 0 }},
                    changed={{ upgradeImpact.manifest_diff_summary.changed ?? 0 }},
                    high_risk={{ upgradeImpact.manifest_diff_summary.high_risk ?? 0 }},
                    blocked={{ upgradeImpact.manifest_diff_summary.blocked ?? 0 }}
                  </div>
                </div>
              </template>
            </el-alert>

            <el-alert
              v-if="upgradeResult"
              :title="t('plugin.ops.packageUpgradeResultTitle')"
              :type="upgradeResult.status === 'upgraded' ? 'success' : 'warning'"
              show-icon
              :closable="false"
              class="mb"
            >
              <template #default>
                <div class="banner-lines">
                  <div><strong>status：</strong>{{ upgradeResult.status }}</div>
                  <div><strong>old_version：</strong>{{ upgradeResult.old_version }}</div>
                  <div><strong>new_version：</strong>{{ upgradeResult.new_version }}</div>
                  <div><strong>new_plugin_status：</strong>{{ upgradeResult.new_plugin_status || '-' }}</div>
                  <div v-if="(upgradeResult.errors || []).length"><strong>errors：</strong>{{ (upgradeResult.errors || []).join('；') }}</div>
                  <div v-if="(upgradeResult.warnings || []).length"><strong>warnings：</strong>{{ (upgradeResult.warnings || []).join('；') }}</div>
                </div>
              </template>
            </el-alert>

            <el-table
              v-loading="upgradeTasksLoading"
              :data="upgradeTasks"
              border
              stripe
              size="small"
              data-testid="plugin-upgrade-tasks-table"
              :empty-text="t('plugin.ops.packageUpgradeTasksEmpty')"
            >
              <el-table-column prop="id" label="id" width="90" />
              <el-table-column prop="status" :label="t('field.status')" width="150" />
              <el-table-column prop="old_version" :label="t('plugin.ops.currentVersion')" width="140" />
              <el-table-column prop="new_version" :label="t('plugin.ops.newVersion')" width="140" />
              <el-table-column prop="created_at" :label="t('field.created_at')" width="180" />
              <el-table-column :label="t('plugin.action')" width="140" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" :disabled="!row?.id" @click="openUpgradeTask(row.id)">{{ t('common.detail') }}</el-button>
                  <el-button link type="warning" :disabled="row?.status !== 'failed'" @click="retryUpgradeTask(row.id)">{{ t('common.retry') }}</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div class="mt" data-testid="plugin-soft-uninstall-panel">
            <el-alert type="warning" show-icon :closable="false" class="mb" :title="t('plugin.ops.softUninstallTip')" />
            <div class="sub-toolbar">
              <el-input v-model="softUninstallReason" :placeholder="t('plugin.ops.softUninstallReasonPlaceholder')" size="small" style="max-width: 420px" />
              <el-button
                v-if="canSoftUninstall"
                size="small"
                type="danger"
                plain
                :loading="softUninstallLoading"
                data-testid="plugin-soft-uninstall-run"
                @click="runSoftUninstall"
              >
                {{ t('plugin.ops.softUninstall') }}
              </el-button>
            </div>
            <el-alert
              v-if="softUninstallResult"
              :title="t('plugin.ops.softUninstallResult')"
              :type="softUninstallResult.status === 'soft_uninstalled' ? 'success' : 'warning'"
              show-icon
              :closable="false"
              class="mb"
            >
              <template #default>
                <div class="banner-lines">
                  <div><strong>status：</strong>{{ softUninstallResult.status }}</div>
                  <div><strong>previous_status：</strong>{{ softUninstallResult.previous_status || '-' }}</div>
                  <div><strong>new_status：</strong>{{ softUninstallResult.new_status || '-' }}</div>
                  <div v-if="(softUninstallResult.errors || []).length"><strong>errors：</strong>{{ (softUninstallResult.errors || []).join('；') }}</div>
                  <div v-if="(softUninstallResult.warnings || []).length"><strong>warnings：</strong>{{ (softUninstallResult.warnings || []).join('；') }}</div>
                </div>
              </template>
            </el-alert>
          </div>
          <el-table v-loading="readinessLoading" :data="readinessResult?.checks || []" border stripe data-testid="plugin-readiness-table">
            <el-table-column prop="title" :label="t('field.name')" min-width="220" />
            <el-table-column :label="t('field.status')" width="120">
              <template #default="{ row }">
                <el-tag :type="readinessTagType(row.status)" effect="plain">{{ readinessStatusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="code" :label="t('plugin.ops.errorCode')" width="220">
              <template #default="{ row }"><span class="mono">{{ row.code || '-' }}</span></template>
            </el-table-column>
            <el-table-column prop="reason" :label="t('plugin.readiness.reason')" min-width="240" />
            <el-table-column prop="suggestion" :label="t('plugin.readiness.suggestion')" min-width="260" />
            <el-table-column :label="t('plugin.action')" width="120">
              <template #default="{ row }">
                <el-button v-if="row.dependency_code" link type="primary" @click="emit('open-plugin', row.dependency_code)">{{ t('plugin.dependencies.openPlugin') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.dependencies')" name="dependencies">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            class="mb"
            :title="t('plugin.dependencies.detailTip')"
          />
          <div class="sub-toolbar">
            <el-tag type="info" effect="plain">{{ t('common.total') }} {{ dependencyRows.length }}</el-tag>
            <el-tag type="danger" effect="plain">{{ t('plugin.dependencies.requiredDep') }} {{ dependencyRows.filter((row) => row.required).length }}</el-tag>
            <el-tag type="warning" effect="plain">{{ t('plugin.dependencies.optionalDep') }} {{ dependencyRows.filter((row) => !row.required).length }}</el-tag>
            <el-tag :type="metricType(plugin.health?.dependency_status)" effect="plain">
              {{ t('plugin.runtime.dependencyStatus') }}：{{ pluginHealthLabel(plugin.health?.dependency_status) }}
            </el-tag>
          </div>
          <el-table :data="dependencyRows" border stripe :empty-text="`暂无${t('plugin.tabs.dependencies')}`" data-testid="plugin-dependencies-table">
            <el-table-column prop="code" :label="t('plugin.code')" min-width="160">
              <template #default="{ row }">
                <span class="mono">{{ row.code }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="pluginName" :label="t('field.name')" min-width="160" />
            <el-table-column prop="currentVersion" :label="t('plugin.dependencies.currentVersion')" width="130" />
            <el-table-column prop="version" :label="t('plugin.dependencies.requiredVersion')" width="150" />
            <el-table-column :label="t('plugin.dependencies.required')" width="110">
              <template #default="{ row }">
                <el-tag :type="row.required ? 'danger' : 'warning'" effect="plain">
                  {{ row.required ? t('plugin.dependencies.requiredDep') : t('plugin.dependencies.optionalDep') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.status')" width="150">
              <template #default="{ row }">
                <el-tag :type="dependencyStatusType(row.status, row.satisfied)" effect="plain">{{ dependencyStatusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" :label="t('plugin.dependencies.reason')" min-width="220" />
            <el-table-column prop="message" :label="t('plugin.dependencies.message')" min-width="220" />
            <el-table-column :label="t('plugin.action')" width="140" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" :disabled="!row.code" @click="emit('open-plugin', row.code)">{{ t('plugin.dependencies.openPlugin') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.contentTypes')" name="contentTypes">
          <el-table :data="plugin.content_type_definitions || []" border stripe :empty-text="`暂无${t('plugin.tabs.contentTypes')}`">
            <el-table-column prop="type" :label="t('field.type')" width="140" />
            <el-table-column prop="name" :label="t('field.name')" width="140" />
            <el-table-column prop="plugin_code" :label="t('field.plugin_code')" width="120" />
            <el-table-column prop="create_permission" :label="t('plugin.contentTypeDefinition.createPermission')" min-width="180" />
            <el-table-column prop="edit_permission" :label="t('plugin.contentTypeDefinition.editPermission')" min-width="160" />
            <el-table-column prop="delete_permission" :label="t('plugin.contentTypeDefinition.deletePermission')" min-width="160" />
            <el-table-column prop="audit_permission" :label="t('plugin.contentTypeDefinition.auditPermission')" min-width="160" />
            <el-table-column prop="seo_type" :label="t('plugin.contentTypeDefinition.seoType')" width="130" />
            <el-table-column :label="t('plugin.contentTypeDefinition.flags')" width="260">
              <template #default="{ row }">
                <el-tag size="small" effect="plain" :type="row.allow_comment ? 'success' : 'info'">{{ t('plugin.contentTypeDefinition.allowComment') }}</el-tag>
                <el-tag size="small" effect="plain" :type="row.allow_like ? 'success' : 'info'" class="ml">{{ t('plugin.contentTypeDefinition.allowLike') }}</el-tag>
                <el-tag size="small" effect="plain" :type="row.allow_favorite ? 'success' : 'info'" class="ml">{{ t('plugin.contentTypeDefinition.allowFavorite') }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.permissions')" name="permissions">
          <div class="sub-toolbar">
            <el-input v-model="permQ" :placeholder="t('plugin.permissionsSearchPlaceholder')" clearable style="max-width: 320px" />
            <el-checkbox v-model="permOnlyMissing" class="ml">{{ t('plugin.permissionsOnlyMissing') }}</el-checkbox>
            <el-checkbox v-model="permOnlyHighRisk" class="ml">{{ t('plugin.permissionsOnlyHighRisk') }}</el-checkbox>
          </div>
          <el-table :data="filteredPermissions" border stripe :empty-text="`暂无${t('plugin.tabs.permissions')}`" :row-class-name="permissionRowClass">
            <el-table-column prop="code" :label="t('field.code')" min-width="240">
              <template #default="{ row }">
                <div class="mono">{{ row.code }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="name" :label="t('field.name')" min-width="160" />
            <el-table-column prop="plugin_code" :label="t('field.plugin_code')" width="120" />
            <el-table-column prop="scope" :label="t('field.scope')" width="150" />
            <el-table-column :label="t('plugin.permissionsOpType')" width="130">
              <template #default="{ row }">
                <el-tag size="small" effect="plain" :type="row._opTypeTag">{{ row._opTypeLabel }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.permissionsOwned')" width="160">
              <template #default="{ row }">
                <el-tag size="small" effect="plain" :type="row._hasPermission ? 'success' : 'danger'">
                  {{ row._hasPermission ? t('common.yes') : t('common.no') }}
                </el-tag>
                <el-tag v-if="row._highRisk" size="small" effect="plain" type="warning" class="ml">{{ t('plugin.permissionsHighRisk') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="description" :label="t('field.description')" min-width="220" />
            <el-table-column :label="t('plugin.action')" width="140" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="copyText(row.code)">{{ t('common.copy') }}</el-button>
                <el-button link type="primary" @click="openPermissionRefs(row)">{{ t('plugin.permissionsRefs') }}</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-dialog v-model="permissionRefsDialog" :title="t('plugin.permissionsRefsTitle')" width="720px">
            <div v-if="selectedPermission" class="mb">
              <div class="mono">{{ selectedPermission.code }}</div>
              <div class="muted">{{ selectedPermission.name || '-' }}</div>
            </div>
            <el-alert v-if="selectedPermission && !selectedPermission._hasPermission" type="warning" show-icon :closable="false" class="mb" :title="t('plugin.permissionsMissingImpactTip')" />
            <el-table :data="selectedPermissionRefs" border stripe :empty-text="t('plugin.permissionsNoRefs')">
              <el-table-column prop="type" :label="t('field.type')" width="130" />
              <el-table-column prop="title" :label="t('field.name')" min-width="220" />
              <el-table-column prop="path" :label="t('field.path')" min-width="260" />
            </el-table>
            <template #footer>
              <el-button @click="permissionRefsDialog = false">{{ t('common.close') }}</el-button>
            </template>
          </el-dialog>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.menus')" name="menus">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            :title="t('plugin.menuVisibilityNote')"
            class="mb"
          />
          <el-table :data="plugin.menus || []" border stripe :empty-text="`暂无${t('plugin.tabs.menus')}`">
            <el-table-column prop="area" :label="t('field.area')" width="120" />
            <el-table-column prop="title" :label="t('field.title')" width="160" />
            <el-table-column prop="path" :label="t('field.path')" min-width="220" />
            <el-table-column prop="permission" :label="t('field.permission')" min-width="200" />
            <el-table-column prop="sort_order" :label="t('field.sort_order')" width="120" />
          </el-table>

          <section class="mt" data-testid="plugin-frontend-menus-preview">
            <div class="sub-toolbar">
              <strong>{{ t('plugin.menuPreviewTitle') }}</strong>
              <el-button size="small" :loading="menuPreviewLoading" data-testid="plugin-menu-preview-refresh" @click="loadMenuPreview">{{ t('plugin.menuPreviewRefresh') }}</el-button>
            </div>
            <el-alert type="info" show-icon :closable="false" class="mb" :title="t('plugin.menuPreviewTip')" />
            <div class="sub-toolbar">
              <el-input v-model="menuPreviewParams.community_slug" :placeholder="t('plugin.menuPreviewCommunity')" clearable style="max-width: 240px" />
              <el-input v-model="menuPreviewParams.category_id" :placeholder="t('plugin.menuPreviewCategory')" clearable style="max-width: 200px" class="ml" />
            </div>
            <el-table v-loading="menuPreviewLoading" :data="menuPreviewRows" border stripe :empty-text="t('common.none')">
              <el-table-column prop="code" :label="t('field.code')" width="180">
                <template #default="{ row }"><span class="mono">{{ row.code || '-' }}</span></template>
              </el-table-column>
              <el-table-column prop="location" :label="t('plugin.menuPreviewLocation')" width="150" />
              <el-table-column prop="title" :label="t('field.title')" min-width="180" />
              <el-table-column prop="route" :label="t('plugin.menuPreviewRoute')" min-width="220">
                <template #default="{ row }"><span class="mono">{{ row.route || '-' }}</span></template>
              </el-table-column>
              <el-table-column prop="content_type" :label="t('plugin.contentType')" width="140">
                <template #default="{ row }"><span class="mono">{{ row.content_type || '-' }}</span></template>
              </el-table-column>
              <el-table-column prop="required_permission" :label="t('field.permission')" min-width="200">
                <template #default="{ row }"><span class="mono">{{ row.required_permission || '-' }}</span></template>
              </el-table-column>
              <el-table-column :label="t('plugin.menuPreviewVisible')" width="110">
                <template #default="{ row }">
                  <el-tag :type="row.visible ? 'success' : 'danger'" effect="plain">{{ row.visible ? t('common.yes') : t('common.no') }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.menuPreviewReason')" min-width="220">
                <template #default="{ row }">
                  <div v-if="!row.visible">
                    <div class="muted">{{ row.reason || '-' }}</div>
                    <div v-if="row.reason_code" class="mono">{{ row.reason_code }}</div>
                  </div>
                  <span v-else class="muted">-</span>
                </template>
              </el-table-column>
            </el-table>
          </section>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.config')" name="config">
          <el-alert
            :title="t('plugin.configCapabilityNote')"
            type="info"
            show-icon
            :closable="false"
            class="mb"
          />

          <el-collapse v-model="configPanels">
            <el-collapse-item name="schema" :title="t('plugin.config.schema')">
              <pre class="json-box">{{ formatJSON(plugin.config_schema || {}) }}</pre>
            </el-collapse-item>
            <el-collapse-item name="resolved" :title="t('plugin.config.resolvedPanel')">
              <pre class="json-box">{{ formatJSON(plugin.resolved_config || {}) }}</pre>
            </el-collapse-item>
          </el-collapse>

          <section class="config-card" data-testid="plugin-global-config-panel">
            <div class="config-card-header">
              <div>
                <h4>{{ t('plugin.config.globalPanel') }}</h4>
                <p>{{ t('plugin.config.globalTip') }}</p>
              </div>
              <div class="config-card-tools">
                <el-tag :type="schemaErrors.length ? 'danger' : 'success'" effect="plain">
                  {{ schemaErrors.length ? t('plugin.capability.schemaInvalid') : t('plugin.capability.schemaValid') }}
                </el-tag>
                <el-button size="small" data-testid="plugin-config-versions-open" @click="configVersionsVisible = true">版本历史</el-button>
                <el-button size="small" @click="reloadConfig">{{ t('plugin.config.resetCurrent') }}</el-button>
                <el-button size="small" data-testid="plugin-global-config-clear" @click="clearGlobalConfig">{{ t('common.clearObject') }}</el-button>
                <el-button size="small" type="primary" data-testid="plugin-global-config-save" :disabled="schemaErrors.length > 0" @click="saveConfig">{{ t('common.save') }}</el-button>
              </div>
            </div>
            <PluginConfigEditor
              v-model="editableConfig"
              :schema="plugin.config_schema || null"
              :default-config="plugin.resolved_config?.default || {}"
              :original-config="jsonValue(plugin.config_json)"
              :effective-config="plugin.resolved_config?.effective || plugin.resolved_config || {}"
              @schema-errors="onSchemaErrors"
            >
              <template #title>
                <strong>{{ t('plugin.config.globalConfig') }}</strong>
              </template>
            </PluginConfigEditor>
          </section>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.hooks')" name="hooks">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            class="mb"
            :title="t('plugin.hooksNote')"
          />
          <el-table v-loading="hooksLoading" :data="hooksRows" border stripe>
            <el-table-column prop="name" :label="t('plugin.tabs.hooks')" min-width="200" />
            <el-table-column :label="t('plugin.hook.declared')" width="130">
              <template #default="{ row }">
                <el-tag :type="row.declared ? 'success' : 'info'">{{ row.declared ? t('plugin.hook.declaredYes') : t('plugin.hook.declaredNo') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.hook.platformHook')" width="130">
              <template #default="{ row }">
                <el-tag :type="row.platformHook ? 'success' : 'warning'">{{ row.platformHook ? t('plugin.hook.platformExists') : t('plugin.hook.platformUnknown') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.hook.handler')" width="150">
              <template #default="{ row }">
                <el-tag :type="row.execution_count > 0 ? 'success' : 'info'">
                  {{ row.execution_count > 0 ? t('plugin.hook.hasExecution') : t('plugin.hook.noExecution') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="mode" :label="t('plugin.hook.mode')" width="120">
              <template #default="{ row }">{{ row.mode === 'blocking' ? t('plugin.hook.blocking') : t('plugin.hook.nonBlocking') }}</template>
            </el-table-column>
            <el-table-column :label="t('plugin.hook.executionFailure')" width="130">
              <template #default="{ row }">{{ row.execution_count || 0 }} / {{ row.failure_count || 0 }}</template>
            </el-table-column>
            <el-table-column :label="t('plugin.hook.failureRate')" width="100">
              <template #default="{ row }">{{ failureRate(row) }}</template>
            </el-table-column>
            <el-table-column :label="t('plugin.hook.avgDuration')" width="110">
              <template #default="{ row }">{{ avgDuration(row) }}</template>
            </el-table-column>
            <el-table-column prop="last_executed_at" :label="t('plugin.hook.lastExecuted')" min-width="160" />
            <el-table-column prop="last_failed_at" :label="t('plugin.hook.lastFailed')" min-width="160" />
            <el-table-column prop="last_error" :label="t('plugin.hook.lastError')" min-width="220" />
            <el-table-column prop="failure_policy" :label="t('plugin.hook.failurePolicy')" width="140" />
            <el-table-column prop="description" :label="t('field.description')" min-width="240" />
          </el-table>
          <div class="hook-exec-head">
            <el-divider content-position="left">{{ t('plugin.hook.recentExecutions') }}</el-divider>
            <el-button size="small" @click="openHookExecutions()">{{ t('plugin.hook.viewAllExecutions') }}</el-button>
          </div>
          <el-table :data="hookRecent" border stripe :empty-text="`暂无${t('plugin.hook.recentExecutions')}`" data-testid="hook-recent-table">
            <el-table-column prop="finished_at" :label="t('plugin.audit.time')" width="170" />
            <el-table-column prop="hook_name" :label="t('plugin.tabs.hooks')" min-width="180" />
            <el-table-column prop="mode" :label="t('plugin.hook.mode')" width="120">
              <template #default="{ row }">{{ row.mode === 'blocking' ? t('plugin.hook.blocking') : t('plugin.hook.nonBlocking') }}</template>
            </el-table-column>
            <el-table-column prop="blocking" :label="t('plugin.hook.blockingFlag')" width="110">
              <template #default="{ row }">
                <el-tag :type="row.blocking ? 'danger' : 'info'" effect="plain">{{ row.blocking ? t('plugin.hook.blocking') : t('plugin.hook.nonBlocking') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.hook.result')" width="90">
              <template #default="{ row }">
                <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? pluginHealthLabel('success') : pluginHealthLabel('failed') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="content_type" :label="t('plugin.contentType')" width="140" />
            <el-table-column prop="content_id" :label="`${t('plugin.content.contentManagement')} ID`" width="120" />
            <el-table-column prop="community_id" :label="t('field.community_id')" width="130" />
            <el-table-column prop="duration_ms" :label="t('plugin.hook.durationMs')" width="100" />
            <el-table-column prop="error_message" :label="t('plugin.hook.error')" min-width="220" />
            <el-table-column :label="t('field.action')" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openHookExecutionDetail(row)">{{ t('common.detail') }}</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-drawer v-model="hookExecDrawer" :title="t('plugin.hook.executionsTitle')" size="920px" data-testid="hook-executions-drawer">
            <div class="hook-exec-filter">
              <el-form :inline="true">
                <el-form-item :label="t('plugin.hook.hookName')">
                  <el-select v-model="hookExecFilters.hook_name" clearable filterable style="width: 220px" data-testid="hook-exec-filter-hook">
                    <el-option v-for="name in allHookNames" :key="name" :label="name" :value="name" />
                  </el-select>
                </el-form-item>
                <el-form-item :label="t('plugin.hook.mode')">
                  <el-select v-model="hookExecFilters.mode" clearable style="width: 160px" data-testid="hook-exec-filter-mode">
                    <el-option label="blocking" value="blocking" />
                    <el-option label="non_blocking" value="non_blocking" />
                  </el-select>
                </el-form-item>
                <el-form-item :label="t('plugin.hook.success')">
                  <el-select v-model="hookExecFilters.success" style="width: 140px" data-testid="hook-exec-filter-success">
                    <el-option :label="t('common.all')" value="all" />
                    <el-option :label="pluginHealthLabel('success')" value="true" />
                    <el-option :label="pluginHealthLabel('failed')" value="false" />
                  </el-select>
                </el-form-item>
                <el-form-item :label="t('plugin.hook.blockingFlag')">
                  <el-select v-model="hookExecFilters.blocking" style="width: 140px" data-testid="hook-exec-filter-blocking">
                    <el-option :label="t('common.all')" value="all" />
                    <el-option :label="t('plugin.hook.blocking')" value="true" />
                    <el-option :label="t('plugin.hook.nonBlocking')" value="false" />
                  </el-select>
                </el-form-item>
                <el-form-item :label="t('plugin.contentType')">
                  <el-input v-model="hookExecFilters.content_type" clearable style="width: 140px" data-testid="hook-exec-filter-content-type" />
                </el-form-item>
                <el-form-item :label="t('plugin.hook.contentId')">
                  <el-input v-model="hookExecFilters.content_id" clearable style="width: 120px" data-testid="hook-exec-filter-content-id" />
                </el-form-item>
                <el-form-item :label="t('field.community_id')">
                  <el-input v-model="hookExecFilters.community_id" clearable style="width: 120px" data-testid="hook-exec-filter-community-id" />
                </el-form-item>
                <el-form-item :label="t('plugin.hook.actorType')">
                  <el-input v-model="hookExecFilters.actor_type" clearable style="width: 140px" data-testid="hook-exec-filter-actor-type" />
                </el-form-item>
                <el-form-item :label="t('plugin.hook.actorId')">
                  <el-input v-model="hookExecFilters.actor_id" clearable style="width: 120px" data-testid="hook-exec-filter-actor-id" />
                </el-form-item>
                <el-form-item :label="t('plugin.audit.timeRange')">
                  <el-date-picker v-model="hookExecFilters.range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" range-separator="-" start-placeholder="start" end-placeholder="end" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="loadHookExecutions(true)">{{ t('common.query') }}</el-button>
                  <el-button @click="resetHookExecutionsFilters">{{ t('common.reset') }}</el-button>
                </el-form-item>
              </el-form>
            </div>

            <el-table v-loading="hookExecLoading" :data="hookExecRows" border stripe data-testid="hook-executions-table">
              <el-table-column prop="id" label="ID" width="90" />
              <el-table-column prop="finished_at" :label="t('plugin.audit.time')" width="170" />
              <el-table-column prop="hook_name" :label="t('plugin.hook.hookName')" min-width="180" />
              <el-table-column prop="mode" :label="t('plugin.hook.mode')" width="140" />
              <el-table-column prop="blocking" :label="t('plugin.hook.blockingFlag')" width="110">
                <template #default="{ row }">
                  <el-tag :type="row.blocking ? 'danger' : 'info'" effect="plain">{{ row.blocking ? t('plugin.hook.blocking') : t('plugin.hook.nonBlocking') }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.hook.result')" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? pluginHealthLabel('success') : pluginHealthLabel('failed') }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="duration_ms" :label="t('plugin.hook.durationMs')" width="110" />
              <el-table-column prop="content_type" :label="t('plugin.contentType')" width="140" />
              <el-table-column prop="content_id" :label="t('plugin.hook.contentId')" width="120" />
              <el-table-column prop="community_id" :label="t('field.community_id')" width="130" />
              <el-table-column prop="actor_type" :label="t('plugin.hook.actorType')" width="140" />
              <el-table-column prop="actor_id" :label="t('plugin.hook.actorId')" width="120" />
              <el-table-column prop="error_message" :label="t('plugin.hook.error')" min-width="220" />
              <el-table-column :label="t('field.action')" width="120" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="openHookExecutionDetail(row)">{{ t('common.detail') }}</el-button>
                </template>
              </el-table-column>
            </el-table>

            <div class="pager">
              <el-pagination
                background
                layout="total, prev, pager, next, sizes"
                :total="hookExecTotal"
                :page-size="hookExecFilters.pageSize"
                :current-page="hookExecFilters.page"
                @update:current-page="(p) => { hookExecFilters.page = p; loadHookExecutions(false); }"
                @update:page-size="(s) => { hookExecFilters.pageSize = s; hookExecFilters.page = 1; loadHookExecutions(false); }"
              />
            </div>
          </el-drawer>

          <el-drawer v-model="hookDrawer" :title="t('plugin.hook.executionDetailTitle')" size="720px" data-testid="hook-execution-detail-drawer">
            <template v-if="hookExecTarget">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="ID">{{ hookExecTarget.id }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.hookName')">{{ hookExecTarget.hook_name }}</el-descriptions-item>
                <el-descriptions-item :label="t('field.plugin_code')">{{ hookExecTarget.plugin_code }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.mode')">{{ hookExecTarget.mode }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.blockingFlag')">
                  <el-tag :type="hookExecTarget.blocking ? 'danger' : 'info'" effect="plain">{{ hookExecTarget.blocking ? t('plugin.hook.blocking') : t('plugin.hook.nonBlocking') }}</el-tag>
                </el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.result')">
                  <el-tag :type="hookExecTarget.success ? 'success' : 'danger'">{{ hookExecTarget.success ? pluginHealthLabel('success') : pluginHealthLabel('failed') }}</el-tag>
                </el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.durationMs')">{{ hookExecTarget.duration_ms }}ms</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.audit.time')">{{ hookExecTarget.finished_at || hookExecTarget.started_at }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.contentType')">{{ hookExecTarget.content_type || '-' }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.contentId')">{{ hookExecTarget.content_id || 0 }}</el-descriptions-item>
                <el-descriptions-item :label="t('field.community_id')">{{ hookExecTarget.community_id || 0 }}</el-descriptions-item>
                <el-descriptions-item :label="t('field.category_id')">{{ hookExecTarget.category_id || 0 }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.actorType')">{{ hookExecTarget.actor_type || '-' }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.actorId')">{{ hookExecTarget.actor_id || 0 }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.audit.requestId')" :span="2">{{ hookExecTarget.request_id || '-' }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.error')" :span="2">
                  <pre class="mono pre">{{ hookExecTarget.error_message || '-' }}</pre>
                </el-descriptions-item>
                <el-descriptions-item :label="t('plugin.hook.metadata')" :span="2">
                  <pre class="mono pre">{{ formatJSON(jsonValue(hookExecTarget.metadata_json)) }}</pre>
                </el-descriptions-item>
              </el-descriptions>

              <div class="detail-actions">
                <el-button v-if="hookExecTarget.content_id" @click="openPluginContent(hookExecTarget)">{{ t('plugin.hook.openPluginContent') }}</el-button>
                <el-button @click="openHookAudit(hookExecTarget)">{{ t('plugin.hook.openAuditLogs') }}</el-button>
              </div>
            </template>
          </el-drawer>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.migrations')" name="migrations">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            class="mb"
            :title="t('plugin.migrationNote')"
          />
          <div class="sub-toolbar">
            <el-tag type="info" effect="plain">{{ t('common.total') }} {{ migrationSummary.total || migrationRows.length }}</el-tag>
            <el-tag type="success" effect="plain">{{ pluginHealthLabel('success') }} {{ migrationSummary.success || 0 }}</el-tag>
            <el-tag type="warning" effect="plain">{{ t('plugin.migration.pending') }} {{ migrationSummary.pending || 0 }}</el-tag>
            <el-tag type="danger" effect="plain">{{ pluginHealthLabel('failed') }} {{ migrationSummary.failed || 0 }}</el-tag>
            <el-button type="primary" size="small" @click="runMigrations">{{ t('plugin.migration.runPending') }}</el-button>
            <el-button size="small" @click="loadMigrations">{{ t('common.refresh') }}</el-button>
          </div>
          <el-table v-loading="migrationsLoading" :data="migrationRows" border stripe :empty-text="`暂无${t('plugin.tabs.migrations')}`">
            <el-table-column prop="migration_name" :label="t('plugin.migration.title')" min-width="180" />
            <el-table-column prop="migration_version" :label="t('plugin.version')" width="120" />
            <el-table-column prop="direction" :label="t('plugin.migration.direction')" width="100" />
            <el-table-column :label="t('field.status')" width="120">
              <template #default="{ row }">
                <el-tag :type="migrationStatusType(row.status)">{{ migrationStatusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="finished_at" :label="t('plugin.migration.lastFinished')" min-width="160" />
            <el-table-column :label="t('plugin.migration.duration')" width="110">
              <template #default="{ row }">{{ row.duration_ms || row.execution_time_ms || 0 }}ms</template>
            </el-table-column>
            <el-table-column :label="t('plugin.migration.rollback')" width="110">
              <template #default="{ row }">
                <el-tag :type="row.rollback_supported ? 'warning' : 'info'" effect="plain">
                  {{ row.rollback_supported ? t('common.support') : t('common.unsupported') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="error_message" :label="t('plugin.migration.errorReason')" min-width="220" />
            <el-table-column prop="description" :label="t('field.description')" min-width="240" />
            <el-table-column :label="t('field.action')" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" :disabled="row.status === 'success'" @click="retryMigration(row)">
                  {{ row.status === 'failed' ? t('common.retry') : t('common.run') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.routes')" name="routes">
          <el-table :data="plugin.routes || []" border stripe :empty-text="`暂无${t('plugin.tabs.routes')}`">
            <el-table-column prop="area" :label="t('field.area')" width="120" />
            <el-table-column prop="method" :label="t('field.method')" width="110" />
            <el-table-column prop="path" :label="t('field.path')" min-width="240" />
            <el-table-column prop="handler" :label="`${t('field.handler')} / 认证`" min-width="240" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.tabs.audit')" name="audit">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            :title="t('plugin.auditNote')"
            class="mb"
          />
          <div class="sub-toolbar">
            <span data-testid="plugin-audit-action-filter" class="audit-filter-wrap">
              <el-input v-model="auditQ.action" :placeholder="`${t('plugin.audit.actionText')}关键字（可选）`" clearable style="max-width: 260px" />
            </span>
            <span data-testid="plugin-audit-community-filter" class="audit-filter-wrap">
              <el-input-number v-model="auditQ.communityId" :min="0" :placeholder="`${t('field.community_id')}（可选）`" controls-position="right" style="width: 200px" />
            </span>
            <el-input v-model="auditQ.actor" :placeholder="`${t('plugin.audit.actor')}（可选）`" clearable style="max-width: 180px" />
            <el-input v-model="auditQ.targetType" :placeholder="`${t('plugin.audit.targetType')}（可选）`" clearable style="max-width: 180px" />
            <span data-testid="plugin-audit-target-filter" class="audit-filter-wrap">
              <el-input-number v-model="auditQ.targetId" :min="0" :placeholder="`${t('plugin.audit.targetId')}（可选）`" controls-position="right" style="width: 180px" />
            </span>
            <span data-testid="plugin-audit-metadata-filter" class="audit-filter-wrap">
              <el-input v-model="auditQ.metadata" :placeholder="`${t('plugin.audit.metadata')}关键字（可选）`" clearable style="max-width: 220px" />
            </span>
            <span data-testid="plugin-audit-request-filter" class="audit-filter-wrap">
              <el-input v-model="auditQ.requestId" :placeholder="`${t('plugin.audit.requestId')}（可选）`" clearable style="max-width: 200px" />
            </span>
            <el-date-picker
              v-model="auditQ.range"
              type="datetimerange"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 360px"
            />
            <el-button @click="loadAudit">{{ t('common.query') }}</el-button>
          </div>
          <el-table v-loading="auditLoading" :data="auditRows" border stripe :empty-text="`暂无${t('plugin.tabs.audit')}`">
            <el-table-column prop="id" label="ID" width="90" />
            <el-table-column prop="created_at" :label="t('plugin.audit.time')" width="170" />
            <el-table-column :label="t('plugin.audit.actor')" width="170">
              <template #default="{ row }">
                {{ row.actor || '-' }}
                <div class="muted">{{ row.actor_type || '-' }} / ID {{ row.actor_id || row.actor_user_id || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="action" :label="t('plugin.audit.actionText')" min-width="180">
              <template #default="{ row }">
                <div>{{ auditActionLabel(row.action) }}</div>
                <div class="muted mono">{{ row.action }}</div>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.audit.scope')" min-width="160">
              <template #default="{ row }">
                <div>{{ t('field.community_id') }} {{ row.community_id || '-' }}</div>
                <div class="muted">{{ t('plugin.audit.requestId') }} {{ metadataValue(row, 'request_id') || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.audit.targetType')" min-width="220">
              <template #default="{ row }">
                <div class="mono">{{ row.target || '-' }}</div>
                <div class="muted">{{ row.target_type || '-' }} / {{ row.target_id || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.audit.diff')" min-width="260">
              <template #default="{ row }">
                <details>
                  <summary class="muted">{{ t('common.view') }}{{ t('plugin.audit.diff') }} / {{ t('plugin.audit.metadata') }}</summary>
                  <pre class="json-box compact">{{ formatJSON(jsonValue(row.old_value)) }}</pre>
                  <pre class="json-box compact">{{ formatJSON(jsonValue(row.new_value)) }}</pre>
                  <pre class="json-box compact">{{ formatJSON(jsonValue(row.metadata_json)) }}</pre>
                </details>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="auditQ.page"
            v-model:page-size="auditQ.pageSize"
            class="pager"
            layout="total, sizes, prev, pager, next, jumper"
            :total="auditTotal"
            @change="loadAudit"
          />
        </el-tab-pane>
        </el-tabs>
      </div>
    </template>
  </el-drawer>

  <PluginConfigVersionsDialog
    v-model="configVersionsVisible"
    :plugin-code="plugin.code"
    scope="global"
    :community-id="0"
  />

  <el-dialog v-model="exportDialogVisible" title="导出本地插件包" width="820px" destroy-on-close data-testid="plugin-export-dialog">
    <template v-if="plugin">
      <el-alert
        type="info"
        show-icon
        :closable="false"
        class="mb"
        title="仅导出声明型插件包；不导出敏感配置、用户数据、运行时代码、外部 SQL，也不生成 zip 或远程发布。"
      />
      <div class="sub-toolbar">
        <el-checkbox v-model="exportForm.include_docs">include_docs</el-checkbox>
        <el-checkbox v-model="exportForm.include_migrations">include_migrations</el-checkbox>
        <el-checkbox v-model="exportForm.include_publisher">include_publisher</el-checkbox>
        <el-checkbox v-model="exportForm.include_signature_stub">include_signature_stub</el-checkbox>
      </div>
      <el-input v-model="exportForm.output_dir" class="mb" placeholder="可选 output_dir，例如 storage/plugins/exports/demo-1.0.0" data-testid="plugin-export-output-dir" />
      <div class="sub-toolbar">
        <el-button :loading="exportLoading" data-testid="plugin-export-dry-run" @click="dryRunExport">Dry-run</el-button>
        <el-button
          type="success"
          :loading="exportLoading"
          :disabled="!exportPreview || exportPreview.status === 'blocked'"
          data-testid="plugin-export-submit"
          @click="confirmExport"
        >
          正式导出
        </el-button>
      </div>
      <el-alert v-if="exportError" type="error" show-icon :closable="false" class="mb" :title="exportError" />
      <section v-if="exportPreview" class="export-result" data-testid="plugin-export-preview">
        <div class="sub-toolbar">
          <el-tag :type="exportPreview.status === 'blocked' ? 'danger' : exportPreview.status === 'warning' ? 'warning' : 'success'" effect="plain">
            {{ exportPreview.status }}
          </el-tag>
          <span class="mono">{{ exportPreview.export_preview?.output_dir || '-' }}</span>
        </div>
        <el-alert type="success" show-icon :closable="false" class="mb" title="安全检查：不包含敏感配置 / 不包含用户数据 / 不包含运行时代码 / 不包含外部 SQL" />
        <el-table :data="exportPreviewFiles" border stripe size="small" data-testid="plugin-export-files">
          <el-table-column prop="path" label="将导出文件" min-width="260" />
        </el-table>
        <el-alert v-if="(exportPreview.warnings || []).length" type="warning" show-icon :closable="false" class="mt" title="warnings">
          <ul class="result-list">
            <li v-for="(item, idx) in exportPreview.warnings" :key="`export-preview-warning-${idx}`">{{ item }}</li>
          </ul>
        </el-alert>
      </section>
      <section v-if="exportResult" class="export-result" data-testid="plugin-export-result">
        <el-alert type="success" show-icon :closable="false" class="mb" :title="exportResult.message || '导出成功'" />
        <el-descriptions :column="2" border>
          <el-descriptions-item label="output_dir">
            <span class="mono">{{ exportResult.output_dir }}</span>
            <el-button link type="primary" data-testid="plugin-export-copy-path" @click="copyText(exportResult.output_dir)">复制</el-button>
          </el-descriptions-item>
          <el-descriptions-item label="checksums">{{ exportResult.checksum_status || '-' }}</el-descriptions-item>
          <el-descriptions-item label="package_dry_run">{{ exportResult.package_dry_run_status || '-' }}</el-descriptions-item>
          <el-descriptions-item label="提示">可将该目录复制到本地插件仓库并执行 dry-run 验证。</el-descriptions-item>
        </el-descriptions>
      </section>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import PluginConfigEditor from './PluginConfigEditor.vue';
import PluginConfigVersionsDialog from './PluginConfigVersionsDialog.vue';
import { dryRunPluginExport, enablePluginFromEnablePrecheck, exportPluginPackage, getPluginUninstallImpact, getPluginUpgradeTask, listPluginPackageCompatChecks, listPluginUpgradeTasks, pluginAuditLogs, pluginHookExecutions, pluginHooks, pluginMenusPreview, pluginMigrations, pluginReadiness, pluginUpgradeImpact, retryPluginUpgradeTask, runPluginEnablePrecheck, runPluginMigrations, softUninstallPlugin, upgradePluginFromPackage, updatePluginConfig } from '@/api/admin';
import { t } from '@/i18n';
import { auditActionLabel, migrationStatusLabel, pluginHealthLabel, pluginStatusLabel } from '@/i18n/formatters';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  plugin: { type: Object, default: null },
  plugins: { type: Array, default: () => [] },
  initialTab: { type: String, default: 'overview' },
});
const emit = defineEmits(['update:modelValue', 'refresh', 'open-plugin']);
const router = useRouter();
const auth = useAuthStore();

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
});

watch(
  () => visible.value,
  (v) => {
    if (v && tab.value === 'audit') loadAudit();
  },
);

const tab = ref('overview');
const permQ = ref('');
const permOnlyMissing = ref(false);
const permOnlyHighRisk = ref(false);
const permissionRefsDialog = ref(false);
const selectedPermission = ref(null);
const selectedPermissionRefs = ref([]);
const schemaErrors = ref([]);
const configVersionsVisible = ref(false);
const configPanels = ref([]);
const editableConfig = ref({});
const auditLoading = ref(false);
const auditRows = ref([]);
const auditTotal = ref(0);
const auditQ = reactive({
  action: '',
  communityId: null,
  actor: '',
  targetType: '',
  targetId: null,
  metadata: '',
  requestId: '',
  range: [],
  page: 1,
  pageSize: 20,
});
const hooksLoading = ref(false);
const hookStats = ref([]);
const hookRecent = ref([]);
const hookExecLoading = ref(false);
const hookExecRows = ref([]);
const hookExecTotal = ref(0);
const hookExecDrawer = ref(false);
const hookExecTarget = ref(null);
const hookDrawer = ref(false);
const readinessLoading = ref(false);
const readinessResult = ref(null);
const menuPreviewLoading = ref(false);
const menuPreviewRows = ref([]);
const enablePrecheckLoading = ref(false);
const enablePrecheckResult = ref(null);
const enableTaskLoading = ref(false);
const softUninstallLoading = ref(false);
const softUninstallReason = ref('');
const softUninstallImpact = ref(null);
const softUninstallResult = ref(null);
const upgradeLoading = ref(false);
const upgradeImpact = ref(null);
const upgradeCompatChecks = ref([]);
const selectedUpgradeCompatCheckID = ref(0);
const upgradeResult = ref(null);
const upgradeTasksLoading = ref(false);
const upgradeTasks = ref([]);
const upgradeTasksTotal = ref(0);
const menuPreviewParams = reactive({
  community_slug: '',
  category_id: '',
});
const hookExecFilters = reactive({
  hook_name: '',
  mode: '',
  success: 'all',
  blocking: 'all',
  content_type: '',
  content_id: '',
  community_id: '',
  actor_type: '',
  actor_id: '',
  range: [],
  page: 1,
  pageSize: 20,
});
const migrationsLoading = ref(false);
const migrationRows = ref([]);
const migrationSummary = ref({});
const exportDialogVisible = ref(false);
const exportLoading = ref(false);
const exportError = ref('');
const exportPreview = ref(null);
const exportResult = ref(null);
const exportForm = reactive({
  include_docs: true,
  include_migrations: true,
  include_publisher: false,
  include_signature_stub: false,
  output_dir: '',
  force: false,
});

const title = computed(() => `${props.plugin?.name || ''} ${t('plugin.detailTitle')}`);
const exportPreviewFiles = computed(() => (exportPreview.value?.export_preview?.files || []).map((path) => ({ path })));
const canRunEnablePrecheck = computed(() => {
  const p = props.plugin;
  if (!p || !p.code) return false;
  if (String(p.status || '').trim() === 'enabled') return false;
  if (String(p.status || '').trim() === 'archived') return false;
  // Needs plugin.write permission; UI remains best-effort.
  return auth?.can ? auth.can('plugin.write') : true;
});

const canSoftUninstall = computed(() => {
  const p = props.plugin;
  if (!p || !p.code) return false;
  if (p.is_system) return false;
  if (String(p.source_type || '').trim() === 'builtin') return false;
  // Archived == already soft-uninstalled.
  if (String(p.status || '').trim() === 'archived') return false;
  return auth?.can ? auth.can('plugin.write') : true;
});

const canRunPackageUpgrade = computed(() => {
  const p = props.plugin;
  if (!p || !p.code) return false;
  if (p.is_system) return false;
  if (String(p.source_type || '').trim() === 'builtin') return false;
  // needs approver permission; UI remains best-effort.
  return auth?.can ? auth.can('plugin.approve') : true;
});

watch(
  () => props.plugin,
  (p) => {
    tab.value = props.initialTab || 'overview';
    permQ.value = '';
    schemaErrors.value = [];
    editableConfig.value = jsonValue(p?.config_json);
    // Reset audit query state for new plugin target.
    auditQ.action = '';
    auditQ.communityId = 0;
    auditQ.actor = '';
    auditQ.targetType = '';
    auditQ.targetId = 0;
    auditQ.metadata = '';
    auditQ.requestId = '';
    auditQ.range = [];
    auditQ.page = 1;
    auditQ.pageSize = 20;
    auditRows.value = [];
    auditTotal.value = 0;
    hookStats.value = [];
    hookRecent.value = [];
    migrationRows.value = [];
    migrationSummary.value = {};
    readinessResult.value = null;
    exportPreview.value = null;
    exportResult.value = null;
    exportError.value = '';
    upgradeImpact.value = null;
    upgradeCompatChecks.value = [];
    selectedUpgradeCompatCheckID.value = 0;
    upgradeResult.value = null;
    upgradeTasks.value = [];
    upgradeTasksTotal.value = 0;
    if (visible.value && tab.value === 'hooks') loadHooks();
    if (visible.value && tab.value === 'migrations') loadMigrations();
    if (visible.value && tab.value === 'readiness') {
      loadUpgradeCompatChecks();
      loadUpgradeTasks();
    }
  },
  { immediate: true },
);

watch(
  () => props.initialTab,
  (t) => {
    if (!visible.value) return;
    if (!t) return;
    tab.value = t;
    if (t === 'audit') loadAudit();
    if (t === 'hooks') loadHooks();
    if (t === 'migrations') loadMigrations();
    if (t === 'readiness') loadReadiness();
  },
);

watch(tab, (t) => {
  if (!visible.value) return;
  if (t === 'audit') loadAudit();
  if (t === 'hooks') loadHooks();
  if (t === 'migrations') loadMigrations();
  if (t === 'readiness') loadReadiness();
  if (t === 'menus') loadMenuPreview();
});

function readinessTagType(status) {
  if (status === 'pass') return 'success';
  if (status === 'warning') return 'warning';
  if (status === 'blocked') return 'danger';
  return 'info';
}

function readinessStatusLabel(status) {
  if (status === 'pass') return t('plugin.readiness.pass');
  if (status === 'warning') return t('plugin.readiness.warning');
  if (status === 'blocked') return t('plugin.readiness.blocked');
  return status || '-';
}

async function loadReadiness() {
  const p = props.plugin;
  if (!p || !p.code) return;
  readinessLoading.value = true;
  try {
    readinessResult.value = await pluginReadiness(p.code, { action: 'enable' });
  } catch (e) {
    readinessResult.value = null;
    ElMessage.warning(String(e?.message || e || t('plugin.readiness.unavailable')));
  } finally {
    readinessLoading.value = false;
  }
}

async function runEnablePrecheck() {
  const p = props.plugin;
  if (!p || !p.code) return;
  enablePrecheckLoading.value = true;
  try {
    enablePrecheckResult.value = await runPluginEnablePrecheck(p.code);
    ElMessage.success(t('plugin.ops.enablePrecheckDone'));
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('common.failed')));
  } finally {
    enablePrecheckLoading.value = false;
  }
}

async function enableFromPrecheck() {
  const id = Number(enablePrecheckResult.value?.id || 0);
  if (!id) return;
  enableTaskLoading.value = true;
  try {
    await ElMessageBox.confirm(t('plugin.ops.enableFromPrecheckConfirmText'), t('plugin.ops.enableFromPrecheckConfirmTitle'), {
      type: 'warning',
      confirmButtonText: t('plugin.ops.enableFromPrecheckConfirm'),
      cancelButtonText: t('common.cancel'),
    });
    await enablePluginFromEnablePrecheck(id);
    ElMessage.success(t('plugin.ops.enableFromPrecheckDone'));
    emit('refresh');
    await loadReadiness();
  } catch (e) {
    if (e === 'cancel') return;
    ElMessage.error(String(e?.message || e || t('common.failed')));
  } finally {
    enableTaskLoading.value = false;
  }
}

async function loadSoftUninstallImpact() {
  const p = props.plugin;
  if (!p || !p.code) return;
  try {
    softUninstallImpact.value = await getPluginUninstallImpact(p.code);
  } catch (e) {
    softUninstallImpact.value = null;
  }
}

async function runSoftUninstall() {
  const p = props.plugin;
  if (!p || !p.code) return;
  softUninstallLoading.value = true;
  try {
    await loadSoftUninstallImpact();
    const lines = [];
    if (softUninstallImpact.value?.impact) {
      const impact = softUninstallImpact.value.impact;
      lines.push(`历史内容数：${impact.existing_contents_count ?? '-'}`);
      lines.push(`已启用子站数：${impact.enabled_communities_count ?? '-'}`);
      lines.push(`待迁移数：${impact.pending_migrations_count ?? '-'}`);
    }
    lines.push('软卸载会将插件归档（archived），并从运行时入口中移除内容类型/菜单/路由/Hook 的可用性。');
    lines.push('软卸载不会删除历史内容、配置、迁移记录或审计日志；不会执行插件代码。');

    await ElMessageBox.confirm(lines.join('\n'), t('plugin.ops.softUninstallConfirmTitle'), {
      type: 'warning',
      confirmButtonText: t('plugin.ops.softUninstallConfirm'),
      cancelButtonText: t('common.cancel'),
    });
    softUninstallResult.value = await softUninstallPlugin(p.code, { version: p.version, reason: String(softUninstallReason.value || '').trim() });
    ElMessage.success(t('plugin.ops.softUninstallDone'));
    emit('refresh');
    await loadReadiness();
  } catch (e) {
    if (e === 'cancel') return;
    ElMessage.error(String(e?.message || e || t('common.failed')));
  } finally {
    softUninstallLoading.value = false;
  }
}

function upgradeCompatLabel(it) {
  if (!it) return '';
  const version = String(it.version || it.Version || '').trim();
  const status = String(it.status || '').trim();
  const canInstall = Boolean(it.can_install ?? it.canInstall);
  return `${version || '-'} / ${status || '-'} / can_install=${canInstall ? 'true' : 'false'} #${it.id}`;
}

async function loadUpgradeCompatChecks() {
  const p = props.plugin;
  if (!p || !p.code) return;
  try {
    const res = await listPluginPackageCompatChecks({ plugin_code: p.code, page: 1, page_size: 50 });
    const items = Array.isArray(res?.items) ? res.items : [];
    // Only show passed+can_install and version > current.
    upgradeCompatChecks.value = items
      .filter((it) => Boolean(it?.can_install))
      .filter((it) => String(it?.plugin_code || '').trim() === String(p.code || '').trim())
      .filter((it) => compareVersionStrings(String(it?.version || ''), String(p.version || '')) > 0)
      .sort((a, b) => -compareVersionStrings(String(a?.version || ''), String(b?.version || '')));
    if (!selectedUpgradeCompatCheckID.value && upgradeCompatChecks.value.length) {
      selectedUpgradeCompatCheckID.value = Number(upgradeCompatChecks.value[0].id || 0);
    }
  } catch (e) {
    upgradeCompatChecks.value = [];
  }
}

function compareVersionStrings(a, b) {
  // Minimal x.y.z compare (v prefix allowed). Keep frontend best-effort; backend is source of truth.
  const norm = (v) => String(v || '').trim().replace(/^v/i, '');
  const pa = norm(a).split('.').map((x) => parseInt(x, 10) || 0);
  const pb = norm(b).split('.').map((x) => parseInt(x, 10) || 0);
  for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
    const da = pa[i] ?? 0;
    const db = pb[i] ?? 0;
    if (da !== db) return da > db ? 1 : -1;
  }
  return 0;
}

async function loadUpgradeImpact() {
  const p = props.plugin;
  if (!p || !p.code) return;
  const id = Number(selectedUpgradeCompatCheckID.value || 0);
  if (!id) return;
  upgradeLoading.value = true;
  try {
    upgradeImpact.value = await pluginUpgradeImpact(p.code, { target_compat_check_id: id });
    ElMessage.success(t('plugin.ops.packageUpgradeImpactDone'));
  } catch (e) {
    upgradeImpact.value = null;
    ElMessage.error(String(e?.message || e || t('common.failed')));
  } finally {
    upgradeLoading.value = false;
  }
}

async function runPackageUpgrade() {
  const p = props.plugin;
  if (!p || !p.code) return;
  const id = Number(selectedUpgradeCompatCheckID.value || 0);
  if (!id) return;
  upgradeLoading.value = true;
  try {
    await loadUpgradeImpact();
    const canUpgrade = Boolean(upgradeImpact.value?.can_upgrade);
    const lines = [];
    lines.push(`plugin_code=${p.code}`);
    if (upgradeImpact.value?.old_version) lines.push(`old_version=${upgradeImpact.value.old_version}`);
    if (upgradeImpact.value?.new_version) lines.push(`new_version=${upgradeImpact.value.new_version}`);
    lines.push(t('plugin.ops.packageUpgradeConfirmTip'));
    if (!canUpgrade) lines.push(t('plugin.ops.packageUpgradeBlockedTip'));
    await ElMessageBox.confirm(lines.join('\n'), t('plugin.ops.packageUpgradeConfirmTitle'), {
      type: 'warning',
      confirmButtonText: t('plugin.ops.packageUpgradeConfirm'),
      cancelButtonText: t('common.cancel'),
    });
    upgradeResult.value = await upgradePluginFromPackage(p.code, { target_compat_check_id: id, reason: t('plugin.ops.packageUpgradeDefaultReason') });
    ElMessage.success(t('plugin.ops.packageUpgradeDone'));
    emit('refresh');
    await loadReadiness();
    await loadUpgradeTasks();
  } catch (e) {
    if (e === 'cancel') return;
    ElMessage.error(String(e?.message || e || t('common.failed')));
  } finally {
    upgradeLoading.value = false;
  }
}

async function loadUpgradeTasks() {
  const p = props.plugin;
  if (!p || !p.code) return;
  upgradeTasksLoading.value = true;
  try {
    const res = await listPluginUpgradeTasks({ plugin_code: p.code, page: 1, page_size: 20 });
    upgradeTasks.value = Array.isArray(res?.items) ? res.items : [];
    upgradeTasksTotal.value = Number(res?.pagination?.total || 0);
  } catch (e) {
    upgradeTasks.value = [];
    upgradeTasksTotal.value = 0;
  } finally {
    upgradeTasksLoading.value = false;
  }
}

async function openUpgradeTask(id) {
  const taskID = Number(id || 0);
  if (!taskID) return;
  try {
    const res = await getPluginUpgradeTask(taskID);
    const lines = [];
    lines.push(`id=${res?.id}`);
    lines.push(`status=${res?.status}`);
    lines.push(`old_version=${res?.old_version}`);
    lines.push(`new_version=${res?.new_version}`);
    if ((res?.errors || []).length) lines.push(`errors=${(res.errors || []).join('；')}`);
    if ((res?.warnings || []).length) lines.push(`warnings=${(res.warnings || []).join('；')}`);
    await ElMessageBox.alert(lines.join('\n'), t('plugin.ops.packageUpgradeTaskDetailTitle'), { type: 'info', confirmButtonText: t('common.close') });
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('common.failed')));
  }
}

async function retryUpgradeTask(id) {
  const taskID = Number(id || 0);
  if (!taskID) return;
  try {
    await ElMessageBox.confirm(t('plugin.ops.packageUpgradeRetryConfirmTip'), t('plugin.ops.packageUpgradeRetryConfirmTitle'), {
      type: 'warning',
      confirmButtonText: t('common.retry'),
      cancelButtonText: t('common.cancel'),
    });
    const res = await retryPluginUpgradeTask(taskID);
    upgradeResult.value = res;
    ElMessage.success(t('plugin.ops.packageUpgradeRetryDone'));
    emit('refresh');
    await loadReadiness();
    await loadUpgradeTasks();
  } catch (e) {
    if (e === 'cancel') return;
    ElMessage.error(String(e?.message || e || t('common.failed')));
  }
}

async function loadMenuPreview() {
  const p = props.plugin;
  if (!p || !p.code) return;
  menuPreviewLoading.value = true;
  try {
    const params = {};
    if (String(menuPreviewParams.community_slug || '').trim()) params.community_slug = String(menuPreviewParams.community_slug || '').trim();
    if (String(menuPreviewParams.category_id || '').trim()) params.category_id = String(menuPreviewParams.category_id || '').trim();
    const data = await pluginMenusPreview(p.code, params);
    menuPreviewRows.value = Array.isArray(data.items) ? data.items : [];
  } catch (e) {
    menuPreviewRows.value = [];
  } finally {
    menuPreviewLoading.value = false;
  }
}

function permissionOpMeta(code) {
  const c = String(code || '');
  const ops = [
    { suffix: '.read', label: t('op.read'), tag: 'info', highRisk: false },
    { suffix: '.create', label: t('op.create'), tag: 'success', highRisk: false },
    { suffix: '.edit', label: t('op.edit'), tag: 'warning', highRisk: false },
    { suffix: '.update', label: t('op.edit'), tag: 'warning', highRisk: false },
    { suffix: '.delete', label: t('op.delete'), tag: 'danger', highRisk: true },
    { suffix: '.audit', label: t('op.audit'), tag: 'warning', highRisk: true },
    { suffix: '.manage', label: t('op.manage'), tag: 'danger', highRisk: true },
    { suffix: '.configure', label: t('op.configure'), tag: 'warning', highRisk: true },
    { suffix: '.write', label: t('op.manage'), tag: 'danger', highRisk: true },
  ];
  for (const item of ops) {
    if (c.endsWith(item.suffix)) return item;
  }
  return { label: t('op.unknown'), tag: 'info', highRisk: false };
}

function buildPermissionRefs(plugin) {
  const refs = new Map();
  const push = (permission, item) => {
    if (!permission) return;
    const key = String(permission);
    if (!refs.has(key)) refs.set(key, []);
    refs.get(key).push(item);
  };
  (plugin?.menus || []).forEach((m) => {
    push(m.permission, { type: 'menu', title: m.title || '-', path: m.path || '-' });
  });
  (plugin?.routes || []).forEach((r) => {
    push(r.permission, { type: 'route', title: `${r.method || ''} ${r.path || '-'}`.trim(), path: r.path || '-' });
  });
  (plugin?.content_type_definitions || []).forEach((ct) => {
    push(ct.create_permission, { type: 'content_type', title: `${ct.type} ${t('op.create')}`, path: '-' });
    push(ct.edit_permission, { type: 'content_type', title: `${ct.type} ${t('op.edit')}`, path: '-' });
    push(ct.delete_permission, { type: 'content_type', title: `${ct.type} ${t('op.delete')}`, path: '-' });
    push(ct.audit_permission, { type: 'content_type', title: `${ct.type} ${t('op.audit')}`, path: '-' });
  });
  return refs;
}

function permissionRowClass({ row }) {
  if (row?._highRisk && !row?._hasPermission) return 'row-warn';
  if (!row?._hasPermission) return 'row-danger';
  if (row?._highRisk) return 'row-warn';
  return '';
}

function openPermissionRefs(row) {
  const refs = buildPermissionRefs(props.plugin);
  selectedPermission.value = row;
  selectedPermissionRefs.value = refs.get(row.code) || [];
  permissionRefsDialog.value = true;
}

const filteredPermissions = computed(() => {
  const q = (permQ.value || '').trim().toLowerCase();
  const list = (props.plugin?.permissions || []).map((p) => {
    const meta = permissionOpMeta(p.code);
    return {
      ...p,
      _hasPermission: auth.can(p.code),
      _opTypeLabel: meta.label,
      _opTypeTag: meta.tag,
      _highRisk: meta.highRisk,
    };
  });
  return list
    .filter((p) => (!q ? true : (p.code || '').toLowerCase().includes(q) || (p.name || '').toLowerCase().includes(q)))
    .filter((p) => (permOnlyMissing.value ? !p._hasPermission : true))
    .filter((p) => (permOnlyHighRisk.value ? p._highRisk : true));
});

const allHookNames = [
  'BeforeCreateContent',
  'AfterCreateContent',
  'BeforeUpdateContent',
  'AfterUpdateContent',
  'BeforeModerateContent',
  'AfterModerateContent',
  'BeforeBuildSEO',
  'AfterBuildSEO',
  'AfterPluginEnabled',
  'AfterPluginDisabled',
  'AfterCreateComment',
  'OnSearchIndex',
  'OnNotificationBuild',
  'OnSEOBuild',
];

// 平台“声明”了这些 Hook，但并不代表当前代码已经在所有流程中完整触发。
// 这里仅用于避免 UI 伪造“平台调用点存在”的结论。
// 是否触发，以后端真实 Dispatch 接入点为准（详见 docs/PLUGIN_ARCHITECTURE.md / docs/TESTING.md）。
const platformDispatchedHooks = new Set([
  // 已确认的接入点（根据当前后端 service/router 的 DispatchHook 调用）
  'BeforeCreateContent',
  'AfterCreateContent',
  'BeforeUpdateContent',
  'AfterUpdateContent',
  'BeforeModerateContent',
  'AfterModerateContent',
  'AfterCreateComment',
  'OnSearchIndex',
  'OnNotificationBuild',
  'OnSEOBuild',
  'AfterPluginEnabled',
  'AfterPluginDisabled',
]);

const hooksRows = computed(() => {
  const declared = new Map((props.plugin?.hooks || []).map((h) => [h.name, h]));
  const stats = new Map((hookStats.value || []).map((h) => [h.hook_name, h]));
  return allHookNames.map((name) => {
    const hook = declared.get(name);
    const stat = stats.get(name) || {};
    return {
      name,
      declared: Boolean(hook),
      platformHook: platformDispatchedHooks.has(name),
      failure_policy: hook?.failure_policy || '-',
      description: hook?.description || '-',
      mode: stat.mode || (hook?.critical ? 'blocking' : 'non_blocking'),
      execution_count: stat.execution_count || 0,
      failure_count: stat.failure_count || 0,
      avg_duration_ms: stat.avg_duration_ms || 0,
      last_executed_at: stat.last_executed_at || '-',
      last_failed_at: stat.last_failed_at || '-',
      last_error: stat.last_error || '-',
    };
  });
});

const dependencyRows = computed(() => {
  if (Array.isArray(props.plugin?.dependency_checks) && props.plugin.dependency_checks.length) {
    return props.plugin.dependency_checks.map((row) => ({
      code: row.code,
      version: row.version || '-',
      required: row.required !== false,
      reason: row.reason || '-',
      pluginName: row.plugin_name || '-',
      currentVersion: row.current_version || '-',
      currentStatus: row.current_status || '',
      status: row.status || '',
      satisfied: Boolean(row.satisfied),
      message: row.message || '-',
    }));
  }
  const declared = Array.isArray(props.plugin?.dependencies) ? props.plugin.dependencies : [];
  const byCode = new Map((props.plugins || []).map((item) => [item.code, item]));
  return declared.map((dep) => {
    const plugin = byCode.get(dep.code) || {};
    const status = dependencyStatus(dep, plugin);
    return {
      code: dep.code,
      version: dep.version || '-',
      required: dep.required !== false,
      reason: dep.reason || '-',
      pluginName: plugin.name || '-',
      currentVersion: plugin.version || '-',
      currentStatus: plugin.status || '',
      status,
      satisfied: status === 'satisfied',
      message: dependencyMessage(status, plugin),
    };
  });
});

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
  if (status === 'error' || status === 'migration_failed' || status === 'config_invalid' || status === 'dependency_missing') return 'danger';
  return 'info';
}

function metricType(status) {
  if (status === 'ok' || status === 'valid') return 'success';
  if (status === 'warning' || status === 'pending' || status === 'hook_warning') return 'warning';
  if (status === 'failed' || status === 'invalid' || status === 'missing' || status === 'hook_error') return 'danger';
  return 'info';
}

function dependencyStatus(dep, plugin) {
  if (!plugin?.code) return dep.required === false ? 'optional_missing' : 'missing';
  if (plugin.status === 'archived') return 'archived';
  if (plugin.status === 'migration_failed') return 'migration_failed';
  if (plugin.status === 'config_invalid') return 'config_invalid';
  if (plugin.status !== 'enabled') return 'disabled';
  return 'satisfied';
}

function dependencyStatusType(status, satisfied) {
  if (satisfied || status === 'satisfied') return 'success';
  if (status === 'optional_missing') return 'warning';
  if (['missing', 'disabled', 'archived', 'migration_failed', 'config_invalid', 'version_mismatch', 'circular_dependency', 'self_dependency'].includes(status)) return 'danger';
  return 'info';
}

function dependencyStatusLabel(status) {
  return t(`plugin.dependencies.status.${status || 'unknown'}`);
}

function dependencyMessage(status, plugin) {
  if (status === 'satisfied') return t('plugin.dependencies.satisfiedTip');
  if (status === 'disabled') return plugin?.status ? pluginStatusLabel(plugin.status) : t('plugin.dependencies.status.disabled');
  return dependencyStatusLabel(status);
}

function migrationStatusType(status) {
  if (status === 'success') return 'success';
  if (status === 'failed') return 'danger';
  if (status === 'running' || status === 'pending') return 'warning';
  return 'info';
}

function jsonValue(v) {
  if (!v) return {};
  if (typeof v === 'string') {
    try {
      return JSON.parse(v);
    } catch {
      return {};
    }
  }
  if (typeof v === 'object') return v;
  return {};
}

async function loadAudit() {
  const p = props.plugin;
  if (!p || !p.code) return;
  auditLoading.value = true;
  try {
    const data = await pluginAuditLogs(p.code, {
      type: 'all',
      action: auditQ.action || '',
      community_id: auditQ.communityId || 0,
      actor: auditQ.actor || '',
      target_type: auditQ.targetType || '',
      target_id: auditQ.targetId || 0,
      metadata: auditQ.metadata || '',
      request_id: auditQ.requestId || '',
      start_time: Array.isArray(auditQ.range) ? auditQ.range[0] || '' : '',
      end_time: Array.isArray(auditQ.range) ? auditQ.range[1] || '' : '',
      page: auditQ.page,
      page_size: auditQ.pageSize,
    });
    auditRows.value = data.items || [];
    auditTotal.value = data.total || 0;
  } finally {
    auditLoading.value = false;
  }
}

async function loadHooks() {
  const p = props.plugin;
  if (!p || !p.code) return;
  hooksLoading.value = true;
  try {
    const data = await pluginHooks(p.code);
    hookStats.value = data.items || [];
    hookRecent.value = data.recent_executions || [];
  } catch (e) {
    hookStats.value = [];
    hookRecent.value = [];
    ElMessage.warning(String(e?.message || e || t('plugin.hook.unavailable')));
  } finally {
    hooksLoading.value = false;
  }
}

function openHookExecutions() {
  hookExecDrawer.value = true;
  // Load first page with current filters.
  loadHookExecutions(true);
}

function resetHookExecutionsFilters() {
  hookExecFilters.hook_name = '';
  hookExecFilters.mode = '';
  hookExecFilters.success = 'all';
  hookExecFilters.blocking = 'all';
  hookExecFilters.content_type = '';
  hookExecFilters.content_id = '';
  hookExecFilters.community_id = '';
  hookExecFilters.actor_type = '';
  hookExecFilters.actor_id = '';
  hookExecFilters.range = [];
  hookExecFilters.page = 1;
  hookExecFilters.pageSize = 20;
  loadHookExecutions(true);
}

async function loadHookExecutions(resetPage) {
  const p = props.plugin;
  if (!p || !p.code) return;
  if (resetPage) hookExecFilters.page = 1;
  hookExecLoading.value = true;
  try {
    const params = {
      hook_name: hookExecFilters.hook_name || '',
      mode: hookExecFilters.mode || '',
      content_type: hookExecFilters.content_type || '',
      content_id: Number(hookExecFilters.content_id || 0) || 0,
      community_id: Number(hookExecFilters.community_id || 0) || 0,
      actor_type: hookExecFilters.actor_type || '',
      actor_id: Number(hookExecFilters.actor_id || 0) || 0,
      page: hookExecFilters.page,
      page_size: hookExecFilters.pageSize,
      start_time: Array.isArray(hookExecFilters.range) ? hookExecFilters.range[0] || '' : '',
      end_time: Array.isArray(hookExecFilters.range) ? hookExecFilters.range[1] || '' : '',
    };
    if (hookExecFilters.success !== 'all') params.success = hookExecFilters.success;
    if (hookExecFilters.blocking !== 'all') params.blocking = hookExecFilters.blocking;
    const data = await pluginHookExecutions(p.code, params);
    hookExecRows.value = data.items || [];
    hookExecTotal.value = data.total || 0;
  } catch (e) {
    hookExecRows.value = [];
    hookExecTotal.value = 0;
    ElMessage.warning(String(e?.message || e || t('plugin.hook.unavailable')));
  } finally {
    hookExecLoading.value = false;
  }
}

function openHookExecutionDetail(row) {
  hookExecTarget.value = row;
  hookDrawer.value = true;
}

function openPluginContent(row) {
  const pluginCode = row?.plugin_code || props.plugin?.code;
  if (!pluginCode) return;
  router.push({ path: `/admin-next/${pluginCode}` });
}

function openHookAudit(row) {
  const pluginCode = row?.plugin_code || props.plugin?.code;
  const hookName = row?.hook_name || '';
  const requestId = row?.request_id || '';
  // Reuse existing audit log page; do not fake exact match.
  router.push({
    path: '/admin-next/audit-logs',
    query: {
      plugin_code: pluginCode,
      action: row?.blocking ? 'plugin.hook.blocked' : 'plugin.hook.failed',
      request_id: requestId,
      metadata: hookName ? `hook_name=${hookName}` : '',
      target_type: 'hooks',
      target: pluginCode ? `hooks#${pluginCode}:${hookName}` : '',
    },
  });
}

async function loadMigrations() {
  const p = props.plugin;
  if (!p || !p.code) return;
  migrationsLoading.value = true;
  try {
    const data = await pluginMigrations(p.code);
    migrationRows.value = data.items || [];
    migrationSummary.value = data.summary || {};
  } catch (e) {
    migrationRows.value = [];
    migrationSummary.value = {};
    ElMessage.warning(String(e?.message || e || t('plugin.migration.unavailable')));
  } finally {
    migrationsLoading.value = false;
  }
}

async function runMigrations() {
  const p = props.plugin;
  if (!p || !p.code) return;
  migrationsLoading.value = true;
  try {
    await runPluginMigrations(p.code);
    ElMessage.success(t('plugin.migration.executeDone'));
    await loadMigrations();
    emit('refresh');
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('plugin.migration.executeFailed')));
  } finally {
    migrationsLoading.value = false;
  }
}

async function retryMigration(row) {
  const p = props.plugin;
  if (!p || !p.code || !row?.migration_name) return;
  migrationsLoading.value = true;
  try {
    await retryPluginMigration(p.code, row.migration_name);
    ElMessage.success(row.status === 'failed' ? t('plugin.migration.retryDone') : t('plugin.migration.executeDone'));
    await loadMigrations();
    emit('refresh');
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('plugin.migration.executeFailed')));
  } finally {
    migrationsLoading.value = false;
  }
}

function failureRate(row) {
  const total = Number(row.execution_count || 0);
  if (!total) return '-';
  return `${Math.round((Number(row.failure_count || 0) / total) * 100)}%`;
}

function avgDuration(row) {
  const value = Number(row.avg_duration_ms || 0);
  if (!Number.isFinite(value)) return '-';
  return `${value.toFixed(value >= 10 ? 0 : 1)}ms`;
}

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
}

function metadataValue(row, key) {
  const meta = jsonValue(row?.metadata_json);
  return meta?.[key] || '';
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(String(text || ''));
    ElMessage.success(t('common.copied'));
  } catch {
    ElMessage.warning(t('common.copyUnsupported'));
  }
}

function onSchemaErrors(errs) {
  schemaErrors.value = Array.isArray(errs) ? errs : [];
}

function reloadConfig() {
  editableConfig.value = jsonValue(props.plugin?.config_json);
  ElMessage.success(t('common.resetDone'));
}

function clearGlobalConfig() {
  editableConfig.value = {};
  ElMessage.success(t('common.clearDone'));
}

async function saveConfig() {
  const p = props.plugin;
  if (!p) return;
  try {
    await updatePluginConfig(p.code, { config_json: editableConfig.value || {} });
    ElMessage.success(t('plugin.config.globalSaved'));
    emit('refresh');
  } catch (e) {
    ElMessage.error(String(e?.message || e || t('common.saveFailed')));
  }
}

function exportPayload() {
  const payload = {
    include_docs: Boolean(exportForm.include_docs),
    include_migrations: Boolean(exportForm.include_migrations),
    include_publisher: Boolean(exportForm.include_publisher),
    include_signature_stub: Boolean(exportForm.include_signature_stub),
    force: Boolean(exportForm.force),
  };
  const out = String(exportForm.output_dir || '').trim();
  if (out) payload.output_dir = out;
  return payload;
}

function openExportDialog() {
  exportDialogVisible.value = true;
  exportError.value = '';
  exportPreview.value = null;
  exportResult.value = null;
  exportForm.include_docs = true;
  exportForm.include_migrations = true;
  exportForm.include_publisher = false;
  exportForm.include_signature_stub = false;
  exportForm.force = false;
  exportForm.output_dir = '';
}

function formatAPIError(e, fallback) {
  const data = e?.response?.data;
  const code = String(data?.code || '').trim();
  const message = String(data?.message || data?.error || e?.message || fallback || '').trim();
  const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
  return [code ? `[${code}]` : '', message, suggestion ? `建议：${suggestion}` : ''].filter(Boolean).join(' ');
}

async function dryRunExport() {
  const p = props.plugin;
  if (!p?.code) return;
  exportLoading.value = true;
  exportError.value = '';
  exportResult.value = null;
  try {
    exportPreview.value = await dryRunPluginExport(p.code, exportPayload());
  } catch (e) {
    exportPreview.value = null;
    exportError.value = formatAPIError(e, '导出 dry-run 失败');
  } finally {
    exportLoading.value = false;
  }
}

async function confirmExport() {
  const p = props.plugin;
  if (!p?.code) return;
  try {
    await ElMessageBox.confirm('确认导出声明型插件包？导出不会包含敏感配置、用户数据、运行时代码或外部 SQL。', '确认导出', {
      confirmButtonText: '确认导出',
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    });
  } catch {
    return;
  }
  exportLoading.value = true;
  exportError.value = '';
  try {
    exportResult.value = await exportPluginPackage(p.code, exportPayload());
    ElMessage.success('导出成功');
  } catch (e) {
    exportResult.value = null;
    exportError.value = formatAPIError(e, '导出失败');
  } finally {
    exportLoading.value = false;
  }
}
</script>

<style scoped>
.hero {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: linear-gradient(135deg, #f8fafc, #eef6ff);
}
.drawer-content {
  min-height: calc(100vh - 96px);
  padding-bottom: 24px;
}
.hero-title {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.hero-title h3 {
  margin: 0;
  font-size: 20px;
  color: #0f172a;
}
.hero-desc {
  margin: 8px 0 0;
  color: #64748b;
}
.hero-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}
.code-pill {
  align-self: flex-start;
  padding: 6px 10px;
  border-radius: 999px;
  background: #0f172a;
  color: #fff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.meta-line {
  margin-top: 10px;
  color: #64748b;
  font-size: 12px;
}
.tabs {
  margin-top: 16px;
}
.tabs :deep(.el-tabs__content) {
  min-height: 420px;
  padding-top: 4px;
}
.tabs :deep(.el-tab-pane) {
  min-height: 360px;
}
.sub-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-start;
  align-items: center;
  margin-bottom: 10px;
}
.audit-filter-wrap {
  display: inline-flex;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: #0f172a;
}
.mb {
  margin-bottom: 12px;
}
.mt {
  margin-top: 12px;
}
.banner-lines {
  display: grid;
  gap: 6px;
  line-height: 1.5;
}
.json-box {
  margin: 0;
  padding: 14px;
  border-radius: 12px;
  background: #0f172a;
  color: #dbeafe;
  max-height: 360px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
}
.config-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 10px;
}
.config-card {
  margin-top: 12px;
  padding: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background: #fff;
}
.config-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}
.config-card-header h4 {
  margin: 0;
  color: #0f172a;
}
.config-card-header p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 13px;
}
.config-card-tools {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  min-width: 320px;
}
.export-panel {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
  padding: 14px;
  border: 1px solid #dbeafe;
  border-radius: 12px;
  background: #f8fbff;
}
.export-panel h4 {
  margin: 0;
  color: #0f172a;
}
.export-panel p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 13px;
}
.export-result {
  margin-top: 12px;
}
.result-list {
  margin: 6px 0 0;
  padding-left: 18px;
}
:global(.plugin-detail-drawer .el-drawer__body) {
  padding-top: 10px;
  overflow: auto;
}
.hook-exec-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}
.hook-exec-filter {
  margin-bottom: 10px;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}
.pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 220px;
  overflow: auto;
}
.detail-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

:global(.plugin-detail-drawer .el-table .row-danger td) {
  background: rgba(245, 108, 108, 0.08);
}
:global(.plugin-detail-drawer .el-table .row-warn td) {
  background: rgba(230, 162, 60, 0.08);
}
</style>
