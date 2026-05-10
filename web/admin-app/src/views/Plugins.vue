<template>
  <section class="page-card">
    <el-alert
      title="插件系统用于扩展 DevHub 的内容类型、权限、菜单、配置和子站能力。禁用插件不会影响历史内容访问，只影响新发布和入口展示。"
      type="info"
      show-icon
      :closable="false"
      class="intro-alert"
    />
    <div class="toolbar">
      <div>
        <h2>系统插件</h2>
        <p>Core 保留通用底座，业务能力通过系统插件声明、启停、权限、菜单和配置扩展。</p>
      </div>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-table v-loading="loading" :data="items" border stripe empty-text="暂无插件">
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
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="260" />
      <el-table-column label="操作" fixed="right" width="300">
        <template #default="{ row }">
          <el-button link type="primary" @click="openManifest(row)">详情</el-button>
          <el-button link type="info" @click="openManifest(row, 'permissions')">权限</el-button>
          <el-button link type="info" @click="openManifest(row, 'menus')">菜单</el-button>
          <el-button link type="primary" @click="openConfig(row)">配置</el-button>
          <el-button v-if="canOpen(row)" link type="primary" @click="openPlugin(row)">管理</el-button>
          <el-button v-if="row.status !== 'enabled'" link type="success" @click="setStatus(row, 'enabled')">启用</el-button>
          <el-button v-if="row.status === 'enabled'" link type="warning" @click="setStatus(row, 'disabled')">禁用</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>

  <el-drawer v-model="manifestDialog" :title="`${manifestTarget?.name || ''} 插件详情`" size="820px">
    <template v-if="manifestTarget">
      <div class="detail-hero">
        <div>
          <div class="detail-title">
            <h3>{{ manifestTarget.name }}</h3>
            <el-tag :type="statusType(manifestTarget.status)">{{ manifestTarget.status }}</el-tag>
            <el-tag v-if="manifestTarget.is_system" type="primary">system</el-tag>
          </div>
          <p>{{ manifestTarget.description || '暂无插件说明' }}</p>
        </div>
        <div class="detail-code">{{ manifestTarget.code }}</div>
      </div>

      <el-tabs v-model="manifestTab" class="detail-tabs">
        <el-tab-pane label="基础信息" name="basic">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="name">{{ manifestTarget.name }}</el-descriptions-item>
            <el-descriptions-item label="code">{{ manifestTarget.code }}</el-descriptions-item>
            <el-descriptions-item label="version">{{ manifestTarget.version }}</el-descriptions-item>
            <el-descriptions-item label="status">{{ manifestTarget.status }}</el-descriptions-item>
            <el-descriptions-item label="is_system">{{ manifestTarget.is_system ? '是' : '否' }}</el-descriptions-item>
            <el-descriptions-item label="min_core_version">{{ manifestTarget.min_core_version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="dependencies" :span="2">{{ (manifestTarget.dependencies || []).join(', ') || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <el-tab-pane label="内容类型" name="contentTypes">
          <el-table :data="manifestTarget.content_type_definitions || []" border stripe empty-text="暂无内容类型声明">
            <el-table-column prop="type" label="type" width="130" />
            <el-table-column prop="name" label="名称" width="130" />
            <el-table-column prop="create_permission" label="create_permission" min-width="190" />
            <el-table-column prop="edit_permission" label="edit_permission" min-width="150" />
            <el-table-column prop="delete_permission" label="delete_permission" min-width="150" />
            <el-table-column prop="seo_type" label="seo_type" width="130" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="权限" name="permissions">
          <el-table :data="manifestTarget.permissions || []" border stripe empty-text="暂无权限声明">
            <el-table-column prop="code" label="permission code" min-width="220" />
            <el-table-column prop="name" label="名称" width="150" />
            <el-table-column prop="scope" label="scope" width="150" />
            <el-table-column prop="description" label="说明" min-width="220" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="菜单" name="menus">
          <el-table :data="manifestTarget.menus || []" border stripe empty-text="暂无菜单声明">
            <el-table-column label="area" width="120">
              <template #default="{ row }">{{ row.area || row.location || '-' }}</template>
            </el-table-column>
            <el-table-column prop="title" label="标题" width="150" />
            <el-table-column prop="path" label="path" min-width="180" />
            <el-table-column prop="permission" label="permission" min-width="190" />
            <el-table-column prop="sort_order" label="排序" width="90" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="配置" name="config">
          <el-alert title="当前仅校验 JSON 格式，schema 强校验和自动表单生成属于后续插件平台能力。" type="info" show-icon :closable="false" class="mb" />
          <div class="json-section">
            <div class="json-head">
              <strong>config_schema</strong>
              <el-button size="small" @click="copyJSON(manifestTarget.config_schema)">复制</el-button>
            </div>
            <pre class="json-box">{{ formatJSON(manifestTarget.config_schema || {}) }}</pre>
          </div>
          <div class="json-section">
            <div class="json-head">
              <strong>resolved_config</strong>
              <el-button size="small" @click="copyJSON(manifestTarget.resolved_config)">复制</el-button>
            </div>
            <pre class="json-box">{{ formatJSON(manifestTarget.resolved_config || {}) }}</pre>
          </div>
        </el-tab-pane>

        <el-tab-pane label="路由" name="routes">
          <el-table :data="manifestTarget.routes || []" border stripe empty-text="暂无路由声明">
            <el-table-column prop="area" label="area" width="120" />
            <el-table-column prop="method" label="method" width="110" />
            <el-table-column prop="path" label="path" min-width="220" />
            <el-table-column prop="handler" label="handler/auth" min-width="220" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="Hooks" name="hooks">
          <el-table :data="hooksWithCoverage(manifestTarget)" border stripe>
            <el-table-column prop="name" label="Hook" min-width="180" />
            <el-table-column label="声明状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.declared ? 'success' : 'info'">{{ row.declared ? '已声明' : '未声明 / 待完善' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="failure_policy" label="失败策略" width="130" />
            <el-table-column prop="description" label="说明" min-width="240" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </template>
  </el-drawer>

  <el-dialog v-model="configDialog" :title="`${configTarget?.name || ''} 全局配置`" width="640px">
    <el-alert title="全局配置会作为默认配置和子站配置之间的中间层，子站 config_json 优先级更高；当前仅校验 JSON 格式，暂不做 config_schema 强校验。" type="info" show-icon :closable="false" class="mb" />
    <div class="json-section">
      <div class="json-head">
        <strong>config_schema 参考</strong>
        <el-button size="small" @click="copyJSON(configTarget?.config_schema)">复制 schema</el-button>
      </div>
      <pre class="json-box compact">{{ formatJSON(configTarget?.config_schema || {}) }}</pre>
    </div>
    <el-input v-model="configText" type="textarea" :rows="10" placeholder="{}" />
    <template #footer>
      <el-button @click="configDialog = false">取消</el-button>
      <el-button @click="formatConfig">格式化</el-button>
      <el-button type="primary" @click="saveConfig">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import { disablePlugin, enablePlugin, plugins, updatePluginConfig } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const router = useRouter();
const items = ref([]);
const loading = ref(false);
const manifestDialog = ref(false);
const manifestTarget = ref(null);
const manifestTab = ref('basic');
const configDialog = ref(false);
const configTarget = ref(null);
const configText = ref('{}');
const hookNames = [
  'BeforeCreateContent',
  'AfterCreateContent',
  'BeforeUpdateContent',
  'AfterUpdateContent',
  'BeforeDeleteContent',
  'AfterDeleteContent',
  'AfterCreateComment',
  'OnSearchIndex',
  'OnNotificationBuild',
  'OnSEOBuild',
];

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
    await ElMessageBox.confirm(
      '全局禁用该插件后，所有子站将不能新发布该插件内容，相关导航、菜单和管理入口会隐藏。历史内容详情页和 SEO 不受影响。',
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

function openManifest(row, tab = 'basic') {
  manifestTarget.value = row;
  manifestTab.value = tab;
  manifestDialog.value = true;
}

function openConfig(row) {
  configTarget.value = row;
  configText.value = editableJSON(row.config_json);
  configDialog.value = true;
}

async function saveConfig() {
  const row = configTarget.value;
  if (!row) return;
  const raw = (configText.value || '').trim();
  try {
    await updatePluginConfig(row.code, { config_json: raw ? JSON.parse(raw) : {} });
  } catch (err) {
    if (err instanceof SyntaxError) {
      ElMessage.error('config_json 不是合法 JSON');
      return;
    }
    throw err;
  }
  ElMessage.success('插件全局配置已保存');
  configDialog.value = false;
  await load();
}

function formatConfig() {
  const raw = (configText.value || '').trim();
  try {
    configText.value = JSON.stringify(raw ? JSON.parse(raw) : {}, null, 2);
  } catch {
    ElMessage.error('config_json 不是合法 JSON，无法格式化');
  }
}

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
}

function editableJSON(value) {
  if (!value) return '{}';
  if (typeof value === 'string') return value.trim() ? value : '{}';
  return formatJSON(value);
}

function statusType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'danger';
  return 'info';
}

function hasConfigSchema(row) {
  const schema = row?.config_schema;
  if (!schema) return false;
  if (Array.isArray(schema)) return schema.length > 0;
  if (typeof schema === 'object') return Object.keys(schema).length > 0;
  return true;
}

function hooksWithCoverage(row) {
  const declared = new Map((row?.hooks || []).map((hook) => [hook.name, hook]));
  return hookNames.map((name) => {
    const hook = declared.get(name);
    return {
      name,
      declared: Boolean(hook),
      failure_policy: hook?.failure_policy || '-',
      description: hook?.description || '未声明 / 待完善',
    };
  });
}

async function copyJSON(value) {
  try {
    await navigator.clipboard.writeText(formatJSON(value || {}));
    ElMessage.success('已复制');
  } catch {
    ElMessage.warning('当前浏览器不支持自动复制');
  }
}

onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 16px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; }
.toolbar h2 { margin: 0 0 6px; }
.toolbar p { margin: 0; color: #64748b; }
.intro-alert { border-radius: 12px; }
.mr-6 { margin-right: 6px; }
.mb { margin-bottom: 12px; }
.muted { color: #64748b; }
.plugin-title { display: grid; gap: 2px; margin-bottom: 6px; }
.plugin-title strong { color: #0f172a; }
.plugin-title span { color: #64748b; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.metric-line { display: flex; flex-wrap: wrap; gap: 6px; }
.detail-hero { display: flex; justify-content: space-between; gap: 16px; padding: 18px; border: 1px solid #e2e8f0; border-radius: 16px; background: linear-gradient(135deg, #f8fafc, #eef6ff); }
.detail-title { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.detail-title h3 { margin: 0; font-size: 20px; color: #0f172a; }
.detail-hero p { margin: 8px 0 0; color: #64748b; }
.detail-code { align-self: flex-start; padding: 6px 10px; border-radius: 999px; background: #0f172a; color: #fff; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.detail-tabs { margin-top: 16px; }
.json-section { display: grid; gap: 8px; margin-bottom: 14px; }
.json-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.json-box { margin: 0; padding: 14px; border-radius: 12px; background: #0f172a; color: #dbeafe; max-height: 360px; overflow: auto; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; line-height: 1.55; white-space: pre-wrap; }
.json-box.compact { max-height: 180px; }
</style>
