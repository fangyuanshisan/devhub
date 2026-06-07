<template>
  <section class="plugin-page" data-testid="plugin-content-hub-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">插件运营</div>
        <h2>内容治理</h2>
        <p class="muted">按插件内容类型聚合入口；具体治理仍进入对应 PluginContent 页面（旧路由保持兼容）。</p>
      </div>
    </div>

    <el-alert v-if="error" type="error" show-icon :closable="false" class="mb" :title="error" />

    <div class="governance-grid" data-testid="plugin-content-hub-grid">
      <article v-for="card in cards" :key="card.key" class="governance-card">
        <div class="governance-card-head">
          <div>
            <h3>{{ card.title }}</h3>
            <p>{{ card.desc }}</p>
          </div>
          <el-tag :type="statusTagType(card.plugin?.status)" effect="plain">{{ pluginStatusLabel(card.plugin?.status) }}</el-tag>
        </div>
        <div class="tag-wrap mb">
          <el-tag type="info" effect="plain">{{ card.pluginCode }}</el-tag>
          <el-tag type="info" effect="plain">{{ card.contentType }}</el-tag>
          <el-tag v-if="card.plugin?.health?.status" :type="healthTagType(card.plugin.health.status)" effect="plain">{{ pluginHealthLabel(card.plugin.health.status) }}</el-tag>
        </div>
        <div class="row-actions">
          <el-button type="primary" plain :data-testid="`plugin-content-hub-open-${card.pluginCode}`" @click="router.push(card.path)">进入治理</el-button>
          <el-button type="info" plain @click="openPlugin(card.plugin, 'overview')">插件详情</el-button>
        </div>
      </article>
    </div>

    <PluginDetailDrawer v-model="drawerVisible" :plugin="drawerPlugin" :plugins="items" :initial-tab="drawerTab" @refresh="load" @open-plugin="openByCode" />
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import PluginDetailDrawer from '@/components/plugin/PluginDetailDrawer.vue';
import { usePluginData } from './usePluginData';
import { pluginHealthLabel, pluginStatusLabel } from '@/i18n/formatters';

const router = useRouter();
const drawerVisible = ref(false);
const drawerPlugin = ref(null);
const drawerTab = ref('overview');
const { items, loading, error, load } = usePluginData();

onMounted(load);

function findPlugin(code) {
  return (items.value || []).find((p) => (p.code || p.plugin_code) === code) || null;
}

const cards = computed(() => ([
  { key: 'qa', title: '问答治理', desc: 'question 内容治理入口', pluginCode: 'qa', contentType: 'question', path: '/qa' },
  { key: 'docs', title: '文档治理', desc: 'document 内容治理入口', pluginCode: 'docs', contentType: 'document', path: '/docs' },
  { key: 'wiki', title: 'Wiki 治理', desc: 'wiki_page 内容治理入口', pluginCode: 'wiki', contentType: 'wiki_page', path: '/wiki' },
  { key: 'projects', title: '项目治理', desc: 'project 内容治理入口', pluginCode: 'projects', contentType: 'project', path: '/projects' },
  { key: 'jobs', title: '招聘治理', desc: 'job 内容治理入口', pluginCode: 'jobs', contentType: 'job', path: '/jobs' },
  { key: 'ai_works', title: 'AI 作品治理', desc: 'ai_work 内容治理入口', pluginCode: 'ai_works', contentType: 'ai_work', path: '/ai-works' },
  { key: 'plugin_a7b0cc04', title: '飞书链接治理', desc: 'feishu_link 内容治理入口', pluginCode: 'plugin_a7b0cc04', contentType: 'feishu_link', path: '/feishu-links' },
].map((c) => ({ ...c, plugin: findPlugin(c.pluginCode) }))));

function openPlugin(row, tab) {
  if (!row) return;
  drawerPlugin.value = row;
  drawerTab.value = tab || 'overview';
  drawerVisible.value = true;
}

function openByCode(code, tab = 'overview') {
  const target = findPlugin(code);
  if (!target) return;
  openPlugin(target, tab);
}

function statusTagType(status) {
  if (status === 'enabled') return 'success';
  if (status === 'disabled') return 'warning';
  if (status === 'archived') return 'info';
  return 'info';
}

function healthTagType(status) {
  if (!status || status === 'healthy') return 'success';
  if (status === 'hook_warning' || status === 'migration_pending' || status === 'dependency_missing') return 'warning';
  if (status === 'hook_error' || status === 'error' || status === 'config_invalid') return 'danger';
  return 'info';
}
</script>
