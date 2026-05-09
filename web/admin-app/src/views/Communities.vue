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
import { ElMessage } from 'element-plus';
import { adminCommunities, adminCommunityCategories, createCommunity, createCommunityCategory, disableCategory, disableCommunity, enableCategory, enableCommunity, reorderCategories, reorderCommunities, updateCategory, updateCommunity } from '@/api/admin';

const router = useRouter();
const keyword = ref('');
const rows = ref([]);
const drawer = ref(false);
const editing = ref({});
const categoryDrawer = ref(false);
const categoryDialog = ref(false);
const currentCommunity = ref({});
const categoryRows = ref([]);
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
