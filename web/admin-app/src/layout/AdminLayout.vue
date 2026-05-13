<template>
  <el-container class="admin-layout">
    <aside class="primary-nav">
      <div class="mini-logo">DH</div>
      <button v-for="item in visibleMenus" :key="item.path" :class="['primary-item', { active: activeGroup.path === item.path }]" @click="router.push(item.path)">
        <el-icon><component :is="icons[item.meta.icon]" /></el-icon>
        <span>{{ item.meta.short || item.meta.title.slice(0, 2) }}</span>
      </button>
    </aside>

    <aside v-if="showSubNav" class="sub-nav">
      <div class="edition">DevHub 标准版</div>
      <div class="sub-title">{{ activeGroup.meta?.title || '控制台' }}</div>
      <button
        v-for="child in activeChildren"
        :key="child.key"
        :class="['sub-item', { active: child.key === activeSub }]"
        @click="handleSubClick(child)"
      >
        {{ child.label }}
      </button>
    </aside>

    <el-container class="workbench">
      <el-header class="header">
        <div class="header-left">
          <el-icon class="fold-icon"><Fold /></el-icon>
          <span class="route-title">{{ activeGroup.meta?.title }}</span>
          <template v-if="showSubNav && activeSubLabel">
            <span class="slash">/</span>
            <span>{{ activeSubLabel }}</span>
          </template>
        </div>
        <div class="header-actions">
          <el-icon><Refresh /></el-icon>
          <el-input v-model="quickSearch" placeholder="搜索" :prefix-icon="Search" clearable class="quick-search" />
          <el-badge is-dot>
            <el-icon><Bell /></el-icon>
          </el-badge>
          <el-icon><FullScreen /></el-icon>
          <el-tag effect="plain">{{ scopeLabel }}</el-tag>
          <el-dropdown>
            <span class="user-entry">{{ auth.user?.nickname || auth.user?.username || 'demo' }} <el-icon><ArrowDown /></el-icon></span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-icon><Setting /></el-icon>
        </div>
      </el-header>

      <div class="tabbar">
        <el-tabs :model-value="$route.path" type="card" closable @tab-remove="removeTab" @tab-click="openTab">
          <el-tab-pane v-for="tab in tabs.tabs" :key="tab.path" :label="tab.title" :name="tab.path" />
        </el-tabs>
        <el-icon class="grid-icon"><Grid /></el-icon>
      </div>

      <el-main class="main">
        <router-view v-slot="{ Component, route }">
          <keep-alive :include="keepAliveNames">
            <component :is="Component" :key="route.name" />
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { ArrowDown, Bell, Briefcase, ChatDotRound, Connection, DataBoard, Document, Fold, FolderOpened, FullScreen, Grid, MagicStick, Notebook, PriceTag, Promotion, QuestionFilled, Refresh, Search, SetUp, Setting, Tickets, TrendCharts, User, UserFilled, Warning } from '@element-plus/icons-vue';
import { plugins as fetchPlugins } from '@/api/admin';
import { menuRoutes } from '@/router';
import { useAuthStore } from '@/stores/auth';
import { useTabsStore } from '@/stores/tabs';

const icons = { Briefcase, ChatDotRound, Connection, DataBoard, Document, FolderOpened, MagicStick, Notebook, PriceTag, Promotion, QuestionFilled, SetUp, Setting, Tickets, TrendCharts, User, UserFilled, Warning };
const subMenus = {
  dashboard: [{ key: 'overview', label: '运营概览' }, { key: 'todo', label: '待办事项' }],
  content: [{ key: 'posts', label: '内容管理' }, { key: 'docs', label: '文档管理' }, { key: 'tags', label: '标签管理' }],
  comments: [{ key: 'all', label: '评论列表' }, { key: 'audit', label: '审核队列' }],
  reports: [{ key: 'pending', label: '待处理' }, { key: 'handled', label: '处理记录' }],
  moderators: [{ key: 'list', label: '版主列表' }, { key: 'scope', label: '子站授权' }],
  auditLogs: [{ key: 'list', label: '审计列表' }, { key: 'filter', label: '治理筛选' }],
  tags: [{ key: 'list', label: '标签列表' }, { key: 'seo', label: 'SEO 配置' }, { key: 'topics', label: '关联内容' }],
  plugins: [
    { key: 'list', label: '插件列表', path: '/plugins/list' },
    { key: 'status', label: '状态治理', path: '/plugins/governance' },
    { key: 'manifest', label: '安装 / 升级', path: '/plugins/manifest' },
    { key: 'diagnostics', label: '诊断与排障', path: '/plugins/diagnostics' },
  ],
  qaPlugin: [{ key: 'questions', label: '问题列表' }, { key: 'answers', label: '回答治理' }],
  docsPlugin: [{ key: 'documents', label: '文档列表' }, { key: 'spaces', label: '空间结构' }],
  wikiPlugin: [{ key: 'pages', label: '页面列表' }, { key: 'versions', label: '版本历史' }],
  projectsPlugin: [{ key: 'projects', label: '项目列表' }, { key: 'review', label: '项目治理' }],
  jobsPlugin: [{ key: 'jobs', label: '招聘列表' }, { key: 'review', label: '招聘治理' }],
  aiWorksPlugin: [{ key: 'works', label: '作品列表' }, { key: 'review', label: '作品治理' }],
  communities: [{ key: 'site', label: '子站配置' }, { key: 'board', label: '板块管理' }],
  sites: [{ key: 'site', label: '子站配置' }, { key: 'board', label: '板块管理' }],
  users: [{ key: 'list', label: '用户管理' }, { key: 'roles', label: '角色权限' }],
  operation: [{ key: 'notice', label: '通知推送' }, { key: 'recommend', label: '推荐位' }],
  statistics: [{ key: 'site', label: '子站统计' }, { key: 'board', label: '板块统计' }],
  system: [{ key: 'setting', label: '基础设置' }, { key: 'log', label: '操作日志' }],
};

const auth = useAuthStore();
const tabs = useTabsStore();
const route = useRoute();
const router = useRouter();
const quickSearch = ref('');
const activeSub = ref('overview');
const pluginStatus = ref({});

const visibleMenus = computed(() => menuRoutes.filter((item) => {
  if (item.meta.hiddenInMenu) return false;
  if (item.meta.pluginCode && pluginStatus.value[item.meta.pluginCode] && pluginStatus.value[item.meta.pluginCode] !== 'enabled') return false;
  return !item.meta.permission || auth.can(item.meta.permission);
}));
const activeGroup = computed(() => {
  if (route.meta?.subNavGroup) return menuRoutes.find((item) => item.name === 'plugins') || menuRoutes.find((item) => item.path === '/plugins') || visibleMenus.value[0] || menuRoutes[0];
  return menuRoutes.find((item) => item.path === route.path) || visibleMenus.value[0] || menuRoutes[0];
});
const activeChildren = computed(() => {
  if (route.meta?.subNavGroup) return subMenus[route.meta.subNavGroup] || [];
  return subMenus[activeGroup.value.name] || [];
});
const showSubNav = computed(() => Boolean(activeChildren.value.length));
const activeSubLabel = computed(() => activeChildren.value.find((item) => item.key === activeSub.value)?.label || activeChildren.value[0]?.label || '');
const keepAliveNames = computed(() => menuRoutes.filter((item) => item.meta.keepAlive).map((item) => item.name));
const scopeLabel = computed(() => {
  const site = new URLSearchParams(window.location.search).get('site') || 'portal';
  return site === 'portal' ? '全局后台' : `${site} 子站`;
});

watch(route, (value) => {
  tabs.add(value);
  const nextChildren = value.meta?.subNavGroup ? (subMenus[value.meta.subNavGroup] || []) : (subMenus[value.name] || []);
  if (value.meta?.subNavKey) {
    activeSub.value = value.meta.subNavKey;
    return;
  }
  activeSub.value = nextChildren[0]?.key || '';
}, { immediate: true });
watch(quickSearch, (value) => {
  const matched = menuRoutes.find((item) => !item.meta.hiddenInMenu && item.meta.title.includes(value));
  if (value && matched) ElMessage.info(`可进入：${matched.meta.title}`);
});

onMounted(async () => {
  if (!auth.authed || !auth.can('plugin.read')) return;
  try {
    const data = await fetchPlugins();
    pluginStatus.value = Object.fromEntries((data.items || []).map((item) => [item.code || item.plugin_code, item.status]));
  } catch (error) {
    pluginStatus.value = {};
  }
});

function removeTab(path) {
  tabs.remove(path);
  if (route.path === path) router.push(tabs.tabs.at(-1)?.path || '/dashboard');
}

function openTab(tab) {
  router.push(tab.props.name);
}

async function logout() {
  await auth.logout();
  router.push('/login');
}

function handleSubClick(child) {
  activeSub.value = child.key;
  if (child.path) router.push(child.path);
}
</script>
