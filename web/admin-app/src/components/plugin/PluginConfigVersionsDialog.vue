<template>
  <el-dialog v-model="visible" width="980px" destroy-on-close :title="title" data-testid="plugin-config-versions-dialog">
    <el-alert type="info" show-icon :closable="false" class="mb" title="提示：历史版本与差异会对敏感字段脱敏；回滚预览不会修改当前配置。" />

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-card shadow="never" class="mb" data-testid="plugin-config-versions-list">
      <template #header>
        <div class="card-head">
          <strong>版本历史</strong>
              <span class="muted">（共 {{ total }} 条）</span>
        </div>
      </template>
      <el-table :data="items" border size="small" v-loading="loading" data-testid="plugin-config-versions-table">
        <el-table-column prop="version_no" label="版本号" width="110" />
        <el-table-column prop="scope" label="范围" width="110">
          <template #default="{ row }">{{ scopeLabel(row.scope) }}</template>
        </el-table-column>
        <el-table-column prop="community_id" label="子站" width="120" />
        <el-table-column prop="source" label="来源" width="130">
          <template #default="{ row }">{{ sourceLabel(row.source) }}</template>
        </el-table-column>
        <el-table-column prop="operator_name" label="操作人" width="160" />
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="变更字段" min-width="220">
          <template #default="{ row }">
            <div class="tag-wrap">
              <el-tag v-for="k in row.changed_keys || []" :key="k" effect="plain">{{ k }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" data-testid="plugin-config-version-detail-btn" @click="openDetail(row)">详情</el-button>
            <el-button link type="warning" data-testid="plugin-config-version-rollback-preview-btn" @click="openRollbackPreview(row)">回滚预览</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="filter-actions" style="justify-content: flex-end; margin-top: 10px">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          layout="prev, pager, next, sizes, total"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          @current-change="fetchList"
          @size-change="() => fetchList(true)"
        />
      </div>
    </el-card>

    <el-drawer v-model="detailVisible" size="70%" destroy-on-close title="版本详情" data-testid="plugin-config-version-detail-drawer">
      <section class="action-panel in-drawer">
        <el-alert v-if="detailError" type="error" show-icon :closable="false" class="mb" :title="detailError" />
        <div v-if="detail" data-testid="plugin-config-version-detail">
          <el-descriptions :column="2" border class="mb">
            <el-descriptions-item label="版本号">{{ detail.version?.version_no }}</el-descriptions-item>
            <el-descriptions-item label="范围">{{ scopeLabel(detail.version?.scope) }}</el-descriptions-item>
            <el-descriptions-item label="子站 ID">{{ detail.version?.community_id || 0 }}</el-descriptions-item>
            <el-descriptions-item label="来源">{{ sourceLabel(detail.version?.source) }}</el-descriptions-item>
            <el-descriptions-item label="操作人">{{ detail.version?.operator_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ detail.version?.created_at || '-' }}</el-descriptions-item>
          </el-descriptions>

          <el-collapse>
            <el-collapse-item title="config_json（脱敏）" name="config">
              <pre class="json-box">{{ pretty(detail.config_json || {}) }}</pre>
            </el-collapse-item>
            <el-collapse-item title="diff（脱敏）" name="diff">
              <el-table :data="detail.diff || []" border size="small">
                <el-table-column prop="path" label="路径" min-width="240" />
                <el-table-column prop="type" label="类型" width="120">
                  <template #default="{ row }">{{ diffTypeLabel(row.type) }}</template>
                </el-table-column>
                <el-table-column prop="before" label="变更前" min-width="240">
                  <template #default="{ row }"><span class="mono">{{ inline(row.before) }}</span></template>
                </el-table-column>
                <el-table-column prop="after" label="变更后" min-width="240">
                  <template #default="{ row }"><span class="mono">{{ inline(row.after) }}</span></template>
                </el-table-column>
              </el-table>
            </el-collapse-item>
          </el-collapse>
        </div>
        <el-skeleton v-else :rows="8" animated />
      </section>
    </el-drawer>

    <el-drawer v-model="rollbackVisible" size="70%" destroy-on-close title="回滚预览（预检）" data-testid="plugin-config-rollback-dryrun-drawer">
      <section class="action-panel in-drawer">
        <el-alert type="info" show-icon :closable="false" class="mb" title="预检不会修改当前配置；会用当前 config_schema 校验目标版本配置。" />
        <el-alert v-if="rollbackError" type="error" show-icon :closable="false" class="mb" :title="rollbackError" />
	        <div v-if="rollback" data-testid="plugin-config-rollback-preview">
	          <el-descriptions :column="2" border class="mb">
		            <el-descriptions-item label="状态">
		              <el-tag :type="rollback.status === 'blocked' ? 'danger' : rollback.status === 'warning' ? 'warning' : 'success'" effect="plain">
		                {{ genericStatusLabel(rollback.status) }}
		              </el-tag>
		            </el-descriptions-item>
		            <el-descriptions-item label="阻断码">{{ rollback.blocked_code || '-' }}</el-descriptions-item>
		            <el-descriptions-item label="插件编码">{{ rollback.plugin_code }}</el-descriptions-item>
		            <el-descriptions-item label="范围">{{ scopeLabel(rollback.scope) }}</el-descriptions-item>
		            <el-descriptions-item label="目标版本">{{ rollback.target_version?.version_no }}</el-descriptions-item>
	          </el-descriptions>

	          <el-alert
	            v-if="rollback.status === 'blocked'"
	            type="error"
	            show-icon
	            :closable="false"
	            class="mb"
	            :title="`回滚预览被阻断：${rollback.blocked_code || 'blocked'}`"
	            :description="rollback.suggestion || ''"
	          />
		          <el-alert v-else-if="rollback.schema_validation && rollback.schema_validation.valid === false" type="error" show-icon :closable="false" class="mb" title="配置模型校验失败（已阻断）" />
		          <el-table :data="rollback.diff || []" border size="small">
		            <el-table-column prop="path" label="路径" min-width="240" />
		            <el-table-column prop="type" label="类型" width="120">
                  <template #default="{ row }">{{ diffTypeLabel(row.type) }}</template>
                </el-table-column>
		            <el-table-column prop="before" label="变更前" min-width="240">
              <template #default="{ row }"><span class="mono">{{ inline(row.before) }}</span></template>
            </el-table-column>
	            <el-table-column prop="after" label="变更后" min-width="240">
              <template #default="{ row }"><span class="mono">{{ inline(row.after) }}</span></template>
            </el-table-column>
          </el-table>
        </div>
        <el-skeleton v-else :rows="8" animated />
      </section>
    </el-drawer>
  </el-dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import { listPluginConfigVersions, getPluginConfigVersionDetail, dryRunPluginConfigRollback, listCommunityPluginConfigVersions, getCommunityPluginConfigVersionDetail, dryRunCommunityPluginConfigRollback } from '@/api/admin';
import { genericStatusLabel } from '@/i18n/formatters';

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  pluginCode: { type: String, required: true },
  scope: { type: String, default: 'global' }, // global | community
  communityId: { type: Number, default: 0 },
});
const emit = defineEmits(['update:modelValue']);

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
});

const title = computed(() => (props.scope === 'community' ? `子站配置版本历史：${props.pluginCode}` : `全局配置版本历史：${props.pluginCode}`));

function scopeLabel(scope) {
  const labels = { global: '全局', community: '子站' };
  return labels[scope] || scope || '-';
}

function sourceLabel(source) {
  const labels = {
    admin: '后台修改',
    import: '导入',
    rollback: '回滚',
    system: '系统',
  };
  return labels[source] || source || '-';
}

function diffTypeLabel(type) {
  const labels = { added: '新增', removed: '删除', changed: '修改', unchanged: '未变化' };
  return labels[type] || type || '-';
}

const loading = ref(false);
const error = ref('');
const items = ref([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

const detailVisible = ref(false);
const detail = ref(null);
const detailError = ref('');

const rollbackVisible = ref(false);
const rollback = ref(null);
const rollbackError = ref('');

watch(
  () => visible.value,
  (v) => {
    if (v) fetchList(true);
  },
);

async function fetchList(reset = false) {
  if (reset) page.value = 1;
  loading.value = true;
  error.value = '';
  try {
    const params = { page: page.value, page_size: pageSize.value };
    const resp = props.scope === 'community'
      ? await listCommunityPluginConfigVersions(props.communityId, props.pluginCode, params)
      : await listPluginConfigVersions(props.pluginCode, params);
    items.value = resp.items || [];
    total.value = resp.pagination?.total ?? 0;
  } catch (e) {
    error.value = formatError(e, '加载版本历史失败');
    items.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
}

async function openDetail(row) {
  detailVisible.value = true;
  detail.value = null;
  detailError.value = '';
  try {
    const id = row?.id;
    detail.value = props.scope === 'community'
      ? await getCommunityPluginConfigVersionDetail(props.communityId, props.pluginCode, id)
      : await getPluginConfigVersionDetail(props.pluginCode, id);
  } catch (e) {
    detailError.value = formatError(e, '加载版本详情失败');
  }
}

async function openRollbackPreview(row) {
  rollbackVisible.value = true;
  rollback.value = null;
  rollbackError.value = '';
  try {
    const id = row?.id;
    rollback.value = props.scope === 'community'
      ? await dryRunCommunityPluginConfigRollback(props.communityId, props.pluginCode, id)
      : await dryRunPluginConfigRollback(props.pluginCode, id);
  } catch (e) {
    rollbackError.value = formatError(e, '回滚预览失败');
  }
}

function formatError(e, fallback) {
  const data = e?.response?.data;
  const code = String(data?.code || '').trim();
  const message = String(data?.message || data?.error || '').trim();
  const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
  const parts = [];
  if (code) parts.push(`[${code}]`);
  if (message) parts.push(message);
  if (suggestion) parts.push(`建议：${suggestion}`);
  return parts.join(' ') || String(e?.message || fallback);
}

function pretty(v) {
  try {
    return JSON.stringify(v ?? {}, null, 2);
  } catch {
    return String(v ?? '');
  }
}

function inline(v) {
  if (v == null) return '-';
  if (typeof v === 'string') return v;
  try {
    return JSON.stringify(v);
  } catch {
    return String(v);
  }
}
</script>
