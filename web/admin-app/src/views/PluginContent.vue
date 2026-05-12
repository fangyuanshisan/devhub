<template>
  <section class="page-card plugin-content-page" data-testid="plugin-content-page">
    <div class="plugin-content-header" data-testid="plugin-content-header">
      <div class="header-main">
        <p class="eyebrow">{{ t('plugin.content.contentManagement') }}</p>
        <h2>{{ plugin?.name || route.meta.title }}</h2>
        <div class="meta-line">
          <span>{{ t('plugin.code') }}：<span class="mono">{{ route.meta.pluginCode }}</span></span>
          <span>{{ t('plugin.status') }}：<el-tag v-if="plugin" :type="statusTagType(plugin.status)" size="small">{{ pluginStatusLabel(plugin.status) }}</el-tag></span>
          <span data-testid="plugin-content-health">{{ t('plugin.content.health') }}：<el-tag :type="statusTagType(healthStatus)" size="small">{{ pluginHealthLabel(healthStatus) }}</el-tag></span>
          <span data-testid="plugin-content-type-count">{{ t('plugin.content.contentTypeCount') }}：{{ contentTypeCount }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button data-testid="plugin-content-back" @click="backToPlugins">{{ t('plugin.content.backToPlugins') }}</el-button>
        <el-button type="primary" data-testid="plugin-content-query" @click="load">{{ t('common.query') }}</el-button>
      </div>
    </div>

    <el-alert
      v-if="plugin?.status === 'disabled'"
      :title="t('plugin.content.disabledHistoryTip')"
      type="warning"
      show-icon
      :closable="false"
      data-testid="plugin-content-disabled-tip"
    />
    <el-alert
      v-if="plugin?.status === 'archived'"
      :title="t('plugin.content.archivedTip')"
      type="warning"
      show-icon
      :closable="false"
      data-testid="plugin-content-archived-tip"
    />

    <div class="filter-panel">
      <el-select v-model="filters.communityId" clearable filterable :placeholder="t('field.community')" style="width: 180px" data-testid="plugin-content-community-filter">
        <el-option v-for="c in communities" :key="c.id" :label="`${c.name} /${c.slug}`" :value="c.id" />
      </el-select>
      <el-select v-model="filters.status" clearable :placeholder="t('plugin.status')" style="width: 150px" data-testid="plugin-content-status-filter">
        <el-option :label="t('common.all')" value="all" />
        <el-option :label="contentStatusLabel('publish')" value="publish" />
        <el-option :label="contentStatusLabel('hidden')" value="hidden" />
        <el-option :label="contentStatusLabel('offline')" value="offline" />
        <el-option :label="contentStatusLabel('pending')" value="pending" />
      </el-select>
      <el-select v-model="filters.contentType" clearable filterable :placeholder="t('plugin.contentType')" style="width: 180px" data-testid="plugin-content-type-filter">
        <el-option v-for="ct in contentTypes" :key="ct" :label="ct" :value="ct" />
      </el-select>
      <el-input v-model="keyword" :placeholder="t('plugin.content.searchPlaceholder')" clearable class="search" data-testid="plugin-content-search" @keyup.enter="load" />
    </div>

    <div class="batch-panel">
      <span class="muted">{{ t('common.selected') }} {{ selectedRows.length }} {{ t('common.selectedItems') }}</span>
      <div class="batch-actions">
        <el-button v-for="action in batchActions" :key="action.name" :type="action.type" :disabled="!selectedRows.length" :data-testid="`plugin-content-batch-${action.name}`" @click="batchUpdate(action.name)">
          {{ action.label }}
        </el-button>
      </div>
      <el-button :disabled="!lastAuditQuery" data-testid="plugin-content-audit" @click="openAuditLogs">{{ t('plugin.content.viewAuditLogs') }}</el-button>
    </div>

    <el-table :data="items" border stripe data-testid="plugin-content-table" @selection-change="onSelectionChange">
      <el-table-column type="selection" width="48" />
      <el-table-column prop="id" :label="t('field.id')" width="80" />
      <el-table-column prop="title" :label="t('field.title')" min-width="260" />
      <el-table-column prop="site" :label="t('field.community')" width="110" />
      <el-table-column prop="board" :label="t('plugin.content.board')" width="110" />
      <el-table-column :label="t('plugin.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="contentStatusType(row.status)" size="small">{{ contentStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.content.governanceFlags')" width="130">
        <template #default="{ row }">
          <el-tag v-if="row.pinned" type="warning" size="small">{{ t('plugin.content.pin') }}</el-tag>
          <el-tag v-if="row.recommended" type="success" size="small">{{ t('plugin.content.feature') }}</el-tag>
          <span v-if="!row.pinned && !row.recommended" class="muted">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="comments" :label="t('plugin.content.comments')" width="110" />
      <el-table-column prop="updated_at" :label="t('plugin.content.updatedAt')" width="170" />
      <el-table-column :label="t('plugin.content.pluginAndType')" width="180">
        <template #default="{ row }">
          <div class="mono">{{ row.plugin_code || '-' }}</div>
          <div class="muted mono">{{ row.content_type || '-' }}</div>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.action')" width="210" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">{{ t('plugin.content.viewDetail') }}</el-button>
          <el-button link type="primary" @click="open(row)">{{ t('common.view') }}</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>

  <el-drawer v-model="detailDrawer" :title="t('plugin.content.detailTitle')" size="620px" data-testid="plugin-content-detail-drawer">
    <el-descriptions v-if="detailTarget" :column="1" border>
      <el-descriptions-item :label="t('field.id')">{{ detailTarget.id }}</el-descriptions-item>
      <el-descriptions-item :label="t('field.title')">{{ detailTarget.title }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.code')">{{ detailTarget.plugin_code || '-' }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.contentType')">{{ detailTarget.content_type || '-' }}</el-descriptions-item>
      <el-descriptions-item :label="t('field.community')">{{ detailTarget.site || '-' }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.content.board')">{{ detailTarget.board || '-' }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.status')">{{ contentStatusLabel(detailTarget.status) }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.content.comments')">{{ detailTarget.comments || 0 }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.content.updatedAt')">{{ detailTarget.updated_at || '-' }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.content.recentGovernance')">
        <el-button link type="primary" data-testid="plugin-content-detail-audit" @click="openAuditLogsForRow(detailTarget)">{{ t('plugin.content.viewAuditLogs') }}</el-button>
      </el-descriptions-item>
      <el-descriptions-item :label="t('field.description')">{{ detailTarget.summary || detailTarget.excerpt || '-' }}</el-descriptions-item>
    </el-descriptions>
    <template #footer>
      <el-button @click="detailDrawer = false">{{ t('common.close') }}</el-button>
      <el-button type="primary" @click="open(detailTarget)">{{ t('common.view') }}</el-button>
    </template>
  </el-drawer>

  <el-dialog v-model="batchResultVisible" :title="t('plugin.content.batchResultTitle')" width="720px" data-testid="plugin-content-batch-result">
    <el-descriptions :column="3" border class="result-summary">
      <el-descriptions-item :label="t('plugin.content.batchAction')">{{ actionLabel(lastBatchAction) }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.content.succeededCount')">{{ batchResult?.updated || 0 }}</el-descriptions-item>
      <el-descriptions-item :label="t('plugin.content.failedCount')">{{ batchResult?.failed || 0 }}</el-descriptions-item>
    </el-descriptions>
    <el-table :data="batchResult?.items || []" border stripe max-height="320">
      <el-table-column prop="id" :label="t('field.id')" width="110" />
      <el-table-column :label="t('plugin.content.resultStatus')" width="120">
        <template #default="{ row }">
          <el-tag :type="row.ok ? 'success' : 'danger'" size="small">{{ row.ok ? t('plugin.content.succeeded') : t('plugin.content.failed') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="error" :label="t('plugin.content.failureReason')" min-width="260">
        <template #default="{ row }">{{ row.error || '-' }}</template>
      </el-table-column>
    </el-table>
    <template #footer>
      <el-button @click="batchResultVisible = false">{{ t('common.close') }}</el-button>
      <el-button type="primary" data-testid="plugin-content-batch-audit" @click="openAuditLogs">{{ t('plugin.content.viewAuditLogs') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import { adminCommunities, batchTopics, plugins, posts } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';
import { t } from '@/i18n';
import { contentStatusLabel, pluginHealthLabel, pluginStatusLabel, statusTagType } from '@/i18n/formatters';

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const items = ref([]);
const keyword = ref('');
const plugin = ref(null);
const communities = ref([]);
const filters = ref({ communityId: null, status: 'all', contentType: '' });
const selectedRows = ref([]);
const detailDrawer = ref(false);
const detailTarget = ref(null);
const lastAuditQuery = ref(null);
const lastBatchAction = ref('');
const batchResult = ref(null);
const batchResultVisible = ref(false);
const contentTypes = ref([]);

const contentTypeCount = computed(() => contentTypes.value.length || 1);
const healthStatus = computed(() => plugin.value?.health_status || plugin.value?.runtime_status || plugin.value?.status || 'unknown');
const batchActions = computed(() => [
  { name: 'hide', label: t('plugin.content.batchHide'), type: 'warning' },
  { name: 'restore', label: t('plugin.content.batchRestore'), type: 'success' },
  { name: 'approve', label: t('plugin.content.approve'), type: 'success' },
  { name: 'reject', label: t('plugin.content.reject'), type: 'danger' },
  { name: 'pin', label: t('plugin.content.pin'), type: 'warning' },
  { name: 'unpin', label: t('plugin.content.unpin'), type: 'info' },
  { name: 'feature', label: t('plugin.content.feature'), type: 'success' },
  { name: 'unfeature', label: t('plugin.content.unfeature'), type: 'info' },
]);

async function load() {
  const pluginList = await plugins();
  const current = (pluginList.items || []).find((item) => item.code === route.meta.pluginCode);
  if (!current) {
    ElMessage.warning(t('plugin.content.disabledTip'));
    router.replace('/plugins');
    return;
  }
  plugin.value = current;
  contentTypes.value = current.content_types?.length ? current.content_types : [route.meta.contentType];
  const permission = current.menus?.find((item) => item.area === 'admin')?.permission || route.meta.permission;
  if (permission && !auth.can(permission)) {
    ElMessage.warning(t('plugin.content.noPermissionTip'));
    router.replace('/plugins');
    return;
  }
  if (!communities.value.length) {
    const commData = await adminCommunities();
    communities.value = commData.items || [];
  }

  const community = filters.value.communityId ? communities.value.find((c) => c.id === filters.value.communityId) : null;
  const site = community?.slug || 'portal';
  const targetType = filters.value.contentType || route.meta.contentType;
  const data = await posts({
    site,
    board: 'all',
    q: keyword.value,
    plugin_code: route.meta.pluginCode,
    content_type: targetType,
    status: filters.value.status || 'all',
  });
  items.value = data.items || data || [];
  selectedRows.value = [];
}

function open(row) {
  if (!row) return;
  window.open(`/topics/${row.id}/`, '_blank');
}

function backToPlugins() {
  router.push('/plugins');
}

function openDetail(row) {
  detailTarget.value = row;
  detailDrawer.value = true;
}

function onSelectionChange(rows) {
  selectedRows.value = rows || [];
}

function actionLabel(action) {
  return t(`plugin.content.${action}`) || action;
}

async function batchUpdate(action) {
  const ids = selectedRows.value.map((row) => row.id).filter(Boolean);
  if (!ids.length) return;
  const label = actionLabel(action);
  await ElMessageBox.confirm(t('plugin.content.batchConfirm', { count: ids.length, action: label }), t('plugin.content.batchConfirmTitle'), {
    type: action === 'hide' || action === 'reject' ? 'warning' : 'info',
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  });
  const result = await batchTopics({
    ids,
    action,
    note: `PluginContent ${action} ${route.meta.pluginCode}/${filters.value.contentType || route.meta.contentType}`,
  });
  batchResult.value = result;
  lastBatchAction.value = action;
  lastAuditQuery.value = {
    plugin_code: route.meta.pluginCode,
    content_type: filters.value.contentType || route.meta.contentType,
    action,
  };
  batchResultVisible.value = true;
  ElMessage.success(t('plugin.content.batchDone'));
  await load();
}

function auditQuery(extra = {}) {
  return {
    action: '批量治理主题',
    target_type: 'topics',
    metadata: route.meta.pluginCode,
    plugin_code: route.meta.pluginCode,
    content_type: filters.value.contentType || route.meta.contentType,
    operation: lastBatchAction.value ? `batch:${lastBatchAction.value}` : undefined,
    ...extra,
  };
}

function openAuditLogs() {
  router.push({
    path: '/admin-next/audit-logs',
    query: auditQuery(lastAuditQuery.value || {}),
  });
}

function openAuditLogsForRow(row) {
  if (!row) return;
  router.push({
    path: '/admin-next/audit-logs',
    query: auditQuery({
      metadata: row.plugin_code || route.meta.pluginCode,
      content_type: row.content_type || filters.value.contentType || route.meta.contentType,
      target_id: row.id,
    }),
  });
}

function contentStatusType(status) {
  if (status === 'publish') return 'success';
  if (status === 'hidden' || status === 'rejected' || status === 'offline') return 'danger';
  if (status === 'pending' || status === 'draft') return 'warning';
  return 'info';
}

watch(() => route.meta.contentType, load);
onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 16px; }
.plugin-content-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
}
.header-main { min-width: 0; }
.eyebrow {
  margin: 0 0 6px;
  color: #64748b;
  font-size: 12px;
}
.plugin-content-header h2 { margin: 0 0 8px; font-size: 22px; }
.meta-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px 14px;
  color: #64748b;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
}
.filter-panel,
.batch-panel {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
}
.batch-panel { justify-content: space-between; }
.batch-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  flex: 1;
}
.search { max-width: 320px; }
.muted { color: #64748b; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.result-summary { margin-bottom: 12px; }
@media (max-width: 900px) {
  .plugin-content-header { flex-direction: column; }
  .header-actions { width: 100%; justify-content: flex-start; }
  .search { max-width: none; width: 100%; }
}
</style>
