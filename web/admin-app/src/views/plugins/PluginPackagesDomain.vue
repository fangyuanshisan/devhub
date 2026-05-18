<template>
  <section class="plugin-domain-page" data-testid="plugin-packages-domain">
    <div class="plugin-domain-header">
      <div>
        <p class="eyebrow">插件管理</p>
        <h2>插件包治理</h2>
        <p class="muted">统一承接远程索引、暂存上传、本地预检、依赖兼容、版本升级和安装审批。</p>
      </div>
    </div>

    <el-alert
      class="domain-alert"
      type="info"
      show-icon
      :closable="false"
      title="远程索引不等于自动安装；暂存区不等于已安装；预检通过不等于可启用；启用前检查通过后才允许启用；升级不会执行第三方代码。"
    />

    <el-tabs v-model="activeTab" class="plugin-domain-tabs" data-testid="plugin-packages-domain-tabs">
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
  h('p', { class: 'muted' }, '正在加载插件包治理内容...'),
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

const defaultTab = 'remote-indexes';
const tabs = [
  { name: 'remote-indexes', label: '远程索引', component: lazyTab(() => import('./PluginRemoteIndexes.vue')) },
  { name: 'remote-packages', label: '远程包下载', component: lazyTab(() => import('./PluginRemotePackages.vue')) },
  { name: 'uploads', label: '暂存上传包', component: lazyTab(() => import('./PluginPackageUploads.vue')) },
  { name: 'install', label: '本地包与预检', component: lazyTab(() => import('./PluginInstallUpgrade.vue')) },
  { name: 'dependencies', label: '依赖 / 兼容性', component: lazyTab(() => import('./PluginDependencies.vue')) },
  { name: 'versions', label: '版本与升级', component: lazyTab(() => import('./PluginVersions.vue')) },
  { name: 'approvals', label: '安装 / 升级审批', component: lazyTab(() => import('./PluginApprovals.vue')) },
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
</style>
