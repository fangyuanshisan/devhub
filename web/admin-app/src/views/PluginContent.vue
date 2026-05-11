<template>
  <section class="page-card" data-testid="plugin-content-page">
    <div class="toolbar">
      <div>
        <h2>{{ route.meta.title }}</h2>
        <p>
          {{ t('plugin.code') }}：<span class="mono">{{ route.meta.pluginCode }}</span>
          · {{ t('plugin.contentType') }}：<span class="mono">{{ route.meta.contentType }}</span>
          · {{ t('plugin.status') }}：<el-tag v-if="plugin" :type="statusType(plugin.status)" size="small">{{ pluginStatusLabel(plugin.status) }}</el-tag>
        </p>
        <p class="muted">{{ t('plugin.content.intro') }}</p>
      </div>
      <div class="tool-actions">
        <el-select v-model="filters.communityId" clearable filterable :placeholder="t('field.community')" style="width: 160px" data-testid="plugin-content-community-filter">
          <el-option v-for="c in communities" :key="c.id" :label="`${c.name} /${c.slug}`" :value="c.id" />
        </el-select>
        <el-select v-model="filters.status" clearable :placeholder="t('plugin.status')" style="width: 140px" data-testid="plugin-content-status-filter">
          <el-option :label="t('common.all')" value="all" />
          <el-option :label="contentStatusLabel('publish')" value="publish" />
          <el-option :label="contentStatusLabel('hidden')" value="hidden" />
          <el-option :label="contentStatusLabel('pending')" value="pending" />
        </el-select>
        <el-select v-model="filters.contentType" clearable filterable :placeholder="t('plugin.contentType')" style="width: 160px" data-testid="plugin-content-type-filter">
          <el-option v-for="ct in contentTypes" :key="ct" :label="ct" :value="ct" />
        </el-select>
        <el-input v-model="keyword" :placeholder="t('plugin.content.searchPlaceholder')" clearable class="search" data-testid="plugin-content-search" @keyup.enter="load" />
        <el-button data-testid="plugin-content-back" @click="backToPlugins">{{ t('plugin.content.backToPlugins') }}</el-button>
        <el-button type="primary" data-testid="plugin-content-query" @click="load">{{ t('common.query') }}</el-button>
      </div>
    </div>
    <div class="batch-bar">
      <span class="muted">{{ t('common.selected') }} {{ selectedRows.length }} {{ t('common.selectedItems') }}</span>
      <el-button type="warning" :disabled="!selectedRows.length" data-testid="plugin-content-batch-hide" @click="batchUpdate('hide')">{{ t('plugin.content.batchHide') }}</el-button>
      <el-button type="success" :disabled="!selectedRows.length" data-testid="plugin-content-batch-restore" @click="batchUpdate('restore')">{{ t('plugin.content.batchRestore') }}</el-button>
      <el-button :disabled="!lastAuditQuery" @click="openAuditLogs">{{ t('plugin.content.viewAuditLogs') }}</el-button>
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
      <el-table-column prop="comments" :label="t('plugin.content.comments')" width="110" />
      <el-table-column prop="updated_at" :label="t('plugin.content.updatedAt')" width="170" />
      <el-table-column :label="t('plugin.content.pluginAndType')" width="180">
        <template #default="{ row }">
          <div class="mono">{{ row.plugin_code || '-' }}</div>
          <div class="muted mono">{{ row.content_type || '-' }}</div>
        </template>
      </el-table-column>
      <el-table-column :label="t('plugin.action')" width="190">
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
      <el-descriptions-item :label="t('field.description')">{{ detailTarget.summary || detailTarget.excerpt || '-' }}</el-descriptions-item>
    </el-descriptions>
    <template #footer>
      <el-button @click="detailDrawer = false">{{ t('common.close') }}</el-button>
      <el-button type="primary" @click="open(detailTarget)">{{ t('common.view') }}</el-button>
    </template>
  </el-drawer>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import { adminCommunities, batchTopics, plugins, posts } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';
import { t } from '@/i18n';
import { contentStatusLabel, pluginStatusLabel } from '@/i18n/formatters';

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
const lastAuditQuery = ref('');
const contentTypes = ref([]);

async function load() {
  const pluginList = await plugins();
  const current = (pluginList.items || []).find((item) => item.code === route.meta.pluginCode);
  if (!current || current.status !== 'enabled') {
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

  const site =
    filters.value.communityId && communities.value.find((c) => c.id === filters.value.communityId)?.slug
      ? communities.value.find((c) => c.id === filters.value.communityId)?.slug
      : 'portal';

  const targetType = filters.value.contentType || route.meta.contentType;
  const data = await posts({ site, board: 'all', q: keyword.value, content_type: targetType, status: filters.value.status || 'all' });
  const list = data.items || data || [];
  // Best-effort: keep only rows belonging to this plugin/type when data comes from legacy adminPosts.
  items.value = list.filter((r) => (r.plugin_code || route.meta.pluginCode) === route.meta.pluginCode || (r.content_type || targetType) === targetType);
  selectedRows.value = [];
}

function open(row) {
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

async function batchUpdate(action) {
  const ids = selectedRows.value.map((row) => row.id).filter(Boolean);
  if (!ids.length) return;
  const actionLabel = action === 'hide' ? t('plugin.content.hide') : t('plugin.content.restore');
  await ElMessageBox.confirm(t('plugin.content.batchConfirm', { count: ids.length, action: actionLabel }), t('plugin.content.batchConfirmTitle'), {
    type: action === 'hide' ? 'warning' : 'info',
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  });
  await batchTopics({ ids, action, note: `PluginContent ${action} ${route.meta.pluginCode}` });
  ElMessage.success(t('plugin.content.batchDone'));
  lastAuditQuery.value = route.meta.pluginCode;
  await load();
}

function openAuditLogs() {
  router.push({
    path: '/audit-logs',
    query: {
      action: '批量治理主题',
      target_type: 'topic',
      metadata: route.meta.pluginCode,
    },
  });
}

function statusType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'danger';
  return 'info';
}

function contentStatusType(status) {
  if (status === 'publish') return 'success';
  if (status === 'hidden' || status === 'rejected') return 'danger';
  if (status === 'pending' || status === 'draft') return 'warning';
  return 'info';
}

watch(() => route.meta.contentType, load);
onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 16px; }
.toolbar { display: flex; gap: 12px; align-items: flex-start; }
.toolbar h2 { margin: 0 0 6px; }
.toolbar p { margin: 0; color: #64748b; }
.muted { color: #64748b; }
.tool-actions { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; margin-left: auto; justify-content: flex-end; }
.search { max-width: 260px; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.batch-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #f8fafc;
}
</style>
