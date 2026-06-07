<template>
  <section class="plugin-domain-page" data-testid="plugin-runtime-domain">
    <div class="plugin-domain-header">
      <div>
        <p class="eyebrow">插件管理</p>
        <h2>运行与审计</h2>
        <p class="muted">追踪插件运行错误、操作历史、Hook 执行、搜索索引和审计日志，排障相关记录都从这里进入。</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="plugin-domain-tabs" data-testid="plugin-runtime-domain-tabs">
      <el-tab-pane v-for="item in visibleTabs" :key="item.name" :label="item.label" :name="item.name" />
    </el-tabs>

    <component :is="activeComponent" />
  </section>
</template>

<script setup>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue';
import { ElAlert, ElButton, ElCard, ElDescriptions, ElDescriptionsItem, ElText } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();

function RecentErrorsGuidance() {
  return h(ElCard, { shadow: 'never', class: 'domain-note-card' }, () => [
    h('h3', { class: 'note-title' }, '最近错误'),
    h(ElText, { type: 'info' }, () => '最近错误用于快速定位插件运行、Webhook 投递和 Hook 执行异常。完整记录仍以操作历史、审计日志和 Hook 排障为主入口。'),
    h(ElAlert, {
      class: 'note-alert',
      type: 'warning',
      showIcon: true,
      closable: false,
      title: '本 Tab 只做错误定位入口，不重复展示全局治理表格。',
    }),
    h(ElDescriptions, { column: 1, border: true, class: 'note-descriptions' }, () => [
      h(ElDescriptionsItem, { label: '操作错误' }, () => '进入“操作历史”查看安装、升级、审批、恢复等结果。'),
      h(ElDescriptionsItem, { label: 'Hook 错误' }, () => '进入“Hook 排障”查看 Hook 执行失败和耗时。'),
      h(ElDescriptionsItem, { label: '审计追踪' }, () => '进入“审计日志”按操作人、时间和结果排查。'),
    ]),
  ]);
}

function AdvancedRuntimeGuidance() {
  const go = (tab, label) => h(ElButton, {
    plain: true,
    type: 'primary',
    onClick: () => router.push({ path: '/plugins/runtime', query: { ...route.query, tab } }),
  }, () => label);
  return h(ElCard, { shadow: 'never', class: 'domain-note-card' }, () => [
    h('h3', { class: 'note-title' }, '高级排障'),
    h(ElText, { type: 'info' }, () => '低频排障入口不再默认铺开；需要定位 Hook、索引或 request_id 时再进入对应工具。'),
    h('div', { class: 'next-actions' }, [
      go('hooks', 'Hook 排障'),
      go('search-index', '搜索索引'),
      go('audit', '审计日志'),
      go('operations', '操作历史'),
    ]),
    h(ElAlert, {
      class: 'note-alert',
      type: 'info',
      showIcon: true,
      closable: false,
      title: '原始 metadata、request_id 和技术 trace 默认折叠，仅用于排障。',
    }),
  ]);
}

const tabLoading = () => h('div', { class: 'tab-state' }, [
  h(ElText, { type: 'info' }, () => '正在加载运行与审计内容...'),
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

const defaultTab = 'errors';
const tabs = [
  { name: 'errors', label: '最近错误', component: RecentErrorsGuidance },
  { name: 'operations', label: '操作历史', component: lazyTab(() => import('./PluginOperations.vue')) },
  { name: 'hooks', label: 'Hook 排障', component: lazyTab(() => import('./PluginHooks.vue')) },
  { name: 'search-index', label: '搜索索引', component: lazyTab(() => import('./PluginSearchIndex.vue')) },
  { name: 'audit', label: '审计日志', component: lazyTab(() => import('./PluginAudit.vue')) },
  { name: 'advanced', label: '高级排障', component: AdvancedRuntimeGuidance },
];
const tabNames = new Set(tabs.map((item) => item.name));
const normalizeTab = (value) => (tabNames.has(String(value || '')) ? String(value) : defaultTab);

const activeTab = ref(normalizeTab(route.query.tab));
const visibleTabs = computed(() => {
  const primary = new Set(['errors', 'operations', 'hooks', 'search-index', 'audit']);
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

.domain-note-card {
  margin-top: 8px;
}

.tab-state {
  margin-top: 12px;
}

.note-title {
  margin: 0 0 8px;
  font-size: 16px;
}

.note-alert,
.note-descriptions {
  margin-top: 12px;
}

.next-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 12px 0;
}
</style>
