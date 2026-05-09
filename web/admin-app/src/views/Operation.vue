<template>
  <el-row :gutter="16">
    <el-col :xs="24" :lg="12"><el-card shadow="never"><template #header>通知推送</template><el-form ref="formRef" :model="form" :rules="rules" label-width="82px"><el-form-item label="范围" prop="scope"><el-select v-model="form.scope"><el-option label="全局" value="portal" /><el-option v-for="s in siteList.filter((s) => s.key !== 'portal')" :key="s.key" :label="s.name" :value="s.key" /></el-select></el-form-item><el-form-item label="标题" prop="title"><el-input v-model="form.title" /></el-form-item><el-form-item label="内容" prop="content"><el-input v-model="form.content" type="textarea" /></el-form-item><el-form-item><el-button type="primary" @click="submit">发送</el-button></el-form-item></el-form></el-card></el-col>
    <el-col :xs="24" :lg="12"><el-card shadow="never"><template #header>通知记录</template><el-table :data="noticeList" stripe><el-table-column prop="site" label="站点" width="100" /><el-table-column prop="title" label="标题" /><el-table-column prop="read" label="状态" width="100" /></el-table></el-card></el-col>
  </el-row>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { notices, pushNotice, sites } from '@/api/admin';
const formRef = ref();
const form = reactive({ scope: 'portal', title: '', content: '' });
const siteList = ref([]);
const noticeList = ref([]);
const rules = { scope: [{ required: true }], title: [{ required: true, message: '请输入标题' }], content: [{ required: true, message: '请输入内容' }] };
async function load() { siteList.value = (await sites()).items || []; noticeList.value = (await notices('portal')).items || []; }
async function submit() { await formRef.value.validate(); await pushNotice(form); Object.assign(form, { scope: 'portal', title: '', content: '' }); await load(); }
load();
</script>
