<template>
  <el-card shadow="never">
    <template #header><div class="card-head"><span>评论管理</span><el-button type="danger" @click="batchDialog = true">批量处理</el-button></div></template>
    <el-form :inline="true"><el-form-item><el-select v-model="status" style="width: 130px"><el-option label="全部" value="all" /><el-option label="正常" value="normal" /><el-option label="违规" value="illegal" /></el-select></el-form-item><el-form-item><el-input v-model="keyword" placeholder="内容 / 作者 / 标题" clearable /></el-form-item></el-form>
    <el-table :data="filtered" border stripe><el-table-column prop="id" label="ID" width="80" sortable /><el-table-column prop="author" label="作者" width="140" /><el-table-column prop="post_title" label="帖子" min-width="220" show-overflow-tooltip /><el-table-column prop="text" label="评论" min-width="260" show-overflow-tooltip /><el-table-column prop="status" label="状态" width="100"><template #default="{ row }"><el-tag>{{ row.status }}</el-tag></template></el-table-column><el-table-column label="操作" width="220"><template #default="{ row }"><el-button link @click="restore(row)">恢复</el-button><el-button link type="warning" @click="hide(row)">隐藏</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column></el-table>
  </el-card>
  <el-dialog v-model="batchDialog" title="批量处理"><el-alert title="可按当前筛选条件批量标记违规、恢复或删除。" type="info" :closable="false" /><template #footer><el-button @click="batchDialog = false">取消</el-button><el-button type="primary" @click="batchDialog = false">确认</el-button></template></el-dialog>
</template>

<script setup>
import { computed, ref } from 'vue';
import { comments, deleteComment, hideComment, restoreComment } from '@/api/admin';

const rows = ref([]);
const status = ref('all');
const keyword = ref('');
const batchDialog = ref(false);
const filtered = computed(() => rows.value.filter((row) => (status.value === 'all' || row.status === status.value) && `${row.text} ${row.author} ${row.post_title}`.toLowerCase().includes(keyword.value.toLowerCase())));
async function load() { rows.value = (await comments('portal')).items || []; }
async function hide(row) { await hideComment(row.id); await load(); }
async function restore(row) { await restoreComment(row.id); await load(); }
async function remove(row) { await deleteComment(row.id); await load(); }
load();
</script>
