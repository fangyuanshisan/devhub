<template>
  <el-card shadow="never">
    <template #header><div class="card-head"><span>评论管理</span><el-button type="danger" :disabled="!selected.length" @click="batchDialog = true">批量处理</el-button></div></template>
    <el-form :inline="true"><el-form-item><el-select v-model="status" style="width: 130px"><el-option label="全部" value="all" /><el-option label="正常" value="normal" /><el-option label="违规" value="illegal" /></el-select></el-form-item><el-form-item><el-input v-model="keyword" placeholder="内容 / 作者 / 标题" clearable /></el-form-item></el-form>
    <el-table :data="filtered" border stripe @selection-change="selected = $event"><el-table-column type="selection" width="46" /><el-table-column prop="id" label="ID" width="80" sortable /><el-table-column prop="author" label="作者" width="140" /><el-table-column prop="post_title" label="帖子" min-width="220" show-overflow-tooltip /><el-table-column prop="text" label="评论" min-width="260" show-overflow-tooltip /><el-table-column prop="status" label="状态" width="100"><template #default="{ row }"><el-tag>{{ row.status }}</el-tag></template></el-table-column><el-table-column label="操作" width="220"><template #default="{ row }"><el-button link @click="restore(row)">恢复</el-button><el-button link type="warning" @click="hide(row)">隐藏</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column></el-table>
  </el-card>
  <el-dialog v-model="batchDialog" title="批量处理" width="420px">
    <el-alert :title="`已选择 ${selected.length} 条评论`" type="info" :closable="false" />
    <el-form label-width="88px" class="mt"><el-form-item label="治理动作"><el-select v-model="batchAction"><el-option label="隐藏评论" value="hide" /><el-option label="恢复评论" value="restore" /><el-option label="删除评论" value="delete" /></el-select></el-form-item><el-form-item label="处理备注"><el-input v-model="batchNote" type="textarea" maxlength="200" show-word-limit placeholder="可选，用于治理审计记录" /></el-form-item></el-form>
    <template #footer><el-button @click="batchDialog = false">取消</el-button><el-button type="primary" :disabled="!selected.length" @click="submitBatch">确认</el-button></template>
  </el-dialog>
</template>

<script setup>
import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { batchComments, comments, deleteComment, hideComment, restoreComment } from '@/api/admin';

const rows = ref([]);
const selected = ref([]);
const status = ref('all');
const keyword = ref('');
const batchDialog = ref(false);
const batchAction = ref('hide');
const batchNote = ref('');
const filtered = computed(() => rows.value.filter((row) => (status.value === 'all' || row.status === status.value) && `${row.text} ${row.author} ${row.post_title}`.toLowerCase().includes(keyword.value.toLowerCase())));
async function load() { rows.value = (await comments('portal')).items || []; }
async function hide(row) { await hideComment(row.id); await load(); }
async function restore(row) { await restoreComment(row.id); await load(); }
async function remove(row) { await deleteComment(row.id); await load(); }
async function submitBatch() {
  if (!selected.value.length) {
    ElMessage.warning('请先选择评论');
    return;
  }
  const data = await batchComments({ ids: selected.value.map((row) => row.id), action: batchAction.value, note: batchNote.value });
  batchDialog.value = false;
  if (data.failed) ElMessage.warning(`已处理 ${data.updated || 0} 条，失败 ${data.failed || 0} 条`);
  else ElMessage.success(`已处理 ${data.updated || 0} 条`);
  selected.value = [];
  await load();
}
load();
</script>
