<template>
  <section class="filter-panel">
    <el-form :inline="true" :model="query">
      <el-form-item label="子站">
        <el-select v-model="query.site" style="width: 180px">
          <el-option label="全部子站" value="" />
          <el-option label="总站" value="portal" />
          <el-option v-for="item in communities" :key="item.slug" :label="item.name" :value="item.slug" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" style="width: 140px">
          <el-option label="默认" value="all" />
          <el-option label="启用" value="enable" />
          <el-option label="禁用" value="disable" />
          <el-option label="已合并" value="merged" />
        </el-select>
      </el-form-item>
      <el-form-item label="搜索">
        <el-input v-model="query.q" clearable placeholder="名称 / slug / 描述" style="width: 220px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
        <el-button type="primary" plain @click="openCreate">新增标签</el-button>
        <el-button plain @click="recalculateAll">全部重算</el-button>
      </el-form-item>
    </el-form>
  </section>

  <el-card shadow="never">
    <template #header>
      <div class="card-head">
        <span>标签管理</span>
        <div style="display: flex; gap: 8px">
          <el-button plain @click="recalculateAll">全部重算</el-button>
          <el-button type="primary" @click="openCreate">新增标签</el-button>
        </div>
      </div>
    </template>
    <el-table :data="rows" stripe border>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column label="标签" min-width="260">
        <template #default="{ row }">
          <div class="content-cell">
            <div class="content-thumb">#</div>
            <div>
              <div class="content-title">{{ row.name }} <span class="content-meta">/{{ row.slug }}</span></div>
              <div class="content-meta">{{ row.description || '暂无说明' }}</div>
              <div v-if="row.status === 'merged' && row.merged_to_name" class="content-meta">已合并到 {{ row.merged_to_name }} / {{ row.merged_to_slug }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="子站" width="120">
        <template #default="{ row }">{{ communityName(row) }}</template>
      </el-table-column>
      <el-table-column label="别名" min-width="200">
        <template #default="{ row }">
          <span class="content-meta">{{ aliasSummary(row.id) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="统计" width="200">
        <template #default="{ row }">
          <div class="content-meta">内容 {{ row.topic_count || row.use_count || 0 }}</div>
          <div class="content-meta">关注 {{ row.follower_count || 0 }} / 热度 {{ row.hot_score || 0 }}</div>
        </template>
      </el-table-column>
      <el-table-column label="SEO" min-width="220">
        <template #default="{ row }">
          <div class="content-meta">{{ row.seo_title || '未设置 SEO title' }}</div>
          <div class="content-meta">{{ row.seo_description || '未设置 SEO description' }}</div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="360">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="openAliases(row)">别名</el-button>
          <el-button link type="primary" @click="openTopics(row)">关联内容</el-button>
          <el-button link type="primary" @click="openFrontend(row)">前台</el-button>
          <el-button link type="primary" @click="recalculateOne(row)">重算</el-button>
          <el-button v-if="row.status !== 'merged'" link type="warning" @click="openMerge(row)">合并</el-button>
          <el-button v-if="row.status !== 'enable' && row.status !== 'merged'" link type="success" @click="enable(row)">启用</el-button>
          <el-button v-else-if="row.status === 'enable'" link type="warning" @click="disable(row)">禁用</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-drawer v-model="drawer" :title="form.id ? '编辑标签' : '新增标签'" size="560px">
    <el-form :model="form" label-width="120px">
      <el-form-item label="所属子站">
        <el-select v-model="form.site" filterable>
          <el-option label="总站" value="portal" />
          <el-option v-for="item in communities" :key="item.slug" :label="item.name" :value="item.slug" />
        </el-select>
      </el-form-item>
      <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="Slug"><el-input v-model="form.slug" placeholder="留空按名称生成" /></el-form-item>
      <el-form-item label="说明"><el-input v-model="form.description" type="textarea" rows="3" /></el-form-item>
      <el-form-item label="排序"><el-input-number v-model="form.sort_order" :min="0" /></el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="form.status">
          <el-radio-button label="enable">启用</el-radio-button>
          <el-radio-button label="disable">禁用</el-radio-button>
          <el-radio-button v-if="form.id && form.status === 'merged'" label="merged">已合并</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-divider>SEO</el-divider>
      <el-form-item label="SEO Title"><el-input v-model="form.seo_title" /></el-form-item>
      <el-form-item label="SEO Description"><el-input v-model="form.seo_description" type="textarea" rows="2" /></el-form-item>
      <el-form-item label="SEO Keywords"><el-input v-model="form.seo_keywords" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="drawer = false">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-drawer>

  <el-drawer v-model="topicsDrawer" :title="`${currentTag.name || ''} 关联内容`" size="760px">
    <el-table :data="topicRows" stripe border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column label="标题" min-width="260">
        <template #default="{ row }">
          <el-link type="primary" :href="`/topics/${row.id}/`" target="_blank">{{ row.title }}</el-link>
          <div class="content-meta">{{ row.summary || row.content || '暂无摘要' }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="content_type" label="类型" width="110" />
      <el-table-column prop="view_count" label="浏览" width="90" />
      <el-table-column prop="comment_count" label="评论" width="90" />
      <el-table-column prop="created_at" label="发布时间" width="170" />
    </el-table>
    <el-pagination v-model:current-page="topicQuery.page" v-model:page-size="topicQuery.page_size" class="pager" layout="total, sizes, prev, pager, next" :total="topicTotal" @change="loadTopics" />
  </el-drawer>

  <el-drawer v-model="aliasDrawer" :title="`${currentTag.name || ''} 标签别名`" size="520px">
    <el-form :inline="true">
      <el-form-item>
        <el-input v-model="aliasForm.alias" placeholder="输入别名，例如 JS" style="width: 260px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="createAlias">添加别名</el-button>
      </el-form-item>
    </el-form>
    <el-table :data="aliasRows" stripe border>
      <el-table-column prop="alias" label="别名" min-width="180" />
      <el-table-column prop="alias_slug" label="Alias Slug" min-width="180" />
      <el-table-column prop="created_at" label="创建时间" width="170" />
      <el-table-column label="操作" width="90">
        <template #default="{ row }">
          <el-button link type="danger" @click="removeAlias(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-drawer>

  <el-dialog v-model="mergeDialog" title="合并标签" width="560px">
    <el-form :model="mergeForm" label-width="110px">
      <el-form-item label="源标签">
        <span>{{ currentTag.name }} / {{ currentTag.slug }}</span>
      </el-form-item>
      <el-form-item label="目标标签">
        <el-select v-model="mergeForm.target_tag_id" filterable placeholder="搜索目标标签" style="width: 100%">
          <el-option v-for="item in mergeCandidates" :key="item.id" :label="`${item.name} / ${item.slug}`" :value="item.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="mergeForm.note" type="textarea" rows="3" placeholder="可选，记录重复标签合并原因" />
      </el-form-item>
      <el-alert title="合并后将迁移 topic_tags 与 follows，并将源标签置为 merged。" type="warning" :closable="false" />
    </el-form>
    <template #footer>
      <el-button @click="mergeDialog = false">取消</el-button>
      <el-button type="warning" @click="submitMerge">确认合并</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  adminCommunities,
  createTag,
  createTagAlias,
  deleteTagAlias,
  disableTag,
  enableTag,
  mergeTag,
  recalculateAllTags,
  recalculateTag,
  tagAliases,
  tagTopics,
  tags,
  updateTag,
} from '@/api/admin';

const query = reactive({ site: '', status: 'all', q: '' });
const topicQuery = reactive({ page: 1, page_size: 10 });
const rows = ref([]);
const communities = ref([]);
const drawer = ref(false);
const topicsDrawer = ref(false);
const aliasDrawer = ref(false);
const mergeDialog = ref(false);
const currentTag = ref({});
const topicRows = ref([]);
const topicTotal = ref(0);
const aliasRows = ref([]);
const aliasCache = ref({});
const mergeCandidates = ref([]);
const aliasForm = reactive({ alias: '' });
const mergeForm = reactive({ target_tag_id: 0, note: '' });
const form = reactive(blankTag());

function blankTag() {
  return { id: 0, site: 'portal', name: '', slug: '', description: '', status: 'enable', sort: 0, sort_order: 0, seo_title: '', seo_description: '', seo_keywords: '' };
}

async function loadBase() {
  const data = await adminCommunities();
  communities.value = data.items || [];
}

async function load() {
  const data = await tags(query);
  rows.value = data.items || [];
  await preloadAliases();
}

async function preloadAliases() {
  const next = { ...aliasCache.value };
  for (const row of rows.value) {
    if (!next[row.id]) {
      try {
        const data = await tagAliases(row.id);
        next[row.id] = data.items || [];
      } catch {
        next[row.id] = [];
      }
    }
  }
  aliasCache.value = next;
}

function reset() {
  Object.assign(query, { site: '', status: 'all', q: '' });
  load();
}

function openCreate() {
  Object.assign(form, blankTag(), { site: query.site || 'portal' });
  drawer.value = true;
}

function openEdit(row) {
  Object.assign(form, blankTag(), row, { sort_order: row.sort_order || row.sort || 0 });
  drawer.value = true;
}

async function save() {
  const payload = { ...form, sort: form.sort_order };
  form.id ? await updateTag(form.id, payload) : await createTag(payload);
  drawer.value = false;
  ElMessage.success('标签已保存');
  await load();
}

async function enable(row) {
  await enableTag(row.id);
  ElMessage.success('标签已启用');
  await load();
}

async function disable(row) {
  await disableTag(row.id);
  ElMessage.success('标签已禁用');
  await load();
}

async function recalculateOne(row) {
  await recalculateTag(row.id);
  ElMessage.success('标签统计已重算');
  await load();
}

async function recalculateAll() {
  await ElMessageBox.confirm('将同步重算全部标签统计，是否继续？', '批量重算', { type: 'warning' });
  const data = await recalculateAllTags();
  ElMessage.success(`已重算 ${data.updated || 0} 个标签`);
  await load();
}

function openFrontend(row) {
  const slug = row.slug || row.name;
  const communitySlug = row.community_slug || (row.site && row.site !== 'portal' ? row.site : '');
  const path = communitySlug ? `/c/${encodeURIComponent(communitySlug)}/tags/${encodeURIComponent(slug)}/` : `/tags/${encodeURIComponent(slug)}/`;
  window.open(path, '_blank');
}

async function openTopics(row) {
  currentTag.value = row;
  Object.assign(topicQuery, { page: 1, page_size: 10 });
  topicsDrawer.value = true;
  await loadTopics();
}

async function loadTopics() {
  if (!currentTag.value.id) return;
  const data = await tagTopics(currentTag.value.id, topicQuery);
  topicRows.value = data.items || [];
  topicTotal.value = data.total || topicRows.value.length;
}

async function openAliases(row) {
  currentTag.value = row;
  aliasForm.alias = '';
  aliasDrawer.value = true;
  await refreshAliases(row.id);
}

async function refreshAliases(tagID) {
  const data = await tagAliases(tagID);
  aliasRows.value = data.items || [];
  aliasCache.value = { ...aliasCache.value, [tagID]: aliasRows.value };
}

async function createAlias() {
  if (!currentTag.value.id) return;
  await createTagAlias(currentTag.value.id, { alias: aliasForm.alias });
  aliasForm.alias = '';
  ElMessage.success('标签别名已添加');
  await refreshAliases(currentTag.value.id);
  await load();
}

async function removeAlias(row) {
  await ElMessageBox.confirm(`确认删除别名 ${row.alias} 吗？`, '删除别名', { type: 'warning' });
  await deleteTagAlias(currentTag.value.id, row.id);
  ElMessage.success('标签别名已删除');
  await refreshAliases(currentTag.value.id);
  await load();
}

function openMerge(row) {
  currentTag.value = row;
  mergeCandidates.value = rows.value.filter((item) => item.id !== row.id && item.site === row.site && item.status === 'enable');
  Object.assign(mergeForm, { target_tag_id: 0, note: '' });
  mergeDialog.value = true;
}

async function submitMerge() {
  if (!mergeForm.target_tag_id) {
    ElMessage.error('请选择目标标签');
    return;
  }
  await ElMessageBox.confirm(`确认将 ${currentTag.value.name} 合并到目标标签吗？`, '合并确认', { type: 'warning' });
  await mergeTag(currentTag.value.id, mergeForm);
  mergeDialog.value = false;
  ElMessage.success('标签已合并');
  await load();
}

function aliasSummary(tagID) {
  const items = aliasCache.value[tagID] || [];
  if (!items.length) return '暂无别名';
  return items.slice(0, 3).map((item) => item.alias).join(' / ');
}

function communityName(row) {
  if (row.community_name) return row.community_name;
  if (!row.site || row.site === 'portal') return '总站';
  return communities.value.find((item) => item.slug === row.site)?.name || row.site;
}

function statusType(status) {
  if (status === 'enable') return 'success';
  if (status === 'merged') return 'warning';
  return 'info';
}

function statusLabel(status) {
  return { enable: '启用', disable: '禁用', merged: '已合并' }[status] || status || '未知';
}

loadBase().then(load);
</script>
