<template>
  <el-dialog v-model="visible" width="980px" destroy-on-close :title="title" data-testid="plugin-config-versions-dialog">
    <el-alert type="info" show-icon :closable="false" class="mb" title="提示：历史版本与 diff 会对敏感字段脱敏；回滚预览不会修改当前配置。" />

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <el-card shadow="never" class="mb" data-testid="plugin-config-versions-list">
      <template #header>
        <div class="card-head">
          <strong>版本历史</strong>
          <span class="muted">(total {{ total }})</span>
        </div>
      </template>
      <el-table :data="items" border size="small" v-loading="loading" data-testid="plugin-config-versions-table">
        <el-table-column prop="version_no" label="version_no" width="110" />
        <el-table-column prop="scope" label="scope" width="110" />
        <el-table-column prop="community_id" label="community" width="120" />
        <el-table-column prop="source" label="source" width="130" />
        <el-table-column prop="operator_name" label="operator" width="160" />
        <el-table-column prop="created_at" label="created_at" width="170" />
        <el-table-column label="changed_keys" min-width="220">
          <template #default="{ row }">
            <div class="tag-wrap">
              <el-tag v-for="k in row.changed_keys || []" :key="k" effect="plain">{{ k }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="actions" width="160" fixed="right">
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
            <el-descriptions-item label="version_no">{{ detail.version?.version_no }}</el-descriptions-item>
            <el-descriptions-item label="scope">{{ detail.version?.scope }}</el-descriptions-item>
            <el-descriptions-item label="community_id">{{ detail.version?.community_id || 0 }}</el-descriptions-item>
            <el-descriptions-item label="source">{{ detail.version?.source || '-' }}</el-descriptions-item>
            <el-descriptions-item label="operator">{{ detail.version?.operator_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="created_at">{{ detail.version?.created_at || '-' }}</el-descriptions-item>
          </el-descriptions>

          <el-collapse>
            <el-collapse-item title="config_json（脱敏）" name="config">
              <pre class="json-box">{{ pretty(detail.config_json || {}) }}</pre>
            </el-collapse-item>
            <el-collapse-item title="diff（脱敏）" name="diff">
              <el-table :data="detail.diff || []" border size="small">
                <el-table-column prop="path" label="path" min-width="240" />
                <el-table-column prop="type" label="type" width="120" />
                <el-table-column prop="before" label="before" min-width="240">
                  <template #default="{ row }"><span class="mono">{{ inline(row.before) }}</span></template>
                </el-table-column>
                <el-table-column prop="after" label="after" min-width="240">
                  <template #default="{ row }"><span class="mono">{{ inline(row.after) }}</span></template>
                </el-table-column>
              </el-table>
            </el-collapse-item>
          </el-collapse>
        </div>
        <el-skeleton v-else :rows="8" animated />
      </section>
    </el-drawer>

    <el-drawer v-model="rollbackVisible" size="70%" destroy-on-close title="回滚预览（dry-run）" data-testid="plugin-config-rollback-dryrun-drawer">
      <section class="action-panel in-drawer">
        <el-alert type="info" show-icon :closable="false" class="mb" title="dry-run 不会修改当前配置；会用当前 config_schema 校验目标版本配置。" />
        <el-alert v-if="rollbackError" type="error" show-icon :closable="false" class="mb" :title="rollbackError" />
	        <div v-if="rollback" data-testid="plugin-config-rollback-preview">
	          <el-descriptions :column="2" border class="mb">
	            <el-descriptions-item label="status">
	              <el-tag :type="rollback.status === 'blocked' ? 'danger' : rollback.status === 'warning' ? 'warning' : 'success'" effect="plain">
	                {{ rollback.status }}
	              </el-tag>
	            </el-descriptions-item>
	            <el-descriptions-item label="blocked_code">{{ rollback.blocked_code || '-' }}</el-descriptions-item>
	            <el-descriptions-item label="plugin_code">{{ rollback.plugin_code }}</el-descriptions-item>
	            <el-descriptions-item label="scope">{{ rollback.scope }}</el-descriptions-item>
	            <el-descriptions-item label="target_version">{{ rollback.target_version?.version_no }}</el-descriptions-item>
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
	          <el-alert v-else-if="rollback.schema_validation && rollback.schema_validation.valid === false" type="error" show-icon :closable="false" class="mb" title="schema 校验失败（blocked）" />
	          <el-table :data="rollback.diff || []" border size="small">
	            <el-table-column prop="path" label="path" min-width="240" />
	            <el-table-column prop="type" label="type" width="120" />
	            <el-table-column prop="before" label="before" min-width="240">
              <template #default="{ row }"><span class="mono">{{ inline(row.before) }}</span></template>
            </el-table-column>
            <el-table-column prop="after" label="after" min-width="240">
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
