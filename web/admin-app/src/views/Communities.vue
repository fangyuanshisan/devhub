<template>
  <section class="filter-panel">
    <el-form :inline="true">
      <el-form-item label="子站搜索">
        <el-input v-model="keyword" placeholder="名称 / slug" clearable style="width: 220px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="load">刷新</el-button>
        <el-button @click="openCreate">新增子站</el-button>
        <el-button :disabled="rows.length < 2" @click="saveOrder">保存排序</el-button>
      </el-form-item>
    </el-form>
  </section>

  <el-card shadow="never">
    <template #header>
      <div class="card-head">
        <span>子站管理</span>
        <el-button type="primary" @click="openCreate">新增子站</el-button>
      </div>
    </template>
    <el-table :data="filteredRows" stripe border>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column label="子站" min-width="220">
        <template #default="{ row }">
          <div class="content-cell">
            <div class="content-thumb">{{ row.logo || row.name?.slice(0, 2) }}</div>
            <div>
              <div class="content-title">{{ row.name }} <span class="content-meta">/{{ row.slug }}</span></div>
              <div class="content-meta">{{ row.slogan || row.description || '暂无简介' }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="统计" min-width="220">
        <template #default="{ row }">
          <span class="content-meta">主题 {{ row.topic_count || 0 }}</span>
          <span class="content-meta">评论 {{ row.comment_count || 0 }}</span>
          <span class="content-meta">关注 {{ row.follower_count || 0 }}</span>
        </template>
      </el-table-column>
      <el-table-column label="主题色" width="120">
        <template #default="{ row }"><el-color-picker v-model="row.theme_color" disabled /></template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : row.status === 2 ? '归档' : '禁用' }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="280">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="openCategories(row)">板块</el-button>
          <el-button link type="primary" :data-testid="`community-plugins-${row.id}`" @click="openPlugins(row)">插件</el-button>
          <el-button link type="primary" @click="openModerators(row)">版主</el-button>
          <el-button link type="primary" @click="openFrontend(row)">前台</el-button>
          <el-button v-if="row.status !== 1" link type="success" @click="enable(row)">启用</el-button>
          <el-button v-else link type="warning" @click="disable(row)">禁用</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-drawer v-model="drawer" :title="editing.id ? '编辑子站' : '新增子站'" size="620px">
    <el-form :model="form" label-width="120px">
      <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="Slug"><el-input v-model="form.slug" :disabled="Boolean(editing.id)" /></el-form-item>
      <el-form-item label="Logo"><el-input v-model="form.logo" /></el-form-item>
      <el-form-item label="封面图"><el-input v-model="form.cover_image" placeholder="/frontend-assets/community-php.jpg" /></el-form-item>
      <el-form-item label="Slogan"><el-input v-model="form.slogan" /></el-form-item>
      <el-form-item label="简介"><el-input v-model="form.description" type="textarea" rows="3" /></el-form-item>
      <el-form-item label="主题色"><el-color-picker v-model="form.theme_color" /><el-input v-model="form.theme_color" style="width: 160px; margin-left: 12px" /></el-form-item>
      <el-form-item label="排序"><el-input-number v-model="form.sort_order" :min="0" /></el-form-item>
      <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio-button :label="1">启用</el-radio-button><el-radio-button :label="0">禁用</el-radio-button><el-radio-button :label="2">归档</el-radio-button></el-radio-group></el-form-item>
      <el-divider>SEO</el-divider>
      <el-form-item label="SEO Title"><el-input v-model="form.seo_title" /></el-form-item>
      <el-form-item label="SEO Description"><el-input v-model="form.seo_description" type="textarea" rows="2" /></el-form-item>
      <el-form-item label="SEO Keywords"><el-input v-model="form.seo_keywords" /></el-form-item>
      <el-divider>公告</el-divider>
      <el-form-item label="公告标题"><el-input v-model="form.announcement_title" /></el-form-item>
      <el-form-item label="公告内容"><el-input v-model="form.announcement_content" type="textarea" rows="3" /></el-form-item>
      <el-form-item label="公告链接"><el-input v-model="form.announcement_url" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="drawer = false">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-drawer>

  <el-drawer v-model="categoryDrawer" :title="`${currentCommunity.name || ''} 板块管理`" size="720px">
    <div class="batch-actions"><el-button type="primary" @click="openCategoryCreate">新增板块</el-button><el-button @click="saveCategoryOrder">保存排序</el-button></div>
    <el-table :data="categoryRows" stripe border>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column prop="name" label="名称" min-width="120" />
      <el-table-column prop="slug" label="Slug" min-width="120" />
      <el-table-column prop="type" label="内容类型" width="120" />
      <el-table-column label="导航" width="90"><template #default="{ row }"><el-tag :type="row.nav_visible ? 'success' : 'info'">{{ row.nav_visible ? '显示' : '隐藏' }}</el-tag></template></el-table-column>
      <el-table-column label="发帖" width="90"><template #default="{ row }"><el-tag :type="row.postable ? 'success' : 'info'">{{ row.postable ? '允许' : '关闭' }}</el-tag></template></el-table-column>
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" fixed="right" width="170">
        <template #default="{ row }">
          <el-button link type="primary" @click="openCategoryEdit(row)">编辑</el-button>
          <el-button v-if="row.status !== 1" link type="success" @click="enableCat(row)">启用</el-button>
          <el-button v-else link type="warning" @click="disableCat(row)">禁用</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-drawer>

  <el-drawer v-model="pluginDrawer" :title="`${currentCommunity.name || ''} 插件配置`" size="860px" data-testid="community-plugin-drawer">
    <div class="drawer-head">
      <div>
        <h3 class="drawer-title">{{ currentCommunity.name }} <span class="muted">/{{ currentCommunity.slug }}</span></h3>
        <p class="muted">插件需同时满足“全局 enabled + 子站 enabled”才可用于发布与菜单展示。禁用不影响历史内容访问，只限制新发布与入口。</p>
      </div>
      <div>
        <el-button class="mr-6" @click="loadPlugins">刷新</el-button>
        <el-button type="primary" :disabled="pluginRows.length < 2" @click="savePluginOrder">保存排序</el-button>
      </div>
    </div>
    <div class="plugin-summary">
      <div><strong>{{ enabledCommunityPlugins }}</strong><span>子站已启用</span></div>
      <div><strong>{{ disabledCommunityPlugins }}</strong><span>子站未启用</span></div>
      <div><strong>{{ globallyDisabledPlugins }}</strong><span>全局禁用</span></div>
    </div>
    <div class="plugin-toolbar">
      <el-segmented v-model="pluginFilters.mode" :options="pluginFilterModes" />
      <el-input v-model="pluginFilters.q" placeholder="搜索 name / code" clearable style="width: 220px" />
      <el-select v-model="pluginFilters.contentType" placeholder="content_type" clearable filterable style="width: 200px">
        <el-option v-for="ct in allPluginContentTypes" :key="ct" :label="ct" :value="ct" />
      </el-select>
    </div>

    <el-table v-loading="pluginLoading" :data="filteredPluginRows" border stripe empty-text="暂无插件">
      <el-table-column prop="sort_order" label="排序" width="150">
        <template #default="{ row, $index }">
          <el-input-number v-model="row.sort_order" :min="0" size="small" controls-position="right" class="sort-input" />
          <div class="sort-actions">
            <el-button link :disabled="indexInAllPlugins(row.code) === 0" @click="movePluginByCode(row.code, -1)">上移</el-button>
            <el-button
              link
              :disabled="indexInAllPlugins(row.code) === pluginRows.length - 1"
              @click="movePluginByCode(row.code, 1)"
            >
              下移
            </el-button>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="插件" min-width="200">
        <template #default="{ row }">
          <div class="plugin-name">
            <strong>{{ row.name }}</strong>
            <span>{{ row.code }} · v{{ row.version || '-' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="内容类型" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="type in row.content_types || []" :key="type" class="mr-6" effect="plain">{{ type }}</el-tag>
          <span v-if="!(row.content_types || []).length" class="muted">无</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="190">
        <template #default="{ row }">
          <div class="status-stack">
            <span><em>全局</em><el-tag :type="statusType(globalStatus(row))">{{ globalStatus(row) }}</el-tag></span>
            <span><em>子站</em><el-tag :type="statusType(communityStatus(row))">{{ communityStatus(row) }}</el-tag></span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="说明" min-width="230">
        <template #default="{ row }">
          <span :class="globalStatus(row) !== 'enabled' ? 'danger-text' : 'muted'">{{ pluginAvailabilityText(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="配置覆盖" width="140">
        <template #default="{ row }">
          <el-tag v-if="hasCommunityConfigOverride(row)" type="warning" effect="plain">已覆盖</el-tag>
          <el-tag v-else type="info" effect="plain">默认</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="210">
        <template #default="{ row }">
          <el-button v-if="communityStatus(row) !== 'enabled'" link type="success" :data-testid="`community-plugin-enable-${row.code}`" :disabled="globalStatus(row) !== 'enabled'" @click="setCommunityPlugin(row, 'enabled')">启用</el-button>
          <el-button v-else link type="warning" :data-testid="`community-plugin-disable-${row.code}`" @click="setCommunityPlugin(row, 'disabled')">禁用</el-button>
          <el-button link type="primary" :data-testid="`community-plugin-config-${row.code}`" @click="openPluginConfig(row)">配置</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-drawer>

  <el-dialog v-model="pluginConfigDialog" :title="`${pluginConfigTarget?.name || ''} 配置`" width="820px" data-testid="community-plugin-config-dialog">
    <el-form label-width="110px">
      <el-alert
        title="子站 config_json 会覆盖全局配置；保存前会校验 JSON 合法性，并在可用时按 config_schema 做基础校验。"
        type="info"
        show-icon
        :closable="false"
        class="mb"
      />
      <el-form-item label="状态">
        <div class="config-status">
          <el-tag :type="statusType(globalStatus(pluginConfigTarget || {}))">全局 {{ globalStatus(pluginConfigTarget || {}) }}</el-tag>
          <el-tag :type="statusType(communityStatus(pluginConfigTarget || {}))">子站 {{ communityStatus(pluginConfigTarget || {}) }}</el-tag>
        </div>
      </el-form-item>
      <el-form-item label="config_schema">
        <pre class="json-box compact">{{ formatJSON(pluginConfigTarget?.config_schema || {}) }}</pre>
      </el-form-item>
      <el-form-item label="config_json">
        <PluginJsonEditor v-model="pluginConfigValue" :schema="pluginConfigTarget?.config_schema || null" @schema-errors="onCommunityConfigSchemaErrors">
          <template #title><strong>子站 config_json</strong></template>
        </PluginJsonEditor>
        <div class="config-actions">
          <el-button @click="clearPluginConfig">清空为 {}</el-button>
          <el-button type="primary" data-testid="community-plugin-config-save" :disabled="communityConfigSchemaErrors.length > 0" @click="savePluginConfig">保存</el-button>
        </div>
      </el-form-item>
      <el-form-item label="resolved_config">
        <pre class="json-box compact">{{ formatJSON(pluginConfigTarget?.resolved_config?.effective || pluginConfigTarget?.resolved_config || {}) }}</pre>
      </el-form-item>
      <el-alert title="禁用提示" type="info" show-icon :closable="false">
        <template #default>禁用子站插件只影响新发布、导航、菜单和管理入口，不影响历史内容访问。</template>
      </el-alert>
    </el-form>
    <template #footer>
      <el-button @click="pluginConfigDialog = false">关闭</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="categoryDialog" :title="categoryForm.id ? '编辑板块' : '新增板块'" width="520px">
    <el-form :model="categoryForm" label-width="100px">
      <el-form-item label="名称"><el-input v-model="categoryForm.name" /></el-form-item>
      <el-form-item label="Slug"><el-input v-model="categoryForm.slug" /></el-form-item>
      <el-form-item label="内容类型"><el-select v-model="categoryForm.content_type"><el-option v-for="item in contentTypes" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
      <el-form-item label="说明"><el-input v-model="categoryForm.description" type="textarea" rows="2" /></el-form-item>
      <el-form-item label="图标"><el-input v-model="categoryForm.icon" /></el-form-item>
      <el-form-item label="排序"><el-input-number v-model="categoryForm.sort_order" :min="0" /></el-form-item>
      <el-form-item label="导航展示"><el-switch v-model="categoryForm.nav_visible" /></el-form-item>
      <el-form-item label="允许发帖"><el-switch v-model="categoryForm.postable" /></el-form-item>
      <el-form-item label="状态"><el-switch v-model="categoryEnabled" active-text="启用" inactive-text="禁用" /></el-form-item>
      <el-form-item label="SEO Title"><el-input v-model="categoryForm.seo_title" /></el-form-item>
      <el-form-item label="SEO Description"><el-input v-model="categoryForm.seo_description" type="textarea" rows="2" /></el-form-item>
    </el-form>
    <template #footer><el-button @click="categoryDialog = false">取消</el-button><el-button type="primary" @click="saveCategory">保存</el-button></template>
  </el-dialog>
</template>

<script setup>
import { computed, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  adminCommunities,
  adminCommunityCategories,
  adminCommunityPlugins,
  communityPluginImpact,
  createCommunity,
  createCommunityCategory,
  disableCategory,
  disableCommunity,
  disableCommunityPlugin,
  enableCategory,
  enableCommunity,
  enableCommunityPlugin,
  reorderCategories,
  reorderCommunities,
  reorderCommunityPlugins,
  updateCategory,
  updateCommunity,
  updateCommunityPluginConfig,
} from '@/api/admin';
import PluginJsonEditor from '@/components/plugin/PluginJsonEditor.vue';

const router = useRouter();
const keyword = ref('');
const rows = ref([]);
const drawer = ref(false);
const editing = ref({});
const categoryDrawer = ref(false);
const categoryDialog = ref(false);
const pluginDrawer = ref(false);
const currentCommunity = ref({});
const categoryRows = ref([]);
const pluginRows = ref([]);
const pluginLoading = ref(false);
const pluginConfigDialog = ref(false);
const pluginConfigTarget = ref(null);
const pluginConfigValue = ref({});
const communityConfigSchemaErrors = ref([]);
const pluginFilters = reactive({ mode: 'all', q: '', contentType: '' });
const pluginFilterModes = [
  { label: '全部', value: 'all' },
  { label: '子站已启用', value: 'community_enabled' },
  { label: '子站未启用', value: 'community_disabled' },
  { label: '全局已禁用', value: 'global_disabled' },
];
const form = reactive(blankCommunity());
const categoryForm = reactive(blankCategory());
const contentTypes = [
  { label: '文章', value: 'article' },
  { label: '问答', value: 'question' },
  { label: '项目', value: 'project' },
  { label: 'AI 作品', value: 'ai_work' },
  { label: '招聘', value: 'job' },
  { label: 'Wiki', value: 'wiki' },
  { label: '文档', value: 'doc' },
];
const filteredRows = computed(() => {
  const q = keyword.value.trim().toLowerCase();
  if (!q) return rows.value;
  return rows.value.filter((row) => `${row.name} ${row.slug}`.toLowerCase().includes(q));
});
const categoryEnabled = computed({ get: () => categoryForm.status === 1, set: (value) => { categoryForm.status = value ? 1 : 0; } });
const enabledCommunityPlugins = computed(() => pluginRows.value.filter((row) => communityStatus(row) === 'enabled').length);
const disabledCommunityPlugins = computed(() => pluginRows.value.filter((row) => communityStatus(row) !== 'enabled').length);
const globallyDisabledPlugins = computed(() => pluginRows.value.filter((row) => globalStatus(row) !== 'enabled').length);

const allPluginContentTypes = computed(() => {
  const seen = new Set();
  for (const p of pluginRows.value || []) {
    for (const ct of p.content_types || []) seen.add(ct);
  }
  return Array.from(seen).sort();
});

const filteredPluginRows = computed(() => {
  const q = (pluginFilters.q || '').trim().toLowerCase();
  return (pluginRows.value || []).filter((p) => {
    if (pluginFilters.mode === 'community_enabled' && communityStatus(p) !== 'enabled') return false;
    if (pluginFilters.mode === 'community_disabled' && communityStatus(p) === 'enabled') return false;
    if (pluginFilters.mode === 'global_disabled' && globalStatus(p) === 'enabled') return false;
    if (pluginFilters.contentType && !(p.content_types || []).includes(pluginFilters.contentType)) return false;
    if (!q) return true;
    return (p.code || '').toLowerCase().includes(q) || (p.name || '').toLowerCase().includes(q);
  });
});

function blankCommunity() {
  return { name: '', slug: '', logo: '', cover_image: '', slogan: '', description: '', theme_color: '#2563eb', seo_title: '', seo_description: '', seo_keywords: '', sort_order: 0, status: 1, announcement_title: '', announcement_content: '', announcement_url: '' };
}
function blankCategory() {
  return { id: 0, name: '', slug: '', type: 'article', content_type: 'article', description: '', icon: '', sort_order: 0, visible: true, nav_visible: true, postable: true, seo_title: '', seo_description: '', status: 1 };
}
async function load() {
  const data = await adminCommunities();
  rows.value = data.items || [];
}
function openCreate() {
  editing.value = {};
  Object.assign(form, blankCommunity(), { sort_order: rows.value.length + 1 });
  drawer.value = true;
}
function openEdit(row) {
  editing.value = row;
  Object.assign(form, blankCommunity(), row);
  drawer.value = true;
}
async function save() {
  const payload = { ...form };
  editing.value.id ? await updateCommunity(editing.value.id, payload) : await createCommunity(payload);
  drawer.value = false;
  ElMessage.success('已保存子站');
  await load();
}
async function enable(row) {
  await enableCommunity(row.id);
  await load();
}
async function disable(row) {
  await disableCommunity(row.id);
  await load();
}
async function saveOrder() {
  await reorderCommunities({ ids: rows.value.map((row) => row.id) });
  ElMessage.success('排序已保存');
  await load();
}
function openFrontend(row) {
  window.open(`/c/${row.slug}/`, '_blank');
}
function openModerators(row) {
  router.push(`/moderators?community_id=${row.id}&community_slug=${row.slug}`);
}
async function openCategories(row) {
  currentCommunity.value = row;
  categoryDrawer.value = true;
  await loadCategories();
}
async function loadCategories() {
  const data = await adminCommunityCategories(currentCommunity.value.id);
  categoryRows.value = data.items || [];
}

async function openPlugins(row) {
  currentCommunity.value = row;
  pluginDrawer.value = true;
  await loadPlugins();
}

async function loadPlugins() {
  pluginLoading.value = true;
  try {
    const data = await adminCommunityPlugins(currentCommunity.value.id);
    pluginRows.value = data.items || [];
  } finally {
    pluginLoading.value = false;
  }
}

async function setCommunityPlugin(row, status) {
  if (status === 'disabled') {
    let impact = null;
    try {
      impact = await communityPluginImpact(currentCommunity.value.id, row.code);
    } catch {
      impact = null;
    }
    const lines = [];
    if (impact) {
      lines.push(`影响板块：${impact.categories_count}`);
      lines.push(`已有内容：${impact.topics_count}（历史仍可访问，SEO 不受影响）`);
      if (typeof impact.pending_topics_count === 'number') lines.push(`审核中内容：${impact.pending_topics_count}`);
    } else {
      lines.push('影响范围统计待后端接口支持或当前环境暂不可用。');
    }
    await ElMessageBox.confirm(
      `禁用该子站插件后，仅当前子站不能新发布该插件内容，当前子站导航、发布入口、版主菜单会隐藏。历史内容详情页和 SEO 不受影响。\n\n${lines.join('\n')}`,
      '禁用确认',
      { type: 'warning', confirmButtonText: '确认禁用', cancelButtonText: '取消' },
    );
  } else {
    if (globalStatus(row) !== 'enabled') {
      ElMessage.warning('该插件已被全局禁用，不能在子站启用。');
      return;
    }
    await ElMessageBox.confirm('启用后，当前子站可以在允许的板块中发布该插件内容，并显示对应导航与菜单入口。是否继续？', '启用确认', {
      type: 'info',
      confirmButtonText: '确认启用',
      cancelButtonText: '取消',
    });
  }
  if (status === 'enabled') await enableCommunityPlugin(currentCommunity.value.id, row.code);
  else await disableCommunityPlugin(currentCommunity.value.id, row.code);
  ElMessage.success('子站插件已更新');
  await loadPlugins();
  if (categoryDrawer.value) await loadCategories();
}

function movePlugin(index, delta) {
  const next = index + delta;
  if (next < 0 || next >= pluginRows.value.length) return;
  const arr = [...pluginRows.value];
  const tmp = arr[index];
  arr[index] = arr[next];
  arr[next] = tmp;
  pluginRows.value = arr;
  normalizePluginSortOrder();
}

function indexInAllPlugins(code) {
  return (pluginRows.value || []).findIndex((p) => p.code === code);
}

function movePluginByCode(code, delta) {
  const idx = indexInAllPlugins(code);
  if (idx < 0) return;
  movePlugin(idx, delta);
}

async function savePluginOrder() {
  const codes = [...pluginRows.value]
    .sort((a, b) => Number(a.sort_order || 0) - Number(b.sort_order || 0))
    .map((p) => p.code);
  await reorderCommunityPlugins(currentCommunity.value.id, { codes });
  ElMessage.success('插件排序已保存');
  await loadPlugins();
}

function openPluginConfig(row) {
  pluginConfigTarget.value = row;
  communityConfigSchemaErrors.value = [];
  pluginConfigValue.value = jsonValue(row?.config_json);
  pluginConfigDialog.value = true;
}

function onCommunityConfigSchemaErrors(errs) {
  communityConfigSchemaErrors.value = errs || [];
}

async function savePluginConfig() {
  const row = pluginConfigTarget.value;
  if (!row) return;
  if (communityConfigSchemaErrors.value.length > 0) {
    ElMessage.error('config_schema 校验失败，无法保存');
    return;
  }
  await updateCommunityPluginConfig(currentCommunity.value.id, row.code, { config_json: pluginConfigValue.value || {} });
  ElMessage.success('插件配置已保存');
  pluginConfigDialog.value = false;
  await loadPlugins();
}

function clearPluginConfig() {
  pluginConfigValue.value = {};
  ElMessage.success('已清空');
}

function normalizePluginSortOrder() {
  pluginRows.value = pluginRows.value.map((row, index) => ({ ...row, sort_order: index + 1 }));
}

function globalStatus(row) {
  return row?.global_status || row?.status || 'disabled';
}

function communityStatus(row) {
  return row?.community_status || row?.status || 'disabled';
}

function statusType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'danger';
  return 'info';
}

function pluginAvailabilityText(row) {
  if (globalStatus(row) !== 'enabled') return '该插件已被全局禁用，不能在子站启用。';
  if (communityStatus(row) !== 'enabled') return '当前子站未启用，不能新发布对应内容。';
  return '当前子站已启用，可用于发布入口、导航和版主菜单。';
}

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
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

function hasCommunityConfigOverride(row) {
  const v = row?.config_json;
  if (!v) return false;
  if (typeof v === 'string') {
    const s = v.trim();
    return s !== '' && s !== '{}' && s !== 'null';
  }
  if (typeof v === 'object') return Object.keys(v).length > 0;
  return false;
}
function openCategoryCreate() {
  Object.assign(categoryForm, blankCategory(), { sort_order: categoryRows.value.length + 1 });
  categoryDialog.value = true;
}
function openCategoryEdit(row) {
  Object.assign(categoryForm, blankCategory(), row, { content_type: row.content_type || row.type });
  categoryDialog.value = true;
}
async function saveCategory() {
  const payload = { ...categoryForm, type: categoryForm.content_type };
  categoryForm.id ? await updateCategory(categoryForm.id, payload) : await createCommunityCategory(currentCommunity.value.id, payload);
  categoryDialog.value = false;
  ElMessage.success('板块已保存');
  await loadCategories();
}
async function enableCat(row) {
  await enableCategory(row.id);
  await loadCategories();
}
async function disableCat(row) {
  await disableCategory(row.id);
  await loadCategories();
}
async function saveCategoryOrder() {
  await reorderCategories({ ids: categoryRows.value.map((row) => row.id) });
  ElMessage.success('板块排序已保存');
  await loadCategories();
}

load();
</script>

<style scoped>
.drawer-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; margin-bottom: 12px; }
.drawer-title { margin: 0 0 6px; }
.muted { color: #64748b; }
.mr-6 { margin-right: 6px; }
.danger-text { color: #b91c1c; }
.plugin-summary { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin-bottom: 14px; }
.plugin-summary div { border: 1px solid #e2e8f0; border-radius: 14px; padding: 12px; background: #f8fafc; display: grid; gap: 4px; }
.plugin-summary strong { font-size: 22px; color: #0f172a; }
.plugin-summary span { color: #64748b; font-size: 12px; }
.plugin-toolbar { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; margin-bottom: 12px; justify-content: flex-end; }
.sort-input { width: 92px; }
.sort-actions { display: flex; gap: 6px; margin-top: 4px; }
.plugin-name { display: grid; gap: 3px; }
.plugin-name strong { color: #0f172a; }
.plugin-name span { color: #64748b; font-size: 12px; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.status-stack { display: grid; gap: 6px; }
.status-stack span { display: flex; align-items: center; gap: 6px; }
.status-stack em { min-width: 32px; color: #64748b; font-style: normal; font-size: 12px; }
.config-status { display: flex; gap: 8px; flex-wrap: wrap; }
.config-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 10px; }
.json-box { margin: 0; width: 100%; box-sizing: border-box; padding: 12px; border-radius: 12px; background: #0f172a; color: #dbeafe; overflow: auto; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; line-height: 1.55; white-space: pre-wrap; }
.json-box.compact { max-height: 180px; }
.mb { margin-bottom: 12px; }
</style>
