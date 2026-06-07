<template>
  <section class="plugin-page" data-testid="plugin-approvals-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">安装与治理</div>
        <h2>审批中心</h2>
        <p class="muted">插件安装/升级等高风险操作的审批与执行入口。审批通过不等于绕过后端校验，执行时仍会重新预检。</p>
      </div>
      <div class="primary-actions">
        <el-button :loading="loading" data-testid="plugin-approvals-refresh" @click="load(true)">刷新</el-button>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" class="mb" title="边界：不会执行第三方代码，不执行 SQL，不动态加载前端资产。" />

    <div class="filter-panel mb" data-testid="plugin-approvals-filters">
      <div>
        <strong>筛选</strong>
        <div class="muted" style="margin-top: 6px">按状态/动作/插件编码筛选审批记录。</div>
      </div>
      <div class="filter-actions" style="gap: 10px">
        <el-input v-model="qPluginCode" placeholder="插件编码" clearable style="max-width: 220px" />
        <el-select v-model="qAction" placeholder="动作" clearable style="width: 140px">
          <el-option label="安装" value="install" />
          <el-option label="升级" value="upgrade" />
        </el-select>
        <el-select v-model="qStatus" placeholder="状态" clearable style="width: 160px">
          <el-option label="待处理" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
          <el-option label="已取消" value="canceled" />
          <el-option label="已执行" value="executed" />
          <el-option label="失败" value="failed" />
        </el-select>
        <el-button type="primary" :loading="loading" data-testid="plugin-approvals-search" @click="load(true)">查询</el-button>
      </div>
    </div>

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-card shadow="never" data-testid="plugin-approvals-card">
      <template #header>
        <div class="card-head">
          <strong>审批列表</strong>
          <span class="muted">（第 {{ page }} 页 / 共 {{ total }} 条）</span>
        </div>
      </template>

      <el-table :data="items" size="small" stripe data-testid="plugin-approvals-table">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column label="动作" width="100">
          <template #default="{ row }">{{ actionLabel(row.action) }}</template>
        </el-table-column>
        <el-table-column prop="plugin_code" label="插件" min-width="180">
          <template #default="{ row }">
            <div style="display: flex; flex-direction: column">
              <strong>{{ row.plugin_code }}</strong>
              <span class="muted" style="font-size: 12px">{{ row.plugin_name || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" effect="plain">{{ genericStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="risk_level" label="风险" width="120">
          <template #default="{ row }">
            <el-tag :type="riskType(row.risk_level)" effect="plain">{{ packageRiskLabel(row.risk_level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="checksum_status" label="校验" width="120">
          <template #default="{ row }">
            <el-tag :type="checksumType(row.checksum_status)" effect="plain">{{ genericStatusLabel(row.checksum_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="requested_at" label="申请时间" width="170" />
        <el-table-column prop="requested_by_name" label="申请人" width="150" />
        <el-table-column prop="reviewed_at" label="审批时间" width="170" />
        <el-table-column prop="reviewed_by_name" label="审批人" width="150" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <span style="display: inline-flex" data-testid="plugin-approvals-detail-btn" @click="openDetail(row.id)">
              <el-button link type="primary">详情</el-button>
            </span>
          </template>
        </el-table-column>
      </el-table>

      <div class="filter-actions" style="justify-content: flex-end; margin-top: 10px">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" layout="prev, pager, next, sizes, total" :total="total" :page-sizes="[10, 20, 50, 100]" @current-change="load()" @size-change="load(true)" />
      </div>
    </el-card>

    <el-drawer v-model="detailOpen" size="720px" data-testid="plugin-approval-detail-drawer">
      <template #header>
        <div style="display: flex; gap: 10px; align-items: center">
          <strong data-testid="plugin-approval-detail-title">审批详情 #{{ detail?.id || '-' }}</strong>
          <el-tag v-if="detail?.status" :type="statusType(detail.status)" effect="plain">{{ genericStatusLabel(detail.status) }}</el-tag>
          <el-tag v-if="detail?.package_risk_level" :type="riskType(detail.package_risk_level)" effect="plain">风险={{ packageRiskLabel(detail.package_risk_level) }}</el-tag>
        </div>
      </template>

      <el-alert v-if="detailError" type="error" show-icon :closable="false" class="mb" :title="detailError" />

      <el-descriptions v-if="detail" :column="2" border class="mb" data-testid="plugin-approval-detail-info">
        <el-descriptions-item label="动作">{{ actionLabel(detail.action) }}</el-descriptions-item>
        <el-descriptions-item label="插件编码">{{ detail.plugin_code }}</el-descriptions-item>
        <el-descriptions-item label="插件名称">{{ detail.plugin_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="插件包路径">{{ detail.package_path || '-' }}</el-descriptions-item>
        <el-descriptions-item label="当前版本">{{ detail.current_version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="目标版本">{{ detail.target_version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="校验状态">{{ genericStatusLabel(detail.package_checksum_status) }}</el-descriptions-item>
        <el-descriptions-item label="申请时间">{{ detail.requested_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="申请人">{{ detail.requested_by_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审批时间">{{ detail.reviewed_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审批人">{{ detail.reviewed_by_name || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-alert v-if="detail?.reason" type="info" show-icon :closable="false" class="mb" title="申请理由">{{ detail.reason }}</el-alert>
      <el-alert v-if="detail?.review_comment" type="warning" show-icon :closable="false" class="mb" title="审批备注">{{ detail.review_comment }}</el-alert>

      <div class="filter-actions mb" style="justify-content: flex-end; gap: 10px" data-testid="plugin-approval-detail-actions">
        <el-button v-if="detail?.status === 'pending'" :disabled="!canApprove" type="success" plain data-testid="plugin-approval-approve" @click="openReview('approve')">通过</el-button>
        <el-button v-if="detail?.status === 'pending'" :disabled="!canApprove" type="danger" plain data-testid="plugin-approval-reject" @click="openReview('reject')">拒绝</el-button>
        <el-button v-if="detail?.status === 'pending'" :disabled="!canCancel" plain data-testid="plugin-approval-cancel" @click="cancelRequest">撤销</el-button>
        <el-button v-if="detail?.status === 'approved'" :disabled="!canApprove" type="primary" data-testid="plugin-approval-execute" @click="executeRequest">执行</el-button>
      </div>

      <el-collapse v-if="detail" accordion>
        <el-collapse-item title="风险报告" name="risk">
          <pre class="json-box" data-testid="plugin-approval-risk-json">{{ prettyJSON(parseJSON(detail.risk_report_json)) }}</pre>
        </el-collapse-item>
        <el-collapse-item title="预检快照" name="dry">
          <pre class="json-box" data-testid="plugin-approval-dryrun-json">{{ prettyJSON(parseJSON(detail.dry_run_json)) }}</pre>
        </el-collapse-item>
        <el-collapse-item title="执行结果" name="exec">
          <pre class="json-box" data-testid="plugin-approval-exec-json">{{ prettyJSON(parseJSON(detail.execute_result_json)) }}</pre>
        </el-collapse-item>
      </el-collapse>
    </el-drawer>

    <el-dialog v-model="reviewOpen" width="560px" data-testid="plugin-approval-review-dialog">
      <template #header>
        <strong>{{ reviewMode === 'approve' ? '审批通过' : '审批拒绝' }}</strong>
      </template>
      <el-input v-model="reviewComment" type="textarea" :rows="4" placeholder="审批备注（拒绝必填）" data-testid="plugin-approval-review-comment" />
      <template #footer>
        <el-button @click="reviewOpen = false">取消</el-button>
        <el-button v-if="reviewMode === 'approve'" type="success" :loading="reviewLoading" data-testid="plugin-approval-review-confirm" @click="confirmApprove">确认通过</el-button>
        <el-button v-else type="danger" :loading="reviewLoading" data-testid="plugin-approval-review-confirm" @click="confirmReject">确认拒绝</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useAuthStore } from '@/stores/auth';
import { approvePluginApproval, cancelPluginApproval, executePluginApproval, getPluginApproval, listPluginApprovals, rejectPluginApproval } from '@/api/admin';
import { genericStatusLabel, packageRiskLabel } from '@/i18n/formatters';

const auth = useAuthStore();
const canApprove = computed(() => auth.can('plugin.approve'));
const canCancel = computed(() => auth.can('plugin.write'));

const loading = ref(false);
const error = ref('');
const items = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);

const qPluginCode = ref('');
const qAction = ref('');
const qStatus = ref('');

const detailOpen = ref(false);
const detailLoading = ref(false);
const detailError = ref('');
const detail = ref(null);

const reviewOpen = ref(false);
const reviewMode = ref('approve');
const reviewComment = ref('');
const reviewLoading = ref(false);

function statusType(status) {
  if (status === 'executed') return 'success';
  if (status === 'approved') return 'warning';
  if (status === 'pending') return 'info';
  if (status === 'failed') return 'danger';
  if (status === 'rejected') return 'danger';
  return 'default';
}

function riskType(level) {
  if (level === 'blocked') return 'danger';
  if (level === 'high') return 'danger';
  if (level === 'medium') return 'warning';
  return 'success';
}

function checksumType(status) {
  if (status === 'ok') return 'success';
  if (status === 'missing') return 'warning';
  if (status === 'failed') return 'danger';
  return 'info';
}

function actionLabel(action) {
  const labels = {
    install: '安装',
    upgrade: '升级',
    package_import: '导入插件包',
    package_promote: '转入本地仓库',
  };
  return labels[action] || action || '-';
}

function parseJSON(raw) {
  try {
    if (!raw) return {};
    if (typeof raw === 'string') return JSON.parse(raw);
    return raw;
  } catch (e) {
    return { _invalid_json: String(e?.message || e), raw };
  }
}

function prettyJSON(v) {
  try {
    return JSON.stringify(v || {}, null, 2);
  } catch (e) {
    return String(v || '');
  }
}

async function load(reset) {
  if (reset) page.value = 1;
  loading.value = true;
  error.value = '';
  try {
    const res = await listPluginApprovals({
      page: page.value,
      page_size: pageSize.value,
      plugin_code: qPluginCode.value || undefined,
      action: qAction.value || undefined,
      status: qStatus.value || undefined,
    });
    items.value = res?.items || [];
    total.value = res?.pagination?.total ?? 0;
  } catch (e) {
    error.value = String(e?.message || '加载审批列表失败');
  } finally {
    loading.value = false;
  }
}

async function openDetail(id) {
  detailOpen.value = true;
  detailLoading.value = true;
  detailError.value = '';
  detail.value = null;
  try {
    const res = await getPluginApproval(id);
    detail.value = res?.request || null;
  } catch (e) {
    detailError.value = String(e?.message || '加载审批详情失败');
  } finally {
    detailLoading.value = false;
  }
}

function openReview(mode) {
  reviewMode.value = mode;
  reviewComment.value = '';
  reviewOpen.value = true;
}

async function confirmApprove() {
  if (!detail.value) return;
  reviewLoading.value = true;
  try {
    const res = await approvePluginApproval(detail.value.id, { comment: reviewComment.value || '' });
    detail.value = res;
    ElMessage.success('审批通过');
    reviewOpen.value = false;
    await load();
  } catch (e) {
    ElMessage.error(String(e?.message || '审批失败'));
  } finally {
    reviewLoading.value = false;
  }
}

async function confirmReject() {
  if (!detail.value) return;
  reviewLoading.value = true;
  try {
    const res = await rejectPluginApproval(detail.value.id, { comment: reviewComment.value || '' });
    detail.value = res;
    ElMessage.success('已拒绝');
    reviewOpen.value = false;
    await load();
  } catch (e) {
    ElMessage.error(String(e?.message || '拒绝失败'));
  } finally {
    reviewLoading.value = false;
  }
}

async function cancelRequest() {
  if (!detail.value) return;
  await ElMessageBox.confirm('确认撤销该审批申请？', '撤销确认', { type: 'warning' });
  try {
    const res = await cancelPluginApproval(detail.value.id, {});
    detail.value = res;
    ElMessage.success('已撤销');
    await load();
  } catch (e) {
    ElMessage.error(String(e?.message || '撤销失败'));
  }
}

async function executeRequest() {
  if (!detail.value) return;
  await ElMessageBox.confirm('确认执行该审批？执行前会重新预检；不会执行第三方代码、SQL 或前端资产。', '执行确认', { type: 'warning' });
  try {
    const res = await executePluginApproval(detail.value.id);
    detail.value = res;
    ElMessage.success('已执行');
    await load();
  } catch (e) {
    ElMessage.error(String(e?.message || '执行失败'));
  }
}

onMounted(() => {
  load(true);
});
</script>
