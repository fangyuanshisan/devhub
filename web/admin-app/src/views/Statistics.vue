<template>
  <el-row :gutter="16">
    <el-col :xs="24" :lg="12"><el-card shadow="never"><template #header>子站数据</template><el-table :data="overview.site_stats || []" stripe><el-table-column prop="site" label="站点" /><el-table-column prop="posts" label="帖子" sortable /><el-table-column prop="views" label="浏览" sortable /><el-table-column prop="likes" label="点赞" sortable /></el-table></el-card></el-col>
    <el-col :xs="24" :lg="12"><el-card shadow="never"><template #header>板块数据</template><el-table :data="overview.board_stats || []" stripe><el-table-column prop="board" label="板块" /><el-table-column prop="posts" label="帖子" sortable /><el-table-column prop="comments" label="评论" sortable /></el-table></el-card></el-col>
  </el-row>
</template>

<script setup>
import { ref } from 'vue';
import { overview as overviewAPI } from '@/api/admin';
const overview = ref({});
overviewAPI('portal').then((data) => (overview.value = data));
</script>
