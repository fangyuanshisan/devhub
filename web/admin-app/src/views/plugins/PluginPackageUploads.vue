<template>
  <section class="plugin-upload-page" data-testid="plugin-package-upload-lifecycle-page">
    <header class="page-head">
      <div>
        <h2>上传包管理</h2>
        <p>zip 上传包只进入安全暂存区；转入本地仓库后仍不等于安装。</p>
      </div>
      <div class="head-actions">
        <el-button :loading="cleanupLoading" data-testid="upload-cleanup" @click="cleanupUploads">清理失效包</el-button>
        <el-button type="primary" :loading="loading" data-testid="upload-refresh" @click="fetchUploads">刷新</el-button>
      </div>
    </header>

    <PluginPackageBoundaryNotice title="上传包只是暂存区文件；不会执行插件代码、不会执行 SQL、不会加载前端资产。" />

    <section class="toolbar">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :limit="1"
        accept=".zip"
        :on-change="onFileChange"
        :on-remove="onFileRemove"
        data-testid="upload-zip-picker"
      >
        <el-button>选择 zip</el-button>
      </el-upload>
      <el-button type="primary" :loading="uploading" data-testid="upload-zip-submit" @click="submitUpload">上传到安全沙箱</el-button>
      <span class="muted">限制：zip 20MB、解压 50MB、单文件 5MB、最多 300 文件；嵌套压缩包、路径穿越和 symlink 会被阻断。</span>
    </section>

    <PluginErrorAlert v-if="errorText" class="mb" :message="errorText" data-testid="upload-error" />

    <PluginFilterBar title="筛选" tip="仅用于列表筛选；详情内操作仍由后端校验状态流转。" testid="upload-filter-bar">
      <el-input v-model="filters.keyword" data-testid="upload-filter-keyword" placeholder="搜索上传 ID / 文件名 / 插件编码" clearable @keyup.enter="fetchUploads" />
      <el-select v-model="filters.status" data-testid="upload-filter-status" clearable placeholder="状态">
        <el-option label="全部" value="all" />
        <el-option v-for="s in statuses" :key="s" :label="genericStatusLabel(s)" :value="s" />
      </el-select>
      <el-select v-model="filters.risk_level" clearable placeholder="风险">
        <el-option label="全部" value="all" />
        <el-option label="低风险" value="low" />
        <el-option label="中风险" value="medium" />
        <el-option label="高风险" value="high" />
        <el-option label="已阻断" value="blocked" />
      </el-select>
      <el-button data-testid="upload-filter-submit" @click="fetchUploads">筛选</el-button>
    </PluginFilterBar>

    <PluginEmptyState v-if="!loading && !items.length" testid="upload-empty-state" description="暂无上传包记录" />

    <el-table v-loading="loading" :data="items" data-testid="upload-list" border>
      <el-table-column prop="upload_id" label="上传 ID" min-width="190" />
      <el-table-column prop="original_filename" label="文件名" min-width="160" />
      <el-table-column prop="package_code" label="插件编码" min-width="120" />
      <el-table-column prop="package_version" label="版本" width="100" />
      <el-table-column prop="status" label="状态" width="160">
        <template #default="{ row }"><PluginStatusTag :value="row.status" /></template>
      </el-table-column>
      <el-table-column prop="risk_level" label="风险" width="110">
        <template #default="{ row }"><PluginRiskTag :level="row.risk_level" /></template>
      </el-table-column>
      <el-table-column prop="checksum_status" label="校验" width="120">
        <template #default="{ row }">{{ genericStatusLabel(row.checksum_status) }}</template>
      </el-table-column>
      <el-table-column prop="signature_status" label="签名" width="130">
        <template #default="{ row }">{{ genericStatusLabel(row.signature_status) }}</template>
      </el-table-column>
      <el-table-column prop="trust_status" label="信任状态" width="110">
        <template #default="{ row }">{{ trustLevelLabel(row.trust_status) }}</template>
      </el-table-column>
      <el-table-column prop="uploaded_by_name" label="上传人" width="120" />
      <el-table-column prop="expires_at" label="过期时间" width="170" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" data-testid="upload-open-detail" @click="openDetail(row.upload_id)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="drawer" size="720px" title="上传包详情" data-testid="upload-detail-drawer">
      <template v-if="detail">
        <el-alert class="mb" type="warning" show-icon :closable="false" title="转入本地仓库只复制到 storage/plugins/packages，不会安装插件；安装仍走本地包审批与安装流程。" />
        <el-descriptions :column="2" border>
          <el-descriptions-item label="上传 ID">{{ detail.record.upload_id }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ genericStatusLabel(detail.record.status) }}</el-descriptions-item>
          <el-descriptions-item label="插件">{{ detail.record.package_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ detail.record.package_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="暂存路径">{{ detail.record.staging_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="本地包路径">{{ detail.record.package_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="转入仓库路径">{{ detail.record.promoted_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审批 ID">{{ detail.record.approval_id || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="actions" data-testid="upload-actions">
          <el-button :disabled="!actionEnabled('rescan')" :loading="actionLoading === 'rescan'" data-testid="upload-rescan" @click="runAction('rescan')">重新扫描</el-button>
          <el-button :disabled="!actionEnabled('submit_approval')" :loading="actionLoading === 'submit_approval'" data-testid="upload-submit-approval" @click="runAction('submit_approval')">提交导入审批</el-button>
          <el-button :disabled="!actionEnabled('approve')" :loading="actionLoading === 'approve'" data-testid="upload-approve" @click="runAction('approve')">审批通过</el-button>
          <el-button :disabled="!actionEnabled('reject')" :loading="actionLoading === 'reject'" data-testid="upload-reject" @click="runAction('reject')">拒绝</el-button>
          <el-button type="primary" :disabled="!actionEnabled('promote')" :loading="actionLoading === 'promote'" data-testid="upload-promote" @click="runAction('promote')">转入本地仓库</el-button>
          <el-button :disabled="!actionEnabled('cancel')" :loading="actionLoading === 'cancel'" data-testid="upload-cancel" @click="runAction('cancel')">取消</el-button>
          <el-button type="danger" :disabled="!actionEnabled('delete')" :loading="actionLoading === 'delete'" data-testid="upload-delete" @click="runAction('delete')">删除</el-button>
        </div>

        <el-alert v-if="disabledReasons" class="mb" type="info" show-icon :closable="false" :title="disabledReasons" data-testid="upload-action-reasons" />

        <el-alert class="mb" type="info" show-icon :closable="false" :title="uploadSuggestion" data-testid="upload-suggestion" />

        <el-collapse class="technical-collapse" data-testid="upload-technical-details">
          <el-collapse-item name="technical">
            <template #title>
              <strong>技术详情</strong>
              <span class="muted ml8">原始扫描、清单文件校验、风险报告和安装预检 JSON</span>
            </template>
            <PluginEmptyState v-if="!hasTechnicalDetails" testid="upload-technical-empty" description="暂无技术详情" />
            <div v-else class="detail-grid">
              <section>
                <h3>zip 扫描</h3>
                <pre data-testid="upload-zip-scan">{{ pretty(detail.zip_scan) }}</pre>
              </section>
              <section>
                <h3>插件包扫描</h3>
                <pre data-testid="upload-file-scan">{{ pretty(detail.file_scan) }}</pre>
              </section>
              <section>
                <h3>校验与签名</h3>
                <pre data-testid="upload-checksum">{{ pretty({ checksum: detail.checksum, signature: detail.signature }) }}</pre>
              </section>
              <section>
                <h3>错误详情 / 风险报告</h3>
                <pre data-testid="upload-risk-report">{{ pretty(detail.risk_report) }}</pre>
              </section>
              <section>
                <h3>清单文件校验</h3>
                <pre data-testid="upload-manifest-validation">{{ pretty(detail.manifest_validation) }}</pre>
              </section>
              <section>
                <h3>安装预检</h3>
                <pre data-testid="upload-dry-run">{{ pretty(detail.install_dry_run) }}</pre>
              </section>
            </div>
          </el-collapse-item>
        </el-collapse>
      </template>
    </el-drawer>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import PluginEmptyState from './components/PluginEmptyState.vue';
import PluginErrorAlert from './components/PluginErrorAlert.vue';
import PluginFilterBar from './components/PluginFilterBar.vue';
import PluginPackageBoundaryNotice from './components/PluginPackageBoundaryNotice.vue';
import PluginRiskTag from './components/PluginRiskTag.vue';
import PluginStatusTag from './components/PluginStatusTag.vue';
import {
  approvePluginPackageUpload,
  cancelPluginPackageUpload,
  cleanupPluginPackageUploads,
  deletePluginPackageUpload,
  getPluginPackageUpload,
  listPluginPackageUploads,
  promotePluginPackageUpload,
  rejectPluginPackageUpload,
  rescanPluginPackageUpload,
  submitPluginPackageUploadApproval,
  uploadPluginPackageZip,
} from '@/api/plugins';
import { genericStatusLabel, trustLevelLabel } from '@/i18n/formatters';
import { pluginMessageText, pluginReasonText } from '@/modules/plugins/statusText';

const statuses = ['uploaded', 'scanned', 'staged', 'blocked', 'approval_pending', 'approval_rejected', 'approved', 'promoted', 'install_approval_pending', 'installed', 'canceled', 'expired', 'deleted', 'failed'];
const filters = reactive({ keyword: '', status: 'all', risk_level: 'all', page: 1, page_size: 20 });
const items = ref([]);
const detail = ref(null);
const drawer = ref(false);
const loading = ref(false);
const uploading = ref(false);
const cleanupLoading = ref(false);
const actionLoading = ref('');
const selectedFile = ref(null);
const uploadRef = ref();
const errorText = ref('');

const actionMap = computed(() => Object.fromEntries((detail.value?.actions || []).map((item) => [item.action, item])));
const disabledReasons = computed(() => (detail.value?.actions || [])
  .filter((item) => !item.enabled && item.reason)
  .map((item) => `${pluginReasonText(item.reason_code)}：${item.reason}`)
  .join('；'));
const hasTechnicalDetails = computed(() => [
  detail.value?.zip_scan,
  detail.value?.file_scan,
  detail.value?.checksum,
  detail.value?.signature,
  detail.value?.risk_report,
  detail.value?.manifest_validation,
  detail.value?.install_dry_run,
].some((item) => item && Object.keys(item || {}).length));
const uploadSuggestion = computed(() => {
  const status = String(detail.value?.record?.status || '');
  if (status === 'promoted') return '该上传包已转入本地仓库；后续仍需走安装审批和启用前检查。';
  if (status === 'blocked' || status === 'failed') return '该上传包存在阻断或失败，请展开技术详情查看错误详情后重新上传或重新扫描。';
  if (status === 'approval_pending' || status === 'install_approval_pending') return '该上传包正在等待审批，请先完成审批再继续。';
  return '先看状态和可用操作；原始 JSON 已折叠到技术详情，适合排障时展开。';
});

onMounted(async () => {
  await fetchUploads();
});

function actionEnabled(name) {
  return Boolean(actionMap.value[name]?.enabled);
}

function pretty(value) {
  return JSON.stringify(value || {}, null, 2);
}

function apiError(error) {
  const data = error?.response?.data?.error || error?.response?.data || {};
  return pluginMessageText(data, error?.message || '操作失败');
}

async function fetchUploads() {
  loading.value = true;
  errorText.value = '';
  try {
    const res = await listPluginPackageUploads(filters);
    items.value = res.items || [];
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    loading.value = false;
  }
}

function onFileChange(file) {
  selectedFile.value = file.raw;
}

function onFileRemove() {
  selectedFile.value = null;
}

async function submitUpload() {
  if (!selectedFile.value) {
    errorText.value = '请先选择 .zip 插件包。';
    return;
  }
  uploading.value = true;
  errorText.value = '';
  try {
    const form = new FormData();
    form.append('file', selectedFile.value);
    const res = await uploadPluginPackageZip(form);
    uploadRef.value?.clearFiles?.();
    selectedFile.value = null;
    ElMessage.success(`上传完成：${genericStatusLabel(res.status || res.record?.status)}`);
    await fetchUploads();
    await openDetail(res.upload_id || res.record?.upload_id);
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    uploading.value = false;
  }
}

async function openDetail(uploadId) {
  if (!uploadId) return;
  const res = await getPluginPackageUpload(uploadId);
  detail.value = res;
  drawer.value = true;
}

async function runAction(name) {
  if (!detail.value?.record?.upload_id) return;
  const uploadId = detail.value.record.upload_id;
  try {
    if (name === 'promote') {
      await ElMessageBox.confirm('确认将该上传包转入本地插件仓库？这不会安装插件，也不会执行第三方代码。', '转入仓库确认', { type: 'warning' });
    }
    if (name === 'delete') {
      await ElMessageBox.confirm('确认删除该上传包记录？此操作不会删除已安装插件，但会影响该上传包后续治理。', '删除确认', { type: 'warning' });
    }
    actionLoading.value = name;
    errorText.value = '';
    if (name === 'rescan') detail.value = await rescanPluginPackageUpload(uploadId);
    if (name === 'submit_approval') detail.value = await submitPluginPackageUploadApproval(uploadId, { reason: '后台上传包导入审批' });
    if (name === 'approve') detail.value = await approvePluginPackageUpload(uploadId, { comment: '同意导入' });
    if (name === 'reject') detail.value = await rejectPluginPackageUpload(uploadId, { comment: '拒绝导入' });
    if (name === 'promote') {
      await promotePluginPackageUpload(uploadId, { force: false });
      detail.value = await getPluginPackageUpload(uploadId);
    }
    if (name === 'cancel') detail.value = await cancelPluginPackageUpload(uploadId);
    if (name === 'delete') detail.value = await deletePluginPackageUpload(uploadId);
    ElMessage.success('操作完成');
    await fetchUploads();
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    errorText.value = apiError(error);
  } finally {
    actionLoading.value = '';
  }
}

async function cleanupUploads() {
  cleanupLoading.value = true;
  errorText.value = '';
  try {
    const res = await cleanupPluginPackageUploads();
    ElMessage.success(`清理完成：${res.cleaned || 0}`);
    await fetchUploads();
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    cleanupLoading.value = false;
  }
}
</script>

<style scoped>
.plugin-upload-page {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.page-head,
.toolbar,
.filters,
.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.download-card {
  padding: 14px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  background: var(--el-bg-color);
}
.download-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 12px;
}
.download-head h3 {
  margin: 0 0 4px;
}
.download-form {
  max-width: 880px;
}
.page-head {
  justify-content: space-between;
}
.page-head h2 {
  margin: 0 0 4px;
}
.page-head p,
.muted {
  color: var(--el-text-color-secondary);
  margin: 0;
}
.filters .el-input {
  width: 280px;
}
.filters .el-select {
  width: 150px;
}
.mb {
  margin-bottom: 12px;
}
.danger {
  color: var(--el-color-danger);
}
.compat-result {
  margin-top: 12px;
}
.detail-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}
.technical-collapse {
  margin-top: 12px;
}
.ml8 {
  margin-left: 8px;
}
.detail-grid h3 {
  margin: 10px 0 6px;
  font-size: 14px;
}
pre {
  max-height: 240px;
  overflow: auto;
  padding: 10px;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  background: var(--el-fill-color-lighter);
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
}
</style>
