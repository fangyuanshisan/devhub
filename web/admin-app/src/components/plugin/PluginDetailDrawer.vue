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
          </div>
          <el-table :data="filteredPermissions" border stripe :empty-text="`暂无${t('plugin.tabs.permissions')}`">
            <el-table-column prop="code" :label="t('field.code')" min-width="240">
              <template #default="{ row }">
                <div class="mono">{{ row.code }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="name" :label="t('field.name')" min-width="160" />
            <el-table-column prop="scope" :label="t('field.scope')" width="150" />
            <el-table-column prop="description" :label="t('field.description')" min-width="220" />
            <el-table-column :label="t('plugin.action')" width="90">
              <template #default="{ row }">
                <el-button link type="primary" @click="copyText(row.code)">{{ t('common.copy') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
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
                <el-button size="small" @click="reloadConfig">{{ t('plugin.config.resetCurrent') }}</el-button>
                <el-button size="small" data-testid="plugin-global-config-clear" @click="clearGlobalConfig">{{ t('common.clearObject') }}</el-button>
                <el-button size="small" type="primary" data-testid="plugin-global-config-save" :disabled="schemaErrors.length > 0" @click="saveConfig">{{ t('common.save') }}</el-button>
              </div>
            </div>
            <PluginJsonEditor
              v-model="editableConfig"
              :schema="plugin.config_schema || null"
              :original-value="jsonValue(plugin.config_json)"
              :resolved-config="plugin.resolved_config?.effective || plugin.resolved_config || {}"
              @schema-errors="onSchemaErrors"
            >
              <template #title>
                <strong>{{ t('plugin.config.globalConfig') }}</strong>
              </template>
            </PluginJsonEditor>
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
          <el-divider>{{ t('plugin.hook.recentExecutions') }}</el-divider>
          <el-table :data="hookRecent" border stripe :empty-text="`暂无${t('plugin.hook.recentExecutions')}`">
            <el-table-column prop="finished_at" :label="t('plugin.audit.time')" width="170" />
            <el-table-column prop="hook_name" :label="t('plugin.tabs.hooks')" min-width="180" />
            <el-table-column prop="mode" :label="t('plugin.hook.mode')" width="120">
              <template #default="{ row }">{{ row.mode === 'blocking' ? t('plugin.hook.blocking') : t('plugin.hook.nonBlocking') }}</template>
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
          </el-table>
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
            <el-table-column prop="handler" :label="`${t('field.handler')} / Auth`" min-width="240" />
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
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import PluginJsonEditor from './PluginJsonEditor.vue';
import { pluginAuditLogs, pluginHooks, pluginMigrations, retryPluginMigration, runPluginMigrations, updatePluginConfig } from '@/api/admin';
import { t } from '@/i18n';
import { auditActionLabel, migrationStatusLabel, pluginHealthLabel, pluginStatusLabel } from '@/i18n/formatters';

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  plugin: { type: Object, default: null },
  initialTab: { type: String, default: 'overview' },
});
const emit = defineEmits(['update:modelValue', 'refresh']);

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
const schemaErrors = ref([]);
const configPanels = ref([]);
const editableConfig = ref({});
const hooksLoading = ref(false);
const hookStats = ref([]);
const hookRecent = ref([]);
const migrationsLoading = ref(false);
const migrationRows = ref([]);
const migrationSummary = ref({});

const title = computed(() => `${props.plugin?.name || ''} ${t('plugin.detailTitle')}`);

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
    if (visible.value && tab.value === 'hooks') loadHooks();
    if (visible.value && tab.value === 'migrations') loadMigrations();
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
  },
);

watch(tab, (t) => {
  if (!visible.value) return;
  if (t === 'audit') loadAudit();
  if (t === 'hooks') loadHooks();
  if (t === 'migrations') loadMigrations();
});

const filteredPermissions = computed(() => {
  const q = (permQ.value || '').trim().toLowerCase();
  const list = props.plugin?.permissions || [];
  if (!q) return list;
  return list.filter((p) => (p.code || '').toLowerCase().includes(q) || (p.name || '').toLowerCase().includes(q));
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
:global(.plugin-detail-drawer .el-drawer__body) {
  padding-top: 10px;
  overflow: auto;
}
</style>
