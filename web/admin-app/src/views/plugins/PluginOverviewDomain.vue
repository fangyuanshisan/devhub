<template>
  <section class="plugin-domain-page" data-testid="plugin-overview-domain">
    <div class="plugin-domain-header">
      <div>
        <p class="eyebrow">插件管理</p>
        <h2>插件总览</h2>
        <p class="muted">统一查看插件状态、配置、前端挂载、内容治理、权限矩阵和开发者工具；低频入口收进页内 Tab。</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="plugin-domain-tabs" data-testid="plugin-overview-domain-tabs">
      <el-tab-pane v-for="item in tabs" :key="item.name" :label="item.label" :name="item.name" />
    </el-tabs>

    <component :is="activeComponent" />
  </section>
</template>

<script setup>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue';
import { ElAlert, ElSkeleton } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();

const tabLoading = () => h('div', { class: 'tab-state' }, [
  h(ElSkeleton, { rows: 4, animated: true }),
  h('p', { class: 'muted' }, '正在加载插件总览内容...'),
]);
const tabError = () => h(ElAlert, {
  class: 'tab-state',
  type: 'error',
  showIcon: true,
  closable: false,
  title: '加载失败，请刷新后重试',
});
const lazyTab = (loader) => defineAsyncComponent({
  loader,
  loadingComponent: tabLoading,
  errorComponent: tabError,
  delay: 120,
  timeout: 20000,
});

const defaultTab = 'overview';
const tabs = [
  { name: 'overview', label: '总览', component: lazyTab(() => import('./PluginOverview.vue')) },
  { name: 'list', label: '插件列表', component: lazyTab(() => import('./PluginList.vue')) },
  { name: 'config', label: '配置中心', component: lazyTab(() => import('./PluginConfigHub.vue')) },
  { name: 'navigation', label: '前端挂载', component: lazyTab(() => import('./PluginNavigation.vue')) },
  { name: 'content', label: '内容治理', component: lazyTab(() => import('./PluginContentHub.vue')) },
  { name: 'permissions', label: '权限矩阵', component: lazyTab(() => import('./PluginPermissions.vue')) },
  { name: 'developer', label: '开发者工具', component: lazyTab(() => import('./PluginDeveloper.vue')) },
];
const tabNames = new Set(tabs.map((item) => item.name));
const normalizeTab = (value) => (tabNames.has(String(value || '')) ? String(value) : defaultTab);

const activeTab = ref(normalizeTab(route.query.tab));
const activeComponent = computed(() => tabs.find((item) => item.name === activeTab.value)?.component || tabs[0].component);

watch(activeTab, async (next) => {
  const query = { ...route.query };
  if (next === defaultTab) delete query.tab;
  else query.tab = next;
  await router.replace({ query });
});

watch(() => route.query.tab, (value) => {
  const next = normalizeTab(value);
  if (activeTab.value !== next) activeTab.value = next;
});
</script>

<style scoped>
.plugin-domain-page {
  padding: 4px 0 12px;
}

.plugin-domain-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0 4px;
}

.plugin-domain-header h2 {
  margin: 0;
  font-size: 18px;
  line-height: 26px;
}

.eyebrow {
  margin: 0 0 4px;
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
}

.muted {
  margin: 4px 0 0;
  color: #6b7280;
}

.plugin-domain-tabs {
  margin-top: 8px;
}

.tab-state {
  margin-top: 12px;
}
</style>
