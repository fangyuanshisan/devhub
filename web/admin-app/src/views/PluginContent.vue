<template>
  <section class="page-card" data-testid="plugin-content-page">
    <div class="toolbar">
      <div>
        <h2>{{ route.meta.title }}</h2>
        <p>
          当前插件：<span class="mono">{{ route.meta.pluginCode }}</span>
          · content_type：<span class="mono">{{ route.meta.contentType }}</span>
          · 状态：<el-tag v-if="plugin" :type="statusType(plugin.status)" size="small">{{ plugin.status }}</el-tag>
        </p>
        <p class="muted">插件内容通过 Core 通用内容表兼容展示；禁用插件不影响历史内容访问，只影响新发布与入口。</p>
      </div>
      <div class="tool-actions">
        <el-select v-model="filters.communityId" clearable filterable placeholder="子站" style="width: 160px" data-testid="plugin-content-community-filter">
          <el-option v-for="c in communities" :key="c.id" :label="`${c.name} /${c.slug}`" :value="c.id" />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="状态" style="width: 140px" data-testid="plugin-content-status-filter">
          <el-option label="全部" value="all" />
          <el-option label="publish" value="publish" />
          <el-option label="hidden" value="hidden" />
          <el-option label="pending" value="pending" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索标题 / 摘要" clearable class="search" data-testid="plugin-content-search" @keyup.enter="load" />
        <el-button data-testid="plugin-content-back" @click="backToPlugins">返回插件</el-button>
        <el-button type="primary" data-testid="plugin-content-query" @click="load">查询</el-button>
      </div>
    </div>
    <el-table :data="items" border stripe data-testid="plugin-content-table">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" min-width="260" />
      <el-table-column prop="site" label="子站" width="110" />
      <el-table-column prop="board" label="板块" width="110" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="comments" label="评论/回答" width="110" />
      <el-table-column prop="updated_at" label="更新时间" width="170" />
      <el-table-column label="plugin/type" width="180">
        <template #default="{ row }">
          <div class="mono">{{ row.plugin_code || '-' }}</div>
          <div class="muted mono">{{ row.content_type || '-' }}</div>
        </template>
      </el-table-column>
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
import { adminCommunities, plugins, posts } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const items = ref([]);
const keyword = ref('');
const plugin = ref(null);
const communities = ref([]);
const filters = ref({ communityId: null, status: 'all' });

async function load() {
  const pluginList = await plugins();
  const current = (pluginList.items || []).find((item) => item.code === route.meta.pluginCode);
  if (!current || current.status !== 'enabled') {
    ElMessage.warning('当前插件未启用，请先在系统插件中启用。');
    router.replace('/plugins');
    return;
  }
  plugin.value = current;
  const permission = current.menus?.find((item) => item.area === 'admin')?.permission || route.meta.permission;
  if (permission && !auth.can(permission)) {
    ElMessage.warning('当前账号无权访问该插件管理页。');
    router.replace('/plugins');
    return;
  }
  if (!communities.value.length) {
    const commData = await adminCommunities();
    communities.value = commData.items || [];
  }

  const site =
    filters.value.communityId && communities.value.find((c) => c.id === filters.value.communityId)?.slug
      ? communities.value.find((c) => c.id === filters.value.communityId)?.slug
      : 'portal';

  const data = await posts({ site, board: 'all', q: keyword.value, content_type: route.meta.contentType, status: filters.value.status || 'all' });
  const list = data.items || data || [];
  // Best-effort: keep only rows belonging to this plugin/type when data comes from legacy adminPosts.
  items.value = list.filter((r) => (r.content_type || route.meta.contentType) === route.meta.contentType);
}

function open(row) {
  window.open(`/topics/${row.id}/`, '_blank');
}

function backToPlugins() {
  router.push('/plugins');
}

function statusType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'danger';
  return 'info';
}

watch(() => route.meta.contentType, load);
onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 16px; }
.toolbar { display: flex; gap: 12px; align-items: flex-start; }
.toolbar h2 { margin: 0 0 6px; }
.toolbar p { margin: 0; color: #64748b; }
.muted { color: #64748b; }
.tool-actions { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; margin-left: auto; justify-content: flex-end; }
.search { max-width: 260px; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
</style>
