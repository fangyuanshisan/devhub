<template>
  <el-row :gutter="16">
    <el-col :xs="24" :lg="10"><el-card shadow="never"><template #header>站点配置</template><el-table :data="siteList" stripe><el-table-column prop="key" label="Key" /><el-table-column prop="name" label="名称" /><el-table-column prop="status" label="状态" /></el-table></el-card></el-col>
    <el-col :xs="24" :lg="14"><el-card shadow="never"><template #header>板块配置</template><el-table :data="boardList" stripe><el-table-column prop="key" label="Key" /><el-table-column prop="name" label="名称" /><el-table-column prop="site" label="站点" /><el-table-column prop="visible" label="显示" /></el-table></el-card></el-col>
  </el-row>
</template>

<script setup>
import { ref } from 'vue';
import { boards, sites } from '@/api/admin';
const siteList = ref([]);
const boardList = ref([]);
Promise.all([sites(), boards()]).then(([s, b]) => {
  siteList.value = s.items || [];
  boardList.value = b.items || [];
});
</script>
