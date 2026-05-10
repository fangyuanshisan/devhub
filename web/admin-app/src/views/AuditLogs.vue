<template>
  <section class="filter-panel" data-testid="admin-audit-logs-page">
    <el-form :inline="true" :model="query" class="filter-form">
      <el-form-item label="类型">
        <el-select v-model="query.type" style="width: 130px">
          <el-option label="全部" value="all" />
          <el-option label="治理" value="audit" />
          <el-option label="运营" value="operation" />
          <el-option label="系统" value="system" />
          <el-option label="登录" value="login" />
        </el-select>
      </el-form-item>
      <el-form-item label="对象">
        <el-select v-model="query.target_type" clearable placeholder="全部" style="width: 130px">
          <el-option label="主题" value="topics" />
          <el-option label="评论" value="comments" />
          <el-option label="举报" value="reports" />
          <el-option label="版主" value="community_moderators" />
          <el-option label="用户" value="users" />
        </el-select>
      </el-form-item>
      <el-form-item label="操作人"><el-input v-model="query.actor" clearable placeholder="用户名" style="width: 130px" /></el-form-item>
      <el-form-item label="身份">
        <el-select v-model="query.actor_type" clearable placeholder="全部" style="width: 130px">
          <el-option label="后台人员" value="admin_user" />
          <el-option label="子站版主" value="moderator" />
          <el-option label="系统" value="system" />
        </el-select>
      </el-form-item>
      <el-form-item label="动作"><el-input v-model="query.action" clearable placeholder="动作关键词" style="width: 150px" data-testid="admin-audit-action-search" /></el-form-item>
      <el-form-item label="目标"><el-input v-model="query.target" clearable placeholder="topics#1" style="width: 150px" /></el-form-item>
      <el-form-item><el-button type="primary" @click="load">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
    </el-form>
  </section>

  <el-card shadow="never">
    <template #header><div class="card-head"><span>治理审计日志</span><el-tag>{{ total }} 条</el-tag></div></template>
    <el-table :data="rows" stripe border>
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column label="操作人" width="170"><template #default="{ row }">{{ row.actor || '-' }}<div class="content-meta">{{ actorTypeName(row.actor_type) }} / ID {{ row.actor_id || row.actor_user_id || '-' }}</div></template></el-table-column>
      <el-table-column label="类型" width="110"><template #default="{ row }"><el-tag :type="typeTag(row.type)">{{ typeName(row.type) }}</el-tag></template></el-table-column>
      <el-table-column prop="action" label="动作" min-width="160" />
      <el-table-column label="目标" min-width="190"><template #default="{ row }"><span>{{ row.target || '-' }}</span><div class="content-meta">{{ targetInfo(row) }}</div></template></el-table-column>
      <el-table-column label="站点" width="120"><template #default="{ row }">{{ row.site || '-' }}<div class="content-meta">CID {{ row.community_id || '-' }}</div></template></el-table-column>
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column prop="created_at" label="时间" width="170" />
    </el-table>
    <el-pagination v-model:current-page="query.page" v-model:page-size="query.page_size" class="pager" layout="total, sizes, prev, pager, next, jumper" :total="total" @change="load" />
  </el-card>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { auditLogs } from '@/api/admin';

const query = reactive({ site: 'portal', type: 'all', target_type: '', actor_type: '', actor: '', action: '', target: '', page: 1, page_size: 20 });
const rows = ref([]);
const total = ref(0);

function typeName(type) {
  return { audit: '治理', operation: '运营', system: '系统', login: '登录' }[type] || type || '日志';
}
function typeTag(type) {
  return { audit: 'warning', operation: 'primary', system: 'info', login: 'success' }[type] || '';
}
function actorTypeName(type) {
  return { admin_user: '后台人员', moderator: '子站版主', system: '系统' }[type] || type || '未知身份';
}
function targetInfo(row) {
  if (row.target_type || row.target_id) return `target_type=${row.target_type || '-'} / target_id=${row.target_id || '-'}`;
  return 'target_type / target_id：未结构化';
}
async function load() {
  const data = await auditLogs(query);
  rows.value = data.items || [];
  total.value = data.total || 0;
}
function reset() {
  Object.assign(query, { site: 'portal', type: 'all', target_type: '', actor_type: '', actor: '', action: '', target: '', page: 1, page_size: 20 });
  load();
}
load();
</script>
