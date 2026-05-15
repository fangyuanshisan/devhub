<template>
  <section class="plugin-remote-index-page" data-testid="plugin-remote-indexes-page">
    <header class="page-head">
      <div>
        <p class="eyebrow">插件包分发前置能力</p>
        <h2>远程插件索引</h2>
        <p>只读取远程 index.json 元数据，展示远程插件、版本、checksum、publisher 与风险提示。</p>
      </div>
      <div class="head-actions">
        <el-button data-testid="remote-index-refresh" :loading="loadingSources" @click="loadSources">刷新</el-button>
        <el-button type="primary" data-testid="remote-index-create" @click="openCreate">新增索引源</el-button>
      </div>
    </header>

    <el-alert
      class="mb"
      type="warning"
      show-icon
      :closable="false"
      title="当前是只读镜像：不下载插件包、不安装插件、不自动更新、不执行代码、不动态加载前端资产，也不会自动信任远程 publisher。package_sha256 只是远程元数据声明。"
    />

    <el-alert v-if="errorText" class="mb" type="error" show-icon :closable="false" :title="errorText" data-testid="remote-index-error" />

    <el-card shadow="never" class="mb">
      <el-form :inline="true" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="sourceFilters.keyword" data-testid="remote-index-keyword" clearable placeholder="source_id / name / URL" @keyup.enter="loadSources" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="sourceFilters.status" clearable placeholder="全部" style="width: 140px">
            <el-option label="enabled" value="enabled" />
            <el-option label="disabled" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadSources">筛选</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="mb">
      <template #header>
        <div class="card-header">
          <span>索引源列表</span>
          <span class="muted">enabled {{ summary.enabled || 0 }} / disabled {{ summary.disabled || 0 }} / failed {{ summary.failed || 0 }}</span>
        </div>
      </template>
      <el-table v-loading="loadingSources" :data="sources" border data-testid="remote-index-list" empty-text="暂无远程索引源">
        <el-table-column prop="source_id" label="source_id" min-width="170" />
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="index_url" label="index_url" min-width="280">
          <template #default="{ row }"><code>{{ row.index_url }}</code></template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }"><el-tag :type="row.status === 'enabled' ? 'success' : 'info'">{{ row.status }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="last_fetch_status" label="fetch" width="120" />
        <el-table-column prop="last_fetch_at" label="last_fetch_at" width="170" />
        <el-table-column label="操作" width="330" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" data-testid="remote-index-fetch" @click="fetchIndex(row)">拉取</el-button>
            <el-button link type="primary" @click="selectSource(row)">插件列表</el-button>
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status !== 'enabled'" link type="success" @click="enableSource(row)">启用</el-button>
            <el-button v-else link type="warning" data-testid="remote-index-disable" @click="disableSource(row)">禁用</el-button>
            <el-button link type="danger" @click="removeSource(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-if="selectedSource" shadow="never">
      <template #header>
        <div class="card-header">
          <span>远程插件列表：{{ selectedSource.name }}</span>
          <el-input v-model="pluginKeyword" class="plugin-search" data-testid="remote-plugin-keyword" clearable placeholder="搜索插件 code / name" @keyup.enter="loadPlugins" />
        </div>
      </template>
      <el-table v-loading="loadingPlugins" :data="plugins" border data-testid="remote-plugin-list" empty-text="暂无远程插件，请先拉取索引">
        <el-table-column prop="code" label="code" min-width="140" />
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="latest_version" label="latest" width="110" />
        <el-table-column prop="publisher_id" label="publisher" min-width="150" />
        <el-table-column label="trust" width="120">
          <template #default="{ row }"><el-tag :type="trustType(row.publisher_trust_status)">{{ row.publisher_trust_status }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="version_status" label="本地状态" width="150" />
        <el-table-column label="Core" width="130">
          <template #default="{ row }"><el-tag :type="row.core_compatibility?.status === 'incompatible' ? 'danger' : 'success'">{{ row.core_compatibility?.status }}</el-tag></template>
        </el-table-column>
        <el-table-column label="风险" min-width="220">
          <template #default="{ row }">
            <el-tag :type="riskType(row.risk_level)">{{ row.risk_level }}</el-tag>
            <span class="muted risk-text">{{ row.risk_summary }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" data-testid="remote-plugin-detail" @click="openPluginDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑索引源' : '新增索引源'" width="720px" data-testid="remote-index-dialog">
      <el-form :model="form" label-width="130px">
        <el-form-item label="source_id" required>
          <el-input v-model="form.source_id" data-testid="remote-index-source-id" placeholder="devhub-official-index" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" data-testid="remote-index-name" />
        </el-form-item>
        <el-form-item label="index_url" required>
          <el-input v-model="form.index_url" data-testid="remote-index-url" placeholder="https://example.com/plugins/index.json" />
        </el-form-item>
        <el-form-item label="homepage">
          <el-input v-model="form.homepage" />
        </el-form-item>
        <el-form-item label="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 180px">
            <el-option label="enabled" value="enabled" />
            <el-option label="disabled" value="disabled" />
          </el-select>
        </el-form-item>
        <el-alert type="info" show-icon :closable="false" title="URL 只允许 http/https；生产建议 HTTPS；localhost、内网、file:// 会被服务端拦截以防 SSRF。" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" data-testid="remote-index-submit" @click="submitSource">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" size="760px" title="远程插件详情" data-testid="remote-plugin-detail-drawer">
      <template v-if="pluginDetail">
        <el-alert class="mb" type="warning" show-icon :closable="false" title="只读展示：本轮不会下载 package_url，不会安装远程插件，不会自动信任 publisher。" />
        <el-descriptions :column="2" border>
          <el-descriptions-item label="code">{{ pluginDetail.plugin.code }}</el-descriptions-item>
          <el-descriptions-item label="name">{{ pluginDetail.plugin.name }}</el-descriptions-item>
          <el-descriptions-item label="latest">{{ pluginDetail.plugin.latest_version }}</el-descriptions-item>
          <el-descriptions-item label="installed">{{ pluginDetail.installed ? pluginDetail.local_version : 'not_installed' }}</el-descriptions-item>
          <el-descriptions-item label="description" :span="2">{{ pluginDetail.plugin.description }}</el-descriptions-item>
        </el-descriptions>
        <h3>版本元数据</h3>
        <el-table :data="pluginDetail.versions || []" border>
          <el-table-column prop="version" label="version" width="110" />
          <el-table-column label="package_url" min-width="260">
            <template #default="{ row }"><code>{{ row.package_url }}</code></template>
          </el-table-column>
          <el-table-column prop="package_sha256" label="package_sha256" min-width="190" />
          <el-table-column prop="publisher_id" label="publisher" min-width="140" />
          <el-table-column label="trust" width="110">
            <template #default="{ row }"><el-tag :type="trustType(row.publisher_trust_status)">{{ row.publisher_trust_status }}</el-tag></template>
          </el-table-column>
          <el-table-column label="verification" width="130">
            <template #default>{{ 'metadata_only' }}</template>
          </el-table-column>
          <el-table-column label="风险" min-width="180">
            <template #default="{ row }">
              <el-tag :type="riskType(row.risk_level)">{{ row.risk_level }}</el-tag>
              <div v-for="item in row.risk_items || []" :key="item.code" class="risk-item">{{ item.code }}：{{ item.message }}</div>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </el-drawer>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  createPluginRemoteIndex,
  deletePluginRemoteIndex,
  disablePluginRemoteIndex,
  enablePluginRemoteIndex,
  fetchPluginRemoteIndex,
  getPluginRemoteIndexPlugin,
  listPluginRemoteIndexes,
  listPluginRemoteIndexPlugins,
  updatePluginRemoteIndex,
} from '@/api/admin';

const loadingSources = ref(false);
const loadingPlugins = ref(false);
const saving = ref(false);
const sources = ref([]);
const plugins = ref([]);
const summary = ref({});
const errorText = ref('');
const selectedSource = ref(null);
const pluginKeyword = ref('');
const dialogVisible = ref(false);
const editingId = ref(0);
const detailVisible = ref(false);
const pluginDetail = ref(null);
const sourceFilters = reactive({ keyword: '', status: '' });
const form = reactive({ source_id: '', name: '', index_url: '', homepage: '', description: '', status: 'enabled', trust_policy: 'readonly' });

const errorMessage = (err, fallback) => {
  const data = err?.response?.data || {};
  return [data.code, data.message || data.error || fallback, data.suggestion].filter(Boolean).join(' ｜ ');
};

const loadSources = async () => {
  loadingSources.value = true;
  errorText.value = '';
  try {
    const res = await listPluginRemoteIndexes({ ...sourceFilters });
    sources.value = res.items || [];
    summary.value = res.summary || {};
    if (!selectedSource.value && sources.value.length) selectedSource.value = sources.value[0];
  } catch (err) {
    errorText.value = errorMessage(err, '加载远程索引失败');
  } finally {
    loadingSources.value = false;
  }
};

const loadPlugins = async () => {
  if (!selectedSource.value?.id) return;
  loadingPlugins.value = true;
  errorText.value = '';
  try {
    const res = await listPluginRemoteIndexPlugins(selectedSource.value.id, { keyword: pluginKeyword.value });
    plugins.value = res.items || [];
  } catch (err) {
    errorText.value = errorMessage(err, '加载远程插件失败');
    plugins.value = [];
  } finally {
    loadingPlugins.value = false;
  }
};

const resetFilters = () => {
  sourceFilters.keyword = '';
  sourceFilters.status = '';
  loadSources();
};

const resetForm = () => {
  Object.assign(form, { source_id: '', name: '', index_url: '', homepage: '', description: '', status: 'enabled', trust_policy: 'readonly' });
};

const openCreate = () => {
  editingId.value = 0;
  resetForm();
  dialogVisible.value = true;
};

const openEdit = (row) => {
  editingId.value = row.id;
  Object.assign(form, { source_id: row.source_id, name: row.name, index_url: row.index_url, homepage: row.homepage, description: row.description, status: row.status, trust_policy: row.trust_policy || 'readonly' });
  dialogVisible.value = true;
};

const submitSource = async () => {
  saving.value = true;
  try {
    if (editingId.value) await updatePluginRemoteIndex(editingId.value, form);
    else await createPluginRemoteIndex(form);
    ElMessage.success('已保存远程索引源');
    dialogVisible.value = false;
    await loadSources();
  } catch (err) {
    ElMessage.error(errorMessage(err, '保存远程索引失败'));
  } finally {
    saving.value = false;
  }
};

const fetchIndex = async (row) => {
  try {
    await fetchPluginRemoteIndex(row.id);
    ElMessage.success('索引拉取完成');
    selectedSource.value = row;
    await loadSources();
    await loadPlugins();
  } catch (err) {
    ElMessage.error(errorMessage(err, '拉取远程索引失败'));
    await loadSources();
  }
};

const selectSource = async (row) => {
  selectedSource.value = row;
  await loadPlugins();
};

const enableSource = async (row) => {
  await enablePluginRemoteIndex(row.id);
  await loadSources();
};

const disableSource = async (row) => {
  await disablePluginRemoteIndex(row.id);
  await loadSources();
};

const removeSource = async (row) => {
  await ElMessageBox.confirm(`删除远程索引源 ${row.source_id}？不会删除任何插件包。`, '确认删除', { type: 'warning' });
  await deletePluginRemoteIndex(row.id);
  ElMessage.success('已删除');
  if (selectedSource.value?.id === row.id) {
    selectedSource.value = null;
    plugins.value = [];
  }
  await loadSources();
};

const openPluginDetail = async (row) => {
  pluginDetail.value = await getPluginRemoteIndexPlugin(selectedSource.value.id, row.code);
  detailVisible.value = true;
};

const trustType = (status) => {
  if (status === 'trusted') return 'success';
  if (status === 'blocked' || status === 'revoked') return 'danger';
  return 'warning';
};

const riskType = (risk) => {
  if (risk === 'blocked' || risk === 'high') return 'danger';
  if (risk === 'warning' || risk === 'medium') return 'warning';
  return 'success';
};

onMounted(async () => {
  await loadSources();
  await loadPlugins();
});
</script>

<style scoped>
.plugin-remote-index-page {
  padding: 20px;
}
.page-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 16px;
}
.page-head h2 {
  margin: 4px 0 8px;
  font-size: 26px;
}
.eyebrow,
.muted {
  color: #64748b;
  margin: 0;
}
.head-actions,
.card-header {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: space-between;
}
.mb {
  margin-bottom: 16px;
}
.plugin-search {
  width: 260px;
}
code {
  color: #334155;
  font-size: 12px;
  word-break: break-all;
}
.risk-text {
  margin-left: 8px;
}
.risk-item {
  font-size: 12px;
  color: #64748b;
  margin-top: 4px;
}
</style>
