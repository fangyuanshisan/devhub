<template>
  <el-drawer v-model="visible" :title="title" size="920px" data-testid="plugin-detail-drawer" class="plugin-detail-drawer">
    <template v-if="plugin">
      <div class="drawer-content">
        <div class="hero">
        <div class="hero-left">
          <div class="hero-title">
            <h3>{{ plugin.name }}</h3>
            <el-tag :type="statusType(plugin.status)">{{ plugin.status }}</el-tag>
            <el-tag :type="healthType(plugin.health?.status)">{{ plugin.health?.status || 'unknown' }}</el-tag>
            <el-tag v-if="plugin.is_system" type="primary">system</el-tag>
          </div>
          <p class="hero-desc">{{ plugin.description || '暂无插件说明' }}</p>
          <div class="hero-metrics">
            <el-tag type="info" effect="plain">content_types {{ (plugin.content_types || []).length }}</el-tag>
            <el-tag type="info" effect="plain">permissions {{ (plugin.permissions || []).length }}</el-tag>
            <el-tag type="info" effect="plain">menus {{ (plugin.menus || []).length }}</el-tag>
            <el-tag :type="(plugin.hooks || []).length ? 'success' : 'info'" effect="plain">hooks {{ (plugin.hooks || []).length }}</el-tag>
          </div>
        </div>
        <div class="hero-right">
          <div class="code-pill">{{ plugin.code }}</div>
          <div class="meta-line">version: {{ plugin.version }}</div>
        </div>
        </div>

        <el-tabs v-model="tab" class="tabs" data-testid="plugin-detail-tabs">
        <el-tab-pane label="概览" name="overview">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="name">{{ plugin.name }}</el-descriptions-item>
            <el-descriptions-item label="plugin_code">{{ plugin.code }}</el-descriptions-item>
            <el-descriptions-item label="version">{{ plugin.version }}</el-descriptions-item>
            <el-descriptions-item label="status">{{ plugin.status }}</el-descriptions-item>
            <el-descriptions-item label="health">{{ plugin.health?.status || '-' }}</el-descriptions-item>
            <el-descriptions-item label="is_system">{{ plugin.is_system ? '是' : '否' }}</el-descriptions-item>
            <el-descriptions-item label="maturity">{{ maturityLabel(plugin) }}</el-descriptions-item>
            <el-descriptions-item label="suggested_action">{{ plugin.health?.suggested_action || '-' }}</el-descriptions-item>
            <el-descriptions-item label="content_types" :span="2">{{ (plugin.content_types || []).join(', ') || '-' }}</el-descriptions-item>
          </el-descriptions>
          <el-alert
            class="mt"
            type="info"
            show-icon
            :closable="false"
            title="成熟度说明：平台治理已接入表示已纳入插件状态/权限/配置/Hook 等平台能力；业务闭环是否完成仍以各插件专项验收为准。"
          />
        </el-tab-pane>

        <el-tab-pane label="运行状态" name="runtime">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            class="mb"
            title="运行状态由全局启停、配置校验、迁移记录、依赖状态和 Hook 失败统计计算；禁用插件不会影响历史内容访问和 SEO。"
          />
          <el-descriptions :column="2" border>
            <el-descriptions-item label="overall">
              <el-tag :type="healthType(plugin.health?.status)">{{ plugin.health?.status || 'unknown' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="suggested_action">{{ plugin.health?.suggested_action || '-' }}</el-descriptions-item>
            <el-descriptions-item label="config_status">
              <el-tag :type="metricType(plugin.health?.config_status)">{{ plugin.health?.config_status || '-' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="migration_status">
              <el-tag :type="metricType(plugin.health?.migration_status)">{{ plugin.health?.migration_status || '-' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="hook_status">
              <el-tag :type="metricType(plugin.health?.hook_status)">{{ plugin.health?.hook_status || '-' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="dependency_status">
              <el-tag :type="metricType(plugin.health?.dependency_status)">{{ plugin.health?.dependency_status || '-' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="pending_migrations">{{ plugin.health?.pending_migrations_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="failed_migrations">{{ plugin.health?.failed_migrations_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="hook_failures">{{ plugin.health?.hook_failure_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="updated_at">{{ plugin.health?.updated_at || '-' }}</el-descriptions-item>
            <el-descriptions-item label="recent_error" :span="2">{{ plugin.health?.recent_error || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <el-tab-pane label="内容类型" name="contentTypes">
          <el-table :data="plugin.content_type_definitions || []" border stripe empty-text="暂无内容类型定义">
            <el-table-column prop="type" label="type" width="140" />
            <el-table-column prop="name" label="name" width="140" />
            <el-table-column prop="plugin_code" label="plugin_code" width="120" />
            <el-table-column prop="create_permission" label="create_permission" min-width="180" />
            <el-table-column prop="edit_permission" label="edit_permission" min-width="160" />
            <el-table-column prop="delete_permission" label="delete_permission" min-width="160" />
            <el-table-column prop="audit_permission" label="audit_permission" min-width="160" />
            <el-table-column prop="seo_type" label="seo_type" width="130" />
            <el-table-column label="flags" width="180">
              <template #default="{ row }">
                <el-tag size="small" effect="plain" :type="row.allow_comment ? 'success' : 'info'">comment</el-tag>
                <el-tag size="small" effect="plain" :type="row.allow_like ? 'success' : 'info'" class="ml">like</el-tag>
                <el-tag size="small" effect="plain" :type="row.allow_favorite ? 'success' : 'info'" class="ml">fav</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="权限" name="permissions">
          <div class="sub-toolbar">
            <el-input v-model="permQ" placeholder="按权限码搜索" clearable style="max-width: 320px" />
          </div>
          <el-table :data="filteredPermissions" border stripe empty-text="暂无权限定义">
            <el-table-column prop="code" label="code" min-width="240">
              <template #default="{ row }">
                <div class="mono">{{ row.code }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="name" min-width="160" />
            <el-table-column prop="scope" label="scope" width="150" />
            <el-table-column prop="description" label="description" min-width="220" />
            <el-table-column label="操作" width="90">
              <template #default="{ row }">
                <el-button link type="primary" @click="copyText(row.code)">复制</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="菜单" name="menus">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            title="菜单最终展示会受插件全局状态、子站状态与当前用户权限影响。"
            class="mb"
          />
          <el-table :data="plugin.menus || []" border stripe empty-text="暂无菜单声明">
            <el-table-column prop="area" label="area" width="120" />
            <el-table-column prop="title" label="title" width="160" />
            <el-table-column prop="path" label="path" min-width="220" />
            <el-table-column prop="permission" label="permission" min-width="200" />
            <el-table-column prop="sort_order" label="sort_order" width="120" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="配置" name="config">
          <el-alert
            title="当前已支持 JSON 合法性与 config_schema 的基础校验；更完整强校验与自动表单渲染仍属于后续插件平台能力。"
            type="info"
            show-icon
            :closable="false"
            class="mb"
          />

          <el-collapse v-model="configPanels">
            <el-collapse-item name="schema" title="config_schema">
              <pre class="json-box">{{ formatJSON(plugin.config_schema || {}) }}</pre>
            </el-collapse-item>
            <el-collapse-item name="global" title="global config_json（可编辑）" data-testid="plugin-global-config-panel">
              <PluginJsonEditor v-model="editableConfig" :schema="plugin.config_schema || null" @schema-errors="onSchemaErrors">
                <template #title>
                  <strong>全局 config_json</strong>
                </template>
              </PluginJsonEditor>
              <div class="config-actions">
                <el-button @click="reloadConfig">重置为当前值</el-button>
                <el-button type="primary" data-testid="plugin-global-config-save" :disabled="schemaErrors.length > 0" @click="saveConfig">保存</el-button>
              </div>
            </el-collapse-item>
            <el-collapse-item name="resolved" title="resolved_config（只读）">
              <pre class="json-box">{{ formatJSON(plugin.resolved_config || {}) }}</pre>
            </el-collapse-item>
          </el-collapse>
        </el-tab-pane>

        <el-tab-pane label="Hooks" name="hooks">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            class="mb"
            title="说明：平台调用点=当前后端代码已确认存在 Dispatch 接入点；运行统计来自 hook_executions。handler 状态仍不伪造，仅展示真实执行记录。"
          />
          <el-table v-loading="hooksLoading" :data="hooksRows" border stripe>
            <el-table-column prop="name" label="Hook" min-width="200" />
            <el-table-column label="manifest 声明" width="130">
              <template #default="{ row }">
                <el-tag :type="row.declared ? 'success' : 'info'">{{ row.declared ? '已声明' : '未声明' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="平台调用点" width="130">
              <template #default="{ row }">
                <el-tag :type="row.platformHook ? 'success' : 'warning'">{{ row.platformHook ? '存在' : '未知/未覆盖' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="handler" width="150">
              <template #default="{ row }">
                <el-tag :type="row.execution_count > 0 ? 'success' : 'info'">
                  {{ row.execution_count > 0 ? '有执行记录' : '暂无记录' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="mode" label="mode" width="120" />
            <el-table-column label="执行/失败" width="130">
              <template #default="{ row }">{{ row.execution_count || 0 }} / {{ row.failure_count || 0 }}</template>
            </el-table-column>
            <el-table-column label="失败率" width="100">
              <template #default="{ row }">{{ failureRate(row) }}</template>
            </el-table-column>
            <el-table-column label="平均耗时" width="110">
              <template #default="{ row }">{{ avgDuration(row) }}</template>
            </el-table-column>
            <el-table-column prop="last_executed_at" label="最近执行" min-width="160" />
            <el-table-column prop="last_failed_at" label="最近失败" min-width="160" />
            <el-table-column prop="last_error" label="最近错误" min-width="220" />
            <el-table-column prop="failure_policy" label="failure_policy" width="140" />
            <el-table-column prop="description" label="说明" min-width="240" />
          </el-table>
          <el-divider>最近执行</el-divider>
          <el-table :data="hookRecent" border stripe empty-text="暂无 Hook 执行记录">
            <el-table-column prop="finished_at" label="时间" width="170" />
            <el-table-column prop="hook_name" label="Hook" min-width="180" />
            <el-table-column prop="mode" label="mode" width="120" />
            <el-table-column label="结果" width="90">
              <template #default="{ row }">
                <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? 'success' : 'failed' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="content_type" label="content_type" width="140" />
            <el-table-column prop="content_id" label="content_id" width="120" />
            <el-table-column prop="community_id" label="community_id" width="130" />
            <el-table-column prop="duration_ms" label="耗时(ms)" width="100" />
            <el-table-column prop="error_message" label="错误" min-width="220" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="迁移" name="migrations">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            class="mb"
            title="当前仅支持内置插件 up migration 的执行记录、失败记录与重试；rollback/down 先保留 rollback_supported 标识，不做真实回滚。"
          />
          <div class="sub-toolbar">
            <el-tag type="info" effect="plain">total {{ migrationSummary.total || migrationRows.length }}</el-tag>
            <el-tag type="success" effect="plain">success {{ migrationSummary.success || 0 }}</el-tag>
            <el-tag type="warning" effect="plain">pending {{ migrationSummary.pending || 0 }}</el-tag>
            <el-tag type="danger" effect="plain">failed {{ migrationSummary.failed || 0 }}</el-tag>
            <el-button type="primary" size="small" @click="runMigrations">执行待迁移</el-button>
            <el-button size="small" @click="loadMigrations">刷新</el-button>
          </div>
          <el-table v-loading="migrationsLoading" :data="migrationRows" border stripe empty-text="暂无迁移声明">
            <el-table-column prop="migration_name" label="迁移" min-width="180" />
            <el-table-column prop="migration_version" label="version" width="120" />
            <el-table-column prop="direction" label="direction" width="100" />
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="migrationStatusType(row.status)">{{ row.status || 'pending' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="finished_at" label="最近完成" min-width="160" />
            <el-table-column label="耗时" width="110">
              <template #default="{ row }">{{ row.duration_ms || row.execution_time_ms || 0 }}ms</template>
            </el-table-column>
            <el-table-column label="rollback" width="110">
              <template #default="{ row }">
                <el-tag :type="row.rollback_supported ? 'warning' : 'info'" effect="plain">
                  {{ row.rollback_supported ? 'supported' : 'no' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="error_message" label="失败原因" min-width="220" />
            <el-table-column prop="description" label="说明" min-width="240" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" :disabled="row.status === 'success'" @click="retryMigration(row)">
                  {{ row.status === 'failed' ? '重试' : '执行' }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="路由" name="routes">
          <el-table :data="plugin.routes || []" border stripe empty-text="暂无路由声明">
            <el-table-column prop="area" label="area" width="120" />
            <el-table-column prop="method" label="method" width="110" />
            <el-table-column prop="path" label="path" min-width="240" />
            <el-table-column prop="handler" label="handler/auth" min-width="240" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="审计" name="audit">
          <el-alert
            type="info"
            show-icon
            :closable="false"
            title="本 Tab 读取插件专用审计接口，覆盖插件启停、配置、Hook 失败和带 plugin_code 的内容治理操作。"
            class="mb"
          />
          <div class="sub-toolbar">
            <el-input v-model="auditQ.action" placeholder="动作关键字（可选）" clearable style="max-width: 260px" />
            <el-input-number v-model="auditQ.communityId" :min="0" placeholder="community_id（可选）" controls-position="right" style="width: 200px" />
            <el-button @click="loadAudit">查询</el-button>
          </div>
          <el-table v-loading="auditLoading" :data="auditRows" border stripe empty-text="暂无审计记录">
            <el-table-column prop="id" label="ID" width="90" />
            <el-table-column prop="created_at" label="时间" width="170" />
            <el-table-column label="操作人" width="170">
              <template #default="{ row }">
                {{ row.actor || '-' }}
                <div class="muted">{{ row.actor_type || '-' }} / ID {{ row.actor_id || row.actor_user_id || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="action" label="动作" min-width="180" />
            <el-table-column label="目标" min-width="220">
              <template #default="{ row }">
                <div class="mono">{{ row.target || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column label="diff" min-width="260">
              <template #default="{ row }">
                <details>
                  <summary class="muted">查看 old/new/metadata</summary>
                  <pre class="json-box compact">{{ formatJSON(jsonValue(row.old_value)) }}</pre>
                  <pre class="json-box compact">{{ formatJSON(jsonValue(row.new_value)) }}</pre>
                  <pre class="json-box compact">{{ formatJSON(jsonValue(row.metadata_json)) }}</pre>
                </details>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="auditQ.page"
            v-model:page-size="auditQ.pageSize"
            class="pager"
            layout="total, sizes, prev, pager, next, jumper"
            :total="auditTotal"
            @change="loadAudit"
          />
        </el-tab-pane>
        </el-tabs>
      </div>
    </template>
  </el-drawer>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import PluginJsonEditor from './PluginJsonEditor.vue';
import { pluginAuditLogs, pluginHooks, pluginMigrations, retryPluginMigration, runPluginMigrations, updatePluginConfig } from '@/api/admin';

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  plugin: { type: Object, default: null },
  initialTab: { type: String, default: 'overview' },
});
const emit = defineEmits(['update:modelValue', 'refresh']);

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
});

watch(
  () => visible.value,
  (v) => {
    if (v && tab.value === 'audit') loadAudit();
  },
);

const tab = ref('overview');
const permQ = ref('');
const schemaErrors = ref([]);
const configPanels = ref(['global']);
const editableConfig = ref({});
const hooksLoading = ref(false);
const hookStats = ref([]);
const hookRecent = ref([]);
const migrationsLoading = ref(false);
const migrationRows = ref([]);
const migrationSummary = ref({});

const title = computed(() => `${props.plugin?.name || ''} 插件详情`);

watch(
  () => props.plugin,
  (p) => {
    tab.value = props.initialTab || 'overview';
    permQ.value = '';
    schemaErrors.value = [];
    editableConfig.value = jsonValue(p?.config_json);
    // Reset audit query state for new plugin target.
    auditQ.action = '';
    auditQ.communityId = 0;
    auditQ.page = 1;
    auditQ.pageSize = 20;
    auditRows.value = [];
    auditTotal.value = 0;
    hookStats.value = [];
    hookRecent.value = [];
    migrationRows.value = [];
    migrationSummary.value = {};
    if (visible.value && tab.value === 'hooks') loadHooks();
    if (visible.value && tab.value === 'migrations') loadMigrations();
  },
  { immediate: true },
);

watch(
  () => props.initialTab,
  (t) => {
    if (!visible.value) return;
    if (!t) return;
    tab.value = t;
    if (t === 'audit') loadAudit();
    if (t === 'hooks') loadHooks();
    if (t === 'migrations') loadMigrations();
  },
);

watch(tab, (t) => {
  if (!visible.value) return;
  if (t === 'audit') loadAudit();
  if (t === 'hooks') loadHooks();
  if (t === 'migrations') loadMigrations();
});

const filteredPermissions = computed(() => {
  const q = (permQ.value || '').trim().toLowerCase();
  const list = props.plugin?.permissions || [];
  if (!q) return list;
  return list.filter((p) => (p.code || '').toLowerCase().includes(q) || (p.name || '').toLowerCase().includes(q));
});

const allHookNames = [
  'BeforeCreateContent',
  'AfterCreateContent',
  'BeforeUpdateContent',
  'AfterUpdateContent',
  'BeforeModerateContent',
  'AfterModerateContent',
  'BeforeBuildSEO',
  'AfterBuildSEO',
  'AfterPluginEnabled',
  'AfterPluginDisabled',
  'AfterCreateComment',
  'OnSearchIndex',
  'OnNotificationBuild',
  'OnSEOBuild',
];

// 平台“声明”了这些 Hook，但并不代表当前代码已经在所有流程中完整触发。
// 这里仅用于避免 UI 伪造“平台调用点存在”的结论。
// 是否触发，以后端真实 Dispatch 接入点为准（详见 docs/PLUGIN_ARCHITECTURE.md / docs/TESTING.md）。
const platformDispatchedHooks = new Set([
  // 已确认的接入点（根据当前后端 service/router 的 DispatchHook 调用）
  'BeforeCreateContent',
  'AfterCreateContent',
  'BeforeUpdateContent',
  'AfterUpdateContent',
  'BeforeModerateContent',
  'AfterModerateContent',
  'AfterCreateComment',
  'OnSearchIndex',
  'OnNotificationBuild',
  'OnSEOBuild',
  'AfterPluginEnabled',
  'AfterPluginDisabled',
]);

const hooksRows = computed(() => {
  const declared = new Map((props.plugin?.hooks || []).map((h) => [h.name, h]));
  const stats = new Map((hookStats.value || []).map((h) => [h.hook_name, h]));
  return allHookNames.map((name) => {
    const hook = declared.get(name);
    const stat = stats.get(name) || {};
    return {
      name,
      declared: Boolean(hook),
      platformHook: platformDispatchedHooks.has(name),
      failure_policy: hook?.failure_policy || '-',
      description: hook?.description || '-',
      mode: stat.mode || (hook?.critical ? 'blocking' : 'non_blocking'),
      execution_count: stat.execution_count || 0,
      failure_count: stat.failure_count || 0,
      avg_duration_ms: stat.avg_duration_ms || 0,
      last_executed_at: stat.last_executed_at || '-',
      last_failed_at: stat.last_failed_at || '-',
      last_error: stat.last_error || '-',
    };
  });
});

function statusType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'danger';
  return 'info';
}

function healthType(status) {
  if (status === 'healthy') return 'success';
  if (status === 'disabled') return 'info';
  if (status === 'warning' || status === 'migration_pending') return 'warning';
  if (status === 'error' || status === 'config_invalid' || status === 'dependency_missing') return 'danger';
  return 'info';
}

function metricType(status) {
  if (status === 'ok' || status === 'valid') return 'success';
  if (status === 'warning' || status === 'pending') return 'warning';
  if (status === 'failed' || status === 'invalid' || status === 'missing') return 'danger';
  return 'info';
}

function migrationStatusType(status) {
  if (status === 'success') return 'success';
  if (status === 'failed') return 'danger';
  if (status === 'running' || status === 'pending') return 'warning';
  return 'info';
}

function maturityLabel(plugin) {
  if (!plugin) return '-';
  if (plugin.code === 'qa' || plugin.code === 'docs' || plugin.code === 'wiki') return '平台治理已接入';
  return '业务闭环待完善';
}

function jsonValue(v) {
  if (!v) return {};
  if (typeof v === 'string') {
    try {
      return JSON.parse(v);
    } catch {
      return {};
    }
  }
  if (typeof v === 'object') return v;
  return {};
}

const auditLoading = ref(false);
const auditRows = ref([]);
const auditTotal = ref(0);
const auditQ = reactive({
  action: '',
  communityId: null,
  page: 1,
  pageSize: 20,
});

async function loadAudit() {
  const p = props.plugin;
  if (!p || !p.code) return;
  auditLoading.value = true;
  try {
    const data = await pluginAuditLogs(p.code, {
      type: 'all',
      action: auditQ.action || '',
      community_id: auditQ.communityId || 0,
      page: auditQ.page,
      page_size: auditQ.pageSize,
    });
    auditRows.value = data.items || [];
    auditTotal.value = data.total || 0;
  } finally {
    auditLoading.value = false;
  }
}

async function loadHooks() {
  const p = props.plugin;
  if (!p || !p.code) return;
  hooksLoading.value = true;
  try {
    const data = await pluginHooks(p.code);
    hookStats.value = data.items || [];
    hookRecent.value = data.recent_executions || [];
  } catch (e) {
    hookStats.value = [];
    hookRecent.value = [];
    ElMessage.warning(String(e?.message || e || 'Hook 统计暂不可用'));
  } finally {
    hooksLoading.value = false;
  }
}

async function loadMigrations() {
  const p = props.plugin;
  if (!p || !p.code) return;
  migrationsLoading.value = true;
  try {
    const data = await pluginMigrations(p.code);
    migrationRows.value = data.items || [];
    migrationSummary.value = data.summary || {};
  } catch (e) {
    migrationRows.value = [];
    migrationSummary.value = {};
    ElMessage.warning(String(e?.message || e || '迁移状态暂不可用'));
  } finally {
    migrationsLoading.value = false;
  }
}

async function runMigrations() {
  const p = props.plugin;
  if (!p || !p.code) return;
  migrationsLoading.value = true;
  try {
    await runPluginMigrations(p.code);
    ElMessage.success('迁移执行完成');
    await loadMigrations();
    emit('refresh');
  } catch (e) {
    ElMessage.error(String(e?.message || e || '迁移执行失败'));
  } finally {
    migrationsLoading.value = false;
  }
}

async function retryMigration(row) {
  const p = props.plugin;
  if (!p || !p.code || !row?.migration_name) return;
  migrationsLoading.value = true;
  try {
    await retryPluginMigration(p.code, row.migration_name);
    ElMessage.success(row.status === 'failed' ? '迁移重试完成' : '迁移执行完成');
    await loadMigrations();
    emit('refresh');
  } catch (e) {
    ElMessage.error(String(e?.message || e || '迁移执行失败'));
  } finally {
    migrationsLoading.value = false;
  }
}

function failureRate(row) {
  const total = Number(row.execution_count || 0);
  if (!total) return '-';
  return `${Math.round((Number(row.failure_count || 0) / total) * 100)}%`;
}

function avgDuration(row) {
  const value = Number(row.avg_duration_ms || 0);
  if (!Number.isFinite(value)) return '-';
  return `${value.toFixed(value >= 10 ? 0 : 1)}ms`;
}

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(String(text || ''));
    ElMessage.success('已复制');
  } catch {
    ElMessage.warning('当前浏览器不支持自动复制');
  }
}

function onSchemaErrors(errs) {
  schemaErrors.value = Array.isArray(errs) ? errs : [];
}

function reloadConfig() {
  editableConfig.value = jsonValue(props.plugin?.config_json);
  ElMessage.success('已重置');
}

async function saveConfig() {
  const p = props.plugin;
  if (!p) return;
  try {
    await updatePluginConfig(p.code, { config_json: editableConfig.value || {} });
    ElMessage.success('插件全局配置已保存');
    emit('refresh');
  } catch (e) {
    ElMessage.error(String(e?.message || e || '保存失败'));
  }
}
</script>

<style scoped>
.hero {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: linear-gradient(135deg, #f8fafc, #eef6ff);
}
.drawer-content {
  min-height: calc(100vh - 96px);
  padding-bottom: 24px;
}
.hero-title {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.hero-title h3 {
  margin: 0;
  font-size: 20px;
  color: #0f172a;
}
.hero-desc {
  margin: 8px 0 0;
  color: #64748b;
}
.hero-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}
.code-pill {
  align-self: flex-start;
  padding: 6px 10px;
  border-radius: 999px;
  background: #0f172a;
  color: #fff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.meta-line {
  margin-top: 10px;
  color: #64748b;
  font-size: 12px;
}
.tabs {
  margin-top: 16px;
}
.tabs :deep(.el-tabs__content) {
  min-height: 420px;
  padding-top: 4px;
}
.tabs :deep(.el-tab-pane) {
  min-height: 360px;
}
.sub-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: #0f172a;
}
.mb {
  margin-bottom: 12px;
}
.mt {
  margin-top: 12px;
}
.json-box {
  margin: 0;
  padding: 14px;
  border-radius: 12px;
  background: #0f172a;
  color: #dbeafe;
  max-height: 360px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
}
.config-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 10px;
}
:global(.plugin-detail-drawer .el-drawer__body) {
  padding-top: 10px;
  overflow: auto;
}
</style>
