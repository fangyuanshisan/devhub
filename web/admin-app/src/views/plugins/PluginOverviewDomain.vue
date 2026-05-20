<template>
  <section class="plugin-domain-page" data-testid="plugin-overview-domain">
    <div class="plugin-domain-header">
      <div>
        <p class="eyebrow">插件管理</p>
        <h2>插件总览</h2>
        <p class="muted">按任务找入口：日常操作先去总览和列表，安装升级去插件包治理，投递与密钥去 Webhook，排障和追踪去运行记录 / 审计。</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="plugin-domain-tabs" data-testid="plugin-overview-domain-tabs">
      <el-tab-pane v-for="item in visibleTabs" :key="item.name" :label="item.label" :name="item.name" />
    </el-tabs>

    <component :is="activeComponent" />
  </section>
</template>

<script setup>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue';
import { ElAlert, ElButton, ElCard, ElDescriptions, ElDescriptionsItem, ElSkeleton } from 'element-plus';
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

function AdvancedGovernance() {
  const go = (tab, label) => h(ElButton, {
    plain: true,
    type: 'primary',
    onClick: () => router.push({ path: '/plugins/overview', query: { ...route.query, tab } }),
  }, () => label);
  return h(ElCard, { shadow: 'never', class: 'domain-note-card' }, () => [
    h('h3', { class: 'note-title' }, '高级治理'),
    h('p', { class: 'muted' }, '低频能力仍可访问，但不再默认铺在插件总览首页。需要改配置、看挂载、查权限或排障时，再进入对应功能区。'),
    h('div', { class: 'next-actions' }, [
      go('config', '配置中心'),
      go('navigation', '前端挂载'),
      go('content', '内容治理'),
      go('permissions', '权限矩阵'),
      go('developer', '开发者工具'),
    ]),
    h(ElDescriptions, { column: 1, border: true, class: 'note-descriptions' }, () => [
      h(ElDescriptionsItem, { label: '日常入口' }, () => '优先使用“总览”和“插件列表”处理启停、配置、异常和详情。'),
      h(ElDescriptionsItem, { label: '变更 / 排障' }, () => '安装升级去插件包治理，投递与密钥去 Webhook 治理，操作历史和审计去运行记录 / 审计。'),
      h(ElDescriptionsItem, { label: '技术内容' }, () => '原始声明、权限引用、前端挂载明细和开发工具归入高级治理或插件详情技术详情。'),
    ]),
  ]);
}

const defaultTab = 'overview';
const tabs = [
  { name: 'overview', label: '总览', component: lazyTab(() => import('./PluginOverview.vue')) },
  { name: 'list', label: '插件列表', component: lazyTab(() => import('./PluginList.vue')) },
  { name: 'advanced', label: '高级治理', component: AdvancedGovernance },
  { name: 'config', label: '配置中心', component: lazyTab(() => import('./PluginConfigHub.vue')) },
  { name: 'navigation', label: '前端挂载', component: lazyTab(() => import('./PluginNavigation.vue')) },
  { name: 'content', label: '内容治理', component: lazyTab(() => import('./PluginContentHub.vue')) },
  { name: 'permissions', label: '权限矩阵', component: lazyTab(() => import('./PluginPermissions.vue')) },
  { name: 'developer', label: '开发者工具', component: lazyTab(() => import('./PluginDeveloper.vue')) },
];
const tabNames = new Set(tabs.map((item) => item.name));
const normalizeTab = (value) => (tabNames.has(String(value || '')) ? String(value) : defaultTab);

const activeTab = ref(normalizeTab(route.query.tab));
const visibleTabs = computed(() => {
  const primary = new Set(['overview', 'list', 'advanced']);
  return tabs.filter((item) => primary.has(item.name) || item.name === activeTab.value);
});
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

.domain-note-card {
  margin-top: 8px;
}

.note-title {
  margin: 0 0 8px;
  font-size: 16px;
}

.next-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 12px 0;
}

.note-descriptions {
  margin-top: 12px;
}
</style>
