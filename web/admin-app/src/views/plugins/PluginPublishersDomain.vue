<template>
  <section class="plugin-domain-page" data-testid="plugin-publishers-domain">
    <div class="plugin-domain-header">
      <div>
        <p class="eyebrow">插件管理</p>
        <h2>发布者与信任</h2>
        <p class="muted">统一管理可信发布者、公钥 key_id、可信级别、影响分析和密钥轮换；依赖兼容已归入插件包治理。</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="plugin-domain-tabs" data-testid="plugin-publishers-domain-tabs">
      <el-tab-pane v-for="item in tabs" :key="item.name" :label="item.label" :name="item.name" />
    </el-tabs>

    <component :is="activeComponent" />
  </section>
</template>

<script setup>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue';
import { ElAlert, ElCard, ElDescriptions, ElDescriptionsItem, ElText } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();

function PublisherGuidance(_, { title, description }) {
  return h(ElCard, { shadow: 'never', class: 'domain-note-card' }, () => [
    h('h3', { class: 'note-title' }, title),
    h(ElText, { type: 'info' }, () => description),
    h(ElAlert, {
      class: 'note-alert',
      type: 'info',
      showIcon: true,
      closable: false,
      title: '发布者列表是主数据入口，本页只提供治理域说明，避免重复展示完整表格。',
    }),
    h(ElDescriptions, { column: 1, border: true, class: 'note-descriptions' }, () => [
      h(ElDescriptionsItem, { label: '主入口' }, () => '发布者列表'),
      h(ElDescriptionsItem, { label: '安全边界' }, () => '不展示私钥或 Secret 明文；吊销、禁用和恢复仍走原有确认流程。'),
    ]),
  ]);
}

const KeysGuidance = (props, context) => PublisherGuidance(props, { ...context, title: '公钥 / key_id', description: '查看发布者公钥标识、key_id 与签名信任关系，请进入“发布者列表”查看详情。' });
const TrustGuidance = (props, context) => PublisherGuidance(props, { ...context, title: '可信级别', description: '可信级别用于区分官方、可信、企业私有和本地开发发布者，列表中的状态 badge 已统一中文化。' });
const ImpactGuidance = (props, context) => PublisherGuidance(props, { ...context, title: '影响分析', description: '影响分析围绕 allowed / blocked plugin codes、过期时间和最近使用情况展开。' });
const tabLoading = () => h('div', { class: 'tab-state' }, [
  h(ElText, { type: 'info' }, () => '正在加载发布者与信任内容...'),
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

const defaultTab = 'list';
const tabs = [
  { name: 'list', label: '发布者列表', component: lazyTab(() => import('./PluginTrustedPublishers.vue')) },
  { name: 'keys', label: '公钥 / key_id', component: KeysGuidance },
  { name: 'trust-level', label: '可信级别', component: TrustGuidance },
  { name: 'impact', label: '影响分析', component: ImpactGuidance },
  { name: 'config-keys', label: '密钥轮换', component: lazyTab(() => import('./PluginConfigKeys.vue')) },
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
</style>
