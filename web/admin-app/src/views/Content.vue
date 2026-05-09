<template>
  <section class="filter-panel">
    <el-form :inline="true" :model="query" class="filter-form">
      <el-form-item label="内容站点">
        <el-select v-model="query.site" style="width: 220px"><el-option label="全部站点" value="portal" /><el-option v-for="s in siteList" :key="s.key" :label="s.name" :value="s.key" /></el-select>
      </el-form-item>
      <el-form-item label="内容板块">
        <el-select v-model="query.board" style="width: 220px"><el-option label="全部" value="all" /><el-option v-for="b in boardList" :key="b.key" :label="b.name" :value="b.key" /></el-select>
      </el-form-item>
      <el-form-item label="更新时间"><el-date-picker v-model="dateRange" type="daterange" start-placeholder="开始日期" end-placeholder="结束日期" /></el-form-item>
      <el-form-item label="内容搜索"><el-select v-model="searchField" style="width: 90px"><el-option label="全部" value="all" /><el-option label="标题" value="title" /><el-option label="作者" value="author" /><el-option label="标签" value="tag" /></el-select></el-form-item>
      <el-form-item><el-input v-model="query.q" placeholder="请输入" clearable style="width: 160px" /></el-form-item>
      <el-form-item><el-button type="primary" @click="load">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
    </el-form>
  </section>

  <el-card shadow="never">
    <el-tabs v-model="query.status" class="status-tabs" @tab-change="load">
      <el-tab-pane label="全部" name="all" />
      <el-tab-pane label="草稿" name="draft" />
      <el-tab-pane label="待审核" name="review" />
      <el-tab-pane label="已发布" name="publish" />
      <el-tab-pane label="已下架" name="offline" />
      <el-tab-pane label="已驳回" name="rejected" />
      <el-tab-pane label="置顶内容" name="pinned" />
      <el-tab-pane label="推荐内容" name="recommended" />
    </el-tabs>
    <div class="batch-actions">
      <el-button type="primary" @click="openCreate">新增内容</el-button>
      <el-button @click="batchDrawer = true">批量审核</el-button>
      <el-button @click="batchDrawer = true">批量下架</el-button>
      <el-button @click="batchDrawer = true">内容导出</el-button>
    </div>
    <el-table :data="rows" stripe @sort-change="sortChange">
      <el-table-column type="expand" width="44"><template #default="{ row }"><div class="content-meta">摘要：{{ row.summary || '暂无摘要' }}</div></template></el-table-column>
      <el-table-column type="selection" width="46" />
      <el-table-column prop="id" label="内容ID | 类型" width="190" sortable="custom">
        <template #default="{ row }"><div>DH{{ row.id }}</div><el-link type="primary">[{{ typeName(row.board) }}]</el-link></template>
      </el-table-column>
      <el-table-column label="内容信息" min-width="320">
        <template #default="{ row }">
          <div class="content-cell"><div class="content-thumb"></div><div><div class="content-title">{{ row.title }}</div><div class="content-meta">{{ row.content || row.summary || '内容摘要...' }}</div></div></div>
        </template>
      </el-table-column>
      <el-table-column label="作者信息" min-width="170"><template #default="{ row }">{{ row.author }} | UID {{ row.id + 100000 }}</template></el-table-column>
      <el-table-column prop="views" label="浏览" width="110" sortable="custom"><template #default="{ row }">{{ row.views || 0 }}</template></el-table-column>
      <el-table-column label="互动" width="130"><template #default="{ row }">{{ row.likes || 0 }} 赞</template></el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="170"><template #default="{ row }">{{ row.updated_at || '--' }}</template></el-table-column>
      <el-table-column prop="status" label="内容状态" width="120"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusName(row.status) }}</el-tag></template></el-table-column>
      <el-table-column label="操作" fixed="right" width="150">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">详情</el-button>
          <el-dropdown>
            <el-button link type="primary">更多<el-icon><ArrowDown /></el-icon></el-button>
            <template #dropdown><el-dropdown-menu>
              <el-dropdown-item @click="toggleFeature(row)">{{ row.recommended ? '取消精华' : '设为精华' }}</el-dropdown-item>
              <el-dropdown-item @click="togglePin(row)">{{ row.pinned ? '取消置顶' : '设为置顶' }}</el-dropdown-item>
              <el-dropdown-item @click="toggleLock(row)">{{ row.comment_locked ? '解锁评论' : '锁定评论' }}</el-dropdown-item>
              <el-dropdown-item @click="toggleVisible(row)">{{ row.status === 'offline' ? '恢复内容' : '隐藏内容' }}</el-dropdown-item>
              <el-dropdown-item @click="openEdit(row)">编辑</el-dropdown-item>
              <el-dropdown-item @click="remove(row)">删除</el-dropdown-item>
            </el-dropdown-menu></template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="query.page" v-model:page-size="query.page_size" class="pager" layout="total, sizes, prev, pager, next, jumper" :total="total" @change="load" />
  </el-card>

  <el-drawer v-model="drawer" :title="editing.id ? '编辑内容' : '新增内容'" size="58%">
    <el-steps :active="step" finish-status="success" simple class="mb"><el-step title="基础信息" /><el-step title="正文编辑" /><el-step title="发布设置" /></el-steps>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="88px">
      <template v-if="step === 0">
        <el-form-item label="站点" prop="site"><el-select v-model="form.site"><el-option v-for="s in siteList.filter((s) => s.key !== 'portal')" :key="s.key" :label="s.name" :value="s.key" /></el-select></el-form-item>
        <el-form-item label="板块" prop="board"><el-select v-model="form.board"><el-option v-for="b in boardList.filter((b) => b.key !== 'all')" :key="b.key" :label="b.name" :value="b.key" /></el-select></el-form-item>
        <el-form-item label="标题" prop="title"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="摘要"><el-input v-model="form.summary" type="textarea" /></el-form-item>
      </template>
      <template v-else-if="step === 1">
        <el-form-item label="正文" prop="content"><MarkdownEditor v-model="form.content" /></el-form-item>
        <el-form-item label="附件"><el-upload action="#" :auto-upload="false" drag><el-icon><UploadFilled /></el-icon><div>拖拽文件到此处或点击上传</div></el-upload></el-form-item>
      </template>
      <template v-else>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio-button label="draft">草稿</el-radio-button><el-radio-button label="review">审核</el-radio-button><el-radio-button label="publish">发布</el-radio-button></el-radio-group></el-form-item>
        <el-form-item label="标签"><el-input v-model="tagText" placeholder="多个标签用逗号分隔" /></el-form-item>
        <el-form-item label="运营位"><el-checkbox v-model="form.pinned">置顶</el-checkbox><el-checkbox v-model="form.recommended">推荐</el-checkbox></el-form-item>
      </template>
    </el-form>
    <template #footer><el-button @click="step = Math.max(0, step - 1)">上一步</el-button><el-button v-if="step < 2" type="primary" @click="next">下一步</el-button><el-button v-else type="primary" @click="save">保存</el-button></template>
  </el-drawer>

  <el-drawer v-model="batchDrawer" title="批量操作" size="360px">
    <el-alert title="这里可扩展批量审核、推荐、下架、迁移板块等动作。" type="info" :closable="false" />
  </el-drawer>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import { ArrowDown, UploadFilled } from '@element-plus/icons-vue';
import MarkdownEditor from '@/components/MarkdownEditor.vue';
import { boards, createPost, deletePost, featureTopic, hideTopic, lockTopicComments, pinTopic, posts, restoreTopic, sites, unlockTopicComments, updatePost } from '@/api/admin';

const query = reactive({ site: 'portal', board: 'all', status: 'all', q: '', page: 1, page_size: 10 });
const dateRange = ref([]);
const searchField = ref('all');
const rows = ref([]);
const total = ref(0);
const siteList = ref([]);
const boardList = ref([]);
const drawer = ref(false);
const batchDrawer = ref(false);
const step = ref(0);
const formRef = ref();
const editing = ref({});
const tagText = ref('');
const form = reactive({ site: 'php', board: 'community', title: '', summary: '', content: '', status: 'publish', pinned: false, recommended: false });
const rules = { site: [{ required: true, message: '请选择站点' }], board: [{ required: true, message: '请选择板块' }], title: [{ required: true, message: '请输入标题' }], content: [{ required: true, message: '请输入正文' }] };

async function loadBase() {
  siteList.value = (await sites()).items || [];
  boardList.value = (await boards()).items || [];
}
async function load() {
  const params = { ...query, status: ['pinned', 'recommended'].includes(query.status) ? 'all' : query.status };
  const data = await posts(params);
  rows.value = data.items || [];
  total.value = data.total || rows.value.length;
}
function reset() {
  Object.assign(query, { site: 'portal', board: 'all', status: 'all', q: '', page: 1, page_size: 10 });
  dateRange.value = [];
  load();
}
function statusName(status) {
  return { draft: '草稿', review: '待审核', publish: '已发布', offline: '已隐藏', hidden: '已隐藏', rejected: '已驳回' }[status] || status || '草稿';
}
function statusType(status) {
  return { publish: 'success', offline: 'info', rejected: 'danger', review: 'warning' }[status] || '';
}
function typeName(board) {
  return { community: '社区帖', qa: '问答', docs: '文档', wiki: 'Wiki', opensource: '开源项目', ai: 'AI 作品', jobs: '招聘内推' }[board] || '内容';
}
function sortChange({ prop, order }) {
  if (!prop || !order) return load();
  rows.value = [...rows.value].sort((a, b) => (order === 'ascending' ? a[prop] - b[prop] : b[prop] - a[prop]));
}
function openCreate() {
  Object.assign(form, { site: 'php', board: 'community', title: '', summary: '', content: '', status: 'publish', pinned: false, recommended: false });
  editing.value = {};
  tagText.value = '';
  step.value = 0;
  drawer.value = true;
}
function openEdit(row) {
  Object.assign(form, row);
  editing.value = row;
  tagText.value = (row.tags || []).join(', ');
  step.value = 0;
  drawer.value = true;
}
async function next() {
  await formRef.value.validate();
  step.value += 1;
}
async function save() {
  await formRef.value.validate();
  const payload = { ...form, tags: tagText.value.split(',').map((tag) => tag.trim()).filter(Boolean) };
  editing.value.id ? await updatePost(editing.value.id, payload) : await createPost(payload);
  drawer.value = false;
  await load();
}
async function remove(row) {
  await ElMessageBox.confirm(`确定删除「${row.title}」？`, '确认删除', { type: 'warning' });
  await deletePost(row.id);
  await load();
}
async function toggleFeature(row) { await featureTopic(row.id); await load(); }
async function togglePin(row) { await pinTopic(row.id); await load(); }
async function toggleLock(row) { row.comment_locked ? await unlockTopicComments(row.id) : await lockTopicComments(row.id); await load(); }
async function toggleVisible(row) { row.status === 'offline' ? await restoreTopic(row.id) : await hideTopic(row.id); await load(); }
loadBase().then(load);
</script>
