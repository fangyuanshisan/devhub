<template>
  <section class="filter-panel">
    <el-form :inline="true" :model="query">
      <el-form-item label="状态">
        <el-select v-model="query.status" style="width: 150px">
          <el-option label="全部" value="all" />
          <el-option label="待处理" value="pending" />
          <el-option label="已接受" value="accepted" />
          <el-option label="已驳回" value="rejected" />
        </el-select>
      </el-form-item>
      <el-form-item label="对象">
        <el-select v-model="query.target_type" style="width: 150px">
          <el-option label="全部" value="all" />
          <el-option label="主题" value="topic" />
          <el-option label="评论" value="comment" />
          <el-option label="用户" value="user" />
          <el-option label="Wiki" value="wiki" />
        </el-select>
      </el-form-item>
      <el-form-item><el-button type="primary" @click="load">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
    </el-form>
  </section>

  <el-card shadow="never">
    <template #header><div class="card-head"><span>举报管理</span><el-tag type="warning">{{ total }} 条</el-tag></div></template>
    <el-table :data="rows" stripe border>
      <el-table-column prop="id" label="ID" width="80" sortable />
      <el-table-column label="对象" width="110"><template #default="{ row }"><el-tag>{{ typeName(row.target_type) }}</el-tag></template></el-table-column>
      <el-table-column label="内容" min-width="260" show-overflow-tooltip>
        <template #default="{ row }"><el-link :href="row.target_url || '/'" target="_blank" type="primary">{{ row.target_title || `${row.target_type}#${row.target_id}` }}</el-link></template>
      </el-table-column>
      <el-table-column prop="reason_type" label="原因" width="120" />
      <el-table-column prop="reason_text" label="说明" min-width="180" show-overflow-tooltip />
      <el-table-column label="子站" width="120"><template #default="{ row }">{{ row.community_name || row.community_slug || '-' }}</template></el-table-column>
      <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusName(row.status) }}</el-tag></template></el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="170" />
      <el-table-column label="操作" fixed="right" width="180">
        <template #default="{ row }">
          <el-button link type="primary" :disabled="row.status !== 'pending'" @click="openHandle(row, 'accepted')">接受</el-button>
          <el-button link :disabled="row.status !== 'pending'" @click="openHandle(row, 'rejected')">驳回</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="query.page" v-model:page-size="query.page_size" class="pager" layout="total, sizes, prev, pager, next, jumper" :total="total" @change="load" />
  </el-card>

  <el-dialog v-model="dialog" :title="handleForm.status === 'accepted' ? '接受举报' : '驳回举报'" width="420px">
    <el-input v-model="handleForm.handle_note" type="textarea" maxlength="500" show-word-limit placeholder="处理备注" />
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" @click="submitHandle">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { handleReport, reports } from '@/api/admin';

const query = reactive({ status: 'pending', target_type: 'all', page: 1, page_size: 10 });
const rows = ref([]);
const total = ref(0);
const dialog = ref(false);
const current = ref(null);
const handleForm = reactive({ status: 'accepted', handle_note: '' });

function typeName(type) { return { topic: '主题', comment: '评论', user: '用户', wiki: 'Wiki' }[type] || type; }
function statusName(status) { return { pending: '待处理', accepted: '已接受', rejected: '已驳回' }[status] || status; }
function statusType(status) { return { pending: 'warning', accepted: 'success', rejected: 'info' }[status] || ''; }
async function load() {
  const data = await reports(query);
  rows.value = data.items || [];
  total.value = data.total || 0;
}
function reset() {
  Object.assign(query, { status: 'pending', target_type: 'all', page: 1, page_size: 10 });
  load();
}
function openHandle(row, status) {
  current.value = row;
  Object.assign(handleForm, { status, handle_note: status === 'accepted' ? '确认违规，已隐藏目标内容' : '未确认违规' });
  dialog.value = true;
}
async function submitHandle() {
  if (!current.value) return;
  await handleReport(current.value.id, handleForm);
  dialog.value = false;
  await load();
}
load();
</script>
