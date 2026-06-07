<template>
  <section class="plugin-domain-page" data-testid="plugin-packages-domain">
    <AdminPageHeader
      title="安装与升级"
      description="围绕插件安装、升级、上传、远程下载、预检、依赖兼容和审批处理包生命周期。"
      :breadcrumbs="['插件管理']"
      testid="plugin-packages-page-header"
    />

    <el-alert
      class="domain-alert"
      type="info"
      show-icon
      :closable="false"
      title="远程索引不等于自动安装；暂存区不等于已安装；预检通过不等于可启用；启用前检查通过后才允许启用；升级不会执行第三方代码。"
    />

    <el-tabs v-model="activeTab" class="plugin-domain-tabs" data-testid="plugin-packages-domain-tabs">
      <el-tab-pane v-for="item in visibleTabs" :key="item.name" :label="item.label" :name="item.name" />
    </el-tabs>

    <component :is="activeComponent" />
  </section>
</template>

<script setup>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue';
import { ElAlert, ElButton, ElCard, ElDescriptions, ElDescriptionsItem, ElSkeleton, ElSteps, ElStep } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import { AdminPageHeader } from '@/components/admin';

const route = useRoute();
const router = useRouter();

const tabLoading = () => h('div', { class: 'tab-state' }, [
  h(ElSkeleton, { rows: 4, animated: true }),
  h('p', { class: 'muted' }, '正在加载安装与升级内容...'),
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

function PackageFlowWorkbench() {
  const step = (tab, label, description) => h(ElButton, {
    plain: true,
    type: 'primary',
    onClick: () => router.push({ path: '/plugins/packages', query: { ...route.query, tab } }),
  }, () => label || description);
  return h(ElCard, { shadow: 'never', class: 'domain-note-card' }, () => [
    h('h3', { class: 'note-title' }, '插件包流程工作台'),
    h('p', { class: 'muted' }, '先判断包处在哪一步，再执行下一步；blocked 包不能转入本地仓库，安装前必须重新执行 install dry-run。'),
    h(ElSteps, { active: 1, simple: true, class: 'flow-steps' }, () => [
      h(ElStep, { title: '上传 / 暂存' }),
      h(ElStep, { title: '预检' }),
      h(ElStep, { title: '转入本地仓库' }),
      h(ElStep, { title: '安装 dry-run' }),
      h(ElStep, { title: '安装 / 升级' }),
    ]),
    h('div', { class: 'next-actions' }, [
      step('install', '执行安装 dry-run'),
      step('versions', '版本与升级'),
      step('uploads', '查看上传包'),
      step('approvals', '审批中心'),
      step('remote-indexes', '远程索引'),
    ]),
    h(ElDescriptions, { column: 1, border: true, class: 'note-descriptions' }, () => [
      h(ElDescriptionsItem, { label: '阻断包' }, () => '查看阻断原因；后端仍会强校验，不能只依赖按钮禁用。'),
      h(ElDescriptionsItem, { label: '本地仓库包' }, () => '安装前必须重新 dry-run，upload 阶段 dry-run 不可替代 install dry-run。'),
      h(ElDescriptionsItem, { label: '技术详情' }, () => 'Manifest、migration plan、warnings、blockers 和原始 JSON 默认放入对应详情区域。'),
    ]),
  ]);
}

const defaultTab = 'flow';
const tabs = [
  { name: 'flow', label: '流程工作台', component: PackageFlowWorkbench },
  { name: 'install', label: '本地包与预检', component: lazyTab(() => import('./PluginInstallUpgrade.vue')) },
  { name: 'versions', label: '版本与升级', component: lazyTab(() => import('./PluginVersions.vue')) },
  { name: 'uploads', label: '上传包', component: lazyTab(() => import('./PluginPackageUploads.vue')) },
  { name: 'remote-indexes', label: '远程索引', component: lazyTab(() => import('./PluginRemoteIndexes.vue')) },
  { name: 'remote-packages', label: '远程包下载', component: lazyTab(() => import('./PluginRemotePackages.vue')) },
  { name: 'dependencies', label: '依赖兼容', component: lazyTab(() => import('./PluginDependencies.vue')) },
  { name: 'approvals', label: '审批中心', component: lazyTab(() => import('./PluginApprovals.vue')) },
];
const tabNames = new Set(tabs.map((item) => item.name));
const normalizeTab = (value) => (tabNames.has(String(value || '')) ? String(value) : defaultTab);

const activeTab = ref(normalizeTab(route.query.tab));
const visibleTabs = computed(() => {
  const primary = new Set(['flow', 'install', 'versions', 'uploads']);
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

.domain-alert,
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

.flow-steps {
  margin: 12px 0;
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
