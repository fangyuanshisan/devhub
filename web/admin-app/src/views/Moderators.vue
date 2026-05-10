<template>
  <section class="filter-panel" data-testid="admin-moderators-page">
    <el-form :inline="true" :model="query">
      <el-form-item label="子站">
        <el-select v-model="query.community_slug" style="width: 180px" data-testid="admin-moderator-community-filter">
          <el-option label="全部子站" value="all" />
          <el-option v-for="site in siteList.filter((item) => item.key !== 'portal')" :key="site.key" :label="site.name" :value="site.key" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" style="width: 130px" data-testid="admin-moderator-status-filter"><el-option label="全部" value="all" /><el-option label="启用" value="1" /><el-option label="停用" value="0" /></el-select>
      </el-form-item>
      <el-form-item><el-button type="primary" @click="load">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
    </el-form>
  </section>

  <el-card shadow="never">
    <template #header><div class="card-head"><span>版主管理</span><el-button type="primary" @click="openCreate">新增版主</el-button></div></template>
    <el-table :data="rows" stripe border data-testid="admin-moderators-table">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column label="子站" min-width="150"><template #default="{ row }">{{ row.community_name || row.community_slug || row.community_id }}</template></el-table-column>
      <el-table-column label="用户" min-width="180"><template #default="{ row }">{{ row.user_nickname || row.user_name || row.user_id }} <span class="content-meta">UID {{ row.user_id }}</span></template></el-table-column>
      <el-table-column label="角色" width="120"><template #default="{ row }"><el-tag>{{ roleName(row.role) }}</el-tag></template></el-table-column>
      <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="170" />
      <el-table-column label="操作" fixed="right" width="170">
        <template #default="{ row }"><el-button link type="primary" @click="openEdit(row)">编辑</el-button><el-button link type="danger" :disabled="row.status === 0" @click="disable(row)">停用</el-button></template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="query.page" v-model:page-size="query.page_size" class="pager" layout="total, sizes, prev, pager, next, jumper" :total="total" @change="load" />
  </el-card>

  <el-dialog v-model="dialog" :title="editing.id ? '编辑版主' : '新增版主'" width="460px" data-testid="admin-moderator-dialog">
    <el-form :model="form" label-width="92px">
      <el-form-item label="子站">
        <el-select v-model="form.community_slug" :disabled="Boolean(editing.id)">
          <el-option v-for="site in siteList.filter((item) => item.key !== 'portal')" :key="site.key" :label="site.name" :value="site.key" />
        </el-select>
      </el-form-item>
      <el-form-item label="用户">
        <el-select v-model="form.user_id" filterable :disabled="Boolean(editing.id)">
          <el-option v-for="user in userList" :key="user.id" :label="`${user.nickname || user.username} / UID ${user.id}`" :value="user.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="角色"><el-radio-group v-model="form.role"><el-radio-button label="moderator">版主</el-radio-button><el-radio-button label="owner">站长</el-radio-button></el-radio-group></el-form-item>
      <el-form-item label="状态"><el-switch v-model="enabled" active-text="启用" inactive-text="停用" /></el-form-item>
    </el-form>
    <template #footer><el-button @click="dialog = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
  </el-dialog>
</template>

<script setup>
import { computed, reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import { createModerator, deleteModerator, moderators, sites, updateModerator, users } from '@/api/admin';

const query = reactive({ community_slug: 'all', status: 'all', page: 1, page_size: 10 });
const rows = ref([]);
const total = ref(0);
const siteList = ref([]);
const userList = ref([]);
const dialog = ref(false);
const editing = ref({});
const form = reactive({ community_slug: 'php', user_id: 1, role: 'moderator', status: 1 });
const enabled = computed({ get: () => form.status === 1, set: (value) => { form.status = value ? 1 : 0; } });

function roleName(role) { return { moderator: '版主', owner: '站长' }[role] || role; }
async function loadBase() {
  siteList.value = (await sites()).items || [];
  userList.value = (await users()).items || [];
}
async function load() {
  const data = await moderators(query);
  rows.value = data.items || [];
  total.value = data.total || 0;
}
function reset() {
  Object.assign(query, { community_slug: 'all', status: 'all', page: 1, page_size: 10 });
  load();
}
function openCreate() {
  editing.value = {};
  Object.assign(form, { community_slug: 'php', user_id: userList.value[0]?.id || 1, role: 'moderator', status: 1 });
  dialog.value = true;
}
function openEdit(row) {
  editing.value = row;
  Object.assign(form, { community_slug: row.community_slug, user_id: row.user_id, role: row.role, status: row.status });
  dialog.value = true;
}
async function save() {
  const payload = { ...form };
  editing.value.id ? await updateModerator(editing.value.id, payload) : await createModerator(payload);
  dialog.value = false;
  await load();
}
async function disable(row) {
  await ElMessageBox.confirm(`确定停用 ${row.user_nickname || row.user_name || row.user_id} 的版主权限？`, '停用版主', { type: 'warning' });
  await deleteModerator(row.id);
  await load();
}
loadBase().then(load);
</script>
