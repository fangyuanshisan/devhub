<template>
  <el-container class="admin-layout">
    <aside class="primary-nav">
      <div class="mini-logo">DH</div>
      <button
        v-for="mod in visibleModules"
        :key="mod.key"
        :class="['primary-item', { active: activeModuleKey === mod.key }]"
        :data-testid="`admin-primary-nav-${mod.key}`"
        @click="router.push(mod.entry)"
      >
        <el-icon><component :is="icons[mod.icon]" /></el-icon>
        <span>{{ mod.short || mod.title.slice(0, 2) }}</span>
      </button>
    </aside>

    <aside v-if="showSubNav" class="sub-nav" data-testid="admin-sub-nav">
      <div class="edition">DevHub 标准版</div>
      <div class="sub-title">{{ activeModule?.title || '控制台' }}</div>
      <template v-for="group in activeGroups" :key="group.key">
        <div class="sub-group-title" :class="{ active: group.key === activeNavGroupKey }" :data-testid="`admin-sub-nav-group-${group.key}`">{{ group.title }}</div>
        <button
          v-for="child in group.items"
          :key="child.key"
          :class="['sub-item', { active: child.key === activeNavPageKey }]"
          @click="router.push(child.path)"
          :data-testid="`admin-sub-nav-${child.key}`"
        >
          {{ child.label }}
        </button>
      </template>
    </aside>

    <el-container class="workbench">
      <el-header class="header">
        <div class="header-left">
          <el-icon class="fold-icon"><Fold /></el-icon>
          <el-breadcrumb separator="/" class="breadcrumb" data-testid="admin-breadcrumb">
            <el-breadcrumb-item v-if="activeModule" data-testid="breadcrumb-module">{{ activeModule.title }}</el-breadcrumb-item>
            <el-breadcrumb-item v-if="activeNavGroup" data-testid="breadcrumb-group">{{ activeNavGroup.title }}</el-breadcrumb-item>
            <el-breadcrumb-item v-if="activeNavPage" data-testid="breadcrumb-page">{{ activeNavPage.label }}</el-breadcrumb-item>
            <el-breadcrumb-item v-if="showLeafBreadcrumb" data-testid="breadcrumb-leaf">{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
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
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { ArrowDown, Bell, Briefcase, ChatDotRound, Connection, DataBoard, Document, Fold, FolderOpened, FullScreen, Grid, MagicStick, Notebook, PriceTag, Promotion, QuestionFilled, Refresh, Search, SetUp, Setting, Tickets, TrendCharts, User, UserFilled, Warning } from '@element-plus/icons-vue';
import { adminModuleGroups, adminModules } from '@/router/adminNav';
import { useAuthStore } from '@/stores/auth';
import { useTabsStore } from '@/stores/tabs';

const icons = { Briefcase, ChatDotRound, Connection, DataBoard, Document, FolderOpened, MagicStick, Notebook, PriceTag, Promotion, QuestionFilled, SetUp, Setting, Tickets, TrendCharts, User, UserFilled, Warning };

const auth = useAuthStore();
const tabs = useTabsStore();
const route = useRoute();
const router = useRouter();
const quickSearch = ref('');

function inferModuleKeyByPath(path) {
  if (path.startsWith('/plugins')) return 'plugins';
  if (path.startsWith('/users')) return 'users';
  if (path.startsWith('/communities') || path.startsWith('/moderators') || path.startsWith('/sites')) return 'communities';
  if (path.startsWith('/system') || path.startsWith('/operation') || path.startsWith('/statistics')) return 'system';
  if (path.startsWith('/audit-logs')) return 'maintenance';
  if (path.startsWith('/content') || path.startsWith('/comments') || path.startsWith('/reports') || path.startsWith('/tags')) return 'content';
  if (path.startsWith('/dashboard')) return 'home';
  return 'home';
}

const activeModuleKey = computed(() => route.meta?.moduleKey || inferModuleKeyByPath(route.path));
const visibleModules = computed(() => adminModules.filter((mod) => !mod.permission || auth.can(mod.permission)));
const activeModule = computed(() => adminModules.find((m) => m.key === activeModuleKey.value) || visibleModules.value[0] || adminModules[0]);
const activeGroups = computed(() => {
  const raw = adminModuleGroups[activeModuleKey.value] || [];
  return raw
    .map((group) => ({
      key: group.key,
      title: group.title,
      items: (group.items || []).filter((item) => !item.permission || auth.can(item.permission)),
    }))
    .filter((group) => group.items.length);
});
const showSubNav = computed(() => activeGroups.value.length > 0);

const activeNavPageKey = computed(() => {
  if (route.meta?.navPageKey) return route.meta.navPageKey;
  const matched = activeGroups.value.flatMap((g) => g.items).find((item) => item.path === route.path);
  return matched?.key || '';
});
const activeNavGroupKey = computed(() => {
  if (route.meta?.navGroupKey) return route.meta.navGroupKey;
  const group = activeGroups.value.find((g) => g.items.some((item) => item.key === activeNavPageKey.value));
  return group?.key || activeGroups.value[0]?.key || '';
});
const activeNavGroup = computed(() => activeGroups.value.find((g) => g.key === activeNavGroupKey.value) || null);
const activeNavPage = computed(() => activeNavGroup.value?.items?.find((item) => item.key === activeNavPageKey.value) || null);

const currentPageTitle = computed(() => {
  const title = route.meta?.breadcrumbTitle || route.meta?.title || activeNavPage.value?.label || '控制台';
  return title;
});
const showLeafBreadcrumb = computed(() => {
  if (!currentPageTitle.value) return false;
  if (!activeNavPage.value?.label) return true;
  return currentPageTitle.value !== activeNavPage.value.label;
});
const keepAliveNames = computed(() => router.getRoutes().filter((r) => r.meta?.keepAlive && r.name).map((r) => r.name));
const scopeLabel = computed(() => {
  const site = new URLSearchParams(window.location.search).get('site') || 'portal';
  return site === 'portal' ? '全局后台' : `${site} 子站`;
});

watch(route, (value) => {
  tabs.add(value);
}, { immediate: true });
watch(quickSearch, (value) => {
  if (!value) return;
  const keyword = value.trim();
  if (!keyword) return;
  const pages = Object.values(adminModuleGroups).flatMap((groups) => groups.flatMap((g) => g.items || []));
  const matched = pages.find((p) => (p.label || '').includes(keyword));
  if (matched) ElMessage.info(`可进入：${matched.label}`);
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
</script>
