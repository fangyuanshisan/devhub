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
          <el-button link type="primary" @click="openPlugins(row)">插件</el-button>
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

  <el-drawer v-model="pluginDrawer" :title="`${currentCommunity.name || ''} 插件管理`" size="720px">
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
    <el-table :data="pluginRows" border stripe>
      <el-table-column prop="sort_order" label="排序" width="120">
        <template #default="{ row, $index }">
          <el-input-number v-model="row.sort_order" :min="0" size="small" />
          <el-button link :disabled="$index === 0" @click="movePlugin($index, -1)">上移</el-button>
          <el-button link :disabled="$index === pluginRows.length - 1" @click="movePlugin($index, 1)">下移</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="code" label="code" width="110" />
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column label="内容类型" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="type in row.content_types || []" :key="type" class="mr-6">{{ type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="全局/子站" width="170">
        <template #default="{ row }">
          <el-tag :type="(row.global_status || row.status) === 'enabled' ? 'success' : 'info'" class="mr-6">{{ row.global_status || row.status }}</el-tag>
          <el-tag :type="(row.community_status || row.status) === 'enabled' ? 'success' : 'info'">{{ row.community_status || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-if="row.community_status !== 'enabled'" link type="success" @click="setCommunityPlugin(row, 'enabled')">启用</el-button>
          <el-button v-else link type="warning" @click="setCommunityPlugin(row, 'disabled')">禁用</el-button>
          <el-button link type="primary" @click="openPluginConfig(row)">配置</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-drawer>

  <el-dialog v-model="pluginConfigDialog" :title="`${pluginConfigTarget?.name || ''} 配置`" width="640px">
    <el-form label-width="110px">
      <el-alert title="当前 config_schema 仅用于声明展示；这里保存时只校验 config_json 是合法 JSON，暂不做 schema 强校验。" type="info" show-icon :closable="false" class="mb" />
      <el-form-item label="config_json">
        <el-input v-model="pluginConfigText" type="textarea" :rows="10" placeholder="{}" />
      </el-form-item>
      <el-alert title="禁用提示" type="info" show-icon :closable="false">
        <template #default>禁用子站插件只影响新发布、导航、菜单和管理入口，不影响历史内容访问。</template>
      </el-alert>
    </el-form>
    <template #footer>
      <el-button @click="pluginConfigDialog = false">取消</el-button>
      <el-button type="primary" @click="savePluginConfig">保存</el-button>
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
import { adminCommunities, adminCommunityCategories, adminCommunityPlugins, createCommunity, createCommunityCategory, disableCategory, disableCommunity, disableCommunityPlugin, enableCategory, enableCommunity, enableCommunityPlugin, reorderCategories, reorderCommunities, reorderCommunityPlugins, updateCategory, updateCommunity, updateCommunityPluginConfig } from '@/api/admin';

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
const pluginConfigDialog = ref(false);
const pluginConfigTarget = ref(null);
const pluginConfigText = ref('{}');
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
  const data = await adminCommunityPlugins(currentCommunity.value.id);
  pluginRows.value = data.items || [];
}

async function setCommunityPlugin(row, status) {
  if (status === 'disabled') {
    await ElMessageBox.confirm(
      '禁用子站插件后，只影响新发布、导航、菜单和管理入口，不影响历史内容访问。是否继续？',
      '禁用确认',
      { type: 'warning' },
    );
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
  pluginConfigText.value = row.config_json && row.config_json.trim() ? row.config_json : '{}';
  pluginConfigDialog.value = true;
}

async function savePluginConfig() {
  const row = pluginConfigTarget.value;
  if (!row) return;
  const raw = (pluginConfigText.value || '').trim();
  if (raw !== '') {
    try {
      JSON.parse(raw);
    } catch (e) {
      ElMessage.error('config_json 不是合法 JSON');
      return;
    }
  }
  await updateCommunityPluginConfig(currentCommunity.value.id, row.code, { config_json: raw ? JSON.parse(raw) : {} });
  ElMessage.success('插件配置已保存');
  pluginConfigDialog.value = false;
  await loadPlugins();
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
</style>
