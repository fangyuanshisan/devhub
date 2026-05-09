<template>
  <div>
    <el-row :gutter="16">
      <el-col v-for="card in cards" :key="card.label" :xs="24" :sm="12" :lg="6">
        <el-card shadow="never" class="stat-card"><span>{{ card.label }}</span><strong>{{ card.value }}</strong><p>{{ card.hint }}</p></el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" class="mt">
      <el-col :xs="24" :lg="15">
        <el-card shadow="never"><template #header>热门内容</template><el-table :data="overview.top_posts || []" stripe><el-table-column prop="title" label="标题" min-width="240" /><el-table-column prop="site" label="站点" width="100" /><el-table-column prop="views" label="浏览" sortable width="100" /><el-table-column prop="likes" label="点赞" sortable width="100" /></el-table></el-card>
      </el-col>
      <el-col :xs="24" :lg="9">
        <el-card shadow="never"><template #header>状态分布</template><el-table :data="statusRows" size="small"><el-table-column prop="name" label="状态" /><el-table-column prop="value" label="数量" sortable /></el-table></el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { overview as overviewAPI } from '@/api/admin';

const overview = ref({});
const cards = computed(() => {
  const stats = overview.value.stats || {};
  const users = overview.value.user_stats || {};
  return [
    { label: '内容总量', value: stats.total_posts || 0, hint: '全部内容' },
    { label: '评论总量', value: stats.total_comments || 0, hint: '含多级回复' },
    { label: '点赞总量', value: stats.total_likes || 0, hint: '互动指标' },
    { label: '用户总量', value: users.total_users || 0, hint: `活跃 ${users.active_users || 0}` },
  ];
});
const statusRows = computed(() => Object.entries(overview.value.status_distribution || {}).map(([name, value]) => ({ name, value })));

onMounted(async () => {
  overview.value = await overviewAPI('portal');
});
</script>
