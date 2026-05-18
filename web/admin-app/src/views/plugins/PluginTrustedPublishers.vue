<template>
  <section class="plugin-page trusted-page" data-testid="plugin-trusted-publishers-page">
    <div class="page-hero">
      <div>
        <p class="eyebrow">插件包安全</p>
        <h1>可信发布者</h1>
        <p class="muted">维护本地可信发布者公钥。包内 publisher.json 只能声明身份，是否可信以这里的公钥记录为准。</p>
      </div>
      <div class="hero-actions">
        <el-button @click="load" data-testid="trusted-publisher-refresh">刷新</el-button>
        <el-button type="primary" @click="openCreate" data-testid="trusted-publisher-create">新增可信发布者</el-button>
      </div>
    </div>

    <el-alert
      class="boundary-alert"
      type="warning"
      show-icon
      :closable="false"
      title="当前只支持本地可信来源管理，不支持远程可信源同步、证书链、远程市场、自动下载或动态加载。"
    />

    <el-card shadow="never" class="toolbar-card">
      <el-form :inline="true" class="toolbar-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" clearable placeholder="发布者 ID / 名称 / key_id" data-testid="trusted-publisher-keyword" @keyup.enter="load" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width: 150px" data-testid="trusted-publisher-status-filter">
            <el-option label="可信" value="trusted" />
            <el-option label="已禁用" value="blocked" />
            <el-option label="已吊销" value="revoked" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="load">筛选</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>发布者列表</span>
          <div class="summary">
            <el-tag type="success">可信 {{ summary.trusted || 0 }}</el-tag>
            <el-tag type="danger">已禁用 {{ summary.blocked || 0 }}</el-tag>
            <el-tag type="warning">已吊销 {{ summary.revoked || 0 }}</el-tag>
          </div>
        </div>
      </template>
      <el-table v-loading="loading" :data="items" border data-testid="trusted-publisher-list" empty-text="暂无可信发布者">
        <el-table-column label="发布者" min-width="240">
          <template #default="{ row }">
            <div class="publisher-main">
              <strong>{{ row.name || row.publisher_id }}</strong>
              <span class="mono muted">{{ row.publisher_id }}</span>
              <span class="mono muted">key_id：{{ row.public_key_id || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="可信级别" width="120">
          <template #default="{ row }">{{ trustLevelLabel(row.trust_level || row.status) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" data-testid="trusted-publisher-status">{{ trustLevelLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="影响范围" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <div>允许：{{ codeListText(row.allowed_plugin_codes, '全部插件') }}</div>
            <div class="muted">阻断：{{ codeListText(row.blocked_plugin_codes, '无') }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="expires_at" label="过期时间" width="170" />
        <el-table-column prop="last_used_at" label="最近使用" width="170" />
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)" data-testid="trusted-publisher-detail">详情</el-button>
            <el-button link type="primary" @click="openEdit(row)" data-testid="trusted-publisher-edit">编辑</el-button>
            <el-dropdown trigger="click" @command="(action) => changeStatus(row, action)">
              <el-button link type="info">更多</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="row.status !== 'blocked'" command="block" data-testid="trusted-publisher-block">禁用</el-dropdown-item>
                  <el-dropdown-item v-if="row.status !== 'revoked'" command="revoke" data-testid="trusted-publisher-revoke">吊销</el-dropdown-item>
                  <el-dropdown-item v-if="row.status !== 'trusted'" command="restore" data-testid="trusted-publisher-restore">恢复</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑可信发布者' : '新增可信发布者'" width="720px" data-testid="trusted-publisher-dialog">
      <el-form :model="form" label-width="150px" data-testid="trusted-publisher-form">
        <el-form-item label="发布者 ID" required>
          <el-input v-model="form.publisher_id" data-testid="trusted-publisher-publisher-id" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" data-testid="trusted-publisher-name" />
        </el-form-item>
        <el-form-item label="主页">
          <el-input v-model="form.homepage" data-testid="trusted-publisher-homepage" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="key_id" required>
          <el-input v-model="form.public_key_id" data-testid="trusted-publisher-key-id" />
        </el-form-item>
        <el-form-item label="算法" required>
          <el-select v-model="form.public_key_algorithm" style="width: 220px">
            <el-option label="ed25519" value="ed25519" />
          </el-select>
        </el-form-item>
        <el-form-item label="公钥" required>
          <el-input v-model="form.public_key" type="textarea" :rows="3" placeholder="Base64 编码的 Ed25519 公钥" data-testid="trusted-publisher-public-key" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" type="textarea" :rows="2" data-testid="trusted-publisher-notes" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit" data-testid="trusted-publisher-submit">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" size="48%" title="可信发布者详情" data-testid="trusted-publisher-detail-drawer">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="发布者 ID">{{ detail.publisher_id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="key_id">{{ detail.public_key_id }}</el-descriptions-item>
        <el-descriptions-item label="可信级别">{{ trustLevelLabel(detail.trust_level || detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="算法">{{ detail.public_key_algorithm }}</el-descriptions-item>
        <el-descriptions-item label="指纹">{{ detail.fingerprint }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType(detail.status)">{{ trustLevelLabel(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="公钥">
          <code class="public-key">{{ detail.public_key }}</code>
        </el-descriptions-item>
        <el-descriptions-item label="允许插件">{{ codeListText(detail.allowed_plugin_codes, '全部') }}</el-descriptions-item>
        <el-descriptions-item label="阻断插件">{{ codeListText(detail.blocked_plugin_codes, '无') }}</el-descriptions-item>
        <el-descriptions-item label="过期时间">{{ detail.expires_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="最近使用">{{ detail.last_used_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="影响分析">会影响使用该 publisher_id / key_id 的插件包验签与安装准入。</el-descriptions-item>
        <el-descriptions-item label="备注">{{ detail.notes || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  blockTrustedPublisher,
  createTrustedPublisher,
  getTrustedPublisher,
  listTrustedPublishers,
  restoreTrustedPublisher,
  revokeTrustedPublisher,
  updateTrustedPublisher,
} from '@/api/admin';
import { trustLevelLabel } from '@/i18n/formatters';

const loading = ref(false);
const saving = ref(false);
const items = ref([]);
const summary = ref({});
const dialogVisible = ref(false);
const detailVisible = ref(false);
const detail = ref(null);
const editingId = ref(null);
const filters = reactive({ keyword: '', status: '' });
const form = reactive(emptyForm());

function emptyForm() {
  return {
    publisher_id: '',
    name: '',
    homepage: '',
    email: '',
    public_key_id: '',
    public_key_algorithm: 'ed25519',
    public_key: '',
    notes: '',
  };
}

function applyForm(record = {}) {
  Object.assign(form, emptyForm(), record);
}

async function load() {
  loading.value = true;
  try {
    const data = await listTrustedPublishers({
      keyword: filters.keyword || undefined,
      status: filters.status || undefined,
      page_size: 100,
    });
    items.value = data?.items || [];
    summary.value = data?.summary || {};
  } finally {
    loading.value = false;
  }
}

function resetFilters() {
  filters.keyword = '';
  filters.status = '';
  load();
}

function openCreate() {
  editingId.value = null;
  applyForm();
  dialogVisible.value = true;
}

function openEdit(row) {
  editingId.value = row.id;
  applyForm(row);
  dialogVisible.value = true;
}

async function openDetail(row) {
  const data = await getTrustedPublisher(row.id);
  detail.value = data;
  detailVisible.value = true;
}

async function submit() {
  saving.value = true;
  try {
    if (editingId.value) await updateTrustedPublisher(editingId.value, form);
    else await createTrustedPublisher(form);
    ElMessage.success('已保存可信发布者');
    dialogVisible.value = false;
    await load();
  } catch (err) {
    ElMessage.error(err?.response?.data?.message || err?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function changeStatus(row, action) {
  const labels = { block: '禁用', revoke: '吊销', restore: '恢复' };
  await ElMessageBox.confirm(`确认 ${labels[action]} ${row.publisher_id} / ${row.public_key_id}？`, '确认操作', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消',
  });
  const payload = { comment: `${labels[action]} from admin UI` };
  if (action === 'block') await blockTrustedPublisher(row.id, payload);
  if (action === 'revoke') await revokeTrustedPublisher(row.id, payload);
  if (action === 'restore') await restoreTrustedPublisher(row.id, payload);
  ElMessage.success('状态已更新');
  await load();
}

function statusType(status) {
  const value = String(status || '').toLowerCase();
  if (value === 'trusted') return 'success';
  if (value === 'blocked') return 'danger';
  if (value === 'revoked') return 'warning';
  return 'info';
}

function codeListText(value, emptyText = '-') {
  if (Array.isArray(value)) return value.length ? value.join(', ') : emptyText;
  if (typeof value === 'string' && value.trim()) return value;
  return emptyText;
}

onMounted(load);
</script>

<style scoped>
.plugin-page {
  padding: 18px;
}
.page-hero,
.toolbar-card,
.table-card {
  border-radius: 14px;
}
.page-hero {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  padding: 18px 20px;
  background: linear-gradient(135deg, #f8fbff, #eef5ff);
  border: 1px solid #dbe7f7;
  margin-bottom: 14px;
}
.eyebrow {
  margin: 0 0 6px;
  color: #409eff;
  font-weight: 700;
}
h1 {
  margin: 0 0 8px;
  font-size: 26px;
}
.muted {
  margin: 0;
  color: #64748b;
}
.hero-actions,
.summary {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.boundary-alert,
.toolbar-card {
  margin-bottom: 14px;
}
.toolbar-form {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.publisher-main {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.public-key {
  word-break: break-all;
  white-space: pre-wrap;
}
</style>
