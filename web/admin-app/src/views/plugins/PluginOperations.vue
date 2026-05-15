<template>
  <div class="plugin-operations" data-testid="plugin-operations-page">
    <div class="page-header">
      <div>
        <h1>插件操作历史</h1>
        <p class="desc">安装/升级前会生成操作快照；失败时可查看恢复预览并执行 cleanup。回滚仅提供 dry-run 预览，不支持 migration down。</p>
      </div>
      <div class="actions">
        <el-button :loading="loading" @click="load">刷新</el-button>
      </div>
    </div>

    <el-card shadow="never" class="filters">
      <div class="filter-row">
        <el-input v-model="filters.plugin_code" placeholder="plugin_code" clearable style="width: 220px" />
        <el-select v-model="filters.operation_type" placeholder="类型" clearable style="width: 160px">
          <el-option label="install" value="install" />
          <el-option label="upgrade" value="upgrade" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 160px">
          <el-option label="created" value="created" />
          <el-option label="applied" value="applied" />
          <el-option label="failed" value="failed" />
          <el-option label="recovered" value="recovered" />
        </el-select>
        <el-button type="primary" :loading="loading" @click="load">查询</el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" data-testid="plugin-operations-table">
        <el-table-column prop="operation_id" label="operation_id" min-width="220" />
        <el-table-column prop="operation_type" label="type" width="110" />
        <el-table-column prop="plugin_code" label="plugin_code" width="180" />
        <el-table-column prop="from_version" label="from" width="110" />
        <el-table-column prop="to_version" label="to" width="110" />
        <el-table-column prop="package_source" label="source" width="140" />
        <el-table-column prop="status" label="status" width="110" />
        <el-table-column prop="error_code" label="error" min-width="180" />
        <el-table-column prop="created_at" label="created_at" width="180" />
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
          <el-descriptions-item label="operation_id">{{ detail.operation_id }}</el-descriptions-item>
          <el-descriptions-item label="type">{{ detail.operation_type }}</el-descriptions-item>
          <el-descriptions-item label="plugin_code">{{ detail.plugin_code }}</el-descriptions-item>
          <el-descriptions-item label="status">{{ detail.status }}</el-descriptions-item>
          <el-descriptions-item label="from_version">{{ detail.from_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="to_version">{{ detail.to_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="package_source">{{ detail.package_source || '-' }}</el-descriptions-item>
          <el-descriptions-item label="created_at">{{ detail.created_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="error_code">{{ detail.error_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="error_message">{{ detail.error_message || '-' }}</el-descriptions-item>
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
            cleanup
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

        <el-alert v-if="!canManage" type="warning" show-icon title="缺少 plugin.manage：只能查看，不能 recover/cleanup/rollback 预览。" />

        <el-divider />

        <h3>Dry-run / 风险快照</h3>
        <pre class="code">{{ pretty(detail.dry_run_json) }}</pre>

        <el-divider />

        <h3>恢复预览</h3>
        <pre class="code">{{ pretty(recoverResult) }}</pre>

        <el-divider />

        <h3>升级回滚 dry-run</h3>
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

const load = async () => {
  loading.value = true;
  try {
    const { data } = await listPluginOperations({
      plugin_code: filters.value.plugin_code || undefined,
      operation_type: filters.value.operation_type || undefined,
      status: filters.value.status || undefined,
      page: 1,
      page_size: 50,
    });
    items.value = data.items || [];
  } catch (e) {
    ElMessage.error(String(e?.response?.data?.message || e?.response?.data?.error || e?.message || '加载失败'));
  } finally {
    loading.value = false;
  }
};

const openDetail = async (operationId) => {
  detailVisible.value = true;
  recoverResult.value = null;
  rollbackResult.value = null;
  try {
    const { data } = await getPluginOperation(operationId);
    detail.value = data;
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
    const { data } = await recoverPluginOperationDryRun(detail.value.operation_id);
    recoverResult.value = data;
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
  await ElMessageBox.confirm('仅清理本次失败操作产生的残留记录；不会删除历史内容，也不支持 migration down。确认执行 cleanup？', '确认 cleanup', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning',
  });
  cleanupLoading.value = true;
  try {
    const { data } = await cleanupPluginOperation(detail.value.operation_id);
    ElMessage.success(data?.status === 'ok' ? 'cleanup 完成' : 'cleanup 已执行');
    await openDetail(detail.value.operation_id);
    await load();
  } catch (e) {
    ElMessage.error(String(e?.response?.data?.message || e?.response?.data?.error || e?.message || 'cleanup 失败'));
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
    const { data } = await dryRunPluginUpgradeRollback(detail.value.plugin_code, { operation_id: detail.value.operation_id });
    rollbackResult.value = data;
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

