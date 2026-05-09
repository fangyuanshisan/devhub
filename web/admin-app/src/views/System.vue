<template>
  <el-row :gutter="16">
    <el-col :xs="24" :lg="10"><el-card shadow="never"><template #header>系统设置</template><el-form :model="form" label-width="110px"><el-form-item label="网站名称"><el-input v-model="form.site_name" /></el-form-item><el-form-item label="版权信息"><el-input v-model="form.copyright" /></el-form-item><el-form-item label="默认条数"><el-input-number v-model="form.default_page_size" /></el-form-item><el-form-item><el-button type="primary" @click="save">保存</el-button></el-form-item></el-form></el-card></el-col>
    <el-col :xs="24" :lg="14"><el-card shadow="never"><template #header>操作日志</template><el-table :data="logList" stripe><el-table-column prop="site" label="站点" width="100" /><el-table-column prop="actor" label="操作人" width="120" /><el-table-column prop="role" label="角色" width="120" /><el-table-column prop="action" label="动作" /><el-table-column prop="target" label="对象" /><el-table-column prop="created_at" label="时间" width="170" /></el-table></el-card></el-col>
  </el-row>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { logs, settings, updateSettings } from '@/api/admin';
const form = reactive({});
const logList = ref([]);
async function load() { Object.assign(form, await settings()); logList.value = (await logs('portal')).items || []; }
async function save() { await updateSettings(form); await load(); }
load();
</script>
