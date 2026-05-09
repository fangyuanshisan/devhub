<template>
  <el-row :gutter="16">
    <el-col :xs="24" :lg="10"><el-card shadow="never"><template #header>系统设置</template><el-form :model="form" label-width="110px"><el-form-item label="网站名称"><el-input v-model="form.site_name" /></el-form-item><el-form-item label="版权信息"><el-input v-model="form.copyright" /></el-form-item><el-form-item label="默认条数"><el-input-number v-model="form.default_page_size" /></el-form-item><el-form-item><el-button type="primary" @click="save">保存</el-button></el-form-item></el-form></el-card></el-col>
    <el-col :xs="24" :lg="14">
      <el-card shadow="never">
        <template #header><div class="card-head"><span>治理审计</span><el-tag>{{ logTotal }} 条</el-tag></div></template>
        <el-form :inline="true" :model="logQuery"><el-form-item><el-select v-model="logQuery.type" style="width: 130px"><el-option label="全部" value="all" /><el-option label="治理" value="audit" /><el-option label="运营" value="operation" /><el-option label="系统" value="system" /><el-option label="登录" value="login" /></el-select></el-form-item><el-form-item><el-input v-model="logQuery.actor" placeholder="操作人" clearable /></el-form-item><el-form-item><el-input v-model="logQuery.target" placeholder="对象" clearable /></el-form-item><el-form-item><el-button @click="loadLogs">筛选</el-button></el-form-item></el-form>
        <el-table :data="logList" stripe><el-table-column prop="site" label="站点" width="100" /><el-table-column prop="actor" label="操作人" width="120" /><el-table-column prop="role" label="角色" width="120" /><el-table-column prop="type" label="类型" width="90" /><el-table-column prop="action" label="动作" /><el-table-column prop="target" label="对象" /><el-table-column prop="created_at" label="时间" width="170" /></el-table>
        <el-pagination v-model:current-page="logQuery.page" v-model:page-size="logQuery.page_size" class="pager" layout="total, prev, pager, next" :total="logTotal" @change="loadLogs" />
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { auditLogs, settings, updateSettings } from '@/api/admin';
const form = reactive({});
const logQuery = reactive({ site: 'portal', type: 'all', actor: '', target: '', page: 1, page_size: 10 });
const logList = ref([]);
const logTotal = ref(0);
async function loadLogs() {
  const data = await auditLogs(logQuery);
  logList.value = data.items || [];
  logTotal.value = data.total || 0;
}
async function load() { Object.assign(form, await settings()); await loadLogs(); }
async function save() { await updateSettings(form); await load(); }
load();
</script>
