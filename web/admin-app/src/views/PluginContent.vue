<template>
  <section class="page-card">
    <div class="toolbar">
      <div>
        <h2>{{ route.meta.title }}</h2>
        <p>插件内容通过 Core 内容表兼容展示，当前筛选：{{ route.meta.contentType }}</p>
      </div>
      <el-input v-model="keyword" placeholder="搜索标题 / 摘要" clearable class="search" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
    </div>
    <el-table :data="items" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" min-width="260" />
      <el-table-column prop="site" label="子站" width="110" />
      <el-table-column prop="board" label="板块" width="110" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="comments" label="评论/回答" width="110" />
      <el-table-column prop="updated_at" label="更新时间" width="170" />
      <el-table-column label="操作" width="130">
        <template #default="{ row }">
          <el-button link type="primary" @click="open(row)">查看</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import { plugins, posts } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const items = ref([]);
const keyword = ref('');

async function load() {
  const pluginList = await plugins();
  const current = (pluginList.items || []).find((item) => item.code === route.meta.pluginCode);
  if (!current || current.status !== 'enabled') {
    ElMessage.warning('当前插件未启用，请先在系统插件中启用。');
    router.replace('/plugins');
    return;
  }
  const permission = current.menus?.find((item) => item.area === 'admin')?.permission || route.meta.permission;
  if (permission && !auth.can(permission)) {
    ElMessage.warning('当前账号无权访问该插件管理页。');
    router.replace('/plugins');
    return;
  }
  const data = await posts({ board: 'all', q: keyword.value, content_type: route.meta.contentType });
  items.value = data.items || data || [];
}

function open(row) {
  window.open(`/topics/${row.id}/`, '_blank');
}

watch(() => route.meta.contentType, load);
onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 16px; }
.toolbar { display: flex; gap: 12px; align-items: center; }
.toolbar h2 { margin: 0 0 6px; }
.toolbar p { margin: 0; color: #64748b; }
.search { max-width: 260px; margin-left: auto; }
</style>
