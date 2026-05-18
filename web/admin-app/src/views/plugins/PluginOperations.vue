<template>
  <div class="plugin-operations" data-testid="plugin-operations-page">
    <div class="page-header">
      <div>
        <h1>插件操作历史</h1>
        <p class="desc">安装/升级前会生成操作快照；失败时可查看恢复预览并清理残留。回滚仅提供预检预览，不支持 migration down。</p>
      </div>
      <div class="actions">
        <el-button :loading="loading" @click="load">刷新</el-button>
      </div>
    </div>

    <el-card shadow="never" class="filters">
      <div class="filter-row">
        <el-input v-model="filters.plugin_code" placeholder="插件编码" clearable style="width: 220px" />
        <el-select v-model="filters.operation_type" placeholder="类型" clearable style="width: 160px">
          <el-option label="安装" value="install" />
          <el-option label="升级" value="upgrade" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 160px">
          <el-option label="已创建" value="created" />
          <el-option label="已应用" value="applied" />
          <el-option label="失败" value="failed" />
          <el-option label="已恢复" value="recovered" />
        </el-select>
        <el-button type="primary" :loading="loading" @click="load">查询</el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" empty-text="暂无操作记录" data-testid="plugin-operations-table">
        <el-table-column prop="operation_id" label="操作 ID" min-width="220" />
        <el-table-column prop="operation_type" label="类型" width="110">
          <template #default="{ row }">{{ operationTypeLabel(row.operation_type) }}</template>
        </el-table-column>
        <el-table-column prop="plugin_code" label="插件编码" width="180" />
        <el-table-column prop="from_version" label="原版本" width="110" />
        <el-table-column prop="to_version" label="目标版本" width="110" />
        <el-table-column prop="package_source" label="来源" width="140">
          <template #default="{ row }">{{ packageSourceLabel(row.package_source) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">{{ genericStatusLabel(row.status) }}</template>
        </el-table-column>
        <el-table-column prop="error_code" label="错误" min-width="180" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" data-testid="plugin-operations-detail-btn" @click="openDetail(row.operation_id)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="!loading && items.length === 0" class="empty">暂无操作记录</div>
    </el-card>

    <el-drawer v-model="detailVisible" size="70%" title="操作详情" data-testid="plugin-operation-detail-drawer">
      <div v-if="detail" class="detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="操作 ID">{{ detail.operation_id }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ operationTypeLabel(detail.operation_type) }}</el-descriptions-item>
          <el-descriptions-item label="插件编码">{{ detail.plugin_code }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ genericStatusLabel(detail.status) }}</el-descriptions-item>
          <el-descriptions-item label="原版本">{{ detail.from_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="目标版本">{{ detail.to_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="来源">{{ packageSourceLabel(detail.package_source) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detail.created_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="错误码">{{ detail.error_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="错误信息">{{ detail.error_message || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-actions">
          <el-button
            data-testid="plugin-operation-recover-dryrun"
            :loading="recoverLoading"
            @click="recoverDryRun"
          >
            恢复预览
          </el-button>
          <el-button
            data-testid="plugin-operation-cleanup"
            type="danger"
            :disabled="!canManage || detail.status !== 'failed'"
            :loading="cleanupLoading"
            @click="doCleanup"
          >
            清理残留
          </el-button>
          <el-button
            data-testid="plugin-operation-rollback-dryrun"
            :disabled="!canManage || detail.operation_type !== 'upgrade'"
            :loading="rollbackLoading"
            @click="rollbackDryRun"
          >
            升级回滚预览
          </el-button>
        </div>

        <el-alert v-if="!canManage" type="warning" show-icon title="缺少 plugin.manage：只能查看，不能执行恢复预览、清理残留或回滚预览。" />

        <el-divider />

        <h3>预检与风险快照</h3>
        <pre class="code">{{ pretty(detail.dry_run_json) }}</pre>

        <el-divider />

        <h3>恢复预览</h3>
        <pre class="code">{{ pretty(recoverResult) }}</pre>

        <el-divider />

        <h3>升级回滚预检</h3>
        <pre class="code" data-testid="plugin-operation-rollback-result">{{ pretty(rollbackResult) }}</pre>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { cleanupPluginOperation, dryRunPluginUpgradeRollback, getPluginOperation, listPluginOperations, recoverPluginOperationDryRun } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';
import { genericStatusLabel } from '@/i18n/formatters';

const auth = useAuthStore();
const canManage = computed(() => auth.can('plugin.manage'));

const loading = ref(false);
const items = ref([]);
const filters = ref({ plugin_code: '', operation_type: '', status: '' });

const detailVisible = ref(false);
const detail = ref(null);
const recoverLoading = ref(false);
const cleanupLoading = ref(false);
const rollbackLoading = ref(false);
const recoverResult = ref(null);
const rollbackResult = ref(null);

function responseData(response) {
  return response?.data && typeof response.data === 'object' ? response.data : response;
}

function listItems(response) {
  const data = responseData(response);
  if (Array.isArray(data?.items)) return data.items.filter(Boolean);
  if (Array.isArray(data)) return data.filter(Boolean);
  return [];
}

function operationTypeLabel(type) {
  const labels = { install: '安装', upgrade: '升级' };
  return labels[type] || type || '-';
}

function packageSourceLabel(source) {
  const labels = {
    local_package: '本地插件包',
    remote_package: '远程插件包',
    upload_package: '上传包',
  };
  return labels[source] || source || '-';
}

const load = async () => {
  loading.value = true;
  try {
    const response = await listPluginOperations({
      plugin_code: filters.value.plugin_code || undefined,
      operation_type: filters.value.operation_type || undefined,
      status: filters.value.status || undefined,
      page: 1,
      page_size: 50,
    });
    items.value = listItems(response);
  } catch (e) {
    items.value = [];
    ElMessage.error(String(e?.response?.data?.message || e?.response?.data?.error || e?.message || '加载失败'));
  } finally {
    loading.value = false;
  }
};

const openDetail = async (operationId) => {
  detailVisible.value = true;
  recoverResult.value = null;
  rollbackResult.value = null;
  detail.value = null;
  try {
    const response = await getPluginOperation(operationId);
    detail.value = responseData(response) || null;
  } catch (e) {
    ElMessage.error(String(e?.response?.data?.message || e?.response?.data?.error || e?.message || '加载详情失败'));
  }
};

const recoverDryRun = async () => {
  if (!detail.value) return;
  if (!canManage.value) {
    ElMessage.warning('缺少 plugin.manage');
    return;
  }
  recoverLoading.value = true;
  try {
    const response = await recoverPluginOperationDryRun(detail.value.operation_id);
    recoverResult.value = responseData(response);
  } catch (e) {
    ElMessage.error(String(e?.response?.data?.message || e?.response?.data?.error || e?.message || '恢复预览失败'));
  } finally {
    recoverLoading.value = false;
  }
};

const doCleanup = async () => {
  if (!detail.value) return;
  if (!canManage.value) {
    ElMessage.warning('缺少 plugin.manage');
    return;
  }
  await ElMessageBox.confirm('仅清理本次失败操作产生的残留记录；不会删除历史内容，也不支持 migration down。确认执行清理残留？', '清理残留确认', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning',
  });
  cleanupLoading.value = true;
  try {
    const response = await cleanupPluginOperation(detail.value.operation_id);
    const data = responseData(response);
    ElMessage.success(data?.status === 'ok' ? '清理完成' : '清理已执行');
    await openDetail(detail.value.operation_id);
    await load();
  } catch (e) {
    ElMessage.error(String(e?.response?.data?.message || e?.response?.data?.error || e?.message || '清理失败'));
  } finally {
    cleanupLoading.value = false;
  }
};

const rollbackDryRun = async () => {
  if (!detail.value) return;
  if (!canManage.value) {
    ElMessage.warning('缺少 plugin.manage');
    return;
  }
  rollbackLoading.value = true;
  try {
    const response = await dryRunPluginUpgradeRollback(detail.value.plugin_code, { operation_id: detail.value.operation_id });
    rollbackResult.value = responseData(response);
  } catch (e) {
    ElMessage.error(String(e?.response?.data?.message || e?.response?.data?.error || e?.message || '回滚预览失败'));
  } finally {
    rollbackLoading.value = false;
  }
};

const pretty = (v) => {
  if (!v) return '';
  if (typeof v === 'string') {
    try {
      return JSON.stringify(JSON.parse(v), null, 2);
    } catch (_) {
      return v;
    }
  }
  return JSON.stringify(v, null, 2);
};

onMounted(load);
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}
.desc {
  margin: 6px 0 0;
  color: #666;
}
.filters {
  margin-bottom: 12px;
}
.filter-row {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.empty {
  padding: 12px;
  color: #888;
}
.detail-actions {
  display: flex;
  gap: 10px;
  margin: 14px 0;
  flex-wrap: wrap;
}
.code {
  background: #0b1021;
  color: #e6e6e6;
  padding: 12px;
  border-radius: 8px;
  overflow: auto;
  max-height: 420px;
}
</style>
