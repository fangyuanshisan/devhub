<template>
  <section class="page-card" data-testid="admin-plugins-page">
    <el-alert
      title="插件系统用于扩展 DevHub 的内容类型、权限、菜单、配置和子站能力。禁用插件不会影响历史内容访问，只影响新发布和入口展示。"
      type="info"
      show-icon
      :closable="false"
      class="intro-alert"
    />

    <div class="stats-grid" data-testid="plugin-stats">
      <div class="stat-card">
        <div class="stat-k">全部插件</div>
        <div class="stat-v">{{ stats.total }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">Enabled</div>
        <div class="stat-v">{{ stats.enabled }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">Disabled</div>
        <div class="stat-v">{{ stats.disabled }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">System</div>
        <div class="stat-v">{{ stats.system }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-k">有 Schema</div>
        <div class="stat-v">{{ stats.hasSchema }}</div>
      </div>
    </div>

    <div class="toolbar">
      <div>
        <h2>系统插件</h2>
        <p>Core 保留通用底座，业务能力通过系统插件声明、启停、权限、菜单和配置扩展。</p>
      </div>
      <div class="tool-actions">
        <el-input v-model="filters.q" data-testid="plugin-search" placeholder="搜索 code / name" clearable style="width: 220px" />
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 140px">
          <el-option label="全部" value="all" />
          <el-option label="enabled" value="enabled" />
          <el-option label="disabled" value="disabled" />
        </el-select>
        <el-select v-model="filters.contentType" placeholder="content_type" clearable filterable style="width: 180px">
          <el-option v-for="ct in allContentTypes" :key="ct" :label="ct" :value="ct" />
        </el-select>
        <el-select v-model="filters.system" placeholder="system" clearable style="width: 140px">
          <el-option label="全部" value="all" />
          <el-option label="仅 system" value="yes" />
          <el-option label="非 system" value="no" />
        </el-select>
        <el-select v-model="filters.hasSchema" placeholder="config_schema" clearable style="width: 160px">
          <el-option label="全部" value="all" />
          <el-option label="有 schema" value="yes" />
          <el-option label="无 schema" value="no" />
        </el-select>
        <el-button @click="load">刷新</el-button>
      </div>
    </div>
    <el-table v-loading="loading" :data="filteredItems" border stripe empty-text="暂无插件">
      <el-table-column label="插件" min-width="210">
        <template #default="{ row }">
          <div class="plugin-title">
            <strong>{{ row.name }}</strong>
            <span>{{ row.code }}</span>
          </div>
          <el-tag v-if="row.is_system" size="small" type="primary">system</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" effect="light">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="运行健康" min-width="170">
        <template #default="{ row }">
          <div class="health-cell">
            <el-tag :type="healthType(row.health?.status)" effect="light">{{ row.health?.status || 'unknown' }}</el-tag>
            <span class="muted">{{ row.health?.suggested_action || '暂无建议' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="内容类型" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="type in row.content_types || []" :key="type" class="mr-6" effect="plain">{{ type }}</el-tag>
          <span v-if="!(row.content_types || []).length" class="muted">无</span>
        </template>
      </el-table-column>
      <el-table-column label="能力摘要" min-width="220">
        <template #default="{ row }">
          <div class="metric-line">
            <el-tag type="info" effect="plain">权限 {{ (row.permissions || []).length }}</el-tag>
            <el-tag type="info" effect="plain">菜单 {{ (row.menus || []).length }}</el-tag>
            <el-tag :type="hasConfigSchema(row) ? 'success' : 'info'" effect="plain">
              schema {{ hasConfigSchema(row) ? '有' : '无' }}
            </el-tag>
            <el-tag :type="(row.hooks || []).length ? 'success' : 'info'" effect="plain">hooks {{ (row.hooks || []).length }}</el-tag>
            <el-tag :type="statusMetricType(row.health?.config_status)" effect="plain">配置 {{ row.health?.config_status || '-' }}</el-tag>
            <el-tag :type="statusMetricType(row.health?.migration_status)" effect="plain">迁移 {{ row.health?.migration_status || '-' }}</el-tag>
            <el-tag :type="statusMetricType(row.health?.hook_status)" effect="plain">Hook {{ row.health?.hook_status || '-' }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="最近错误" min-width="220">
        <template #default="{ row }">
          <span class="muted">{{ row.health?.recent_error || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="260" />
      <el-table-column label="操作" fixed="right" width="300">
        <template #default="{ row }">
          <el-button link type="primary" :data-testid="`plugin-detail-${row.code}`" @click="openManifest(row)">详情</el-button>
          <el-button link type="info" @click="openManifest(row, 'permissions')">权限</el-button>
          <el-button link type="info" @click="openManifest(row, 'menus')">菜单</el-button>
          <el-button link type="primary" @click="openManifest(row, 'config')">配置</el-button>
          <el-button v-if="canOpen(row)" link type="primary" :data-testid="`plugin-manage-${row.code}`" @click="openPlugin(row)">管理</el-button>
          <el-button v-if="row.status !== 'enabled'" link type="success" :data-testid="`plugin-enable-${row.code}`" @click="setStatus(row, 'enabled')">启用</el-button>
          <el-button v-if="row.status === 'enabled'" link type="warning" :data-testid="`plugin-disable-${row.code}`" @click="setStatus(row, 'disabled')">禁用</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>

  <PluginDetailDrawer v-model="manifestDialog" :plugin="manifestTarget" :initial-tab="manifestInitialTab" @refresh="load" />
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import { disablePlugin, enablePlugin, pluginImpact, plugins } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';
import PluginDetailDrawer from '@/components/plugin/PluginDetailDrawer.vue';

const auth = useAuthStore();
const router = useRouter();
const items = ref([]);
const filters = reactive({
  q: '',
  status: 'all',
  contentType: '',
  system: 'all',
  hasSchema: 'all',
});
const loading = ref(false);
const manifestDialog = ref(false);
const manifestTarget = ref(null);
const manifestInitialTab = ref('overview');

const allContentTypes = computed(() => {
  const seen = new Set();
  for (const p of items.value || []) {
    for (const ct of p.content_types || []) seen.add(ct);
  }
  return Array.from(seen).sort();
});

const filteredItems = computed(() => {
  const q = (filters.q || '').trim().toLowerCase();
  return (items.value || []).filter((p) => {
    if (filters.status && filters.status !== 'all' && p.status !== filters.status) return false;
    if (filters.contentType && filters.contentType !== '' && !(p.content_types || []).includes(filters.contentType)) return false;
    if (filters.system === 'yes' && !p.is_system) return false;
    if (filters.system === 'no' && p.is_system) return false;
    const hasSchema = hasConfigSchema(p);
    if (filters.hasSchema === 'yes' && !hasSchema) return false;
    if (filters.hasSchema === 'no' && hasSchema) return false;
    if (!q) return true;
    return (p.code || '').toLowerCase().includes(q) || (p.name || '').toLowerCase().includes(q);
  });
});

const stats = computed(() => {
  const list = items.value || [];
  const total = list.length;
  const enabled = list.filter((p) => p.status === 'enabled').length;
  const disabled = list.filter((p) => p.status === 'disabled').length;
  const system = list.filter((p) => p.is_system).length;
  const hasSchema = list.filter((p) => hasConfigSchema(p)).length;
  return { total, enabled, disabled, system, hasSchema };
});

async function load() {
  loading.value = true;
  try {
    const data = await plugins();
    items.value = data.items || [];
  } finally {
    loading.value = false;
  }
}

async function setStatus(row, status) {
  if (status === 'disabled') {
    let impact = null;
    try {
      impact = await pluginImpact(row.code);
    } catch {
      impact = null;
    }
    const lines = [];
    if (impact) {
      lines.push(`当前启用子站：${impact.enabled_communities_count ?? 0}`);
      lines.push(`当前未启用子站：${impact.disabled_communities_count ?? 0}`);
      lines.push(`将阻止发布的板块：${impact.categories_count ?? 0}`);
      lines.push(`已有历史内容：${impact.existing_contents_count ?? impact.topics_count ?? 0}（历史仍可访问，SEO 不受影响）`);
      if (typeof impact.recent_contents_count === 'number') lines.push(`近 7 天内容：${impact.recent_contents_count}`);
      if (typeof impact.pending_contents_count === 'number') lines.push(`审核中内容：${impact.pending_contents_count}`);
      if (typeof impact.configs_count === 'number') lines.push(`配置覆盖记录：${impact.configs_count}`);
      if (typeof impact.pending_migrations_count === 'number') lines.push(`待执行迁移：${impact.pending_migrations_count}`);
      lines.push(`菜单声明：${impact.menus_count}（frontend ${impact.frontend_menus_count} / moderator ${impact.moderator_menus_count} / admin ${impact.admin_menus_count}）`);
      if (typeof impact.recent_hook_errors_count === 'number') lines.push(`近期 Hook 错误：${impact.recent_hook_errors_count}`);
    } else {
      lines.push('影响范围统计待后端接口支持或当前环境暂不可用。');
    }
    await ElMessageBox.confirm(
      `全局禁用该插件后，所有子站将不能新发布该插件内容，相关导航、菜单和管理入口会隐藏。历史内容详情页和 SEO 不受影响。\n\n${lines.join('\n')}`,
      '禁用确认',
      { type: 'warning', confirmButtonText: '确认禁用', cancelButtonText: '取消' },
    );
  } else {
    await ElMessageBox.confirm('启用该插件后，全局可用性恢复；具体子站仍需单独启用后才能发布对应内容。是否继续？', '启用确认', {
      type: 'info',
      confirmButtonText: '确认启用',
      cancelButtonText: '取消',
    });
  }
  if (status === 'enabled') await enablePlugin(row.code);
  else await disablePlugin(row.code);
  ElMessage.success('插件状态已更新');
  await load();
}

function canOpen(row) {
  const target = adminMenu(row)?.path;
  return Boolean(target && row.status === 'enabled' && hasPermission(row));
}

function hasPermission(row) {
  const required = adminMenu(row)?.permission || row.permissions?.[0]?.code || '';
  return !required || auth.can(required);
}

function adminMenu(row) {
  return (row.menus || []).find((item) => (item.area || item.location) === 'admin');
}

function openPlugin(row) {
  const target = adminMenu(row)?.path;
  if (!target) return;
  router.push(target);
}

function openManifest(row, tab = 'overview') {
  manifestTarget.value = row;
  manifestInitialTab.value = tab || 'overview';
  manifestDialog.value = true;
}

function statusType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'danger';
  return 'info';
}

function healthType(status) {
  if (status === 'healthy') return 'success';
  if (status === 'disabled') return 'info';
  if (status === 'warning' || status === 'migration_pending' || status === 'hook_warning') return 'warning';
  if (status === 'hook_error') return 'danger';
  if (status === 'error' || status === 'config_invalid' || status === 'dependency_missing') return 'danger';
  return 'info';
}

function statusMetricType(status) {
  if (status === 'ok' || status === 'valid') return 'success';
  if (status === 'warning' || status === 'pending' || status === 'hook_warning') return 'warning';
  if (status === 'failed' || status === 'invalid' || status === 'missing' || status === 'hook_error') return 'danger';
  return 'info';
}

function hasConfigSchema(row) {
  const schema = row?.config_schema;
  if (!schema) return false;
  if (Array.isArray(schema)) return schema.length > 0;
  if (typeof schema === 'object') return Object.keys(schema).length > 0;
  return true;
}

onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: flex-end; gap: 16px; }
.tool-actions { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; justify-content: flex-end; }
.toolbar h2 { margin: 0 0 6px; }
.toolbar p { margin: 0; color: #64748b; }
.intro-alert { border-radius: 12px; }
.stats-grid { display: grid; grid-template-columns: repeat(5, minmax(150px, 1fr)); gap: 12px; }
.stat-card { min-height: 76px; border: 1px solid #e2e8f0; border-radius: 14px; background: #fff; padding: 10px 14px; }
.stat-k { color: #64748b; font-size: 12px; }
.stat-v { color: #0f172a; font-size: 22px; font-weight: 700; margin-top: 4px; }
.mr-6 { margin-right: 6px; }
.mb { margin-bottom: 12px; }
.muted { color: #64748b; }
.plugin-title { display: grid; gap: 2px; margin-bottom: 6px; }
.plugin-title strong { color: #0f172a; }
.plugin-title span { color: #64748b; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.metric-line { display: flex; flex-wrap: wrap; gap: 6px; }
.health-cell { display: grid; gap: 6px; }
.page-card :deep(.el-table__cell) { padding: 8px 0; }
.page-card :deep(.el-table .cell) { line-height: 1.35; }
.json-box { margin: 0; padding: 14px; border-radius: 12px; background: #0f172a; color: #dbeafe; max-height: 360px; overflow: auto; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; line-height: 1.55; white-space: pre-wrap; }
.json-box.compact { max-height: 180px; }
</style>
