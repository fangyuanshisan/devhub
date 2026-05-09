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
          <el-option label="全部" value="all" />
          <el-option label="启用" value="enable" />
          <el-option label="禁用" value="disable" />
        </el-select>
      </el-form-item>
      <el-form-item label="搜索">
        <el-input v-model="query.q" clearable placeholder="名称 / slug / 描述" style="width: 220px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
        <el-button type="primary" plain @click="openCreate">新增标签</el-button>
      </el-form-item>
    </el-form>
  </section>

  <el-card shadow="never">
    <template #header>
      <div class="card-head">
        <span>标签管理</span>
        <el-button type="primary" @click="openCreate">新增标签</el-button>
      </div>
    </template>
    <el-table :data="rows" stripe border>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column label="标签" min-width="240">
        <template #default="{ row }">
          <div class="content-cell">
            <div class="content-thumb">#</div>
            <div>
              <div class="content-title">{{ row.name }} <span class="content-meta">/{{ row.slug }}</span></div>
              <div class="content-meta">{{ row.description || '暂无说明' }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="子站" width="130">
        <template #default="{ row }">{{ communityName(row) }}</template>
      </el-table-column>
      <el-table-column label="统计" width="170">
        <template #default="{ row }">
          <span class="content-meta">内容 {{ row.topic_count || row.use_count || 0 }}</span>
          <span class="content-meta">关注 {{ row.follower_count || 0 }}</span>
        </template>
      </el-table-column>
      <el-table-column label="SEO" min-width="220">
        <template #default="{ row }">
          <div class="content-meta">{{ row.seo_title || '未设置 SEO title' }}</div>
          <div class="content-meta">{{ row.seo_description || '未设置 SEO description' }}</div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }"><el-tag :type="row.status === 'enable' ? 'success' : 'info'">{{ row.status === 'enable' ? '启用' : '禁用' }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="260">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="openTopics(row)">关联内容</el-button>
          <el-button link type="primary" @click="openFrontend(row)">前台</el-button>
          <el-button v-if="row.status !== 'enable'" link type="success" @click="enable(row)">启用</el-button>
          <el-button v-else link type="warning" @click="disable(row)">禁用</el-button>
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
      <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio-button label="enable">启用</el-radio-button><el-radio-button label="disable">禁用</el-radio-button></el-radio-group></el-form-item>
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
</template>

<script setup>
import { reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { adminCommunities, createTag, disableTag, enableTag, tagTopics, tags, updateTag } from '@/api/admin';

const query = reactive({ site: '', status: 'all', q: '' });
const topicQuery = reactive({ page: 1, page_size: 10 });
const rows = ref([]);
const communities = ref([]);
const drawer = ref(false);
const topicsDrawer = ref(false);
const currentTag = ref({});
const topicRows = ref([]);
const topicTotal = ref(0);
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

function openFrontend(row) {
  const slug = row.slug || row.name;
  const queryString = row.community_slug || (row.site && row.site !== 'portal' ? row.site : '');
  window.open(`/tags/${encodeURIComponent(slug)}/${queryString ? `?community_slug=${encodeURIComponent(queryString)}` : ''}`, '_blank');
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

function communityName(row) {
  if (row.community_name) return row.community_name;
  if (!row.site || row.site === 'portal') return '总站';
  return communities.value.find((item) => item.slug === row.site)?.name || row.site;
}

loadBase().then(load);
</script>
