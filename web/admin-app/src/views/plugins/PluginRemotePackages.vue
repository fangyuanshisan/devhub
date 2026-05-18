<template>
  <section class="plugin-remote-packages" data-testid="plugin-remote-packages-page">
    <header class="page-head">
      <div>
        <h2>远程插件包</h2>
        <p class="muted">只读治理：下载到受控暂存区，进行 sha256、预检、兼容性检查与 detached signature 验签；不会安装、启用或执行插件代码。</p>
      </div>
      <div class="head-actions">
        <el-button :loading="loading" data-testid="remote-packages-refresh" @click="refreshActive">刷新</el-button>
      </div>
    </header>

    <PluginPackageBoundaryNotice title="安全边界：不下载依赖、不安装插件、不执行第三方代码、不执行 SQL、不加载前端资产。" />

    <el-tabs v-model="activeTab" class="page-tabs" data-testid="remote-packages-tabs" @tab-change="syncQuery">
      <el-tab-pane name="downloads" label="下载记录" />
      <el-tab-pane name="compat" label="兼容性检查" />
      <el-tab-pane name="signatures" label="签名验签" />
    </el-tabs>

    <section v-if="activeTab === 'downloads'" class="card" data-testid="remote-packages-downloads">
      <div class="card-head">
        <div>
          <h3>下载到暂存区</h3>
          <p class="muted">仅允许 https 与 .zip/.tar.gz/.tgz（默认 20MB）；签名文件 URL 仅允许 https + .json（默认 64KB）。</p>
        </div>
        <el-button :loading="downloading" type="primary" data-testid="download-to-staging" @click="submitDownload">下载到暂存区</el-button>
      </div>

      <el-form class="download-form" :model="downloadForm" label-width="110px">
        <el-form-item label="插件编码">
          <el-input v-model="downloadForm.plugin_code" data-testid="download-plugin-code" placeholder="demo_notice" />
        </el-form-item>
        <el-form-item label="版本">
          <el-input v-model="downloadForm.version" data-testid="download-version" placeholder="1.0.0" />
        </el-form-item>
        <el-form-item label="插件包 URL">
          <el-input v-model="downloadForm.package_url" data-testid="download-package-url" placeholder="https://example.com/demo_notice-1.0.0.zip" />
        </el-form-item>
        <el-form-item label="sha256">
          <el-input v-model="downloadForm.sha256" data-testid="download-sha256" placeholder="建议填写；缺失会标记缺少 sha256" />
        </el-form-item>
        <el-form-item label="签名文件 URL">
          <el-input
            v-model="downloadForm.signature_url"
            data-testid="download-signature-url"
            placeholder="可选：https://example.com/demo_notice-1.0.0.devhub-signature.json"
          />
        </el-form-item>
      </el-form>

      <el-tabs v-model="downloadStatus" class="status-tabs" data-testid="remote-packages-download-status" @tab-change="syncQuery">
        <el-tab-pane name="all" label="全部" />
        <el-tab-pane name="downloaded" label="已下载" />
        <el-tab-pane name="checksum_failed" label="校验失败" />
        <el-tab-pane name="checksum_missing" label="缺少 sha256" />
        <el-tab-pane name="failed" label="失败" />
      </el-tabs>

      <el-table v-loading="loading" :data="stagingItems" data-testid="staging-list" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="plugin_code" label="插件编码" min-width="130" />
        <el-table-column prop="version" label="版本" width="110" />
        <el-table-column prop="status" label="状态" width="160">
          <template #default="{ row }"><PluginStatusTag :value="row.status" /></template>
        </el-table-column>
        <el-table-column prop="file_size" label="大小" width="110" />
        <el-table-column prop="sha256_actual" label="sha256" min-width="220" show-overflow-tooltip />
        <el-table-column prop="staging_path" label="暂存路径" min-width="240" show-overflow-tooltip />
        <el-table-column prop="error_message" label="错误" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" data-testid="staging-delete" @click="deleteStaging(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <PluginEmptyState v-if="!loading && !stagingItems.length" testid="staging-empty-state" description="暂无暂存区下载记录" />
    </section>

    <section v-else-if="activeTab === 'compat'" class="card" data-testid="remote-packages-compat">
      <div class="card-head">
        <div>
          <h3>依赖 / 兼容性检查</h3>
          <p class="muted">只接受预检通过的记录；只做检查，不安装、不启用、不注册任何声明。</p>
        </div>
        <el-button :loading="compatLoading" data-testid="compat-refresh" @click="fetchCompatChecks">刷新</el-button>
      </div>

      <el-form class="download-form" :model="compatForm" label-width="110px">
        <el-form-item label="预检 ID">
          <el-input v-model="compatForm.precheck_id" data-testid="compat-precheck-id" placeholder="预检通过记录 ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="compatRunning" data-testid="run-compat-check" @click="submitCompatCheck">执行兼容性检查</el-button>
          <span class="muted">检查结果会落库，后续安装/升级必须依赖“允许安装”为是。</span>
        </el-form-item>
      </el-form>

      <el-table v-loading="compatLoading" :data="compatItems" data-testid="compat-list" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="plugin_code" label="插件编码" min-width="130" />
        <el-table-column prop="version" label="版本" width="110" />
        <el-table-column prop="status" label="状态" width="170">
          <template #default="{ row }">{{ genericStatusLabel(row.status) }}</template>
        </el-table-column>
        <el-table-column prop="can_install" label="允许安装" width="110">
          <template #default="{ row }">{{ row.can_install ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column prop="core_version" label="当前 Core" width="120" />
        <el-table-column prop="compatible_core_version" label="兼容 Core" min-width="170" show-overflow-tooltip />
        <el-table-column prop="finished_at" label="完成时间" width="170" />
      </el-table>

      <pre v-if="compatResult" class="json-box" data-testid="compat-result">{{ pretty(compatResult) }}</pre>
    </section>

    <section v-else class="card" data-testid="remote-packages-signatures">
      <div class="card-head">
        <div>
          <h3>detached signature 验签</h3>
          <p class="muted">对预检通过的解压目录执行 devhub-signature.json 验签；验签通过才允许进入默认安装/升级链路。</p>
        </div>
        <el-button :loading="signatureLoading" data-testid="signature-refresh" @click="fetchSignatures">刷新验签记录</el-button>
      </div>

      <el-form class="download-form" :model="signatureForm" label-width="110px">
        <el-form-item label="预检 ID">
          <el-input v-model="signatureForm.precheck_id" data-testid="signature-precheck-id" placeholder="预检通过记录 ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="signatureRunning" data-testid="run-signature-verify" @click="submitVerifySignature">执行验签</el-button>
          <span class="muted">不会安装、不会启用、不会执行插件代码；签名文件来自签名文件 URL 或包内 devhub-signature.json。</span>
        </el-form-item>
      </el-form>

      <el-table v-loading="signatureLoading" :data="signatureItems" data-testid="signature-list" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="plugin_code" label="插件编码" min-width="130" />
        <el-table-column prop="version" label="版本" width="110" />
        <el-table-column prop="status" label="状态" width="190">
          <template #default="{ row }">{{ genericStatusLabel(row.status) }}</template>
        </el-table-column>
        <el-table-column prop="publisher_id" label="发布者 ID" min-width="150" show-overflow-tooltip />
        <el-table-column prop="key_id" label="key_id" min-width="160" show-overflow-tooltip />
        <el-table-column prop="verified_at" label="验签时间" width="170" />
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" data-testid="signature-delete" @click="deleteSignature(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <pre v-if="signatureResult" class="json-box" data-testid="signature-result">{{ pretty(signatureResult) }}</pre>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import {
  deletePluginPackageSignature,
  deletePluginPackageStaging,
  downloadPluginPackageToStaging,
  listPluginPackageCompatChecks,
  listPluginPackageSignatures,
  listPluginPackageStaging,
  runPluginPackageCompatCheck,
  verifyPluginPackageSignature,
} from '@/api/admin';
import PluginPackageBoundaryNotice from './components/PluginPackageBoundaryNotice.vue';
import PluginEmptyState from './components/PluginEmptyState.vue';
import PluginStatusTag from './components/PluginStatusTag.vue';
import { genericStatusLabel } from '@/i18n/formatters';

const route = useRoute();
const router = useRouter();

const activeTab = ref(route.query.tab || 'downloads');
const downloadStatus = ref(route.query.status || 'all');

const downloadForm = reactive({ plugin_code: '', version: '', package_url: '', sha256: '', signature_url: '' });
const compatForm = reactive({ precheck_id: '' });
const signatureForm = reactive({ precheck_id: '' });

const loading = ref(false);
const downloading = ref(false);
const stagingItems = ref([]);

const compatLoading = ref(false);
const compatRunning = ref(false);
const compatItems = ref([]);
const compatResult = ref(null);

const signatureLoading = ref(false);
const signatureRunning = ref(false);
const signatureItems = ref([]);
const signatureResult = ref(null);

const stagingParams = computed(() => (downloadStatus.value && downloadStatus.value !== 'all' ? { status: downloadStatus.value } : {}));

watch(() => route.query.tab, (value) => {
  if (value && value !== activeTab.value) activeTab.value = value;
});
watch(() => route.query.status, (value) => {
  if (value && value !== downloadStatus.value) downloadStatus.value = value;
});

function syncQuery() {
  const query = { ...route.query, tab: activeTab.value };
  if (activeTab.value === 'downloads') query.status = downloadStatus.value;
  else delete query.status;
  router.replace({ query });
  refreshActive();
}

function pretty(obj) {
  try {
    return JSON.stringify(obj, null, 2);
  } catch (e) {
    return String(obj || '');
  }
}

async function refreshActive() {
  if (activeTab.value === 'downloads') return fetchStaging();
  if (activeTab.value === 'compat') return fetchCompatChecks();
  return fetchSignatures();
}

async function fetchStaging() {
  loading.value = true;
  try {
    const res = await listPluginPackageStaging(stagingParams.value);
    stagingItems.value = res.items || [];
  } catch (error) {
    ElMessage.error(error?.message || '暂存区记录获取失败');
  } finally {
    loading.value = false;
  }
}

async function submitDownload() {
  downloading.value = true;
  try {
    await downloadPluginPackageToStaging({ ...downloadForm });
    ElMessage.success('已提交下载');
    await fetchStaging();
  } catch (error) {
    ElMessage.error(error?.message || '下载失败');
  } finally {
    downloading.value = false;
  }
}

async function deleteStaging(id) {
  try {
    await deletePluginPackageStaging(id);
    ElMessage.success('已删除');
    await fetchStaging();
  } catch (error) {
    ElMessage.error(error?.message || '删除失败');
  }
}

async function fetchCompatChecks() {
  compatLoading.value = true;
  try {
    const res = await listPluginPackageCompatChecks({});
    compatItems.value = res.items || [];
  } catch (error) {
    ElMessage.error(error?.message || '兼容性检查列表获取失败');
  } finally {
    compatLoading.value = false;
  }
}

async function submitCompatCheck() {
  compatRunning.value = true;
  compatResult.value = null;
  try {
    const res = await runPluginPackageCompatCheck(compatForm.precheck_id);
    compatResult.value = res;
    ElMessage.success('兼容性检查已执行');
    await fetchCompatChecks();
  } catch (error) {
    ElMessage.error(error?.message || '兼容性检查失败');
  } finally {
    compatRunning.value = false;
  }
}

async function fetchSignatures() {
  signatureLoading.value = true;
  try {
    const res = await listPluginPackageSignatures({});
    signatureItems.value = res.items || [];
  } catch (error) {
    ElMessage.error(error?.message || '验签记录获取失败');
  } finally {
    signatureLoading.value = false;
  }
}

async function submitVerifySignature() {
  signatureRunning.value = true;
  signatureResult.value = null;
  try {
    const res = await verifyPluginPackageSignature(signatureForm.precheck_id);
    signatureResult.value = res;
    ElMessage.success('验签已执行');
    await fetchSignatures();
  } catch (error) {
    ElMessage.error(error?.message || '验签失败');
  } finally {
    signatureRunning.value = false;
  }
}

async function deleteSignature(id) {
  try {
    await deletePluginPackageSignature(id);
    ElMessage.success('已删除');
    await fetchSignatures();
  } catch (error) {
    ElMessage.error(error?.message || '删除失败');
  }
}

onMounted(() => {
  refreshActive();
});
</script>

<style scoped>
.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}
.muted {
  color: #6b7280;
}
.card {
  margin-top: 10px;
  padding: 14px;
  border: 1px solid #edf0f5;
  border-radius: 12px;
  background: #fff;
}
.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}
.download-form {
  margin-bottom: 10px;
}
.json-box {
  margin-top: 10px;
  padding: 10px;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #0b1020;
  color: #e5e7eb;
  font-size: 12px;
  line-height: 1.45;
  white-space: pre-wrap;
  word-break: break-word;
}
.status-tabs {
  margin-bottom: 10px;
}
</style>
