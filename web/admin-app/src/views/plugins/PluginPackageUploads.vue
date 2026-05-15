<template>
  <section class="plugin-upload-page" data-testid="plugin-package-upload-lifecycle-page">
    <header class="page-head">
      <div>
        <h2>上传包管理</h2>
        <p>zip 上传包只进入安全 staging；promote 后才进入本地插件仓库，promote 不等于安装。</p>
      </div>
      <div class="head-actions">
        <el-button :loading="cleanupLoading" data-testid="upload-cleanup" @click="cleanupUploads">清理失效包</el-button>
        <el-button type="primary" :loading="loading" data-testid="upload-refresh" @click="fetchUploads">刷新</el-button>
      </div>
    </header>

    <PluginPackageBoundaryNotice title="上传包只是 staging；不会执行插件代码、不会执行 SQL、不会加载前端资产。" />

    <section class="download-card" data-testid="remote-package-download-card">
      <div class="download-head">
        <div>
          <h3>远程插件包下载到 staging</h3>
          <p class="muted">仅下载到受控 staging，并校验 URL、SSRF、大小和 sha256；不会安装、启用、解压执行或加载插件代码。</p>
        </div>
        <el-button :loading="stagingLoading" data-testid="staging-refresh" @click="fetchStaging">刷新 staging</el-button>
      </div>
      <el-form class="download-form" :model="downloadForm" label-width="110px">
        <el-form-item label="plugin_code">
          <el-input v-model="downloadForm.plugin_code" data-testid="download-plugin-code" placeholder="demo_notice" />
        </el-form-item>
        <el-form-item label="version">
          <el-input v-model="downloadForm.version" data-testid="download-version" placeholder="1.0.0" />
        </el-form-item>
        <el-form-item label="package_url">
          <el-input v-model="downloadForm.package_url" data-testid="download-package-url" placeholder="https://example.com/demo_notice-1.0.0.zip" />
        </el-form-item>
        <el-form-item label="sha256">
          <el-input v-model="downloadForm.sha256" data-testid="download-sha256" placeholder="建议填写；缺失会标记 checksum_missing" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="downloading" data-testid="download-to-staging" @click="submitDownload">下载到 staging</el-button>
          <span class="muted">仅允许 https 与 .zip / .tar.gz / .tgz，默认最大 20MB。</span>
        </el-form-item>
      </el-form>
      <el-table v-loading="stagingLoading" :data="stagingItems" data-testid="staging-list" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="plugin_code" label="code" min-width="120" />
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="status" label="状态" width="150">
          <template #default="{ row }"><PluginStatusTag :value="row.status" /></template>
        </el-table-column>
        <el-table-column prop="file_size" label="大小" width="100" />
        <el-table-column prop="sha256_actual" label="sha256" min-width="220" show-overflow-tooltip />
        <el-table-column prop="staging_path" label="staging" min-width="220" show-overflow-tooltip />
        <el-table-column prop="error_message" label="错误" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" data-testid="staging-delete" @click="deleteStaging(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </section>

    <section class="download-card" data-testid="plugin-package-compat-card">
      <div class="download-head">
        <div>
          <h3>插件依赖 / 兼容性检查</h3>
          <p class="muted">只接受 precheck passed 的记录；检查依赖、Core 版本、权限、菜单、路由、Hook、config_schema 和 migrations，不安装、不启用、不注册任何声明。</p>
        </div>
        <el-button :loading="compatLoading" data-testid="compat-refresh" @click="fetchCompatChecks">刷新检查记录</el-button>
      </div>
      <el-form class="download-form" :model="compatForm" label-width="110px">
        <el-form-item label="precheck_id">
          <el-input v-model="compatForm.precheck_id" data-testid="compat-precheck-id" placeholder="上一轮预检通过记录 ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="compatRunning" data-testid="run-compat-check" @click="submitCompatCheck">执行兼容性检查</el-button>
          <span class="muted">can_install 由后端计算；本页不会显示安装按钮。</span>
        </el-form-item>
      </el-form>
      <el-table v-loading="compatLoading" :data="compatItems" data-testid="compat-check-list" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="package_precheck_id" label="precheck" width="100" />
        <el-table-column prop="plugin_code" label="code" min-width="120" />
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="status" label="状态" width="150">
          <template #default="{ row }"><PluginStatusTag :value="row.status" /></template>
        </el-table-column>
        <el-table-column prop="can_install" label="can_install" width="120">
          <template #default="{ row }">{{ row.can_install ? 'true' : 'false' }}</template>
        </el-table-column>
        <el-table-column prop="core_version" label="Core" width="100" />
        <el-table-column prop="compatible_core_version" label="兼容范围" min-width="140" />
        <el-table-column label="blockers / warnings" min-width="260">
          <template #default="{ row }">
            <span class="danger">{{ (row.errors || []).slice(0, 1).join('；') }}</span>
            <span v-if="!(row.errors || []).length" class="muted">{{ (row.warnings || []).slice(0, 1).join('；') || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" data-testid="compat-delete" @click="deleteCompat(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <pre v-if="compatResult" class="compat-result" data-testid="compat-check-result">{{ pretty(compatResult) }}</pre>
    </section>

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

    <section class="filters">
      <el-input v-model="filters.keyword" data-testid="upload-filter-keyword" placeholder="搜索 upload_id / 文件名 / code" clearable @keyup.enter="fetchUploads" />
      <el-select v-model="filters.status" data-testid="upload-filter-status" clearable placeholder="状态">
        <el-option label="全部" value="all" />
        <el-option v-for="s in statuses" :key="s" :label="s" :value="s" />
      </el-select>
      <el-select v-model="filters.risk_level" clearable placeholder="风险">
        <el-option label="全部" value="all" />
        <el-option label="low" value="low" />
        <el-option label="medium" value="medium" />
        <el-option label="high" value="high" />
        <el-option label="blocked" value="blocked" />
      </el-select>
      <el-button data-testid="upload-filter-submit" @click="fetchUploads">筛选</el-button>
    </section>

    <el-table v-loading="loading" :data="items" data-testid="upload-list" border>
      <el-table-column prop="upload_id" label="upload_id" min-width="190" />
      <el-table-column prop="original_filename" label="文件名" min-width="160" />
      <el-table-column prop="package_code" label="code" min-width="120" />
      <el-table-column prop="package_version" label="版本" width="100" />
      <el-table-column prop="status" label="状态" width="160">
        <template #default="{ row }"><PluginStatusTag :value="row.status" /></template>
      </el-table-column>
      <el-table-column prop="risk_level" label="风险" width="110">
        <template #default="{ row }"><PluginRiskTag :level="row.risk_level" /></template>
      </el-table-column>
      <el-table-column prop="checksum_status" label="checksum" width="120" />
      <el-table-column prop="signature_status" label="signature" width="130" />
      <el-table-column prop="trust_status" label="trust" width="110" />
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
        <el-alert class="mb" type="warning" show-icon :closable="false" title="promote 只复制到 storage/plugins/packages，不会安装插件；安装仍走本地包审批与安装流程。" />
        <el-descriptions :column="2" border>
          <el-descriptions-item label="upload_id">{{ detail.record.upload_id }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ detail.record.status }}</el-descriptions-item>
          <el-descriptions-item label="插件">{{ detail.record.package_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ detail.record.package_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="staging">{{ detail.record.staging_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="package">{{ detail.record.package_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="promoted">{{ detail.record.promoted_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="approval">{{ detail.record.approval_id || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="actions" data-testid="upload-actions">
          <el-button :disabled="!actionEnabled('rescan')" :loading="actionLoading === 'rescan'" data-testid="upload-rescan" @click="runAction('rescan')">重新扫描</el-button>
          <el-button :disabled="!actionEnabled('submit_approval')" :loading="actionLoading === 'submit_approval'" data-testid="upload-submit-approval" @click="runAction('submit_approval')">提交导入审批</el-button>
          <el-button :disabled="!actionEnabled('approve')" :loading="actionLoading === 'approve'" data-testid="upload-approve" @click="runAction('approve')">审批通过</el-button>
          <el-button :disabled="!actionEnabled('reject')" :loading="actionLoading === 'reject'" data-testid="upload-reject" @click="runAction('reject')">拒绝</el-button>
          <el-button type="primary" :disabled="!actionEnabled('promote')" :loading="actionLoading === 'promote'" data-testid="upload-promote" @click="runAction('promote')">promote 到本地仓库</el-button>
          <el-button :disabled="!actionEnabled('cancel')" :loading="actionLoading === 'cancel'" data-testid="upload-cancel" @click="runAction('cancel')">取消</el-button>
          <el-button type="danger" :disabled="!actionEnabled('delete')" :loading="actionLoading === 'delete'" data-testid="upload-delete" @click="runAction('delete')">删除</el-button>
        </div>

        <el-alert v-if="disabledReasons" class="mb" type="info" show-icon :closable="false" :title="disabledReasons" data-testid="upload-action-reasons" />

        <div class="detail-grid">
          <section>
            <h3>zip scan</h3>
            <pre data-testid="upload-zip-scan">{{ pretty(detail.zip_scan) }}</pre>
          </section>
          <section>
            <h3>package scan</h3>
            <pre data-testid="upload-file-scan">{{ pretty(detail.file_scan) }}</pre>
          </section>
          <section>
            <h3>checksum / signature</h3>
            <pre data-testid="upload-checksum">{{ pretty({ checksum: detail.checksum, signature: detail.signature }) }}</pre>
          </section>
          <section>
            <h3>risk_report</h3>
            <pre data-testid="upload-risk-report">{{ pretty(detail.risk_report) }}</pre>
          </section>
          <section>
            <h3>manifest validate</h3>
            <pre data-testid="upload-manifest-validation">{{ pretty(detail.manifest_validation) }}</pre>
          </section>
          <section>
            <h3>dry-run</h3>
            <pre data-testid="upload-dry-run">{{ pretty(detail.install_dry_run) }}</pre>
          </section>
        </div>
      </template>
    </el-drawer>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import PluginErrorAlert from './components/PluginErrorAlert.vue';
import PluginPackageBoundaryNotice from './components/PluginPackageBoundaryNotice.vue';
import PluginRiskTag from './components/PluginRiskTag.vue';
import PluginStatusTag from './components/PluginStatusTag.vue';
import {
  approvePluginPackageUpload,
  cancelPluginPackageUpload,
  cleanupPluginPackageUploads,
  deletePluginPackageUpload,
  deletePluginPackageCompatCheck,
  deletePluginPackageStaging,
  downloadPluginPackageToStaging,
  getPluginPackageUpload,
  listPluginPackageCompatChecks,
  listPluginPackageUploads,
  listPluginPackageStaging,
  promotePluginPackageUpload,
  rejectPluginPackageUpload,
  rescanPluginPackageUpload,
  runPluginPackageCompatCheck,
  submitPluginPackageUploadApproval,
  uploadPluginPackageZip,
} from '@/api/plugins';

const statuses = ['uploaded', 'scanned', 'staged', 'blocked', 'approval_pending', 'approval_rejected', 'approved', 'promoted', 'install_approval_pending', 'installed', 'canceled', 'expired', 'deleted', 'failed'];
const filters = reactive({ keyword: '', status: 'all', risk_level: 'all', page: 1, page_size: 20 });
const downloadForm = reactive({ plugin_code: '', version: '', package_url: '', sha256: '' });
const compatForm = reactive({ precheck_id: '' });
const items = ref([]);
const stagingItems = ref([]);
const compatItems = ref([]);
const compatResult = ref(null);
const detail = ref(null);
const drawer = ref(false);
const loading = ref(false);
const uploading = ref(false);
const downloading = ref(false);
const stagingLoading = ref(false);
const compatLoading = ref(false);
const compatRunning = ref(false);
const cleanupLoading = ref(false);
const actionLoading = ref('');
const selectedFile = ref(null);
const uploadRef = ref();
const errorText = ref('');

const actionMap = computed(() => Object.fromEntries((detail.value?.actions || []).map((item) => [item.action, item])));
const disabledReasons = computed(() => (detail.value?.actions || []).filter((item) => !item.enabled && item.reason).map((item) => `${item.reason_code}: ${item.reason}`).join('；'));

onMounted(async () => {
  await fetchUploads();
  await fetchStaging();
  await fetchCompatChecks();
});

function actionEnabled(name) {
  return Boolean(actionMap.value[name]?.enabled);
}

function pretty(value) {
  return JSON.stringify(value || {}, null, 2);
}

function apiError(error) {
  const data = error?.response?.data?.error || error?.response?.data || {};
  const code = data.code || 'unknown_error';
  const message = data.message || error?.message || '操作失败';
  const suggestion = data.suggestion ? ` 建议：${data.suggestion}` : '';
  return `[${code}] ${message}${suggestion}`;
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

async function fetchStaging() {
  stagingLoading.value = true;
  errorText.value = '';
  try {
    const res = await listPluginPackageStaging({ page: 1, page_size: 20 });
    stagingItems.value = res.items || [];
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    stagingLoading.value = false;
  }
}

async function fetchCompatChecks() {
  compatLoading.value = true;
  errorText.value = '';
  try {
    const res = await listPluginPackageCompatChecks({ page: 1, page_size: 20 });
    compatItems.value = res.items || [];
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    compatLoading.value = false;
  }
}

async function submitCompatCheck() {
  compatRunning.value = true;
  errorText.value = '';
  compatResult.value = null;
  try {
    const res = await runPluginPackageCompatCheck(compatForm.precheck_id);
    compatResult.value = res;
    ElMessage.success(`兼容性检查完成：${res.status}`);
    await fetchCompatChecks();
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    compatRunning.value = false;
  }
}

async function deleteCompat(id) {
  if (!id) return;
  compatLoading.value = true;
  errorText.value = '';
  try {
    await deletePluginPackageCompatCheck(id);
    ElMessage.success('兼容性检查记录已删除');
    await fetchCompatChecks();
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    compatLoading.value = false;
  }
}

async function submitDownload() {
  downloading.value = true;
  errorText.value = '';
  try {
    const res = await downloadPluginPackageToStaging({ ...downloadForm });
    ElMessage.success(`下载完成：${res.status}`);
    await fetchStaging();
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    downloading.value = false;
  }
}

async function deleteStaging(id) {
  if (!id) return;
  stagingLoading.value = true;
  errorText.value = '';
  try {
    await deletePluginPackageStaging(id);
    ElMessage.success('staging 文件已删除');
    await fetchStaging();
  } catch (error) {
    errorText.value = apiError(error);
  } finally {
    stagingLoading.value = false;
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
    errorText.value = '[plugin_package_upload_invalid_type] 请先选择 .zip 插件包';
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
    ElMessage.success(`上传完成：${res.status || res.record?.status}`);
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
  actionLoading.value = name;
  errorText.value = '';
  try {
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
